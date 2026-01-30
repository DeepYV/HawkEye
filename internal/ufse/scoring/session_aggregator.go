/**
 * Session-Level Frustration Aggregator
 *
 * Author: AI Issue Solver
 * Responsibility: Aggregate multiple signals within a time window for session-level frustration detection
 *
 * Key Features:
 * - Sliding time window aggregation
 * - Configurable signal weights for unequal contribution
 * - Session-level context evaluation
 * - Deterministic and explainable scoring
 */

package scoring

import (
	"math"
	"sync"
	"time"
)

const (
	// DefaultTimeWindowSeconds is the default time window for aggregating signals
	DefaultTimeWindowSeconds = 30

	// DefaultMinSignalsForFrustration is the minimum number of signals needed to trigger frustration
	DefaultMinSignalsForFrustration = 2

	// DefaultAggregatedScoreThreshold is the minimum aggregated score to consider frustration
	DefaultAggregatedScoreThreshold = 0.5
)

// SignalWeight represents configurable weights for each signal type
type SignalWeight struct {
	Type   string  // Signal type (rage, blocked, abandonment, confusion, form_loop, rage_bait)
	Weight float64 // Weight contribution (0.0 to 1.0)
}

// DefaultSignalWeights provides the default weights for each signal type
var DefaultSignalWeights = map[string]float64{
	"rage":        0.35,
	"rage_bait":   0.50, // Higher weight - dark pattern detection
	"blocked":     0.40,
	"abandonment": 0.30,
	"confusion":   0.15, // Lower weight - often ambiguous
	"form_loop":   0.35,
}

// AggregatorConfig holds configuration for the session aggregator
type AggregatorConfig struct {
	// TimeWindowSeconds defines how long signals are considered together
	TimeWindowSeconds int

	// MinSignalsForFrustration is the minimum number of signals needed
	MinSignalsForFrustration int

	// AggregatedScoreThreshold is the minimum score to consider frustration
	AggregatedScoreThreshold float64

	// SignalWeights maps signal types to their weight contribution
	SignalWeights map[string]float64

	// EnableDecay applies time-based decay to older signals in the window
	EnableDecay bool

	// DecayHalfLifeSeconds is how long until a signal loses half its weight
	DecayHalfLifeSeconds int
}

// DefaultAggregatorConfig returns the default aggregator configuration
func DefaultAggregatorConfig() AggregatorConfig {
	return AggregatorConfig{
		TimeWindowSeconds:        DefaultTimeWindowSeconds,
		MinSignalsForFrustration: DefaultMinSignalsForFrustration,
		AggregatedScoreThreshold: DefaultAggregatedScoreThreshold,
		SignalWeights:            DefaultSignalWeights,
		EnableDecay:              true,
		DecayHalfLifeSeconds:     15, // 15-second half-life
	}
}

// AggregatedSignal represents a signal with its timestamp and metadata
type AggregatedSignal struct {
	Type       string
	Timestamp  time.Time
	Strength   float64 // 0.0 to 1.0 signal strength
	Route      string
	Details    map[string]interface{}
}

// AggregationResult contains the result of session-level aggregation
type AggregationResult struct {
	// TotalScore is the aggregated frustration score (0.0 to 1.0)
	TotalScore float64

	// SignalCount is the number of signals in the time window
	SignalCount int

	// SignalBreakdown shows contribution by signal type
	SignalBreakdown map[string]float64

	// IsFrustrated indicates if the threshold was exceeded
	IsFrustrated bool

	// Confidence is the confidence level of the detection
	Confidence string // "low", "medium", "high"

	// Reasoning explains why the decision was made
	Reasoning string

	// TimeWindowStart is when the aggregation window started
	TimeWindowStart time.Time

	// TimeWindowEnd is when the aggregation window ended
	TimeWindowEnd time.Time
}

// SessionAggregator aggregates signals at the session level
type SessionAggregator struct {
	mu      sync.RWMutex
	config  AggregatorConfig
	signals []AggregatedSignal
}

// NewSessionAggregator creates a new session aggregator with default config
func NewSessionAggregator() *SessionAggregator {
	return &SessionAggregator{
		config:  DefaultAggregatorConfig(),
		signals: make([]AggregatedSignal, 0),
	}
}

