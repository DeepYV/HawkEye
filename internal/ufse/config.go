/**
 * UFSE Configuration
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Configuration for frustration detection system
 */

package ufse

import (
	"strconv"
	"time"

	"github.com/your-org/frustration-engine/internal/config"
)

// DetectionConfig holds configuration for detection thresholds
type DetectionConfig struct {
	// Feature flags
	UseEnhancedDetection     bool
	UseSessionAggregation    bool // Enable session-level aggregation

	// Rage detection thresholds
	RageHighMinClicks            int
	RageHighTimeWindow          time.Duration
	RageHighMaxTimeBetweenClicks time.Duration

	RageMediumMinClicks            int
	RageMediumTimeWindow           time.Duration
	RageMediumMaxTimeBetweenClicks time.Duration

	RageLowMinClicks            int
	RageLowTimeWindow           time.Duration
	RageLowMaxTimeBetweenClicks time.Duration

	// Rage bait detection
	RageBaitEnabled            bool
	RageBaitMinClicks          int
	RageBaitTimeWindow         time.Duration
	RageBaitMaxTimeBetweenClicks time.Duration
	MinDarkPatternScore        float64

	// Correlation
	SingleSignalStrengthThreshold float64
	CorrelationTimeWindow        time.Duration

	// Confidence
	EmitMediumConfidence bool
	MediumScoreRange     [2]int // [min, max]
	HighScoreRange      [2]int // [min, max]

	// Session-level aggregation settings
	SessionTimeWindowSeconds        int     // Time window for aggregating signals (default: 30)
	SessionMinSignalsForFrustration int     // Minimum signals needed (default: 2)
	SessionScoreThreshold           float64 // Score threshold (default: 0.5)
	SessionEnableDecay              bool    // Enable time-based decay (default: true)
	SessionDecayHalfLifeSeconds     int     // Decay half-life (default: 15)

	// Signal weights (0.0 to 1.0)
	WeightRage        float64
	WeightRageBait    float64
	WeightBlocked     float64
	WeightAbandonment float64
	WeightConfusion   float64
	WeightFormLoop    float64

	// Detection sensitivity
	SensitivityLevel string // "low", "medium", "high" (default: "medium")

	// Environment
	Environment string // "development", "staging", "production"
}

// DefaultDetectionConfig returns default configuration
func DefaultDetectionConfig() DetectionConfig {
	return DetectionConfig{
		UseEnhancedDetection:  true,  // Feature flag: enabled by default
		UseSessionAggregation: true,  // Session aggregation: enabled by default

		// High strength rage
		RageHighMinClicks:            5,
		RageHighTimeWindow:           2 * time.Second,
		RageHighMaxTimeBetweenClicks: 300 * time.Millisecond,

		// Medium strength rage
		RageMediumMinClicks:            4,
		RageMediumTimeWindow:           3 * time.Second,
		RageMediumMaxTimeBetweenClicks: 500 * time.Millisecond,

		// Low strength rage
		RageLowMinClicks:            3,
		RageLowTimeWindow:           5 * time.Second,
		RageLowMaxTimeBetweenClicks: 800 * time.Millisecond,

		// Rage bait
		RageBaitEnabled:            true,
		RageBaitMinClicks:          3,
		RageBaitTimeWindow:         5 * time.Second,
		RageBaitMaxTimeBetweenClicks: 1000 * time.Millisecond,
		MinDarkPatternScore:       0.6,

		// Correlation
		SingleSignalStrengthThreshold: 0.8,
		CorrelationTimeWindow:          30 * time.Second,

		// Confidence
		EmitMediumConfidence: true,
		MediumScoreRange:     [2]int{0, 50},
		HighScoreRange:       [2]int{51, 100},

		// Session-level aggregation
		SessionTimeWindowSeconds:        30,
		SessionMinSignalsForFrustration: 2,
		SessionScoreThreshold:           0.5,
		SessionEnableDecay:              true,
		SessionDecayHalfLifeSeconds:     15,

		// Signal weights
		WeightRage:        0.35,
		WeightRageBait:    0.50,
		WeightBlocked:     0.40,
		WeightAbandonment: 0.30,
		WeightConfusion:   0.15,
		WeightFormLoop:    0.35,

		// Default sensitivity
		SensitivityLevel: "medium",

		// Default environment
		Environment: "production",
	}
}

