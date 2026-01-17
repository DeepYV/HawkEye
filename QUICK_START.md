# Quick Start Guide

Get the Frustration Engine running in 5 minutes.

## Prerequisites

- Node.js 18+ (for frontend SDK)
- Go 1.21+ (for backend services)
- Docker (optional, for databases)

## Step 1: Start Backend Services

### Option A: Using Docker Compose (Easiest)

```bash
# Start all services
docker-compose up -d

# Check services are running
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8084/health
```

### Option B: Manual Start

```bash
# 1. Start databases
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:15
docker run -d -p 9000:9000 clickhouse/clickhouse-server:latest

# 2. Start services (in separate terminals)
go run ./cmd/incident-store --port=8084 --dsn="postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable"
go run ./cmd/ufse --port=8082 --incident-store-url=http://localhost:8084
go run ./cmd/session-manager --port=8081 --ufse-url=http://localhost:8082
go run ./cmd/event-ingestion --port=8080 --session-manager=http://localhost:8081 --clickhouse="clickhouse://localhost:9000"
```

## Step 2: Get API Key

For testing, a default API key is already configured: `test-api-key`

For production, create an API key through the admin API (implement as needed).

## Step 3: Install Frontend SDK

```bash
npm install @frustration-engine/observer-sdk
```

## Step 4: Initialize SDK in Your App

### React / Next.js

```typescript
import { useEffect } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default function App() {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: 'test-api-key', // Use your API key
      ingestionUrl: 'http://localhost:8080/v1/events', // Your Event Ingestion API URL
    });
  }, []);

  return <YourApp />;
}
```

### Vanilla JavaScript

```html
<script src="https://cdn.yourdomain.com/frustration-observer.min.js"></script>
<script>
  FrustrationObserver.init({
    apiKey: 'test-api-key',
    ingestionUrl: 'http://localhost:8080/v1/events',
  });
</script>
```

## Step 5: Test It

1. **Open your app** in a browser
2. **Interact with your app** (click buttons, fill forms, trigger errors)
3. **Check logs**:

```bash
# Event Ingestion
tail -f /tmp/event-ingestion.log

# Session Manager
tail -f /tmp/session-manager.log

# UFSE
tail -f /tmp/ufse.log
```

## Step 6: View Incidents

After 15 minutes (session completion timeout), check for incidents:

```bash
# Query incidents
curl http://localhost:8084/v1/incidents?project_id=test-project
```

## That's It!

Your app is now tracking user frustration. Incidents will appear in:
- Incident Store API
- Jira/Linear (if configured)

## Next Steps

- Read [INTEGRATION_GUIDE.md](./INTEGRATION_GUIDE.md) for detailed integration
- Configure ticket export to Jira/Linear
- Set up monitoring and alerts
- Tune frustration detection thresholds
