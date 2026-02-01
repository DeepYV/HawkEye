// Package ingest handles event ingestion from the SDK.
//
// It validates incoming events, stores them, and forwards them to the
// session manager for aggregation.
package ingest

import (
	"context"
	"log"

	"github.com/your-org/frustration-engine/internal/metrics"
	"github.com/your-org/frustration-engine/internal/storage"
	"github.com/your-org/frustration-engine/internal/types"
	pkgtypes "github.com/your-org/frustration-engine/pkg/types"
)

// SessionForwarder sends events to the session manager.
type SessionForwarder interface {
	AddEvents(projectID, sessionID string, events []types.Event)
}

// Handler validates and processes incoming event batches.
type Handler struct {
	store     storage.EventStore
	forwarder SessionForwarder
}

// NewHandler creates a new ingest handler.
func NewHandler(store storage.EventStore, forwarder SessionForwarder) *Handler {
	return &Handler{store: store, forwarder: forwarder}
}

// Ingest validates, stores, and forwards a batch of events.
// Accepts pkg/types.Event from the HTTP layer, converts to internal types.
func (h *Handler) Ingest(ctx context.Context, projectID string, events []pkgtypes.Event) (int, error) {
	// Convert and validate
	valid := make([]types.Event, 0, len(events))
	for _, e := range events {
		if e.EventType == "" || e.SessionID == "" {
			continue
		}
		valid = append(valid, types.Event{
			EventType:      e.EventType,
			Timestamp:      e.Timestamp,
			SessionID:      e.SessionID,
			Route:          e.Route,
			Target:         types.EventTarget(e.Target),
			Metadata:       e.Metadata,
			Environment:    e.Environment,
			IdempotencyKey: e.IdempotencyKey,
		})
	}

	if len(valid) == 0 {
		return 0, nil
	}

	// Store events
	if err := h.store.StoreEvents(ctx, projectID, valid); err != nil {
		log.Printf("[ingest] storage error: %v", err)
		return 0, err
	}

	metrics.EventsIngested.Add(float64(len(valid)))

	// Group by session and forward to session manager
	grouped := make(map[string][]types.Event)
	for _, e := range valid {
		grouped[e.SessionID] = append(grouped[e.SessionID], e)
	}
	for sessionID, sessionEvents := range grouped {
		h.forwarder.AddEvents(projectID, sessionID, sessionEvents)
	}

	return len(valid), nil
}
