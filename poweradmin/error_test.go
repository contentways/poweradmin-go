// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			name: "without details",
			err:  &APIError{StatusCode: 404, Message: "not found"},
			want: "poweradmin: HTTP 404: not found",
		},
		{
			name: "with details",
			err:  &APIError{StatusCode: 422, Message: "invalid", Details: "bad name"},
			want: "poweradmin: HTTP 422: invalid (bad name)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain 404", &APIError{StatusCode: 404}, true},
		{"plain 500", &APIError{StatusCode: 500}, false},
		{"wrapped 404", fmt.Errorf("wrap: %w", &APIError{StatusCode: 404}), true},
		{"non-api error", errors.New("oops"), false},
		{"string fallback HTTP 404", errors.New("got HTTP 404 oops"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantMessage string
		wantDetails string
	}{
		{
			name:        "poweradmin v2 error envelope",
			body:        `{"success":false,"message":"Zone already exists","data":null}`,
			wantMessage: "Zone already exists",
		},
		{
			name:        "error object takes precedence",
			body:        `{"success":false,"message":"Request failed","error":{"message":"Invalid name","details":"label too long"}}`,
			wantMessage: "Invalid name",
			wantDetails: "label too long",
		},
		{
			name:        "non-JSON body from proxy",
			body:        "<html>502 Bad Gateway</html>\n",
			wantMessage: "<html>502 Bad Gateway</html>",
		},
		{
			name:        "envelope without any message",
			body:        `{"success":false}`,
			wantMessage: `{"success":false}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newAPIError(409, []byte(tt.body))
			if got.StatusCode != 409 {
				t.Errorf("StatusCode = %d, want 409", got.StatusCode)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", got.Message, tt.wantMessage)
			}
			if got.Details != tt.wantDetails {
				t.Errorf("Details = %q, want %q", got.Details, tt.wantDetails)
			}
		})
	}
}
