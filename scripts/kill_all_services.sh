#!/bin/bash

# Kill All Services Script
# Stops all Frustration Engine services and databases

set -e

echo "=== Stopping All Frustration Engine Services ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Ports used by services
PORTS=(8080 8081 8082 8084 8085)

# Function to kill process on a port
kill_port() {
    local port=$1
    local pid=$(lsof -ti :$port 2>/dev/null)
    
    if [ -n "$pid" ]; then
        echo -e "${YELLOW}   Killing process on port $port (PID: $pid)...${NC}"
        kill -9 $pid 2>/dev/null || true
        sleep 1
        echo -e "${GREEN}   ✅ Port $port cleared${NC}"
    else
        echo -e "${BLUE}   ℹ️  Port $port is already free${NC}"
    fi
}

# Kill all service ports
echo -e "${BLUE}Stopping microservices...${NC}"
for port in "${PORTS[@]}"; do
    kill_port $port
done

# Stop Docker containers
echo ""
echo -e "${BLUE}Stopping Docker containers...${NC}"

# Stop PostgreSQL
if docker ps -a --filter "name=frustration-postgres" --format "{{.ID}}" | grep -q .; then
    echo -e "${YELLOW}   Stopping frustration-postgres...${NC}"
    docker stop frustration-postgres > /dev/null 2>&1 || true
    docker rm frustration-postgres > /dev/null 2>&1 || true
    echo -e "${GREEN}   ✅ PostgreSQL container stopped${NC}"
else
    echo -e "${BLUE}   ℹ️  PostgreSQL container not running${NC}"
fi

# Stop ClickHouse
if docker ps -a --filter "name=frustration-clickhouse" --format "{{.ID}}" | grep -q .; then
    echo -e "${YELLOW}   Stopping frustration-clickhouse...${NC}"
    docker stop frustration-clickhouse > /dev/null 2>&1 || true
    docker rm frustration-clickhouse > /dev/null 2>&1 || true
    echo -e "${GREEN}   ✅ ClickHouse container stopped${NC}"
else
    echo -e "${BLUE}   ℹ️  ClickHouse container not running${NC}"
fi

# Kill any Go processes that might be running
echo ""
echo -e "${BLUE}Checking for Go processes...${NC}"
GO_PROCESSES=$(pgrep -f "go run.*cmd/(event-ingestion|session-manager|ufse|incident-store|ticket-exporter)" 2>/dev/null || true)

if [ -n "$GO_PROCESSES" ]; then
    echo -e "${YELLOW}   Killing Go service processes...${NC}"
    echo "$GO_PROCESSES" | xargs kill -9 2>/dev/null || true
    echo -e "${GREEN}   ✅ Go processes killed${NC}"
else
    echo -e "${BLUE}   ℹ️  No Go service processes found${NC}"
fi

# Final check
echo ""
echo -e "${BLUE}Final port check...${NC}"
ALL_CLEAR=true
for port in "${PORTS[@]}"; do
    if lsof -ti :$port > /dev/null 2>&1; then
        echo -e "${RED}   ⚠️  Port $port is still in use${NC}"
        ALL_CLEAR=false
    fi
done

if [ "$ALL_CLEAR" = true ]; then
    echo -e "${GREEN}   ✅ All ports are clear${NC}"
fi

echo ""
echo -e "${GREEN}=== All Services Stopped ===${NC}"
echo ""
