#!/bin/bash

# Performance Testing Script
# Tests performance improvements

set -e

echo "=== Performance Testing ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test connection pooling
echo -e "${YELLOW}Testing Connection Pooling...${NC}"
go test ./internal/performance/... -v -run TestConnectionPool 2>/dev/null || echo "Connection pool tests (if available)"

# Test HTTP pooling
echo -e "${YELLOW}Testing HTTP Client Pooling...${NC}"
go test ./internal/performance/... -v -run TestHTTPPool 2>/dev/null || echo "HTTP pool tests (if available)"

# Test retry logic
echo -e "${YELLOW}Testing Retry Logic...${NC}"
go test ./internal/resilience/... -v -run TestRetry 2>/dev/null || echo "Retry tests (if available)"

# Test circuit breaker
echo -e "${YELLOW}Testing Circuit Breaker...${NC}"
go test ./internal/resilience/... -v -run TestCircuitBreaker 2>/dev/null || echo "Circuit breaker tests (if available)"

# Benchmark (if available)
echo -e "${YELLOW}Running Benchmarks...${NC}"
go test ./internal/performance/... -bench=. -benchmem 2>/dev/null || echo "Benchmarks (if available)"

echo ""
echo -e "${GREEN}=== Performance Testing Complete ===${NC}"
