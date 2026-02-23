# HelixAgent Architecture

A deep dive into HelixAgent's system architecture for technical architects and senior engineers.

---

## Executive Summary

HelixAgent is an AI-powered ensemble LLM service written in Go that acts as a **Virtual LLM Provider**. It exposes a single unified model (`helixagent/helixagent-debate`) that internally orchestrates multiple language models through AI debate to produce consensus-driven responses.

**Key Architectural Principles:**
- Virtual Provider abstraction
- Real-time provider verification
- Consensus through AI debate
- Graceful degradation
- Protocol-first design

---

## System Overview

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                              Client Applications                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ OpenAI SDK   │  │   cURL/HTTP  │  │  gRPC Client │  │  AI CLI Tool │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼─────────────────┼─────────────────┼─────────────────┼──────────────┘
          │                 │                 │                 │
          └─────────────────┴─────────────────┴─────────────────┘
                                      │
                           ┌──────────▼──────────┐
                           │      API Gateway     │
                           │   (Gin HTTP + gRPC)  │
                           │  - Authentication    │
                           │  - Rate Limiting     │
                           │  - Request Routing   │
                           └──────────┬──────────┘
                                      │
          ┌───────────────────────────┼───────────────────────────┐
          │                           │                           │
┌─────────▼─────────┐    ┌───────────▼───────────┐    ┌──────────▼──────────┐
│   REST Handlers   │    │    gRPC Services      │    │  Protocol Handlers  │
│  /v1/chat/...     │    │  LLMFacade service    │    │  MCP, LSP, ACP      │
└─────────┬─────────┘    └───────────┬───────────┘    └──────────┬──────────┘
          │                          │                           │
          └──────────────────────────┼───────────────────────────┘
                                     │
                          ┌──────────▼──────────┐
                          │   Ensemble Service   │
                          │  - AI Debate Engine  │
                          │  - Consensus Voting  │
                          │  - Response Merge    │
                          └──────────┬──────────┘
                                     │
          ┌──────────────────────────┼──────────────────────────┐
          │                          │                          │
┌─────────▼─────────┐    ┌──────────▼──────────┐    ┌──────────▼──────────┐
│ Provider Registry │    │   Debate Service    │    │   Memory Manager    │
│ - Provider Store  │    │  - Team Selection   │    │  - Short-term       │
│ - Health Checks   │    │  - Round Execution  │    │  - Long-term        │
│ - Scoring         │    │  - Validation       │    │  - Entity Graph     │
└─────────┬─────────┘    └──────────┬──────────┘    └──────────┬──────────┘
          │                         │                          │
          │              ┌──────────▼──────────┐               │
          │              │  LLMsVerifier       │               │
          │              │  (Submodule)        │               │
          │              │  - Real API Tests   │               │
          │              │  - Score Generation │               │
          │              └──────────┬──────────┘               │
          │                         │                          │
          └─────────────────────────┼──────────────────────────┘
                                    │
     ┌───────────────────┬──────────┼──────────┬───────────────────┐
     │                   │          │          │                   │
┌────▼────┐  ┌───────────▼───┐  ┌───▼───┐  ┌───▼───┐  ┌───────────▼───┐
│ Claude  │  │   DeepSeek    │  │Gemini │  │Mistral│  │  OpenRouter   │
└─────────┘  └───────────────┘  └───────┘  └───────┘  └───────────────┘
                    LLM Providers (10+)
