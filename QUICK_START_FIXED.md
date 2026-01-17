# Quick Start (Fixed for Port Conflicts)

**Problem Solved:** Port 5432 was already in use by another project.

---

## ✅ Solution: Use Port 5434 for PostgreSQL

The databases are now running on:
- **PostgreSQL:** Port **5434** (instead of 5432)
- **ClickHouse:** Port **9001** (instead of 9000)

---

## 🚀 Quick Start Commands

### Step 1: Start Databases (Auto-detects available ports)

```bash
./scripts/start_databases.sh
```

This will:
- Automatically find available ports (5434, 9001)
- Start PostgreSQL and ClickHouse
- Show you the connection strings

### Step 2: Start All Services

```bash
# Set the ports (from Step 1 output)
export POSTGRES_PORT=5434
export CLICKHOUSE_PORT=9001

# Start all services
./scripts/start_services_custom_port.sh
```

Or manually:

```bash
# Start Incident Store
go run ./cmd/incident-store/main.go \
  -port=8084 \
  -dsn="postgres://postgres:postgres@localhost:5434/frustration_engine?sslmode=disable" &

# Start UFSE
go run ./cmd/ufse/main.go \
  -port=8082 \
  -incident-store-url=http://localhost:8084 &

# Start Session Manager
go run ./cmd/session-manager/main.go \
  -port=8081 \
  -ufse-url=http://localhost:8082 &

# Start Event Ingestion
go run ./cmd/event-ingestion/main.go \
  -port=8080 \
  -clickhouse="localhost:9001" \
  -session-manager="http://localhost:8081" &

# Start Ticket Exporter
ADAPTER=noop go run ./cmd/ticket-exporter/main.go \
  -port=8085 \
  -incident-store-url=http://localhost:8084 &
```

---

## ✅ Verify Everything is Running

```bash
# Check databases
docker ps | grep -E "frustration-postgres|frustration-clickhouse"

# Check services
curl http://localhost:8080/health  # Event Ingestion
curl http://localhost:8081/health  # Session Manager
curl http://localhost:8082/health  # UFSE
curl http://localhost:8084/health  # Incident Store
curl http://localhost:8085/health  # Ticket Exporter
```

---

## 📋 Current Port Configuration

| Service | Port | Status |
|---------|------|--------|
| PostgreSQL | **5434** | ✅ Running |
| ClickHouse | **9001** | ✅ Running |
| Event Ingestion | 8080 | ✅ Running |
| Session Manager | 8081 | ✅ Running |
| UFSE | 8082 | ✅ Running |
| Incident Store | 8084 | ✅ Running |
| Ticket Exporter | 8085 | ✅ Running |

---

## 🔧 Connection Strings

**PostgreSQL:**
```
postgres://postgres:postgres@localhost:5434/frustration_engine?sslmode=disable
```

**ClickHouse:**
```
clickhouse://localhost:9001
```

---

## 🧪 Test It

```bash
# Send a test event
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

```bash
# Stop Go services
./scripts/stop_services.sh

# Stop databases
docker stop frustration-postgres frustration-clickhouse
```

---

## 📚 Related Documentation

- **[PORT_CONFLICT_SOLUTION.md](./PORT_CONFLICT_SOLUTION.md)** - Full port conflict solutions
- **[RUN_AND_INTEGRATE.md](./RUN_AND_INTEGRATE.md)** - Complete integration guide
- **[CREDENTIALS.md](./CREDENTIALS.md)** - All credentials reference

---

**✅ Everything is now running on custom ports to avoid conflicts!**
