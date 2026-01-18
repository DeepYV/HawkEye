/**
 * Enhanced Confidence Evaluation
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Evaluate confidence with support for Medium confidence incidents
 * 
 * Enhanced Features:
 * - Progressive confidence levels (High/Medium/Low)
 * - Signal strength-based confidence
 * - Medium confidence incidents can be emitted
 */

package scoring

import (
	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// EvaluateEnhancedConfidence evaluates confidence with enhanced logic
func EvaluateEnhancedConfidence(group correlation.CorrelatedGroup) ConfidenceLevel {
	// Check signal count
	if len(group.Signals) == 0 {
		return ConfidenceLow
	}

	// Single signal case
	if len(group.Signals) == 1 {
		signal := group.Signals[0]
		strengthScore := getSignalStrengthScoreFromSignal(signal)
		
		// High-strength single signals can be Medium confidence
		if strengthScore >= 0.8 {
			// Check for clear failure point
			if hasClearFailurePointForSignal(signal) {
				return ConfidenceMedium
			}
		}
		
		return ConfidenceLow
	}

	// Multi-signal case
	// Check for strong correlation
	if !hasStrongCorrelation(group) {
		// Check if signals have high strength
		avgStrength := calculateAverageStrength(group)
		if avgStrength >= 0.7 {
			return ConfidenceMedium
		}
		return ConfidenceLow
	}

	// Check for clear failure point
	if !hasClearFailurePoint(group) {
		// Still Medium confidence if signal strength is high
		avgStrength := calculateAverageStrength(group)
		if avgStrength >= 0.7 {
			return ConfidenceMedium
		}
		return ConfidenceLow
	}

	// High confidence requires all conditions
	return ConfidenceHigh
}

// calculateAverageStrength calculates average strength score for a group
func calculateAverageStrength(group correlation.CorrelatedGroup) float64 {
	if len(group.Signals) == 0 {
		return 0.0
	}

	totalStrength := 0.0
	for _, signal := range group.Signals {
		totalStrength += getSignalStrengthScoreFromSignal(signal)
	}

	return totalStrength / float64(len(group.Signals))
}

// hasClearFailurePointForSignal checks if a single signal has a clear failure point
func hasClearFailurePointForSignal(signal signals.QualifiedSignal) bool {
	// System feedback signals indicate clear failure points
	if signal.SystemFeedback {
		return true
	}

	// Rage bait signals have clear failure points
	if signal.Type == "rage_bait" {
		return true
	}

	// High-strength rage signals
	if signal.Type == "rage" {
		if details := signal.Details; details != nil {
			if strength, ok := details["strength"].(string); ok {
				if strength == "high" {
					return true
				}
			}
		}
	}

	// Blocked signals have clear failure points
	if signal.Type == "blocked" {
		return true
	}

	return false
}

// IsMediumOrHighConfidence checks if confidence is Medium or High
func IsMediumOrHighConfidence(confidence ConfidenceLevel) bool {
	return confidence == ConfidenceMedium || confidence == ConfidenceHigh
}
