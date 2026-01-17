/**
 * Refined Rage Signal Detection
 * 
 * Author: Team Alpha (Bob, Charlie)
 * Responsibility: Production-grade rage signal detection with zero false alarms
 * 
 * Handles all edge cases: legitimate rapid clicks, accessibility, gaming, etc.
 */

package signals

import (
	"log"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Refined thresholds
	rageMinInteractionsRefined = 4 // Increased from 3 to reduce false positives
	rageTimeWindowRefined      = 3 * time.Second // Reduced from 5s for precision
	rageMaxTimeBetweenClicks    = 500 * time.Millisecond // Max time between clicks
)

// RefinedRageDetector detects rage signals with comprehensive false alarm prevention
type RefinedRageDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewRefinedRageDetector creates a new refined rage detector
func NewRefinedRageDetector() *RefinedRageDetector {
	return &RefinedRageDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectRageInteractionRefined detects rage with comprehensive edge case handling
func (d *RefinedRageDetector) DetectRageInteractionRefined(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Group events by target
	targetGroups := make(map[string][]ClassifiedEvent)
	for _, event := range classified {
		if event.Category == CategoryInteraction && event.Event.EventType == "click" {
			targetID := getTargetIDFromClassified(event)
			targetGroups[targetID] = append(targetGroups[targetID], event)
		}
	}

	// Check each target group
	for targetID, events := range targetGroups {
		if len(events) < rageMinInteractionsRefined {
			continue
		}

		// Check for rapid interactions within time window
		for i := 0; i <= len(events)-rageMinInteractionsRefined; i++ {
			firstEvent := events[i]
			lastEvent := events[i+rageMinInteractionsRefined-1]

			// Parse timestamps
			t1, err1 := time.Parse(time.RFC3339, firstEvent.Event.Timestamp)
			t2, err2 := time.Parse(time.RFC3339, lastEvent.Event.Timestamp)
			if err1 != nil || err2 != nil {
				continue
			}

			timeWindow := t2.Sub(t1)
			if timeWindow > rageTimeWindowRefined {
				continue // Too slow, not rage
			}

			// Check time between consecutive clicks
			allRapid := true
			for j := i; j < i+rageMinInteractionsRefined-1; j++ {
				tj, _ := time.Parse(time.RFC3339, events[j].Event.Timestamp)
				tj1, _ := time.Parse(time.RFC3339, events[j+1].Event.Timestamp)
				if tj1.Sub(tj) > rageMaxTimeBetweenClicks {
					allRapid = false
					break
				}
			}

			if !allRapid {
				continue // Not all clicks are rapid
			}

			// Check for success feedback between clicks
			hasSuccessFeedback := checkSuccessFeedbackBetweenClicks(events[i:i+rageMinInteractionsRefined], session.Events, t1, t2)
			if hasSuccessFeedback {
				continue // Success occurred, not frustration
			}

			// Create candidate signal
			candidate := CandidateSignal{
				Type:      "rage",
				Timestamp: t1.Unix(),
				Route:     firstEvent.Event.Route,
				Details: map[string]interface{}{
					"targetID":        targetID,
					"interactionCount": rageMinInteractionsRefined,
					"timeWindow":       timeWindow.String(),
				},
			}

			// Check for false alarms
			if isFalse, reason := d.falseAlarmPreventer.IsFalseAlarm(candidate, session, convertToEvents(classified)); isFalse {
				// Log false alarm for analysis
				log.Printf("[Rage Detection] False alarm prevented: %s", reason)
				continue
			}

			candidates = append(candidates, candidate)
			break // Only one rage signal per target
		}
	}

	return candidates
}

// checkSuccessFeedbackBetweenClicks checks if there's success feedback between clicks
func checkSuccessFeedbackBetweenClicks(clickEvents []ClassifiedEvent, allEvents []types.Event, firstTime, lastTime time.Time) bool {
	if len(clickEvents) < 2 {
		return false
	}

	// Check for success indicators between first and last click
	for _, event := range allEvents {
		eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			continue
		}

		if eventTime.After(firstTime) && eventTime.Before(lastTime) {
			// Check for success indicators
			if event.EventType == "navigation" {
				// Navigation change might indicate success
				if event.Route != clickEvents[0].Event.Route {
					return true // Route changed, likely success
				}
			}
			if event.EventType == "network" {
				if metadata := event.Metadata; metadata != nil {
					if status, ok := metadata["status"].(float64); ok {
						if status >= 200 && status < 300 {
							return true // Successful network request
						}
					}
				}
			}
		}
	}

	return false
}

// convertToEvents converts classified events to regular events
func convertToEvents(classified []ClassifiedEvent) []types.Event {
	events := make([]types.Event, len(classified))
	for i, ce := range classified {
		events[i] = ce.Event
	}
	return events
}

// getTargetIDFromClassified gets target ID from classified event
func getTargetIDFromClassified(event ClassifiedEvent) string {
	// Use helper function from helpers.go which takes ClassifiedEvent
	return getTargetID(event)
}
