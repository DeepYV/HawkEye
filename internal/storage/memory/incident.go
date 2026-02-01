package memory

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/your-org/frustration-engine/pkg/types"
)

// IncidentStore stores incidents in memory for development and testing.
type IncidentStore struct {
	mu        sync.RWMutex
	incidents []types.Incident
}

// NewIncidentStore creates a new in-memory incident store.
func NewIncidentStore() *IncidentStore {
	log.Println("[storage/memory] initialised in-memory incident store")
	return &IncidentStore{incidents: make([]types.Incident, 0, 256)}
}

// Save persists an incident.
func (s *IncidentStore) Save(ctx context.Context, incident types.Incident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if incident.Status == "" {
		incident.Status = "draft"
	}
	if incident.CreatedAt.IsZero() {
		incident.CreatedAt = time.Now()
	}
	incident.UpdatedAt = time.Now()

	// Deduplicate by incident ID
	for i, existing := range s.incidents {
		if existing.IncidentID == incident.IncidentID {
			s.incidents[i] = incident
			log.Printf("[storage/memory] updated incident %s", incident.IncidentID)
			return nil
		}
	}

	s.incidents = append(s.incidents, incident)
	log.Printf("[storage/memory] stored incident %s (score: %d, confidence: %s)",
		incident.IncidentID, incident.FrustrationScore, incident.ConfidenceLevel)
	return nil
}

// Query returns incidents matching the given filter.
func (s *IncidentStore) Query(ctx context.Context, filter types.Filter) ([]types.Incident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []types.Incident
	for _, inc := range s.incidents {
		if filter.ProjectID != "" && inc.ProjectID != filter.ProjectID {
			continue
		}
		if filter.Status != "" && inc.Status != filter.Status {
			continue
		}
		if filter.MinConfidence > 0 && inc.ConfidenceScore < filter.MinConfidence {
			continue
		}
		if filter.Suppressed != nil && inc.Suppressed != *filter.Suppressed {
			continue
		}
		result = append(result, inc)
	}

	// Apply limit/offset
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	} else if filter.Offset >= len(result) {
		result = nil
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, nil
}

// Close is a no-op.
func (s *IncidentStore) Close() error { return nil }
