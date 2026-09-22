// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
)

func TestZoneGetByID(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5" {
			t.Errorf("path = %s, want /api/v2/zones/5", r.URL.Path)
		}
		// Single-zone GET wraps the object under data.zone.
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"zone": map[string]any{
				"id":   5,
				"name": "example.com",
				"type": "MASTER",
			},
		})
	})
	z, resp, err := client.Zone.GetByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Errorf("response status = %v", resp)
	}
	if z.ID != 5 || z.Name != "example.com" || z.Type != ZoneTypeMaster {
		t.Errorf("zone = %+v", z)
	}
}

func TestZoneGetByIDNative(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"zone": map[string]any{
				"id":   6,
				"name": "native.example.com",
				"type": "NATIVE",
			},
		})
	})
	z, _, err := client.Zone.GetByID(context.Background(), 6)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if z.Type != ZoneTypeNative {
		t.Errorf("type = %q, want %q", z.Type, ZoneTypeNative)
	}
}

func TestZoneGetByIDNotFound(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, "no such zone")
	})
	_, _, err := client.Zone.GetByID(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true (err=%v)", err)
	}
}

func TestZoneListSinglePage(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/zones" {
			t.Errorf("path = %s", r.URL.Path)
		}
		// List: data is a flat array, pagination is at envelope level.
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{
				"zones": []map[string]any{
					{"id": 1, "name": "a.com", "type": "MASTER"},
					{"id": 2, "name": "b.com", "type": "SLAVE"},
				},
			},
			map[string]int{"current_page": 1, "per_page": 100, "total": 2, "last_page": 1},
		)
	})
	zones, resp, err := client.Zone.List(context.Background(), ListOpts{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(zones) != 2 {
		t.Fatalf("len(zones) = %d, want 2", len(zones))
	}
	if zones[0].Name != "a.com" || zones[1].Type != ZoneTypeSlave {
		t.Errorf("zones = %+v", zones)
	}
	if resp.Meta.Pagination == nil {
		t.Fatal("Meta.Pagination is nil")
	}
	if resp.Meta.Pagination.Total != 2 || resp.Meta.Pagination.LastPage != 1 {
		t.Errorf("Pagination = %+v", resp.Meta.Pagination)
	}
}

func TestZoneAllIteratesPages(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{
				"zones": []map[string]any{
					{"id": page, "name": "z" + strconv.Itoa(page) + ".com", "type": "MASTER"},
				},
			},
			map[string]int{"current_page": page, "per_page": 1, "total": 2, "last_page": 2},
		)
	})
	zones, err := client.Zone.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(zones) != 2 {
		t.Fatalf("len(zones) = %d, want 2", len(zones))
	}
	if zones[0].ID != 1 || zones[1].ID != 2 {
		t.Errorf("ids = [%d,%d], want [1,2]", zones[0].ID, zones[1].ID)
	}
}

