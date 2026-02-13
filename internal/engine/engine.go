// Package engine contains the pure frustration detection logic.
//
// The engine processes completed sessions through a multi-stage pipeline:
//  1. Event classification
//  2. Candidate signal detection
//  3. Signal qualification
//  4. Signal correlation
//  5. Scoring & confidence evaluation
//  6. Incident emission
//
// This package has no I/O side effects — it takes a session and returns
// zero or more incidents. All persistence is handled by the caller.
package engine

import (
	"sort"
	"strconv"
	"time"

	"github.com/your-org/frustration-engine/internal/metrics"
	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/ufse/emission"
	"github.com/your-org/frustration-engine/internal/ufse/scoring"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
	oldtypes "github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/pkg/types"
)

// DetectFrustration processes a session and returns detected incidents.
// This is a pure function with no side effects beyond metric counters.
func DetectFrustration(session types.Session) []*types.Incident {
	start := time.Now()
	defer func() {
		metrics.ProcessingLatency.Observe(time.Since(start).Seconds())
	}()

	metrics.SessionsProcessed.Inc()

	// Convert to old types for compatibility with existing detectors
	oldSession := toOldSession(session)

	// Step 1: classify events
	classified := classifyEvents(oldSession.Events)
	if len(classified) == 0 {
		return nil
	}

	// Step 2: detect candidate signals
	candidates := signals.DetectCandidateSignals(classified, oldSession)
	if len(candidates) == 0 {
		return nil
	}

	for _, c := range candidates {
		metrics.SignalsDetected.WithLabelValues(c.Type).Inc()
	}

	// Step 3: qualify signals
	qualified := signals.QualifySignals(candidates, classified)
	if len(qualified) == 0 {
		discarded := len(candidates)
		if discarded > 0 {
			metrics.SignalsDiscarded.WithLabelValues("qualification_failed").Add(float64(discarded))
		}
		return nil
	}

	discarded := len(candidates) - len(qualified)
	if discarded > 0 {
		metrics.SignalsDiscarded.WithLabelValues("qualification_failed").Add(float64(discarded))
	}

	// Step 4: correlate signals
	groups := correlation.CorrelateSignals(qualified)
	if len(groups) == 0 {
		metrics.SignalsDiscarded.WithLabelValues("correlation_failed").Add(float64(len(qualified)))
		return nil
	}

	// Step 5–6: score and emit
	var incidents []*types.Incident
	for _, group := range groups {
		score := scoring.CalculateScore(group, oldSession.StartTime, oldSession.EndTime)
		severity := scoring.DetermineSeverityType(group)
		confidence := scoring.EvaluateConfidence(group)

		if !scoring.IsHighConfidence(confidence) {
			metrics.SignalsDiscarded.WithLabelValues("low_confidence").Add(float64(len(group.Signals)))
			continue
		}

		failurePoint, ok := scoring.DetermineFailurePoint(group)
		if !ok {
			metrics.SignalsDiscarded.WithLabelValues("ambiguous_failure_point").Add(float64(len(group.Signals)))
			continue
		}

		oldIncident, ok := emission.EmitIncident(
			session.SessionID,
			session.ProjectID,
			group,
			score,
			severity,
			failurePoint,
		)
		if !ok {
			metrics.SignalsDiscarded.WithLabelValues("explanation_failed").Add(float64(len(group.Signals)))
			continue
		}

		metrics.IncidentsDetected.Inc()
		incidents = append(incidents, fromOldIncident(oldIncident))
	}

	return incidents
}

// classifyEvents converts raw events into classified events for signal detection.
func classifyEvents(events []oldtypes.Event) []signals.ClassifiedEvent {
	classified := make([]signals.ClassifiedEvent, 0, len(events))
	for _, event := range events {
		ts, ok := parseEventTimestamp(event.Timestamp)
		if !ok {
			// Events without a parseable timestamp cannot be safely used in
			// temporal correlation logic and are skipped to reduce false positives.
			continue
		}
		category := classifyEventType(event.EventType, event.Metadata)
		classified = append(classified, signals.ClassifiedEvent{
			Event:     event,
			Category:  category,
			Timestamp: ts,
			Route:     event.Route,
		})
	}

	// Ensure deterministic chronological ordering for time-window based
	// detectors, even when ingestion batches arrive out of order.
	sort.SliceStable(classified, func(i, j int) bool {
		return classified[i].Timestamp.Before(classified[j].Timestamp)
	})

	return classified
}

func parseEventTimestamp(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}

	if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return ts, true
	}

	if millis, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return time.UnixMilli(millis).UTC(), true
	}

	return time.Time{}, false
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
		if metadata != nil {
			if status, ok := metadata["status"].(float64); ok && status >= 400 {
				return signals.CategorySystemFeedback
			}
			if _, ok := metadata["error"]; ok {
				return signals.CategorySystemFeedback
			}
		}
		return signals.CategoryInteraction
	}
}

// --- type conversion helpers ---

func toOldSession(s types.Session) oldtypes.Session {
	oldEvents := make([]oldtypes.Event, len(s.Events))
	for i, e := range s.Events {
		oldEvents[i] = oldtypes.Event{
			EventType:      e.EventType,
			Timestamp:      e.Timestamp,
			SessionID:      e.SessionID,
			Route:          e.Route,
			Target:         oldtypes.EventTarget(e.Target),
			Metadata:       e.Metadata,
			Environment:    e.Environment,
			IdempotencyKey: e.IdempotencyKey,
		}
	}

	oldTransitions := make([]oldtypes.RouteTransition, len(s.RouteTransitions))
	for i, rt := range s.RouteTransitions {
		oldTransitions[i] = oldtypes.RouteTransition{
			From:      rt.From,
			To:        rt.To,
			Timestamp: rt.Timestamp,
		}
	}

	return oldtypes.Session{
		SessionID:        s.SessionID,
		ProjectID:        s.ProjectID,
		State:            s.State,
		Events:           oldEvents,
		StartTime:        s.StartTime,
		EndTime:          s.EndTime,
		LastActivity:     s.LastActivity,
		RouteTransitions: oldTransitions,
		Metadata:         s.Metadata,
	}
}

func fromOldIncident(old *oldtypes.Incident) *types.Incident {
	details := make([]types.SignalDetail, len(old.SignalDetails))
	for i, d := range old.SignalDetails {
		details[i] = types.SignalDetail{
			Type:      d.Type,
			Timestamp: d.Timestamp,
			Route:     d.Route,
			Details:   d.Details,
		}
	}
	return &types.Incident{
		IncidentID:          old.IncidentID,
		SessionID:           old.SessionID,
		ProjectID:           old.ProjectID,
		FrustrationScore:    old.FrustrationScore,
		ConfidenceLevel:     old.ConfidenceLevel,
		TriggeringSignals:   old.TriggeringSignals,
		PrimaryFailurePoint: old.PrimaryFailurePoint,
		SeverityType:        old.SeverityType,
		Timestamp:           old.Timestamp,
		Explanation:         old.Explanation,
		SignalDetails:       details,
		Status:              old.Status,
		ConfidenceScore:     old.ConfidenceScore,
		Suppressed:          old.Suppressed,
		CreatedAt:           old.CreatedAt,
		UpdatedAt:           old.UpdatedAt,
	}
}
