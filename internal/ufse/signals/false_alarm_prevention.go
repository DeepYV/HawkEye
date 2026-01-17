/**
 * False Alarm Prevention
 * 
 * Author: Principal Engineer + Team Alpha/Beta
 * Responsibility: Prevent false alarms through comprehensive pattern recognition
 * 
 * Goal: Zero false alarms (< 0.1% false positive rate)
 */

package signals

import (
	"fmt"
	"strings"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// FalseAlarmPreventer prevents false alarms through pattern recognition
type FalseAlarmPreventer struct {
	whitelistPatterns []PatternMatcher
	blacklistPatterns []PatternMatcher
	contextCheckers   []ContextChecker
}

// PatternMatcher matches patterns to identify legitimate vs frustration behavior
type PatternMatcher interface {
	Matches(signal CandidateSignal, events []types.Event) bool
	IsWhitelist() bool
	Description() string
}

// ContextChecker checks context to prevent false alarms
type ContextChecker interface {
	Check(signal CandidateSignal, session types.Session) (bool, string) // (isFalseAlarm, reason)
}

// NewFalseAlarmPreventer creates a new false alarm preventer
func NewFalseAlarmPreventer() *FalseAlarmPreventer {
	return &FalseAlarmPreventer{
		whitelistPatterns: []PatternMatcher{
			&DoubleClickPattern{},
			&AccessibilityPattern{},
			&GamingPattern{},
			&SearchPattern{},
			&ComparisonShoppingPattern{},
		},
		blacklistPatterns: []PatternMatcher{
			&BotPattern{},
			&CrawlerPattern{},
		},
		contextCheckers: []ContextChecker{
			&LoadingStateChecker{},
			&DisabledButtonChecker{},
			&FormValidationChecker{},
			&NetworkRetryChecker{},
			&LegitimateNavigationChecker{},
		},
	}
}

// IsFalseAlarm checks if a signal is a false alarm
func (f *FalseAlarmPreventer) IsFalseAlarm(signal CandidateSignal, session types.Session, events []types.Event) (bool, string) {
	// Check whitelist patterns (legitimate behavior)
	for _, pattern := range f.whitelistPatterns {
		if pattern.Matches(signal, events) {
			return true, fmt.Sprintf("whitelist pattern: %s", pattern.Description())
		}
	}

	// Check blacklist patterns (known false patterns)
	for _, pattern := range f.blacklistPatterns {
		if pattern.Matches(signal, events) {
			return true, fmt.Sprintf("blacklist pattern: %s", pattern.Description())
		}
	}

	// Check context
	for _, checker := range f.contextCheckers {
		if isFalse, reason := checker.Check(signal, session); isFalse {
			return true, reason
		}
	}

	return false, ""
}

// DoubleClickPattern matches legitimate double-click behavior
type DoubleClickPattern struct{}

func (p *DoubleClickPattern) Matches(signal CandidateSignal, events []types.Event) bool {
	if signal.Type != "rage" {
		return false
	}

	// Get target ID from signal details
	targetID, ok := signal.Details["targetID"].(string)
	if !ok {
		return false
	}

	// Check for double-click pattern (2 clicks within 500ms)
	clicks := getClicksForTarget(events, targetID)
	if len(clicks) >= 2 {
		// Parse timestamps
		t1, err1 := time.Parse(time.RFC3339, clicks[0].Timestamp)
		t2, err2 := time.Parse(time.RFC3339, clicks[1].Timestamp)
		if err1 == nil && err2 == nil {
			timeDiff := t2.Sub(t1)
			if timeDiff > 0 && timeDiff < 500*time.Millisecond {
				return true // Legitimate double-click
			}
		}
	}

	return false
}

func (p *DoubleClickPattern) IsWhitelist() bool { return true }
func (p *DoubleClickPattern) Description() string { return "double-click pattern" }

// AccessibilityPattern matches accessibility tool usage
type AccessibilityPattern struct{}

func (p *AccessibilityPattern) Matches(signal CandidateSignal, events []types.Event) bool {
	// Check for accessibility indicators in events
	for _, event := range events {
		if metadata := event.Metadata; metadata != nil {
			if ua, ok := metadata["userAgent"].(string); ok {
				uaLower := strings.ToLower(ua)
				accessibilityIndicators := []string{
					"screen reader", "nvda", "jaws", "voiceover",
					"talkback", "orca", "narrator",
				}
				for _, indicator := range accessibilityIndicators {
					if strings.Contains(uaLower, indicator) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (p *AccessibilityPattern) IsWhitelist() bool { return true }
func (p *AccessibilityPattern) Description() string { return "accessibility tool usage" }

// GamingPattern matches gaming application rapid clicks
type GamingPattern struct{}

func (p *GamingPattern) Matches(signal CandidateSignal, events []types.Event) bool {
	if signal.Type != "rage" {
		return false
	}

	// Check route for gaming indicators
	for _, event := range events {
		route := strings.ToLower(event.Route)
		gamingIndicators := []string{"/game", "/play", "/gaming"}
		for _, indicator := range gamingIndicators {
			if strings.Contains(route, indicator) {
				// In gaming contexts, rapid clicks are expected
				return true
			}
		}
	}

	return false
}

func (p *GamingPattern) IsWhitelist() bool { return true }
func (p *GamingPattern) Description() string { return "gaming application pattern" }

// SearchPattern matches search functionality usage
type SearchPattern struct{}

func (p *SearchPattern) Matches(signal CandidateSignal, events []types.Event) bool {
	// Check for search-related navigation
	for _, event := range events {
		if event.EventType == "navigation" {
			route := strings.ToLower(event.Route)
			if strings.Contains(route, "/search") || strings.Contains(route, "/filter") {
				return true
			}
		}
	}
	return false
}

func (p *SearchPattern) IsWhitelist() bool { return true }
func (p *SearchPattern) Description() string { return "search functionality pattern" }

// ComparisonShoppingPattern matches comparison shopping behavior
type ComparisonShoppingPattern struct{}

func (p *ComparisonShoppingPattern) Matches(signal CandidateSignal, events []types.Event) bool {
	// Check for back-and-forth navigation in shopping contexts
	if signal.Type != "confusion" {
		return false
	}

	routeCounts := make(map[string]int)
	for _, event := range events {
		if event.EventType == "navigation" {
			routeCounts[event.Route]++
		}
	}

	// Multiple visits to same routes in shopping context
	shoppingRoutes := []string{"/product", "/compare", "/cart", "/checkout"}
	shoppingContext := false
	for _, route := range shoppingRoutes {
		if routeCounts[route] > 1 {
			shoppingContext = true
			break
		}
	}

	return shoppingContext
}

func (p *ComparisonShoppingPattern) IsWhitelist() bool { return true }
func (p *ComparisonShoppingPattern) Description() string { return "comparison shopping pattern" }

// BotPattern matches bot/crawler behavior
type BotPattern struct{}

func (p *BotPattern) Matches(signal CandidateSignal, events []types.Event) bool {
	for _, event := range events {
		if metadata := event.Metadata; metadata != nil {
			if ua, ok := metadata["userAgent"].(string); ok {
				uaLower := strings.ToLower(ua)
				botIndicators := []string{
					"bot", "crawler", "spider", "scraper",
					"googlebot", "bingbot", "slurp",
				}
				for _, indicator := range botIndicators {
					if strings.Contains(uaLower, indicator) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (p *BotPattern) IsWhitelist() bool { return false }
func (p *BotPattern) Description() string { return "bot/crawler pattern" }

// CrawlerPattern matches web crawler behavior
type CrawlerPattern struct{}

func (p *CrawlerPattern) Matches(signal CandidateSignal, events []types.Event) bool {
	// Similar to BotPattern but more specific
	return (&BotPattern{}).Matches(signal, events)
}

func (p *CrawlerPattern) IsWhitelist() bool { return false }
func (p *CrawlerPattern) Description() string { return "crawler pattern" }

// LoadingStateChecker checks if clicks are on loading elements
type LoadingStateChecker struct{}

func (c *LoadingStateChecker) Check(signal CandidateSignal, session types.Session) (bool, string) {
	if signal.Type != "rage" {
		return false, ""
	}

	// Get target ID from signal details
	targetID, ok := signal.Details["targetID"].(string)
	if !ok {
		return false, ""
	}

	// Check if target has loading state
	for _, event := range session.Events {
		if event.Target.ID == targetID {
			if metadata := event.Metadata; metadata != nil {
				if loading, ok := metadata["loading"].(bool); ok && loading {
					return true, "click on loading element"
				}
			}
		}
	}

	return false, ""
}

// DisabledButtonChecker checks if clicks are on disabled buttons
type DisabledButtonChecker struct{}

func (c *DisabledButtonChecker) Check(signal CandidateSignal, session types.Session) (bool, string) {
	if signal.Type != "rage" {
		return false, ""
	}

	// Get target ID from signal details
	targetID, ok := signal.Details["targetID"].(string)
	if !ok {
		return false, ""
	}

	// Check if target is disabled
	for _, event := range session.Events {
		if event.Target.ID == targetID {
			if metadata := event.Metadata; metadata != nil {
				if disabled, ok := metadata["disabled"].(bool); ok && disabled {
					return true, "click on disabled button"
				}
			}
		}
	}

	return false, ""
}

// FormValidationChecker checks if errors are from form validation
type FormValidationChecker struct{}

func (c *FormValidationChecker) Check(signal CandidateSignal, session types.Session) (bool, string) {
	if signal.Type != "blocked" {
		return false, ""
	}

	// Check if error is validation error (user error, not system error)
	for _, event := range session.Events {
		if event.EventType == "error" {
			if metadata := event.Metadata; metadata != nil {
				if errorType, ok := metadata["errorType"].(string); ok {
					if errorType == "validation" || errorType == "client_error" {
						return true, "form validation error (user error)"
					}
				}
			}
		}
	}

	return false, ""
}

// NetworkRetryChecker checks if retries are from network issues
type NetworkRetryChecker struct{}

func (c *NetworkRetryChecker) Check(signal CandidateSignal, session types.Session) (bool, string) {
	if signal.Type != "blocked" {
		return false, ""
	}

	// Check if there are network errors followed by successful retries
	hasNetworkError := false
	hasSuccessAfterError := false

	for i, event := range session.Events {
		if event.EventType == "error" {
			if metadata := event.Metadata; metadata != nil {
				if errorType, ok := metadata["errorType"].(string); ok {
					if errorType == "network" || errorType == "timeout" {
						hasNetworkError = true
						// Check if there's success after this error
						for j := i + 1; j < len(session.Events) && j < i+10; j++ {
							if session.Events[j].EventType == "navigation" {
								// Success navigation after error
								hasSuccessAfterError = true
								break
							}
						}
					}
				}
			}
		}
	}

	if hasNetworkError && hasSuccessAfterError {
		return true, "network retry succeeded (not frustration)"
	}

	return false, ""
}

// LegitimateNavigationChecker checks if navigation is legitimate
type LegitimateNavigationChecker struct{}

func (c *LegitimateNavigationChecker) Check(signal CandidateSignal, session types.Session) (bool, string) {
	if signal.Type != "abandonment" {
		return false, ""
	}

	// Check if user navigated to external link (expected behavior)
	for _, event := range session.Events {
		if event.EventType == "navigation" {
			if metadata := event.Metadata; metadata != nil {
				if external, ok := metadata["external"].(bool); ok && external {
					return true, "external navigation (expected)"
				}
				if share, ok := metadata["share"].(bool); ok && share {
					return true, "share action (positive signal)"
				}
				if bookmark, ok := metadata["bookmark"].(bool); ok && bookmark {
					return true, "bookmark action (positive signal)"
				}
			}
		}
	}

	return false, ""
}

// Helper functions

func getClicksForTarget(events []types.Event, targetID string) []types.Event {
	var clicks []types.Event
	for _, event := range events {
		if event.EventType == "click" && event.Target.ID == targetID {
			clicks = append(clicks, event)
		}
	}
	return clicks
}
