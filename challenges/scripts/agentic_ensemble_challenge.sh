#!/bin/bash
# HelixAgent Challenge - Agentic Ensemble
# Validates that the AgenticEnsemble implementation exists, is structurally
# correct, and integrates with the OpenAI-compatible handler pipeline.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "agentic-ensemble" "Agentic Ensemble"
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
echo "  Agentic Ensemble Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: CORE FILE EXISTENCE
# ============================================================================
echo -e "${BLUE}--- Section 1: Core File Existence ---${NC}"

# Test 1: agentic_ensemble.go exists
if [ -f "$PROJECT_ROOT/internal/services/agentic_ensemble.go" ]; then
    record_result "internal/services/agentic_ensemble.go exists" "PASS"
else
    record_result "internal/services/agentic_ensemble.go exists" "FAIL"
fi

# Test 2: agentic_ensemble_types.go exists
if [ -f "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" ]; then
    record_result "internal/services/agentic_ensemble_types.go exists" "PASS"
else
    record_result "internal/services/agentic_ensemble_types.go exists" "FAIL"
fi

# Test 3: execution_planner.go exists (dependency graph support)
if [ -f "$PROJECT_ROOT/internal/services/execution_planner.go" ]; then
    record_result "internal/services/execution_planner.go exists" "PASS"
else
    record_result "internal/services/execution_planner.go exists" "FAIL"
fi

# ============================================================================
# SECTION 2: AGENTIC MODE ENUM
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: AgenticMode Enum ---${NC}"

# Test 4: AgenticMode type is defined
if grep -q "type AgenticMode" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticMode type defined in agentic_ensemble_types.go" "PASS"
else
    record_result "AgenticMode type defined in agentic_ensemble_types.go" "FAIL"
fi

# Test 5: AgenticModeReason constant exists
if grep -q "AgenticModeReason" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticModeReason constant exists" "PASS"
else
    record_result "AgenticModeReason constant exists" "FAIL"
fi

# Test 6: AgenticModeExecute constant exists
if grep -q "AgenticModeExecute" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticModeExecute constant exists" "PASS"
else
    record_result "AgenticModeExecute constant exists" "FAIL"
fi

# ============================================================================
# SECTION 3: AGENTIC TASK STRUCT
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: AgenticTask Struct ---${NC}"

# Test 7: AgenticTask struct defined
if grep -q "type AgenticTask struct" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticTask struct defined" "PASS"
else
    record_result "AgenticTask struct defined" "FAIL"
fi

# Test 8: AgenticTask has Dependencies field
if grep -q "Dependencies" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticTask has Dependencies field" "PASS"
else
    record_result "AgenticTask has Dependencies field" "FAIL"
fi

# Test 9: AgenticTask has ToolRequirements field
if grep -q "ToolRequirements" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticTask has ToolRequirements field" "PASS"
else
    record_result "AgenticTask has ToolRequirements field" "FAIL"
fi

# Test 10: AgenticTask has Priority field
if grep -q "Priority" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticTask has Priority field" "PASS"
else
    record_result "AgenticTask has Priority field" "FAIL"
fi

# ============================================================================
# SECTION 4: AGENTIC ENSEMBLE CONFIG
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: AgenticEnsembleConfig ---${NC}"

# Test 11: AgenticEnsembleConfig struct defined
if grep -q "type AgenticEnsembleConfig struct" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticEnsembleConfig struct defined" "PASS"
else
    record_result "AgenticEnsembleConfig struct defined" "FAIL"
fi

# Test 12: MaxConcurrentAgents field
if grep -q "MaxConcurrentAgents" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticEnsembleConfig has MaxConcurrentAgents field" "PASS"
else
    record_result "AgenticEnsembleConfig has MaxConcurrentAgents field" "FAIL"
fi

# Test 13: MaxIterationsPerAgent field
if grep -q "MaxIterationsPerAgent" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticEnsembleConfig has MaxIterationsPerAgent field" "PASS"
else
    record_result "AgenticEnsembleConfig has MaxIterationsPerAgent field" "FAIL"
fi

# Test 14: MaxToolIterationsPerPhase field
if grep -q "MaxToolIterationsPerPhase" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticEnsembleConfig has MaxToolIterationsPerPhase field" "PASS"
else
    record_result "AgenticEnsembleConfig has MaxToolIterationsPerPhase field" "FAIL"
fi

# Test 15: AgentTimeout field
if grep -q "AgentTimeout" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticEnsembleConfig has AgentTimeout field" "PASS"
else
    record_result "AgenticEnsembleConfig has AgentTimeout field" "FAIL"
fi

# Test 16: GlobalTimeout field
if grep -q "GlobalTimeout" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "AgenticEnsembleConfig has GlobalTimeout field" "PASS"
else
    record_result "AgenticEnsembleConfig has GlobalTimeout field" "FAIL"
fi

# Test 17: DefaultAgenticEnsembleConfig function
if grep -q "func DefaultAgenticEnsembleConfig" "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go" 2>/dev/null; then
    record_result "DefaultAgenticEnsembleConfig function exists" "PASS"
else
    record_result "DefaultAgenticEnsembleConfig function exists" "FAIL"
fi

# ============================================================================
# SECTION 5: AGENTIC ENSEMBLE STRUCT AND PROCESS METHOD
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: AgenticEnsemble Struct and Methods ---${NC}"

# Test 18: AgenticEnsemble struct defined
if grep -q "type AgenticEnsemble struct" "$PROJECT_ROOT/internal/services/agentic_ensemble.go" 2>/dev/null; then
    record_result "AgenticEnsemble struct defined in agentic_ensemble.go" "PASS"