func TestZoneCreate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/v2/zones" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("body: %v", err)
		}
		if got["name"] != "new.com" || got["type"] != "MASTER" {
			t.Errorf("body = %v", got)
		}
		writeEnvelope(t, w, http.StatusCreated, map[string]any{"zone_id": 42})
	})
	id, _, err := client.Zone.Create(context.Background(), ZoneCreateOpts{
		Name: "new.com",
		Type: ZoneTypeMaster,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
}

func TestZoneCreateWireFields(t *testing.T) {
	tests := []struct {
		name string
		opts ZoneCreateOpts
		want map[string]any
	}{
		{
			name: "minimal",
			opts: ZoneCreateOpts{Name: "example.com", Type: ZoneTypeMaster},
			want: map[string]any{"name": "example.com", "type": "MASTER"},
		},
		{
			name: "slave with template, dnssec and explicit owner",
			opts: ZoneCreateOpts{
				Name: "example.org", Type: ZoneTypeSlave, Masters: "192.0.2.1:5300",
				Account: "acme", Description: "prod", TemplateID: 3, EnableDNSSEC: true,
				OwnerUserID: new(7), GroupIDs: []int{2, 5},
			},
			want: map[string]any{
				"name": "example.org", "type": "SLAVE", "master": "192.0.2.1:5300",
				"account": "acme", "description": "prod", "template": 3, "enable_dnssec": true,
				"owner_user_id": 7, "group_ids": []any{float64(2), float64(5)},
			},
		},
		{
			name: "group-only zone sends explicit null owner",
			opts: ZoneCreateOpts{Name: "example.net", Type: ZoneTypeNative, WithoutUserOwner: true, GroupIDs: []int{2}},
			want: map[string]any{
				"name": "example.net", "type": "NATIVE", "owner_user_id": nil, "group_ids": []any{float64(2)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]any
			client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				decodeBody(t, r, &got)
				writeEnvelope(t, w, http.StatusCreated, map[string]any{"zone_id": 42})
			})
			id, _, err := client.Zone.Create(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if id != 42 {
				t.Errorf("id = %d, want 42", id)
			}
			if !equalJSONMaps(got, tt.want) {
				t.Errorf("body = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZoneCreateOwnerValidation(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request expected for invalid options")
	})
	for name, opts := range map[string]ZoneCreateOpts{
		"owner and no owner":      {Name: "a.com", OwnerUserID: new(1), WithoutUserOwner: true, GroupIDs: []int{1}},
		"no owner without groups": {Name: "a.com", WithoutUserOwner: true},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := client.Zone.Create(context.Background(), opts); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestZoneUpdate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if got["description"] != "updated" {
			t.Errorf("description = %v, want 'updated'", got["description"])
		}
		if _, present := got["type"]; present {
			t.Errorf("type should be omitted in update")
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"zone": map[string]any{"id": 5, "name": "x.com", "type": "MASTER", "description": "updated"},
		})
	})
	desc := "updated"
	z, _, err := client.Zone.Update(context.Background(), 5, ZoneUpdateOpts{Description: &desc})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if z.Description != "updated" {
		t.Errorf("description = %q", z.Description)
	}
}

// The update endpoint reads "master" (singular) and "name"; it silently
// ignores "masters" and "account".
func TestZoneUpdateWireFields(t *testing.T) {
	var got map[string]any
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"zone": map[string]any{
				"id": 5, "name": "new.example.com", "type": "SLAVE",
				"masters": "192.0.2.1,192.0.2.2:5300", "account": nil, "description": nil,
			},
		})
	})
	name := "new.example.com"
	typ := ZoneTypeSlave
	masters := "192.0.2.1,192.0.2.2:5300"
	z, _, err := client.Zone.Update(context.Background(), 5, ZoneUpdateOpts{
		Name: &name, Type: &typ, Masters: &masters,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	want := map[string]any{"name": name, "type": "SLAVE", "master": masters}
	if len(got) != len(want) {
		t.Errorf("body = %v, want exactly %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("body[%q] = %v, want %v", k, got[k], v)
		}
	}
	if _, present := got["masters"]; present {
		t.Error(`body must not contain "masters"; the update endpoint ignores it`)
	}
	if z.Name != name || z.Masters != masters || z.Account != "" {
		t.Errorf("zone = %+v", z)
	}
}

func TestZoneDelete(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/7" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if _, err := client.Zone.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestZoneGetByName(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{
				"zones": []map[string]any{
					{"id": 1, "name": "a.com", "type": "MASTER"},
					{"id": 2, "name": "target.com", "type": "MASTER"},
				},
			},
			map[string]int{"current_page": 1, "per_page": 100, "total": 2, "last_page": 1},
		)
	})
	z, _, err := client.Zone.GetByName(context.Background(), "target.com")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if z.ID != 2 {
		t.Errorf("ID = %d, want 2", z.ID)
	}
}

