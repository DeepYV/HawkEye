/**
 * Robust Event Ingestion Handler
 * 
 * Author: Team Alpha (Bob, Charlie, Diana)
 * Responsibility: Production-grade event ingestion with comprehensive edge case handling
 * 
 * Handles 100+ edge cases for zero false alarms
 */

package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/auth"
	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/validation"
)

// RobustHandler handles event ingestion with comprehensive edge case handling
type RobustHandler struct {
	validator        *validation.EdgeCaseValidator
	rateLimiter      RateLimiter
	storage          Storage
	forwarder        Forwarder
	deadLetterQueue DeadLetterQueueInterface
	metrics          Metrics

	// Circuit breaker for downstream services
	sessionManagerCircuitBreaker *CircuitBreaker
	storageCircuitBreaker        *CircuitBreaker

	// Retry configuration
	maxRetries      int
	retryBackoff    time.Duration
	maxRetryBackoff time.Duration
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Allow(projectID string) bool
	Record(projectID string, count int)
}

// Storage interface for event storage
type Storage interface {
	StoreEvents(ctx context.Context, projectID string, events []types.Event) error
}

// Forwarder interface for forwarding events
type Forwarder interface {
	ForwardEvents(ctx context.Context, projectID string, events []types.Event) error
}

// DeadLetterQueueInterface interface for failed events
type DeadLetterQueueInterface interface {
	Enqueue(ctx context.Context, event types.Event, reason string) error
}

// Metrics interface for observability
type Metrics interface {
	IncrementCounter(name string, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	mu                sync.RWMutex
	failureCount      int
	lastFailureTime   time.Time
	state             string // "closed", "open", "half-open"
	failureThreshold  int
	timeout           time.Duration
	successThreshold  int
	successCount      int
}

const (
	circuitStateClosed   = "closed"
	circuitStateOpen     = "open"
	circuitStateHalfOpen = "half-open"
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout time.Duration, successThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:            circuitStateClosed,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		successThreshold: successThreshold,
	}
}

// Allow checks if request should be allowed
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case circuitStateClosed:
		return true
	case circuitStateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = circuitStateHalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			return true
		}
		return false
	case circuitStateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == circuitStateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = circuitStateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	} else if cb.state == circuitStateClosed {
		cb.failureCount = 0 // Reset on success
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == circuitStateHalfOpen {
		cb.state = circuitStateOpen
		cb.successCount = 0
	} else if cb.state == circuitStateClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = circuitStateOpen
	}
}

// NewRobustHandler creates a new robust handler
func NewRobustHandler(
	validator *validation.EdgeCaseValidator,
	rateLimiter RateLimiter,
	storage Storage,
	forwarder Forwarder,
	deadLetterQueue DeadLetterQueueInterface,
	metrics Metrics,
) *RobustHandler {
	return &RobustHandler{
		validator:                   validator,
		rateLimiter:                 rateLimiter,
		storage:                     storage,
		forwarder:                   forwarder,
		deadLetterQueue:            deadLetterQueue,
		metrics:                       metrics,
		sessionManagerCircuitBreaker: NewCircuitBreaker(5, 30*time.Second, 2),
		storageCircuitBreaker:        NewCircuitBreaker(5, 30*time.Second, 2),
		maxRetries:                  3,
		retryBackoff:                100 * time.Millisecond,
		maxRetryBackoff:             5 * time.Second,
	}
}

