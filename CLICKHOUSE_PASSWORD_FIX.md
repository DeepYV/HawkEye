# ClickHouse Password Fix

## Problem
ClickHouse is asking for a password, but the default setup should work without one.

## Solution

### Option 1: Restart ClickHouse Container (Recommended)

```bash
# Stop and remove existing container
docker stop frustration-clickhouse
docker rm frustration-clickhouse

# Start fresh (no password required)
docker run -d \
  --name frustration-clickhouse \
  -p 9001:9000 \
  -p 8124:8123 \
  clickhouse/clickhouse-server:latest

# Wait a few seconds for it to start
sleep 5

# Test connection
clickhouse-client --host localhost --port 9001 --query "SELECT 1"
```

### Option 2: Use Password (If Required)

If your ClickHouse instance requires a password, set it:

```bash
# Set password environment variable
export CLICKHOUSE_PASSWORD=your_password
export CLICKHOUSE_USERNAME=default

# Then start services
./scripts/start_services_custom_port.sh
```

### Option 3: Use Log-Only Mode (For Testing)

If you just want to test without ClickHouse:

```bash
# Start Event Ingestion with log-only mode
go run ./cmd/event-ingestion/main.go \
  -port=8080 \
  -clickhouse="log-only" \
  -session-manager="http://localhost:8081"
```

This will log events to console instead of storing them.

---

## Quick Fix Command

```bash
# Kill everything
./scripts/kill_all_services.sh

# Start databases fresh
./scripts/start_databases.sh

# Start services
export POSTGRES_PORT=5434
export CLICKHOUSE_PORT=9001
./scripts/start_services_custom_port.sh
```

---

## Updated Code

The code now supports password from environment:
- `CLICKHOUSE_PASSWORD` - Password (empty by default)
- `CLICKHOUSE_USERNAME` - Username (default: "default")

No changes needed - it will work with or without password!
