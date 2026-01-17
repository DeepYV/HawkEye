/**
 * Observability Metrics
 * 
 * Author: Henry Wilson (Team Beta)
 * Responsibility: Internal observability (critical for trust)
 */

package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ExportAttempts tracks export attempts
	ExportAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ticket_exporter_attempts_total",
		Help: "Total number of export attempts",
	})

	// ExportsSuccessful tracks successful exports
	ExportsSuccessful = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ticket_exporter_successful_total",
		Help: "Total number of successful exports",
	})

	// ExportsSkipped tracks skipped exports with reason
	ExportsSkipped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ticket_exporter_skipped_total",
			Help: "Total number of skipped exports by reason",
		},
		[]string{"reason"},
	)

	// ExportFailures tracks export failures
	ExportFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ticket_exporter_failures_total",
		Help: "Total number of export failures",
	})

	// SessionsProcessed tracks sessions processed by UFSE
	SessionsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ufse_sessions_processed_total",
		Help: "Total number of sessions processed",
	})

	// SignalsDetected tracks signals detected
	SignalsDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ufse_signals_detected_total",
		Help: "Total number of signals detected",
	})

	// SignalsDiscarded tracks signals discarded with reason
	SignalsDiscarded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ufse_signals_discarded_total",
			Help: "Total number of signals discarded by reason",
		},
		[]string{"reason"},
	)

	// IncidentsEmitted tracks incidents emitted
	IncidentsEmitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ufse_incidents_emitted_total",
		Help: "Total number of incidents emitted",
	})

	// ProcessingDuration tracks processing duration
	ProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "ufse_processing_duration_seconds",
		Help: "Duration of session processing in seconds",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
	})
)