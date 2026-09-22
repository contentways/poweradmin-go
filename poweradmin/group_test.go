// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"net/http"
	"strconv"
	"testing"
)

func TestGroupMembers(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/groups/3/members" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{"members": []map[string]any{
			{"user_id": 1, "username": "alice", "fullname": "Alice A."},
			{"user_id": 2, "username": "bob"},
		}})
	})
	members, _, err := client.Group.Members(context.Background(), 3)
	if err != nil {
		t.Fatalf("Members: %v", err)
	}
	if len(members) != 2 || members[0].Username != "alice" {
		t.Errorf("members = %+v", members)
	}
}

func TestGroupAddRemoveMember(t *testing.T) {
	var sawAdd, sawRemove bool
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/groups/3/members":
			sawAdd = true
			writeEnvelope(t, w, http.StatusOK, nil)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/groups/3/members/9":
			sawRemove = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	if _, err := client.Group.AddMember(context.Background(), 3, 9); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if _, err := client.Group.RemoveMember(context.Background(), 3, 9); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	if !sawAdd || !sawRemove {
		t.Errorf("sawAdd=%v sawRemove=%v", sawAdd, sawRemove)
	}
}

func TestGroupGetByID(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/groups/3" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"group": map[string]any{
				"id":   3,
				"name": "ops",
			},
		})
	})
	group, _, err := client.Group.GetByID(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if group.ID != 3 || group.Name != "ops" {
		t.Errorf("group = %+v", group)
	}
}

func TestGroupAll(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"groups": []map[string]any{
				{"id": page, "name": "group" + strconv.Itoa(page)},
			}},
			map[string]int{"current_page": page, "per_page": 1, "total": 2, "last_page": 2},
		)
	})
	groups, err := client.Group.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("len(groups) = %d, want 2", len(groups))
	}
}

func TestGroupCreate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/groups" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusCreated, map[string]any{
			"group": map[string]any{"id": 7, "name": "devs"},
		})
	})
	id, _, err := client.Group.Create(context.Background(), GroupCreateOpts{
		Name: "devs", Description: "Developers",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
}

func TestGroupUpdate(t *testing.T) {
	var got map[string]any
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v2/groups/3" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		decodeBody(t, r, &got)
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"group": map[string]any{"id": 3, "name": "ops-updated", "description": "", "perm_templ_id": 4},
		})
	})
	group, _, err := client.Group.Update(context.Background(), 3, GroupUpdateOpts{
		Name: new("ops-updated"), Description: new(""), PermTemplID: new(4),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	// An empty description is an explicit "clear", so it must be sent.
	if want := map[string]any{"name": "ops-updated", "description": "", "perm_templ_id": 4}; !equalJSONMaps(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
	if group.Name != "ops-updated" || group.PermTemplID != 4 {
		t.Errorf("group = %+v", group)
	}
}

func TestGroupDelete(t *testing.T) {
	deleted := false
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v2/groups/3" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		deleted = true
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	_, err := client.Group.Delete(context.Background(), 3)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("DELETE was not called")
	}
}

func TestGroupZones(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/groups/3/zones" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{"zones": []map[string]any{
			{"zone_id": 1, "zone_name": "example.com", "zone_type": "NATIVE"},
		}})
	})
	zones, _, err := client.Group.Zones(context.Background(), 3)
	if err != nil {
		t.Fatalf("Zones: %v", err)
	}
	if len(zones) != 1 || zones[0].ZoneName != "example.com" {
		t.Errorf("zones = %+v", zones)
	}
}

func TestGroupAddRemoveZone(t *testing.T) {
	var sawAdd, sawRemove bool
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/groups/3/zones":
			sawAdd = true
			writeEnvelope(t, w, http.StatusOK, nil)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/groups/3/zones/1":
			sawRemove = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	if _, err := client.Group.AddZone(context.Background(), 3, 1); err != nil {
		t.Fatalf("AddZone: %v", err)
	}
	if _, err := client.Group.RemoveZone(context.Background(), 3, 1); err != nil {
		t.Fatalf("RemoveZone: %v", err)
	}
	if !sawAdd || !sawRemove {
		t.Errorf("sawAdd=%v sawRemove=%v", sawAdd, sawRemove)
	}
}