// NewSessionAggregatorWithConfig creates a session aggregator with custom config
func NewSessionAggregatorWithConfig(config AggregatorConfig) *SessionAggregator {
	// Fill in defaults for any missing values
	if config.TimeWindowSeconds <= 0 {
		config.TimeWindowSeconds = DefaultTimeWindowSeconds
	}
	if config.MinSignalsForFrustration <= 0 {
		config.MinSignalsForFrustration = DefaultMinSignalsForFrustration
	}
	if config.AggregatedScoreThreshold <= 0 {
		config.AggregatedScoreThreshold = DefaultAggregatedScoreThreshold
	}
	if config.SignalWeights == nil {
		config.SignalWeights = DefaultSignalWeights
	}
	if config.DecayHalfLifeSeconds <= 0 {
		config.DecayHalfLifeSeconds = 15
	}

	return &SessionAggregator{
		config:  config,
		signals: make([]AggregatedSignal, 0),
	}
}

// AddSignal adds a signal to the aggregator
func (sa *SessionAggregator) AddSignal(signal AggregatedSignal) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// Set default strength if not provided
	if signal.Strength <= 0 {
		signal.Strength = 0.5 // Default medium strength
	}

	sa.signals = append(sa.signals, signal)
}

// Aggregate computes the session-level frustration score
func (sa *SessionAggregator) Aggregate(referenceTime time.Time) AggregationResult {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	result := AggregationResult{
		SignalBreakdown: make(map[string]float64),
		TimeWindowEnd:   referenceTime,
		TimeWindowStart: referenceTime.Add(-time.Duration(sa.config.TimeWindowSeconds) * time.Second),
	}

	// Filter signals within the time window
	windowStart := result.TimeWindowStart
	relevantSignals := make([]AggregatedSignal, 0)

	for _, signal := range sa.signals {
		if signal.Timestamp.After(windowStart) && !signal.Timestamp.After(referenceTime) {
			relevantSignals = append(relevantSignals, signal)
		}
	}

	result.SignalCount = len(relevantSignals)

	// Check minimum signal count
	if result.SignalCount < sa.config.MinSignalsForFrustration {
		result.IsFrustrated = false
		result.Confidence = "low"
		result.Reasoning = "Insufficient signals in time window for frustration detection"
		return result
	}

	// Calculate weighted score with optional decay
	totalWeight := 0.0
	maxPossibleWeight := 0.0

	for _, signal := range relevantSignals {
		weight := sa.getSignalWeight(signal.Type)
		strength := signal.Strength

		// Apply time decay if enabled
		if sa.config.EnableDecay {
			decayFactor := sa.calculateDecay(signal.Timestamp, referenceTime)
			strength *= decayFactor
		}

		contribution := weight * strength
		totalWeight += contribution
		maxPossibleWeight += weight

		// Track breakdown by type
		result.SignalBreakdown[signal.Type] += contribution
	}

	// Normalize score to 0-1 range
	if maxPossibleWeight > 0 {
		result.TotalScore = totalWeight / maxPossibleWeight
	}

	// Determine if frustrated based on threshold
	result.IsFrustrated = result.TotalScore >= sa.config.AggregatedScoreThreshold

	// Determine confidence level based on multiple factors
	result.Confidence = sa.determineConfidence(result)

	// Generate reasoning
	result.Reasoning = sa.generateReasoning(result, relevantSignals)

	return result
}

// AggregateSignals aggregates a slice of signals (stateless operation)
func (sa *SessionAggregator) AggregateSignals(signals []AggregatedSignal, referenceTime time.Time) AggregationResult {
	result := AggregationResult{
		SignalBreakdown: make(map[string]float64),
		TimeWindowEnd:   referenceTime,
		TimeWindowStart: referenceTime.Add(-time.Duration(sa.config.TimeWindowSeconds) * time.Second),
	}

	// Filter signals within the time window
	windowStart := result.TimeWindowStart
	relevantSignals := make([]AggregatedSignal, 0)

	for _, signal := range signals {
		if signal.Timestamp.After(windowStart) && !signal.Timestamp.After(referenceTime) {
			relevantSignals = append(relevantSignals, signal)
		}
	}

	result.SignalCount = len(relevantSignals)

	// Check minimum signal count
	if result.SignalCount < sa.config.MinSignalsForFrustration {
		result.IsFrustrated = false
		result.Confidence = "low"
		result.Reasoning = "Insufficient signals in time window for frustration detection"
		return result
	}

	// Calculate weighted score with optional decay
	totalWeight := 0.0
	maxPossibleWeight := 0.0

	for _, signal := range relevantSignals {
		weight := sa.getSignalWeight(signal.Type)
		strength := signal.Strength
		if strength <= 0 {
			strength = 0.5 // Default
		}

		// Apply time decay if enabled
		if sa.config.EnableDecay {
			decayFactor := sa.calculateDecay(signal.Timestamp, referenceTime)
			strength *= decayFactor
		}

		contribution := weight * strength
		totalWeight += contribution
		maxPossibleWeight += weight

		// Track breakdown by type
		result.SignalBreakdown[signal.Type] += contribution
	}

	// Normalize score to 0-1 range
	if maxPossibleWeight > 0 {
		result.TotalScore = totalWeight / maxPossibleWeight
	}

	// Determine if frustrated based on threshold
	result.IsFrustrated = result.TotalScore >= sa.config.AggregatedScoreThreshold

	// Determine confidence level based on multiple factors
	result.Confidence = sa.determineConfidence(result)

	// Generate reasoning
	result.Reasoning = sa.generateReasoning(result, relevantSignals)

	return result
}

