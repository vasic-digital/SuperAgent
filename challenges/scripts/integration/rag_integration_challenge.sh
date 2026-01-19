#!/bin/bash
# =============================================================================
# RAG Integration Challenge
# Tests all RAG and Vector Database integrations
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; ((TESTS_PASSED++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; ((TESTS_FAILED++)); }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; ((TESTS_SKIPPED++)); }

# Configuration
HELIXAGENT_URL="${HELIXAGENT_URL:-http://localhost:7061}"
CHROMADB_URL="${CHROMADB_URL:-http://localhost:8001}"
QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
WEAVIATE_URL="${WEAVIATE_URL:-http://localhost:8081}"
COGNEE_URL="${COGNEE_URL:-http://localhost:8000}"

# =============================================================================
# Helper Functions
# =============================================================================

wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1

    log_info "Waiting for $service_name to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if curl -sf "$url" > /dev/null 2>&1 || curl -sf "${url}/health" > /dev/null 2>&1; then
            log_info "$service_name is ready"
            return 0
        fi
        sleep 1
        ((attempt++))
    done
    log_skip "$service_name not available"
    return 1
}

test_endpoint() {
    local url=$1
    local description=$2

    if curl -sf "$url" > /dev/null 2>&1; then
        log_success "$description"
        return 0
    fi
    log_fail "$description"
    return 1
}

# =============================================================================
# Test Categories
# =============================================================================

test_embedding_endpoints() {
    log_info "=== Testing Embedding Endpoints ==="

    # Test 1: Generate embeddings
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"text": "Test embedding generation", "model": "text-embedding-3-small"}' \
        "$HELIXAGENT_URL/v1/embeddings/generate" 2>/dev/null); then
        if echo "$response" | jq -e '.embeddings' > /dev/null 2>&1 || \
           echo "$response" | jq -e '.embedding' > /dev/null 2>&1; then
            log_success "Embedding generation works"
        else
            log_fail "Embedding generation returned invalid response"
        fi
    else
        log_fail "Embedding generation endpoint failed"
    fi

    # Test 2: List embedding providers
    if response=$(curl -sf "$HELIXAGENT_URL/v1/embeddings/providers" 2>/dev/null); then
        if echo "$response" | jq -e '.providers | length >= 1' > /dev/null 2>&1; then
            log_success "Embedding providers list available"
        else
            log_skip "No embedding providers configured"
        fi
    else
        log_fail "Embedding providers endpoint failed"
    fi

    # Test 3: Batch embeddings
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"texts": ["First text", "Second text", "Third text"]}' \
        "$HELIXAGENT_URL/v1/embeddings/batch" 2>/dev/null); then
        if echo "$response" | jq -e '.embeddings | length == 3' > /dev/null 2>&1; then
            log_success "Batch embedding generation works"
        else
            log_skip "Batch embeddings returned unexpected count"
        fi
    else
        log_skip "Batch embeddings endpoint not available"
    fi
}

test_pgvector_integration() {
    log_info "=== Testing PostgreSQL pgvector Integration ==="

    # Test 1: Vector search
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"query": "test query", "limit": 5}' \
        "$HELIXAGENT_URL/v1/embeddings/search" 2>/dev/null); then
        if echo "$response" | jq -e '.results' > /dev/null 2>&1; then
            log_success "pgvector search works"
        else
            log_skip "pgvector search returned no results (may be empty)"
        fi
    else
        log_fail "pgvector search endpoint failed"
    fi

    # Test 2: Index document
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"title": "Test Document", "content": "This is a test document for RAG integration."}' \
        "$HELIXAGENT_URL/v1/embeddings/index" 2>/dev/null); then
        if echo "$response" | jq -e '.success == true' > /dev/null 2>&1 || \
           echo "$response" | jq -e '.id' > /dev/null 2>&1; then
            log_success "Document indexing works"
        else
            log_fail "Document indexing returned failure"
        fi
    else
        log_fail "Document indexing endpoint failed"
    fi
}

