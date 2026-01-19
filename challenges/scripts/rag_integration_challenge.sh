#!/bin/bash

# RAG Integration Challenge
# This challenge verifies that the RAG system with all components is working correctly.
# All tests MUST pass for the challenge to succeed.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "RAG Integration Challenge"
echo "=========================================="
echo ""

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "Testing: $test_name... "

    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

cd "$PROJECT_ROOT"

echo "1. Running RAG Package Unit Tests"
echo "-----------------------------------"

# Test RAG pipeline
run_test "RAG Pipeline Creation" "go test -v -run TestPipeline_Creation ./internal/rag/..."
run_test "RAG Pipeline Chunking" "go test -v -run TestPipeline_ChunkDocument ./internal/rag/..."
run_test "RAG Pipeline Initialize" "go test -v -run TestPipeline_Initialize ./internal/rag/..."

# Test Advanced RAG
run_test "Advanced RAG Config" "go test -v -run TestDefaultConfigs ./internal/rag/..."
run_test "Advanced RAG Initialize" "go test -v -run TestAdvancedRAG_Initialize ./internal/rag/..."
run_test "Advanced RAG Query Expansion" "go test -v -run TestAdvancedRAG_ExpandQuery ./internal/rag/..."
run_test "Advanced RAG Re-ranking" "go test -v -run TestAdvancedRAG_ReRank ./internal/rag/..."
run_test "Advanced RAG Compression" "go test -v -run TestAdvancedRAG_CompressContext ./internal/rag/..."
run_test "Advanced RAG Full Workflow" "go test -v -run TestAdvancedRAG_FullWorkflow ./internal/rag/..."

# Test helper functions
run_test "Tokenization" "go test -v -run TestTokenize ./internal/rag/..."
run_test "Sentence Splitting" "go test -v -run TestSplitIntoSentences ./internal/rag/..."
run_test "String Similarity" "go test -v -run TestSimilarity ./internal/rag/..."

# Test HyDE
run_test "HyDE Config" "go test -v -run TestDefaultHyDEConfig ./internal/rag/..."
run_test "HyDE Generator Creation" "go test -v -run TestNewHyDEGenerator ./internal/rag/..."
run_test "HyDE Expand Query" "go test -v -run TestHyDEGenerator_ExpandQuery ./internal/rag/..."
run_test "HyDE Aggregation Methods" "go test -v -run TestHyDEGenerator_AggregationMethods ./internal/rag/..."

echo ""
echo "2. Running RAG API Handler Tests"
echo "---------------------------------"

run_test "RAG Handler Creation" "go test -v -run TestNewRAGHandler ./internal/handlers/..."
run_test "RAG Handler Health" "go test -v -run TestRAGHandler_Health ./internal/handlers/..."
run_test "RAG Handler ChunkDocument" "go test -v -run TestRAGHandler_ChunkDocument ./internal/handlers/..."
run_test "RAG Handler ExpandQuery" "go test -v -run TestRAGHandler_ExpandQuery ./internal/handlers/..."
run_test "RAG Handler Search" "go test -v -run TestRAGHandler_Search ./internal/handlers/..."
run_test "RAG Handler ReRank" "go test -v -run TestRAGHandler_ReRank ./internal/handlers/..."
run_test "RAG Handler Compress" "go test -v -run TestRAGHandler_CompressContext ./internal/handlers/..."

echo ""
echo "3. Running RAG Integration Tests"
echo "---------------------------------"

run_test "Pipeline Integration" "go test -v -run TestRAGPipeline_Integration ./tests/integration/rag_integration_test.go"
run_test "Advanced RAG Integration" "go test -v -run TestAdvancedRAG_Integration ./tests/integration/rag_integration_test.go"
run_test "Embedding Registry Integration" "go test -v -run TestEmbeddingModelRegistry_Integration ./tests/integration/rag_integration_test.go"
run_test "End-to-End RAG" "go test -v -run TestRAGWithAdvancedRAG_EndToEnd ./tests/integration/rag_integration_test.go"
run_test "Concurrent Operations" "go test -v -run TestConcurrentRAGOperations ./tests/integration/rag_integration_test.go"

echo ""
echo "4. Running Vector Database Adapter Tests"
echo "-----------------------------------------"

run_test "ChromaDB Adapter" "go test -v -run TestChromaAdapter ./internal/mcp/servers/..."
run_test "Qdrant Adapter" "go test -v -run TestQdrantAdapter ./internal/mcp/servers/..."
run_test "Weaviate Adapter" "go test -v -run TestWeaviateAdapter ./internal/mcp/servers/..."

echo ""
echo "5. Running Embedding Model Tests"
echo "---------------------------------"

run_test "Embedding Registry Creation" "go test -v -run TestNewEmbeddingModelRegistry ./internal/embeddings/models/..."
run_test "Embedding Model Registration" "go test -v -run TestEmbeddingModelRegistry_Register ./internal/embeddings/models/..."
run_test "Embedding Fallback Chain" "go test -v -run TestEmbeddingModelRegistry_EncodeWithFallback ./internal/embeddings/models/..."

echo ""
echo "=========================================="
echo "CHALLENGE RESULTS"
echo "=========================================="
echo ""
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}=========================================="
    echo "CHALLENGE PASSED!"
    echo "==========================================${NC}"
    exit 0
else
    echo -e "${RED}=========================================="
    echo "CHALLENGE FAILED!"
    echo "==========================================${NC}"
    echo ""
    echo "Please fix the failing tests before proceeding."
    exit 1
fi
