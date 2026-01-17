/**
 * Failure Handling
 * 
 * Author: Grace Lee (Team Beta)
 * Responsibility: Handle export failures gracefully
 * 
 * Rules:
 * - On API failure: retry with backoff
 * - On partial success: recover via idempotency
 * - On permanent failure: mark export_failed
 */

package exporter

import (
	"context"
	"errors"
	"log"

	"github.com/your-org/frustration-engine/internal/types"
)

var (
	ErrExportFailed = errors.New("export failed")
	ErrPermanentFailure = errors.New("permanent export failure")
)

// HandleExportFailure handles export failure
func HandleExportFailure(
	ctx context.Context,
	incident types.Incident,
	err error,
	store Store,
) {
	// Log failure
	log.Printf("[Ticket Exporter] Export failed for incident %s: %v", incident.IncidentID, err)

	// Retry with exponential backoff
	retryErr := RetryWithBackoff(ctx, func() error {
		// Attempt export again (idempotency key prevents duplicates)
		return err // Placeholder - would retry actual export
	})

	if retryErr != nil {
		// Permanent failure - mark in store
		if markErr := store.MarkExportFailed(ctx, incident.IncidentID); markErr != nil {
			log.Printf("[Ticket Exporter] Failed to mark export_failed: %v", markErr)
		}
	}
}