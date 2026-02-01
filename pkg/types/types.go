// Package types provides shared domain types for HawkEye.
// These types form the contract between all internal modules and are SDK-compatible.
package types

import "time"

// Event represents a normalized event from the SDK.
type Event struct {
	EventType      string                 `json:"eventType"`
	Timestamp      string                 `json:"timestamp"`
	SessionID      string                 `json:"sessionId"`
	Route          string                 `json:"route"`
	Target         EventTarget            `json:"target"`
	Metadata       map[string]interface{} `json:"metadata"`
	Environment    string                 `json:"environment,omitempty"`
	IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
}

// EventTarget represents the target of an event.
type EventTarget struct {
	Type     string `json:"type"`
	ID       string `json:"id,omitempty"`
	Selector string `json:"selector,omitempty"`
	TagName  string `json:"tagName,omitempty"`
}

// IngestRequest represents an incoming batch request from the SDK.
type IngestRequest struct {
	APIKey     string  `json:"api_key"`
	SDKVersion string  `json:"sdk_version,omitempty"`
	AppID      string  `json:"app_id,omitempty"`
	Events     []Event `json:"events"`
}

// IngestResponse represents the API response for event ingestion.
type IngestResponse struct {
	Success   bool   `json:"success"`
	Processed int    `json:"processed,omitempty"`
	Message   string `json:"message,omitempty"`
}

// Session represents a completed user session.
type Session struct {
	SessionID        string                 `json:"sessionId"`
	ProjectID        string                 `json:"projectId"`
	State            string                 `json:"state"`
	Events           []Event                `json:"events"`
	StartTime        time.Time              `json:"startTime"`
	EndTime          time.Time              `json:"endTime"`
	LastActivity     time.Time              `json:"lastActivity"`
	RouteTransitions []RouteTransition      `json:"routeTransitions"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RouteTransition represents a route change during a session.
type RouteTransition struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionState constants.
const (
	SessionStateActive    = "active"
	SessionStateIdle      = "idle"
	SessionStateCompleted = "completed"
)

// Incident represents a detected frustration incident.
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
	Status              string         `json:"status"`
	ConfidenceScore     float64        `json:"confidenceScore"`
	Suppressed          bool           `json:"suppressed"`
	ExternalTicketID    string         `json:"externalTicketId,omitempty"`
	ExternalSystem      string         `json:"externalSystem,omitempty"`
	ExportedAt          *time.Time     `json:"exportedAt,omitempty"`
	ExportFailed        bool           `json:"exportFailed"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
}

// SignalDetail provides details about a detected signal.
type SignalDetail struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Route     string    `json:"route"`
	Details   string    `json:"details"`
}

// QueryRequest represents a query for incidents.
type QueryRequest struct {
	ProjectID     string  `json:"projectId,omitempty"`
	Status        string  `json:"status,omitempty"`
	MinConfidence float64 `json:"minConfidence,omitempty"`
	Suppressed    *bool   `json:"suppressed,omitempty"`
	Exported      *bool   `json:"exported,omitempty"`
	Limit         int     `json:"limit,omitempty"`
	Offset        int     `json:"offset,omitempty"`
}

// QueryResponse represents a query response.
type QueryResponse struct {
	Incidents []Incident `json:"incidents"`
	Total     int        `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// Filter is an alias for QueryRequest used by the IncidentStore interface.
type Filter = QueryRequest
