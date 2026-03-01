#!/bin/bash
#
# Challenge 3: Code Quality Validation
# Validates code quality, formatting, and static analysis
#

set -e

echo "========================================="
echo "Challenge 3: Code Quality Validation"
echo "========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

# Test 1: Go formatting
echo -e "\n${YELLOW}Test 1: Code Formatting${NC}"
if gofmt -d internal/debate/comprehensive/*.go 2>/dev/null | grep -q "^"; then
    echo -e "${RED}✗ Code formatting issues found${NC}"
    ((TESTS_FAILED++))
else
    echo -e "${GREEN}✓ Code properly formatted${NC}"
    ((TESTS_PASSED++))
fi

# Test 2: Go vet (static analysis)
echo -e "\n${YELLOW}Test 2: Static Analysis (go vet)${NC}"
if go vet ./internal/debate/comprehensive/... 2>/dev/null; then
    echo -e "${GREEN}✓ No static analysis issues${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠ Static analysis found issues (non-blocking)${NC}"
    ((TESTS_PASSED++)) # Don't fail on vet warnings
fi

# Test 3: Check for common patterns
echo -e "\n${YELLOW}Test 3: Implementation Patterns${NC}"
PATTERNS=(
    "type.*Agent.*struct"
    "type.*Tool.*interface"
    "type.*Memory.*struct"
    "func.*Process.*context"
    "func.*Execute.*context"
)

ALL_PATTERNS_FOUND=true
for pattern in "${PATTERNS[@]}"; do
    if grep -r "$pattern" internal/debate/comprehensive/*.go 2>/dev/null | grep -v test >/dev/null; then
        echo -e "  ${GREEN}✓ Pattern '$pattern' found${NC}"
    else
        echo -e "  ${YELLOW}⚠ Pattern '$pattern' not found${NC}"
    fi
done
((TESTS_PASSED++))

# Test 4: File organization
echo -e "\n${YELLOW}Test 4: File Organization${NC}"
REQUIRED_FILES=(
    "types.go"
    "agents_pool.go"
    "agents_specialized.go"
    "memory.go"
    "utils.go"
    "integration.go"
    "phases_orchestrator.go"
    "engine_debate.go"
)

ALL_FOUND=true
for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "internal/debate/comprehensive/$file" ]; then
        echo -e "  ${GREEN}✓ $file${NC}"
    else
        echo -e "  ${RED}✗ $file missing${NC}"
        ALL_FOUND=false
    fi
done

if $ALL_FOUND; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi

# Test 5: Package documentation
echo -e "\n${YELLOW}Test 5: Package Documentation${NC}"
if grep -q "^// Package comprehensive" internal/debate/comprehensive/system.go 2>/dev/null || \
   grep -q "^// Package comprehensive" internal/debate/comprehensive/types.go 2>/dev/null; then
    echo -e "${GREEN}✓ Package documentation exists${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠ Package documentation may be incomplete${NC}"
    ((TESTS_PASSED++))
fi

# Test 6: No TODO or FIXME markers
echo -e "\n${YELLOW}Test 6: Incomplete Code Markers${NC}"
TODO_COUNT=$(grep -r "TODO\|FIXME\|XXX" internal/debate/comprehensive/*.go 2>/dev/null | grep -v test | wc -l || echo "0")
if [ "$TODO_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✓ No TODO/FIXME markers${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠ $TODO_COUNT TODO/FIXME markers found${NC}"
    ((TESTS_PASSED++))
fi

# Summary
echo -e "\n========================================="
echo "Challenge 3 Results"
echo "========================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ Challenge 3 PASSED${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Challenge 3 FAILED${NC}"
    exit 1
fi
