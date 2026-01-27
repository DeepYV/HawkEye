/**
 * Detection Reasoning Exposure
 *
 * Author: Enhanced Detection Team
 * Responsibility: Expose detection reasoning for admin/debug tools
 *
 * Features:
 * - Detailed breakdown of why an incident was created
 * - Step-by-step detection pipeline trace
 * - Signal qualification details
 * - Correlation logic explanation
 * - Confidence calculation breakdown
 */

package ufse

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/your-org/frustration-engine/internal/ufse/signals"
)

// DetectionReasoning contains detailed reasoning for a detection decision
type DetectionReasoning struct {
	// Summary
	Outcome          string    `json:"outcome"` // "emitted", "suppressed", "shadow"
	IncidentID       string    `json:"incidentId,omitempty"`
	SessionID        string    `json:"sessionId"`
	Timestamp        time.Time `json:"timestamp"`
	ProcessingTimeMs int64     `json:"processingTimeMs"`

	// Pipeline steps
	Steps []PipelineStep `json:"steps"`

	// Signal details
	DetectedSignals   []SignalReasoning   `json:"detectedSignals"`
	QualifiedSignals  []SignalReasoning   `json:"qualifiedSignals"`
	CorrelatedSignals []SignalReasoning   `json:"correlatedSignals"`

	// Scoring details
	ScoreBreakdown    ScoreBreakdown    `json:"scoreBreakdown"`
	ConfidenceDetails ConfidenceDetails `json:"confidenceDetails"`

	// Final decision
	FinalDecision DecisionDetails `json:"finalDecision"`
}

// PipelineStep describes a step in the detection pipeline
type PipelineStep struct {
	StepNumber  int       `json:"stepNumber"`
	Name        string    `json:"name"`
	Status      string    `json:"status"` // "passed", "failed", "skipped"
	Description string    `json:"description"`
	InputCount  int       `json:"inputCount"`
	OutputCount int       `json:"outputCount"`
	Duration    string    `json:"duration"`
	Details     string    `json:"details,omitempty"`
}

// SignalReasoning describes the reasoning for a signal
type SignalReasoning struct {
	Type            string                 `json:"type"`
	Timestamp       time.Time              `json:"timestamp"`
	Route           string                 `json:"route"`
	Strength        float64                `json:"strength"`
	Details         map[string]interface{} `json:"details"`
	QualificationReason string             `json:"qualificationReason,omitempty"`
	CorrelationReason   string             `json:"correlationReason,omitempty"`
}

// ScoreBreakdown shows how the frustration score was calculated
type ScoreBreakdown struct {
	BaseScore        int    `json:"baseScore"`
	SignalCountScore int    `json:"signalCountScore"`
	TypeWeightScore  int    `json:"typeWeightScore"`
	DurationScore    int    `json:"durationScore"`
	ErrorBonus       int    `json:"errorBonus"`
	DecayPenalty     int    `json:"decayPenalty"`
	SessionPenalty   int    `json:"sessionPenalty"`
	FinalScore       int    `json:"finalScore"`
	Explanation      string `json:"explanation"`
}

// ConfidenceDetails shows how confidence was determined
type ConfidenceDetails struct {
	Level            string   `json:"level"` // "Low", "Medium", "High"
	HasMultipleSignals bool   `json:"hasMultipleSignals"`
	HasSystemFeedback  bool   `json:"hasSystemFeedback"`
	HasClearFailure    bool   `json:"hasClearFailure"`
	SignalTypes        []string `json:"signalTypes"`
	Factors            []string `json:"factors"`
}

// DecisionDetails describes the final emission decision
type DecisionDetails struct {
	ShouldEmit       bool              `json:"shouldEmit"`
	Reason           string            `json:"reason"`
	SuppressionReason SuppressionReason `json:"suppressionReason,omitempty"`
	RouteConfig      string            `json:"routeConfig,omitempty"`
	ShadowMode       bool              `json:"shadowMode"`
}

