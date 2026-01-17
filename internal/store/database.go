/**
 * Database Operations
 * 
 * Author: Bob Williams (Team Alpha)
 * Responsibility: Database connection and operations
 */

package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Store represents the incident store
type Store struct {
	DB *sql.DB
	db *sql.DB // Keep for backward compatibility
}

// NewStore creates a new store instance
func NewStore(dsn string) (*Store, error) {
	// If DSN is empty or "log-only", use logging mode
	if dsn == "" || dsn == "log-only" {
		fmt.Println("[Incident Store] Running in LOG-ONLY mode (no database connection)")
		return &Store{DB: nil, db: nil}, nil
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("[Incident Store] Failed to open database, falling back to LOG-ONLY mode: %v\n", err)
		return &Store{DB: nil, db: nil}, nil
	}

	// Test connection
	if err := db.Ping(); err != nil {
		fmt.Printf("[Incident Store] Failed to ping database, falling back to LOG-ONLY mode: %v\n", err)
		return &Store{DB: nil, db: nil}, nil
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	store := &Store{DB: db, db: db}

	// Initialize schema
	if err := store.InitializeSchema(); err != nil {
		fmt.Printf("[Incident Store] Failed to initialize schema, falling back to LOG-ONLY mode: %v\n", err)
		return &Store{DB: nil, db: nil}, nil
	}

	fmt.Println("[Incident Store] Connected to PostgreSQL successfully")
	return store, nil
}

// InitializeSchema initializes the database schema
func (s *Store) InitializeSchema() error {
	if s.db == nil {
		return nil // Log-only mode
	}

	if _, err := s.db.Exec(CreateIncidentsTable); err != nil {
		return fmt.Errorf("failed to create incidents table: %w", err)
	}

	if _, err := s.db.Exec(CreateIndexes); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db == nil {
		return nil // Log-only mode
	}
	return s.db.Close()
}