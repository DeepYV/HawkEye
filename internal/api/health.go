/**
 * Health Check Handler
 * 
 * Author: Grace Lee (Team Beta)
 * Responsibility: Health check endpoint
 */

package api

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/frustration-engine/internal/store"
)

// HealthHandler handles health checks
type HealthHandler struct {
	store *store.Store
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(store *store.Store) *HealthHandler {
	return &HealthHandler{store: store}
}

// HealthCheck handles health check requests
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Log-only mode
	if h.store.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"mode":    "log-only",
			"database": "not connected",
		})
		return
	}

	// Check database connection
	if err := h.store.DB.Ping(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	// Check if we can query
	var count int
	if err := h.store.DB.QueryRow("SELECT COUNT(*) FROM incidents").Scan(&count); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database query failed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"database": "connected",
	})
}