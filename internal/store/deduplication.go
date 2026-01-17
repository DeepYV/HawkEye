/**
 * Deduplication Logic
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Deduplication of incidents
 */

package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// DeduplicationEngine handles incident deduplication
type DeduplicationEngine struct {
	store *Store
}

// NewDeduplicationEngine creates a new deduplication engine
func NewDeduplicationEngine(store *Store) *DeduplicationEngine {
	return &DeduplicationEngine{store: store}
}

// CheckDuplicate checks if a similar incident already exists
func (e *DeduplicationEngine) CheckDuplicate(ctx context.Context, incident types.Incident) (bool, string, error) {
	// Log-only mode
	if e.store.db == nil {
		fmt.Printf("[Incident Store] LOG-ONLY: Would check duplicate for incident %s\n", incident.IncidentID)
		return false, "", nil // No duplicates in log-only mode
	}

	// Check for exact duplicate (same incident ID)
	query := `
		SELECT incident_id
		FROM incidents
		WHERE incident_id = $1
		LIMIT 1
	`

	var existingID string
	err := e.store.db.QueryRowContext(ctx, query, incident.IncidentID).Scan(&existingID)
	if err == nil {
		return true, existingID, nil // Exact duplicate found
	}
	if err != sql.ErrNoRows {
		return false, "", fmt.Errorf("failed to check duplicate: %w", err)
	}

	// Check for similar incident (same session, same failure point, within time window)
	// This is a simple deduplication - can be enhanced later
	query = `
		SELECT incident_id
		FROM incidents
		WHERE session_id = $1
		  AND primary_failure_point = $2
		  AND timestamp BETWEEN $3 AND $4
		LIMIT 1
	`

	timeWindow := 5 * time.Minute // 5 minutes
	timeStart := incident.Timestamp.Add(-timeWindow)
	timeEnd := incident.Timestamp.Add(timeWindow)

	err = e.store.db.QueryRowContext(ctx, query,
		incident.SessionID,
		incident.PrimaryFailurePoint,
		timeStart,
		timeEnd,
	).Scan(&existingID)

	if err == nil {
		return true, existingID, nil // Similar incident found
	}
	if err != sql.ErrNoRows {
		return false, "", fmt.Errorf("failed to check similar incident: %w", err)
	}

	return false, "", nil // No duplicate found
}
