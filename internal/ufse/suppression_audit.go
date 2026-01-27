/**
 * Suppression Tracking and Auditing
 *
 * Author: Enhanced Detection Team
 * Responsibility: Track why incidents are suppressed and provide audit trail
 *
 * Features:
 * - Log every suppression with detailed reason
 * - Track false-positive suppressions over time
 * - Provide metrics for tuning suppression rules
 * - Enable debugging of detection decisions
 */

package ufse

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// SuppressionReason describes why a signal/incident was suppressed
type SuppressionReason string

const (
	ReasonLowConfidence        SuppressionReason = "low_confidence"
	ReasonMediumConfidence     SuppressionReason = "medium_confidence"
	ReasonFalseAlarmDetected   SuppressionReason = "false_alarm_detected"
	ReasonBelowThreshold       SuppressionReason = "below_threshold"
	ReasonDecayedSignal        SuppressionReason = "decayed_signal"
	ReasonSessionTooShort      SuppressionReason = "session_too_short"
	ReasonNoSystemFeedback     SuppressionReason = "no_system_feedback"
	ReasonShadowMode           SuppressionReason = "shadow_mode"
	ReasonRouteDisabled        SuppressionReason = "route_disabled"
	ReasonDuplicate            SuppressionReason = "duplicate"
	ReasonManualSuppression    SuppressionReason = "manual_suppression"
	ReasonExplanationFailed    SuppressionReason = "explanation_failed"
	ReasonCorrelationFailed    SuppressionReason = "correlation_failed"
)

