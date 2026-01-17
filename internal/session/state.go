/**
 * Session State Machine
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Session state management
 */

package session

import (
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// StateMachine manages session state transitions
type StateMachine struct {
	mu sync.RWMutex
}

// SessionState represents internal session state
type SessionState struct {
	State            types.SessionState
	SessionID        string
	ProjectID        string
	Events           []types.Event
	StartTime        time.Time
	LastActivity     time.Time
	RouteTransitions []types.RouteTransition
	CurrentRoute     string
	mu               sync.RWMutex
}

// NewSessionState creates a new session state
func NewSessionState(sessionID, projectID string) *SessionState {
	now := time.Now()
	return &SessionState{
		State:            types.SessionStateActive,
		SessionID:        sessionID,
		ProjectID:        projectID,
		Events:           make([]types.Event, 0),
		StartTime:        now,
		LastActivity:     now,
		RouteTransitions: make([]types.RouteTransition, 0),
		CurrentRoute:     "",
	}
}

// Transition transitions session to new state
func (s *SessionState) Transition(newState types.SessionState) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate transition
	if !isValidTransition(s.State, newState) {
		return false
	}

	s.State = newState
	if newState == types.SessionStateCompleted {
		// Freeze session on completion
		s.LastActivity = time.Now()
	}

	return true
}

// isValidTransition checks if state transition is valid
func isValidTransition(from, to types.SessionState) bool {
	// State machine rules:
	// Active -> Idle (allowed)
	// Active -> Completed (allowed)
	// Idle -> Active (allowed, on new event)
	// Idle -> Completed (allowed)
	// Completed -> (no transitions, frozen)

	if from == types.SessionStateCompleted {
		return false // Completed sessions are frozen
	}

	if from == types.SessionStateActive && to == types.SessionStateIdle {
		return true
	}

	if from == types.SessionStateActive && to == types.SessionStateCompleted {
		return true
	}

	if from == types.SessionStateIdle && to == types.SessionStateActive {
		return true
	}

	if from == types.SessionStateIdle && to == types.SessionStateCompleted {
		return true
	}

	return false
}

// IsCompleted checks if session is completed
func (s *SessionState) IsCompleted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State == types.SessionStateCompleted
}

// IsActive checks if session is active
func (s *SessionState) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State == types.SessionStateActive
}
