/**
 * Retry Logic with Exponential Backoff
 * 
 * Author: Principal Engineer + Team Beta
 * Responsibility: Retry logic for transient failures
 */

package resilience

import (
	"context"
	"errors"
	"log"
	"math"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries      int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffMultiplier float64
	RetryableErrors []error
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:       3,
		InitialBackoff:   100 * time.Millisecond,
		MaxBackoff:       5 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// Retry executes a function with retry logic
func Retry(ctx context.Context, fn RetryableFunc, config RetryConfig) error {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err, config) {
			return err // Not retryable, fail immediately
		}

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate backoff with jitter
		backoffDuration := calculateBackoff(backoff, config.MaxBackoff, attempt)
		
		log.Printf("[Retry] Attempt %d/%d failed: %v, retrying in %v", 
			attempt+1, config.MaxRetries+1, err, backoffDuration)

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffDuration):
			// Continue to next attempt
		}

		// Exponential backoff
		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	return lastErr
}

// isRetryableError checks if error is retryable
func isRetryableError(err error, config RetryConfig) bool {
	// If no specific retryable errors configured, retry all
	if len(config.RetryableErrors) == 0 {
		return true
	}

	// Check if error matches any retryable error
	for _, retryableErr := range config.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}

// calculateBackoff calculates backoff duration with jitter
func calculateBackoff(baseBackoff, maxBackoff time.Duration, attempt int) time.Duration {
	// Exponential backoff: base * 2^attempt
	backoff := time.Duration(float64(baseBackoff) * math.Pow(2, float64(attempt)))
	
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	// Add jitter (±20%)
	jitter := time.Duration(float64(backoff) * 0.2)
	backoff = backoff + time.Duration(float64(jitter) * (math.Sin(float64(attempt)) + 1) / 2)

	return backoff
}
