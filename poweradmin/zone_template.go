// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// ZoneTemplate represents a Poweradmin zone template.
type ZoneTemplate struct {
	ID          int
	Name        string
	Description string
	Owner       int
	IsGlobal    bool
	ZonesLinked int
}

// ZoneTemplateRecord is a record entry within a [ZoneTemplate]. Placeholders
// like [ZONE], [NS1], [SERIAL] are expanded at zone creation time.
type ZoneTemplateRecord struct {
	ID       int
	Name     string
	Type     string
	Content  string
	TTL      int
	Priority int
}

// ZoneTemplateCreateOpts configures a template creation request.
type ZoneTemplateCreateOpts struct {
	Name        string
	Description string
	IsGlobal    bool
}

// ZoneTemplateUpdateOpts configures a template update.
type ZoneTemplateUpdateOpts struct {
	Name        string
	Description string
	IsGlobal    bool
}

// ZoneTemplateRecordOpts configures a template-record create/update.
type ZoneTemplateRecordOpts struct {
	Name     string
	Type     string
	Content  string
	TTL      int
	Priority int
}

// ZoneTemplateClient provides access to the zone-template endpoints.
type ZoneTemplateClient struct {
	client *Client
}

// GetByName returns a single [ZoneTemplate] by name.
// The template endpoint is not paginated, so this is a single-page scan.
func (c *ZoneTemplateClient) GetByName(ctx context.Context, name string) (*ZoneTemplate, *Response, error) {
	templates, resp, err := c.List(ctx)
	if err != nil {
		return nil, resp, err
	}
	for _, tpl := range templates {
		if tpl.Name == name {
			return tpl, resp, nil
		}
	}
	return nil, resp, &APIError{StatusCode: 404, Message: fmt.Sprintf("zone template not found: %s", name)}
}

// GetByID returns a single [ZoneTemplate].
func (c *ZoneTemplateClient) GetByID(ctx context.Context, id int) (*ZoneTemplate, *Response, error) {
	var result schema.ZoneTemplateResponse
	resp, err := c.client.get(ctx, fmt.Sprintf("zone-templates/%d", id), &result)
	if err != nil {
		return nil, resp, err
	}
	tpl := ZoneTemplateFromSchema(result.Template)
	return &tpl, resp, nil
}

// List returns all [ZoneTemplate]s.
func (c *ZoneTemplateClient) List(ctx context.Context) ([]*ZoneTemplate, *Response, error) {
	var result schema.ZoneTemplateListResponse
	resp, err := c.client.get(ctx, "zone-templates", &result)
	if err != nil {
		return nil, resp, err
	}
	templates := make([]*ZoneTemplate, len(result.Templates))
	for i, s := range result.Templates {
		tpl := ZoneTemplateFromSchema(s)
		templates[i] = &tpl
	}
	return templates, resp, nil
}

// Create creates a new [ZoneTemplate].
func (c *ZoneTemplateClient) Create(ctx context.Context, opts ZoneTemplateCreateOpts) (*ZoneTemplate, *Response, error) {
	req := schema.ZoneTemplateRequest{
		Name:        opts.Name,
		Description: opts.Description,
		IsGlobal:    opts.IsGlobal,
	}
	// The create endpoint returns only the new ID; the rest of the template is
	// the input echoed back.
	var result schema.ZoneTemplateCreateResponse
	resp, err := c.client.post(ctx, "zone-templates", req, &result)
	if err != nil {
		return nil, resp, err
	}
	tpl := ZoneTemplate{
		ID:          result.ID,
		Name:        opts.Name,
		Description: opts.Description,
		IsGlobal:    opts.IsGlobal,
	}
	return &tpl, resp, nil
}

