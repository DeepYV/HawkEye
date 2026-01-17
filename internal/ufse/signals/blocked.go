/**
 * Blocked Progress Signal Detection
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Detect blocked progress candidate signals
 *
 * Rules:
 * - Action attempt
 * - System rejection (error or validation)
 * - Retry attempt
 * - Ignore first failure
 */

package signals

import (
	"time"
)

const (
	blockedTimeWindow = 30 * time.Second // Time window for action → rejection → retry
)

// DetectBlockedProgress detects blocked progress candidate signals
func DetectBlockedProgress(classified []ClassifiedEvent) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Find action attempts (form_submit, click on submit buttons)
	actionEvents := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		if isActionAttempt(event) {
			actionEvents = append(actionEvents, event)
		}
	}

	// For each action, check for rejection followed by retry
	for _, action := range actionEvents {
		// Look for system rejection after action
		rejection := findSystemRejection(classified, action.Timestamp, blockedTimeWindow)
		if rejection == nil {
			continue
		}

		// Look for retry after rejection (ignore first failure)
		retry := findRetry(classified, rejection.Timestamp, blockedTimeWindow, action)
		if retry == nil {
			continue
		}

		// Candidate blocked progress signal detected
		candidates = append(candidates, CandidateSignal{
			Type:      "blocked",
			Timestamp: action.Timestamp.Unix(),
			Route:     action.Route,
			Details: map[string]interface{}{
				"action_type":    action.Event.EventType,
				"rejection_type": getRejectionType(rejection),
				"retry_count":    1, // At least one retry
			},
		})
	}

	return candidates
}

// isActionAttempt checks if event is an action attempt
func isActionAttempt(event ClassifiedEvent) bool {
	// Form submit is always an action attempt
	if event.Event.EventType == "form_submit" {
		return true
	}

	// Click on submit button
	if event.Event.EventType == "click" {
		if target := event.Event.Target; target.Type == "button" {
			// Check if it's a submit button (by selector or type)
			if target.Selector != "" {
				// Could check for submit-related selectors
				return true
			}
		}
	}

	return false
}

// findSystemRejection finds system rejection after action
func findSystemRejection(classified []ClassifiedEvent, after time.Time, window time.Duration) *ClassifiedEvent {
	windowEnd := after.Add(window)
	for _, event := range classified {
		if event.Timestamp.After(after) && event.Timestamp.Before(windowEnd) {
			if event.Category == CategorySystemFeedback {
				// Check for error or rejection
				if isRejectionEvent(event) {
					return &event
				}
			}
		}
	}
	return nil
}

// isRejectionEvent checks if event represents a rejection (local version)
func isRejectionEvent(event ClassifiedEvent) bool {
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

// getRejectionType gets the type of rejection
func getRejectionType(event *ClassifiedEvent) string {
	if event.Event.EventType == "error" {
		return "javascript_error"
	}
	if event.Event.EventType == "network" {
		if metadata := event.Event.Metadata; metadata != nil {
			if _, ok := metadata["status"].(float64); ok {
				return "http_error"
			}
		}
		return "network_error"
	}
	return "unknown"
}

// findRetry finds retry attempt after rejection
func findRetry(classified []ClassifiedEvent, after time.Time, window time.Duration, originalAction ClassifiedEvent) *ClassifiedEvent {
	windowEnd := after.Add(window)
	for _, event := range classified {
		if event.Timestamp.After(after) && event.Timestamp.Before(windowEnd) {
			// Check if it's the same type of action
			if event.Event.EventType == originalAction.Event.EventType {
				// Check if it's the same target
				if getTargetID(event) == getTargetID(originalAction) {
					return &event
				}
			}
		}
	}
	return nil
}
