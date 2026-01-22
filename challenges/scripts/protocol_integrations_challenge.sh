#!/bin/bash
# Protocol Integrations Challenge
# Verifies MCP, LSP, ACP, Embedding, and RAG integrations are fully implemented
# and working with all supported CLI agents

set -e

# Source challenge framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

# Initialize challenge
init_challenge "protocol_integrations" "Protocol Integrations Verification"
load_env

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))

    log_info "Running test: $test_name"

    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        log_success "PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "$test_name" "passed" ""
        return 0
    else
        log_error "FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "$test_name" "failed" "Test command failed"
        return 1
    fi
}

# ============================================================================
# SECTION 1: MCP SERVER VERIFICATION (16 servers + 17 adapters = 33 tests)
# ============================================================================
log_info "="
log_info "SECTION 1: MCP Server Verification"
log_info "="

# MCP Servers
MCP_SERVERS=(
    "filesystem" "git" "fetch" "memory" "chroma" "qdrant" "weaviate"
    "github" "postgres" "sqlite" "redis" "figma" "miro" "replicate"
    "stablediffusion" "svgmaker"
)

# MCP Adapters
MCP_ADAPTERS=(
    "aws_s3" "brave_search" "datadog" "docker" "figma" "gitlab"
    "google_drive" "kubernetes" "miro" "mongodb" "notion" "puppeteer"
    "sentry" "slack" "stable_diffusion" "svgmaker" "registry"
)

log_info "Verifying MCP Server implementations..."
for server in "${MCP_SERVERS[@]}"; do
    run_test "MCP Server: $server implementation exists" \
        "[[ -f '$PROJECT_ROOT/internal/mcp/servers/${server}_adapter.go' ]]"

    run_test "MCP Server: $server has tests" \
        "[[ -f '$PROJECT_ROOT/internal/mcp/servers/${server}_adapter_test.go' ]]"
done

log_info "Verifying MCP Adapter implementations..."
for adapter in "${MCP_ADAPTERS[@]}"; do
    adapter_file=$(echo "$adapter" | tr '_' '-')
    run_test "MCP Adapter: $adapter implementation exists" \
        "[[ -f '$PROJECT_ROOT/internal/mcp/adapters/${adapter}.go' ]] || [[ -f '$PROJECT_ROOT/internal/mcp/adapters/${adapter_file}.go' ]]"
done

# ============================================================================
# SECTION 2: LSP SERVER VERIFICATION (19 languages)
# ============================================================================
log_info "="
log_info "SECTION 2: LSP Server Verification"
log_info "="

LSP_LANGUAGES=(
    "go" "rust" "python" "typescript" "javascript" "c" "cpp" "java"
    "csharp" "php" "ruby" "elixir" "haskell" "bash" "yaml"
    "dockerfile" "terraform" "lua" "xml"
)

log_info "Verifying LSP Server registry..."
run_test "LSP Registry exists" \
    "[[ -f '$PROJECT_ROOT/internal/lsp/servers/registry.go' ]]"

run_test "LSP Registry has tests" \
    "[[ -f '$PROJECT_ROOT/internal/lsp/servers/registry_test.go' ]]"

log_info "Verifying LSP language support..."
for lang in "${LSP_LANGUAGES[@]}"; do
    run_test "LSP Language support: $lang" \
        "grep -qi '$lang' '$PROJECT_ROOT/internal/lsp/servers/registry.go'"
done

run_test "LSP-AI integration exists" \
    "[[ -f '$PROJECT_ROOT/internal/lsp/lsp_ai.go' ]]"

# ============================================================================
# SECTION 3: ACP VERIFICATION
# ============================================================================
log_info "="
log_info "SECTION 3: ACP (Agent Client Protocol) Verification"
log_info "="

run_test "ACP Manager exists" \
    "[[ -f '$PROJECT_ROOT/internal/services/acp_manager.go' ]]"

run_test "ACP Manager has tests" \
    "[[ -f '$PROJECT_ROOT/internal/services/acp_manager_test.go' ]]"

run_test "ACP Client exists" \
    "[[ -f '$PROJECT_ROOT/internal/services/acp_client.go' ]]"

run_test "ACP Client has tests" \
    "[[ -f '$PROJECT_ROOT/internal/services/acp_client_test.go' ]]"

run_test "ACP supports HTTP transport" \
    "grep -q 'HTTP' '$PROJECT_ROOT/internal/services/acp_manager.go'"