// Update updates an existing [ZoneTemplate].
//
// The update endpoint returns no data, so Update reads the template back with
// an additional GET to return the persisted state.
func (c *ZoneTemplateClient) Update(ctx context.Context, id int, opts ZoneTemplateUpdateOpts) (*ZoneTemplate, *Response, error) {
	req := schema.ZoneTemplateRequest{
		Name:        opts.Name,
		Description: opts.Description,
		IsGlobal:    opts.IsGlobal,
	}
	resp, err := c.client.put(ctx, fmt.Sprintf("zone-templates/%d", id), req, nil)
	if err != nil {
		return nil, resp, err
	}
	tpl, _, err := c.GetByID(ctx, id)
	if err != nil {
		return nil, resp, err
	}
	return tpl, resp, nil
}

// Delete deletes a [ZoneTemplate].
func (c *ZoneTemplateClient) Delete(ctx context.Context, id int) (*Response, error) {
	return c.client.delete(ctx, fmt.Sprintf("zone-templates/%d", id))
}

// Records returns all records belonging to the template.
func (c *ZoneTemplateClient) Records(ctx context.Context, templateID int) ([]*ZoneTemplateRecord, *Response, error) {
	var result schema.ZoneTemplateRecordListResponse
	resp, err := c.client.get(ctx, fmt.Sprintf("zone-templates/%d/records", templateID), &result)
	if err != nil {
		return nil, resp, err
	}
	records := make([]*ZoneTemplateRecord, len(result.Records))
	for i, s := range result.Records {
		rec := ZoneTemplateRecordFromSchema(s)
		records[i] = &rec
	}
	return records, resp, nil
}

// GetRecord returns a single template record.
func (c *ZoneTemplateClient) GetRecord(ctx context.Context, templateID, recordID int) (*ZoneTemplateRecord, *Response, error) {
	var result schema.ZoneTemplateRecordResponse
	resp, err := c.client.get(ctx, fmt.Sprintf("zone-templates/%d/records/%d", templateID, recordID), &result)
	if err != nil {
		return nil, resp, err
	}
	rec := ZoneTemplateRecordFromSchema(result.Record)
	return &rec, resp, nil
}

// CreateRecord adds a record to a template.
func (c *ZoneTemplateClient) CreateRecord(ctx context.Context, templateID int, opts ZoneTemplateRecordOpts) (*ZoneTemplateRecord, *Response, error) {
	req := schema.ZoneTemplateRecordRequest{
		Name:     opts.Name,
		Type:     opts.Type,
		Content:  opts.Content,
		TTL:      opts.TTL,
		Priority: opts.Priority,
	}
	// The create endpoint returns only the new ID; echo back the input fields.
	var result schema.ZoneTemplateRecordCreateResponse
	resp, err := c.client.post(ctx, fmt.Sprintf("zone-templates/%d/records", templateID), req, &result)
	if err != nil {
		return nil, resp, err
	}
	rec := ZoneTemplateRecord{
		ID:       result.ID,
		Name:     opts.Name,
		Type:     opts.Type,
		Content:  opts.Content,
		TTL:      opts.TTL,
		Priority: opts.Priority,
	}
	return &rec, resp, nil
}

// UpdateRecord replaces a template record.
//
// The update endpoint returns no data, so UpdateRecord reads the record back
// with an additional GET to return the persisted state.
func (c *ZoneTemplateClient) UpdateRecord(ctx context.Context, templateID, recordID int, opts ZoneTemplateRecordOpts) (*ZoneTemplateRecord, *Response, error) {
	req := schema.ZoneTemplateRecordRequest{
		Name:     opts.Name,
		Type:     opts.Type,
		Content:  opts.Content,
		TTL:      opts.TTL,
		Priority: opts.Priority,
	}
	resp, err := c.client.put(ctx, fmt.Sprintf("zone-templates/%d/records/%d", templateID, recordID), req, nil)
	if err != nil {
		return nil, resp, err
	}
	rec, _, err := c.GetRecord(ctx, templateID, recordID)
	if err != nil {
		return nil, resp, err
	}
	return rec, resp, nil
}

// DeleteRecord removes a record from a template.
func (c *ZoneTemplateClient) DeleteRecord(ctx context.Context, templateID, recordID int) (*Response, error) {
	return c.client.delete(ctx, fmt.Sprintf("zone-templates/%d/records/%d", templateID, recordID))
}
