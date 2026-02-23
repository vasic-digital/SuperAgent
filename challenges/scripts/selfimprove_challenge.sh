#!/usr/bin/env bash
# selfimprove_challenge.sh - Validates the SelfImprove module extraction and adapter integration
# Tests the digital.vasic.selfimprove module, its adapter, and integration with HelixAgent.
# Does NOT require running infrastructure.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$PROJECT_ROOT/SelfImprove"
ADAPTER_DIR="$PROJECT_ROOT/internal/adapters/selfimprove"
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
echo "SelfImprove Module Challenge"
echo "Validates digital.vasic.selfimprove extraction"
echo "=========================================="
echo ""

# ------------------------------------------------------------------
# Section 1: Module Directory Structure
# ------------------------------------------------------------------
echo "--- Section 1: Module Directory Structure ---"

check_dir "$MODULE_DIR" "SelfImprove/ module directory exists"
check_file "$MODULE_DIR/go.mod" "SelfImprove/go.mod exists"
check_file "$MODULE_DIR/README.md" "SelfImprove/README.md exists"
check_file "$MODULE_DIR/CLAUDE.md" "SelfImprove/CLAUDE.md exists"
check_file "$MODULE_DIR/AGENTS.md" "SelfImprove/AGENTS.md exists"

echo ""

# ------------------------------------------------------------------
# Section 2: Module Source Files
# ------------------------------------------------------------------
echo "--- Section 2: Module Source Files ---"

check_file "$MODULE_DIR/selfimprove/reward.go"      "SelfImprove/selfimprove/reward.go exists"
check_file "$MODULE_DIR/selfimprove/optimizer.go"   "SelfImprove/selfimprove/optimizer.go exists"
check_file "$MODULE_DIR/selfimprove/feedback.go"    "SelfImprove/selfimprove/feedback.go exists"
check_file "$MODULE_DIR/selfimprove/types.go"       "SelfImprove/selfimprove/types.go exists"
check_file "$MODULE_DIR/selfimprove/integration.go" "SelfImprove/selfimprove/integration.go exists"

echo ""

# ------------------------------------------------------------------
# Section 3: Module go.mod Content
# ------------------------------------------------------------------
echo "--- Section 3: Module go.mod Content ---"

check_grep "digital.vasic.selfimprove" "$MODULE_DIR/go.mod" \
    "go.mod declares module digital.vasic.selfimprove"

echo ""

# ------------------------------------------------------------------
# Section 4: Exported Types
# ------------------------------------------------------------------
echo "--- Section 4: Exported Types ---"

check_grep "type AIRewardModel struct" "$MODULE_DIR/selfimprove/reward.go" \
    "reward.go exports AIRewardModel struct"
check_grep "type TrainingExample struct" "$MODULE_DIR/selfimprove/types.go" \
    "types.go exports TrainingExample struct"
check_grep "type DimensionType" "$MODULE_DIR/selfimprove/types.go" \
    "types.go exports DimensionType"

echo ""

# ------------------------------------------------------------------
# Section 5: Challenge Script in Module
# ------------------------------------------------------------------
echo "--- Section 5: Challenge Script in Module ---"

check_file "$MODULE_DIR/challenges/scripts/selfimprove_challenge.sh" \
    "SelfImprove/challenges/scripts/selfimprove_challenge.sh exists"

echo ""

# ------------------------------------------------------------------
# Section 6: Adapter Files
# ------------------------------------------------------------------
echo "--- Section 6: Adapter Files ---"

check_file "$ADAPTER_DIR/adapter.go"      "internal/adapters/selfimprove/adapter.go exists"
check_file "$ADAPTER_DIR/adapter_test.go" "internal/adapters/selfimprove/adapter_test.go exists"

echo ""

# ------------------------------------------------------------------
# Section 7: Adapter Exported API
# ------------------------------------------------------------------
echo "--- Section 7: Adapter Exported API ---"

check_grep "func New(" "$ADAPTER_DIR/adapter.go" "adapter.go exports New function"
check_grep "func.*NewRewardModel(" "$ADAPTER_DIR/adapter.go" "adapter.go exports NewRewardModel method"
check_grep "func.*Train(" "$ADAPTER_DIR/adapter.go" "adapter.go exports Train method"
check_grep "digital.vasic.selfimprove" "$ADAPTER_DIR/adapter.go" \
    "adapter.go imports digital.vasic.selfimprove"

echo ""

# ------------------------------------------------------------------
# Section 8: Root go.mod Directives
# ------------------------------------------------------------------
echo "--- Section 8: Root go.mod Directives ---"

check_grep "digital.vasic.selfimprove" "$PROJECT_ROOT/go.mod" \
    "root go.mod has require directive for digital.vasic.selfimprove"
check_grep "digital.vasic.selfimprove => ./SelfImprove" "$PROJECT_ROOT/go.mod" \
    "root go.mod has replace directive for digital.vasic.selfimprove"

echo ""

# ------------------------------------------------------------------
# Section 9: Module Build
# ------------------------------------------------------------------
echo "--- Section 9: Module Build ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./...) &>/dev/null; then
    pass "SelfImprove module builds successfully (cd SelfImprove && go build ./...)"
else
    fail "SelfImprove module build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 10: Module Tests
# ------------------------------------------------------------------
echo "--- Section 10: Module Tests ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test ./... -short -count=1 -timeout 120s) &>/dev/null; then
    pass "SelfImprove module tests pass (go test ./... -short -count=1)"
else
    fail "SelfImprove module tests failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 11: Adapter Build
# ------------------------------------------------------------------
echo "--- Section 11: Adapter Build ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go build ./internal/adapters/selfimprove/) &>/dev/null; then
    pass "SelfImprove adapter builds successfully"
else
    fail "SelfImprove adapter build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 12: Adapter Tests
# ------------------------------------------------------------------
echo "--- Section 12: Adapter Tests ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -timeout 120s \
    ./internal/adapters/selfimprove/) &>/dev/null; then
    pass "SelfImprove adapter tests pass"
else
    fail "SelfImprove adapter tests failed"
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
