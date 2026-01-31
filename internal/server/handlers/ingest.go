/**
 * Event Ingestion Handler
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: HTTP handler for event ingestion
 */

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/your-org/frustration-engine/internal/auth"
	"github.com/your-org/frustration-engine/internal/forwarding"
	"github.com/your-org/frustration-engine/internal/storage"
	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/validation"
)

// Handler handles HTTP requests
type Handler struct {
	storage   storage.EventStore
	forwarder *forwarding.Manager
	authStore *auth.Store
}

// NewHandler creates a new handler
func NewHandler(storage storage.EventStore, forwarder *forwarding.Manager, authStore *auth.Store) *Handler {
	return &Handler{
		storage:   storage,
		forwarder: forwarder,
		authStore: authStore,
	}
}

// IngestEvents handles event ingestion requests
func (h *Handler) IngestEvents(w http.ResponseWriter, r *http.Request) {
	// Get project ID from context (set by auth middleware)
	projectID := auth.GetProjectID(r.Context())
	if projectID == "" {
		// Should not happen if auth middleware works correctly
		respondSuccess(w, 0)
		return
	}

	// Decode request
	var req types.IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Malformed request - drop silently
		respondSuccess(w, 0)
		return
	}

	// Validate batch
	validEvents, validationErrors := validation.ValidateBatch(req.Events)
	if len(validationErrors) > 0 {
		// Some events invalid - continue with valid ones
		// Errors are logged internally but not exposed
	}

	// Re-validate privacy
	privacyValidEvents, privacyErrors := validation.ValidatePrivacy(validEvents)
	if len(privacyErrors) > 0 {
		// Events with PII dropped
		// Errors are logged internally but not exposed
	}

	if len(privacyValidEvents) == 0 {
		// No valid events - return success anyway
		respondSuccess(w, 0)
		return
	}

	// Persist events (async, non-blocking)
	go func() {
		ctx := r.Context()
		if err := h.storage.StoreEvents(ctx, projectID, privacyValidEvents); err != nil {
			// Log error with context (not exposed to client)
			log.Printf("[Event Ingestion] Failed to store events: project_id=%s, event_count=%d, error=%v",
				projectID, len(privacyValidEvents), err)
			// Events are dropped on storage failure
		}
	}()

	// Forward to Session Manager (async, non-blocking)
	go func() {
		ctx := r.Context()
		// Use request context so operation cancels if request is cancelled
		h.forwarder.ForwardEvents(ctx, projectID, privacyValidEvents)
	}()

	// Return success immediately
	respondSuccess(w, len(privacyValidEvents))
}

// respondSuccess returns generic success response
func respondSuccess(w http.ResponseWriter, processed int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := types.IngestResponse{
		Success:   true,
		Processed: processed,
	}

	json.NewEncoder(w).Encode(response)
}
