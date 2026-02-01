// Package http provides the HTTP server and routing for HawkEye.
//
// All routes are mounted on a single port. The server exposes:
//   - POST /v1/events    — event ingestion from SDK
//   - GET  /v1/incidents — query detected incidents
//   - GET  /health       — health check
//   - GET  /metrics      — Prometheus metrics
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/your-org/frustration-engine/internal/incident"
	"github.com/your-org/frustration-engine/internal/ingest"
	"github.com/your-org/frustration-engine/internal/metrics"
	"github.com/your-org/frustration-engine/pkg/types"
)

type ctxKey int

const projectIDKey ctxKey = iota

// Server wraps the HTTP server with HawkEye route handlers.
type Server struct {
	router    chi.Router
	ingest    *ingest.Handler
	incidents *incident.Service
	apiKey    string
}

// NewServer creates a new HTTP server with all routes configured.
func NewServer(ingestHandler *ingest.Handler, incidentSvc *incident.Service, apiKey string, devMode bool) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		ingest:    ingestHandler,
		incidents: incidentSvc,
		apiKey:    apiKey,
	}

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(30 * time.Second))
	s.router.Use(s.metricsMiddleware)

	if devMode {
		s.router.Use(corsMiddleware)
	}

	// Public endpoints
	s.router.Get("/health", s.handleHealth)
	s.router.Handle("/metrics", promhttp.Handler())

	// Authenticated endpoints
	s.router.Group(func(r chi.Router) {
		r.Use(s.apiKeyAuth)
		r.Post("/v1/events", s.handleIngest)
	})

	// Incident query
	s.router.Get("/v1/incidents", s.handleQueryIncidents)

	return s
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("[http] listening on %s", addr)
	return srv.ListenAndServe()
}

// Handler returns the underlying http.Handler for testing or custom servers.
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "hawkeye",
		"mode":    "single-binary",
	})
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	var req types.IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, types.IngestResponse{
			Success: false, Message: "invalid request body",
		})
		return
	}

	pid, _ := r.Context().Value(projectIDKey).(string)
	if pid == "" {
		pid = req.AppID
	}
	if pid == "" {
		pid = "default"
	}

	processed, err := s.ingest.Ingest(r.Context(), pid, req.Events)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, types.IngestResponse{
			Success: false, Message: "storage error",
		})
		return
	}

	writeJSON(w, http.StatusOK, types.IngestResponse{
		Success:   true,
		Processed: processed,
	})
}

func (s *Server) handleQueryIncidents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := types.QueryRequest{
		ProjectID: q.Get("projectId"),
		Status:    q.Get("status"),
	}
	if v := q.Get("limit"); v != "" {
		filter.Limit, _ = strconv.Atoi(v)
	}
	if v := q.Get("offset"); v != "" {
		filter.Offset, _ = strconv.Atoi(v)
	}
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	incidents, total, err := s.incidents.Query(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}

	writeJSON(w, http.StatusOK, types.QueryResponse{
		Incidents: incidents,
		Total:     total,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	})
}

// --- middleware ---

func (s *Server) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}
		if key == "" {
			if auth := r.Header.Get("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
				key = auth[7:]
			}
		}

		if key != s.apiKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid API key"})
			return
		}

		ctx := context.WithValue(r.Context(), projectIDKey, "default")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		duration := time.Since(start).Seconds()

		path := r.URL.Path
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, fmt.Sprintf("%d", ww.Status())).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
