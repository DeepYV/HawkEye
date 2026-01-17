/**
 * Session Manager
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Main session manager service
 */

package session

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// Manager manages all sessions
type Manager struct {
	storage        *Storage
	emissionChan   chan *types.Session
	edgeCaseHandler *EdgeCaseHandler
	mu             sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewManager creates a new session manager
func NewManager() *Manager {
	return &Manager{
		storage:         NewStorage(),
		emissionChan:    make(chan *types.Session, 1000),
		edgeCaseHandler: NewEdgeCaseHandler(),
		stopChan:        make(chan struct{}),
	}
}

// Start starts the session manager
func (m *Manager) Start(ctx context.Context) {
	// Start state update ticker
	m.wg.Add(1)
	go m.stateUpdateLoop(ctx)

	// Start cleanup ticker
	m.wg.Add(1)
	go m.cleanupLoop(ctx)

	// Start emission loop
	m.wg.Add(1)
	go m.emissionLoop(ctx)
}

// Stop stops the session manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

// AddEvents adds events to session with comprehensive edge case handling
func (m *Manager) AddEvents(projectID, sessionID string, events []types.Event) {
	if sessionID == "" {
		log.Printf("[Session Manager] Dropping events due to missing session ID. Project: %s, Events: %d", projectID, len(events))
		return
	}

	// Get or create session
	session := m.storage.GetOrCreate(sessionID, projectID)
	
	// Check for session collision (different project ID)
	if session.ProjectID != projectID && len(session.Events) > 0 {
		if m.edgeCaseHandler.HandleSessionCollision(session, projectID) {
			// Force complete old session and create new one
			session.Transition(types.SessionStateCompleted)
			session = m.storage.GetOrCreate(sessionID, projectID)
		}
	}

	wasNew := len(session.Events) == 0

	// Handle out-of-order events
	events = m.edgeCaseHandler.HandleOutOfOrderEvents(events)

	// Add events with edge case handling
	addedCount := 0
	for _, event := range events {
		// Validate event has session ID
		if event.SessionID == "" {
			log.Printf("[Session Manager] Dropping event due to missing event session ID. Project: %s, Event Type: %s", projectID, event.EventType)
			continue
		}

		// Validate session ID matches
		if event.SessionID != sessionID {
			log.Printf("[Session Manager] Dropping event due to session ID mismatch. Expected: %s, Got: %s, Event Type: %s", sessionID, event.SessionID, event.EventType)
			continue
		}

		// Handle late events (after session completion)
		if session.IsCompleted() {
			if !m.edgeCaseHandler.HandleLateEvent(session, event) {
				continue // Event too late, drop it
			}
		}

		// Handle clock skew
		hasClockSkew, adjustedTime := m.edgeCaseHandler.HandleClockSkew(session, event)
		if hasClockSkew {
			// Adjust timestamp
			event.Timestamp = adjustedTime.Format(time.RFC3339)
		} else if adjustedTime.IsZero() {
			// Very old event, drop it
			log.Printf("[Session Manager] Dropping very old event. Project: %s, Event Type: %s", projectID, event.EventType)
			continue
		}

		// Check memory pressure
		if m.edgeCaseHandler.CheckMemoryPressure(session) {
			log.Printf("[Session Manager] Forcing session completion due to memory pressure. Session: %s", sessionID)
			session.Transition(types.SessionStateCompleted)
			break
		}

		// Add event to session
		if session.AddEvent(event) {
			addedCount++
		}
	}

	// Process events (sort, deduplicate)
	session.ProcessEvents()

	// Log for debugging
	if wasNew {
		log.Printf("[Session Manager] Created new session %s (project: %s) with %d events", sessionID, projectID, addedCount)
	} else if addedCount > 0 {
		log.Printf("[Session Manager] Added %d events to session %s (total: %d)", addedCount, sessionID, len(session.Events))
	}
}

// GetEmissionChannel returns the channel for emitted sessions
func (m *Manager) GetEmissionChannel() <-chan *types.Session {
	return m.emissionChan
}

// Get gets session by ID (for testing)
func (m *Manager) Get(sessionID string) (*SessionState, bool) {
	return m.storage.Get(sessionID)
}

// stateUpdateLoop periodically updates session states
func (m *Manager) stateUpdateLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.updateAllSessionStates()
		}
	}
}

// updateAllSessionStates updates all session states
func (m *Manager) updateAllSessionStates() {
	// Get all session IDs first (with read lock)
	m.storage.mu.RLock()
	sessionIDs := make([]string, 0, len(m.storage.sessions))
	for sessionID := range m.storage.sessions {
		sessionIDs = append(sessionIDs, sessionID)
	}
	m.storage.mu.RUnlock()

	// Update each session (sessions have their own locks)
	now := time.Now()
	for _, sessionID := range sessionIDs {
		// Get session (with proper locking)
		session, exists := m.storage.Get(sessionID)
		if !exists {
			continue
		}
		// Update state (session has its own mutex)
		session.UpdateState(now)
	}
}

// cleanupLoop periodically cleans up old sessions
func (m *Manager) cleanupLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			// Cleanup sessions older than 24 hours
			m.storage.Cleanup(24 * time.Hour)
		}
	}
}

// emissionLoop emits completed sessions
func (m *Manager) emissionLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.emitCompletedSessions()
		}
	}
}

// emitCompletedSessions emits all completed sessions
func (m *Manager) emitCompletedSessions() {
	completed := m.storage.GetAllCompleted()

	for _, sessionState := range completed {
		if sessionState.CanEmit() {
			// Convert to session
			session := sessionState.ToSession()

			// Emit (non-blocking)
			select {
			case m.emissionChan <- session:
				// Emitted successfully, remove from storage
				log.Printf("[Session Manager] Emitting completed session %s (project: %s, events: %d)", 
					session.SessionID, session.ProjectID, len(session.Events))
				m.storage.Remove(sessionState.SessionID)
			default:
				// Channel full, skip (will retry next tick)
				log.Printf("[Session Manager] Emission channel full, will retry session %s", sessionState.SessionID)
			}
		}
	}
}
