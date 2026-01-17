/**
 * Confidence Evaluation
 * 
 * Author: Grace Lee (Team Beta)
 * Responsibility: Evaluate confidence level (Low/Medium/High)
 * 
 * Rules:
 * - Low → discard
 * - Medium → discard
 * - High → emit
 * - High confidence requires:
 *   - Strong signal correlation
 *   - Clear failure point
 *   - No successful resolution later
 */

package scoring

import (
	"github.com/your-org/frustration-engine/internal/ufse/correlation"
)

// ConfidenceLevel represents confidence level
type ConfidenceLevel string

const (
	ConfidenceLow    ConfidenceLevel = "Low"
	ConfidenceMedium ConfidenceLevel = "Medium"
	ConfidenceHigh   ConfidenceLevel = "High"
)

// EvaluateConfidence evaluates confidence level for correlated group
func EvaluateConfidence(group correlation.CorrelatedGroup) ConfidenceLevel {
	// Check signal count (more signals = higher confidence)
	if len(group.Signals) < 2 {
		return ConfidenceLow
	}

	// Check for strong correlation
	if !hasStrongCorrelation(group) {
		return ConfidenceLow
	}

	// Check for clear failure point
	if !hasClearFailurePoint(group) {
		return ConfidenceMedium
	}

	// High confidence requires all conditions
	return ConfidenceHigh
}

// hasStrongCorrelation checks for strong signal correlation
func hasStrongCorrelation(group correlation.CorrelatedGroup) bool {
	// At least 2 signals
	if len(group.Signals) < 2 {
		return false
	}

	// At least one system feedback signal
	if !group.HasSystemFeedback {
		return false
	}

	// Multiple different signal types (stronger correlation)
	signalTypes := make(map[string]bool)
	for _, signal := range group.Signals {
		signalTypes[signal.Type] = true
	}

	// If we have multiple types, correlation is stronger
	return len(signalTypes) >= 2
}

// hasClearFailurePoint checks if there's a clear failure point
func hasClearFailurePoint(group correlation.CorrelatedGroup) bool {
	// System feedback signals indicate clear failure points
	if group.HasSystemFeedback {
		return true
	}

	// Blocked progress signals have clear failure points
	for _, signal := range group.Signals {
		if signal.Type == "blocked" {
			return true
		}
	}

	return false
}

// IsHighConfidence checks if confidence is High
func IsHighConfidence(confidence ConfidenceLevel) bool {
	return confidence == ConfidenceHigh
}