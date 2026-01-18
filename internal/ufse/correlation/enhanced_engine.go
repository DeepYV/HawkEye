/**
 * Enhanced Signal Correlation Engine
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Correlate signals with support for single-signal high-strength cases
 * 
 * Enhanced Features:
 * - Single-signal correlation for high-strength signals
 * - Signal strength-based correlation
 * - Context-aware correlation windows
 * - Progressive correlation (High/Medium/Low)
 */

package correlation

import (
	"time"

	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

const (
	enhancedCorrelationTimeWindow = 30 * time.Second
	singleSignalStrengthThreshold = 0.8 // Signal strength score threshold for single-signal correlation
)

// EnhancedCorrelateSignals correlates signals with enhanced logic
func EnhancedCorrelateSignals(qualified []signals.QualifiedSignal) []CorrelatedGroup {
	if len(qualified) == 0 {
		return nil
	}

	groups := make([]CorrelatedGroup, 0)

	// Step 1: Check for high-strength single signals
	singleSignalGroups := correlateSingleSignals(qualified)
	groups = append(groups, singleSignalGroups...)

	// Step 2: Standard multi-signal correlation
	if len(qualified) >= 2 {
		multiSignalGroups := correlateMultiSignals(qualified)
		groups = append(groups, multiSignalGroups...)
	}

	// Step 3: Remove duplicates and merge overlapping groups
	mergedGroups := mergeOverlappingGroups(groups)

	return mergedGroups
}

// correlateSingleSignals correlates high-strength single signals
func correlateSingleSignals(qualified []signals.QualifiedSignal) []CorrelatedGroup {
	groups := make([]CorrelatedGroup, 0)

	for _, signal := range qualified {
		// Check if signal has high strength
		strengthScore := getSignalStrengthScore(signal)
		
		if strengthScore >= singleSignalStrengthThreshold {
			// Check if it's a high-strength signal type
			if isHighStrengthSignalType(signal) {
				// Create single-signal group
				group := CorrelatedGroup{
					Signals:        []signals.QualifiedSignal{signal},
					TimeWindow:     enhancedCorrelationTimeWindow,
					Route:          signal.Route,
					HasSystemFeedback: signal.SystemFeedback,
				}
				groups = append(groups, group)
			}
		}
	}

	return groups
}

// correlateMultiSignals performs standard multi-signal correlation
func correlateMultiSignals(qualified []signals.QualifiedSignal) []CorrelatedGroup {
	groups := make([]CorrelatedGroup, 0)

	// Group signals by route and time window
	routeGroups := groupByRoute(qualified)
	for _, routeSignals := range routeGroups {
		// Correlate within time windows
		correlated := correlateInTimeWindow(routeSignals)
		groups = append(groups, correlated...)
	}

	// Filter groups that meet requirements (relaxed for enhanced version)
	validGroups := make([]CorrelatedGroup, 0)
	for _, group := range groups {
		if isValidEnhancedCorrelation(group) {
			validGroups = append(validGroups, group)
		}
	}

	return validGroups
}

// isValidEnhancedCorrelation checks if correlation group meets enhanced requirements
func isValidEnhancedCorrelation(group CorrelatedGroup) bool {
	// Requirement 1: At least 1 signal (relaxed from 2)
	if len(group.Signals) < 1 {
		return false
	}

	// Requirement 2: For single signals, must be high-strength
	if len(group.Signals) == 1 {
		strengthScore := getSignalStrengthScore(group.Signals[0])
		if strengthScore < singleSignalStrengthThreshold {
			return false
		}
		// Single high-strength signals don't require system feedback
		group.HasSystemFeedback = group.Signals[0].SystemFeedback
		return true
	}

	// Requirement 3: For multi-signal, prefer system feedback but not required if signal strength is high
	if len(group.Signals) >= 2 {
		hasSystemFeedback := false
		avgStrengthScore := 0.0
		
		for _, signal := range group.Signals {
			if signal.SystemFeedback {
				hasSystemFeedback = true
			}
			avgStrengthScore += getSignalStrengthScore(signal)
		}
		avgStrengthScore = avgStrengthScore / float64(len(group.Signals))

		// If average strength is high, system feedback not required
		if avgStrengthScore >= 0.7 {
			group.HasSystemFeedback = hasSystemFeedback
			return true
		}

		// Otherwise, require system feedback
		if !hasSystemFeedback {
			return false
		}

		group.HasSystemFeedback = hasSystemFeedback
		return true
	}

	return false
}

// getSignalStrengthScore extracts signal strength score from signal details
func getSignalStrengthScore(signal signals.QualifiedSignal) float64 {
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
	}
	
	// Default strength based on signal type
	switch signal.Type {
	case "rage_bait":
		return 0.8 // Rage bait is high-strength by default
	case "rage":
		return 0.6 // Medium strength by default
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

// isHighStrengthSignalType checks if signal type is considered high-strength
func isHighStrengthSignalType(signal signals.QualifiedSignal) bool {
	// Rage bait is always high-strength
	if signal.Type == "rage_bait" {
		return true
	}
	
	// High-strength rage signals
	if signal.Type == "rage" {
		if details := signal.Details; details != nil {
			if strength, ok := details["strength"].(string); ok {
				return strength == "high"
			}
		}
	}
	
	return false
}

// mergeOverlappingGroups merges overlapping correlation groups
func mergeOverlappingGroups(groups []CorrelatedGroup) []CorrelatedGroup {
	if len(groups) == 0 {
		return groups
	}

	merged := make([]CorrelatedGroup, 0)
	used := make(map[int]bool)

	for i, group1 := range groups {
		if used[i] {
			continue
		}

		mergedGroup := group1
		used[i] = true

		// Find overlapping groups
		for j, group2 := range groups {
			if used[j] || i == j {
				continue
			}

			if areGroupsOverlapping(group1, group2) {
				// Merge groups
				mergedGroup.Signals = append(mergedGroup.Signals, group2.Signals...)
				used[j] = true
			}
		}

		merged = append(merged, mergedGroup)
	}

	return merged
}

// areGroupsOverlapping checks if two groups overlap
func areGroupsOverlapping(group1, group2 CorrelatedGroup) bool {
	// Same route
	if group1.Route != group2.Route {
		return false
	}

	// Check time overlap
	timeWindow1 := getGroupTimeWindow(group1)
	timeWindow2 := getGroupTimeWindow(group2)

	return timeWindowsOverlap(timeWindow1, timeWindow2)
}

// getGroupTimeWindow gets the time window for a group
func getGroupTimeWindow(group CorrelatedGroup) (time.Time, time.Time) {
	if len(group.Signals) == 0 {
		return time.Time{}, time.Time{}
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

	return earliest, latest.Add(group.TimeWindow)
}

// timeWindowsOverlap checks if two time windows overlap
func timeWindowsOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

// Helper function to get signal strength score (duplicated from enhanced_calculator for now)
func getSignalStrengthScore(signal signals.QualifiedSignal) float64 {
	if details := signal.Details; details != nil {
		if strengthScore, ok := details["strengthScore"].(float64); ok {
			return strengthScore
		}
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
		if darkPatternScore, ok := details["darkPatternScore"].(float64); ok {
			return darkPatternScore
		}
	}
	
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
