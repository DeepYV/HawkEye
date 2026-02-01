// Package app wires together all HawkEye modules into a runnable application.
//
// This is the composition root — the only place where concrete types are chosen
// and connected. All other packages depend only on interfaces.
package app

import (
	"context"
	"log"

	"github.com/your-org/frustration-engine/internal/config"
	"github.com/your-org/frustration-engine/internal/engine"
	hawkhttp "github.com/your-org/frustration-engine/internal/http"
	"github.com/your-org/frustration-engine/internal/incident"
	"github.com/your-org/frustration-engine/internal/ingest"
	"github.com/your-org/frustration-engine/internal/metrics"
	"github.com/your-org/frustration-engine/internal/session"
	memstorage "github.com/your-org/frustration-engine/internal/storage/memory"
	oldtypes "github.com/your-org/frustration-engine/internal/types"
	"github.com/your-org/frustration-engine/pkg/types"
)

// App is the fully wired HawkEye application.
type App struct {
	Server         *hawkhttp.Server
	SessionManager *session.Manager
	IncidentSvc    *incident.Service
	cfg            *config.Config
	cancel         context.CancelFunc
}

// New builds the application from configuration.
func New(cfg *config.Config) *App {
	eventStore := memstorage.New()
	incidentStore := memstorage.NewIncidentStore()
	sessionMgr := session.NewManager()
	incidentSvc := incident.NewService(incidentStore)
	ingestHandler := ingest.NewHandler(eventStore, sessionMgr)
	server := hawkhttp.NewServer(ingestHandler, incidentSvc, cfg.APIKey, cfg.Dev)

	return &App{
		Server:         server,
		SessionManager: sessionMgr,
		IncidentSvc:    incidentSvc,
		cfg:            cfg,
	}
}

// Start begins background processing (session manager, engine pipeline).
func (a *App) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	a.SessionManager.Start(ctx)

	// Session → Engine → Incident Store pipeline
	go func() {
		ch := a.SessionManager.GetEmissionChannel()
		for {
			select {
			case <-ctx.Done():
				return
			case sess := <-ch:
				if sess == nil {
					continue
				}
				newSess := convertSession(sess)
				log.Printf("[app] processing session %s through engine", newSess.SessionID)
				metrics.EventQueueDepth.Set(0)

				incidents := engine.DetectFrustration(newSess)
				for _, inc := range incidents {
					if err := a.IncidentSvc.Store(ctx, *inc); err != nil {
						log.Printf("[app] failed to store incident: %v", err)
					} else {
						log.Printf("[app] incident stored: %s (score: %d, confidence: %s)",
							inc.IncidentID, inc.FrustrationScore, inc.ConfidenceLevel)
					}
				}
			}
		}
	}()
}

// Stop shuts down the application gracefully.
func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	a.SessionManager.Stop()
}

// convertSession bridges the old internal/types.Session to pkg/types.Session.
func convertSession(old *oldtypes.Session) types.Session {
	events := make([]types.Event, len(old.Events))
	for i, e := range old.Events {
		events[i] = types.Event{
			EventType:      e.EventType,
			Timestamp:      e.Timestamp,
			SessionID:      e.SessionID,
			Route:          e.Route,
			Target:         types.EventTarget(e.Target),
			Metadata:       e.Metadata,
			Environment:    e.Environment,
			IdempotencyKey: e.IdempotencyKey,
		}
	}

	transitions := make([]types.RouteTransition, len(old.RouteTransitions))
	for i, rt := range old.RouteTransitions {
		transitions[i] = types.RouteTransition{
			From:      rt.From,
			To:        rt.To,
			Timestamp: rt.Timestamp,
		}
	}

	return types.Session{
		SessionID:        old.SessionID,
		ProjectID:        old.ProjectID,
		State:            old.State,
		Events:           events,
		StartTime:        old.StartTime,
		EndTime:          old.EndTime,
		LastActivity:     old.LastActivity,
		RouteTransitions: transitions,
		Metadata:         old.Metadata,
	}
}
