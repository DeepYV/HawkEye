/**
 * Signal Detection Helpers
 *
 * Shared helper functions for signal detection
 */

package signals

import (
	"strings"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// ClassifiedEvent represents an event with its classification
type ClassifiedEvent struct {
	Event     types.Event
	Category  string // "interaction", "system_feedback", "navigation", "performance"
	Timestamp time.Time
	Route     string
}

// Category constants
const (
	CategoryInteraction    = "interaction"
	CategorySystemFeedback = "system_feedback"
	CategoryNavigation     = "navigation"
	CategoryPerformance    = "performance"
)

// getTargetID gets a unique identifier for the event target
func getTargetID(event ClassifiedEvent) string {
	if event.Event.Target.ID != "" {
		return event.Event.Target.ID
	}
	if event.Event.Target.Selector != "" {
		return event.Event.Target.Selector
	}
	return event.Event.Target.Type
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// isRejection checks if event represents a rejection
func isRejection(event ClassifiedEvent) bool {
	if event.Event.EventType == "error" {
		return true
	}
	if event.Event.EventType == "network" {
		if metadata := event.Event.Metadata; metadata != nil {
			if status, ok := metadata["status"].(float64); ok {
				if status >= 400 {
					return true // HTTP error
				}
			}
			if _, ok := metadata["error"]; ok {
				return true // Network error
			}
		}
	}
	return false
}

// isSuccessPage checks if route indicates success
func isSuccessPage(route string) bool {
	successKeywords := []string{"success", "thank", "complete", "confirmation"}
	for _, keyword := range successKeywords {
		if contains(route, keyword) {
			return true
		}
	}
	return false
}

// isRejectionEventRefined checks if event represents a rejection (refined version)
func isRejectionEventRefined(event ClassifiedEvent) bool {
	return isRejection(event)
}
