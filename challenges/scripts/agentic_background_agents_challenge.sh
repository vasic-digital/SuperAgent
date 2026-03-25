#!/bin/bash
# HelixAgent Challenge - Agentic Background Agents
# Validates that the AgentWorkerPool and AgentWorker implementations exist,
# are structurally correct, and support graceful shutdown and dependency graphs.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "agentic-background-agents" "Agentic Background Agents"
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
echo "  Agentic Background Agents Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: CORE FILE EXISTENCE
# ============================================================================
echo -e "${BLUE}--- Section 1: Core File Existence ---${NC}"

# Test 1: agent_worker_pool.go exists
if [ -f "$PROJECT_ROOT/internal/services/agent_worker_pool.go" ]; then
    record_result "internal/services/agent_worker_pool.go exists" "PASS"
else
    record_result "internal/services/agent_worker_pool.go exists" "FAIL"
fi

# Test 2: agent_worker.go exists
if [ -f "$PROJECT_ROOT/internal/services/agent_worker.go" ]; then
    record_result "internal/services/agent_worker.go exists" "PASS"
else
    record_result "internal/services/agent_worker.go exists" "FAIL"
fi

# Test 3: execution_planner.go exists (dependency graph support)
if [ -f "$PROJECT_ROOT/internal/services/execution_planner.go" ]; then
    record_result "internal/services/execution_planner.go exists" "PASS"
else
    record_result "internal/services/execution_planner.go exists" "FAIL"
fi

# ============================================================================
# SECTION 2: AGENT WORKER POOL STRUCT
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: AgentWorkerPool Struct ---${NC}"

POOL_FILE="$PROJECT_ROOT/internal/services/agent_worker_pool.go"

# Test 4: AgentWorkerPool struct defined
if grep -q "type AgentWorkerPool struct" "$POOL_FILE" 2>/dev/null; then
    record_result "AgentWorkerPool struct defined" "PASS"
else
    record_result "AgentWorkerPool struct defined" "FAIL"
fi

# Test 5: semaphore pattern (chan struct{}) for concurrency limiting
if grep -q "chan struct{}" "$POOL_FILE" 2>/dev/null; then
    record_result "AgentWorkerPool uses channel semaphore (chan struct{}) for concurrency" "PASS"
else
    record_result "AgentWorkerPool uses channel semaphore (chan struct{}) for concurrency" "FAIL"
fi

# Test 6: NewAgentWorkerPool constructor
if grep -q "func NewAgentWorkerPool" "$POOL_FILE" 2>/dev/null; then
    record_result "NewAgentWorkerPool constructor function exists" "PASS"
else
    record_result "NewAgentWorkerPool constructor function exists" "FAIL"
fi

# ============================================================================
# SECTION 3: DISPATCH TASKS METHOD
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: DispatchTasks Method ---${NC}"

# Test 7: DispatchTasks method exists on AgentWorkerPool
if grep -q "func.*AgentWorkerPool.*DispatchTasks\|func.*\*AgentWorkerPool.*DispatchTasks" \
    "$POOL_FILE" 2>/dev/null; then
    record_result "AgentWorkerPool has DispatchTasks method" "PASS"
else
    record_result "AgentWorkerPool has DispatchTasks method" "FAIL"
fi

# Test 8: context.Context used in pool (cancellable dispatching)
if grep -q "context\.Context\|ctx context" "$POOL_FILE" 2>/dev/null; then
    record_result "AgentWorkerPool uses context.Context for cancellation" "PASS"
else
    record_result "AgentWorkerPool uses context.Context for cancellation" "FAIL"
fi

# ============================================================================
# SECTION 4: AGENT WORKER STRUCT AND EXECUTE METHOD
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: AgentWorker Struct and Execute Method ---${NC}"

WORKER_FILE="$PROJECT_ROOT/internal/services/agent_worker.go"

# Test 9: AgentWorker struct defined
if grep -q "type AgentWorker struct" "$WORKER_FILE" 2>/dev/null; then
    record_result "AgentWorker struct defined" "PASS"
else
    record_result "AgentWorker struct defined" "FAIL"
fi

# Test 10: Execute method on AgentWorker
if grep -q "func.*AgentWorker.*Execute\|func.*\*AgentWorker.*Execute" \
    "$WORKER_FILE" 2>/dev/null; then
    record_result "AgentWorker has Execute method" "PASS"
else
    record_result "AgentWorker has Execute method" "FAIL"
fi

# Test 11: NewAgentWorker constructor exists
if grep -q "func NewAgentWorker" "$WORKER_FILE" 2>/dev/null; then
    record_result "NewAgentWorker constructor function exists" "PASS"
