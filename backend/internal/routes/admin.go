package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ── GET /api/admin/members ────────────────────────────────────────────────────
// Proxies KonsumZcy's customer list (operational source of truth).
func (h *Handlers) ListMembers(c *gin.Context) {
	// KonsumZcy list endpoint is GET /api/customers?search=&limit=&offset=.
	// Forward the query string verbatim.
	path := "/api/customers"
	if q := c.Request.URL.RawQuery; q != "" {
		path += "?" + q
	}
	var out any
	if err := h.Konsum.RawGet(c.Request.Context(), path, &out); err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
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
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name"`
	Limit       int    `json:"limit"`        // total redemptions (e.g. 200)
	PerCustomer int    `json:"per_customer"` // default 1
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
	body := map[string]any{
		"code": req.Code,
		"type": "fixed",
		"discount": map[string]any{
			"type":     "percentage",
			"value":    100,
			"apply_to": "line_items",
		},
		"usage":    map[string]any{"max_total": req.Limit, "max_per_customer": perCustomer},
		"metadata": map[string]any{"campaign": req.Name, "label": req.Name},
	}
	promo, err := h.Promo.CreatePromo(c.Request.Context(), body)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaign": promo})
}
