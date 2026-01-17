/**
 * Authentication Middleware
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: API key authentication middleware
 */

package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const projectIDKey contextKey = "project_id"

// APIKeyAuth middleware validates API key from request
func APIKeyAuth(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from header or body
			apiKey := extractAPIKey(r)

			if apiKey == "" {
				// Silent rejection - return generic success
				respondSilentSuccess(w)
				return
			}

			// Validate API key
			projectID, err := store.ValidateAPIKey(r.Context(), apiKey)
			if err != nil {
				// Silent rejection - never expose validation errors
				respondSilentSuccess(w)
				return
			}

			// Add project ID to context
			ctx := context.WithValue(r.Context(), projectIDKey, projectID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetProjectID extracts project ID from context
func GetProjectID(ctx context.Context) string {
	if projectID, ok := ctx.Value(projectIDKey).(string); ok {
		return projectID
	}
	return ""
}

// extractAPIKey extracts API key from request
func extractAPIKey(r *http.Request) string {
	// Try header first
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	// Try Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// respondSilentSuccess returns generic success without exposing errors
func respondSilentSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
