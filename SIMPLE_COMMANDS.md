# Simple Commands - Quick Reference

One-line commands to manage everything.

---

## 🛑 Kill All Ports & Services

```bash
./scripts/kill_all_services.sh
```

**Or manually:**
```bash
# Kill all service ports
lsof -ti :8080 :8081 :8082 :8084 :8085 | xargs kill -9 2>/dev/null || true

# Stop Docker containers
docker stop frustration-postgres frustration-clickhouse 2>/dev/null || true
docker rm frustration-postgres frustration-clickhouse 2>/dev/null || true
```

---

## 🗄️ PostgreSQL

**Start:**
```bash
docker run -d --name frustration-postgres -p 5434:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=frustration_engine postgres:15
```

**Stop:**
```bash
docker stop frustration-postgres && docker rm frustration-postgres
```

**Check:**
```bash
docker ps | grep frustration-postgres
```

---

## 📊 ClickHouse

**Start (removes old container first):**
```bash
docker stop frustration-clickhouse 2>/dev/null || true
docker rm frustration-clickhouse 2>/dev/null || true
docker run -d --name frustration-clickhouse -p 9001:9000 -p 8124:8123 clickhouse/clickhouse-server:latest
```

**Stop:**
```bash
docker stop frustration-clickhouse && docker rm frustration-clickhouse
```

**Check:**
```bash
docker ps | grep frustration-clickhouse
```

---

## 🚀 Services

**Start All (One Command):**
```bash
export POSTGRES_PORT=5434 && export CLICKHOUSE_PORT=9001 && ./scripts/start_services_custom_port.sh
```

**Or Start One by One:**

**Incident Store:**
```bash
go run ./cmd/incident-store/main.go -port=8084 -dsn="postgres://postgres:postgres@localhost:5434/frustration_engine?sslmode=disable"
```

**UFSE:**
```bash
go run ./cmd/ufse/main.go -port=8082 -incident-store-url=http://localhost:8084
```

**Session Manager:**
```bash
go run ./cmd/session-manager/main.go -port=8081 -ufse-url=http://localhost:8082
```

**Event Ingestion:**
```bash
go run ./cmd/event-ingestion/main.go -port=8080 -clickhouse="localhost:9001" -session-manager="http://localhost:8081"
```

**Ticket Exporter:**
```bash
ADAPTER=noop go run ./cmd/ticket-exporter/main.go -port=8085 -incident-store-url=http://localhost:8084
```

---

## ✅ Quick Start (All-in-One)

```bash
# 1. Kill everything first
./scripts/kill_all_services.sh

# 2. Start databases
./scripts/start_databases.sh

# 3. Start services
export POSTGRES_PORT=5434 && export CLICKHOUSE_PORT=9001 && ./scripts/start_services_custom_port.sh
```

---

## 🧪 Test

```bash
curl http://localhost:8080/health
```

---

## 📋 Ports

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

**That's it! Simple and clean.** 🚀
