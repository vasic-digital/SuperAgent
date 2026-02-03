# Extracted Modules Catalog

HelixAgent's functionality is decomposed into **20 independent Go modules**, each with its own repository, tests, and documentation. All modules are integrated as git submodules with `replace` directives in the root `go.mod` for local development.

## Module Index

| # | Module | Go Path | Directory | Packages | Phase |
|---|--------|---------|-----------|----------|-------|
| 1 | EventBus | `digital.vasic.eventbus` | `EventBus/` | 4 | Foundation |
| 2 | Concurrency | `digital.vasic.concurrency` | `Concurrency/` | 6 | Foundation |
| 3 | Observability | `digital.vasic.observability` | `Observability/` | 5 | Foundation |
| 4 | Auth | `digital.vasic.auth` | `Auth/` | 5 | Foundation |
| 5 | Storage | `digital.vasic.storage` | `Storage/` | 4 | Foundation |
| 6 | Streaming | `digital.vasic.streaming` | `Streaming/` | 6 | Foundation |
| 7 | Security | `digital.vasic.security` | `Security/` | 5 | Infrastructure |
| 8 | VectorDB | `digital.vasic.vectordb` | `VectorDB/` | 5 | Infrastructure |
| 9 | Embeddings | `digital.vasic.embeddings` | `Embeddings/` | 7 | Infrastructure |
| 10 | Database | `digital.vasic.database` | `Database/` | 7 | Infrastructure |
| 11 | Cache | `digital.vasic.cache` | `Cache/` | 5 | Infrastructure |
| 12 | Messaging | `digital.vasic.messaging` | `Messaging/` | 5 | Services |
| 13 | Formatters | `digital.vasic.formatters` | `Formatters/` | 6 | Services |
| 14 | MCP | `digital.vasic.mcp` | `MCP_Module/` | 6 | Services |
| 15 | RAG | `digital.vasic.rag` | `RAG/` | 5 | Integration |
| 16 | Memory | `digital.vasic.memory` | `Memory/` | 4 | Integration |
| 17 | Optimization | `digital.vasic.optimization` | `Optimization/` | 6 | Integration |
| 18 | Plugins | `digital.vasic.plugins` | `Plugins/` | 5 | Integration |
| 19 | Containers | `digital.vasic.containers` | `Containers/` | 12 | Pre-existing |
| 20 | Challenges | `digital.vasic.challenges` | `Challenges/` | 12 | Pre-existing |

**Total: 20 modules, 118 packages**

---

## Phase 1: Foundation Modules (Zero Dependencies)

These modules have no dependencies on other extracted modules.

### EventBus (`digital.vasic.eventbus`)

Pub/sub event system with synchronous/async dispatch, topic filtering, and middleware.

| Package | Purpose |
|---------|---------|
| `pkg/bus` | Core event bus with publish/subscribe |
| `pkg/event` | Event type definitions |
| `pkg/filter` | Topic-based event filtering with glob patterns |
| `pkg/middleware` | Event middleware (logging, metrics, retry) |

**Patterns**: Observer, Mediator, Pub/Sub, Middleware chain

### Concurrency (`digital.vasic.concurrency`)

Concurrency primitives: worker pools, priority queues, rate limiters, circuit breakers, semaphores.

| Package | Purpose |
|---------|---------|
| `pkg/pool` | Worker pool with task submission, batching |
| `pkg/queue` | Generic thread-safe priority task queue |
| `pkg/limiter` | Rate limiting (token bucket, sliding window) |
| `pkg/breaker` | Circuit breaker (closed/open/half-open) |
| `pkg/semaphore` | Weighted semaphore for resource access |
| `pkg/monitor` | System resource monitoring (CPU, memory, disk) |

**Patterns**: Worker Pool, Semaphore, Circuit Breaker, Rate Limiter

### Observability (`digital.vasic.observability`)

Application observability: distributed tracing, metrics, structured logging, health checks.

