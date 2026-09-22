// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr string
	}{
		{
			name:    "missing base URL",
			opts:    []Option{WithAPIKey("k")},
			wantErr: "base URL is required",
		},
		{
			name:    "missing auth",
			opts:    []Option{WithBaseURL("https://x")},
			wantErr: "authentication is required",
		},
		{
			name: "ok with api key",
			opts: []Option{WithBaseURL("https://x"), WithAPIKey("k")},
		},
		{
			name: "ok with basic auth",
			opts: []Option{WithBaseURL("https://x"), WithBasicAuth("u", "p")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.opts...)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestBuildURL(t *testing.T) {
	c, err := NewClient(WithBaseURL("https://dns.example.com/"), WithAPIKey("k"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if g := c.buildURL("/zones"); g != "https://dns.example.com/api/v2/zones" {
		t.Fatalf("buildURL = %q", g)
	}

	c2, _ := NewClient(WithBaseURL("https://dns.example.com"), WithAPIKey("k"), WithAPIVersion("v3"))
	if g := c2.buildURL("groups/5"); g != "https://dns.example.com/api/v3/groups/5" {
		t.Fatalf("custom version: %q", g)
	}
}

func TestBearerHeader(t *testing.T) {
	var gotAuth, gotAPIKey, gotUA string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAPIKey = r.Header.Get("X-API-Key")
		gotUA = r.Header.Get("User-Agent")
		writeEnvelope(t, w, http.StatusOK, map[string]any{"zones": []any{}})
	})
	if _, _, err := client.Zone.List(context.Background(), ListOpts{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want Bearer test-token", gotAuth)
	}
	if gotAPIKey != "" {
		t.Errorf("X-API-Key should be empty, got %q", gotAPIKey)
	}
	if !strings.HasPrefix(gotUA, "poweradmin-go/") {
		t.Errorf("User-Agent = %q, want prefix poweradmin-go/", gotUA)
	}
}

func TestBasicAuthHeader(t *testing.T) {
	var gotUser, gotPass string
	var present bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, present = r.BasicAuth()
		writeEnvelope(t, w, http.StatusOK, map[string]any{"zones": []any{}})
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(WithBaseURL(srv.URL), WithBasicAuth("alice", "s3cret"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, _, err := c.Zone.List(context.Background(), ListOpts{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if !present || gotUser != "alice" || gotPass != "s3cret" {
		t.Errorf("BasicAuth = (%q,%q,%v); want (alice,s3cret,true)", gotUser, gotPass, present)
	}
}

func TestParseAPIError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusBadRequest, "invalid input")
	})
	_, _, err := client.Zone.GetByID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Message, "invalid input") {
		t.Errorf("Message = %q, want to contain 'invalid input'", apiErr.Message)
	}
}

func TestParseNoContent(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	if _, err := client.Zone.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestParseSuccessFalseEnvelope(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// 2xx but success:false in the envelope — should still be an error.
		_, _ = w.Write([]byte(`{"success":false,"error":{"message":"nope"}}`))
	})
	_, _, err := client.Zone.GetByID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error = %v, want 'nope'", err)
	}
}

func TestErrorMessageIsNotRawBody(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusConflict, "Zone already exists")
	})
	_, _, err := client.Zone.Create(context.Background(), ZoneCreateOpts{Name: "example.com", Type: ZoneTypeMaster})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Message != "Zone already exists" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Zone already exists")
	}
	if want := "poweradmin: HTTP 409: Zone already exists"; err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}
