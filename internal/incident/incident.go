// Package incident provides persistence and querying for detected incidents.
package incident

import (
	"context"
	"log"

	"github.com/your-org/frustration-engine/internal/storage"
	"github.com/your-org/frustration-engine/pkg/types"
)

// Service handles incident storage and retrieval.
type Service struct {
	store storage.IncidentStore
}

// NewService creates a new incident service.
func NewService(store storage.IncidentStore) *Service {
	return &Service{store: store}
}

// Store persists an incident.
func (s *Service) Store(ctx context.Context, incident types.Incident) error {
	if err := s.store.Save(ctx, incident); err != nil {
		log.Printf("[incident] failed to store incident %s: %v", incident.IncidentID, err)
		return err
	}
	return nil
}

// Query retrieves incidents matching the filter.
func (s *Service) Query(ctx context.Context, filter types.Filter) ([]types.Incident, int, error) {
	incidents, err := s.store.Query(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return incidents, len(incidents), nil
}
