/**
 * API Handler
 * 
 * Author: Henry Wilson (Team Beta)
 * Responsibility: HTTP API for configuration and status
 */

package api

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/frustration-engine/internal/exporter"
)

// Handler handles HTTP requests
type Handler struct {
	engine *exporter.Engine
}

// NewHandler creates a new handler
func NewHandler(engine *exporter.Engine) *Handler {
	return &Handler{
		engine: engine,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// TriggerExport manually triggers export (for testing/admin)
func (h *Handler) TriggerExport(w http.ResponseWriter, r *http.Request) {
	// Trigger export of eligible incidents
	h.engine.ExportEligible(10) // Max 10 per manual trigger

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "export triggered"})
}