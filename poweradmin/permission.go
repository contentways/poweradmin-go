// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// Permission represents a Poweradmin permission.
type Permission struct {
	ID    int
	Name  string
	Descr string
}

// PermissionClient provides access to the permission-related Poweradmin API endpoints.
type PermissionClient struct {
	client *Client
}

// GetByID returns a single [Permission] by ID.
func (p *PermissionClient) GetByID(ctx context.Context, id int) (*Permission, *Response, error) {
	var result schema.PermissionResponse
	resp, err := p.client.get(ctx, fmt.Sprintf("permissions/%d", id), &result)
	if err != nil {
		return nil, resp, err
	}
	perm := PermissionFromSchema(result.Permission)
	return &perm, resp, nil
}

// List returns one page of [Permission]s.
func (p *PermissionClient) List(ctx context.Context, opts ListOpts) ([]*Permission, *Response, error) {
	path := appendQuery("permissions", opts.values())
	var result schema.PermissionListResponse
	resp, err := p.client.get(ctx, path, &result)
	if err != nil {
		return nil, resp, err
	}
	perms := make([]*Permission, len(result.Permissions))
	for i, s := range result.Permissions {
		perm := PermissionFromSchema(s)
		perms[i] = &perm
	}
	return perms, resp, nil
}

// All returns all [Permission]s across all pages.
func (p *PermissionClient) All(ctx context.Context) ([]*Permission, error) {
	var all []*Permission
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		perms, resp, err := p.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, perms...)
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.Page >= resp.Meta.Pagination.LastPage {
			return all, nil
		}
		opts.Page++
	}
}
