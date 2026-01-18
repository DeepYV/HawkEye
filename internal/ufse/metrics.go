/**
 * Enhanced Detection Metrics
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Metrics for enhanced frustration detection
 */

package ufse

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Enhanced detection metrics
	enhancedSignalsDetected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ufse_enhanced_signals_detected_total",
			Help: "Total number of signals detected by enhanced detectors",
		},
		[]string{"signal_type", "strength"},
	)
	
	rageBaitDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ufse_rage_bait_detected_total",
			Help: "Total number of rage bait signals detected",
		},
	)
	
	singleSignalIncidents = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ufse_single_signal_incidents_total",
			Help: "Total number of incidents from single-signal correlation",
		},
	)
	
	mediumConfidenceIncidents = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ufse_medium_confidence_incidents_total",
			Help: "Total number of Medium confidence incidents emitted",
		},
	)
	
	signalStrengthScore = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ufse_signal_strength_score",
			Help:    "Distribution of signal strength scores",
			Buckets: []float64{0.0, 0.2, 0.4, 0.6, 0.8, 1.0},
		},
		[]string{"signal_type"},
	)
	
	darkPatternScore = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ufse_dark_pattern_score",
			Help:    "Distribution of dark pattern scores for rage bait",
			Buckets: []float64{0.0, 0.2, 0.4, 0.6, 0.8, 1.0},
		},
	)
	
	enhancedDetectionEnabled = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ufse_enhanced_detection_enabled",
			Help: "Whether enhanced detection is enabled (1=enabled, 0=disabled)",
		},
	)
)

// RecordEnhancedSignal records an enhanced signal detection
func RecordEnhancedSignal(signalType, strength string) {
	enhancedSignalsDetected.WithLabelValues(signalType, strength).Inc()
}

// RecordRageBait records a rage bait detection
func RecordRageBait() {
	rageBaitDetected.Inc()
}

// RecordSingleSignalIncident records a single-signal incident
func RecordSingleSignalIncident() {
	singleSignalIncidents.Inc()
}

// RecordMediumConfidenceIncident records a Medium confidence incident
func RecordMediumConfidenceIncident() {
	mediumConfidenceIncidents.Inc()
}

// RecordSignalStrength records a signal strength score
func RecordSignalStrength(signalType string, score float64) {
	signalStrengthScore.WithLabelValues(signalType).Observe(score)
}

// RecordDarkPatternScore records a dark pattern score
func RecordDarkPatternScore(score float64) {
	darkPatternScore.Observe(score)
}

// UpdateEnhancedDetectionStatus updates the enhanced detection status gauge
func UpdateEnhancedDetectionStatus(enabled bool) {
	if enabled {
		enhancedDetectionEnabled.Set(1)
	} else {
		enhancedDetectionEnabled.Set(0)
	}
}
