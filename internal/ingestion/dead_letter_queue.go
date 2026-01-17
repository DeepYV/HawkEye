/**
 * Dead Letter Queue
 * 
 * Author: Team Alpha (Bob, Charlie, Diana)
 * Responsibility: Store failed events for later analysis and retry
 * 
 * Handles events that fail validation or cannot be processed
 */

package ingestion

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// DeadLetterQueue stores failed events
type DeadLetterQueue struct {
	mu      sync.RWMutex
	queue   []DeadLetterEvent
	maxSize int
	storage DeadLetterStorage
}

// DeadLetterEvent represents a failed event
type DeadLetterEvent struct {
	Event     types.Event
	Reason    string
	Timestamp time.Time
	RetryCount int
	ProjectID string
}

// DeadLetterStorage interface for persistent storage
type DeadLetterStorage interface {
	Store(ctx context.Context, event DeadLetterEvent) error
	Query(ctx context.Context, projectID string, limit int) ([]DeadLetterEvent, error)
}

// NewDeadLetterQueue creates a new dead letter queue
func NewDeadLetterQueue(maxSize int, storage DeadLetterStorage) *DeadLetterQueue {
	return &DeadLetterQueue{
		queue:   make([]DeadLetterEvent, 0, maxSize),
		maxSize: maxSize,
		storage: storage,
	}
}

// Enqueue adds an event to the dead letter queue
func (dlq *DeadLetterQueue) Enqueue(ctx context.Context, event types.Event, reason string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	// Get project ID from context or event metadata
	projectID := getProjectIDFromContext(ctx)
	if projectID == "" {
		projectID = "unknown"
	}

	dlEvent := DeadLetterEvent{
		Event:      event,
		Reason:     reason,
		Timestamp:  time.Now(),
		RetryCount: 0,
		ProjectID:  projectID,
	}

	// Store persistently if storage is available
	if dlq.storage != nil {
		if err := dlq.storage.Store(ctx, dlEvent); err != nil {
			log.Printf("[DeadLetterQueue] Failed to store event: %v", err)
		}
	}

	// Add to in-memory queue
	if len(dlq.queue) >= dlq.maxSize {
		// Remove oldest event
		dlq.queue = dlq.queue[1:]
	}
	dlq.queue = append(dlq.queue, dlEvent)

	log.Printf("[DeadLetterQueue] Enqueued event: session=%s, reason=%s", event.SessionID, reason)
	return nil
}

// GetFailedEvents returns failed events for a project
func (dlq *DeadLetterQueue) GetFailedEvents(ctx context.Context, projectID string, limit int) ([]DeadLetterEvent, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	if dlq.storage != nil {
		return dlq.storage.Query(ctx, projectID, limit)
	}

	// Return from in-memory queue
	var results []DeadLetterEvent
	for _, event := range dlq.queue {
		if event.ProjectID == projectID {
			results = append(results, event)
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// getProjectIDFromContext extracts project ID from context
func getProjectIDFromContext(ctx context.Context) string {
	// Implementation depends on auth middleware
	if projectID := ctx.Value("project_id"); projectID != nil {
		if pid, ok := projectID.(string); ok {
			return pid
		}
	}
	return ""
}

// InMemoryDeadLetterStorage is an in-memory implementation for testing
type InMemoryDeadLetterStorage struct {
	mu     sync.RWMutex
	events []DeadLetterEvent
}

// NewInMemoryDeadLetterStorage creates a new in-memory storage
func NewInMemoryDeadLetterStorage() *InMemoryDeadLetterStorage {
	return &InMemoryDeadLetterStorage{
		events: make([]DeadLetterEvent, 0),
	}
}

// Store stores an event
func (s *InMemoryDeadLetterStorage) Store(ctx context.Context, event DeadLetterEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

// Query queries events
func (s *InMemoryDeadLetterStorage) Query(ctx context.Context, projectID string, limit int) ([]DeadLetterEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []DeadLetterEvent
	for _, event := range s.events {
		if event.ProjectID == projectID {
			results = append(results, event)
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// LogDeadLetterStorage logs events instead of storing (for testing)
type LogDeadLetterStorage struct{}

// NewLogDeadLetterStorage creates a new log-based storage
func NewLogDeadLetterStorage() *LogDeadLetterStorage {
	return &LogDeadLetterStorage{}
}

// Store logs an event
func (s *LogDeadLetterStorage) Store(ctx context.Context, event DeadLetterEvent) error {
	eventJSON, _ := json.Marshal(event.Event)
	log.Printf("[DeadLetterQueue] LOG-ONLY: Event failed - session=%s, reason=%s, event=%s",
		event.Event.SessionID, event.Reason, string(eventJSON))
	return nil
}

// Query returns empty (log-only mode)
func (s *LogDeadLetterStorage) Query(ctx context.Context, projectID string, limit int) ([]DeadLetterEvent, error) {
	return []DeadLetterEvent{}, nil
}
