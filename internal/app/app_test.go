package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/frustration-engine/internal/config"
	"github.com/your-org/frustration-engine/pkg/types"
)

func TestApp_HealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:   "0",
		APIKey: "test-key",
		Dev:    true,
	}

	application := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application.Start(ctx)
	defer application.Stop()

	srv := httptest.NewServer(application.Server.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("health status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "healthy" {
		t.Errorf("health status = %v, want 'healthy'", body["status"])
	}
}

func TestApp_MetricsEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:   "0",
		APIKey: "test-key",
		Dev:    true,
	}

	application := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application.Start(ctx)
	defer application.Stop()

	srv := httptest.NewServer(application.Server.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("metrics request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("metrics status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestApp_IngestEvents(t *testing.T) {
	cfg := &config.Config{
		Port:   "0",
		APIKey: "test-key",
		Dev:    true,
	}

	application := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application.Start(ctx)
	defer application.Stop()

	srv := httptest.NewServer(application.Server.Handler())
	defer srv.Close()

	// Send events
	payload := types.IngestRequest{
		APIKey: "test-key",
		Events: []types.Event{
			{
				EventType: "click",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "session-1",
				Route:     "/home",
				Target:    types.EventTarget{Type: "button", ID: "cta"},
			},
			{
				EventType: "scroll",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "session-1",
				Route:     "/home",
			},
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", srv.URL+"/v1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("ingest request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ingest status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result types.IngestResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if !result.Success {
		t.Errorf("ingest success = false, want true")
	}
	if result.Processed != 2 {
		t.Errorf("ingest processed = %d, want 2", result.Processed)
	}
}

func TestApp_IngestUnauthorized(t *testing.T) {
	cfg := &config.Config{
		Port:   "0",
		APIKey: "correct-key",
		Dev:    true,
	}

	application := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application.Start(ctx)
	defer application.Stop()

	srv := httptest.NewServer(application.Server.Handler())
	defer srv.Close()

	payload := types.IngestRequest{Events: []types.Event{{EventType: "click", SessionID: "s1"}}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", srv.URL+"/v1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "wrong-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestApp_QueryIncidents(t *testing.T) {
	cfg := &config.Config{
		Port:   "0",
		APIKey: "test-key",
		Dev:    true,
	}

	application := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application.Start(ctx)
	defer application.Stop()

	srv := httptest.NewServer(application.Server.Handler())
	defer srv.Close()

	// Query incidents (should be empty initially)
	resp, err := http.Get(srv.URL + "/v1/incidents")
	if err != nil {
		t.Fatalf("query request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("query status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result types.QueryResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Total != 0 {
		t.Errorf("expected 0 incidents initially, got %d", result.Total)
	}
}
