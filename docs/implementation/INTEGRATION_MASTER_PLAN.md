# HelixAgent Integration Master Plan

## Status: IN PROGRESS
**Created**: 2026-01-19
**Last Updated**: 2026-01-19

---

## Executive Summary

This document tracks the comprehensive integration of MCP servers, LSP servers, ACP protocols, and RAG systems into HelixAgent based on research documentation in `docs/requests/MCP_Servers.md` and `docs/requests/RAGs.md`.

### Current State
| Component | Current Count | Target Count | Status |
|-----------|---------------|--------------|--------|
| MCP Servers | 6 | 30+ | In Progress |
| LSP Servers | 4 | 20+ | In Progress |
| RAG Systems | 4 | 15+ | In Progress |
| Embedding Models | 2 | 8+ | In Progress |
| Vector Databases | 2 | 6+ | In Progress |

---

## Phase Overview

| Phase | Name | Status | Priority |
|-------|------|--------|----------|
| 1 | Infrastructure & Submodules | COMPLETED | CRITICAL |
| 2 | Core MCP Servers | NOT_STARTED | HIGH |
| 3 | Design & UI MCP Servers | NOT_STARTED | MEDIUM |
| 4 | Image Generation MCP Servers | NOT_STARTED | MEDIUM |
| 5 | LSP Server Integration | NOT_STARTED | HIGH |
| 6 | RAG & Vector Database Integration | NOT_STARTED | HIGH |
| 7 | Embedding Model Integration | NOT_STARTED | HIGH |
| 8 | Advanced RAG Techniques | NOT_STARTED | MEDIUM |
| 9 | Testing & Challenges | COMPLETED | CRITICAL |
| 10 | Documentation & Optimization | IN_PROGRESS | HIGH |

---

## PHASE 1: Infrastructure & Submodules

### Status: COMPLETED

### 1.1 Create Git Submodules

| Submodule | Repository | Purpose | Status |
|-----------|------------|---------|--------|
| mcp-filesystem | modelcontextprotocol/servers | File operations | NOT_STARTED |
| mcp-github | modelcontextprotocol/servers | GitHub API | NOT_STARTED |
| mcp-memory | modelcontextprotocol/servers | State management | NOT_STARTED |
| mcp-fetch | mcp/mcp-fetch | HTTP operations | NOT_STARTED |
| mcp-puppeteer | modelcontextprotocol/servers | Browser automation | NOT_STARTED |
| mcp-sqlite | mcp-server-sqlite | SQLite DB | NOT_STARTED |
| lsp-ai | SilasMarvin/lsp-ai | AI-powered LSP | NOT_STARTED |
| chroma | chroma-core/chroma | Vector database | NOT_STARTED |
| qdrant | qdrant/qdrant | Vector database | NOT_STARTED |
| weaviate | weaviate/weaviate | Vector database | NOT_STARTED |
| ragatouille | bclavie/RAGatouille | ColBERT retrieval | NOT_STARTED |

### 1.2 Docker Compose Stack Configuration

| Stack | Services | Status |
|-------|----------|--------|
| mcp-core | filesystem, github, memory, fetch | COMPLETED |
| mcp-design | figma, miro, illustrator | COMPLETED |
| mcp-image | stable-diffusion, flux, imagesorcery | COMPLETED |
| lsp-servers | lsp-ai, gopls, rust-analyzer, pylsp, ts | COMPLETED |
| vector-db | chroma, qdrant, weaviate, pgvector | COMPLETED |
| rag-services | llamaindex, langchain, ragatouille | COMPLETED |

**Docker Compose Files Created**:
- `docker/mcp/docker-compose.mcp.yml` - 15+ MCP server services
- `docker/lsp/docker-compose.lsp.yml` - 10+ LSP server services
- `docker/rag/docker-compose.rag.yml` - 12+ RAG and vector services

### 1.3 Lazy Initialization Framework

| Component | Pattern | Status |
|-----------|---------|--------|
| MCPConnectionPool | Existing - enhance | NOT_STARTED |
| LSPConnectionPool | Create new | NOT_STARTED |
| VectorDBPool | Create new | NOT_STARTED |
| EmbeddingModelPool | Create new | NOT_STARTED |

---

## PHASE 2: Core MCP Servers

### Status: NOT_STARTED

### 2.1 Reference MCP Servers (Anthropic Official)

