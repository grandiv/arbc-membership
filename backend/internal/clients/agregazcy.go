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

const agregaTenant = "tana-arabica"

// Emit fires an event with no specific target. See EmitFor.
func (a *AgregaZcy) Emit(action, targetEntity string, metadata map[string]any) {
	a.EmitFor(action, targetEntity, "", metadata)
}

// EmitFor fires a single event to AgregaZcy POST /api/v1/internal/events with a
// targetId (e.g. the member's phone) so per-target metrics can be derived later.
// BEST-EFFORT: failures are logged + swallowed; runs in its own goroutine.
func (a *AgregaZcy) EmitFor(action, targetEntity, targetID string, metadata map[string]any) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	if _, ok := metadata["campaign"]; !ok {
		metadata["campaign"] = a.campaign
	}
	// AgregaZcy's IngestRequest uses camelCase keys (targetEntity, targetId,
	// sourceService, tenantId) + a free-form metadata bag.
	body := map[string]any{
		"action":        action,
		"targetEntity":  targetEntity,
		"targetId":      targetID,
		"metadata":      metadata,
		"sourceService": "arbc-membership",
		"tenantId":      agregaTenant,
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.h.do(ctx, "AgregaZcy", "POST", "/api/v1/internal/events", body, nil); err != nil {
			slog.Warn("agregazcy emit failed (ignored)", "action", action, "err", err)
		}
	}()
}

// MemberMetric is the derived visit/spend rollup for one member (by phone).
type MemberMetric struct {
	Visits int
	Spend  float64
}

// MemberStats derives per-member visit count + total spend from the
// visit.recorded events AgregaZcy already stores — KonsumZcy holds no metrics.
// Returns a map keyed by member phone (the event targetId).
func (a *AgregaZcy) MemberStats(ctx context.Context) (map[string]MemberMetric, error) {
	var env struct {
		Data struct {
			Events []struct {
				TargetID string         `json:"targetId"`
				Metadata map[string]any `json:"metadata"`
			} `json:"events"`
		} `json:"data"`
	}
	// Pull visit events for this tenant (high limit; small-scale product).
	path := "/api/v1/internal/events?tenantId=" + agregaTenant + "&action=visit.recorded&limit=10000"
	if err := a.h.do(ctx, "AgregaZcy", "GET", path, nil, &env); err != nil {
		return nil, err
	}
	out := make(map[string]MemberMetric)
	for _, e := range env.Data.Events {
		if e.TargetID == "" {
			continue
		}
		m := out[e.TargetID]
		m.Visits++
		if amt, ok := e.Metadata["amount"].(float64); ok {
			m.Spend += amt
		}
		out[e.TargetID] = m
	}
	return out, nil
}
