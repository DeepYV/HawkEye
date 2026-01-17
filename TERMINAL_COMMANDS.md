# Terminal Commands - Copy & Paste

Simple commands to run the Frustration Engine in your terminal.

---

## 🚀 Quick Start (All-in-One)

### Step 1: Start Databases

```bash
cd /Users/deepakyadav/ucfp
./scripts/start_databases.sh
```

This will start PostgreSQL (port 5434) and ClickHouse (port 9001).

---

### Step 2: Start All Services

**Option A: Using Script (Easiest)**
```bash
export POSTGRES_PORT=5434
export CLICKHOUSE_PORT=9001
./scripts/start_services_custom_port.sh
```

**Option B: Manual (One by One)**

Open 5 terminal windows/tabs and run:

**Terminal 1 - Incident Store:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/incident-store/main.go -port=8084 -dsn="postgres://postgres:postgres@localhost:5434/frustration_engine?sslmode=disable"
```

**Terminal 2 - UFSE:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/ufse/main.go -port=8082 -incident-store-url=http://localhost:8084
```

**Terminal 3 - Session Manager:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/session-manager/main.go -port=8081 -ufse-url=http://localhost:8082
```

**Terminal 4 - Event Ingestion:**
```bash
cd /Users/deepakyadav/ucfp
go run ./cmd/event-ingestion/main.go -port=8080 -clickhouse="localhost:9001" -session-manager="http://localhost:8081"
```

**Note:** ClickHouse DSN should be `host:port` format (e.g., `localhost:9001`), not `clickhouse://host:port`. The code now supports both formats.

**Terminal 5 - Ticket Exporter:**
```bash
cd /Users/deepakyadav/ucfp
ADAPTER=noop go run ./cmd/ticket-exporter/main.go -port=8085 -incident-store-url=http://localhost:8084
```

**Note:** The `-incident-store-url` flag is now properly defined. If you get an error, make sure you're using the latest code.

---

## ✅ Verify Services

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8084/health
curl http://localhost:8085/health
```

---

## 🛑 Stop All Services

**Quick Stop:**
```bash
./scripts/kill_all_services.sh
```

This will:
- Kill all processes on ports 8080, 8081, 8082, 8084, 8085
- Stop PostgreSQL and ClickHouse Docker containers
- Clean up all Go service processes

**Manual Stop:**
```bash
# Kill processes on specific ports
lsof -ti :8080 | xargs kill -9
lsof -ti :8081 | xargs kill -9
lsof -ti :8082 | xargs kill -9
lsof -ti :8084 | xargs kill -9
lsof -ti :8085 | xargs kill -9

# Stop Docker containers
docker stop frustration-postgres frustration-clickhouse
docker rm frustration-postgres frustration-clickhouse
```

---

## 🧪 Test It

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

---

## 🛑 Stop Services

**Stop Go Services:**
```bash
cd /Users/deepakyadav/ucfp
./scripts/stop_services.sh
```

**Stop Databases:**
```bash
docker stop frustration-postgres frustration-clickhouse
```

**Or stop everything:**
```bash
cd /Users/deepakyadav/ucfp
./scripts/stop_services.sh
docker stop frustration-postgres frustration-clickhouse
```

---

## 📋 Port Reference

| Service | Port |
|---------|------|
| PostgreSQL | 5434 |
| ClickHouse | 9001 |
| Event Ingestion | 8080 |
| Session Manager | 8081 |
| UFSE | 8082 |
| Incident Store | 8084 |
| Ticket Exporter | 8085 |

---

## 🔑 Credentials

- **PostgreSQL:** `postgres` / `postgres`
- **API Key:** `test-api-key`
- **ClickHouse:** No password

---

## 📝 Quick Copy-Paste Sequence

```bash
# 1. Start databases
cd /Users/deepakyadav/ucfp
./scripts/start_databases.sh

# 2. Start services (in background or separate terminals)
export POSTGRES_PORT=5434
export CLICKHOUSE_PORT=9001
./scripts/start_services_custom_port.sh

# 3. Verify
curl http://localhost:8080/health

# 4. Test
curl -X POST http://localhost:8080/v1/events \
  -H "X-API-Key: test-api-key" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"test","events":[]}'
```

---

**That's it! Copy and paste these commands into your terminal.** 🚀