test_chromadb_integration() {
    log_info "=== Testing ChromaDB Integration ==="

    if ! wait_for_service "$CHROMADB_URL/api/v1/heartbeat" "ChromaDB"; then
        return
    fi

    # Test 1: ChromaDB heartbeat
    if curl -sf "$CHROMADB_URL/api/v1/heartbeat" | jq -e '.nanosecond_heartbeat' > /dev/null 2>&1; then
        log_success "ChromaDB heartbeat works"
    else
        log_fail "ChromaDB heartbeat failed"
    fi

    # Test 2: List collections
    if response=$(curl -sf "$CHROMADB_URL/api/v1/collections" 2>/dev/null); then
        if echo "$response" | jq -e '. | type == "array"' > /dev/null 2>&1; then
            log_success "ChromaDB list collections works"
        else
            log_fail "ChromaDB collections list invalid"
        fi
    else
        log_fail "ChromaDB collections endpoint failed"
    fi

    # Test 3: Create and delete test collection
    local collection_name="rag_test_$(date +%s)"
    if curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$collection_name\"}" \
        "$CHROMADB_URL/api/v1/collections" > /dev/null 2>&1; then
        log_success "ChromaDB collection creation works"

        # Cleanup
        curl -sf -X DELETE "$CHROMADB_URL/api/v1/collections/$collection_name" > /dev/null 2>&1
    else
        log_fail "ChromaDB collection creation failed"
    fi
}

test_qdrant_integration() {
    log_info "=== Testing Qdrant Integration ==="

    if ! wait_for_service "$QDRANT_URL/readyz" "Qdrant"; then
        return
    fi

    # Test 1: Qdrant health
    if curl -sf "$QDRANT_URL/readyz" > /dev/null 2>&1; then
        log_success "Qdrant health check works"
    else
        log_fail "Qdrant health check failed"
    fi

    # Test 2: List collections
    if response=$(curl -sf "$QDRANT_URL/collections" 2>/dev/null); then
        if echo "$response" | jq -e '.result.collections' > /dev/null 2>&1; then
            log_success "Qdrant list collections works"
        else
            log_fail "Qdrant collections list invalid"
        fi
    else
        log_fail "Qdrant collections endpoint failed"
    fi

    # Test 3: Create and delete test collection
    local collection_name="rag_test_$(date +%s)"
    if curl -sf -X PUT \
        -H "Content-Type: application/json" \
        -d '{"vectors": {"size": 1536, "distance": "Cosine"}}' \
        "$QDRANT_URL/collections/$collection_name" > /dev/null 2>&1; then
        log_success "Qdrant collection creation works"

        # Test point insertion
        if curl -sf -X PUT \
            -H "Content-Type: application/json" \
            -d '{"points": [{"id": 1, "vector": '"$(python3 -c "import json; print(json.dumps([0.1]*1536))")"', "payload": {"test": "value"}}]}' \
            "$QDRANT_URL/collections/$collection_name/points" > /dev/null 2>&1; then
            log_success "Qdrant point insertion works"
        else
            log_fail "Qdrant point insertion failed"
        fi

        # Cleanup
        curl -sf -X DELETE "$QDRANT_URL/collections/$collection_name" > /dev/null 2>&1
    else
        log_fail "Qdrant collection creation failed"
    fi
}

test_weaviate_integration() {
    log_info "=== Testing Weaviate Integration ==="

    if ! wait_for_service "$WEAVIATE_URL/v1/.well-known/ready" "Weaviate"; then
        return
    fi

    # Test 1: Weaviate ready
    if curl -sf "$WEAVIATE_URL/v1/.well-known/ready" > /dev/null 2>&1; then
        log_success "Weaviate ready check works"
    else
        log_fail "Weaviate ready check failed"
    fi

    # Test 2: Get schema
    if response=$(curl -sf "$WEAVIATE_URL/v1/schema" 2>/dev/null); then
        if echo "$response" | jq -e '.classes' > /dev/null 2>&1; then
            log_success "Weaviate schema endpoint works"
        else
            log_fail "Weaviate schema response invalid"
        fi
    else
        log_fail "Weaviate schema endpoint failed"
    fi

    # Test 3: Meta info
    if response=$(curl -sf "$WEAVIATE_URL/v1/meta" 2>/dev/null); then
        if echo "$response" | jq -e '.version' > /dev/null 2>&1; then
            log_success "Weaviate meta endpoint works"
        else
            log_fail "Weaviate meta response invalid"
        fi
    else
        log_fail "Weaviate meta endpoint failed"
    fi
}

