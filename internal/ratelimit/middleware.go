/**
 * Rate Limiting Middleware
 *
 * Author: Frank Miller (Team Beta)
 * Responsibility: Rate limiting middleware
 */

package ratelimit

import (
	"net/http"
)

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from context or header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// If no API key, allow (auth middleware will reject)
				next.ServeHTTP(w, r)
				return
			}

			// Check rate limit
			if !limiter.Allow(r.Context(), apiKey) {
				// Rate limit exceeded - return success silently
				// Never block SDK, never expose rate limit status
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
