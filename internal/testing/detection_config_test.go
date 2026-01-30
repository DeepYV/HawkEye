/**
 * Detection Configuration Tests
 *
 * Author: AI Issue Solver
 * Responsibility: Test configurable frustration detection settings
 */

package testing

import (
	"os"
	"testing"

	"github.com/your-org/frustration-engine/internal/ufse"
)

func TestDefaultDetectionConfig(t *testing.T) {
	// Clear any existing environment variables
	os.Unsetenv("HAWKEYE_TIME_WINDOW_SECONDS")
	os.Unsetenv("HAWKEYE_MIN_SIGNALS")
	os.Unsetenv("HAWKEYE_SCORE_THRESHOLD")
	os.Unsetenv("HAWKEYE_SENSITIVITY")

	// Reload config
	ufse.ReloadConfig()

	config := ufse.GetConfig()

	if config.SessionTimeWindowSeconds != 30 {
		t.Errorf("Expected SessionTimeWindowSeconds 30, got %d", config.SessionTimeWindowSeconds)
	}
	if config.SessionMinSignalsForFrustration != 2 {
		t.Errorf("Expected SessionMinSignalsForFrustration 2, got %d", config.SessionMinSignalsForFrustration)
	}
	if config.SessionScoreThreshold != 0.5 {
		t.Errorf("Expected SessionScoreThreshold 0.5, got %f", config.SessionScoreThreshold)
	}
	if config.SessionEnableDecay != true {
		t.Error("Expected SessionEnableDecay to be true by default")
	}
}

func TestConfigFromEnvironment(t *testing.T) {
	// Set custom environment variables
	os.Setenv("HAWKEYE_TIME_WINDOW_SECONDS", "60")
	os.Setenv("HAWKEYE_MIN_SIGNALS", "3")
	os.Setenv("HAWKEYE_SCORE_THRESHOLD", "0.7")
	os.Setenv("HAWKEYE_ENABLE_DECAY", "false")

	defer func() {
		os.Unsetenv("HAWKEYE_TIME_WINDOW_SECONDS")
		os.Unsetenv("HAWKEYE_MIN_SIGNALS")
		os.Unsetenv("HAWKEYE_SCORE_THRESHOLD")
		os.Unsetenv("HAWKEYE_ENABLE_DECAY")
	}()

	// Reload config
	ufse.ReloadConfig()

	config := ufse.GetConfig()

	if config.SessionTimeWindowSeconds != 60 {
		t.Errorf("Expected SessionTimeWindowSeconds 60, got %d", config.SessionTimeWindowSeconds)
	}
	if config.SessionMinSignalsForFrustration != 3 {
		t.Errorf("Expected SessionMinSignalsForFrustration 3, got %d", config.SessionMinSignalsForFrustration)
	}
	if config.SessionScoreThreshold != 0.7 {
		t.Errorf("Expected SessionScoreThreshold 0.7, got %f", config.SessionScoreThreshold)
	}
	if config.SessionEnableDecay != false {
		t.Error("Expected SessionEnableDecay to be false")
	}
}

func TestSensitivityPresetLow(t *testing.T) {
	os.Setenv("HAWKEYE_SENSITIVITY", "low")
	defer os.Unsetenv("HAWKEYE_SENSITIVITY")

	ufse.ReloadConfig()
	config := ufse.GetConfig()

	// Low sensitivity should have higher threshold
	if config.SessionScoreThreshold != 0.7 {
		t.Errorf("Expected threshold 0.7 for low sensitivity, got %f", config.SessionScoreThreshold)
	}
	if config.SessionMinSignalsForFrustration != 3 {
		t.Errorf("Expected min signals 3 for low sensitivity, got %d", config.SessionMinSignalsForFrustration)
	}
}

func TestSensitivityPresetHigh(t *testing.T) {
	os.Setenv("HAWKEYE_SENSITIVITY", "high")
	defer os.Unsetenv("HAWKEYE_SENSITIVITY")

	ufse.ReloadConfig()
	config := ufse.GetConfig()

	// High sensitivity should have lower threshold
	if config.SessionScoreThreshold != 0.3 {
		t.Errorf("Expected threshold 0.3 for high sensitivity, got %f", config.SessionScoreThreshold)
	}
	if config.SessionMinSignalsForFrustration != 1 {
		t.Errorf("Expected min signals 1 for high sensitivity, got %d", config.SessionMinSignalsForFrustration)
	}
}

