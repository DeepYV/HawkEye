/**
 * Enhanced Frustration Scoring
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Calculate frustration score with signal strength weighting
 * 
 * Enhanced Features:
 * - Signal strength-based scoring
 * - Context-aware multipliers
 * - Rage bait bonus scoring
 * - Progressive score ranges
 */

package scoring

import (
	"time"

	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// Enhanced signal weights (includes rage_bait)
var enhancedSignalWeights = map[string]float64{
	"rage":       0.3,
	"rage_bait":  0.5, // Higher weight for rage bait
	"blocked":    0.4,
	"abandonment": 0.3,
	"confusion":   0.1,
}

// CalculateEnhancedScore calculates frustration score with enhanced logic
func CalculateEnhancedScore(group correlation.CorrelatedGroup, sessionStart, sessionEnd time.Time) int {
	score := 0.0

	// Factor 1: Signal count with strength weighting
	signalCountScore := calculateSignalCountScore(group)
	score += signalCountScore

	// Factor 2: Signal type weights with strength multipliers
	typeScore := calculateEnhancedTypeScore(group)
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

	// Factor 5: Rage bait bonus
	if hasRageBait(group) {
		score += 20.0 // Bonus for rage bait detection
	}

	// Factor 6: Signal strength multiplier
	strengthMultiplier := calculateStrengthMultiplier(group)
	score = score * strengthMultiplier

	// Normalize to 0-100
	if score > 100.0 {
		score = 100.0
	}

	return int(score)
}

// calculateSignalCountScore calculates score based on signal count and strength
func calculateSignalCountScore(group correlation.CorrelatedGroup) float64 {
	if len(group.Signals) == 0 {
		return 0.0
	}

	// Base score from count
	baseScore := float64(len(group.Signals)) * 10.0

	// Apply strength weighting
	strengthWeight := 0.0
	for _, signal := range group.Signals {
		strengthScore := getSignalStrengthScoreFromSignal(signal)
		strengthWeight += strengthScore
	}
	strengthWeight = strengthWeight / float64(len(group.Signals))

	// Weighted score
	weightedScore := baseScore * (0.7 + strengthWeight*0.3)

	if weightedScore > 50.0 {
		return 50.0
	}

	return weightedScore
}

// calculateEnhancedTypeScore calculates type score with strength multipliers
func calculateEnhancedTypeScore(group correlation.CorrelatedGroup) float64 {
	typeScore := 0.0

	for _, signal := range group.Signals {
		weight := enhancedSignalWeights[signal.Type]
		if weight == 0 {
			weight = 0.2 // Default weight
		}

		// Apply strength multiplier
		strengthScore := getSignalStrengthScoreFromSignal(signal)
		strengthMultiplier := 0.7 + strengthScore*0.3 // 0.7x to 1.0x multiplier

		signalScore := weight * 20.0 * strengthMultiplier
		typeScore += signalScore
	}

	if typeScore > 30.0 {
		typeScore = 30.0
	}

	return typeScore
}

// calculateStrengthMultiplier calculates overall strength multiplier for the group
func calculateStrengthMultiplier(group correlation.CorrelatedGroup) float64 {
	if len(group.Signals) == 0 {
		return 1.0
	}

	avgStrength := 0.0
	for _, signal := range group.Signals {
		avgStrength += getSignalStrengthScoreFromSignal(signal)
	}
	avgStrength = avgStrength / float64(len(group.Signals))

	// Multiplier: 0.8x to 1.2x based on strength
	return 0.8 + (avgStrength * 0.4)
}

// getSignalStrengthScoreFromSignal extracts strength score from signal
func getSignalStrengthScoreFromSignal(signal signals.QualifiedSignal) float64 {
	// Extract strength score from signal details
	if details := signal.Details; details != nil {
		if strengthScore, ok := details["strengthScore"].(float64); ok {
			return strengthScore
		}
		// Check for strength level
		if strength, ok := details["strength"].(string); ok {
			switch strength {
			case "high":
				return 0.9
			case "medium":
				return 0.6
			case "low":
				return 0.4
			}
		}
		// Check for dark pattern score (rage bait)
		if darkPatternScore, ok := details["darkPatternScore"].(float64); ok {
			return darkPatternScore
		}
	}
	
	// Default strength based on signal type
	switch signal.Type {
	case "rage_bait":
		return 0.9
	case "rage":
		return 0.6
	case "blocked":
		return 0.7
	case "abandonment":
		return 0.5
	case "confusion":
		return 0.3
	default:
		return 0.5
	}
}

// hasRageBait checks if group contains rage bait signals
func hasRageBait(group correlation.CorrelatedGroup) bool {
	for _, signal := range group.Signals {
		if signal.Type == "rage_bait" {
			return true
		}
	}
	return false
}

// DetermineEnhancedSeverityType determines severity type with rage bait support
func DetermineEnhancedSeverityType(group correlation.CorrelatedGroup) string {
	// Check for rage bait (dark pattern)
	if hasRageBait(group) {
		return "Dark Pattern"
	}

	// Check for system feedback signals (bugs)
	if group.HasSystemFeedback {
		// Check if it's a performance issue
		for _, signal := range group.Signals {
			if signal.Type == "confusion" {
				// Would need to check details for excessive scrolling
				return "Performance"
			}
		}
		return "Bug"
	}

	// Otherwise UX issue
	return "UX"
}
