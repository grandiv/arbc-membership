package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KreaZcy/arbc-membership-backend/internal/clients"
)

// ── POST /api/claim ───────────────────────────────────────────────────────────
// The campaign's single staff action: capture the customer (Nama + HP) AND claim
// their free cup in one call. Register and redeem stay as generic primitives;
// "claim" is the campaign-specific orchestration that lives in the BFF.
//
// Forward-compatible: the customer is a real KonsumZcy record (the future member),
// keyed by phone, with other profile fields left null to be enriched later.
// Once-per-phone + the 200 cap are enforced by PromoZcy.
type claimRequest struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Domisili string `json:"domisili"` // → KonsumZcy address (generic location)
	Umur     int    `json:"umur"`     // age → derived date_of_birth (durable form)
}

func (h *Handlers) Claim(c *gin.Context) {
	var req claimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	ctx := c.Request.Context()
	phone := strings.TrimSpace(req.Phone)
	name := strings.TrimSpace(req.Name)

	// Map brand fields → generic KonsumZcy fields (no engine change):
	//   Domisili → address; Umur → date_of_birth (Jan 1 of birth year).
	var domisili, dob *string
	if d := strings.TrimSpace(req.Domisili); d != "" {
		domisili = &d
	}
	if req.Umur > 0 && req.Umur < 130 {
		y := time.Now().Year() - req.Umur
		s := fmt.Sprintf("%04d-01-01", y)
		dob = &s
	}

	// 1. Capture the data (always — this is the point; idempotent by phone).
	member, err := h.Konsum.RegisterProfile(ctx, phone, name, nil, dob, domisili)
	if err != nil {
		fail(c, err)
		return
	}

	// 2. The active campaign.
	camp, err := h.Promo.ActiveCampaign(ctx)
	if err != nil {
		fail(c, err)
		return
	}
	if camp == nil {
		c.JSON(http.StatusOK, gin.H{"claimed": false, "reason": "no_campaign", "member": member})
		return
	}

	// 3. Attempt the claim for this phone (PromoZcy enforces cap + once-per-phone).
	price := h.coffeePrice()
	res, err := h.Promo.Validate(ctx, camp.Code, phone, price, 0)
	if err != nil {
		fail(c, err)
		return
	}
	if !res.Valid {
		reason := "ineligible"
		switch {
		case strings.Contains(res.Message, "per-customer"):
			reason = "already_claimed"
		case strings.Contains(res.Message, "usage limit"):
			reason = "exhausted"
		}
		c.JSON(http.StatusOK, gin.H{"claimed": false, "reason": reason, "member": member, "remaining": camp.Remaining()})
		return
	}
	if err := h.Promo.Apply(ctx, camp.Code, phone, name, price); err != nil {
		fail(c, err)
		return
	}

	// 4. Analytics (best-effort). The cup is FREE, so the member's spend is 0;
	//    the giveaway VALUE (cup price) is recorded on the voucher.redeemed event,
	//    not as member spend — keeps total_spend honest for future membership.
	h.Agrega.EmitFor("voucher.redeemed", "promo", camp.Code, map[string]any{"value": price, "campaign": camp.Code})
	h.Agrega.EmitFor("visit.recorded", "member", phone, map[string]any{"amount": 0})

	remaining := camp.Remaining()
	if remaining > 0 {
		remaining-- // reflect this claim
	}
	c.JSON(http.StatusOK, gin.H{"claimed": true, "member": member, "remaining": remaining})
}

// ── POST /api/register ────────────────────────────────────────────────────────
// The data-for-value loop: capture member data → issue a voucher → deliver it.

type registerRequest struct {
	Phone    string  `json:"phone" binding:"required"`
	Name     string  `json:"name" binding:"required"`
	Email    *string `json:"email,omitempty"`
	IGHandle *string `json:"ig_handle,omitempty"`
	Dob      *string `json:"dob,omitempty"` // ISO "YYYY-MM-DD"; for birthday promos
}

