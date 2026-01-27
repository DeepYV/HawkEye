/**
 * Signal Decay Logic
 *
 * Author: Enhanced Detection Team
 * Responsibility: Apply time-based decay to signals for noise reduction
 *
 * Rationale:
 * - Old signals are less indicative of current frustration
 * - One-off signals that don't repeat may be user error, not frustration
 * - Decay helps prevent false positives from stale data
 *
 * Decay Formula:
 * DecayedStrength = OriginalStrength * DecayFactor
 * DecayFactor = 1 / (1 + DecayRate * TimeSinceSignal)
 *
 * Configuration:
 * - Signals older than MaxSignalAge are completely discarded
 * - DecayRate controls how fast signals lose significance
 */

package scoring

import (
	"time"
)

const (
	// MaxSignalAge is the maximum age of a signal before it's discarded
	MaxSignalAge = 5 * time.Minute

	// HalfLifeSignal is the time at which signal strength drops to 50%
	HalfLifeSignal = 60 * time.Second

	// MinDecayedStrength is the minimum strength after decay (below this, discard)
	MinDecayedStrength = 0.1

	// RecentSignalBoost is a multiplier for very recent signals (within 10s)
	RecentSignalBoost = 1.2
	RecentSignalWindow = 10 * time.Second
)

// SignalDecayConfig holds configuration for signal decay
type SignalDecayConfig struct {
	MaxAge           time.Duration
	HalfLife         time.Duration
	MinStrength      float64
	RecentBoost      float64
	RecentWindow     time.Duration
	EnableDecay      bool
}

// DefaultDecayConfig returns default decay configuration
func DefaultDecayConfig() SignalDecayConfig {
	return SignalDecayConfig{
		MaxAge:       MaxSignalAge,
		HalfLife:     HalfLifeSignal,
		MinStrength:  MinDecayedStrength,
		RecentBoost:  RecentSignalBoost,
		RecentWindow: RecentSignalWindow,
		EnableDecay:  true,
	}
}

// SignalDecayer applies decay to signals based on age
type SignalDecayer struct {
	config SignalDecayConfig
}

// NewSignalDecayer creates a new signal decayer with default config
func NewSignalDecayer() *SignalDecayer {
	return &SignalDecayer{
		config: DefaultDecayConfig(),
	}
}

// NewSignalDecayerWithConfig creates a new signal decayer with custom config
func NewSignalDecayerWithConfig(config SignalDecayConfig) *SignalDecayer {
	return &SignalDecayer{
		config: config,
	}
}

// DecayedSignal represents a signal with decay applied
type DecayedSignal struct {
	OriginalStrength float64
	DecayedStrength  float64
	Age              time.Duration
	IsStale          bool
	DecayFactor      float64
}

// ApplyDecay applies decay to a signal based on its age
func (d *SignalDecayer) ApplyDecay(originalStrength float64, signalTimestamp time.Time, referenceTime time.Time) DecayedSignal {
	age := referenceTime.Sub(signalTimestamp)

	// Check if signal is too old
	if age > d.config.MaxAge {
		return DecayedSignal{
			OriginalStrength: originalStrength,
			DecayedStrength:  0,
			Age:              age,
			IsStale:          true,
			DecayFactor:      0,
		}
	}

	// If decay is disabled, return original strength
	if !d.config.EnableDecay {
		return DecayedSignal{
			OriginalStrength: originalStrength,
			DecayedStrength:  originalStrength,
			Age:              age,
			IsStale:          false,
			DecayFactor:      1.0,
		}
	}

	// Calculate decay factor using half-life formula
	// DecayFactor = 0.5 ^ (age / halfLife)
	decayFactor := exponentialDecay(age, d.config.HalfLife)

	// Apply recent signal boost
	if age < d.config.RecentWindow {
		decayFactor *= d.config.RecentBoost
		if decayFactor > 1.0 {
			decayFactor = 1.0
		}
	}

	decayedStrength := originalStrength * decayFactor

	// Check if below minimum threshold
	isStale := decayedStrength < d.config.MinStrength

	return DecayedSignal{
		OriginalStrength: originalStrength,
		DecayedStrength:  decayedStrength,
		Age:              age,
		IsStale:          isStale,
		DecayFactor:      decayFactor,
	}
}

// FilterStaleSignals filters out signals that have decayed below threshold
func (d *SignalDecayer) FilterStaleSignals(signals []SignalWithTimestamp, referenceTime time.Time) []SignalWithTimestamp {
	filtered := make([]SignalWithTimestamp, 0, len(signals))

	for _, signal := range signals {
		decayed := d.ApplyDecay(signal.Strength, signal.Timestamp, referenceTime)
		if !decayed.IsStale {
			signal.Strength = decayed.DecayedStrength
			filtered = append(filtered, signal)
		}
	}

	return filtered
}

// SignalWithTimestamp represents a signal with its timestamp and strength
type SignalWithTimestamp struct {
	Type      string
	Timestamp time.Time
	Strength  float64
	Route     string
	Details   map[string]interface{}
}

// exponentialDecay calculates exponential decay factor
// Returns value between 0 and 1
func exponentialDecay(age time.Duration, halfLife time.Duration) float64 {
	if halfLife <= 0 {
		return 1.0
	}

	// Formula: 0.5 ^ (age / halfLife)
	exponent := float64(age) / float64(halfLife)
	return pow(0.5, exponent)
}

// pow calculates x^y for small positive exponents
func pow(x, y float64) float64 {
	if y == 0 {
		return 1
	}
	if y == 1 {
		return x
	}
	if x == 0 {
		return 0
	}

	// Use natural logarithm for accurate calculation
	// x^y = e^(y * ln(x))
	return exp(y * ln(x))
}

// ln calculates natural logarithm using Taylor series
func ln(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x == 1 {
		return 0
	}

	// For 0.5, ln(0.5) ≈ -0.693147
	// Use the known value for common case
	if x == 0.5 {
		return -0.693147180559945
	}

	// Taylor series approximation for ln(1+u) where u = x-1
	// ln(1+u) = u - u^2/2 + u^3/3 - u^4/4 + ...
	u := x - 1
	result := 0.0
	term := u
	for i := 1; i <= 20; i++ {
		if i%2 == 1 {
			result += term / float64(i)
		} else {
			result -= term / float64(i)
		}
		term *= u
	}
	return result
}

// exp calculates e^x using Taylor series
func exp(x float64) float64 {
	if x == 0 {
		return 1
	}

	// Taylor series: e^x = 1 + x + x^2/2! + x^3/3! + ...
	result := 1.0
	term := 1.0
	for i := 1; i <= 30; i++ {
		term *= x / float64(i)
		result += term
		// Early termination for convergence
		if term < 1e-15 && term > -1e-15 {
			break
		}
	}
	return result
}

// CalculateDecayedScore calculates a frustration score with decay applied
func (d *SignalDecayer) CalculateDecayedScore(signals []SignalWithTimestamp, referenceTime time.Time) float64 {
	if len(signals) == 0 {
		return 0
	}

	totalScore := 0.0

	for _, signal := range signals {
		decayed := d.ApplyDecay(signal.Strength, signal.Timestamp, referenceTime)
		if !decayed.IsStale {
			totalScore += decayed.DecayedStrength * 20 // Base multiplier per signal
		}
	}

	// Cap at 100
	if totalScore > 100 {
		return 100
	}

	return totalScore
}
