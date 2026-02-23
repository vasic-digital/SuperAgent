# HelixAgent Architecture

**Version:** 1.2.0  
**Last Updated:** 2026-02-23

## Overview

HelixAgent is an AI-powered ensemble LLM service written in Go that combines responses from multiple language models using intelligent aggregation strategies.

## Core Architecture

### Entry Points

| Component | Location | Description |
|-----------|----------|-------------|
| Main Application | `cmd/helixagent/` | Primary server binary |
| API Server | `cmd/api/` | REST API server |
| gRPC Server | `cmd/grpc-server/` | gRPC protocol server |
| Cognee Mock | `cmd/cognee-mock/` | Mock Cognee service |
| Sanity Check | `cmd/sanity-check/` | System validation tool |
| MCP Bridge | `cmd/mcp-bridge/` | MCP protocol bridge |
| Constitution Generator | `cmd/generate-constitution/` | Constitution file generator |

### Internal Packages

```
internal/
├── adapters/          # Bridge layer to extracted modules
├── agents/            # CLI agent registry (48 agents)
├── background/        # Task queue, worker pool, resource monitor
├── bigdata/           # Infinite context, distributed memory
├── cache/             # Redis + in-memory caching
├── concurrency/       # Worker pools, semaphores, rate limiters
├── database/          # PostgreSQL/pgx repositories
├── debate/            # AI debate orchestrator framework
├── embedding/         # 6 embedding providers
├── formatters/        # 32+ code formatters
├── handlers/          # HTTP handlers
├── llm/               # LLM provider implementations
├── mcp/               # MCP adapters (45+)
├── memory/            # Mem0-style with entity graphs
├── messaging/         # Kafka + RabbitMQ abstraction
├── middleware/        # Auth, rate limiting, CORS
├── models/            # Data models and enums
├── notifications/     # SSE, WebSocket, Webhooks
├── observability/     # OpenTelemetry, Jaeger, Zipkin
├── optimization/      # GPT-Cache, Outlines, SGLang
├── plugins/           # Hot-reloadable plugin system
├── rag/               # Hybrid retrieval
├── routing/           # Semantic routing
├── security/          # Red team framework, guardrails
├── services/          # Business logic
├── streaming/         # SSE, WebSocket, gRPC streaming
├── tools/             # Tool schema registry (21 tools)
├── vectordb/          # Qdrant, Pinecone, Milvus, pgvector
└── verifier/          # Startup verification orchestrator
```

## Key Components

### LLM Provider Registry

- **22 dedicated providers**: Claude, Chutes, DeepSeek, Gemini, Mistral, OpenRouter, Qwen, ZAI, Zen, Cerebras, Ollama, AI21, Anthropic, Cohere, Fireworks, Groq, HuggingFace, OpenAI, Perplexity, Replicate, Together, xAI
- **Generic OpenAI-compatible**: 17+ additional providers
- **Dynamic model discovery**: 3-tier (Provider API → models.dev → fallback)

### AI Debate System

- **5 positions × 5 LLMs** = 25 total LLMs
- **Multi-pass validation**: Initial → Validation → Polish → Final
- **Orchestrator**: Multi-topology (mesh/star/chain), phase protocol

### Extracted Modules

20 independent modules with zero shared dependencies:

| Phase | Modules |
|-------|---------|
| Foundation | EventBus, Concurrency, Observability, Auth, Storage, Streaming |
| Infrastructure | Security, VectorDB, Embeddings, Database, Cache |
| Services | Messaging, Formatters, MCP |
| Integration | RAG, Memory, Optimization, Plugins |
| Pre-existing | Containers, Challenges |

## Data Flow

```
Request → Handler → Middleware → Service Layer
    ↓
Provider Registry → LLM Provider (22+) → Response
    ↓
Ensemble Strategy → Debate System → Aggregation
    ↓
Cache → Database → Response
```

## Deployment

- **Container Runtime**: Docker / Podman / Kubernetes
- **Build**: All release builds in containers for reproducibility
- **Configuration**: YAML files + environment variables

## See Also

- [MODULES.md](MODULES.md) - Detailed module documentation
- [CLAUDE.md](../CLAUDE.md) - Development guidelines
- [AGENTS.md](../AGENTS.md) - Agent configuration
