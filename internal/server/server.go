/**
 * HTTP Server Setup
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: HTTP server setup & routing
 */

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/frustration-engine/internal/auth"
	"github.com/your-org/frustration-engine/internal/forwarding"
	"github.com/your-org/frustration-engine/internal/ratelimit"
	"github.com/your-org/frustration-engine/internal/security"
	"github.com/your-org/frustration-engine/internal/server/handlers"
	"github.com/your-org/frustration-engine/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	router      *chi.Mux
	authStore   *auth.Store
	storage     *storage.Storage
	forwarder   *forwarding.Manager
	rateLimiter *ratelimit.Limiter
}

// Config holds server configuration
type Config struct {
	Port              string
	ClickHouseDSN     string
	SessionManagerURL string
	RateLimitRPS      int
	RateLimitBurst    int
	Environment       string // "development", "staging", "production"
	CORSAllowedOrigins []string
}

// NewServer creates a new server instance
func NewServer(cfg Config) (*Server, error) {
	// Initialize components
	authStore := auth.NewStore()
	
	// Load API keys from environment (production) or use test key for development
	// In production, API keys should be loaded from database
	testAPIKey := os.Getenv("TEST_API_KEY")
	if testAPIKey != "" {
		// Only add test API key if explicitly set in environment (for local dev/testing)
		authStore.AddAPIKey(testAPIKey, "test-project")
		log.Printf("[Server] Test API key loaded from environment (development mode)")
	}

	// Initialize ClickHouse storage
	ctx := context.Background()
	storage, err := storage.NewStorage(ctx, cfg.ClickHouseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize forwarding manager
	forwarder := forwarding.NewManager(cfg.SessionManagerURL, 5)

	// Initialize rate limiter
	rateLimiter := ratelimit.NewLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	// Create handler
	handler := handlers.NewHandler(storage, forwarder, authStore)

	// Setup router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(10 * time.Second))

	// CORS middleware (before security headers for preflight support)
	corsConfig := getCORSConfig(cfg)
	router.Use(CORSMiddleware(corsConfig))

	// Security headers middleware
	router.Use(security.SecurityHeadersMiddleware)

	// Rate limiting (before auth to save resources)
	router.Use(ratelimit.RateLimitMiddleware(rateLimiter))

	// Authentication
	router.Use(auth.APIKeyAuth(authStore))

	// Routes
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"event-ingestion"}`)
	})
	router.Post("/v1/events", handler.IngestEvents)

	return &Server{
		router:      router,
		authStore:   authStore,
		storage:     storage,
		forwarder:   forwarder,
		rateLimiter: rateLimiter,
	}, nil
}

// Start starts the HTTP server with graceful shutdown support
func (s *Server) Start(port string) error {
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down Event Ingestion API...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	return server.ListenAndServe()
}

// getCORSConfig returns the appropriate CORS configuration based on environment
func getCORSConfig(cfg Config) CORSConfig {
	// Check environment from config or environment variable
	env := cfg.Environment
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = os.Getenv("HAWKEYE_ENVIRONMENT")
	}

	// Development environment - allow localhost
	if env == "development" || env == "dev" || env == "local" {
		config := LocalDevCORSConfig()
		// Add any additional configured origins
		if len(cfg.CORSAllowedOrigins) > 0 {
			config.AllowedOrigins = append(config.AllowedOrigins, cfg.CORSAllowedOrigins...)
		}
		log.Printf("[Server] CORS enabled for local development (localhost allowed)")
		return config
	}

	// Production environment - only explicit origins
	config := DefaultCORSConfig()
	config.AllowLocalhost = false
	config.AllowedOrigins = cfg.CORSAllowedOrigins
	return config
}
