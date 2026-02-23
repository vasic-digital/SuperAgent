# HelixAgent Architecture

## System Overview

HelixAgent is a multi-provider LLM orchestration platform that provides unified access to multiple AI providers with intelligent ensemble capabilities and advanced tooling support.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        HelixAgent                               │
│                    LLM Orchestration Platform                   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   REST API  │  │   gRPC API  │  │  WebSocket  │             │
│  │   (Gin)     │  │             │  │             │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Auth & Rate │  │  Request    │  │  Response   │             │
│  │   Limiting  │  │ Validation  │  │ Processing │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                    Core Services Layer                          │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │ Provider    │  │ Ensemble    │  │ Context     │  │ Memory  │ │
│  │ Registry    │  │ Service     │  │ Manager     │  │ Service │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │ MCP Manager │  │ LSP Client  │  │ Tool        │  │ Security │ │
│  │             │  │             │  │ Registry    │  │ Sandbox  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
│                                                                 │
│  ┌─────────────┐                                                │
│  │ Integration │                                                │
│  │ Orchestrator│                                                │
│  └─────────────┘                                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                   Provider Layer                                │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │  DeepSeek   │  │    Qwen     │  │ OpenRouter  │  │ Claude  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐                               │
│  │   Gemini    │  │   Ollama    │                               │
│  └─────────────┘  └─────────────┘                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                   Infrastructure Layer                          │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │ PostgreSQL  │  │    Redis    │  │ Prometheus  │  │ Grafana │ │
│  │  (Primary)  │  │  (Cache)    │  │ (Metrics)   │  │ (Dash)  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐                               │
│  │   Docker    │  │ Kubernetes  │                               │
│  │ (Container) │  │ (Orch)      │                               │
│  └─────────────┘  └─────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

## Entry Points

HelixAgent has multiple entry points for different use cases:

| Entry Point | Location | Purpose | Production Ready |
|-------------|----------|---------|------------------|
| **Main Server** | `cmd/helixagent/main.go` | Production server with full AI Debate ensemble, LLMsVerifier integration, all protocols | Yes |
| **gRPC Server** | `cmd/grpc-server/main.go` | High-performance gRPC API server | Yes |
| **Demo API** | `cmd/api/main.go` | **DEMO ONLY** - Returns mock/hardcoded responses for API exploration | No |

### Production Entry Point

```bash
# Run the main production server
go run cmd/helixagent/main.go

# Or build and run
make build
./helixagent
```

The main server (`cmd/helixagent/main.go`) provides:
- Full AI Debate ensemble with 25 LLMs
- LLMsVerifier startup verification pipeline
- All protocol support (MCP, LSP, ACP, Embeddings)
- Real provider integrations (Claude, DeepSeek, Gemini, etc.)
- Production-ready error handling and observability

### Demo Server (Not for Production)

The `cmd/api/main.go` server is a **demonstration only** implementation that:
- Returns hardcoded/mock responses
- Does NOT connect to real LLM backends
- Is useful for API structure exploration and client development
- Should NEVER be deployed in production

## Component Details

### API Layer
- **REST API (Gin)**: OpenAI-compatible endpoints for chat completions, model management
- **gRPC API**: High-performance internal communication
- **WebSocket**: Real-time streaming responses

### Middleware Layer
- **Authentication**: JWT-based user authentication
- **Rate Limiting**: Request throttling and abuse prevention
- **Request Validation**: Input sanitization and schema validation
- **Response Processing**: Output formatting and metadata injection

### Core Services

#### Provider Registry
- Manages LLM provider registration and configuration
- Health monitoring and failover
- Load balancing and routing

#### Ensemble Service
- Multi-provider response aggregation
- Confidence-weighted voting strategies
- Fallback mechanisms

#### Context Manager
- Multi-source context aggregation
- ML-based relevance scoring
- Context compression and optimization

#### Memory Service
- User session management
- Conversation history
- Caching layer

#### MCP Manager
- Model Context Protocol server management
- Tool discovery and registration
- Secure tool execution

#### LSP Client
- Language Server Protocol integration
- Code intelligence and analysis
- Multi-language support

#### Tool Registry
- Dynamic tool discovery
- Validation and security checks
- Dependency management

