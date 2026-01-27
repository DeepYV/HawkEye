/**
 * Session Thresholds for Incident Creation
 *
 * Author: Enhanced Detection Team
 * Responsibility: Enforce minimum session thresholds before incident creation
 *
 * Rationale:
 * - Very short sessions may not provide enough context
 * - Single-event sessions are more likely to be bots or accidental visits
 * - Minimum thresholds help filter out noise
 *
 * Thresholds:
 * - MinSessionDuration: Minimum time a session must last
 * - MinEventCount: Minimum number of events in a session
 * - MinInteractionCount: Minimum user interactions (clicks, inputs, etc.)
 */

package scoring

import (
	"time"
)

const (
	// DefaultMinSessionDuration is the minimum session duration for incident creation
	DefaultMinSessionDuration = 5 * time.Second

	// DefaultMinEventCount is the minimum number of events in a session
	DefaultMinEventCount = 3

	// DefaultMinInteractionCount is the minimum user interactions
	DefaultMinInteractionCount = 2

	// ShortSessionPenalty is the penalty multiplier for short sessions
	ShortSessionPenalty = 0.5
)

// SessionThresholdConfig holds configuration for session thresholds
type SessionThresholdConfig struct {
	MinDuration        time.Duration
	MinEventCount      int
	MinInteractionCount int
	ShortSessionPenalty float64
	Enabled            bool
}

// DefaultSessionThresholdConfig returns default session threshold configuration
func DefaultSessionThresholdConfig() SessionThresholdConfig {
	return SessionThresholdConfig{
		MinDuration:         DefaultMinSessionDuration,
		MinEventCount:       DefaultMinEventCount,
		MinInteractionCount: DefaultMinInteractionCount,
		ShortSessionPenalty: ShortSessionPenalty,
		Enabled:             true,
	}
}

// SessionMetrics holds metrics about a session for threshold checking
type SessionMetrics struct {
	Duration          time.Duration
	EventCount        int
	InteractionCount  int
	NavigationCount   int
	ErrorCount        int
	FirstEventTime    time.Time
	LastEventTime     time.Time
}

// SessionThresholdChecker checks if a session meets minimum thresholds
type SessionThresholdChecker struct {
	config SessionThresholdConfig
}

// NewSessionThresholdChecker creates a new session threshold checker
func NewSessionThresholdChecker() *SessionThresholdChecker {
	return &SessionThresholdChecker{
		config: DefaultSessionThresholdConfig(),
	}
}

// NewSessionThresholdCheckerWithConfig creates a checker with custom config
func NewSessionThresholdCheckerWithConfig(config SessionThresholdConfig) *SessionThresholdChecker {
	return &SessionThresholdChecker{
		config: config,
	}
}

// ThresholdResult contains the result of threshold checking
type ThresholdResult struct {
	MeetsThreshold bool
	Reason         string
	ScoreModifier  float64 // Multiplier to apply to score (1.0 = no change)
	Metrics        SessionMetrics
}

// CheckThresholds checks if session metrics meet minimum thresholds
func (c *SessionThresholdChecker) CheckThresholds(metrics SessionMetrics) ThresholdResult {
	if !c.config.Enabled {
		return ThresholdResult{
			MeetsThreshold: true,
			Reason:         "threshold checking disabled",
			ScoreModifier:  1.0,
			Metrics:        metrics,
		}
	}

	// Check duration
	if metrics.Duration < c.config.MinDuration {
		return ThresholdResult{
			MeetsThreshold: false,
			Reason:         "session duration below minimum threshold",
			ScoreModifier:  c.config.ShortSessionPenalty,
			Metrics:        metrics,
		}
	}

	// Check event count
	if metrics.EventCount < c.config.MinEventCount {
		return ThresholdResult{
			MeetsThreshold: false,
			Reason:         "event count below minimum threshold",
			ScoreModifier:  c.config.ShortSessionPenalty,
			Metrics:        metrics,
		}
	}

	// Check interaction count
	if metrics.InteractionCount < c.config.MinInteractionCount {
		return ThresholdResult{
			MeetsThreshold: false,
			Reason:         "interaction count below minimum threshold",
			ScoreModifier:  c.config.ShortSessionPenalty,
			Metrics:        metrics,
		}
	}

	return ThresholdResult{
		MeetsThreshold: true,
		Reason:         "all thresholds met",
		ScoreModifier:  1.0,
		Metrics:        metrics,
	}
}

// CalculateModifiedScore applies threshold-based modification to score
func (c *SessionThresholdChecker) CalculateModifiedScore(baseScore int, metrics SessionMetrics) int {
	result := c.CheckThresholds(metrics)

	if result.MeetsThreshold {
		return baseScore
	}

	// Apply penalty
	modifiedScore := float64(baseScore) * result.ScoreModifier

	return int(modifiedScore)
}

// IsInteractionEvent checks if an event type is a user interaction
func IsInteractionEvent(eventType string) bool {
	interactionTypes := map[string]bool{
		"click":       true,
		"input":       true,
		"form_submit": true,
		"scroll":      true,
	}
	return interactionTypes[eventType]
}

// CalculateSessionMetrics calculates metrics from session events
func CalculateSessionMetrics(events []EventForMetrics) SessionMetrics {
	if len(events) == 0 {
		return SessionMetrics{}
	}

	metrics := SessionMetrics{
		EventCount:       len(events),
		InteractionCount: 0,
		NavigationCount:  0,
		ErrorCount:       0,
	}

	for i, event := range events {
		if i == 0 {
			metrics.FirstEventTime = event.Timestamp
		}
		metrics.LastEventTime = event.Timestamp

		if IsInteractionEvent(event.EventType) {
			metrics.InteractionCount++
		}

		if event.EventType == "navigation" {
			metrics.NavigationCount++
		}

		if event.EventType == "error" || event.EventType == "network" {
			// Check if it's an error response
			if event.IsError {
				metrics.ErrorCount++
			}
		}
	}

	metrics.Duration = metrics.LastEventTime.Sub(metrics.FirstEventTime)

	return metrics
}

// EventForMetrics represents an event with minimal data for metrics calculation
type EventForMetrics struct {
	EventType string
	Timestamp time.Time
	IsError   bool
}
