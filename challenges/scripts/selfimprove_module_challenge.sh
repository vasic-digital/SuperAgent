#!/bin/bash
# HelixAgent Challenge: SelfImprove Module
# Tests: ~12 tests across 4 sections
# Validates: Module build, reward model, feedback types, key interfaces,
#            tests, coverage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

MODULE_DIR="$PROJECT_ROOT/SelfImprove"
PKG_DIR="$MODULE_DIR/selfimprove"

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
   grep -q 'module digital.vasic.selfimprove' "$MODULE_DIR/go.mod"; then
    pass "go.mod exists with module digital.vasic.selfimprove"
else
    fail "go.mod missing or incorrect module name"
fi

# Test 1.2: All source files exist (types, reward, feedback, optimizer, integration)
if [ -f "$PKG_DIR/types.go" ] && \
   [ -f "$PKG_DIR/reward.go" ] && \
   [ -f "$PKG_DIR/feedback.go" ] && \
   [ -f "$PKG_DIR/optimizer.go" ] && \
   [ -f "$PKG_DIR/integration.go" ]; then
    pass "All 5 source files exist (types, reward, feedback, optimizer, integration)"
else
    fail "Missing one or more source files"
fi

# Test 1.3: Test files exist
if [ -f "$PKG_DIR/reward_test.go" ] && \
   [ -f "$PKG_DIR/types_test.go" ]; then
    pass "Test files exist (reward_test.go, types_test.go)"
else
    fail "Missing one or more test files"
fi

#===============================================================================
# Section 2: Reward Model (3 tests)
#===============================================================================
section "Section 2: Reward Model"

# Test 2.1: RewardModel interface exists
if grep -q 'type RewardModel interface' "$PKG_DIR/types.go"; then
    pass "RewardModel interface defined"
else
    fail "RewardModel interface missing"
fi

# Test 2.2: AIRewardModel implementation and constructor
if grep -q 'type AIRewardModel struct' "$PKG_DIR/reward.go" && \
   grep -q 'func NewAIRewardModel(' "$PKG_DIR/reward.go"; then
    pass "AIRewardModel implementation and NewAIRewardModel constructor exist"
else
    fail "AIRewardModel or constructor missing"
fi

# Test 2.3: Score, ScoreWithDimensions, Compare, and Train methods
if grep -q 'func (rm \*AIRewardModel) Score(' "$PKG_DIR/reward.go" && \
   grep -q 'func (rm \*AIRewardModel) ScoreWithDimensions(' "$PKG_DIR/reward.go" && \
   grep -q 'func (rm \*AIRewardModel) Compare(' "$PKG_DIR/reward.go" && \
   grep -q 'func (rm \*AIRewardModel) Train(' "$PKG_DIR/reward.go"; then
    pass "Score, ScoreWithDimensions, Compare, and Train methods exist"
else
    fail "One or more RewardModel methods missing"
fi

#===============================================================================
# Section 3: Feedback Types and Interfaces (3 tests)
#===============================================================================
section "Section 3: Feedback Types and Interfaces"

# Test 3.1: Feedback and TrainingExample types
if grep -q 'type Feedback struct' "$PKG_DIR/types.go" && \
   grep -q 'type TrainingExample struct' "$PKG_DIR/types.go" && \
   grep -q 'type PreferencePair struct' "$PKG_DIR/types.go"; then
    pass "Feedback, TrainingExample, and PreferencePair types defined"
else
    fail "Feedback or training types missing"
fi

# Test 3.2: FeedbackCollector and PolicyOptimizer interfaces
if grep -q 'type FeedbackCollector interface' "$PKG_DIR/types.go" && \
   grep -q 'type PolicyOptimizer interface' "$PKG_DIR/types.go"; then
    pass "FeedbackCollector and PolicyOptimizer interfaces defined"
else
    fail "FeedbackCollector or PolicyOptimizer interface missing"
fi

# Test 3.3: DimensionType constants (accuracy, relevance, helpfulness, harmless, honest, coherence)
if grep -q 'DimensionAccuracy' "$PKG_DIR/types.go" && \
   grep -q 'DimensionRelevance' "$PKG_DIR/types.go" && \
   grep -q 'DimensionHelpfulness' "$PKG_DIR/types.go" && \
   grep -q 'DimensionHarmless' "$PKG_DIR/types.go" && \
   grep -q 'DimensionHonest' "$PKG_DIR/types.go" && \
   grep -q 'DimensionCoherence' "$PKG_DIR/types.go"; then
    pass "All 6 dimension types defined (accuracy, relevance, helpfulness, harmless, honest, coherence)"
else
    fail "One or more dimension types missing"
fi

#===============================================================================
# Section 4: Build, Tests, and Coverage (3 tests)
#===============================================================================
section "Section 4: Build, Tests, and Coverage"

# Test 4.1: Module compiles successfully
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./... >/dev/null 2>&1); then
    pass "SelfImprove module compiles successfully"
else
    fail "SelfImprove module compilation failed"
fi

# Test 4.2: All tests pass
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -count=1 -timeout 120s ./... >/dev/null 2>&1); then
    pass "All SelfImprove module tests pass"
else
    fail "SelfImprove module tests failed"
fi

# Test 4.3: Test coverage >= 90%
COVERAGE_OUTPUT=$(cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -coverprofile=/tmp/selfimprove_coverage.out ./... 2>/dev/null)
if [ -f /tmp/selfimprove_coverage.out ]; then
    COVERAGE=$(cd "$MODULE_DIR" && go tool cover -func=/tmp/selfimprove_coverage.out 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
    COVERAGE_INT=${COVERAGE%%.*}
    if [ "$COVERAGE_INT" -ge 90 ]; then
        pass "Test coverage >= 90% (${COVERAGE}%)"
    else
        fail "Test coverage ${COVERAGE}% (expected >= 90%)"
    fi
    rm -f /tmp/selfimprove_coverage.out
else
    fail "Could not generate coverage report"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}SelfImprove Module Challenge Results${NC}"
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
