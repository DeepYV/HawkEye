#!/bin/bash

# Comprehensive QA Test Suite
# Executes all QA testing scenarios

set -e

echo "=== QA Test Suite Execution ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
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

# 1. Unit Tests
echo "=== 1. Unit Tests ==="
run_test "All Unit Tests" "go test ./internal/testing/... -v"

# 2. Build Verification
echo "=== 2. Build Verification ==="
run_test "Code Compilation" "go build ./..."

# 3. Linting (if available)
echo "=== 3. Code Quality ==="
run_test "Go Vet" "go vet ./..."

# 4. Test Coverage
echo "=== 4. Test Coverage ==="
run_test "Coverage Check" "go test ./internal/testing/... -cover"

# 5. Integration Test Framework
echo "=== 5. Integration Tests ==="
run_test "Integration Test Framework" "go test ./internal/testing/... -run TestIntegration"

# Summary
echo "=== QA Test Summary ==="
echo -e "${YELLOW}Total Tests: $TOTAL${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All QA tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some QA tests failed${NC}"
    exit 1
fi
