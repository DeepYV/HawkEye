/**
 * Processing Pipeline
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Orchestrate processing pipeline in exact order
 *
 * Pipeline order (MUST NOT BE CHANGED):
 * 1. Event classification
 * 2. Candidate signal detection
 * 3. Signal qualification
 * 4. Signal correlation
 * 5. Scoring & confidence evaluation
 * 6. Incident emission (or discard)
 */

package ufse

import (
	"time"

	"github.com/your-org/frustration-engine/internal/observability"
	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/ufse/emission"
	"github.com/your-org/frustration-engine/internal/ufse/scoring"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// ProcessSession processes a session through the pipeline
func ProcessSession(session types.Session) []*types.Incident {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		observability.ProcessingDuration.Observe(duration)
	}()

	incidents := make([]*types.Incident, 0)

	// Track session processed
	observability.SessionsProcessed.Inc()

	// Step 1: Event classification
	classified := classifyEvents(session.Events)
	if len(classified) == 0 {
		return incidents // No events, no incidents
	}

	// Step 2: Candidate signal detection (using refined detectors)
	candidates := signals.DetectCandidateSignals(classified, session)
	if len(candidates) == 0 {
		return incidents // No candidates, no incidents
	}

	// Signals detected
	observability.SignalsDetected.Add(float64(len(candidates)))

	// Step 3: Signal qualification
	qualified := signals.QualifySignals(candidates, classified)
	if len(qualified) == 0 {
		// Track discarded signals
		observability.SignalsDiscarded.WithLabelValues("qualification_failed").Add(float64(len(candidates)))
		return incidents // No qualified signals, no incidents
	}

	// Track discarded (candidates - qualified)
	discardedCount := len(candidates) - len(qualified)
	if discardedCount > 0 {
		observability.SignalsDiscarded.WithLabelValues("qualification_failed").Add(float64(discardedCount))
	}

	// Step 4: Signal correlation
	correlatedGroups := correlation.CorrelateSignals(qualified)
	if len(correlatedGroups) == 0 {
		// Track discarded (no valid correlation)
		observability.SignalsDiscarded.WithLabelValues("correlation_failed").Add(float64(len(qualified)))
		return incidents // No valid correlations, no incidents
	}

	// Step 5: Scoring & confidence evaluation
	for _, group := range correlatedGroups {
		// Calculate frustration score
		frustrationScore := scoring.CalculateScore(group, session.StartTime, session.EndTime)

		// Determine severity type
		severityType := scoring.DetermineSeverityType(group)

		// Evaluate confidence
		confidence := scoring.EvaluateConfidence(group)

		// Only proceed if High confidence
		if !scoring.IsHighConfidence(confidence) {
			// Track discarded (low/medium confidence)
			observability.SignalsDiscarded.WithLabelValues("low_confidence").Add(float64(len(group.Signals)))
			continue // Discard if not High confidence
		}

		// Determine failure point
		failurePoint, ok := scoring.DetermineFailurePoint(group)
		if !ok {
			// Track discarded (ambiguous failure point)
			observability.SignalsDiscarded.WithLabelValues("ambiguous_failure_point").Add(float64(len(group.Signals)))
			continue // Cannot determine failure point → discard
		}

		// Step 6: Incident emission
		incident, ok := emission.EmitIncident(
			session.SessionID,
			session.ProjectID,
			group,
			frustrationScore,
			severityType,
			failurePoint,
		)
		if !ok {
			// Track discarded (explanation failed)
			observability.SignalsDiscarded.WithLabelValues("explanation_failed").Add(float64(len(group.Signals)))
			continue // Cannot emit (explanation failed) → discard
		}

		// Track incident emitted
		observability.IncidentsEmitted.Inc()
		incidents = append(incidents, incident)
	}

	return incidents
}

// classifyEvents classifies events into categories
func classifyEvents(events []types.Event) []signals.ClassifiedEvent {
	classified := make([]signals.ClassifiedEvent, 0, len(events))

	for _, event := range events {
		timestamp, _ := time.Parse(time.RFC3339, event.Timestamp)
		category := classifyEventType(event.EventType, event.Metadata)

		classified = append(classified, signals.ClassifiedEvent{
			Event:     event,
			Category:  category,
			Timestamp: timestamp,
			Route:     event.Route,
		})
	}

	return classified
}

func classifyEventType(eventType string, metadata map[string]interface{}) string {
	switch eventType {
	case "click", "input", "scroll", "form_submit":
		return signals.CategoryInteraction
	case "error", "network_error", "network_success", "slow_response":
		return signals.CategorySystemFeedback
	case "navigation", "route_change":
		return signals.CategoryNavigation
	case "long_task", "performance", "loading":
		return signals.CategoryPerformance
	default:
		// Check metadata for system feedback
		if metadata != nil {
			if status, ok := metadata["status"].(float64); ok {
				if status >= 400 {
					return signals.CategorySystemFeedback
				}
			}
			if _, ok := metadata["error"]; ok {
				return signals.CategorySystemFeedback
			}
		}
		return signals.CategoryInteraction
	}
}
