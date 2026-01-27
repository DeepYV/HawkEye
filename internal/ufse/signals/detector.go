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
	Type      string // "rage", "blocked", "abandonment", "confusion", "form_loop"
	Timestamp int64  // Unix timestamp
	Route     string
	Details   map[string]interface{}
}

// DetectCandidateSignals detects all candidate signals in classified events
// Uses enhanced detectors for improved frustration detection
func DetectCandidateSignals(classified []ClassifiedEvent, session types.Session) []CandidateSignal {
	candidates := make([]CandidateSignal, 0)

	// Use enhanced detectors for better accuracy
	enhancedRageDetector := NewEnhancedRageDetector()
	rageBaitDetector := NewRageBaitDetector()
	blockedDetector := NewRefinedBlockedDetector()
	abandonmentDetector := NewRefinedAbandonmentDetector()
	confusionDetector := NewRefinedConfusionDetector()
	formLoopDetector := NewFormLoopDetector()

	// Detect each type of candidate signal with enhanced detection
	candidates = append(candidates, enhancedRageDetector.DetectRageMultiTier(classified, session)...)
	candidates = append(candidates, rageBaitDetector.DetectRageBait(classified, session)...)
	candidates = append(candidates, blockedDetector.DetectBlockedProgressRefined(classified, session)...)
	candidates = append(candidates, abandonmentDetector.DetectAbandonmentRefined(classified, session)...)
	candidates = append(candidates, confusionDetector.DetectConfusionRefined(classified, session)...)
	candidates = append(candidates, formLoopDetector.DetectFormLoops(classified, session)...)

	return candidates
}
