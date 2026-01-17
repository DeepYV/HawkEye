/**
 * Event Types and Structures
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Event type definitions
 */

package types

import "time"

// IngestRequest represents the incoming batch request from SDK
type IngestRequest struct {
	APIKey     string  `json:"api_key"`
	SDKVersion string  `json:"sdk_version,omitempty"`
	AppID      string  `json:"app_id,omitempty"`
	Events     []Event `json:"events"`
}

// Event represents a normalized event from the SDK
type Event struct {
	EventType string                 `json:"eventType"`
	Timestamp string                 `json:"timestamp"`
	SessionID string                 `json:"sessionId"`
	Route     string                 `json:"route"`
	Target    EventTarget            `json:"target"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// EventTarget represents the target of an event
type EventTarget struct {
	Type     string `json:"type"`
	ID       string `json:"id,omitempty"`
	Selector string `json:"selector,omitempty"`
	TagName  string `json:"tagName,omitempty"`
}

// IngestResponse represents the API response
type IngestResponse struct {
	Success   bool   `json:"success"`
	Processed int    `json:"processed,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Reason  string
	EventID int // Index in batch
}

// StoredEvent represents an event stored in ClickHouse
type StoredEvent struct {
	ID          string
	ProjectID   string
	APIKey      string
	Event       Event
	ReceivedAt  time.Time
	ProcessedAt time.Time
}
