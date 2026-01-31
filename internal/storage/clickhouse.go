/**
 * ClickHouse Storage
 *
 * Author: Grace Lee (Team Beta)
 * Responsibility: Event persistence to ClickHouse
 */

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/your-org/frustration-engine/internal/config"
	"github.com/your-org/frustration-engine/internal/types"
)

// Storage handles event persistence
type Storage struct {
	conn    driver.Conn
	logOnly bool
}

// NewStorage creates a new storage instance
func NewStorage(ctx context.Context, dsn string) (*Storage, error) {
	// Support log-only mode for testing
	if dsn == "log-only" {
		log.Printf("[ClickHouse Storage] Running in LOG-ONLY mode - events will be logged, not stored")
		return &Storage{
			logOnly: true,
		}, nil
	}

	// Parse DSN - handle both "clickhouse://host:port" and "host:port" formats
	addr := dsn
	if len(dsn) > 12 && dsn[:12] == "clickhouse://" {
		// Remove "clickhouse://" prefix
		addr = dsn[12:]
	}

	// Get password from environment if set, otherwise use empty (default ClickHouse setup)
	password := config.GetEnv("CLICKHOUSE_PASSWORD", "")
	username := config.GetEnv("CLICKHOUSE_USERNAME", "default")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "events",
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Create table if not exists
	if err := createTable(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Storage{conn: conn, logOnly: false}, nil
}

// StoreEvents stores events in ClickHouse
func (s *Storage) StoreEvents(ctx context.Context, projectID string, events []types.Event) error {
	if s.logOnly {
		log.Printf("[ClickHouse Storage] LOG-ONLY: Would store %d events for project %s", len(events), projectID)
		for i, event := range events {
			if i < 3 { // Log first 3 events as sample
				log.Printf("[ClickHouse Storage] Sample event: type=%s, session=%s, route=%s",
					event.EventType, event.SessionID, event.Route)
			}
		}
		return nil
	}
	if len(events) == 0 {
		return nil
	}

	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO events")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, event := range events {
		// Serialize metadata to JSON string
		metadataJSON := "{}"
		if event.Metadata != nil {
			metadataBytes, err := json.Marshal(event.Metadata)
			if err != nil {
				// If serialization fails, use empty JSON object
				metadataJSON = "{}"
			} else {
				metadataJSON = string(metadataBytes)
			}
		}

		if err := batch.Append(
			uuid.New().String(),   // id
			projectID,             // project_id
			event.EventType,       // event_type
			event.SessionID,       // session_id
			event.Route,           // route
			event.Target.Type,     // target_type
			event.Target.ID,       // target_id
			event.Target.Selector, // target_selector
			metadataJSON,          // metadata (JSON string)
			time.Now(),            // received_at
			time.Now(),            // processed_at
		); err != nil {
			return fmt.Errorf("failed to append event: %w", err)
		}
	}

	return batch.Send()
}

// Close closes the ClickHouse connection.
func (s *Storage) Close() error {
	if s.logOnly || s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

// createTable creates the events table if it doesn't exist
func createTable(ctx context.Context, conn driver.Conn) error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id String,
		project_id String,
		event_type String,
		session_id String,
		route String,
		target_type String,
		target_id String,
		target_selector String,
		metadata String,
		received_at DateTime,
		processed_at DateTime
	) ENGINE = MergeTree()
	ORDER BY (project_id, session_id, received_at)
	PARTITION BY toYYYYMM(received_at)
	TTL received_at + INTERVAL 30 DAY
	`
	return conn.Exec(ctx, query)
}
