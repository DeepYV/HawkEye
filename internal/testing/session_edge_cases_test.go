/**
 * Session Management Edge Case Tests
 * 
 * Author: QA Engineer + All Engineers
 * Responsibility: Comprehensive testing of session management edge cases
 */

package testing

import (
	"testing"
	"time"

	"github.com/your-org/frustration-engine/internal/session"
	"github.com/your-org/frustration-engine/internal/types"
)

// TestSessionEdgeCases_LateEvents tests late event handling
func TestSessionEdgeCases_LateEvents(t *testing.T) {
	manager := session.NewManager()
	projectID := "test-project"
	sessionID := "test-session"

	// Create session and add initial events
	events := []types.Event{
		{
			EventType: "click",
			SessionID: sessionID,
			Timestamp: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			Route:     "/test",
			Target:    types.EventTarget{Type: "button", ID: "btn1"},
		},
	}
	manager.AddEvents(projectID, sessionID, events)

	// Force session to complete
	s, exists := manager.Get(sessionID)
	if !exists {
		t.Fatal("Session not found")
	}
	s.Transition(types.SessionStateCompleted)
	
	// Wait a bit for state to settle
	time.Sleep(100 * time.Millisecond)

	// Try to add late event (within tolerance)
	lateEvent := types.Event{
		EventType: "click",
		SessionID: sessionID,
		Timestamp: time.Now().Add(-30 * time.Minute).Format(time.RFC3339), // 30 min ago
		Route:     "/test",
		Target:    types.EventTarget{Type: "button", ID: "btn2"},
	}
	manager.AddEvents(projectID, sessionID, []types.Event{lateEvent})

	// Event should be dropped (session is completed)
	if len(s.Events) > 1 {
		t.Errorf("Late event should be dropped, but found %d events", len(s.Events))
	}
}

// TestSessionEdgeCases_ClockSkew tests clock skew handling
func TestSessionEdgeCases_ClockSkew(t *testing.T) {
	handler := session.NewEdgeCaseHandler()
	projectID := "test-project"
	sessionID := "test-session"
	s := session.NewSessionState(sessionID, projectID)

	// Event with future timestamp (clock skew)
	futureEvent := types.Event{
		EventType: "click",
		SessionID: sessionID,
		Timestamp: time.Now().Add(10 * time.Minute).Format(time.RFC3339), // 10 min in future
		Route:     "/test",
		Target:    types.EventTarget{Type: "button", ID: "btn1"},
	}

	hasSkew, adjustedTime := handler.HandleClockSkew(s, futureEvent)
	if !hasSkew {
		t.Error("Expected clock skew to be detected")
	}
	if adjustedTime.IsZero() {
		t.Error("Expected adjusted time to be set")
	}
}

// TestSessionEdgeCases_OutOfOrder tests out-of-order event handling
func TestSessionEdgeCases_OutOfOrder(t *testing.T) {
	handler := session.NewEdgeCaseHandler()

	events := []types.Event{
		{
			EventType: "click",
			SessionID: "session1",
			Timestamp: time.Now().Add(5 * time.Second).Format(time.RFC3339),
			Route:     "/test",
			Target:    types.EventTarget{Type: "button", ID: "btn1"},
		},
		{
			EventType: "click",
			SessionID: "session1",
			Timestamp: time.Now().Format(time.RFC3339),
			Route:     "/test",
			Target:    types.EventTarget{Type: "button", ID: "btn2"},
		},
	}

	sorted := handler.HandleOutOfOrderEvents(events)
	if len(sorted) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(sorted))
	}

	// First event should have earlier timestamp
	t1, _ := time.Parse(time.RFC3339, sorted[0].Timestamp)
	t2, _ := time.Parse(time.RFC3339, sorted[1].Timestamp)
	if t1.After(t2) {
		t.Error("Events should be sorted by timestamp")
	}
}

// TestSessionEdgeCases_MemoryPressure tests memory pressure handling
func TestSessionEdgeCases_MemoryPressure(t *testing.T) {
	handler := session.NewEdgeCaseHandler()
	sessionID := "test-session"
	s := session.NewSessionState(sessionID, "test-project")

	// Add many events to trigger memory pressure
	for i := 0; i < 10001; i++ {
		s.Events = append(s.Events, types.Event{
			EventType: "click",
			SessionID: sessionID,
			Timestamp: time.Now().Format(time.RFC3339),
			Route:     "/test",
			Target:    types.EventTarget{Type: "button", ID: "btn1"},
		})
	}

	if !handler.CheckMemoryPressure(s) {
		t.Error("Expected memory pressure to be detected")
	}
}

// TestSessionEdgeCases_SessionCollision tests session collision handling
func TestSessionEdgeCases_SessionCollision(t *testing.T) {
	handler := session.NewEdgeCaseHandler()
	sessionID := "test-session"
	s := session.NewSessionState(sessionID, "project1")

	// Try to create session with different project ID
	shouldCreateNew := handler.HandleSessionCollision(s, "project2")
	if !shouldCreateNew {
		t.Error("Expected new session to be created on collision")
	}

	// Session should be forced to complete
	if s.State != types.SessionStateCompleted {
		t.Error("Expected session to be completed on collision")
	}
}

// TestSessionEdgeCases_ConcurrentUpdates tests concurrent update handling
func TestSessionEdgeCases_ConcurrentUpdates(t *testing.T) {
	manager := session.NewManager()
	projectID := "test-project"
	sessionID := "test-session"

	// Add events concurrently (simulated)
	events1 := []types.Event{
		{
			EventType: "click",
			SessionID: sessionID,
			Timestamp: time.Now().Format(time.RFC3339),
			Route:     "/test",
			Target:    types.EventTarget{Type: "button", ID: "btn1"},
		},
	}
	events2 := []types.Event{
		{
			EventType: "click",
			SessionID: sessionID,
			Timestamp: time.Now().Add(1 * time.Second).Format(time.RFC3339),
			Route:     "/test",
			Target:    types.EventTarget{Type: "button", ID: "btn2"},
		},
	}

	manager.AddEvents(projectID, sessionID, events1)
	manager.AddEvents(projectID, sessionID, events2)

	s, exists := manager.Get(sessionID)
	if !exists {
		t.Fatal("Session not found")
	}

	if len(s.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(s.Events))
	}
}
