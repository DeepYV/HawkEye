/**
 * UFSE Engine
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Main UFSE service
 */

package ufse

import (
	"context"
	"fmt"
	"sync"

	"github.com/your-org/frustration-engine/internal/types"
)

// Engine is the main UFSE engine
type Engine struct {
	mu                sync.RWMutex
	incidentForwarder *IncidentForwarder
}

// NewEngine creates a new UFSE engine
func NewEngine(incidentStoreURL string) *Engine {
	return &Engine{
		incidentForwarder: NewIncidentForwarder(incidentStoreURL),
	}
}

// ProcessSession processes a completed session and returns incidents
func (e *Engine) ProcessSession(ctx context.Context, session types.Session) []*types.Incident {
	// Process through pipeline (deterministic)
	incidents := ProcessSession(session)

	// Forward incidents to Incident Store using request context
	for _, incident := range incidents {
		if err := e.incidentForwarder.ForwardIncident(ctx, incident); err != nil {
			// Log error but continue processing other incidents
			fmt.Printf("[UFSE] Failed to forward incident %s: %v\n", incident.IncidentID, err)
		}
	}

	return incidents
}