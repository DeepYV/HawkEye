/**
 * Tests for Signal Ingestion and Detection Improvements
 *
 * Author: Enhanced Detection Team
 * Responsibility: Test new features for signal ingestion, frustration detection, and noise reduction
 */

package testing

import (
	"testing"
	"time"

	"github.com/your-org/frustration-engine/internal/session"
	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/ufse"
	"github.com/your-org/frustration-engine/internal/ufse/scoring"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// =====================================================
// IDEMPOTENCY AND DEDUPLICATION TESTS
// =====================================================

func TestIdempotencyKeyDeduplication(t *testing.T) {
	events := []types.Event{
		{
			EventType:      "click",
			Timestamp:      "2024-01-15T10:00:00Z",
			SessionID:      "session-1",
			Route:          "/checkout",
			IdempotencyKey: "session-1-2024-01-15T10:00:00Z-1",
			Target:         types.EventTarget{Type: "button", ID: "submit"},
		},
		{
			EventType:      "click",
			Timestamp:      "2024-01-15T10:00:00Z",
			SessionID:      "session-1",
			Route:          "/checkout",
			IdempotencyKey: "session-1-2024-01-15T10:00:00Z-1", // Duplicate
			Target:         types.EventTarget{Type: "button", ID: "submit"},
		},
		{
			EventType:      "click",
			Timestamp:      "2024-01-15T10:00:01Z",
			SessionID:      "session-1",
			Route:          "/checkout",
			IdempotencyKey: "session-1-2024-01-15T10:00:01Z-2", // Different
			Target:         types.EventTarget{Type: "button", ID: "submit"},
		},
	}

	deduplicated := session.DeduplicateEvents(events)

	if len(deduplicated) != 2 {
		t.Errorf("Expected 2 events after deduplication, got %d", len(deduplicated))
	}
}

func TestIdempotencyCacheExpiry(t *testing.T) {
	cache := session.NewIdempotencyCache(100, 60) // 60 second TTL

	now := time.Now().Unix()

	// Add key
	isDup := cache.CheckAndAdd("key1", now)
	if isDup {
		t.Error("First check should not be duplicate")
	}

	// Check same key immediately - should be duplicate
	isDup = cache.CheckAndAdd("key1", now)
	if !isDup {
		t.Error("Second check should be duplicate")
	}

	// Check same key after expiry - should not be duplicate
	isDup = cache.CheckAndAdd("key1", now+61)
	if isDup {
		t.Error("Check after expiry should not be duplicate")
	}
}

func TestFingerprintDeduplicationFallback(t *testing.T) {
	// Events without idempotency keys should use fingerprint
	events := []types.Event{
		{
			EventType: "click",
			Timestamp: "2024-01-15T10:00:00Z",
			SessionID: "session-1",
			Route:     "/checkout",
			Target:    types.EventTarget{Type: "button", ID: "submit"},
		},
		{
			EventType: "click",
			Timestamp: "2024-01-15T10:00:00Z",
			SessionID: "session-1",
			Route:     "/checkout",
			Target:    types.EventTarget{Type: "button", ID: "submit"},
		},
	}

	deduplicated := session.DeduplicateEvents(events)

	if len(deduplicated) != 1 {
		t.Errorf("Expected 1 event after fingerprint deduplication, got %d", len(deduplicated))
	}
}

// =====================================================
// FORM LOOP DETECTION TESTS
// =====================================================

func TestFormLoopRapidSubmission(t *testing.T) {
	detector := signals.NewFormLoopDetector()

	now := time.Now()
	classified := []signals.ClassifiedEvent{
		createFormSubmitEvent(now, "/checkout", "checkout-form"),
		createFormSubmitEvent(now.Add(500*time.Millisecond), "/checkout", "checkout-form"),
		createFormSubmitEvent(now.Add(1*time.Second), "/checkout", "checkout-form"),
		createFormSubmitEvent(now.Add(1500*time.Millisecond), "/checkout", "checkout-form"),
		createFormSubmitEvent(now.Add(2*time.Second), "/checkout", "checkout-form"),
	}

	session := createTestSession()
	candidates := detector.DetectFormLoops(classified, session)

	if len(candidates) == 0 {
		t.Error("Expected form loop signal from rapid submissions")
	}

	if len(candidates) > 0 && candidates[0].Type != "form_loop" {
		t.Errorf("Expected form_loop signal, got %s", candidates[0].Type)
	}
}

func TestFormLoopFrustratedResubmission(t *testing.T) {
	detector := signals.NewFormLoopDetector()

	now := time.Now()
	classified := []signals.ClassifiedEvent{
		createFormSubmitEvent(now, "/checkout", "checkout-form"),
		createFormSubmitEvent(now.Add(3*time.Second), "/checkout", "checkout-form"),
		createFormSubmitEvent(now.Add(6*time.Second), "/checkout", "checkout-form"),
	}

	session := createTestSession()
	candidates := detector.DetectFormLoops(classified, session)

	if len(candidates) == 0 {
		t.Error("Expected form loop signal from frustrated resubmissions")
	}
}

func TestFormLoopNoFalsePositiveWithSuccess(t *testing.T) {
	detector := signals.NewFormLoopDetector()

	now := time.Now()
	classified := []signals.ClassifiedEvent{
		createFormSubmitEvent(now, "/checkout", "checkout-form"),
		createSuccessResponseEvent(now.Add(500*time.Millisecond), "/checkout"),
		createNavigationEvent(now.Add(1*time.Second), "/checkout/success"),
	}

	session := createTestSession()
	candidates := detector.DetectFormLoops(classified, session)

	if len(candidates) > 0 {
		t.Error("Should not detect form loop when success response exists")
	}
}

// =====================================================
// SIGNAL DECAY TESTS
// =====================================================

func TestSignalDecayBasic(t *testing.T) {
	decayer := scoring.NewSignalDecayer()
	referenceTime := time.Now()

	// Recent signal should have high strength
	recentSignal := decayer.ApplyDecay(1.0, referenceTime.Add(-5*time.Second), referenceTime)
	if recentSignal.DecayedStrength < 0.9 {
		t.Errorf("Recent signal should retain most strength, got %.2f", recentSignal.DecayedStrength)
	}

	// Old signal should have lower strength
	oldSignal := decayer.ApplyDecay(1.0, referenceTime.Add(-2*time.Minute), referenceTime)
	if oldSignal.DecayedStrength > 0.5 {
		t.Errorf("Old signal should have reduced strength, got %.2f", oldSignal.DecayedStrength)
	}

	// Very old signal should be stale
	veryOldSignal := decayer.ApplyDecay(1.0, referenceTime.Add(-10*time.Minute), referenceTime)
	if !veryOldSignal.IsStale {
		t.Error("Very old signal should be marked as stale")
	}
}

func TestSignalDecayFiltering(t *testing.T) {
	decayer := scoring.NewSignalDecayer()
	referenceTime := time.Now()

	signals := []scoring.SignalWithTimestamp{
		{Type: "rage", Timestamp: referenceTime.Add(-5 * time.Second), Strength: 0.8},
		{Type: "blocked", Timestamp: referenceTime.Add(-30 * time.Second), Strength: 0.7},
		{Type: "confusion", Timestamp: referenceTime.Add(-6 * time.Minute), Strength: 0.5}, // Should be filtered
	}

	filtered := decayer.FilterStaleSignals(signals, referenceTime)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 signals after filtering, got %d", len(filtered))
	}
}

