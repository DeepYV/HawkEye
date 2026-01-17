/**
 * Event Forwarding Manager
 * 
 * Author: Henry Wilson (Team Beta)
 * Responsibility: Forward events to Session Manager
 */

package forwarding

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// Manager handles forwarding events to Session Manager
type Manager struct {
	sessionManagerURL string
	forwardingQueue   chan forwardingTask
	workerCount       int
}

type forwardingTask struct {
	projectID string
	events     []types.Event
}

// NewManager creates a new forwarding manager
func NewManager(sessionManagerURL string, workerCount int) *Manager {
	m := &Manager{
		sessionManagerURL: sessionManagerURL,
		forwardingQueue:   make(chan forwardingTask, 1000),
		workerCount:       workerCount,
	}

	// Start workers
	for i := 0; i < workerCount; i++ {
		go m.worker()
	}

	return m
}

// ForwardEvents queues events for forwarding
func (m *Manager) ForwardEvents(ctx context.Context, projectID string, events []types.Event) {
	// Non-blocking queue - if full, events are dropped
	select {
	case m.forwardingQueue <- forwardingTask{projectID: projectID, events: events}:
		// Queued successfully
	default:
		// Queue full - drop events silently
		// Reliability > immediacy
	}
}

// worker processes forwarding tasks
func (m *Manager) worker() {
	for task := range m.forwardingQueue {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		if err := m.forwardToSessionManager(ctx, task.projectID, task.events); err != nil {
			// Log error but don't retry indefinitely
			// Events are already persisted, so this is best-effort
		}
		
		cancel()
	}
}

// forwardToSessionManager forwards events to Session Manager
func (m *Manager) forwardToSessionManager(ctx context.Context, projectID string, events []types.Event) error {
	if m.sessionManagerURL == "" {
		// Log and skip if URL not configured
		log.Printf("[Event Ingestion] Session Manager URL not configured, logging events instead")
		for _, event := range events {
			log.Printf("[Event Ingestion] Event: %s (session: %s, type: %s)", event.EventType, event.SessionID, event.Target.Type)
		}
		return nil
	}

	// Group events by session to maintain order
	sessionGroups := make(map[string][]types.Event)
	for _, event := range events {
		sessionGroups[event.SessionID] = append(sessionGroups[event.SessionID], event)
	}

	// Forward each session's events
	for sessionID, sessionEvents := range sessionGroups {
		payload := map[string]interface{}{
			"project_id": projectID,
			"session_id": sessionID,
			"events":     sessionEvents,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[Event Ingestion] Failed to marshal payload: %v", err)
			continue
		}

		// HTTP POST to Session Manager
		url := m.sessionManagerURL + "/v1/sessions/events"
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
		if err != nil {
			log.Printf("[Event Ingestion] Failed to create request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		// Use HTTP client with connection pooling
		client := &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Event Ingestion] Failed to forward to Session Manager (session: %s): %v", sessionID, err)
			// Log events for debugging
			log.Printf("[Event Ingestion] Events logged: %d events for session %s", len(sessionEvents), sessionID)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("[Event Ingestion] Session Manager returned status %d for session %s", resp.StatusCode, sessionID)
			continue
		}

		log.Printf("[Event Ingestion] Successfully forwarded %d events to Session Manager (session: %s)", len(sessionEvents), sessionID)
	}

	return nil
}