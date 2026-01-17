/**
 * Session Completion Detection
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Session completion rules and detection
 */

package session

import (
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// IdleTimeout is the time before session becomes idle
	// For testing, use shorter timeout if TEST_MODE env var is set
	IdleTimeout = 5 * time.Minute

	// CompletionTimeout is the time after idle before session completes
	// For testing, use shorter timeout if TEST_MODE env var is set
	CompletionTimeout = 10 * time.Minute

	// MaxSessionDuration is the maximum session duration
	MaxSessionDuration = 4 * time.Hour
)

// ShouldComplete checks if session should be completed
func ShouldComplete(state *SessionState, now time.Time) bool {
	state.mu.RLock()
	defer state.mu.RUnlock()

	// Already completed
	if state.State == types.SessionStateCompleted {
		return false
	}

	// Check explicit completion conditions
	if shouldCompleteByDuration(state, now) {
		return true
	}

	if shouldCompleteByIdleTimeout(state, now) {
		return true
	}

	return false
}

// shouldCompleteByDuration checks if session exceeded max duration
func shouldCompleteByDuration(state *SessionState, now time.Time) bool {
	return now.Sub(state.StartTime) > MaxSessionDuration
}

// shouldCompleteByIdleTimeout checks if session exceeded completion timeout
func shouldCompleteByIdleTimeout(state *SessionState, now time.Time) bool {
	if state.State == types.SessionStateIdle {
		// If idle, check if completion timeout exceeded
		return now.Sub(state.LastActivity) > CompletionTimeout
	}

	// If active, check if total idle time exceeded
	return now.Sub(state.LastActivity) > (IdleTimeout + CompletionTimeout)
}

// ShouldBecomeIdle checks if session should become idle
func ShouldBecomeIdle(state *SessionState, now time.Time) bool {
	state.mu.RLock()
	defer state.mu.RUnlock()

	// Only active sessions can become idle
	if state.State != types.SessionStateActive {
		return false
	}

	// Check if idle timeout exceeded
	return now.Sub(state.LastActivity) > IdleTimeout
}

// CheckForSessionReset checks if event indicates session reset
func CheckForSessionReset(event types.Event) bool {
	// Check for explicit session reset event
	if event.EventType == "navigation" {
		// Navigation events might indicate page reload
		// Check metadata for explicit reset
		if metadata := event.Metadata; metadata != nil {
			if reset, ok := metadata["reset"].(bool); ok && reset {
				return true
			}
		}
	}

	return false
}