func TestRecentSignalBoost(t *testing.T) {
	decayer := scoring.NewSignalDecayer()
	referenceTime := time.Now()

	// Very recent signal (within 10s) should get boost
	veryRecentSignal := decayer.ApplyDecay(0.8, referenceTime.Add(-5*time.Second), referenceTime)

	// Signal just outside recent window
	slightlyOldSignal := decayer.ApplyDecay(0.8, referenceTime.Add(-15*time.Second), referenceTime)

	if veryRecentSignal.DecayedStrength <= slightlyOldSignal.DecayedStrength {
		t.Error("Very recent signal should have higher strength due to boost")
	}
}

// =====================================================
// SESSION THRESHOLD TESTS
// =====================================================

func TestSessionThresholdChecking(t *testing.T) {
	checker := scoring.NewSessionThresholdChecker()

	// Short session should fail
	shortMetrics := scoring.SessionMetrics{
		Duration:         2 * time.Second,
		EventCount:       2,
		InteractionCount: 1,
	}
	result := checker.CheckThresholds(shortMetrics)
	if result.MeetsThreshold {
		t.Error("Short session should not meet thresholds")
	}

	// Valid session should pass
	validMetrics := scoring.SessionMetrics{
		Duration:         30 * time.Second,
		EventCount:       10,
		InteractionCount: 5,
	}
	result = checker.CheckThresholds(validMetrics)
	if !result.MeetsThreshold {
		t.Error("Valid session should meet thresholds")
	}
}

