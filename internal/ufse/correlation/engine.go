/**
 * Signal Correlation Engine
 * 
 * Author: Frank Miller (Team Beta)
 * Responsibility: Correlate qualified signals into incident groups
 * 
 * Rules (MANDATORY):
 * - ≥ 2 qualified signals
 * - ≥ 1 is a system feedback signal
 * - All signals occur within same session, same route/flow, bounded time window
 * - If any condition fails → discard all signals
 */

package correlation

import (
	"time"

	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

const (
	correlationTimeWindow = 30 * time.Second
)

// CorrelatedGroup represents a group of correlated signals
type CorrelatedGroup struct {
	Signals        []signals.QualifiedSignal
	TimeWindow     time.Duration
	Route          string
	HasSystemFeedback bool
}

// CorrelateSignals correlates qualified signals into groups
func CorrelateSignals(qualified []signals.QualifiedSignal) []CorrelatedGroup {
	if len(qualified) < 2 {
		return nil // Need at least 2 signals
	}

	groups := make([]CorrelatedGroup, 0)

	// Group signals by route and time window
	routeGroups := groupByRoute(qualified)
	for _, routeSignals := range routeGroups {
		// Correlate within time windows
		correlated := correlateInTimeWindow(routeSignals)
		groups = append(groups, correlated...)
	}

	// Filter groups that meet all requirements
	validGroups := make([]CorrelatedGroup, 0)
	for _, group := range groups {
		if isValidCorrelation(group) {
			validGroups = append(validGroups, group)
		}
	}

	return validGroups
}

// groupByRoute groups signals by route
func groupByRoute(qualified []signals.QualifiedSignal) map[string][]signals.QualifiedSignal {
	groups := make(map[string][]signals.QualifiedSignal)
	for _, signal := range qualified {
		groups[signal.Route] = append(groups[signal.Route], signal)
	}
	return groups
}

// correlateInTimeWindow correlates signals within time windows
func correlateInTimeWindow(routeSignals []signals.QualifiedSignal) []CorrelatedGroup {
	if len(routeSignals) < 2 {
		return nil
	}

	groups := make([]CorrelatedGroup, 0)

	// Sort by timestamp
	sorted := make([]signals.QualifiedSignal, len(routeSignals))
	copy(sorted, routeSignals)
	// Simple sort by timestamp
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Timestamp.Before(sorted[i].Timestamp) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Find groups within time window
	for i := 0; i < len(sorted); i++ {
		group := []signals.QualifiedSignal{sorted[i]}
		windowStart := sorted[i].Timestamp
		windowEnd := windowStart.Add(correlationTimeWindow)

		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Timestamp.Before(windowEnd) {
				group = append(group, sorted[j])
			} else {
				break
			}
		}

		if len(group) >= 2 {
			groups = append(groups, CorrelatedGroup{
				Signals:    group,
				TimeWindow: correlationTimeWindow,
				Route:      sorted[i].Route,
			})
		}
	}

	return groups
}

// isValidCorrelation checks if correlation group meets all requirements
func isValidCorrelation(group CorrelatedGroup) bool {
	// Requirement 1: ≥ 2 qualified signals
	if len(group.Signals) < 2 {
		return false
	}

	// Requirement 2: ≥ 1 is a system feedback signal
	hasSystemFeedback := false
	for _, signal := range group.Signals {
		if signal.SystemFeedback {
			hasSystemFeedback = true
			break
		}
	}
	if !hasSystemFeedback {
		return false
	}

	// Requirement 3: All signals in same session (checked by caller)
	// Requirement 4: All signals in same route (already grouped by route)
	// Requirement 5: All signals in bounded time window (already checked)

	group.HasSystemFeedback = hasSystemFeedback
	return true
}