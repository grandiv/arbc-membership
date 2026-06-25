// Package clients holds thin HTTP clients for the KreaZcy engines the BFF wires.
// All brand vocabulary stays here / in routes — the engines only ever see their
// own neutral payloads. Upstream errors are sanitized so stack traces never reach
// the browser.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// urlEscape escapes a query-string value.
func urlEscape(s string) string { return url.QueryEscape(s) }

// IsNotFound reports whether err is an UpstreamError with a 404 status. The BFF
// uses it to tell "this customer doesn't exist yet" apart from a real failure,
// so a claim can decide between an insert (new capture) and leaving an existing
// record untouched.
func IsNotFound(err error) bool {
	var ue *UpstreamError
	if errors.As(err, &ue) {
		return ue.Status == http.StatusNotFound
	}
	return false
}

// httpDo is the shared transport for every engine client.
type httpDo struct {
	base string
	hc   *http.Client
}

func newHTTP(base string) httpDo {
	return httpDo{
		base: base,
		hc:   &http.Client{Timeout: 8 * time.Second},
	}
}

// UpstreamError is returned when an engine responds with a non-2xx status. The BFF
// maps it to a clean client-facing error without leaking the upstream body.
type UpstreamError struct {
	Engine string
	Status int
	Body   string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("%s returned %d", e.Engine, e.Status)
}

// do issues a JSON request and decodes a 2xx JSON response into out (may be nil).
func (h httpDo) do(ctx context.Context, engine, method, path string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal %s request: %w", engine, err)
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, h.base+path, rdr)
	if err != nil {
		return fmt.Errorf("build %s request: %w", engine, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := h.hc.Do(req)
	if err != nil {
		// Transport failure (engine down, DNS, timeout) → treat as upstream
		// unavailable so the BFF returns a clean 502, not a 500.
		return &UpstreamError{Engine: engine, Status: 0, Body: err.Error()}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &UpstreamError{Engine: engine, Status: resp.StatusCode, Body: string(raw)}
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode %s response: %w", engine, err)
		}
	}
	return nil
}
