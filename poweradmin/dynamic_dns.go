// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"strings"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// DynamicDNSUpdateOpts configures a dynamic DNS update.
type DynamicDNSUpdateOpts struct {
	// Hostname is the fully qualified name to update, e.g. "host.example.com".
	// It must lie in a zone the authenticated user owns.
	Hostname string
	// IPv4 and IPv6 are the addresses the hostname should resolve to. At
	// least one address is required unless Dualstack is set.
	IPv4 []string
	IPv6 []string
	// Dualstack removes the records of the address family that has no value,
	// e.g. the AAAA records when only IPv4 is given.
	Dualstack bool
}

// DynamicDNSResult is the outcome of a dynamic DNS update.
type DynamicDNSResult struct {
	Hostname    string
	ZoneID      int
	AppliedIPv4 []string
	AppliedIPv6 []string
	// Changed is false when the records already matched the addresses.
	Changed bool
}

// DynamicDNSClient updates A and AAAA records the way a dynamic DNS client does.
type DynamicDNSClient struct {
	client *Client
}

// Update sets the A and AAAA records of a hostname to the given addresses,
// creating or replacing records as needed.
//
// Requires Poweradmin 4.5 or later and the zone_content_edit_own (or
// zone_content_edit_own_as_client) permission; only zones the user owns are
// considered. Unlike other endpoints, user_is_ueberuser alone is not enough:
// Poweradmin 4.5.0 answers 403 for an administrator without one of these
// permissions. A scoped API key needs the create and update operations, and
// delete as well when the update removes records. The server answers 404 when
// no owned zone contains the hostname, 403 for secondary and consumer zones,
// and 409 when the update would change nothing and no matching records exist
// ([IsConflict]).
func (d *DynamicDNSClient) Update(ctx context.Context, opts DynamicDNSUpdateOpts) (*DynamicDNSResult, *Response, error) {
	req := schema.DynamicDNSUpdateRequest{
		Hostname:  opts.Hostname,
		IPv4:      strings.Join(opts.IPv4, ","),
		IPv6:      strings.Join(opts.IPv6, ","),
		Dualstack: opts.Dualstack,
	}
	var result schema.DynamicDNSUpdateResponse
	resp, err := d.client.post(ctx, "dynamic-dns", req, &result)
	if err != nil {
		return nil, resp, err
	}
	return &DynamicDNSResult{
		Hostname:    result.Hostname,
		ZoneID:      result.ZoneID,
		AppliedIPv4: result.AppliedIPv4,
		AppliedIPv6: result.AppliedIPv6,
		Changed:     result.Changed,
	}, resp, nil
}
