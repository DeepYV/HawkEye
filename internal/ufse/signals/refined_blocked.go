/**
 * Refined Blocked Progress Signal Detection
 * 
 * Author: Team Alpha (Bob, Charlie)
 * Responsibility: Production-grade blocked progress detection with zero false alarms
 * 
 * Handles all edge cases: form validation, network retries, legitimate retry logic, etc.
 */

package signals

import (
	"log"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Refined thresholds
	blockedMinRetriesRefined = 2 // At least 2 retries (ignore first failure)
	blockedTimeWindowRefined  = 30 * time.Second
	blockedMaxTimeBetweenRetries = 5 * time.Second // Max time between retries
)

// RefinedBlockedDetector detects blocked progress signals with comprehensive false alarm prevention
type RefinedBlockedDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewRefinedBlockedDetector creates a new refined blocked detector
func NewRefinedBlockedDetector() *RefinedBlockedDetector {
	return &RefinedBlockedDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectBlockedProgressRefined detects blocked progress with comprehensive edge case handling
func (d *RefinedBlockedDetector) DetectBlockedProgressRefined(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Find action attempts (form_submit, click on submit buttons)
	actionEvents := make([]ClassifiedEvent, 0)
	for _, event := range classified {
		if isActionAttemptRefined(event) {
			actionEvents = append(actionEvents, event)
		}
	}

	// For each action, check for rejection followed by retries
	for _, action := range actionEvents {
		// Look for system rejection after action
		rejection := findSystemRejectionRefined(classified, action.Timestamp, blockedTimeWindowRefined)
		if rejection == nil {
			continue
		}

		// Look for retries after rejection (must have at least blockedMinRetriesRefined retries)
		retries := findRetriesRefined(classified, rejection.Timestamp, blockedTimeWindowRefined, action)
		if len(retries) < blockedMinRetriesRefined {
			continue
		}

		// Check time between retries (should be reasonable)
		if !areRetriesReasonable(retries) {
			continue
		}

		// Create candidate signal
		candidate := CandidateSignal{
			Type:      "blocked",
			Timestamp: action.Timestamp.Unix(),
			Route:     action.Route,
			Details: map[string]interface{}{
				"action_type":    action.Event.EventType,
				"rejection_type": getRejectionTypeRefined(rejection),
				"retry_count":    len(retries),
				"first_retry":    retries[0].Timestamp.Format(time.RFC3339),
			},
		}

		// Check for false alarms
		if isFalse, reason := d.falseAlarmPreventer.IsFalseAlarm(candidate, session, convertToEvents(classified)); isFalse {
			log.Printf("[Blocked Detection] False alarm prevented: %s", reason)
			continue
		}

		candidates = append(candidates, candidate)
	}

	return candidates
}

// isActionAttemptRefined checks if event is an action attempt (refined)
func isActionAttemptRefined(event ClassifiedEvent) bool {
	// Form submit is always an action attempt
	if event.Event.EventType == "form_submit" {
		return true
	}

	// Click on submit button
	if event.Event.EventType == "click" {
		if target := event.Event.Target; target.Type == "button" {
			// Check if it's a submit button (by selector, type, or ID)
			if target.Selector != "" {
				selectorLower := target.Selector
				if contains(selectorLower, "submit") || contains(selectorLower, "btn-submit") {
					return true
				}
			}
			if target.ID != "" {
				idLower := target.ID
				if contains(idLower, "submit") || contains(idLower, "btn-submit") {
					return true
				}
			}
		}
	}

	return false
}

// findSystemRejectionRefined finds system rejection after action (refined)
func findSystemRejectionRefined(classified []ClassifiedEvent, after time.Time, window time.Duration) *ClassifiedEvent {
	windowEnd := after.Add(window)
	for _, event := range classified {
		if event.Timestamp.After(after) && event.Timestamp.Before(windowEnd) {
			if event.Category == CategorySystemFeedback {
				// Check for error or rejection
				if isRejectionEventRefined(event) {
					return &event
				}
			}
		}
	}
	return nil
}

// Note: isRejectionEventRefined is now in helpers.go

// findRetriesRefined finds all retry attempts after rejection
func findRetriesRefined(classified []ClassifiedEvent, after time.Time, window time.Duration, originalAction ClassifiedEvent) []ClassifiedEvent {
	windowEnd := after.Add(window)
	retries := make([]ClassifiedEvent, 0)

	for _, event := range classified {
		if event.Timestamp.After(after) && event.Timestamp.Before(windowEnd) {
			// Check if it's the same type of action
			if event.Event.EventType == originalAction.Event.EventType {
				// Check if it's the same target
				if getTargetIDFromClassified(event) == getTargetIDFromClassified(originalAction) {
					retries = append(retries, event)
				}
			}
		}
	}

	return retries
}

// areRetriesReasonable checks if retries are within reasonable time windows
func areRetriesReasonable(retries []ClassifiedEvent) bool {
	if len(retries) < 2 {
		return true
	}

	for i := 0; i < len(retries)-1; i++ {
		timeDiff := retries[i+1].Timestamp.Sub(retries[i].Timestamp)
		if timeDiff > blockedMaxTimeBetweenRetries {
			// Retries too far apart, might not be frustration
			return false
		}
	}

	return true
}

// getRejectionTypeRefined gets the type of rejection (refined)
func getRejectionTypeRefined(event *ClassifiedEvent) string {
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

// Note: contains function is in helpers.go
