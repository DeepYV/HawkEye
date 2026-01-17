/**
 * Session Emission
 *
 * Author: Frank Miller (Team Beta)
 * Responsibility: Session emission and serialization
 */

package session

import (
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// ToSession converts SessionState to Session for emission
func (s *SessionState) ToSession() *types.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Process events before emission
	// (sort, deduplicate, order)
	events := make([]types.Event, len(s.Events))
	copy(events, s.Events)

	// Sort and deduplicate
	events = SortEventsByTimestamp(events)
	events = DeduplicateEvents(events)

	// Determine end time
	endTime := s.LastActivity
	if len(events) > 0 {
		// Use last event timestamp if available
		if lastEventTime, err := time.Parse(time.RFC3339, events[len(events)-1].Timestamp); err == nil {
			if lastEventTime.After(endTime) {
				endTime = lastEventTime
			}
		}
	}

	return &types.Session{
		SessionID:        s.SessionID,
		ProjectID:        s.ProjectID,
		State:            string(s.State),
		Events:           events,
		StartTime:        s.StartTime,
		EndTime:          endTime,
		LastActivity:     s.LastActivity,
		RouteTransitions: s.RouteTransitions,
		Metadata: map[string]interface{}{
			"event_count": len(events),
		},
	}
}

// CanEmit checks if session can be emitted
func (s *SessionState) CanEmit() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Only completed sessions can be emitted
	return s.State == types.SessionStateCompleted && len(s.Events) > 0
}
