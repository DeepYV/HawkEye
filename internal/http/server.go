// Package http provides the HTTP server and routing for HawkEye.
//
// All routes are mounted on a single port. The server exposes:
//   - POST /v1/events    — event ingestion from SDK
//   - GET  /v1/issues    — ranked issues
//   - GET  /v1/flows     — flow summaries
//   - GET  /             — dashboard
//   - GET  /v1/incidents — query detected incidents
//   - GET  /health       — health check
//   - GET  /metrics      — Prometheus metrics
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/your-org/frustration-engine/internal/incident"
	"github.com/your-org/frustration-engine/internal/ingest"
	"github.com/your-org/frustration-engine/internal/issues"
	"github.com/your-org/frustration-engine/internal/metrics"
	"github.com/your-org/frustration-engine/pkg/types"
	"github.com/your-org/frustration-engine/web"
)

type ctxKey int

const projectIDKey ctxKey = iota

// Server wraps the HTTP server with HawkEye route handlers.
type Server struct {
	router    chi.Router
	ingest    *ingest.Handler
	incidents *incident.Service
	issues    *issues.Store
	apiKey    string
}

// NewServer creates a new HTTP server with all routes configured.
func NewServer(ingestHandler *ingest.Handler, incidentSvc *incident.Service, issueStore *issues.Store, apiKey string, devMode bool) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		ingest:    ingestHandler,
		incidents: incidentSvc,
		issues:    issueStore,
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
	s.router.Get("/", s.handleIndex)
	s.router.Get("/app.js", s.handleAppJS)
	s.router.Get("/v1/issues", s.handleListIssues)
	s.router.Get("/v1/issues/{id}", s.handleGetIssue)
	s.router.Get("/v1/flows", s.handleFlows)
	s.router.Options("/v1/events", s.handleEventsPreflight)

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
	// The browser reads this response after a cross-origin preflight.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.IngestResponse{
			Success: false, Message: "invalid request body",
		})
		return
	}

	events, legacy, err := issues.DecodeBatch(body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, types.IngestResponse{
			Success: false, Message: "invalid request body",
		})
		return
	}
	if legacy {
		var req types.IngestRequest
		if err := json.Unmarshal(body, &req); err != nil {
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
		return
	}

	if s.issues == nil {
		writeJSON(w, http.StatusInternalServerError, types.IngestResponse{
			Success: false, Message: "storage error",
		})
		return
	}
	processed := s.issues.Apply(time.Now(), events)
	writeJSON(w, http.StatusOK, types.IngestResponse{
		Success:   true,
		Processed: processed,
	})
}

func (s *Server) handleListIssues(w http.ResponseWriter, r *http.Request) {
	window, ok := parseDayWindow(r.URL.Query().Get("window"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid window"})
		return
	}
	if s.issues == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	result := s.issues.List(issues.Query{
		Now:        time.Now(),
		Window:     window,
		IncludeLow: r.URL.Query().Get("include") == "low",
		Limit:      parseLimit(r.URL.Query().Get("limit")),
	})
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetIssue(w http.ResponseWriter, r *http.Request) {
	id, ok := normalizeIssueID(chi.URLParam(r, "id"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if s.issues == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	iss, found := s.issues.Get(time.Now(), id, 7*24*time.Hour)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, iss)
}

func (s *Server) handleFlows(w http.ResponseWriter, r *http.Request) {
	window, ok := parseDayWindow(r.URL.Query().Get("window"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid window"})
		return
	}
	if s.issues == nil {
		writeJSON(w, http.StatusOK, issues.FlowReport{WindowMs: window.Milliseconds(), Flows: []issues.FlowSummary{}})
		return
	}
	writeJSON(w, http.StatusOK, s.issues.Flows(time.Now(), window))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	serveEmbedded(w, "index.html", "text/html; charset=utf-8")
}

func (s *Server) handleAppJS(w http.ResponseWriter, r *http.Request) {
	serveEmbedded(w, "app.js", "application/javascript; charset=utf-8")
}

func (s *Server) handleEventsPreflight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.WriteHeader(http.StatusNoContent)
}

func serveEmbedded(w http.ResponseWriter, name, contentType string) {
	b, err := web.Files.ReadFile(name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func parseDayWindow(v string) (time.Duration, bool) {
	if v == "" {
		v = "7d"
	}
	if len(v) != 2 || v[1] != 'd' || v[0] < '1' || v[0] > '7' {
		return 0, false
	}
	days := int(v[0] - '0')
	return time.Duration(days) * 24 * time.Hour, true
}

func parseLimit(v string) int {
	if v == "" {
		return 50
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 50
	}
	if n > 200 {
		return 200
	}
	return n
}

func normalizeIssueID(id string) (string, bool) {
	if len(id) != 64 {
		return "", false
	}
	for _, c := range id {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return "", false
		}
	}
	return strings.ToLower(id), true
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
