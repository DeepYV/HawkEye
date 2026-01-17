/**
 * Adapter Interface
 * 
 * Author: Frank Miller (Team Beta)
 * Responsibility: Define adapter interface for pluggable ticket systems
 */

package adapters

import (
	"context"

	"github.com/your-org/frustration-engine/internal/types"
)

// Adapter interface for ticket system integrations
type Adapter interface {
	// CreateTicket creates a ticket in the external system
	// Returns external ticket ID and error
	// Must be idempotent (same incident = same ticket ID)
	CreateTicket(ctx context.Context, incident types.Incident, ticket types.Ticket, idempotencyKey string) (string, error)

	// GetTicket retrieves ticket by ID (for idempotency check)
	GetTicket(ctx context.Context, ticketID string) (*TicketInfo, error)

	// Name returns adapter name (e.g., "jira", "linear")
	Name() string
}

// TicketInfo represents ticket information
type TicketInfo struct {
	ID      string
	Title   string
	Status  string
	URL     string
}