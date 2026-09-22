// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"errors"
	"fmt"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// ZoneType represents the type of a DNS zone.
type ZoneType string

const (
	ZoneTypeMaster ZoneType = "MASTER"
	ZoneTypeSlave  ZoneType = "SLAVE"
	ZoneTypeNative ZoneType = "NATIVE"
)

// Zone represents a DNS zone in Poweradmin.
type Zone struct {
	ID           int
	Name         string
	Type         ZoneType
	Masters      string
	Account      string
	Description  string
	SOASerial    int
	DNSSECSigned bool
}

// ZoneCreateOpts configures a zone creation request.
type ZoneCreateOpts struct {
	Name string
	Type ZoneType
	// Masters is a comma-separated list of master servers for SLAVE zones,
	// e.g. "192.0.2.1,192.0.2.2:5300" or "[2001:db8::1]:5300".
	Masters     string
	Account     string
	Description string
	// TemplateID applies a zone template (see [ZoneTemplateClient]); 0 means
	// no template.
	TemplateID   int
	EnableDNSSEC bool
	// OwnerUserID assigns a specific user as owner. When nil, the API makes
	// the authenticated user the owner.
	OwnerUserID *int
	// WithoutUserOwner creates a group-only zone with no user owner. It
	// requires a non-empty GroupIDs and a server zone ownership mode that
	// allows groups. It cannot be combined with OwnerUserID.
	WithoutUserOwner bool
	// GroupIDs assigns groups as zone owners.
	GroupIDs []int
}

// ZoneUpdateOpts configures a zone update request.
// Only non-nil pointer fields are sent to the API.
//
// The zone account cannot be changed through the API; set it via
// [ZoneCreateOpts] when creating the zone.
type ZoneUpdateOpts struct {
	// Name renames the zone (FQDN).
	Name *string
	Type *ZoneType
	// Masters is a comma-separated list of master servers for SLAVE zones,
	// e.g. "192.0.2.1,192.0.2.2:5300" or "[2001:db8::1]:5300".
	Masters     *string
	Description *string
}

// ZoneDNSSEC represents the DNSSEC status of a zone.
type ZoneDNSSEC struct {
	Enabled   bool
	DSRecords []DSRecord
	DNSKey    *string
}

// DSRecord represents a DS record for registry submission.
type DSRecord struct {
	KeyTag     int
	Algorithm  int
	DigestType int
	Digest     string
}

// ZoneMetadata represents the values stored under a single metadata kind
// for a zone (e.g. ALLOW-AXFR-FROM, TSIG-ALLOW-AXFR).
type ZoneMetadata struct {
	Kind   string
	Values []string
}

// ZoneClient provides access to the zone-related Poweradmin API endpoints.
type ZoneClient struct {
	client *Client
}

// GetByID returns a single [Zone] by its numeric ID.
func (z *ZoneClient) GetByID(ctx context.Context, id int) (*Zone, *Response, error) {
	var result schema.ZoneResponse
	resp, err := z.client.get(ctx, fmt.Sprintf("zones/%d", id), &result)
	if err != nil {
		return nil, resp, err
	}
	zone := ZoneFromSchema(result.Zone)
	return &zone, resp, nil
}

// GetByName returns a single [Zone] by its DNS name.
// Performs a list + linear search across all pages — no dedicated API endpoint exists.
func (z *ZoneClient) GetByName(ctx context.Context, name string) (*Zone, *Response, error) {
	// TODO: replace with a server-side filter once the Poweradmin API gains one.
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		zones, resp, err := z.List(ctx, opts)
		if err != nil {
			return nil, resp, err
		}
		for _, zone := range zones {
			if zone.Name == name {
				return zone, resp, nil
			}
		}
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.Page >= resp.Meta.Pagination.LastPage {
			return nil, resp, &APIError{StatusCode: 404, Message: fmt.Sprintf("zone not found: %s", name)}
		}
		opts.Page++
	}
}

