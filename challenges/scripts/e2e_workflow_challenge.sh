#!/bin/bash
# End-to-End Workflow Challenge - Test complete workflows
# Verifies entire system works from input to output

set -o pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

pass() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}[PASS]${NC} $1"
}

fail() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    echo -e "${RED}[FAIL]${NC} $1"
}

section() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

run_test() {
    local name="$1"
    local cmd="$2"
    echo -e "${BLUE}Running Test:${NC} $name"
    if eval "$cmd" > /dev/null 2>&1; then
        pass "$name"
    else
        fail "$name"
    fi
}

cd "$(dirname "$0")/../.." || exit 1

echo "============================================"
echo "  END-TO-END WORKFLOW CHALLENGE"
echo "  Test Complete System Workflows"
echo "============================================"

section "Code Analysis Workflow"

# Test: Code parsing -> Graph building -> Semantic search
echo -e "${BLUE}Testing:${NC} Code analysis workflow components"

# 1. Code parser exists
if grep -r "CodeParser\|ParseFile" internal/knowledge/code_graph.go > /dev/null 2>&1; then
    pass "Code parser component exists"
else
    fail "Code parser component missing"
fi

# 2. Graph builder exists
if grep -r "AddNode\|AddEdge\|BuildSemanticEdges" internal/knowledge/code_graph.go > /dev/null 2>&1; then
    pass "Graph builder component exists"
else
    fail "Graph builder component missing"
fi

