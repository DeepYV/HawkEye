/**
 * Session Management Edge Cases
 * 
 * Author: Team Beta (Frank, Grace, Henry)
 * Responsibility: Handle edge cases in session management
 * 
 * Covers: late events, clock skew, concurrent updates, memory pressure
 */

package session

import (
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Late event tolerance - events arriving after session completion
	LateEventTolerance = 1 * time.Hour

	// Clock skew tolerance - maximum acceptable clock difference
	ClockSkewTolerance = 5 * time.Minute

	// Maximum events per session before forcing completion
	MaxEventsPerSession = 10000

	// Maximum session duration before forcing completion
	MaxSessionDurationHard = 24 * time.Hour
)

// EdgeCaseHandler handles edge cases in session management
type EdgeCaseHandler struct {
	mu sync.RWMutex
}

// NewEdgeCaseHandler creates a new edge case handler
func NewEdgeCaseHandler() *EdgeCaseHandler {
	return &EdgeCaseHandler{}
}

// HandleLateEvent handles events arriving after session completion
func (h *EdgeCaseHandler) HandleLateEvent(session *SessionState, event types.Event) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Check if session is completed
	if session.State != types.SessionStateCompleted {
		return false // Not a late event
	}

	// Parse event timestamp
	eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		log.Printf("[Session EdgeCase] Invalid timestamp in late event: %v", err)
		return false
	}

	// Use LastActivity as session end time
	sessionEndTime := session.LastActivity

	timeSinceCompletion := eventTime.Sub(sessionEndTime)
	if timeSinceCompletion > LateEventTolerance {
		log.Printf("[Session EdgeCase] Event too late: %v after completion, dropping", timeSinceCompletion)
		return false // Drop event
	}

	// Event is within tolerance but session is completed
	// Log for analysis but don't add to session
	log.Printf("[Session EdgeCase] Late event within tolerance: %v after completion, logging only", timeSinceCompletion)
	return false // Don't add to completed session
}

// HandleClockSkew handles clock skew between services
func (h *EdgeCaseHandler) HandleClockSkew(session *SessionState, event types.Event) (bool, time.Time) {
	// Parse event timestamp
	eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		return false, time.Time{}
	}

	now := time.Now()

	// Check for future timestamps (clock skew indicator)
	if eventTime.After(now.Add(ClockSkewTolerance)) {
		log.Printf("[Session EdgeCase] Future timestamp detected (clock skew): event=%v, now=%v, diff=%v",
			eventTime, now, eventTime.Sub(now))
		// Adjust to current time
		return true, now
	}

	// Check for very old timestamps
	if eventTime.Before(now.Add(-MaxPastTimestamp)) {
		log.Printf("[Session EdgeCase] Very old timestamp: event=%v, now=%v, diff=%v",
			eventTime, now, now.Sub(eventTime))
		return false, time.Time{} // Drop very old events
	}

	return false, eventTime
}

// HandleConcurrentUpdate handles concurrent session updates
func (h *EdgeCaseHandler) HandleConcurrentUpdate(session *SessionState, event types.Event) bool {
	// Session state has its own mutex, so concurrent updates are handled
	// This function can be used for additional validation or logging
	return true
}

// CheckMemoryPressure checks if session should be forced to complete due to memory pressure
func (h *EdgeCaseHandler) CheckMemoryPressure(session *SessionState) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Check event count
	if len(session.Events) > MaxEventsPerSession {
		log.Printf("[Session EdgeCase] Forcing completion due to event count: %d", len(session.Events))
		return true
	}

	// Check session duration
	sessionDuration := time.Since(session.StartTime)
	if sessionDuration > MaxSessionDurationHard {
		log.Printf("[Session EdgeCase] Forcing completion due to duration: %v", sessionDuration)
		return true
	}

	return false
}

// HandleOutOfOrderEvents handles events arriving out of order
func (h *EdgeCaseHandler) HandleOutOfOrderEvents(events []types.Event) []types.Event {
	// Sort events by timestamp
	sorted := make([]types.Event, len(events))
	copy(sorted, events)

	// Simple bubble sort (can be optimized)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			t1, err1 := time.Parse(time.RFC3339, sorted[i].Timestamp)
			t2, err2 := time.Parse(time.RFC3339, sorted[j].Timestamp)
			if err1 == nil && err2 == nil {
				if t2.Before(t1) {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}

	return sorted
}

// ValidateEventOrder validates event ordering within tolerance
func (h *EdgeCaseHandler) ValidateEventOrder(events []types.Event) bool {
	if len(events) < 2 {
		return true
	}

	for i := 0; i < len(events)-1; i++ {
		t1, err1 := time.Parse(time.RFC3339, events[i].Timestamp)
		t2, err2 := time.Parse(time.RFC3339, events[i+1].Timestamp)
		if err1 != nil || err2 != nil {
			continue // Skip invalid timestamps
		}

		// Allow small out-of-order (within 1 second) due to network delays
		if t2.Before(t1.Add(-1 * time.Second)) {
			log.Printf("[Session EdgeCase] Significant out-of-order event detected: %v before %v",
				events[i+1].Timestamp, events[i].Timestamp)
			return false
		}
	}

	return true
}

// HandleSessionCollision handles multiple sessions with same ID
func (h *EdgeCaseHandler) HandleSessionCollision(existing *SessionState, newProjectID string) bool {
	// If project ID differs, it's a collision
	if existing.ProjectID != newProjectID {
		log.Printf("[Session EdgeCase] Session ID collision detected: session=%s, existing_project=%s, new_project=%s",
			existing.SessionID, existing.ProjectID, newProjectID)
		// Force complete existing session
		existing.State = types.SessionStateCompleted
		return true // Create new session
	}

	return false // Same project, continue with existing session
}

// MaxPastTimestamp is the maximum age for events
const MaxPastTimestamp = 30 * 24 * time.Hour
