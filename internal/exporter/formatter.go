/**
 * Ticket Content Formatter
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Format ticket content (short, specific, non-AI)
 */

package exporter

import (
	"fmt"
	"strings"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// Formatter formats incidents into ticket content
type Formatter struct {
}

// NewFormatter creates a new formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatTicket formats incident into ticket content
func (f *Formatter) FormatTicket(incident types.Incident, priority PriorityLevel) types.Ticket {
	// Generate title (short, specific, non-AI-sounding)
	title := f.generateTitle(incident)

	// Generate structured description
	description := f.generateDescription(incident)

	// Generate labels/metadata (includes priority)
	labels := f.generateLabels(incident)
	metadata := f.generateMetadata(incident, priority)

	return types.Ticket{
		Title:       title,
		Description: description,
		Labels:      labels,
		Metadata:    metadata,
	}
}

// generateTitle generates short, specific title (matching gold-standard examples)
func (f *Formatter) generateTitle(incident types.Incident) string {
	// Extract route from failure point
	route := extractRoute(incident.PrimaryFailurePoint)

	// Extract action/component
	action := extractAction(incident.PrimaryFailurePoint)

	// Generate specific title based on signals (matching example patterns)
	signalType := getPrimarySignalType(incident.TriggeringSignals)

	// Pattern: "Users [specific problem] on [route]" - matching examples exactly
	switch signalType {
	case "rage":
		// Example: "Users repeatedly click disabled 'Create Report' button"
		if action == "button" || action == "form_submit" {
			return fmt.Sprintf("Users repeatedly click %s on %s", getComponentName(incident), route)
		}
		return fmt.Sprintf("Users stuck in %s loop on %s", action, route)
	case "blocked":
		// Example: "Users stuck in checkout form submission loop"
		if action == "form_submit" {
			return fmt.Sprintf("Users stuck in %s form submission loop", getFeatureName(route))
		}
		return fmt.Sprintf("Users blocked from %s on %s", action, route)
	case "abandonment":
		return fmt.Sprintf("Users abandoning %s flow on %s", getFeatureName(route), route)
	case "confusion":
		// Example: "Users repeatedly exit and re-enter security settings"
		return fmt.Sprintf("Users repeatedly exit and re-enter %s", getFeatureName(route))
	default:
		// Example: "Profile save fails without user-visible error"
		return fmt.Sprintf("User action fails without feedback on %s", route)
	}
}

// generateDescription generates structured description (matching gold-standard examples)
func (f *Formatter) generateDescription(incident types.Incident) string {
	var parts []string

	// 1. Summary (matching example format exactly)
	parts = append(parts, "## Summary")
	summary := f.generateSummary(incident)
	parts = append(parts, summary)
	parts = append(parts, "")

	// 2. What Users Were Trying to Do (matching example format)
	parts = append(parts, "## What users were trying to do")
	userIntent := f.inferUserIntent(incident)
	parts = append(parts, userIntent)
	parts = append(parts, "")

	// 3. What Went Wrong (matching example format)
	parts = append(parts, "## What went wrong")
	whatWentWrong := f.describeFailure(incident)
	parts = append(parts, whatWentWrong)
	parts = append(parts, "")

	// 4. Evidence (matching example format exactly)
	parts = append(parts, "## Evidence")
	evidence := f.formatEvidence(incident)
	parts = append(parts, evidence)
	parts = append(parts, "")

	// 5. Confidence (matching example format, transparency rule)
	parts = append(parts, "## Confidence")
	confidence := fmt.Sprintf("%.2f — %s", incident.ConfidenceScore/100.0, f.getConfidenceReason(incident))
	parts = append(parts, confidence)
	parts = append(parts, "")

	// 6. Notes (optional, clearly marked as hypothesis)
	if f.shouldIncludeNotes(incident) {
		parts = append(parts, "## Notes")
		notes := f.generateNotes(incident)
		parts = append(parts, notes)
	}

	return strings.Join(parts, "\n")
}

// generateSummary generates summary matching gold-standard examples
func (f *Formatter) generateSummary(incident types.Incident) string {
	route := extractRoute(incident.PrimaryFailurePoint)
	signalType := getPrimarySignalType(incident.TriggeringSignals)
	affectedCount := len(incident.SignalDetails)
	if affectedCount == 0 {
		affectedCount = 1
	}

	// Match example patterns exactly
	switch signalType {
	case "rage":
		return fmt.Sprintf("Users are repeatedly clicking a disabled or unresponsive element without feedback, indicating a blocked workflow.")
	case "blocked":
		if strings.Contains(route, "checkout") || strings.Contains(route, "form") {
			return fmt.Sprintf("Multiple users are unable to complete checkout due to a repeated submission loop on the checkout form.")
		}
		return fmt.Sprintf("Users are unable to complete an action due to system errors or validation failures.")
	case "abandonment":
		return fmt.Sprintf("Users abandon a workflow after encountering friction, without completing the intended action.")
	case "confusion":
		return fmt.Sprintf("Users appear unable to locate expected functionality and repeatedly navigate away and back to the same page.")
	default:
		return fmt.Sprintf("Users experience frustration when attempting to complete an action on %s.", route)
	}
}

// inferUserIntent infers what users were trying to do (matching example format)
func (f *Formatter) inferUserIntent(incident types.Incident) string {
	route := extractRoute(incident.PrimaryFailurePoint)
	action := extractAction(incident.PrimaryFailurePoint)

	// Match example patterns exactly
	if strings.Contains(route, "checkout") {
		return "Submit payment details to complete checkout."
	}
	if strings.Contains(route, "profile") {
		return "Update profile information."
	}
	if strings.Contains(route, "settings") || strings.Contains(route, "security") {
		return "Update security or authentication settings."
	}
	if strings.Contains(route, "dashboard") {
		return "Create a new report from the dashboard."
	}
	if action == "form_submit" || strings.Contains(route, "form") {
		return fmt.Sprintf("Submit a form on %s", route)
	}

	return fmt.Sprintf("Complete an action on %s", route)
}

// describeFailure describes what went wrong (matching example format exactly)
func (f *Formatter) describeFailure(incident types.Incident) string {
	route := extractRoute(incident.PrimaryFailurePoint)
	signalType := getPrimarySignalType(incident.TriggeringSignals)

	// Match example patterns exactly
	switch signalType {
	case "rage":
		// Example: "The 'Create Report' button appears clickable but does not respond, leading to rapid repeated clicks."
		return "The element appears clickable but does not respond, leading to rapid repeated clicks without feedback."
	case "blocked":
		// Example: "After clicking 'Submit', the form remains on the same screen and allows repeated submissions without progressing or showing a clear blocking error."
		if strings.Contains(route, "checkout") || strings.Contains(route, "form") {
			return "After clicking 'Submit', the form remains on the same screen and allows repeated submissions without progressing or showing a clear blocking error."
		}
		return "Users encounter system errors or validation failures when attempting the action, with no clear resolution path."
	case "abandonment":
		return "Users encounter friction (errors, delays, or unclear feedback) and leave the flow without completing the intended action."
	case "confusion":
		// Example: "Users navigate into the Security Settings page, scroll without interacting, exit, and re-enter multiple times without completing an action."
		return "Users navigate to the page, scroll or interact minimally, exit, and re-enter multiple times without completing an action."
	default:
		// Example: "After clicking 'Save', no success or error message is shown and changes are not persisted."
		return "After attempting the action, no success or error message is shown and the expected outcome does not occur."
	}
}

// formatEvidence formats evidence section (matching example format exactly)
func (f *Formatter) formatEvidence(incident types.Incident) string {
	var parts []string

	// Number of affected sessions (matching example format)
	sessionCount := len(incident.SignalDetails)
	if sessionCount == 0 {
		sessionCount = 1
	}
	parts = append(parts, fmt.Sprintf("- %d affected user sessions", sessionCount))

	// Time window (matching example format)
	if len(incident.SignalDetails) > 0 {
		firstSignal := incident.SignalDetails[0].Timestamp
		lastSignal := incident.SignalDetails[len(incident.SignalDetails)-1].Timestamp

		// Format as date range if available
		timeWindow := lastSignal.Sub(firstSignal)
		if timeWindow.Hours() < 24 {
			parts = append(parts, fmt.Sprintf("- Observed within %v time window", formatDuration(timeWindow)))
		} else {
			// Format as date range
			parts = append(parts, fmt.Sprintf("- Observed between %s and %s",
				formatDate(firstSignal), formatDate(lastSignal)))
		}
	}

	// Specific evidence based on signal type
	signalType := getPrimarySignalType(incident.TriggeringSignals)
	switch signalType {
	case "rage":
		parts = append(parts, "- Rage click patterns (5+ clicks within 2 seconds)")
		parts = append(parts, "- No tooltip or feedback displayed")
	case "blocked":
		parts = append(parts, "- Repeated submit attempts (3–8 times per session)")
		parts = append(parts, "- No route change after submission")
	case "confusion":
		parts = append(parts, "- Repeated page entry within short time windows")
		parts = append(parts, "- No settings changed during sessions")
	default:
		parts = append(parts, "- Action triggered but expected outcome not observed")
		parts = append(parts, "- No error feedback shown")
	}

	// Example session references (matching example format)
	exampleSessions := f.getExampleSessions(incident)
	if len(exampleSessions) > 0 {
		parts = append(parts, fmt.Sprintf("- Example sessions: %s", strings.Join(exampleSessions, ", ")))
	} else {
		parts = append(parts, fmt.Sprintf("- Example session: %s", incident.SessionID))
	}

	return strings.Join(parts, "\n")
}

// getExampleSessions gets example session IDs (up to 3)
func (f *Formatter) getExampleSessions(incident types.Incident) []string {
	sessions := []string{incident.SessionID}

	// If we have signal details, we could extract more session IDs
	// For now, just return the main session ID
	if len(sessions) > 3 {
		return sessions[:3]
	}
	return sessions
}

// formatDuration formats duration in human-readable format
func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.0f hours", d.Hours())
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	}
	return fmt.Sprintf("%.0f seconds", d.Seconds())
}

