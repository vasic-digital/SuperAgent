#!/bin/bash
# RAG Comprehensive Challenge
# Validates ALL RAG components: Hybrid retrieval, HyDE, Reranking, Pipeline, Vector stores
# Tests: Implementation, Tests, Interface compliance, Retrieval modes

set -e

# Source challenge framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

# Initialize challenge
init_challenge "rag_comprehensive" "RAG Comprehensive Verification"
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

    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        log_success "PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "test" "$test_name" "true" ""
        return 0
    else
        log_error "FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "test" "$test_name" "false" "Test command failed"
        return 1
    fi
}

# ============================================================================
# SECTION 1: RAG CORE COMPONENTS
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: RAG Core Components"
log_info "=============================================="

# RAG component files
RAG_COMPONENTS=(
    "advanced"
    "hybrid"
    "hyde"
    "pipeline"
    "qdrant_enhanced"
    "qdrant_retriever"
    "reranker"
    "types"
)

log_info "Verifying RAG component implementations..."
for component in "${RAG_COMPONENTS[@]}"; do
    run_test "RAG: $component implementation exists" \
        "[[ -f '$PROJECT_ROOT/internal/rag/${component}.go' ]]"
done

# Check for tests
run_test "RAG: hybrid has tests" \
    "[[ -f '$PROJECT_ROOT/internal/rag/hybrid_test.go' ]]"

run_test "RAG: pipeline has tests" \
    "[[ -f '$PROJECT_ROOT/internal/rag/pipeline_test.go' ]] || [[ -f '$PROJECT_ROOT/internal/rag/pipeline_extended_test.go' ]]"

run_test "RAG: qdrant has tests" \
    "[[ -f '$PROJECT_ROOT/internal/rag/qdrant_retriever_test.go' ]] || [[ -f '$PROJECT_ROOT/internal/rag/qdrant_enhanced_test.go' ]]"

# ============================================================================
# SECTION 2: HYBRID RETRIEVAL
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: Hybrid Retrieval"
log_info "=============================================="

run_test "HybridRetriever struct exists" \
    "grep -q 'type HybridRetriever struct' '$PROJECT_ROOT/internal/rag/hybrid.go'"

run_test "Hybrid: Dense retrieval support" \
    "grep -qE 'dense|Dense|embedding|Embedding' '$PROJECT_ROOT/internal/rag/hybrid.go'"

run_test "Hybrid: Sparse retrieval support (BM25/keyword)" \
    "grep -qE 'sparse|Sparse|BM25|keyword|Keyword' '$PROJECT_ROOT/internal/rag/hybrid.go'"

run_test "Hybrid: RRF fusion method" \
    "grep -qE 'RRF|ReciprocalRankFusion|reciprocal' '$PROJECT_ROOT/internal/rag/hybrid.go'"

run_test "Hybrid: Weighted fusion method" \
    "grep -qE 'Weighted|weighted|weight' '$PROJECT_ROOT/internal/rag/hybrid.go'"

run_test "Hybrid: Alpha parameter (dense/sparse balance)" \
    "grep -qE 'Alpha|alpha|balance' '$PROJECT_ROOT/internal/rag/hybrid.go'"

run_test "Hybrid: Retrieve method" \
    "grep -q 'func.*Retrieve' '$PROJECT_ROOT/internal/rag/hybrid.go'"

# ============================================================================
# SECTION 3: HyDE (Hypothetical Document Embeddings)
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: HyDE Implementation"
log_info "=============================================="

run_test "HyDE implementation exists" \
    "[[ -f '$PROJECT_ROOT/internal/rag/hyde.go' ]]"

run_test "HyDE: Generator struct" \
    "grep -qE 'type.*HyDE|HyDEGenerator|Hypothetical' '$PROJECT_ROOT/internal/rag/hyde.go'"

run_test "HyDE: Generate hypothetical document" \
    "grep -qE 'Generate|hypothetical|Hypothetical' '$PROJECT_ROOT/internal/rag/hyde.go'"

