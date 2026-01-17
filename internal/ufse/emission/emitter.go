/**
 * Incident Emission
 * 
 * Author: Henry Wilson (Team Beta)
 * Responsibility: Emit frustration incidents
 * 
 * Rules:
 * - Emit at most one incident per session per failure point
 * - Each incident includes all required fields
 * - Never emit partial incidents
 * - If explanation unclear → discard
 */

package emission

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/types"
)

// EmitIncident creates and emits an incident
func EmitIncident(
	sessionID string,
	projectID string,
	group correlation.CorrelatedGroup,
	frustrationScore int,
	severityType string,
	failurePoint string,
) (*types.Incident, bool) {
	// Generate explanation
	explanation, ok := GenerateExplanation(group, failurePoint, "High")
	if !ok {
		return nil, false // Cannot generate explanation → discard
	}

	// Create signal details
	signalDetails := CreateSignalDetails(group)

	// Get triggering signal types
	triggeringSignals := make([]string, 0, len(group.Signals))
	for _, signal := range group.Signals {
		triggeringSignals = append(triggeringSignals, signal.Type)
	}

	// Create incident
	incident := &types.Incident{
		IncidentID:       uuid.New().String(),
		SessionID:        sessionID,
		ProjectID:        projectID,
		FrustrationScore:  frustrationScore,
		ConfidenceLevel:   "High",
		TriggeringSignals: triggeringSignals,
		PrimaryFailurePoint: failurePoint,
		SeverityType:     severityType,
		Timestamp:        time.Now(),
		Explanation:      explanation,
		SignalDetails:    signalDetails,
	}

	return incident, true
}