#### Security Sandbox
- Isolated execution environment
- Resource limits and monitoring
- Command validation

#### Integration Orchestrator
- Workflow orchestration
- Parallel processing
- Error handling and recovery

### Provider Layer
- **DeepSeek**: Chinese LLM provider
- **Qwen**: Alibaba's LLM series
- **OpenRouter**: Multi-provider marketplace
- **Claude**: Anthropic's advanced models
- **Gemini**: Google's multimodal models
- **Ollama**: Local model execution

### Infrastructure Layer
- **PostgreSQL**: Primary data storage
- **Redis**: Caching and session storage
- **Prometheus**: Metrics collection
- **Grafana**: Monitoring dashboards
- **Docker**: Containerization
- **Kubernetes**: Orchestration

## Data Flow

```
User Request → API Gateway → Authentication → Rate Limiting → Request Validation
    ↓
Provider Selection → Ensemble Configuration → Context Building
    ↓
Parallel Provider Calls → Response Collection → Voting/Scoring
    ↓
Response Processing → Context Update → Memory Storage
    ↓
Final Response → User
```

## Security Architecture

```
┌─────────────────────────────────────────────────┐
│              Security Layers                    │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────┐    │
│  │         Application Layer              │    │
│  │  ┌─────────────┐  ┌─────────────┐       │    │
│  │  │ Input       │  │ Output      │       │    │
│  │  │ Validation  │  │ Sanitization│       │    │
│  │  └─────────────┘  └─────────────┘       │    │
│  └─────────────────────────────────────────┘    │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────┐    │
│  │         Service Layer                   │    │
│  │  ┌─────────────┐  ┌─────────────┐       │    │
│  │  │ Auth & Auth │  │ Tool        │       │    │
│  │  │             │  │ Validation  │       │    │
│  │  └─────────────┘  └─────────────┘       │    │
│  └─────────────────────────────────────────┘    │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────┐    │
│  │         Infrastructure Layer            │    │
│  │  ┌─────────────┐  ┌─────────────┐       │    │
│  │  │ Network     │  │ Container   │       │    │
│  │  │ Security    │  │ Isolation   │       │    │
│  │  └─────────────┘  └─────────────┘       │    │
│  └─────────────────────────────────────────┘    │
└─────────────────────────────────────────────────┘
```

## Deployment Architecture

```
┌─────────────────────────────────────────────────┐
│              Production Deployment             │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │   Load      │  │  HelixAgent │  │  Load   │ │
│  │  Balancer   │  │   API       │  │  Balancer│ │
│  │  (Nginx)    │  │   Servers   │  │  (Nginx) │ │
│  └─────────────┘  └─────────────┘  └─────────┘ │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │ PostgreSQL  │  │    Redis    │  │  Redis  │ │
│  │  Cluster    │  │  Cluster    │  │  Cache  │ │
│  └─────────────┘  └─────────────┘  └─────────┘ │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │ Monitoring  │  │  Logging    │  │ Backup  │ │
│  │  Stack      │  │  Stack      │  │  System │ │
│  └─────────────┘  └─────────────┘  └─────────┘ │
└─────────────────────────────────────────────────┘
```

## Extracted Modules Layer

HelixAgent's functionality is decomposed into **25 independent Go modules** organized by phase. Each
module is a separate Go project with its own `go.mod`, tests, and documentation, integrated via git
submodules and `replace` directives. Bridge adapters in `internal/adapters/` connect internal types
to each module. Full catalog: [`docs/MODULES.md`](../MODULES.md).

