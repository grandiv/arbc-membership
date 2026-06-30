package routes

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/KreaZcy/arbc-membership-backend/internal/clients"
)

// ── GET /api/admin/members ────────────────────────────────────────────────────
// Members come from KonsumZcy (identity), but their visit/spend METRICS are
// derived from AgregaZcy (the metrics engine) and merged in — KonsumZcy holds no
// metrics for this product. Clean separation: identity vs. analytics.
func (h *Handlers) ListMembers(c *gin.Context) {
	ctx := c.Request.Context()
	// Page through ALL customers — KonsumZcy caps a single request at 100, so a
	// naive fetch silently truncates the dashboard once past 100 pendaftar.
	list, err := h.Konsum.ListAllCustomers(ctx)
	if err != nil {
		fail(c, err)
		return
	}
	// Overlay AgregaZcy-derived visit/spend/menu (best-effort; zeros if it fails).
	if stats, err := h.Agrega.MemberStats(ctx); err == nil {
		for i := range list.Data {
			if m, ok := stats[list.Data[i].Phone]; ok {
				list.Data[i].OrderCount = m.Visits
				list.Data[i].TotalSpend = m.Spend
				list.Data[i].Menu = m.Menu
			}
		}
	}
	// Overlay the queue number + live production status, keyed by phone (one
	// claim per phone, so one ticket; if somehow more, the highest number wins).
	if events, err := h.Agrega.ListTickets(ctx); err == nil {
		byPhone := map[string]*QueueTicket{}
		for _, t := range foldTickets(events) {
			if t.Phone == "" {
				continue
			}
			if ex, ok := byPhone[t.Phone]; !ok || t.Number > ex.Number {
				byPhone[t.Phone] = t
			}
		}
		for i := range list.Data {
			if t, ok := byPhone[list.Data[i].Phone]; ok {
				list.Data[i].QueueNumber = t.Number
				list.Data[i].QueueStatus = t.Status
			}
		}
	}
	c.JSON(http.StatusOK, list)
}