```

---

## Core Components

### Entry Points

| Component | Path | Purpose |
|-----------|------|---------|
| Main Application | `cmd/helixagent/` | Service entry point |
| API Server | `cmd/api/` | HTTP API server |
| gRPC Server | `cmd/grpc-server/` | gRPC service |

### Service Layer (`internal/services/`)

| Service | Responsibility |
|---------|----------------|
| `provider_registry` | Provider lifecycle, health checks, scoring |
| `ensemble` | Ensemble orchestration, parallel execution |
| `debate_service` | AI debate coordination |
| `debate_team_config` | Dynamic team selection |
| `llm_intent_classifier` | Semantic intent detection |
| `context_manager` | Conversation context |
| `mcp_client` | MCP protocol handling |
| `lsp_manager` | LSP server management |
| `plugin_system` | Plugin lifecycle |
| `boot_manager` | Service startup orchestration |
| `health_checker` | Health monitoring |

### Provider Layer (`internal/llm/providers/`)

Each provider implements the `LLMProvider` interface:

```go
type LLMProvider interface {
    Complete(ctx context.Context, req *Request) (*Response, error)
    CompleteStream(ctx context.Context, req *Request) (<-chan *StreamChunk, error)
    HealthCheck(ctx context.Context) error
    GetCapabilities() *Capabilities
    ValidateConfig() error
}
```

**Supported Providers:**
- `claude/` - Anthropic Claude (with CLI proxy support)
- `deepseek/` - DeepSeek Chat and Coder
- `gemini/` - Google Gemini
- `mistral/` - Mistral AI
- `openrouter/` - OpenRouter aggregator
- `qwen/` - Alibaba Qwen (with ACP support)
- `zai/` - ZAI models
- `zen/` - Zen/OpenCode (with HTTP server mode)
- `cerebras/` - Cerebras inference
- `ollama/` - Local Ollama models

---

## Architectural Patterns

### Virtual LLM Provider Pattern

HelixAgent presents a single unified interface while internally managing multiple providers:

```
┌─────────────────────────────────────────────────────┐
│                   Client View                        │
│   Model: helixagent/helixagent-debate               │
│   Endpoint: /v1/chat/completions                    │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                  Internal View                       │
│   Primary 1: Claude (Score: 8.5)                    │
│     └── Fallback: DeepSeek                          │
│   Primary 2: Gemini (Score: 8.2)                    │
│     └── Fallback: Mistral                           │
│   Primary 3: DeepSeek (Score: 8.0)                  │
│   Primary 4: Mistral (Score: 7.8)                   │
│   Primary 5: OpenRouter/Llama (Score: 7.5)          │
└─────────────────────────────────────────────────────┘
```

### AI Debate Orchestration

Multi-round debate with validation passes:

```
┌──────────────────────────────────────────────────────────────────────┐
│                        AI Debate Flow                                 │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  User Prompt ─────▶ [INITIAL ROUND]                                  │
│                      │                                               │
│                      ├─▶ LLM 1 Response                              │
│                      ├─▶ LLM 2 Response                              │
│                      ├─▶ LLM 3 Response                              │
│                      ├─▶ LLM 4 Response                              │
│                      └─▶ LLM 5 Response                              │
│                             │                                        │
│                             ▼                                        │
│                    [VALIDATION ROUND]                                │
│                      │                                               │
│                      ├─▶ Cross-review responses                      │
│                      └─▶ Identify disagreements                      │
│                             │                                        │
│                             ▼                                        │
│                    [POLISH ROUND]                                    │
│                      │                                               │
│                      └─▶ Refine based on feedback                    │
│                             │                                        │
│                             ▼                                        │
│                    [FINAL CONSENSUS]                                 │
│                      │                                               │
│                      ├─▶ Confidence-weighted voting                  │
│                      └─▶ Response synthesis                          │
│                             │                                        │
│                             ▼                                        │
│                    Final Response ────▶ User                         │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### Debate Orchestrator Patterns

| Topology | Description | Use Case |
|----------|-------------|----------|
| **Mesh** | All participants debate each other | Complex topics |
| **Star** | Central coordinator with spokes | Quick consensus |
| **Chain** | Sequential refinement | Deep analysis |

| Phase | Purpose |
|-------|---------|
| Proposal | Initial response generation |
| Critique | Cross-review and feedback |
| Review | Incorporate feedback |
| Synthesis | Final consensus building |

### Circuit Breaker Pattern

Fault tolerance for provider failures:

```
┌─────────────────────────────────────────────────────┐
│                 Circuit Breaker                      │
├─────────────────────────────────────────────────────┤
│                                                     │
│   CLOSED ──────▶ OPEN ──────▶ HALF-OPEN            │
│     │              │              │                 │
│     │              │              │                 │
│   Normal        Fail Fast      Test               │
│   Operation     (no calls)     Recovery            │
│     │              │              │                 │
│     │              │              │                 │
│     ▼              ▼              ▼                 │
│   Success:      Timeout:       Success:            │
│   Stay CLOSED   Stay OPEN      Go CLOSED           │
│                                                     │
│   Failure:      Failure:                           │
│   Count++       N/A            Go OPEN             │
│   If threshold                                     │
│   → Go OPEN                                        │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Semantic Intent Detection

LLM-based classification with pattern fallback:

```
User Input
    │
    ▼
┌───────────────────┐
│ LLM Classifier    │ ──▶ Classified Intent
│ (Primary)         │
└─────────┬─────────┘
          │
          │ (On Failure)
          ▼
┌───────────────────┐
│ Pattern Matcher   │ ──▶ Classified Intent
│ (Fallback)        │
└───────────────────┘
```

---

## Data Flow

### Request Processing Flow

```
1. Client Request (HTTP/gRPC)
        │
        ▼
2. Authentication Middleware
   - API Key validation
   - JWT verification
   - Rate limit check
        │
        ▼
