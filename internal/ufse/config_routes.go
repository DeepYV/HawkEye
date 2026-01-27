/**
 * Route-Specific Threshold Configuration
 *
 * Author: Enhanced Detection Team
 * Responsibility: Allow per-route/feature detection thresholds
 *
 * Rationale:
 * - Different routes have different expected user behavior
 * - Checkout flows may need stricter rage detection
 * - Gaming/interactive features may need relaxed click thresholds
 * - Allows fine-tuning without global changes
 *
 * Configuration Structure:
 * - Route patterns support glob-style matching
 * - More specific patterns take precedence
 * - Fallback to global defaults if no match
 */

package ufse

import (
	"regexp"
	"strings"
	"time"
)

// RouteConfig holds detection configuration for a specific route
type RouteConfig struct {
	// Route matching
	Pattern        string         // Glob pattern (e.g., "/checkout/*", "/api/**")
	compiledRegex  *regexp.Regexp // Compiled regex from pattern

	// Rage detection overrides
	RageMinClicks            *int
	RageTimeWindow           *time.Duration
	RageMaxTimeBetweenClicks *time.Duration

	// Form loop detection overrides
	FormLoopMinSubmissions *int
	FormLoopTimeWindow     *time.Duration

	// Confidence overrides
	MinConfidenceForEmit *string // "Low", "Medium", "High"

	// Feature flags
	DisableRageDetection      bool
	DisableBlockedDetection   bool
	DisableAbandonmentDetection bool
	DisableConfusionDetection bool
	DisableFormLoopDetection  bool

	// Shadow mode (detect but don't emit)
	ShadowModeEnabled bool

	// Priority (higher = checked first)
	Priority int
}

// RouteConfigManager manages route-specific configurations
type RouteConfigManager struct {
	configs       []RouteConfig
	defaultConfig DetectionConfig
}

// NewRouteConfigManager creates a new route config manager
func NewRouteConfigManager(defaultConfig DetectionConfig) *RouteConfigManager {
	return &RouteConfigManager{
		configs:       make([]RouteConfig, 0),
		defaultConfig: defaultConfig,
	}
}

