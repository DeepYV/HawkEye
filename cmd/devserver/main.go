/**
 * HawkEye Dev Server - Single-Port Unified Backend
 *
 * Combines all four services (Event Ingestion, Session Manager, UFSE,
 * Incident Store) behind a single HTTP port for local development.
 *
 * No external database dependencies required:
 *   - Events stored in memory (replaces ClickHouse)
 *   - Incidents logged to console (replaces PostgreSQL)
 *
 * Usage:
 *   go run ./cmd/devserver
 *   go run ./cmd/devserver --port 8080 --api-key my-key
 *
 * Environment variables:
 *   HAWKEYE_MODE=single-node   (default for this binary)
 *   ENVIRONMENT=development    (default for this binary)
 *   PORT=8080                  (server port)
 *   TEST_API_KEY=...           (API key for SDK authentication)
 */

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/frustration-engine/internal/api"
	"github.com/your-org/frustration-engine/internal/auth"
	"github.com/your-org/frustration-engine/internal/config"
	"github.com/your-org/frustration-engine/internal/forwarding"
	"github.com/your-org/frustration-engine/internal/ratelimit"
	"github.com/your-org/frustration-engine/internal/security"
	"github.com/your-org/frustration-engine/internal/server/handlers"
	corsMiddleware "github.com/your-org/frustration-engine/internal/server"
	"github.com/your-org/frustration-engine/internal/session"
	"github.com/your-org/frustration-engine/internal/storage"
	"github.com/your-org/frustration-engine/internal/store"
	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/ufse"
)

