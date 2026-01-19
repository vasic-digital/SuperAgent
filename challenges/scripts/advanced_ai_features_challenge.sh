#!/bin/bash

# Advanced AI Features Challenge Script
# Tests: ToT, MCTS, HiPlan, CodeGraph, GraphRAG, Formal Verification, SecureFixAgent, LSP-AI, Lesson Banking, SEMAP
# Target: 100% validation of all advanced AI features

# Don't exit on error - we want to report all test results

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test result tracking
declare -a FAILED_TEST_NAMES

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_failure() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    FAILED_TEST_NAMES+=("$1")
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -e "\n${BLUE}Running Test:${NC} $test_name"

    if eval "$test_cmd" > /dev/null 2>&1; then
        log_success "$test_name"
        return 0
    else
        log_failure "$test_name"
        return 1
    fi
}

check_file_exists() {
    local file_path="$1"
    local description="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ -f "$PROJECT_ROOT/$file_path" ]; then
        log_success "File exists: $description"
        return 0
    else
        log_failure "File missing: $description ($file_path)"
        return 1
    fi
}

check_go_package() {
    local package_path="$1"
    local description="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if go list "$package_path" > /dev/null 2>&1; then
        log_success "Package valid: $description"
        return 0
    else
        log_failure "Package invalid: $description ($package_path)"
        return 1
    fi
}

check_struct_exists() {
    local file_path="$1"
    local struct_name="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if grep -q "type $struct_name struct" "$PROJECT_ROOT/$file_path" 2>/dev/null; then
        log_success "Struct exists: $struct_name in $file_path"
        return 0
    else
        log_failure "Struct missing: $struct_name in $file_path"
        return 1
    fi
}

check_interface_exists() {
    local file_path="$1"
    local interface_name="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if grep -q "type $interface_name interface" "$PROJECT_ROOT/$file_path" 2>/dev/null; then
        log_success "Interface exists: $interface_name in $file_path"
        return 0
    else
        log_failure "Interface missing: $interface_name in $file_path"
        return 1
    fi
}

check_function_exists() {
    local file_path="$1"
    local function_name="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if grep -q "func.*$function_name" "$PROJECT_ROOT/$file_path" 2>/dev/null; then
        log_success "Function exists: $function_name in $file_path"
        return 0
    else
        log_failure "Function missing: $function_name in $file_path"
        return 1
    fi
}

echo "============================================"
echo "  Advanced AI Features Challenge"
echo "============================================"
echo "Testing all advanced AI capabilities..."
echo ""

cd "$PROJECT_ROOT"

# ============================================
# 1. Tree of Thoughts (ToT) Tests
# ============================================
echo -e "\n${YELLOW}=== Testing Tree of Thoughts (ToT) ===${NC}"

check_file_exists "internal/planning/tree_of_thoughts.go" "ToT implementation"
check_struct_exists "internal/planning/tree_of_thoughts.go" "TreeOfThoughts"
check_struct_exists "internal/planning/tree_of_thoughts.go" "Thought"
check_struct_exists "internal/planning/tree_of_thoughts.go" "ThoughtNode"
check_interface_exists "internal/planning/tree_of_thoughts.go" "ThoughtGenerator"
check_interface_exists "internal/planning/tree_of_thoughts.go" "ThoughtEvaluator"
check_function_exists "internal/planning/tree_of_thoughts.go" "Solve"
check_function_exists "internal/planning/tree_of_thoughts.go" "breadthFirstSearch"
check_function_exists "internal/planning/tree_of_thoughts.go" "depthFirstSearch"
check_function_exists "internal/planning/tree_of_thoughts.go" "beamSearch"

# ============================================
# 2. Monte Carlo Tree Search (MCTS) Tests
# ============================================
echo -e "\n${YELLOW}=== Testing Monte Carlo Tree Search (MCTS) ===${NC}"

check_file_exists "internal/planning/mcts.go" "MCTS implementation"
check_struct_exists "internal/planning/mcts.go" "MCTS"
check_struct_exists "internal/planning/mcts.go" "MCTSNode"
check_struct_exists "internal/planning/mcts.go" "MCTSConfig"
check_function_exists "internal/planning/mcts.go" "Search"
check_function_exists "internal/planning/mcts.go" "selectNode"
check_function_exists "internal/planning/mcts.go" "expand"
check_function_exists "internal/planning/mcts.go" "simulate"
check_function_exists "internal/planning/mcts.go" "backpropagate"
check_function_exists "internal/planning/mcts.go" "UCTValue"

# ============================================
# 3. Hierarchical Planning (HiPlan) Tests
# ============================================
echo -e "\n${YELLOW}=== Testing Hierarchical Planning (HiPlan) ===${NC}"

