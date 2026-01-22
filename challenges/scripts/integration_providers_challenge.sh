#!/bin/bash
# Integration Providers Challenge
# VALIDATES: New embedding providers, vector stores, and MCP adapters
# Tests the implementation of all new integration providers (45 tests)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Integration Providers Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: Embedding providers, Vector stores, MCP adapters"
log_info ""

# ============================================================================
# Section 1: Embedding Providers - Cohere
# ============================================================================

log_info "=============================================="
log_info "Section 1: Embedding Providers - Cohere"
log_info "=============================================="

# Test 1: CohereEmbedding struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: CohereEmbedding struct exists"
if grep -q "type CohereEmbedding struct" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "CohereEmbedding struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "CohereEmbedding struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: NewCohereEmbedding function exists
TOTAL=$((TOTAL + 1))
log_info "Test 2: NewCohereEmbedding function exists"
if grep -q "func NewCohereEmbedding" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "NewCohereEmbedding function exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewCohereEmbedding function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 3: Cohere API endpoint configured
TOTAL=$((TOTAL + 1))
log_info "Test 3: Cohere API endpoint configured"
if grep -q "api.cohere.com" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "Cohere API endpoint configured"
    PASSED=$((PASSED + 1))
else
    log_error "Cohere API endpoint NOT configured!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Embedding Providers - Voyage
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Embedding Providers - Voyage"
log_info "=============================================="

# Test 4: VoyageEmbedding struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 4: VoyageEmbedding struct exists"
if grep -q "type VoyageEmbedding struct" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "VoyageEmbedding struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "VoyageEmbedding struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 5: NewVoyageEmbedding function exists
TOTAL=$((TOTAL + 1))
log_info "Test 5: NewVoyageEmbedding function exists"
if grep -q "func NewVoyageEmbedding" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "NewVoyageEmbedding function exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewVoyageEmbedding function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 6: Voyage API endpoint configured
TOTAL=$((TOTAL + 1))
log_info "Test 6: Voyage API endpoint configured"
if grep -q "voyageai.com" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "Voyage API endpoint configured"
    PASSED=$((PASSED + 1))
else
    log_error "Voyage API endpoint NOT configured!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: Embedding Providers - Jina
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: Embedding Providers - Jina"
log_info "=============================================="

# Test 7: JinaEmbedding struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 7: JinaEmbedding struct exists"
if grep -q "type JinaEmbedding struct" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "JinaEmbedding struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "JinaEmbedding struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 8: NewJinaEmbedding function exists
TOTAL=$((TOTAL + 1))
log_info "Test 8: NewJinaEmbedding function exists"
if grep -q "func NewJinaEmbedding" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "NewJinaEmbedding function exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewJinaEmbedding function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: Jina API endpoint configured
TOTAL=$((TOTAL + 1))
log_info "Test 9: Jina API endpoint configured"
if grep -q "jina.ai" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "Jina API endpoint configured"
    PASSED=$((PASSED + 1))
else
    log_error "Jina API endpoint NOT configured!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Embedding Providers - Google
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Embedding Providers - Google"
log_info "=============================================="

# Test 10: GoogleEmbedding struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 10: GoogleEmbedding struct exists"
if grep -q "type GoogleEmbedding struct" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "GoogleEmbedding struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "GoogleEmbedding struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 11: NewGoogleEmbedding function exists
TOTAL=$((TOTAL + 1))
log_info "Test 11: NewGoogleEmbedding function exists"
if grep -q "func NewGoogleEmbedding" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "NewGoogleEmbedding function exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewGoogleEmbedding function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 12: Google Vertex AI support
TOTAL=$((TOTAL + 1))
log_info "Test 12: Google Vertex AI support"
if grep -q "aiplatform.googleapis.com\|vertex" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "Google Vertex AI support configured"
    PASSED=$((PASSED + 1))
else
    log_error "Google Vertex AI support NOT configured!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: Embedding Providers - AWS Bedrock
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 5: Embedding Providers - AWS Bedrock"
log_info "=============================================="

