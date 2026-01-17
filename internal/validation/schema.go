/**
 * Event Schema Validation
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Event validation & schema checking
 */

package validation

import (
	"fmt"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// Allowed event types
var allowedEventTypes = map[string]bool{
	"click":       true,
	"scroll":      true,
	"input":       true,
	"form_submit": true,
	"navigation":  true,
	"error":       true,
	"network":     true,
	"performance": true,
	"loading":     true,
}

// MaxEventSize is the maximum size of a single event in bytes
const MaxEventSize = 10 * 1024 // 10KB

// MaxBatchSize is the maximum number of events in a batch
const MaxBatchSize = 100

// MaxPayloadSize is the maximum total payload size in bytes
const MaxPayloadSize = 500 * 1024 // 500KB

// ValidateEvent validates a single event
func ValidateEvent(event types.Event, index int) *types.ValidationError {
	// Check event type
	if event.EventType == "" {
		return &types.ValidationError{
			Field:   "eventType",
			Reason:  "missing required field",
			EventID: index,
		}
	}

	if !allowedEventTypes[event.EventType] {
		return &types.ValidationError{
			Field:   "eventType",
			Reason:  fmt.Sprintf("unknown event type: %s", event.EventType),
			EventID: index,
		}
	}

	// Check timestamp
	if event.Timestamp == "" {
		return &types.ValidationError{
			Field:   "timestamp",
			Reason:  "missing required field",
			EventID: index,
		}
	}

	// Validate timestamp format (ISO 8601)
	if _, err := time.Parse(time.RFC3339, event.Timestamp); err != nil {
		return &types.ValidationError{
			Field:   "timestamp",
			Reason:  "invalid timestamp format",
			EventID: index,
		}
	}

	// Check session ID
	if event.SessionID == "" {
		return &types.ValidationError{
			Field:   "sessionId",
			Reason:  "missing required field",
			EventID: index,
		}
	}

	// Check route
	if event.Route == "" {
		return &types.ValidationError{
			Field:   "route",
			Reason:  "missing required field",
			EventID: index,
		}
	}

	// Check target
	if event.Target.Type == "" {
		return &types.ValidationError{
			Field:   "target.type",
			Reason:  "missing required field",
			EventID: index,
		}
	}

	return nil
}

// ValidateBatch validates a batch of events
func ValidateBatch(events []types.Event) ([]types.Event, []types.ValidationError) {
	validEvents := make([]types.Event, 0, len(events))
	errors := make([]types.ValidationError, 0)

	// Check batch size
	if len(events) > MaxBatchSize {
		// Drop entire batch if too large
		return validEvents, []types.ValidationError{
			{
				Field:  "events",
				Reason: fmt.Sprintf("batch size exceeds maximum: %d", MaxBatchSize),
			},
		}
	}

	// Validate each event
	for i, event := range events {
		if err := ValidateEvent(event, i); err != nil {
			errors = append(errors, *err)
			continue // Drop invalid event
		}

		validEvents = append(validEvents, event)
	}

	return validEvents, errors
}
