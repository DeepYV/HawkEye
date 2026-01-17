/**
 * Incident Ingestion Handler
 * 
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Handle incident ingestion from UFSE
 */

package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/your-org/frustration-engine/internal/store"
	"github.com/your-org/frustration-engine/internal/types"
)

// IngestHandler handles incident ingestion
type IngestHandler struct {
	engine *store.StoreEngine
	dedup  *store.DeduplicationEngine
}

// NewIngestHandler creates a new ingest handler
func NewIngestHandler(engine *store.StoreEngine, dedup *store.DeduplicationEngine) *IngestHandler {
	return &IngestHandler{
		engine: engine,
		dedup:  dedup,
	}
}

// IngestIncident handles incident ingestion requests
func (h *IngestHandler) IngestIncident(w http.ResponseWriter, r *http.Request) {
	var req types.IncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	incident := req.Incident

	// Validate incident
	if err := validateIncident(incident); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check for duplicates
	isDuplicate, existingID, err := h.dedup.CheckDuplicate(r.Context(), incident)
	if err != nil {
		log.Printf("[Incident Store] Deduplication check failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if isDuplicate {
		// Return existing incident ID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"duplicate":   true,
			"incident_id": existingID,
		})
		return
	}

	// Store incident (default status: draft)
	if incident.Status == "" {
		incident.Status = store.StatusDraft
	}

	// Store incident
	if err := h.engine.StoreIncident(r.Context(), incident); err != nil {
		log.Printf("[Incident Store] Failed to store incident: %v", err)
		http.Error(w, "failed to store incident", http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"duplicate":   false,
		"incident_id": incident.IncidentID,
	})
}

// validateIncident validates incident structure
func validateIncident(incident types.Incident) error {
	if incident.IncidentID == "" {
		return fmt.Errorf("incident_id is required")
	}
	if incident.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if incident.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if incident.ConfidenceLevel == "" {
		return fmt.Errorf("confidence_level is required")
	}
	if incident.PrimaryFailurePoint == "" {
		return fmt.Errorf("primary_failure_point is required")
	}
	if incident.SeverityType == "" {
		return fmt.Errorf("severity_type is required")
	}
	if incident.Explanation == "" {
		return fmt.Errorf("explanation is required")
	}
	return nil
}