# Test 13: BedrockEmbedding struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 13: BedrockEmbedding struct exists"
if grep -q "type BedrockEmbedding struct" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "BedrockEmbedding struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "BedrockEmbedding struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 14: NewBedrockEmbedding function exists
TOTAL=$((TOTAL + 1))
log_info "Test 14: NewBedrockEmbedding function exists"
if grep -q "func NewBedrockEmbedding" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "NewBedrockEmbedding function exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewBedrockEmbedding function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 15: AWS Bedrock Titan model support
TOTAL=$((TOTAL + 1))
log_info "Test 15: AWS Bedrock Titan model support"
if grep -q "titan-embed\|amazon.titan" "$PROJECT_ROOT/internal/embedding/providers.go" 2>/dev/null; then
    log_success "AWS Bedrock Titan model support"
    PASSED=$((PASSED + 1))
else
    log_error "AWS Bedrock Titan model support NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 6: Vector Store - Pinecone
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 6: Vector Store - Pinecone"
log_info "=============================================="

# Test 16: Pinecone Client struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 16: Pinecone Client struct exists"
if grep -q "type Client struct" "$PROJECT_ROOT/internal/vectordb/pinecone/client.go" 2>/dev/null; then
    log_success "Pinecone Client struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "Pinecone Client struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 17: Pinecone Connect method exists
TOTAL=$((TOTAL + 1))
log_info "Test 17: Pinecone Connect method exists"
if grep -q "func (c \*Client) Connect" "$PROJECT_ROOT/internal/vectordb/pinecone/client.go" 2>/dev/null; then
    log_success "Pinecone Connect method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Pinecone Connect method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 18: Pinecone Upsert method exists
TOTAL=$((TOTAL + 1))
log_info "Test 18: Pinecone Upsert method exists"
if grep -q "func (c \*Client) Upsert" "$PROJECT_ROOT/internal/vectordb/pinecone/client.go" 2>/dev/null; then
    log_success "Pinecone Upsert method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Pinecone Upsert method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 19: Pinecone Query method exists
TOTAL=$((TOTAL + 1))
log_info "Test 19: Pinecone Query method exists"
if grep -q "func (c \*Client) Query" "$PROJECT_ROOT/internal/vectordb/pinecone/client.go" 2>/dev/null; then
    log_success "Pinecone Query method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Pinecone Query method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 20: Pinecone tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 20: Pinecone tests exist"
if [ -f "$PROJECT_ROOT/internal/vectordb/pinecone/client_test.go" ]; then
    log_success "Pinecone tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "Pinecone tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 7: Vector Store - Milvus
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 7: Vector Store - Milvus"
log_info "=============================================="

# Test 21: Milvus Client struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 21: Milvus Client struct exists"
if grep -q "type Client struct" "$PROJECT_ROOT/internal/vectordb/milvus/client.go" 2>/dev/null; then
    log_success "Milvus Client struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "Milvus Client struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 22: Milvus Connect method exists
TOTAL=$((TOTAL + 1))
log_info "Test 22: Milvus Connect method exists"
if grep -q "func (c \*Client) Connect" "$PROJECT_ROOT/internal/vectordb/milvus/client.go" 2>/dev/null; then
    log_success "Milvus Connect method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Milvus Connect method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 23: Milvus CreateCollection method exists
TOTAL=$((TOTAL + 1))
log_info "Test 23: Milvus CreateCollection method exists"
if grep -q "func (c \*Client) CreateCollection" "$PROJECT_ROOT/internal/vectordb/milvus/client.go" 2>/dev/null; then
    log_success "Milvus CreateCollection method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Milvus CreateCollection method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 24: Milvus Search method exists
TOTAL=$((TOTAL + 1))
log_info "Test 24: Milvus Search method exists"
if grep -q "func (c \*Client) Search" "$PROJECT_ROOT/internal/vectordb/milvus/client.go" 2>/dev/null; then
    log_success "Milvus Search method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Milvus Search method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 25: Milvus tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 25: Milvus tests exist"
if [ -f "$PROJECT_ROOT/internal/vectordb/milvus/client_test.go" ]; then
    log_success "Milvus tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "Milvus tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 8: Vector Store - pgvector
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 8: Vector Store - pgvector"
log_info "=============================================="

# Test 26: pgvector Client struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 26: pgvector Client struct exists"
if grep -q "type Client struct" "$PROJECT_ROOT/internal/vectordb/pgvector/client.go" 2>/dev/null; then
    log_success "pgvector Client struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "pgvector Client struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 27: pgvector Connect method exists
