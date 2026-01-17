/**
 * Incident Store Engine
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Main store operations
 */

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// StoreEngine handles all store operations
type StoreEngine struct {
	store *Store
}

// NewStoreEngine creates a new store engine
func NewStoreEngine(store *Store) *StoreEngine {
	return &StoreEngine{store: store}
}

// StoreIncident stores an incident
func (e *StoreEngine) StoreIncident(ctx context.Context, incident types.Incident) error {
	// Log-only mode
	if e.store.db == nil {
		fmt.Printf("[Incident Store] LOG-ONLY: Would store incident %s (session: %s, score: %d, confidence: %s)\n",
			incident.IncidentID, incident.SessionID, incident.FrustrationScore, incident.ConfidenceLevel)
		return nil
	}

	// Set defaults
	if incident.Status == "" {
		incident.Status = StatusDraft
	}
	if incident.CreatedAt.IsZero() {
		incident.CreatedAt = time.Now()
	}
	incident.UpdatedAt = time.Now()

	// Convert to JSON
	triggeringSignalsJSON, err := types.TriggeringSignalsToJSON(incident.TriggeringSignals)
	if err != nil {
		return fmt.Errorf("failed to marshal triggering signals: %w", err)
	}

	signalDetailsJSON, err := json.Marshal(incident.SignalDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal signal details: %w", err)
	}

	query := `
		INSERT INTO incidents (
			incident_id, session_id, project_id, frustration_score,
			confidence_level, confidence_score, triggering_signals,
			primary_failure_point, severity_type, timestamp, explanation,
			signal_details, status, suppressed, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (incident_id) DO UPDATE SET
			updated_at = $16,
			status = $13,
			suppressed = $14
	`

	_, err = e.store.db.ExecContext(ctx, query,
		incident.IncidentID,
		incident.SessionID,
		incident.ProjectID,
		incident.FrustrationScore,
		incident.ConfidenceLevel,
		incident.ConfidenceScore,
		triggeringSignalsJSON,
		incident.PrimaryFailurePoint,
		incident.SeverityType,
		incident.Timestamp,
		incident.Explanation,
		signalDetailsJSON,
		incident.Status,
		incident.Suppressed,
		incident.CreatedAt,
		incident.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store incident: %w", err)
	}

	return nil
}

