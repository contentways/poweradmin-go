// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func dnssecKeyJSON(id int, active bool) map[string]any {
	return map[string]any{
		"id":           id,
		"type":         "csk",
		"keytag":       46395,
		"algorithm":    "ecdsa256",
		"algorithm_id": 13,
		"bits":         256,
		"active":       active,
	}
}

func TestDNSSECListKeys(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/zones/7/dnssec/keys" {
			t.Errorf("request = %s %s, want GET /api/v2/zones/7/dnssec/keys", r.Method, r.URL.Path)
		}
		unknownAlgorithm := dnssecKeyJSON(2, false)
		unknownAlgorithm["algorithm"] = nil
		unknownAlgorithm["algorithm_id"] = 253
		writeEnvelope(t, w, http.StatusOK, []map[string]any{dnssecKeyJSON(1, true), unknownAlgorithm})
	})

	keys, _, err := client.DNSSEC.ListKeys(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("len(keys) = %d, want 2", len(keys))
	}
	want := DNSSECKey{ID: 1, Type: DNSSECKeyTypeCSK, KeyTag: 46395, Algorithm: "ecdsa256", AlgorithmID: 13, Bits: 256, Active: true}
	if *keys[0] != want {
		t.Errorf("keys[0] = %+v, want %+v", *keys[0], want)
	}
	if keys[1].Algorithm != "" || keys[1].AlgorithmID != 253 {
		t.Errorf("unknown algorithm = %q/%d, want \"\"/253", keys[1].Algorithm, keys[1].AlgorithmID)
	}
}

func TestDNSSECAddKey(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/zones/7/dnssec/keys" {
			t.Errorf("request = %s %s, want POST /api/v2/zones/7/dnssec/keys", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if req["type"] != "csk" || req["algorithm"] != "ecdsa256" || req["bits"] != float64(256) {
			t.Errorf("body = %s", body)
		}
		writeEnvelope(t, w, http.StatusCreated, dnssecKeyJSON(3, false))
	})

	key, resp, err := client.DNSSEC.AddKey(context.Background(), 7, DNSSECKeyCreateOpts{
		Type:      DNSSECKeyTypeCSK,
		Algorithm: "ecdsa256",
		Bits:      256,
	})
	if err != nil {
		t.Fatalf("AddKey: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
	if key.ID != 3 || key.Active {
		t.Errorf("key = %+v, want ID 3, inactive", *key)
	}
}

func TestDNSSECAddKeyValidationError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusBadRequest, "ECDSA P-256 algorithm must use 256 bits")
	})

	_, _, err := client.DNSSEC.AddKey(context.Background(), 7, DNSSECKeyCreateOpts{
		Type: DNSSECKeyTypeCSK, Algorithm: "ecdsa256", Bits: 384,
	})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != http.StatusBadRequest || apiErr.Message != "ECDSA P-256 algorithm must use 256 bits" {
		t.Fatalf("err = %v, want 400 APIError with server message", err)
	}
}

func TestDNSSECSetKeyActive(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v2/zones/7/dnssec/keys/3" {
			t.Errorf("request = %s %s, want PATCH /api/v2/zones/7/dnssec/keys/3", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"active":true}` {
			t.Errorf("body = %s, want {\"active\":true}", body)
		}
		writeEnvelope(t, w, http.StatusOK, dnssecKeyJSON(3, true))
	})

	key, _, err := client.DNSSEC.SetKeyActive(context.Background(), 7, 3, true)
	if err != nil {
		t.Fatalf("SetKeyActive: %v", err)
	}
	if !key.Active {
		t.Errorf("active = false, want true")
	}
}

func TestDNSSECDeleteKey(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v2/zones/7/dnssec/keys/3" {
			t.Errorf("request = %s %s, want DELETE /api/v2/zones/7/dnssec/keys/3", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, nil)
	})

	if _, err := client.DNSSEC.DeleteKey(context.Background(), 7, 3); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}
}

func TestDNSSECGetKeyNotFound(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, "DNSSEC key not found")
	})

	_, _, err := client.DNSSEC.GetKey(context.Background(), 7, 999)
	if !IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestDNSSECRectify(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/zones/7/dnssec/rectify" {
			t.Errorf("request = %s %s, want POST /api/v2/zones/7/dnssec/rectify", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, nil)
	})

	if _, err := client.DNSSEC.Rectify(context.Background(), 7); err != nil {
		t.Fatalf("Rectify: %v", err)
	}
}

func TestZoneGetDNSSECPresigned(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"enabled":    true,
			"presigned":  true,
			"ds_records": []map[string]any{},
			"dnskey":     nil,
		})
	})

	d, _, err := client.Zone.GetDNSSEC(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetDNSSEC: %v", err)
	}
	if !d.Presigned {
		t.Errorf("presigned = false, want true")
	}
}
