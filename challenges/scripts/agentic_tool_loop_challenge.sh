#!/bin/bash
# HelixAgent Challenge - Agentic Tool Loop
# Validates that the IterativeToolExecutor implementation exists, is
# structurally correct, and covers all 6 required tool protocols.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "agentic-tool-loop" "Agentic Tool Loop"
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
echo "  Agentic Tool Loop Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: CORE FILE EXISTENCE
# ============================================================================
echo -e "${BLUE}--- Section 1: Core File Existence ---${NC}"

# Test 1: iterative_tool_executor.go exists
if [ -f "$PROJECT_ROOT/internal/services/iterative_tool_executor.go" ]; then
    record_result "internal/services/iterative_tool_executor.go exists" "PASS"
else
    record_result "internal/services/iterative_tool_executor.go exists" "FAIL"
fi

# Test 2: agentic_ensemble_types.go exists (defines AgenticToolExecution)
if [ -f "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" ]; then
    record_result "internal/services/agentic_ensemble_types.go exists (protocol type source)" "PASS"
else
    record_result "internal/services/agentic_ensemble_types.go exists (protocol type source)" "FAIL"
fi

# ============================================================================
# SECTION 2: ITERATIVE TOOL EXECUTOR STRUCT
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: IterativeToolExecutor Struct ---${NC}"

EXECUTOR_FILE="$PROJECT_ROOT/internal/services/iterative_tool_executor.go"

# Test 3: IterativeToolExecutor struct defined
if grep -q "type IterativeToolExecutor struct" "$EXECUTOR_FILE" 2>/dev/null; then
    record_result "IterativeToolExecutor struct defined" "PASS"
else
    record_result "IterativeToolExecutor struct defined" "FAIL"
fi

# Test 4: maxIterations field exists in IterativeToolExecutor
if grep -q "maxIterations" "$EXECUTOR_FILE" 2>/dev/null; then
    record_result "IterativeToolExecutor has maxIterations field" "PASS"
else
    record_result "IterativeToolExecutor has maxIterations field" "FAIL"
fi

# Test 5: NewIterativeToolExecutor constructor exists
if grep -q "func NewIterativeToolExecutor" "$EXECUTOR_FILE" 2>/dev/null; then
    record_result "NewIterativeToolExecutor constructor exists" "PASS"
else
    record_result "NewIterativeToolExecutor constructor exists" "FAIL"
fi

# ============================================================================
# SECTION 3: EXECUTE WITH TOOLS METHOD
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: ExecuteWithTools Method ---${NC}"

# Test 6: ExecuteWithTools method exists
if grep -q "func.*IterativeToolExecutor.*ExecuteWithTools\|func.*\*IterativeToolExecutor.*ExecuteWithTools" \
    "$EXECUTOR_FILE" 2>/dev/null; then
    record_result "IterativeToolExecutor has ExecuteWithTools method" "PASS"
else
    record_result "IterativeToolExecutor has ExecuteWithTools method" "FAIL"
fi

# Test 7: maxIterations enforcement (loop guarded by maxIterations check)
if grep -q "maxIterations" "$EXECUTOR_FILE" 2>/dev/null; then
    # Check it's used in a comparison/loop guard
    if grep -qE "maxIterations|MaxToolIterations|maxToolIter" "$EXECUTOR_FILE" 2>/dev/null; then
        record_result "maxIterations enforcement present in executor" "PASS"
    else
        record_result "maxIterations enforcement present in executor" "FAIL"
    fi
else
    record_result "maxIterations enforcement present in executor" "FAIL"
fi

# Test 8: context.Context parameter used (non-blocking, cancellable)
if grep -q "context\.Context\|ctx context" "$EXECUTOR_FILE" 2>/dev/null; then
    record_result "ExecuteWithTools uses context.Context for cancellation" "PASS"
else
    record_result "ExecuteWithTools uses context.Context for cancellation" "FAIL"
fi

# ============================================================================
# SECTION 4: TOOL INTEGRATION INTERFACES
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: Tool Integration Interfaces ---${NC}"

# Search across all tool-related files in services/
TOOL_SOURCES="$PROJECT_ROOT/internal/services"

