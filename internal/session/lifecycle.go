/**
 * Session Lifecycle Management
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Session lifecycle and event handling
 */

package session

import (
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// AddEvent adds an event to session
func (s *SessionState) AddEvent(event types.Event) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reject events if session is completed
	if s.State == types.SessionStateCompleted {
		return false // Late event, drop silently
	}

	// Validate session ID matches
	if event.SessionID != s.SessionID {
		return false // Wrong session, drop silently
	}

	// Check for session reset
	if CheckForSessionReset(event) {
		// Force completion
		s.State = types.SessionStateCompleted
		return false
	}

	// Update last activity
	s.LastActivity = time.Now()

	// Transition from idle to active if needed
	if s.State == types.SessionStateIdle {
		s.State = types.SessionStateActive
	}

	// Track route transitions
	if event.Route != s.CurrentRoute && s.CurrentRoute != "" {
		s.RouteTransitions = append(s.RouteTransitions, types.RouteTransition{
			From:      s.CurrentRoute,
			To:        event.Route,
			Timestamp: time.Now(),
		})
	}
	s.CurrentRoute = event.Route

	// Add event (will be sorted and deduplicated later)
	s.Events = append(s.Events, event)

	return true
}

// ProcessEvents processes events: sort, deduplicate, and order
func (s *SessionState) ProcessEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Sort events by timestamp
	s.Events = SortEventsByTimestamp(s.Events)

	// Handle timestamp conflicts
	s.Events = PreserveOrderWithTimestampConflicts(s.Events)

	// Deduplicate events
	s.Events = DeduplicateEvents(s.Events)
}

// UpdateState updates session state based on time
func (s *SessionState) UpdateState(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if should become idle
	if s.State == types.SessionStateActive && ShouldBecomeIdle(s, now) {
		s.State = types.SessionStateIdle
	}

	// Check if should complete
	if ShouldComplete(s, now) {
		s.State = types.SessionStateCompleted
	}
}
