/**
 * Shared Configuration Utilities
 *
 * Author: Principal Engineer
 * Responsibility: Common configuration functions
 */

package config

import "os"

// GetEnv gets environment variable or returns default
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
