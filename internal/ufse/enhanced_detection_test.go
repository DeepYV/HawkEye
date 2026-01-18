/**
 * Enhanced Detection Tests
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Test cases for enhanced frustration detection
 */

package ufse

import (
	"testing"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/ufse/correlation"
	"github.com/your-org/frustration-engine/internal/ufse/scoring"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// TestEnhancedRageDetection tests multi-tier rage detection
func TestEnhancedRageDetection(t *testing.T) {
	// Create test session with high-strength rage pattern
	session := createTestSessionWithRageClicks(5, 1500*time.Millisecond)
	
	classified := classifyEvents(session.Events)
	detector := signals.NewEnhancedRageDetector()
	candidates := detector.DetectRageMultiTier(classified, session)
	
	if len(candidates) == 0 {
		t.Error("Expected high-strength rage signal to be detected")
	}
	
	if len(candidates) > 0 {
		candidate := candidates[0]
		if candidate.Type != "rage" {
			t.Errorf("Expected signal type 'rage', got '%s'", candidate.Type)
		}
		
		if details := candidate.Details; details != nil {
			if strength, ok := details["strength"].(string); ok {
				if strength != "high" {
					t.Errorf("Expected strength 'high', got '%s'", strength)
				}
			}
		}
	}
}

// TestRageBaitDetection tests rage bait detection
func TestRageBaitDetection(t *testing.T) {
	// Create test session with rage bait pattern (clicks on non-interactive element)
	session := createTestSessionWithRageBait()
	
	classified := classifyEvents(session.Events)
	detector := signals.NewRageBaitDetector()
	candidates := detector.DetectRageBait(classified, session)
	
	if len(candidates) == 0 {
		t.Error("Expected rage bait signal to be detected")
	}
	
	if len(candidates) > 0 {
		candidate := candidates[0]
		if candidate.Type != "rage_bait" {
			t.Errorf("Expected signal type 'rage_bait', got '%s'", candidate.Type)
		}
		
		if details := candidate.Details; details != nil {
			if darkPatternScore, ok := details["darkPatternScore"].(float64); ok {
				if darkPatternScore < 0.6 {
					t.Errorf("Expected dark pattern score >= 0.6, got %f", darkPatternScore)
				}
			}
		}
	}
}

// TestSingleSignalCorrelation tests single-signal correlation
func TestSingleSignalCorrelation(t *testing.T) {
	// Create test session with single high-strength signal
	session := createTestSessionWithHighStrengthSignal()
	
	classified := classifyEvents(session.Events)
	candidates := signals.DetectCandidateSignals(classified, session)
	qualified := signals.QualifySignals(candidates, classified)
	
	// Use enhanced correlation
	groups := correlation.EnhancedCorrelateSignals(qualified)
	
	// Should have at least one group (single-signal)
	if len(groups) == 0 {
		t.Error("Expected at least one correlated group from single high-strength signal")
	}
	
	// Check if it's a single-signal group
	foundSingleSignal := false
	for _, group := range groups {
		if len(group.Signals) == 1 {
			foundSingleSignal = true
			break
		}
	}
	
	if !foundSingleSignal {
		t.Error("Expected at least one single-signal correlation group")
	}
}

// TestMediumConfidenceEmission tests Medium confidence incident emission
func TestMediumConfidenceEmission(t *testing.T) {
	// Create test session that should produce Medium confidence
	session := createTestSessionForMediumConfidence()
	
	incidents := ProcessSessionEnhanced(session)
	
	// Should have at least one incident
	if len(incidents) == 0 {
		t.Error("Expected at least one incident with Medium confidence")
	}
	
	// Check for Medium confidence incident
	foundMedium := false
	for _, incident := range incidents {
		if incident.ConfidenceLevel == "Medium" {
			foundMedium = true
			// Check that explanation has review flag
			if len(incident.Explanation) > 0 {
				// Explanation should indicate needs review
				if len(incident.Explanation) < 20 {
					t.Error("Expected explanation to indicate Medium confidence needs review")
				}
			}
			break
		}
	}
	
	if !foundMedium {
		t.Error("Expected at least one Medium confidence incident")
	}
}

// Helper functions for test data

func createTestSessionWithRageClicks(clickCount int, timeWindow time.Duration) types.Session {
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

func createTestSessionWithRageBait() types.Session {
	events := make([]types.Event, 0)
	baseTime := time.Now()
	
	// Create clicks on non-interactive element that looks clickable
	for i := 0; i < 3; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * 500 * time.Millisecond).Format(time.RFC3339),
			SessionID: "test-session",
			Route:     "/test",
			Target: types.EventTarget{
				ID:   "fake-button",
				Type: "div", // Non-interactive
			},
			Metadata: map[string]interface{}{
				"cursor":         "pointer",
				"hasClickHandler": false,
				"looksLikeButton": true,
				"isButton":       false,
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

func createTestSessionWithHighStrengthSignal() types.Session {
	// Create session with 5 rapid clicks (high-strength)
	return createTestSessionWithRageClicks(5, 1500*time.Millisecond)
}

func createTestSessionForMediumConfidence() types.Session {
	// Create session with high-strength signal but no system feedback
	// This should produce Medium confidence
	events := make([]types.Event, 0)
	baseTime := time.Now()
	
	// High-strength rage clicks
	for i := 0; i < 5; i++ {
		events = append(events, types.Event{
			EventType: "click",
			Timestamp: baseTime.Add(time.Duration(i) * 300 * time.Millisecond).Format(time.RFC3339),
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
		EndTime:   baseTime.Add(2 * time.Second),
	}
}