func TestSignalWeightsFromEnvironment(t *testing.T) {
	os.Setenv("HAWKEYE_WEIGHT_RAGE", "0.8")
	os.Setenv("HAWKEYE_WEIGHT_CONFUSION", "0.05")

	defer func() {
		os.Unsetenv("HAWKEYE_WEIGHT_RAGE")
		os.Unsetenv("HAWKEYE_WEIGHT_CONFUSION")
	}()

	ufse.ReloadConfig()
	config := ufse.GetConfig()

	if config.WeightRage != 0.8 {
		t.Errorf("Expected WeightRage 0.8, got %f", config.WeightRage)
	}
	if config.WeightConfusion != 0.05 {
		t.Errorf("Expected WeightConfusion 0.05, got %f", config.WeightConfusion)
	}
}

func TestInvalidEnvironmentValues(t *testing.T) {
	// Set invalid values - should be ignored
	os.Setenv("HAWKEYE_TIME_WINDOW_SECONDS", "invalid")
	os.Setenv("HAWKEYE_SCORE_THRESHOLD", "2.0") // Out of range
	os.Setenv("HAWKEYE_MIN_SIGNALS", "-1")       // Negative

	defer func() {
		os.Unsetenv("HAWKEYE_TIME_WINDOW_SECONDS")
		os.Unsetenv("HAWKEYE_SCORE_THRESHOLD")
		os.Unsetenv("HAWKEYE_MIN_SIGNALS")
	}()

	ufse.ReloadConfig()
	config := ufse.GetConfig()

	// Should fall back to defaults
	if config.SessionTimeWindowSeconds != 30 {
		t.Errorf("Expected SessionTimeWindowSeconds 30 (default), got %d", config.SessionTimeWindowSeconds)
	}
	if config.SessionScoreThreshold != 0.5 {
		t.Errorf("Expected SessionScoreThreshold 0.5 (default), got %f", config.SessionScoreThreshold)
	}
	if config.SessionMinSignalsForFrustration != 2 {
		t.Errorf("Expected SessionMinSignalsForFrustration 2 (default), got %d", config.SessionMinSignalsForFrustration)
	}
}

func TestFeatureFlags(t *testing.T) {
	// Test enabled by default
	os.Unsetenv("HAWKEYE_ENHANCED_DETECTION")
	os.Unsetenv("HAWKEYE_SESSION_AGGREGATION")
	os.Unsetenv("UFSE_ENHANCED_DETECTION")
	ufse.ReloadConfig()

	if !ufse.IsEnhancedDetectionEnabled() {
		t.Error("Expected enhanced detection to be enabled by default")
	}
	if !ufse.IsSessionAggregationEnabled() {
		t.Error("Expected session aggregation to be enabled by default")
	}

	// Test disabled via environment
	os.Setenv("HAWKEYE_ENHANCED_DETECTION", "false")
	os.Setenv("HAWKEYE_SESSION_AGGREGATION", "0")
	defer func() {
		os.Unsetenv("HAWKEYE_ENHANCED_DETECTION")
		os.Unsetenv("HAWKEYE_SESSION_AGGREGATION")
	}()

	ufse.ReloadConfig()

	if ufse.IsEnhancedDetectionEnabled() {
		t.Error("Expected enhanced detection to be disabled")
	}
	if ufse.IsSessionAggregationEnabled() {
		t.Error("Expected session aggregation to be disabled")
	}
}

func TestEnvironmentDetection(t *testing.T) {
	os.Setenv("HAWKEYE_ENVIRONMENT", "development")
	defer os.Unsetenv("HAWKEYE_ENVIRONMENT")

	ufse.ReloadConfig()

	if !ufse.IsDevelopmentEnvironment() {
		t.Error("Expected IsDevelopmentEnvironment to return true")
	}
}

func TestEnvironmentDetectionFallback(t *testing.T) {
	os.Unsetenv("HAWKEYE_ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "dev")
	defer os.Unsetenv("ENVIRONMENT")

	ufse.ReloadConfig()

	if !ufse.IsDevelopmentEnvironment() {
		t.Error("Expected IsDevelopmentEnvironment to return true with ENVIRONMENT=dev")
	}
}
