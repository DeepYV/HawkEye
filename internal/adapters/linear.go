/**
 * Linear Adapter
 * 
 * Author: Frank Miller (Team Beta)
 * Responsibility: Linear integration
 */

package adapters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// LinearAdapter implements Linear integration
type LinearAdapter struct {
	apiURL     string
	apiKey     string
	teamID     string
	client     *http.Client
}

// NewLinearAdapter creates a new Linear adapter
func NewLinearAdapter(apiURL, apiKey, teamID string) *LinearAdapter {
	return &LinearAdapter{
		apiURL: apiURL,
		apiKey: apiKey,
		teamID: teamID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns adapter name
func (l *LinearAdapter) Name() string {
	return "linear"
}

// CreateTicket creates a ticket in Linear
func (l *LinearAdapter) CreateTicket(ctx context.Context, incident types.Incident, ticket types.Ticket, idempotencyKey string) (string, error) {
	// Check if ticket already exists (idempotency)
	existingTicketID, err := l.findTicketByKey(ctx, idempotencyKey)
	if err == nil && existingTicketID != "" {
		return existingTicketID, nil // Ticket already exists
	}

	// Build Linear issue
	issue := l.buildLinearIssue(ticket, incident)

	// GraphQL mutation to create issue
	query := `
		mutation CreateIssue($input: IssueCreateInput!) {
			issueCreate(input: $input) {
				success
				issue {
					id
					identifier
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": issue,
	}

	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	// TODO: Implement actual GraphQL request
	_ = reqBody

	return "", fmt.Errorf("Linear integration not yet implemented")
}

// buildLinearIssue builds Linear issue structure
func (l *LinearAdapter) buildLinearIssue(ticket types.Ticket, incident types.Incident) map[string]interface{} {
	// Get priority from ticket metadata (already mapped by PriorityMapper)
	priority := ticket.Metadata["priority"]
	linearPriority := mapPriorityToLinear(priority)

	issue := map[string]interface{}{
		"teamId":     l.teamID,
		"title":      ticket.Title,
		"description": ticket.Description,
		"priority":   linearPriority,
		"labels":     ticket.Labels,
	}

	// Add metadata as custom fields
	for key, value := range ticket.Metadata {
		issue[key] = value
	}

	return issue
}

// findTicketByKey finds ticket by idempotency key
func (l *LinearAdapter) findTicketByKey(ctx context.Context, idempotencyKey string) (string, error) {
	// TODO: GraphQL query to find issue by custom field matching idempotency key
	return "", nil
}

// GetTicket retrieves ticket by ID
func (l *LinearAdapter) GetTicket(ctx context.Context, ticketID string) (*TicketInfo, error) {
	// TODO: GraphQL query to get issue by ID
	return nil, fmt.Errorf("not implemented")
}

// mapPriorityToLinear maps priority level to Linear priority (0-4 scale)
func mapPriorityToLinear(priority string) int {
	switch priority {
	case "P0":
		return 4 // Urgent
	case "P1":
		return 3 // High
	case "P2":
		return 2 // Medium
	case "P3":
		return 1 // Low
	case "P4":
		return 0 // Backlog
	default:
		return 2
	}
}