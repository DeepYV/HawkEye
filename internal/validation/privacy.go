/**
 * Privacy Validation
 *
 * Author: Diana Prince (Team Alpha)
 * Responsibility: Re-validate privacy, detect PII
 */

package validation

import (
	"github.com/your-org/frustration-engine/internal/types"
)

// Disallowed metadata keys that indicate PII
var disallowedKeys = map[string]bool{
	"email":           true,
	"password":        true,
	"creditcard":      true,
	"credit_card":     true,
	"cardnumber":      true,
	"ssn":             true,
	"social_security": true,
	"phone":           true,
	"phonenumber":     true,
	"address":         true,
	"zipcode":         true,
	"postalcode":      true,
}

// ValidatePrivacy re-validates events for PII
func ValidatePrivacy(events []types.Event) ([]types.Event, []types.ValidationError) {
	validEvents := make([]types.Event, 0, len(events))
	errors := make([]types.ValidationError, 0)

	for i, event := range events {
		// Check metadata for disallowed keys
		if hasDisallowedKeys(event.Metadata) {
			errors = append(errors, types.ValidationError{
				Field:   "metadata",
				Reason:  "contains disallowed PII keys",
				EventID: i,
			})
			continue // Drop event with PII
		}

		// Check for PII patterns in string values
		if hasPIIPatterns(event) {
			errors = append(errors, types.ValidationError{
				Field:   "metadata",
				Reason:  "contains PII patterns",
				EventID: i,
			})
			continue // Drop event with PII
		}

		validEvents = append(validEvents, event)
	}

	return validEvents, errors
}

// hasDisallowedKeys checks if metadata contains disallowed keys
func hasDisallowedKeys(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}

	for key := range metadata {
		// Case-insensitive check
		lowerKey := toLower(key)
		if disallowedKeys[lowerKey] {
			return true
		}
	}

	return false
}

// hasPIIPatterns checks for PII patterns in event data
func hasPIIPatterns(event types.Event) bool {
	// Check metadata values for email patterns
	if metadata := event.Metadata; metadata != nil {
		for _, value := range metadata {
			if str, ok := value.(string); ok {
				if containsEmailPattern(str) || containsCreditCardPattern(str) {
					return true
				}
			}
		}
	}

	return false
}

// containsEmailPattern checks for email pattern
func containsEmailPattern(s string) bool {
	// Simple email pattern check
	for i := 0; i < len(s)-3; i++ {
		if s[i] == '@' {
			return true
		}
	}
	return false
}

// containsCreditCardPattern checks for credit card pattern
func containsCreditCardPattern(s string) bool {
	// Simple check for 4 groups of 4 digits
	digitCount := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digitCount++
		} else {
			digitCount = 0
		}
		if digitCount >= 13 {
			return true
		}
	}
	return false
}

// toLower converts string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}
