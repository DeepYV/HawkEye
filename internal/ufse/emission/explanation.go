/**
 * Explainability Generation
 * 
 * Author: Henry Wilson (Team Beta)
 * Responsibility: Generate human-readable explanations
 * 
 * Every emitted incident MUST include:
 * - Which signals fired
 * - Why correlation passed
 * - Why confidence is High
 * - What failed first
 * 
 * If explanation cannot be produced → discard
 */

package emission

import (
	"fmt"
	"strings"

	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/types"
)

// GenerateExplanation generates explanation for incident
func GenerateExplanation(
	group correlation.CorrelatedGroup,
	failurePoint string,
	confidence string,
) (string, bool) {
	if len(group.Signals) == 0 {
		return "", false
	}

	var parts []string

	// Part 1: Which signals fired
	signalTypes := make([]string, 0, len(group.Signals))
	for _, signal := range group.Signals {
		signalTypes = append(signalTypes, signal.Type)
	}
	parts = append(parts, fmt.Sprintf("Signals detected: %s", strings.Join(signalTypes, ", ")))

	// Part 2: Why correlation passed
	correlationReason := fmt.Sprintf(
		"Correlation passed: %d signals detected within %v time window on route '%s'",
		len(group.Signals),
		group.TimeWindow,
		group.Route,
	)
	if group.HasSystemFeedback {
		correlationReason += " with system feedback"
	}
	parts = append(parts, correlationReason)

	// Part 3: Why confidence is High
	confidenceReason := "High confidence: "
	if len(group.Signals) >= 2 {
		confidenceReason += fmt.Sprintf("multiple signals (%d) ", len(group.Signals))
	}
	if group.HasSystemFeedback {
		confidenceReason += "with system feedback, "
	}
	confidenceReason += "clear failure point identified"
	parts = append(parts, confidenceReason)

	// Part 4: What failed first
	parts = append(parts, fmt.Sprintf("Primary failure point: %s", failurePoint))

	explanation := strings.Join(parts, ". ")
	return explanation, true
}

// CreateSignalDetails creates signal details for incident
func CreateSignalDetails(group correlation.CorrelatedGroup) []types.SignalDetail {
	details := make([]types.SignalDetail, 0, len(group.Signals))
	for _, signal := range group.Signals {
		detailStr := signal.Type
		if signal.Details != nil {
			if targetID, ok := signal.Details["target_id"].(string); ok {
				detailStr += " on " + targetID
			}
		}

		details = append(details, types.SignalDetail{
			Type:      signal.Type,
			Timestamp: signal.Timestamp,
			Route:     signal.Route,
			Details:   detailStr,
		})
	}
	return details
}