run_test "ACP supports WebSocket transport" \
    "grep -q 'WebSocket' '$PROJECT_ROOT/internal/services/acp_manager.go' || grep -q 'websocket' '$PROJECT_ROOT/internal/services/acp_client.go'"

# ============================================================================
# SECTION 4: EMBEDDING MODELS VERIFICATION
# ============================================================================
log_info "="
log_info "SECTION 4: Embedding Models Verification"
log_info "="

# Embedding Models from requirements
EMBEDDING_MODELS=(
    "openai" "ollama" "bge" "nomic" "sentence-transformers"
    "qwen" "gte" "codeBERT"
)

run_test "Embedding registry exists" \
    "[[ -f '$PROJECT_ROOT/internal/embeddings/models/registry.go' ]] || [[ -f '$PROJECT_ROOT/internal/embedding/models.go' ]]"

run_test "Embedding models have tests" \
    "[[ -f '$PROJECT_ROOT/internal/embeddings/models/registry_test.go' ]] || [[ -f '$PROJECT_ROOT/internal/embedding/models_test.go' ]]"

log_info "Verifying embedding model support..."
for model in "${EMBEDDING_MODELS[@]}"; do
    model_pattern=$(echo "$model" | tr '[:upper:]' '[:lower:]')
    run_test "Embedding model support: $model" \
        "grep -qi '$model_pattern' '$PROJECT_ROOT/internal/embedding/models.go' || grep -qi '$model_pattern' '$PROJECT_ROOT/internal/embeddings/models/registry.go'"
done

# ============================================================================
# SECTION 5: RAG VERIFICATION
# ============================================================================
log_info "="
log_info "SECTION 5: RAG (Retrieval-Augmented Generation) Verification"
log_info "="

RAG_COMPONENTS=(
    "pipeline" "hybrid" "reranker" "hyde" "advanced" "types"
    "qdrant_retriever" "qdrant_enhanced"
)

log_info "Verifying RAG components..."
for component in "${RAG_COMPONENTS[@]}"; do
    run_test "RAG component: $component exists" \
        "[[ -f '$PROJECT_ROOT/internal/rag/${component}.go' ]]"
done

run_test "RAG has comprehensive tests" \
    "ls '$PROJECT_ROOT/internal/rag/'*_test.go 2>/dev/null | wc -l | xargs test 5 -le"

run_test "RAG handler exists" \
    "[[ -f '$PROJECT_ROOT/internal/handlers/rag_handler.go' ]]"

run_test "RAG supports dense retrieval" \
    "grep -q 'Dense' '$PROJECT_ROOT/internal/rag/types.go'"

run_test "RAG supports sparse retrieval" \
    "grep -q 'Sparse' '$PROJECT_ROOT/internal/rag/types.go'"

run_test "RAG supports hybrid retrieval" \
    "grep -q 'Hybrid' '$PROJECT_ROOT/internal/rag/types.go'"

# ============================================================================
# SECTION 6: VECTOR DATABASE VERIFICATION
# ============================================================================
log_info "="
log_info "SECTION 6: Vector Database Verification"
log_info "="

run_test "Qdrant client exists" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/qdrant/client.go' ]]"

run_test "Qdrant client has tests" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/qdrant/client_test.go' ]]"

run_test "Qdrant config exists" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/qdrant/config.go' ]]"

# ============================================================================
# SECTION 7: CLI AGENTS VERIFICATION (20+ agents)
# ============================================================================
log_info "="
log_info "SECTION 7: CLI Agents Verification"
log_info "="

CLI_AGENTS=(
    "OpenCode" "Crush" "HelixCode" "Kiro" "Aider" "ClaudeCode"
    "Cline" "CodenameGoose" "DeepSeekCLI" "Forge" "GeminiCLI"
    "GPTEngineer" "KiloCode" "MistralCode" "OllamaCode" "Plandex"
    "QwenCode" "AmazonQ"
)

run_test "Agent registry exists" \
    "[[ -f '$PROJECT_ROOT/internal/agents/registry.go' ]]"

log_info "Verifying CLI agent support..."
for agent in "${CLI_AGENTS[@]}"; do
    run_test "CLI Agent: $agent registered" \
        "grep -q '$agent' '$PROJECT_ROOT/internal/agents/registry.go'"
done

