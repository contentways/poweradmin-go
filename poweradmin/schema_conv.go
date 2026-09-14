// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import "github.com/contentways/poweradmin-go/v3/poweradmin/schema"

// This file holds the schema ↔ domain conversion functions.
// Hand-written rather than goverter-generated to avoid a bootstrap dependency
// where the package can't compile without the generated converter, but
// goverter can't generate without the package compiling.
//
// Add a new pair whenever a new resource type is introduced.

func ZoneFromSchema(s schema.Zone) Zone {
	return Zone{
		ID:           s.ID,
		Name:         s.Name,
		Type:         ZoneType(s.Type),
		Masters:      s.Masters,
		Account:      s.Account,
		Description:  s.Description,
		SOASerial:    s.SOASerial,
		DNSSECSigned: s.DNSSECSigned,
	}
}

func ZoneToSchema(z Zone) schema.Zone {
	return schema.Zone{
		ID:           z.ID,
		Name:         z.Name,
		Type:         string(z.Type),
		Masters:      z.Masters,
		Account:      z.Account,
		Description:  z.Description,
		SOASerial:    z.SOASerial,
		DNSSECSigned: z.DNSSECSigned,
	}
}

func RecordFromSchema(s schema.Record) Record {
	return Record{
		ID:       s.ID,
		ZoneID:   s.ZoneID,
		Name:     s.Name,
		Type:     s.Type,
		Content:  s.Content,
		TTL:      s.TTL,
		Priority: s.Priority,
		Disabled: s.Disabled,
		Auth:     s.Auth,
	}
}

func RecordToSchema(r Record) schema.Record {
	return schema.Record{
		ID:       r.ID,
		ZoneID:   r.ZoneID,
		Name:     r.Name,
		Type:     r.Type,
		Content:  r.Content,
		TTL:      r.TTL,
		Priority: r.Priority,
		Disabled: r.Disabled,
		Auth:     r.Auth,
	}
}

func RRSetFromSchema(s schema.RRSet) RRSet {
	records := make([]RRSetRecord, len(s.Records))
	for i, sr := range s.Records {
		records[i] = RRSetRecord{Content: sr.Content, Priority: sr.Priority, Disabled: sr.Disabled}
	}
	return RRSet{
		Name:    s.Name,
		Type:    s.Type,
		TTL:     s.TTL,
		Records: records,
	}
}

func UserFromSchema(s schema.User) User {
	return User{
		ID:          s.UserID,
		Username:    s.Username,
		Fullname:    s.Fullname,
		Email:       s.Email,
		Description: s.Description,
		Active:      s.Active,
		PermTempl:   s.PermTempl,
		UseLdap:     s.UseLdap,
		IsAdmin:     s.IsAdmin,
		Permissions: s.Permissions,
		ZoneCount:   s.ZoneCount,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func GroupFromSchema(s schema.Group) Group {
	return Group{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		PermTemplID: s.PermTemplID,
		MemberCount: s.MemberCount,
		ZoneCount:   s.ZoneCount,
		CreatedAt:   s.CreatedAt,
	}
}

func GroupMemberFromSchema(s schema.GroupMember) GroupMember {
	return GroupMember{
		UserID:   s.UserID,
		Username: s.Username,
		JoinedAt: s.JoinedAt,
	}
}

func GroupZoneFromSchema(s schema.GroupZone) GroupZone {
	return GroupZone{
		ZoneID:   s.ZoneID,
		ZoneName: s.ZoneName,
		ZoneType: s.ZoneType,
	}
}

func PermissionFromSchema(s schema.Permission) Permission {
	return Permission{
		ID:    s.ID,
		Name:  s.Name,
		Descr: s.Descr,
	}
}

func ZoneDNSSECFromSchema(s schema.ZoneDNSSECResponse) ZoneDNSSEC {
	records := make([]DSRecord, len(s.DSRecords))
	for i, r := range s.DSRecords {
		records[i] = DSRecord{
			KeyTag:     r.KeyTag,
			Algorithm:  r.Algorithm,
			DigestType: r.DigestType,
			Digest:     r.Digest,
		}
	}
	return ZoneDNSSEC{
		Enabled:   s.Enabled,
		DSRecords: records,
		DNSKey:    s.DNSKey,
	}
}
