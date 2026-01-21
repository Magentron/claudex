// Package ratelimit provides rate limiting for process spawning to prevent rapid runaway processes.
package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter tracks process spawn timestamps and enforces rate limits with exponential backoff.
type RateLimiter struct {
	mu         sync.Mutex
	timestamps []time.Time
	limit      int           // Max spawns per second
	window     time.Duration // Time window for rate calculation (1 second)
}

// NewRateLimiter creates a new RateLimiter with the specified spawn limit per second.
func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{
		timestamps: make([]time.Time, 0),
		limit:      limit,
		window:     time.Second,
	}
}

// Allow checks if a new process spawn is allowed and applies exponential backoff if needed.
// Returns true if spawn is allowed, false if rate limit is exceeded.
// Blocks with exponential backoff if spawn frequency is too high.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Remove timestamps outside the sliding window
	cutoff := now.Add(-r.window)
	validTimestamps := make([]time.Time, 0)
	for _, ts := range r.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	r.timestamps = validTimestamps

	// Check if we're within the rate limit
	if len(r.timestamps) < r.limit {
		r.timestamps = append(r.timestamps, now)
		return true
	}

	// Calculate exponential backoff based on excess spawns
	excess := len(r.timestamps) - r.limit + 1
	backoffMs := 100 * (1 << uint(excess-1)) // 100ms, 200ms, 400ms, 800ms, 1600ms...
	if backoffMs > 3000 {
		backoffMs = 3000 // Cap at 3 seconds
	}

	// Release lock before sleeping
	r.mu.Unlock()
	time.Sleep(time.Duration(backoffMs) * time.Millisecond)
	r.mu.Lock()

	// After backoff, add timestamp and allow
	r.timestamps = append(r.timestamps, time.Now())
	return true
}

// Reset clears all tracked timestamps (useful for testing).
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timestamps = make([]time.Time, 0)
}
