/**
 * Shadow Mode Detection Support
 *
 * Author: Enhanced Detection Team
 * Responsibility: Support shadow-mode detectors that detect but don't export
 *
 * Shadow Mode Features:
 * - Detectors run and log results but don't create incidents
 * - Useful for testing new detection logic before production
 * - Allows A/B testing of detection changes
 * - Provides metrics for tuning thresholds
 *
 * Implementation:
 * - ShadowDetector wraps any detector and intercepts emissions
 * - ShadowMetrics tracks detection counts, false positive rates, etc.
 * - ShadowReporter generates reports comparing shadow vs production
 */

package ufse

import (
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// ShadowMode represents shadow mode state for a detector
type ShadowMode struct {
	mu           sync.RWMutex
	enabled      bool
	detectorName string
	metrics      *ShadowMetrics
}

// ShadowMetrics tracks metrics for shadow mode detection
type ShadowMetrics struct {
	mu sync.RWMutex

	// Detection counts
	TotalDetections      int64
	HighConfidenceCount  int64
	MediumConfidenceCount int64
	LowConfidenceCount   int64

	// Signal type breakdown
	SignalTypeCounts map[string]int64

	// Route breakdown
	RouteDetections map[string]int64

	// Time-based metrics
	DetectionsPerHour   map[string]int64 // hour key -> count
	FirstDetectionTime  time.Time
	LastDetectionTime   time.Time

	// Comparison with production
	WouldHaveEmitted    int64 // Count that would have been emitted
	WouldHaveSuppressed int64 // Count that would have been suppressed
}

// NewShadowMode creates a new shadow mode instance
func NewShadowMode(detectorName string) *ShadowMode {
	return &ShadowMode{
		enabled:      false,
		detectorName: detectorName,
		metrics: &ShadowMetrics{
			SignalTypeCounts:  make(map[string]int64),
			RouteDetections:   make(map[string]int64),
			DetectionsPerHour: make(map[string]int64),
		},
	}
}

// Enable enables shadow mode
func (s *ShadowMode) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
	log.Printf("[Shadow Mode] Enabled for detector: %s", s.detectorName)
}

// Disable disables shadow mode
func (s *ShadowMode) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
	log.Printf("[Shadow Mode] Disabled for detector: %s", s.detectorName)
}

// IsEnabled checks if shadow mode is enabled
func (s *ShadowMode) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// RecordDetection records a shadow detection
func (s *ShadowMode) RecordDetection(candidate signals.CandidateSignal, confidenceLevel string, wouldEmit bool) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	now := time.Now()

	// Update counts
	s.metrics.TotalDetections++

	switch confidenceLevel {
	case "High":
		s.metrics.HighConfidenceCount++
	case "Medium":
		s.metrics.MediumConfidenceCount++
	case "Low":
		s.metrics.LowConfidenceCount++
	}

	// Signal type
	s.metrics.SignalTypeCounts[candidate.Type]++

	// Route
	s.metrics.RouteDetections[candidate.Route]++

	// Time tracking
	hourKey := now.Format("2006-01-02-15")
	s.metrics.DetectionsPerHour[hourKey]++

	if s.metrics.FirstDetectionTime.IsZero() {
		s.metrics.FirstDetectionTime = now
	}
	s.metrics.LastDetectionTime = now

	// Emission tracking
	if wouldEmit {
		s.metrics.WouldHaveEmitted++
	} else {
		s.metrics.WouldHaveSuppressed++
	}

	// Log shadow detection
	log.Printf("[Shadow Mode] Detection recorded: detector=%s, type=%s, route=%s, confidence=%s, would_emit=%v",
		s.detectorName, candidate.Type, candidate.Route, confidenceLevel, wouldEmit)
}

// GetMetrics returns a copy of the current metrics
func (s *ShadowMode) GetMetrics() ShadowMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Create a copy
	copy := ShadowMetrics{
		TotalDetections:       s.metrics.TotalDetections,
		HighConfidenceCount:   s.metrics.HighConfidenceCount,
		MediumConfidenceCount: s.metrics.MediumConfidenceCount,
		LowConfidenceCount:    s.metrics.LowConfidenceCount,
		FirstDetectionTime:    s.metrics.FirstDetectionTime,
		LastDetectionTime:     s.metrics.LastDetectionTime,
		WouldHaveEmitted:      s.metrics.WouldHaveEmitted,
		WouldHaveSuppressed:   s.metrics.WouldHaveSuppressed,
		SignalTypeCounts:      make(map[string]int64),
		RouteDetections:       make(map[string]int64),
		DetectionsPerHour:     make(map[string]int64),
	}

	for k, v := range s.metrics.SignalTypeCounts {
		copy.SignalTypeCounts[k] = v
	}
	for k, v := range s.metrics.RouteDetections {
		copy.RouteDetections[k] = v
	}
	for k, v := range s.metrics.DetectionsPerHour {
		copy.DetectionsPerHour[k] = v
	}

	return copy
}

