/**
 * Session Types
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Type definitions for UFSE input
 */

package types

import "time"

// Session represents a completed session from Session Manager
// Note: Event and EventTarget are defined in events.go
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

// RouteTransition represents a route change
type RouteTransition struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionState represents the state of a session
type SessionState string

const (
	SessionStateActive    SessionState = "active"
	SessionStateIdle      SessionState = "idle"
	SessionStateCompleted SessionState = "completed"
)