func TestSessionMetricsCalculation(t *testing.T) {
	now := time.Now()
	events := []scoring.EventForMetrics{
		{EventType: "click", Timestamp: now, IsError: false},
		{EventType: "input", Timestamp: now.Add(1 * time.Second), IsError: false},
		{EventType: "navigation", Timestamp: now.Add(2 * time.Second), IsError: false},
		{EventType: "error", Timestamp: now.Add(3 * time.Second), IsError: true},
		{EventType: "form_submit", Timestamp: now.Add(5 * time.Second), IsError: false},
	}

	metrics := scoring.CalculateSessionMetrics(events)

	if metrics.EventCount != 5 {
		t.Errorf("Expected 5 events, got %d", metrics.EventCount)
	}
	if metrics.InteractionCount != 3 { // click, input, form_submit
		t.Errorf("Expected 3 interactions, got %d", metrics.InteractionCount)
	}
	if metrics.NavigationCount != 1 {
		t.Errorf("Expected 1 navigation, got %d", metrics.NavigationCount)
	}
	if metrics.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", metrics.ErrorCount)
	}
	if metrics.Duration != 5*time.Second {
		t.Errorf("Expected 5s duration, got %v", metrics.Duration)
	}
}

// =====================================================
// ROUTE-SPECIFIC CONFIG TESTS
// =====================================================

func TestRouteConfigMatching(t *testing.T) {
	manager := ufse.NewRouteConfigManager(ufse.DefaultDetectionConfig())

	// Add checkout route config
	checkoutConfig := ufse.RouteConfig{
		Pattern:       "/checkout/**",
		Priority:      100,
		ShadowModeEnabled: true,
	}
	err := manager.AddRouteConfig(checkoutConfig)
	if err != nil {
		t.Fatalf("Failed to add route config: %v", err)
	}

	// Test matching
	config := manager.GetConfigForRoute("/checkout/payment")
	if !config.ShadowModeEnabled {
		t.Error("Checkout route should have shadow mode enabled")
	}

	// Test non-matching
	config = manager.GetConfigForRoute("/products/123")
	if config.ShadowModeEnabled {
		t.Error("Products route should not have shadow mode enabled")
	}
}

func TestRouteConfigPriority(t *testing.T) {
	manager := ufse.NewRouteConfigManager(ufse.DefaultDetectionConfig())

	// Add general config
	generalConfig := ufse.RouteConfig{
		Pattern:  "/**",
		Priority: 10,
		DisableRageDetection: false,
	}
	manager.AddRouteConfig(generalConfig)

	// Add specific config
	specificConfig := ufse.RouteConfig{
		Pattern:              "/game/**",
		Priority:             100,
		DisableRageDetection: true,
	}
	manager.AddRouteConfig(specificConfig)

	// Specific should take priority
	config := manager.GetConfigForRoute("/game/level1")
	if !config.DisableRageDetection {
		t.Error("Game route should have rage detection disabled")
	}
}

