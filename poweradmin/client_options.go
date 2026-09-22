// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"
)

// Option configures a [Client].
type Option func(*Client)

// WithBaseURL sets the base URL of the Poweradmin instance.
// Trailing slashes are stripped.
//
//	poweradmin.WithBaseURL("https://dns.example.com")
func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(u, "/")
	}
}

// WithAPIKey configures Bearer token / X-API-Key authentication.
// This is the recommended authentication method.
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithBasicAuth configures HTTP basic authentication.
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// WithAPIVersion overrides the API version prefix (default: "v2").
func WithAPIVersion(version string) Option {
	return func(c *Client) {
		c.apiVersion = version
	}
}

// WithHTTPClient replaces the default [http.Client].
// Use this to inject a custom transport, proxy, or test double.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithTimeout sets the HTTP client timeout (default: 30s).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithRetry enables automatic retry on transient failures. maxAttempts is the
// total number of attempts including the initial one — values < 2 disable
// retrying. Backoff defaults to [DefaultBackoff]; override via
// [WithRetryBackoff].
//
// HTTP 429 is retried for every method. Network errors and 5xx responses are
// only retried for idempotent methods (GET, PUT, DELETE), because a failed
// POST or PATCH may already have been applied by the server.
func WithRetry(maxAttempts int) Option {
	return func(c *Client) {
		if maxAttempts < 2 {
			c.retry = nil
			return
		}
		if c.retry == nil {
			c.retry = &retryConfig{backoff: DefaultBackoff}
		}
		c.retry.maxAttempts = maxAttempts
	}
}

// WithRetryBackoff sets a custom backoff strategy. Only takes effect if
// [WithRetry] has also been called.
func WithRetryBackoff(fn Backoff) Option {
	return func(c *Client) {
		if c.retry == nil {
			c.retry = &retryConfig{maxAttempts: 3}
		}
		c.retry.backoff = fn
	}
}

// WithDebugWriter enables HTTP request/response logging to w. One line is
// written per request with method, path, status code, and duration.
// Suitable destinations: os.Stderr, a *log.Logger via log.Writer(), or any
// io.Writer your application wires up (e.g. slog backed by a buffer).
func WithDebugWriter(w io.Writer) Option {
	return func(c *Client) {
		c.debugWriter = w
	}
}

// WithInsecure disables TLS certificate verification.
// Do not use in production.
// Clones the existing transport if one is set so other settings are preserved.
func WithInsecure() Option {
	return func(c *Client) {
		base, ok := c.httpClient.Transport.(*http.Transport)
		var t *http.Transport
		if ok && base != nil {
			t = base.Clone()
		} else {
			t = http.DefaultTransport.(*http.Transport).Clone()
		}
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{} //nolint:gosec
		}
		t.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec
		c.httpClient.Transport = t
	}
}
