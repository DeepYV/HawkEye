/**
 * Rage Bait Detection
 * 
 * Author: Enhanced Detection Team
 * Responsibility: Detect intentionally frustrating content (rage bait, dark patterns)
 * 
 * Rage Bait Indicators:
 * - Clicks on non-interactive elements that look clickable
 * - Multiple rapid clicks on misleading UI elements
 * - Patterns matching known dark pattern signatures
 * - Content designed to frustrate users
 */

package signals

import (
	"log"
	"strings"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Rage bait detection thresholds
	rageBaitMinClicks            = 3
	rageBaitTimeWindow           = 5 * time.Second
	rageBaitMaxTimeBetweenClicks = 1000 * time.Millisecond
	
	// Dark pattern indicators
	minDarkPatternScore = 0.6
)

// RageBaitDetector detects rage bait and dark pattern signals
type RageBaitDetector struct {
	falseAlarmPreventer *FalseAlarmPreventer
}

// NewRageBaitDetector creates a new rage bait detector
func NewRageBaitDetector() *RageBaitDetector {
	return &RageBaitDetector{
		falseAlarmPreventer: NewFalseAlarmPreventer(),
	}
}

// DetectRageBait detects rage bait patterns
func (d *RageBaitDetector) DetectRageBait(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)
	
	// Group events by target
	targetGroups := make(map[string][]ClassifiedEvent)
	for _, event := range classified {
		if event.Category == CategoryInteraction && event.Event.EventType == "click" {
			targetID := getTargetIDFromClassified(event)
			targetGroups[targetID] = append(targetGroups[targetID], event)
		}
	}
	
	// Detect rage bait patterns
	for targetID, events := range targetGroups {
		if len(events) < rageBaitMinClicks {
			continue
		}
		
		// Check for rapid clicks on potentially misleading elements
		for i := 0; i <= len(events)-rageBaitMinClicks; i++ {
			firstEvent := events[i]
			lastEvent := events[i+rageBaitMinClicks-1]
			
			// Parse timestamps
			t1, err1 := time.Parse(time.RFC3339, firstEvent.Event.Timestamp)
			t2, err2 := time.Parse(time.RFC3339, lastEvent.Event.Timestamp)
			if err1 != nil || err2 != nil {
				continue
			}
			
			timeWindow := t2.Sub(t1)
			if timeWindow > rageBaitTimeWindow {
				continue
			}
			
			// Check if this looks like rage bait
			isRageBait, darkPatternScore := d.analyzeRageBaitPattern(
				events[i:i+rageBaitMinClicks],
				session,
				firstEvent.Event,
			)
			
			if !isRageBait {
				continue
			}
			
			// Check for success feedback (less strict for rage bait)
			hasSuccessFeedback := checkSuccessFeedbackBetweenClicks(
				events[i:i+rageBaitMinClicks],
				session.Events,
				t1,
				t2,
			)
			
			// For rage bait, even if there's success feedback, it might still be rage bait
			// if the dark pattern score is high
			if hasSuccessFeedback && darkPatternScore < minDarkPatternScore {
				continue
			}
			
			// Create candidate signal
			candidate := CandidateSignal{
				Type:      "rage_bait",
				Timestamp: t1.Unix(),
				Route:     firstEvent.Event.Route,
				Details: map[string]interface{}{
					"targetID":         targetID,
					"interactionCount":  rageBaitMinClicks,
					"timeWindow":        timeWindow.String(),
					"darkPatternScore": darkPatternScore,
					"isDarkPattern":    darkPatternScore >= minDarkPatternScore,
				},
			}
			
			// Check for false alarms (less strict for rage bait)
			if isFalse, reason := d.checkRageBaitFalseAlarms(candidate, session, convertToEvents(classified)); isFalse {
				log.Printf("[Rage Bait Detection] False alarm prevented: %s", reason)
				continue
			}
			
			candidates = append(candidates, candidate)
			break // Only one rage bait signal per target
		}
	}
	
	return candidates
}

