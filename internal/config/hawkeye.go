// Package config handles environment and flag-based configuration for HawkEye.
package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all configuration for the HawkEye server.
type Config struct {
	Port         string
	APIKey       string
	StorageMode  string // "memory" or "clickhouse"
	IncidentDSN  string // PostgreSQL DSN or "" for log-only
	Dev          bool   // development mode: memory storage, debug logging, wide CORS
	LogLevel     string // "debug", "info", "warn", "error"
}

// Load reads configuration from flags and environment variables.
// Flags take precedence over environment variables.
func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Port, "port", getEnv("PORT", "8080"), "Server port")
	flag.StringVar(&cfg.APIKey, "api-key", getEnv("TEST_API_KEY", "dev-api-key"), "API key for SDK authentication")
	flag.StringVar(&cfg.StorageMode, "storage", getEnv("HAWKEYE_STORAGE", "memory"), "Event storage: memory or clickhouse")
	flag.StringVar(&cfg.IncidentDSN, "incident-dsn", getEnv("INCIDENT_DSN", ""), "PostgreSQL DSN for incidents (empty = log-only)")
	flag.BoolVar(&cfg.Dev, "dev", getEnvBool("HAWKEYE_DEV", true), "Enable development mode")
	flag.StringVar(&cfg.LogLevel, "log-level", getEnv("LOG_LEVEL", "info"), "Log level: debug, info, warn, error")
	flag.Parse()

	return cfg
}

// Print prints the configuration summary to stdout.
func (c *Config) Print() {
	incidentMode := "log-only"
	if c.IncidentDSN != "" {
		incidentMode = "postgresql"
	}

	fmt.Println("=============================================================")
	fmt.Println("  HawkEye Frustration Detection Engine")
	fmt.Println("=============================================================")
	fmt.Printf("  Port:          %s\n", c.Port)
	fmt.Printf("  API Key:       %s\n", c.APIKey)
	fmt.Printf("  Event Storage: %s\n", c.StorageMode)
	fmt.Printf("  Incidents:     %s\n", incidentMode)
	fmt.Printf("  Dev Mode:      %v\n", c.Dev)
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  Endpoints:")
	fmt.Printf("    POST http://localhost:%s/v1/events       (event ingestion)\n", c.Port)
	fmt.Printf("    GET  http://localhost:%s/v1/incidents     (query incidents)\n", c.Port)
	fmt.Printf("    GET  http://localhost:%s/health           (health check)\n", c.Port)
	fmt.Printf("    GET  http://localhost:%s/metrics          (prometheus)\n", c.Port)
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("  SDK config:")
	fmt.Printf("    ingestionUrl: 'http://localhost:%s'\n", c.Port)
	fmt.Printf("    apiKey:       '%s'\n", c.APIKey)
	fmt.Println("=============================================================")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	switch v {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}