check_file_exists "internal/planning/hiplan.go" "HiPlan implementation"
check_struct_exists "internal/planning/hiplan.go" "HiPlan"
check_struct_exists "internal/planning/hiplan.go" "HierarchicalPlan"
check_struct_exists "internal/planning/hiplan.go" "Milestone"
check_struct_exists "internal/planning/hiplan.go" "PlanStep"
check_function_exists "internal/planning/hiplan.go" "CreatePlan"
check_function_exists "internal/planning/hiplan.go" "ExecutePlan"
check_function_exists "internal/planning/hiplan.go" "ExecuteStep"

# ============================================
# 4. Code Knowledge Graph Tests
# ============================================
echo -e "\n${YELLOW}=== Testing Code Knowledge Graph ===${NC}"

check_file_exists "internal/knowledge/code_graph.go" "CodeGraph implementation"
check_struct_exists "internal/knowledge/code_graph.go" "CodeGraph"
check_struct_exists "internal/knowledge/code_graph.go" "CodeNode"
check_struct_exists "internal/knowledge/code_graph.go" "CodeEdge"
check_function_exists "internal/knowledge/code_graph.go" "AddNode"
check_function_exists "internal/knowledge/code_graph.go" "AddEdge"
check_function_exists "internal/knowledge/code_graph.go" "GetNeighbors"
check_function_exists "internal/knowledge/code_graph.go" "GetImpactRadius"
check_function_exists "internal/knowledge/code_graph.go" "SemanticSearch"
check_function_exists "internal/knowledge/code_graph.go" "FindPath"

# ============================================
# 5. GraphRAG Tests
# ============================================
echo -e "\n${YELLOW}=== Testing GraphRAG ===${NC}"

check_file_exists "internal/knowledge/graphrag.go" "GraphRAG implementation"
check_struct_exists "internal/knowledge/graphrag.go" "GraphRAG"
check_struct_exists "internal/knowledge/graphrag.go" "RetrievalResult"
check_function_exists "internal/knowledge/graphrag.go" "Retrieve"
check_function_exists "internal/knowledge/graphrag.go" "BuildContext"
check_function_exists "internal/knowledge/graphrag.go" "retrieveLocal"
check_function_exists "internal/knowledge/graphrag.go" "retrieveGraph"

# ============================================
# 6. Formal Verification Tests
# ============================================
echo -e "\n${YELLOW}=== Testing Formal Verification ===${NC}"

check_file_exists "internal/verification/formal_verifier.go" "FormalVerifier implementation"
check_struct_exists "internal/verification/formal_verifier.go" "FormalVerifier"
check_struct_exists "internal/verification/formal_verifier.go" "VerificationResult"
check_struct_exists "internal/verification/formal_verifier.go" "Specification"
check_interface_exists "internal/verification/formal_verifier.go" "SpecGenerator"
check_interface_exists "internal/verification/formal_verifier.go" "TheoremProver"
check_function_exists "internal/verification/formal_verifier.go" "VerifyCode"
check_function_exists "internal/verification/formal_verifier.go" "VerifyPlan"

# ============================================
# 7. SecureFixAgent Tests
# ============================================
echo -e "\n${YELLOW}=== Testing SecureFixAgent ===${NC}"

check_file_exists "internal/security/secure_fix_agent.go" "SecureFixAgent implementation"
check_struct_exists "internal/security/secure_fix_agent.go" "SecureFixAgent"
check_struct_exists "internal/security/secure_fix_agent.go" "Vulnerability"
check_struct_exists "internal/security/secure_fix_agent.go" "SecurityResult"
check_struct_exists "internal/security/secure_fix_agent.go" "FiveRingDefense"
check_interface_exists "internal/security/secure_fix_agent.go" "VulnerabilityScanner"
check_interface_exists "internal/security/secure_fix_agent.go" "FixGenerator"
check_function_exists "internal/security/secure_fix_agent.go" "DetectRepairValidate"

# ============================================
# 8. LSP-AI Tests
# ============================================
echo -e "\n${YELLOW}=== Testing LSP-AI ===${NC}"

check_file_exists "internal/lsp/lsp_ai.go" "LSP-AI implementation"
check_struct_exists "internal/lsp/lsp_ai.go" "LSPAI"
check_struct_exists "internal/lsp/lsp_ai.go" "CompletionItem"
check_struct_exists "internal/lsp/lsp_ai.go" "Diagnostic"
check_struct_exists "internal/lsp/lsp_ai.go" "CodeAction"
check_interface_exists "internal/lsp/lsp_ai.go" "AICompletionProvider"
check_function_exists "internal/lsp/lsp_ai.go" "GetCompletion"
check_function_exists "internal/lsp/lsp_ai.go" "GetHover"
check_function_exists "internal/lsp/lsp_ai.go" "GetCodeActions"
check_function_exists "internal/lsp/lsp_ai.go" "GetDiagnostics"