// LoadDetectionConfig loads configuration from environment variables
func LoadDetectionConfig() DetectionConfig {
	cfg := DefaultDetectionConfig()

	// Feature flags
	if val := config.GetEnv("UFSE_ENHANCED_DETECTION", ""); val != "" {
		cfg.UseEnhancedDetection = val == "true" || val == "1"
	}
	if val := config.GetEnv("HAWKEYE_ENHANCED_DETECTION", ""); val != "" {
		cfg.UseEnhancedDetection = val == "true" || val == "1"
	}
	if val := config.GetEnv("HAWKEYE_SESSION_AGGREGATION", ""); val != "" {
		cfg.UseSessionAggregation = val == "true" || val == "1"
	}
	
	// Rage detection thresholds
	if val := config.GetEnv("RAGE_HIGH_MIN_CLICKS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageHighMinClicks = i
		}
	}
	
	if val := config.GetEnv("RAGE_HIGH_TIME_WINDOW_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageHighTimeWindow = time.Duration(i) * time.Millisecond
		}
	}
	
	if val := config.GetEnv("RAGE_HIGH_MAX_TIME_BETWEEN_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageHighMaxTimeBetweenClicks = time.Duration(i) * time.Millisecond
		}
	}
	
	// Medium strength
	if val := config.GetEnv("RAGE_MEDIUM_MIN_CLICKS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageMediumMinClicks = i
		}
	}
	
	if val := config.GetEnv("RAGE_MEDIUM_TIME_WINDOW_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageMediumTimeWindow = time.Duration(i) * time.Millisecond
		}
	}
	
	if val := config.GetEnv("RAGE_MEDIUM_MAX_TIME_BETWEEN_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageMediumMaxTimeBetweenClicks = time.Duration(i) * time.Millisecond
		}
	}
	
	// Low strength
	if val := config.GetEnv("RAGE_LOW_MIN_CLICKS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageLowMinClicks = i
		}
	}
	
	if val := config.GetEnv("RAGE_LOW_TIME_WINDOW_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageLowTimeWindow = time.Duration(i) * time.Millisecond
		}
	}
	
	if val := config.GetEnv("RAGE_LOW_MAX_TIME_BETWEEN_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageLowMaxTimeBetweenClicks = time.Duration(i) * time.Millisecond
		}
	}
	
	// Rage bait
	if val := config.GetEnv("RAGE_BAIT_ENABLED", ""); val != "" {
		cfg.RageBaitEnabled = val == "true" || val == "1"
	}
	
	if val := config.GetEnv("RAGE_BAIT_MIN_CLICKS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageBaitMinClicks = i
		}
	}
	
	if val := config.GetEnv("RAGE_BAIT_TIME_WINDOW_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageBaitTimeWindow = time.Duration(i) * time.Millisecond
		}
	}
	
	if val := config.GetEnv("RAGE_BAIT_MAX_TIME_BETWEEN_MS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.RageBaitMaxTimeBetweenClicks = time.Duration(i) * time.Millisecond
		}
	}
	
	if val := config.GetEnv("MIN_DARK_PATTERN_SCORE", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.MinDarkPatternScore = f
		}
	}
	
	// Correlation
	if val := config.GetEnv("SINGLE_SIGNAL_STRENGTH_THRESHOLD", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.SingleSignalStrengthThreshold = f
		}
	}
	
	if val := config.GetEnv("CORRELATION_TIME_WINDOW_SECONDS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.CorrelationTimeWindow = time.Duration(i) * time.Second
		}
	}
	
	// Confidence
	if val := config.GetEnv("EMIT_MEDIUM_CONFIDENCE", ""); val != "" {
		cfg.EmitMediumConfidence = val == "true" || val == "1"
	}

	// Session-level aggregation
	if val := config.GetEnv("HAWKEYE_TIME_WINDOW_SECONDS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			cfg.SessionTimeWindowSeconds = i
		}
	}

	if val := config.GetEnv("HAWKEYE_MIN_SIGNALS", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			cfg.SessionMinSignalsForFrustration = i
		}
	}

	if val := config.GetEnv("HAWKEYE_SCORE_THRESHOLD", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f > 0 && f <= 1 {
			cfg.SessionScoreThreshold = f
		}
	}

	if val := config.GetEnv("HAWKEYE_ENABLE_DECAY", ""); val != "" {
		cfg.SessionEnableDecay = val == "true" || val == "1"
	}

	if val := config.GetEnv("HAWKEYE_DECAY_HALF_LIFE", ""); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			cfg.SessionDecayHalfLifeSeconds = i
		}
	}

	// Signal weights
	if val := config.GetEnv("HAWKEYE_WEIGHT_RAGE", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f <= 1 {
			cfg.WeightRage = f
		}
	}

	if val := config.GetEnv("HAWKEYE_WEIGHT_RAGE_BAIT", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f <= 1 {
			cfg.WeightRageBait = f
		}
	}

	if val := config.GetEnv("HAWKEYE_WEIGHT_BLOCKED", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f <= 1 {
			cfg.WeightBlocked = f
		}
	}

	if val := config.GetEnv("HAWKEYE_WEIGHT_ABANDONMENT", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f <= 1 {
			cfg.WeightAbandonment = f
		}
	}

	if val := config.GetEnv("HAWKEYE_WEIGHT_CONFUSION", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f <= 1 {
			cfg.WeightConfusion = f
		}
	}

	if val := config.GetEnv("HAWKEYE_WEIGHT_FORM_LOOP", ""); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f <= 1 {
			cfg.WeightFormLoop = f
		}
	}

	// Sensitivity preset
	if val := config.GetEnv("HAWKEYE_SENSITIVITY", ""); val != "" {
		cfg.SensitivityLevel = val
		applySensitivityPreset(&cfg, val)
	}

	// Environment
	if val := config.GetEnv("HAWKEYE_ENVIRONMENT", ""); val != "" {
		cfg.Environment = val
	} else if val := config.GetEnv("ENVIRONMENT", ""); val != "" {
		cfg.Environment = val
	}

	return cfg
}

