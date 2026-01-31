#!/bin/bash

# HawkEye Dev Server - Quick Start Script
#
# Starts all HawkEye services on a single port with no external dependencies.
# No ClickHouse, no PostgreSQL required.
#
# Usage:
#   ./scripts/start_dev.sh
#   PORT=9090 ./scripts/start_dev.sh
#   TEST_API_KEY=my-secret-key ./scripts/start_dev.sh

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PORT="${PORT:-8080}"
TEST_API_KEY="${TEST_API_KEY:-dev-api-key}"

echo -e "${GREEN}=== HawkEye Dev Server ===${NC}"
echo ""
echo -e "${BLUE}Starting single-port development server...${NC}"
echo -e "  Port:    ${PORT}"
echo -e "  API Key: ${TEST_API_KEY}"
echo -e "  Storage: in-memory (no databases needed)"
echo ""
echo -e "${YELLOW}SDK configuration:${NC}"
echo -e "  ingestionUrl: 'http://localhost:${PORT}'"
echo -e "  apiKey:       '${TEST_API_KEY}'"
echo ""

# Set environment for development
export ENVIRONMENT=development
export HAWKEYE_MODE=single-node

# Build and run
echo -e "${BLUE}Building dev server...${NC}"
go build -o /tmp/hawkeye-devserver ./cmd/devserver

echo -e "${GREEN}Starting dev server on http://localhost:${PORT}${NC}"
echo ""
exec /tmp/hawkeye-devserver \
  --port "${PORT}" \
  --api-key "${TEST_API_KEY}" \
  --storage memory