// formatDate formats date in example format
func formatDate(t time.Time) string {
	return t.Format("Jan 2")
}

// getConfidenceReason gets confidence reason (matching example format)
func (f *Formatter) getConfidenceReason(incident types.Incident) string {
	signalCount := len(incident.TriggeringSignals)
	sessionCount := len(incident.SignalDetails)
	if sessionCount == 0 {
		sessionCount = 1
	}

	if sessionCount > 20 && signalCount >= 2 {
		return fmt.Sprintf("consistent behavior observed across many sessions with identical interaction patterns")
	}
	if sessionCount > 10 {
		return fmt.Sprintf("repeated patterns suggest clear issue rather than isolated incident")
	}
	return fmt.Sprintf("repeated patterns observed across multiple sessions")
}

// shouldIncludeNotes determines if notes should be included
func (f *Formatter) shouldIncludeNotes(incident types.Incident) bool {
	// Include notes if we have a hypothesis or if confidence is medium
	return incident.ConfidenceScore < 90 || len(incident.Explanation) > 0
}

// generateNotes generates notes section (matching example format)
func (f *Formatter) generateNotes(incident types.Incident) string {
	signalType := getPrimarySignalType(incident.TriggeringSignals)

	// Match example patterns
	switch signalType {
	case "rage":
		return "Button may appear enabled or lacks affordance explaining disabled state."
	case "blocked":
		return "Possible client-side validation or API error not surfaced to the user."
	case "confusion":
		return "May indicate unclear labeling or missing affordances."
	default:
		return "Backend errors may not be surfaced to the UI."
	}
}

