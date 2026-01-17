/**
 * Abandonment Signal Detection
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Detect abandonment candidate signals
 *
 * Rules:
 * - Flow start
 * - Friction event occurs
 * - Session ends without completion
 * - Completion must be explicit
 */

package signals

import (
	"time"

)

const (
	abandonmentTimeWindow = 60 * time.Second // Time window for flow start → friction → abandonment
)

// DetectAbandonment detects abandonment candidate signals
func DetectAbandonment(classified []ClassifiedEvent) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Identify flow starts (form_submit, navigation to checkout, etc.)
	flowStarts := findFlowStarts(classified)
	if len(flowStarts) == 0 {
		return candidates
	}

	// For each flow start, check for friction and abandonment
	for _, flowStart := range flowStarts {
		// Find friction events after flow start
		friction := findFriction(classified, flowStart.Timestamp, abandonmentTimeWindow)
		if friction == nil {
			continue
		}

		// Check if flow was completed (explicit completion required)
		completed := checkFlowCompletion(classified, flowStart.Timestamp)
		if completed {
			continue // Flow completed, not abandonment
		}

		// Candidate abandonment signal detected
		candidates = append(candidates, CandidateSignal{
			Type:      "abandonment",
			Timestamp: flowStart.Timestamp.Unix(),
			Route:     flowStart.Route,
			Details: map[string]interface{}{
				"flow_type":     getFlowType(flowStart),
				"friction_type": getFrictionType(friction),
			},
		})
	}

	return candidates
}

// findFlowStarts identifies flow start events
func findFlowStarts(classified []ClassifiedEvent) []ClassifiedEvent {
	starts := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		// Form submit indicates flow start
		if event.Event.EventType == "form_submit" {
			starts = append(starts, event)
		}
		// Navigation to checkout/checkout-related routes
		if event.Category == CategoryNavigation {
			if isCheckoutFlow(event.Route) {
				starts = append(starts, event)
			}
		}
	}
	return starts
}

// isCheckoutFlow checks if route indicates checkout flow
func isCheckoutFlow(route string) bool {
	checkoutKeywords := []string{"checkout", "payment", "purchase", "buy", "cart"}
	for _, keyword := range checkoutKeywords {
		if contains(route, keyword) {
			return true
		}
	}
	return false
}

// findFriction finds friction events (errors, delays, rejections)
func findFriction(classified []ClassifiedEvent, after time.Time, window time.Duration) *ClassifiedEvent {
	windowEnd := after.Add(window)
	for _, event := range classified {
		if event.Timestamp.After(after) && event.Timestamp.Before(windowEnd) {
			// System feedback errors
			if event.Category == CategorySystemFeedback {
				if isRejection(event) {
					return &event
				}
			}
			// Performance issues (slow loading)
			if event.Category == CategoryPerformance {
				if metadata := event.Event.Metadata; metadata != nil {
					if duration, ok := metadata["duration"].(float64); ok {
						if duration > 3000 { // 3 seconds
							return &event
						}
					}
				}
			}
		}
	}
	return nil
}

// checkFlowCompletion checks if flow was explicitly completed
func checkFlowCompletion(classified []ClassifiedEvent, flowStartTime time.Time) bool {
	for _, event := range classified {
		if event.Timestamp.After(flowStartTime) {
			// Check for success indicators
			if event.Category == CategorySystemFeedback {
				if metadata := event.Event.Metadata; metadata != nil {
					if status, ok := metadata["status"].(float64); ok {
						if status >= 200 && status < 300 {
							// Check if it's a completion endpoint
							if isCompletionEndpoint(event.Event.Route) {
								return true
							}
						}
					}
				}
			}
			// Navigation to success/thank-you page
			if event.Category == CategoryNavigation {
				if isSuccessPageRoute(event.Route) {
					return true
				}
			}
		}
	}
	return false
}

// isCompletionEndpoint checks if route is a completion endpoint
func isCompletionEndpoint(route string) bool {
	completionKeywords := []string{"success", "thank", "complete", "confirmation", "receipt"}
	for _, keyword := range completionKeywords {
		if contains(route, keyword) {
			return true
		}
	}
	return false
}

// isSuccessPage checks if route is a success page
func isSuccessPageRoute(route string) bool {
	return isCompletionEndpoint(route)
}

// getFlowType gets the type of flow
func getFlowType(event ClassifiedEvent) string {
	if event.Event.EventType == "form_submit" {
		return "form_submission"
	}
	if isCheckoutFlow(event.Route) {
		return "checkout"
	}
	return "unknown"
}

// getFrictionType gets the type of friction
func getFrictionType(event *ClassifiedEvent) string {
	if event.Category == CategorySystemFeedback {
		return "system_error"
	}
	if event.Category == CategoryPerformance {
		return "performance_issue"
	}
	return "unknown"
}
