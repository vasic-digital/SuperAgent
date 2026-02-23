#!/usr/bin/env bash
# benchmark_challenge.sh - Validates the Benchmark module extraction and adapter integration
# Tests the digital.vasic.benchmark module, its adapter, and integration with HelixAgent.
# Does NOT require running infrastructure.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$PROJECT_ROOT/Benchmark"
ADAPTER_DIR="$PROJECT_ROOT/internal/adapters/benchmark"
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
echo "Benchmark Module Challenge"
echo "Validates digital.vasic.benchmark extraction"
echo "=========================================="
echo ""

# ------------------------------------------------------------------
# Section 1: Module Directory Structure
# ------------------------------------------------------------------
echo "--- Section 1: Module Directory Structure ---"

check_dir "$MODULE_DIR" "Benchmark/ module directory exists"
check_file "$MODULE_DIR/go.mod" "Benchmark/go.mod exists"
check_file "$MODULE_DIR/README.md" "Benchmark/README.md exists"
check_file "$MODULE_DIR/CLAUDE.md" "Benchmark/CLAUDE.md exists"
check_file "$MODULE_DIR/AGENTS.md" "Benchmark/AGENTS.md exists"

echo ""

# ------------------------------------------------------------------
# Section 2: Module Source Files
# ------------------------------------------------------------------
echo "--- Section 2: Module Source Files ---"

check_file "$MODULE_DIR/benchmark/runner.go"      "Benchmark/benchmark/runner.go exists"
check_file "$MODULE_DIR/benchmark/types.go"       "Benchmark/benchmark/types.go exists"
check_file "$MODULE_DIR/benchmark/integration.go" "Benchmark/benchmark/integration.go exists"

echo ""

# ------------------------------------------------------------------
# Section 3: Module go.mod Content
# ------------------------------------------------------------------
echo "--- Section 3: Module go.mod Content ---"

check_grep "digital.vasic.benchmark" "$MODULE_DIR/go.mod" "go.mod declares module digital.vasic.benchmark"

echo ""

# ------------------------------------------------------------------
# Section 4: Exported Types
# ------------------------------------------------------------------
echo "--- Section 4: Exported Types ---"

check_grep "type BenchmarkType" "$MODULE_DIR/benchmark/types.go" \
    "types.go exports BenchmarkType"
check_grep "type StandardBenchmarkRunner struct" "$MODULE_DIR/benchmark/runner.go" \
    "runner.go exports StandardBenchmarkRunner struct"
check_grep "type BenchmarkSystem struct" "$MODULE_DIR/benchmark/integration.go" \
    "integration.go exports BenchmarkSystem struct"

echo ""

# ------------------------------------------------------------------
# Section 5: Challenge Script in Module
# ------------------------------------------------------------------
echo "--- Section 5: Challenge Script in Module ---"

check_file "$MODULE_DIR/challenges/scripts/benchmark_challenge.sh" \
    "Benchmark/challenges/scripts/benchmark_challenge.sh exists"

echo ""

# ------------------------------------------------------------------
# Section 6: Adapter Files
# ------------------------------------------------------------------
echo "--- Section 6: Adapter Files ---"

check_file "$ADAPTER_DIR/adapter.go"      "internal/adapters/benchmark/adapter.go exists"
check_file "$ADAPTER_DIR/adapter_test.go" "internal/adapters/benchmark/adapter_test.go exists"

echo ""

# ------------------------------------------------------------------
# Section 7: Adapter Exported API
# ------------------------------------------------------------------
echo "--- Section 7: Adapter Exported API ---"

check_grep "func New(" "$ADAPTER_DIR/adapter.go" "adapter.go exports New function"
check_grep "func.*Initialize(" "$ADAPTER_DIR/adapter.go" "adapter.go exports Initialize method"
check_grep "func.*ListBenchmarks(" "$ADAPTER_DIR/adapter.go" "adapter.go exports ListBenchmarks method"
check_grep "digital.vasic.benchmark" "$ADAPTER_DIR/adapter.go" \
    "adapter.go imports digital.vasic.benchmark"

echo ""

# ------------------------------------------------------------------
# Section 8: Root go.mod Directives
# ------------------------------------------------------------------
echo "--- Section 8: Root go.mod Directives ---"

check_grep "digital.vasic.benchmark" "$PROJECT_ROOT/go.mod" \
    "root go.mod has require directive for digital.vasic.benchmark"
check_grep "digital.vasic.benchmark => ./Benchmark" "$PROJECT_ROOT/go.mod" \
    "root go.mod has replace directive for digital.vasic.benchmark"

echo ""

# ------------------------------------------------------------------
# Section 9: Module Build
# ------------------------------------------------------------------
echo "--- Section 9: Module Build ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./...) &>/dev/null; then
    pass "Benchmark module builds successfully (cd Benchmark && go build ./...)"
else
    fail "Benchmark module build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 10: Module Tests
# ------------------------------------------------------------------
echo "--- Section 10: Module Tests ---"

if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test ./... -short -count=1 -timeout 120s) &>/dev/null; then
    pass "Benchmark module tests pass (go test ./... -short -count=1)"
else
    fail "Benchmark module tests failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 11: Adapter Build
# ------------------------------------------------------------------
echo "--- Section 11: Adapter Build ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go build ./internal/adapters/benchmark/) &>/dev/null; then
    pass "Benchmark adapter builds successfully"
else
    fail "Benchmark adapter build failed"
fi

echo ""

# ------------------------------------------------------------------
# Section 12: Adapter Tests
# ------------------------------------------------------------------
echo "--- Section 12: Adapter Tests ---"

if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -timeout 120s \
    ./internal/adapters/benchmark/) &>/dev/null; then
    pass "Benchmark adapter tests pass"
else
    fail "Benchmark adapter tests failed"
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
