/**
 * Export Rate Limiting
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Rate limiting for exports (never spam)
 */

package exporter

import (
	"sync"
	"time"
)

// RateLimiter limits export rate per project
type RateLimiter struct {
	mu           sync.RWMutex
	lastExport   map[string]time.Time
	maxPerMinute int
	minInterval  time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxPerMinute int) *RateLimiter {
	minInterval := time.Minute / time.Duration(maxPerMinute)
	if minInterval < time.Second {
		minInterval = time.Second
	}

	return &RateLimiter{
		lastExport:   make(map[string]time.Time),
		maxPerMinute: maxPerMinute,
		minInterval:  minInterval,
	}
}

// Allow checks if export is allowed for project
func (r *RateLimiter) Allow(projectID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	lastExport, exists := r.lastExport[projectID]

	if !exists {
		// First export for this project
		r.lastExport[projectID] = now
		return true
	}

	// Check if enough time has passed
	if now.Sub(lastExport) < r.minInterval {
		return false // Rate limited
	}

	// Update last export time
	r.lastExport[projectID] = now
	return true
}
