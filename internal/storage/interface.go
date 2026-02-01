/**
 * Event Storage Interface
 *
 * Responsibility: Pluggable storage interface for event persistence.
 * Allows swapping between ClickHouse (production) and in-memory (development).
 */

package storage

import (
	"context"

	"github.com/your-org/frustration-engine/internal/types"
	pkgtypes "github.com/your-org/frustration-engine/pkg/types"
)

// EventStore defines the interface for event persistence backends.
// Production uses ClickHouse; local development uses in-memory storage.
type EventStore interface {
	// StoreEvents persists a batch of events for a given project.
	StoreEvents(ctx context.Context, projectID string, events []types.Event) error

	// Close releases any resources held by the storage backend.
	Close() error
}

// IncidentStore persists and queries detected incidents.
type IncidentStore interface {
	Save(ctx context.Context, incident pkgtypes.Incident) error
	Query(ctx context.Context, filter pkgtypes.Filter) ([]pkgtypes.Incident, error)
	Close() error
}
