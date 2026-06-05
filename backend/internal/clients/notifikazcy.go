package clients

import "context"

// NotifikaZcy delivers the voucher code to the member (Resend email today, WhatsApp
// later via InterakZcy). Template-driven: the BFF sends a template slug + data bag.
type NotifikaZcy struct {
	h       httpDo
	enabled bool
}

func NewNotifikaZcy(base string, enabled bool) *NotifikaZcy {
	return &NotifikaZcy{h: newHTTP(base), enabled: enabled}
}

// SendVoucher emails a member their voucher code. When delivery is disabled
// (NOTIFY_ENABLED=false), it's a no-op so the BFF can return the code in-band
// instead — lets us ship before Resend is wired.
func (n *NotifikaZcy) SendVoucher(ctx context.Context, toEmail, name, code, expiresAt string) error {
	if !n.enabled || toEmail == "" {
		return nil
	}
	body := map[string]any{
		"template": "voucher-issued",
		"to":       toEmail,
		"data": map[string]any{
			"name":      name,
			"code":      code,
			"expiresAt": expiresAt,
		},
	}
	return n.h.do(ctx, "NotifikaZcy", "POST", "/api/v1/send", body, nil)
}
