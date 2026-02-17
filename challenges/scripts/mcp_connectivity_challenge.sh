#!/bin/bash
# HelixAgent Challenge: MCP Connectivity & SSE Endpoints
# Tests: ~20 tests across 4 sections
# Validates: SSE endpoint registration, URL correctness, config completeness, router wiring

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

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

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}MCP Connectivity Challenge${NC}"
echo -e "${BLUE}========================================${NC}"

#===============================================================================
# Section 1: SSE Endpoint Registration (9 tests)
#===============================================================================
section "Section 1: SSE Endpoint Registration"

SSE_FILE="$PROJECT_ROOT/internal/handlers/protocol_sse.go"

# Test 1.1: MCP SSE route registered
if grep -q 'router.GET("/mcp"' "$SSE_FILE" && grep -q 'router.POST("/mcp"' "$SSE_FILE"; then
    pass "MCP SSE route registered (GET + POST)"
else
    fail "MCP SSE route missing"
fi

# Test 1.2: ACP SSE route registered
if grep -q 'router.GET("/acp"' "$SSE_FILE" && grep -q 'router.POST("/acp"' "$SSE_FILE"; then
    pass "ACP SSE route registered (GET + POST)"
else
    fail "ACP SSE route missing"
fi

# Test 1.3: LSP SSE route registered
if grep -q 'router.GET("/lsp"' "$SSE_FILE" && grep -q 'router.POST("/lsp"' "$SSE_FILE"; then
    pass "LSP SSE route registered (GET + POST)"
else
    fail "LSP SSE route missing"
fi

# Test 1.4: Embeddings SSE route registered
if grep -q 'router.GET("/embeddings"' "$SSE_FILE" && grep -q 'router.POST("/embeddings"' "$SSE_FILE"; then
    pass "Embeddings SSE route registered (GET + POST)"
else
    fail "Embeddings SSE route missing"
fi

# Test 1.5: Vision SSE route registered
if grep -q 'router.GET("/vision"' "$SSE_FILE" && grep -q 'router.POST("/vision"' "$SSE_FILE"; then
    pass "Vision SSE route registered (GET + POST)"
else
    fail "Vision SSE route missing"
fi

# Test 1.6: Cognee SSE route registered
if grep -q 'router.GET("/cognee"' "$SSE_FILE" && grep -q 'router.POST("/cognee"' "$SSE_FILE"; then
    pass "Cognee SSE route registered (GET + POST)"
else
    fail "Cognee SSE route missing"
fi

# Test 1.7: RAG SSE route registered
if grep -q 'router.GET("/rag"' "$SSE_FILE" && grep -q 'router.POST("/rag"' "$SSE_FILE"; then
    pass "RAG SSE route registered (GET + POST)"
else
    fail "RAG SSE route missing"
fi

# Test 1.8: Formatters SSE route registered
if grep -q 'router.GET("/formatters"' "$SSE_FILE" && grep -q 'router.POST("/formatters"' "$SSE_FILE"; then
    pass "Formatters SSE route registered (GET + POST)"
else
    fail "Formatters SSE route missing"
fi

# Test 1.9: Monitoring SSE route registered
if grep -q 'router.GET("/monitoring"' "$SSE_FILE" && grep -q 'router.POST("/monitoring"' "$SSE_FILE"; then
    pass "Monitoring SSE route registered (GET + POST)"
else
    fail "Monitoring SSE route missing"
fi

#===============================================================================
# Section 2: URL Correctness (4 tests)
#===============================================================================
section "Section 2: URL Correctness"

MAIN_FILE="$PROJECT_ROOT/cmd/helixagent/main.go"
GENERATOR_FILE="$PROJECT_ROOT/LLMsVerifier/llm-verifier/pkg/cliagents/generator.go"

# Test 2.1: helixagent-formatters uses /v1/formatters not /v1/format
if grep -A2 'helixagent-formatters' "$MAIN_FILE" | grep -q '/v1/formatters'; then
    pass "helixagent-formatters URL uses /v1/formatters (not /v1/format)"
else
    fail "helixagent-formatters URL incorrect"
fi

