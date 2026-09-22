// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"net/http"
	"testing"
	"time"
)

func TestWithBaseURLTrimsTrailingSlash(t *testing.T) {
	c, err := NewClient(WithBaseURL("https://x.example.com////"), WithAPIKey("k"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.baseURL != "https://x.example.com" {
		t.Errorf("baseURL = %q, want trailing slashes stripped", c.baseURL)
	}
}

func TestWithTimeout(t *testing.T) {
	c, err := NewClient(WithBaseURL("https://x"), WithAPIKey("k"), WithTimeout(7*time.Second))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.httpClient.Timeout != 7*time.Second {
		t.Errorf("Timeout = %v, want 7s", c.httpClient.Timeout)
	}
}

func TestWithInsecureClonesExistingTransport(t *testing.T) {
	// Caller-supplied transport with a custom setting we want to keep.
	base := &http.Transport{MaxIdleConns: 99}
	httpc := &http.Client{Transport: base}

	c, err := NewClient(
		WithBaseURL("https://x"),
		WithAPIKey("k"),
		WithHTTPClient(httpc),
		WithInsecure(),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	tr, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is %T, want *http.Transport", c.httpClient.Transport)
	}
	if tr.MaxIdleConns != 99 {
		t.Errorf("MaxIdleConns = %d, want 99 (transport should be cloned, not replaced)", tr.MaxIdleConns)
	}
	if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify not set")
	}
	if tr == base {
		t.Error("Transport was not cloned (same pointer)")
	}
}

func TestWithInsecureWithoutPriorTransport(t *testing.T) {
	c, err := NewClient(WithBaseURL("https://x"), WithAPIKey("k"), WithInsecure())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	tr, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is %T", c.httpClient.Transport)
	}
	if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify not set")
	}
}

func TestWithRetryOptions(t *testing.T) {
	c, err := NewClient(WithBaseURL("https://dns.example.com"), WithAPIKey("k"), WithRetry(4))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.retry == nil || c.retry.maxAttempts != 4 || c.retry.backoff == nil {
		t.Fatalf("retry = %+v, want 4 attempts with default backoff", c.retry)
	}

	c, _ = NewClient(WithBaseURL("https://dns.example.com"), WithAPIKey("k"), WithRetry(4), WithRetry(1))
	if c.retry != nil {
		t.Errorf("WithRetry(1) should disable retries, got %+v", c.retry)
	}

	backoff := func(int) time.Duration { return time.Second }
	c, _ = NewClient(WithBaseURL("https://dns.example.com"), WithAPIKey("k"), WithRetryBackoff(backoff))
	if c.retry == nil || c.retry.maxAttempts != 3 || c.retry.backoff(0) != time.Second {
		t.Errorf("WithRetryBackoff alone = %+v, want 3 attempts with custom backoff", c.retry)
	}

	c, _ = NewClient(WithBaseURL("https://dns.example.com"), WithAPIKey("k"), WithRetryBackoff(backoff), WithRetry(6))
	if c.retry.maxAttempts != 6 || c.retry.backoff(0) != time.Second {
		t.Errorf("WithRetryBackoff then WithRetry = %+v", c.retry)
	}
}
