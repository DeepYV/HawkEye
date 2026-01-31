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

## Integration Guide

This section covers how to integrate HawkEye into your application. Choose the setup that matches your architecture.

### Prerequisites

- **Go 1.21+** for the backend services
- **Node.js 18+** and npm for the frontend SDK

### Step 1: Install the SDK

Since the SDK is not published to the public npm registry, install it locally:

```bash
# Clone the repository (private — requires access)
git clone https://github.com/DeepYV/HawkEye.git
cd HawkEye

# Build the SDK
cd sdk && npm install && npm run build

# Option A: Install from local path (in your frontend project)
cd /path/to/your-frontend-app
npm install /path/to/HawkEye/sdk

# Option B: Use npm link
cd /path/to/HawkEye/sdk && npm link
cd /path/to/your-frontend-app && npm link @hawkeye/observer-sdk
```

### Step 2: Choose Your Architecture

---

### Option A: Single Port (Everything on Port 3000)

Use this when your frontend and HawkEye backend run behind a **single port** — ideal for quick local development and testing. The dev server bundles all HawkEye services (Event Ingestion, Session Manager, UFSE, Incident Store) into one process with in-memory storage. No ClickHouse or PostgreSQL required.

**Start the dev server on port 3000:**

```bash
# Using the startup script
PORT=3000 ./scripts/start_dev.sh

# Or run directly
go run ./cmd/devserver --port 3000 --api-key dev-api-key --storage memory
```

**Initialize the SDK in your app (React example):**

```jsx
import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: 'dev-api-key',
      ingestionUrl: 'http://localhost:3000',  // same port as your app
      enableDebug: true,
    });

    return () => teardown();
  }, []);

  return <div>Your App</div>;
}
```

**Next.js example:**

```jsx
// pages/_app.js
import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function MyApp({ Component, pageProps }) {
  useEffect(() => {
    if (typeof window !== 'undefined') {
      initFrustrationObserver({
        apiKey: 'dev-api-key',
        ingestionUrl: 'http://localhost:3000',
      });
    }
    return () => teardown();
  }, []);

  return <Component {...pageProps} />;
}

export default MyApp;
```

**Vanilla JavaScript example:**

```html
<script type="module">
  import { initFrustrationObserver } from '@hawkeye/observer-sdk';

  initFrustrationObserver({
    apiKey: 'dev-api-key',
    ingestionUrl: 'http://localhost:3000',
  });
</script>
```

**Available endpoints (all on port 3000):**

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/v1/events` | Receives events from the SDK |
| `GET` | `/v1/incidents` | Query detected frustration incidents |
| `GET` | `/health` | Health check for all services |

---

### Option B: Two Ports (Frontend: 3000, Backend: 8080)

Use this when your frontend and backend run on **separate ports** — closer to a production-like setup. The frontend app runs on port 3000 (React, Next.js, etc.) and the HawkEye backend runs on port 8080.

**Terminal 1 — Start the HawkEye backend (port 8080):**

```bash
# Quick option: single-process dev server
./scripts/start_dev.sh
# Default port is 8080, no databases needed

# Or start all services individually for full pipeline:
export ENVIRONMENT=development
go run ./cmd/event-ingestion   # port 8080
go run ./cmd/session-manager   # port 8081 (in a separate terminal)
go run ./cmd/ufse              # port 8082 (in a separate terminal)
go run ./cmd/incident-store    # port 8084 (in a separate terminal)
```

**Terminal 2 — Start your frontend app (port 3000):**

```bash
# React
cd your-react-app && npm start        # runs on http://localhost:3000

# Next.js
cd your-nextjs-app && npm run dev     # runs on http://localhost:3000

# Vite
cd your-vite-app && npm run dev       # runs on http://localhost:5173
```

**Initialize the SDK pointing to the backend port:**

```jsx
import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: 'dev-api-key',
      ingestionUrl: 'http://localhost:8080',  // backend port, NOT your app port
      enableDebug: true,
    });

    return () => teardown();
  }, []);

  return <div>Your App</div>;
}
```

**Next.js with environment variables:**

```bash
# .env.local
NEXT_PUBLIC_HAWKEYE_API_KEY=dev-api-key
NEXT_PUBLIC_HAWKEYE_INGESTION_URL=http://localhost:8080
```

```jsx
// pages/_app.js
import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function MyApp({ Component, pageProps }) {
  useEffect(() => {
    if (typeof window !== 'undefined') {
      initFrustrationObserver({
        apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
        ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
      });
    }
    return () => teardown();
  }, []);

  return <Component {...pageProps} />;
}

