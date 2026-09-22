// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// newTestClient spins up an httptest server with the given handler and returns
// a Client pointed at it. The server is closed when the test finishes.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	return newTestClientWithOpts(t, handler)
}

// newTestClientWithOpts is like newTestClient but lets the test add extra options
// (e.g. WithRetry, WithDebugWriter) on top of the default baseURL+apiKey.
func newTestClientWithOpts(t *testing.T, handler http.HandlerFunc, extra ...Option) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	opts := append([]Option{
		WithBaseURL(srv.URL),
		WithAPIKey("test-token"),
	}, extra...)

	c, err := NewClient(opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// writeEnvelope writes a Poweradmin-style success envelope wrapping data.
func writeEnvelope(t *testing.T, w http.ResponseWriter, status int, data any) {
	t.Helper()
	writeEnvelopeWithPagination(t, w, status, data, nil)
}

// writeEnvelopeWithPagination writes a Poweradmin-style success envelope
// including a top-level pagination object (mirrors the real API).
func writeEnvelopeWithPagination(t *testing.T, w http.ResponseWriter, status int, data any, pagination map[string]int) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	payload := map[string]any{"success": true}
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			t.Fatalf("marshal data: %v", err)
		}
		payload["data"] = json.RawMessage(raw)
	}
	if pagination != nil {
		payload["pagination"] = pagination
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("encode envelope: %v", err)
	}
}

// writeError writes an error envelope in the shape Poweradmin's v2 API
// actually returns: no "error" object, the message sits at envelope level.
func writeError(t *testing.T, w http.ResponseWriter, status int, message string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	payload := map[string]any{
		"success": false,
		"message": message,
		"data":    nil,
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("encode error envelope: %v", err)
	}
}

// decodeBody decodes the JSON request body into v.
func decodeBody(t *testing.T, r *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}

// equalJSONMaps compares a decoded JSON object against the expected keys and
// values. Numbers must be given as float64-compatible values (int is fine).
func equalJSONMaps(got, want map[string]any) bool {
	if len(got) != len(want) {
		return false
	}
	for k, w := range want {
		g, ok := got[k]
		if !ok {
			return false
		}
		switch wv := w.(type) {
		case int:
			if gf, ok := g.(float64); !ok || gf != float64(wv) {
				return false
			}
		default:
			if !reflect.DeepEqual(g, w) {
				return false
			}
		}
	}
	return true
}
