/**
 * Ticket Types
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Ticket content types
 */

package types

// Ticket represents a formatted ticket ready for export
type Ticket struct {
	Title       string
	Description string
	Labels      []string
	Metadata    map[string]string
}

// TicketContent represents structured ticket content
type TicketContent struct {
	Summary             string
	WhatUsersWereTrying string
	WhatWentWrong       string
	Evidence            Evidence
	Confidence          ConfidenceInfo
	Notes               string // Optional AI hypothesis, clearly marked
}

// Evidence provides evidence for the incident
type Evidence struct {
	AffectedSessions   int
	TimeWindow         string
	ExampleSessionRefs []string
}

// ConfidenceInfo provides confidence information
type ConfidenceInfo struct {
	Score  float64
	Reason string
}
