#!/bin/bash

# Start Databases Script
# Starts PostgreSQL and ClickHouse with conflict detection

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Default ports
POSTGRES_PORT=${POSTGRES_PORT:-5434}
CLICKHOUSE_PORT=${CLICKHOUSE_PORT:-9001}
CLICKHOUSE_HTTP_PORT=${CLICKHOUSE_HTTP_PORT:-8124}

echo -e "${BLUE}=== Starting Databases ===" && echo ""

# Function to check if port is available
check_port() {
    local port=$1
    if lsof -i :$port > /dev/null 2>&1; then
        return 1  # Port is taken
    else
        return 0  # Port is free
    fi
}

# Function to find available port
find_available_port() {
    local start_port=$1
    local port=$start_port
    while ! check_port $port; do
        port=$((port + 1))
        if [ $port -gt $((start_port + 10)) ]; then
            echo -e "${RED}❌ Could not find available port starting from $start_port${NC}"
            exit 1
        fi
    done
    echo $port
}

# Check and find available PostgreSQL port
if ! check_port $POSTGRES_PORT; then
    echo -e "${YELLOW}⚠️  Port $POSTGRES_PORT is already in use${NC}"
    POSTGRES_PORT=$(find_available_port $POSTGRES_PORT)
    echo -e "${BLUE}   Using port $POSTGRES_PORT instead${NC}"
fi

# Check and find available ClickHouse port
if ! check_port $CLICKHOUSE_PORT; then
    echo -e "${YELLOW}⚠️  Port $CLICKHOUSE_PORT is already in use${NC}"
    CLICKHOUSE_PORT=$(find_available_port $CLICKHOUSE_PORT)
    echo -e "${BLUE}   Using port $CLICKHOUSE_PORT instead${NC}"
fi

# Start PostgreSQL
echo -e "${BLUE}Starting PostgreSQL on port $POSTGRES_PORT...${NC}"

# Remove existing container if it exists and is stopped
if docker ps -a | grep -q "frustration-postgres"; then
    if ! docker ps | grep -q "frustration-postgres"; then
        echo -e "${YELLOW}   Removing stopped container...${NC}"
        docker rm frustration-postgres > /dev/null 2>&1 || true
    else
        echo -e "${GREEN}✅ PostgreSQL container already running${NC}"
        exit 0
    fi
fi

# Start PostgreSQL
docker run -d \
    --name frustration-postgres \
    -p ${POSTGRES_PORT}:5432 \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=frustration_engine \
    postgres:15 > /dev/null 2>&1

sleep 3

# Verify PostgreSQL is running
if docker ps | grep -q "frustration-postgres"; then
    echo -e "${GREEN}✅ PostgreSQL started on port $POSTGRES_PORT${NC}"
else
    echo -e "${RED}❌ Failed to start PostgreSQL${NC}"
    exit 1
fi

# Start ClickHouse
echo -e "${BLUE}Starting ClickHouse on port $CLICKHOUSE_PORT...${NC}"

# Remove existing container if it exists and is stopped
if docker ps -a | grep -q "frustration-clickhouse"; then
    if ! docker ps | grep -q "frustration-clickhouse"; then
        echo -e "${YELLOW}   Removing stopped container...${NC}"
        docker rm frustration-clickhouse > /dev/null 2>&1 || true
    else
        echo -e "${GREEN}✅ ClickHouse container already running${NC}"
        echo ""
        echo -e "${GREEN}=== Databases Ready ===" && echo ""
        echo "PostgreSQL: localhost:$POSTGRES_PORT"
        echo "ClickHouse: localhost:$CLICKHOUSE_PORT"
        exit 0
    fi
fi

# Start ClickHouse (without password requirement for default user)
docker run -d \
    --name frustration-clickhouse \
    -p ${CLICKHOUSE_PORT}:9000 \
    -p ${CLICKHOUSE_HTTP_PORT}:8123 \
    -e CLICKHOUSE_DB=events \
    clickhouse/clickhouse-server:latest > /dev/null 2>&1

sleep 3

# Verify ClickHouse is running
if docker ps | grep -q "frustration-clickhouse"; then
    echo -e "${GREEN}✅ ClickHouse started on port $CLICKHOUSE_PORT${NC}"
else
    echo -e "${RED}❌ Failed to start ClickHouse${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== Databases Ready ===" && echo ""
echo "PostgreSQL: localhost:$POSTGRES_PORT"
echo "  Connection: postgres://postgres:postgres@localhost:$POSTGRES_PORT/frustration_engine?sslmode=disable"
echo ""
echo "ClickHouse: localhost:$CLICKHOUSE_PORT"
echo "  Connection: clickhouse://localhost:$CLICKHOUSE_PORT"
echo ""
echo "Export these for services:"
echo "  export POSTGRES_PORT=$POSTGRES_PORT"
echo "  export CLICKHOUSE_PORT=$CLICKHOUSE_PORT"
