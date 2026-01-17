/**
 * Jira Adapter
 * 
 * Author: Frank Miller (Team Beta)
 * Responsibility: Jira integration
 */

package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// JiraAdapter implements Jira integration
type JiraAdapter struct {
	baseURL    string
	apiToken   string
	projectKey string
	client     *http.Client
}

// NewJiraAdapter creates a new Jira adapter
func NewJiraAdapter(baseURL, apiToken, projectKey string) *JiraAdapter {
	return &JiraAdapter{
		baseURL:    baseURL,
		apiToken:   apiToken,
		projectKey: projectKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns adapter name
func (j *JiraAdapter) Name() string {
	return "jira"
}

// CreateTicket creates a ticket in Jira
func (j *JiraAdapter) CreateTicket(ctx context.Context, incident types.Incident, ticket types.Ticket, idempotencyKey string) (string, error) {
	// Check if ticket already exists (idempotency)
	existingTicketID, err := j.findTicketByKey(ctx, idempotencyKey)
	if err == nil && existingTicketID != "" {
		return existingTicketID, nil // Ticket already exists
	}

	// Create Jira issue
	issue := j.buildJiraIssue(ticket, incident)

	// POST to Jira API
	url := fmt.Sprintf("%s/rest/api/3/issue", j.baseURL)
	reqBody, err := json.Marshal(issue)
	if err != nil {
		return "", fmt.Errorf("failed to marshal issue: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+j.apiToken)
	req.Body = http.MaxBytesReader(nil, req.Body, 1*1024*1024) // 1MB limit

	// TODO: Implement actual HTTP request
	// For now, return placeholder
	_ = reqBody

	return "", fmt.Errorf("Jira integration not yet implemented")
}

// buildJiraIssue builds Jira issue structure
func (j *JiraAdapter) buildJiraIssue(ticket types.Ticket, incident types.Incident) map[string]interface{} {
	// Get priority from ticket metadata (already mapped by PriorityMapper)
	priority := ticket.Metadata["priority"]
	jiraPriority := mapPriorityToJira(priority)

	// Build issue fields
	fields := map[string]interface{}{
		"project": map[string]string{
			"key": j.projectKey,
		},
		"summary":     ticket.Title,
		"description": map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{"type": "text", "text": ticket.Description},
					},
				},
			},
		},
		"issuetype": map[string]string{
			"name": "Bug", // Or "Task" based on severity
		},
		"priority": map[string]string{
			"name": jiraPriority,
		},
		"labels": ticket.Labels,
	}

	// Add custom fields from metadata
	for key, value := range ticket.Metadata {
		fields[key] = value
	}

	return map[string]interface{}{
		"fields": fields,
	}
}

// findTicketByKey finds ticket by idempotency key
func (j *JiraAdapter) findTicketByKey(ctx context.Context, idempotencyKey string) (string, error) {
	// TODO: Search Jira for ticket with custom field matching idempotency key
	// This ensures idempotency
	return "", nil
}

// GetTicket retrieves ticket by ID
func (j *JiraAdapter) GetTicket(ctx context.Context, ticketID string) (*TicketInfo, error) {
	// TODO: GET /rest/api/3/issue/{ticketID}
	return nil, fmt.Errorf("not implemented")
}

// mapPriorityToJira maps priority level to Jira priority
func mapPriorityToJira(priority string) string {
	switch priority {
	case "P0":
		return "Blocker"
	case "P1":
		return "Highest"
	case "P2":
		return "High"
	case "P3":
		return "Medium"
	case "P4":
		return "Low"
	default:
		return "Medium"
	}
}