// List returns one page of [Zone]s.
// Note: /v2/zones wraps the array under data.zones — unlike the other list
// endpoints in this API. The wrapper is unmarshalled here so callers see a
// plain slice.
func (z *ZoneClient) List(ctx context.Context, opts ListOpts) ([]*Zone, *Response, error) {
	path := appendQuery("zones", opts.values())
	var result schema.ZoneListResponse
	resp, err := z.client.get(ctx, path, &result)
	if err != nil {
		return nil, resp, err
	}
	zones := make([]*Zone, len(result.Zones))
	for i, s := range result.Zones {
		zone := ZoneFromSchema(s)
		zones[i] = &zone
	}
	return zones, resp, nil
}

// All returns all [Zone]s across all pages.
func (z *ZoneClient) All(ctx context.Context) ([]*Zone, error) {
	var all []*Zone
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		zones, resp, err := z.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, zones...)
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.Page >= resp.Meta.Pagination.LastPage {
			return all, nil
		}
		opts.Page++
	}
}

// Create creates a new [Zone] and returns the new ID.
// Call [ZoneClient.GetByID] to fetch the full object.
func (z *ZoneClient) Create(ctx context.Context, opts ZoneCreateOpts) (int, *Response, error) {
	if opts.WithoutUserOwner && opts.OwnerUserID != nil {
		return 0, nil, errors.New("poweradmin: ZoneCreateOpts: OwnerUserID and WithoutUserOwner are mutually exclusive")
	}
	if opts.WithoutUserOwner && len(opts.GroupIDs) == 0 {
		return 0, nil, errors.New("poweradmin: ZoneCreateOpts: WithoutUserOwner requires GroupIDs")
	}
	req := schema.ZoneCreateRequest{
		Name:         opts.Name,
		Type:         string(opts.Type),
		Master:       opts.Masters,
		Account:      opts.Account,
		Description:  opts.Description,
		Template:     opts.TemplateID,
		EnableDNSSEC: opts.EnableDNSSEC,
		GroupIDs:     opts.GroupIDs,
	}
	switch {
	case opts.WithoutUserOwner:
		req.OwnerUserID = schema.OwnerUserIDNull()
	case opts.OwnerUserID != nil:
		req.OwnerUserID = schema.OwnerUserIDValue(*opts.OwnerUserID)
	}
	var result schema.ZoneCreateResponse
	resp, err := z.client.post(ctx, "zones", req, &result)
	if err != nil {
		return 0, resp, err
	}
	return result.ZoneID, resp, nil
}

// Update updates an existing [Zone] and returns the updated state.
func (z *ZoneClient) Update(ctx context.Context, id int, opts ZoneUpdateOpts) (*Zone, *Response, error) {
	req := schema.ZoneUpdateRequest{
		Name:        opts.Name,
		Master:      opts.Masters,
		Description: opts.Description,
	}
	if opts.Type != nil {
		t := string(*opts.Type)
		req.Type = &t
	}

	var result schema.ZoneResponse
	resp, err := z.client.put(ctx, fmt.Sprintf("zones/%d", id), req, &result)
	if err != nil {
		return nil, resp, err
	}
	zone := ZoneFromSchema(result.Zone)
	return &zone, resp, nil
}

// Delete deletes the [Zone] with the given ID.
func (z *ZoneClient) Delete(ctx context.Context, id int) (*Response, error) {
	return z.client.delete(ctx, fmt.Sprintf("zones/%d", id))
}

// ── Zone Owners ──────────────────────────────────────────────────────────────

// ZoneOwner represents a user who owns a zone.
type ZoneOwner struct {
	UserID   int
	Username string
	Fullname string
}

// Owners returns all owners of the zone.
func (z *ZoneClient) Owners(ctx context.Context, zoneID int) ([]*ZoneOwner, *Response, error) {
	var result schema.ZoneOwnerListResponse
	resp, err := z.client.get(ctx, fmt.Sprintf("zones/%d/owners", zoneID), &result)
	if err != nil {
		return nil, resp, err
	}
	owners := make([]*ZoneOwner, len(result.Owners))
	for i, s := range result.Owners {
		owners[i] = &ZoneOwner{UserID: s.UserID, Username: s.Username, Fullname: s.Fullname}
	}
	return owners, resp, nil
}