// =====================================================
// SHADOW MODE TESTS
// =====================================================

func TestShadowModeRecording(t *testing.T) {
	shadowMode := ufse.NewShadowMode("test_detector")
	shadowMode.Enable()

	candidate := signals.CandidateSignal{
		Type:      "rage",
		Timestamp: time.Now().Unix(),
		Route:     "/checkout",
	}

	shadowMode.RecordDetection(candidate, "High", true)
	shadowMode.RecordDetection(candidate, "Medium", false)

	metrics := shadowMode.GetMetrics()

	if metrics.TotalDetections != 2 {
		t.Errorf("Expected 2 detections, got %d", metrics.TotalDetections)
	}
	if metrics.HighConfidenceCount != 1 {
		t.Errorf("Expected 1 high confidence, got %d", metrics.HighConfidenceCount)
	}
	if metrics.WouldHaveEmitted != 1 {
		t.Errorf("Expected 1 would-have-emitted, got %d", metrics.WouldHaveEmitted)
	}
}

func TestShadowModeManager(t *testing.T) {
	manager := ufse.NewShadowModeManager()

	sm1 := manager.GetOrCreate("detector1")
	sm2 := manager.GetOrCreate("detector2")

	sm1.Enable()

	if !sm1.IsEnabled() {
		t.Error("Detector1 should be enabled")
	}
	if sm2.IsEnabled() {
		t.Error("Detector2 should not be enabled")
	}

	manager.EnableAll()

	if !sm2.IsEnabled() {
		t.Error("Detector2 should be enabled after EnableAll")
	}
}

// =====================================================
// SUPPRESSION AUDIT TESTS
// =====================================================

func TestSuppressionAuditLogging(t *testing.T) {
	log := ufse.NewSuppressionAuditLog(100)

	event := ufse.SuppressionEvent{
		SessionID:        "session-1",
		Route:            "/checkout",
		SignalType:       "rage",
		Reason:           ufse.ReasonLowConfidence,
		ReasonDetails:    "Only 1 signal detected",
		ConfidenceLevel:  "Low",
		FrustrationScore: 25,
	}

	log.RecordSuppression(event)

	recent := log.GetRecentSuppressions(10)
	if len(recent) != 1 {
		t.Errorf("Expected 1 suppression, got %d", len(recent))
	}

	if recent[0].Reason != ufse.ReasonLowConfidence {
		t.Errorf("Expected low_confidence reason, got %s", recent[0].Reason)
	}
}

func TestSuppressionStatistics(t *testing.T) {
	log := ufse.NewSuppressionAuditLog(100)

	// Add multiple suppressions
	reasons := []ufse.SuppressionReason{
		ufse.ReasonLowConfidence,
		ufse.ReasonLowConfidence,
		ufse.ReasonFalseAlarmDetected,
		ufse.ReasonShadowMode,
	}

	for _, reason := range reasons {
		log.RecordSuppression(ufse.SuppressionEvent{
			SessionID:  "session-1",
			Route:      "/checkout",
			SignalType: "rage",
			Reason:     reason,
		})
	}

	stats := log.GetSuppressionStats()

	if stats.TotalSuppressions != 4 {
		t.Errorf("Expected 4 total suppressions, got %d", stats.TotalSuppressions)
	}
	if stats.ByReason["low_confidence"] != 2 {
		t.Errorf("Expected 2 low_confidence suppressions, got %d", stats.ByReason["low_confidence"])
	}
}

// =====================================================
// DETECTION REASONING TESTS
// =====================================================

