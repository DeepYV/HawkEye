/**
 * Incident Types
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Type definitions for incidents from Incident Store
 */

package types

import (
	"encoding/json"
	"time"
)

// Incident represents an incident from Incident Store
type Incident struct {
	IncidentID          string         `json:"incidentId"`
	SessionID           string         `json:"sessionId"`
	ProjectID           string         `json:"projectId"`
	FrustrationScore    int            `json:"frustrationScore"`
	ConfidenceLevel     string         `json:"confidenceLevel"`
	TriggeringSignals   []string       `json:"triggeringSignals"`
	PrimaryFailurePoint string         `json:"primaryFailurePoint"`
	SeverityType        string         `json:"severityType"`
	Timestamp           time.Time      `json:"timestamp"`
	Explanation         string         `json:"explanation"`
	SignalDetails       []SignalDetail `json:"signalDetails"`
	Status              string         `json:"status"`          // "confirmed", "draft", etc.
	ConfidenceScore     float64        `json:"confidenceScore"` // 0-100
	Suppressed          bool           `json:"suppressed"`
	ExternalTicketID    string         `json:"externalTicketId,omitempty"`
	ExternalSystem      string         `json:"externalSystem,omitempty"` // "jira" or "linear"
	ExportedAt          *time.Time     `json:"exportedAt,omitempty"`
	ExportFailed        bool           `json:"exportFailed"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
}

// SignalDetail provides details about a signal
type SignalDetail struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Route     string    `json:"route"`
	Details   string    `json:"details"`
}

// QueryRequest represents a query request for incidents
type QueryRequest struct {
	ProjectID     string  `json:"projectId,omitempty"`
	Status        string  `json:"status,omitempty"`
	MinConfidence float64 `json:"minConfidence,omitempty"`
	Suppressed    *bool   `json:"suppressed,omitempty"`
	Exported      *bool   `json:"exported,omitempty"`
	Limit         int     `json:"limit,omitempty"`
	Offset        int     `json:"offset,omitempty"`
}

// TriggeringSignalsToJSON converts signals to JSON
func TriggeringSignalsToJSON(signals []string) ([]byte, error) {
	return json.Marshal(signals)
}

// TriggeringSignalsFromJSON converts JSON to signals
func TriggeringSignalsFromJSON(data []byte) ([]string, error) {
	var signals []string
	err := json.Unmarshal(data, &signals)
	return signals, err
}

// SignalDetailsFromJSON converts JSON to SignalDetails
func SignalDetailsFromJSON(data []byte) ([]SignalDetail, error) {
	var details []SignalDetail
	err := json.Unmarshal(data, &details)
	return details, err
}

// IncidentRequest represents an incident ingestion request
type IncidentRequest struct {
	Incident Incident `json:"incident"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	Incidents []Incident `json:"incidents"`
	Total     int        `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}
