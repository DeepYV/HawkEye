/**
 * Export Scheduler
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Periodic export scheduling
 *
 * Strategy:
 * - Periodic poll (configurable interval)
 * - Max tickets per interval
 * - Predictable timing
 */

package exporter

import (
	"context"
	"time"
)

// Scheduler schedules periodic exports
type Scheduler struct {
	interval       time.Duration
	maxPerInterval int
	exporter       *Engine
}

// NewScheduler creates a new scheduler
func NewScheduler(interval time.Duration, maxPerInterval int, exporter *Engine) *Scheduler {
	return &Scheduler{
		interval:       interval,
		maxPerInterval: maxPerInterval,
		exporter:       exporter,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Export eligible incidents (up to max per interval)
			s.exporter.ExportEligible(s.maxPerInterval)
		}
	}
}