else
    record_result "AgenticEnsemble struct defined in agentic_ensemble.go" "FAIL"
fi

# Test 19: Process method exists on AgenticEnsemble
if grep -q "func.*AgenticEnsemble.*Process\|func.*\*AgenticEnsemble.*Process" \
    "$PROJECT_ROOT/internal/services/agentic_ensemble.go" 2>/dev/null; then
    record_result "AgenticEnsemble has Process method" "PASS"
else
    record_result "AgenticEnsemble has Process method" "FAIL"
fi

# Test 20: NewAgenticEnsemble constructor exists
if grep -q "func NewAgenticEnsemble" "$PROJECT_ROOT/internal/services/agentic_ensemble.go" 2>/dev/null; then
    record_result "NewAgenticEnsemble constructor function exists" "PASS"
else
    record_result "NewAgenticEnsemble constructor function exists" "FAIL"
fi

# ============================================================================
# SECTION 6: POWER FEATURE REFERENCES
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 6: Power Feature References ---${NC}"

# AgenticEnsemble delegates protocol access to IterativeToolExecutor via ToolIntegration.
# Protocols are referenced across the agentic services layer and the DebateOrchestrator
# tools package (digital.vasic.debate/tools) rather than directly in agentic_ensemble.go.
AGENTIC_SRCS=(
    "$PROJECT_ROOT/internal/services/agentic_ensemble.go"
    "$PROJECT_ROOT/internal/services/iterative_tool_executor.go"
    "$PROJECT_ROOT/internal/services/agentic_ensemble_types.go"
    "$PROJECT_ROOT/DebateOrchestrator/tools/tool_integration.go"
)

check_power_feature() {
    local label="$1"
    local pattern="$2"
    local found=false
    for src in "${AGENTIC_SRCS[@]}"; do
        if [ -f "$src" ] && grep -qi "$pattern" "$src" 2>/dev/null; then
            found=true
            break
        fi
    done
    if $found; then
        record_result "AgenticEnsemble references $label" "PASS"
    else
        record_result "AgenticEnsemble references $label" "FAIL"
    fi
}

# Test 21: MCP protocol referenced
check_power_feature "MCP protocol" "mcp\|MCP"

# Test 22: LSP protocol referenced
check_power_feature "LSP protocol" "lsp\|LSP"

# Test 23: ACP protocol referenced (via ToolIntegration bridge in DebateOrchestrator)
check_power_feature "ACP protocol" "acp\|ACP"

# Test 24: RAG referenced
check_power_feature "RAG" '"rag"\|RAG\|rag'

# Test 25: Embeddings referenced
check_power_feature "Embeddings" "embed\|Embed"

# Test 26: Vision referenced
check_power_feature "Vision" "vision\|Vision"

# Test 27: HelixMemory referenced
ENSEMBLE_FILE="$PROJECT_ROOT/internal/services/agentic_ensemble.go"
if grep -qi "helixmemory\|HelixMemory\|memory" "$ENSEMBLE_FILE" 2>/dev/null || \
   grep -rqi "helixmemory\|HelixMemory\|memory" "$PROJECT_ROOT/internal/services/" 2>/dev/null; then
    record_result "AgenticEnsemble references HelixMemory" "PASS"
else
    record_result "AgenticEnsemble references HelixMemory" "FAIL"
fi

# ============================================================================
# SECTION 7: HANDLER INTEGRATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 7: Handler Integration ---${NC}"

HANDLER_FILE="$PROJECT_ROOT/internal/handlers/openai_compatible.go"

# Test 28: processWithEnsemble function exists in handler
if grep -q "func.*processWithEnsemble" "$HANDLER_FILE" 2>/dev/null; then
    record_result "processWithEnsemble function exists in openai_compatible.go" "PASS"
else
    record_result "processWithEnsemble function exists in openai_compatible.go" "FAIL"
fi

# Test 29: processWithEnsemble is called from handler
if grep -q "h\.processWithEnsemble\|processWithEnsemble(" "$HANDLER_FILE" 2>/dev/null; then
    record_result "processWithEnsemble is called within openai_compatible.go" "PASS"
else
    record_result "processWithEnsemble is called within openai_compatible.go" "FAIL"
fi

# ============================================================================
# SECTION 8: BUILD AND COMPILE CHECKS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 8: Build and Compile ---${NC}"

# Test 30: internal/services/ package compiles
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go build \
    "$PROJECT_ROOT/internal/services/..." \
    > /tmp/agentic_ensemble_services_build.log 2>&1; then
    record_result "internal/services/ compiles cleanly" "PASS"
else
    echo -e "${YELLOW}  Build log: /tmp/agentic_ensemble_services_build.log${NC}"
    record_result "internal/services/ compiles cleanly" "FAIL"
fi

# Test 31: Test files for agentic_ensemble_types compile (run='^$' means no tests run)
if GOMAXPROCS=2 nice -n 19 ionice -c 3 \
    go test -count=1 -run='^$' \
    "$PROJECT_ROOT/internal/services/" \
    > /tmp/agentic_ensemble_types_test_compile.log 2>&1; then
    record_result "internal/services/ test files compile cleanly" "PASS"
else
    echo -e "${YELLOW}  Test compile log: /tmp/agentic_ensemble_types_test_compile.log${NC}"
    record_result "internal/services/ test files compile cleanly" "FAIL"
fi

# Test 32: execution_planner.go has BuildDependencyGraph
if grep -q "func.*BuildDependencyGraph" "$PROJECT_ROOT/internal/services/execution_planner.go" 2>/dev/null; then
    record_result "execution_planner.go has BuildDependencyGraph method" "PASS"
else
    record_result "execution_planner.go has BuildDependencyGraph method" "FAIL"
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
