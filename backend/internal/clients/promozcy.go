package clients

import "context"

// PromoZcy is the voucher / promo engine. The BFF speaks its neutral promo model;
// brand campaigns ("200 free coffee") are just promos with a usage cap + expiry.
type PromoZcy struct{ h httpDo }

func NewPromoZcy(base string) *PromoZcy { return &PromoZcy{h: newHTTP(base)} }

// Promo is a partial view of PromoZcy's promo model.
type Promo struct {
	ID         string         `json:"id"`
	Code       string         `json:"code"`
	Type       string         `json:"type"`
	IsActive   bool           `json:"is_active"`
	UsageCount int            `json:"usage_count"`
	Usage      struct {
		MaxTotal       *int `json:"max_total,omitempty"`
		MaxPerCustomer *int `json:"max_per_customer,omitempty"`
	} `json:"usage"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
}

// Remaining returns how many redemptions are left (-1 = unlimited).
func (p *Promo) Remaining() int {
	if p.Usage.MaxTotal == nil {
		return -1
	}
	r := *p.Usage.MaxTotal - p.UsageCount
	if r < 0 {
		r = 0
	}
	return r
}

type promoEnvelope struct {
	Data Promo `json:"data"`
}

// ValidateResult is PromoZcy's response to a validate call (shape kept loose).
type ValidateResult struct {
	Valid          bool    `json:"valid"`
	DiscountAmount float64 `json:"discount_amount"`
	Message        string  `json:"message,omitempty"`
}

// CreatePromo registers a new promo/campaign. Calls POST /api/promos.
// The caller passes the full neutral PromoZcy CreatePromoRequest body.
func (p *PromoZcy) CreatePromo(ctx context.Context, body map[string]any) (*Promo, error) {
	var env promoEnvelope
	if err := p.h.do(ctx, "PromoZcy", "POST", "/api/v1/admin/promos", body, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// ListPromos returns the raw promo list (used by the admin dashboard proxy).
func (p *PromoZcy) ListPromos(ctx context.Context) (any, error) {
	var out any
	if err := p.h.do(ctx, "PromoZcy", "GET", "/api/v1/admin/promos", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// listTyped decodes the promo list into typed Promo structs.
func (p *PromoZcy) listTyped(ctx context.Context) ([]Promo, error) {
	var env struct {
		Data []Promo `json:"data"`
	}
	if err := p.h.do(ctx, "PromoZcy", "GET", "/api/v1/admin/promos", nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// ActiveCampaign returns the current free-cup campaign: the newest active promo
// tagged kind=campaign. Returns nil (no error) when none is configured. This is
// what registration checks for eligibility and what phone-redeem applies.
func (p *PromoZcy) ActiveCampaign(ctx context.Context) (*Promo, error) {
	promos, err := p.listTyped(ctx)
	if err != nil {
		return nil, err
	}
	var best *Promo
	for i := range promos {
		pr := promos[i]
		if !pr.IsActive {
			continue
		}
		if kind, _ := pr.Metadata["kind"].(string); kind != "campaign" {
			continue
		}
		// Newest wins (CreatedAt is RFC3339, lexically sortable).
		if best == nil || pr.CreatedAt > best.CreatedAt {
			cp := pr
			best = &cp
		}
	}
	return best, nil
}

// Validate checks a voucher code against a purchase. Calls POST /api/promos/validate.
func (p *PromoZcy) Validate(ctx context.Context, code, phone string, subtotal float64, orderCount int) (*ValidateResult, error) {
	body := map[string]any{
		"code":                 code,
		"subtotal":             subtotal,
		"customer_phone":       phone,
		"customer_order_count": orderCount,
	}
	var env struct {
		Data ValidateResult `json:"data"`
	}
	if err := p.h.do(ctx, "PromoZcy", "POST", "/api/v1/internal/promos/validate", body, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Apply records a redemption against a code. Calls POST /internal/promos/apply.
func (p *PromoZcy) Apply(ctx context.Context, code, phone, name string, subtotal float64) error {
	body := map[string]any{
		"promoCode":     code,
		"customerPhone": phone,
		"customerName":  name,
		"subtotal":      subtotal,
	}
	return p.h.do(ctx, "PromoZcy", "POST", "/api/v1/internal/promos/apply", body, nil)
}
