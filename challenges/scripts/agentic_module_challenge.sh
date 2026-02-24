#!/bin/bash
# HelixAgent Challenge: Agentic Module
# Tests: ~12 tests across 4 sections
# Validates: Module build, workflow orchestration types, node types,
#            key interfaces, tests, coverage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

MODULE_DIR="$PROJECT_ROOT/Agentic"
PKG_DIR="$MODULE_DIR/agentic"

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
   grep -q 'module digital.vasic.agentic' "$MODULE_DIR/go.mod"; then
    pass "go.mod exists with module digital.vasic.agentic"
else
    fail "go.mod missing or incorrect module name"
fi

# Test 1.2: Source file exists
if [ -f "$PKG_DIR/workflow.go" ]; then
    pass "workflow.go source file exists"
else
    fail "workflow.go source file missing"
fi

# Test 1.3: Test file exists
if [ -f "$PKG_DIR/workflow_test.go" ]; then
    pass "workflow_test.go test file exists"
else
    fail "workflow_test.go test file missing"
fi

#===============================================================================
# Section 2: Workflow Orchestration Types (3 tests)
#===============================================================================
section "Section 2: Workflow Orchestration Types"

# Test 2.1: Workflow struct and constructor
if grep -q 'type Workflow struct' "$PKG_DIR/workflow.go" && \
   grep -q 'func NewWorkflow(' "$PKG_DIR/workflow.go"; then
    pass "Workflow struct and NewWorkflow constructor exist"
else
    fail "Workflow struct or constructor missing"
fi

# Test 2.2: WorkflowGraph, WorkflowState, and WorkflowConfig types
if grep -q 'type WorkflowGraph struct' "$PKG_DIR/workflow.go" && \
   grep -q 'type WorkflowState struct' "$PKG_DIR/workflow.go" && \
   grep -q 'type WorkflowConfig struct' "$PKG_DIR/workflow.go"; then
    pass "WorkflowGraph, WorkflowState, and WorkflowConfig types defined"
else
    fail "Workflow graph/state/config types missing"
fi

# Test 2.3: Node, Edge, NodeInput, NodeOutput types
if grep -q 'type Node struct' "$PKG_DIR/workflow.go" && \
   grep -q 'type Edge struct' "$PKG_DIR/workflow.go" && \
   grep -q 'type NodeInput struct' "$PKG_DIR/workflow.go" && \
   grep -q 'type NodeOutput struct' "$PKG_DIR/workflow.go"; then
    pass "Node, Edge, NodeInput, and NodeOutput types defined"
else
    fail "Node/Edge/NodeInput/NodeOutput types missing"
fi

#===============================================================================
# Section 3: Node Types and Methods (3 tests)
#===============================================================================
section "Section 3: Node Types and Workflow Methods"

# Test 3.1: All 6 node types defined
if grep -q 'NodeTypeAgent' "$PKG_DIR/workflow.go" && \
   grep -q 'NodeTypeTool' "$PKG_DIR/workflow.go" && \
   grep -q 'NodeTypeCondition' "$PKG_DIR/workflow.go" && \
   grep -q 'NodeTypeParallel' "$PKG_DIR/workflow.go" && \
   grep -q 'NodeTypeHuman' "$PKG_DIR/workflow.go" && \
   grep -q 'NodeTypeSubgraph' "$PKG_DIR/workflow.go"; then
    pass "All 6 node types defined (Agent, Tool, Condition, Parallel, Human, Subgraph)"
else
    fail "One or more node types missing"
fi

# Test 3.2: Key workflow methods exist (AddNode, AddEdge, Execute)
if grep -q 'func (w \*Workflow) AddNode(' "$PKG_DIR/workflow.go" && \
   grep -q 'func (w \*Workflow) AddEdge(' "$PKG_DIR/workflow.go" && \
   grep -q 'func (w \*Workflow) Execute(' "$PKG_DIR/workflow.go"; then
    pass "Key methods exist (AddNode, AddEdge, Execute)"
else
    fail "One or more key methods missing"
fi

# Test 3.3: Checkpoint support (RestoreFromCheckpoint)
if grep -q 'type Checkpoint struct' "$PKG_DIR/workflow.go" && \
   grep -q 'func (w \*Workflow) RestoreFromCheckpoint(' "$PKG_DIR/workflow.go"; then
    pass "Checkpoint support exists (Checkpoint struct, RestoreFromCheckpoint)"
else
    fail "Checkpoint support missing"
fi

#===============================================================================
# Section 4: Build, Tests, and Coverage (3 tests)
#===============================================================================
section "Section 4: Build, Tests, and Coverage"

# Test 4.1: Module compiles successfully
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./... >/dev/null 2>&1); then
    pass "Agentic module compiles successfully"
else
    fail "Agentic module compilation failed"
fi

# Test 4.2: All tests pass
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -count=1 -timeout 120s ./... >/dev/null 2>&1); then
    pass "All Agentic module tests pass"
else
    fail "Agentic module tests failed"
fi

# Test 4.3: Test coverage >= 90%
COVERAGE_OUTPUT=$(cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -coverprofile=/tmp/agentic_coverage.out ./... 2>/dev/null)
if [ -f /tmp/agentic_coverage.out ]; then
    COVERAGE=$(cd "$MODULE_DIR" && go tool cover -func=/tmp/agentic_coverage.out 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
    COVERAGE_INT=${COVERAGE%%.*}
    if [ "$COVERAGE_INT" -ge 90 ]; then
        pass "Test coverage >= 90% (${COVERAGE}%)"
    else
        fail "Test coverage ${COVERAGE}% (expected >= 90%)"
    fi
    rm -f /tmp/agentic_coverage.out
else
    fail "Could not generate coverage report"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Agentic Module Challenge Results${NC}"
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