3. Request Handler
   - Parse request
   - Validate parameters
   - Build internal request
        │
        ▼
4. Ensemble Service
   - Select debate team
   - Parallel provider calls
   - Collect responses
        │
        ▼
5. Debate Engine
   - Multi-round debate
   - Validation pass
   - Consensus voting
        │
        ▼
6. Response Handler
   - Format response
   - Stream if requested
   - Add metadata
        │
        ▼
7. Client Response
```

### Startup Verification Flow

```
1. Application Start
        │
        ▼
2. Load Configuration
   - Environment variables
   - YAML config files
   - Secret management
        │
        ▼
3. Initialize Infrastructure
   - Database connections
   - Redis cache
   - Vector stores
        │
        ▼
4. LLMsVerifier Startup
   - Discover providers
   - Verify with real API calls
   - Score providers (8-test pipeline)
        │
        ▼
5. Provider Ranking
   - Apply scoring weights
   - Select top 5 primaries
   - Assign fallbacks
        │
        ▼
6. Debate Team Formation
   - Configure debate group
   - Initialize debate service
        │
        ▼
7. Health Checks
   - Verify all services
   - Report status
        │
        ▼
8. Ready to Serve
```

---

## Submodules Architecture

HelixAgent uses a modular architecture with 25+ extracted modules:

### Foundation Modules (Phase 1)

| Module | Package | Purpose |
|--------|---------|---------|
| EventBus | `digital.vasic.eventbus` | Pub/sub messaging |
| Concurrency | `digital.vasic.concurrency` | Worker pools, rate limiters |
| Observability | `digital.vasic.observability` | Tracing, metrics, logging |
| Auth | `digital.vasic.auth` | JWT, API key, OAuth |
| Storage | `digital.vasic.storage` | S3, MinIO, filesystem |
| Streaming | `digital.vasic.streaming` | SSE, WebSocket, gRPC |

### Infrastructure Modules (Phase 2)

| Module | Package | Purpose |
|--------|---------|---------|
| Security | `digital.vasic.security` | Guardrails, PII detection |
| VectorDB | `digital.vasic.vectordb` | Qdrant, Pinecone, Milvus |
| Embeddings | `digital.vasic.embeddings` | OpenAI, Cohere, Voyage |
| Database | `digital.vasic.database` | PostgreSQL, SQLite |
| Cache | `digital.vasic.cache` | Redis, in-memory |

### Service Modules (Phase 3)

| Module | Package | Purpose |
|--------|---------|---------|
| Messaging | `digital.vasic.messaging` | Kafka, RabbitMQ |
| Formatters | `digital.vasic.formatters` | Code formatting |
| MCP | `digital.vasic.mcp` | Model Context Protocol |

### Integration Modules (Phase 4)

| Module | Package | Purpose |
|--------|---------|---------|
| RAG | `digital.vasic.rag` | Retrieval-augmented generation |
| Memory | `digital.vasic.memory` | Mem0-style memory |
| Optimization | `digital.vasic.optimization` | GPT-Cache, streaming |
| Plugins | `digital.vasic.plugins` | Plugin system |

### AI/ML Modules (Phase 5)

| Module | Package | Purpose |
|--------|---------|---------|
| Agentic | `digital.vasic.agentic` | Graph-based agentic workflow orchestration |
| LLMOps | `digital.vasic.llmops` | Evaluation pipelines, A/B experiments, prompt versioning |
| SelfImprove | `digital.vasic.selfimprove` | RLHF, reward modelling, feedback-driven optimization |
| Planning | `digital.vasic.planning` | HiPlan, MCTS, Tree of Thoughts planning algorithms |
| Benchmark | `digital.vasic.benchmark` | SWE-bench, HumanEval, MMLU benchmarking suites |

### Pre-existing Modules

| Module | Package | Purpose |
|--------|---------|---------|
| Containers | `digital.vasic.containers` | Docker/Podman/K8s |
| Challenges | `digital.vasic.challenges` | Testing framework |

---

## Technology Stack

### Core Technologies

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24+ | Primary language |
| Gin | 1.11.0 | HTTP framework |
| PostgreSQL | 15 | Primary database |
| Redis | 7 | Caching layer |
| gRPC | Latest | Binary protocol |

### Infrastructure

| Technology | Purpose |
|------------|---------|
| Docker/Podman | Containerization |
| Kubernetes | Orchestration |
| Kafka | Message streaming |
| ClickHouse | Analytics |
| Neo4j | Knowledge graph |

### Observability

| Technology | Purpose |
|------------|---------|
| Prometheus | Metrics |
| Grafana | Dashboards |
| OpenTelemetry | Distributed tracing |
| Jaeger | Trace visualization |
| Langfuse | LLM analytics |

---

## Deployment Architecture

### Single Instance

```
┌─────────────────────────────────────────┐
│              Host Machine                │
│  ┌─────────────────────────────────────┐ │
│  │          HelixAgent                  │ │
│  │  ┌─────────┐  ┌─────────┐          │ │
│  │  │HTTP:8080│  │gRPC:9090│          │ │
│  │  └─────────┘  └─────────┘          │ │
│  └─────────────────────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐       │
│  │ PostgreSQL  │  │    Redis    │       │
│  │   :5432     │  │    :6379    │       │
│  └─────────────┘  └─────────────┘       │
└─────────────────────────────────────────┘
```

### High Availability

```
┌────────────────────────────────────────────────────────────────────┐
│                         Load Balancer                               │
│                      (nginx/HAProxy/ALB)                           │
└──────────────────────────────┬─────────────────────────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
    ┌────▼────┐          ┌─────▼─────┐         ┌────▼────┐
    │ Helix 1 │          │  Helix 2  │         │ Helix 3 │
    │ (HTTP)  │          │  (HTTP)   │         │ (HTTP)  │
    └────┬────┘          └─────┬─────┘         └────┬────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
         ┌────▼────┐     ┌─────▼─────┐    ┌────▼────┐
         │PostgreSQL│     │   Redis   │    │  Kafka  │
         │ Cluster  │     │  Cluster  │    │ Cluster │
         └──────────┘     └───────────┘    └─────────┘
