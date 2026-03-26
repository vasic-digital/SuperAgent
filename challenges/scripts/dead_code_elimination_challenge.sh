#!/bin/bash
# HelixAgent Challenge - Dead Code Elimination
# Validates that unused packages have been removed and all remaining packages
# are properly imported by handlers or adapters, compile cleanly, and are
# registered in the router.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "dead-code-elimination" "Dead Code Elimination"
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
echo "  Dead Code Elimination Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: REMOVED PACKAGES NO LONGER EXIST
# ============================================================================
echo -e "${BLUE}--- Section 1: Removed Packages ---${NC}"

# Test 1: internal/embedding/ directory does NOT exist (was removed, replaced by Embeddings module)
if [ ! -d "$PROJECT_ROOT/internal/embedding" ]; then
    record_result "internal/embedding/ directory does NOT exist (correctly removed)" "PASS"
else
    record_result "internal/embedding/ directory does NOT exist (correctly removed)" "FAIL"
fi

# ============================================================================
# SECTION 2: PACKAGES ARE PROPERLY IMPORTED BY HANDLERS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: Package Imports in Handlers ---${NC}"

# Test 2: internal/agentic/ is imported by agentic_handler.go
if grep -q '"dev.helix.agent/internal/agentic"' "$PROJECT_ROOT/internal/handlers/agentic_handler.go" 2>/dev/null; then
    record_result "internal/agentic/ is imported by agentic_handler.go" "PASS"
else
    record_result "internal/agentic/ is imported by agentic_handler.go" "FAIL"
fi

# Test 3: internal/benchmark/ is imported by benchmark_handler.go
if grep -q '"dev.helix.agent/internal/benchmark"' "$PROJECT_ROOT/internal/handlers/benchmark_handler.go" 2>/dev/null; then
    record_result "internal/benchmark/ is imported by benchmark_handler.go" "PASS"
else
    record_result "internal/benchmark/ is imported by benchmark_handler.go" "FAIL"
fi

# Test 4: internal/llmops/ is imported by llmops_handler.go
if grep -q '"dev.helix.agent/internal/llmops"' "$PROJECT_ROOT/internal/handlers/llmops_handler.go" 2>/dev/null; then
    record_result "internal/llmops/ is imported by llmops_handler.go" "PASS"
else
    record_result "internal/llmops/ is imported by llmops_handler.go" "FAIL"
fi

# Test 5: internal/planning/ is imported by planning_handler.go
if grep -q '"dev.helix.agent/internal/planning"' "$PROJECT_ROOT/internal/handlers/planning_handler.go" 2>/dev/null; then
    record_result "internal/planning/ is imported by planning_handler.go" "PASS"
else
    record_result "internal/planning/ is imported by planning_handler.go" "FAIL"
fi

# ============================================================================
# SECTION 3: PACKAGES ARE PROPERLY IMPORTED BY ADAPTERS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: Package Imports in Adapters ---${NC}"

# Test 6: internal/adapters/mcp/ adapter exists (observability/events adapters were removed in a0258674)
if [ -d "$PROJECT_ROOT/internal/adapters/mcp" ]; then
    record_result "internal/adapters/mcp/ adapter directory exists" "PASS"
else
    record_result "internal/adapters/mcp/ adapter directory exists" "FAIL"
fi

# Test 7: internal/adapters/security/ adapter exists
if [ -d "$PROJECT_ROOT/internal/adapters/security" ]; then
    record_result "internal/adapters/security/ adapter directory exists" "PASS"
else
    record_result "internal/adapters/security/ adapter directory exists" "FAIL"
fi

# ============================================================================
# SECTION 4: HANDLER FILES EXIST
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: Handler Files Exist ---${NC}"

# Test 8: agentic_handler.go exists
if [ -f "$PROJECT_ROOT/internal/handlers/agentic_handler.go" ]; then
    record_result "agentic_handler.go exists" "PASS"
else
    record_result "agentic_handler.go exists" "FAIL"
fi

# Test 9: planning_handler.go exists
if [ -f "$PROJECT_ROOT/internal/handlers/planning_handler.go" ]; then
    record_result "planning_handler.go exists" "PASS"
else
    record_result "planning_handler.go exists" "FAIL"
fi

# Test 10: llmops_handler.go exists
if [ -f "$PROJECT_ROOT/internal/handlers/llmops_handler.go" ]; then
    record_result "llmops_handler.go exists" "PASS"
else
    record_result "llmops_handler.go exists" "FAIL"
fi

# Test 11: benchmark_handler.go exists
if [ -f "$PROJECT_ROOT/internal/handlers/benchmark_handler.go" ]; then
    record_result "benchmark_handler.go exists" "PASS"
else
    record_result "benchmark_handler.go exists" "FAIL"
fi

# ============================================================================
# SECTION 5: ADAPTER FILES EXIST
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: Adapter Files Exist ---${NC}"

# Test 12: adapters/containers/adapter.go exists (centralized container management)
if [ -f "$PROJECT_ROOT/internal/adapters/containers/adapter.go" ]; then
    record_result "adapters/containers/adapter.go exists" "PASS"
else
    record_result "adapters/containers/adapter.go exists" "FAIL"
fi

# Test 13: adapters/mcp/adapter.go or similar file exists in adapters/mcp/
if ls "$PROJECT_ROOT/internal/adapters/mcp/"*.go > /dev/null 2>&1; then
    record_result "adapters/mcp/ contains Go source files" "PASS"
else
    record_result "adapters/mcp/ contains Go source files" "FAIL"
fi

# ============================================================================
# SECTION 6: COMPILATION AND ROUTER REGISTRATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 6: Compilation and Router Registration ---${NC}"

# Test 14: All handler packages compile cleanly
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go build "$PROJECT_ROOT/internal/handlers/..." > /tmp/dead_code_handlers_build.log 2>&1; then
    record_result "internal/handlers/ compiles cleanly" "PASS"
else
    record_result "internal/handlers/ compiles cleanly" "FAIL"
fi

# Test 15: router.go references the new handlers (agentic, planning, llmops, benchmark)
router_refs=0
for handler_name in "agenticHandler" "planningHandler" "llmopsHandler" "benchmarkHandler"; do
    if grep -q "$handler_name" "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
        router_refs=$((router_refs + 1))
    fi
done
if [ "$router_refs" -eq 4 ]; then
    record_result "router.go references all 4 new handlers (agentic, planning, llmops, benchmark)" "PASS"
else
    record_result "router.go references all 4 new handlers (agentic, planning, llmops, benchmark)" "FAIL"
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
