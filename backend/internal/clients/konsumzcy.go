package clients

import "context"

// KonsumZcy is the member store. Its Customer entity carries the data we collect
// plus lifetime tracking (order_count → visits, total_spend → spend).
type KonsumZcy struct{ h httpDo }

func NewKonsumZcy(base string) *KonsumZcy { return &KonsumZcy{h: newHTTP(base)} }

// Customer mirrors KonsumZcy's customer model (the fields the BFF cares about).
type Customer struct {
	ID            string  `json:"id"`
	CustomerID    string  `json:"customer_id"`
	Phone         string  `json:"phone"`
	Name          string  `json:"name"`
	Email         *string `json:"email,omitempty"`
	Address       *string `json:"address,omitempty"`
	OrderCount    int     `json:"order_count"`
	TotalSpend    float64 `json:"total_spend"`
	LastOrderDate *string `json:"last_order_date,omitempty"`
}

type customerEnvelope struct {
	Data Customer `json:"data"`
}

// RegisterProfile upserts a member profile by phone, independent of any order.
// Calls KonsumZcy POST /api/customers.
func (k *KonsumZcy) RegisterProfile(ctx context.Context, phone, name string, email *string) (*Customer, error) {
	body := map[string]any{"phone": phone, "name": name}
	if email != nil {
		body["email"] = *email
	}
	var env customerEnvelope
	if err := k.h.do(ctx, "KonsumZcy", "POST", "/api/customers", body, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// RawGet proxies an arbitrary GET to KonsumZcy and decodes into out. Used by the
// admin members list, which forwards the upstream query string verbatim.
func (k *KonsumZcy) RawGet(ctx context.Context, path string, out any) error {
	return k.h.do(ctx, "KonsumZcy", "GET", path, nil, out)
}

// GetByPhone fetches a member profile by phone for the barista lookup.
// Calls KonsumZcy GET /api/customers-by-phone?phone=...
func (k *KonsumZcy) GetByPhone(ctx context.Context, phone string) (*Customer, error) {
	var env customerEnvelope
	if err := k.h.do(ctx, "KonsumZcy", "GET", "/api/customers-by-phone?phone="+urlEscape(phone), nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
