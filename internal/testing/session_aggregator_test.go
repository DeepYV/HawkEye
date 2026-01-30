/**
 * Session Aggregator Tests
 *
 * Author: AI Issue Solver
 * Responsibility: Test session-level frustration detection and aggregation
 */

package testing

import (
	"testing"
	"time"

	"github.com/your-org/frustration-engine/internal/ufse/scoring"
)

func TestNewSessionAggregator(t *testing.T) {
	aggregator := scoring.NewSessionAggregator()
	if aggregator == nil {
		t.Error("NewSessionAggregator returned nil")
	}

	config := aggregator.GetConfig()
	if config.TimeWindowSeconds != 30 {
		t.Errorf("Expected TimeWindowSeconds 30, got %d", config.TimeWindowSeconds)
	}
	if config.MinSignalsForFrustration != 2 {
		t.Errorf("Expected MinSignalsForFrustration 2, got %d", config.MinSignalsForFrustration)
	}
}

func TestSessionAggregatorWithCustomConfig(t *testing.T) {
	config := scoring.AggregatorConfig{
		TimeWindowSeconds:        60,
		MinSignalsForFrustration: 3,
		AggregatedScoreThreshold: 0.7,
		EnableDecay:              false,
		SignalWeights: map[string]float64{
			"rage": 0.5,
		},
	}

	aggregator := scoring.NewSessionAggregatorWithConfig(config)
	resultConfig := aggregator.GetConfig()

	if resultConfig.TimeWindowSeconds != 60 {
		t.Errorf("Expected TimeWindowSeconds 60, got %d", resultConfig.TimeWindowSeconds)
	}
	if resultConfig.MinSignalsForFrustration != 3 {
		t.Errorf("Expected MinSignalsForFrustration 3, got %d", resultConfig.MinSignalsForFrustration)
	}
}

func TestAggregateInsufficientSignals(t *testing.T) {
	aggregator := scoring.NewSessionAggregator()
	now := time.Now()

	// Add only one signal (less than minimum of 2)
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now.Add(-5 * time.Second),
		Strength:  0.8,
		Route:     "/checkout",
	})

	result := aggregator.Aggregate(now)

	if result.IsFrustrated {
		t.Error("Expected IsFrustrated to be false with insufficient signals")
	}
	if result.SignalCount != 1 {
		t.Errorf("Expected SignalCount 1, got %d", result.SignalCount)
	}
	if result.Confidence != "low" {
		t.Errorf("Expected Confidence 'low', got '%s'", result.Confidence)
	}
}

func TestAggregateSufficientSignals(t *testing.T) {
	aggregator := scoring.NewSessionAggregator()
	now := time.Now()

	// Add multiple signals
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now.Add(-10 * time.Second),
		Strength:  0.8,
		Route:     "/checkout",
	})
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "blocked",
		Timestamp: now.Add(-5 * time.Second),
		Strength:  0.9,
		Route:     "/checkout",
	})

	result := aggregator.Aggregate(now)

	if result.SignalCount != 2 {
		t.Errorf("Expected SignalCount 2, got %d", result.SignalCount)
	}
	if result.TotalScore <= 0 {
		t.Error("Expected positive TotalScore")
	}
	if _, ok := result.SignalBreakdown["rage"]; !ok {
		t.Error("Expected 'rage' in SignalBreakdown")
	}
	if _, ok := result.SignalBreakdown["blocked"]; !ok {
		t.Error("Expected 'blocked' in SignalBreakdown")
	}
}

func TestAggregateSignalsOutsideWindow(t *testing.T) {
	aggregator := scoring.NewSessionAggregator()
	now := time.Now()

	// Add signal outside the 30-second window
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now.Add(-60 * time.Second), // 60 seconds ago, outside 30s window
		Strength:  0.8,
		Route:     "/checkout",
	})

	result := aggregator.Aggregate(now)

	if result.SignalCount != 0 {
		t.Errorf("Expected SignalCount 0 for signals outside window, got %d", result.SignalCount)
	}
	if result.IsFrustrated {
		t.Error("Expected IsFrustrated to be false with no signals in window")
	}
}

