// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"net/http"
	"testing"
)

func TestUserGetByName(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"users": []map[string]any{
				{"user_id": 1, "username": "alice"},
				{"user_id": 2, "username": "bob"},
			}},
			map[string]int{"current_page": 1, "per_page": 100, "total": 2, "last_page": 1},
		)
	})
	u, _, err := client.User.GetByName(context.Background(), "bob")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if u.ID != 2 || u.Username != "bob" {
		t.Errorf("user = %+v", u)
	}
}

func TestUserGetByNameNotFound(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"users": []map[string]any{{"user_id": 1, "username": "alice"}}},
			map[string]int{"current_page": 1, "per_page": 100, "total": 1, "last_page": 1},
		)
	})
	_, _, err := client.User.GetByName(context.Background(), "missing")
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, err = %v", err)
	}
}

func TestGroupGetByName(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"groups": []map[string]any{
				{"id": 1, "name": "ops"},
				{"id": 2, "name": "dev"},
			}},
			map[string]int{"current_page": 1, "per_page": 100, "total": 2, "last_page": 1},
		)
	})
	g, _, err := client.Group.GetByName(context.Background(), "dev")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if g.ID != 2 {
		t.Errorf("group = %+v", g)
	}
}

func TestZoneTemplateGetByName(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(t, w, http.StatusOK, map[string]any{"templates": []map[string]any{
			{"id": 1, "name": "Default", "description": "x"},
			{"id": 2, "name": "Custom", "description": "y"},
		}})
	})
	tpl, _, err := client.ZoneTemplate.GetByName(context.Background(), "Custom")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if tpl.ID != 2 {
		t.Errorf("template = %+v", tpl)
	}
}

func TestPermissionTemplateGetByName(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(t, w, http.StatusOK, map[string]any{"templates": []map[string]any{
			{"id": 1, "name": "Admin", "descr": "x"},
			{"id": 2, "name": "Viewer", "descr": "y"},
		}})
	})
	tpl, _, err := client.PermissionTemplate.GetByName(context.Background(), "Viewer")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if tpl.ID != 2 {
		t.Errorf("template = %+v", tpl)
	}
}

// GetByName sends the exact-match filter; a filtering server answers with a
// single, unpaginated result.
func TestZoneGetByNameUsesServerFilter(t *testing.T) {
	var gotName string
	var calls int
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		gotName = r.URL.Query().Get("name")
		writeEnvelope(t, w, http.StatusOK, map[string]any{"zones": []map[string]any{
			{"id": 9, "name": "target.com", "type": "MASTER", "created_at": "2026-01-01 00:00:00"},
		}})
	})
	z, _, err := client.Zone.GetByName(context.Background(), "target.com")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if gotName != "target.com" || calls != 1 {
		t.Errorf("name filter = %q, calls = %d; want target.com, 1", gotName, calls)
	}
	if z.ID != 9 || z.CreatedAt != "2026-01-01 00:00:00" {
		t.Errorf("zone = %+v", z)
	}
}

func TestUserGetByNameUsesServerFilter(t *testing.T) {
	var gotUsername string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotUsername = r.URL.Query().Get("username")
		writeEnvelope(t, w, http.StatusOK, map[string]any{"users": []map[string]any{
			{"user_id": 4, "username": "carol"},
		}})
	})
	u, _, err := client.User.GetByName(context.Background(), "carol")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if gotUsername != "carol" || u.ID != 4 {
		t.Errorf("username filter = %q, user = %+v", gotUsername, u)
	}
}

func TestUserGetByNameFilteredNotFound(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(t, w, http.StatusOK, map[string]any{"users": []map[string]any{}})
	})
	if _, _, err := client.User.GetByName(context.Background(), "nobody"); !IsNotFound(err) {
		t.Errorf("err = %v, want not found", err)
	}
}