| Package | Purpose |
|---------|---------|
| `pkg/trace` | OpenTelemetry tracing (OTLP/Jaeger/Zipkin/stdout) |
| `pkg/metrics` | Prometheus metrics (counters, histograms, gauges) |
| `pkg/logging` | Structured logging with correlation IDs |
| `pkg/health` | Health check aggregation |
| `pkg/analytics` | ClickHouse analytics adapter |

**Patterns**: Strategy, Adapter, Null Object, Aggregator, Graceful Degradation

### Auth (`digital.vasic.auth`)

Authentication and token management.

| Package | Purpose |
|---------|---------|
| `pkg/jwt` | JWT creation, validation, refresh |
| `pkg/apikey` | API key authentication |
| `pkg/oauth` | OAuth2 token management |
| `pkg/middleware` | HTTP auth middleware |
| `pkg/token` | Token abstraction and revocation |

**Patterns**: Strategy, Factory, Decorator

### Storage (`digital.vasic.storage`)

Object storage abstraction with multiple backends.

| Package | Purpose |
|---------|---------|
| `pkg/object` | Unified object store interface |
| `pkg/local` | Local filesystem storage |
| `pkg/s3` | S3-compatible storage (AWS S3, MinIO) |
| `pkg/provider` | Cloud provider abstraction |

**Patterns**: Adapter, Factory, Strategy

### Streaming (`digital.vasic.streaming`)

Real-time streaming and transport protocols.

| Package | Purpose |
|---------|---------|
| `pkg/sse` | Server-Sent Events with reconnection |
| `pkg/websocket` | WebSocket with ping/pong and rooms |
| `pkg/grpc` | gRPC streaming (unary, server, client, bidi) |
| `pkg/webhook` | Webhook dispatch with retry and signing |
| `pkg/http` | HTTP client with retry, timeout, circuit breaker |
| `pkg/transport` | Transport abstraction (HTTP/gRPC/WebSocket) |

**Patterns**: Observer, Strategy, Adapter

---

## Phase 2: Infrastructure Modules

Larger modules with no cross-module dependencies.

### Security (`digital.vasic.security`)

Content security, guardrails, and policy enforcement.

| Package | Purpose |
|---------|---------|
| `pkg/guardrails` | Configurable content guardrails engine |
| `pkg/pii` | PII detection and redaction |
| `pkg/content` | Input/output content filtering |
| `pkg/policy` | Security policy enforcement |
| `pkg/scanner` | Vulnerability scanning integration |

**Patterns**: Chain of Responsibility, Strategy, Proxy

### VectorDB (`digital.vasic.vectordb`)

Unified vector store with 4 backend adapters.

| Package | Purpose |
|---------|---------|
| `pkg/client` | Unified VectorStore interface |
| `pkg/qdrant` | Qdrant adapter |
| `pkg/pinecone` | Pinecone adapter |
| `pkg/milvus` | Milvus adapter |
| `pkg/pgvector` | pgvector (PostgreSQL) adapter |

**Patterns**: Adapter, Factory, Repository

### Embeddings (`digital.vasic.embeddings`)

Embedding generation with 6 provider adapters.

| Package | Purpose |
|---------|---------|
| `pkg/provider` | Unified EmbeddingProvider interface |
| `pkg/openai` | OpenAI embeddings |
| `pkg/cohere` | Cohere embeddings |
| `pkg/voyage` | Voyage AI embeddings |
| `pkg/jina` | Jina AI embeddings |
| `pkg/google` | Google Vertex AI embeddings |
| `pkg/bedrock` | AWS Bedrock embeddings |

**Patterns**: Strategy, Factory, Adapter

### Database (`digital.vasic.database`)

Database abstraction with PostgreSQL, SQLite, and utilities.

