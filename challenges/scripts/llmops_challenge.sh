#!/usr/bin/env bash
# llmops_challenge.sh - Validates the LLMOps module extraction and adapter integration
# Tests the digital.vasic.llmops module, its adapter, and integration with HelixAgent.
# Does NOT require running infrastructure.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$PROJECT_ROOT/LLMOps"
ADAPTER_DIR="$PROJECT_ROOT/internal/adapters/llmops"
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
echo "LLMOps Module Challenge"
echo "Validates digital.vasic.llmops extraction"
echo "=========================================="
echo ""

# ------------------------------------------------------------------
# Section 1: Module Directory Structure
# ------------------------------------------------------------------
echo "--- Section 1: Module Directory Structure ---"

check_dir "$MODULE_DIR" "LLMOps/ module directory exists"
check_file "$MODULE_DIR/go.mod" "LLMOps/go.mod exists"
check_file "$MODULE_DIR/README.md" "LLMOps/README.md exists"
check_file "$MODULE_DIR/CLAUDE.md" "LLMOps/CLAUDE.md exists"
check_file "$MODULE_DIR/AGENTS.md" "LLMOps/AGENTS.md exists"

echo ""

# ------------------------------------------------------------------
# Section 2: Module Source Files
# ------------------------------------------------------------------
echo "--- Section 2: Module Source Files ---"

check_file "$MODULE_DIR/llmops/evaluator.go"    "LLMOps/llmops/evaluator.go exists"
check_file "$MODULE_DIR/llmops/experiments.go"  "LLMOps/llmops/experiments.go exists"
check_file "$MODULE_DIR/llmops/prompts.go"      "LLMOps/llmops/prompts.go exists"
check_file "$MODULE_DIR/llmops/types.go"        "LLMOps/llmops/types.go exists"
check_file "$MODULE_DIR/llmops/integration.go"  "LLMOps/llmops/integration.go exists"

echo ""

# ------------------------------------------------------------------
# Section 3: Module go.mod Content
# ------------------------------------------------------------------
echo "--- Section 3: Module go.mod Content ---"

check_grep "digital.vasic.llmops" "$MODULE_DIR/go.mod" "go.mod declares module digital.vasic.llmops"

echo ""

# ------------------------------------------------------------------
# Section 4: Exported Types
# ------------------------------------------------------------------
echo "--- Section 4: Exported Types ---"

check_grep "type InMemoryContinuousEvaluator struct" "$MODULE_DIR/llmops/evaluator.go" \
    "evaluator.go exports InMemoryContinuousEvaluator struct"
check_grep "type InMemoryExperimentManager struct" "$MODULE_DIR/llmops/experiments.go" \
    "experiments.go exports InMemoryExperimentManager struct"
check_grep "type Dataset struct" "$MODULE_DIR/llmops/types.go" \
    "types.go exports Dataset struct"

echo ""

# ------------------------------------------------------------------
# Section 5: Challenge Script in Module
# ------------------------------------------------------------------
echo "--- Section 5: Challenge Script in Module ---"

check_file "$MODULE_DIR/challenges/scripts/llmops_challenge.sh" \
    "LLMOps/challenges/scripts/llmops_challenge.sh exists"

echo ""

# ------------------------------------------------------------------
# Section 6: Adapter Files
# ------------------------------------------------------------------
echo "--- Section 6: Adapter Files ---"

check_file "$ADAPTER_DIR/adapter.go"      "internal/adapters/llmops/adapter.go exists"
check_file "$ADAPTER_DIR/adapter_test.go" "internal/adapters/llmops/adapter_test.go exists"

echo ""

# ------------------------------------------------------------------
# Section 7: Adapter Exported API
# ------------------------------------------------------------------
echo "--- Section 7: Adapter Exported API ---"

check_grep "func New(" "$ADAPTER_DIR/adapter.go" "adapter.go exports New function"
check_grep "func.*NewEvaluator(" "$ADAPTER_DIR/adapter.go" "adapter.go exports NewEvaluator method"
check_grep "digital.vasic.llmops" "$ADAPTER_DIR/adapter.go" "adapter.go imports digital.vasic.llmops"

echo ""

# ------------------------------------------------------------------
# Section 8: Root go.mod Directives
# ------------------------------------------------------------------
echo "--- Section 8: Root go.mod Directives ---"

check_grep "digital.vasic.llmops" "$PROJECT_ROOT/go.mod" \
    "root go.mod has require directive for digital.vasic.llmops"
check_grep "digital.vasic.llmops => ./LLMOps" "$PROJECT_ROOT/go.mod" \
    "root go.mod has replace directive for digital.vasic.llmops"

echo ""

# ------------------------------------------------------------------
# Section 9: Module Build
# ------------------------------------------------------------------
echo "--- Section 9: Module Build ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./...) &>/dev/null; then
    pass "LLMOps module builds successfully (cd LLMOps && go build ./...)"
else
    fail "LLMOps module build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 10: Module Tests
# ------------------------------------------------------------------
echo "--- Section 10: Module Tests ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test ./... -short -count=1 -timeout 120s) &>/dev/null; then
    pass "LLMOps module tests pass (go test ./... -short -count=1)"
else
    fail "LLMOps module tests failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 11: Adapter Build
# ------------------------------------------------------------------
echo "--- Section 11: Adapter Build ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go build ./internal/adapters/llmops/) &>/dev/null; then
    pass "LLMOps adapter builds successfully"
else
    fail "LLMOps adapter build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 12: Adapter Tests
# ------------------------------------------------------------------
echo "--- Section 12: Adapter Tests ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -timeout 120s \
    ./internal/adapters/llmops/) &>/dev/null; then
    pass "LLMOps adapter tests pass"
else
    fail "LLMOps adapter tests failed"
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