func TestZoneGetByNameNotFound(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"zones": []map[string]any{}},
			map[string]int{"current_page": 1, "per_page": 100, "total": 0, "last_page": 1},
		)
	})
	_, _, err := client.Zone.GetByName(context.Background(), "missing.com")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true (err=%v)", err)
	}
}

func TestZoneOwners(t *testing.T) {
	var sawAdd, sawDelete bool
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/zones/5/owners":
			writeEnvelope(t, w, http.StatusOK, map[string]any{"owners": []map[string]any{
				{"user_id": 1, "username": "alice", "fullname": "Alice A."},
			}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/zones/5/owners":
			sawAdd = true
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["user_id"] == nil {
				t.Errorf("expected user_id in body, got %v", got)
			}
			writeEnvelope(t, w, http.StatusOK, nil)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/zones/5/owners/7":
			sawDelete = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	owners, _, err := client.Zone.Owners(context.Background(), 5)
	if err != nil {
		t.Fatalf("Owners: %v", err)
	}
	if len(owners) != 1 || owners[0].Username != "alice" {
		t.Errorf("owners = %+v", owners)
	}
	if _, err := client.Zone.AddOwner(context.Background(), 5, 7); err != nil {
		t.Fatalf("AddOwner: %v", err)
	}
	if _, err := client.Zone.RemoveOwner(context.Background(), 5, 7); err != nil {
		t.Fatalf("RemoveOwner: %v", err)
	}
	if !sawAdd || !sawDelete {
		t.Errorf("sawAdd=%v sawDelete=%v", sawAdd, sawDelete)
	}
}

func TestZoneAddOwners(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/zones/5/owners" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		ids, ok := got["user_ids"]
		if !ok {
			t.Errorf("expected user_ids in body, got %v", got)
		}
		// JSON numbers decode as float64
		if ids.([]any)[0].(float64) != 3 || ids.([]any)[1].(float64) != 4 {
			t.Errorf("user_ids = %v", ids)
		}
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	if _, err := client.Zone.AddOwners(context.Background(), 5, []int{3, 4}); err != nil {
		t.Fatalf("AddOwners: %v", err)
	}
}

func TestZoneGetDNSSEC(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5/dnssec" {
			t.Errorf("path = %s, want /api/v2/zones/5/dnssec", r.URL.Path)
		}
		dnskey := "257 3 13 abc123"
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"enabled": true,
			"ds_records": []map[string]any{
				{"key_tag": 12345, "algorithm": 13, "digest_type": 2, "digest": "ABC123"},
			},
			"dnskey": dnskey,
		})
	})
	d, resp, err := client.Zone.GetDNSSEC(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetDNSSEC: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Errorf("response status = %v", resp)
	}
	if !d.Enabled {
		t.Errorf("enabled = false, want true")
	}
	if len(d.DSRecords) != 1 {
		t.Fatalf("ds_records len = %d, want 1", len(d.DSRecords))
	}
	if d.DSRecords[0].KeyTag != 12345 {
		t.Errorf("key_tag = %d, want 12345", d.DSRecords[0].KeyTag)
	}
	if d.DNSKey == nil || *d.DNSKey != "257 3 13 abc123" {
		t.Errorf("dnskey = %v", d.DNSKey)
	}
}

func TestZoneSetDNSSEC(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5/dnssec" {
			t.Errorf("path = %s, want /api/v2/zones/5/dnssec", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if req["enabled"] != true {
			t.Errorf("enabled = %v, want true", req["enabled"])
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"enabled":    true,
			"ds_records": []map[string]any{},
			"dnskey":     nil,
		})
	})
	d, resp, err := client.Zone.SetDNSSEC(context.Background(), 5, true)
	if err != nil {
		t.Fatalf("SetDNSSEC: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Errorf("response status = %v", resp)
	}
	if !d.Enabled {
		t.Errorf("enabled = false, want true")
	}
}