// ── GET /api/admin/menu-stats ─────────────────────────────────────────────────
// Which free drink is more popular. Derived from AgregaZcy's claim events.
func (h *Handlers) MenuStats(c *gin.Context) {
	tally, err := h.Agrega.MenuTally(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	if tally == nil {
		tally = map[string]int{}
	}
	c.JSON(http.StatusOK, gin.H{"data": tally})
}

// ── Production queue (POS-style board) ────────────────────────────────────────
// The production house's live board, folded from ticket.* events. Brand-shaped
// (number/name/menu) here in the BFF; AgregaZcy only ever sees neutral events.

// QueueTicket is one open ticket on the production board.
type QueueTicket struct {
	TicketID  string `json:"ticketId"`
	Number    int    `json:"number"`
	Name      string `json:"name"`
	Menu      string `json:"menu"`
	Phone     string `json:"phone"`
	Status    string `json:"status"` // "waiting" | "ready" | "done"
	CreatedAt string `json:"createdAt"`
}

// foldTickets collapses the ticket.* event stream into one final state per
// ticketId (newest status wins; done is terminal). Shared by the board and the
// pendaftar overlay.
func foldTickets(events []clients.TicketEvent) map[string]*QueueTicket {
	m := map[string]*QueueTicket{}
	created := map[string]bool{}
	for _, e := range events {
		t := m[e.TicketID]
		if t == nil {
			t = &QueueTicket{TicketID: e.TicketID, Status: "waiting"}
			m[e.TicketID] = t
		}
		switch e.Action {
		case "ticket.created":
			created[e.TicketID] = true
			t.CreatedAt = e.Timestamp
			if v, ok := e.Metadata["number"].(float64); ok {
				t.Number = int(v)
			}
			if v, ok := e.Metadata["name"].(string); ok {
				t.Name = v
			}
			if v, ok := e.Metadata["menu"].(string); ok {
				t.Menu = v
			}
			if v, ok := e.Metadata["phone"].(string); ok {
				t.Phone = v
			}
		case "ticket.ready":
			if t.Status != "done" {
				t.Status = "ready"
			}
		case "ticket.done":
			t.Status = "done"
		}
	}
	// Drop tickets that never had a created event (shouldn't happen, defensive).
	for id := range m {
		if !created[id] {
			delete(m, id)
		}
	}
	return m
}

// GET /api/admin/queue — open tickets (created/ready, not done), lowest number first.
func (h *Handlers) ProductionQueue(c *gin.Context) {
	events, err := h.Agrega.ListTickets(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	out := make([]QueueTicket, 0)
	for _, t := range foldTickets(events) {
		if t.Status == "done" {
			continue
		}
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	c.JSON(http.StatusOK, gin.H{"data": out})
}

// POST /api/admin/queue/:id/ready — drink is made; call the customer.
func (h *Handlers) TicketReady(c *gin.Context) {
	h.Agrega.EmitFor("ticket.ready", "ticket", c.Param("id"), map[string]any{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/admin/queue/:id/done — handed over; clear from the board.
func (h *Handlers) TicketDone(c *gin.Context) {
	h.Agrega.EmitFor("ticket.done", "ticket", c.Param("id"), map[string]any{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── GET /api/admin/campaigns ──────────────────────────────────────────────────
func (h *Handlers) ListCampaigns(c *gin.Context) {
	out, err := h.Promo.ListPromos(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// ── POST /api/admin/campaigns ─────────────────────────────────────────────────
// Create a campaign like "200 free coffee". The admin sends the brand-shaped
// fields; the BFF translates to PromoZcy's neutral promo model.
type createCampaignRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name"`
	Limit       int     `json:"limit"`         // total redemptions (e.g. 200); 0 = unlimited
	PerCustomer int     `json:"per_customer"`  // default 1
	DiscountType string `json:"discount_type"` // "free" (default) | "percent" | "fixed"
	DiscountValue float64 `json:"discount_value"` // % for percent, Rp for fixed
}

func (h *Handlers) CreateCampaign(c *gin.Context) {
	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	// Guard against a SECOND active campaign. Two campaigns each carry their own
	// usage cap, which silently doubles the giveaway and splits the dashboard
	// stats. Refuse if one is already running — the accidental KOPI* duplicate
	// (created during a deploy blip) must not happen again.
	if existing, err := h.Promo.ActiveCampaign(c.Request.Context()); err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"code": "CAMPAIGN_EXISTS", "message": "sudah ada kampanye aktif — nonaktifkan dulu sebelum membuat yang baru"})
		return
	}
	perCustomer := req.PerCustomer
	if perCustomer <= 0 {
		perCustomer = 1
	}

	// Translate the brand-shaped discount into PromoZcy's neutral model.
	// "free" = 100% off the cup; "percent"/"fixed" use the given value.
	promoType, discType, discValue := "percentage", "percentage", req.DiscountValue
	switch req.DiscountType {
	case "fixed":
		promoType, discType = "fixed", "fixed"
	case "percent":
		// percentage off, value as-is
	default: // "free" or empty
		discValue = 100
	}

	usage := map[string]any{"max_per_customer": perCustomer}
	if req.Limit > 0 {
		usage["max_total"] = req.Limit
	}

	body := map[string]any{
		"code": req.Code,
		"type": promoType,
		"discount": map[string]any{
			"type":     discType,
			"value":    discValue,
			"apply_to": "subtotal",
		},
		"usage": usage,
		// kind=campaign marks this as the redeemable free-cup campaign the BFF
		// auto-selects for registration eligibility + phone redemption.
		"metadata": map[string]any{"campaign": req.Name, "label": req.Name, "kind": "campaign"},
	}
	promo, err := h.Promo.CreatePromo(c.Request.Context(), body)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaign": promo})
}
