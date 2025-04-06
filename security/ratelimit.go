package security

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new RateLimiter instance
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed based on rate limiting rules
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get the requests for this key
	requests, ok := rl.requests[key]
	if !ok {
		// First request for this key
		rl.requests[key] = []time.Time{now}
		return true
	}

	// Filter out requests outside the window
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	// Check if we're over the limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add the new request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	return true
}

// Reset clears all rate limiting data
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.requests = make(map[string][]time.Time)
}

// HTTPRateLimiter provides rate limiting for HTTP clients
type HTTPRateLimiter struct {
	limiter *RateLimiter
}

// NewHTTPRateLimiter creates a new HTTPRateLimiter instance
func NewHTTPRateLimiter(limit int, window time.Duration) *HTTPRateLimiter {
	return &HTTPRateLimiter{
		limiter: NewRateLimiter(limit, window),
	}
}

// WrapHTTPClient wraps an HTTP client with rate limiting
func (rl *HTTPRateLimiter) WrapHTTPClient(client *http.Client, key string) *http.Client {
	// Create a custom transport that applies rate limiting
	transport := &rateLimitedTransport{
		base:    client.Transport,
		limiter: rl.limiter,
		key:     key,
	}

	// Create a new client with the custom transport
	return &http.Client{
		Transport:     transport,
		Timeout:       client.Timeout,
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
	}
}

// rateLimitedTransport is a custom HTTP transport that applies rate limiting
type rateLimitedTransport struct {
	base    http.RoundTripper
	limiter *RateLimiter
	key     string
}

// RoundTrip implements the http.RoundTripper interface
func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Apply rate limiting
	if !t.limiter.Allow(t.key) {
		return nil, fmt.Errorf("rate limit exceeded for %s", t.key)
	}

	// Use the base transport if available, otherwise use the default
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	// Forward the request to the base transport
	return base.RoundTrip(req)
}
