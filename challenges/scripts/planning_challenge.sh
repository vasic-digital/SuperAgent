#!/usr/bin/env bash
# planning_challenge.sh - Validates the Planning module extraction and adapter integration
# Tests the digital.vasic.planning module, its adapter, and integration with HelixAgent.
# Does NOT require running infrastructure.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$PROJECT_ROOT/Planning"
ADAPTER_DIR="$PROJECT_ROOT/internal/adapters/planning"
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

echo "=========================================="
echo "Planning Module Challenge"
echo "Validates digital.vasic.planning extraction"
echo "=========================================="
echo ""

# ------------------------------------------------------------------
# Section 1: Module Directory Structure
# ------------------------------------------------------------------
echo "--- Section 1: Module Directory Structure ---"

check_dir "$MODULE_DIR" "Planning/ module directory exists"
check_file "$MODULE_DIR/go.mod" "Planning/go.mod exists"
check_file "$MODULE_DIR/README.md" "Planning/README.md exists"
check_file "$MODULE_DIR/CLAUDE.md" "Planning/CLAUDE.md exists"
check_file "$MODULE_DIR/AGENTS.md" "Planning/AGENTS.md exists"

echo ""

# ------------------------------------------------------------------
# Section 2: Module Source Files
# ------------------------------------------------------------------
echo "--- Section 2: Module Source Files ---"

check_file "$MODULE_DIR/planning/hiplan.go"           "Planning/planning/hiplan.go exists"
check_file "$MODULE_DIR/planning/mcts.go"             "Planning/planning/mcts.go exists"
check_file "$MODULE_DIR/planning/tree_of_thoughts.go" "Planning/planning/tree_of_thoughts.go exists"

echo ""

# ------------------------------------------------------------------
# Section 3: Module go.mod Content
# ------------------------------------------------------------------
echo "--- Section 3: Module go.mod Content ---"

check_grep "digital.vasic.planning" "$MODULE_DIR/go.mod" "go.mod declares module digital.vasic.planning"

echo ""

# ------------------------------------------------------------------
# Section 4: Exported Types
# ------------------------------------------------------------------
echo "--- Section 4: Exported Types ---"

check_grep "type HiPlan struct" "$MODULE_DIR/planning/hiplan.go" \
    "hiplan.go exports HiPlan struct"
check_grep "type MCTS struct" "$MODULE_DIR/planning/mcts.go" \
    "mcts.go exports MCTS struct"
check_grep "type TreeOfThoughts struct" "$MODULE_DIR/planning/tree_of_thoughts.go" \
    "tree_of_thoughts.go exports TreeOfThoughts struct"

echo ""

# ------------------------------------------------------------------
# Section 5: Challenge Script in Module
# ------------------------------------------------------------------
echo "--- Section 5: Challenge Script in Module ---"

check_file "$MODULE_DIR/challenges/scripts/planning_challenge.sh" \
    "Planning/challenges/scripts/planning_challenge.sh exists"

echo ""

# ------------------------------------------------------------------
# Section 6: Adapter Files
# ------------------------------------------------------------------
echo "--- Section 6: Adapter Files ---"

check_file "$ADAPTER_DIR/adapter.go"      "internal/adapters/planning/adapter.go exists"
check_file "$ADAPTER_DIR/adapter_test.go" "internal/adapters/planning/adapter_test.go exists"

echo ""

# ------------------------------------------------------------------
# Section 7: Adapter Exported API
# ------------------------------------------------------------------
echo "--- Section 7: Adapter Exported API ---"

check_grep "func New(" "$ADAPTER_DIR/adapter.go" "adapter.go exports New function"
check_grep "func.*NewHiPlan(" "$ADAPTER_DIR/adapter.go" "adapter.go exports NewHiPlan method"
check_grep "func.*NewMCTS(" "$ADAPTER_DIR/adapter.go" "adapter.go exports NewMCTS method"
check_grep "func.*NewTreeOfThoughts(" "$ADAPTER_DIR/adapter.go" "adapter.go exports NewTreeOfThoughts method"
check_grep "digital.vasic.planning" "$ADAPTER_DIR/adapter.go" \
    "adapter.go imports digital.vasic.planning"

echo ""

# ------------------------------------------------------------------
# Section 8: Root go.mod Directives
# ------------------------------------------------------------------
echo "--- Section 8: Root go.mod Directives ---"

check_grep "digital.vasic.planning" "$PROJECT_ROOT/go.mod" \
    "root go.mod has require directive for digital.vasic.planning"
check_grep "digital.vasic.planning => ./Planning" "$PROJECT_ROOT/go.mod" \
    "root go.mod has replace directive for digital.vasic.planning"

echo ""

# ------------------------------------------------------------------
# Section 9: Module Build
# ------------------------------------------------------------------
echo "--- Section 9: Module Build ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./...) &>/dev/null; then
    pass "Planning module builds successfully (cd Planning && go build ./...)"
else
    fail "Planning module build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 10: Module Tests
# ------------------------------------------------------------------
echo "--- Section 10: Module Tests ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test ./... -short -count=1 -timeout 120s) &>/dev/null; then
    pass "Planning module tests pass (go test ./... -short -count=1)"
else
    fail "Planning module tests failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 11: Adapter Build
# ------------------------------------------------------------------
echo "--- Section 11: Adapter Build ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go build ./internal/adapters/planning/) &>/dev/null; then
    pass "Planning adapter builds successfully"
else
    fail "Planning adapter build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 12: Adapter Tests
# ------------------------------------------------------------------
echo "--- Section 12: Adapter Tests ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -timeout 120s \
    ./internal/adapters/planning/) &>/dev/null; then
    pass "Planning adapter tests pass"
else
    fail "Planning adapter tests failed"
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