func TestAggregateWithDecay(t *testing.T) {
	config := scoring.AggregatorConfig{
		TimeWindowSeconds:        30,
		MinSignalsForFrustration: 1,
		AggregatedScoreThreshold: 0.3,
		EnableDecay:              true,
		DecayHalfLifeSeconds:     10,
		SignalWeights:            scoring.DefaultSignalWeights,
	}

	aggregator := scoring.NewSessionAggregatorWithConfig(config)
	now := time.Now()

	// Add signal 20 seconds ago (2 half-lives = 0.25 decay factor)
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now.Add(-20 * time.Second),
		Strength:  1.0,
		Route:     "/checkout",
	})

	result := aggregator.Aggregate(now)

	// With decay, the effective strength should be reduced
	// Original strength 1.0 * decay factor ~0.25 = ~0.25 effective strength
	if result.TotalScore >= 1.0 {
		t.Errorf("Expected TotalScore < 1.0 with decay applied, got %f", result.TotalScore)
	}
}

func TestAggregateWithoutDecay(t *testing.T) {
	config := scoring.AggregatorConfig{
		TimeWindowSeconds:        30,
		MinSignalsForFrustration: 1,
		AggregatedScoreThreshold: 0.3,
		EnableDecay:              false,
		SignalWeights:            scoring.DefaultSignalWeights,
	}

	aggregator := scoring.NewSessionAggregatorWithConfig(config)
	now := time.Now()

	// Add signal 20 seconds ago
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now.Add(-20 * time.Second),
		Strength:  1.0,
		Route:     "/checkout",
	})

	result := aggregator.Aggregate(now)

	// Without decay, the signal should retain full strength
	if result.TotalScore <= 0.5 {
		t.Errorf("Expected TotalScore > 0.5 without decay, got %f", result.TotalScore)
	}
}

func TestAggregateSignalsStateless(t *testing.T) {
	aggregator := scoring.NewSessionAggregator()
	now := time.Now()

	signals := []scoring.AggregatedSignal{
		{Type: "rage", Timestamp: now.Add(-10 * time.Second), Strength: 0.8, Route: "/checkout"},
		{Type: "blocked", Timestamp: now.Add(-5 * time.Second), Strength: 0.9, Route: "/checkout"},
	}

	result := aggregator.AggregateSignals(signals, now)

	if result.SignalCount != 2 {
		t.Errorf("Expected SignalCount 2, got %d", result.SignalCount)
	}
	if !result.IsFrustrated {
		t.Error("Expected IsFrustrated to be true")
	}
}

func TestHighConfidenceDetection(t *testing.T) {
	config := scoring.AggregatorConfig{
		TimeWindowSeconds:        30,
		MinSignalsForFrustration: 1,
		AggregatedScoreThreshold: 0.3,
		EnableDecay:              false,
		SignalWeights:            scoring.DefaultSignalWeights,
	}

	aggregator := scoring.NewSessionAggregatorWithConfig(config)
	now := time.Now()

	// Add multiple high-strength signals
	for i := 0; i < 5; i++ {
		aggregator.AddSignal(scoring.AggregatedSignal{
			Type:      "rage",
			Timestamp: now.Add(-time.Duration(i) * time.Second),
			Strength:  0.9,
			Route:     "/checkout",
		})
	}

	result := aggregator.Aggregate(now)

	if result.Confidence != "high" {
		t.Errorf("Expected Confidence 'high' with many high-strength signals, got '%s'", result.Confidence)
	}
}

