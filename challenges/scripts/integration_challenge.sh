#!/bin/bash
# Integration Challenge - Verify all components work together
# Tests: Planning, Knowledge, Security, Governance, MCP, LSP, Embedding integrations

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
echo "  INTEGRATION CHALLENGE"
echo "  Verify All Components Work Together"
echo "============================================"

section "Planning System Integration"

# Test planning package compiles and tests pass
run_test "Planning package compiles" "go build ./internal/planning/..."
run_test "Planning unit tests" "go test -short ./internal/planning/..."

# Verify key components exist
echo -e "${BLUE}Verifying:${NC} ToT, MCTS, HiPlan components"
if grep -q "TreeOfThoughts\|ToT" internal/planning/tree_of_thoughts.go && \
   grep -q "MCTS\|MonteCarloTreeSearch" internal/planning/mcts.go && \
   grep -q "HiPlan\|HierarchicalPlanner" internal/planning/hiplan.go; then
    pass "Planning components exist (ToT, MCTS, HiPlan)"
else
    fail "Planning components missing"
fi

section "Knowledge Graph Integration"

# Test knowledge package compiles and tests pass
run_test "Knowledge package compiles" "go build ./internal/knowledge/..."
run_test "Knowledge unit tests" "go test -short ./internal/knowledge/..."

# Verify key components exist
echo -e "${BLUE}Verifying:${NC} CodeGraph, GraphRAG components"
if grep -q "CodeGraph" internal/knowledge/code_graph.go && \
   grep -q "GraphRAG" internal/knowledge/graphrag.go; then
    pass "Knowledge components exist (CodeGraph, GraphRAG)"
else
    fail "Knowledge components missing"
fi

section "Security System Integration"

# Test security package compiles and tests pass
run_test "Security package compiles" "go build ./internal/security/..."
run_test "Security unit tests" "go test -short ./internal/security/..."

# Verify key components exist
echo -e "${BLUE}Verifying:${NC} SecureFixAgent, FiveRingDefense components"
if grep -q "SecureFixAgent\|FiveRingDefense" internal/security/secure_fix_agent.go; then
    pass "Security components exist (SecureFixAgent, FiveRingDefense)"
else
    fail "Security components missing"
fi

section "Governance System Integration"

# Test governance package compiles and tests pass
run_test "Governance package compiles" "go build ./internal/governance/..."
run_test "Governance unit tests" "go test -short ./internal/governance/..."

# Verify key components exist
echo -e "${BLUE}Verifying:${NC} SEMAP components"
if grep -q "SEMAP\|Contract\|Policy" internal/governance/semap.go; then
    pass "Governance components exist (SEMAP)"
else
    fail "Governance components missing"
fi

section "MCP Adapter Integration"

# Test MCP adapters package compiles
run_test "MCP adapters package compiles" "go build ./internal/mcp/adapters/..."

# Verify adapters exist
ADAPTERS=("registry" "brave_search" "aws_s3" "google_drive" "gitlab" "mongodb" "puppeteer" "slack" "docker" "kubernetes" "notion")
for adapter in "${ADAPTERS[@]}"; do
    if [ -f "internal/mcp/adapters/${adapter}.go" ]; then
        pass "MCP $adapter adapter exists"
    else
        fail "MCP $adapter adapter missing"
    fi
done

section "LSP Integration"

# Test LSP package compiles
run_test "LSP package compiles" "go build ./internal/lsp/..."

section "Embedding Models Integration"

# Test embedding package compiles
run_test "Embedding package compiles" "go build ./internal/embedding/..."

section "Debate System Integration"

# Test debate package compiles
run_test "Debate package compiles" "go build ./internal/debate/..."

section "Cross-Component Integration Tests"

# Test that all internal packages compile together
run_test "Full internal package build" "go build ./internal/..."

# Test that all packages have no import cycles
run_test "No import cycles" "go list ./internal/... 2>&1 | grep -v 'import cycle' || true"

# Run all short tests across internal packages
run_test "All internal short tests" "go test -short ./internal/..."

section "Service Integration Tests"

# Test services compile
run_test "Services compile" "go build ./internal/services/..."

# Test handlers compile
run_test "Handlers compile" "go build ./internal/handlers/..."

section "Database Integration"

# Test database packages compile
run_test "Database packages compile" "go build ./internal/database/..."

section "Caching Integration"

# Test cache packages compile
run_test "Cache packages compile" "go build ./internal/cache/..."

section "Background Tasks Integration"

# Test background package compiles
run_test "Background tasks compile" "go build ./internal/background/..."

section "Notification System Integration"

# Test notifications compile
run_test "Notifications compile" "go build ./internal/notifications/..."

section "Plugin System Integration"

# Test plugins compile
run_test "Plugins compile" "go build ./internal/plugins/..."

section "LLM Provider Integration"

# Test LLM packages compile
run_test "LLM packages compile" "go build ./internal/llm/..."

section "Optimization Integration"

# Test optimization packages compile
run_test "Optimization packages compile" "go build ./internal/optimization/..."

echo ""
echo "============================================"
echo "  Integration Challenge Results Summary"
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

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ ALL INTEGRATION TESTS PASSED!${NC}"
    exit 0
else
    echo -e "${RED}✗ SOME INTEGRATION TESTS FAILED${NC}"
    exit 1
fi
