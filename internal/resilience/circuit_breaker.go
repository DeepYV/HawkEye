/**
 * Circuit Breaker
 * 
 * Author: Principal Engineer + Team Beta
 * Responsibility: Circuit breaker pattern for downstream services
 */

package resilience

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreakerState represents circuit breaker state
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	name              string
	failureThreshold  int
	successThreshold  int
	resetTimeout      time.Duration
	halfOpenTimeout   time.Duration
	
	mu                sync.RWMutex
	state             CircuitBreakerState
	failures          int
	successes         int
	lastFailureTime   time.Time
	lastStateChange   time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold, successThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		resetTimeout:     resetTimeout,
		halfOpenTimeout:  resetTimeout / 2,
		state:            StateClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if circuit is open
	if !cb.Allow() {
		return ErrCircuitOpen
	}

	// Execute function
	err := fn()

	// Record result
	if err != nil {
		cb.RecordFailure()
	} else {
		cb.RecordSuccess()
	}

	return err
}

// Allow checks if request is allowed
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if reset timeout has passed
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			// Transition to half-open
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.resetTimeout {
				cb.state = StateHalfOpen
				cb.successes = 0
				cb.lastStateChange = time.Now()
				log.Printf("[CircuitBreaker] %s: Transitioning to HALF-OPEN", cb.name)
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return cb.state == StateHalfOpen
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.successThreshold {
			// Transition to closed
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.lastStateChange = time.Now()
			log.Printf("[CircuitBreaker] %s: Transitioning to CLOSED (recovered)", cb.name)
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.failureThreshold {
			// Transition to open
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
			log.Printf("[CircuitBreaker] %s: Transitioning to OPEN (failures: %d)", cb.name, cb.failures)
		}
	case StateHalfOpen:
		// Any failure in half-open transitions back to open
		cb.state = StateOpen
		cb.successes = 0
		cb.lastStateChange = time.Now()
		log.Printf("[CircuitBreaker] %s: Transitioning to OPEN (failure in half-open)", cb.name)
	}
}

// State returns current circuit breaker state
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() (state CircuitBreakerState, failures, successes int) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state, cb.failures, cb.successes
}
