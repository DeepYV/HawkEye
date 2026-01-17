# Port Conflict Solution

Your port 5432 is already in use by another PostgreSQL instance. Here are solutions:

---

## ✅ Solution 1: Use Different Ports (Recommended)

### Quick Start with Custom Ports

```bash
# Use the custom port script
./scripts/start_services_custom_port.sh
```

This will:
- Start PostgreSQL on port **5433** (instead of 5432)
- Start ClickHouse on port **9001** (instead of 9000)
- Configure all services to use these ports

### Manual Start with Custom Ports

```bash
# 1. Start PostgreSQL on port 5433
docker run -d \
  --name frustration-postgres \
  -p 5433:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=frustration_engine \
  postgres:15

# 2. Start ClickHouse on port 9001
docker run -d \
  --name frustration-clickhouse \
  -p 9001:9000 \
  -p 8124:8123 \
  clickhouse/clickhouse-server:latest

# 3. Start services with custom ports
go run ./cmd/incident-store/main.go \
  -port=8084 \
  -dsn="postgres://postgres:postgres@localhost:5433/frustration_engine?sslmode=disable" &

go run ./cmd/event-ingestion/main.go \
  -port=8080 \
  -clickhouse="clickhouse://localhost:9001" \
  -session-manager="http://localhost:8081" &
```

---

## ✅ Solution 2: Use Existing PostgreSQL Instance

If your existing PostgreSQL is accessible, you can use it:

### Option A: Use Existing Database

```bash
# Check if you can connect to existing PostgreSQL
psql -h localhost -U postgres -d postgres

# Create frustration_engine database in existing PostgreSQL
psql -h localhost -U postgres -c "CREATE DATABASE frustration_engine;"

# Start Incident Store using existing PostgreSQL (port 5432)
go run ./cmd/incident-store/main.go \
  -port=8084 \
  -dsn="postgres://postgres:YOUR_PASSWORD@localhost:5432/frustration_engine?sslmode=disable"
```

**Note:** Replace `YOUR_PASSWORD` with your existing PostgreSQL password.

### Option B: Use Different Database Name

If you want to keep using port 5432 but with a different database:

```bash
# Create a new database in your existing PostgreSQL
psql -h localhost -U postgres -c "CREATE DATABASE frustration_engine;"

# Use it (assuming default postgres user/password)
go run ./cmd/incident-store/main.go \
  -port=8084 \
  -dsn="postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable"
```

---

## ✅ Solution 3: Stop Existing PostgreSQL (If Not Needed)

If the other PostgreSQL instance is not needed:

```bash
# Find and stop the container
docker ps | grep postgres
docker stop discern-postgres-1  # or whatever the container name is

# Or stop by port
lsof -ti:5432 | xargs kill

# Then use the default port 5432
./scripts/start_services.sh
```

---

## 🔧 Environment Variables for Custom Ports

You can set custom ports via environment variables:

```bash
# Set custom ports
export POSTGRES_PORT=5433
export CLICKHOUSE_PORT=9001
export CLICKHOUSE_HTTP_PORT=8124

# Run the custom port script
./scripts/start_services_custom_port.sh
```

Or create a `.env` file:

```bash
# .env
POSTGRES_PORT=5433
CLICKHOUSE_PORT=9001
CLICKHOUSE_HTTP_PORT=8124
DATABASE_URL=postgres://postgres:postgres@localhost:5433/frustration_engine?sslmode=disable
CLICKHOUSE_DSN=clickhouse://localhost:9001
```

---

## 📋 Port Reference

| Service | Default Port | Custom Port Option |
|---------|--------------|-------------------|
| PostgreSQL | 5432 | 5433, 5434, etc. |
| ClickHouse (native) | 9000 | 9001, 9002, etc. |
| ClickHouse (HTTP) | 8123 | 8124, 8125, etc. |
| Event Ingestion | 8080 | - |
| Session Manager | 8081 | - |
| UFSE | 8082 | - |
| Incident Store | 8084 | - |
| Ticket Exporter | 8085 | - |

---

## 🧪 Test Custom Port Setup

```bash
# Test PostgreSQL connection on custom port
psql -h localhost -p 5433 -U postgres -d frustration_engine -c "SELECT 1;"

# Test ClickHouse on custom port
curl http://localhost:8124/ping

# Test Incident Store
curl http://localhost:8084/health
```

---

## 💡 Recommended Approach

**For Development:** Use Solution 1 (different ports) to avoid conflicts:
```bash
./scripts/start_services_custom_port.sh
```

**For Production:** Use Solution 2 (existing database) or dedicated database instance.

---

## 📚 Related Files

- `scripts/start_services_custom_port.sh` - Script with custom ports
- `CREDENTIALS.md` - All credentials reference
- `RUN_AND_INTEGRATE.md` - Full integration guide

---

**Quick Fix:** Run `./scripts/start_services_custom_port.sh` to use ports 5433 and 9001! 🚀
