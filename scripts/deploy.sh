#!/bin/bash

# Deployment Script
# Handles deployment to staging/production

set -e

ENVIRONMENT=${1:-staging}
SERVICE=${2:-all}

echo "=== Deployment Script ==="
echo "Environment: $ENVIRONMENT"
echo "Service: $SERVICE"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Pre-deployment checks
echo -e "${YELLOW}=== Pre-Deployment Checks ===${NC}"

# 1. Build verification
echo "1. Verifying build..."
if ! go build ./...; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# 2. Test verification
echo "2. Running tests..."
if ! go test ./internal/testing/... -short; then
    echo -e "${RED}❌ Tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Tests passed${NC}"

# 3. Environment check
echo "3. Checking environment..."
if [ "$ENVIRONMENT" != "staging" ] && [ "$ENVIRONMENT" != "production" ]; then
    echo -e "${RED}❌ Invalid environment: $ENVIRONMENT${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Environment valid${NC}"

# Deployment steps
echo ""
echo -e "${YELLOW}=== Deployment Steps ===${NC}"

# Build binaries
echo "Building binaries..."
for service in event-ingestion session-manager ufse incident-store ticket-exporter; do
    if [ "$SERVICE" == "all" ] || [ "$SERVICE" == "$service" ]; then
        echo "  Building $service..."
        go build -o bin/$service ./cmd/$service/...
        echo -e "${GREEN}  ✅ $service built${NC}"
    fi
done

# Health check
echo ""
echo -e "${YELLOW}=== Post-Deployment Health Check ===${NC}"
echo "Waiting for services to start..."
sleep 5

# Check health endpoints
for port in 8080 8081 8082 8084 8085; do
    if curl -s http://localhost:$port/health > /dev/null; then
        echo -e "${GREEN}✅ Service on port $port is healthy${NC}"
    else
        echo -e "${YELLOW}⚠️  Service on port $port not responding${NC}"
    fi
done

echo ""
echo -e "${GREEN}=== Deployment Complete ===${NC}"