export default MyApp;
```

**Vanilla JavaScript:**

```html
<script type="module">
  import { initFrustrationObserver } from '@hawkeye/observer-sdk';

  initFrustrationObserver({
    apiKey: 'dev-api-key',
    ingestionUrl: 'http://localhost:8080',
  });
</script>
```

**Service ports (two-port mode):**

| Service            | Default Port | Purpose                                                |
|--------------------|-------------|--------------------------------------------------------|
| Event Ingestion API | 8080        | Receives events from the frontend SDK (`POST /v1/events`) |
| Session Manager     | 8081        | Aggregates raw events into sessions                     |
| UFSE Engine         | 8082        | Detects frustration patterns and scores incidents       |
| Incident Store      | 8084        | Persists and queries incidents                          |
| **Your Frontend App** | **3000**  | React / Next.js (or any port your framework uses)       |

> The SDK appends `/v1/events` automatically, so you only need to provide the base URL.

---

### Step 3: Verify the Integration

Once both the backend and frontend are running, open your app in a browser and interact with it (click around, scroll, navigate between pages). You should see:

1. **In the browser console** (if `enableDebug: true`): log messages showing events being captured and sent.
2. **In the backend terminal**: log messages showing events being received at `/v1/events`.
3. **Health check**: Visit `http://localhost:<backend-port>/health` to confirm the backend is running.

### SDK Configuration Options

```javascript
initFrustrationObserver({
  apiKey: 'your-api-key',            // required — any string in development
  ingestionUrl: 'http://localhost:8080', // required — backend base URL
  enableDebug: true,                 // optional — logs events to console (default: false)
  batchSize: 10,                     // optional — events per batch (default: 10)
  batchInterval: 5000,              // optional — ms between batch sends (default: 5000)
  projectId: 'my-project',          // optional — project identifier
  appId: 'my-app',                  // optional — app identifier
  environment: 'development',       // optional — auto-detected from hostname if omitted
});
```

### SDK API Methods

| Method | Description |
|--------|-------------|
| `initFrustrationObserver(config)` | Start tracking user interactions |
| `teardown()` | Stop tracking and flush pending events |
| `captureEvent(type, target, metadata?)` | Manually track a custom event |
| `getSessionId()` | Get the current session ID |
| `flushEvents()` | Immediately send all pending events |

### Environment Variables

#### Backend

| Variable                | Default                  | Description                                          |
|-------------------------|--------------------------|------------------------------------------------------|
| `ENVIRONMENT`           | `production`             | Set to `development` to enable local CORS and debug  |
| `PORT`                  | `8080`                   | HTTP listen port                                     |
| `CLICKHOUSE_DSN`        | `localhost:9000`         | ClickHouse connection string (optional in dev)       |
| `SESSION_MANAGER_URL`   | `http://localhost:8081`  | Where to forward sessions                            |
| `UFSE_URL`              | `http://localhost:8082`  | UFSE endpoint (used by Session Manager)              |
| `INCIDENT_STORE_URL`    | `http://localhost:8084`  | Incident Store endpoint (used by UFSE)               |
| `DATABASE_URL`          | _(required in prod)_     | PostgreSQL DSN for Incident Store                    |
| `CORS_ALLOWED_ORIGINS`  | _(empty)_                | Additional allowed origins (comma-separated)         |

#### Frontend SDK (Next.js)

| Variable                                | Description                           |
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

### Quick-Start Checklist

1. Clone the repository and install Go 1.21+.
2. Build the SDK: `cd sdk && npm install && npm run build`.
3. Install the SDK in your frontend app: `npm install /path/to/HawkEye/sdk`.
4. Start the backend: `./scripts/start_dev.sh` (port 8080) or `PORT=3000 ./scripts/start_dev.sh` (port 3000).
5. Initialize the SDK with `ingestionUrl` pointing to the backend port.
6. Open your app in a browser and interact — events will flow to the backend.

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
