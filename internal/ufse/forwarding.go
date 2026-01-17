/**
 * Incident Forwarding
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Forward incidents to Incident Store
 */

package ufse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// IncidentForwarder forwards incidents to Incident Store
type IncidentForwarder struct {
	incidentStoreURL string
	client          *http.Client
}

// NewIncidentForwarder creates a new incident forwarder
func NewIncidentForwarder(incidentStoreURL string) *IncidentForwarder {
	return &IncidentForwarder{
		incidentStoreURL: incidentStoreURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ForwardIncident forwards an incident to Incident Store
func (f *IncidentForwarder) ForwardIncident(ctx context.Context, incident *types.Incident) error {
	if f.incidentStoreURL == "" {
		// Log and skip if URL not configured
		log.Printf("[UFSE] Incident Store URL not configured, logging incident instead")
		log.Printf("[UFSE] Incident: %s (session: %s, score: %d, confidence: %s)", 
			incident.IncidentID, incident.SessionID, incident.FrustrationScore, incident.ConfidenceLevel)
		return nil
	}

	// Create incident payload matching Module 5 format
	incidentPayload := map[string]interface{}{
		"incident": map[string]interface{}{
			"incidentId":          incident.IncidentID,
			"sessionId":            incident.SessionID,
			"projectId":            incident.ProjectID,
			"frustrationScore":     incident.FrustrationScore,
			"confidenceLevel":      incident.ConfidenceLevel,
			"confidenceScore":      100.0, // High confidence = 100.0
			"triggeringSignals":    incident.TriggeringSignals,
			"primaryFailurePoint": incident.PrimaryFailurePoint,
			"severityType":         incident.SeverityType,
			"timestamp":            incident.Timestamp.Format(time.RFC3339),
			"explanation":          incident.Explanation,
			"signalDetails":       incident.SignalDetails,
			"status":               "draft", // Will be confirmed later
			"suppressed":           false,
		},
	}

	data, err := json.Marshal(incidentPayload)
	if err != nil {
		log.Printf("[UFSE] Failed to marshal incident: %v", err)
		return fmt.Errorf("failed to marshal incident: %w", err)
	}

	// POST to Incident Store
	url := f.incidentStoreURL + "/v1/incidents"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[UFSE] Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		log.Printf("[UFSE] Failed to forward incident %s: %v", incident.IncidentID, err)
		log.Printf("[UFSE] Incident logged: %s (session: %s, score: %d)", 
			incident.IncidentID, incident.SessionID, incident.FrustrationScore)
		return fmt.Errorf("failed to forward incident: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[UFSE] Incident Store returned status %d for incident %s", resp.StatusCode, incident.IncidentID)
		return fmt.Errorf("Incident Store returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if duplicate, ok := result["duplicate"].(bool); ok && duplicate {
			log.Printf("[UFSE] Incident %s is duplicate, skipped", incident.IncidentID)
		} else {
			log.Printf("[UFSE] Successfully forwarded incident %s to Incident Store", incident.IncidentID)
		}
	}

	return nil
}