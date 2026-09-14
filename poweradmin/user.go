// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"

	"github.com/contentways/poweradmin-go/v3/poweradmin/schema"
)

// User represents a Poweradmin user.
type User struct {
	ID          int
	Username    string
	Fullname    string
	Email       string
	Description string
	Active      bool
	PermTempl   int
	UseLdap     bool
	IsAdmin     bool
	Permissions []string
	ZoneCount   int
	CreatedAt   string
	UpdatedAt   string
}

// UserCreateOpts configures a user creation request.
type UserCreateOpts struct {
	Username    string
	Password    string
	Fullname    string
	Email       string
	Description string
	Active      bool
	PermTempl   int
	UseLdap     bool
}

// UserUpdateOpts configures a user update request.
type UserUpdateOpts struct {
	Username    string
	Password    string
	Fullname    string
	Email       string
	Description string
	Active      *bool
	PermTempl   int
	UseLdap     *bool
}

// UserClient provides access to the user-related Poweradmin API endpoints.
type UserClient struct {
	client *Client
}

// GetByName returns a single [User] by username.
// Paginates the list endpoint server-side and matches client-side; the API
// has no dedicated username lookup.
func (u *UserClient) GetByName(ctx context.Context, username string) (*User, *Response, error) {
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		users, resp, err := u.List(ctx, opts)
		if err != nil {
			return nil, resp, err
		}
		for _, user := range users {
			if user.Username == username {
				return user, resp, nil
			}
		}
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.Page >= resp.Meta.Pagination.LastPage {
			return nil, resp, &APIError{StatusCode: 404, Message: fmt.Sprintf("user not found: %s", username)}
		}
		opts.Page++
	}
}

// GetByID returns a single [User] by ID.
func (u *UserClient) GetByID(ctx context.Context, id int) (*User, *Response, error) {
	var result schema.UserResponse
	resp, err := u.client.get(ctx, fmt.Sprintf("users/%d", id), &result)
	if err != nil {
		return nil, resp, err
	}
	user := UserFromSchema(result.User)
	return &user, resp, nil
}

// List returns one page of [User]s.
func (u *UserClient) List(ctx context.Context, opts ListOpts) ([]*User, *Response, error) {
	path := appendQuery("users", opts.values())
	var result schema.UserListResponse
	resp, err := u.client.get(ctx, path, &result)
	if err != nil {
		return nil, resp, err
	}
	users := make([]*User, len(result.Users))
	for i, s := range result.Users {
		user := UserFromSchema(s)
		users[i] = &user
	}
	return users, resp, nil
}

// All returns all [User]s across all pages.
func (u *UserClient) All(ctx context.Context) ([]*User, error) {
	var all []*User
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		users, resp, err := u.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, users...)
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.Page >= resp.Meta.Pagination.LastPage {
			return all, nil
		}
		opts.Page++
	}
}

// Create creates a new [User] and returns the new ID.
func (u *UserClient) Create(ctx context.Context, opts UserCreateOpts) (int, *Response, error) {
	req := schema.UserCreateRequest{
		Username:    opts.Username,
		Password:    opts.Password,
		Fullname:    opts.Fullname,
		Email:       opts.Email,
		Description: opts.Description,
		Active:      opts.Active,
		PermTempl:   opts.PermTempl,
		UseLdap:     opts.UseLdap,
	}
	var result schema.UserCreateResponse
	resp, err := u.client.post(ctx, "users", req, &result)
	if err != nil {
		return 0, resp, err
	}
	return result.UserID, resp, nil
}

// Update updates an existing [User] and returns the updated state.
func (u *UserClient) Update(ctx context.Context, id int, opts UserUpdateOpts) (*User, *Response, error) {
	req := schema.UserUpdateRequest{
		Username:    opts.Username,
		Password:    opts.Password,
		Fullname:    opts.Fullname,
		Email:       opts.Email,
		Description: opts.Description,
		Active:      opts.Active,
		PermTempl:   opts.PermTempl,
		UseLdap:     opts.UseLdap,
	}
	var result schema.UserResponse
	resp, err := u.client.put(ctx, fmt.Sprintf("users/%d", id), req, &result)
	if err != nil {
		return nil, resp, err
	}
	user := UserFromSchema(result.User)
	return &user, resp, nil
}

// Delete deletes the [User] with the given ID.
func (u *UserClient) Delete(ctx context.Context, id int) (*Response, error) {
	return u.client.delete(ctx, fmt.Sprintf("users/%d", id))
}

// SetPermissionTemplate assigns a permission template to a user via PATCH.
// This is a partial update; other user fields are untouched.
func (u *UserClient) SetPermissionTemplate(ctx context.Context, id, permTemplID int) (*Response, error) {
	req := schema.UserPatchRequest{PermTempl: permTemplID}
	return u.client.patch(ctx, fmt.Sprintf("users/%d", id), req, nil)
}
