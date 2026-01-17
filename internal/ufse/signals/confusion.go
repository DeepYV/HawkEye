/**
 * Confusion Signal Detection
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Detect confusion candidate signals
 *
 * Rules:
 * - Route oscillation (back and forth)
 * - Excessive scrolling or hovering
 * - No progress events
 * - Confusion is low severity by default
 */

package signals

import (
	"time"

)

const (
	confusionMinOscillations = 3 // Minimum route changes back and forth
	confusionTimeWindow      = 60 * time.Second
	confusionMinScrolls      = 10 // Excessive scrolling threshold
)

// DetectConfusion detects confusion candidate signals
func DetectConfusion(classified []ClassifiedEvent) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Detect route oscillation
	oscillation := detectRouteOscillation(classified)
	if oscillation != nil {
		candidates = append(candidates, *oscillation)
	}

	// Detect excessive scrolling
	excessiveScroll := detectExcessiveScrolling(classified)
	if excessiveScroll != nil {
		candidates = append(candidates, *excessiveScroll)
	}

	return candidates
}

// detectRouteOscillation detects back-and-forth navigation
func detectRouteOscillation(classified []ClassifiedEvent) *CandidateSignal {
	navigationEvents := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		if event.Category == CategoryNavigation {
			navigationEvents = append(navigationEvents, event)
		}
	}

	if len(navigationEvents) < confusionMinOscillations {
		return nil
	}

	// Check for oscillation pattern (A → B → A → B)
	routes := make([]string, 0, len(navigationEvents))
	for _, event := range navigationEvents {
		routes = append(routes, event.Route)
	}

	oscillations := countOscillations(routes)
	if oscillations >= confusionMinOscillations {
		return &CandidateSignal{
			Type:      "confusion",
			Timestamp: navigationEvents[0].Timestamp.Unix(),
			Route:     navigationEvents[0].Route,
			Details: map[string]interface{}{
				"type":         "route_oscillation",
				"oscillations": oscillations,
			},
		}
	}

	return nil
}

// countOscillations counts back-and-forth route changes
func countOscillations(routes []string) int {
	if len(routes) < 2 {
		return 0
	}

	oscillations := 0
	for i := 1; i < len(routes); i++ {
		// Check if route changed
		if routes[i] != routes[i-1] {
			// Check if it's oscillating (going back to previous route)
			if i >= 2 && routes[i] == routes[i-2] {
				oscillations++
			}
		}
	}

	return oscillations
}

// detectExcessiveScrolling detects excessive scrolling without progress
func detectExcessiveScrolling(classified []ClassifiedEvent) *CandidateSignal {
	scrollEvents := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		if event.Event.EventType == "scroll" {
			scrollEvents = append(scrollEvents, event)
		}
	}

	if len(scrollEvents) < confusionMinScrolls {
		return nil
	}

	// Check time window
	if len(scrollEvents) > 0 {
		firstScroll := scrollEvents[0]
		lastScroll := scrollEvents[len(scrollEvents)-1]
		timeDiff := lastScroll.Timestamp.Sub(firstScroll.Timestamp)

		if timeDiff < confusionTimeWindow {
			// Check for progress events (clicks, form submissions)
			hasProgress := false
			for _, event := range classified {
				if event.Timestamp.After(firstScroll.Timestamp) && event.Timestamp.Before(lastScroll.Timestamp) {
					if event.Event.EventType == "click" || event.Event.EventType == "form_submit" {
						hasProgress = true
						break
					}
				}
			}

			if !hasProgress {
				// Excessive scrolling without progress
				return &CandidateSignal{
					Type:      "confusion",
					Timestamp: firstScroll.Timestamp.Unix(),
					Route:     firstScroll.Route,
					Details: map[string]interface{}{
						"type":         "excessive_scrolling",
						"scroll_count": len(scrollEvents),
					},
				}
			}
		}
	}

	return nil
}
