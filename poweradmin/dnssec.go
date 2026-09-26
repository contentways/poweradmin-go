// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// DNSSECKeyType is the role of a DNSSEC key.
type DNSSECKeyType string

const (
	// DNSSECKeyTypeKSK is a key signing key: it signs the DNSKEY RRset.
	DNSSECKeyTypeKSK DNSSECKeyType = "ksk"
	// DNSSECKeyTypeZSK is a zone signing key: it signs all other RRsets.
	DNSSECKeyTypeZSK DNSSECKeyType = "zsk"
	// DNSSECKeyTypeCSK is a combined signing key that acts as both KSK and ZSK.
	DNSSECKeyTypeCSK DNSSECKeyType = "csk"
)

// DNSSECKey is a DNSSEC key of a zone.
type DNSSECKey struct {
	ID     int
	Type   DNSSECKeyType
	KeyTag int
	// Algorithm is the PowerDNS algorithm name (e.g. "ecdsa256"). It is empty
	// for algorithms Poweradmin does not know by name; AlgorithmID is always set.
	Algorithm   string
	AlgorithmID int
	Bits        int
	Active      bool
}

// DNSSECKeyCreateOpts describes a new DNSSEC key.
//
// Poweradmin validates the combination of Algorithm and Bits, e.g. "ecdsa256"
// requires 256 bits and the RSA algorithms 1024 or 2048 bits.
type DNSSECKeyCreateOpts struct {
	Type      DNSSECKeyType
	Algorithm string
	Bits      int
}

// DNSSECClient manages DNSSEC keys and rectification of zones.
//
// Enabling or disabling DNSSEC for a zone is done with [ZoneClient.SetDNSSEC].
// All methods here require Poweradmin 4.5 or later with the PowerDNS API
// configured; older servers answer 404, a server without the PowerDNS API 501.
type DNSSECClient struct {
	client *Client
}

// ListKeys returns all DNSSEC keys of the zone with the given ID.
func (d *DNSSECClient) ListKeys(ctx context.Context, zoneID int) ([]*DNSSECKey, *Response, error) {
	var result []schema.DNSSECKey
	resp, err := d.client.get(ctx, fmt.Sprintf("zones/%d/dnssec/keys", zoneID), &result)
	if err != nil {
		return nil, resp, err
	}
	keys := make([]*DNSSECKey, len(result))
	for i, k := range result {
		key := DNSSECKeyFromSchema(k)
		keys[i] = &key
	}
	return keys, resp, nil
}

// GetKey returns a single DNSSEC key of the zone.
func (d *DNSSECClient) GetKey(ctx context.Context, zoneID, keyID int) (*DNSSECKey, *Response, error) {
	var result schema.DNSSECKey
	resp, err := d.client.get(ctx, fmt.Sprintf("zones/%d/dnssec/keys/%d", zoneID, keyID), &result)
	if err != nil {
		return nil, resp, err
	}
	key := DNSSECKeyFromSchema(result)
	return &key, resp, nil
}

// AddKey creates a new DNSSEC key for the zone and returns it.
//
// PowerDNS creates keys inactive; use [DNSSECClient.SetKeyActive] to activate
// the new key.
func (d *DNSSECClient) AddKey(ctx context.Context, zoneID int, opts DNSSECKeyCreateOpts) (*DNSSECKey, *Response, error) {
	req := schema.DNSSECKeyAddRequest{
		Type:      string(opts.Type),
		Algorithm: opts.Algorithm,
		Bits:      opts.Bits,
	}
	var result schema.DNSSECKey
	resp, err := d.client.post(ctx, fmt.Sprintf("zones/%d/dnssec/keys", zoneID), req, &result)
	if err != nil {
		return nil, resp, err
	}
	key := DNSSECKeyFromSchema(result)
	return &key, resp, nil
}

// SetKeyActive activates or deactivates a DNSSEC key and returns its new state.
// Setting the state the key already has is not an error.
func (d *DNSSECClient) SetKeyActive(ctx context.Context, zoneID, keyID int, active bool) (*DNSSECKey, *Response, error) {
	req := schema.DNSSECKeyUpdateRequest{Active: active}
	var result schema.DNSSECKey
	resp, err := d.client.patch(ctx, fmt.Sprintf("zones/%d/dnssec/keys/%d", zoneID, keyID), req, &result)
	if err != nil {
		return nil, resp, err
	}
	key := DNSSECKeyFromSchema(result)
	return &key, resp, nil
}

// DeleteKey deletes a DNSSEC key of the zone.
func (d *DNSSECClient) DeleteKey(ctx context.Context, zoneID, keyID int) (*Response, error) {
	return d.client.delete(ctx, fmt.Sprintf("zones/%d/dnssec/keys/%d", zoneID, keyID))
}

// Rectify recalculates the DNSSEC ordering and auth fields of a signed zone.
// The server answers 409 for unsigned, presigned and secondary zones.
func (d *DNSSECClient) Rectify(ctx context.Context, zoneID int) (*Response, error) {
	return d.client.post(ctx, fmt.Sprintf("zones/%d/dnssec/rectify", zoneID), nil, nil)
}