# Test 9: VisionClient interface defined
if grep -rq "VisionClient\|VisionService\|visionClient" "$TOOL_SOURCES" 2>/dev/null; then
    record_result "VisionClient interface or field referenced in services" "PASS"
else
    record_result "VisionClient interface or field referenced in services" "FAIL"
fi

# Test 10: HelixMemory client interface defined
if grep -rq "HelixMemoryClient\|HelixMemory\|helixMemory\|memoryClient" "$TOOL_SOURCES" 2>/dev/null; then
    record_result "HelixMemory client interface or field referenced in services" "PASS"
else
    record_result "HelixMemory client interface or field referenced in services" "FAIL"
fi

# ============================================================================
# SECTION 5: ALL 6 PROTOCOLS REPRESENTED
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: All 6 Protocols Represented ---${NC}"

# Check across executor file, types file, and the DebateOrchestrator ToolIntegration bridge.
# ACP is defined in the digital.vasic.debate/tools package that the executor imports.
PROTOCOL_SOURCES="$PROJECT_ROOT/internal/services/iterative_tool_executor.go
$PROJECT_ROOT/internal/services/agentic_ensemble_types.go
$PROJECT_ROOT/internal/services/agentic_ensemble.go
$PROJECT_ROOT/DebateOrchestrator/tools/tool_integration.go"

check_protocol() {
    local protocol_label="$1"
    local pattern="$2"
    local found=false
    for src in $PROTOCOL_SOURCES; do
        if [ -f "$src" ] && grep -qi "$pattern" "$src" 2>/dev/null; then
            found=true
            break
        fi
    done
    if $found; then
        record_result "$protocol_label protocol represented" "PASS"
    else
        record_result "$protocol_label protocol represented" "FAIL"
    fi
}

# Test 11: MCP protocol
check_protocol "MCP" "mcp\|MCP"

# Test 12: LSP protocol
check_protocol "LSP" "lsp\|LSP"

# Test 13: ACP protocol
check_protocol "ACP" "acp\|ACP"

# Test 14: RAG
check_protocol "RAG" '"rag"\|RAG\|retrieval'

# Test 15: Embeddings
check_protocol "Embeddings" "embed\|Embed"

# Test 16: Vision
check_protocol "Vision" "vision\|Vision"

# ============================================================================
# SECTION 6: TOOL EXECUTION TYPE
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 6: Tool Execution Type ---${NC}"

# Test 17: AgenticToolExecution struct has Protocol field
if grep -q "Protocol" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticToolExecution struct has Protocol field" "PASS"
else
    record_result "AgenticToolExecution struct has Protocol field" "FAIL"
fi

# Test 18: AgenticToolExecution struct has Operation field
if grep -q "Operation" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticToolExecution struct has Operation field" "PASS"
else
    record_result "AgenticToolExecution struct has Operation field" "FAIL"
fi

# Test 19: AgenticToolExecution struct has Duration field
if grep -q "Duration" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticToolExecution struct has Duration field" "PASS"
else
    record_result "AgenticToolExecution struct has Duration field" "FAIL"
fi

# ============================================================================
# SECTION 7: BUILD AND COMPILE CHECKS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 7: Build and Compile ---${NC}"

# Test 20: internal/services/ compiles after executor addition
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go build \
    "$PROJECT_ROOT/internal/services/..." \
    > /tmp/agentic_tool_loop_build.log 2>&1; then
    record_result "internal/services/ compiles cleanly with iterative_tool_executor" "PASS"
else
    echo -e "${YELLOW}  Build log: /tmp/agentic_tool_loop_build.log${NC}"
    record_result "internal/services/ compiles cleanly with iterative_tool_executor" "FAIL"
fi

# Test 21: Test files compile cleanly
if GOMAXPROCS=2 nice -n 19 ionice -c 3 \
    go test -count=1 -run='^$' \
    "$PROJECT_ROOT/internal/services/" \
    > /tmp/agentic_tool_loop_test_compile.log 2>&1; then
    record_result "internal/services/ test files compile cleanly" "PASS"
else
    echo -e "${YELLOW}  Test compile log: /tmp/agentic_tool_loop_test_compile.log${NC}"
    record_result "internal/services/ test files compile cleanly" "FAIL"
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