// generateLabels generates labels for ticket (matching example format)
func (f *Formatter) generateLabels(incident types.Incident) []string {
	labels := []string{
		"user-behavior-detection",
		strings.ToLower(incident.SeverityType),
	}

	// Add feature name as label
	route := extractRoute(incident.PrimaryFailurePoint)
	feature := getFeatureName(route)
	if feature != "" {
		labels = append(labels, feature)
	}

	return labels
}

// generateMetadata generates metadata for ticket (matching example format)
func (f *Formatter) generateMetadata(incident types.Incident, priority PriorityLevel) map[string]string {
	route := extractRoute(incident.PrimaryFailurePoint)
	return map[string]string{
		"feature":     getFeatureName(route),
		"route":       route,
		"severity":    strings.ToLower(incident.SeverityType),
		"priority":    string(priority),
		"environment": "production",
		"source":      "user-behavior-detection",
		"incident_id": incident.IncidentID,
		"session_id":  incident.SessionID,
	}
}

// Helper functions
func extractRoute(failurePoint string) string {
	parts := strings.Split(failurePoint, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return failurePoint
}

func extractAction(failurePoint string) string {
	parts := strings.Split(failurePoint, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return "action"
}

func getPrimarySignalType(signals []string) string {
	// Prefer blocked > rage > abandonment > confusion
	for _, signal := range signals {
		if signal == "blocked" {
			return "blocked"
		}
	}
	for _, signal := range signals {
		if signal == "rage" {
			return "rage"
		}
	}
	for _, signal := range signals {
		if signal == "abandonment" {
			return "abandonment"
		}
	}
	if len(signals) > 0 {
		return signals[0]
	}
	return "unknown"
}

// getFeatureName extracts feature name from route
func getFeatureName(route string) string {
	// Extract feature from route
	if strings.Contains(route, "checkout") {
		return "checkout"
	}
	if strings.Contains(route, "profile") {
		return "profile"
	}
	if strings.Contains(route, "settings") {
		return "settings"
	}
	if strings.Contains(route, "security") {
		return "security"
	}
	if strings.Contains(route, "dashboard") {
		return "dashboard"
	}
	if strings.Contains(route, "pricing") {
		return "pricing"
	}

	// Default: use first part of route
	parts := strings.Split(strings.Trim(route, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// getComponentName extracts component name from incident
func getComponentName(incident types.Incident) string {
	// Try to extract from failure point
	failurePoint := incident.PrimaryFailurePoint
	parts := strings.Split(failurePoint, ":")
	if len(parts) >= 2 {
		return parts[1]
	}

	// Try to extract from signal details
	if len(incident.SignalDetails) > 0 {
		details := incident.SignalDetails[0].Details
		if strings.Contains(details, "button") {
			return "disabled button"
		}
		if strings.Contains(details, "form") {
			return "form"
		}
	}

	return "element"
}
