// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
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

	// PermTemplName is the name of PermTempl (Poweradmin 4.5+).
	PermTemplName string
	// Groups lists the user's group memberships (Poweradmin 4.5+).
	Groups []UserGroup
}

// UserGroup is a group a [User] belongs to.
type UserGroup struct {
	ID   int
	Name string
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
	// GroupIDs and GroupNames add the new user to these groups (Poweradmin
	// 4.5+). Names must match exactly, including case. Assigning groups needs
	// the user_is_ueberuser permission; without it the whole request fails
	// with 403. At most 50 groups in total.
	GroupIDs   []int
	GroupNames []string
}

// UserCreateResult is the outcome of [UserClient.CreateWithResult].
type UserCreateResult struct {
	UserID int
	// Groups lists the memberships that were actually created. The user is
	// created even if adding it to a group fails, so this can be shorter
	// than the requested groups.
	Groups []UserGroup
}

// UserUpdateOpts configures a user update request.
// Only non-nil fields are sent; omitted fields keep their current value.
type UserUpdateOpts struct {
	Username    *string
	Password    *string
	Fullname    *string
	Email       *string
	Description *string
	Active      *bool
	PermTempl   *int
	UseLdap     *bool
}

// UserDeleteOpts configures a user deletion.
type UserDeleteOpts struct {
	// TransferToUserID receives the zones owned by the deleted user. The API
	// rejects the deletion with 400 if the user owns zones and this is nil.
	TransferToUserID *int
}

// UserClient provides access to the user-related Poweradmin API endpoints.
type UserClient struct {
	client *Client
}

// GetByName returns a single [User] by username.
//
// It uses the server-side exact-match filter (?username=) of the list
// endpoint and still compares usernames client-side, so it also works
// against servers that ignore the filter; there it falls back to paging
// through all users.
func (u *UserClient) GetByName(ctx context.Context, username string) (*User, *Response, error) {
	opts := ListOpts{Page: 1, PerPage: 100}
	for {
		q := opts.values()
		q.Set("username", username)
		users, resp, err := u.list(ctx, q)
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
	return u.list(ctx, opts.values())
}

func (u *UserClient) list(ctx context.Context, query url.Values) ([]*User, *Response, error) {
	path := appendQuery("users", query)
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
	result, resp, err := u.CreateWithResult(ctx, opts)
	if err != nil {
		return 0, resp, err
	}
	return result.UserID, resp, nil
}

// CreateWithResult creates a new [User] like [UserClient.Create] and also
// returns the group memberships that were created.
func (u *UserClient) CreateWithResult(ctx context.Context, opts UserCreateOpts) (*UserCreateResult, *Response, error) {
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
	for _, id := range opts.GroupIDs {
		req.Groups = append(req.Groups, id)
	}
	for _, name := range opts.GroupNames {
		req.Groups = append(req.Groups, name)
	}
	var result schema.UserCreateResponse
	resp, err := u.client.post(ctx, "users", req, &result)
	if err != nil {
		return nil, resp, err
	}
	return &UserCreateResult{
		UserID: result.UserID,
		Groups: userGroupsFromSchema(result.Groups),
	}, resp, nil
}

// Update updates an existing [User] and returns the updated state.
//
// The update endpoint only returns the user ID, so Update reads the user back
// with an additional GET to return the persisted state.
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
	resp, err := u.client.put(ctx, fmt.Sprintf("users/%d", id), req, nil)
	if err != nil {
		return nil, resp, err
	}
	user, _, err := u.GetByID(ctx, id)
	if err != nil {
		return nil, resp, err
	}
	return user, resp, nil
}

// Delete deletes the [User] with the given ID and returns the number of zones
// that were transferred to [UserDeleteOpts.TransferToUserID].
func (u *UserClient) Delete(ctx context.Context, id int, opts UserDeleteOpts) (int, *Response, error) {
	var body any
	if opts.TransferToUserID != nil {
		body = schema.UserDeleteRequest{TransferToUserID: opts.TransferToUserID}
	}
	var result schema.UserDeleteResponse
	resp, err := u.client.deleteWithBody(ctx, fmt.Sprintf("users/%d", id), body, &result)
	if err != nil {
		return 0, resp, err
	}
	return result.ZonesAffected, resp, nil
}

// SetPermissionTemplate assigns a permission template to a user via PATCH.
// This is a partial update; other user fields are untouched.
func (u *UserClient) SetPermissionTemplate(ctx context.Context, id, permTemplID int) (*Response, error) {
	req := schema.UserPatchRequest{PermTempl: permTemplID}
	return u.client.patch(ctx, fmt.Sprintf("users/%d", id), req, nil)
}