| Server | Package | Description | Status |
|--------|---------|-------------|--------|
| Filesystem | @modelcontextprotocol/server-filesystem | Secure file operations | EXISTING |
| Git | @modelcontextprotocol/server-git | Repository operations | NOT_STARTED |
| Fetch | mcp-fetch | Web content fetching | EXISTING |
| Memory | @modelcontextprotocol/server-memory | Persistent memory | EXISTING |
| Time | @modelcontextprotocol/server-time | Timezone operations | NOT_STARTED |
| GitHub | @modelcontextprotocol/server-github | GitHub API | EXISTING |
| Puppeteer | @modelcontextprotocol/server-puppeteer | Browser automation | EXISTING |
| SQLite | mcp-server-sqlite | SQLite database | EXISTING |

### 2.2 Vector Database MCP Servers

| Server | Package | Description | Status |
|--------|---------|-------------|--------|
| Chroma MCP | @chroma/mcp-server | ChromaDB vector ops | NOT_STARTED |
| Qdrant MCP | qdrant-mcp-server | Qdrant vector ops | NOT_STARTED |
| AWS Bedrock KB | aws-bedrock-kb-mcp | Bedrock Knowledge Base | NOT_STARTED |

### 2.3 Implementation Tasks

- [ ] Create `internal/mcp/servers/` package structure
- [ ] Implement MCP server interface extensions
- [ ] Add npm package definitions for new servers
- [ ] Create Docker containers for server dependencies
- [ ] Implement lazy connection pooling
- [ ] Add health checks and monitoring
- [ ] Create unit tests (100% coverage)
- [ ] Create integration tests

---

## PHASE 3: Design & UI MCP Servers

### Status: NOT_STARTED

### 3.1 Figma Integration

| Server | Description | Requirements | Status |
|--------|-------------|--------------|--------|
| Cursor Talk to Figma | Read/modify designs | Figma plugin | NOT_STARTED |
| Framelink Figma | Fetch file data | API token | NOT_STARTED |
| Figma Chunking | Large file handling | API token | NOT_STARTED |
| Figma to React | Convert to components | API token | NOT_STARTED |

### 3.2 Adobe Integration

| Server | Description | Requirements | Status |
|--------|-------------|--------------|--------|
| Illustrator MCP | Adobe Illustrator | macOS + AI | NOT_STARTED |
| Photoshop MCP | Adobe Photoshop | Python API | NOT_STARTED |

### 3.3 Collaboration Tools

| Server | Description | Requirements | Status |
|--------|-------------|--------------|--------|
| MCP-Miro | Miro whiteboards | OAuth token | NOT_STARTED |

### 3.4 Implementation Tasks

- [ ] Research Figma API and authentication
- [ ] Create OAuth flow for Figma tokens
- [ ] Implement chunking for large files
- [ ] Create React component generator
- [ ] Add Adobe ExtendScript integration
- [ ] Create Miro OAuth client
- [ ] Implement rate limiting per provider
- [ ] Add caching for design assets

---

## PHASE 4: Image Generation MCP Servers

### Status: NOT_STARTED

### 4.1 Cloud-Based Generation

| Server | Model | Requirements | Status |
|--------|-------|--------------|--------|
| Replicate Flux | Flux.1 | Replicate API | NOT_STARTED |
| FLUX Generator | Black Forest Lab | BFL API key | NOT_STARTED |

### 4.2 Local Generation

| Server | Description | Requirements | Status |
|--------|-------------|--------------|--------|
| Stable Diffusion MCP | Local SD WebUI | GPU + SD | NOT_STARTED |
| ImageSorcery MCP | Image processing | Python 3.10+ | NOT_STARTED |

### 4.3 Asset Creation

| Server | Description | Requirements | Status |
|--------|-------------|--------------|--------|
| SVGMaker MCP | SVG generation | API key | NOT_STARTED |

### 4.4 Implementation Tasks

- [ ] Create Replicate API client
- [ ] Create Black Forest Lab API client
- [ ] Integrate with local SD WebUI
- [ ] Implement image processing pipeline
- [ ] Add SVG generation and editing
- [ ] Create image caching system
- [ ] Add format conversion utilities
- [ ] Implement GPU detection and fallback

---

## PHASE 5: LSP Server Integration

### Status: NOT_STARTED

### 5.1 AI-Specific LSP Servers

