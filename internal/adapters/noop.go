/**
 * No-Op Adapter
 * 
 * Author: Test Team
 * Responsibility: Test adapter that logs but doesn't actually create tickets
 */

package adapters

import (
	"context"
	"fmt"
	"log"

	"github.com/your-org/frustration-engine/internal/types"
)

// NoOpAdapter is a test adapter that logs but doesn't create tickets
type NoOpAdapter struct {
	name string
}

// NewNoOpAdapter creates a new no-op adapter
func NewNoOpAdapter() *NoOpAdapter {
	return &NoOpAdapter{
		name: "noop",
	}
}

// Name returns the adapter name
func (a *NoOpAdapter) Name() string {
	return a.name
}

// CreateTicket logs the ticket creation but doesn't actually create it
func (a *NoOpAdapter) CreateTicket(ctx context.Context, incident types.Incident, ticket types.Ticket, idempotencyKey string) (string, error) {
	log.Printf("[NoOp Adapter] Would create ticket for incident %s (idempotency: %s)", incident.IncidentID, idempotencyKey)
	log.Printf("[NoOp Adapter] Ticket title: %s", ticket.Title)
	return fmt.Sprintf("test-ticket-%s", incident.IncidentID), nil
}

// GetTicket returns nil (no-op)
func (a *NoOpAdapter) GetTicket(ctx context.Context, ticketID string) (*TicketInfo, error) {
	return nil, fmt.Errorf("noop adapter does not support GetTicket")
}