// GetEligibleIncidents gets incidents eligible for export
func (e *StoreEngine) GetEligibleIncidents(ctx context.Context, req types.QueryRequest) ([]types.Incident, error) {
	// Log-only mode
	if e.store.db == nil {
		fmt.Printf("[Incident Store] LOG-ONLY: Query for eligible incidents (status: %s, exported: %v)\n",
			req.Status, req.Exported != nil && *req.Exported == false)
		return []types.Incident{}, nil
	}

	query := `
		SELECT 
			incident_id, session_id, project_id, frustration_score,
			confidence_level, confidence_score, triggering_signals,
			primary_failure_point, severity_type, timestamp, explanation,
			signal_details, status, suppressed, external_ticket_id,
			external_system, exported_at, export_failed, created_at, updated_at
		FROM incidents
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	// Build WHERE clause
	if req.ProjectID != "" {
		query += fmt.Sprintf(" AND project_id = $%d", argPos)
		args = append(args, req.ProjectID)
		argPos++
	}

	if req.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, req.Status)
		argPos++
	} else {
		// Default: only confirmed incidents
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, StatusConfirmed)
		argPos++
	}

	if req.MinConfidence > 0 {
		query += fmt.Sprintf(" AND confidence_score >= $%d", argPos)
		args = append(args, req.MinConfidence)
		argPos++
	}

	if req.Suppressed != nil {
		query += fmt.Sprintf(" AND suppressed = $%d", argPos)
		args = append(args, *req.Suppressed)
		argPos++
	} else {
		// Default: not suppressed
		query += fmt.Sprintf(" AND suppressed = $%d", argPos)
		args = append(args, false)
		argPos++
	}

	if req.Exported != nil {
		if *req.Exported {
			query += fmt.Sprintf(" AND external_ticket_id IS NOT NULL")
		} else {
			query += fmt.Sprintf(" AND external_ticket_id IS NULL")
		}
	} else {
		// Default: not exported
		query += fmt.Sprintf(" AND external_ticket_id IS NULL")
	}

	// Order by timestamp (oldest first)
	query += " ORDER BY timestamp ASC"

	// Limit and offset
	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, req.Limit)
		argPos++
	} else {
		query += " LIMIT 100" // Default limit
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, req.Offset)
	}

	rows, err := e.store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []types.Incident
	for rows.Next() {
		incident, err := e.scanIncident(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, incident)
	}

	return incidents, nil
}

// MarkExported marks an incident as exported
func (e *StoreEngine) MarkExported(ctx context.Context, incidentID, externalTicketID, externalSystem string) error {
	// Log-only mode
	if e.store.db == nil {
		fmt.Printf("[Incident Store] LOG-ONLY: Would mark incident %s as exported (ticket: %s, system: %s)\n",
			incidentID, externalTicketID, externalSystem)
		return nil
	}

	query := `
		UPDATE incidents
		SET external_ticket_id = $1,
		    external_system = $2,
		    exported_at = $3,
		    updated_at = $3
		WHERE incident_id = $4
	`

	now := time.Now()
	_, err := e.store.db.ExecContext(ctx, query, externalTicketID, externalSystem, now, incidentID)
	if err != nil {
		return fmt.Errorf("failed to mark incident as exported: %w", err)
	}

	return nil
}

// MarkExportFailed marks an incident export as failed
func (e *StoreEngine) MarkExportFailed(ctx context.Context, incidentID string) error {
	// Log-only mode
	if e.store.db == nil {
		fmt.Printf("[Incident Store] LOG-ONLY: Would mark incident %s export as failed\n", incidentID)
		return nil
	}

	query := `
		UPDATE incidents
		SET export_failed = TRUE,
		    updated_at = $1
		WHERE incident_id = $2
	`

	now := time.Now()
	_, err := e.store.db.ExecContext(ctx, query, now, incidentID)
	if err != nil {
		return fmt.Errorf("failed to mark export as failed: %w", err)
	}

	return nil
}

// ConfirmIncident confirms an incident (changes status from draft to confirmed)
func (e *StoreEngine) ConfirmIncident(ctx context.Context, incidentID string) error {
	// Log-only mode
	if e.store.db == nil {
		fmt.Printf("[Incident Store] LOG-ONLY: Would confirm incident %s\n", incidentID)
		return nil
	}

	query := `
		UPDATE incidents
		SET status = $1,
		    updated_at = $2
		WHERE incident_id = $3
	`

	now := time.Now()
	_, err := e.store.db.ExecContext(ctx, query, StatusConfirmed, now, incidentID)
	if err != nil {
		return fmt.Errorf("failed to confirm incident: %w", err)
	}

	return nil
}

// scanIncident scans a row into an Incident
func (e *StoreEngine) scanIncident(rows *sql.Rows) (types.Incident, error) {
	var incident types.Incident
	var triggeringSignalsJSON []byte
	var signalDetailsJSON []byte
	var exportedAt sql.NullTime

	err := rows.Scan(
		&incident.IncidentID,
		&incident.SessionID,
		&incident.ProjectID,
		&incident.FrustrationScore,
		&incident.ConfidenceLevel,
		&incident.ConfidenceScore,
		&triggeringSignalsJSON,
		&incident.PrimaryFailurePoint,
		&incident.SeverityType,
		&incident.Timestamp,
		&incident.Explanation,
		&signalDetailsJSON,
		&incident.Status,
		&incident.Suppressed,
		&incident.ExternalTicketID,
		&incident.ExternalSystem,
		&exportedAt,
		&incident.ExportFailed,
		&incident.CreatedAt,
		&incident.UpdatedAt,
	)

	if err != nil {
		return incident, err
	}

	// Parse JSON fields
	incident.TriggeringSignals, err = types.TriggeringSignalsFromJSON(triggeringSignalsJSON)
	if err != nil {
		return incident, fmt.Errorf("failed to parse triggering signals: %w", err)
	}

	incident.SignalDetails, err = types.SignalDetailsFromJSON(signalDetailsJSON)
	if err != nil {
		return incident, fmt.Errorf("failed to parse signal details: %w", err)
	}

	if exportedAt.Valid {
		incident.ExportedAt = &exportedAt.Time
	}

	return incident, nil
}