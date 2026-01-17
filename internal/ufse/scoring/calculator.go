/**
 * Frustration Scoring
 * 
 * Author: Grace Lee (Team Beta)
 * Responsibility: Calculate frustration score (0-100) deterministically
 * 
 * Scoring considers:
 * - Signal count
 * - Signal type weights
 * - Duration of struggle
 * - Presence of errors
 * - No randomness, no learning
 */

package scoring

import (
	"time"

	"github.com/your-org/frustration-engine/internal/ufse/correlation"
)

// Signal weights (deterministic)
var signalWeights = map[string]float64{
	"rage":       0.3,
	"blocked":    0.4,
	"abandonment": 0.3,
	"confusion":   0.1, // Low severity by default
}

// CalculateScore calculates frustration score (0-100)
func CalculateScore(group correlation.CorrelatedGroup, sessionStart, sessionEnd time.Time) int {
	score := 0.0

	// Factor 1: Signal count (more signals = higher score)
	signalCount := float64(len(group.Signals))
	score += signalCount * 10.0 // Max 50 points for 5+ signals
	if score > 50.0 {
		score = 50.0
	}

	// Factor 2: Signal type weights
	typeScore := 0.0
	for _, signal := range group.Signals {
		if weight, ok := signalWeights[signal.Type]; ok {
			typeScore += weight * 20.0
		}
	}
	if typeScore > 30.0 {
		typeScore = 30.0
	}
	score += typeScore

	// Factor 3: Duration of struggle
	duration := calculateStruggleDuration(group)
	durationScore := duration.Minutes() * 2.0 // 2 points per minute
	if durationScore > 10.0 {
		durationScore = 10.0
	}
	score += durationScore

	// Factor 4: Presence of errors (system feedback signals)
	if group.HasSystemFeedback {
		score += 10.0
	}

	// Normalize to 0-100
	if score > 100.0 {
		score = 100.0
	}

	return int(score)
}

// calculateStruggleDuration calculates duration of struggle
func calculateStruggleDuration(group correlation.CorrelatedGroup) time.Duration {
	if len(group.Signals) == 0 {
		return 0
	}

	earliest := group.Signals[0].Timestamp
	latest := group.Signals[0].Timestamp

	for _, signal := range group.Signals {
		if signal.Timestamp.Before(earliest) {
			earliest = signal.Timestamp
		}
		if signal.Timestamp.After(latest) {
			latest = signal.Timestamp
		}
	}

	return latest.Sub(earliest)
}

// DetermineSeverityType determines severity type
func DetermineSeverityType(group correlation.CorrelatedGroup) string {
	// Check for system feedback signals (bugs)
	if group.HasSystemFeedback {
		// Check if it's a performance issue
		for _, signal := range group.Signals {
			if signal.Type == "confusion" {
				if details := signal.Details; details != nil {
					if details["type"] == "excessive_scrolling" {
						return "Performance"
					}
				}
			}
		}
		return "Bug"
	}

	// Otherwise UX issue
	return "UX"
}