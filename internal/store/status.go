/**
 * Status Management
 * 
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: Incident status management
 */

package store

const (
	StatusDraft     = "draft"
	StatusConfirmed = "confirmed"
	StatusSuppressed = "suppressed"
)

// ValidateStatus validates incident status
func ValidateStatus(status string) bool {
	switch status {
	case StatusDraft, StatusConfirmed, StatusSuppressed:
		return true
	default:
		return false
	}
}