/**
 * Incident Store - Main Entry Point
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Application bootstrap
 */

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/frustration-engine/internal/api"
	"github.com/your-org/frustration-engine/internal/config"
	"github.com/your-org/frustration-engine/internal/store"
)

func main() {
	// Configuration
	port := flag.String("port", config.GetEnv("PORT", "8084"), "Server port")
	dsn := flag.String("dsn", config.GetEnv("DATABASE_URL", "log-only"), "PostgreSQL DSN (use 'log-only' for testing without DB)")
	flag.Parse()

	// Create store (will use log-only mode if DSN is empty or "log-only")
	storeInstance, err := store.NewStore(*dsn)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer storeInstance.Close()

	if *dsn == "" || *dsn == "log-only" {
		log.Println("Running in LOG-ONLY mode - incidents will be logged, not stored in database")
	}

	// Create store engine
	engine := store.NewStoreEngine(storeInstance)

	// Create deduplication engine
	dedup := store.NewDeduplicationEngine(storeInstance)

	// Setup HTTP server with timeouts
	handler := setupIncidentRoutes(engine, dedup, storeInstance)
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down Incident Store...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("Starting Incident Store on port %s", *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

// setupIncidentRoutes sets up routes for Incident Store
func setupIncidentRoutes(engine *store.StoreEngine, dedup *store.DeduplicationEngine, healthStore *store.Store) http.Handler {
	ingestHandler := api.NewIngestHandler(engine, dedup)
	queryHandler := api.NewQueryHandler(engine)
	healthHandler := api.NewHealthHandler(healthStore)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Get("/health", healthHandler.HealthCheck)
	router.Post("/v1/incidents", ingestHandler.IngestIncident)
	router.Get("/v1/incidents", queryHandler.QueryIncidents)
	return router
}