TOTAL=$((TOTAL + 1))
log_info "Test 27: pgvector Connect method exists"
if grep -q "func (c \*Client) Connect" "$PROJECT_ROOT/internal/vectordb/pgvector/client.go" 2>/dev/null; then
    log_success "pgvector Connect method exists"
    PASSED=$((PASSED + 1))
else
    log_error "pgvector Connect method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 28: pgvector CreateTable method exists
TOTAL=$((TOTAL + 1))
log_info "Test 28: pgvector CreateTable method exists"
if grep -q "func (c \*Client) CreateTable" "$PROJECT_ROOT/internal/vectordb/pgvector/client.go" 2>/dev/null; then
    log_success "pgvector CreateTable method exists"
    PASSED=$((PASSED + 1))
else
    log_error "pgvector CreateTable method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 29: pgvector Search method exists
TOTAL=$((TOTAL + 1))
log_info "Test 29: pgvector Search method exists"
if grep -q "func (c \*Client) Search" "$PROJECT_ROOT/internal/vectordb/pgvector/client.go" 2>/dev/null; then
    log_success "pgvector Search method exists"
    PASSED=$((PASSED + 1))
else
    log_error "pgvector Search method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 30: pgvector tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 30: pgvector tests exist"
if [ -f "$PROJECT_ROOT/internal/vectordb/pgvector/client_test.go" ]; then
    log_success "pgvector tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "pgvector tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 9: MCP Adapter - Linear
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 9: MCP Adapter - Linear"
log_info "=============================================="

# Test 31: LinearAdapter struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 31: LinearAdapter struct exists"
if grep -q "type LinearAdapter struct" "$PROJECT_ROOT/internal/mcp/adapters/linear.go" 2>/dev/null; then
    log_success "LinearAdapter struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "LinearAdapter struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 32: LinearClient interface exists
TOTAL=$((TOTAL + 1))
log_info "Test 32: LinearClient interface exists"
if grep -q "type LinearClient interface" "$PROJECT_ROOT/internal/mcp/adapters/linear.go" 2>/dev/null; then
    log_success "LinearClient interface exists"
    PASSED=$((PASSED + 1))
else
    log_error "LinearClient interface NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 33: Linear get_issue tool exists
TOTAL=$((TOTAL + 1))
log_info "Test 33: Linear get_issue tool exists"
if grep -q "linear_get_issue" "$PROJECT_ROOT/internal/mcp/adapters/linear.go" 2>/dev/null; then
    log_success "Linear get_issue tool exists"
    PASSED=$((PASSED + 1))
else
    log_error "Linear get_issue tool NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 34: Linear tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 34: Linear tests exist"
if [ -f "$PROJECT_ROOT/internal/mcp/adapters/linear_test.go" ]; then
    log_success "Linear tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "Linear tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 10: MCP Adapter - Asana
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 10: MCP Adapter - Asana"
log_info "=============================================="

# Test 35: AsanaAdapter struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 35: AsanaAdapter struct exists"
if grep -q "type AsanaAdapter struct" "$PROJECT_ROOT/internal/mcp/adapters/asana.go" 2>/dev/null; then
    log_success "AsanaAdapter struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "AsanaAdapter struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 36: AsanaClient interface exists
TOTAL=$((TOTAL + 1))
log_info "Test 36: AsanaClient interface exists"
if grep -q "type AsanaClient interface" "$PROJECT_ROOT/internal/mcp/adapters/asana.go" 2>/dev/null; then
    log_success "AsanaClient interface exists"
    PASSED=$((PASSED + 1))
else
    log_error "AsanaClient interface NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 37: Asana get_task tool exists
TOTAL=$((TOTAL + 1))
log_info "Test 37: Asana get_task tool exists"
if grep -q "asana_get_task" "$PROJECT_ROOT/internal/mcp/adapters/asana.go" 2>/dev/null; then
    log_success "Asana get_task tool exists"
    PASSED=$((PASSED + 1))
else
    log_error "Asana get_task tool NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 38: Asana tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 38: Asana tests exist"
