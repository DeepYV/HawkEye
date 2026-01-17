/**
 * UFSE Forwarding
 * 
 * Author: Henry Wilson (Team Beta)
 * Responsibility: Forward completed sessions to UFSE
 */

package forwarding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/your-org/frustration-engine/internal/types"
)

// Forwarder forwards sessions to UFSE
type Forwarder struct {
	ufseURL string
	client  *http.Client
}

// NewForwarder creates a new forwarder
func NewForwarder(ufseURL string) *Forwarder {
	return &Forwarder{
		ufseURL: ufseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ForwardSession forwards a session to UFSE
func (f *Forwarder) ForwardSession(ctx context.Context, session *types.Session) error {
	if f.ufseURL == "" {
		// Log and skip if URL not configured
		log.Printf("[Session Manager] UFSE URL not configured, logging session instead")
		log.Printf("[Session Manager] Session: %s (project: %s, events: %d)", session.SessionID, session.ProjectID, len(session.Events))
		return nil
	}

	// Serialize session
	data, err := json.Marshal(session)
	if err != nil {
		log.Printf("[Session Manager] Failed to marshal session: %v", err)
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// POST to UFSE
	url := f.ufseURL + "/v1/sessions/process"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[Session Manager] Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		log.Printf("[Session Manager] Failed to forward to UFSE (session: %s): %v", session.SessionID, err)
		log.Printf("[Session Manager] Session logged: %s (project: %s, events: %d)", session.SessionID, session.ProjectID, len(session.Events))
		return fmt.Errorf("failed to forward session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Session Manager] UFSE returned status %d for session %s", resp.StatusCode, session.SessionID)
		return fmt.Errorf("UFSE returned status %d", resp.StatusCode)
	}

	// Read response to check for incidents
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if count, ok := result["count"].(float64); ok {
			log.Printf("[Session Manager] Successfully forwarded session %s to UFSE, %d incidents detected", session.SessionID, int(count))
		}
	}

	return nil
}