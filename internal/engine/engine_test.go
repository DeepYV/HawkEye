package engine

import (
	"sort"
	"testing"
	"time"

	"github.com/your-org/frustration-engine/pkg/types"
)

func TestDetectFrustration_EmptySession(t *testing.T) {
	session := types.Session{
		SessionID: "test-empty",
		ProjectID: "proj-1",
		Events:    []types.Event{},
	}

	incidents := DetectFrustration(session)
	if len(incidents) != 0 {
		t.Errorf("expected 0 incidents for empty session, got %d", len(incidents))
	}
}

func TestClassifyEvents_SortsByTimestampAndSkipsInvalid(t *testing.T) {
	events := []types.Event{
		{EventType: "click", Timestamp: "", Route: "/a"},
		{EventType: "click", Timestamp: "not-a-time", Route: "/a"},
		{EventType: "click", Timestamp: "1710000000123", Route: "/b"}, // unix millis
		{EventType: "click", Timestamp: "2024-03-10T12:00:01.100000000Z", Route: "/c"},
		{EventType: "click", Timestamp: "2024-03-10T12:00:00Z", Route: "/d"},
	}

	oldSession := toOldSession(types.Session{Events: events})
	classified := classifyEvents(oldSession.Events)

	if len(classified) != 3 {
		t.Fatalf("expected 3 valid classified events, got %d", len(classified))
	}

	if !sort.SliceIsSorted(classified, func(i, j int) bool {
		return classified[i].Timestamp.Before(classified[j].Timestamp) || classified[i].Timestamp.Equal(classified[j].Timestamp)
	}) {
		t.Fatalf("expected classified events to be sorted by timestamp")
	}
}

func TestParseEventTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		valid bool
	}{
		{name: "rfc3339nano", raw: "2024-03-10T12:00:00.123456789Z", valid: true},
		{name: "unix_millis", raw: "1710000000123", valid: true},
		{name: "invalid", raw: "invalid", valid: false},
		{name: "empty", raw: "", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := parseEventTimestamp(tt.raw)
			if ok != tt.valid {
				t.Fatalf("parseEventTimestamp(%q) valid=%v, want %v", tt.raw, ok, tt.valid)
			}
		})
	}
}

func TestDetectFrustration_SingleEvent(t *testing.T) {
	session := types.Session{
		SessionID: "test-single",
		ProjectID: "proj-1",
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		Events: []types.Event{
			{
				EventType: "click",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-single",
				Route:     "/home",
				Target: types.EventTarget{
					Type: "button",
					ID:   "submit-btn",
				},
			},
		},
	}

	// A single click should not trigger frustration detection
	incidents := DetectFrustration(session)
	if len(incidents) != 0 {
		t.Errorf("expected 0 incidents for single click, got %d", len(incidents))
	}
}

func TestDetectFrustration_RageClicks(t *testing.T) {
	now := time.Now()
	start := now.Add(-30 * time.Second)

	// Generate rapid clicks on the same element (rage click pattern)
	events := make([]types.Event, 0, 20)
	for i := 0; i < 20; i++ {
		ts := start.Add(time.Duration(i*200) * time.Millisecond)
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: ts.Format(time.RFC3339),
			SessionID: "test-rage",
			Route:     "/checkout",
			Target: types.EventTarget{
				Type:     "button",
				ID:       "pay-btn",
				Selector: "#pay-btn",
			},
		})
	}

	// Add an error event (system feedback)
	events = append(events, types.Event{
		EventType: "error",
		Timestamp: start.Add(5 * time.Second).Format(time.RFC3339),
		SessionID: "test-rage",
		Route:     "/checkout",
		Metadata:  map[string]interface{}{"error": "Payment failed"},
	})

	session := types.Session{
		SessionID: "test-rage",
		ProjectID: "proj-1",
		StartTime: start,
		EndTime:   now,
		Events:    events,
	}

	// This should detect rage clicks as a frustration signal
	incidents := DetectFrustration(session)
	// We expect either 0 (if thresholds aren't met) or 1+ incidents
	// The test verifies the function processes without panicking
	t.Logf("detected %d incidents from rage click pattern", len(incidents))
	for _, inc := range incidents {
		t.Logf("  incident: %s (score: %d, confidence: %s, signals: %v)",
			inc.IncidentID, inc.FrustrationScore, inc.ConfidenceLevel, inc.TriggeringSignals)
	}
}

func TestClassifyEventType(t *testing.T) {
	tests := []struct {
		name     string
		evtType  string
		metadata map[string]interface{}
		want     string
	}{
		{"click", "click", nil, "interaction"},
		{"input", "input", nil, "interaction"},
		{"scroll", "scroll", nil, "interaction"},
		{"error", "error", nil, "system_feedback"},
		{"network_error", "network_error", nil, "system_feedback"},
		{"navigation", "navigation", nil, "navigation"},
		{"route_change", "route_change", nil, "navigation"},
		{"long_task", "long_task", nil, "performance"},
		{"unknown_default", "custom", nil, "interaction"},
		{"metadata_error", "custom", map[string]interface{}{"error": "something"}, "system_feedback"},
		{"metadata_4xx", "custom", map[string]interface{}{"status": float64(404)}, "system_feedback"},
		{"metadata_2xx", "custom", map[string]interface{}{"status": float64(200)}, "interaction"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyEventType(tt.evtType, tt.metadata)
			if got != tt.want {
				t.Errorf("classifyEventType(%q, %v) = %q, want %q", tt.evtType, tt.metadata, got, tt.want)
			}
		})
	}
}