// DetectionReasoningBuilder builds detection reasoning incrementally
type DetectionReasoningBuilder struct {
	reasoning   DetectionReasoning
	startTime   time.Time
	currentStep int
}

// NewDetectionReasoningBuilder creates a new reasoning builder
func NewDetectionReasoningBuilder(sessionID string) *DetectionReasoningBuilder {
	return &DetectionReasoningBuilder{
		reasoning: DetectionReasoning{
			SessionID:        sessionID,
			Timestamp:        time.Now(),
			Steps:            make([]PipelineStep, 0),
			DetectedSignals:  make([]SignalReasoning, 0),
			QualifiedSignals: make([]SignalReasoning, 0),
			CorrelatedSignals: make([]SignalReasoning, 0),
		},
		startTime: time.Now(),
	}
}

// StartStep starts tracking a pipeline step
func (b *DetectionReasoningBuilder) StartStep(name, description string, inputCount int) *StepTracker {
	b.currentStep++
	return &StepTracker{
		builder:    b,
		stepNumber: b.currentStep,
		name:       name,
		description: description,
		inputCount: inputCount,
		startTime:  time.Now(),
	}
}

// StepTracker tracks a single pipeline step
type StepTracker struct {
	builder     *DetectionReasoningBuilder
	stepNumber  int
	name        string
	description string
	inputCount  int
	startTime   time.Time
}

// Complete marks the step as complete
func (s *StepTracker) Complete(outputCount int, details string) {
	duration := time.Since(s.startTime)
	status := "passed"
	if outputCount == 0 {
		status = "failed"
	}

	s.builder.reasoning.Steps = append(s.builder.reasoning.Steps, PipelineStep{
		StepNumber:  s.stepNumber,
		Name:        s.name,
		Status:      status,
		Description: s.description,
		InputCount:  s.inputCount,
		OutputCount: outputCount,
		Duration:    duration.String(),
		Details:     details,
	})
}

// Skip marks the step as skipped
func (s *StepTracker) Skip(reason string) {
	s.builder.reasoning.Steps = append(s.builder.reasoning.Steps, PipelineStep{
		StepNumber:  s.stepNumber,
		Name:        s.name,
		Status:      "skipped",
		Description: s.description,
		InputCount:  s.inputCount,
		OutputCount: 0,
		Duration:    "0s",
		Details:     reason,
	})
}

// AddDetectedSignal adds a detected signal
func (b *DetectionReasoningBuilder) AddDetectedSignal(candidate signals.CandidateSignal) {
	b.reasoning.DetectedSignals = append(b.reasoning.DetectedSignals, SignalReasoning{
		Type:      candidate.Type,
		Timestamp: time.Unix(candidate.Timestamp, 0),
		Route:     candidate.Route,
		Details:   candidate.Details,
	})
}

// AddQualifiedSignal adds a qualified signal
func (b *DetectionReasoningBuilder) AddQualifiedSignal(signal signals.QualifiedSignal, reason string) {
	b.reasoning.QualifiedSignals = append(b.reasoning.QualifiedSignals, SignalReasoning{
		Type:                signal.Type,
		Timestamp:           signal.Timestamp,
		Route:               signal.Route,
		Strength:            signal.Strength,
		Details:             signal.Details,
		QualificationReason: reason,
	})
}

// AddCorrelatedSignal adds a correlated signal
func (b *DetectionReasoningBuilder) AddCorrelatedSignal(signal signals.QualifiedSignal, reason string) {
	b.reasoning.CorrelatedSignals = append(b.reasoning.CorrelatedSignals, SignalReasoning{
		Type:              signal.Type,
		Timestamp:         signal.Timestamp,
		Route:             signal.Route,
		Strength:          signal.Strength,
		Details:           signal.Details,
		CorrelationReason: reason,
	})
}

// SetScoreBreakdown sets the score breakdown
func (b *DetectionReasoningBuilder) SetScoreBreakdown(breakdown ScoreBreakdown) {
	b.reasoning.ScoreBreakdown = breakdown
}

