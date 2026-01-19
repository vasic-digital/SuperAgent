#!/bin/bash

# MCP Servers Integration Challenge
# This challenge verifies that all MCP server adapters are working correctly.
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
echo "MCP Servers Integration Challenge"
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

echo "1. Vector Database MCP Servers"
echo "-------------------------------"

# ChromaDB
run_test "ChromaDB Adapter Creation" "go test -v -run TestChromaAdapter_NewChromaAdapter ./internal/mcp/servers/..."
run_test "ChromaDB Add Documents" "go test -v -run TestChromaAdapter_AddDocuments ./internal/mcp/servers/..."
run_test "ChromaDB Query" "go test -v -run TestChromaAdapter_Query ./internal/mcp/servers/..."
run_test "ChromaDB MCP Tools" "go test -v -run TestChromaAdapter_GetMCPTools ./internal/mcp/servers/..."

# Qdrant
run_test "Qdrant Adapter Creation" "go test -v -run TestQdrantAdapter_NewQdrantAdapter ./internal/mcp/servers/..."
run_test "Qdrant Upsert Points" "go test -v -run TestQdrantAdapter_UpsertPoints ./internal/mcp/servers/..."
run_test "Qdrant Search" "go test -v -run TestQdrantAdapter_Search ./internal/mcp/servers/..."
run_test "Qdrant MCP Tools" "go test -v -run TestQdrantAdapter_GetMCPTools ./internal/mcp/servers/..."

# Weaviate
run_test "Weaviate Adapter Creation" "go test -v -run TestWeaviateAdapter_NewWeaviateAdapter ./internal/mcp/servers/..."
run_test "Weaviate Add Objects" "go test -v -run TestWeaviateAdapter_AddObjects ./internal/mcp/servers/..."
run_test "Weaviate Search" "go test -v -run TestWeaviateAdapter_Search ./internal/mcp/servers/..."
run_test "Weaviate MCP Tools" "go test -v -run TestWeaviateAdapter_GetMCPTools ./internal/mcp/servers/..."

echo ""
echo "2. Design & UI MCP Servers"
echo "---------------------------"

# Figma
run_test "Figma Adapter Creation" "go test -v -run TestFigmaAdapter_NewFigmaAdapter ./internal/mcp/servers/..."
run_test "Figma Get File" "go test -v -run TestFigmaAdapter_GetFile ./internal/mcp/servers/..."
run_test "Figma Get Components" "go test -v -run TestFigmaAdapter_GetComponents ./internal/mcp/servers/..."
run_test "Figma MCP Tools" "go test -v -run TestFigmaAdapter_GetMCPTools ./internal/mcp/servers/..."

# Miro
run_test "Miro Adapter Creation" "go test -v -run TestMiroAdapter_NewMiroAdapter ./internal/mcp/servers/..."
run_test "Miro Get Boards" "go test -v -run TestMiroAdapter_GetBoards ./internal/mcp/servers/..."
run_test "Miro Board Items" "go test -v -run TestMiroAdapter_GetBoardItems ./internal/mcp/servers/..."
run_test "Miro MCP Tools" "go test -v -run TestMiroAdapter_GetMCPTools ./internal/mcp/servers/..."

echo ""
echo "3. Image Generation MCP Servers"
echo "--------------------------------"

# Replicate
run_test "Replicate Adapter Creation" "go test -v -run TestReplicateAdapter_NewReplicateAdapter ./internal/mcp/servers/..."
run_test "Replicate Create Prediction" "go test -v -run TestReplicateAdapter_CreatePrediction ./internal/mcp/servers/..."
run_test "Replicate Get Prediction" "go test -v -run TestReplicateAdapter_GetPrediction ./internal/mcp/servers/..."
run_test "Replicate MCP Tools" "go test -v -run TestReplicateAdapter_GetMCPTools ./internal/mcp/servers/..."

# Stable Diffusion
run_test "SD Adapter Creation" "go test -v -run TestStableDiffusionAdapter_NewStableDiffusionAdapter ./internal/mcp/servers/..."
run_test "SD Text-to-Image" "go test -v -run TestStableDiffusionAdapter_TextToImage ./internal/mcp/servers/..."
run_test "SD Get Models" "go test -v -run TestStableDiffusionAdapter_GetModels ./internal/mcp/servers/..."
run_test "SD MCP Tools" "go test -v -run TestStableDiffusionAdapter_GetMCPTools ./internal/mcp/servers/..."

echo ""
echo "4. Core MCP Servers (Filesystem, Git, Memory, SVGMaker)"
echo "--------------------------------------------------------"

# Filesystem
run_test "Filesystem Adapter Creation" "go test -v -run TestNewFilesystemAdapter ./internal/mcp/servers/..."
run_test "Filesystem Read File" "go test -v -run TestFilesystemAdapter_ReadFile ./internal/mcp/servers/..."
run_test "Filesystem Write File" "go test -v -run TestFilesystemAdapter_WriteFile ./internal/mcp/servers/..."
run_test "Filesystem MCP Tools" "go test -v -run TestFilesystemAdapter_GetMCPTools ./internal/mcp/servers/..."

# Git
run_test "Git Adapter Creation" "go test -v -run TestNewGitAdapter ./internal/mcp/servers/..."
run_test "Git Status" "go test -v -run TestGitAdapter_Status ./internal/mcp/servers/..."
run_test "Git Commit" "go test -v -run TestGitAdapter_Commit ./internal/mcp/servers/..."
run_test "Git MCP Tools" "go test -v -run TestGitAdapter_GetMCPTools ./internal/mcp/servers/..."

