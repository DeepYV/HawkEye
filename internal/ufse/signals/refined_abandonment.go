/**
 * Refined Abandonment Signal Detection
 * 
 * Author: Team Alpha (Diana)
 * Responsibility: Production-grade abandonment detection with zero false alarms
 * 
 * Handles all edge cases: intentional navigation, external links, tab switching, etc.
 */

package signals

import (
	"log"
	"strings"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Refined thresholds
	abandonmentTimeWindowRefined = 60 * time.Second
	abandonmentMinFrictionEvents  = 1 // At least one friction event
)

// RefinedAbandonmentDetector detects abandonment signals with comprehensive false alarm prevention
type RefinedAbandonmentDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewRefinedAbandonmentDetector creates a new refined abandonment detector
func NewRefinedAbandonmentDetector() *RefinedAbandonmentDetector {
	return &RefinedAbandonmentDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectAbandonmentRefined detects abandonment with comprehensive edge case handling
func (d *RefinedAbandonmentDetector) DetectAbandonmentRefined(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Identify flow starts (form_submit, navigation to checkout, etc.)
	flowStarts := findFlowStartsRefined(classified)
	if len(flowStarts) == 0 {
		return candidates
	}

	// For each flow start, check for friction and abandonment
	for _, flowStart := range flowStarts {
		// Find friction events after flow start
		frictionEvents := findFrictionEventsRefined(classified, flowStart.Timestamp, abandonmentTimeWindowRefined)
		if len(frictionEvents) < abandonmentMinFrictionEvents {
			continue
		}

		// Check if flow was completed (explicit completion required)
		completed := checkFlowCompletionRefined(classified, flowStart.Timestamp, session)
		if completed {
			continue // Flow completed, not abandonment
		}

		// Check for intentional navigation (external links, sharing, bookmarking)
		if isIntentionalNavigation(session) {
			continue // Intentional navigation, not abandonment
		}

		// Create candidate signal
		candidate := CandidateSignal{
			Type:      "abandonment",
			Timestamp: flowStart.Timestamp.Unix(),
			Route:     flowStart.Route,
			Details: map[string]interface{}{
				"flow_type":      getFlowTypeRefined(flowStart),
				"friction_count": len(frictionEvents),
				"friction_types": getFrictionTypes(frictionEvents),
			},
		}

		// Check for false alarms
		if isFalse, reason := d.falseAlarmPreventer.IsFalseAlarm(candidate, session, convertToEvents(classified)); isFalse {
			log.Printf("[Abandonment Detection] False alarm prevented: %s", reason)
			continue
		}

		candidates = append(candidates, candidate)
	}

	return candidates
}

// findFlowStartsRefined identifies flow start events (refined)
func findFlowStartsRefined(classified []ClassifiedEvent) []ClassifiedEvent {
	starts := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		// Form submit indicates flow start
		if event.Event.EventType == "form_submit" {
			starts = append(starts, event)
		}
		// Navigation to checkout/checkout-related routes
		if event.Category == CategoryNavigation {
			if isCheckoutFlowRefined(event.Route) {
				starts = append(starts, event)
			}
		}
	}
	return starts
}

// isCheckoutFlowRefined checks if route indicates checkout flow (refined)
func isCheckoutFlowRefined(route string) bool {
	routeLower := strings.ToLower(route)
	checkoutKeywords := []string{"checkout", "payment", "purchase", "buy", "cart", "order"}
	for _, keyword := range checkoutKeywords {
		if strings.Contains(routeLower, keyword) {
			return true
		}
	}
	return false
}

// findFrictionEventsRefined finds friction events (errors, delays, rejections) - refined
func findFrictionEventsRefined(classified []ClassifiedEvent, after time.Time, window time.Duration) []ClassifiedEvent {
	windowEnd := after.Add(window)
	frictionEvents := make([]ClassifiedEvent, 0)

	for _, event := range classified {
		if event.Timestamp.After(after) && event.Timestamp.Before(windowEnd) {
			// System feedback errors
			if event.Category == CategorySystemFeedback {
				if isRejectionEventRefined(event) {
					frictionEvents = append(frictionEvents, event)
				}
			}
			// Performance issues (slow loading)
			if event.Category == CategoryPerformance {
				if metadata := event.Event.Metadata; metadata != nil {
					if duration, ok := metadata["duration"].(float64); ok {
						if duration > 3000 { // 3 seconds
							frictionEvents = append(frictionEvents, event)
						}
					}
				}
			}
		}
	}

	return frictionEvents
}

// checkFlowCompletionRefined checks if flow was explicitly completed (refined)
func checkFlowCompletionRefined(classified []ClassifiedEvent, flowStartTime time.Time, session types.Session) bool {
	for _, event := range session.Events {
		eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			continue
		}

		if eventTime.After(flowStartTime) {
			// Check for success indicators
			if event.EventType == "network" {
				if metadata := event.Metadata; metadata != nil {
					if status, ok := metadata["status"].(float64); ok {
						if status >= 200 && status < 300 {
							// Check if it's a completion endpoint
							if isCompletionEndpointRefined(event.Route) {
								return true
							}
						}
					}
				}
			}
			// Navigation to success/thank-you page
			if event.EventType == "navigation" {
				if isSuccessPageRouteRefined(event.Route) {
					return true
				}
			}
		}
	}
	return false
}

// isCompletionEndpointRefined checks if route is a completion endpoint (refined)
func isCompletionEndpointRefined(route string) bool {
	routeLower := strings.ToLower(route)
	completionKeywords := []string{"success", "thank", "complete", "confirmation", "receipt", "order-complete"}
	for _, keyword := range completionKeywords {
		if strings.Contains(routeLower, keyword) {
			return true
		}
	}
	return false
}

// isSuccessPageRouteRefined checks if route is a success page (refined)
func isSuccessPageRouteRefined(route string) bool {
	return isCompletionEndpointRefined(route)
}

// getFlowTypeRefined gets the type of flow (refined)
func getFlowTypeRefined(event ClassifiedEvent) string {
	if event.Event.EventType == "form_submit" {
		return "form_submission"
	}
	if isCheckoutFlowRefined(event.Route) {
		return "checkout"
	}
	return "unknown"
}

// getFrictionTypes gets the types of friction events
func getFrictionTypes(frictionEvents []ClassifiedEvent) []string {
	types := make([]string, 0, len(frictionEvents))
	for _, event := range frictionEvents {
		if event.Category == CategorySystemFeedback {
			types = append(types, "system_error")
		}
		if event.Category == CategoryPerformance {
			types = append(types, "performance_issue")
		}
	}
	return types
}

// isIntentionalNavigation checks if navigation was intentional (external links, sharing, etc.)
func isIntentionalNavigation(session types.Session) bool {
	for _, event := range session.Events {
		if event.EventType == "navigation" {
			if metadata := event.Metadata; metadata != nil {
				if external, ok := metadata["external"].(bool); ok && external {
					return true // External link
				}
				if share, ok := metadata["share"].(bool); ok && share {
					return true // Share action
				}
				if bookmark, ok := metadata["bookmark"].(bool); ok && bookmark {
					return true // Bookmark action
				}
			}
		}
	}
	return false
}

// Note: isRejectionEventRefined is now in helpers.go
