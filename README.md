# HawkEye - Real-Time Frustration Detection Engine

HawkEye detects user frustration in real time by analysing behavioural signals — rage clicks, navigation loops, dead ends, error cascades, and scroll thrashing — so product teams can fix UX issues before they impact retention.

## Architecture

```text
Browser (SDK)  ──▶  HawkEye Server  ──▶  Incidents
                    ┌──────────────┐
  events ──▶        │   Ingest     │
                    │   Session    │
                    │   Engine     │  ◀── pure detection logic
                    │   Incident   │
                    └──────────────┘
                    Single binary,
                    single port.
```

One process. One port. Well-isolated modules:

| Module | Purpose |
|--------|---------|
| `internal/ingest` | Validate & store SDK events |
| `internal/session` | Aggregate events into sessions |
| `internal/engine` | Frustration detection (pure functions, no I/O) |
| `internal/incident` | Persist & query detected incidents |
| `internal/http` | HTTP server & routing |
| `internal/metrics` | Prometheus instrumentation |
| `internal/storage` | Pluggable storage (memory / ClickHouse / PostgreSQL) |
| `internal/app` | Wires everything together |
| `pkg/types` | Shared domain types (SDK-compatible) |

## Quick Start

### 1. Start the backend

```bash
go run ./cmd/hawkeye --dev
```

This starts HawkEye in dev mode: in-memory storage, no database dependencies, CORS wide open, debug logging.

**Endpoints:**

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/v1/events` | Event ingestion (requires API key) |
| `GET` | `/v1/incidents` | Query detected incidents |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |

### 2. Install the SDK

```bash
npm install @hawkeye/observer-sdk
```

### 3. Add to your app

```javascript
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

initFrustrationObserver({
  apiKey: 'dev-api-key',
  ingestionUrl: 'http://localhost:8080',
});
```

That's it. HawkEye captures clicks, scrolls, errors, navigation and form interactions automatically.

## Local Development

```bash
# Start with defaults (port 8080, in-memory, api key = dev-api-key)
go run ./cmd/hawkeye --dev

# Custom port and API key
go run ./cmd/hawkeye --dev --port 3001 --api-key my-secret

# Run tests
go test ./...
```

### Verify it works

```bash
# Health check
curl http://localhost:8080/health

# Send a test event
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-api-key" \
  -d '{"events":[{"eventType":"click","timestamp":"2024-01-01T00:00:00Z","sessionId":"test","route":"/home","target":{"type":"button","id":"cta"}}]}'

# Query incidents
curl http://localhost:8080/v1/incidents

# Prometheus metrics
curl http://localhost:8080/metrics
```

## Production Deployment

For production, configure external storage:

```bash
go run ./cmd/hawkeye \
  --port 8080 \
  --api-key "$HAWKEYE_API_KEY" \
  --storage clickhouse \
  --incident-dsn "postgres://user:pass@host:5432/hawkeye?sslmode=require"
```

The legacy multi-service deployment (separate binaries for event-ingestion, session-manager, ufse, incident-store) is still available under `cmd/` for backward compatibility.

## Observability

HawkEye exposes Prometheus metrics at `/metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `hawkeye_events_ingested_total` | counter | Events received from SDK |
| `hawkeye_sessions_created_total` | counter | Sessions created |
| `hawkeye_sessions_processed_total` | counter | Sessions processed by engine |
| `hawkeye_incidents_detected_total` | counter | Frustration incidents detected |
| `hawkeye_processing_latency_seconds` | histogram | Session processing latency |
| `hawkeye_event_queue_depth` | gauge | Event processing queue depth |
| `hawkeye_http_requests_total` | counter | HTTP requests by method/path/status |
| `hawkeye_http_request_duration_seconds` | histogram | HTTP request duration |

### Logs

Structured logs with correlation via `session_id`. Zero business logic in logs.

### Traces

OpenTelemetry-ready (disabled by default, enable via environment).

## Configuration

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--port` | `PORT` | `8080` | Server port |
| `--api-key` | `TEST_API_KEY` | `dev-api-key` | API key for SDK auth |
| `--storage` | `HAWKEYE_STORAGE` | `memory` | Event storage backend |
| `--incident-dsn` | `INCIDENT_DSN` | `` (log-only) | PostgreSQL DSN for incidents |
| `--dev` | `HAWKEYE_DEV` | `true` | Dev mode (memory, debug, wide CORS) |
| `--log-level` | `LOG_LEVEL` | `info` | Log level |

## SDK Integration Examples

See `examples/` for framework-specific guides:

- [React](examples/react/) — custom hook + component
- [Next.js](examples/nextjs/) — SSR-safe initialization
- [Vanilla JS](examples/vanilla-js/) — plain HTML/JS

## Integration with Other Applications

HawkEye is designed to integrate in three common ways depending on your architecture.

### 1) Frontend app integration (browser SDK)

Use this for React, Next.js, Vue, Angular, or plain JS apps.

```javascript
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

initFrustrationObserver({
  apiKey: process.env.HAWKEYE_API_KEY,
  ingestionUrl: 'https://hawkeye.your-company.com',
});
```

The SDK sends behavioral events to `POST /v1/events`, and your app can query incidents from `GET /v1/incidents` (typically via your backend).

### 2) Backend-to-backend integration (no browser SDK)

If you already collect product analytics events, forward selected events server-side to HawkEye:

```bash
curl -X POST https://hawkeye.your-company.com/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $HAWKEYE_API_KEY" \
  -d '{
    "events": [
      {
        "eventType": "error",
        "timestamp": "2024-01-01T00:00:00Z",
        "sessionId": "sess-123",
        "route": "/checkout",
        "metadata": {"error": "payment_failed"}
      }
    ]
  }'
```

This is useful for mobile backends, API gateways, and event pipelines (Kafka/Segment/ETL workers).

### 3) Incident export to ticketing/ops systems

HawkEye supports adapters and exporter components that can route detected incidents to downstream tools.

- Adapter interfaces: `internal/adapters/interface.go`
- Built-in adapters: JIRA and Linear (`internal/adapters/jira.go`, `internal/adapters/linear.go`)
- Export engine/scheduler: `internal/exporter/engine.go`, `internal/exporter/scheduler.go`

Typical pattern:
1. HawkEye detects incidents.
2. Exporter applies eligibility/priority/rate-limit rules.
3. Adapter creates/updates external tickets.

### Recommended production integration pattern

- Run HawkEye behind your API gateway at a stable internal domain.
- Keep API keys in secrets manager, not source code.
- Use ClickHouse/PostgreSQL backends in production.
- Let your application backend own incident read APIs for tenant-aware access control.

## Storage Abstraction

```go
// Event persistence
type EventStore interface {
    StoreEvents(ctx context.Context, projectID string, events []types.Event) error
    Close() error
}

// Incident persistence
type IncidentStore interface {
    Save(ctx context.Context, incident types.Incident) error
    Query(ctx context.Context, filter types.Filter) ([]types.Incident, error)
    Close() error
}
```

Implementations: `memory` (dev & tests), `clickhouse` (production events), `postgresql` (production incidents).

No storage logic inside business rules.

## Testing

```bash
# All tests
go test ./...

# Engine unit tests (pure functions)
go test ./internal/engine/ -v

# Storage tests
go test ./internal/storage/... -v

# Integration tests (HTTP endpoints)
go test ./internal/app/ -v
```

## Contributing

1. Fork and create a feature branch
2. Keep files under ~200 lines
3. No deep nesting, no reflection, no hidden magic
4. Comments explain *why*, not *what*
5. Run `go test ./...` before submitting

## License

MIT