func TestZoneGetDNSSECNotFound(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, "zone not found")
	})
	_, _, err := client.Zone.GetDNSSEC(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true (err=%v)", err)
	}
}

func TestZoneSetDNSSECError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusInternalServerError, "failed to update DNSSEC status")
	})
	_, _, err := client.Zone.SetDNSSEC(context.Background(), 5, true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestZoneListMetadata(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5/metadata" {
			t.Errorf("path = %s, want /api/v2/zones/5/metadata", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"metadata": []map[string]any{
				{"kind": "ALLOW-AXFR-FROM", "values": []string{"192.0.2.10", "AUTO-NS"}},
			},
		})
	})
	metadata, resp, err := client.Zone.ListMetadata(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListMetadata: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Errorf("response status = %v", resp)
	}
	if len(metadata) != 1 {
		t.Fatalf("metadata len = %d, want 1", len(metadata))
	}
	if metadata[0].Kind != "ALLOW-AXFR-FROM" {
		t.Errorf("kind = %s, want ALLOW-AXFR-FROM", metadata[0].Kind)
	}
	if len(metadata[0].Values) != 2 || metadata[0].Values[0] != "192.0.2.10" {
		t.Errorf("values = %v", metadata[0].Values)
	}
}

func TestZoneGetMetadata(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5/metadata/ALLOW-AXFR-FROM" {
			t.Errorf("path = %s, want /api/v2/zones/5/metadata/ALLOW-AXFR-FROM", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"kind":   "ALLOW-AXFR-FROM",
			"values": []string{"192.0.2.10", "AUTO-NS"},
		})
	})
	m, resp, err := client.Zone.GetMetadata(context.Background(), 5, "ALLOW-AXFR-FROM")
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Errorf("response status = %v", resp)
	}
	if m.Kind != "ALLOW-AXFR-FROM" {
		t.Errorf("kind = %s, want ALLOW-AXFR-FROM", m.Kind)
	}
	if len(m.Values) != 2 {
		t.Errorf("values len = %d, want 2", len(m.Values))
	}
}

func TestZoneGetMetadataNotFound(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, "zone or metadata kind not found")
	})
	_, _, err := client.Zone.GetMetadata(context.Background(), 99, "ALLOW-AXFR-FROM")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true (err=%v)", err)
	}
}

func TestZoneSetMetadata(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5/metadata/ALLOW-AXFR-FROM" {
			t.Errorf("path = %s, want /api/v2/zones/5/metadata/ALLOW-AXFR-FROM", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		values, _ := req["values"].([]any)
		if len(values) != 2 {
			t.Errorf("values = %v, want 2 entries", req["values"])
		}
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	resp, err := client.Zone.SetMetadata(context.Background(), 5, "ALLOW-AXFR-FROM", []string{"192.0.2.10", "AUTO-NS"})
	if err != nil {
		t.Fatalf("SetMetadata: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Errorf("response status = %v", resp)
	}
}

func TestZoneSetMetadataError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusForbidden, "forbidden or read-only metadata kind")
	})
	_, err := client.Zone.SetMetadata(context.Background(), 5, "ALLOW-AXFR-FROM", []string{"192.0.2.10"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestZoneDeleteMetadata(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5/metadata/ALLOW-AXFR-FROM" {
			t.Errorf("path = %s, want /api/v2/zones/5/metadata/ALLOW-AXFR-FROM", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	resp, err := client.Zone.DeleteMetadata(context.Background(), 5, "ALLOW-AXFR-FROM")
	if err != nil {
		t.Fatalf("DeleteMetadata: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Errorf("response status = %v", resp)
	}
}

func TestZoneDeleteMetadataError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusForbidden, "forbidden or read-only metadata kind")
	})
	_, err := client.Zone.DeleteMetadata(context.Background(), 5, "ALLOW-AXFR-FROM")
	if err == nil {
		t.Fatal("expected error")
	}
}
