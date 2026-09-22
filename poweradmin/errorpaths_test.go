// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// allOperations calls every service method once. Each entry returns only the
// error so the table can cover the whole API surface uniformly.
func allOperations(c *Client) map[string]func(context.Context) error {
	e := func(_ any, _ *Response, err error) error { return err }
	r := func(_ *Response, err error) error { return err }
	s := func(_ any, err error) error { return err }
	return map[string]func(context.Context) error{
		"Zone.GetByID":        func(ctx context.Context) error { return e(c.Zone.GetByID(ctx, 1)) },
		"Zone.GetByName":      func(ctx context.Context) error { return e(c.Zone.GetByName(ctx, "a.com")) },
		"Zone.List":           func(ctx context.Context) error { return e(c.Zone.List(ctx, ListOpts{})) },
		"Zone.All":            func(ctx context.Context) error { return s(c.Zone.All(ctx)) },
		"Zone.Create":         func(ctx context.Context) error { return e(c.Zone.Create(ctx, ZoneCreateOpts{Name: "a.com"})) },
		"Zone.Update":         func(ctx context.Context) error { return e(c.Zone.Update(ctx, 1, ZoneUpdateOpts{})) },
		"Zone.Delete":         func(ctx context.Context) error { return r(c.Zone.Delete(ctx, 1)) },
		"Zone.Owners":         func(ctx context.Context) error { return e(c.Zone.Owners(ctx, 1)) },
		"Zone.AddOwner":       func(ctx context.Context) error { return r(c.Zone.AddOwner(ctx, 1, 2)) },
		"Zone.AddOwners":      func(ctx context.Context) error { return r(c.Zone.AddOwners(ctx, 1, []int{2})) },
		"Zone.RemoveOwner":    func(ctx context.Context) error { return r(c.Zone.RemoveOwner(ctx, 1, 2)) },
		"Zone.GetDNSSEC":      func(ctx context.Context) error { return e(c.Zone.GetDNSSEC(ctx, 1)) },
		"Zone.SetDNSSEC":      func(ctx context.Context) error { return e(c.Zone.SetDNSSEC(ctx, 1, true)) },
		"Zone.ListMetadata":   func(ctx context.Context) error { return e(c.Zone.ListMetadata(ctx, 1)) },
		"Zone.GetMetadata":    func(ctx context.Context) error { return e(c.Zone.GetMetadata(ctx, 1, "ALLOW-AXFR-FROM")) },
		"Zone.SetMetadata":    func(ctx context.Context) error { return r(c.Zone.SetMetadata(ctx, 1, "ALLOW-AXFR-FROM", nil)) },
		"Zone.DeleteMetadata": func(ctx context.Context) error { return r(c.Zone.DeleteMetadata(ctx, 1, "ALLOW-AXFR-FROM")) },

		"Record.GetByID": func(ctx context.Context) error { return e(c.Record.GetByID(ctx, 1, "2")) },
		"Record.List":    func(ctx context.Context) error { return e(c.Record.List(ctx, 1, RecordListOpts{})) },
		"Record.All":     func(ctx context.Context) error { return s(c.Record.All(ctx, 1)) },
		"Record.Create":  func(ctx context.Context) error { return e(c.Record.Create(ctx, 1, RecordCreateOpts{})) },
		"Record.Update":  func(ctx context.Context) error { return e(c.Record.Update(ctx, 1, "2", RecordUpdateOpts{})) },
		"Record.Delete":  func(ctx context.Context) error { return r(c.Record.Delete(ctx, 1, "2")) },
		"Record.Bulk":    func(ctx context.Context) error { return e(c.Record.Bulk(ctx, 1, nil)) },

		"RRSet.Get":    func(ctx context.Context) error { return e(c.RRSet.Get(ctx, 1, "www", "A")) },
		"RRSet.List":   func(ctx context.Context) error { return e(c.RRSet.List(ctx, 1)) },
		"RRSet.Upsert": func(ctx context.Context) error { return r(c.RRSet.Upsert(ctx, 1, RRSetUpsertOpts{})) },
		"RRSet.Delete": func(ctx context.Context) error { return r(c.RRSet.Delete(ctx, 1, "www", "A")) },

		"User.GetByName": func(ctx context.Context) error { return e(c.User.GetByName(ctx, "alice")) },
		"User.GetByID":   func(ctx context.Context) error { return e(c.User.GetByID(ctx, 1)) },
		"User.List":      func(ctx context.Context) error { return e(c.User.List(ctx, ListOpts{})) },
		"User.All":       func(ctx context.Context) error { return s(c.User.All(ctx)) },
		"User.Create":    func(ctx context.Context) error { return e(c.User.Create(ctx, UserCreateOpts{})) },
		"User.Update":    func(ctx context.Context) error { return e(c.User.Update(ctx, 1, UserUpdateOpts{})) },
		"User.Delete": func(ctx context.Context) error {
			_, resp, err := c.User.Delete(ctx, 1, UserDeleteOpts{})
			return r(resp, err)
		},
		"User.SetPermissionTemplate": func(ctx context.Context) error { return r(c.User.SetPermissionTemplate(ctx, 1, 2)) },

		"Group.GetByName":    func(ctx context.Context) error { return e(c.Group.GetByName(ctx, "ops")) },
		"Group.GetByID":      func(ctx context.Context) error { return e(c.Group.GetByID(ctx, 1)) },
		"Group.List":         func(ctx context.Context) error { return e(c.Group.List(ctx, ListOpts{})) },
		"Group.All":          func(ctx context.Context) error { return s(c.Group.All(ctx)) },
		"Group.Create":       func(ctx context.Context) error { return e(c.Group.Create(ctx, GroupCreateOpts{})) },
		"Group.Update":       func(ctx context.Context) error { return e(c.Group.Update(ctx, 1, GroupUpdateOpts{})) },
		"Group.Delete":       func(ctx context.Context) error { return r(c.Group.Delete(ctx, 1)) },
		"Group.Members":      func(ctx context.Context) error { return e(c.Group.Members(ctx, 1)) },
		"Group.AddMember":    func(ctx context.Context) error { return r(c.Group.AddMember(ctx, 1, 2)) },
		"Group.RemoveMember": func(ctx context.Context) error { return r(c.Group.RemoveMember(ctx, 1, 2)) },
		"Group.Zones":        func(ctx context.Context) error { return e(c.Group.Zones(ctx, 1)) },
		"Group.AddZone":      func(ctx context.Context) error { return r(c.Group.AddZone(ctx, 1, 2)) },
		"Group.RemoveZone":   func(ctx context.Context) error { return r(c.Group.RemoveZone(ctx, 1, 2)) },

		"Permission.GetByID": func(ctx context.Context) error { return e(c.Permission.GetByID(ctx, 1)) },
		"Permission.List":    func(ctx context.Context) error { return e(c.Permission.List(ctx, ListOpts{})) },
		"Permission.All":     func(ctx context.Context) error { return s(c.Permission.All(ctx)) },

		"PermissionTemplate.GetByName": func(ctx context.Context) error { return e(c.PermissionTemplate.GetByName(ctx, "x")) },
		"PermissionTemplate.GetByID":   func(ctx context.Context) error { return e(c.PermissionTemplate.GetByID(ctx, 1)) },
		"PermissionTemplate.List":      func(ctx context.Context) error { return e(c.PermissionTemplate.List(ctx)) },
		"PermissionTemplate.Create":    func(ctx context.Context) error { return e(c.PermissionTemplate.Create(ctx, PermissionTemplateOpts{})) },
		"PermissionTemplate.Update": func(ctx context.Context) error {
			return e(c.PermissionTemplate.Update(ctx, 1, PermissionTemplateOpts{}))
		},
		"PermissionTemplate.Delete": func(ctx context.Context) error { return r(c.PermissionTemplate.Delete(ctx, 1)) },

		"ZoneTemplate.GetByName": func(ctx context.Context) error { return e(c.ZoneTemplate.GetByName(ctx, "x")) },
		"ZoneTemplate.GetByID":   func(ctx context.Context) error { return e(c.ZoneTemplate.GetByID(ctx, 1)) },
		"ZoneTemplate.List":      func(ctx context.Context) error { return e(c.ZoneTemplate.List(ctx)) },
		"ZoneTemplate.Create":    func(ctx context.Context) error { return e(c.ZoneTemplate.Create(ctx, ZoneTemplateCreateOpts{})) },
		"ZoneTemplate.Update":    func(ctx context.Context) error { return e(c.ZoneTemplate.Update(ctx, 1, ZoneTemplateUpdateOpts{})) },
		"ZoneTemplate.Delete":    func(ctx context.Context) error { return r(c.ZoneTemplate.Delete(ctx, 1)) },
		"ZoneTemplate.Records":   func(ctx context.Context) error { return e(c.ZoneTemplate.Records(ctx, 1)) },
		"ZoneTemplate.GetRecord": func(ctx context.Context) error { return e(c.ZoneTemplate.GetRecord(ctx, 1, 2)) },
		"ZoneTemplate.CreateRecord": func(ctx context.Context) error {
			return e(c.ZoneTemplate.CreateRecord(ctx, 1, ZoneTemplateRecordOpts{}))
		},
		"ZoneTemplate.UpdateRecord": func(ctx context.Context) error {
			return e(c.ZoneTemplate.UpdateRecord(ctx, 1, 2, ZoneTemplateRecordOpts{}))
		},
		"ZoneTemplate.DeleteRecord": func(ctx context.Context) error { return r(c.ZoneTemplate.DeleteRecord(ctx, 1, 2)) },
	}
}

