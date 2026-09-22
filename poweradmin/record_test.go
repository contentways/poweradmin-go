// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"net/http"
	"strconv"
	"testing"
)

func TestRecordListUsesZoneScope(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/zones/5/records" {
			t.Errorf("path = %s, want /api/v2/zones/5/records", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{"records": []map[string]any{
			{"id": "rec-1", "name": "host", "type": "A", "content": "1.2.3.4", "ttl": 300},
		}})
	})
	records, _, err := client.Record.List(context.Background(), 5, RecordListOpts{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 1 || records[0].Type != "A" {
		t.Errorf("records = %+v", records)
	}
}

func TestRecordListTypeFilter(t *testing.T) {
	var gotType string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotType = r.URL.Query().Get("type")
		writeEnvelope(t, w, http.StatusOK, map[string]any{"records": []map[string]any{}})
	})
	if _, _, err := client.Record.List(context.Background(), 5, RecordListOpts{Type: "MX"}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotType != "MX" {
		t.Errorf("type filter = %q, want MX", gotType)
	}
}

func TestRecordAllIteratesPages(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"records": []map[string]any{
				{"id": "rec-" + strconv.Itoa(page), "name": "r" + strconv.Itoa(page), "type": "A", "ttl": 300},
			}},
			map[string]int{"current_page": page, "per_page": 1, "total": 3, "last_page": 3},
		)
	})
	records, err := client.Record.All(context.Background(), 5)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(records))
	}
}

func TestRecordCreate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/zones/5/records" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		// Create response wraps the new record under data.record.
		writeEnvelope(t, w, http.StatusCreated, map[string]any{
			"record": map[string]any{
				"id":      "rec-99",
				"zone_id": 5,
				"name":    "www.example.com",
				"type":    "A",
			},
		})
	})
	id, _, err := client.Record.Create(context.Background(), 5, RecordCreateOpts{
		Name: "www.example.com", Type: "A", Content: "1.2.3.4", TTL: 300,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "rec-99" {
		t.Errorf("id = %s, want rec-99", id)
	}
}

func TestRecordBulk(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/v2/zones/5/records/bulk" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"created": 1,
			"updated": 0,
			"deleted": 1,
			"failed":  0,
		})
	})
	result, _, err := client.Record.Bulk(context.Background(), 5, []BulkRecordOperation{
		{Action: "create", Name: "a", Type: "A", Content: "1.1.1.1", TTL: 300},
		{Action: "delete", RecordID: "rec-99"},
	})
	if err != nil {
		t.Fatalf("Bulk: %v", err)
	}
	if result.Created != 1 || result.Deleted != 1 {
		t.Errorf("result = %+v, want created=1 deleted=1", result)
	}
}

func TestRecordGetByID(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/zones/5/records/rec-99" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"record": map[string]any{
				"id":      "rec-99",
				"zone_id": 5,
				"name":    "www.example.com",
				"type":    "A",
				"content": "1.2.3.4",
				"ttl":     300,
			},
		})
	})
	rec, _, err := client.Record.GetByID(context.Background(), 5, "rec-99")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if rec.ID != "rec-99" || rec.Type != "A" || rec.Content != "1.2.3.4" {
		t.Errorf("rec = %+v", rec)
	}
}

func TestRecordUpdate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v2/zones/5/records/rec-99" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"record": map[string]any{
				"id":      "rec-99",
				"zone_id": 5,
				"name":    "www.example.com",
				"type":    "A",
				"content": "5.6.7.8",
				"ttl":     600,
			},
		})
	})
	rec, _, err := client.Record.Update(context.Background(), 5, "rec-99", RecordUpdateOpts{
		Name:    Ptr("www.example.com"),
		Type:    Ptr("A"),
		Content: Ptr("5.6.7.8"),
		TTL:     Ptr(600),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if rec.Content != "5.6.7.8" || rec.TTL != 600 {
		t.Errorf("rec = %+v", rec)
	}
}

func TestRecordDelete(t *testing.T) {
	deleted := false
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v2/zones/5/records/rec-99" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		deleted = true
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	_, err := client.Record.Delete(context.Background(), 5, "rec-99")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("DELETE was not called")
	}
}

// Poweradmin emits purely numeric record IDs as JSON numbers.
func TestRecordNumericIDs(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeEnvelope(t, w, http.StatusOK, map[string]any{"records": []map[string]any{
				{"id": 1234, "name": "www.example.com", "type": "A", "content": "192.0.2.1", "ttl": 300, "priority": nil, "disabled": false},
			}})
		case http.MethodPost:
			writeEnvelope(t, w, http.StatusCreated, map[string]any{"record": map[string]any{
				"id": 1235, "zone_id": 5, "name": "api.example.com", "type": "A", "content": "192.0.2.2", "ttl": 300,
			}})
		}
	})

	records, _, err := client.Record.List(context.Background(), 5, RecordListOpts{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 1 || records[0].ID != "1234" {
		t.Errorf("records = %+v, want ID 1234", records)
	}

	id, _, err := client.Record.Create(context.Background(), 5, RecordCreateOpts{
		Name: "api.example.com", Type: "A", Content: "192.0.2.2", TTL: 300,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "1235" {
		t.Errorf("Create ID = %q, want 1235", id)
	}
}

// Only the fields set in RecordUpdateOpts are sent; an explicit zero value
// (e.g. disabled=false) must still reach the server.
func TestRecordUpdateSendsOnlySetFields(t *testing.T) {
	var got map[string]any
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		decodeBody(t, r, &got)
		writeEnvelope(t, w, http.StatusOK, map[string]any{"record": map[string]any{"id": 99, "content": "192.0.2.10"}})
	})
	if _, _, err := client.Record.Update(context.Background(), 5, "99", RecordUpdateOpts{
		Content:  Ptr("192.0.2.10"),
		Disabled: Ptr(false),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if want := map[string]any{"content": "192.0.2.10", "disabled": false}; !equalJSONMaps(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
}
