/**
 * Enhanced Event Forwarding Manager with Retry and Circuit Breaker
 * 
 * Author: Principal Engineer + Team Beta
 * Responsibility: Enhanced forwarding with resilience patterns
 */

package forwarding

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/your-org/frustration-engine/internal/performance"
	"github.com/your-org/frustration-engine/internal/resilience"
	"github.com/your-org/frustration-engine/internal/types"
)

// EnhancedManager handles forwarding with retry and circuit breaker
type EnhancedManager struct {
	sessionManagerURL string
	forwardingQueue   chan forwardingTask
	workerCount       int
	clientPool        *performance.HTTPClientPool
	circuitBreaker    *resilience.CircuitBreaker
	retryConfig       resilience.RetryConfig
}

// NewEnhancedManager creates a new enhanced forwarding manager
func NewEnhancedManager(sessionManagerURL string, workerCount int) *EnhancedManager {
	return &EnhancedManager{
		sessionManagerURL: sessionManagerURL,
		forwardingQueue:   make(chan forwardingTask, 1000),
		workerCount:       workerCount,
		clientPool:        performance.NewHTTPClientPool(),
		circuitBreaker:    resilience.NewCircuitBreaker("session-manager", 5, 2, 30*time.Second),
		retryConfig:       resilience.DefaultRetryConfig(),
	}
}

// ForwardEvents queues events for forwarding with enhanced resilience
func (m *EnhancedManager) ForwardEvents(ctx context.Context, projectID string, events []types.Event) {
	// Non-blocking queue
	select {
	case m.forwardingQueue <- forwardingTask{projectID: projectID, events: events}:
		// Queued successfully
	default:
		log.Printf("[Enhanced Forwarding] Queue full, dropping %d events", len(events))
	}
}

// StartWorkers starts worker goroutines
func (m *EnhancedManager) StartWorkers() {
	for i := 0; i < m.workerCount; i++ {
		go m.worker()
	}
}

// worker processes forwarding tasks with retry and circuit breaker
func (m *EnhancedManager) worker() {
	for task := range m.forwardingQueue {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		// Use circuit breaker and retry
		err := m.circuitBreaker.Execute(ctx, func() error {
			return resilience.Retry(ctx, func() error {
				return m.forwardToSessionManager(ctx, task.projectID, task.events)
			}, m.retryConfig)
		})
		
		if err != nil {
			log.Printf("[Enhanced Forwarding] Failed to forward after retries: %v", err)
		}
		
		cancel()
	}
}

// forwardToSessionManager forwards events with connection pooling
func (m *EnhancedManager) forwardToSessionManager(ctx context.Context, projectID string, events []types.Event) error {
	if m.sessionManagerURL == "" {
		return nil
	}

	// Group events by session
	sessionGroups := make(map[string][]types.Event)
	for _, event := range events {
		sessionGroups[event.SessionID] = append(sessionGroups[event.SessionID], event)
	}

	// Get HTTP client from pool
	client := m.clientPool.GetClient(m.sessionManagerURL)

	// Forward each session's events
	for sessionID, sessionEvents := range sessionGroups {
		payload := map[string]interface{}{
			"project_id": projectID,
			"session_id": sessionID,
			"events":     sessionEvents,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[Enhanced Forwarding] Failed to marshal: %v", err)
			continue
		}

		url := m.sessionManagerURL + "/v1/sessions/events"
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("[Enhanced Forwarding] Status %d for session %s", resp.StatusCode, sessionID)
			continue
		}
	}

	return nil
}

// Close closes the manager and cleans up resources
func (m *EnhancedManager) Close() {
	close(m.forwardingQueue)
	m.clientPool.Close()
}