// SuppressionEvent records a suppression decision
type SuppressionEvent struct {
	ID               string            `json:"id"`
	Timestamp        time.Time         `json:"timestamp"`
	SessionID        string            `json:"sessionId"`
	Route            string            `json:"route"`
	SignalType       string            `json:"signalType"`
	SignalStrength   float64           `json:"signalStrength"`
	Reason           SuppressionReason `json:"reason"`
	ReasonDetails    string            `json:"reasonDetails"`
	DetectorName     string            `json:"detectorName"`
	ConfidenceLevel  string            `json:"confidenceLevel"`
	FrustrationScore int               `json:"frustrationScore"`
	RelatedSignals   []string          `json:"relatedSignals"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// SuppressionAuditLog maintains an audit log of suppressions
type SuppressionAuditLog struct {
	mu      sync.RWMutex
	events  []SuppressionEvent
	maxSize int

	// Aggregated stats
	countByReason    map[SuppressionReason]int64
	countBySignalType map[string]int64
	countByRoute     map[string]int64
	countByHour      map[string]int64

	// Running totals
	totalSuppressions int64
	startTime         time.Time
}

// NewSuppressionAuditLog creates a new suppression audit log
func NewSuppressionAuditLog(maxSize int) *SuppressionAuditLog {
	return &SuppressionAuditLog{
		events:           make([]SuppressionEvent, 0, maxSize),
		maxSize:          maxSize,
		countByReason:    make(map[SuppressionReason]int64),
		countBySignalType: make(map[string]int64),
		countByRoute:     make(map[string]int64),
		countByHour:      make(map[string]int64),
		startTime:        time.Now(),
	}
}

// RecordSuppression records a suppression event
func (l *SuppressionAuditLog) RecordSuppression(event SuppressionEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Generate ID if not set
	if event.ID == "" {
		event.ID = generateSuppressionID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Append event
	l.events = append(l.events, event)

	// Trim if over max size
	if len(l.events) > l.maxSize {
		l.events = l.events[len(l.events)-l.maxSize:]
	}

	// Update aggregated stats
	l.totalSuppressions++
	l.countByReason[event.Reason]++
	l.countBySignalType[event.SignalType]++
	l.countByRoute[event.Route]++

	hourKey := event.Timestamp.Format("2006-01-02-15")
	l.countByHour[hourKey]++

	// Log suppression for debugging
	log.Printf("[Suppression Audit] signal=%s, route=%s, reason=%s, details=%s",
		event.SignalType, event.Route, event.Reason, event.ReasonDetails)
}

// GetRecentSuppressions returns recent suppression events
func (l *SuppressionAuditLog) GetRecentSuppressions(limit int) []SuppressionEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if limit <= 0 || limit > len(l.events) {
		limit = len(l.events)
	}

	start := len(l.events) - limit
	if start < 0 {
		start = 0
	}

	result := make([]SuppressionEvent, limit)
	copy(result, l.events[start:])
	return result
}

// GetSuppressionsByReason returns suppressions filtered by reason
func (l *SuppressionAuditLog) GetSuppressionsByReason(reason SuppressionReason, limit int) []SuppressionEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]SuppressionEvent, 0)

	for i := len(l.events) - 1; i >= 0 && len(result) < limit; i-- {
		if l.events[i].Reason == reason {
			result = append(result, l.events[i])
		}
	}

	return result
}

// GetSuppressionStats returns aggregated suppression statistics
func (l *SuppressionAuditLog) GetSuppressionStats() SuppressionStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := SuppressionStats{
		TotalSuppressions:  l.totalSuppressions,
		StartTime:          l.startTime,
		CurrentTime:        time.Now(),
		ByReason:           make(map[string]int64),
		BySignalType:       make(map[string]int64),
		TopRoutes:          make([]RouteSuppressionCount, 0),
		HourlyDistribution: make([]HourlyCount, 0),
	}

	// Copy reason counts
	for reason, count := range l.countByReason {
		stats.ByReason[string(reason)] = count
	}

	// Copy signal type counts
	for signalType, count := range l.countBySignalType {
		stats.BySignalType[signalType] = count
	}

	// Get top routes (up to 10)
	type routeCount struct {
		route string
		count int64
	}
	routeCounts := make([]routeCount, 0, len(l.countByRoute))
	for route, count := range l.countByRoute {
		routeCounts = append(routeCounts, routeCount{route, count})
	}
	// Sort by count (simple bubble sort for small lists)
	for i := 0; i < len(routeCounts); i++ {
		for j := i + 1; j < len(routeCounts); j++ {
			if routeCounts[j].count > routeCounts[i].count {
				routeCounts[i], routeCounts[j] = routeCounts[j], routeCounts[i]
			}
		}
	}
	// Take top 10
	limit := 10
	if len(routeCounts) < limit {
		limit = len(routeCounts)
	}
	for i := 0; i < limit; i++ {
		stats.TopRoutes = append(stats.TopRoutes, RouteSuppressionCount{
			Route: routeCounts[i].route,
			Count: routeCounts[i].count,
		})
	}

	// Hourly distribution (last 24 hours)
	now := time.Now()
	for i := 23; i >= 0; i-- {
		hourTime := now.Add(-time.Duration(i) * time.Hour)
		hourKey := hourTime.Format("2006-01-02-15")
		count := l.countByHour[hourKey]
		stats.HourlyDistribution = append(stats.HourlyDistribution, HourlyCount{
			Hour:  hourTime.Format("15:00"),
			Count: count,
		})
	}

	return stats
}

// SuppressionStats contains aggregated suppression statistics
type SuppressionStats struct {
	TotalSuppressions  int64                    `json:"totalSuppressions"`
	StartTime          time.Time                `json:"startTime"`
	CurrentTime        time.Time                `json:"currentTime"`
	ByReason           map[string]int64         `json:"byReason"`
	BySignalType       map[string]int64         `json:"bySignalType"`
	TopRoutes          []RouteSuppressionCount  `json:"topRoutes"`
	HourlyDistribution []HourlyCount            `json:"hourlyDistribution"`
}

// RouteSuppressionCount counts suppressions per route
type RouteSuppressionCount struct {
	Route string `json:"route"`
	Count int64  `json:"count"`
}

// HourlyCount counts suppressions per hour
type HourlyCount struct {
	Hour  string `json:"hour"`
	Count int64  `json:"count"`
}

// ToJSON converts stats to JSON
func (s *SuppressionStats) ToJSON() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// generateSuppressionID generates a unique suppression ID
func generateSuppressionID() string {
	return time.Now().Format("20060102150405.000000")
}

// CreateSuppressionEvent creates a suppression event from a candidate signal
func CreateSuppressionEvent(
	candidate signals.CandidateSignal,
	sessionID string,
	reason SuppressionReason,
	details string,
	confidenceLevel string,
	frustrationScore int,
) SuppressionEvent {
	// Extract signal strength from details if available
	var signalStrength float64
	if candidate.Details != nil {
		if strength, ok := candidate.Details["signal_strength"].(float64); ok {
			signalStrength = strength
		}
	}

	return SuppressionEvent{
		Timestamp:        time.Now(),
		SessionID:        sessionID,
		Route:            candidate.Route,
		SignalType:       candidate.Type,
		SignalStrength:   signalStrength,
		Reason:           reason,
		ReasonDetails:    details,
		DetectorName:     candidate.Type + "_detector",
		ConfidenceLevel:  confidenceLevel,
		FrustrationScore: frustrationScore,
		RelatedSignals:   []string{candidate.Type},
		Metadata:         candidate.Details,
	}
}

// Global suppression audit log instance
var globalSuppressionLog = NewSuppressionAuditLog(10000)

// GetGlobalSuppressionLog returns the global suppression audit log
func GetGlobalSuppressionLog() *SuppressionAuditLog {
	return globalSuppressionLog
}

// RecordGlobalSuppression records a suppression to the global log
func RecordGlobalSuppression(event SuppressionEvent) {
	globalSuppressionLog.RecordSuppression(event)
}
