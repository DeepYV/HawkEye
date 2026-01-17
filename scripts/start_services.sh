#!/bin/bash

# Start All Services Script
# Starts all services for end-to-end testing

set -e

echo "=== Starting All Services ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if databases are running
check_database() {
    local name=$1
    local port=$2
    local check_cmd=$3
    
    if eval "$check_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $name is running${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  $name not detected on port $port${NC}"
        echo -e "${BLUE}   Start it with: docker run -d --name frustration-$name -p $port:$port ...${NC}"
        return 1
    fi
}

# Check databases
echo -e "${BLUE}Checking databases...${NC}"
check_database "PostgreSQL" 5432 "psql -h localhost -U postgres -d frustration_engine -c 'SELECT 1' 2>/dev/null" || true
check_database "ClickHouse" 9000 "curl -s http://localhost:9000 > /dev/null" || true
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
start_service "Incident Store" 8084 "go run ./cmd/incident-store/main.go -port=8084 -dsn=\"postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable\""

# Start UFSE
start_service "UFSE" 8082 "go run ./cmd/ufse/main.go -port=8082 -incident-store-url=http://localhost:8084"

# Start Session Manager
start_service "Session Manager" 8081 "go run ./cmd/session-manager/main.go -port=8081 -ufse-url=http://localhost:8082"

# Start Event Ingestion
start_service "Event Ingestion" 8080 "go run ./cmd/event-ingestion/main.go -port=8080 -clickhouse=\"clickhouse://localhost:9000\" -session-manager=\"http://localhost:8081\""

# Start Ticket Exporter (with NoOp adapter for testing)
start_service "Ticket Exporter" 8085 "ADAPTER=noop go run ./cmd/ticket-exporter/main.go -port=8085 -incident-store-url=http://localhost:8084"

echo ""
echo -e "${GREEN}=== All Services Started ===${NC}"
echo ""
echo "Service Logs:"
echo "  Event Ingestion: tail -f /tmp/Event\\ Ingestion.log"
echo "  Session Manager: tail -f /tmp/Session\\ Manager.log"
echo "  UFSE: tail -f /tmp/UFSE.log"
echo "  Incident Store: tail -f /tmp/Incident\\ Store.log"
echo "  Ticket Exporter: tail -f /tmp/Ticket\\ Exporter.log"
echo ""
echo "To stop services: ./scripts/stop_services.sh"
