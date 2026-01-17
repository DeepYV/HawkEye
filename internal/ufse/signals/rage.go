/**
 * Rage Interaction Signal Detection
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Detect rage interaction candidate signals
 *
 * Rules:
 * - ≥ N interactions (N = 3 minimum)
 * - Same target
 * - Short time window (5 seconds)
 * - No success feedback between
 */

package signals

import (
	"time"

)

const (
	rageMinInteractions = 3
	rageTimeWindow      = 5 * time.Second
)

// DetectRageInteraction detects rage interaction candidate signals
func DetectRageInteraction(classified []ClassifiedEvent) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Group events by target
	targetGroups := make(map[string][]ClassifiedEvent)
	for _, event := range classified {
		if event.Category == CategoryInteraction {
			targetID := getTargetID(event)
			targetGroups[targetID] = append(targetGroups[targetID], event)
		}
	}

	// Check each target group for rage pattern
	for targetID, events := range targetGroups {
		if len(events) < rageMinInteractions {
			continue
		}

		// Check for rapid interactions within time window
		for i := 0; i <= len(events)-rageMinInteractions; i++ {
			firstEvent := events[i]
			lastEvent := events[i+rageMinInteractions-1]

			// Check time window
			timeDiff := lastEvent.Timestamp.Sub(firstEvent.Timestamp)
			if timeDiff > rageTimeWindow {
				continue
			}

			// Check for success feedback between interactions
			hasSuccessFeedback := checkSuccessFeedback(classified, firstEvent.Timestamp, lastEvent.Timestamp)
			if hasSuccessFeedback {
				continue // Not rage if success occurred
			}

			// Candidate rage signal detected
			candidates = append(candidates, CandidateSignal{
				Type:      "rage",
				Timestamp: firstEvent.Timestamp.Unix(),
				Route:     firstEvent.Route,
				Details: map[string]interface{}{
					"target_id":    targetID,
					"interactions": len(events[i : i+rageMinInteractions]),
					"time_window":  timeDiff.Seconds(),
				},
			})
		}
	}

	return candidates
}

// checkSuccessFeedback checks if there was success feedback between timestamps
func checkSuccessFeedback(classified []ClassifiedEvent, start, end time.Time) bool {
	for _, event := range classified {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			// Check for success indicators
			if event.Category == CategorySystemFeedback {
				if metadata := event.Event.Metadata; metadata != nil {
					if status, ok := metadata["status"].(float64); ok {
						if status >= 200 && status < 300 {
							return true // Success response
						}
					}
				}
			}
			// Check for navigation away (might indicate success)
			if event.Category == CategoryNavigation {
				// Navigation after interaction might indicate success
				return true
			}
		}
	}
	return false
}