// applySensitivityPreset applies sensitivity preset values
func applySensitivityPreset(cfg *DetectionConfig, level string) {
	switch level {
	case "low":
		// Low sensitivity = fewer false positives
		cfg.SessionScoreThreshold = 0.7
		cfg.SessionMinSignalsForFrustration = 3
		cfg.SessionTimeWindowSeconds = 20
	case "high":
		// High sensitivity = catch more frustration
		cfg.SessionScoreThreshold = 0.3
		cfg.SessionMinSignalsForFrustration = 1
		cfg.SessionTimeWindowSeconds = 45
	case "medium":
		// Medium is the default
		cfg.SessionScoreThreshold = 0.5
		cfg.SessionMinSignalsForFrustration = 2
		cfg.SessionTimeWindowSeconds = 30
	}
}

// GetConfig returns the current detection configuration
// This can be called from anywhere in the package
var currentConfig = func() DetectionConfig {
	cfg := LoadDetectionConfig()
	UpdateEnhancedDetectionStatus(cfg.UseEnhancedDetection)
	return cfg
}()

// GetConfig returns the current configuration
func GetConfig() DetectionConfig {
	return currentConfig
}

// ReloadConfig reloads configuration from environment
func ReloadConfig() {
	currentConfig = LoadDetectionConfig()
	UpdateEnhancedDetectionStatus(currentConfig.UseEnhancedDetection)
}

// IsEnhancedDetectionEnabled checks if enhanced detection is enabled
func IsEnhancedDetectionEnabled() bool {
	return currentConfig.UseEnhancedDetection
}

// IsSessionAggregationEnabled checks if session aggregation is enabled
func IsSessionAggregationEnabled() bool {
	return currentConfig.UseSessionAggregation
}

// IsDevelopmentEnvironment returns true if running in development mode
func IsDevelopmentEnvironment() bool {
	env := currentConfig.Environment
	return env == "development" || env == "dev"
}
