/**
 * Improved Rage Detection
 * 
 * Expert Team Improvements:
 * - Velocity-based scoring
 * - Context-aware thresholds
 * - Element importance weighting
 * - Advanced false positive filtering
 */

package signals

import (
	"log"
	"math"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// ImprovedRageDetector with advanced features
type ImprovedRageDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewImprovedRageDetector creates an improved rage detector
func NewImprovedRageDetector() *ImprovedRageDetector {
	return &ImprovedRageDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectRageImproved detects rage with improved algorithms
func (d *ImprovedRageDetector) DetectRageImproved(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Group events by target
	targetGroups := make(map[string][]ClassifiedEvent)
	for _, event := range classified {
		if event.Category == CategoryInteraction && event.Event.EventType == "click" {
			targetID := getTargetIDFromClassified(event)
			targetGroups[targetID] = append(targetGroups[targetID], event)
		}
	}

	// Detect with improved algorithms
	for targetID, events := range targetGroups {
		if len(events) < 3 {
			continue
		}

		// Calculate velocity pattern
		velocityScore := d.calculateVelocityScore(events)
		if velocityScore < 0.5 {
			continue // Not rapid enough
		}

		// Check for frustration pattern
		if d.isFrustrationPattern(events, session, velocityScore) {
			// Calculate improved strength score
			strengthScore := d.calculateImprovedStrengthScore(events, session, velocityScore)

			// Create candidate signal
			candidate := CandidateSignal{
				Type:      "rage",
				Timestamp: d.getFirstTimestamp(events),
				Route:     events[0].Event.Route,
				Details: map[string]interface{}{
					"targetID":         targetID,
					"interactionCount":  len(events),
					"velocityScore":    velocityScore,
					"strengthScore":    strengthScore,
					"strength":         d.determineStrengthLevel(strengthScore),
				},
			}

			// Advanced false alarm prevention
			if isFalse, reason := d.advancedFalseAlarmCheck(candidate, session, events); isFalse {
				log.Printf("[Improved Rage Detection] False alarm prevented: %s", reason)
				continue
			}

			candidates = append(candidates, candidate)
		}
	}

	return candidates
}

// calculateVelocityScore calculates velocity-based score (0.0-1.0)
func (d *ImprovedRageDetector) calculateVelocityScore(events []ClassifiedEvent) float64 {
	if len(events) < 2 {
		return 0.0
	}

	var intervals []time.Duration
	for i := 0; i < len(events)-1; i++ {
		t1, _ := time.Parse(time.RFC3339, events[i].Event.Timestamp)
		t2, _ := time.Parse(time.RFC3339, events[i+1].Event.Timestamp)
		intervals = append(intervals, t2.Sub(t1))
	}

	// Calculate average velocity
	avgInterval := time.Duration(0)
	for _, interval := range intervals {
		avgInterval += interval
	}
	avgInterval = avgInterval / time.Duration(len(intervals))

	// Calculate acceleration (decreasing intervals = frustration)
	acceleration := 0.0
	if len(intervals) >= 2 {
		for i := 0; i < len(intervals)-1; i++ {
			diff := float64(intervals[i] - intervals[i+1])
			acceleration += diff
		}
		acceleration = acceleration / float64(len(intervals)-1)
	}

	// Score based on velocity and acceleration
	velocityScore := 1.0 - (float64(avgInterval) / 1000.0) // Normalize to ms
	if velocityScore < 0 {
		velocityScore = 0
	}
	if velocityScore > 1.0 {
		velocityScore = 1.0
	}

	// Boost score if acceleration is negative (getting faster = frustration)
	if acceleration < 0 {
		velocityScore *= 1.2
		if velocityScore > 1.0 {
			velocityScore = 1.0
		}
	}

	return velocityScore
}

// isFrustrationPattern checks if pattern indicates frustration
func (d *ImprovedRageDetector) isFrustrationPattern(
	events []ClassifiedEvent,
	session types.Session,
	velocityScore float64,
) bool {
	// Check time window
	firstTime, _ := time.Parse(time.RFC3339, events[0].Event.Timestamp)
	lastTime, _ := time.Parse(time.RFC3339, events[len(events)-1].Event.Timestamp)
	timeWindow := lastTime.Sub(firstTime)

	// Adaptive time window based on velocity
	maxTimeWindow := 5 * time.Second
	if velocityScore > 0.8 {
		maxTimeWindow = 3 * time.Second // High velocity = shorter window
	}

	if timeWindow > maxTimeWindow {
		return false
	}

	// Check for success feedback
	hasSuccessFeedback := checkSuccessFeedbackBetweenClicks(
		events,
		session.Events,
		firstTime,
		lastTime,
	)

	return !hasSuccessFeedback
}

// calculateImprovedStrengthScore calculates improved strength score
func (d *ImprovedRageDetector) calculateImprovedStrengthScore(
	events []ClassifiedEvent,
	session types.Session,
	velocityScore float64,
) float64 {
	score := 0.0

	// Factor 1: Click count (30%)
	clickScore := math.Min(float64(len(events))/10.0, 1.0)
	score += clickScore * 0.3

	// Factor 2: Velocity (40%)
	score += velocityScore * 0.4

	// Factor 3: Element importance (20%)
	importanceScore := d.getElementImportanceScore(events[0].Event)
	score += importanceScore * 0.2

	// Factor 4: Context (10%)
	contextScore := d.getContextScore(session)
	score += contextScore * 0.1

	return math.Min(score, 1.0)
}

// getElementImportanceScore scores element importance
func (d *ImprovedRageDetector) getElementImportanceScore(event types.Event) float64 {
	target := event.Target

	// High importance elements
	highImportance := []string{"button", "a", "input[type='submit']"}
	for _, elem := range highImportance {
		if target.Type == elem {
			return 1.0
		}
	}

	// Medium importance
	mediumImportance := []string{"select", "input"}
	for _, elem := range mediumImportance {
		if target.Type == elem {
			return 0.7
		}
	}

	// Low importance (decorative)
	return 0.3
}

// getContextScore scores user context
func (d *ImprovedRageDetector) getContextScore(session types.Session) float64 {
	// Check for error context
	for _, event := range session.Events {
		if event.EventType == "error" {
			return 1.0 // High frustration in error context
		}
	}

	// Check for form context
	for _, event := range session.Events {
		if event.EventType == "form_submit" {
			return 0.8 // Medium-high frustration in form context
		}
	}

	return 0.5 // Default
}

// determineStrengthLevel determines strength level from score
func (d *ImprovedRageDetector) determineStrengthLevel(score float64) string {
	if score >= 0.8 {
		return "high"
	}
	if score >= 0.6 {
		return "medium"
	}
	return "low"
}

// advancedFalseAlarmCheck performs advanced false alarm checking
func (d *ImprovedRageDetector) advancedFalseAlarmCheck(
	candidate CandidateSignal,
	session types.Session,
	events []ClassifiedEvent,
) (bool, string) {
	// Check for loading states
	if d.isLoadingState(events, session) {
		return true, "loading state detected"
	}

	// Check for disabled elements
	if d.isDisabledElement(events) {
		return true, "disabled element detected"
	}

	// Check for accessibility tools
	if d.isAccessibilityTool(session) {
		return true, "accessibility tool detected"
	}

	// Check for gaming context
	if d.isGamingContext(session) {
		return true, "gaming context detected"
	}

	return false, ""
}

// Helper methods
func (d *ImprovedRageDetector) isLoadingState(events []ClassifiedEvent, session types.Session) bool {
	for _, event := range events {
		if metadata := event.Event.Metadata; metadata != nil {
			if loading, ok := metadata["loading"].(bool); ok && loading {
				return true
			}
		}
	}
	return false
}

func (d *ImprovedRageDetector) isDisabledElement(events []ClassifiedEvent) bool {
	for _, event := range events {
		if metadata := event.Event.Metadata; metadata != nil {
			if disabled, ok := metadata["disabled"].(bool); ok && disabled {
				return true
			}
		}
	}
	return false
}

func (d *ImprovedRageDetector) isAccessibilityTool(session types.Session) bool {
	for _, event := range session.Events {
		if metadata := event.Metadata; metadata != nil {
			if ua, ok := metadata["userAgent"].(string); ok {
				accessibilityTools := []string{"screen reader", "nvda", "jaws", "voiceover"}
				for _, tool := range accessibilityTools {
					if contains(ua, tool) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (d *ImprovedRageDetector) isGamingContext(session types.Session) bool {
	for _, event := range session.Events {
		route := event.Route
		gamingRoutes := []string{"/game", "/play", "/gaming"}
		for _, gr := range gamingRoutes {
			if contains(route, gr) {
				return true
			}
		}
	}
	return false
}

func (d *ImprovedRageDetector) getFirstTimestamp(events []ClassifiedEvent) int64 {
	if len(events) == 0 {
		return 0
	}
	t, _ := time.Parse(time.RFC3339, events[0].Event.Timestamp)
	return t.Unix()
}
