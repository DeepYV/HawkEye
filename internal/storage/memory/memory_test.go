package memory

import (
	"context"
	"testing"

	"github.com/your-org/frustration-engine/internal/types"
	pkgtypes "github.com/your-org/frustration-engine/pkg/types"
)

func TestEventStore_StoreAndCount(t *testing.T) {
	store := New()
	defer store.Close()

	ctx := context.Background()

	events := []types.Event{
		{EventType: "click", SessionID: "s1", Route: "/home"},
		{EventType: "scroll", SessionID: "s1", Route: "/home"},
	}

	if err := store.StoreEvents(ctx, "proj-1", events); err != nil {
		t.Fatalf("StoreEvents failed: %v", err)
	}

	if got := store.Count(); got != 2 {
		t.Errorf("Count() = %d, want 2", got)
	}
}

func TestEventStore_EmptyBatch(t *testing.T) {
	store := New()
	defer store.Close()

	if err := store.StoreEvents(context.Background(), "proj-1", nil); err != nil {
		t.Fatalf("StoreEvents with nil should not fail: %v", err)
	}
	if got := store.Count(); got != 0 {
		t.Errorf("Count() = %d after empty store, want 0", got)
	}
}

func TestIncidentStore_SaveAndQuery(t *testing.T) {
	store := NewIncidentStore()
	defer store.Close()

	ctx := context.Background()

	incident := pkgtypes.Incident{
		IncidentID:       "inc-1",
		SessionID:        "s1",
		ProjectID:        "proj-1",
		FrustrationScore: 85,
		ConfidenceLevel:  "High",
		ConfidenceScore:  90.0,
		Status:           "draft",
	}

	if err := store.Save(ctx, incident); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	results, err := store.Query(ctx, pkgtypes.Filter{ProjectID: "proj-1"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(results))
	}
	if results[0].IncidentID != "inc-1" {
		t.Errorf("expected incident ID 'inc-1', got '%s'", results[0].IncidentID)
	}
}

func TestIncidentStore_Deduplication(t *testing.T) {
	store := NewIncidentStore()
	ctx := context.Background()

	inc := pkgtypes.Incident{
		IncidentID:       "inc-dup",
		SessionID:        "s1",
		ProjectID:        "proj-1",
		FrustrationScore: 50,
	}

	// Save twice with same ID
	store.Save(ctx, inc)
	inc.FrustrationScore = 75
	store.Save(ctx, inc)

	results, _ := store.Query(ctx, pkgtypes.Filter{ProjectID: "proj-1"})
	if len(results) != 1 {
		t.Fatalf("expected 1 incident after dedup, got %d", len(results))
	}
	if results[0].FrustrationScore != 75 {
		t.Errorf("expected updated score 75, got %d", results[0].FrustrationScore)
	}
}

func TestIncidentStore_QueryFilter(t *testing.T) {
	store := NewIncidentStore()
	ctx := context.Background()

	store.Save(ctx, pkgtypes.Incident{IncidentID: "a", ProjectID: "proj-1", Status: "draft"})
	store.Save(ctx, pkgtypes.Incident{IncidentID: "b", ProjectID: "proj-2", Status: "confirmed"})
	store.Save(ctx, pkgtypes.Incident{IncidentID: "c", ProjectID: "proj-1", Status: "confirmed"})

	// Filter by project
	results, _ := store.Query(ctx, pkgtypes.Filter{ProjectID: "proj-1"})
	if len(results) != 2 {
		t.Errorf("expected 2 results for proj-1, got %d", len(results))
	}

	// Filter by status
	results, _ = store.Query(ctx, pkgtypes.Filter{Status: "confirmed"})
	if len(results) != 2 {
		t.Errorf("expected 2 confirmed results, got %d", len(results))
	}

	// Filter with limit
	results, _ = store.Query(ctx, pkgtypes.Filter{Limit: 1})
	if len(results) != 1 {
		t.Errorf("expected 1 result with limit=1, got %d", len(results))
	}
}
