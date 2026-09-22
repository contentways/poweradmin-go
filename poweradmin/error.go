// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// APIError represents an error returned by the Poweradmin API.
type APIError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("poweradmin: HTTP %d: %s (%s)", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("poweradmin: HTTP %d: %s", e.StatusCode, e.Message)
}

// IsNotFound reports whether err is a 404 Not Found API error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.StatusCode == 404
	}
	// Fallback for string-wrapped errors.
	return strings.Contains(err.Error(), "HTTP 404")
}

// newAPIError builds an [APIError] from an error response body.
//
// Poweradmin's v2 API reports errors as {"success": false, "message": "...",
// "data": null} without an "error" object. An "error" object is still honoured
// when present. If the body is not a JSON envelope at all (e.g. an HTML error
// page from a reverse proxy), the raw body is used as the message.
func newAPIError(status int, body []byte) *APIError {
	var env schema.APIResponse
	if err := json.Unmarshal(body, &env); err != nil {
		return &APIError{StatusCode: status, Message: strings.TrimSpace(string(body))}
	}
	if env.Error != nil && env.Error.Message != "" {
		return &APIError{StatusCode: status, Message: env.Error.Message, Details: env.Error.Details}
	}
	if env.Message != "" {
		return &APIError{StatusCode: status, Message: env.Message}
	}
	return &APIError{StatusCode: status, Message: strings.TrimSpace(string(body))}
}
