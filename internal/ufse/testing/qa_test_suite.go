/**
 * QA Test Suite for Enhanced Detection
 * 
 * Comprehensive test suite covering:
 * - Detection accuracy
 * - False positive/negative rates
 * - Performance benchmarks
 * - Edge cases
 */

package testing

import (
	"testing"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/ufse"
)

// TestDetectionAccuracy tests detection accuracy
func TestDetectionAccuracy(t *testing.T) {
	testCases := []struct {
		name           string
		session        types.Session
		expectedDetect bool
		expectedType   string
	}{
		{
			name: "High-strength rage clicks",
			session: createRageClickSession(5, 1500*time.Millisecond),
			expectedDetect: true,
			expectedType:   "rage",
		},
		{
			name: "Rage bait pattern",
			session: createRageBaitSession(),
			expectedDetect: true,
			expectedType:   "rage_bait",
		},
		{
			name: "Legitimate double-click",
			session: createDoubleClickSession(),
			expectedDetect: false,
		},
		{
			name: "Gaming rapid clicks",
			session: createGamingSession(),
			expectedDetect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			incidents := ufse.ProcessSessionEnhanced(tc.session)
			
			if tc.expectedDetect {
				if len(incidents) == 0 {
					t.Errorf("Expected detection but none found")
				} else if tc.expectedType != "" {
					// Check incident type
					found := false
					for _, incident := range incidents {
						for _, signal := range incident.TriggeringSignals {
							if signal == tc.expectedType {
								found = true
								break
							}
						}
					}
					if !found {
						t.Errorf("Expected signal type %s not found", tc.expectedType)
					}
				}
			} else {
				if len(incidents) > 0 {
					t.Errorf("Expected no detection but found %d incidents", len(incidents))
				}
			}
		})
	}
}

// TestFalsePositiveRate tests false positive rate
func TestFalsePositiveRate(t *testing.T) {
	legitimateScenarios := []types.Session{
		createDoubleClickSession(),
		createGamingSession(),
		createAccessibilitySession(),
		createSearchSession(),
		createComparisonShoppingSession(),
	}

	falsePositives := 0
	for _, session := range legitimateScenarios {
		incidents := ufse.ProcessSessionEnhanced(session)
		if len(incidents) > 0 {
			falsePositives++
		}
	}

	falsePositiveRate := float64(falsePositives) / float64(len(legitimateScenarios))
	if falsePositiveRate > 0.01 { // 1% threshold
		t.Errorf("False positive rate too high: %.2f%% (expected <1%%)", falsePositiveRate*100)
	}
}

// TestFalseNegativeRate tests false negative rate
func TestFalseNegativeRate(t *testing.T) {
	frustrationScenarios := []types.Session{
		createRageClickSession(5, 1500*time.Millisecond),
		createRageBaitSession(),
		createBlockedProgressSession(),
		createAbandonmentSession(),
	}

	falseNegatives := 0
	for _, session := range frustrationScenarios {
		incidents := ufse.ProcessSessionEnhanced(session)
		if len(incidents) == 0 {
			falseNegatives++
		}
	}

	falseNegativeRate := float64(falseNegatives) / float64(len(frustrationScenarios))
	if falseNegativeRate > 0.15 { // 15% threshold
		t.Errorf("False negative rate too high: %.2f%% (expected <15%%)", falseNegativeRate*100)
	}
}

// TestPerformance benchmarks performance
func TestPerformance(t *testing.T) {
	session := createLargeSession(1000) // 1000 events

	start := time.Now()
	incidents := ufse.ProcessSessionEnhanced(session)
	duration := time.Since(start)

	if duration > 1*time.Second {
		t.Errorf("Processing too slow: %v (expected <1s)", duration)
	}

	if len(incidents) == 0 {
		t.Log("No incidents detected in large session (may be expected)")
	}
}

// Helper functions for test data

func createRageClickSession(clickCount int, timeWindow time.Duration) types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	for i := 0; i < clickCount; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * (timeWindow / time.Duration(clickCount))).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     "/test",
			Target: types.EventTarget{
				ID:   "test-button",
				Type: "button",
			},
			Metadata: map[string]interface{}{},
		})
	}

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(timeWindow),
	}
}

func createRageBaitSession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	for i := 0; i < 3; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * 500 * time.Millisecond).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     "/test",
			Target: types.EventTarget{
				ID:   "fake-button",
				Type: "div",
			},
			Metadata: map[string]interface{}{
				"cursor":          "pointer",
				"hasClickHandler": false,
				"looksLikeButton": true,
				"isButton":        false,
			},
		})
	}

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(2 * time.Second),
	}
}

