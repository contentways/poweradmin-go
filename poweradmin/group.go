// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// Group represents a Poweradmin user group.
type Group struct {
	ID          int
	Name        string
	Description string
	PermTemplID int
	MemberCount int
	ZoneCount   int
	CreatedAt   string
}

// GroupMember represents a member of a group.
type GroupMember struct {
	UserID   int
	Username string
	JoinedAt string
}

// GroupZone represents a zone associated with a group.
type GroupZone struct {
	ZoneID   int
	ZoneName string
	ZoneType string
}

// GroupCreateOpts configures a group creation request.
type GroupCreateOpts struct {
	Name        string
	Description string
	PermTemplID int
}

// GroupUpdateOpts configures a group update request.
// Description is a *string so an empty description can be sent explicitly.
type GroupUpdateOpts struct {
	Name        string
	Description *string
}

// GroupClient provides access to the group-related Poweradmin API endpoints.
type GroupClient struct {
	client *Client
}

// GetByName returns a single [Group] by name.
// Paginates the list endpoint and matches client-side.
func (g *GroupClient) GetByName(ctx context.Context, name string) (*Group, *Response, error) {
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		groups, resp, err := g.List(ctx, opts)
		if err != nil {
			return nil, resp, err
		}
		for _, group := range groups {
			if group.Name == name {
				return group, resp, nil
			}
		}
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.Page >= resp.Meta.Pagination.LastPage {
			return nil, resp, &APIError{StatusCode: 404, Message: fmt.Sprintf("group not found: %s", name)}
		}
		opts.Page++
	}
}

// GetByID returns a single [Group] by ID.
func (g *GroupClient) GetByID(ctx context.Context, id int) (*Group, *Response, error) {
	var result schema.GroupResponse
	resp, err := g.client.get(ctx, fmt.Sprintf("groups/%d", id), &result)
	if err != nil {
		return nil, resp, err
	}
	group := GroupFromSchema(result.Group)
	return &group, resp, nil
}

// List returns one page of [Group]s.
func (g *GroupClient) List(ctx context.Context, opts ListOpts) ([]*Group, *Response, error) {
	path := appendQuery("groups", opts.values())
	var result schema.GroupListResponse
	resp, err := g.client.get(ctx, path, &result)
	if err != nil {
		return nil, resp, err
	}
	groups := make([]*Group, len(result.Groups))
	for i, s := range result.Groups {
		group := GroupFromSchema(s)
		groups[i] = &group
	}
	return groups, resp, nil
}

// All returns all [Group]s across all pages.
func (g *GroupClient) All(ctx context.Context) ([]*Group, error) {
	var all []*Group
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		groups, resp, err := g.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, groups...)
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.Page >= resp.Meta.Pagination.LastPage {
			return all, nil
		}
		opts.Page++
	}
}

// Create creates a new [Group] and returns the new ID.
func (g *GroupClient) Create(ctx context.Context, opts GroupCreateOpts) (int, *Response, error) {
	req := schema.GroupCreateRequest{
		Name:        opts.Name,
		Description: opts.Description,
		PermTemplID: opts.PermTemplID,
	}
	var result schema.GroupCreateResponse
	resp, err := g.client.post(ctx, "groups", req, &result)
	if err != nil {
		return 0, resp, err
	}
	return result.Group.ID, resp, nil
}

// Update updates an existing [Group] and returns the updated state.
func (g *GroupClient) Update(ctx context.Context, id int, opts GroupUpdateOpts) (*Group, *Response, error) {
	req := schema.GroupUpdateRequest{
		Name:        opts.Name,
		Description: opts.Description,
	}
	var result schema.GroupResponse
	resp, err := g.client.put(ctx, fmt.Sprintf("groups/%d", id), req, &result)
	if err != nil {
		return nil, resp, err
	}
	group := GroupFromSchema(result.Group)
	return &group, resp, nil
}

// Delete deletes the [Group] with the given ID.
func (g *GroupClient) Delete(ctx context.Context, id int) (*Response, error) {
	return g.client.delete(ctx, fmt.Sprintf("groups/%d", id))
}

// Members returns the [GroupMember]s of a group.
func (g *GroupClient) Members(ctx context.Context, id int) ([]*GroupMember, *Response, error) {
	var result schema.GroupMemberListResponse
	resp, err := g.client.get(ctx, fmt.Sprintf("groups/%d/members", id), &result)
	if err != nil {
		return nil, resp, err
	}
	members := make([]*GroupMember, len(result.Members))
	for i, s := range result.Members {
		m := GroupMemberFromSchema(s)
		members[i] = &m
	}
	return members, resp, nil
}

// AddMember adds a user to a group.
func (g *GroupClient) AddMember(ctx context.Context, groupID, userID int) (*Response, error) {
	req := schema.GroupMemberRequest{UserID: userID}
	return g.client.post(ctx, fmt.Sprintf("groups/%d/members", groupID), req, nil)
}

// RemoveMember removes a user from a group.
func (g *GroupClient) RemoveMember(ctx context.Context, groupID, userID int) (*Response, error) {
	return g.client.delete(ctx, fmt.Sprintf("groups/%d/members/%d", groupID, userID))
}

// Zones returns the zones associated with a group.
func (g *GroupClient) Zones(ctx context.Context, id int) ([]*GroupZone, *Response, error) {
	var result schema.GroupZoneListResponse
	resp, err := g.client.get(ctx, fmt.Sprintf("groups/%d/zones", id), &result)
	if err != nil {
		return nil, resp, err
	}
	zones := make([]*GroupZone, len(result.Zones))
	for i, s := range result.Zones {
		z := GroupZoneFromSchema(s)
		zones[i] = &z
	}
	return zones, resp, nil
}

// AddZone associates a zone with a group.
func (g *GroupClient) AddZone(ctx context.Context, groupID, zoneID int) (*Response, error) {
	req := schema.GroupZoneRequest{ZoneID: zoneID}
	return g.client.post(ctx, fmt.Sprintf("groups/%d/zones", groupID), req, nil)
}

// RemoveZone disassociates a zone from a group.
func (g *GroupClient) RemoveZone(ctx context.Context, groupID, zoneID int) (*Response, error) {
	return g.client.delete(ctx, fmt.Sprintf("groups/%d/zones/%d", groupID, zoneID))
}
