# HawkEye - Frustration Detection Engine

HawkEye is a real-time user frustration detection system designed to help product teams identify and fix UX issues before they impact user retention.

## What is HawkEye?

HawkEye analyzes user behavior patterns to detect frustration signals like:
- **Rage clicks**: Rapid, repeated clicks indicating confusion
- **Error cascades**: Multiple errors in quick succession
- **Navigation loops**: Users going back and forth between pages
- **Dead ends**: Users getting stuck without clear next steps
- **Scroll thrashing**: Excessive scrolling suggesting users can't find what they need

## Quick Start

### 1. Install the SDK

```bash
npm install @hawkeye/observer-sdk
```

### 2. Initialize in Your App

```javascript
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

initFrustrationObserver({
  apiKey: 'your-api-key-here',
  ingestionUrl: 'https://your-hawkeye-instance.com',
});
```

### 3. Deploy the Backend

```bash
# Clone the repository
git clone https://github.com/DeepYV/HawkEye.git
cd HawkEye

# Build and run
go build -o hawkeye ./cmd/server
./hawkeye
```

## Components

### 🎯 Observer SDK (JavaScript)

The frontend SDK that captures user interaction signals.

- **Lightweight**: ~15KB gzipped
- **Privacy-focused**: No PII, no content capture
- **Framework-agnostic**: Works with React, Vue, Next.js, vanilla JS
- **Zero-config**: Sensible defaults out of the box

[Read SDK Documentation →](./sdk/README.md)

### 🔧 Frustration Engine (Go)

The backend engine that processes signals and detects frustration patterns.

- **Real-time processing**: Instant frustration detection
- **Scalable**: Handles high-volume event streams
- **Extensible**: Plugin-based architecture
- **Observable**: Built-in metrics and tracing

[Read Backend Documentation →](./docs/backend.md) _(coming soon)_

## Integration Examples

### React

```jsx
import { useEffect } from 'react';
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: 'your-api-key',
      ingestionUrl: 'https://api.hawkeye.example.com',
    });
  }, []);

  return <div>Your App</div>;
}
```

[Full React Example →](./examples/react/README.md)

### Next.js

```jsx
// pages/_app.js
function MyApp({ Component, pageProps }) {
  useEffect(() => {
    if (typeof window !== 'undefined') {
      initFrustrationObserver({
        apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
        ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
      });
    }
  }, []);

  return <Component {...pageProps} />;
}
```

[Full Next.js Example →](./examples/nextjs/README.md)

### Vanilla JavaScript

```html
<script type="module">
  import { initFrustrationObserver } from 'https://unpkg.com/@hawkeye/observer-sdk';

  initFrustrationObserver({
    apiKey: 'your-api-key',
    ingestionUrl: 'https://api.hawkeye.example.com',
  });
</script>
```

[Full Vanilla JS Example →](./examples/vanilla-js/README.md)

## Local Development Setup

This section explains how to run and integrate HawkEye locally when the frontend and backend run on separate ports.

### Prerequisites

- **Go 1.21+** for the backend services
- **Node.js 18+** and npm for the frontend SDK
- **ClickHouse** for event storage (optional — the ingestion service logs events if unavailable)
- **PostgreSQL** for incident storage (optional — use `log-only` mode for testing without a database)

### Service Ports

| Service            | Default Port | Purpose                                                |
|--------------------|-------------|--------------------------------------------------------|
| Event Ingestion API | 8080        | Receives events from the frontend SDK (`POST /v1/events`) |
| Session Manager     | 8081        | Aggregates raw events into sessions                     |
| UFSE Engine         | 8082        | Detects frustration patterns and scores incidents       |
| Incident Store      | 8084        | Persists and queries incidents                          |
| **Your Frontend App** | **3000**  | React / Next.js (or any port your framework uses)       |

### Starting the Backend

```bash
# Clone the repository (private — requires access)
git clone https://github.com/DeepYV/HawkEye.git
cd HawkEye

# Set development environment
export ENVIRONMENT=development

# Start each service (in separate terminals or via a process manager)
go run ./cmd/event-ingestion
go run ./cmd/session-manager
go run ./cmd/ufse
go run ./cmd/incident-store
```

Each service reads its port from the `PORT` environment variable and falls back to the defaults listed above.

### Connecting the Frontend SDK

Point the SDK's `ingestionUrl` at the local Event Ingestion API:

```javascript
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

initFrustrationObserver({
  apiKey: 'test-key',                        // any string works in development
  ingestionUrl: 'http://localhost:8080',      // Event Ingestion API
  enableDebug: true,                         // enable console logging
});
```

The SDK appends `/v1/events` automatically, so you only need to provide the base URL.

For **Next.js**, use environment variables so the URL is easy to change per environment:

```bash
# .env.local
NEXT_PUBLIC_HAWKEYE_API_KEY=test-key
NEXT_PUBLIC_HAWKEYE_INGESTION_URL=http://localhost:8080
```

```javascript
initFrustrationObserver({
  apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
  ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
});
```

### Environment Variables

#### Backend (Event Ingestion)