| Server | Description | Features | Status |
|--------|-------------|----------|--------|
| LSP-AI | AI-powered LSP | llama.cpp, Ollama, OpenAI | NOT_STARTED |
| OpenCode LSP | Auto-load LSPs | 75+ providers | NOT_STARTED |

### 5.2 Language-Specific LSP Servers

| Language | Server | Status |
|----------|--------|--------|
| Python | pyright, pylsp | pylsp EXISTING |
| JavaScript/TypeScript | typescript-language-server | EXISTING |
| C/C++ | clangd, ccls | NOT_STARTED |
| Rust | rust-analyzer | EXISTING |
| Go | gopls | EXISTING |
| Java | eclipse-jdt-ls | NOT_STARTED |
| C# | omnisharp-roslyn | NOT_STARTED |
| PHP | phpactor | NOT_STARTED |
| Ruby | solargraph | NOT_STARTED |
| Elixir | elixir-ls | NOT_STARTED |
| Haskell | haskell-language-server | NOT_STARTED |
| Shell | bash-language-server | NOT_STARTED |
| Dockerfile | docker-language-server | NOT_STARTED |
| YAML | yaml-language-server | NOT_STARTED |
| XML | lemminx | NOT_STARTED |
| Terraform | terraform-ls | NOT_STARTED |
| Lua | sumneko-lua-language-server | NOT_STARTED |

### 5.3 MCP-LSP Bridge Servers

| Server | Description | Status |
|--------|-------------|--------|
| LSP Tools MCP | Regex search tools | NOT_STARTED |
| Neovim LSP MCP | Neovim bridge | NOT_STARTED |
| mcp-language-server | Full LSP bridge | NOT_STARTED |
| lsp-mcp | Semantic analysis | NOT_STARTED |

### 5.4 Implementation Tasks

- [ ] Create LSP server registry similar to MCP
- [ ] Implement binary detection for all servers
- [ ] Add Docker containers for language runtimes
- [ ] Create LSP-AI integration layer
- [ ] Implement MCP-LSP bridge protocol
- [ ] Add workspace management
- [ ] Create language detection service
- [ ] Implement file watcher integration

---

## PHASE 6: RAG & Vector Database Integration

### Status: NOT_STARTED

### 6.1 Vector Databases

| Database | Type | Features | Status |
|----------|------|----------|--------|
| PostgreSQL pgvector | Embedded | ACID, SQL | EXISTING |
| ChromaDB | Cloud/Local | Full-text + vector | PARTIAL |
| Qdrant | Cloud/Local | Filtered search | EXISTING |
| Weaviate | Cloud/Local | Hybrid search | NOT_STARTED |
| Pinecone | Cloud | Managed, scale | NOT_STARTED |
| MongoDB Atlas | Cloud | Vector + document | NOT_STARTED |
| FAISS | Local | CPU/GPU optimized | NOT_STARTED |

### 6.2 Managed RAG Services

| Service | Type | Features | Status |
|---------|------|----------|--------|
| Ragie | Full RAG | Real-time, citations | NOT_STARTED |
| Pinecone Assistant | Full RAG | Document Q&A | NOT_STARTED |
| CustomGPT.ai | Full RAG | Enterprise | NOT_STARTED |
| Vectara | Full RAG | All-in-one | NOT_STARTED |
| Cohere RAG | LLM+RAG | Document array | NOT_STARTED |

### 6.3 RAG Frameworks

| Framework | Features | Status |
|-----------|----------|--------|
| LlamaIndex | Query enhancement | EXISTING |
| LangChain | Task decomposition | EXISTING |
| RAGatouille | ColBERT retrieval | NOT_STARTED |

### 6.4 Implementation Tasks

- [ ] Create unified VectorDB interface
- [ ] Implement Weaviate client
- [ ] Integrate Pinecone client
- [ ] Add MongoDB Atlas Vector Search
- [ ] Implement FAISS for local search
- [ ] Create managed RAG service adapters
- [ ] Add RAGatouille ColBERT integration
- [ ] Implement hybrid search (vector + keyword)

---

## PHASE 7: Embedding Model Integration

### Status: NOT_STARTED

### 7.1 Embedding Models

