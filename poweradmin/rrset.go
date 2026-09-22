// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// RRSet represents a DNS Resource Record Set. RRSets group all records that
// share the same name and type within a zone. The Poweradmin API identifies
// them by (name, type) — there is no numeric ID.
type RRSet struct {
	Name    string
	Type    string
	TTL     int
	Records []RRSetRecord
}

// RRSetRecord is one entry within an [RRSet].
type RRSetRecord struct {
	Content  string
	Priority int
	Disabled bool
}

// RRSetUpsertOpts configures a PUT request that creates or replaces an RRSet.
type RRSetUpsertOpts struct {
	Name    string
	Type    string
	TTL     int
	Records []RRSetRecord
}

// RRSetClient provides access to the rrset-related Poweradmin API endpoints.
type RRSetClient struct {
	client *Client
}

// Get returns a single [RRSet] identified by name and type within a zone.
func (r *RRSetClient) Get(ctx context.Context, zoneID int, name, recordType string) (*RRSet, *Response, error) {
	var result schema.RRSetResponse
	resp, err := r.client.get(ctx, fmt.Sprintf("zones/%d/rrsets/%s/%s", zoneID, name, recordType), &result)
	if err != nil {
		return nil, resp, err
	}
	rr := RRSetFromSchema(result.RRSet)
	return &rr, resp, nil
}

// List returns all [RRSet]s in a zone.
func (r *RRSetClient) List(ctx context.Context, zoneID int) ([]*RRSet, *Response, error) {
	var result schema.RRSetListResponse
	resp, err := r.client.get(ctx, fmt.Sprintf("zones/%d/rrsets", zoneID), &result)
	if err != nil {
		return nil, resp, err
	}
	rrsets := make([]*RRSet, len(result.RRSets))
	for i, s := range result.RRSets {
		rr := RRSetFromSchema(s)
		rrsets[i] = &rr
	}
	return rrsets, resp, nil
}

// Upsert creates or replaces an [RRSet] via PUT. The API treats this as an
// idempotent create-or-update on (name, type) within the zone.
func (r *RRSetClient) Upsert(ctx context.Context, zoneID int, opts RRSetUpsertOpts) (*Response, error) {
	records := make([]schema.RRSetRecord, len(opts.Records))
	for i, rec := range opts.Records {
		records[i] = schema.RRSetRecord{
			Content:  rec.Content,
			Priority: rec.Priority,
			Disabled: rec.Disabled,
		}
	}
	req := schema.RRSetUpsertRequest{
		Name:    opts.Name,
		Type:    opts.Type,
		TTL:     opts.TTL,
		Records: records,
	}
	return r.client.put(ctx, fmt.Sprintf("zones/%d/rrsets", zoneID), req, nil)
}

// Delete deletes the [RRSet] identified by name and type within a zone.
func (r *RRSetClient) Delete(ctx context.Context, zoneID int, name, recordType string) (*Response, error) {
	return r.client.delete(ctx, fmt.Sprintf("zones/%d/rrsets/%s/%s", zoneID, name, recordType))
}