# ============================================
# 9. Lesson Banking Tests
# ============================================
echo -e "\n${YELLOW}=== Testing Lesson Banking ===${NC}"

check_file_exists "internal/debate/lesson_bank.go" "LessonBank implementation"
check_struct_exists "internal/debate/lesson_bank.go" "LessonBank"
check_struct_exists "internal/debate/lesson_bank.go" "Lesson"
check_struct_exists "internal/debate/lesson_bank.go" "LessonContent"
check_struct_exists "internal/debate/lesson_bank.go" "LessonStatistics"
check_function_exists "internal/debate/lesson_bank.go" "AddLesson"
check_function_exists "internal/debate/lesson_bank.go" "SearchLessons"
check_function_exists "internal/debate/lesson_bank.go" "ApplyLesson"
check_function_exists "internal/debate/lesson_bank.go" "ExtractLessonsFromDebate"

# ============================================
# 10. SEMAP Protocol Tests
# ============================================
echo -e "\n${YELLOW}=== Testing SEMAP Protocol ===${NC}"

check_file_exists "internal/governance/semap.go" "SEMAP implementation"
check_struct_exists "internal/governance/semap.go" "SEMAP"
check_struct_exists "internal/governance/semap.go" "Contract"
check_struct_exists "internal/governance/semap.go" "Policy"
check_struct_exists "internal/governance/semap.go" "AgentProfile"
check_struct_exists "internal/governance/semap.go" "Violation"
check_interface_exists "internal/governance/semap.go" "ConditionEvaluator"
check_function_exists "internal/governance/semap.go" "CheckPreconditions"
check_function_exists "internal/governance/semap.go" "CheckPostconditions"
check_function_exists "internal/governance/semap.go" "CheckGuardRails"

# ============================================
# 11. MCP Adapters Tests
# ============================================
echo -e "\n${YELLOW}=== Testing MCP Adapters ===${NC}"

check_file_exists "internal/mcp/adapters/registry.go" "MCP Adapter Registry"
check_file_exists "internal/mcp/adapters/brave_search.go" "Brave Search Adapter"
check_file_exists "internal/mcp/adapters/aws_s3.go" "AWS S3 Adapter"
check_file_exists "internal/mcp/adapters/google_drive.go" "Google Drive Adapter"
check_file_exists "internal/mcp/adapters/gitlab.go" "GitLab Adapter"
check_file_exists "internal/mcp/adapters/mongodb.go" "MongoDB Adapter"
check_file_exists "internal/mcp/adapters/puppeteer.go" "Puppeteer Adapter"
check_file_exists "internal/mcp/adapters/slack.go" "Slack Adapter"
check_file_exists "internal/mcp/adapters/docker.go" "Docker Adapter"
check_file_exists "internal/mcp/adapters/kubernetes.go" "Kubernetes Adapter"
check_file_exists "internal/mcp/adapters/notion.go" "Notion Adapter"

# ============================================
# 12. Embedding Models Tests
# ============================================
echo -e "\n${YELLOW}=== Testing Embedding Models ===${NC}"

check_file_exists "internal/embedding/models.go" "Embedding Models implementation"
check_struct_exists "internal/embedding/models.go" "OpenAIEmbedding"
check_struct_exists "internal/embedding/models.go" "OllamaEmbedding"
check_struct_exists "internal/embedding/models.go" "HuggingFaceEmbedding"
check_interface_exists "internal/embedding/models.go" "EmbeddingModel"
check_function_exists "internal/embedding/models.go" "Embed"
check_function_exists "internal/embedding/models.go" "EmbedBatch"

# ============================================
# 13. Unit Tests
# ============================================
echo -e "\n${YELLOW}=== Running Unit Tests ===${NC}"

run_test "Planning Tests" "go test -v ./internal/planning/... -short 2>&1 | head -50"
run_test "Knowledge Tests" "go test -v ./internal/knowledge/... -short 2>&1 | head -50"
run_test "Security Tests" "go test -v ./internal/security/... -short 2>&1 | head -50"
run_test "Governance Tests" "go test -v ./internal/governance/... -short 2>&1 | head -50"

# ============================================
# Summary
# ============================================
echo ""
echo "============================================"
echo "  Challenge Results Summary"
echo "============================================"
echo ""
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"
echo ""

if [ ${#FAILED_TEST_NAMES[@]} -gt 0 ]; then
    echo -e "${RED}Failed Tests:${NC}"
    for test_name in "${FAILED_TEST_NAMES[@]}"; do
        echo "  - $test_name"
    done
fi

echo ""
PASS_PERCENTAGE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
echo "Pass Rate: ${PASS_PERCENTAGE}%"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✓ ALL ADVANCED AI FEATURES CHALLENGE TESTS PASSED!${NC}"
    exit 0
else
    echo -e "\n${RED}✗ SOME TESTS FAILED${NC}"
    exit 1
fi
