// Package memory provides an in-memory event store for development and testing.
package memory

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// storedEvent is an event with reception metadata.
type storedEvent struct {
	ID         string
	ProjectID  string
	Event      types.Event
	ReceivedAt time.Time
}

// EventStore stores events in memory. Implements storage.EventStore.
type EventStore struct {
	mu     sync.RWMutex
	events []storedEvent
}

// New creates a new in-memory event store.
func New() *EventStore {
	log.Println("[storage/memory] initialised in-memory event store")
	return &EventStore{events: make([]storedEvent, 0, 1024)}
}

// StoreEvents persists a batch of events.
func (s *EventStore) StoreEvents(ctx context.Context, projectID string, events []types.Event) error {
	if len(events) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, e := range events {
		s.events = append(s.events, storedEvent{
			ID:         e.IdempotencyKey,
			ProjectID:  projectID,
			Event:      e,
			ReceivedAt: now,
		})
	}
	log.Printf("[storage/memory] stored %d events for project %s (total: %d)", len(events), projectID, len(s.events))
	return nil
}

// Count returns the total number of stored events.
func (s *EventStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

// Close is a no-op for in-memory storage.
func (s *EventStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("[storage/memory] closing (%d events discarded)", len(s.events))
	s.events = nil
	return nil
}
