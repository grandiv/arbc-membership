package routes

import (
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ── POST /api/register ────────────────────────────────────────────────────────
// The data-for-value loop: capture member data → issue a voucher → deliver it.

type registerRequest struct {
	Phone    string  `json:"phone" binding:"required"`
	Name     string  `json:"name" binding:"required"`
	Email    *string `json:"email,omitempty"`
	IGHandle *string `json:"ig_handle,omitempty"`
	Dob      *string `json:"dob,omitempty"` // ISO "YYYY-MM-DD"; for birthday promos
}

// RegisterMember upserts the member in KonsumZcy, issues a per-member voucher in
// PromoZcy, fires it off via NotifikaZcy, and logs analytics events to AgregaZcy.
func (h *Handlers) RegisterMember(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	ctx := c.Request.Context()

	// 1. Store the data (the thing we actually want).
	member, err := h.Konsum.RegisterProfile(ctx, req.Phone, req.Name, req.Email, req.Dob)
	if err != nil {
		fail(c, err)
		return
	}

	// 2. Issue a single-use voucher tied to this member.
	code := genCode("ARBC")
	campaign := h.Cfg.DefaultCampaign
	promoBody := map[string]any{
		"code": code,
		"type": "fixed",
		"discount": map[string]any{
			"type":     "percentage",
			"value":    100, // 100% off one coffee = free
			"apply_to": "line_items",
		},
		"usage":    map[string]any{"max_total": 1, "max_per_customer": 1},
		"metadata": map[string]any{"campaign": campaign, "member_phone": req.Phone},
	}
	if _, err := h.Promo.CreatePromo(ctx, promoBody); err != nil {
		// The member is already saved; surface the voucher failure but don't lose data.
		fail(c, err)
		return
	}

	// 3. Deliver the code (no-op when NOTIFY_ENABLED=false → returned in-band below).
	var emailVal string
	if req.Email != nil {
		emailVal = *req.Email
	}
	_ = h.Notify.SendVoucher(ctx, emailVal, req.Name, code, "")

	// 4. Analytics (best-effort, never blocks).
	meta := map[string]any{"campaign": campaign}
	if req.IGHandle != nil {
		meta["ig_handle"] = *req.IGHandle
	}
	// Tag the birth month (e.g. "07") so birthday cohorts are queryable for
	// monthly birthday-promo campaigns, without exposing the full DOB.
	if req.Dob != nil && len(*req.Dob) >= 7 {
		meta["birth_month"] = (*req.Dob)[5:7]
	}
	h.Agrega.Emit("member.registered", "member", meta)
	h.Agrega.Emit("voucher.issued", "promo", map[string]any{"campaign": campaign, "code": code})

	c.JSON(http.StatusOK, gin.H{
		"member": member,
		"voucher": gin.H{
			"code":      code,
			"delivered": h.Cfg.NotifyEnabled && emailVal != "",
		},
	})
}

// ── POST /api/lookup ──────────────────────────────────────────────────────────
// Barista types a phone → see the member + what they're eligible for.

type lookupRequest struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *Handlers) Lookup(c *gin.Context) {
	var req lookupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	member, err := h.Konsum.GetByPhone(c.Request.Context(), req.Phone)
	if err != nil {
		fail(c, err)
		return
	}
	// Eligible-promo listing is a future enhancement (needs a PromoZcy
	// by-member query); for now we return the member so the barista can confirm.
	c.JSON(http.StatusOK, gin.H{"member": member, "eligiblePromos": []any{}})
}

// ── POST /api/redeem ──────────────────────────────────────────────────────────
// Barista redeems either a campaign code or (future) a phone-eligible promo.

type redeemRequest struct {
	Code  string `json:"code" binding:"required"`
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

func (h *Handlers) Redeem(c *gin.Context) {
	var req redeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	ctx := c.Request.Context()
	price := h.coffeePrice()

	// 1. Validate the code against the purchase.
	res, err := h.Promo.Validate(ctx, req.Code, req.Phone, price, 0)
	if err != nil {
		fail(c, err)
		return
	}
	if !res.Valid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": "VOUCHER_INVALID", "message": res.Message})
		return
	}

	// 2. Record the redemption.
	if err := h.Promo.Apply(ctx, req.Code, req.Phone, req.Name, price); err != nil {
		fail(c, err)
		return
	}

	// 3. Analytics.
	h.Agrega.Emit("voucher.redeemed", "promo", map[string]any{"code": req.Code, "amount": price})
	if req.Phone != "" {
		h.Agrega.Emit("visit.recorded", "member", map[string]any{"amount": price})
	}

	c.JSON(http.StatusOK, gin.H{
		"redeemed":       true,
		"discountAmount": res.DiscountAmount,
	})
}

// genCode returns a short, human-typeable voucher code like "ARBC-7K2QM9".
func genCode(prefix string) string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	s := strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
	return prefix + "-" + s
}