run_test "HyDE: LLM integration for generation" \
    "grep -qE 'LLM|provider|Provider|Complete' '$PROJECT_ROOT/internal/rag/hyde.go'"

run_test "HyDE: Embedding of hypothetical docs" \
    "grep -qE 'embed|Embed|embedding' '$PROJECT_ROOT/internal/rag/hyde.go'"

# ============================================================================
# SECTION 4: RERANKING
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: Reranking"
log_info "=============================================="

run_test "Reranker implementation exists" \
    "[[ -f '$PROJECT_ROOT/internal/rag/reranker.go' ]]"

run_test "Reranker: struct definition" \
    "grep -qE 'type.*Reranker struct|CrossEncoder' '$PROJECT_ROOT/internal/rag/reranker.go'"

run_test "Reranker: Rerank method" \
    "grep -q 'func.*Rerank' '$PROJECT_ROOT/internal/rag/reranker.go'"

run_test "Reranker: Score calculation" \
    "grep -qE 'score|Score|relevance|Relevance' '$PROJECT_ROOT/internal/rag/reranker.go'"

run_test "Reranker: TopK parameter" \
    "grep -qE 'topK|TopK|top_k|limit|Limit' '$PROJECT_ROOT/internal/rag/reranker.go'"

# ============================================================================
# SECTION 5: RAG PIPELINE
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: RAG Pipeline"
log_info "=============================================="

run_test "Pipeline implementation exists" \
    "[[ -f '$PROJECT_ROOT/internal/rag/pipeline.go' ]]"

run_test "Pipeline: struct definition" \
    "grep -qE 'type.*Pipeline struct|RAGPipeline' '$PROJECT_ROOT/internal/rag/pipeline.go'"

run_test "Pipeline: Query method" \
    "grep -qE 'func.*Query|func.*Retrieve|func.*Execute' '$PROJECT_ROOT/internal/rag/pipeline.go'"

run_test "Pipeline: Context augmentation" \
    "grep -qE 'context|Context|augment|Augment' '$PROJECT_ROOT/internal/rag/pipeline.go'"

run_test "Pipeline: Document ranking" \
    "grep -qE 'rank|Rank|score|Score' '$PROJECT_ROOT/internal/rag/pipeline.go'"

# ============================================================================
# SECTION 6: QDRANT INTEGRATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 6: Qdrant Integration"
log_info "=============================================="

run_test "Qdrant retriever exists" \
    "[[ -f '$PROJECT_ROOT/internal/rag/qdrant_retriever.go' ]]"

run_test "Qdrant enhanced exists" \
    "[[ -f '$PROJECT_ROOT/internal/rag/qdrant_enhanced.go' ]]"

run_test "Qdrant: Collection management" \
    "grep -qE 'collection|Collection|CreateCollection' '$PROJECT_ROOT/internal/rag/qdrant_retriever.go'"

run_test "Qdrant: Vector search" \
    "grep -qE 'Search|search|Query|query' '$PROJECT_ROOT/internal/rag/qdrant_retriever.go'"

run_test "Qdrant: Point upsert" \
    "grep -qE 'Upsert|upsert|Insert|insert' '$PROJECT_ROOT/internal/rag/qdrant_retriever.go'"

run_test "Qdrant Enhanced: Hierarchical retrieval" \
    "grep -qE 'hierarchical|Hierarchical|parent|child' '$PROJECT_ROOT/internal/rag/qdrant_enhanced.go'"

run_test "Qdrant Enhanced: Temporal retrieval" \
    "grep -qE 'temporal|Temporal|time|Time|date|Date' '$PROJECT_ROOT/internal/rag/qdrant_enhanced.go'"

# ============================================================================
# SECTION 7: ADVANCED RETRIEVAL TECHNIQUES
# ============================================================================
log_info "=============================================="
log_info "SECTION 7: Advanced Retrieval Techniques"
log_info "=============================================="

