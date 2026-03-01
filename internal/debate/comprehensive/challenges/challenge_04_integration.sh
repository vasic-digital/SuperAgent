#!/bin/bash
#
# Challenge 4: Integration Validation
# Validates integration between components
#

set -e

echo "========================================="
echo "Challenge 4: Integration Validation"
echo "========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

# Test 1: Integration tests pass
echo -e "\n${YELLOW}Test 1: Integration Tests${NC}"
if go test ./internal/debate/comprehensive/... -run "Integration|Manager" 2>/dev/null | grep -q "PASS"; then
    echo -e "${GREEN}✓ Integration tests pass${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Integration tests failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 2: Phase orchestration tests
echo -e "\n${YELLOW}Test 2: Phase Orchestration${NC}"
if go test ./internal/debate/comprehensive/... -run "Phase" 2>/dev/null | grep -q "PASS"; then
    echo -e "${GREEN}✓ Phase orchestration works${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Phase orchestration issues${NC}"
    ((TESTS_FAILED++))
fi

# Test 3: Agent pool integration
echo -e "\n${YELLOW}Test 3: Agent Pool Integration${NC}"
if go test ./internal/debate/comprehensive/... -run "AgentPool" 2>/dev/null | grep -q "PASS"; then
    echo -e "${GREEN}✓ Agent pool integration works${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Agent pool integration issues${NC}"
    ((TESTS_FAILED++))
fi

# Test 4: Memory system integration
echo -e "\n${YELLOW}Test 4: Memory System Integration${NC}"
if go test ./internal/debate/comprehensive/... -run "Memory" 2>/dev/null | grep -q "PASS"; then
    echo -e "${GREEN}✓ Memory system integration works${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Memory system integration issues${NC}"
    ((TESTS_FAILED++))
fi

# Test 5: Consensus algorithm
echo -e "\n${YELLOW}Test 5: Consensus Algorithm${NC}"
if go test ./internal/debate/comprehensive/... -run "Consensus" 2>/dev/null | grep -q "PASS"; then
    echo -e "${GREEN}✓ Consensus algorithm works${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Consensus algorithm issues${NC}"
    ((TESTS_FAILED++))
fi

# Test 6: Tool registry integration
echo -e "\n${YELLOW}Test 6: Tool Registry${NC}"
if go test ./internal/debate/comprehensive/... -run "ToolRegistry" 2>/dev/null | grep -q "PASS"; then
    echo -e "${GREEN}✓ Tool registry integration works${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Tool registry integration issues${NC}"
    ((TESTS_FAILED++))
fi

# Summary
echo -e "\n========================================="
echo "Challenge 4 Results"
echo "========================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ Challenge 4 PASSED${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Challenge 4 FAILED${NC}"
    exit 1
fi
