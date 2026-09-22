// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT

// Package schema contains the raw JSON types that map 1:1 to the Poweradmin API.
// These types are not meant to be used directly by consumers of the SDK.
// Use the domain types in the parent package instead.
//
// Shapes follow Poweradmin API v2, as standardized in Poweradmin 4.3.0:
//   - Envelope: {success, message, data, pagination, meta, error}
//   - Pagination lives at envelope level, not inside data.
//   - List endpoints wrap the array under a named key inside data
//     (e.g. data.zones, data.records, data.rrsets, data.owners).
//   - Single-resource endpoints wrap the object under its resource key
//     (e.g. data.zone, data.record, data.rrset).
//
// Releases before 4.3.0 returned several of these collections as bare arrays;
// this client targets the 4.3.0+ wrapped shape. See the version matrix in the
// README.
package schema

import "encoding/json"

// APIResponse represents the Poweradmin API envelope. Data is RawMessage so
// callers can unmarshal it into the concrete shape without a marshal roundtrip.
type APIResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data,omitempty"`
	Pagination *Pagination     `json:"pagination,omitempty"`
	Meta       *APIMeta        `json:"meta,omitempty"`
	Error      *ErrorPayload   `json:"error,omitempty"`
}

type APIMeta struct {
	Timestamp string `json:"timestamp,omitempty"`
}

type ErrorPayload struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

// ── Zone ─────────────────────────────────────────────────────────────────────

type Zone struct {
	ID           int    `json:"id,omitempty"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Masters      string `json:"masters,omitempty"`
	Account      string `json:"account,omitempty"`
	Description  string `json:"description,omitempty"`
	SOASerial    int    `json:"soa_serial,omitempty"`
	DNSSECSigned bool   `json:"dnssec_signed,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
}

type ZoneResponse struct {
	Zone Zone `json:"zone"`
}

type ZoneListResponse struct {
	Zones []Zone `json:"zones"`
}

type ZoneCreateRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Masters     string `json:"masters,omitempty"`
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"`
}

type ZoneCreateResponse struct {
	ZoneID int `json:"zone_id"`
}

// ZoneUpdateRequest is sent via PUT /v2/zones/{id}. The server only reads
// name, type, master (singular, unlike the "masters" field it returns) and
// description. The account cannot be changed after creation.
type ZoneUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Master      *string `json:"master,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ── DNSSEC ───────────────────────────────────────────────────────────────────

type DSRecord struct {
	KeyTag     int    `json:"key_tag"`
	Algorithm  int    `json:"algorithm"`
	DigestType int    `json:"digest_type"`
	Digest     string `json:"digest"`
}

type ZoneDNSSECResponse struct {
	Enabled   bool       `json:"enabled"`
	DSRecords []DSRecord `json:"ds_records"`
	DNSKey    *string    `json:"dnskey"`
}

type ZoneDNSSECSetRequest struct {
	Enabled bool `json:"enabled"`
}

// ── Zone Metadata ────────────────────────────────────────────────────────────

type ZoneMetadataEntry struct {
	Kind   string   `json:"kind"`
	Values []string `json:"values"`
}

type ZoneMetadataListResponse struct {
	Metadata []ZoneMetadataEntry `json:"metadata"`
}

type ZoneMetadataResponse struct {
	Kind   string   `json:"kind"`
	Values []string `json:"values"`
}

type ZoneMetadataSetRequest struct {
	Values []string `json:"values"`
}

// ── Record ───────────────────────────────────────────────────────────────────

type Record struct {
	ID       RecordID `json:"id,omitempty"`
	ZoneID   int64    `json:"zone_id,omitempty"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Content  string   `json:"content"`
	TTL      int      `json:"ttl"`
	Priority int      `json:"priority,omitempty"`
	Disabled bool     `json:"disabled"`
	Auth     bool     `json:"auth,omitempty"`
}

// RecordResponse is used for both single-record GET and the POST create
// response. The API wraps the record under data.record in both cases.
type RecordResponse struct {
	Record Record `json:"record"`
}

// RecordListResponse wraps the array returned by GET /v2/zones/{id}/records.
type RecordListResponse struct {
	Records []Record `json:"records"`
}

type RecordCreateRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	TTL       int    `json:"ttl"`
	Priority  int    `json:"priority,omitempty"`
	Disabled  bool   `json:"disabled,omitempty"`
	CreatePTR bool   `json:"create_ptr,omitempty"`
}

// RecordUpdateRequest is sent via PUT /v2/zones/{id}/records/{id}. Omitted
// fields keep their current value on the server.
type RecordUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Type     *string `json:"type,omitempty"`
	Content  *string `json:"content,omitempty"`
	TTL      *int    `json:"ttl,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Disabled *bool   `json:"disabled,omitempty"`
}

type BulkRecordOperation struct {
	Action   string `json:"action"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type BulkRecordsRequest struct {
	Operations []BulkRecordOperation `json:"operations"`
}

type BulkRecordsResponse struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Deleted int      `json:"deleted"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// ── RRSet ────────────────────────────────────────────────────────────────────
// RRSet (Resource Record Set) groups records sharing a name+type. The API
// identifies them by name+type within a zone, not by ID. Records are objects,
// not strings.

type RRSet struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	TTL     int           `json:"ttl"`
	Records []RRSetRecord `json:"records"`
}

type RRSetRecord struct {
	Content  string `json:"content"`
	Priority int    `json:"priority,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

// RRSetResponse wraps a single RRSet returned under data.rrset.
type RRSetResponse struct {
	RRSet RRSet `json:"rrset"`
}

// RRSetListResponse wraps the array returned under data.rrsets.
type RRSetListResponse struct {
	RRSets []RRSet `json:"rrsets"`
}

// RRSetUpsertRequest is sent via PUT to create-or-replace a full RRSet.
type RRSetUpsertRequest struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	TTL     int           `json:"ttl"`
	Records []RRSetRecord `json:"records"`
}

// ── User ─────────────────────────────────────────────────────────────────────

type User struct {
	UserID      int      `json:"user_id,omitempty"`
	Username    string   `json:"username"`
	Fullname    string   `json:"fullname"`
	Email       string   `json:"email"`
	Description string   `json:"description,omitempty"`
	Active      bool     `json:"active"`
	PermTempl   int      `json:"perm_templ,omitempty"`
	UseLdap     bool     `json:"use_ldap,omitempty"`
	IsAdmin     bool     `json:"is_admin,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	ZoneCount   int      `json:"zone_count,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

type UserResponse struct {
	User User `json:"user"`
}

// UserListResponse wraps the array returned under data.users.
type UserListResponse struct {
	Users []User `json:"users"`
}

type UserCreateResponse struct {
	UserID int `json:"user_id"`
}

type UserCreateRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active"`
	PermTempl   int    `json:"perm_templ,omitempty"`
	UseLdap     bool   `json:"use_ldap,omitempty"`
}