```

### Kubernetes Deployment

```yaml
# Simplified structure
Namespace: helixagent
├── Deployment: helixagent-api (replicas: 3)
├── Deployment: helixagent-grpc (replicas: 2)
├── Deployment: helixagent-worker (replicas: 5)
├── StatefulSet: postgresql (replicas: 3)
├── StatefulSet: redis (replicas: 3)
├── Service: helixagent-api (ClusterIP)
├── Service: helixagent-grpc (ClusterIP)
├── Ingress: helixagent-ingress
├── ConfigMap: helixagent-config
├── Secret: helixagent-secrets
├── HPA: helixagent-api-hpa
└── PDB: helixagent-pdb
```

---

## Security Architecture

### Authentication Flow

```
┌─────────────────────────────────────────────────────────┐
│                    Client Request                        │
│                Authorization: Bearer <token>             │
└────────────────────────────┬────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│                  Auth Middleware                         │
├─────────────────────────────────────────────────────────┤
│  1. Extract token from header                           │
│  2. Determine token type (JWT/API Key)                  │
│  3. Validate token                                      │
│  4. Extract claims/permissions                          │
│  5. Attach to context                                   │
└────────────────────────────┬────────────────────────────┘
                             │
                    ┌────────┴────────┐
                    │                 │
               JWT Token         API Key
                    │                 │
              ┌─────▼─────┐    ┌─────▼─────┐
              │ Verify    │    │  Lookup   │
              │ Signature │    │  in DB    │
              └─────┬─────┘    └─────┬─────┘
                    │                │
                    └────────┬───────┘
                             │
                             ▼
                    ┌────────────────┐
                    │ Request Handler│
                    │  (Authorized)  │
                    └────────────────┘
```

### Data Protection

| Layer | Protection |
|-------|------------|
| Transport | TLS 1.3 |
| API | Authentication, rate limiting |
| Data | Encryption at rest |
| Secrets | Vault/KMS integration |
| PII | Detection and redaction |

---

## Performance Characteristics

### Latency Profile

| Operation | P50 | P95 | P99 |
|-----------|-----|-----|-----|
| Health check | 5ms | 10ms | 20ms |
| Cache hit | 10ms | 50ms | 100ms |
| Single LLM call | 500ms | 2s | 5s |
| Full debate (5 LLMs) | 2s | 5s | 10s |

### Throughput

| Configuration | Requests/sec |
|---------------|--------------|
| Single instance | 100-500 |
| 3-node cluster | 300-1500 |
| 10-node cluster | 1000-5000 |

### Scaling Factors

- **Horizontal**: Add more instances
- **Caching**: Semantic cache for repeated queries
- **Async**: Background processing for non-blocking ops
- **Streaming**: Token-by-token response delivery

---

## Related Documentation

- [CLAUDE.md](/CLAUDE.md) - Technical reference
- [API Documentation](/docs/api/README.md) - API reference
- [Deployment Guide](/docs/deployment/README.md) - Deployment instructions
- [Security Guide](/docs/website/SECURITY.md) - Security details

---

**Last Updated**: February 2026
**Version**: 1.0.0
