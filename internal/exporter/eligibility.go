/**
 * Export Eligibility Checker
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Strict eligibility checking
 *
 * Rules:
 * - status = confirmed
 * - confidence_score ≥ export_threshold
 * - NOT suppressed
 * - NOT already exported
 * - Export rate limits respected
 */

package exporter

import (
	"github.com/your-org/frustration-engine/internal/types"
)

const (
	// Confidence bands (system-level)
	minConfidenceForExport = 0.70 // No ticket below 0.70, ever
	defaultExportThreshold = 0.70 // Default threshold (0.70-0.79 eligible)
)

// EligibilityChecker checks if incident is eligible for export
type EligibilityChecker struct {
	exportThreshold float64
	rateLimiter     *RateLimiter
}

// NewEligibilityChecker creates a new eligibility checker
func NewEligibilityChecker(exportThreshold float64, rateLimiter *RateLimiter) *EligibilityChecker {
	if exportThreshold == 0 {
		exportThreshold = defaultExportThreshold
	}
	return &EligibilityChecker{
		exportThreshold: exportThreshold,
		rateLimiter:     rateLimiter,
	}
}

// IsEligible checks if incident is eligible for export
func (e *EligibilityChecker) IsEligible(incident types.Incident) (bool, string) {
	// Rule 1: status = confirmed
	if incident.Status != "confirmed" {
		return false, "status_not_confirmed"
	}

	// Rule 2: confidence_score ≥ 0.70 (non-negotiable system rule)
	// Confidence is 0-100 in our system, so 0.70 = 70.0
	if incident.ConfidenceScore < minConfidenceForExport*100 {
		return false, "confidence_below_minimum"
	}

	// Rule 3: confidence_score ≥ export_threshold (configurable per customer)
	if incident.ConfidenceScore < e.exportThreshold*100 {
		return false, "confidence_below_threshold"
	}

	// Rule 3: NOT suppressed
	if incident.Suppressed {
		return false, "suppressed"
	}

	// Rule 4: NOT already exported
	if incident.ExternalTicketID != "" {
		return false, "already_exported"
	}

	// Rule 5: Export rate limits respected
	if !e.rateLimiter.Allow(incident.ProjectID) {
		return false, "rate_limit_exceeded"
	}

	// All conditions met
	return true, ""
}
