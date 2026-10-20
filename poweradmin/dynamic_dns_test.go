// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestDynamicDNSUpdate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/dynamic-dns" {
			t.Errorf("request = %s %s, want POST /api/v2/dynamic-dns", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		want := map[string]any{
			"hostname":  "host.example.com",
			"ipv4":      "192.0.2.1,192.0.2.2",
			"dualstack": true,
		}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %v, want %v", body, want)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"hostname":     "host.example.com",
			"zone_id":      3,
			"applied_ipv4": []string{"192.0.2.1", "192.0.2.2"},
			"applied_ipv6": []string{},
			"changed":      true,
		})
	})

	res, _, err := client.DynamicDNS.Update(context.Background(), DynamicDNSUpdateOpts{
		Hostname:  "host.example.com",
		IPv4:      []string{"192.0.2.1", "192.0.2.2"},
		Dualstack: true,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if res.ZoneID != 3 || !res.Changed || len(res.AppliedIPv4) != 2 || len(res.AppliedIPv6) != 0 {
		t.Errorf("result = %+v", *res)
	}
}

func TestDynamicDNSUpdateConflict(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusConflict, "Update did not produce any change and no matching records exist")
	})

	_, _, err := client.DynamicDNS.Update(context.Background(), DynamicDNSUpdateOpts{
		Hostname: "host.example.com",
		IPv6:     []string{"2001:db8::1"},
	})
	if !IsConflict(err) {
		t.Fatalf("err = %v, want a 409 conflict", err)
	}
}