// AddRouteConfig adds a route-specific configuration
func (m *RouteConfigManager) AddRouteConfig(config RouteConfig) error {
	// Compile pattern to regex
	regex, err := patternToRegex(config.Pattern)
	if err != nil {
		return err
	}
	config.compiledRegex = regex

	// Insert sorted by priority (highest first)
	inserted := false
	for i, existing := range m.configs {
		if config.Priority > existing.Priority {
			m.configs = append(m.configs[:i], append([]RouteConfig{config}, m.configs[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		m.configs = append(m.configs, config)
	}

	return nil
}

// GetConfigForRoute returns the configuration for a specific route
func (m *RouteConfigManager) GetConfigForRoute(route string) MergedRouteConfig {
	// Find matching config
	for _, config := range m.configs {
		if config.compiledRegex != nil && config.compiledRegex.MatchString(route) {
			return m.mergeConfigs(config)
		}
	}

	// Return default config
	return m.defaultToMerged()
}

// MergedRouteConfig is the final configuration with all overrides applied
type MergedRouteConfig struct {
	// Rage detection
	RageMinClicks            int
	RageTimeWindow           time.Duration
	RageMaxTimeBetweenClicks time.Duration

	// Form loop detection
	FormLoopMinSubmissions int
	FormLoopTimeWindow     time.Duration

	// Confidence
	MinConfidenceForEmit string

	// Feature flags
	DisableRageDetection      bool
	DisableBlockedDetection   bool
	DisableAbandonmentDetection bool
	DisableConfusionDetection bool
	DisableFormLoopDetection  bool

	// Shadow mode
	ShadowModeEnabled bool

	// Route info
	MatchedPattern string
}

// mergeConfigs merges route config with defaults
func (m *RouteConfigManager) mergeConfigs(routeConfig RouteConfig) MergedRouteConfig {
	merged := m.defaultToMerged()
	merged.MatchedPattern = routeConfig.Pattern

	// Apply overrides
	if routeConfig.RageMinClicks != nil {
		merged.RageMinClicks = *routeConfig.RageMinClicks
	}
	if routeConfig.RageTimeWindow != nil {
		merged.RageTimeWindow = *routeConfig.RageTimeWindow
	}
	if routeConfig.RageMaxTimeBetweenClicks != nil {
		merged.RageMaxTimeBetweenClicks = *routeConfig.RageMaxTimeBetweenClicks
	}
	if routeConfig.FormLoopMinSubmissions != nil {
		merged.FormLoopMinSubmissions = *routeConfig.FormLoopMinSubmissions
	}
	if routeConfig.FormLoopTimeWindow != nil {
		merged.FormLoopTimeWindow = *routeConfig.FormLoopTimeWindow
	}
	if routeConfig.MinConfidenceForEmit != nil {
		merged.MinConfidenceForEmit = *routeConfig.MinConfidenceForEmit
	}

	// Feature flags (always override if set)
	merged.DisableRageDetection = routeConfig.DisableRageDetection
	merged.DisableBlockedDetection = routeConfig.DisableBlockedDetection
	merged.DisableAbandonmentDetection = routeConfig.DisableAbandonmentDetection
	merged.DisableConfusionDetection = routeConfig.DisableConfusionDetection
	merged.DisableFormLoopDetection = routeConfig.DisableFormLoopDetection
	merged.ShadowModeEnabled = routeConfig.ShadowModeEnabled

	return merged
}

// defaultToMerged converts default config to merged config
func (m *RouteConfigManager) defaultToMerged() MergedRouteConfig {
	return MergedRouteConfig{
		RageMinClicks:            m.defaultConfig.RageHighMinClicks,
		RageTimeWindow:           m.defaultConfig.RageHighTimeWindow,
		RageMaxTimeBetweenClicks: m.defaultConfig.RageHighMaxTimeBetweenClicks,
		FormLoopMinSubmissions:   3, // Default
		FormLoopTimeWindow:       30 * time.Second,
		MinConfidenceForEmit:     "High",
		DisableRageDetection:     false,
		DisableBlockedDetection:  false,
		DisableAbandonmentDetection: false,
		DisableConfusionDetection: false,
		DisableFormLoopDetection: false,
		ShadowModeEnabled:        false,
		MatchedPattern:           "",
	}
}

// patternToRegex converts a glob-style pattern to regex
func patternToRegex(pattern string) (*regexp.Regexp, error) {
	// Escape regex special characters except * and ?
	escaped := regexp.QuoteMeta(pattern)

	// Convert glob patterns to regex
	// ** matches any characters including /
	escaped = strings.ReplaceAll(escaped, `\*\*`, `.*`)
	// * matches any characters except /
	escaped = strings.ReplaceAll(escaped, `\*`, `[^/]*`)
	// ? matches single character
	escaped = strings.ReplaceAll(escaped, `\?`, `.`)

	// Anchor the pattern
	escaped = "^" + escaped + "$"

	return regexp.Compile(escaped)
}

// IsDetectorDisabled checks if a detector is disabled for a route
func (c *MergedRouteConfig) IsDetectorDisabled(detectorType string) bool {
	switch detectorType {
	case "rage":
		return c.DisableRageDetection
	case "blocked":
		return c.DisableBlockedDetection
	case "abandonment":
		return c.DisableAbandonmentDetection
	case "confusion":
		return c.DisableConfusionDetection
	case "form_loop":
		return c.DisableFormLoopDetection
	default:
		return false
	}
}

// PredefinedRouteConfigs returns common route configurations
func PredefinedRouteConfigs() []RouteConfig {
	// Helper to create pointers
	intPtr := func(i int) *int { return &i }
	durPtr := func(d time.Duration) *time.Duration { return &d }
	strPtr := func(s string) *string { return &s }

	return []RouteConfig{
		// Checkout/Payment flows - stricter detection
		{
			Pattern:                  "/checkout/**",
			RageMinClicks:            intPtr(3), // Lower threshold
			RageTimeWindow:           durPtr(5 * time.Second),
			MinConfidenceForEmit:     strPtr("Medium"), // Emit medium confidence
			Priority:                 100,
		},
		{
			Pattern:                  "/payment/**",
			RageMinClicks:            intPtr(3),
			RageTimeWindow:           durPtr(5 * time.Second),
			MinConfidenceForEmit:     strPtr("Medium"),
			Priority:                 100,
		},

		// Gaming/Interactive - relaxed click detection
		{
			Pattern:               "/game/**",
			RageMinClicks:         intPtr(10), // Higher threshold
			DisableRageDetection:  true,       // Or disable entirely
			Priority:              90,
		},

		// API routes - disable most detection
		{
			Pattern:                   "/api/**",
			DisableRageDetection:      true,
			DisableConfusionDetection: true,
			DisableAbandonmentDetection: true,
			Priority:                  80,
		},

		// Admin/Debug routes - shadow mode
		{
			Pattern:           "/admin/**",
			ShadowModeEnabled: true,
			Priority:          70,
		},
	}
}
