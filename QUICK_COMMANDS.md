# Quick Commands Reference

Quick reference for running and integrating the Frustration Engine.

---

## 🚀 Start Services

### Quick Start (All Services)
```bash
./scripts/start_services.sh
```

### Manual Start (One by One)
```bash
# 1. Start databases
docker run -d --name frustration-postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=frustration_engine postgres:15
docker run -d --name frustration-clickhouse -p 9000:9000 -p 8123:8123 clickhouse/clickhouse-server:latest

# 2. Start services
go run ./cmd/incident-store/main.go -port=8084 -dsn="postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable" &
go run ./cmd/ufse/main.go -port=8082 -incident-store-url=http://localhost:8084 &
go run ./cmd/session-manager/main.go -port=8081 -ufse-url=http://localhost:8082 &
go run ./cmd/event-ingestion/main.go -port=8080 -clickhouse="clickhouse://localhost:9000" -session-manager="http://localhost:8081" &
ADAPTER=noop go run ./cmd/ticket-exporter/main.go -port=8085 -incident-store-url=http://localhost:8084 &
```

---

## ✅ Verify Services

```bash
# Check all health endpoints
curl http://localhost:8080/health  # Event Ingestion
curl http://localhost:8081/health  # Session Manager
curl http://localhost:8082/health  # UFSE
curl http://localhost:8084/health  # Incident Store
curl http://localhost:8085/health  # Ticket Exporter
```

---

## 🧪 Test Integration

### Send Test Event
```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "project_id": "test-project",
    "events": [{
      "eventType": "click",
      "timestamp": "2024-01-16T10:00:00Z",
      "sessionId": "test-session-123",
      "route": "/test",
      "target": {"type": "button", "id": "test-btn"}
    }]
  }'
```

### Query Incidents
```bash
curl http://localhost:8084/v1/incidents?project_id=test-project
```

---

## 📱 Frontend Integration

### React/Next.js
```typescript
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

useEffect(() => {
  initFrustrationObserver({
    apiKey: 'test-api-key',
    ingestionUrl: 'http://localhost:8080/v1/events',
  });
}, []);
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

---

## 📊 View Logs

```bash
tail -f /tmp/Event\ Ingestion.log
tail -f /tmp/Session\ Manager.log
tail -f /tmp/UFSE.log
tail -f /tmp/Incident\ Store.log
tail -f /tmp/Ticket\ Exporter.log
```

---

## 🛑 Stop Services

```bash
./scripts/stop_services.sh
```

Or manually:
```bash
pkill -f "go run.*event-ingestion"
pkill -f "go run.*session-manager"
pkill -f "go run.*ufse"
pkill -f "go run.*incident-store"
pkill -f "go run.*ticket-exporter"
```

---

## 🔧 Environment Variables

```bash
# Event Ingestion
export PORT=8080
export CLICKHOUSE_DSN="clickhouse://localhost:9000"
export SESSION_MANAGER_URL="http://localhost:8081"

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
export ADAPTER="noop"  # or "jira" or "linear"
```

---

## 📚 Full Documentation

- **[RUN_AND_INTEGRATE.md](./RUN_AND_INTEGRATE.md)** - Complete run and integration guide
- **[INTEGRATION_GUIDE.md](./INTEGRATION_GUIDE.md)** - Detailed integration documentation
- **[QUICK_START.md](./QUICK_START.md)** - 5-minute quick start

---

**Quick Start:** `./scripts/start_services.sh` 🚀