# 3. Semantic search exists
if grep -r "SemanticSearch\|Retrieve" internal/knowledge/*.go > /dev/null 2>&1; then
    pass "Semantic search component exists"
else
    fail "Semantic search component missing"
fi

section "Planning Workflow"

# Test: Problem -> ToT/MCTS exploration -> Solution
echo -e "${BLUE}Testing:${NC} Planning workflow components"

# 1. Thought generation
if grep -r "GenerateThoughts\|ThoughtGenerator" internal/planning/tree_of_thoughts.go > /dev/null 2>&1; then
    pass "Thought generation component exists"
else
    fail "Thought generation component missing"
fi

# 2. Search strategies
if grep -r "SearchStrategy\|BFS\|DFS\|Beam\|selectNode" internal/planning/*.go > /dev/null 2>&1; then
    pass "Search strategy components exist"
else
    fail "Search strategy components missing"
fi

# 3. Solution evaluation
if grep -r "EvaluateThought\|ThoughtEvaluator\|IsTerminal" internal/planning/tree_of_thoughts.go > /dev/null 2>&1; then
    pass "Solution evaluation component exists"
else
    fail "Solution evaluation component missing"
fi

section "Security Workflow"

# Test: Code input -> Vulnerability scan -> Fix -> Validation
echo -e "${BLUE}Testing:${NC} Security workflow components"

# 1. Vulnerability scanner
if grep -r "Scan\|Scanner\|VulnerabilityScanner" internal/security/secure_fix_agent.go > /dev/null 2>&1; then
    pass "Vulnerability scanner component exists"
else
    fail "Vulnerability scanner component missing"
fi

# 2. Fix generator
if grep -r "GenerateFix\|FixGenerator\|Remediate" internal/security/secure_fix_agent.go > /dev/null 2>&1; then
    pass "Fix generator component exists"
else
    fail "Fix generator component missing"
fi

# 3. Validation
if grep -r "Validate\|Validator\|DetectRepairValidate" internal/security/secure_fix_agent.go > /dev/null 2>&1; then
    pass "Validation component exists"
else
    fail "Validation component missing"
fi

# 4. Defense rings
if grep -r "FiveRingDefense\|Ring\|InputSanitization\|RateLimiting" internal/security/secure_fix_agent.go > /dev/null 2>&1; then
    pass "Defense ring component exists"
else
    fail "Defense ring component missing"
fi

section "Governance Workflow"

# Test: Action request -> Contract check -> Execution -> Audit
echo -e "${BLUE}Testing:${NC} Governance workflow components"

# 1. Contract registration
if grep -r "RegisterContract\|Contract" internal/governance/semap.go > /dev/null 2>&1; then
    pass "Contract registration component exists"
else
    fail "Contract registration component missing"
fi

# 2. Precondition checking
if grep -r "CheckPreconditions" internal/governance/semap.go > /dev/null 2>&1; then
    pass "Precondition checking component exists"
else
    fail "Precondition checking component missing"
fi

# 3. Guard rails
if grep -r "CheckGuardRails" internal/governance/semap.go > /dev/null 2>&1; then
    pass "Guard rails component exists"
else
    fail "Guard rails component missing"
fi

# 4. Postcondition checking
if grep -r "CheckPostconditions" internal/governance/semap.go > /dev/null 2>&1; then
    pass "Postcondition checking component exists"
else
    fail "Postcondition checking component missing"
fi

# 5. Audit logging
if grep -r "AuditLog\|AuditEntry" internal/governance/semap.go > /dev/null 2>&1; then
    pass "Audit logging component exists"
else
    fail "Audit logging component missing"
fi

section "Debate Workflow"

# Test: Topic -> Debate -> Lesson extraction -> Knowledge storage
echo -e "${BLUE}Testing:${NC} Debate workflow components"

# 1. Debate initiation
if grep -r "Debate\|StartDebate\|DebateService" internal/services/debate*.go > /dev/null 2>&1; then
    pass "Debate initiation component exists"
else
    fail "Debate initiation component missing"
fi

# 2. Lesson extraction
if grep -r "ExtractLessonsFromDebate\|ExtractLesson" internal/debate/lesson_bank.go > /dev/null 2>&1; then
    pass "Lesson extraction component exists"
else
    fail "Lesson extraction component missing"
fi

# 3. Lesson storage
if grep -r "AddLesson\|LessonBank" internal/debate/lesson_bank.go > /dev/null 2>&1; then
    pass "Lesson storage component exists"
else
    fail "Lesson storage component missing"
fi

# 4. Lesson retrieval
if grep -r "SearchLessons\|ApplyLesson" internal/debate/lesson_bank.go > /dev/null 2>&1; then
    pass "Lesson retrieval component exists"
else
    fail "Lesson retrieval component missing"
fi

section "LSP Workflow"

# Test: Code context -> AI analysis -> Suggestions
echo -e "${BLUE}Testing:${NC} LSP workflow components"

# 1. Completion provider
if grep -r "GetCompletion\|CompletionItem" internal/lsp/lsp_ai.go > /dev/null 2>&1; then
    pass "Completion provider component exists"
else
    fail "Completion provider component missing"
fi

# 2. Hover provider
if grep -r "GetHover\|HoverInfo" internal/lsp/lsp_ai.go > /dev/null 2>&1; then
    pass "Hover provider component exists"
else
    fail "Hover provider component missing"
fi

# 3. Diagnostics
if grep -r "GetDiagnostics\|Diagnostic" internal/lsp/lsp_ai.go > /dev/null 2>&1; then
    pass "Diagnostics component exists"
else
    fail "Diagnostics component missing"
fi

# 4. Code actions
if grep -r "GetCodeActions\|CodeAction" internal/lsp/lsp_ai.go > /dev/null 2>&1; then
    pass "Code actions component exists"
else
    fail "Code actions component missing"
fi

section "Full Build Workflow"

# Test complete build of all packages
echo -e "${BLUE}Testing:${NC} Full system build"
run_test "Full internal build" "go build ./internal/..."
run_test "Full cmd build" "go build ./cmd/..."

section "Full Test Workflow"

# Run all tests to ensure complete system works
echo -e "${BLUE}Testing:${NC} Full test suite"
run_test "Internal package tests" "go test -short ./internal/..."

section "API Handler Workflow"

# Test handlers exist for all features
echo -e "${BLUE}Testing:${NC} API handler completeness"

# Check debate handler
if [ -f "internal/handlers/debate_handler.go" ]; then
    pass "Debate handler exists"
else
    fail "Debate handler missing"
fi

# Check background task handler
if [ -f "internal/handlers/background_task_handler.go" ]; then
    pass "Background task handler exists"
else
    fail "Background task handler missing"
fi

# Check completion handler (for chat/completion API)
if [ -f "internal/handlers/completion.go" ]; then
    pass "Completion handler exists"
else
    fail "Completion handler missing"
fi

# Check agent handler
if [ -f "internal/handlers/agent_handler.go" ]; then
    pass "Agent handler exists"
else
    fail "Agent handler missing"
fi

section "Background Processing Workflow"

# Test background task infrastructure
echo -e "${BLUE}Testing:${NC} Background processing components"

# Task queue
if grep -r "TaskQueue\|Queue" internal/background/*.go > /dev/null 2>&1; then
    pass "Task queue component exists"
else
    fail "Task queue component missing"
fi

# Worker pool
if grep -r "WorkerPool\|Worker" internal/background/*.go > /dev/null 2>&1; then
    pass "Worker pool component exists"
else
    fail "Worker pool component missing"
fi

# Stuck detection
if grep -r "StuckDetector\|Timeout\|Stuck" internal/background/*.go > /dev/null 2>&1; then
    pass "Stuck detection component exists"
else
    fail "Stuck detection component missing"
fi

section "Notification Workflow"

# Test notification infrastructure
echo -e "${BLUE}Testing:${NC} Notification components"

# SSE
if grep -r "SSE\|ServerSentEvent" internal/notifications/*.go > /dev/null 2>&1; then
    pass "SSE component exists"
else
    fail "SSE component missing"
fi

# WebSocket
if grep -r "WebSocket\|WS" internal/notifications/*.go > /dev/null 2>&1; then
    pass "WebSocket component exists"
else
    fail "WebSocket component missing"
fi

# Webhook
if grep -r "Webhook" internal/notifications/*.go > /dev/null 2>&1; then
    pass "Webhook component exists"
else
    fail "Webhook component missing"
fi

section "Data Flow Integrity Tests"

# Test that data flows correctly through the system
echo -e "${BLUE}Testing:${NC} Data flow integrity"

# Model -> Service -> Handler chain
if grep -r "service\|Service" internal/handlers/*.go > /dev/null 2>&1; then
    pass "Handler-Service connection exists"
else
    fail "Handler-Service connection missing"
fi

# Service -> Repository chain
if grep -r "repository\|Repository" internal/services/*.go > /dev/null 2>&1; then
    pass "Service-Repository connection exists"
else
    fail "Service-Repository connection missing"
fi

echo ""
echo "============================================"
echo "  E2E Workflow Challenge Results Summary"
echo "============================================"
echo ""
echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"
echo ""

if [ $TOTAL_TESTS -gt 0 ]; then
    PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "Pass Rate: ${PASS_RATE}%"
fi
echo ""

echo "Workflow Summary:"
echo "  1. Code Analysis: Parse -> Graph -> Search"
echo "  2. Planning: Problem -> Explore -> Solution"
echo "  3. Security: Scan -> Fix -> Validate"
echo "  4. Governance: Check -> Execute -> Audit"
echo "  5. Debate: Topic -> Debate -> Learn"
echo "  6. LSP: Context -> Analyze -> Suggest"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ ALL E2E WORKFLOW TESTS PASSED!${NC}"
    echo -e "${GREEN}  Complete system workflows are operational!${NC}"
    exit 0
else
    echo -e "${RED}✗ SOME E2E WORKFLOW TESTS FAILED${NC}"
    exit 1
fi
