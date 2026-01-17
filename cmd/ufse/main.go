/**
 * UFSE - Main Entry Point
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Application bootstrap
 */

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/frustration-engine/internal/config"
	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/ufse"
)

func main() {
	// Configuration
	port := flag.String("port", config.GetEnv("PORT", "8082"), "Server port")
	incidentStoreURL := flag.String("incident-store-url", config.GetEnv("INCIDENT_STORE_URL", ""), "Incident Store URL")
	flag.Parse()

	// Create UFSE engine
	engine := ufse.NewEngine(*incidentStoreURL)
	
	if *incidentStoreURL == "" {
		log.Println("WARNING: Incident Store URL not configured, incidents will be logged only")
	}

	// Setup HTTP server with timeouts
	handler := setupUFSERoutes(engine)
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

		log.Println("Shutting down UFSE...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("Starting UFSE on port %s", *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

// setupUFSERoutes sets up routes for UFSE
func setupUFSERoutes(engine *ufse.Engine) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "ufse"})
	})
	router.Post("/v1/sessions/process", func(w http.ResponseWriter, r *http.Request) {
		var session types.Session
		if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
			log.Printf("[UFSE] Failed to decode session: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
			return
		}
		
		// Process session with request context
		incidents := engine.ProcessSession(r.Context(), session)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count": len(incidents),
			"incidents": incidents,
		})
	})
	return router
}