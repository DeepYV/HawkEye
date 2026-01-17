/**
 * PostgreSQL Schema
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Database schema definition
 */

package store

const (
	// CreateIncidentsTable creates the incidents table
	CreateIncidentsTable = `
	CREATE TABLE IF NOT EXISTS incidents (
		incident_id VARCHAR(255) PRIMARY KEY,
		session_id VARCHAR(255) NOT NULL,
		project_id VARCHAR(255) NOT NULL,
		frustration_score INTEGER NOT NULL,
		confidence_level VARCHAR(50) NOT NULL,
		confidence_score DECIMAL(5,2) NOT NULL,
		triggering_signals JSONB NOT NULL,
		primary_failure_point TEXT NOT NULL,
		severity_type VARCHAR(50) NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		explanation TEXT NOT NULL,
		signal_details JSONB NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'draft',
		suppressed BOOLEAN NOT NULL DEFAULT FALSE,
		external_ticket_id VARCHAR(255),
		external_system VARCHAR(50),
		exported_at TIMESTAMP,
		export_failed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	`

	// CreateIndexes creates indexes for efficient queries
	CreateIndexes = `
	CREATE INDEX IF NOT EXISTS idx_incidents_project_id ON incidents(project_id);
	CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
	CREATE INDEX IF NOT EXISTS idx_incidents_confidence_score ON incidents(confidence_score);
	CREATE INDEX IF NOT EXISTS idx_incidents_suppressed ON incidents(suppressed);
	CREATE INDEX IF NOT EXISTS idx_incidents_external_ticket_id ON incidents(external_ticket_id);
	CREATE INDEX IF NOT EXISTS idx_incidents_created_at ON incidents(created_at);
	CREATE INDEX IF NOT EXISTS idx_incidents_project_status ON incidents(project_id, status);
	CREATE INDEX IF NOT EXISTS idx_incidents_project_exported ON incidents(project_id, external_ticket_id);
	`
)