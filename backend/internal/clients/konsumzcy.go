package clients

import (
	"context"
	"fmt"
)

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
	DateOfBirth   *string `json:"date_of_birth,omitempty"`
	OrderCount    int     `json:"order_count"`
	TotalSpend    float64 `json:"total_spend"`
	LastOrderDate *string `json:"last_order_date,omitempty"`
	// Menu is a BFF-only overlay (not a KonsumZcy field): the free drink this
	// customer chose, derived from AgregaZcy. Empty if they never claimed.
	Menu string `json:"menu,omitempty"`
	// Queue overlay (also BFF-only, derived from ticket events): the customer's
	// queue number and live production status.
	QueueNumber int    `json:"queue_number,omitempty"`
	QueueStatus string `json:"queue_status,omitempty"` // waiting|ready|done
}

type customerEnvelope struct {
	Data Customer `json:"data"`
}

// RegisterProfile upserts a member profile by phone, independent of any order.
// Calls KonsumZcy POST /api/customers. `address` carries domisili (a generic
// customer location field); `dateOfBirth` is the durable form of age (umur).
func (k *KonsumZcy) RegisterProfile(ctx context.Context, phone, name string, email, dateOfBirth, address *string) (*Customer, error) {
	body := map[string]any{"phone": phone, "name": name}
	if email != nil {
		body["email"] = *email
	}
	if dateOfBirth != nil {
		body["date_of_birth"] = *dateOfBirth
	}
	if address != nil {
		body["address"] = *address
	}
	var env customerEnvelope
	if err := k.h.do(ctx, "KonsumZcy", "POST", "/api/customers", body, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// RawGet proxies an arbitrary GET to KonsumZcy and decodes into out.
func (k *KonsumZcy) RawGet(ctx context.Context, path string, out any) error {
	return k.h.do(ctx, "KonsumZcy", "GET", path, nil, out)
}

// CustomerList mirrors KonsumZcy's paginated customer list response.
type CustomerList struct {
	Data  []Customer `json:"data"`
	Total int64      `json:"total"`
	Limit int        `json:"limit"`
	Skip  int        `json:"skip"`
}

// ListCustomers fetches the member list (typed), forwarding the query string.
func (k *KonsumZcy) ListCustomers(ctx context.Context, rawQuery string) (*CustomerList, error) {
	path := "/api/customers"
	if rawQuery != "" {
		path += "?" + rawQuery
	}
	var out CustomerList
	if err := k.h.do(ctx, "KonsumZcy", "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAllCustomers fetches EVERY customer by paging. KonsumZcy clamps `limit` to
// 100, so a single request silently truncates once there are >100 members —
// page through with offset until Total is collected.
func (k *KonsumZcy) ListAllCustomers(ctx context.Context) (*CustomerList, error) {
	const page = 100 // KonsumZcy's hard per-request cap
	all := &CustomerList{}
	for offset := 0; ; offset += page {
		var out CustomerList
		path := fmt.Sprintf("/api/customers?limit=%d&offset=%d", page, offset)
		if err := k.h.do(ctx, "KonsumZcy", "GET", path, nil, &out); err != nil {
			return nil, err
		}
		all.Data = append(all.Data, out.Data...)
		all.Total = out.Total
		if len(out.Data) == 0 || int64(len(all.Data)) >= out.Total {
			break
		}
	}
	all.Limit = len(all.Data)
	return all, nil
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