// SetConfidenceDetails sets the confidence details
func (b *DetectionReasoningBuilder) SetConfidenceDetails(details ConfidenceDetails) {
	b.reasoning.ConfidenceDetails = details
}

// SetFinalDecision sets the final decision
func (b *DetectionReasoningBuilder) SetFinalDecision(decision DecisionDetails) {
	b.reasoning.FinalDecision = decision

	if decision.ShouldEmit {
		b.reasoning.Outcome = "emitted"
	} else if decision.ShadowMode {
		b.reasoning.Outcome = "shadow"
	} else {
		b.reasoning.Outcome = "suppressed"
	}
}

// SetIncidentID sets the incident ID
func (b *DetectionReasoningBuilder) SetIncidentID(id string) {
	b.reasoning.IncidentID = id
}

// Build returns the complete reasoning
func (b *DetectionReasoningBuilder) Build() DetectionReasoning {
	b.reasoning.ProcessingTimeMs = time.Since(b.startTime).Milliseconds()
	return b.reasoning
}

// ToJSON converts reasoning to JSON
func (r *DetectionReasoning) ToJSON() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// ToHumanReadable converts reasoning to human-readable format
func (r *DetectionReasoning) ToHumanReadable() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Detection Reasoning ===\n"))
	sb.WriteString(fmt.Sprintf("Session: %s\n", r.SessionID))
	sb.WriteString(fmt.Sprintf("Outcome: %s\n", r.Outcome))
	sb.WriteString(fmt.Sprintf("Processing Time: %dms\n\n", r.ProcessingTimeMs))

	sb.WriteString("--- Pipeline Steps ---\n")
	for _, step := range r.Steps {
		status := "✓"
		if step.Status == "failed" {
			status = "✗"
		} else if step.Status == "skipped" {
			status = "○"
		}
		sb.WriteString(fmt.Sprintf("%s Step %d: %s (%s)\n", status, step.StepNumber, step.Name, step.Duration))
		sb.WriteString(fmt.Sprintf("   %s (in: %d, out: %d)\n", step.Description, step.InputCount, step.OutputCount))
		if step.Details != "" {
			sb.WriteString(fmt.Sprintf("   Details: %s\n", step.Details))
		}
	}

	sb.WriteString("\n--- Detected Signals ---\n")
	for _, signal := range r.DetectedSignals {
		sb.WriteString(fmt.Sprintf("• %s @ %s (route: %s)\n", signal.Type, signal.Timestamp.Format("15:04:05"), signal.Route))
	}

	sb.WriteString("\n--- Qualified Signals ---\n")
	for _, signal := range r.QualifiedSignals {
		sb.WriteString(fmt.Sprintf("• %s (strength: %.2f) - %s\n", signal.Type, signal.Strength, signal.QualificationReason))
	}

	sb.WriteString("\n--- Score Breakdown ---\n")
	sb.WriteString(fmt.Sprintf("Signal Count Score: %d\n", r.ScoreBreakdown.SignalCountScore))
	sb.WriteString(fmt.Sprintf("Type Weight Score: %d\n", r.ScoreBreakdown.TypeWeightScore))
	sb.WriteString(fmt.Sprintf("Duration Score: %d\n", r.ScoreBreakdown.DurationScore))
	sb.WriteString(fmt.Sprintf("Error Bonus: %d\n", r.ScoreBreakdown.ErrorBonus))
	sb.WriteString(fmt.Sprintf("Decay Penalty: %d\n", r.ScoreBreakdown.DecayPenalty))
	sb.WriteString(fmt.Sprintf("Session Penalty: %d\n", r.ScoreBreakdown.SessionPenalty))
	sb.WriteString(fmt.Sprintf("Final Score: %d\n", r.ScoreBreakdown.FinalScore))
	sb.WriteString(fmt.Sprintf("Explanation: %s\n", r.ScoreBreakdown.Explanation))

	sb.WriteString("\n--- Confidence ---\n")
	sb.WriteString(fmt.Sprintf("Level: %s\n", r.ConfidenceDetails.Level))
	sb.WriteString(fmt.Sprintf("Multiple Signals: %v\n", r.ConfidenceDetails.HasMultipleSignals))
	sb.WriteString(fmt.Sprintf("System Feedback: %v\n", r.ConfidenceDetails.HasSystemFeedback))
	sb.WriteString(fmt.Sprintf("Clear Failure: %v\n", r.ConfidenceDetails.HasClearFailure))
	sb.WriteString(fmt.Sprintf("Factors: %s\n", strings.Join(r.ConfidenceDetails.Factors, ", ")))

	sb.WriteString("\n--- Final Decision ---\n")
	sb.WriteString(fmt.Sprintf("Should Emit: %v\n", r.FinalDecision.ShouldEmit))
	sb.WriteString(fmt.Sprintf("Reason: %s\n", r.FinalDecision.Reason))
	if r.FinalDecision.SuppressionReason != "" {
		sb.WriteString(fmt.Sprintf("Suppression Reason: %s\n", r.FinalDecision.SuppressionReason))
	}
	if r.FinalDecision.ShadowMode {
		sb.WriteString("Mode: SHADOW (detection only, no export)\n")
	}

	return sb.String()
}

