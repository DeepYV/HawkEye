/**
 * Enhanced Multi-Tier Rage Detection
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Detect frustration at multiple intensity levels with improved accuracy
 * 
 * Features:
 * - Multi-tier detection (High/Medium/Low strength)
 * - Signal strength scoring
 * - Context-aware thresholds
 * - Better detection of subtle frustration patterns
 */

package signals

import (
	"log"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// RageStrengthLevel represents the strength level of a rage signal
type RageStrengthLevel string

const (
	RageStrengthHigh   RageStrengthLevel = "high"
	RageStrengthMedium RageStrengthLevel = "medium"
	RageStrengthLow    RageStrengthLevel = "low"
)

// Multi-tier thresholds
const (
	// High strength: Very intense frustration
	rageHighMinClicks              = 5
	rageHighTimeWindow             = 2 * time.Second
	rageHighMaxTimeBetweenClicks   = 300 * time.Millisecond
	
	// Medium strength: Standard frustration (current refined thresholds)
	rageMediumMinClicks            = 4
	rageMediumTimeWindow           = 3 * time.Second
	rageMediumMaxTimeBetweenClicks = 500 * time.Millisecond
	
	// Low strength: Subtle frustration patterns
	rageLowMinClicks               = 3
	rageLowTimeWindow              = 5 * time.Second
	rageLowMaxTimeBetweenClicks    = 800 * time.Millisecond
)

// EnhancedRageDetector detects rage signals at multiple strength levels
type EnhancedRageDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewEnhancedRageDetector creates a new enhanced rage detector
func NewEnhancedRageDetector() *EnhancedRageDetector {
	return &EnhancedRageDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectRageMultiTier detects rage signals at all strength levels
func (d *EnhancedRageDetector) DetectRageMultiTier(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)
	
	// Group events by target
	targetGroups := make(map[string][]ClassifiedEvent)
	for _, event := range classified {
		if event.Category == CategoryInteraction && event.Event.EventType == "click" {
			targetID := getTargetIDFromClassified(event)
			targetGroups[targetID] = append(targetGroups[targetID], event)
		}
	}
	
	// Detect at each strength level
	for targetID, events := range targetGroups {
		// Try high strength first (most specific)
		if candidate := d.detectAtStrength(events, targetID, RageStrengthHigh, session, classified); candidate != nil {
			candidates = append(candidates, *candidate)
			continue // Only one signal per target
		}
		
		// Try medium strength
		if candidate := d.detectAtStrength(events, targetID, RageStrengthMedium, session, classified); candidate != nil {
			candidates = append(candidates, *candidate)
			continue
		}
		
		// Try low strength
		if candidate := d.detectAtStrength(events, targetID, RageStrengthLow, session, classified); candidate != nil {
			candidates = append(candidates, *candidate)
			continue
		}
	}
	
	return candidates
}

// detectAtStrength detects rage at a specific strength level
func (d *EnhancedRageDetector) detectAtStrength(
	events []ClassifiedEvent,
	targetID string,
	strength RageStrengthLevel,
	session types.Session,
	classified []ClassifiedEvent,
) *CandidateSignal {
	
	// Get thresholds for strength level
	minClicks, timeWindow, maxTimeBetween := d.getThresholds(strength)
	
	if len(events) < minClicks {
		return nil
	}
	
	// Check for rapid interactions within time window
	for i := 0; i <= len(events)-minClicks; i++ {
		firstEvent := events[i]
		lastEvent := events[i+minClicks-1]
		
		// Parse timestamps
		t1, err1 := time.Parse(time.RFC3339, firstEvent.Event.Timestamp)
		t2, err2 := time.Parse(time.RFC3339, lastEvent.Event.Timestamp)
		if err1 != nil || err2 != nil {
			continue
		}
		
		timeWindowActual := t2.Sub(t1)
		if timeWindowActual > timeWindow {
			continue // Too slow for this strength level
		}
		
		// Check time between consecutive clicks
		allRapid := true
		for j := i; j < i+minClicks-1; j++ {
			tj, _ := time.Parse(time.RFC3339, events[j].Event.Timestamp)
			tj1, _ := time.Parse(time.RFC3339, events[j+1].Event.Timestamp)
			if tj1.Sub(tj) > maxTimeBetween {
				allRapid = false
				break
			}
		}
		
		if !allRapid {
			continue // Not all clicks are rapid enough
		}
		
		// Check for success feedback between clicks
		hasSuccessFeedback := checkSuccessFeedbackBetweenClicks(events[i:i+minClicks], session.Events, t1, t2)
		if hasSuccessFeedback {
			continue // Success occurred, not frustration
		}
		
		// Calculate signal strength score
		strengthScore := d.calculateStrengthScore(timeWindowActual, minClicks, maxTimeBetween, events[i:i+minClicks])
		
		// Create candidate signal
		candidate := &CandidateSignal{
			Type:      "rage",
			Timestamp: t1.Unix(),
			Route:     firstEvent.Event.Route,
			Details: map[string]interface{}{
				"targetID":         targetID,
				"interactionCount":  minClicks,
				"timeWindow":        timeWindowActual.String(),
				"strength":          string(strength),
				"strengthScore":    strengthScore,
			},
		}
		
		// Check for false alarms (less strict for high-strength signals)
		if strength != RageStrengthHigh {
			if isFalse, reason := d.falseAlarmPreventer.IsFalseAlarm(*candidate, session, convertToEvents(classified)); isFalse {
				log.Printf("[Enhanced Rage Detection] False alarm prevented (%s): %s", strength, reason)
				continue
			}
		} else {
			// For high-strength signals, only check critical false alarms
			if isFalse, reason := d.checkCriticalFalseAlarms(*candidate, session, convertToEvents(classified)); isFalse {
				log.Printf("[Enhanced Rage Detection] Critical false alarm prevented (%s): %s", strength, reason)
				continue
			}
		}
		
		return candidate
	}
	
	return nil
}

// getThresholds returns thresholds for a strength level
func (d *EnhancedRageDetector) getThresholds(strength RageStrengthLevel) (minClicks int, timeWindow time.Duration, maxTimeBetween time.Duration) {
	switch strength {
	case RageStrengthHigh:
		return rageHighMinClicks, rageHighTimeWindow, rageHighMaxTimeBetweenClicks
	case RageStrengthMedium:
		return rageMediumMinClicks, rageMediumTimeWindow, rageMediumMaxTimeBetweenClicks
	case RageStrengthLow:
		return rageLowMinClicks, rageLowTimeWindow, rageLowMaxTimeBetweenClicks
	default:
		return rageMediumMinClicks, rageMediumTimeWindow, rageMediumMaxTimeBetweenClicks
	}
}

// calculateStrengthScore calculates a signal strength score (0.0-1.0)
func (d *EnhancedRageDetector) calculateStrengthScore(
	timeWindow time.Duration,
	clickCount int,
	maxTimeBetween time.Duration,
	clickEvents []ClassifiedEvent,
) float64 {
	score := 0.0
	
	// Factor 1: Click intensity (more clicks = higher score)
	clickScore := float64(clickCount) / 10.0 // Normalize to 0-1
	if clickScore > 1.0 {
		clickScore = 1.0
	}
	score += clickScore * 0.4
	
	// Factor 2: Time compression (faster = higher score)
	avgTimeBetween := timeWindow / time.Duration(len(clickEvents)-1)
	timeScore := 1.0 - (float64(avgTimeBetween) / float64(maxTimeBetween))
	if timeScore < 0 {
		timeScore = 0
	}
	if timeScore > 1.0 {
		timeScore = 1.0
	}
	score += timeScore * 0.4
	
	// Factor 3: Consistency (more consistent = higher score)
	consistencyScore := d.calculateConsistency(clickEvents)
	score += consistencyScore * 0.2
	
	return score
}

// calculateConsistency calculates consistency of click timing
func (d *EnhancedRageDetector) calculateConsistency(clickEvents []ClassifiedEvent) float64 {
	if len(clickEvents) < 2 {
		return 0.0
	}
	
	var intervals []time.Duration
	for i := 0; i < len(clickEvents)-1; i++ {
		t1, _ := time.Parse(time.RFC3339, clickEvents[i].Event.Timestamp)
		t2, _ := time.Parse(time.RFC3339, clickEvents[i+1].Event.Timestamp)
		intervals = append(intervals, t2.Sub(t1))
	}
	
	if len(intervals) == 0 {
		return 0.0
	}
	
	// Calculate coefficient of variation (lower = more consistent)
	avg := time.Duration(0)
	for _, interval := range intervals {
		avg += interval
	}
	avg = avg / time.Duration(len(intervals))
	
	variance := 0.0
	for _, interval := range intervals {
		diff := float64(interval - avg)
		variance += diff * diff
	}
	variance = variance / float64(len(intervals))
	stdDev := variance
	
	if avg == 0 {
		return 1.0
	}
	
	coefficientOfVariation := stdDev / float64(avg)
	
	// Convert to consistency score (lower CV = higher consistency)
	consistency := 1.0 - (coefficientOfVariation / 2.0) // Normalize
	if consistency < 0 {
		consistency = 0
	}
	if consistency > 1.0 {
		consistency = 1.0
	}
	
	return consistency
}

// checkCriticalFalseAlarms checks only for critical false alarm patterns
func (d *EnhancedRageDetector) checkCriticalFalseAlarms(
	signal CandidateSignal,
	session types.Session,
	events []types.Event,
) (bool, string) {
	// Only check for bot/crawler patterns and accessibility tools
	// Skip other checks for high-strength signals
	
	// Check bot patterns
	botPattern := &BotPattern{}
	if botPattern.Matches(signal, events) {
		return true, "bot/crawler pattern"
	}
	
	// Check accessibility (but be less strict)
	accessibilityPattern := &AccessibilityPattern{}
	if accessibilityPattern.Matches(signal, events) {
		// For high-strength signals, only filter if it's clearly accessibility tool
		// Don't filter for general accessibility indicators
		return false, ""
	}
	
	return false, ""
}
