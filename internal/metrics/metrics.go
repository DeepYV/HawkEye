// Package metrics provides Prometheus instrumentation for HawkEye.
//
// All metrics follow the hawkeye_ namespace convention and are registered
// automatically via promauto.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EventsIngested counts total events received.
	EventsIngested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hawkeye_events_ingested_total",
		Help: "Total number of events ingested from SDK",
	})

	// SessionsCreated counts sessions created by the session manager.
	SessionsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hawkeye_sessions_created_total",
		Help: "Total number of sessions created",
	})

	// SessionsProcessed counts sessions processed by the engine.
	SessionsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hawkeye_sessions_processed_total",
		Help: "Total number of sessions processed by detection engine",
	})

	// IncidentsDetected counts incidents emitted by the engine.
	IncidentsDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "hawkeye_incidents_detected_total",
		Help: "Total number of frustration incidents detected",
	})

	// ProcessingLatency tracks session processing duration.
	ProcessingLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "hawkeye_processing_latency_seconds",
		Help:    "Session processing latency in seconds",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	})

	// EventQueueDepth tracks the current depth of the event processing queue.
	EventQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hawkeye_event_queue_depth",
		Help: "Current depth of the event processing queue",
	})

	// SignalsDetected counts signals detected with type label.
	SignalsDetected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hawkeye_signals_detected_total",
		Help: "Total signals detected by type",
	}, []string{"type"})

	// SignalsDiscarded counts signals discarded with reason label.
	SignalsDiscarded = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hawkeye_signals_discarded_total",
		Help: "Total signals discarded by reason",
	}, []string{"reason"})

	// HTTPRequestsTotal counts HTTP requests by method and status.
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hawkeye_http_requests_total",
		Help: "Total HTTP requests by method, path, and status",
	}, []string{"method", "path", "status"})

	// HTTPRequestDuration tracks HTTP request latency.
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "hawkeye_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)
