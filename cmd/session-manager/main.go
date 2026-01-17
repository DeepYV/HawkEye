/**
 * Session Manager - Main Entry Point
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
	"github.com/your-org/frustration-engine/internal/forwarding"
	"github.com/your-org/frustration-engine/internal/session"
	"github.com/your-org/frustration-engine/internal/types"
)

func main() {
	// Configuration
	port := flag.String("port", config.GetEnv("PORT", "8081"), "Server port")
	ufseURL := flag.String("ufse-url", config.GetEnv("UFSE_URL", "http://localhost:8082"), "UFSE URL")
	flag.Parse()

	// Create session manager
	sessionManager := session.NewManager()

	// Start session manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionManager.Start(ctx)

	// Setup forwarding
	forwarder := forwarding.NewForwarder(*ufseURL)

	// Start forwarding loop
	go forwardSessions(ctx, sessionManager, forwarder)

	// Setup HTTP server with timeouts
	handler := setupSessionRoutes(sessionManager)
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

		log.Println("Shutting down Session Manager...")
		cancel()
		sessionManager.Stop()
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("Starting Session Manager on port %s", *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

// forwardSessions forwards completed sessions to UFSE
func forwardSessions(ctx context.Context, sessionManager *session.Manager, forwarder *forwarding.Forwarder) {
	emissionChan := sessionManager.GetEmissionChannel()

	for {
		select {
		case <-ctx.Done():
			return
		case session := <-emissionChan:
			if err := forwarder.ForwardSession(ctx, session); err != nil {
				log.Printf("Failed to forward session %s: %v", session.SessionID, err)
				// Continue processing other sessions
			}
		}
	}
}

// setupSessionRoutes sets up routes for Session Manager
func setupSessionRoutes(sessionManager *session.Manager) http.Handler {
	handler := &sessionHandler{sessionManager: sessionManager}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "session-manager"})
	})
	router.Post("/v1/sessions/events", handler.ingestEvents)
	return router
}

type sessionHandler struct {
	sessionManager *session.Manager
}

func (h *sessionHandler) ingestEvents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID string          `json:"project_id"`
		SessionID string          `json:"session_id"`
		Events    []types.Event   `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.sessionManager.AddEvents(req.ProjectID, req.SessionID, req.Events)
	w.WriteHeader(http.StatusOK)
}