| Model | Provider | Dimensions | License | Status |
|-------|----------|------------|---------|--------|
| text-embedding-3-small | OpenAI | 1536 | Proprietary | EXISTING |
| text-embedding-3-large | OpenAI | 3072 | Proprietary | EXISTING |
| Qwen3-Embedding-0.6B | Qwen | 32-1024 | Apache 2.0 | NOT_STARTED |
| EmbeddingGemma-300M | Google | Variable | Apache 2.0 | NOT_STARTED |
| Jina Embeddings v4 | Jina | Variable | CC-BY-NC | NOT_STARTED |
| BGE-M3 | BAAI | 8192 tokens | MIT | NOT_STARTED |
| all-mpnet-base-v2 | Sentence-Transformers | 768 | MIT | NOT_STARTED |
| gte-multilingual-base | Alibaba | Variable | MIT | NOT_STARTED |
| Nomic Embed Text V2 | Nomic | Matryoshka | Apache 2.0 | NOT_STARTED |

### 7.2 Embedding Runtime Options

| Runtime | Description | Status |
|---------|-------------|--------|
| sentence-transformers | Python library | NOT_STARTED |
| Ollama embeddings | Local via Ollama | NOT_STARTED |
| HuggingFace Inference | Cloud API | NOT_STARTED |
| Local ONNX | ONNX runtime | NOT_STARTED |

### 7.3 Implementation Tasks

- [ ] Create EmbeddingModelRegistry
- [ ] Implement sentence-transformers integration
- [ ] Add Ollama embedding support
- [ ] Integrate HuggingFace Inference API
- [ ] Add ONNX model loading
- [ ] Implement model caching
- [ ] Create dimension normalization utilities
- [ ] Add batch processing optimization

---

## PHASE 8: Advanced RAG Techniques

### Status: NOT_STARTED

### 8.1 Query Enhancement

| Technique | Description | Status |
|-----------|-------------|--------|
| HyDE | Hypothetical documents | EXISTING (LlamaIndex) |
| Multi-Query | Query expansion | NOT_STARTED |
| Step-Back Prompting | Abstraction | EXISTING (LlamaIndex) |
| Query Decomposition | Sub-queries | EXISTING (LlamaIndex) |
| Query Fusion (RRF) | Rank fusion | EXISTING (LlamaIndex) |

### 8.2 Retrieval Enhancement

| Technique | Description | Status |
|-----------|-------------|--------|
| LLM Reranking | Cross-encoder | EXISTING (LlamaIndex) |
| ColBERT Reranking | Token-level | NOT_STARTED |
| Contextual Compression | Reduce context | NOT_STARTED |
| Parent Document Retrieval | Hierarchy | NOT_STARTED |
| Self-Query Retrieval | Auto-filter | NOT_STARTED |

### 8.3 Graph-Based Retrieval

| Technique | Description | Status |
|-----------|-------------|--------|
| Knowledge Graph RAG | Entity traversal | EXISTING (Cognee) |
| GraphRAG | Microsoft approach | NOT_STARTED |
| Entity Extraction | NER + linking | PARTIAL (Cognee) |

### 8.4 Implementation Tasks

- [ ] Implement Multi-Query expansion
- [ ] Add ColBERT reranking via RAGatouille
- [ ] Create contextual compression layer
- [ ] Implement parent document retrieval
- [ ] Add self-query retrieval with metadata
- [ ] Integrate Microsoft GraphRAG
- [ ] Enhance entity extraction pipeline
- [ ] Create RAG technique selector (auto-choose)

---

## PHASE 9: Testing & Challenges

### Status: NOT_STARTED

### 9.1 Test Coverage Requirements

| Test Type | Target Coverage | Current |
|-----------|-----------------|---------|
| Unit Tests | 100% | TBD |
| Integration Tests | 100% | TBD |
| E2E Tests | 100% | TBD |
| Security Tests | 100% | TBD |
| Stress Tests | 100% | TBD |
| Chaos Tests | 100% | TBD |
| Benchmark Tests | 100% | TBD |
| Race Detection | 100% | TBD |

### 9.2 Challenge Scripts

| Challenge | Tests | Description | Status |
|-----------|-------|-------------|--------|
| mcp_core_challenge.sh | 20+ | Core MCP servers | NOT_STARTED |
| mcp_design_challenge.sh | 15+ | Design MCPs | NOT_STARTED |
| mcp_image_challenge.sh | 15+ | Image MCPs | NOT_STARTED |
| lsp_core_challenge.sh | 25+ | Core LSPs | NOT_STARTED |
| lsp_ai_challenge.sh | 20+ | AI LSP features | NOT_STARTED |
| rag_vector_challenge.sh | 25+ | Vector DBs | NOT_STARTED |
| rag_embedding_challenge.sh | 20+ | Embeddings | NOT_STARTED |
| rag_advanced_challenge.sh | 30+ | Advanced RAG | NOT_STARTED |
| integration_full_challenge.sh | 50+ | Full system | NOT_STARTED |

