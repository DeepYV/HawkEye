/**
 * Rate Limiting
 *
 * Author: Frank Miller (Team Beta)
 * Responsibility: Per-API-key rate limiting with burst tolerance
 */

package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements token bucket rate limiting per API key
type Limiter struct {
	mu              sync.RWMutex
	limiters        map[string]*tokenBucket
	rate            int // requests per second
	burst           int // burst capacity
	cleanupInterval time.Duration
}

// tokenBucket implements token bucket algorithm
type tokenBucket struct {
	tokens     float64
	capacity   float64
	rate       float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewLimiter creates a new rate limiter
func NewLimiter(rate, burst int) *Limiter {
	limiter := &Limiter{
		limiters:        make(map[string]*tokenBucket),
		rate:            rate,
		burst:           burst,
		cleanupInterval: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if request is allowed for given API key
func (l *Limiter) Allow(ctx context.Context, apiKey string) bool {
	if apiKey == "" {
		return false
	}

	// Get or create bucket for API key
	bucket := l.getBucket(apiKey)
	return bucket.allow()
}

// getBucket gets or creates token bucket for API key
func (l *Limiter) getBucket(apiKey string) *tokenBucket {
	l.mu.RLock()
	bucket, exists := l.limiters[apiKey]
	l.mu.RUnlock()

	if exists {
		return bucket
	}

	// Create new bucket
	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists := l.limiters[apiKey]; exists {
		return bucket
	}

	bucket = &tokenBucket{
		tokens:     float64(l.burst),
		capacity:   float64(l.burst),
		rate:       float64(l.rate),
		lastRefill: time.Now(),
	}

	l.limiters[apiKey] = bucket
	return bucket
}

// allow checks if token bucket allows request
func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// Refill tokens based on elapsed time
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.rate)
	tb.lastRefill = now

	// Check if we have tokens
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// cleanup removes old buckets periodically
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		// Remove buckets that haven't been used recently
		// (Simplified - in production, track last access time)
		if len(l.limiters) > 10000 {
			// Reset if too many buckets (simple cleanup)
			l.limiters = make(map[string]*tokenBucket)
		}
		l.mu.Unlock()
	}
}

// min returns minimum of two floats
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
