/**
 * Session Storage
 *
 * Author: Grace Lee (Team Beta)
 * Responsibility: Session storage and memory management
 */

package session

import (
	"sync"
	"time"
)

// Storage manages session storage
type Storage struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
}

// NewStorage creates a new session storage
func NewStorage() *Storage {
	return &Storage{
		sessions: make(map[string]*SessionState),
	}
}

// GetOrCreate gets existing session or creates new one
func (s *Storage) GetOrCreate(sessionID, projectID string) *SessionState {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	if session, exists := s.sessions[sessionID]; exists {
		return session
	}

	// Create new session
	session := NewSessionState(sessionID, projectID)
	s.sessions[sessionID] = session
	return session
}

// Get gets session by ID
func (s *Storage) Get(sessionID string) (*SessionState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	return session, exists
}

// Remove removes session from storage
func (s *Storage) Remove(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

// GetAllCompleted gets all completed sessions
func (s *Storage) GetAllCompleted() []*SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	completed := make([]*SessionState, 0)
	for _, session := range s.sessions {
		if session.IsCompleted() {
			completed = append(completed, session)
		}
	}

	return completed
}

// Cleanup removes old completed sessions
func (s *Storage) Cleanup(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, session := range s.sessions {
		if session.IsCompleted() {
			// Remove if older than maxAge
			if now.Sub(session.StartTime) > maxAge {
				delete(s.sessions, sessionID)
			}
		}
	}
}

// Count returns the number of active sessions
func (s *Storage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
