// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"testing"
)

func TestUserGetByID(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/users/5" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"user": map[string]any{
				"user_id":  5,
				"username": "alice",
				"fullname": "Alice A.",
				"email":    "alice@example.com",
				"active":   true,
			},
		})
	})
	user, _, err := client.User.GetByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if user.ID != 5 || user.Username != "alice" {
		t.Errorf("user = %+v", user)
	}
}

func TestUserList(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"users": []map[string]any{
				{"user_id": 1, "username": "alice"},
				{"user_id": 2, "username": "bob"},
			}},
			map[string]int{"current_page": 1, "per_page": 100, "total": 2, "last_page": 1},
		)
	})
	users, _, err := client.User.List(context.Background(), ListOpts{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("len(users) = %d, want 2", len(users))
	}
}

func TestUserAll(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		writeEnvelopeWithPagination(t, w, http.StatusOK,
			map[string]any{"users": []map[string]any{
				{"user_id": page, "username": "user" + strconv.Itoa(page)},
			}},
			map[string]int{"current_page": page, "per_page": 1, "total": 2, "last_page": 2},
		)
	})
	users, err := client.User.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("len(users) = %d, want 2", len(users))
	}
}

func TestUserCreate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/users" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusCreated, map[string]any{"user_id": 42})
	})
	id, _, err := client.User.Create(context.Background(), UserCreateOpts{
		Username: "charlie", Password: "secret", Email: "charlie@example.com", Active: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
}

// PUT /v2/users/{id} only returns {"user_id": ...}; Update reads the user back.
func TestUserUpdate(t *testing.T) {
	var putBody map[string]any
	var calls []string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method)
		if r.URL.Path != "/api/v2/users/5" {
			t.Errorf("path = %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodPut:
			decodeBody(t, r, &putBody)
			writeEnvelope(t, w, http.StatusOK, map[string]any{"user_id": 5})
		case http.MethodGet:
			writeEnvelope(t, w, http.StatusOK, map[string]any{"user": map[string]any{
				"user_id": 5, "username": "alice", "email": "newalice@example.com", "active": true,
			}})
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})
	email := "newalice@example.com"
	user, _, err := client.User.Update(context.Background(), 5, UserUpdateOpts{Email: new(email), Active: new(true)})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(calls) != 2 || calls[0] != http.MethodPut || calls[1] != http.MethodGet {
		t.Errorf("calls = %v, want [PUT GET]", calls)
	}
	if want := map[string]any{"email": email, "active": true}; !equalJSONMaps(putBody, want) {
		t.Errorf("PUT body = %v, want %v", putBody, want)
	}
	if user.ID != 5 || user.Email != email || !user.Active {
		t.Errorf("user = %+v", user)
	}
}

func TestUserDelete(t *testing.T) {
	var body []byte
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v2/users/5" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		body, _ = io.ReadAll(r.Body)
		writeEnvelope(t, w, http.StatusOK, map[string]any{"zones_affected": 0})
	})
	n, _, err := client.User.Delete(context.Background(), 5, UserDeleteOpts{})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("body = %s, want no body without transfer target", body)
	}
	if n != 0 {
		t.Errorf("zones affected = %d, want 0", n)
	}
}

func TestUserDeleteTransfersZones(t *testing.T) {
	var got map[string]any
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		decodeBody(t, r, &got)
		writeEnvelope(t, w, http.StatusOK, map[string]any{"zones_affected": 3})
	})
	n, _, err := client.User.Delete(context.Background(), 5, UserDeleteOpts{TransferToUserID: new(2)})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if want := map[string]any{"transfer_to_user_id": 2}; !equalJSONMaps(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
	if n != 3 {
		t.Errorf("zones affected = %d, want 3", n)
	}
}

func TestUserDeleteOwnsZonesError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusBadRequest, "User owns zones. Please specify transfer_to_user_id to transfer zones to another user.")
	})
	_, _, err := client.User.Delete(context.Background(), 5, UserDeleteOpts{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("err = %v, want 400 APIError", err)
	}
}

func TestUserSetPermissionTemplate(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v2/users/5" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		writeEnvelope(t, w, http.StatusOK, nil)
	})
	_, err := client.User.SetPermissionTemplate(context.Background(), 5, 3)
	if err != nil {
		t.Fatalf("SetPermissionTemplate: %v", err)
	}
}
