#!/bin/bash
# HelixAgent Challenge: LLMOps Module
# Tests: ~15 tests across 5 sections
# Validates: Module build, evaluator, experiments, prompts components,
#            key types and interfaces, tests, coverage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

MODULE_DIR="$PROJECT_ROOT/LLMOps"
PKG_DIR="$MODULE_DIR/llmops"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}[FAIL]${NC} $1"
}

section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

#===============================================================================
# Section 1: Module Structure (3 tests)
#===============================================================================
section "Section 1: Module Structure"

# Test 1.1: go.mod exists with correct module name
if [ -f "$MODULE_DIR/go.mod" ] && \
   grep -q 'module digital.vasic.llmops' "$MODULE_DIR/go.mod"; then
    pass "go.mod exists with module digital.vasic.llmops"
else
    fail "go.mod missing or incorrect module name"
fi

# Test 1.2: All source files exist (types, evaluator, experiments, prompts, integration)
if [ -f "$PKG_DIR/types.go" ] && \
   [ -f "$PKG_DIR/evaluator.go" ] && \
   [ -f "$PKG_DIR/experiments.go" ] && \
   [ -f "$PKG_DIR/prompts.go" ] && \
   [ -f "$PKG_DIR/integration.go" ]; then
    pass "All 5 source files exist (types, evaluator, experiments, prompts, integration)"
else
    fail "Missing one or more source files"
fi

# Test 1.3: All test files exist
if [ -f "$PKG_DIR/evaluator_test.go" ] && \
   [ -f "$PKG_DIR/experiments_test.go" ] && \
   [ -f "$PKG_DIR/prompts_test.go" ]; then
    pass "All 3 test files exist"
else
    fail "Missing one or more test files"
fi

#===============================================================================
# Section 2: Evaluator Component (3 tests)
#===============================================================================
section "Section 2: Continuous Evaluator"

# Test 2.1: ContinuousEvaluator interface exists
if grep -q 'type ContinuousEvaluator interface' "$PKG_DIR/types.go"; then
    pass "ContinuousEvaluator interface defined"
else
    fail "ContinuousEvaluator interface missing"
fi

# Test 2.2: InMemoryContinuousEvaluator implementation and constructor
if grep -q 'type InMemoryContinuousEvaluator struct' "$PKG_DIR/evaluator.go" && \
   grep -q 'func NewInMemoryContinuousEvaluator(' "$PKG_DIR/evaluator.go"; then
    pass "InMemoryContinuousEvaluator implementation and constructor exist"
else
    fail "InMemoryContinuousEvaluator or constructor missing"
fi

# Test 2.3: EvaluationRun and EvaluationResults types
if grep -q 'type EvaluationRun struct' "$PKG_DIR/types.go" && \
   grep -q 'type EvaluationResults struct' "$PKG_DIR/types.go" && \
   grep -q 'type SampleResult struct' "$PKG_DIR/types.go"; then
    pass "EvaluationRun, EvaluationResults, and SampleResult types defined"
else
    fail "Evaluation types missing"
fi

#===============================================================================
# Section 3: Experiments Component (3 tests)
#===============================================================================
section "Section 3: Experiment Manager"

# Test 3.1: ExperimentManager interface exists
if grep -q 'type ExperimentManager interface' "$PKG_DIR/types.go"; then
    pass "ExperimentManager interface defined"
else
    fail "ExperimentManager interface missing"
fi

# Test 3.2: InMemoryExperimentManager implementation and constructor
if grep -q 'type InMemoryExperimentManager struct' "$PKG_DIR/experiments.go" && \
   grep -q 'func NewInMemoryExperimentManager(' "$PKG_DIR/experiments.go"; then
    pass "InMemoryExperimentManager implementation and constructor exist"
else
    fail "InMemoryExperimentManager or constructor missing"
fi

# Test 3.3: Experiment, Variant, and ExperimentResult types
if grep -q 'type Experiment struct' "$PKG_DIR/types.go" && \
   grep -q 'type Variant struct' "$PKG_DIR/types.go" && \
   grep -q 'type ExperimentResult struct' "$PKG_DIR/types.go"; then
    pass "Experiment, Variant, and ExperimentResult types defined"
else
    fail "Experiment types missing"
fi

#===============================================================================
# Section 4: Prompts Component (3 tests)
#===============================================================================
section "Section 4: Prompt Registry"

# Test 4.1: PromptRegistry interface exists
if grep -q 'type PromptRegistry interface' "$PKG_DIR/types.go"; then
    pass "PromptRegistry interface defined"
else
    fail "PromptRegistry interface missing"
fi

# Test 4.2: InMemoryPromptRegistry implementation and constructor
if grep -q 'type InMemoryPromptRegistry struct' "$PKG_DIR/prompts.go" && \
   grep -q 'func NewInMemoryPromptRegistry(' "$PKG_DIR/prompts.go"; then
    pass "InMemoryPromptRegistry implementation and constructor exist"
else
    fail "InMemoryPromptRegistry or constructor missing"
fi

# Test 4.3: PromptVersion and DatasetManager types
if grep -q 'type PromptVersion struct' "$PKG_DIR/types.go" && \
   grep -q 'type DatasetManager interface' "$PKG_DIR/types.go" && \
   grep -q 'type Dataset struct' "$PKG_DIR/types.go"; then
    pass "PromptVersion, DatasetManager, and Dataset types defined"
else
    fail "Prompt or dataset types missing"
fi

#===============================================================================
# Section 5: Build, Tests, and Coverage (3 tests)
#===============================================================================
section "Section 5: Build, Tests, and Coverage"

# Test 5.1: Module compiles successfully
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./... >/dev/null 2>&1); then
    pass "LLMOps module compiles successfully"
else
    fail "LLMOps module compilation failed"
fi

# Test 5.2: All tests pass
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -count=1 -timeout 120s ./... >/dev/null 2>&1); then
    pass "All LLMOps module tests pass"
else
    fail "LLMOps module tests failed"
fi

# Test 5.3: Test coverage >= 90%
COVERAGE_OUTPUT=$(cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -coverprofile=/tmp/llmops_coverage.out ./... 2>/dev/null)
if [ -f /tmp/llmops_coverage.out ]; then
    COVERAGE=$(cd "$MODULE_DIR" && go tool cover -func=/tmp/llmops_coverage.out 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
    COVERAGE_INT=${COVERAGE%%.*}
    if [ "$COVERAGE_INT" -ge 90 ]; then
        pass "Test coverage >= 90% (${COVERAGE}%)"
    else
        fail "Test coverage ${COVERAGE}% (expected >= 90%)"
    fi
    rm -f /tmp/llmops_coverage.out
else
    fail "Could not generate coverage report"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}LLMOps Module Challenge Results${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Total:  $TOTAL"
echo -e "  ${GREEN}Passed: $PASSED${NC}"
if [ "$FAILED" -gt 0 ]; then
    echo -e "  ${RED}Failed: $FAILED${NC}"
    exit 1
else
    echo -e "  Failed: 0"
fi
echo ""
echo -e "${GREEN}All tests passed!${NC}"