# Memory
run_test "Memory Adapter Creation" "go test -v -run TestNewMemoryAdapter ./internal/mcp/servers/..."
run_test "Memory Create Entity" "go test -v -run TestMemoryAdapter_CreateEntity ./internal/mcp/servers/..."
run_test "Memory Create Relation" "go test -v -run TestMemoryAdapter_CreateRelation ./internal/mcp/servers/..."
run_test "Memory MCP Tools" "go test -v -run TestMemoryAdapter_GetMCPTools ./internal/mcp/servers/..."

# SVGMaker
run_test "SVGMaker Adapter Creation" "go test -v -run TestNewSVGMakerAdapter ./internal/mcp/servers/..."
run_test "SVGMaker Create SVG" "go test -v -run TestSVGMakerAdapter_CreateSVG ./internal/mcp/servers/..."
run_test "SVGMaker Bar Chart" "go test -v -run TestSVGMakerAdapter_CreateBarChart ./internal/mcp/servers/..."
run_test "SVGMaker MCP Tools" "go test -v -run TestSVGMakerAdapter_GetMCPTools ./internal/mcp/servers/..."

echo ""
echo "5. Database MCP Servers (PostgreSQL, SQLite)"
echo "---------------------------------------------"

# PostgreSQL
run_test "PostgreSQL Adapter Creation" "go test -v -run TestNewPostgresAdapter ./internal/mcp/servers/..."
run_test "PostgreSQL Schema Check" "go test -v -run TestPostgresAdapter_IsSchemaAllowed ./internal/mcp/servers/..."
run_test "PostgreSQL Query Check" "go test -v -run TestPostgresAdapter_IsReadOnlyQuery ./internal/mcp/servers/..."
run_test "PostgreSQL MCP Tools" "go test -v -run TestPostgresAdapter_GetMCPTools ./internal/mcp/servers/..."

# SQLite
run_test "SQLite Adapter Creation" "go test -v -run TestNewSQLiteAdapter ./internal/mcp/servers/..."
run_test "SQLite Initialize InMemory" "go test -v -run TestSQLiteAdapter_Initialize_InMemory ./internal/mcp/servers/..."
run_test "SQLite Query" "go test -v -run TestSQLiteAdapter_Query ./internal/mcp/servers/..."
run_test "SQLite MCP Tools" "go test -v -run TestSQLiteAdapter_GetMCPTools ./internal/mcp/servers/..."

echo ""
echo "6. API Integration MCP Servers (GitHub, Fetch)"
echo "-----------------------------------------------"

# GitHub
run_test "GitHub Adapter Creation" "go test -v -run TestNewGitHubAdapter ./internal/mcp/servers/..."
run_test "GitHub Get User" "go test -v -run TestGitHubAdapter_GetUser_WithMockServer ./internal/mcp/servers/..."
run_test "GitHub List Repos" "go test -v -run TestGitHubAdapter_ListRepositories_WithMockServer ./internal/mcp/servers/..."
run_test "GitHub MCP Tools" "go test -v -run TestGitHubAdapter_GetMCPTools ./internal/mcp/servers/..."

# Fetch
run_test "Fetch Adapter Creation" "go test -v -run TestNewFetchAdapter ./internal/mcp/servers/..."
run_test "Fetch URL" "go test -v -run TestFetchAdapter_Fetch_WithMockServer ./internal/mcp/servers/..."
run_test "Fetch Extract Links" "go test -v -run TestFetchAdapter_ExtractLinks ./internal/mcp/servers/..."
run_test "Fetch MCP Tools" "go test -v -run TestFetchAdapter_GetMCPTools ./internal/mcp/servers/..."

echo ""
echo "7. Cache MCP Servers (Redis)"
echo "----------------------------"

# Redis
run_test "Redis Adapter Creation" "go test -v -run TestNewRedisAdapter ./internal/mcp/servers/..."
run_test "Redis Key Prefix" "go test -v -run TestRedisAdapter_prefixKey ./internal/mcp/servers/..."
run_test "Redis MCP Tools" "go test -v -run TestRedisAdapter_GetMCPTools ./internal/mcp/servers/..."
run_test "Redis Capabilities" "go test -v -run TestRedisAdapter_GetCapabilities ./internal/mcp/servers/..."

echo ""
echo "8. LSP Server Registry"
echo "-----------------------"

run_test "LSP Registry Creation" "go test -v -run TestNewLSPServerRegistry ./internal/lsp/servers/..."
run_test "LSP Server Registration" "go test -v -run TestLSPServerRegistry_Register ./internal/lsp/servers/..."
run_test "LSP Completion" "go test -v -run TestLSPServerRegistry_Completion ./internal/lsp/servers/..."
run_test "LSP Diagnostics" "go test -v -run TestLSPServerRegistry_Diagnostics ./internal/lsp/servers/..."

echo ""
echo "9. Embedding Model Registry"
echo "----------------------------"

run_test "Embedding Registry Creation" "go test -v -run TestNewEmbeddingModelRegistry ./internal/embeddings/models/..."
run_test "Embedding Model Registration" "go test -v -run TestEmbeddingModelRegistry_Register ./internal/embeddings/models/..."
run_test "Embedding Encode" "go test -v -run TestEmbeddingModelRegistry_Encode ./internal/embeddings/models/..."
run_test "Embedding Fallback" "go test -v -run TestEmbeddingModelRegistry_EncodeWithFallback ./internal/embeddings/models/..."

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