// DetectionReasoningStore stores reasoning for debugging
type DetectionReasoningStore struct {
	reasonings []DetectionReasoning
	maxSize    int
	bySession  map[string][]int // sessionID -> indices
}

// NewDetectionReasoningStore creates a new reasoning store
func NewDetectionReasoningStore(maxSize int) *DetectionReasoningStore {
	return &DetectionReasoningStore{
		reasonings: make([]DetectionReasoning, 0, maxSize),
		maxSize:    maxSize,
		bySession:  make(map[string][]int),
	}
}

// Store stores a reasoning
func (s *DetectionReasoningStore) Store(reasoning DetectionReasoning) {
	idx := len(s.reasonings)
	s.reasonings = append(s.reasonings, reasoning)

	// Update session index
	s.bySession[reasoning.SessionID] = append(s.bySession[reasoning.SessionID], idx)

	// Trim if needed (simple approach - could be improved)
	if len(s.reasonings) > s.maxSize {
		s.reasonings = s.reasonings[len(s.reasonings)-s.maxSize:]
		// Rebuild session index
		s.bySession = make(map[string][]int)
		for i, r := range s.reasonings {
			s.bySession[r.SessionID] = append(s.bySession[r.SessionID], i)
		}
	}
}

// GetBySession returns reasonings for a session
func (s *DetectionReasoningStore) GetBySession(sessionID string) []DetectionReasoning {
	indices, ok := s.bySession[sessionID]
	if !ok {
		return nil
	}

	result := make([]DetectionReasoning, 0, len(indices))
	for _, idx := range indices {
		if idx < len(s.reasonings) {
			result = append(result, s.reasonings[idx])
		}
	}
	return result
}

// GetRecent returns recent reasonings
func (s *DetectionReasoningStore) GetRecent(limit int) []DetectionReasoning {
	if limit <= 0 || limit > len(s.reasonings) {
		limit = len(s.reasonings)
	}

	start := len(s.reasonings) - limit
	if start < 0 {
		start = 0
	}

	result := make([]DetectionReasoning, limit)
	copy(result, s.reasonings[start:])
	return result
}

// Global reasoning store
var globalReasoningStore = NewDetectionReasoningStore(1000)

// GetGlobalReasoningStore returns the global reasoning store
func GetGlobalReasoningStore() *DetectionReasoningStore {
	return globalReasoningStore
}

// StoreGlobalReasoning stores a reasoning to the global store
func StoreGlobalReasoning(reasoning DetectionReasoning) {
	globalReasoningStore.Store(reasoning)
}