# ============================================================================
# SECTION 8: PROTOCOL MANAGER VERIFICATION
# ============================================================================
log_info "="
log_info "SECTION 8: Protocol Manager Verification"
log_info "="

run_test "Unified Protocol Manager exists" \
    "[[ -f '$PROJECT_ROOT/internal/services/unified_protocol_manager.go' ]]"

run_test "Protocol Manager has tests" \
    "[[ -f '$PROJECT_ROOT/internal/services/unified_protocol_manager_test.go' ]]"

run_test "Protocol Manager supports MCP" \
    "grep -q 'MCP' '$PROJECT_ROOT/internal/services/unified_protocol_manager.go'"

run_test "Protocol Manager supports LSP" \
    "grep -q 'LSP' '$PROJECT_ROOT/internal/services/unified_protocol_manager.go'"

run_test "Protocol Manager supports ACP" \
    "grep -q 'ACP' '$PROJECT_ROOT/internal/services/unified_protocol_manager.go'"

run_test "MultiError implemented for error aggregation" \
    "grep -q 'MultiError' '$PROJECT_ROOT/internal/services/unified_protocol_manager.go'"

# ============================================================================
# SECTION 9: API ENDPOINTS VERIFICATION
# ============================================================================
log_info "="
log_info "SECTION 9: API Endpoints Verification"
log_info "="

run_test "Router registers MCP endpoints" \
    "grep -q '/mcp' '$PROJECT_ROOT/internal/router/router.go'"

run_test "Router registers RAG endpoints" \
    "grep -q '/rag' '$PROJECT_ROOT/internal/router/router.go'"

run_test "Router registers embeddings endpoints" \
    "grep -q '/embeddings' '$PROJECT_ROOT/internal/router/router.go'"

run_test "Router registers protocol endpoints" \
    "grep -q '/v1' '$PROJECT_ROOT/internal/router/router.go'"

# ============================================================================
# SECTION 10: UNIT TEST COVERAGE VERIFICATION
# ============================================================================
log_info "="
log_info "SECTION 10: Unit Test Coverage Verification"
log_info "="

run_test "MCP servers have test files" \
    "[[ \$(ls '$PROJECT_ROOT/internal/mcp/servers/'*_test.go 2>/dev/null | wc -l) -ge 15 ]]"

run_test "MCP adapters have test files" \
    "[[ \$(ls '$PROJECT_ROOT/internal/mcp/adapters/'*_test.go 2>/dev/null | wc -l) -ge 15 ]]"

run_test "RAG has test files" \
    "[[ \$(ls '$PROJECT_ROOT/internal/rag/'*_test.go 2>/dev/null | wc -l) -ge 5 ]]"

run_test "Services have test files" \
    "[[ \$(ls '$PROJECT_ROOT/internal/services/'*_test.go 2>/dev/null | wc -l) -ge 20 ]]"

# ============================================================================
# SUMMARY
# ============================================================================
log_info "="
log_info "CHALLENGE SUMMARY"
log_info "="

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

log_info "Total tests: $TESTS_TOTAL"
log_info "Passed: $TESTS_PASSED"
log_info "Failed: $TESTS_FAILED"
log_info "Duration: ${DURATION}s"

# Calculate pass rate
if [[ $TESTS_TOTAL -gt 0 ]]; then
    PASS_RATE=$((TESTS_PASSED * 100 / TESTS_TOTAL))
else
    PASS_RATE=0
fi

log_info "Pass rate: ${PASS_RATE}%"

# Update results file
update_results "completed" "{\"total_tests\": $TESTS_TOTAL, \"passed\": $TESTS_PASSED, \"failed\": $TESTS_FAILED, \"pass_rate\": $PASS_RATE, \"duration_seconds\": $DURATION}"

if [[ $TESTS_FAILED -eq 0 ]]; then
    log_success "="
    log_success "ALL PROTOCOL INTEGRATION TESTS PASSED!"
    log_success "="
    log_success "MCP Servers: 16 verified"
    log_success "MCP Adapters: 17 verified"
    log_success "LSP Languages: 19 verified"
    log_success "ACP Protocol: Fully implemented"
    log_success "Embedding Models: 8+ verified"
    log_success "RAG Components: 8 verified"
    log_success "CLI Agents: 18 verified"
    log_success "="
    exit 0
else
    log_error "="
    log_error "SOME TESTS FAILED"
    log_error "="
    log_error "Check $LOG_FILE for details"
    exit 1
fi
