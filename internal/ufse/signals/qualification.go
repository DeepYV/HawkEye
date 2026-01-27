/**
 * Signal Qualification
 * 
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Qualify candidate signals
 * 
 * Qualification requires:
 * - Temporal proximity
 * - Clear cause-effect relationship
 * - Absence of success resolution
 */

package signals

import (
	"time"

)

// QualifiedSignal represents a qualified signal
type QualifiedSignal struct {
	Type              string
	Timestamp         time.Time
	Route             string
	Details           map[string]interface{}
	SystemFeedback    bool    // Whether this is a system feedback signal
	Strength          float64 // Signal strength (0-1)
}

// QualifySignals qualifies candidate signals
func QualifySignals(candidates []CandidateSignal, classified []ClassifiedEvent) []QualifiedSignal {
	qualified := make([]QualifiedSignal, 0)

	for _, candidate := range candidates {
		if isQualified(candidate, classified) {
			qualified = append(qualified, QualifiedSignal{
				Type:           candidate.Type,
				Timestamp:      time.Unix(candidate.Timestamp, 0),
				Route:          candidate.Route,
				Details:        candidate.Details,
				SystemFeedback: isSystemFeedbackSignal(candidate, classified),
			})
		}
	}

	return qualified
}

// isQualified checks if candidate signal is qualified
func isQualified(candidate CandidateSignal, classified []ClassifiedEvent) bool {
	candidateTime := time.Unix(candidate.Timestamp, 0)

	// Check temporal proximity (events should be close in time)
	if !hasTemporalProximity(candidateTime, classified) {
		return false
	}

	// Check for clear cause-effect relationship
	if !hasCauseEffectRelationship(candidate, classified) {
		return false
	}

	// Check absence of success resolution
	if hasSuccessResolution(candidateTime, classified) {
		return false // Signal resolved, not qualified
	}

	return true
}

// hasTemporalProximity checks if events are temporally close
func hasTemporalProximity(candidateTime time.Time, classified []ClassifiedEvent) bool {
	proximityWindow := 30 * time.Second
	windowStart := candidateTime.Add(-proximityWindow)
	windowEnd := candidateTime.Add(proximityWindow)

	hasEventsInWindow := false
	for _, event := range classified {
		if event.Timestamp.After(windowStart) && event.Timestamp.Before(windowEnd) {
			hasEventsInWindow = true
			break
		}
	}

	return hasEventsInWindow
}

// hasCauseEffectRelationship checks for clear cause-effect
func hasCauseEffectRelationship(candidate CandidateSignal, classified []ClassifiedEvent) bool {
	candidateTime := time.Unix(candidate.Timestamp, 0)
	window := 10 * time.Second

	// For rage and blocked, check for system feedback nearby
	if candidate.Type == "rage" || candidate.Type == "blocked" {
		for _, event := range classified {
			if event.Timestamp.After(candidateTime) && event.Timestamp.Before(candidateTime.Add(window)) {
				if event.Category == CategorySystemFeedback {
					return true // System feedback provides cause
				}
			}
		}
	}

	// For abandonment, check for friction before abandonment
	if candidate.Type == "abandonment" {
		// Friction should have occurred (checked in detection)
		return true
	}

	// For confusion, check for lack of progress
	if candidate.Type == "confusion" {
		return true // Confusion is self-evident from pattern
	}

	return false
}

// hasSuccessResolution checks if issue was resolved successfully
func hasSuccessResolution(candidateTime time.Time, classified []ClassifiedEvent) bool {
	// Look for success indicators after candidate signal
	for _, event := range classified {
		if event.Timestamp.After(candidateTime) {
			// Success navigation
			if event.Category == CategoryNavigation {
				if isSuccessPageRoute(event.Route) {
					return true
				}
			}
			// Success response
			if event.Category == CategorySystemFeedback {
				if metadata := event.Event.Metadata; metadata != nil {
					if status, ok := metadata["status"].(float64); ok {
						if status >= 200 && status < 300 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// isSystemFeedbackSignal checks if signal is related to system feedback
func isSystemFeedbackSignal(candidate CandidateSignal, classified []ClassifiedEvent) bool {
	candidateTime := time.Unix(candidate.Timestamp, 0)
	window := 10 * time.Second

	for _, event := range classified {
		if event.Timestamp.After(candidateTime.Add(-window)) && event.Timestamp.Before(candidateTime.Add(window)) {
			if event.Category == CategorySystemFeedback {
				return true
			}
		}
	}

	return false
}
