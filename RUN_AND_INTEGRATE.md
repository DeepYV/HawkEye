# Run & Integrate Guide

Complete guide to run the Frustration Engine and integrate it into your application.

---

## 🚀 Quick Start - Run All Services

### Option 1: Using the Start Script (Easiest)

```bash
# Make script executable
chmod +x scripts/start_services.sh

# Start all services
./scripts/start_services.sh
```

This will start all 5 services:
- **Event Ingestion** (Port 8080)
- **Session Manager** (Port 8081)
- **UFSE** (Port 8082)
- **Incident Store** (Port 8084)
- **Ticket Exporter** (Port 8085)

### Option 2: Manual Start (Step by Step)

#### Step 1: Start Databases

```bash
# PostgreSQL (for Incident Store)
docker run -d \
  --name frustration-postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=frustration_engine \
  postgres:15

# ClickHouse (for Event Ingestion)
docker run -d \
  --name frustration-clickhouse \
  -p 9000:9000 \
  -p 8123:8123 \
  clickhouse/clickhouse-server:latest
```

#### Step 2: Start Services (in separate terminals or background)

**Terminal 1 - Incident Store:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/incident-store/main.go \
  -port=8084 \
  -dsn="postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable"
```

**Terminal 2 - UFSE:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/ufse/main.go \
  -port=8082 \
  -incident-store-url=http://localhost:8084
```

**Terminal 3 - Session Manager:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/session-manager/main.go \
  -port=8081 \
  -ufse-url=http://localhost:8082
```

**Terminal 4 - Event Ingestion:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/event-ingestion/main.go \
  -port=8080 \
  -clickhouse="clickhouse://localhost:9000" \
  -session-manager="http://localhost:8081"
```

**Terminal 5 - Ticket Exporter (Optional):**
```bash
cd /Users/deepakyadav/ucfp
# For testing (NoOp adapter - no external dependencies)
ADAPTER=noop go run ./cmd/ticket-exporter/main.go -port=8085 -incident-store-url=http://localhost:8084

# For production (Jira)
go run ./cmd/ticket-exporter/main.go \
  -port=8085 \
  -incident-store-url=http://localhost:8084 \
  -adapter=jira \
  -jira-url="https://your-domain.atlassian.net" \
  -jira-token="your-token" \
  -jira-project="PROJ"

# For production (Linear)
go run ./cmd/ticket-exporter/main.go \
  -port=8085 \
  -incident-store-url=http://localhost:8084 \
  -adapter=linear \
  -linear-url="https://api.linear.app/graphql" \
  -linear-key="your-key" \
  -linear-team="your-team-id"
```

### Option 3: Build Binaries and Run

```bash
# Build all services
go build -o bin/event-ingestion ./cmd/event-ingestion
go build -o bin/session-manager ./cmd/session-manager
go build -o bin/ufse ./cmd/ufse
go build -o bin/incident-store ./cmd/incident-store
go build -o bin/ticket-exporter ./cmd/ticket-exporter

# Run services
./bin/incident-store -port=8084 -dsn="postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable" &
./bin/ufse -port=8082 -incident-store-url=http://localhost:8084 &
./bin/session-manager -port=8081 -ufse-url=http://localhost:8082 &
./bin/event-ingestion -port=8080 -clickhouse="clickhouse://localhost:9000" -session-manager="http://localhost:8081" &
./bin/ticket-exporter -port=8085 -incident-store-url=http://localhost:8084 -adapter=noop &
```

---

## ✅ Verify Services Are Running

```bash
# Check all service health endpoints
curl http://localhost:8080/health  # Event Ingestion
curl http://localhost:8081/health # Session Manager
curl http://localhost:8082/health # UFSE
curl http://localhost:8084/health # Incident Store
curl http://localhost:8085/health # Ticket Exporter
```

Expected response: `{"status":"healthy"}`

---

## 🔧 Environment Variables

Create a `.env` file (optional, or set in your shell):

