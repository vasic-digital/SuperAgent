#!/bin/bash
# HelixAgent Challenge - New Endpoints
# Validates that all new API endpoints (agentic, planning, llmops, benchmark)
# are properly registered in the router, have correct handler constructors,
# RegisterXRoutes functions, and corresponding test files.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "new-endpoints" "New Endpoints"
    load_env
    FRAMEWORK_LOADED=true
else
    FRAMEWORK_LOADED=false
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

record_result() {
    local test_name="$1"
    local status="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "${GREEN}[PASS]${NC} $test_name"
        if [ "$FRAMEWORK_LOADED" = "true" ]; then
            record_assertion "test" "$test_name" "true" ""
        fi
    else
        FAILED=$((FAILED + 1))
        echo -e "${RED}[FAIL]${NC} $test_name"
        if [ "$FRAMEWORK_LOADED" = "true" ]; then
            record_assertion "test" "$test_name" "false" "Test failed"
        fi
    fi
}

echo "=========================================="
echo "  New Endpoints Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: ROUTER REGISTRATION
# ============================================================================
echo -e "${BLUE}--- Section 1: Router Registration ---${NC}"

# Test 1: router.go contains agentic route group
if grep -q 'agenticHandler\|RegisterAgenticRoutes' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    record_result "router.go contains agentic route group" "PASS"
else
    record_result "router.go contains agentic route group" "FAIL"
fi

# Test 2: router.go contains planning route group
if grep -q 'planningHandler\|RegisterPlanningRoutes' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    record_result "router.go contains planning route group" "PASS"
else
    record_result "router.go contains planning route group" "FAIL"
fi

# Test 3: router.go contains llmops route group
if grep -q 'llmopsHandler\|RegisterLLMOpsRoutes' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    record_result "router.go contains llmops route group" "PASS"
else
    record_result "router.go contains llmops route group" "FAIL"
fi

# Test 4: router.go contains benchmark route group
if grep -q 'benchmarkHandler\|RegisterBenchmarkRoutes' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    record_result "router.go contains benchmark route group" "PASS"
else
    record_result "router.go contains benchmark route group" "FAIL"
fi

# ============================================================================
# SECTION 2: RegisterXRoutes FUNCTIONS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: RegisterXRoutes Functions ---${NC}"

# Test 5: agentic_handler.go has RegisterAgenticRoutes function
if grep -q 'func RegisterAgenticRoutes' "$PROJECT_ROOT/internal/handlers/agentic_handler.go" 2>/dev/null; then
    record_result "agentic_handler.go has RegisterAgenticRoutes function" "PASS"
else
    record_result "agentic_handler.go has RegisterAgenticRoutes function" "FAIL"
fi

# Test 6: planning_handler.go has RegisterPlanningRoutes function
if grep -q 'func RegisterPlanningRoutes' "$PROJECT_ROOT/internal/handlers/planning_handler.go" 2>/dev/null; then
    record_result "planning_handler.go has RegisterPlanningRoutes function" "PASS"
else
    record_result "planning_handler.go has RegisterPlanningRoutes function" "FAIL"
fi

# Test 7: llmops_handler.go has RegisterLLMOpsRoutes function
if grep -q 'func RegisterLLMOpsRoutes' "$PROJECT_ROOT/internal/handlers/llmops_handler.go" 2>/dev/null; then
    record_result "llmops_handler.go has RegisterLLMOpsRoutes function" "PASS"
else
    record_result "llmops_handler.go has RegisterLLMOpsRoutes function" "FAIL"
fi

# Test 8: benchmark_handler.go has RegisterBenchmarkRoutes function
if grep -q 'func RegisterBenchmarkRoutes' "$PROJECT_ROOT/internal/handlers/benchmark_handler.go" 2>/dev/null; then
    record_result "benchmark_handler.go has RegisterBenchmarkRoutes function" "PASS"
else
    record_result "benchmark_handler.go has RegisterBenchmarkRoutes function" "FAIL"
fi

# ============================================================================
# SECTION 3: HANDLER CONSTRUCTORS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: Handler Constructors ---${NC}"

# Test 9: NewAgenticHandler constructor exists
if grep -q 'func NewAgenticHandler' "$PROJECT_ROOT/internal/handlers/agentic_handler.go" 2>/dev/null; then
    record_result "NewAgenticHandler constructor exists" "PASS"
else
    record_result "NewAgenticHandler constructor exists" "FAIL"
fi

# Test 10: NewPlanningHandler constructor exists
if grep -q 'func NewPlanningHandler' "$PROJECT_ROOT/internal/handlers/planning_handler.go" 2>/dev/null; then
    record_result "NewPlanningHandler constructor exists" "PASS"
else
    record_result "NewPlanningHandler constructor exists" "FAIL"
fi

# Test 11: NewLLMOpsHandler constructor exists
if grep -q 'func NewLLMOpsHandler' "$PROJECT_ROOT/internal/handlers/llmops_handler.go" 2>/dev/null; then
    record_result "NewLLMOpsHandler constructor exists" "PASS"
else
    record_result "NewLLMOpsHandler constructor exists" "FAIL"
fi

# Test 12: NewBenchmarkHandler constructor exists
if grep -q 'func NewBenchmarkHandler' "$PROJECT_ROOT/internal/handlers/benchmark_handler.go" 2>/dev/null; then
    record_result "NewBenchmarkHandler constructor exists" "PASS"
else
    record_result "NewBenchmarkHandler constructor exists" "FAIL"
fi

# ============================================================================
# SECTION 4: TEST FILES
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: Test Files ---${NC}"

# Verify all handler test files exist and contain test functions
test_files_ok=0
test_files_total=0
for handler in agentic planning llmops benchmark; do
    test_files_total=$((test_files_total + 1))
    test_file="$PROJECT_ROOT/internal/handlers/${handler}_handler_test.go"
    if [ -f "$test_file" ] && grep -q 'func Test' "$test_file" 2>/dev/null; then
        test_files_ok=$((test_files_ok + 1))
    fi
done

# We count this as a single combined assertion per the spec (12 total)
# but validate all four in one pass
if [ "$test_files_ok" -eq "$test_files_total" ]; then
    record_result "All 4 handler test files exist and contain test functions" "PASS"
else
    record_result "All 4 handler test files exist and contain test functions ($test_files_ok/$test_files_total)" "FAIL"
fi

# ============================================================================
# SECTION 5: COMPILATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: Compilation ---${NC}"

# Compile check for all handlers
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go build "$PROJECT_ROOT/internal/handlers/..." > /tmp/new_endpoints_build.log 2>&1; then
    record_result "go build ./internal/handlers/... compiles cleanly" "PASS"
else
    record_result "go build ./internal/handlers/... compiles cleanly" "FAIL"
fi

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

if [ "$FRAMEWORK_LOADED" = "true" ]; then
    record_metric "total_tests" "$TOTAL"
    record_metric "passed_tests" "$PASSED"
    record_metric "failed_tests" "$FAILED"

    if [ $FAILED -eq 0 ]; then
        finalize_challenge "PASSED"
    else
        finalize_challenge "FAILED"
    fi
fi

if [ $FAILED -gt 0 ]; then
    exit 1
fi
exit 0