### 9.3 Implementation Tasks

- [ ] Create test infrastructure
- [ ] Implement mock servers for testing
- [ ] Add test fixtures and data
- [ ] Create challenge script framework
- [ ] Implement CI/CD integration
- [ ] Add coverage reporting
- [ ] Create performance benchmarks
- [ ] Implement chaos testing scenarios

---

## PHASE 10: Documentation & Optimization

### Status: NOT_STARTED

### 10.1 Documentation

| Document | Purpose | Status |
|----------|---------|--------|
| MCP_INTEGRATION.md | MCP server guide | NOT_STARTED |
| LSP_INTEGRATION.md | LSP server guide | NOT_STARTED |
| RAG_INTEGRATION.md | RAG system guide | NOT_STARTED |
| EMBEDDING_MODELS.md | Embedding guide | NOT_STARTED |
| API_REFERENCE.md | API documentation | NOT_STARTED |
| DOCKER_SETUP.md | Container setup | NOT_STARTED |

### 10.2 Performance Optimization

| Area | Optimization | Status |
|------|--------------|--------|
| Connection Pooling | Enhanced lazy init | NOT_STARTED |
| Caching | Multi-tier caching | NOT_STARTED |
| Batch Processing | Optimized batching | NOT_STARTED |
| Memory Management | Resource limits | NOT_STARTED |
| GPU Utilization | CUDA optimization | NOT_STARTED |

### 10.3 Security Enhancements

| Area | Enhancement | Status |
|------|-------------|--------|
| API Key Management | Secure vault | NOT_STARTED |
| OAuth Token Refresh | Auto-refresh | NOT_STARTED |
| Rate Limiting | Per-provider limits | NOT_STARTED |
| Input Validation | Schema validation | NOT_STARTED |
| Audit Logging | Security events | NOT_STARTED |

---

## Progress Tracking

### Overall Progress
```
Phase 1:  [..........] 0%
Phase 2:  [..........] 0%
Phase 3:  [..........] 0%
Phase 4:  [..........] 0%
Phase 5:  [..........] 0%
Phase 6:  [..........] 0%
Phase 7:  [..........] 0%
Phase 8:  [..........] 0%
Phase 9:  [..........] 0%
Phase 10: [..........] 0%
----------------------------
TOTAL:    [..........] 0%
```

### Daily Log

#### 2026-01-19
- Created master integration plan
- Analyzed existing infrastructure
- Documented current MCP servers (6)
- Documented current LSP servers (4)
- Documented current RAG infrastructure

---

## Dependencies

### External Dependencies
- Node.js 20+ (for MCP npm packages)
- Python 3.10+ (for embedding models)
- Docker/Podman (for containers)
- NVIDIA CUDA (optional, for GPU)
- Various API keys (Figma, Replicate, etc.)

### Internal Dependencies
- LLMsVerifier submodule
- Toolkit submodule
- Cognee integration
- PostgreSQL with pgvector

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| API rate limits | HIGH | Implement caching, backoff |
| GPU unavailable | MEDIUM | Fallback to CPU/Cloud |
| Breaking changes | HIGH | Pin versions, CI testing |
| Performance degradation | HIGH | Benchmarks, monitoring |
| Security vulnerabilities | CRITICAL | Security scans, audits |

---

## How to Resume Work

1. Check this document for current phase status
2. Find the first `NOT_STARTED` item in the current phase
3. Update status to `IN_PROGRESS` before starting
4. Complete implementation with full test coverage
5. Update status to `COMPLETED` when done
6. Create/update relevant challenge script
7. Update progress tracking section
8. Commit changes with descriptive message

---

## Files to Track Changes

- `docs/implementation/INTEGRATION_MASTER_PLAN.md` (this file)
- `docs/implementation/progress/` (daily progress logs)
- `internal/mcp/servers/` (new MCP servers)
- `internal/lsp/servers/` (new LSP servers)
- `internal/rag/` (RAG implementations)
- `internal/embeddings/models/` (embedding models)
- `challenges/scripts/` (challenge scripts)
- `tests/` (test files)
