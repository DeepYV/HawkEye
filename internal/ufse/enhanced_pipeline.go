/**
 * Enhanced Processing Pipeline
 *
 * Author: Enhanced Detection Team
 * Responsibility: Orchestrate enhanced processing pipeline with improved detection
 *
 * Enhanced Pipeline Features:
 * - Multi-tier rage detection
 * - Rage bait detection
 * - Enhanced correlation (supports single-signal)
 * - Progressive confidence (Medium+ incidents emitted)
 * - Signal strength-based scoring
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

// ProcessSessionEnhanced processes a session through the enhanced pipeline
func ProcessSessionEnhanced(session types.Session) []*types.Incident {
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

	// Step 2: Enhanced candidate signal detection
	candidates := signals.DetectCandidateSignals(classified, session)
	if len(candidates) == 0 {
		return incidents // No candidates, no incidents
	}

	// Signals detected
	observability.SignalsDetected.Add(float64(len(candidates)))
	
	// Record enhanced signal metrics
	for _, candidate := range candidates {
		strength := "medium" // default
		if details := candidate.Details; details != nil {
			if s, ok := details["strength"].(string); ok {
				strength = s
			}
		}
		RecordEnhancedSignal(candidate.Type, strength)
		
		// Record signal strength score
		if details := candidate.Details; details != nil {
			if score, ok := details["strengthScore"].(float64); ok {
				RecordSignalStrength(candidate.Type, score)
			}
			// Record dark pattern score for rage bait
			if candidate.Type == "rage_bait" {
				if score, ok := details["darkPatternScore"].(float64); ok {
					RecordDarkPatternScore(score)
					RecordRageBait()
				}
			}
		}
	}

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

	// Step 4: Enhanced signal correlation (supports single-signal)
	correlatedGroups := correlation.EnhancedCorrelateSignals(qualified)
	if len(correlatedGroups) == 0 {
		// Track discarded (no valid correlation)
		observability.SignalsDiscarded.WithLabelValues("correlation_failed").Add(float64(len(qualified)))
		return incidents // No valid correlations, no incidents
	}
	
	// Record single-signal incidents
	for _, group := range correlatedGroups {
		if len(group.Signals) == 1 {
			RecordSingleSignalIncident()
		}
	}

	// Step 5: Enhanced scoring & confidence evaluation
	for _, group := range correlatedGroups {
		// Calculate enhanced frustration score
		frustrationScore := scoring.CalculateEnhancedScore(group, session.StartTime, session.EndTime)

		// Determine enhanced severity type
		severityType := scoring.DetermineEnhancedSeverityType(group)

		// Evaluate enhanced confidence (supports Medium confidence)
		confidence := scoring.EvaluateEnhancedConfidence(group)

		// Proceed if Medium or High confidence (enhanced: emit Medium confidence)
		if !scoring.IsMediumOrHighConfidence(confidence) {
			// Track discarded (low confidence)
			observability.SignalsDiscarded.WithLabelValues("low_confidence").Add(float64(len(group.Signals)))
			continue // Discard if Low confidence
		}
		
		// Record Medium confidence incidents
		if confidence == scoring.ConfidenceMedium {
			RecordMediumConfidenceIncident()
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

		// Set confidence level
		incident.ConfidenceLevel = string(confidence)
		
		// For Medium confidence, set needs_review flag in explanation
		if confidence == scoring.ConfidenceMedium {
			incident.Explanation = "[NEEDS REVIEW - Medium Confidence] " + incident.Explanation
		}

		// Track incident emitted
		observability.IncidentsEmitted.Inc()
		incidents = append(incidents, incident)
	}

	return incidents
}
