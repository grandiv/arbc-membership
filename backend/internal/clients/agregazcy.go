package clients

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
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

// eventBody builds the AgregaZcy IngestRequest. It uses camelCase keys
// (targetEntity, targetId, sourceService, tenantId) + a free-form metadata bag.
func (a *AgregaZcy) eventBody(action, targetEntity, targetID string, metadata map[string]any) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}
	if _, ok := metadata["campaign"]; !ok {
		metadata["campaign"] = a.campaign
	}
	return map[string]any{
		"action":        action,
		"targetEntity":  targetEntity,
		"targetId":      targetID,
		"metadata":      metadata,
		"sourceService": "arbc-membership",
		"tenantId":      agregaTenant,
	}
}

// EmitFor fires a single event to AgregaZcy POST /api/v1/internal/events with a
// targetId (e.g. the member's phone) so per-target metrics can be derived later.
// BEST-EFFORT: failures are logged + swallowed; runs in its own goroutine.
func (a *AgregaZcy) EmitFor(action, targetEntity, targetID string, metadata map[string]any) {
	body := a.eventBody(action, targetEntity, targetID, metadata)
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
	Menu   string // the menu chosen on their (most recent) claim, if recorded
}

// rawEvent is one event as read back from AgregaZcy's timeline.
type rawEvent struct {
	Action    string         `json:"action"`
	TargetID  string         `json:"targetId"`
	Timestamp string         `json:"timestamp"`
	Metadata  map[string]any `json:"metadata"`
}

// agregaPageSize is AgregaZcy's hard cap per query: it silently clamps page_size
// to 200 regardless of what we ask. So any read that can exceed 200 events MUST
// paginate — see fetchAllEvents. (At 1 event/claim that's fine until 200 claims;
// ticket reads hit it at ~66 claims since each claim emits 3 ticket events.)
const agregaPageSize = 200

// fetchAllEvents pages through every event matching filterQuery (e.g.
// "action=visit.recorded" or "target_entity=ticket") until totalCount is
// collected, defeating AgregaZcy's 200-per-page cap. tenant_id is added here.
func (a *AgregaZcy) fetchAllEvents(ctx context.Context, filterQuery string) ([]rawEvent, error) {
	var all []rawEvent
	for page := 1; ; page++ {
		var env struct {
			Data struct {
				Events     []rawEvent `json:"events"`
				TotalCount int64      `json:"totalCount"`
			} `json:"data"`
		}
		path := fmt.Sprintf("/api/v1/internal/events?tenant_id=%s&%s&page_size=%d&page=%d",
			agregaTenant, filterQuery, agregaPageSize, page)
		if err := a.h.do(ctx, "AgregaZcy", "GET", path, nil, &env); err != nil {
			return nil, err
		}
		all = append(all, env.Data.Events...)
		// Stop when we've collected everything, or a page came back empty (guard
		// against an off-by-one or a shrinking dataset between pages).
		if len(env.Data.Events) == 0 || int64(len(all)) >= env.Data.TotalCount {
			break
		}
	}
	return all, nil
}

// visitEvents pulls ALL visit.recorded events for this tenant (paginated). They
// carry the member phone as targetId. NOTE: AgregaZcy's internal query parses
// snake_case params (tenant_id/action/page_size); camelCase or `limit` are
// silently ignored.
func (a *AgregaZcy) visitEvents(ctx context.Context) ([]rawEvent, error) {
	return a.fetchAllEvents(ctx, "action=visit.recorded")
}

// ── Production queue (POS-style ticket board) ─────────────────────────────────
// The free-cup claim enqueues a ticket; the production house works the board and
// "calls" customers. The ticket lifecycle is modelled as neutral AgregaZcy events
// (ticket.created → ticket.ready → ticket.done) keyed by ticketId (targetId), so
// no new engine/store is needed and the engine stays brand-agnostic.

// TicketEvent is one ticket lifecycle event read back from the timeline.
type TicketEvent struct {
	Action    string
	TicketID  string
	Timestamp string
	Metadata  map[string]any
}

type ticketQueryEnvelope struct {
	Data struct {
		Events []struct {
			Action    string         `json:"action"`
			TargetID  string         `json:"targetId"`
			Timestamp string         `json:"timestamp"`
			Metadata  map[string]any `json:"metadata"`
		} `json:"events"`
		TotalCount int64 `json:"totalCount"`
	} `json:"data"`
}

// CountTicketsSince returns how many ticket.created events exist at/after
// startRFC3339 — used to assign the next daily-sequential queue number.
func (a *AgregaZcy) CountTicketsSince(ctx context.Context, startRFC3339 string) (int, error) {
	var env ticketQueryEnvelope
	path := "/api/v1/internal/events?tenant_id=" + agregaTenant +
		"&action=ticket.created&start=" + url.QueryEscape(startRFC3339) + "&page_size=1"
	if err := a.h.do(ctx, "AgregaZcy", "GET", path, nil, &env); err != nil {
		return 0, err
	}
	return int(env.Data.TotalCount), nil
}

// ListTickets returns ALL ticket lifecycle events (created/ready/done) for the
// production board to fold into a live queue. Paginated — each claim emits 3
// ticket events, so this exceeds AgregaZcy's 200/page cap at ~66 claims.
func (a *AgregaZcy) ListTickets(ctx context.Context) ([]TicketEvent, error) {
	events, err := a.fetchAllEvents(ctx, "target_entity=ticket")
	if err != nil {
		return nil, err
	}
	out := make([]TicketEvent, 0, len(events))
	for _, e := range events {
		out = append(out, TicketEvent{Action: e.Action, TicketID: e.TargetID, Timestamp: e.Timestamp, Metadata: e.Metadata})
	}
	return out, nil
}

// MemberStats derives per-member visit count + total spend + chosen menu from the
// visit.recorded events AgregaZcy already stores — KonsumZcy holds no metrics.
// Returns a map keyed by member phone (the event targetId).
func (a *AgregaZcy) MemberStats(ctx context.Context) (map[string]MemberMetric, error) {
	events, err := a.visitEvents(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]MemberMetric)
	for _, e := range events {
		if e.TargetID == "" {
			continue
		}
		m := out[e.TargetID]
		m.Visits++
		if amt, ok := e.Metadata["amount"].(float64); ok {
			m.Spend += amt
		}
		if menu, ok := e.Metadata["menu"].(string); ok && menu != "" {
			m.Menu = menu // last writer wins → the most recent claim's menu
		}
		out[e.TargetID] = m
	}
	return out, nil
}

// MenuTally counts how many free cups were claimed per menu, derived from the
// same visit.recorded events. Powers the "which drink is more popular" view.
func (a *AgregaZcy) MenuTally(ctx context.Context) (map[string]int, error) {
	events, err := a.visitEvents(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int)
	for _, e := range events {
		if menu, ok := e.Metadata["menu"].(string); ok && menu != "" {
			out[menu]++
		}
	}
	return out, nil
}