# Test 2.2: deepwiki uses /mcp not /sse
if grep 'deepwiki' "$MAIN_FILE" | grep -q '/mcp"'; then
    if ! grep 'deepwiki' "$MAIN_FILE" | grep -q '/sse"'; then
        pass "deepwiki URL uses /mcp (not /sse)"
    else
        fail "deepwiki URL still contains /sse"
    fi
else
    fail "deepwiki URL does not use /mcp"
fi

# Test 2.3: No uvx commands in main.go MCP definitions
if ! grep -q '"uvx"' "$MAIN_FILE"; then
    pass "No uvx commands in main.go (all converted to npx)"
else
    fail "Found uvx commands in main.go"
fi

# Test 2.4: Generator has correct URLs
if grep 'helixagent-formatters' "$GENERATOR_FILE" | grep -q '/v1/formatters' && \
   grep 'deepwiki' "$GENERATOR_FILE" | grep -q '/mcp"'; then
    pass "LLMsVerifier generator has correct URLs"
else
    fail "LLMsVerifier generator has stale URLs"
fi

#===============================================================================
# Section 3: Config Completeness (4 tests)
#===============================================================================
section "Section 3: Config Completeness"

# Test 3.1: All 3 new handler fields in ProtocolSSEHandler struct
if grep -q 'ragHandler.*\*RAGHandler' "$SSE_FILE" && \
   grep -q 'formattersHandler.*\*FormattersHandler' "$SSE_FILE" && \
   grep -q 'monitoringHandler.*\*MonitoringHandler' "$SSE_FILE"; then
    pass "All 3 new handler fields in ProtocolSSEHandler struct"
else
    fail "Missing handler fields in ProtocolSSEHandler struct"
fi

# Test 3.2: Setter methods exist for new handlers
if grep -q 'func.*SetRAGHandler' "$SSE_FILE" && \
   grep -q 'func.*SetFormattersHandler' "$SSE_FILE" && \
   grep -q 'func.*SetMonitoringHandler' "$SSE_FILE"; then
    pass "Setter methods exist for all 3 new handlers"
else
    fail "Missing setter methods for new handlers"
fi

# Test 3.3: Protocol switch cases include new protocols
CAPS_COUNT=$(grep -c 'case "rag"\|case "formatters"\|case "monitoring"' "$SSE_FILE" || true)
if [ "$CAPS_COUNT" -ge 9 ]; then
    pass "All 3 new protocols in all 3 switch statements ($CAPS_COUNT cases)"
else
    fail "Missing protocol switch cases (found $CAPS_COUNT, expected 9)"
fi

# Test 3.4: All new protocols have tools defined
if grep -q 'func.*getRAGTools' "$SSE_FILE" && \
   grep -q 'func.*getFormattersTools' "$SSE_FILE" && \
   grep -q 'func.*getMonitoringTools' "$SSE_FILE"; then
    pass "Tool getter functions exist for all 3 new protocols"
else
    fail "Missing tool getter functions"
fi

#===============================================================================
# Section 4: Router Wiring (3 tests)
#===============================================================================
section "Section 4: Router Wiring"

ROUTER_FILE="$PROJECT_ROOT/internal/router/router.go"

# Test 4.1: SetRAGHandler called in router
if grep -q 'SetRAGHandler' "$ROUTER_FILE"; then
    pass "SetRAGHandler called in router.go"
else
    fail "SetRAGHandler not called in router.go"
fi

# Test 4.2: SetFormattersHandler called in router
if grep -q 'SetFormattersHandler' "$ROUTER_FILE"; then
    pass "SetFormattersHandler called in router.go"
else
    fail "SetFormattersHandler not called in router.go"
fi

# Test 4.3: SetMonitoringHandler called in router
if grep -q 'SetMonitoringHandler' "$ROUTER_FILE"; then
    pass "SetMonitoringHandler called in router.go"
else
    fail "SetMonitoringHandler not called in router.go"
fi

#===============================================================================
# Summary
#===============================================================================
echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}Results${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Total:  ${TOTAL}"
echo -e "  Passed: ${GREEN}${PASSED}${NC}"
echo -e "  Failed: ${RED}${FAILED}${NC}"

if [ "$FAILED" -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}${FAILED} test(s) failed!${NC}"
    exit 1
fi
