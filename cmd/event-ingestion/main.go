/**
 * Event Ingestion API - Main Entry Point
 *
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Application bootstrap
 */

package main

import (
	"flag"
	"log"
	"strings"

	"github.com/your-org/frustration-engine/internal/config"
	"github.com/your-org/frustration-engine/internal/server"
)

func main() {
	// Configuration from environment or flags
	port := flag.String("port", config.GetEnv("PORT", "8080"), "Server port")
	clickhouseDSN := flag.String("clickhouse", config.GetEnv("CLICKHOUSE_DSN", "localhost:9000"), "ClickHouse DSN")
	sessionManagerURL := flag.String("session-manager", config.GetEnv("SESSION_MANAGER_URL", "http://localhost:8081"), "Session Manager URL")
	rateLimitRPS := flag.Int("rate-limit-rps", 1000, "Rate limit requests per second")
	rateLimitBurst := flag.Int("rate-limit-burst", 2000, "Rate limit burst capacity")
	environment := flag.String("environment", config.GetEnv("ENVIRONMENT", "production"), "Environment (development/staging/production)")

	flag.Parse()

	// Parse CORS allowed origins from environment
	corsOrigins := []string{}
	if originsEnv := config.GetEnv("CORS_ALLOWED_ORIGINS", ""); originsEnv != "" {
		for _, origin := range strings.Split(originsEnv, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				corsOrigins = append(corsOrigins, origin)
			}
		}
	}

	// Create server
	cfg := server.Config{
		Port:               *port,
		ClickHouseDSN:      *clickhouseDSN,
		SessionManagerURL:  *sessionManagerURL,
		RateLimitRPS:       *rateLimitRPS,
		RateLimitBurst:     *rateLimitBurst,
		Environment:        *environment,
		CORSAllowedOrigins: corsOrigins,
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	log.Printf("Starting Event Ingestion API on port %s", *port)
	if err := srv.Start(*port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
