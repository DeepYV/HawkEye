#!/bin/bash

# Stop All Services Script
# Stops all running Frustration Engine services

set -e

echo "=== Stopping All Services ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Stop services
stop_service() {
    local service=$1
    local pattern=$2
    
    echo -e "${YELLOW}Stopping $service...${NC}"
    
    if pkill -f "$pattern" 2>/dev/null; then
        echo -e "${GREEN}✅ $service stopped${NC}"
        sleep 1
        return 0
    else
        echo -e "${YELLOW}⚠️  $service not running${NC}"
        return 0
    fi
}

# Stop Event Ingestion
stop_service "Event Ingestion" "go run.*event-ingestion"

# Stop Session Manager
stop_service "Session Manager" "go run.*session-manager"

# Stop UFSE
stop_service "UFSE" "go run.*ufse"

# Stop Incident Store
stop_service "Incident Store" "go run.*incident-store"

# Stop Ticket Exporter
stop_service "Ticket Exporter" "go run.*ticket-exporter"

# Also try to kill by port (in case processes are running differently)
for port in 8080 8081 8082 8084 8085; do
    pid=$(lsof -ti:$port 2>/dev/null || true)
    if [ ! -z "$pid" ]; then
        echo -e "${YELLOW}Killing process on port $port (PID: $pid)...${NC}"
        kill $pid 2>/dev/null || true
        sleep 1
    fi
done

echo ""
echo -e "${GREEN}=== All Services Stopped ===${NC}"
echo ""
