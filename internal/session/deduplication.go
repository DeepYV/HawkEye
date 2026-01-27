/**
 * Event Deduplication
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Event deduplication logic
 */

package session

import (
	"crypto/sha256"
	"fmt"

	"github.com/your-org/frustration-engine/internal/types"
)

// DeduplicateEvents removes duplicate events, keeping first occurrence
// Uses idempotency keys if available, falls back to fingerprint-based deduplication
func DeduplicateEvents(events []types.Event) []types.Event {
	seenIdempotencyKeys := make(map[string]bool)
	seenFingerprints := make(map[string]bool)
	result := make([]types.Event, 0, len(events))

	for _, event := range events {
		// Prefer idempotency key if available (client-provided, most reliable)
		if event.IdempotencyKey != "" {
			if !seenIdempotencyKeys[event.IdempotencyKey] {
				seenIdempotencyKeys[event.IdempotencyKey] = true
				result = append(result, event)
			}
			// Otherwise, drop duplicate (conservative: keep first)
			continue
		}

		// Fall back to fingerprint-based deduplication
		fingerprint := createEventFingerprint(event)
		if !seenFingerprints[fingerprint] {
			seenFingerprints[fingerprint] = true
			result = append(result, event)
		}
		// Otherwise, drop duplicate (conservative: keep first)
	}

	return result
}

// createEventFingerprint creates a unique fingerprint for an event
func createEventFingerprint(event types.Event) string {
	// Create fingerprint from key fields
	key := fmt.Sprintf("%s:%s:%s:%s:%s",
		event.EventType,
		event.Timestamp,
		event.SessionID,
		event.Route,
		event.Target.Type,
	)

	// Include target ID if present
	if event.Target.ID != "" {
		key += ":" + event.Target.ID
	}

	// Include selector if present
	if event.Target.Selector != "" {
		key += ":" + event.Target.Selector
	}

	// Hash the key for consistent fingerprint
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// IsDuplicate checks if an event is a duplicate
// Uses idempotency key if available, falls back to fingerprint comparison
func IsDuplicate(event types.Event, existingEvents []types.Event) bool {
	// Prefer idempotency key comparison
	if event.IdempotencyKey != "" {
		for _, existing := range existingEvents {
			if existing.IdempotencyKey == event.IdempotencyKey {
				return true
			}
		}
		return false
	}

	// Fall back to fingerprint comparison
	fingerprint := createEventFingerprint(event)

	for _, existing := range existingEvents {
		if createEventFingerprint(existing) == fingerprint {
			return true
		}
	}

	return false
}

// IdempotencyCache provides time-bounded idempotency key tracking
// to prevent duplicate signals at scale
type IdempotencyCache struct {
	keys       map[string]int64 // key -> expiry timestamp (unix)
	maxSize    int
	ttlSeconds int64
}

// NewIdempotencyCache creates a new idempotency cache
func NewIdempotencyCache(maxSize int, ttlSeconds int64) *IdempotencyCache {
	return &IdempotencyCache{
		keys:       make(map[string]int64),
		maxSize:    maxSize,
		ttlSeconds: ttlSeconds,
	}
}

// CheckAndAdd checks if key exists (returns true if duplicate)
// and adds it if not found
func (c *IdempotencyCache) CheckAndAdd(key string, timestamp int64) bool {
	// Clean expired keys periodically
	if len(c.keys) > c.maxSize/2 {
		c.cleanExpired(timestamp)
	}

	// Check if key exists and not expired
	if expiry, exists := c.keys[key]; exists {
		if timestamp < expiry {
			return true // Duplicate
		}
		// Key expired, allow reuse
	}

	// Add/update key with TTL
	c.keys[key] = timestamp + c.ttlSeconds

	return false // Not a duplicate
}

// cleanExpired removes expired keys from the cache
func (c *IdempotencyCache) cleanExpired(currentTimestamp int64) {
	for key, expiry := range c.keys {
		if currentTimestamp >= expiry {
			delete(c.keys, key)
		}
	}
}