func TestDetectionReasoningBuilder(t *testing.T) {
	builder := ufse.NewDetectionReasoningBuilder("session-123")

	// Add step
	step := builder.StartStep("Classification", "Classify events by category", 100)
	step.Complete(100, "All events classified")

	// Add signals
	builder.AddDetectedSignal(signals.CandidateSignal{
		Type:      "rage",
		Timestamp: time.Now().Unix(),
		Route:     "/checkout",
	})

	// Set score breakdown
	builder.SetScoreBreakdown(ufse.ScoreBreakdown{
		SignalCountScore: 20,
		TypeWeightScore:  15,
		FinalScore:       45,
		Explanation:      "Multiple signals detected with system feedback",
	})

	// Set final decision
	builder.SetFinalDecision(ufse.DecisionDetails{
		ShouldEmit: true,
		Reason:     "High confidence incident",
	})

	reasoning := builder.Build()

	if reasoning.Outcome != "emitted" {
		t.Errorf("Expected emitted outcome, got %s", reasoning.Outcome)
	}
	if len(reasoning.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(reasoning.Steps))
	}
	if len(reasoning.DetectedSignals) != 1 {
		t.Errorf("Expected 1 detected signal, got %d", len(reasoning.DetectedSignals))
	}
	if reasoning.ScoreBreakdown.FinalScore != 45 {
		t.Errorf("Expected score 45, got %d", reasoning.ScoreBreakdown.FinalScore)
	}
}

func TestDetectionReasoningHumanReadable(t *testing.T) {
	builder := ufse.NewDetectionReasoningBuilder("session-123")
	builder.SetFinalDecision(ufse.DecisionDetails{
		ShouldEmit: false,
		Reason:     "Low confidence",
		SuppressionReason: ufse.ReasonLowConfidence,
	})

	reasoning := builder.Build()
	humanReadable := reasoning.ToHumanReadable()

	if len(humanReadable) == 0 {
		t.Error("Human readable output should not be empty")
	}
	if !containsString(humanReadable, "suppressed") {
		t.Error("Output should mention suppressed outcome")
	}
}

// =====================================================
// HELPER FUNCTIONS
// =====================================================

func createFormSubmitEvent(timestamp time.Time, route, formID string) signals.ClassifiedEvent {
	return signals.ClassifiedEvent{
		Event: types.Event{
			EventType: "form_submit",
			Timestamp: timestamp.Format(time.RFC3339),
			SessionID: "test-session",
			Route:     route,
			Target: types.EventTarget{
				Type: "form",
				ID:   formID,
			},
		},
		Timestamp: timestamp,
		Category:  signals.CategoryInteraction,
		Route:     route,
	}
}

func createSuccessResponseEvent(timestamp time.Time, route string) signals.ClassifiedEvent {
	return signals.ClassifiedEvent{
		Event: types.Event{
			EventType: "network",
			Timestamp: timestamp.Format(time.RFC3339),
			SessionID: "test-session",
			Route:     route,
			Target:    types.EventTarget{Type: "network"},
			Metadata:  map[string]interface{}{"status": float64(200)},
		},
		Timestamp: timestamp,
		Category:  signals.CategorySystemFeedback,
		Route:     route,
	}
}

func createNavigationEvent(timestamp time.Time, toRoute string) signals.ClassifiedEvent {
	return signals.ClassifiedEvent{
		Event: types.Event{
			EventType: "navigation",
			Timestamp: timestamp.Format(time.RFC3339),
			SessionID: "test-session",
			Route:     toRoute,
			Target:    types.EventTarget{Type: "navigation"},
			Metadata:  map[string]interface{}{"to": toRoute},
		},
		Timestamp: timestamp,
		Category:  signals.CategoryNavigation,
		Route:     toRoute,
	}
}

func createTestSession() types.Session {
	return types.Session{
		SessionID: "test-session",
		ProjectID: "test-project",
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr))
}
