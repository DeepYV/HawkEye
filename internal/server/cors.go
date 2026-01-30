/**
 * CORS Middleware for Local Development
 *
 * Author: AI Issue Solver
 * Responsibility: Handle CORS for local development integration
 *
 * Features:
 * - Auto-detect development environment
 * - Support localhost on multiple ports
 * - Configurable allowed origins
 * - Secure defaults for production
 */

package server

import (
	"net/http"
	"os"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins (can include localhost)
	AllowedOrigins []string

	// AllowLocalhost enables automatic localhost support for development
	AllowLocalhost bool

	// AllowedMethods is the list of allowed HTTP methods
	AllowedMethods []string

	// AllowedHeaders is the list of allowed request headers
	AllowedHeaders []string

	// ExposeHeaders is the list of headers exposed to the browser
	ExposeHeaders []string

	// AllowCredentials indicates if credentials (cookies, auth) are allowed
	AllowCredentials bool

	// MaxAge is the preflight cache duration in seconds
	MaxAge int
}

// DefaultCORSConfig returns the default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{},
		AllowLocalhost:   true, // Enable by default for easier local dev
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour preflight cache
	}
}

// LocalDevCORSConfig returns CORS configuration optimized for local development
func LocalDevCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",  // React default
			"http://localhost:3001",  // Alternative React
			"http://localhost:5173",  // Vite default
			"http://localhost:5174",  // Vite alternative
			"http://localhost:8080",  // Vue default
			"http://localhost:4200",  // Angular default
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		},
		AllowLocalhost:   true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key", "X-Request-ID", "X-Requested-With"},
		ExposeHeaders:    []string{"X-Request-ID", "X-Session-ID"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours for development
	}
}

// ProductionCORSConfig returns CORS configuration for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowedOrigins:   allowedOrigins,
		AllowLocalhost:   false, // Disable localhost in production
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := isOriginAllowed(origin, config)

			if allowed {
				// Set CORS headers
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Handle preflight
				if r.Method == http.MethodOptions {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
					if len(config.ExposeHeaders) > 0 {
						w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
					}
					if config.MaxAge > 0 {
						w.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}

				// Set expose headers for non-preflight requests
				if len(config.ExposeHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if an origin is allowed
func isOriginAllowed(origin string, config CORSConfig) bool {
	if origin == "" {
		return false
	}

	// Check explicit allowed origins
	for _, allowed := range config.AllowedOrigins {
		if origin == allowed {
			return true
		}
		// Support wildcard subdomain matching
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	// Check localhost if enabled
	if config.AllowLocalhost && isLocalhostOrigin(origin) {
		return true
	}

	return false
}

// isLocalhostOrigin checks if an origin is a localhost origin
func isLocalhostOrigin(origin string) bool {
	localhostPatterns := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"http://0.0.0.0",
		"https://0.0.0.0",
	}

	for _, pattern := range localhostPatterns {
		if strings.HasPrefix(origin, pattern) {
			return true
		}
	}

	return false
}

// GetCORSConfigFromEnv loads CORS configuration from environment variables
func GetCORSConfigFromEnv() CORSConfig {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("HAWKEYE_ENVIRONMENT")
	}

	// Development environment
	if env == "development" || env == "dev" || env == "local" {
		config := LocalDevCORSConfig()

		// Allow additional origins from environment
		if additionalOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); additionalOrigins != "" {
			origins := strings.Split(additionalOrigins, ",")
			for _, origin := range origins {
				origin = strings.TrimSpace(origin)
				if origin != "" {
					config.AllowedOrigins = append(config.AllowedOrigins, origin)
				}
			}
		}

		return config
	}

	// Production environment
	config := DefaultCORSConfig()
	config.AllowLocalhost = false // Disable localhost in production

	// Load allowed origins from environment
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		originList := strings.Split(origins, ",")
		for _, origin := range originList {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				config.AllowedOrigins = append(config.AllowedOrigins, origin)
			}
		}
	}

	return config
}

// LocalDevelopmentMiddleware combines CORS and other development-friendly settings
func LocalDevelopmentMiddleware() func(http.Handler) http.Handler {
	corsConfig := LocalDevCORSConfig()

	return func(next http.Handler) http.Handler {
		corsHandler := CORSMiddleware(corsConfig)(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add development-friendly headers
			w.Header().Set("X-Environment", "development")
			w.Header().Set("X-HawkEye-Version", "dev")

			corsHandler.ServeHTTP(w, r)
		})
	}
}