// AddOwner adds a single user as owner of the zone.
func (z *ZoneClient) AddOwner(ctx context.Context, zoneID, userID int) (*Response, error) {
	req := schema.ZoneOwnerAddRequest{UserID: userID}
	return z.client.post(ctx, fmt.Sprintf("zones/%d/owners", zoneID), req, nil)
}

// AddOwners adds multiple users as owners in a single request.
func (z *ZoneClient) AddOwners(ctx context.Context, zoneID int, userIDs []int) (*Response, error) {
	req := schema.ZoneOwnerAddRequest{UserIDs: userIDs}
	return z.client.post(ctx, fmt.Sprintf("zones/%d/owners", zoneID), req, nil)
}

// RemoveOwner removes a user from the zone's owners.
func (z *ZoneClient) RemoveOwner(ctx context.Context, zoneID, userID int) (*Response, error) {
	return z.client.delete(ctx, fmt.Sprintf("zones/%d/owners/%d", zoneID, userID))
}

// ── DNSSEC ───────────────────────────────────────────────────────────────────

// GetDNSSEC returns the DNSSEC status of the zone with the given ID.
func (z *ZoneClient) GetDNSSEC(ctx context.Context, id int) (*ZoneDNSSEC, *Response, error) {
	var result schema.ZoneDNSSECResponse
	resp, err := z.client.get(ctx, fmt.Sprintf("zones/%d/dnssec", id), &result)
	if err != nil {
		return nil, resp, err
	}
	dnssec := ZoneDNSSECFromSchema(result)
	return &dnssec, resp, nil
}

// SetDNSSEC enables or disables DNSSEC for the zone with the given ID.
func (z *ZoneClient) SetDNSSEC(ctx context.Context, id int, enabled bool) (*ZoneDNSSEC, *Response, error) {
	req := schema.ZoneDNSSECSetRequest{Enabled: enabled}
	var result schema.ZoneDNSSECResponse
	resp, err := z.client.post(ctx, fmt.Sprintf("zones/%d/dnssec", id), req, &result)
	if err != nil {
		return nil, resp, err
	}
	dnssec := ZoneDNSSECFromSchema(result)
	return &dnssec, resp, nil
}

// ListMetadata returns all metadata entries for the zone with the given ID.
func (z *ZoneClient) ListMetadata(ctx context.Context, zoneID int) ([]*ZoneMetadata, *Response, error) {
	var result schema.ZoneMetadataListResponse
	resp, err := z.client.get(ctx, fmt.Sprintf("zones/%d/metadata", zoneID), &result)
	if err != nil {
		return nil, resp, err
	}
	metadata := make([]*ZoneMetadata, len(result.Metadata))
	for i, m := range result.Metadata {
		metadata[i] = &ZoneMetadata{Kind: m.Kind, Values: m.Values}
	}
	return metadata, resp, nil
}

// GetMetadata returns the values stored under a specific metadata kind
// (e.g. ALLOW-AXFR-FROM) for the zone with the given ID.
func (z *ZoneClient) GetMetadata(ctx context.Context, zoneID int, kind string) (*ZoneMetadata, *Response, error) {
	var result schema.ZoneMetadataResponse
	resp, err := z.client.get(ctx, fmt.Sprintf("zones/%d/metadata/%s", zoneID, kind), &result)
	if err != nil {
		return nil, resp, err
	}
	return &ZoneMetadata{Kind: result.Kind, Values: result.Values}, resp, nil
}

// SetMetadata creates or replaces all values for a metadata kind on the zone
// with the given ID.
func (z *ZoneClient) SetMetadata(ctx context.Context, zoneID int, kind string, values []string) (*Response, error) {
	req := schema.ZoneMetadataSetRequest{Values: values}
	return z.client.put(ctx, fmt.Sprintf("zones/%d/metadata/%s", zoneID, kind), req, nil)
}

// DeleteMetadata deletes all values for a metadata kind on the zone with the
// given ID.
func (z *ZoneClient) DeleteMetadata(ctx context.Context, zoneID int, kind string) (*Response, error) {
	return z.client.delete(ctx, fmt.Sprintf("zones/%d/metadata/%s", zoneID, kind))
}