| Variable                | Default                  | Description                                          |
|-------------------------|--------------------------|------------------------------------------------------|
| `ENVIRONMENT`           | `production`             | Set to `development` to enable local CORS and debug  |
| `PORT`                  | `8080`                   | HTTP listen port                                     |
| `CLICKHOUSE_DSN`        | `localhost:9000`         | ClickHouse connection string                         |
| `SESSION_MANAGER_URL`   | `http://localhost:8081`  | Where to forward sessions                            |
| `CORS_ALLOWED_ORIGINS`  | _(empty)_                | Additional allowed origins (comma-separated)         |

#### Backend (Session Manager / UFSE / Incident Store)

| Variable              | Default                  | Description                              |
|-----------------------|--------------------------|------------------------------------------|
| `PORT`                | `8081` / `8082` / `8084` | HTTP listen port (per service)           |
| `UFSE_URL`            | `http://localhost:8082`  | UFSE endpoint (used by Session Manager)  |
| `INCIDENT_STORE_URL`  | `http://localhost:8084`  | Incident Store endpoint (used by UFSE)   |
| `DATABASE_URL`        | _(required in prod)_     | PostgreSQL DSN for Incident Store        |

#### Frontend SDK

| Variable (Next.js example)              | Description                           |
|-----------------------------------------|---------------------------------------|
| `NEXT_PUBLIC_HAWKEYE_API_KEY`           | API key sent with every event batch   |
| `NEXT_PUBLIC_HAWKEYE_INGESTION_URL`     | Base URL of the Event Ingestion API   |

### CORS in Local Development

When `ENVIRONMENT` is set to `development` (or `dev` / `local`), the backend automatically allows requests from common localhost origins:

- `http://localhost:3000` (React)
- `http://localhost:5173` (Vite)
- `http://localhost:8080` (Vue)
- `http://localhost:4200` (Angular)
- `http://127.0.0.1:*`

If your frontend runs on a different port, either:
1. Set `CORS_ALLOWED_ORIGINS=http://localhost:<port>` when starting the backend, or
2. Rely on the automatic localhost detection — any `localhost` or `127.0.0.1` origin is accepted in development mode.

In **production**, set `ENVIRONMENT=production` and list your origins explicitly via `CORS_ALLOWED_ORIGINS`.

### Private Repository Considerations

- The SDK is not published to the public npm registry. Install it locally from the `sdk/` directory:
  ```bash
  cd sdk && npm install && npm run build
  # Then in your frontend project:
  npm install ../HawkEye/sdk
  ```
  Or use `npm link`:
  ```bash
  cd HawkEye/sdk && npm link
  cd your-frontend-app && npm link @hawkeye/observer-sdk
  ```
- Backend services are built from source — no external binary distribution is needed.
- All inter-service communication happens over `localhost` by default, so no external network access is required during development.

### Quick-Start Checklist

1. Clone the repository and install Go 1.21+.
2. Set `export ENVIRONMENT=development`.
3. Start the Event Ingestion API: `go run ./cmd/event-ingestion`.
4. In your frontend app, initialize the SDK with `ingestionUrl: 'http://localhost:8080'`.
5. Open your app in a browser and interact with it — events will flow to the backend.
6. (Optional) Start the remaining services to enable full frustration detection.

## Architecture

```
┌──────────────────┐
│   Your Web App   │
│  (React/Next.js) │
└────────┬─────────┘
         │
         │ (Observer SDK captures signals)
         │
         ▼
┌──────────────────┐
│ Ingestion API    │
│ /v1/events       │
└────────┬─────────┘
         │
         │ (Events batched and queued)
         │
         ▼
┌──────────────────┐
│ Frustration      │
│ Engine (UFSE)    │
│ - Pattern Match  │
│ - Scoring        │
│ - Detection      │
└────────┬─────────┘
         │
         │ (Alerts and insights)
         │
         ▼
┌──────────────────┐
│ Webhooks / Alerts│
│ Dashboard / API  │
└──────────────────┘
```

## Features

### Privacy-First Design

- **No PII collection**: Only behavioral patterns
- **No content capture**: Input values never recorded
- **GDPR compliant**: Transparent data handling
- **User control**: Easy opt-out mechanism

### Production-Ready

- **Safe defaults**: No false alarms out of the box
- **Graceful degradation**: Fails silently without breaking your app
- **Low overhead**: Minimal performance impact
- **Battle-tested**: Used in production environments

### Developer-Friendly

- **10-minute setup**: Get started quickly
- **Framework examples**: React, Next.js, Vanilla JS
- **TypeScript support**: Full type definitions included
- **Comprehensive docs**: Clear guides and API reference

## Use Cases

### Product Teams

Identify where users struggle and prioritize fixes based on real frustration data.

### Engineering Teams

Detect errors and UX issues before they escalate into support tickets.

### Design Teams

Validate designs with real user behavior data, not just assumptions.

## Documentation

- [SDK Documentation](./sdk/README.md) - Complete SDK API reference
- [React Integration](./examples/react/README.md) - React setup guide
- [Next.js Integration](./examples/nextjs/README.md) - Next.js with SSR
- [Vanilla JS Integration](./examples/vanilla-js/README.md) - Plain JavaScript

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Support

- **GitHub Issues**: For bug reports and feature requests
- **Discussions**: For questions and community support
- **Security**: Report security issues to security@hawkeye.example.com

---

**Built with ❤️ for better user experiences**