func createDoubleClickSession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	// Two clicks within 300ms (legitimate double-click)
	events = append(events, types.Event{
		EventType: "click",
		Timestamp: baseTime.Format(time.RFC3339),
		SessionID: "test-session",
		Route:     "/test",
		Target: types.EventTarget{
			ID:   "file-item",
			Type: "div",
		},
		Metadata: map[string]interface{}{},
	})
	events = append(events, types.Event{
		EventType: "click",
		Timestamp: baseTime.Add(250 * time.Millisecond).Format(time.RFC3339),
		SessionID: "test-session",
		Route:     "/test",
		Target: types.EventTarget{
			ID:   "file-item",
			Type: "div",
		},
		Metadata: map[string]interface{}{},
	})

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(1 * time.Second),
	}
}

func createGamingSession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	for i := 0; i < 10; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * 100 * time.Millisecond).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     "/game/play",
			Target: types.EventTarget{
				ID:   "game-button",
				Type: "button",
			},
			Metadata: map[string]interface{}{},
		})
	}

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(1 * time.Second),
	}
}

func createAccessibilitySession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	for i := 0; i < 5; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * 400 * time.Millisecond).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     "/test",
			Target: types.EventTarget{
				ID:   "nav-item",
				Type: "button",
			},
			Metadata: map[string]interface{}{
				"userAgent": "NVDA Screen Reader",
			},
		})
	}

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(2 * time.Second),
	}
}

func createSearchSession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	for i := 0; i < 4; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * 500 * time.Millisecond).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     "/search",
			Target: types.EventTarget{
				ID:   "filter-button",
				Type: "button",
			},
			Metadata: map[string]interface{}{},
		})
	}

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(2 * time.Second),
	}
}

func createComparisonShoppingSession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	routes := []string{"/product/1", "/product/2", "/product/1", "/product/2"}
	for i, route := range routes {
		events = append(events, types.Event{
			EventType: "navigation",
			Timestamp: baseTime.Add(time.Duration(i) * 2 * time.Second).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     route,
			Target: types.EventTarget{
				Type: "a",
			},
			Metadata: map[string]interface{}{},
		})
	}

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(10 * time.Second),
	}
}

func createBlockedProgressSession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	// Multiple clicks with error
	events = append(events, types.Event{
		EventType: "click",
		Timestamp: baseTime.Format(time.RFC3339),
		SessionID: "test-session",
		Route:     "/checkout",
		Target: types.EventTarget{
			ID:   "submit-button",
			Type: "button",
		},
		Metadata: map[string]interface{}{},
	})
	events = append(events, types.Event{
		EventType: "error",
		Timestamp: baseTime.Add(500 * time.Millisecond).Format(time.RFC3339),
		SessionID: "test-session",
		Route:     "/checkout",
		Metadata: map[string]interface{}{
			"error": "Validation failed",
		},
	})
	events = append(events, types.Event{
		EventType: "click",
		Timestamp: baseTime.Add(1 * time.Second).Format(time.RFC3339),
		SessionID: "test-session",
		Route:     "/checkout",
		Target: types.EventTarget{
			ID:   "submit-button",
			Type: "button",
		},
		Metadata: map[string]interface{}{},
	})

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(2 * time.Second),
	}
}

func createAbandonmentSession() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	// Start checkout but abandon
	events = append(events, types.Event{
		EventType: "navigation",
		Timestamp: baseTime.Format(time.RFC3339),
		SessionID: "test-session",
		Route:     "/checkout",
		Metadata: map[string]interface{}{},
	})
	events = append(events, types.Event{
		EventType: "navigation",
		Timestamp: baseTime.Add(30 * time.Second).Format(time.RFC3339),
		SessionID: "test-session",
		Route:     "/",
		Metadata: map[string]interface{}{},
	})

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(35 * time.Second),
	}
}

func createLargeSession(eventCount int) types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()

	for i := 0; i < eventCount; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * 100 * time.Millisecond).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     "/test",
			Target: types.EventTarget{
				ID:   "element",
				Type: "div",
			},
			Metadata: map[string]interface{}{},
		})
	}

	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
		Events:    events,
		StartTime: baseTime,
		EndTime:   baseTime.Add(time.Duration(eventCount) * 100 * time.Millisecond),
	}
}
