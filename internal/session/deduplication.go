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
func DeduplicateEvents(events []types.Event) []types.Event {
	seen := make(map[string]bool)
	result := make([]types.Event, 0, len(events))

	for _, event := range events {
		// Create fingerprint for event
		fingerprint := createEventFingerprint(event)

		// If not seen, add to result
		if !seen[fingerprint] {
			seen[fingerprint] = true
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
func IsDuplicate(event types.Event, existingEvents []types.Event) bool {
	fingerprint := createEventFingerprint(event)

	for _, existing := range existingEvents {
		if createEventFingerprint(existing) == fingerprint {
			return true
		}
	}

	return false
}
