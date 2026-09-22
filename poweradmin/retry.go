// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"math/rand/v2"
	"net/http"
	"time"
)

// Backoff returns the delay before the next retry attempt.
// attempt is the zero-based index of the *next* attempt (0 = first retry).
type Backoff func(attempt int) time.Duration

// retryConfig holds retry behavior; nil means "no retries".
type retryConfig struct {
	maxAttempts int
	backoff     Backoff
}

// DefaultBackoff is exponential with ±25% jitter, hard-capped at 10s.
// attempt 0 → ~100ms, 1 → ~200ms, 2 → ~400ms, 3 → ~800ms, ..., ≤ 10s.
func DefaultBackoff(attempt int) time.Duration {
	const (
		base       = 100 * time.Millisecond
		maxBackoff = 10 * time.Second
	)
	shift := min(attempt,
		// base << 16 ≈ 109min — well beyond max
		16)
	d := base << shift
	jitterRange := int64(d / 2)
	if jitterRange > 0 {
		d += time.Duration(rand.Int64N(jitterRange)) - d/4
	}
	if d > maxBackoff {
		return maxBackoff
	}
	if d < base {
		return base
	}
	return d
}

// shouldRetryStatus reports whether a response with the given status code may
// be retried. 429 (Too Many Requests) means the server rejected the request
// without processing it, so it is always retried. 5xx is only retried for
// idempotent methods: a POST that failed with 502 may still have created the
// zone or record, and replaying it would create a duplicate.
func shouldRetryStatus(method string, code int) bool {
	if code == http.StatusTooManyRequests {
		return true
	}
	return code >= 500 && code < 600 && isIdempotent(method)
}

// shouldRetryError reports whether a transport error may be retried. The
// request may have reached the server, so only idempotent methods qualify.
func shouldRetryError(method string) bool {
	return isIdempotent(method)
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}
