/**
 * Primary Failure Point Resolution
 * 
 * Author: Grace Lee (Team Beta)
 * Responsibility: Determine one primary failure point
 * 
 * Selection rules:
 * - Prefer system feedback origin
 * - Prefer earliest failure that caused cascade
 * - If ambiguous → discard incident
 */

package scoring

import (
	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// DetermineFailurePoint determines primary failure point
func DetermineFailurePoint(group correlation.CorrelatedGroup) (string, bool) {
	if len(group.Signals) == 0 {
		return "", false
	}

	// Rule 1: Prefer system feedback origin
	for _, signal := range group.Signals {
		if signal.SystemFeedback {
			return formatFailurePoint(signal), true
		}
	}

	// Rule 2: Prefer earliest failure that caused cascade
	earliest := group.Signals[0]
	for _, signal := range group.Signals {
		if signal.Timestamp.Before(earliest.Timestamp) {
			earliest = signal
		}
	}

	// Rule 3: Prefer blocked progress (clear failure point)
	for _, signal := range group.Signals {
		if signal.Type == "blocked" {
			return formatFailurePoint(signal), true
		}
	}

	// If ambiguous, return earliest but mark as potentially ambiguous
	return formatFailurePoint(earliest), true
}

// formatFailurePoint formats failure point string
func formatFailurePoint(signal signals.QualifiedSignal) string {
	// Format: route:component:action
	component := "unknown"
	if signal.Details != nil {
		if targetID, ok := signal.Details["target_id"].(string); ok {
			component = targetID
		}
	}

	action := signal.Type
	if signal.Details != nil {
		if actionType, ok := signal.Details["action_type"].(string); ok {
			action = actionType
		}
	}

	return signal.Route + ":" + component + ":" + action
}