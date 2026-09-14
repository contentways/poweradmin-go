// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v3/poweradmin/schema"
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
	Name        string
	Type        ZoneType
	Masters     string
	Account     string
	Description string
	Template    string
}

// ZoneUpdateOpts configures a zone update request.
// Only non-nil pointer fields are sent to the API.
type ZoneUpdateOpts struct {
	Type        *ZoneType
	Masters     *string
	Account     *string
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
	req := schema.ZoneCreateRequest{
		Name:        opts.Name,
		Type:        string(opts.Type),
		Masters:     opts.Masters,
		Account:     opts.Account,
		Description: opts.Description,
		Template:    opts.Template,
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
	req := schema.ZoneUpdateRequest{}
	if opts.Type != nil {
		t := string(*opts.Type)
		req.Type = &t
	}
	req.Masters = opts.Masters
	req.Account = opts.Account
	req.Description = opts.Description

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
