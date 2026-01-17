#!/bin/bash

# Start All Services Script (Custom Ports)
# Starts all services with custom database ports to avoid conflicts

set -e

echo "=== Starting All Services (Custom Ports) ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Custom ports (change these if needed)
POSTGRES_PORT=${POSTGRES_PORT:-5433}
CLICKHOUSE_PORT=${CLICKHOUSE_PORT:-9001}
CLICKHOUSE_HTTP_PORT=${CLICKHOUSE_HTTP_PORT:-8124}

echo -e "${BLUE}Using custom ports:${NC}"
echo -e "  PostgreSQL: ${POSTGRES_PORT}"
echo -e "  ClickHouse: ${CLICKHOUSE_PORT} (native), ${CLICKHOUSE_HTTP_PORT} (HTTP)"
echo ""

# Check if databases are running
check_database() {
    local name=$1
    local port=$2
    local check_cmd=$3
    
    if eval "$check_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $name is running on port $port${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  $name not detected on port $port${NC}"
        return 1
    fi
}

# Start databases first (handles port conflicts automatically)
echo -e "${BLUE}Starting databases...${NC}"
if [ -f "$(dirname "$0")/start_databases.sh" ]; then
    source "$(dirname "$0")/start_databases.sh"
    # Get the actual ports used
    POSTGRES_PORT=${POSTGRES_PORT:-5434}
    CLICKHOUSE_PORT=${CLICKHOUSE_PORT:-9001}
else
    # Fallback: try to start databases manually
    if ! docker ps | grep -q "frustration-postgres"; then
        echo -e "${YELLOW}⚠️  PostgreSQL container not found${NC}"
        echo -e "${BLUE}   Starting PostgreSQL on port ${POSTGRES_PORT}...${NC}"
        # Remove stopped container if exists
        docker rm frustration-postgres > /dev/null 2>&1 || true
        docker run -d \
            --name frustration-postgres \
            -p ${POSTGRES_PORT}:5432 \
            -e POSTGRES_PASSWORD=postgres \
            -e POSTGRES_DB=frustration_engine \
            postgres:15 > /dev/null 2>&1
        sleep 3
        echo -e "${GREEN}✅ PostgreSQL started on port ${POSTGRES_PORT}${NC}"
    else
        echo -e "${GREEN}✅ PostgreSQL container already running${NC}"
        # Get actual port from running container
        POSTGRES_PORT=$(docker port frustration-postgres 5432/tcp | cut -d: -f2)
    fi

    if ! docker ps | grep -q "frustration-clickhouse"; then
        echo -e "${YELLOW}⚠️  ClickHouse container not found${NC}"
        echo -e "${BLUE}   Starting ClickHouse on port ${CLICKHOUSE_PORT}...${NC}"
        # Remove stopped container if exists
        docker rm frustration-clickhouse > /dev/null 2>&1 || true
        docker run -d \
            --name frustration-clickhouse \
            -p ${CLICKHOUSE_PORT}:9000 \
            -p ${CLICKHOUSE_HTTP_PORT}:8123 \
            clickhouse/clickhouse-server:latest > /dev/null 2>&1
        sleep 3
        echo -e "${GREEN}✅ ClickHouse started on port ${CLICKHOUSE_PORT}${NC}"
    else
        echo -e "${GREEN}✅ ClickHouse container already running${NC}"
    fi
fi
echo ""

# Start services in background
start_service() {
    local service=$1
    local port=$2
    local cmd=$3
    
    echo -e "${YELLOW}Starting $service on port $port...${NC}"
    
    # Change to project root directory
    cd "$(dirname "$0")/.."
    
    if eval "$cmd" > /tmp/${service}.log 2>&1 & then
        echo -e "${GREEN}✅ $service started (PID: $!)${NC}"
        sleep 2
        return 0
    else
        echo -e "${RED}❌ Failed to start $service${NC}"
        return 1
    fi
}

# Change to project root
cd "$(dirname "$0")/.."

# Start Incident Store (first - other services depend on it)
# Using custom PostgreSQL port
start_service "Incident Store" 8084 "go run ./cmd/incident-store/main.go -port=8084 -dsn=\"postgres://postgres:postgres@localhost:${POSTGRES_PORT}/frustration_engine?sslmode=disable\""

# Start UFSE
start_service "UFSE" 8082 "go run ./cmd/ufse/main.go -port=8082 -incident-store-url=http://localhost:8084"

# Start Session Manager
start_service "Session Manager" 8081 "go run ./cmd/session-manager/main.go -port=8081 -ufse-url=http://localhost:8082"

# Start Event Ingestion (using custom ClickHouse port)
start_service "Event Ingestion" 8080 "go run ./cmd/event-ingestion/main.go -port=8080 -clickhouse=\"localhost:${CLICKHOUSE_PORT}\" -session-manager=\"http://localhost:8081\""

# Start Ticket Exporter (with NoOp adapter for testing)
start_service "Ticket Exporter" 8085 "ADAPTER=noop go run ./cmd/ticket-exporter/main.go -port=8085 -incident-store-url=http://localhost:8084"

echo ""
echo -e "${GREEN}=== All Services Started ===${NC}"
echo ""
echo "Database Ports:"
echo "  PostgreSQL: ${POSTGRES_PORT}"
echo "  ClickHouse: ${CLICKHOUSE_PORT} (native), ${CLICKHOUSE_HTTP_PORT} (HTTP)"
echo ""
echo "Service Logs:"
echo "  Event Ingestion: tail -f /tmp/Event\\ Ingestion.log"
echo "  Session Manager: tail -f /tmp/Session\\ Manager.log"
echo "  UFSE: tail -f /tmp/UFSE.log"
echo "  Incident Store: tail -f /tmp/Incident\\ Store.log"
echo "  Ticket Exporter: tail -f /tmp/Ticket\\ Exporter.log"
echo ""
echo "To stop services: ./scripts/stop_services.sh"
