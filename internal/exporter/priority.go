/**
 * Priority Mapping
 * 
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Map confidence + severity to ticket priority
 * 
 * Core Principle:
 * - Confidence determines whether we create a ticket
 * - Severity determines how urgent it is
 * - Priority is never based on confidence alone
 */

package exporter

import (
	"github.com/your-org/frustration-engine/internal/types"
)

// PriorityLevel represents ticket priority
type PriorityLevel string

const (
	PriorityP0 PriorityLevel = "P0" // Urgent (Critical + High Confidence)
	PriorityP1 PriorityLevel = "P1" // High
	PriorityP2 PriorityLevel = "P2" // Medium
	PriorityP3 PriorityLevel = "P3" // Low
	PriorityP4 PriorityLevel = "P4" // Backlog
)

// PriorityMapper maps confidence + severity to priority
type PriorityMapper struct {
}

// NewPriorityMapper creates a new priority mapper
func NewPriorityMapper() *PriorityMapper {
	return &PriorityMapper{}
}

// MapPriority maps confidence and severity to priority
func (p *PriorityMapper) MapPriority(confidenceScore float64, severity string) PriorityLevel {
	// Confidence bands (0-100 scale, convert to 0-1 for mapping)
	confidence := confidenceScore / 100.0

	// Map based on confidence band and severity
	if confidence >= 0.90 {
		// ≥ 0.90: Certain
		return p.mapHighConfidence(severity)
	} else if confidence >= 0.80 {
		// 0.80-0.89: Very likely
		return p.mapVeryLikely(severity)
	} else {
		// 0.70-0.79: Likely issue
		return p.mapLikely(severity)
	}
}

// mapHighConfidence maps priority for confidence ≥ 0.90
func (p *PriorityMapper) mapHighConfidence(severity string) PriorityLevel {
	switch severity {
	case "Critical":
		return PriorityP0 // Urgent
	case "High":
		return PriorityP1
	case "Medium":
		return PriorityP2
	case "Low":
		return PriorityP3
	default:
		return PriorityP2
	}
}

// mapVeryLikely maps priority for confidence 0.80-0.89
func (p *PriorityMapper) mapVeryLikely(severity string) PriorityLevel {
	switch severity {
	case "Critical":
		return PriorityP1
	case "High":
		return PriorityP1
	case "Medium":
		return PriorityP2
	case "Low":
		return PriorityP3
	default:
		return PriorityP3
	}
}

// mapLikely maps priority for confidence 0.70-0.79
func (p *PriorityMapper) mapLikely(severity string) PriorityLevel {
	switch severity {
	case "Critical":
		return PriorityP2
	case "High":
		return PriorityP2
	case "Medium":
		return PriorityP3
	case "Low":
		return PriorityP4 // Backlog
	default:
		return PriorityP4
	}
}

// GetJiraPriority maps priority to Jira priority
func (p *PriorityMapper) GetJiraPriority(priority PriorityLevel) string {
	switch priority {
	case PriorityP0:
		return "Blocker"
	case PriorityP1:
		return "Highest"
	case PriorityP2:
		return "High"
	case PriorityP3:
		return "Medium"
	case PriorityP4:
		return "Low"
	default:
		return "Medium"
	}
}

// GetLinearPriority maps priority to Linear priority (0-4 scale)
func (p *PriorityMapper) GetLinearPriority(priority PriorityLevel) int {
	switch priority {
	case PriorityP0:
		return 4 // Urgent
	case PriorityP1:
		return 3 // High
	case PriorityP2:
		return 2 // Medium
	case PriorityP3:
		return 1 // Low
	case PriorityP4:
		return 0 // Backlog
	default:
		return 2
	}
}

// GetPriorityForIncident gets priority for an incident
func (p *PriorityMapper) GetPriorityForIncident(incident types.Incident) PriorityLevel {
	return p.MapPriority(incident.ConfidenceScore, incident.SeverityType)
}