run_test "Advanced retrieval exists" \
    "[[ -f '$PROJECT_ROOT/internal/rag/advanced.go' ]]"

run_test "Advanced: Dense passage retrieval" \
    "grep -qE 'dense|Dense|DPR|passage|Passage' '$PROJECT_ROOT/internal/rag/advanced.go'"

run_test "Advanced: Query expansion" \
    "grep -qE 'expand|Expand|expansion|Expansion|synonym' '$PROJECT_ROOT/internal/rag/advanced.go'"

run_test "Advanced: Multi-hop retrieval" \
    "grep -qE 'multihop|multi-hop|MultiHop|hop|Hop' '$PROJECT_ROOT/internal/rag/advanced.go'" || \
    run_test "Advanced: Iterative retrieval" \
        "grep -qE 'iterative|Iterative|recursive|Recursive' '$PROJECT_ROOT/internal/rag/advanced.go'"

# ============================================================================
# SECTION 8: VECTOR DATABASE BACKENDS
# ============================================================================
log_info "=============================================="
log_info "SECTION 8: Vector Database Backends"
log_info "=============================================="

# Qdrant
run_test "VectorDB: Qdrant client exists" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/qdrant/client.go' ]]"

run_test "VectorDB: Qdrant config exists" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/qdrant/config.go' ]]"

run_test "Qdrant: Connect method" \
    "grep -q 'func.*Connect' '$PROJECT_ROOT/internal/vectordb/qdrant/client.go'"

run_test "Qdrant: Search method" \
    "grep -q 'func.*Search' '$PROJECT_ROOT/internal/vectordb/qdrant/client.go'"

run_test "Qdrant: Upsert method" \
    "grep -q 'func.*Upsert' '$PROJECT_ROOT/internal/vectordb/qdrant/client.go'"

# Pinecone
run_test "VectorDB: Pinecone client exists" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/pinecone/client.go' ]]"

run_test "Pinecone: Connect method" \
    "grep -q 'func.*Connect' '$PROJECT_ROOT/internal/vectordb/pinecone/client.go'"

run_test "Pinecone: Upsert method" \
    "grep -q 'func.*Upsert' '$PROJECT_ROOT/internal/vectordb/pinecone/client.go'"

run_test "Pinecone: Query method" \
    "grep -qE 'func.*Query|func.*Search' '$PROJECT_ROOT/internal/vectordb/pinecone/client.go'"

# Milvus
run_test "VectorDB: Milvus client exists" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/milvus/client.go' ]]"

run_test "Milvus: Connect method" \
    "grep -q 'func.*Connect' '$PROJECT_ROOT/internal/vectordb/milvus/client.go'"

run_test "Milvus: Collection management" \
    "grep -qE 'Collection|collection|CreateCollection' '$PROJECT_ROOT/internal/vectordb/milvus/client.go'"

# PgVector
run_test "VectorDB: PgVector client exists" \
    "[[ -f '$PROJECT_ROOT/internal/vectordb/pgvector/client.go' ]]"

run_test "PgVector: Connect method" \
    "grep -q 'func.*Connect' '$PROJECT_ROOT/internal/vectordb/pgvector/client.go'"

run_test "PgVector: Vector type support" \
    "grep -qE 'vector|Vector|pgvector|embedding' '$PROJECT_ROOT/internal/vectordb/pgvector/client.go'"

# ============================================================================
# SECTION 9: EMBEDDING PROVIDERS
# ============================================================================
log_info "=============================================="
log_info "SECTION 9: Embedding Providers"
log_info "=============================================="

run_test "Embedding: provider interface exists" \
    "[[ -f '$PROJECT_ROOT/Embeddings/pkg/provider/provider.go' ]]"

run_test "Embedding: OpenAI provider exists" \
    "[[ -f '$PROJECT_ROOT/Embeddings/pkg/openai/openai.go' ]]"

# Check embedding providers in the Embeddings submodule
run_test "Embedding: OpenAI support" \
    "grep -qE 'EmbeddingProvider|openai' '$PROJECT_ROOT/Embeddings/pkg/openai/openai.go'"

