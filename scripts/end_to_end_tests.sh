#!/bin/bash

# End-to-End Signal Testing Script
# Executes comprehensive end-to-end tests for all signals

set -e

echo "=== End-to-End Signal Testing ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results
TOTAL=0
PASSED=0
FAILED=0

# Function to run test
run_test() {
    local test_name=$1
    local test_command=$2
    
    ((TOTAL++))
    echo -e "${BLUE}Running: $test_name${NC}"
    
    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PASS: $test_name${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAIL: $test_name${NC}"
        ((FAILED++))
    fi
    echo ""
}

# Check services are running
echo "=== Service Health Check ==="
check_service() {
    local service=$1
    local port=$2
    
    if curl -s http://localhost:$port/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $service (port $port): Running${NC}"
        return 0
    else
        echo -e "${RED}❌ $service (port $port): Not running${NC}"
        return 1
    fi
}

check_service "Event Ingestion" 8080
check_service "Session Manager" 8081
check_service "UFSE" 8082
check_service "Incident Store" 8084
check_service "Ticket Exporter" 8085

echo ""

# Run end-to-end tests
echo "=== End-to-End Signal Tests ==="

# Note: These tests are skipped by default (require running services)
# Uncomment t.Skip() in test files to run them

run_test "Rage Signal E2E" "go test ./internal/testing/... -v -run TestEndToEnd_RageSignal"
run_test "Blocked Progress E2E" "go test ./internal/testing/... -v -run TestEndToEnd_BlockedProgressSignal"
run_test "Abandonment E2E" "go test ./internal/testing/... -v -run TestEndToEnd_AbandonmentSignal"
run_test "Confusion E2E" "go test ./internal/testing/... -v -run TestEndToEnd_ConfusionSignal"
run_test "Combined Signals E2E" "go test ./internal/testing/... -v -run TestEndToEnd_AllSignalsCombined"
run_test "False Alarm Prevention E2E" "go test ./internal/testing/... -v -run TestEndToEnd_FalseAlarmPrevention"

# Summary
echo "=== Test Summary ==="
echo -e "${YELLOW}Total Tests: $TOTAL${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All end-to-end tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some end-to-end tests failed${NC}"
    echo ""
    echo "Note: End-to-end tests require all services to be running."
    echo "Start services with: ./scripts/start_services.sh"
    exit 1
fi
