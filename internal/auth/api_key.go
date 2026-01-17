/**
 * API Key Management
 *
 * Author: Charlie Brown (Team Alpha)
 * Responsibility: API key authentication
 */

package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"sync"
)

var (
	ErrInvalidAPIKey = errors.New("invalid API key")
	ErrMissingAPIKey = errors.New("missing API key")
)

// APIKey represents an API key with project association
type APIKey struct {
	Key       string
	ProjectID string
	Active    bool
}

// Store manages API keys (in-memory for now, should be DB-backed in production)
type Store struct {
	mu   sync.RWMutex
	keys map[string]*APIKey
}

// NewStore creates a new API key store
func NewStore() *Store {
	return &Store{
		keys: make(map[string]*APIKey),
	}
}

// ValidateAPIKey validates an API key and returns project ID
func (s *Store) ValidateAPIKey(ctx context.Context, apiKey string) (string, error) {
	if apiKey == "" {
		return "", ErrMissingAPIKey
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keys[apiKey]
	if !exists {
		// Use constant-time comparison to prevent timing attacks
		_ = subtle.ConstantTimeCompare([]byte(apiKey), []byte("invalid"))
		return "", ErrInvalidAPIKey
	}

	if !key.Active {
		return "", ErrInvalidAPIKey
	}

	return key.ProjectID, nil
}

// AddAPIKey adds an API key to the store (for testing/admin)
func (s *Store) AddAPIKey(apiKey, projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[apiKey] = &APIKey{
		Key:       apiKey,
		ProjectID: projectID,
		Active:    true,
	}
}

// LoadAPIKeysFromDB loads API keys from database (placeholder)
// In production, this would query PostgreSQL
func (s *Store) LoadAPIKeysFromDB(ctx context.Context) error {
	// TODO: Load from PostgreSQL
	// For now, we'll use in-memory store
	return nil
}
