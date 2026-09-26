// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"net/http"
	"testing"
)

func TestServerStatus(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/server/status" {
			t.Errorf("request = %s %s, want GET /api/v2/server/status", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("metrics"); got != "uptime,udp-queries" {
			t.Errorf("metrics = %q, want uptime,udp-queries", got)
		}
		if got := r.URL.Query().Get("include"); got != "slaves" {
			t.Errorf("include = %q, want slaves", got)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"running":        true,
			"server_id":      "localhost",
			"daemon_type":    "authoritative",
			"version":        "4.9.17",
			"uptime_seconds": 349,
			"metrics":        map[string]string{"uptime": "349", "udp-queries": "12"},
			"slaves": []map[string]any{
				{"ip": "192.0.2.1", "status": "unreachable", "last_checked": "2026-09-25 22:33:58", "error": "Network unreachable"},
			},
		})
	})

	status, _, err := client.Server.Status(context.Background(), ServerStatusOpts{
		Metrics:       []string{"uptime", "udp-queries"},
		IncludeSlaves: true,
	})
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Running || status.Version != "4.9.17" || status.DaemonType != "authoritative" {
		t.Errorf("status = %+v", *status)
	}
	if status.UptimeSeconds == nil || *status.UptimeSeconds != 349 {
		t.Errorf("uptime = %v, want 349", status.UptimeSeconds)
	}
	if status.Metrics["udp-queries"] != "12" {
		t.Errorf("metrics = %v", status.Metrics)
	}
	if len(status.Slaves) != 1 || status.Slaves[0].Status != "unreachable" {
		t.Errorf("slaves = %+v", status.Slaves)
	}
}

func TestServerStatusNoOptsSendsNoQuery(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want empty", r.URL.RawQuery)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{"running": true})
	})

	if _, _, err := client.Server.Status(context.Background(), ServerStatusOpts{}); err != nil {
		t.Fatalf("Status: %v", err)
	}
}

func TestServerStatusUnavailable(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusServiceUnavailable, "PowerDNS server is not reachable")
	})

	_, _, err := client.Server.Status(context.Background(), ServerStatusOpts{})
	if !IsServiceUnavailable(err) {
		t.Fatalf("err = %v, want service unavailable", err)
	}
}
