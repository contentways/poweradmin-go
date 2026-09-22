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

func TestZoneTemplateList(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/zone-templates" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{"templates": []map[string]any{
			{"id": 1, "name": "Default", "description": "Default tpl", "owner": 1, "is_global": false, "zones_linked": 3},
		}})
	})
	templates, _, err := client.ZoneTemplate.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(templates) != 1 || templates[0].Name != "Default" || templates[0].ZonesLinked != 3 {
		t.Errorf("templates = %+v", templates)
	}
}

func TestZoneTemplateGetByID(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/zone-templates/7" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"template": map[string]any{"id": 7, "name": "X", "description": "d"},
		})
	})
	tpl, _, err := client.ZoneTemplate.GetByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if tpl.ID != 7 || tpl.Name != "X" {
		t.Errorf("tpl = %+v", tpl)
	}
}

func TestZoneTemplateCreate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/zone-templates" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if got["name"] != "Web" {
			t.Errorf("body = %v", got)
		}
		// Create reports only the new ID.
		writeEnvelope(t, w, http.StatusCreated, map[string]any{"id": 42})
	})
	tpl, _, err := client.ZoneTemplate.Create(context.Background(), ZoneTemplateCreateOpts{
		Name: "Web", Description: "web servers", IsGlobal: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tpl.ID != 42 || tpl.Name != "Web" || tpl.Description != "web servers" || !tpl.IsGlobal {
		t.Errorf("tpl = %+v", tpl)
	}
}

func TestZoneTemplateCreateRecord(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/zone-templates/1/records" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusCreated, map[string]any{"id": 7})
	})
	rec, _, err := client.ZoneTemplate.CreateRecord(context.Background(), 1, ZoneTemplateRecordOpts{
		Name: "[ZONE]", Type: "A", Content: "[NS1]", TTL: 3600,
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	if rec.ID != 7 || rec.Type != "A" || rec.Content != "[NS1]" || rec.TTL != 3600 {
		t.Errorf("rec = %+v", rec)
	}
}

func TestZoneTemplateRecords(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/zone-templates/1/records" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{"records": []map[string]any{
			{"id": 1, "name": "[ZONE]", "type": "SOA", "content": "[NS1] [HOSTMASTER] [SERIAL] 28800 7200 604800 86400", "ttl": 86400},
		}})
	})
	records, _, err := client.ZoneTemplate.Records(context.Background(), 1)
	if err != nil {
		t.Fatalf("Records: %v", err)
	}
	if len(records) != 1 || records[0].Type != "SOA" {
		t.Errorf("records = %+v", records)
	}
}

// PUT /v2/zone-templates/{id} returns data: null; Update reads the template back.
func TestZoneTemplateUpdate(t *testing.T) {
	var calls []string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method)
		if r.URL.Path != "/api/v2/zone-templates/7" {
			t.Errorf("path = %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodPut:
			writeEnvelope(t, w, http.StatusOK, nil)
		case http.MethodGet:
			writeEnvelope(t, w, http.StatusOK, map[string]any{
				"template": map[string]any{
					"id": 7, "name": "Updated", "description": "new desc",
					"owner": 1, "is_global": true, "zones_linked": 0,
				},
			})
		}
	})
	tpl, _, err := client.ZoneTemplate.Update(context.Background(), 7, ZoneTemplateUpdateOpts{
		Name: "Updated", Description: "new desc", IsGlobal: true,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(calls) != 2 || calls[1] != http.MethodGet {
		t.Errorf("calls = %v, want [PUT GET]", calls)
	}
	if tpl.ID != 7 || tpl.Name != "Updated" || !tpl.IsGlobal {
		t.Errorf("tpl = %+v", tpl)
	}
}

func TestZoneTemplateDelete(t *testing.T) {
	deleted := false
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v2/zone-templates/7" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		deleted = true
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	_, err := client.ZoneTemplate.Delete(context.Background(), 7)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("DELETE was not called")
	}
}

func TestZoneTemplateGetRecord(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/zone-templates/1/records/5" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"record": map[string]any{
				"id": 5, "name": "[ZONE]", "type": "A",
				"content": "1.2.3.4", "ttl": 3600, "priority": 0,
			},
		})
	})
	rec, _, err := client.ZoneTemplate.GetRecord(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if rec.ID != 5 || rec.Type != "A" || rec.Content != "1.2.3.4" {
		t.Errorf("rec = %+v", rec)
	}
}

// PUT /v2/zone-templates/{id}/records/{id} returns data: null; UpdateRecord
// reads the record back.
func TestZoneTemplateUpdateRecord(t *testing.T) {
	var calls []string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method)
		if r.URL.Path != "/api/v2/zone-templates/1/records/5" {
			t.Errorf("path = %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodPut:
			writeEnvelope(t, w, http.StatusOK, nil)
		case http.MethodGet:
			writeEnvelope(t, w, http.StatusOK, map[string]any{
				"record": map[string]any{
					"id": 5, "name": "[ZONE]", "type": "A",
					"content": "9.9.9.9", "ttl": 7200, "priority": 0,
				},
			})
		}
	})
	rec, _, err := client.ZoneTemplate.UpdateRecord(context.Background(), 1, 5, ZoneTemplateRecordOpts{
		Name: "[ZONE]", Type: "A", Content: "9.9.9.9", TTL: 7200,
	})
	if err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}
	if len(calls) != 2 || calls[1] != http.MethodGet {
		t.Errorf("calls = %v, want [PUT GET]", calls)
	}
	if rec.Content != "9.9.9.9" || rec.TTL != 7200 {
		t.Errorf("rec = %+v", rec)
	}
}

func TestZoneTemplateDeleteRecord(t *testing.T) {
	deleted := false
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v2/zone-templates/1/records/5" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		deleted = true
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	_, err := client.ZoneTemplate.DeleteRecord(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}
	if !deleted {
		t.Error("DELETE was not called")
	}
}
