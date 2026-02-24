#!/bin/bash
# HelixAgent Challenge: Benchmark Module
# Tests: ~12 tests across 4 sections
# Validates: Module build, benchmark runner, types, built-in benchmarks,
#            key interfaces, tests, coverage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

MODULE_DIR="$PROJECT_ROOT/Benchmark"
PKG_DIR="$MODULE_DIR/benchmark"

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
   grep -q 'module digital.vasic.benchmark' "$MODULE_DIR/go.mod"; then
    pass "go.mod exists with module digital.vasic.benchmark"
else
    fail "go.mod missing or incorrect module name"
fi

# Test 1.2: All source files exist (types, runner, integration)
if [ -f "$PKG_DIR/types.go" ] && \
   [ -f "$PKG_DIR/runner.go" ] && \
   [ -f "$PKG_DIR/integration.go" ]; then
    pass "All 3 source files exist (types, runner, integration)"
else
    fail "Missing one or more source files"
fi

# Test 1.3: Test files exist
if [ -f "$PKG_DIR/runner_test.go" ] && \
   [ -f "$PKG_DIR/types_test.go" ]; then
    pass "Test files exist (runner_test.go, types_test.go)"
else
    fail "Missing one or more test files"
fi

#===============================================================================
# Section 2: Benchmark Runner (3 tests)
#===============================================================================
section "Section 2: Benchmark Runner"

# Test 2.1: BenchmarkRunner interface exists
if grep -q 'type BenchmarkRunner interface' "$PKG_DIR/types.go"; then
    pass "BenchmarkRunner interface defined"
else
    fail "BenchmarkRunner interface missing"
fi

# Test 2.2: StandardBenchmarkRunner implementation and constructor
if grep -q 'type StandardBenchmarkRunner struct' "$PKG_DIR/runner.go" && \
   grep -q 'func NewStandardBenchmarkRunner(' "$PKG_DIR/runner.go"; then
    pass "StandardBenchmarkRunner implementation and constructor exist"
else
    fail "StandardBenchmarkRunner or constructor missing"
fi

# Test 2.3: Key runner methods (ListBenchmarks, CreateRun, StartRun, CompareRuns)
if grep -q 'func (r \*StandardBenchmarkRunner) ListBenchmarks(' "$PKG_DIR/runner.go" && \
   grep -q 'func (r \*StandardBenchmarkRunner) CreateRun(' "$PKG_DIR/runner.go" && \
   grep -q 'func (r \*StandardBenchmarkRunner) StartRun(' "$PKG_DIR/runner.go" && \
   grep -q 'func (r \*StandardBenchmarkRunner) CompareRuns(' "$PKG_DIR/runner.go"; then
    pass "Key runner methods exist (ListBenchmarks, CreateRun, StartRun, CompareRuns)"
else
    fail "One or more key runner methods missing"
fi

#===============================================================================
# Section 3: Built-in Benchmarks and Types (3 tests)
#===============================================================================
section "Section 3: Built-in Benchmarks and Types"

# Test 3.1: All benchmark type constants registered
if grep -q 'BenchmarkTypeSWEBench' "$PKG_DIR/types.go" && \
   grep -q 'BenchmarkTypeHumanEval' "$PKG_DIR/types.go" && \
   grep -q 'BenchmarkTypeMMLU' "$PKG_DIR/types.go" && \
   grep -q 'BenchmarkTypeGSM8K' "$PKG_DIR/types.go" && \
   grep -q 'BenchmarkTypeCustom' "$PKG_DIR/types.go"; then
    pass "Benchmark type constants defined (SWEBench, HumanEval, MMLU, GSM8K, Custom)"
else
    fail "One or more benchmark type constants missing"
fi

# Test 3.2: Built-in benchmarks initialized in runner (SWE-Bench, HumanEval, MMLU, GSM8K)
if grep -q 'initBuiltInBenchmarks' "$PKG_DIR/runner.go" && \
   grep -q 'createSWEBenchTasks' "$PKG_DIR/runner.go" && \
   grep -q 'createHumanEvalTasks' "$PKG_DIR/runner.go" && \
   grep -q 'createMMLUTasks' "$PKG_DIR/runner.go" && \
   grep -q 'createGSM8KTasks' "$PKG_DIR/runner.go"; then
    pass "Built-in benchmarks initialized (SWE-Bench, HumanEval, MMLU, GSM8K)"
else
    fail "Built-in benchmark initialization missing"
fi

# Test 3.3: Key data types (BenchmarkTask, BenchmarkResult, BenchmarkRun, BenchmarkSummary)
if grep -q 'type BenchmarkTask struct' "$PKG_DIR/types.go" && \
   grep -q 'type BenchmarkResult struct' "$PKG_DIR/types.go" && \
   grep -q 'type BenchmarkRun struct' "$PKG_DIR/types.go" && \
   grep -q 'type BenchmarkSummary struct' "$PKG_DIR/types.go"; then
    pass "Key data types defined (BenchmarkTask, BenchmarkResult, BenchmarkRun, BenchmarkSummary)"
else
    fail "One or more key data types missing"
fi

#===============================================================================
# Section 4: Build, Tests, and Coverage (3 tests)
#===============================================================================
section "Section 4: Build, Tests, and Coverage"

# Test 4.1: Module compiles successfully
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go build ./... >/dev/null 2>&1); then
    pass "Benchmark module compiles successfully"
else
    fail "Benchmark module compilation failed"
fi

# Test 4.2: All tests pass
if (cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -count=1 -timeout 120s ./... >/dev/null 2>&1); then
    pass "All Benchmark module tests pass"
else
    fail "Benchmark module tests failed"
fi

# Test 4.3: Test coverage >= 90%
COVERAGE_OUTPUT=$(cd "$MODULE_DIR" && GOMAXPROCS=2 nice -n 19 go test -coverprofile=/tmp/benchmark_coverage.out ./... 2>/dev/null)
if [ -f /tmp/benchmark_coverage.out ]; then
    COVERAGE=$(cd "$MODULE_DIR" && go tool cover -func=/tmp/benchmark_coverage.out 2>/dev/null | grep total | awk '{print $3}' | tr -d '%')
    COVERAGE_INT=${COVERAGE%%.*}
    if [ "$COVERAGE_INT" -ge 90 ]; then
        pass "Test coverage >= 90% (${COVERAGE}%)"
    else
        fail "Test coverage ${COVERAGE}% (expected >= 90%)"
    fi
    rm -f /tmp/benchmark_coverage.out
else
    fail "Could not generate coverage report"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Benchmark Module Challenge Results${NC}"
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
