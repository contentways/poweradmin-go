// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
		"dnskey":       "257 3 13 mdsswUyr3DPW132mOi8V9xESWE8jTo0dxCjjnopKl+GqJxpVXckHAeF+KkxLbxILfDLUT0rAK9iUzy1L53eKGQ==",
		"ds":           []string{"46395 13 2 3dd8ee7d9ab0c6d8e4b2fd8a7e1cb3a2b7b0d4e5f6a7b8c9d0e1f2a3b4c5d6e7"},
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
		unknownAlgorithm["dnskey"] = nil
		unknownAlgorithm["ds"] = []string{}
		writeEnvelope(t, w, http.StatusOK, []map[string]any{dnssecKeyJSON(1, true), unknownAlgorithm})
	})

	keys, _, err := client.DNSSEC.ListKeys(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("len(keys) = %d, want 2", len(keys))
	}
	k := keys[0]
	if k.ID != 1 || k.Type != DNSSECKeyTypeCSK || k.KeyTag != 46395 || k.Algorithm != "ecdsa256" ||
		k.AlgorithmID != 13 || k.Bits != 256 || !k.Active {
		t.Errorf("keys[0] = %+v", *k)
	}
	if k.DNSKey == nil || !strings.HasPrefix(*k.DNSKey, "257 3 13 ") {
		t.Errorf("keys[0].DNSKey = %v, want 257 3 13 …", k.DNSKey)
	}
	if len(k.DS) != 1 || !strings.HasPrefix(k.DS[0], "46395 13 2 ") {
		t.Errorf("keys[0].DS = %v", k.DS)
	}
	if keys[1].Algorithm != "" || keys[1].AlgorithmID != 253 {
		t.Errorf("unknown algorithm = %q/%d, want \"\"/253", keys[1].Algorithm, keys[1].AlgorithmID)
	}
	if keys[1].DNSKey != nil || keys[1].DS == nil || len(keys[1].DS) != 0 {
		t.Errorf("keys[1] DNSKey/DS = %v/%v, want nil/[]", keys[1].DNSKey, keys[1].DS)
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
		if req["type"] != "csk" || req["algorithm"] != "ecdsa256" || req["bits"] != float64(256) || req["active"] != true {
			t.Errorf("body = %s", body)
		}
		writeEnvelope(t, w, http.StatusCreated, dnssecKeyJSON(3, true))
	})

	key, resp, err := client.DNSSEC.AddKey(context.Background(), 7, DNSSECKeyCreateOpts{
		Type:      DNSSECKeyTypeCSK,
		Algorithm: "ecdsa256",
		Bits:      256,
		Active:    true,
	})
	if err != nil {
		t.Fatalf("AddKey: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
	if key.ID != 3 || !key.Active || len(key.DS) != 1 {
		t.Errorf("key = %+v, want ID 3, active, with DS", *key)
	}
}

func TestDNSSECAddKeyOmitsActiveByDefault(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "active") {
			t.Errorf("body = %s, want no active field", body)
		}
		writeEnvelope(t, w, http.StatusCreated, dnssecKeyJSON(4, false))
	})

	if _, _, err := client.DNSSEC.AddKey(context.Background(), 7, DNSSECKeyCreateOpts{
		Type: DNSSECKeyTypeZSK, Algorithm: "ecdsa256", Bits: 256,
	}); err != nil {
		t.Fatalf("AddKey: %v", err)
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

func TestDNSSECListKeysEmpty(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(t, w, http.StatusOK, []map[string]any{})
	})

	keys, _, err := client.DNSSEC.ListKeys(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if keys == nil || len(keys) != 0 {
		t.Errorf("keys = %v, want empty slice", keys)
	}
}

func TestDNSSECGetKey(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/zones/7/dnssec/keys/3" {
			t.Errorf("request = %s %s, want GET /api/v2/zones/7/dnssec/keys/3", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, dnssecKeyJSON(3, true))
	})

	key, _, err := client.DNSSEC.GetKey(context.Background(), 7, 3)
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if key.ID != 3 || key.Type != DNSSECKeyTypeCSK || key.KeyTag != 46395 || !key.Active {
		t.Errorf("key = %+v", *key)
	}
	if key.DNSKey == nil || len(key.DS) != 1 {
		t.Errorf("DNSKey/DS = %v/%v, want both set", key.DNSKey, key.DS)
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