// analyzeRageBaitPattern analyzes if a pattern indicates rage bait
func (d *RageBaitDetector) analyzeRageBaitPattern(
	clickEvents []ClassifiedEvent,
	session types.Session,
	firstEvent types.Event,
) (bool, float64) {
	
	darkPatternScore := 0.0
	indicators := 0
	
	// Indicator 1: Non-interactive element that looks clickable
	if d.isNonInteractiveButClickable(firstEvent) {
		darkPatternScore += 0.3
		indicators++
	}
	
	// Indicator 2: Misleading UI element (e.g., fake button, deceptive link)
	if d.isMisleadingUIElement(firstEvent) {
		darkPatternScore += 0.3
		indicators++
	}
	
	// Indicator 3: Known dark pattern signatures
	if d.matchesDarkPatternSignature(firstEvent, session) {
		darkPatternScore += 0.4
		indicators++
	}
	
	// Indicator 4: Clickbait content (multiple clicks on content elements)
	if d.isClickbaitContent(firstEvent, clickEvents) {
		darkPatternScore += 0.2
		indicators++
	}
	
	// Indicator 5: Element that should be interactive but isn't responding
	if d.isUnresponsiveInteractiveElement(firstEvent, clickEvents, session) {
		darkPatternScore += 0.3
		indicators++
	}
	
	// Require at least 2 indicators or high dark pattern score
	isRageBait := indicators >= 2 || darkPatternScore >= minDarkPatternScore
	
	return isRageBait, darkPatternScore
}

// isNonInteractiveButClickable checks if element is non-interactive but looks clickable
func (d *RageBaitDetector) isNonInteractiveButClickable(event types.Event) bool {
	target := event.Target
	
	// Check element type
	nonInteractiveTypes := []string{"div", "span", "p", "h1", "h2", "h3", "h4", "h5", "h6", "img"}
	for _, t := range nonInteractiveTypes {
		if strings.EqualFold(target.Type, t) {
			// Check if it has clickable styling indicators
			if metadata := event.Metadata; metadata != nil {
				if cursor, ok := metadata["cursor"].(string); ok {
					if cursor == "pointer" {
						return true // Non-interactive but styled as clickable
					}
				}
				if hasClickHandler, ok := metadata["hasClickHandler"].(bool); ok && !hasClickHandler {
					if cursor, ok := metadata["cursor"].(string); ok && cursor == "pointer" {
						return true
					}
				}
			}
		}
	}
	
	return false
}

// isMisleadingUIElement checks if element is misleading (fake button, etc.)
func (d *RageBaitDetector) isMisleadingUIElement(event types.Event) bool {
	target := event.Target
	
	// Check for fake button indicators
	if metadata := event.Metadata; metadata != nil {
		// Fake button: looks like button but isn't
		if looksLikeButton, ok := metadata["looksLikeButton"].(bool); ok && looksLikeButton {
			if isActuallyButton, ok := metadata["isButton"].(bool); ok && !isActuallyButton {
				return true
			}
		}
		
		// Deceptive link: link that doesn't do what it says
		if isLink, ok := metadata["isLink"].(bool); ok && isLink {
			if isDeceptive, ok := metadata["isDeceptive"].(bool); ok && isDeceptive {
				return true
			}
		}
	}
	
	// Check class names for dark pattern indicators
	if target.Selector != "" {
		selectorLower := strings.ToLower(target.Selector)
		darkPatternClasses := []string{
			"fake-button", "fake-link", "deceptive",
			"misleading", "trick", "bait",
		}
		for _, pattern := range darkPatternClasses {
			if strings.Contains(selectorLower, pattern) {
				return true
			}
		}
	}
	
	return false
}

// matchesDarkPatternSignature checks if element matches known dark pattern signatures
func (d *RageBaitDetector) matchesDarkPatternSignature(event types.Event, session types.Session) bool {
	// Check for common dark patterns
	
	// Pattern 1: Roach motel (easy to get in, hard to get out)
	if d.isRoachMotelPattern(event, session) {
		return true
	}
	
	// Pattern 2: Misdirection (distracting elements)
	if d.isMisdirectionPattern(event, session) {
		return true
	}
	
	// Pattern 3: Forced continuity (hard to cancel)
	if d.isForcedContinuityPattern(event, session) {
		return true
	}
	
	// Pattern 4: Sneak into basket (items added without clear consent)
	if d.isSneakIntoBasketPattern(event, session) {
		return true
	}
	
	return false
}

// isRoachMotelPattern checks for roach motel dark pattern
func (d *RageBaitDetector) isRoachMotelPattern(event types.Event, session types.Session) bool {
	// Check if user is trying to exit/cancel but having difficulty
	route := strings.ToLower(event.Route)
	exitRoutes := []string{"/cancel", "/exit", "/close", "/back"}
	for _, exitRoute := range exitRoutes {
		if strings.Contains(route, exitRoute) {
			// Check for multiple attempts
			attemptCount := 0
			for _, e := range session.Events {
				if e.Route == event.Route && e.EventType == "click" {
					attemptCount++
				}
			}
			return attemptCount >= 3
		}
	}
	return false
}

