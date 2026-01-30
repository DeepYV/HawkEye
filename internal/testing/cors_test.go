/**
 * CORS Middleware Tests
 *
 * Author: AI Issue Solver
 * Responsibility: Test CORS configuration for local development
 */

package testing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/your-org/frustration-engine/internal/server"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := server.DefaultCORSConfig()

	if !config.AllowLocalhost {
		t.Error("Expected AllowLocalhost to be true by default")
	}
	if len(config.AllowedMethods) == 0 {
		t.Error("Expected AllowedMethods to be non-empty")
	}
	if len(config.AllowedHeaders) == 0 {
		t.Error("Expected AllowedHeaders to be non-empty")
	}
}

func TestLocalDevCORSConfig(t *testing.T) {
	config := server.LocalDevCORSConfig()

	if !config.AllowLocalhost {
		t.Error("Expected AllowLocalhost to be true for local dev")
	}
	if len(config.AllowedOrigins) == 0 {
		t.Error("Expected AllowedOrigins to include common localhost ports")
	}

	// Check for common development ports
	hasPort3000 := false
	hasPort5173 := false
	for _, origin := range config.AllowedOrigins {
		if origin == "http://localhost:3000" {
			hasPort3000 = true
		}
		if origin == "http://localhost:5173" {
			hasPort5173 = true
		}
	}

	if !hasPort3000 {
		t.Error("Expected http://localhost:3000 in allowed origins")
	}
	if !hasPort5173 {
		t.Error("Expected http://localhost:5173 in allowed origins")
	}
}

func TestProductionCORSConfig(t *testing.T) {
	origins := []string{"https://app.example.com", "https://www.example.com"}
	config := server.ProductionCORSConfig(origins)

	if config.AllowLocalhost {
		t.Error("Expected AllowLocalhost to be false in production")
	}
	if len(config.AllowedOrigins) != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", len(config.AllowedOrigins))
	}
}

func TestCORSMiddlewareAllowsValidOrigin(t *testing.T) {
	config := server.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowLocalhost:   false,
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	handler := server.CORSMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected CORS header to be set for allowed origin")
	}
	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Expected credentials header to be set")
	}
}

func TestCORSMiddlewareBlocksInvalidOrigin(t *testing.T) {
	config := server.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowLocalhost:   false,
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}

	handler := server.CORSMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Expected no CORS header for disallowed origin")
	}
}

func TestCORSMiddlewareAllowsLocalhost(t *testing.T) {
	config := server.CORSConfig{
		AllowedOrigins: []string{},
		AllowLocalhost: true, // Enable localhost
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}

	handler := server.CORSMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test various localhost formats
	localhostOrigins := []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"http://127.0.0.1:3000",
		"http://localhost:5173",
	}

	for _, origin := range localhostOrigins {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", origin)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Access-Control-Allow-Origin") != origin {
			t.Errorf("Expected CORS header for localhost origin %s", origin)
		}
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	config := server.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowLocalhost:   true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-API-Key"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	handler := server.CORSMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for preflight")
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for preflight, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header for preflight")
	}
	if rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Expected Access-Control-Allow-Headers header for preflight")
	}
}

func TestGetCORSConfigFromEnvDevelopment(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	config := server.GetCORSConfigFromEnv()

	if !config.AllowLocalhost {
		t.Error("Expected AllowLocalhost to be true in development")
	}
}

func TestGetCORSConfigFromEnvProduction(t *testing.T) {
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,https://www.example.com")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
	}()

	config := server.GetCORSConfigFromEnv()

	if config.AllowLocalhost {
		t.Error("Expected AllowLocalhost to be false in production")
	}
	if len(config.AllowedOrigins) != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", len(config.AllowedOrigins))
	}
}

func TestGetCORSConfigFromEnvWithAdditionalOrigins(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://custom:4000")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
	}()

	config := server.GetCORSConfigFromEnv()

	hasCustomOrigin := false
	for _, origin := range config.AllowedOrigins {
		if origin == "http://custom:4000" {
			hasCustomOrigin = true
			break
		}
	}

	if !hasCustomOrigin {
		t.Error("Expected custom origin to be added to allowed origins")
	}
}

func TestLocalDevelopmentMiddleware(t *testing.T) {
	handler := server.LocalDevelopmentMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Environment") != "development" {
		t.Error("Expected X-Environment header for development middleware")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("Expected CORS header for localhost")
	}
}

func TestWildcardSubdomainMatching(t *testing.T) {
	config := server.CORSConfig{
		AllowedOrigins: []string{"*.example.com"},
		AllowLocalhost: false,
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}

	handler := server.CORSMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test subdomain
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Error("Expected wildcard subdomain to be allowed")
	}
}
