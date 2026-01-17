/**
 * Signal Detector
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Coordinate candidate signal detection
 */

package signals

import (
	"github.com/your-org/frustration-engine/internal/types"
)

// CandidateSignal represents a candidate signal (not yet qualified)
type CandidateSignal struct {
	Type      string // "rage", "blocked", "abandonment", "confusion"
	Timestamp int64  // Unix timestamp
	Route     string
	Details   map[string]interface{}
}

// DetectCandidateSignals detects all candidate signals in classified events
// Uses refined detectors for production-grade detection with zero false alarms
func DetectCandidateSignals(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Use refined detectors for better accuracy
	rageDetector := NewRefinedRageDetector()
	blockedDetector := NewRefinedBlockedDetector()
	abandonmentDetector := NewRefinedAbandonmentDetector()
	confusionDetector := NewRefinedConfusionDetector()

	// Detect each type of candidate signal with refined detection
	candidates = append(candidates, rageDetector.DetectRageInteractionRefined(classified, session)...)
	candidates = append(candidates, blockedDetector.DetectBlockedProgressRefined(classified, session)...)
	candidates = append(candidates, abandonmentDetector.DetectAbandonmentRefined(classified, session)...)
	candidates = append(candidates, confusionDetector.DetectConfusionRefined(classified, session)...)

	return candidates
}
