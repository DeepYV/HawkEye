#!/bin/bash

# CI/CD Pipeline Script
# Automated testing and deployment pipeline

set -e

echo "=== CI/CD Pipeline ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

STAGE=$1

case $STAGE in
    test)
        echo -e "${BLUE}=== Test Stage ===${NC}"
        go test ./internal/testing/... -v -cover
        go vet ./...
        echo -e "${GREEN}✅ Test stage complete${NC}"
        ;;
    
    build)
        echo -e "${BLUE}=== Build Stage ===${NC}"
        go build ./...
        echo -e "${GREEN}✅ Build stage complete${NC}"
        ;;
    
    deploy-staging)
        echo -e "${BLUE}=== Deploy to Staging ===${NC}"
        ./scripts/deploy.sh staging
        echo -e "${GREEN}✅ Staging deployment complete${NC}"
        ;;
    
    deploy-production)
        echo -e "${BLUE}=== Deploy to Production ===${NC}"
        ./scripts/deploy.sh production
        echo -e "${GREEN}✅ Production deployment complete${NC}"
        ;;
    
    all)
        echo -e "${BLUE}=== Full Pipeline ===${NC}"
        echo "1. Testing..."
        go test ./internal/testing/... -v
        echo "2. Building..."
        go build ./...
        echo "3. Deploying to staging..."
        ./scripts/deploy.sh staging
        echo -e "${GREEN}✅ Full pipeline complete${NC}"
        ;;
    
    *)
        echo "Usage: $0 {test|build|deploy-staging|deploy-production|all}"
        exit 1
        ;;
esac