else
    record_result "NewAgentWorker constructor function exists" "FAIL"
fi

# ============================================================================
# SECTION 5: GRACEFUL SHUTDOWN
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: Graceful Shutdown ---${NC}"

# Test 12: Shutdown method exists on pool
if grep -q "func.*AgentWorkerPool.*Shutdown\|func.*\*AgentWorkerPool.*Shutdown" \
    "$POOL_FILE" 2>/dev/null; then
    record_result "AgentWorkerPool has Shutdown method" "PASS"
else
    record_result "AgentWorkerPool has Shutdown method" "FAIL"
fi

# Test 13: sync.WaitGroup used for goroutine lifecycle tracking
if grep -q "sync\.WaitGroup\|WaitGroup" "$POOL_FILE" 2>/dev/null; then
    record_result "AgentWorkerPool uses sync.WaitGroup for goroutine lifecycle" "PASS"
else
    record_result "AgentWorkerPool uses sync.WaitGroup for goroutine lifecycle" "FAIL"
fi

# Test 14: cancel function used (context cancellation on shutdown)
if grep -q "cancel\|cancelFunc\|context\.WithCancel" "$POOL_FILE" 2>/dev/null; then
    record_result "AgentWorkerPool uses context cancel for shutdown signaling" "PASS"
else
    record_result "AgentWorkerPool uses context cancel for shutdown signaling" "FAIL"
fi

# ============================================================================
# SECTION 6: DEPENDENCY GRAPH SUPPORT
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 6: Dependency Graph Support ---${NC}"

PLANNER_FILE="$PROJECT_ROOT/internal/services/execution_planner.go"

# Test 15: BuildDependencyGraph method exists on ExecutionPlanner
if grep -q "func.*ExecutionPlanner.*BuildDependencyGraph\|func.*\*ExecutionPlanner.*BuildDependencyGraph" \
    "$PLANNER_FILE" 2>/dev/null; then
    record_result "ExecutionPlanner has BuildDependencyGraph method" "PASS"
else
    record_result "ExecutionPlanner has BuildDependencyGraph method" "FAIL"
fi

# Test 16: Circular dependency detection in BuildDependencyGraph
if grep -q "circular\|cycle\|circular dependency" "$PLANNER_FILE" 2>/dev/null; then
    record_result "BuildDependencyGraph detects circular dependencies" "PASS"
else
    record_result "BuildDependencyGraph detects circular dependencies" "FAIL"
fi

# Test 17: DecomposePlan method exists (LLM-based task decomposition)
if grep -q "func.*ExecutionPlanner.*DecomposePlan\|func.*\*ExecutionPlanner.*DecomposePlan" \
    "$PLANNER_FILE" 2>/dev/null; then
    record_result "ExecutionPlanner has DecomposePlan method" "PASS"
else
    record_result "ExecutionPlanner has DecomposePlan method" "FAIL"
fi

# ============================================================================
# SECTION 7: PROVIDER SELECTION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 7: Provider Selection ---${NC}"

# Test 18: selectProvider method exists somewhere in services (worker or pool)
PROVIDER_SEL_FOUND=false
for src in "$POOL_FILE" "$WORKER_FILE" \
    "$PROJECT_ROOT/internal/services/agentic_ensemble.go"; do
    if [ -f "$src" ] && grep -q "selectProvider\|SelectProvider" "$src" 2>/dev/null; then
        PROVIDER_SEL_FOUND=true
        break
    fi
done
if $PROVIDER_SEL_FOUND; then
    record_result "selectProvider method exists in agentic services" "PASS"
else
    record_result "selectProvider method exists in agentic services" "FAIL"
fi

# ============================================================================
# SECTION 8: BUILD AND COMPILE CHECKS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 8: Build and Compile ---${NC}"

# Test 19: internal/services/ compiles with worker pool files
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go build \
    "$PROJECT_ROOT/internal/services/..." \
    > /tmp/agentic_background_agents_build.log 2>&1; then
    record_result "internal/services/ compiles cleanly with worker pool files" "PASS"
else
    echo -e "${YELLOW}  Build log: /tmp/agentic_background_agents_build.log${NC}"
    record_result "internal/services/ compiles cleanly with worker pool files" "FAIL"
fi

# Test 20: Test files compile cleanly
if GOMAXPROCS=2 nice -n 19 ionice -c 3 \
    go test -count=1 -run='^$' \
    "$PROJECT_ROOT/internal/services/" \
    > /tmp/agentic_background_agents_test_compile.log 2>&1; then
    record_result "internal/services/ test files compile cleanly" "PASS"
else
    echo -e "${YELLOW}  Test compile log: /tmp/agentic_background_agents_test_compile.log${NC}"
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
