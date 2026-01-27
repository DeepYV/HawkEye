/**
 * Form Submission Loop Detection
 *
 * Author: Enhanced Detection Team
 * Responsibility: Detect form submission loops and repeated failed actions
 *
 * Pattern: User repeatedly submits the same form or performs the same action
 * without success, indicating UI confusion or persistent errors.
 *
 * Types of loops:
 * 1. Rapid form resubmission (same form, short intervals)
 * 2. Form submission with no response (UI doesn't indicate submission)
 * 3. Form submission with stale data (resubmitting unchanged values)
 */

package signals

import (
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Form loop thresholds
	formLoopMinSubmissions  = 3                  // Minimum form submissions to trigger
	formLoopTimeWindow      = 30 * time.Second   // Time window for detecting loops
	formLoopMaxTimeBetween  = 5 * time.Second    // Max time between submissions in a loop
	formLoopMinRapidCount   = 4                  // Min rapid submissions (within 10s)
	formLoopRapidWindow     = 10 * time.Second   // Window for rapid submissions
	formLoopRapidMaxBetween = 2 * time.Second    // Max time between rapid submissions
)

// FormLoopDetector detects form submission loop patterns
type FormLoopDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewFormLoopDetector creates a new form loop detector
func NewFormLoopDetector() *FormLoopDetector {
	return &FormLoopDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectFormLoops detects form submission loop signals
func (d *FormLoopDetector) DetectFormLoops(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Group form submissions by target (same form)
	formSubmissions := d.groupFormSubmissions(classified)

	for targetKey, submissions := range formSubmissions {
		if len(submissions) < formLoopMinSubmissions {
			continue
		}

		// Check for different loop patterns
		if loopCandidate := d.detectRapidLoop(submissions, targetKey); loopCandidate != nil {
			if isFalse, _ := d.falseAlarmPreventer.IsFalseAlarm(*loopCandidate, session, convertToEvents(classified)); !isFalse {
				candidates = append(candidates, *loopCandidate)
			}
			continue // Don't double-detect
		}

		if loopCandidate := d.detectFrustratedLoop(submissions, classified, targetKey); loopCandidate != nil {
			if isFalse, _ := d.falseAlarmPreventer.IsFalseAlarm(*loopCandidate, session, convertToEvents(classified)); !isFalse {
				candidates = append(candidates, *loopCandidate)
			}
		}
	}

	return candidates
}

// groupFormSubmissions groups form submissions by target
func (d *FormLoopDetector) groupFormSubmissions(classified []ClassifiedEvent) map[string][]ClassifiedEvent {
	groups := make(map[string][]ClassifiedEvent)

	for _, event := range classified {
		if !isFormSubmission(event) {
			continue
		}

		targetKey := getFormTargetKey(event)
		groups[targetKey] = append(groups[targetKey], event)
	}

	return groups
}

// detectRapidLoop detects rapid form submissions (user clicking submit repeatedly)
func (d *FormLoopDetector) detectRapidLoop(submissions []ClassifiedEvent, targetKey string) *CandidateSignal {
	if len(submissions) < formLoopMinRapidCount {
		return nil
	}

	// Find sequences of rapid submissions
	for i := 0; i <= len(submissions)-formLoopMinRapidCount; i++ {
		windowEnd := submissions[i].Timestamp.Add(formLoopRapidWindow)

		// Count submissions in rapid window
		rapidCount := 0
		lastTimestamp := submissions[i].Timestamp
		allWithinThreshold := true

		for j := i; j < len(submissions) && submissions[j].Timestamp.Before(windowEnd); j++ {
			if j > i {
				timeDiff := submissions[j].Timestamp.Sub(lastTimestamp)
				if timeDiff > formLoopRapidMaxBetween {
					allWithinThreshold = false
					break
				}
			}
			lastTimestamp = submissions[j].Timestamp
			rapidCount++
		}

		if rapidCount >= formLoopMinRapidCount && allWithinThreshold {
			return &CandidateSignal{
				Type:      "form_loop",
				Timestamp: submissions[i].Timestamp.Unix(),
				Route:     submissions[i].Route,
				Details: map[string]interface{}{
					"loop_type":        "rapid_submission",
					"target_key":       targetKey,
					"submission_count": rapidCount,
					"window_seconds":   formLoopRapidWindow.Seconds(),
					"signal_strength":  calculateFormLoopStrength(rapidCount, true),
				},
			}
		}
	}

	return nil
}

// detectFrustratedLoop detects frustrated form submission patterns
// (user keeps trying but with no visible success)
func (d *FormLoopDetector) detectFrustratedLoop(submissions []ClassifiedEvent, allEvents []ClassifiedEvent, targetKey string) *CandidateSignal {
	if len(submissions) < formLoopMinSubmissions {
		return nil
	}

	// Find sequences within the time window
	for i := 0; i <= len(submissions)-formLoopMinSubmissions; i++ {
		windowEnd := submissions[i].Timestamp.Add(formLoopTimeWindow)

		// Count submissions and check for success
		loopCount := 0
		lastSubmission := submissions[i]
		allReasonablySpaced := true
		hasSuccessResponse := false

		for j := i; j < len(submissions) && submissions[j].Timestamp.Before(windowEnd); j++ {
			if j > i {
				timeDiff := submissions[j].Timestamp.Sub(lastSubmission.Timestamp)
				if timeDiff > formLoopMaxTimeBetween {
					allReasonablySpaced = false
					break
				}
			}

			// Check for success between submissions
			if j > i {
				hasSuccessResponse = hasSuccessResponse || hasSuccessBetween(allEvents, lastSubmission.Timestamp, submissions[j].Timestamp)
			}

			lastSubmission = submissions[j]
			loopCount++
		}

		// Only report if no success and reasonable spacing
		if loopCount >= formLoopMinSubmissions && allReasonablySpaced && !hasSuccessResponse {
			return &CandidateSignal{
				Type:      "form_loop",
				Timestamp: submissions[i].Timestamp.Unix(),
				Route:     submissions[i].Route,
				Details: map[string]interface{}{
					"loop_type":        "frustrated_resubmission",
					"target_key":       targetKey,
					"submission_count": loopCount,
					"window_seconds":   formLoopTimeWindow.Seconds(),
					"signal_strength":  calculateFormLoopStrength(loopCount, false),
				},
			}
		}
	}

	return nil
}

// isFormSubmission checks if event is a form submission
func isFormSubmission(event ClassifiedEvent) bool {
	return event.Event.EventType == "form_submit"
}

// getFormTargetKey creates a unique key for the form target
func getFormTargetKey(event ClassifiedEvent) string {
	target := event.Event.Target

	// Prefer ID, then selector, then route
	if target.ID != "" {
		return event.Route + ":" + target.ID
	}
	if target.Selector != "" {
		return event.Route + ":" + target.Selector
	}
	return event.Route + ":unknown_form"
}

// hasSuccessBetween checks if there was a success response between two timestamps
func hasSuccessBetween(events []ClassifiedEvent, start, end time.Time) bool {
	for _, event := range events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			// Check for success indicators
			if event.Event.EventType == "network" {
				if metadata := event.Event.Metadata; metadata != nil {
					if status, ok := metadata["status"].(float64); ok {
						if status >= 200 && status < 300 {
							return true
						}
					}
				}
			}

			// Navigation away from form could indicate success
			if event.Event.EventType == "navigation" {
				if metadata := event.Event.Metadata; metadata != nil {
					if newRoute, ok := metadata["to"].(string); ok {
						// Success pages typically contain these terms
						if contains(newRoute, "success") || contains(newRoute, "confirm") ||
							contains(newRoute, "thank") || contains(newRoute, "complete") {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// calculateFormLoopStrength calculates signal strength for form loops
func calculateFormLoopStrength(submissionCount int, isRapid bool) float64 {
	baseStrength := float64(submissionCount) / 10.0 // More submissions = higher strength

	if isRapid {
		baseStrength *= 1.5 // Rapid submissions are more indicative of frustration
	}

	// Cap at 1.0
	if baseStrength > 1.0 {
		return 1.0
	}
	return baseStrength
}