| Package | Purpose |
|---------|---------|
| `pkg/database` | Unified Database interface |
| `pkg/postgres` | PostgreSQL via pgx/v5 |
| `pkg/sqlite` | SQLite via modernc.org/sqlite |
| `pkg/pool` | Connection pool management |
| `pkg/migration` | Schema migration runner |
| `pkg/repository` | Generic Repository[T] pattern |
| `pkg/query` | Type-safe query builder |

**Patterns**: Repository, Factory, Adapter, Builder

### Cache (`digital.vasic.cache`)

Caching with Redis, in-memory, and distributed modes.

| Package | Purpose |
|---------|---------|
| `pkg/cache` | Unified Cache interface |
| `pkg/redis` | Redis cache via go-redis/v9 |
| `pkg/memory` | In-memory cache with LRU/LFU eviction |
| `pkg/distributed` | Distributed cache with consistency |
| `pkg/policy` | TTL policies (fixed, sliding, adaptive) |

**Patterns**: Strategy, Decorator, Proxy

---

## Phase 3: Service Modules

### Messaging (`digital.vasic.messaging`)

Message broker abstraction with Kafka and RabbitMQ.

| Package | Purpose |
|---------|---------|
| `pkg/broker` | Unified MessageBroker interface |
| `pkg/producer` | Message producer abstraction |
| `pkg/consumer` | Consumer with group rebalancing |
| `pkg/kafka` | Kafka via segmentio/kafka-go |
| `pkg/rabbitmq` | RabbitMQ via amqp091-go |

**Patterns**: Adapter, Observer, Factory

### Formatters (`digital.vasic.formatters`)

Code formatting framework with 32+ formatters.

| Package | Purpose |
|---------|---------|
| `pkg/formatter` | Formatter interface and types |
| `pkg/registry` | Formatter registry with discovery |
| `pkg/native` | Native formatters (Go, Python, JS, etc.) |
| `pkg/service` | Service-based formatters (Docker containers) |
| `pkg/executor` | Format execution pipeline |
| `pkg/cache` | Format result caching |

**Patterns**: Strategy, Registry, Factory, Chain of Responsibility

### MCP (`digital.vasic.mcp`)

Model Context Protocol adapter framework.

| Package | Purpose |
|---------|---------|
| `pkg/adapter` | MCP adapter interface |
| `pkg/client` | MCP protocol client |
| `pkg/server` | MCP protocol server |
| `pkg/config` | Container config generation |
| `pkg/registry` | Adapter registry with discovery |
| `pkg/protocol` | JSON-RPC 2.0 transport |

**Patterns**: Adapter, Factory, Registry, Facade

---

## Phase 4: Integration Modules

### RAG (`digital.vasic.rag`)

Retrieval-Augmented Generation pipeline.

| Package | Purpose |
|---------|---------|
| `pkg/retriever` | Document retrieval interface |
| `pkg/reranker` | Result reranking (cross-encoder, MMR) |
| `pkg/chunker` | Document chunking (fixed, recursive, semantic) |
| `pkg/pipeline` | RAG pipeline composition |
| `pkg/hybrid` | Hybrid retrieval (semantic + keyword) |

**Patterns**: Facade, Strategy, Template Method, Pipeline

### Memory (`digital.vasic.memory`)

Mem0-style memory management with entity graphs.

| Package | Purpose |
|---------|---------|
| `pkg/mem0` | Mem0-compatible memory store |
| `pkg/entity` | Entity definitions and relationships |
| `pkg/graph` | Entity graph with relationship tracking |
| `pkg/store` | Memory storage abstraction |

**Patterns**: Strategy, Facade, Repository

### Optimization (`digital.vasic.optimization`)

LLM optimization: caching, structured output, prompt compression.

| Package | Purpose |
|---------|---------|
| `pkg/gptcache` | GPT-Cache semantic caching |
| `pkg/outlines` | Outlines structured output constraints |
| `pkg/streaming` | Streaming optimizations |
| `pkg/sglang` | SGLang integration |
| `pkg/adapter` | LlamaIndex/LangChain adapters |
| `pkg/prompt` | Prompt optimization and compression |