```bash
# Event Ingestion
export PORT=8080
export CLICKHOUSE_DSN="clickhouse://localhost:9000"
export SESSION_MANAGER_URL="http://localhost:8081"
export RATE_LIMIT_RPS=1000
export RATE_LIMIT_BURST=2000

# Session Manager
export PORT=8081
export UFSE_URL="http://localhost:8082"

# UFSE
export PORT=8082
export INCIDENT_STORE_URL="http://localhost:8084"

# Incident Store
export PORT=8084
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable"

# Ticket Exporter
export PORT=8085
export INCIDENT_STORE_URL="http://localhost:8084"
export ADAPTER="noop"  # or "jira" or "linear"

# For Jira
export JIRA_URL="https://your-domain.atlassian.net"
export JIRA_TOKEN="your-token"
export JIRA_PROJECT="PROJ"

# For Linear
export LINEAR_URL="https://api.linear.app/graphql"
export LINEAR_KEY="your-key"
export LINEAR_TEAM="your-team-id"

# Test API Key (for development)
export TEST_API_KEY="test-api-key"
```

---

## 📱 Integrate into Your Application

### Step 1: Get Your API Key

For testing, use the default API key: `test-api-key`

For production, create an API key through your admin interface or Incident Store API.

### Step 2: Install Frontend SDK

**NPM:**
```bash
npm install @frustration-engine/observer-sdk
```

**Yarn:**
```bash
yarn add @frustration-engine/observer-sdk
```

**CDN (HTML):**
```html
<script src="https://cdn.yourdomain.com/frustration-observer.min.js"></script>
```

### Step 3: Initialize SDK in Your App

#### React / Next.js

```typescript
// app/layout.tsx (Next.js App Router) or pages/_app.tsx (Pages Router)
'use client';

import { useEffect } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default function RootLayout({ children }) {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY || 'test-api-key',
      ingestionUrl: process.env.NEXT_PUBLIC_FRUSTRATION_INGESTION_URL || 'http://localhost:8080/v1/events',
      enableDebug: process.env.NODE_ENV === 'development',
    });
  }, []);

  return <html>{children}</html>;
}
```

#### React (Create React App)

```typescript
// src/index.tsx
import { StrictMode } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

if (process.env.REACT_APP_FRUSTRATION_API_KEY) {
  initFrustrationObserver({
    apiKey: process.env.REACT_APP_FRUSTRATION_API_KEY,
    ingestionUrl: process.env.REACT_APP_FRUSTRATION_INGESTION_URL || 'http://localhost:8080/v1/events',
  });
}

ReactDOM.render(
  <StrictMode>
    <App />
  </StrictMode>,
  document.getElementById('root')
);
```

#### Vue 3

```typescript
// src/main.ts
import { createApp } from 'vue';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';
import App from './App.vue';

if (import.meta.env.VITE_FRUSTRATION_API_KEY) {
  initFrustrationObserver({
    apiKey: import.meta.env.VITE_FRUSTRATION_API_KEY,
    ingestionUrl: import.meta.env.VITE_FRUSTRATION_INGESTION_URL || 'http://localhost:8080/v1/events',
  });
}

createApp(App).mount('#app');
```

#### Angular

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
    initFrustrationObserver({
      apiKey: environment.frustrationApiKey,
      ingestionUrl: environment.frustrationIngestionUrl || 'http://localhost:8080/v1/events',
    });
  }
}
```

#### Vanilla JavaScript / HTML

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
      apiKey: 'test-api-key',
      ingestionUrl: 'http://localhost:8080/v1/events',
    });
  </script>
</body>
</html>
```

---

## 🧪 Test Your Integration

### 1. Send a Test Event

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "project_id": "test-project",
    "events": [
      {
        "eventType": "click",
        "timestamp": "2024-01-16T10:00:00Z",
        "sessionId": "test-session-123",
        "route": "/test",
        "target": {
          "type": "button",
          "id": "test-btn",
          "text": "Click Me"
        },
        "metadata": {
          "userId": "user-123"
        }
      }
    ]
  }'
```

### 2. Check Service Logs

```bash
# Event Ingestion
tail -f /tmp/Event\ Ingestion.log

# Session Manager
tail -f /tmp/Session\ Manager.log

