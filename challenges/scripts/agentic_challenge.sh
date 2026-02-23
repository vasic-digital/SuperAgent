#!/usr/bin/env bash
# agentic_challenge.sh - Validates the Agentic module extraction and adapter integration
# Tests the digital.vasic.agentic module, its adapter, and integration with HelixAgent.
# Does NOT require running infrastructure.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$PROJECT_ROOT/Agentic"
ADAPTER_DIR="$PROJECT_ROOT/internal/adapters/agentic"
PASS=0
FAIL=0

# Resource limits per CLAUDE.md
export GOMAXPROCS=2

pass() {
    PASS=$((PASS + 1))
    echo "  [PASS] $1"
}

fail() {
    FAIL=$((FAIL + 1))
    echo "  [FAIL] $1"
}

check_file() {
    if [ -f "$1" ]; then
        pass "$2"
    else
        fail "$2"
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        pass "$2"
    else
        fail "$2"
    fi
}

check_grep() {
    if grep -q "$1" "$2" 2>/dev/null; then
        pass "$3"
    else
        fail "$3"
    fi
}

run_cmd() {
    local desc="$1"; shift
    if "$@" &>/dev/null; then
        pass "$desc"
    else
        fail "$desc"
    fi
}

echo "=========================================="
echo "Agentic Module Challenge"
echo "Validates digital.vasic.agentic extraction"
echo "=========================================="
echo ""

# ------------------------------------------------------------------
# Section 1: Module Directory Structure
# ------------------------------------------------------------------
echo "--- Section 1: Module Directory Structure ---"

check_dir "$MODULE_DIR" "Agentic/ module directory exists"
check_file "$MODULE_DIR/go.mod" "Agentic/go.mod exists"
check_file "$MODULE_DIR/README.md" "Agentic/README.md exists"
check_file "$MODULE_DIR/CLAUDE.md" "Agentic/CLAUDE.md exists"
check_file "$MODULE_DIR/AGENTS.md" "Agentic/AGENTS.md exists"

echo ""

# ------------------------------------------------------------------
# Section 2: Module Source Files
# ------------------------------------------------------------------
echo "--- Section 2: Module Source Files ---"

check_file "$MODULE_DIR/agentic/workflow.go" "Agentic/agentic/workflow.go exists"

echo ""

# ------------------------------------------------------------------
# Section 3: Module go.mod Content
# ------------------------------------------------------------------
echo "--- Section 3: Module go.mod Content ---"

check_grep "digital.vasic.agentic" "$MODULE_DIR/go.mod" "go.mod declares module digital.vasic.agentic"

echo ""

# ------------------------------------------------------------------
# Section 4: Exported Types in workflow.go
# ------------------------------------------------------------------
echo "--- Section 4: Exported Types in workflow.go ---"

check_grep "type Workflow struct" "$MODULE_DIR/agentic/workflow.go" "workflow.go exports Workflow struct"
check_grep "type WorkflowConfig struct" "$MODULE_DIR/agentic/workflow.go" "workflow.go exports WorkflowConfig struct"
check_grep "type WorkflowState struct" "$MODULE_DIR/agentic/workflow.go" "workflow.go exports WorkflowState struct"

echo ""

# ------------------------------------------------------------------
# Section 5: Challenge Script in Module
# ------------------------------------------------------------------
echo "--- Section 5: Challenge Script in Module ---"

check_file "$MODULE_DIR/challenges/scripts/agentic_workflow_challenge.sh" \
    "Agentic/challenges/scripts/agentic_workflow_challenge.sh exists"

echo ""

# ------------------------------------------------------------------
# Section 6: Adapter Files
# ------------------------------------------------------------------
echo "--- Section 6: Adapter Files ---"

check_file "$ADAPTER_DIR/adapter.go" "internal/adapters/agentic/adapter.go exists"
check_file "$ADAPTER_DIR/adapter_test.go" "internal/adapters/agentic/adapter_test.go exists"

echo ""

# ------------------------------------------------------------------
# Section 7: Adapter Exported API
# ------------------------------------------------------------------
echo "--- Section 7: Adapter Exported API ---"

check_grep "func New(" "$ADAPTER_DIR/adapter.go" "adapter.go exports New function"
check_grep "func.*NewWorkflow(" "$ADAPTER_DIR/adapter.go" "adapter.go exports NewWorkflow method"
check_grep "digital.vasic.agentic" "$ADAPTER_DIR/adapter.go" "adapter.go imports digital.vasic.agentic"

echo ""

# ------------------------------------------------------------------
# Section 8: Root go.mod Directives
# ------------------------------------------------------------------
echo "--- Section 8: Root go.mod Directives ---"

check_grep "digital.vasic.agentic" "$PROJECT_ROOT/go.mod" \
    "root go.mod has require directive for digital.vasic.agentic"
check_grep "digital.vasic.agentic => ./Agentic" "$PROJECT_ROOT/go.mod" \
    "root go.mod has replace directive for digital.vasic.agentic"

echo ""

# ------------------------------------------------------------------
# Section 9: Module Build
# ------------------------------------------------------------------
echo "--- Section 9: Module Build ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./...) &>/dev/null; then
    pass "Agentic module builds successfully (cd Agentic && go build ./...)"
else
    fail "Agentic module build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 10: Module Tests
# ------------------------------------------------------------------
echo "--- Section 10: Module Tests ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test ./... -short -count=1 -timeout 120s) &>/dev/null; then
    pass "Agentic module tests pass (go test ./... -short -count=1)"
else
    fail "Agentic module tests failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 11: Adapter Build
# ------------------------------------------------------------------
echo "--- Section 11: Adapter Build ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go build ./internal/adapters/agentic/) &>/dev/null; then
    pass "Agentic adapter builds successfully"
else
    fail "Agentic adapter build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 12: Adapter Tests
# ------------------------------------------------------------------
echo "--- Section 12: Adapter Tests ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -timeout 120s \
    ./internal/adapters/agentic/) &>/dev/null; then
    pass "Agentic adapter tests pass"
else
    fail "Agentic adapter tests failed"
fi

echo ""

# ------------------------------------------------------------------
# Summary
# ------------------------------------------------------------------
TOTAL=$((PASS + FAIL))
echo "=========================================="
echo "SUMMARY"
echo "=========================================="
echo "Results: $PASS passed, $FAIL failed out of $TOTAL total"
echo "=========================================="

if [ "$FAIL" -eq 0 ]; then
    echo "ALL CHECKS PASSED"
    exit 0
else
    echo "CHALLENGE FAILED"
    exit 1
fi
