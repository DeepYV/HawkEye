#!/bin/bash

# Run End-to-End Tests with Services
# Starts services, runs tests, cleans up

set -e

echo "=== End-to-End Test Execution ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Cleanup function
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up services...${NC}"
    pkill -f "go run.*event-ingestion" 2>/dev/null || true
    pkill -f "go run.*session-manager" 2>/dev/null || true
    pkill -f "go run.*ufse" 2>/dev/null || true
    pkill -f "go run.*incident-store" 2>/dev/null || true
    pkill -f "go run.*ticket-exporter" 2>/dev/null || true
    sleep 2
    echo -e "${GREEN}Cleanup complete${NC}"
}

trap cleanup EXIT

# Start services
echo -e "${BLUE}Starting services...${NC}"

cd cmd/event-ingestion && go run main.go -port=8080 > /tmp/event-ingestion.log 2>&1 &
EVENT_PID=$!
sleep 2

cd ../session-manager && go run main.go -port=8081 > /tmp/session-manager.log 2>&1 &
SESSION_PID=$!
sleep 2

cd ../ufse && go run main.go -port=8082 > /tmp/ufse.log 2>&1 &
UFSE_PID=$!
sleep 2

cd ../incident-store && go run main.go -port=8084 > /tmp/incident-store.log 2>&1 &
INCIDENT_PID=$!
sleep 2

cd ../ticket-exporter && go run main.go -port=8085 > /tmp/ticket-exporter.log 2>&1 &
TICKET_PID=$!
sleep 3

# Verify services
echo -e "${BLUE}Verifying services...${NC}"
for port in 8080 8081 8082 8084 8085; do
    if curl -s http://localhost:$port/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Port $port: Healthy${NC}"
    else
        echo -e "${RED}❌ Port $port: Not responding${NC}"
    fi
done

echo ""
echo -e "${BLUE}Running end-to-end tests...${NC}"
echo ""

# Run tests
cd ../../..
go test ./internal/testing/... -v -run TestEndToEnd

echo ""
echo -e "${GREEN}=== Test Execution Complete ===${NC}"