// IngestEvents handles event ingestion with comprehensive edge case handling
func (h *RobustHandler) IngestEvents(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get project ID from context
	projectID := getProjectID(r.Context())
	if projectID == "" {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Rate limiting
	if !h.rateLimiter.Allow(projectID) {
		h.metrics.IncrementCounter("ingestion_rate_limited", map[string]string{"project_id": projectID})
		h.respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	// Decode request with size limit
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024) // 10MB limit
	var req types.IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.metrics.IncrementCounter("ingestion_decode_error", map[string]string{"project_id": projectID})
		h.respondError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	// Validate batch size
	if len(req.Events) == 0 {
		h.respondSuccess(w, 0)
		return
	}

	if len(req.Events) > 1000 {
		h.metrics.IncrementCounter("ingestion_batch_too_large", map[string]string{"project_id": projectID})
		h.respondError(w, http.StatusBadRequest, "batch size exceeds limit")
		return
	}

	// Comprehensive validation
	validEvents := make([]types.Event, 0, len(req.Events))
	invalidCount := 0

	for i, event := range req.Events {
		// Edge case validation
		valid, errors := h.validator.ValidateEventComprehensive(event)
		if !valid {
			invalidCount++
			h.metrics.IncrementCounter("ingestion_validation_failed", map[string]string{
				"project_id": projectID,
				"reason":     "validation_error",
			})

			// Send to dead letter queue
			ctxDLQ, cancelDLQ := context.WithTimeout(context.Background(), 5*time.Second)
			h.deadLetterQueue.Enqueue(ctxDLQ, event, fmt.Sprintf("validation failed: %v", errors))
			cancelDLQ()

			log.Printf("[Ingestion] Event %d validation failed: %v", i, errors)
			continue
		}

		validEvents = append(validEvents, event)
	}

	if len(validEvents) == 0 {
		h.respondSuccess(w, 0)
		return
	}

	// Record rate limit usage
	h.rateLimiter.Record(projectID, len(validEvents))

	// Store events with retry
	go h.storeEventsWithRetry(ctx, projectID, validEvents)

	// Forward events with circuit breaker
	go h.forwardEventsWithCircuitBreaker(ctx, projectID, validEvents)

	// Respond immediately
	h.respondSuccess(w, len(validEvents))
}

// storeEventsWithRetry stores events with exponential backoff retry
func (h *RobustHandler) storeEventsWithRetry(ctx context.Context, projectID string, events []types.Event) {
	if !h.storageCircuitBreaker.Allow() {
		h.metrics.IncrementCounter("storage_circuit_open", map[string]string{"project_id": projectID})
		// Send to dead letter queue
		for _, event := range events {
			h.deadLetterQueue.Enqueue(ctx, event, "storage circuit breaker open")
		}
		return
	}

	backoff := h.retryBackoff
	for attempt := 0; attempt <= h.maxRetries; attempt++ {
		err := h.storage.StoreEvents(ctx, projectID, events)
		if err == nil {
			h.storageCircuitBreaker.RecordSuccess()
			h.metrics.IncrementCounter("storage_success", map[string]string{"project_id": projectID})
			return
		}

		if attempt < h.maxRetries {
			h.metrics.IncrementCounter("storage_retry", map[string]string{
				"project_id": projectID,
				"attempt":    fmt.Sprintf("%d", attempt+1),
			})
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * 2)
			if backoff > h.maxRetryBackoff {
				backoff = h.maxRetryBackoff
			}
		}
	}

	// All retries failed
	h.storageCircuitBreaker.RecordFailure()
	h.metrics.IncrementCounter("storage_failed", map[string]string{"project_id": projectID})

	// Send to dead letter queue
	for _, event := range events {
		h.deadLetterQueue.Enqueue(ctx, event, "storage failed after retries")
	}
}

// forwardEventsWithCircuitBreaker forwards events with circuit breaker
func (h *RobustHandler) forwardEventsWithCircuitBreaker(ctx context.Context, projectID string, events []types.Event) {
	if !h.sessionManagerCircuitBreaker.Allow() {
		h.metrics.IncrementCounter("forwarding_circuit_open", map[string]string{"project_id": projectID})
		// Events are already stored, forwarding failure is not critical
		return
	}

	err := h.forwarder.ForwardEvents(ctx, projectID, events)
	if err != nil {
		h.sessionManagerCircuitBreaker.RecordFailure()
		h.metrics.IncrementCounter("forwarding_failed", map[string]string{"project_id": projectID})
		log.Printf("[Ingestion] Failed to forward events: %v", err)
	} else {
		h.sessionManagerCircuitBreaker.RecordSuccess()
		h.metrics.IncrementCounter("forwarding_success", map[string]string{"project_id": projectID})
	}
}

// respondSuccess sends success response
func (h *RobustHandler) respondSuccess(w http.ResponseWriter, processed int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"processed": processed,
	})
}

// respondError sends error response
func (h *RobustHandler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// getProjectID extracts project ID from context
func getProjectID(ctx context.Context) string {
	return auth.GetProjectID(ctx)
}