// UserUpdateRequest is sent via PUT /v2/users/{id}. Omitted fields keep their
// current value on the server; the server answers with data.user_id only.
type UserUpdateRequest struct {
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
	Fullname    *string `json:"fullname,omitempty"`
	Email       *string `json:"email,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
	PermTempl   *int    `json:"perm_templ,omitempty"`
	UseLdap     *bool   `json:"use_ldap,omitempty"`
}

// UserDeleteRequest is the optional body of DELETE /v2/users/{id}. The server
// requires TransferToUserID when the user still owns zones.
type UserDeleteRequest struct {
	TransferToUserID *int `json:"transfer_to_user_id,omitempty"`
}

// UserDeleteResponse is returned under data by DELETE /v2/users/{id}.
type UserDeleteResponse struct {
	ZonesAffected int `json:"zones_affected"`
}

// UserPatchRequest is sent via PATCH /v2/users/{id} for partial updates such as
// assigning a permission template.
type UserPatchRequest struct {
	PermTempl int `json:"perm_templ,omitempty"`
}

// ── Permission ───────────────────────────────────────────────────────────────

type Permission struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Descr string `json:"descr"`
}

type PermissionResponse struct {
	Permission Permission `json:"permission"`
}

// PermissionListResponse wraps the array returned under data.permissions.
type PermissionListResponse struct {
	Permissions []Permission `json:"permissions"`
}

// ── Permission Template ──────────────────────────────────────────────────────

type PermissionTemplate struct {
	ID           int          `json:"id,omitempty"`
	Name         string       `json:"name"`
	Descr        string       `json:"descr"`
	TemplateType string       `json:"template_type,omitempty"`
	Permissions  []Permission `json:"permissions,omitempty"`
}

type PermissionTemplateResponse struct {
	Template PermissionTemplate `json:"template"`
}

// PermissionTemplateListResponse wraps the array returned under data.templates.
type PermissionTemplateListResponse struct {
	Templates []PermissionTemplate `json:"templates"`
}

type PermissionTemplateRequest struct {
	Name         string `json:"name"`
	Descr        string `json:"descr"`
	TemplateType string `json:"template_type,omitempty"`
	Permissions  []int  `json:"permissions,omitempty"`
}

// ── Group ────────────────────────────────────────────────────────────────────

type Group struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PermTemplID int    `json:"perm_templ_id,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`
	ZoneCount   int    `json:"zone_count,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type GroupResponse struct {
	Group Group `json:"group"`
}

// GroupListResponse wraps the array returned under data.groups.
type GroupListResponse struct {
	Groups []Group `json:"groups"`
}

type GroupCreateResponse struct {
	Group Group `json:"group"`
}

type GroupCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PermTemplID int    `json:"perm_templ_id"`
}

// GroupUpdateRequest is sent via PUT /v2/groups/{id}. Omitted fields keep
// their current value on the server.
type GroupUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	PermTemplID *int    `json:"perm_templ_id,omitempty"`
}

type GroupMember struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Fullname string `json:"fullname,omitempty"`
	Email    string `json:"email,omitempty"`
	JoinedAt string `json:"joined_at,omitempty"`
}

type GroupMemberRequest struct {
	UserID int `json:"user_id"`
}

// GroupMemberListResponse wraps the array returned under data.members.
type GroupMemberListResponse struct {
	Members []GroupMember `json:"members"`
}

type GroupZone struct {
	ZoneID    int    `json:"zone_id"`
	ZoneName  string `json:"zone_name"`
	ZoneType  string `json:"zone_type"`
	CreatedAt string `json:"created_at,omitempty"`
}

type GroupZoneRequest struct {
	ZoneID int `json:"zone_id"`
}

// GroupZoneListResponse wraps the array returned under data.zones.
type GroupZoneListResponse struct {
	Zones []GroupZone `json:"zones"`
}

// ── Zone Owner ───────────────────────────────────────────────────────────────

type ZoneOwner struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Fullname string `json:"fullname,omitempty"`
}

// ZoneOwnerAddRequest supports single or batch add.
type ZoneOwnerAddRequest struct {
	UserID  int   `json:"user_id,omitempty"`
	UserIDs []int `json:"user_ids,omitempty"`
}

// ZoneOwnerListResponse wraps the array returned under data.owners.
type ZoneOwnerListResponse struct {
	Owners []ZoneOwner `json:"owners"`
}

// ── Zone Template ────────────────────────────────────────────────────────────

type ZoneTemplate struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       int    `json:"owner,omitempty"`
	IsGlobal    bool   `json:"is_global,omitempty"`
	ZonesLinked int    `json:"zones_linked,omitempty"`
}

type ZoneTemplateResponse struct {
	Template ZoneTemplate `json:"template"`
}

// ZoneTemplateListResponse wraps the array returned under data.templates.
type ZoneTemplateListResponse struct {
	Templates []ZoneTemplate `json:"templates"`
}

type ZoneTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsGlobal    bool   `json:"is_global,omitempty"`
}

// ZoneTemplateCreateResponse is returned by POST /v2/zone-templates, which
// reports only the new ID rather than the full object.
type ZoneTemplateCreateResponse struct {
	ID int `json:"id"`
}

type ZoneTemplateRecord struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

type ZoneTemplateRecordResponse struct {
	Record ZoneTemplateRecord `json:"record"`
}

// ZoneTemplateRecordListResponse wraps the array returned under data.records.
type ZoneTemplateRecordListResponse struct {
	Records []ZoneTemplateRecord `json:"records"`
}

type ZoneTemplateRecordRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// ZoneTemplateRecordCreateResponse is returned by
// POST /v2/zone-templates/{id}/records, which reports only the new record ID.
type ZoneTemplateRecordCreateResponse struct {
	ID int `json:"id"`
}