**Patterns**: Proxy, Decorator, Strategy

### Plugins (`digital.vasic.plugins`)

Plugin system with lifecycle management and sandboxing.

| Package | Purpose |
|---------|---------|
| `pkg/plugin` | Plugin interface with Init/Start/Stop lifecycle |
| `pkg/registry` | Plugin registry with versioning |
| `pkg/loader` | Dynamic plugin loading |
| `pkg/sandbox` | Plugin sandboxing and isolation |
| `pkg/structured` | Structured output parsing/validation |

**Patterns**: Abstract Factory, Registry, Strategy, Template Method

---

## Pre-existing Modules

### Containers (`digital.vasic.containers`)

Generic container orchestration framework. 12 packages covering runtime abstraction (Docker/Podman/K8s), health checking (TCP/HTTP/gRPC), compose orchestration, lifecycle management, resource monitoring, event bus, service discovery, and boot management.

### Challenges (`digital.vasic.challenges`)

Generic challenge framework. 12 packages covering challenge interface, assertion engine (16 evaluators), registry with dependency ordering, runner (sequential/parallel/pipeline), reporting (MD/JSON/HTML), structured logging, env management, live monitoring, metrics, and plugin system.

---

## Development

### Testing All Modules

```bash
# Test a single module
cd EventBus && go test ./... -count=1 -race && cd ..

# Test all modules
for mod in EventBus Concurrency Observability Auth Storage Streaming \
           Security VectorDB Embeddings Database Cache \
           Messaging Formatters MCP_Module RAG Memory Optimization Plugins; do
  echo "Testing $mod..."
  (cd $mod && go test ./... -count=1 -race -short)
done

# Test HelixAgent with all modules
go build ./cmd/... ./internal/...
go test ./internal/... -short -count=1
```

### Module Dependencies in go.mod

All modules use `replace` directives for local development:

```go
replace digital.vasic.eventbus => ./EventBus
replace digital.vasic.concurrency => ./Concurrency
replace digital.vasic.observability => ./Observability
// ... etc
```

### Adding a New Module

1. Create directory with `go mod init digital.vasic.<name>`
2. Add packages under `pkg/`
3. Write tests (table-driven, testify, `-race`)
4. Create CLAUDE.md, AGENTS.md, README.md, docs/
5. Add `require` + `replace` in root go.mod
6. Add submodule entry in .gitmodules
7. Run `go mod tidy && go mod vendor`

## Architecture Diagram

```
HelixAgent (dev.helix.agent)
├── Foundation Layer (zero dependencies)
│   ├── EventBus ─── Pub/Sub event system
│   ├── Concurrency ─── Pools, limiters, breakers
│   ├── Observability ─── Tracing, metrics, logging
│   ├── Auth ─── JWT, API key, OAuth
│   ├── Storage ─── S3, local filesystem
│   └── Streaming ─── SSE, WS, gRPC, webhooks
├── Infrastructure Layer
│   ├── Security ─── Guardrails, PII, policies
│   ├── VectorDB ─── Qdrant, Pinecone, Milvus, pgvector
│   ├── Embeddings ─── 6 embedding providers
│   ├── Database ─── PostgreSQL, SQLite
│   └── Cache ─── Redis, in-memory
├── Services Layer
│   ├── Messaging ─── Kafka, RabbitMQ
│   ├── Formatters ─── 32+ code formatters
│   └── MCP ─── Model Context Protocol
├── Integration Layer
│   ├── RAG ─── Retrieval-Augmented Generation
│   ├── Memory ─── Mem0-style memory
│   ├── Optimization ─── GPT-Cache, prompt optimization
│   └── Plugins ─── Plugin system
└── Pre-existing
    ├── Containers ─── Container orchestration
    └── Challenges ─── Challenge framework
```