// RegisterMember does ONE thing: create/update the member profile. Campaign
// perks (free cup, etc.) are a separate concern — see GET /api/campaign and the
// phone-redemption flow. There is no code; the member's phone is their identity.
func (h *Handlers) RegisterMember(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	ctx := c.Request.Context()

	member, err := h.Konsum.RegisterProfile(ctx, req.Phone, req.Name, req.Email, req.Dob, nil)
	if err != nil {
		fail(c, err)
		return
	}

	// Analytics (best-effort, never blocks).
	meta := map[string]any{}
	if req.IGHandle != nil {
		meta["ig_handle"] = *req.IGHandle
	}
	// Tag the birth month (e.g. "07") so birthday cohorts are queryable.
	if req.Dob != nil && len(*req.Dob) >= 7 {
		meta["birth_month"] = (*req.Dob)[5:7]
	}
	h.Agrega.Emit("member.registered", "member", meta)

	c.JSON(http.StatusOK, gin.H{"member": member})
}

// ── GET /api/campaign ─────────────────────────────────────────────────────────
// The active free-cup campaign (if any). A separate concern from registration:
// the /join page reads this to show the perk; the booth redeems by phone.
func (h *Handlers) ActiveCampaignInfo(c *gin.Context) {
	camp, err := h.Promo.ActiveCampaign(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	if camp == nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}
	remaining := camp.Remaining()
	c.JSON(http.StatusOK, gin.H{
		"active":    remaining != 0,
		"label":     campaignLabel(camp),
		"remaining": remaining,
	})
}

// campaignLabel pulls a human label off a campaign promo's metadata.
func campaignLabel(p *clients.Promo) string {
	if p == nil {
		return ""
	}
	if l, ok := p.Metadata["label"].(string); ok && l != "" {
		return l
	}
	return p.Code
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
	// Tell the barista whether this member can still claim the active campaign.
	camp, _ := h.Promo.ActiveCampaign(c.Request.Context())
	freeCup := gin.H{"eligible": false}
	if camp != nil {
		freeCup = gin.H{"eligible": camp.Remaining() != 0, "label": campaignLabel(camp), "remaining": camp.Remaining()}
	}
	c.JSON(http.StatusOK, gin.H{"member": member, "freeCup": freeCup})
}

// ── POST /api/redeem ──────────────────────────────────────────────────────────
// Barista redeems by phone against the active campaign (default), or by an
// explicit code. Amount is the real order total (defaults to the base cup price)
// so the discount + spend analytics reflect what was actually ordered.

type redeemRequest struct {
	Phone  string   `json:"phone"`
	Code   string   `json:"code"`           // optional; defaults to the active campaign
	Name   string   `json:"name"`
	Amount *float64 `json:"amount"`         // optional; defaults to COFFEE_PRICE
}

func (h *Handlers) Redeem(c *gin.Context) {
	var req redeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	ctx := c.Request.Context()

	// Resolve the code: explicit code wins, else the active campaign.
	code := strings.TrimSpace(req.Code)
	if code == "" {
		camp, err := h.Promo.ActiveCampaign(ctx)
		if err != nil {
			fail(c, err)
			return
		}
		if camp == nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"code": "NO_CAMPAIGN", "message": "tidak ada kampanye aktif"})
			return
		}
		code = camp.Code
	}
	if req.Phone == "" && req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "phone or code required"})
		return
	}

	// Real order amount (what the customer ordered); default to the base price.
	amount := h.coffeePrice()
	if req.Amount != nil && *req.Amount > 0 {
		amount = *req.Amount
	}

	// 1. Validate against the purchase.
	res, err := h.Promo.Validate(ctx, code, req.Phone, amount, 0)
	if err != nil {
		fail(c, err)
		return
	}
	if !res.Valid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": "VOUCHER_INVALID", "message": res.Message})
		return
	}

	// 2. Record the redemption.
	if err := h.Promo.Apply(ctx, code, req.Phone, req.Name, amount); err != nil {
		fail(c, err)
		return
	}

	// 3. Analytics — AgregaZcy is the metrics ledger. visit.recorded carries the
	//    member phone as target so per-member visit/spend can be derived from it
	//    (KonsumZcy stays a pure identity store).
	h.Agrega.EmitFor("voucher.redeemed", "promo", code, map[string]any{"amount": amount, "discount": res.DiscountAmount})
	if req.Phone != "" {
		h.Agrega.EmitFor("visit.recorded", "member", req.Phone, map[string]any{"amount": amount})
	}

	c.JSON(http.StatusOK, gin.H{
		"redeemed":       true,
		"discountAmount": res.DiscountAmount,
		"orderAmount":    amount,
	})
}
