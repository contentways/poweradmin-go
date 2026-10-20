// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

const (
	defaultAPIVersion = "v2"
	defaultTimeout    = 30 * time.Second
	userAgent         = "poweradmin-go/" + Version
)

// Client is the Poweradmin API client.
// Create one via [NewClient].
type Client struct {
	baseURL     string
	apiVersion  string
	apiKey      string
	username    string
	password    string
	httpClient  *http.Client
	retry       *retryConfig
	debugWriter io.Writer

	// Per-resource clients
	Zone               IZoneClient
	DNSSEC             IDNSSECClient
	ZoneTemplate       IZoneTemplateClient
	Record             IRecordClient
	RRSet              IRRSetClient
	User               IUserClient
	Group              IGroupClient
	Permission         IPermissionClient
	PermissionTemplate IPermissionTemplateClient
	Server             IServerClient
}

// NewClient creates a new Poweradmin API client.
// At minimum, [WithBaseURL] and either [WithAPIKey] or [WithBasicAuth] must be provided.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		apiVersion: defaultAPIVersion,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.baseURL == "" {
		return nil, fmt.Errorf("poweradmin: base URL is required (use WithBaseURL)")
	}
	if _, err := url.Parse(c.baseURL); err != nil {
		return nil, fmt.Errorf("poweradmin: invalid base URL: %w", err)
	}
	if c.apiKey == "" && c.username == "" {
		return nil, fmt.Errorf("poweradmin: authentication is required (use WithAPIKey or WithBasicAuth)")
	}

	c.Zone = &ZoneClient{client: c}
	c.ZoneTemplate = &ZoneTemplateClient{client: c}
	c.Record = &RecordClient{client: c}
	c.RRSet = &RRSetClient{client: c}
	c.User = &UserClient{client: c}
	c.Group = &GroupClient{client: c}
	c.Permission = &PermissionClient{client: c}
	c.PermissionTemplate = &PermissionTemplateClient{client: c}
	c.DNSSEC = &DNSSECClient{client: c}
	c.Server = &ServerClient{client: c}

	return c, nil
}

// buildURL constructs the full URL for an API endpoint path.
func (c *Client) buildURL(path string) string {
	path = strings.TrimLeft(path, "/")
	return fmt.Sprintf("%s/api/%s/%s", c.baseURL, c.apiVersion, path)
}

// do executes an HTTP request and returns the raw [http.Response] wrapped in a
// [Response]. The caller is responsible for closing the body — typically via
// the [Client.parse] helper. When retry is configured, transient failures
// (network errors, HTTP 429, 5xx) trigger replays with backoff up to the
// configured attempt limit.
func (c *Client) do(ctx context.Context, method, path string, body any) (*Response, error) {
	var reqBody []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("poweradmin: marshal request body: %w", err)
		}
		reqBody = b
	}

	reqURL := c.buildURL(path)
	maxAttempts := 1
	if c.retry != nil {
		maxAttempts = c.retry.maxAttempts
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.retry.backoff(attempt - 1)
			c.debugf("retry attempt %d after %s", attempt, delay)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		var rdr io.Reader
		if reqBody != nil {
			rdr = bytes.NewReader(reqBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, reqURL, rdr)
		if err != nil {
			return nil, fmt.Errorf("poweradmin: create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent)

		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		} else {
			req.SetBasicAuth(c.username, c.password)
		}

		start := time.Now()
		httpResp, err := c.httpClient.Do(req)
		dur := time.Since(start)

		if err != nil {
			c.debugf("%s %s -> ERR %v (%s)", method, path, err, dur)
			// Context cancellation: don't retry.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			if !shouldRetryError(method) {
				return nil, err
			}
			lastErr = err
			continue
		}

		c.debugf("%s %s -> %d (%s)", method, path, httpResp.StatusCode, dur)

		if shouldRetryStatus(method, httpResp.StatusCode) && attempt+1 < maxAttempts {
			if err := httpResp.Body.Close(); err != nil {
				c.debugf("close response body: %v", err)
			}
			lastErr = fmt.Errorf("poweradmin: HTTP %d", httpResp.StatusCode)
			continue
		}
		return &Response{Response: httpResp}, nil
	}
	return nil, lastErr
}

func (c *Client) debugf(format string, args ...any) {
	if c.debugWriter == nil {
		return
	}
	_, _ = fmt.Fprintf(c.debugWriter, "[poweradmin] "+format+"\n", args...)
}

// parse reads the response body, validates the envelope, and unmarshals the
// inner data into result. A nil result is valid for operations that return no
// data (e.g. DELETE). Envelope-level pagination is attached to
// resp.Meta.Pagination; error responses become an [APIError].
func (c *Client) parse(resp *Response, result any) error {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.debugf("close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("poweradmin: read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return newAPIError(resp.StatusCode, body)
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("poweradmin: parse API response: %w", err)
	}
	if !apiResp.Success {
		return newAPIError(resp.StatusCode, body)
	}

	if apiResp.Pagination != nil {
		resp.Meta.Pagination = &Pagination{
			Page:     apiResp.Pagination.CurrentPage,
			PerPage:  apiResp.Pagination.PerPage,
			Total:    apiResp.Pagination.Total,
			LastPage: apiResp.Pagination.LastPage,
		}
	}

	if result != nil && len(apiResp.Data) > 0 {
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("poweradmin: unmarshal response data: %w", err)
		}
	}

	return nil
}

func (c *Client) get(ctx context.Context, path string, result any) (*Response, error) {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if err := c.parse(resp, result); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) post(ctx context.Context, path string, body, result any) (*Response, error) {
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	if err := c.parse(resp, result); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) put(ctx context.Context, path string, body, result any) (*Response, error) {
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}
	if err := c.parse(resp, result); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) patch(ctx context.Context, path string, body, result any) (*Response, error) {
	resp, err := c.do(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}
	if err := c.parse(resp, result); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) delete(ctx context.Context, path string) (*Response, error) {
	return c.deleteWithBody(ctx, path, nil, nil)
}

// deleteWithBody sends a DELETE with an optional JSON body. A few endpoints
// (e.g. DELETE /v2/users/{id}) take options in the request body.
func (c *Client) deleteWithBody(ctx context.Context, path string, body, result any) (*Response, error) {
	resp, err := c.do(ctx, http.MethodDelete, path, body)
	if err != nil {
		return nil, err
	}
	if err := c.parse(resp, result); err != nil {
		return resp, err
	}
	return resp, nil
}
