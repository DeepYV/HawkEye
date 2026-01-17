/**
 * Refined Confusion Signal Detection
 * 
 * Author: Team Alpha (Diana)
 * Responsibility: Production-grade confusion detection with zero false alarms
 * 
 * Handles all edge cases: legitimate browsing, search, comparison shopping, etc.
 */

package signals

import (
	"log"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Refined thresholds
	confusionMinOscillationsRefined = 4 // Increased from 3
	confusionTimeWindowRefined      = 60 * time.Second
	confusionMinScrollsRefined      = 15 // Increased from 10
	confusionMinScrollTimeWindow    = 30 * time.Second // Time window for excessive scrolling
)

// RefinedConfusionDetector detects confusion signals with comprehensive false alarm prevention
type RefinedConfusionDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewRefinedConfusionDetector creates a new refined confusion detector
func NewRefinedConfusionDetector() *RefinedConfusionDetector {
	return &RefinedConfusionDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectConfusionRefined detects confusion with comprehensive edge case handling
func (d *RefinedConfusionDetector) DetectConfusionRefined(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Detect route oscillation (refined)
	oscillation := detectRouteOscillationRefined(classified)
	if oscillation != nil {
		// Check for false alarms
		if isFalse, reason := d.falseAlarmPreventer.IsFalseAlarm(*oscillation, session, convertToEvents(classified)); isFalse {
			log.Printf("[Confusion Detection] False alarm prevented (oscillation): %s", reason)
		} else {
			candidates = append(candidates, *oscillation)
		}
	}

	// Detect excessive scrolling (refined)
	excessiveScroll := detectExcessiveScrollingRefined(classified, session)
	if excessiveScroll != nil {
		// Check for false alarms
		if isFalse, reason := d.falseAlarmPreventer.IsFalseAlarm(*excessiveScroll, session, convertToEvents(classified)); isFalse {
			log.Printf("[Confusion Detection] False alarm prevented (scrolling): %s", reason)
		} else {
			candidates = append(candidates, *excessiveScroll)
		}
	}

	return candidates
}

// detectRouteOscillationRefined detects back-and-forth navigation (refined)
func detectRouteOscillationRefined(classified []ClassifiedEvent) *CandidateSignal {
	navigationEvents := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		if event.Category == CategoryNavigation {
			navigationEvents = append(navigationEvents, event)
		}
	}

	if len(navigationEvents) < confusionMinOscillationsRefined {
		return nil
	}

	// Check for oscillation pattern (A → B → A → B)
	routes := make([]string, 0, len(navigationEvents))
	for _, event := range navigationEvents {
		routes = append(routes, event.Route)
	}

	oscillations := countOscillationsRefined(routes)
	if oscillations >= confusionMinOscillationsRefined {
		return &CandidateSignal{
			Type:      "confusion",
			Timestamp: navigationEvents[0].Timestamp.Unix(),
			Route:     navigationEvents[0].Route,
			Details: map[string]interface{}{
				"type":         "route_oscillation",
				"oscillations":  oscillations,
				"route_count":   len(routes),
			},
		}
	}

	return nil
}

// countOscillationsRefined counts back-and-forth route changes (refined)
func countOscillationsRefined(routes []string) int {
	if len(routes) < 2 {
		return 0
	}

	oscillations := 0
	for i := 2; i < len(routes); i++ {
		// Check if route changed
		if routes[i] != routes[i-1] {
			// Check if it's oscillating (going back to previous route)
			if routes[i] == routes[i-2] {
				oscillations++
			}
		}
	}

	return oscillations
}

// detectExcessiveScrollingRefined detects excessive scrolling without progress (refined)
func detectExcessiveScrollingRefined(classified []ClassifiedEvent, session types.Session) *CandidateSignal {
	scrollEvents := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		if event.Event.EventType == "scroll" {
			scrollEvents = append(scrollEvents, event)
		}
	}

	if len(scrollEvents) < confusionMinScrollsRefined {
		return nil
	}

	// Check time window
	if len(scrollEvents) > 0 {
		firstScroll := scrollEvents[0]
		lastScroll := scrollEvents[len(scrollEvents)-1]
		timeDiff := lastScroll.Timestamp.Sub(firstScroll.Timestamp)

		if timeDiff < confusionMinScrollTimeWindow {
			// Check for progress events (clicks, form submissions, navigation)
			hasProgress := checkProgressEvents(session.Events, firstScroll.Timestamp, lastScroll.Timestamp)
			if !hasProgress {
				// Excessive scrolling without progress
				return &CandidateSignal{
					Type:      "confusion",
					Timestamp: firstScroll.Timestamp.Unix(),
					Route:     firstScroll.Route,
					Details: map[string]interface{}{
						"type":         "excessive_scrolling",
						"scroll_count": len(scrollEvents),
						"time_window":   timeDiff.String(),
					},
				}
			}
		}
	}

	return nil
}

// checkProgressEvents checks if there are progress events between timestamps
func checkProgressEvents(events []types.Event, start, end time.Time) bool {
	for _, event := range events {
		eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			continue
		}

		if eventTime.After(start) && eventTime.Before(end) {
			// Check for progress indicators
			if event.EventType == "click" {
				// Check if it's a meaningful click (not just scrolling)
				if event.Target.Type != "body" && event.Target.Type != "html" {
					return true
				}
			}
			if event.EventType == "form_submit" {
				return true
			}
			if event.EventType == "navigation" {
				// Navigation indicates progress
				return true
			}
		}
	}
	return false
}