// isMisdirectionPattern checks for misdirection dark pattern
func (d *RageBaitDetector) isMisdirectionPattern(event types.Event, session types.Session) bool {
	// Check for clicks on distracting elements
	if metadata := event.Metadata; metadata != nil {
		if isDistracting, ok := metadata["isDistracting"].(bool); ok && isDistracting {
			return true
		}
	}
	return false
}

// isForcedContinuityPattern checks for forced continuity dark pattern
func (d *RageBaitDetector) isForcedContinuityPattern(event types.Event, session types.Session) bool {
	// Check for subscription/cancel flows with difficulty
	route := strings.ToLower(event.Route)
	if strings.Contains(route, "cancel") || strings.Contains(route, "unsubscribe") {
		// Multiple clicks on cancel suggests forced continuity
		clickCount := 0
		for _, e := range session.Events {
			if strings.Contains(strings.ToLower(e.Route), "cancel") && e.EventType == "click" {
				clickCount++
			}
		}
		return clickCount >= 3
	}
	return false
}

// isSneakIntoBasketPattern checks for sneak into basket dark pattern
func (d *RageBaitDetector) isSneakIntoBasketPattern(event types.Event, session types.Session) bool {
	// Check for items added without clear user action
	route := strings.ToLower(event.Route)
	if strings.Contains(route, "cart") || strings.Contains(route, "basket") {
		// Check if items were added without corresponding add-to-cart clicks
		addToCartClicks := 0
		cartItems := 0
		
		for _, e := range session.Events {
			if e.EventType == "click" {
				if strings.Contains(strings.ToLower(e.Route), "add") || 
				   strings.Contains(strings.ToLower(e.Route), "cart") {
					addToCartClicks++
				}
			}
			if e.EventType == "navigation" && strings.Contains(strings.ToLower(e.Route), "cart") {
				if metadata := e.Metadata; metadata != nil {
					if items, ok := metadata["itemCount"].(float64); ok {
						cartItems = int(items)
					}
				}
			}
		}
		
		// If more items than add-to-cart clicks, might be sneak into basket
		return cartItems > addToCartClicks && cartItems > 0
	}
	return false
}

// isClickbaitContent checks if clicks are on clickbait content
func (d *RageBaitDetector) isClickbaitContent(event types.Event, clickEvents []ClassifiedEvent) bool {
	// Check if clicking on content elements (not interactive elements)
	target := event.Target
	contentTypes := []string{"article", "div", "span", "p", "h1", "h2", "h3"}
	
	for _, ct := range contentTypes {
		if strings.EqualFold(target.Type, ct) {
			// Multiple clicks on content suggests clickbait
			return len(clickEvents) >= 3
		}
	}
	
	return false
}

// isUnresponsiveInteractiveElement checks if element should be interactive but isn't responding
func (d *RageBaitDetector) isUnresponsiveInteractiveElement(
	event types.Event,
	clickEvents []ClassifiedEvent,
	session types.Session,
) bool {
	target := event.Target
	
	// Check if it's an interactive element type
	interactiveTypes := []string{"button", "a", "input", "select"}
	isInteractive := false
	for _, it := range interactiveTypes {
		if strings.EqualFold(target.Type, it) {
			isInteractive = true
			break
		}
	}
	
	if !isInteractive {
		return false
	}
	
	// Check if there are multiple clicks without response
	if len(clickEvents) >= 3 {
		// Check for lack of response (no navigation, no network activity)
		hasResponse := false
		firstTime, _ := time.Parse(time.RFC3339, clickEvents[0].Event.Timestamp)
		lastTime, _ := time.Parse(time.RFC3339, clickEvents[len(clickEvents)-1].Event.Timestamp)
		
		for _, e := range session.Events {
			eventTime, err := time.Parse(time.RFC3339, e.Timestamp)
			if err != nil {
				continue
			}
			
			if eventTime.After(firstTime) && eventTime.Before(lastTime.Add(2*time.Second)) {
				if e.EventType == "navigation" || e.EventType == "network" {
					hasResponse = true
					break
				}
			}
		}
		
		return !hasResponse // Unresponsive if no response after multiple clicks
	}
	
	return false
}

// checkRageBaitFalseAlarms checks for false alarms specific to rage bait
func (d *RageBaitDetector) checkRageBaitFalseAlarms(
	signal CandidateSignal,
	session types.Session,
	events []types.Event,
) (bool, string) {
	// Only check for critical false alarms (bots, etc.)
	botPattern := &BotPattern{}
	if botPattern.Matches(signal, events) {
		return true, "bot/crawler pattern"
	}
	
	return false, ""
}