func main() {
	port := flag.String("port", config.GetEnv("PORT", "8080"), "Server port")
	apiKey := flag.String("api-key", config.GetEnv("TEST_API_KEY", "dev-api-key"), "API key for authentication")
	storageMode := flag.String("storage", config.GetEnv("HAWKEYE_STORAGE", "memory"), "Event storage backend: memory or log-only")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("=============================================================")
	fmt.Println("  HawkEye Dev Server (single-port mode)")
	fmt.Println("=============================================================")
	fmt.Printf("  Port:        %s\n", *port)
	fmt.Printf("  API Key:     %s\n", *apiKey)
	fmt.Printf("  Storage:     %s (events), log-only (incidents)\n", *storageMode)
	fmt.Printf("  Mode:        HAWKEYE_MODE=single-node\n")
	fmt.Printf("  Environment: development\n")
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  Endpoints:")
	fmt.Printf("    POST http://localhost:%s/v1/events       (event ingestion)\n", *port)
	fmt.Printf("    GET  http://localhost:%s/v1/incidents     (query incidents)\n", *port)
	fmt.Printf("    GET  http://localhost:%s/health           (health check)\n", *port)
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  SDK config:")
	fmt.Printf("    ingestionUrl: 'http://localhost:%s'\n", *port)
	fmt.Printf("    apiKey:       '%s'\n", *apiKey)
	fmt.Println("=============================================================")

	// --- Event Storage (replaces ClickHouse) ---
	var eventStore storage.EventStore
	switch *storageMode {
	case "memory":
		eventStore = storage.NewMemoryStorage()
	default:
		// log-only: use ClickHouse storage in log-only mode
		ctx := context.Background()
		var err error
		s, err := storage.NewStorage(ctx, "log-only")
		if err != nil {
			log.Fatalf("Failed to create log-only storage: %v", err)
		}
		eventStore = s
	}

	// --- Auth ---
	authStore := auth.NewStore()
	authStore.AddAPIKey(*apiKey, "dev-project")
	log.Printf("[DevServer] API key registered: %s -> project 'dev-project'", *apiKey)

	// --- Session Manager (in-process) ---
	sessionManager := session.NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sessionManager.Start(ctx)

	// --- UFSE Engine (in-process, no HTTP forwarding) ---
	// Pass empty incident store URL so UFSE logs incidents instead of
	// forwarding via HTTP. We handle incident storage directly below.
	ufseEngine := ufse.NewEngine("")

	// --- Incident Store (log-only, no PostgreSQL needed) ---
	incidentStore, err := store.NewStore("log-only")
	if err != nil {
		log.Fatalf("Failed to create incident store: %v", err)
	}
	defer incidentStore.Close()
	storeEngine := store.NewStoreEngine(incidentStore)
	dedupEngine := store.NewDeduplicationEngine(incidentStore)

	// --- Forwarding: Event Ingestion → Session Manager (in-process) ---
	// We create a forwarding.Manager pointed at ourselves; but instead of
	// using HTTP, we'll handle the session-manager route internally.
	// For simplicity, we create the manager pointed at our own address.
	selfURL := fmt.Sprintf("http://localhost:%s", *port)
	forwarder := forwarding.NewManager(selfURL, 3)

	// --- Session → UFSE forwarding (in-process goroutine) ---
	go func() {
		emissionChan := sessionManager.GetEmissionChannel()
		for {
			select {
			case <-ctx.Done():
				return
			case sess := <-emissionChan:
				if sess == nil {
					continue
				}
				log.Printf("[DevServer] Processing session %s through UFSE pipeline", sess.SessionID)
				incidents := ufseEngine.ProcessSession(ctx, *sess)
				// Store incidents directly (in-process, no HTTP)
				for _, incident := range incidents {
					if err := storeEngine.StoreIncident(ctx, *incident); err != nil {
						log.Printf("[DevServer] Failed to store incident: %v", err)
					} else {
						log.Printf("[DevServer] Incident stored: %s (score: %d, confidence: %s)",
							incident.IncidentID, incident.FrustrationScore, incident.ConfidenceLevel)
					}
				}
			}
		}
	}()

	// --- HTTP Router (single port, all routes) ---
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	// CORS: allow all localhost origins in dev mode
	corsConfig := corsMiddleware.LocalDevCORSConfig()
	router.Use(corsMiddleware.CORSMiddleware(corsConfig))

	// Security headers
	router.Use(security.SecurityHeadersMiddleware)

	// Rate limiting (relaxed for dev)
	rateLimiter := ratelimit.NewLimiter(10000, 20000)
	router.Use(ratelimit.RateLimitMiddleware(rateLimiter))

	// --- Health endpoint (no auth required) ---
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"service": "hawkeye-devserver",
			"mode":    "single-node",
			"services": map[string]string{
				"event-ingestion": "embedded",
				"session-manager": "embedded",
				"ufse":            "embedded",
				"incident-store":  "embedded (log-only)",
			},
		})
	})

	// --- Authenticated routes ---
	router.Group(func(r chi.Router) {
		r.Use(auth.APIKeyAuth(authStore))

		// Event ingestion (Module 1)
		ingestHandler := handlers.NewHandler(eventStore, forwarder, authStore)
		r.Post("/v1/events", ingestHandler.IngestEvents)
	})

	// --- Session Manager internal route (for self-forwarding) ---
	sessionHandler := &devSessionHandler{sessionManager: sessionManager}
	router.Post("/v1/sessions/events", sessionHandler.ingestEvents)

	// --- Incident Store endpoints (Module 5) ---
	incidentIngestHandler := api.NewIngestHandler(storeEngine, dedupEngine)
	incidentQueryHandler := api.NewQueryHandler(storeEngine)
	router.Post("/v1/incidents", incidentIngestHandler.IngestIncident)
	router.Get("/v1/incidents", incidentQueryHandler.QueryIncidents)

	// --- Start HTTP Server ---
	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		fmt.Println("\n[DevServer] Shutting down...")
		cancel()
		sessionManager.Stop()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("[DevServer] Shutdown error: %v", err)
		}
	}()

	log.Printf("[DevServer] Listening on http://localhost:%s", *port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[DevServer] Server failed: %v", err)
	}

	fmt.Println("[DevServer] Stopped.")
}

// devSessionHandler handles session manager routes in the dev server.
type devSessionHandler struct {
	sessionManager *session.Manager
}

func (h *devSessionHandler) ingestEvents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID string       `json:"project_id"`
		SessionID string       `json:"session_id"`
		Events    []types.Event `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.sessionManager.AddEvents(req.ProjectID, req.SessionID, req.Events)
	w.WriteHeader(http.StatusOK)
}

// init parses CORS origins from environment (unused but keeps compat with
// the patterns established in the event-ingestion service).
func parseCORSOrigins() []string {
	var origins []string
	if originsEnv := config.GetEnv("CORS_ALLOWED_ORIGINS", ""); originsEnv != "" {
		for _, origin := range strings.Split(originsEnv, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				origins = append(origins, origin)
			}
		}
	}
	return origins
}
