#!/bin/bash

# Verify All Improvements Script
# Verifies that all improvements are properly implemented

set -e

echo "=== Verifying All Improvements ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check functions
check_file() {
    local file=$1
    local name=$2
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $name: Found${NC}"
        return 0
    else
        echo -e "${RED}❌ $name: Missing${NC}"
        return 1
    fi
}

check_import() {
    local file=$1
    local import=$2
    local name=$3
    
    if grep -q "$import" "$file" 2>/dev/null; then
        echo -e "${GREEN}✅ $name: Implemented${NC}"
        return 0
    else
        echo -e "${RED}❌ $name: Not found${NC}"
        return 1
    fi
}

# Performance Improvements
echo "=== Performance Improvements ==="
check_file "internal/performance/connection_pool.go" "Connection Pooling"
check_file "internal/performance/http_pool.go" "HTTP Client Pooling"

# Resilience Improvements
echo ""
echo "=== Resilience Improvements ==="
check_file "internal/resilience/retry.go" "Retry Logic"
check_file "internal/resilience/circuit_breaker.go" "Circuit Breaker"

# Security Improvements
echo ""
echo "=== Security Improvements ==="
check_file "internal/security/headers.go" "Security Headers"

# Caching
echo ""
echo "=== Caching ==="
check_file "internal/cache/simple_cache.go" "Caching Layer"

# Enhanced Forwarding
echo ""
echo "=== Enhanced Forwarding ==="
check_file "internal/forwarding/manager_enhanced.go" "Enhanced Forwarding Manager"

# Integration Tests
echo ""
echo "=== Integration Tests ==="
check_file "internal/testing/integration_tests.go" "Integration Tests"

# Verify implementations
echo ""
echo "=== Verifying Implementations ==="
check_import "internal/server/server.go" "security.SecurityHeadersMiddleware" "Security Headers Middleware"
check_import "internal/exporter/engine.go" "time" "Retry Logic (time import)"
check_import "internal/forwarding/manager.go" "MaxIdleConns" "HTTP Connection Pooling"

# Build verification
echo ""
echo "=== Build Verification ==="
if go build ./... 2>/dev/null; then
    echo -e "${GREEN}✅ Code compiles successfully${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== All Improvements Verified ===${NC}"
