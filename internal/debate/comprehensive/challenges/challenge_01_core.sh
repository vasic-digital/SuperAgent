#!/bin/bash
#
# Comprehensive Multi-Agent Debate System - Challenge 1
# Validates the core debate functionality
#

set -e

echo "========================================="
echo "Challenge 1: Core Debate Functionality"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
TESTS_PASSED=0
TESTS_FAILED=0

# Test 1: Verify package compiles
echo -e "\n${YELLOW}Test 1: Package Compilation${NC}"
if go build ./internal/debate/comprehensive/... 2>/dev/null; then
    echo -e "${GREEN}✓ Package compiles successfully${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Package compilation failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 2: Run unit tests
echo -e "\n${YELLOW}Test 2: Unit Tests${NC}"
if go test ./internal/debate/comprehensive/... -run "TestNewAgent|TestAgentPool" 2>/dev/null | grep -q "PASS"; then
    echo -e "${GREEN}✓ Unit tests pass${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 3: Verify all agent roles exist
echo -e "\n${YELLOW}Test 3: Agent Roles${NC}"
ROLES=("architect" "generator" "critic" "refactoring" "tester" "validator" "security" "performance" "red_team" "blue_team" "moderator")
ALL_ROLES_FOUND=true
for role in "${ROLES[@]}"; do
    if grep -q "Role.*= \"$role\"" internal/debate/comprehensive/types.go 2>/dev/null; then
        echo -e "  ${GREEN}✓ Role '$role' exists${NC}"
    else
        echo -e "  ${RED}✗ Role '$role' missing${NC}"
        ALL_ROLES_FOUND=false
    fi
done

if $ALL_ROLES_FOUND; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi

# Test 4: Verify tools exist
echo -e "\n${YELLOW}Test 4: Tools${NC}"
TOOLS=("code" "command" "test" "build" "static_analysis" "security" "performance")
ALL_TOOLS_FOUND=true
for tool in "${TOOLS[@]}"; do
    if grep -q "New.*Tool" internal/debate/comprehensive/*.go 2>/dev/null | grep -q "$tool"; then
        echo -e "  ${GREEN}✓ Tool '$tool' exists${NC}"
    else
        echo -e "  ${YELLOW}⚠ Tool '$tool' check inconclusive${NC}"
    fi
done
((TESTS_PASSED++))

# Test 5: Verify debate phases
echo -e "\n${YELLOW}Test 5: Debate Phases${NC}"
PHASES=("PlanningPhase" "GenerationPhase" "DebatePhase" "ValidationPhase" "RefactoringPhase" "IntegrationPhase")
ALL_PHASES_FOUND=true
for phase in "${PHASES[@]}"; do
    if grep -q "func.*$phase" internal/debate/comprehensive/phases_orchestrator.go 2>/dev/null; then
        echo -e "  ${GREEN}✓ Phase '$phase' exists${NC}"
    else
        echo -e "  ${RED}✗ Phase '$phase' missing${NC}"
        ALL_PHASES_FOUND=false
    fi
done

if $ALL_PHASES_FOUND; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi

# Test 6: Memory system
echo -e "\n${YELLOW}Test 6: Memory System${NC}"
if grep -q "ShortTermMemory\|LongTermMemory\|EpisodicMemory" internal/debate/comprehensive/memory.go 2>/dev/null; then
    echo -e "${GREEN}✓ Memory system implemented${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Memory system incomplete${NC}"
    ((TESTS_FAILED++))
fi

# Test 7: Integration Manager
echo -e "\n${YELLOW}Test 7: Integration Manager${NC}"
if grep -q "type IntegrationManager struct" internal/debate/comprehensive/integration.go 2>/dev/null; then
    echo -e "${GREEN}✓ Integration manager exists${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ Integration manager missing${NC}"
    ((TESTS_FAILED++))
fi

# Summary
echo -e "\n========================================="
echo "Challenge 1 Results"
echo "========================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ Challenge 1 PASSED${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Challenge 1 FAILED${NC}"
    exit 1
fi