if [ -f "$PROJECT_ROOT/internal/mcp/adapters/asana_test.go" ]; then
    log_success "Asana tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "Asana tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 39: Asana registered in adapter registry
TOTAL=$((TOTAL + 1))
log_info "Test 39: Asana registered in adapter registry"
if grep -q '"asana"' "$PROJECT_ROOT/internal/mcp/adapters/registry.go" 2>/dev/null; then
    log_success "Asana registered in adapter registry"
    PASSED=$((PASSED + 1))
else
    log_error "Asana NOT registered in adapter registry!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 11: MCP Adapter - Jira
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 11: MCP Adapter - Jira"
log_info "=============================================="

# Test 40: JiraAdapter struct exists
TOTAL=$((TOTAL + 1))
log_info "Test 40: JiraAdapter struct exists"
if grep -q "type JiraAdapter struct" "$PROJECT_ROOT/internal/mcp/adapters/jira.go" 2>/dev/null; then
    log_success "JiraAdapter struct exists"
    PASSED=$((PASSED + 1))
else
    log_error "JiraAdapter struct NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 41: JiraClient interface exists
TOTAL=$((TOTAL + 1))
log_info "Test 41: JiraClient interface exists"
if grep -q "type JiraClient interface" "$PROJECT_ROOT/internal/mcp/adapters/jira.go" 2>/dev/null; then
    log_success "JiraClient interface exists"
    PASSED=$((PASSED + 1))
else
    log_error "JiraClient interface NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 42: Jira get_issue tool exists
TOTAL=$((TOTAL + 1))
log_info "Test 42: Jira get_issue tool exists"
if grep -q "jira_get_issue" "$PROJECT_ROOT/internal/mcp/adapters/jira.go" 2>/dev/null; then
    log_success "Jira get_issue tool exists"
    PASSED=$((PASSED + 1))
else
    log_error "Jira get_issue tool NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 43: Jira tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 43: Jira tests exist"
if [ -f "$PROJECT_ROOT/internal/mcp/adapters/jira_test.go" ]; then
    log_success "Jira tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "Jira tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 44: Jira registered in adapter registry
TOTAL=$((TOTAL + 1))
log_info "Test 44: Jira registered in adapter registry"
if grep -q '"jira"' "$PROJECT_ROOT/internal/mcp/adapters/registry.go" 2>/dev/null; then
    log_success "Jira registered in adapter registry"
    PASSED=$((PASSED + 1))
else
    log_error "Jira NOT registered in adapter registry!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 12: Go Tests Validation
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 12: Go Tests Validation"
log_info "=============================================="

# Test 45: All embedding provider tests pass
TOTAL=$((TOTAL + 1))
log_info "Test 45: All embedding provider tests pass"
cd "$PROJECT_ROOT"
if go test -count=1 -timeout 60s ./internal/embedding/... > /dev/null 2>&1; then
    log_success "All embedding provider tests pass"
    PASSED=$((PASSED + 1))
else
    log_error "Embedding provider tests FAILED!"
    FAILED=$((FAILED + 1))
fi

# Test 46: All vector store tests pass
TOTAL=$((TOTAL + 1))
log_info "Test 46: All vector store tests pass"
if go test -count=1 -timeout 60s ./internal/vectordb/... > /dev/null 2>&1; then
    log_success "All vector store tests pass"
    PASSED=$((PASSED + 1))
else
    log_error "Vector store tests FAILED!"
    FAILED=$((FAILED + 1))
fi

# Test 47: All MCP adapter tests pass
TOTAL=$((TOTAL + 1))
log_info "Test 47: All MCP adapter tests pass"
if go test -count=1 -timeout 60s ./internal/mcp/adapters/... > /dev/null 2>&1; then
    log_success "All MCP adapter tests pass"
    PASSED=$((PASSED + 1))
else
    log_error "MCP adapter tests FAILED!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Summary
# ============================================================================

log_info ""
log_info "=============================================="
log_info "CHALLENGE SUMMARY"
log_info "=============================================="
log_info "Total tests: $TOTAL"
log_info "Passed: $PASSED"
log_info "Failed: $FAILED"

if [ $FAILED -eq 0 ]; then
    log_success "=============================================="
    log_success "ALL TESTS PASSED!"
    log_success "=============================================="
    exit 0
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED!"
    log_error "=============================================="
    exit 1
fi