```
HelixAgent (dev.helix.agent)
├── Foundation Layer (zero dependencies)
│   ├── EventBus         (digital.vasic.eventbus)      — Pub/Sub event system
│   ├── Concurrency      (digital.vasic.concurrency)   — Pools, limiters, breakers
│   ├── Observability    (digital.vasic.observability) — Tracing, metrics, logging
│   ├── Auth             (digital.vasic.auth)          — JWT, API key, OAuth
│   ├── Storage          (digital.vasic.storage)       — S3, local filesystem
│   └── Streaming        (digital.vasic.streaming)     — SSE, WS, gRPC, webhooks
├── Infrastructure Layer
│   ├── Security         (digital.vasic.security)      — Guardrails, PII, policies
│   ├── VectorDB         (digital.vasic.vectordb)      — Qdrant, Pinecone, Milvus
│   ├── Embeddings       (digital.vasic.embeddings)    — 6 embedding providers
│   ├── Database         (digital.vasic.database)      — PostgreSQL, SQLite
│   └── Cache            (digital.vasic.cache)         — Redis, in-memory
├── Services Layer
│   ├── Messaging        (digital.vasic.messaging)     — Kafka, RabbitMQ
│   ├── Formatters       (digital.vasic.formatters)    — 32+ code formatters
│   └── MCP              (digital.vasic.mcp)           — Model Context Protocol
├── Integration Layer
│   ├── RAG              (digital.vasic.rag)           — Retrieval-Augmented Generation
│   ├── Memory           (digital.vasic.memory)        — Mem0-style memory
│   ├── Optimization     (digital.vasic.optimization)  — GPT-Cache, prompt optimization
│   └── Plugins          (digital.vasic.plugins)       — Plugin system
├── AI/ML Layer (Phase 5)
│   ├── Agentic          (digital.vasic.agentic)       — Graph-based workflow orchestration
│   ├── LLMOps          (digital.vasic.llmops)         — Evaluation, experiments, datasets
│   ├── SelfImprove      (digital.vasic.selfimprove)   — Reward model, RLHF
│   ├── Planning         (digital.vasic.planning)      — HiPlan, MCTS, Tree of Thoughts
│   └── Benchmark        (digital.vasic.benchmark)     — LLM benchmarking, leaderboards
└── Pre-existing
    ├── Containers       (digital.vasic.containers)    — Container orchestration
    └── Challenges       (digital.vasic.challenges)    — Challenge framework
```

### AI/ML Phase 5 Modules

Five modules added to support advanced AI/ML capabilities beyond inference:

- **Agentic** (`Agentic/`, `digital.vasic.agentic`): Graph-based agentic workflow orchestration.
  Multi-step workflow execution with conditional branching, parallel nodes, state management, and
  retry logic. Used for complex AI task pipelines. Adapter: `internal/adapters/agentic/adapter.go`.

- **LLMOps** (`LLMOps/`, `digital.vasic.llmops`): LLM operations and observability framework.
  Continuous evaluation pipelines, A/B experiment management with statistical significance testing,
  dataset management (golden sets, synthetic data), and prompt versioning.
  Adapter: `internal/adapters/llmops/adapter.go`.

- **SelfImprove** (`SelfImprove/`, `digital.vasic.selfimprove`): AI self-improvement via
  RLHF-style feedback collection, reward model training with dimension-weighted scoring, and an
  optimizer that adjusts model parameters based on accumulated feedback signals.
  Adapter: `internal/adapters/selfimprove/adapter.go`.

- **Planning** (`Planning/`, `digital.vasic.planning`): AI planning algorithms for complex task
  decomposition. Hierarchical planning (HiPlan with milestone-based decomposition), Monte Carlo
  Tree Search (MCTS for code action optimization), and Tree of Thoughts (multi-path reasoning).
  Adapter: `internal/adapters/planning/adapter.go`.

- **Benchmark** (`Benchmark/`, `digital.vasic.benchmark`): LLM benchmarking and evaluation
  framework. Industry-standard benchmarks (SWE-bench, HumanEval, MBPP, LMSYS, HellaSwag, MMLU,
  GSM8K, MATH), custom benchmark support, provider comparison, and leaderboard generation.
  Adapter: `internal/adapters/benchmark/adapter.go`.

## Performance Characteristics

- **Latency**: <100ms for cached requests, <2s for ensemble responses
- **Throughput**: 1000+ requests/second per instance
- **Availability**: 99.9% SLA with multi-region deployment
- **Scalability**: Horizontal scaling with Kubernetes

## Monitoring & Observability

- **Metrics**: Prometheus for system and business metrics
- **Logging**: Structured logging with correlation IDs
- **Tracing**: Distributed tracing for request flows
- **Alerts**: Automated alerting for anomalies
- **Dashboards**: Grafana dashboards for real-time monitoring

---

For implementation details, see the [HelixAgent source code](https://dev.helix.agent).