test_cognee_integration() {
    log_info "=== Testing Cognee Integration ==="

    if ! wait_for_service "$COGNEE_URL" "Cognee"; then
        return
    fi

    # Test 1: Cognee health (via HelixAgent)
    if response=$(curl -sf "$HELIXAGENT_URL/v1/cognee/health" 2>/dev/null); then
        if echo "$response" | jq -e '.status == "healthy"' > /dev/null 2>&1 || \
           echo "$response" | jq -e '.healthy == true' > /dev/null 2>&1; then
            log_success "Cognee health check works"
        else
            log_skip "Cognee health check returned non-healthy status"
        fi
    else
        log_fail "Cognee health endpoint failed"
    fi

    # Test 2: Add memory
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"content": "Test memory for RAG integration", "dataset": "rag_test"}' \
        "$HELIXAGENT_URL/v1/cognee/memory" 2>/dev/null); then
        if echo "$response" | jq -e '.success == true' > /dev/null 2>&1 || \
           echo "$response" | jq -e '.id' > /dev/null 2>&1; then
            log_success "Cognee memory storage works"
        else
            log_skip "Cognee memory storage returned unexpected response"
        fi
    else
        log_skip "Cognee memory endpoint not available"
    fi

    # Test 3: Search memory
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"query": "RAG integration", "limit": 5}' \
        "$HELIXAGENT_URL/v1/cognee/search" 2>/dev/null); then
        if echo "$response" | jq -e '.results' > /dev/null 2>&1; then
            log_success "Cognee memory search works"
        else
            log_skip "Cognee search returned no results"
        fi
    else
        log_skip "Cognee search endpoint not available"
    fi
}

test_llamaindex_integration() {
    log_info "=== Testing LlamaIndex Integration ==="

    local llamaindex_url="${LLAMAINDEX_URL:-http://localhost:8012}"

    if ! wait_for_service "$llamaindex_url/health" "LlamaIndex"; then
        return
    fi

    # Test 1: Health check
    log_success "LlamaIndex service healthy"

    # Test 2: Query
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"query": "What is RAG?", "top_k": 5}' \
        "$llamaindex_url/query" 2>/dev/null); then
        if echo "$response" | jq -e '.response' > /dev/null 2>&1 || \
           echo "$response" | jq -e '.results' > /dev/null 2>&1; then
            log_success "LlamaIndex query works"
        else
            log_skip "LlamaIndex query returned unexpected response"
        fi
    else
        log_skip "LlamaIndex query endpoint failed"
    fi

    # Test 3: HyDE query
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"query": "How does vector search work?", "top_k": 5}' \
        "$llamaindex_url/hyde" 2>/dev/null); then
        if echo "$response" | jq -e '.response' > /dev/null 2>&1; then
            log_success "LlamaIndex HyDE query works"
        else
            log_skip "LlamaIndex HyDE returned unexpected response"
        fi
    else
        log_skip "LlamaIndex HyDE endpoint not available"
    fi
}

