#!/bin/bash
# HelixAgent Challenge: Planning Module
# Tests: ~15 tests across 5 sections
# Validates: Module build, 3 planning algorithms (HiPlan, MCTS, TreeOfThoughts),
#            key interfaces, exported types, tests, coverage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

MODULE_DIR="$PROJECT_ROOT/Planning"
PKG_DIR="$MODULE_DIR/planning"

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
   grep -q 'module digital.vasic.planning' "$MODULE_DIR/go.mod"; then
    pass "go.mod exists with module digital.vasic.planning"
else
    fail "go.mod missing or incorrect module name"
fi

# Test 1.2: All three algorithm source files exist
if [ -f "$PKG_DIR/hiplan.go" ] && \
   [ -f "$PKG_DIR/mcts.go" ] && \
   [ -f "$PKG_DIR/tree_of_thoughts.go" ]; then
    pass "All 3 algorithm source files exist (hiplan, mcts, tree_of_thoughts)"
else
    fail "Missing one or more algorithm source files"
fi

# Test 1.3: All three test files exist
if [ -f "$PKG_DIR/hiplan_test.go" ] && \
   [ -f "$PKG_DIR/mcts_test.go" ] && \
   [ -f "$PKG_DIR/tree_of_thoughts_test.go" ]; then
    pass "All 3 test files exist"
else
    fail "Missing one or more test files"
fi

#===============================================================================
# Section 2: HiPlan Algorithm (3 tests)
#===============================================================================
section "Section 2: HiPlan (Hierarchical Planning)"

# Test 2.1: HiPlan struct and constructor exist
if grep -q 'type HiPlan struct' "$PKG_DIR/hiplan.go" && \
   grep -q 'func NewHiPlan(' "$PKG_DIR/hiplan.go"; then
    pass "HiPlan struct and NewHiPlan constructor exist"
else
    fail "HiPlan struct or constructor missing"
fi

# Test 2.2: MilestoneGenerator and StepExecutor interfaces
if grep -q 'type MilestoneGenerator interface' "$PKG_DIR/hiplan.go" && \
   grep -q 'type StepExecutor interface' "$PKG_DIR/hiplan.go"; then
    pass "MilestoneGenerator and StepExecutor interfaces defined"
else
    fail "MilestoneGenerator or StepExecutor interface missing"
fi

# Test 2.3: Key types: Milestone, PlanStep, HierarchicalPlan, PlanResult
if grep -q 'type Milestone struct' "$PKG_DIR/hiplan.go" && \
   grep -q 'type PlanStep struct' "$PKG_DIR/hiplan.go" && \
   grep -q 'type HierarchicalPlan struct' "$PKG_DIR/hiplan.go" && \
   grep -q 'type PlanResult struct' "$PKG_DIR/hiplan.go"; then
    pass "Key types defined (Milestone, PlanStep, HierarchicalPlan, PlanResult)"
else
    fail "One or more key types missing"
fi

#===============================================================================
# Section 3: MCTS Algorithm (3 tests)
#===============================================================================
section "Section 3: MCTS (Monte Carlo Tree Search)"

# Test 3.1: MCTS struct and constructor exist
if grep -q 'type MCTS struct' "$PKG_DIR/mcts.go" && \
   grep -q 'func NewMCTS(' "$PKG_DIR/mcts.go"; then
    pass "MCTS struct and NewMCTS constructor exist"
else
    fail "MCTS struct or constructor missing"
fi

# Test 3.2: MCTSActionGenerator, MCTSRewardFunction, MCTSRolloutPolicy interfaces
if grep -q 'type MCTSActionGenerator interface' "$PKG_DIR/mcts.go" && \
   grep -q 'type MCTSRewardFunction interface' "$PKG_DIR/mcts.go" && \
   grep -q 'type MCTSRolloutPolicy interface' "$PKG_DIR/mcts.go"; then
    pass "MCTS strategy interfaces defined (ActionGenerator, RewardFunction, RolloutPolicy)"
else
    fail "One or more MCTS interfaces missing"
fi

# Test 3.3: MCTSNode and MCTSResult types with Search method
if grep -q 'type MCTSNode struct' "$PKG_DIR/mcts.go" && \
   grep -q 'type MCTSResult struct' "$PKG_DIR/mcts.go" && \
   grep -q 'func (m \*MCTS) Search(' "$PKG_DIR/mcts.go"; then
    pass "MCTSNode, MCTSResult types and Search method exist"
else
    fail "MCTSNode, MCTSResult, or Search method missing"
fi

#===============================================================================
# Section 4: Tree of Thoughts Algorithm (3 tests)
#===============================================================================
section "Section 4: Tree of Thoughts"

# Test 4.1: TreeOfThoughts struct and constructor exist
if grep -q 'type TreeOfThoughts struct' "$PKG_DIR/tree_of_thoughts.go" && \
   grep -q 'func NewTreeOfThoughts(' "$PKG_DIR/tree_of_thoughts.go"; then
    pass "TreeOfThoughts struct and NewTreeOfThoughts constructor exist"
else
    fail "TreeOfThoughts struct or constructor missing"
fi

# Test 4.2: ThoughtGenerator and ThoughtEvaluator interfaces
if grep -q 'type ThoughtGenerator interface' "$PKG_DIR/tree_of_thoughts.go" && \
   grep -q 'type ThoughtEvaluator interface' "$PKG_DIR/tree_of_thoughts.go"; then
    pass "ThoughtGenerator and ThoughtEvaluator interfaces defined"
else
    fail "ThoughtGenerator or ThoughtEvaluator interface missing"
fi

# Test 4.3: Thought, ToTResult types and Solve method
if grep -q 'type Thought struct' "$PKG_DIR/tree_of_thoughts.go" && \
   grep -q 'type ToTResult struct' "$PKG_DIR/tree_of_thoughts.go" && \
   grep -q 'func (t \*TreeOfThoughts) Solve(' "$PKG_DIR/tree_of_thoughts.go"; then
    pass "Thought, ToTResult types and Solve method exist"
else
    fail "Thought, ToTResult, or Solve method missing"
fi

#===============================================================================
# Section 5: Build, Tests, and Coverage (3 tests)
#===============================================================================
section "Section 5: Build, Tests, and Coverage"

# Test 5.1: Module compiles successfully
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./... >/dev/null 2>&1); then
    pass "Planning module compiles successfully"
else
    fail "Planning module compilation failed"
fi

# Test 5.2: All tests pass
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -count=1 -timeout 120s ./... >/dev/null 2>&1); then
    pass "All Planning module tests pass"
else
    fail "Planning module tests failed"
fi

# Test 5.3: Test coverage >= 90%
COVERAGE_OUTPUT=$(cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -coverprofile=/tmp/planning_coverage.out ./... 2>/dev/null)
if [ -f /tmp/planning_coverage.out ]; then
    COVERAGE=$(cd "$MODULE_DIR" && go tool cover -func=/tmp/planning_coverage.out 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
    COVERAGE_INT=${COVERAGE%%.*}
    if [ "$COVERAGE_INT" -ge 90 ]; then
        pass "Test coverage >= 90% (${COVERAGE}%)"
    else
        fail "Test coverage ${COVERAGE}% (expected >= 90%)"
    fi
    rm -f /tmp/planning_coverage.out
else
    fail "Could not generate coverage report"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Planning Module Challenge Results${NC}"
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
