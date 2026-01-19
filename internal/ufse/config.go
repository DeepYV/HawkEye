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
	UseEnhancedDetection bool
	
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
}

// DefaultDetectionConfig returns default configuration
func DefaultDetectionConfig() DetectionConfig {
	return DetectionConfig{
		UseEnhancedDetection: false, // Feature flag: disabled by default
		
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
	}
}

// LoadDetectionConfig loads configuration from environment variables
func LoadDetectionConfig() DetectionConfig {
	cfg := DefaultDetectionConfig()
	
	// Feature flag
	if val := config.GetEnv("UFSE_ENHANCED_DETECTION", ""); val != "" {
		cfg.UseEnhancedDetection = val == "true" || val == "1"
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
	
	return cfg
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