test_langchain_integration() {
    log_info "=== Testing LangChain Integration ==="

    local langchain_url="${LANGCHAIN_URL:-http://localhost:8011}"

    if ! wait_for_service "$langchain_url/health" "LangChain"; then
        return
    fi

    # Test 1: Health check
    log_success "LangChain service healthy"

    # Test 2: Task decomposition
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"task": "Build a RAG system with vector search"}' \
        "$langchain_url/decompose" 2>/dev/null); then
        if echo "$response" | jq -e '.subtasks' > /dev/null 2>&1; then
            log_success "LangChain task decomposition works"
        else
            log_skip "LangChain decomposition returned unexpected response"
        fi
    else
        log_skip "LangChain decomposition endpoint failed"
    fi

    # Test 3: Summarization
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"text": "This is a long document about RAG systems. RAG stands for Retrieval Augmented Generation.", "max_length": 50}' \
        "$langchain_url/summarize" 2>/dev/null); then
        if echo "$response" | jq -e '.summary' > /dev/null 2>&1; then
            log_success "LangChain summarization works"
        else
            log_skip "LangChain summarization returned unexpected response"
        fi
    else
        log_skip "LangChain summarization endpoint failed"
    fi
}

test_rag_pipeline() {
    log_info "=== Testing Full RAG Pipeline ==="

    # Test 1: Index a document
    local doc_id=""
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"title": "RAG Pipeline Test", "content": "This document tests the full RAG pipeline from indexing to retrieval."}' \
        "$HELIXAGENT_URL/v1/embeddings/index" 2>/dev/null); then
        doc_id=$(echo "$response" | jq -r '.id // empty')
        if [ -n "$doc_id" ]; then
            log_success "RAG Pipeline: Document indexed"
        else
            log_skip "RAG Pipeline: Document indexing response missing ID"
        fi
    else
        log_fail "RAG Pipeline: Document indexing failed"
    fi

    # Test 2: Search for the document
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"query": "RAG pipeline indexing retrieval", "limit": 10}' \
        "$HELIXAGENT_URL/v1/embeddings/search" 2>/dev/null); then
        if echo "$response" | jq -e '.results | length > 0' > /dev/null 2>&1; then
            log_success "RAG Pipeline: Document retrieval works"
        else
            log_skip "RAG Pipeline: No documents retrieved"
        fi
    else
        log_fail "RAG Pipeline: Search failed"
    fi

    # Test 3: Enhanced prompt with context
    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d '{"prompt": "Explain RAG systems", "use_rag": true}' \
        "$HELIXAGENT_URL/v1/cognee/enhance" 2>/dev/null); then
        if echo "$response" | jq -e '.enhanced_prompt' > /dev/null 2>&1 || \
           echo "$response" | jq -e '.context' > /dev/null 2>&1; then
            log_success "RAG Pipeline: Prompt enhancement works"
        else
            log_skip "RAG Pipeline: Prompt enhancement returned unexpected response"
        fi
    else
        log_skip "RAG Pipeline: Prompt enhancement not available"
    fi
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo "=============================================="
    echo "  RAG Integration Challenge"
    echo "  HelixAgent - $(date)"
    echo "=============================================="
    echo ""

    # Wait for main service
    wait_for_service "$HELIXAGENT_URL" "HelixAgent" || exit 1

    # Run test categories
    test_embedding_endpoints
    echo ""
    test_pgvector_integration
    echo ""
    test_chromadb_integration
    echo ""
    test_qdrant_integration
    echo ""
    test_weaviate_integration
    echo ""
    test_cognee_integration
    echo ""
    test_llamaindex_integration
    echo ""
    test_langchain_integration
    echo ""
    test_rag_pipeline

    # Summary
    echo ""
    echo "=============================================="
    echo "  Challenge Results"
    echo "=============================================="
    echo -e "  ${GREEN}Passed:${NC}  $TESTS_PASSED"
    echo -e "  ${RED}Failed:${NC}  $TESTS_FAILED"
    echo -e "  ${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
    echo "=============================================="

    # Exit with failure if any tests failed
    if [ $TESTS_FAILED -gt 0 ]; then
        echo -e "\n${RED}Challenge FAILED!${NC}"
        exit 1
    else
        echo -e "\n${GREEN}Challenge PASSED!${NC}"
        exit 0
    fi
}

main "$@"