// ResetMetrics resets all metrics
func (s *ShadowMode) ResetMetrics() {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.TotalDetections = 0
	s.metrics.HighConfidenceCount = 0
	s.metrics.MediumConfidenceCount = 0
	s.metrics.LowConfidenceCount = 0
	s.metrics.SignalTypeCounts = make(map[string]int64)
	s.metrics.RouteDetections = make(map[string]int64)
	s.metrics.DetectionsPerHour = make(map[string]int64)
	s.metrics.FirstDetectionTime = time.Time{}
	s.metrics.LastDetectionTime = time.Time{}
	s.metrics.WouldHaveEmitted = 0
	s.metrics.WouldHaveSuppressed = 0
}

// ShadowIncident represents an incident that was detected but not emitted
type ShadowIncident struct {
	Incident        types.Incident
	DetectedAt      time.Time
	WouldHaveEmitted bool
	SuppressedReason string
	DetectorName     string
}

// ShadowIncidentStore stores shadow incidents for analysis
type ShadowIncidentStore struct {
	mu        sync.RWMutex
	incidents []ShadowIncident
	maxSize   int
}

// NewShadowIncidentStore creates a new shadow incident store
func NewShadowIncidentStore(maxSize int) *ShadowIncidentStore {
	return &ShadowIncidentStore{
		incidents: make([]ShadowIncident, 0, maxSize),
		maxSize:   maxSize,
	}
}

// Store stores a shadow incident
func (s *ShadowIncidentStore) Store(incident ShadowIncident) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Append incident
	s.incidents = append(s.incidents, incident)

	// Trim if over max size
	if len(s.incidents) > s.maxSize {
		s.incidents = s.incidents[len(s.incidents)-s.maxSize:]
	}
}

// GetRecent returns recent shadow incidents
func (s *ShadowIncidentStore) GetRecent(limit int) []ShadowIncident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.incidents) {
		limit = len(s.incidents)
	}

	// Return most recent
	start := len(s.incidents) - limit
	if start < 0 {
		start = 0
	}

	result := make([]ShadowIncident, limit)
	copy(result, s.incidents[start:])
	return result
}

// GetByRoute returns shadow incidents for a specific route
func (s *ShadowIncidentStore) GetByRoute(route string, limit int) []ShadowIncident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ShadowIncident, 0)

	for i := len(s.incidents) - 1; i >= 0 && len(result) < limit; i-- {
		if s.incidents[i].Incident.PrimaryFailurePoint == route ||
			containsRoute(s.incidents[i].Incident.PrimaryFailurePoint, route) {
			result = append(result, s.incidents[i])
		}
	}

	return result
}

// containsRoute checks if failure point contains the route
func containsRoute(failurePoint, route string) bool {
	// Simple contains check
	return len(failurePoint) >= len(route) && failurePoint[:len(route)] == route
}

// ShadowModeManager manages shadow mode across all detectors
type ShadowModeManager struct {
	mu            sync.RWMutex
	shadowModes   map[string]*ShadowMode
	incidentStore *ShadowIncidentStore
}

// NewShadowModeManager creates a new shadow mode manager
func NewShadowModeManager() *ShadowModeManager {
	return &ShadowModeManager{
		shadowModes:   make(map[string]*ShadowMode),
		incidentStore: NewShadowIncidentStore(1000),
	}
}

// GetOrCreate gets or creates shadow mode for a detector
func (m *ShadowModeManager) GetOrCreate(detectorName string) *ShadowMode {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sm, exists := m.shadowModes[detectorName]; exists {
		return sm
	}

	sm := NewShadowMode(detectorName)
	m.shadowModes[detectorName] = sm
	return sm
}

// EnableAll enables shadow mode for all detectors
func (m *ShadowModeManager) EnableAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sm := range m.shadowModes {
		sm.Enable()
	}
}

// DisableAll disables shadow mode for all detectors
func (m *ShadowModeManager) DisableAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sm := range m.shadowModes {
		sm.Disable()
	}
}

// GetAllMetrics returns metrics for all shadow mode detectors
func (m *ShadowModeManager) GetAllMetrics() map[string]ShadowMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]ShadowMetrics)
	for name, sm := range m.shadowModes {
		result[name] = sm.GetMetrics()
	}
	return result
}

// StoreIncident stores a shadow incident
func (m *ShadowModeManager) StoreIncident(incident ShadowIncident) {
	m.incidentStore.Store(incident)
}

// GetRecentIncidents returns recent shadow incidents
func (m *ShadowModeManager) GetRecentIncidents(limit int) []ShadowIncident {
	return m.incidentStore.GetRecent(limit)
}
