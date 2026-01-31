package storage

import (
	"context"
	"sync"
	"testing"

	"github.com/your-org/frustration-engine/internal/types"
)

func TestNewMemoryStorage(t *testing.T) {
	m := NewMemoryStorage()
	if m == nil {
		t.Fatal("expected non-nil MemoryStorage")
	}
	if m.Count() != 0 {
		t.Fatalf("expected 0 events, got %d", m.Count())
	}
}

func TestMemoryStorage_StoreEvents(t *testing.T) {
	m := NewMemoryStorage()
	ctx := context.Background()

	events := []types.Event{
		{EventType: "click", SessionID: "s1", Route: "/home", Target: types.EventTarget{Type: "button"}},
		{EventType: "scroll", SessionID: "s1", Route: "/home", Target: types.EventTarget{Type: "page"}},
	}

	if err := m.StoreEvents(ctx, "proj-1", events); err != nil {
		t.Fatalf("StoreEvents failed: %v", err)
	}

	if m.Count() != 2 {
		t.Fatalf("expected 2 events, got %d", m.Count())
	}

	// Verify events are retrievable
	stored := m.GetEvents()
	if len(stored) != 2 {
		t.Fatalf("expected 2 stored events, got %d", len(stored))
	}
	if stored[0].ProjectID != "proj-1" {
		t.Fatalf("expected project ID 'proj-1', got %q", stored[0].ProjectID)
	}
	if stored[0].Event.EventType != "click" {
		t.Fatalf("expected event type 'click', got %q", stored[0].Event.EventType)
	}
}

func TestMemoryStorage_StoreEventsEmpty(t *testing.T) {
	m := NewMemoryStorage()
	ctx := context.Background()

	if err := m.StoreEvents(ctx, "proj-1", nil); err != nil {
		t.Fatalf("StoreEvents with nil should succeed: %v", err)
	}
	if m.Count() != 0 {
		t.Fatal("expected 0 events after storing nil")
	}
}

func TestMemoryStorage_GetEventsByProject(t *testing.T) {
	m := NewMemoryStorage()
	ctx := context.Background()

	m.StoreEvents(ctx, "proj-a", []types.Event{
		{EventType: "click", SessionID: "s1"},
	})
	m.StoreEvents(ctx, "proj-b", []types.Event{
		{EventType: "scroll", SessionID: "s2"},
		{EventType: "error", SessionID: "s2"},
	})

	a := m.GetEventsByProject("proj-a")
	if len(a) != 1 {
		t.Fatalf("expected 1 event for proj-a, got %d", len(a))
	}

	b := m.GetEventsByProject("proj-b")
	if len(b) != 2 {
		t.Fatalf("expected 2 events for proj-b, got %d", len(b))
	}

	c := m.GetEventsByProject("proj-c")
	if len(c) != 0 {
		t.Fatalf("expected 0 events for proj-c, got %d", len(c))
	}
}

func TestMemoryStorage_Close(t *testing.T) {
	m := NewMemoryStorage()
	ctx := context.Background()

	m.StoreEvents(ctx, "proj-1", []types.Event{
		{EventType: "click", SessionID: "s1"},
	})

	if err := m.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	m := NewMemoryStorage()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.StoreEvents(ctx, "proj-1", []types.Event{
				{EventType: "click", SessionID: "s1"},
			})
		}(i)
	}
	wg.Wait()

	if m.Count() != 100 {
		t.Fatalf("expected 100 events after concurrent writes, got %d", m.Count())
	}
}

func TestMemoryStorage_ImplementsEventStore(t *testing.T) {
	// Compile-time check that MemoryStorage implements EventStore
	var _ EventStore = (*MemoryStorage)(nil)
}

func TestClickHouseStorage_ImplementsEventStore(t *testing.T) {
	// Compile-time check that ClickHouse Storage implements EventStore
	var _ EventStore = (*Storage)(nil)
}