# UFSE
tail -f /tmp/UFSE.log

# Incident Store
tail -f /tmp/Incident\ Store.log
```

### 3. Query Incidents (after session completes - ~15 minutes)

```bash
# Get all incidents
curl http://localhost:8084/v1/incidents

# Get incidents by project
curl http://localhost:8084/v1/incidents?project_id=test-project

# Get incidents by status
curl http://localhost:8084/v1/incidents?status=confirmed
```

---

## 📡 API Endpoints

### Event Ingestion API (Port 8080)

**POST /v1/events** - Send events
```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{"project_id": "test", "events": [...]}'
```

**GET /health** - Health check
```bash
curl http://localhost:8080/health
```

### Session Manager API (Port 8081)

**GET /health** - Health check
```bash
curl http://localhost:8081/health
```

### UFSE API (Port 8082)

**GET /health** - Health check
```bash
curl http://localhost:8082/health
```

### Incident Store API (Port 8084)

**GET /v1/incidents** - List incidents
```bash
curl http://localhost:8084/v1/incidents?project_id=test-project&status=confirmed
```

**GET /health** - Health check
```bash
curl http://localhost:8084/health
```

### Ticket Exporter API (Port 8085)

**GET /health** - Health check
```bash
curl http://localhost:8085/health
```

---

## 🛑 Stop Services

### If using start script:
```bash
./scripts/stop_services.sh
```

### If running manually:
```bash
# Find and kill processes
pkill -f "go run.*event-ingestion"
pkill -f "go run.*session-manager"
pkill -f "go run.*ufse"
pkill -f "go run.*incident-store"
pkill -f "go run.*ticket-exporter"

# Or kill by port
lsof -ti:8080 | xargs kill
lsof -ti:8081 | xargs kill
lsof -ti:8082 | xargs kill
lsof -ti:8084 | xargs kill
lsof -ti:8085 | xargs kill
```

### Stop databases:
```bash
docker stop frustration-postgres frustration-clickhouse
docker rm frustration-postgres frustration-clickhouse
```

---

## 🔍 Troubleshooting

### Services Not Starting

1. **Check if ports are in use:**
```bash
lsof -i :8080
lsof -i :8081
lsof -i :8082
lsof -i :8084
lsof -i :8085
```

2. **Check database connections:**
```bash
# PostgreSQL
psql -h localhost -U postgres -d frustration_engine

# ClickHouse
curl http://localhost:9000
```

3. **Check logs:**
```bash
tail -f /tmp/Event\ Ingestion.log
tail -f /tmp/Session\ Manager.log
```

### Events Not Being Processed

1. **Verify API key:**
```bash
# Test with curl
curl -X POST http://localhost:8080/v1/events \
  -H "X-API-Key: test-api-key" \
  -H "Content-Type: application/json" \
  -d '{"project_id": "test", "events": []}'
```

2. **Check service health:**
```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8084/health
```

3. **Verify service connectivity:**
- Event Ingestion → Session Manager
- Session Manager → UFSE
- UFSE → Incident Store

### Incidents Not Being Created

1. **Sessions need time to complete** (15 minutes after last event)
2. **Check UFSE logs** for signal detection
3. **Verify correlation rules** (≥2 signals, ≥1 system feedback)
4. **Check confidence threshold** (only High confidence incidents are emitted)

---

## 📚 Next Steps

1. **Configure Production API Keys** - Replace `test-api-key` with production keys
2. **Set up Ticket Export** - Configure Jira or Linear integration
3. **Enable Monitoring** - Set up Prometheus/Grafana
4. **Tune Thresholds** - Adjust frustration detection based on your data
5. **Set up Alerts** - Configure alerts for high-confidence incidents

---

## 📖 Additional Documentation

- **[Integration Guide](./INTEGRATION_GUIDE.md)** - Detailed integration documentation
- **[Quick Start](./QUICK_START.md)** - 5-minute quick start
- **[Architecture](./ARCHITECTURE.md)** - System architecture
- **[Testing Guide](./TESTING_GUIDE.md)** - Testing instructions

---

**Ready to integrate!** 🚀