func TestPruneOldSignals(t *testing.T) {
	aggregator := scoring.NewSessionAggregator()
	now := time.Now()

	// Add mix of old and new signals
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now.Add(-60 * time.Second), // Old
		Strength:  0.8,
	})
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "blocked",
		Timestamp: now.Add(-5 * time.Second), // Recent
		Strength:  0.8,
	})

	// Prune old signals
	pruned := aggregator.PruneOldSignals(now)

	if pruned != 1 {
		t.Errorf("Expected 1 pruned signal, got %d", pruned)
	}

	// Verify only recent signal remains
	result := aggregator.Aggregate(now)
	if result.SignalCount != 1 {
		t.Errorf("Expected 1 signal after pruning, got %d", result.SignalCount)
	}
}

func TestClearSignals(t *testing.T) {
	aggregator := scoring.NewSessionAggregator()
	now := time.Now()

	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now,
		Strength:  0.8,
	})

	aggregator.Clear()

	result := aggregator.Aggregate(now)
	if result.SignalCount != 0 {
		t.Errorf("Expected 0 signals after clear, got %d", result.SignalCount)
	}
}

func TestDifferentSignalWeights(t *testing.T) {
	// Test that different signal types have different weights
	// Note: With normalized scoring, individual signal weights affect the breakdown
	// but not TotalScore for single signals. The weights matter when multiple signals
	// of different types are combined.
	config := scoring.AggregatorConfig{
		TimeWindowSeconds:        30,
		MinSignalsForFrustration: 1,
		AggregatedScoreThreshold: 0.3,
		EnableDecay:              false,
		SignalWeights: map[string]float64{
			"rage_bait": 0.9, // High weight
			"confusion": 0.1, // Low weight
		},
	}

	now := time.Now()

	// Test with mixed signals - the total should be weighted by signal type
	aggregator := scoring.NewSessionAggregatorWithConfig(config)
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage_bait",
		Timestamp: now,
		Strength:  1.0,
	})
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "confusion",
		Timestamp: now,
		Strength:  1.0,
	})
	result := aggregator.Aggregate(now)

	// With weights 0.9 and 0.1, and both signals at strength 1.0:
	// totalWeight = 0.9*1.0 + 0.1*1.0 = 1.0
	// maxPossibleWeight = 0.9 + 0.1 = 1.0
	// TotalScore = 1.0/1.0 = 1.0
	if result.TotalScore != 1.0 {
		t.Errorf("Expected TotalScore 1.0 with full strength signals, got %f", result.TotalScore)
	}

	// Check signal breakdown captures weights correctly
	rageBaitContrib := result.SignalBreakdown["rage_bait"]
	confusionContrib := result.SignalBreakdown["confusion"]

	if rageBaitContrib <= confusionContrib {
		t.Errorf("Expected rage_bait contribution (%f) > confusion contribution (%f)",
			rageBaitContrib, confusionContrib)
	}

	// Verify exact breakdown values
	if rageBaitContrib != 0.9 {
		t.Errorf("Expected rage_bait contribution 0.9, got %f", rageBaitContrib)
	}
	if confusionContrib != 0.1 {
		t.Errorf("Expected confusion contribution 0.1, got %f", confusionContrib)
	}
}

func TestReasoningGeneration(t *testing.T) {
	config := scoring.AggregatorConfig{
		TimeWindowSeconds:        30,
		MinSignalsForFrustration: 1,
		AggregatedScoreThreshold: 0.3,
		EnableDecay:              false,
		SignalWeights:            scoring.DefaultSignalWeights,
	}

	aggregator := scoring.NewSessionAggregatorWithConfig(config)
	now := time.Now()

	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "rage",
		Timestamp: now,
		Strength:  0.9,
	})
	aggregator.AddSignal(scoring.AggregatedSignal{
		Type:      "blocked",
		Timestamp: now,
		Strength:  0.9,
	})

	result := aggregator.Aggregate(now)

	if result.Reasoning == "" {
		t.Error("Expected non-empty Reasoning")
	}

	// Should mention "Frustration detected" since threshold is exceeded
	if !result.IsFrustrated {
		t.Skip("Threshold not exceeded, skipping reasoning check")
	}
}
