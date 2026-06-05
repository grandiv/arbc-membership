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
	Metadata   map[string]any `json:"metadata,omitempty"`
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

// ListPromos returns all promos/campaigns. Calls GET /api/promos.
func (p *PromoZcy) ListPromos(ctx context.Context) (any, error) {
	var out any
	if err := p.h.do(ctx, "PromoZcy", "GET", "/api/v1/admin/promos", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
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