// Every operation must surface an API error as *APIError with the server's
// status and message.
func TestAllOperationsReturnAPIErrors(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusInternalServerError, "database unavailable")
	})
	for name, op := range allOperations(client) {
		t.Run(name, func(t *testing.T) {
			err := op(context.Background())
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("err = %v, want *APIError", err)
			}
			if apiErr.StatusCode != http.StatusInternalServerError || apiErr.Message != "database unavailable" {
				t.Errorf("APIError = %+v", apiErr)
			}
		})
	}
}

// A 2xx response whose data does not match the expected shape must fail
// loudly instead of returning zero values.
func TestAllOperationsRejectMalformedData(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"ok","data":"not-an-object"}`))
	})
	// Operations without a response body to decode cannot detect this.
	skip := map[string]bool{
		"Zone.Delete": true, "Zone.AddOwner": true, "Zone.AddOwners": true, "Zone.RemoveOwner": true,
		"Zone.SetMetadata": true, "Zone.DeleteMetadata": true, "Record.Delete": true,
		"RRSet.Upsert": true, "RRSet.Delete": true, "User.SetPermissionTemplate": true,
		"Group.Delete": true, "Group.AddMember": true, "Group.RemoveMember": true,
		"Group.AddZone": true, "Group.RemoveZone": true, "PermissionTemplate.Delete": true,
		"ZoneTemplate.Delete": true, "ZoneTemplate.DeleteRecord": true,
	}
	for name, op := range allOperations(client) {
		if skip[name] {
			continue
		}
		t.Run(name, func(t *testing.T) {
			err := op(context.Background())
			if err == nil || !strings.Contains(err.Error(), "unmarshal response data") {
				t.Errorf("err = %v, want unmarshal error", err)
			}
		})
	}
}

// Update operations that read the object back must report a failing GET.
func TestReadBackFailures(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeError(t, w, http.StatusNotFound, "gone")
		default:
			writeEnvelope(t, w, http.StatusOK, nil)
		}
	})
	ctx := context.Background()
	ops := map[string]func() error{
		"User.Update":         func() error { _, _, err := client.User.Update(ctx, 1, UserUpdateOpts{}); return err },
		"ZoneTemplate.Update": func() error { _, _, err := client.ZoneTemplate.Update(ctx, 1, ZoneTemplateUpdateOpts{}); return err },
		"ZoneTemplate.UpdateRecord": func() error {
			_, _, err := client.ZoneTemplate.UpdateRecord(ctx, 1, 2, ZoneTemplateRecordOpts{})
			return err
		},
		"PermissionTemplate.Create": func() error {
			_, _, err := client.PermissionTemplate.Create(ctx, PermissionTemplateOpts{Name: "x"})
			return err
		},
		"PermissionTemplate.Update": func() error {
			_, _, err := client.PermissionTemplate.Update(ctx, 1, PermissionTemplateOpts{})
			return err
		},
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			if err := op(); !IsNotFound(err) {
				t.Errorf("err = %v, want 404", err)
			}
		})
	}
}

// GetByName on the list-scanning lookups reports 404 when nothing matches.
func TestGetByNameNotFoundEverywhere(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(t, w, http.StatusOK, map[string]any{
			"groups": []any{}, "templates": []any{}, "zones": []any{}, "users": []any{},
		})
	})
	ctx := context.Background()
	for name, op := range map[string]func() error{
		"Group":              func() error { _, _, err := client.Group.GetByName(ctx, "x"); return err },
		"PermissionTemplate": func() error { _, _, err := client.PermissionTemplate.GetByName(ctx, "x"); return err },
		"ZoneTemplate":       func() error { _, _, err := client.ZoneTemplate.GetByName(ctx, "x"); return err },
		"Zone":               func() error { _, _, err := client.Zone.GetByName(ctx, "x"); return err },
		"User":               func() error { _, _, err := client.User.GetByName(ctx, "x"); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := op(); !IsNotFound(err) {
				t.Errorf("err = %v, want not found", err)
			}
		})
	}
}

func TestParseInvalidJSON(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html>ok</html>"))
	})
	_, _, err := client.Zone.GetByID(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "parse API response") {
		t.Errorf("err = %v, want parse error", err)
	}
}

// Transport errors (server unreachable) are returned as-is for every verb.
func TestTransportErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	c, err := NewClient(WithBaseURL(url), WithAPIKey("k"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	for name, op := range allOperations(c) {
		t.Run(name, func(t *testing.T) {
			err := op(context.Background())
			var apiErr *APIError
			if err == nil || errors.As(err, &apiErr) {
				t.Errorf("err = %v, want transport error", err)
			}
		})
	}
}