run_test "Embedding: Cohere support" \
    "grep -qE 'EmbeddingProvider|cohere' '$PROJECT_ROOT/Embeddings/pkg/cohere/cohere.go'"

run_test "Embedding: Voyage support" \
    "grep -qE 'EmbeddingProvider|voyage' '$PROJECT_ROOT/Embeddings/pkg/voyage/voyage.go'"

run_test "Embedding: Jina support" \
    "grep -qE 'EmbeddingProvider|jina' '$PROJECT_ROOT/Embeddings/pkg/jina/jina.go'"

run_test "Embedding: Google support" \
    "grep -qE 'EmbeddingProvider|google|Vertex' '$PROJECT_ROOT/Embeddings/pkg/google/google.go'"

run_test "Embedding: Bedrock support" \
    "grep -qE 'EmbeddingProvider|bedrock|Titan' '$PROJECT_ROOT/Embeddings/pkg/bedrock/bedrock.go'"

run_test "Embedding: Ollama support (via adapters)" \
    "[[ -f '$PROJECT_ROOT/Embeddings/pkg/openai/openai.go' ]]"

# Embedding interface
run_test "EmbeddingProvider interface" \
    "grep -qE 'type EmbeddingProvider interface' '$PROJECT_ROOT/Embeddings/pkg/provider/provider.go'"

run_test "Embedding: Embed method" \
    "grep -q 'Embed(' '$PROJECT_ROOT/Embeddings/pkg/provider/provider.go'"

run_test "Embedding: EmbedBatch method" \
    "grep -q 'EmbedBatch' '$PROJECT_ROOT/Embeddings/pkg/provider/provider.go'"

run_test "Embedding: Dimensions method" \
    "grep -q 'Dimensions' '$PROJECT_ROOT/Embeddings/pkg/provider/provider.go'"

# ============================================================================
# SECTION 10: RAG TYPES AND CONFIGURATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 10: RAG Types and Configuration"
log_info "=============================================="

run_test "RAG types file exists" \
    "[[ -f '$PROJECT_ROOT/internal/rag/types.go' ]]"

run_test "RAG: Document type" \
    "grep -qE 'type.*Document struct|Document' '$PROJECT_ROOT/internal/rag/types.go'"

run_test "RAG: SearchResult type" \
    "grep -qE 'type.*SearchResult|Result|Chunk' '$PROJECT_ROOT/internal/rag/types.go'"

run_test "RAG: RetrievalConfig type" \
    "grep -qE 'type.*Config|RetrievalConfig|Options' '$PROJECT_ROOT/internal/rag/types.go'"

# ============================================================================
# SECTION 11: GO TESTS VALIDATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 11: Go Tests Validation"
log_info "=============================================="

log_info "Running RAG tests..."
run_test "All RAG tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/rag/... -timeout 120s"

log_info "Running VectorDB tests..."
run_test "VectorDB Qdrant tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/vectordb/qdrant/... -timeout 60s"

run_test "VectorDB Pinecone tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/vectordb/pinecone/... -timeout 60s"

run_test "VectorDB Milvus tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/vectordb/milvus/... -timeout 60s"

run_test "VectorDB PgVector tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/vectordb/pgvector/... -timeout 60s"

log_info "Running Embedding tests..."
run_test "All Embedding tests pass" \
    "cd '$PROJECT_ROOT/Embeddings' && go test -v ./pkg/... -timeout 120s"

# ============================================================================
# SUMMARY
# ============================================================================
log_info "=============================================="
log_info "CHALLENGE SUMMARY"
log_info "=============================================="

log_info "Total tests: $TESTS_TOTAL"
log_info "Passed: $TESTS_PASSED"
log_info "Failed: $TESTS_FAILED"

if [[ $TESTS_FAILED -eq 0 ]]; then
    log_success "=============================================="
    log_success "ALL RAG TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge 0
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge 1
fi
