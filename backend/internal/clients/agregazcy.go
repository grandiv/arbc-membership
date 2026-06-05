package clients

import (
	"context"
	"log/slog"
	"time"
)

// AgregaZcy is the analytics layer: an append-only event timeline + forecasting.
// The BFF fires neutral events here; engines never know what a "member" is.
type AgregaZcy struct {
	h        httpDo
	campaign string // tags every event for attribution
}

func NewAgregaZcy(base, defaultCampaign string) *AgregaZcy {
	return &AgregaZcy{h: newHTTP(base), campaign: defaultCampaign}
}

// Emit fires a single event to AgregaZcy POST /api/v1/internal/events. It is
// BEST-EFFORT: failures are logged and swallowed so analytics can never block or
// fail a registration/redemption. Runs in its own goroutine with a short timeout.
func (a *AgregaZcy) Emit(action, targetEntity string, metadata map[string]any) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	if _, ok := metadata["campaign"]; !ok {
		metadata["campaign"] = a.campaign
	}
	// AgregaZcy's IngestRequest uses camelCase keys (targetEntity, sourceService,
	// tenantId) and a free-form metadata bag. tenantId is the brand for this
	// single-tenant product.
	body := map[string]any{
		"action":        action,
		"targetEntity":  targetEntity,
		"metadata":      metadata,
		"sourceService": "arbc-membership",
		"tenantId":      "tana-arabica",
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.h.do(ctx, "AgregaZcy", "POST", "/api/v1/internal/events", body, nil); err != nil {
			slog.Warn("agregazcy emit failed (ignored)", "action", action, "err", err)
		}
	}()
}
