/**
 * Ticket Exporter Engine
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Main exporter service
 */

package exporter

import (
	"context"
	"log"
	"time"

	"github.com/your-org/frustration-engine/internal/adapters"
	"github.com/your-org/frustration-engine/internal/observability"
	"github.com/your-org/frustration-engine/internal/types"
)

// Engine is the main ticket exporter engine
type Engine struct {
	store           Store
	eligibility     *EligibilityChecker
	formatter       *Formatter
	priorityMapper  *PriorityMapper
	adapter         adapters.Adapter
	rateLimiter     *RateLimiter
}

// NewEngine creates a new exporter engine
func NewEngine(store Store, adapter adapters.Adapter, exportThreshold float64, maxPerMinute int) *Engine {
	rateLimiter := NewRateLimiter(maxPerMinute)
	eligibility := NewEligibilityChecker(exportThreshold, rateLimiter)
	formatter := NewFormatter()
	priorityMapper := NewPriorityMapper()

	return &Engine{
		store:        store,
		eligibility:  eligibility,
		formatter:   formatter,
		priorityMapper: priorityMapper,
		adapter:     adapter,
		rateLimiter: rateLimiter,
	}
}

// ExportEligible exports eligible incidents (up to max count)
func (e *Engine) ExportEligible(maxCount int) {
	ctx := context.Background()

	// Get eligible incidents
	incidents, err := e.store.GetEligibleIncidents(ctx)
	if err != nil {
		log.Printf("[Ticket Exporter] Failed to get eligible incidents: %v", err)
		return
	}

	exported := 0
	for _, incident := range incidents {
		if exported >= maxCount {
			break // Rate limit reached
		}

		// Check eligibility
		eligible, reason := e.eligibility.IsEligible(incident)
		if !eligible {
			// Track skipped export
			observability.ExportsSkipped.WithLabelValues(reason).Inc()
			continue
		}

		// Track export attempt
		observability.ExportAttempts.Inc()

		// Map priority (confidence + severity)
		priority := e.priorityMapper.GetPriorityForIncident(incident)

		// Format ticket (includes priority in metadata)
		ticket := e.formatter.FormatTicket(incident, priority)

		// Generate idempotency key (one incident = one ticket)
		idempotencyKey := generateIdempotencyKey(incident)

		// Export to adapter with retry
		var externalTicketID string
		var exportErr error
		
		// Retry logic with exponential backoff
		maxRetries := 3
		backoff := 100 * time.Millisecond
		for attempt := 0; attempt <= maxRetries; attempt++ {
			externalTicketID, exportErr = e.adapter.CreateTicket(ctx, incident, ticket, idempotencyKey)
			if exportErr == nil {
				break // Success
			}
			
			if attempt < maxRetries {
				// Exponential backoff
				time.Sleep(backoff)
				backoff = backoff * 2
				if backoff > 5*time.Second {
					backoff = 5 * time.Second
				}
			}
		}
		
		if exportErr != nil {
			// Track failure
			observability.ExportFailures.Inc()
			
			// Handle failure (retry, mark failed)
			HandleExportFailure(ctx, incident, exportErr, e.store)
			continue
		}

		// Mark as exported in store
		if err := e.store.MarkExported(ctx, incident.IncidentID, externalTicketID, e.adapter.Name()); err != nil {
			log.Printf("[Ticket Exporter] Failed to mark incident as exported: %v", err)
			// Continue - ticket was created, just metadata update failed
		}

		// Track successful export
		observability.ExportsSuccessful.Inc()
		exported++
	}
}

// generateIdempotencyKey generates idempotency key for incident
func generateIdempotencyKey(incident types.Incident) string {
	// Use incident ID as idempotency key (one incident = one ticket)
	return "incident_" + incident.IncidentID
}