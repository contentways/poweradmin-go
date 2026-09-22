// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"reflect"
	"testing"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

func TestZoneRoundtrip(t *testing.T) {
	in := schema.Zone{
		ID:          7,
		Name:        "example.com",
		Type:        "MASTER",
		Masters:     "1.2.3.4",
		Account:     "acct",
		Description: "desc",
		CreatedAt:   "2026-01-01 12:00:00",
	}
	domain := ZoneFromSchema(in)
	if domain.Type != ZoneTypeMaster {
		t.Errorf("Type = %q, want MASTER", domain.Type)
	}
	back := ZoneToSchema(domain)
	if !reflect.DeepEqual(in, back) {
		t.Errorf("roundtrip mismatch:\n got %#v\nwant %#v", back, in)
	}
}

func TestUserFromSchemaMapsUserID(t *testing.T) {
	in := schema.User{
		UserID:   13,
		Username: "alice",
		Email:    "alice@example.com",
		IsAdmin:  true,
	}
	got := UserFromSchema(in)
	if got.ID != 13 {
		t.Errorf("ID = %d, want 13", got.ID)
	}
	if got.Username != "alice" || got.Email != "alice@example.com" || !got.IsAdmin {
		t.Errorf("field mapping wrong: %+v", got)
	}
}

func TestRecordRoundtrip(t *testing.T) {
	in := schema.Record{
		ID:       "rec-1",
		ZoneID:   2,
		Name:     "host.example.com",
		Type:     "A",
		Content:  "10.0.0.1",
		TTL:      300,
		Priority: 0,
	}
	if back := RecordToSchema(RecordFromSchema(in)); !reflect.DeepEqual(in, back) {
		t.Errorf("roundtrip mismatch:\n got %#v\nwant %#v", back, in)
	}
}

func TestGroupAndPermissionConv(t *testing.T) {
	g := schema.Group{ID: 1, Name: "ops", Description: "team", PermTemplID: 2}
	if got := GroupFromSchema(g); got.ID != 1 || got.Name != "ops" || got.PermTemplID != 2 {
		t.Errorf("Group conv wrong: %+v", got)
	}

	p := schema.Permission{ID: 9, Name: "zone_master_add", Descr: "Add master zones"}
	if got := PermissionFromSchema(p); got.ID != 9 || got.Name != "zone_master_add" {
		t.Errorf("Permission conv wrong: %+v", got)
	}
}

func TestGroupMemberAndZoneFromSchema(t *testing.T) {
	m := GroupMemberFromSchema(schema.GroupMember{
		UserID: 1, Username: "admin", Fullname: "Administrator", Email: "admin@example.com", JoinedAt: "2026-01-01",
	})
	if m != (GroupMember{UserID: 1, Username: "admin", Fullname: "Administrator", Email: "admin@example.com", JoinedAt: "2026-01-01"}) {
		t.Errorf("member = %+v", m)
	}
	z := GroupZoneFromSchema(schema.GroupZone{ZoneID: 2, ZoneName: "example.com", ZoneType: "MASTER", CreatedAt: "2026-01-02"})
	if z != (GroupZone{ZoneID: 2, ZoneName: "example.com", ZoneType: "MASTER", CreatedAt: "2026-01-02"}) {
		t.Errorf("zone = %+v", z)
	}
}

func TestTemplateConversions(t *testing.T) {
	pt := PermissionTemplateFromSchema(schema.PermissionTemplate{
		ID: 1, Name: "Zone Admin", Descr: "desc", TemplateType: "group",
		Permissions: []schema.Permission{{ID: 3, Name: "zone_content_view_own", Descr: "view"}},
	})
	if pt.ID != 1 || pt.TemplateType != "group" || len(pt.Permissions) != 1 || pt.Permissions[0].Name != "zone_content_view_own" {
		t.Errorf("permission template = %+v", pt)
	}
	if empty := PermissionTemplateFromSchema(schema.PermissionTemplate{}); empty.Permissions == nil || len(empty.Permissions) != 0 {
		t.Errorf("Permissions = %#v, want empty non-nil slice", empty.Permissions)
	}

	zt := ZoneTemplateFromSchema(schema.ZoneTemplate{ID: 2, Name: "Default", Description: "d", Owner: 1, IsGlobal: true, ZonesLinked: 5})
	if zt != (ZoneTemplate{ID: 2, Name: "Default", Description: "d", Owner: 1, IsGlobal: true, ZonesLinked: 5}) {
		t.Errorf("zone template = %+v", zt)
	}

	rec := ZoneTemplateRecordFromSchema(schema.ZoneTemplateRecord{ID: 7, Name: "[ZONE]", Type: "MX", Content: "mail.[ZONE]", TTL: 3600, Priority: 10})
	if rec != (ZoneTemplateRecord{ID: 7, Name: "[ZONE]", Type: "MX", Content: "mail.[ZONE]", TTL: 3600, Priority: 10}) {
		t.Errorf("zone template record = %+v", rec)
	}
}