// getSignalWeight returns the weight for a signal type
func (sa *SessionAggregator) getSignalWeight(signalType string) float64 {
	if weight, ok := sa.config.SignalWeights[signalType]; ok {
		return weight
	}
	return 0.2 // Default weight for unknown types
}

// calculateDecay calculates the time-based decay factor
func (sa *SessionAggregator) calculateDecay(signalTime, referenceTime time.Time) float64 {
	elapsed := referenceTime.Sub(signalTime).Seconds()
	halfLife := float64(sa.config.DecayHalfLifeSeconds)

	// Exponential decay: factor = 0.5^(elapsed/halfLife)
	factor := math.Pow(0.5, elapsed/halfLife)

	// Clamp to minimum of 0.1 to not completely discard older signals
	if factor < 0.1 {
		return 0.1
	}
	return factor
}

// determineConfidence determines the confidence level of the detection
func (sa *SessionAggregator) determineConfidence(result AggregationResult) string {
	// Higher confidence with more signals and higher score
	score := result.TotalScore
	count := result.SignalCount

	if score >= 0.8 && count >= 4 {
		return "high"
	}
	if score >= 0.6 && count >= 3 {
		return "high"
	}
	if score >= 0.5 && count >= 2 {
		return "medium"
	}
	if score >= 0.3 {
		return "medium"
	}
	return "low"
}

// generateReasoning generates a human-readable explanation
func (sa *SessionAggregator) generateReasoning(result AggregationResult, signals []AggregatedSignal) string {
	if !result.IsFrustrated {
		if result.SignalCount < sa.config.MinSignalsForFrustration {
			return "No frustration detected: insufficient signals in the time window"
		}
		return "No frustration detected: aggregated score below threshold"
	}

	// Count signal types
	typeCounts := make(map[string]int)
	for _, signal := range signals {
		typeCounts[signal.Type]++
	}

	// Build reasoning
	reasoning := "Frustration detected: "
	first := true
	for signalType, count := range typeCounts {
		if !first {
			reasoning += ", "
		}
		if count == 1 {
			reasoning += signalType
		} else {
			reasoning += signalType + " (x" + string(rune('0'+count)) + ")"
		}
		first = false
	}
	reasoning += " signals aggregated within time window"

	return reasoning
}

// Clear removes all signals from the aggregator
func (sa *SessionAggregator) Clear() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.signals = make([]AggregatedSignal, 0)
}

// PruneOldSignals removes signals older than the time window
func (sa *SessionAggregator) PruneOldSignals(referenceTime time.Time) int {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	cutoff := referenceTime.Add(-time.Duration(sa.config.TimeWindowSeconds) * time.Second)
	newSignals := make([]AggregatedSignal, 0)
	pruned := 0

	for _, signal := range sa.signals {
		if signal.Timestamp.After(cutoff) {
			newSignals = append(newSignals, signal)
		} else {
			pruned++
		}
	}

	sa.signals = newSignals
	return pruned
}

// GetConfig returns the current configuration (for debugging/testing)
func (sa *SessionAggregator) GetConfig() AggregatorConfig {
	return sa.config
}

// UpdateConfig updates the aggregator configuration
func (sa *SessionAggregator) UpdateConfig(config AggregatorConfig) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.config = config
}
