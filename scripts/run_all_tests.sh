#!/bin/bash

# Run All Tests Script
# Executes comprehensive test suite

set -e

echo "=== Running Comprehensive Test Suite ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0

# Function to run tests
run_tests() {
    local test_path=$1
    local test_name=$2
    
    echo -e "${YELLOW}Running: $test_name${NC}"
    if go test "$test_path" -v; then
        echo -e "${GREEN}✅ $test_name: PASSED${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ $test_name: FAILED${NC}"
        ((FAILED++))
    fi
    echo ""
}

# Run all test suites
echo "=== Unit Tests ==="
run_tests "./internal/testing/..." "Unit Tests"

echo "=== Integration Tests ==="
run_tests "./internal/testing/..." "Integration Tests"

# Summary
echo "=== Test Summary ==="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
