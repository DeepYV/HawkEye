/**
 * Query Handler
 * 
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Handle queries for Ticket Exporter
 */

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/your-org/frustration-engine/internal/store"
	"github.com/your-org/frustration-engine/internal/types"
)

// QueryHandler handles incident queries
type QueryHandler struct {
	engine *store.StoreEngine
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(engine *store.StoreEngine) *QueryHandler {
	return &QueryHandler{engine: engine}
}

// QueryIncidents handles incident query requests
func (h *QueryHandler) QueryIncidents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := types.QueryRequest{}

	if projectID := r.URL.Query().Get("project_id"); projectID != "" {
		req.ProjectID = projectID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = status
	}

	if minConfidence := r.URL.Query().Get("min_confidence"); minConfidence != "" {
		if val, err := strconv.ParseFloat(minConfidence, 64); err == nil {
			req.MinConfidence = val
		}
	}

	if suppressed := r.URL.Query().Get("suppressed"); suppressed != "" {
		val := suppressed == "true"
		req.Suppressed = &val
	}

	if exported := r.URL.Query().Get("exported"); exported != "" {
		val := exported == "true"
		req.Exported = &val
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			req.Limit = val
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			req.Offset = val
		}
	}

	// Query incidents
	incidents, err := h.engine.GetEligibleIncidents(r.Context(), req)
	if err != nil {
		http.Error(w, "failed to query incidents", http.StatusInternalServerError)
		return
	}

	// Return response
	response := types.QueryResponse{
		Incidents: incidents,
		Total:     len(incidents),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}