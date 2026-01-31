package storage_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/frustration-engine/internal/auth"
	"github.com/your-org/frustration-engine/internal/server/handlers"
	"github.com/your-org/frustration-engine/internal/forwarding"
	"github.com/your-org/frustration-engine/internal/storage"
	"github.com/your-org/frustration-engine/internal/types"
)

// TestDevServerIntegration verifies the single-port flow:
// SDK → Event Ingestion (with memory storage) → responds 200.
func TestDevServerIntegration(t *testing.T) {
	// Setup in-memory storage
	memStore := storage.NewMemoryStorage()

	// Setup auth
	authStore := auth.NewStore()
	authStore.AddAPIKey("test-key", "test-project")

	// Get a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	selfURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	forwarder := forwarding.NewManager(selfURL, 1)
	handler := handlers.NewHandler(memStore, forwarder, authStore)

	// Setup router (minimal, like dev server)
	router := chi.NewRouter()
	router.Use(chiMiddleware.Recoverer)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	router.Group(func(r chi.Router) {
		r.Use(auth.APIKeyAuth(authStore))
		r.Post("/v1/events", handler.IngestEvents)
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// --- Test health endpoint ---
	resp, err := http.Get(selfURL + "/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /health, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// --- Test event ingestion ---
	payload := types.IngestRequest{
		APIKey: "test-key",
		Events: []types.Event{
			{
				EventType: "click",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-1",
				Route:     "/dashboard",
				Target:    types.EventTarget{Type: "button", ID: "submit-btn"},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", selfURL+"/v1/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("event ingestion request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /v1/events, got %d", resp.StatusCode)
	}

	var result types.IngestResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Fatal("expected success=true in response")
	}

	// Give async storage a moment to complete
	time.Sleep(200 * time.Millisecond)

	// Verify events were stored in memory
	if memStore.Count() != 1 {
		t.Fatalf("expected 1 event in memory storage, got %d", memStore.Count())
	}

	stored := memStore.GetEvents()
	if stored[0].Event.EventType != "click" {
		t.Fatalf("expected event type 'click', got %q", stored[0].Event.EventType)
	}
	if stored[0].ProjectID != "test-project" {
		t.Fatalf("expected project ID 'test-project', got %q", stored[0].ProjectID)
	}
}
