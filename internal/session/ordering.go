/**
 * Event Ordering
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Event ordering and timestamp sorting
 */

package session

import (
	"sort"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// SortEventsByTimestamp sorts events by timestamp
func SortEventsByTimestamp(events []types.Event) []types.Event {
	// Create a copy to avoid mutating original
	sorted := make([]types.Event, len(events))
	copy(sorted, events)

	sort.Slice(sorted, func(i, j int) bool {
		ti, err1 := time.Parse(time.RFC3339, sorted[i].Timestamp)
		tj, err2 := time.Parse(time.RFC3339, sorted[j].Timestamp)

		// If timestamp parsing fails, preserve ingestion order
		if err1 != nil || err2 != nil {
			return i < j
		}

		// If timestamps are equal, preserve ingestion order
		if ti.Equal(tj) {
			return i < j
		}

		return ti.Before(tj)
	})

	return sorted
}

// PreserveOrderWithTimestampConflicts handles timestamp conflicts
func PreserveOrderWithTimestampConflicts(events []types.Event) []types.Event {
	// Group events by timestamp
	timestampGroups := make(map[string][]types.Event)
	order := make([]string, 0)

	for _, event := range events {
		ts := event.Timestamp
		if _, exists := timestampGroups[ts]; !exists {
			order = append(order, ts)
		}
		timestampGroups[ts] = append(timestampGroups[ts], event)
	}

	// Sort timestamps
	sort.Slice(order, func(i, j int) bool {
		ti, err1 := time.Parse(time.RFC3339, order[i])
		tj, err2 := time.Parse(time.RFC3339, order[j])

		if err1 != nil || err2 != nil {
			return i < j
		}

		return ti.Before(tj)
	})

	// Reconstruct events preserving order within timestamp groups
	result := make([]types.Event, 0, len(events))
	for _, ts := range order {
		result = append(result, timestampGroups[ts]...)
	}

	return result
}
