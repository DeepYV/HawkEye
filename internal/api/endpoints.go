/**
 * API Endpoints - Ticket Exporter
 * 
 * Author: Henry Wilson (Team Beta)
 * Responsibility: API endpoint setup for Ticket Exporter
 */

package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/frustration-engine/internal/exporter"
)

// SetupRoutes sets up API routes for Ticket Exporter
func SetupRoutes(engine *exporter.Engine) http.Handler {
	handler := NewHandler(engine)

	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Routes
	router.Get("/health", handler.HealthCheck)
	router.Post("/v1/export/trigger", handler.TriggerExport)

	return router
}