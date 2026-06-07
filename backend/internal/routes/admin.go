package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ── GET /api/admin/members ────────────────────────────────────────────────────
// Members come from KonsumZcy (identity), but their visit/spend METRICS are
// derived from AgregaZcy (the metrics engine) and merged in — KonsumZcy holds no
// metrics for this product. Clean separation: identity vs. analytics.
func (h *Handlers) ListMembers(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.Konsum.ListCustomers(ctx, c.Request.URL.RawQuery)
	if err != nil {
		fail(c, err)
		return
	}
	// Overlay AgregaZcy-derived visit/spend (best-effort; zeros if it fails).
	if stats, err := h.Agrega.MemberStats(ctx); err == nil {
		for i := range list.Data {
			if m, ok := stats[list.Data[i].Phone]; ok {
				list.Data[i].OrderCount = m.Visits
				list.Data[i].TotalSpend = m.Spend
			}
		}
	}
	c.JSON(http.StatusOK, list)
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
