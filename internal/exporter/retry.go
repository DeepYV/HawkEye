/**
 * Retry Logic
 *
 * Author: Grace Lee (Team Beta)
 * Responsibility: Retry with exponential backoff
 *
 * Rules:
 * - Exponential backoff
 * - Max retry cap
 * - Never create duplicates
 */

package exporter

import (
	"context"
	"time"
)

const (
	maxRetries     = 3
	initialBackoff = 1 * time.Second
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
}

// RetryWithBackoff retries operation with exponential backoff
func RetryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == maxRetries-1 {
			break
		}

		// Exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay *= 2 // Double the delay
		}
	}

	return lastErr
}
