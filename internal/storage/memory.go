/**
 * In-Memory Event Storage
 *
 * Responsibility: Lightweight event persistence for local development.
 * Stores events in memory with no external dependencies.
 * Not intended for production use.
 */

package storage

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// MemoryEvent is a stored event with metadata.
type MemoryEvent struct {
	ID          string
	ProjectID   string
	Event       types.Event
	ReceivedAt  time.Time
}

// MemoryStorage stores events in memory for local development.
type MemoryStorage struct {
	mu     sync.RWMutex
	events []MemoryEvent
}

// NewMemoryStorage creates a new in-memory event storage.
func NewMemoryStorage() *MemoryStorage {
	log.Printf("[Memory Storage] Running in-memory event storage (development mode)")
	return &MemoryStorage{
		events: make([]MemoryEvent, 0, 1024),
	}
}

// StoreEvents stores events in memory.
func (m *MemoryStorage) StoreEvents(ctx context.Context, projectID string, events []types.Event) error {
	if len(events) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, event := range events {
		m.events = append(m.events, MemoryEvent{
			ID:         event.IdempotencyKey,
			ProjectID:  projectID,
			Event:      event,
			ReceivedAt: now,
		})
	}

	log.Printf("[Memory Storage] Stored %d events for project %s (total: %d)", len(events), projectID, len(m.events))
	return nil
}

// GetEvents returns all stored events (useful for debugging/testing).
func (m *MemoryStorage) GetEvents() []MemoryEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]MemoryEvent, len(m.events))
	copy(result, m.events)
	return result
}

// GetEventsByProject returns events for a specific project.
func (m *MemoryStorage) GetEventsByProject(projectID string) []MemoryEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []MemoryEvent
	for _, e := range m.events {
		if e.ProjectID == projectID {
			result = append(result, e)
		}
	}
	return result
}

// Count returns the total number of stored events.
func (m *MemoryStorage) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.events)
}

// Close is a no-op for in-memory storage.
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.Printf("[Memory Storage] Closing in-memory storage (%d events discarded)", len(m.events))
	m.events = nil
	return nil
}
