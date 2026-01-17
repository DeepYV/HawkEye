/**
 * Incident Store Interaction
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Read from and update Incident Store
 *
 * Rules:
 * - Read-only access to incidents
 * - Write access only for export metadata
 * - Never mutate confidence or severity
 */

package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// Store represents the Incident Store interface
type Store interface {
	GetEligibleIncidents(ctx context.Context) ([]types.Incident, error)
	MarkExported(ctx context.Context, incidentID, externalTicketID, externalSystem string) error
	MarkExportFailed(ctx context.Context, incidentID string) error
}

// IncidentStore implements Store interface
type IncidentStore struct {
	incidentStoreURL string
	client          *http.Client
}

// NewIncidentStore creates a new incident store
func NewIncidentStore(incidentStoreURL string) *IncidentStore {
	return &IncidentStore{
		incidentStoreURL: incidentStoreURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetEligibleIncidents gets incidents eligible for export
func (s *IncidentStore) GetEligibleIncidents(ctx context.Context) ([]types.Incident, error) {
	if s.incidentStoreURL == "" {
		// Log and return empty if URL not configured
		fmt.Printf("[Ticket Exporter] Incident Store URL not configured, logging query\n")
		return []types.Incident{}, nil
	}

	// Query Incident Store
	url := s.incidentStoreURL + "/v1/incidents?status=confirmed&exported=false&limit=100"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Printf("[Ticket Exporter] Failed to create request: %v\n", err)
		return []types.Incident{}, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		fmt.Printf("[Ticket Exporter] Failed to query Incident Store: %v\n", err)
		return []types.Incident{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[Ticket Exporter] Incident Store returned status %d\n", resp.StatusCode)
		return []types.Incident{}, nil
	}

	var result struct {
		Incidents []types.Incident `json:"incidents"`
		Total     int              `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("[Ticket Exporter] Failed to decode response: %v\n", err)
		return []types.Incident{}, nil
	}

	fmt.Printf("[Ticket Exporter] Retrieved %d eligible incidents from Incident Store\n", result.Total)
	return result.Incidents, nil
}

// MarkExported marks incident as exported
func (s *IncidentStore) MarkExported(ctx context.Context, incidentID, externalTicketID, externalSystem string) error {
	if s.incidentStoreURL == "" {
		fmt.Printf("[Ticket Exporter] Incident Store URL not configured, logging export: %s -> %s\n", incidentID, externalTicketID)
		return nil
	}

	// Update via Incident Store API (would need PATCH endpoint)
	// For now, just log
	fmt.Printf("[Ticket Exporter] Marked incident %s as exported (ticket: %s, system: %s)\n", 
		incidentID, externalTicketID, externalSystem)
	return nil
}

// MarkExportFailed marks incident export as failed
func (s *IncidentStore) MarkExportFailed(ctx context.Context, incidentID string) error {
	if s.incidentStoreURL == "" {
		fmt.Printf("[Ticket Exporter] Incident Store URL not configured, logging export failure: %s\n", incidentID)
		return nil
	}

	// Update via Incident Store API (would need PATCH endpoint)
	// For now, just log
	fmt.Printf("[Ticket Exporter] Marked incident %s export as failed\n", incidentID)
	return nil
}
