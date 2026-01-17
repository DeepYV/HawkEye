# Integration Guide

This guide explains how to integrate the Frustration Engine into your applications.

## Table of Contents

1. [Frontend SDK Integration](#frontend-sdk-integration)
2. [Backend Services Setup](#backend-services-setup)
3. [API Configuration](#api-configuration)
4. [Framework-Specific Examples](#framework-specific-examples)
5. [Testing Your Integration](#testing-your-integration)

---

## Frontend SDK Integration

### Overview

The Frontend Observer SDK captures user interaction events and sends them to the Event Ingestion API. It's designed to be:
- **Lightweight**: <10KB gzipped
- **Non-intrusive**: Zero performance impact
- **Privacy-safe**: Automatically masks sensitive data
- **Framework-agnostic**: Works with React, Next.js, Vue, Angular, or vanilla JS

### Quick Start

#### 1. Install the SDK

```bash
npm install @frustration-engine/observer-sdk
# or
yarn add @frustration-engine/observer-sdk
```

#### 2. Initialize in Your App

**React / Next.js:**

```typescript
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

// In your app entry point (App.tsx, _app.tsx, etc.)
useEffect(() => {
  initFrustrationObserver({
    apiKey: process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY,
    ingestionUrl: process.env.NEXT_PUBLIC_FRUSTRATION_INGESTION_URL || 'https://api.yourdomain.com/v1/events',
  });
}, []);
```

**Vanilla JavaScript:**

```html
<script src="https://cdn.yourdomain.com/frustration-observer.min.js"></script>
<script>
  FrustrationObserver.init({
    apiKey: 'your-api-key-here',
    ingestionUrl: 'https://api.yourdomain.com/v1/events',
  });
</script>
```

**Vue.js:**

```javascript
// main.js or App.vue
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default {
  mounted() {
    initFrustrationObserver({
      apiKey: process.env.VUE_APP_FRUSTRATION_API_KEY,
      ingestionUrl: process.env.VUE_APP_FRUSTRATION_INGESTION_URL,
    });
  },
};
```

**Angular:**

```typescript
// app.component.ts
import { Component, OnInit } from '@angular/core';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
})
export class AppComponent implements OnInit {
  ngOnInit() {
    initFrustrationObserver({
      apiKey: environment.frustrationApiKey,
      ingestionUrl: environment.frustrationIngestionUrl,
    });
  }
}
```

### Configuration Options

```typescript
interface ObserverConfig {
  // Required
  apiKey: string;
  
  // Optional
  ingestionUrl?: string; // Default: https://api.yourdomain.com/v1/events
  batchSize?: number; // Default: 10
  batchInterval?: number; // Default: 5000ms
  enableDebug?: boolean; // Default: false
  
  // Privacy options
  maskInputs?: boolean; // Default: true
  sensitiveSelectors?: string[]; // Additional selectors to mask
  
  // Event filtering
  ignoredRoutes?: string[]; // Routes to ignore
  ignoredSelectors?: string[]; // DOM selectors to ignore
}
```

### Environment Variables

Create a `.env` file (or equivalent):

```bash
# Frontend (.env.local or .env)
NEXT_PUBLIC_FRUSTRATION_API_KEY=your-api-key-here
NEXT_PUBLIC_FRUSTRATION_INGESTION_URL=https://api.yourdomain.com/v1/events

# Vue
VUE_APP_FRUSTRATION_API_KEY=your-api-key-here
VUE_APP_FRUSTRATION_INGESTION_URL=https://api.yourdomain.com/v1/events

# React (Create React App)
REACT_APP_FRUSTRATION_API_KEY=your-api-key-here
REACT_APP_FRUSTRATION_INGESTION_URL=https://api.yourdomain.com/v1/events
```

---

## Backend Services Setup

### Architecture Overview

The backend consists of 5 microservices:

1. **Event Ingestion API** (Port 8080) - Receives events from SDK
2. **Session Manager** (Port 8081) - Groups events into sessions
3. **UFSE** (Port 8082) - Detects frustration signals
4. **Incident Store** (Port 8084) - Stores validated incidents
5. **Ticket Exporter** (Port 8085) - Exports incidents to Jira/Linear

### Deployment Options

#### Option 1: Docker Compose (Recommended for Development)

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: frustration_engine
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  clickhouse:
    image: clickhouse/clickhouse-server:latest
    ports:
      - "9000:9000"
      - "8123:8123"
    volumes:
      - clickhouse_data:/var/lib/clickhouse

  event-ingestion:
    build: .
    command: ./cmd/event-ingestion
    environment:
      PORT: 8080
      CLICKHOUSE_DSN: clickhouse://clickhouse:9000
      SESSION_MANAGER_URL: http://session-manager:8081
    ports:
      - "8080:8080"
    depends_on:
      - clickhouse
      - session-manager

  session-manager:
    build: .
    command: ./cmd/session-manager
    environment:
      PORT: 8081
      UFSE_URL: http://ufse:8082
    ports:
      - "8081:8081"
    depends_on:
      - ufse

  ufse:
    build: .
    command: ./cmd/ufse
    environment:
      PORT: 8082
      INCIDENT_STORE_URL: http://incident-store:8084
    ports:
      - "8082:8082"
    depends_on:
      - incident-store

  incident-store:
    build: .
    command: ./cmd/incident-store
    environment:
      PORT: 8084
      DATABASE_URL: postgres://postgres:postgres@postgres:5432/frustration_engine?sslmode=disable
    ports:
      - "8084:8084"
    depends_on:
      - postgres

  ticket-exporter:
    build: .
    command: ./cmd/ticket-exporter
    environment:
      PORT: 8085
      INCIDENT_STORE_URL: http://incident-store:8084
      JIRA_URL: https://your-domain.atlassian.net
      JIRA_EMAIL: your-email@example.com
      JIRA_API_TOKEN: your-api-token
      LINEAR_API_KEY: your-linear-api-key
    ports:
      - "8085:8085"
    depends_on:
      - incident-store
```

#### Option 2: Kubernetes (Production)

See `k8s/` directory for Kubernetes manifests.

#### Option 3: Manual Deployment

```bash
# 1. Start databases
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:15
docker run -d -p 9000:9000 clickhouse/clickhouse-server:latest

# 2. Build services
go build -o bin/event-ingestion ./cmd/event-ingestion
go build -o bin/session-manager ./cmd/session-manager
go build -o bin/ufse ./cmd/ufse
go build -o bin/incident-store ./cmd/incident-store
go build -o bin/ticket-exporter ./cmd/ticket-exporter

# 3. Run services
./bin/incident-store --port=8084 --dsn="postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable" &
./bin/ufse --port=8082 --incident-store-url=http://localhost:8084 &
./bin/session-manager --port=8081 --ufse-url=http://localhost:8082 &
./bin/event-ingestion --port=8080 --session-manager=http://localhost:8081 --clickhouse="clickhouse://localhost:9000" &
./bin/ticket-exporter --port=8085 --incident-store-url=http://localhost:8084 &
```

---

## API Configuration

### 1. Get Your API Key

API keys are managed through the Incident Store API (or admin interface):

```bash
# Create API key (admin endpoint - implement as needed)
curl -X POST http://localhost:8084/v1/admin/api-keys \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "your-project-id",
    "name": "Production API Key"
  }'
```

### 2. Configure Frontend SDK

Use the API key in your frontend application:

```typescript
initFrustrationObserver({
  apiKey: 'your-api-key-here',
  ingestionUrl: 'https://api.yourdomain.com/v1/events',
});
```

### 3. Configure Ticket Exporter

Set up Jira or Linear integration:

**Jira:**
```bash
export JIRA_URL=https://your-domain.atlassian.net
export JIRA_EMAIL=your-email@example.com
export JIRA_API_TOKEN=your-api-token
export JIRA_PROJECT_KEY=PROJ
```

**Linear:**
```bash
export LINEAR_API_KEY=your-linear-api-key
export LINEAR_TEAM_ID=your-team-id
```

---

## Framework-Specific Examples

### Next.js (App Router)

```typescript
// app/layout.tsx
'use client';

import { useEffect } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default function RootLayout({ children }) {
  useEffect(() => {
    if (process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY) {
      initFrustrationObserver({
        apiKey: process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY,
        ingestionUrl: process.env.NEXT_PUBLIC_FRUSTRATION_INGESTION_URL,
        enableDebug: process.env.NODE_ENV === 'development',
      });
    }
  }, []);

  return <html>{children}</html>;
}
```

### Next.js (Pages Router)

```typescript
// pages/_app.tsx
import { useEffect } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default function App({ Component, pageProps }) {
  useEffect(() => {
    if (process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY) {
      initFrustrationObserver({
        apiKey: process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY,
        ingestionUrl: process.env.NEXT_PUBLIC_FRUSTRATION_INGESTION_URL,
      });
    }
  }, []);

  return <Component {...pageProps} />;
}
```

### React (Create React App)

```typescript
// src/index.tsx
import { StrictMode } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

if (process.env.REACT_APP_FRUSTRATION_API_KEY) {
  initFrustrationObserver({
    apiKey: process.env.REACT_APP_FRUSTRATION_API_KEY,
    ingestionUrl: process.env.REACT_APP_FRUSTRATION_INGESTION_URL,
  });
}

ReactDOM.render(
  <StrictMode>
    <App />
  </StrictMode>,
  document.getElementById('root')
);
```

### Vue 3

```typescript
// src/main.ts
import { createApp } from 'vue';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';
import App from './App.vue';

if (import.meta.env.VITE_FRUSTRATION_API_KEY) {
  initFrustrationObserver({
    apiKey: import.meta.env.VITE_FRUSTRATION_API_KEY,
    ingestionUrl: import.meta.env.VITE_FRUSTRATION_INGESTION_URL,
  });
}

createApp(App).mount('#app');
```

### Angular

```typescript
// src/app/app.component.ts
import { Component, OnInit } from '@angular/core';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';
import { environment } from '../environments/environment';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
})
export class AppComponent implements OnInit {
  ngOnInit() {
    if (environment.frustrationApiKey) {
      initFrustrationObserver({
        apiKey: environment.frustrationApiKey,
        ingestionUrl: environment.frustrationIngestionUrl,
      });
    }
  }
}
```

### Vanilla JavaScript / HTML

```html
<!DOCTYPE html>
<html>
<head>
  <title>My App</title>
</head>
<body>
  <!-- Your app content -->
  
  <!-- Load SDK -->
  <script src="https://cdn.yourdomain.com/frustration-observer.min.js"></script>
  <script>
    FrustrationObserver.init({
      apiKey: 'your-api-key-here',
      ingestionUrl: 'https://api.yourdomain.com/v1/events',
    });
  </script>
</body>
</html>
```

---

## Testing Your Integration

### 1. Verify SDK is Loaded

Open browser console and check:

```javascript
// Should see SDK initialization message
// Check network tab for requests to /v1/events
```

### 2. Send Test Event

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "events": [
      {
        "eventType": "click",
        "timestamp": "2024-01-16T10:00:00Z",
        "sessionId": "test-session",
        "route": "/test",
        "target": {
          "type": "button",
          "id": "test-btn"
        },
        "metadata": {}
      }
    ]
  }'
```

### 3. Check Service Health

```bash
# Event Ingestion
curl http://localhost:8080/health

# Session Manager
curl http://localhost:8081/health

# UFSE
curl http://localhost:8082/health

# Incident Store
curl http://localhost:8084/health
```

### 4. Monitor Logs

```bash
# Event Ingestion
tail -f /tmp/event-ingestion.log

# Session Manager
tail -f /tmp/session-manager.log

# UFSE
tail -f /tmp/ufse.log

# Incident Store
tail -f /tmp/incident-store.log
```

---

## Production Checklist

- [ ] API keys are stored securely (environment variables, secrets manager)
- [ ] HTTPS is enabled for all API endpoints
- [ ] Rate limiting is configured appropriately
- [ ] Database backups are configured
- [ ] Monitoring and alerting are set up
- [ ] Log aggregation is configured
- [ ] CORS is properly configured for your frontend domain
- [ ] Privacy compliance (GDPR, CCPA) is verified
- [ ] Error tracking is integrated
- [ ] Performance monitoring is enabled

---

## Troubleshooting

### SDK Not Sending Events

1. Check browser console for errors
2. Verify API key is correct
3. Check network tab for failed requests
4. Verify CORS is configured on Event Ingestion API

### Events Not Reaching Session Manager

1. Check Event Ingestion logs: `tail -f /tmp/event-ingestion.log`
2. Verify `SESSION_MANAGER_URL` is configured
3. Check Session Manager is running: `curl http://localhost:8081/health`

### Sessions Not Completing

- Sessions complete 15 minutes after last event (5 min idle + 10 min completion)
- For testing, you can reduce timeouts in `internal/session/completion.go`

### Incidents Not Being Created

1. Check UFSE logs for signal detection
2. Verify correlation rules are met (≥2 signals, ≥1 system feedback)
3. Check confidence threshold (only High confidence incidents are emitted)

---

## Support

For issues or questions:
- Check logs in `/tmp/*.log`
- Review service health endpoints
- Verify all environment variables are set correctly

---

## Next Steps

1. **Set up monitoring**: Configure Prometheus/Grafana for metrics
2. **Configure ticket export**: Set up Jira or Linear integration
3. **Tune thresholds**: Adjust frustration detection thresholds based on your data
4. **Set up alerts**: Configure alerts for high-confidence incidents
