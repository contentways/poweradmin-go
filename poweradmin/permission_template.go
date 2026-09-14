// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v3/poweradmin/schema"
)

// PermissionTemplate represents a named bundle of permissions that can be
// assigned to a user or group.
type PermissionTemplate struct {
	ID           int
	Name         string
	Descr        string
	TemplateType string // "user" or "group"
	Permissions  []Permission
}

// PermissionTemplateOpts configures a create or update request.
type PermissionTemplateOpts struct {
	Name         string
	Descr        string
	TemplateType string // "user" (default) or "group"
	Permissions  []int  // permission IDs to assign
}

// PermissionTemplateClient provides access to the permission-template endpoints.
type PermissionTemplateClient struct {
	client *Client
}

// GetByName returns a single [PermissionTemplate] by name.
// Falls back to a list scan since there is no dedicated by-name endpoint.
// Note: list responses do not include the full permission list — call
// [PermissionTemplateClient.GetByID] on the result to get permissions.
func (c *PermissionTemplateClient) GetByName(ctx context.Context, name string) (*PermissionTemplate, *Response, error) {
	templates, resp, err := c.List(ctx)
	if err != nil {
		return nil, resp, err
	}
	for _, tpl := range templates {
		if tpl.Name == name {
			return tpl, resp, nil
		}
	}
	return nil, resp, &APIError{StatusCode: 404, Message: fmt.Sprintf("permission template not found: %s", name)}
}

// GetByID returns a single [PermissionTemplate] with its full permission list.
func (c *PermissionTemplateClient) GetByID(ctx context.Context, id int) (*PermissionTemplate, *Response, error) {
	var result schema.PermissionTemplateResponse
	resp, err := c.client.get(ctx, fmt.Sprintf("permission-templates/%d", id), &result)
	if err != nil {
		return nil, resp, err
	}
	tpl := permissionTemplateFromSchema(result.Template)
	return &tpl, resp, nil
}

// List returns all [PermissionTemplate]s. The permission list is typically not
// populated in list responses — use [PermissionTemplateClient.GetByID] for that.
func (c *PermissionTemplateClient) List(ctx context.Context) ([]*PermissionTemplate, *Response, error) {
	var result schema.PermissionTemplateListResponse
	resp, err := c.client.get(ctx, "permission-templates", &result)
	if err != nil {
		return nil, resp, err
	}
	templates := make([]*PermissionTemplate, len(result.Templates))
	for i, s := range result.Templates {
		tpl := permissionTemplateFromSchema(s)
		templates[i] = &tpl
	}
	return templates, resp, nil
}

// Create creates a new [PermissionTemplate].
func (c *PermissionTemplateClient) Create(ctx context.Context, opts PermissionTemplateOpts) (*PermissionTemplate, *Response, error) {
	req := schema.PermissionTemplateRequest{
		Name:         opts.Name,
		Descr:        opts.Descr,
		TemplateType: opts.TemplateType,
		Permissions:  opts.Permissions,
	}
	// The create endpoint returns an empty body, so look the new template up by
	// name. That list lookup omits the permission list — call GetByID on the
	// result if you need it.
	resp, err := c.client.post(ctx, "permission-templates", req, nil)
	if err != nil {
		return nil, resp, err
	}
	tpl, _, err := c.GetByName(ctx, opts.Name)
	if err != nil {
		return nil, resp, err
	}
	return tpl, resp, nil
}

// Update updates an existing [PermissionTemplate].
func (c *PermissionTemplateClient) Update(ctx context.Context, id int, opts PermissionTemplateOpts) (*PermissionTemplate, *Response, error) {
	req := schema.PermissionTemplateRequest{
		Name:         opts.Name,
		Descr:        opts.Descr,
		TemplateType: opts.TemplateType,
		Permissions:  opts.Permissions,
	}
	// The update endpoint returns an empty body; read the updated template back
	// by ID so callers get the persisted state, including its permissions.
	resp, err := c.client.put(ctx, fmt.Sprintf("permission-templates/%d", id), req, nil)
	if err != nil {
		return nil, resp, err
	}
	tpl, _, err := c.GetByID(ctx, id)
	if err != nil {
		return nil, resp, err
	}
	return tpl, resp, nil
}

// Delete deletes a [PermissionTemplate].
func (c *PermissionTemplateClient) Delete(ctx context.Context, id int) (*Response, error) {
	return c.client.delete(ctx, fmt.Sprintf("permission-templates/%d", id))
}

func permissionTemplateFromSchema(s schema.PermissionTemplate) PermissionTemplate {
	perms := make([]Permission, len(s.Permissions))
	for i, p := range s.Permissions {
		perms[i] = PermissionFromSchema(p)
	}
	return PermissionTemplate{
		ID:           s.ID,
		Name:         s.Name,
		Descr:        s.Descr,
		TemplateType: s.TemplateType,
		Permissions:  perms,
	}
}
