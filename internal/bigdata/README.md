# bigdata

Package `bigdata` is the integration layer that orchestrates HelixAgent's big-data subsystems: infinite context, distributed memory, knowledge graph streaming, ClickHouse analytics, and cross-session learning. It provides a unified lifecycle (initialize, start, stop, health check) and exposes HTTP endpoints via Gin for all subsystem operations.

## Architecture

`BigDataIntegration` is the central coordinator. Each subsystem can be independently enabled or disabled via `IntegrationConfig` flags. Components communicate through a Kafka `MessageBroker` and are initialized lazily during the `Initialize` call.

### Key Types

- **`BigDataIntegration`** -- Manages lifecycle of all big-data components with per-component timeout on shutdown.
- **`IntegrationConfig`** -- Feature flags and connection settings for Kafka, ClickHouse, Neo4j, context engine, and learning.
- **`Handler`** -- Gin HTTP handler exposing REST endpoints for all subsystems.
- **`DebateIntegration`** -- Bridges debate context with the infinite context engine.

### Components

| Component             | Config Flag                 | Dependencies                  |
|-----------------------|-----------------------------|-------------------------------|
| Infinite Context      | `EnableInfiniteContext`      | Kafka, ProviderRegistry (LLM) |
| Distributed Memory    | `EnableDistributedMemory`   | Kafka                         |
| Knowledge Graph       | `EnableKnowledgeGraph`      | Kafka, Neo4j                  |
| ClickHouse Analytics  | `EnableAnalytics`           | ClickHouse                    |
| Cross-Session Learning| `EnableCrossLearning`       | Kafka                         |

## Public API

```go
// Construction and lifecycle
NewBigDataIntegration(config *IntegrationConfig, broker messaging.MessageBroker, logger *logrus.Logger) (*BigDataIntegration, error)
Initialize(ctx context.Context) error
Start(ctx context.Context) error
Stop(ctx context.Context) error
HealthCheck(ctx context.Context) map[string]string
IsRunning() bool

// Component accessors
GetInfiniteContext() *conversation.InfiniteContextEngine
GetDistributedMemory() *memory.DistributedMemoryManager
GetKnowledgeGraph() *knowledge.StreamingKnowledgeGraph
GetAnalytics() *analytics.ClickHouseAnalytics
GetCrossLearner() *learning.CrossSessionLearner
```

## HTTP Endpoints

Registered via `Handler.RegisterRoutes(router)`:

| Method | Path                               | Description                       |
|--------|-------------------------------------|-----------------------------------|
| POST   | `/v1/context/replay`               | Replay conversation with compression |
| GET    | `/v1/context/stats/:conversation_id` | Context statistics              |
| GET    | `/v1/memory/sync/status`           | Distributed memory sync status    |
| POST   | `/v1/memory/sync/force`            | Force memory synchronization      |
| GET    | `/v1/knowledge/related/:entity_id` | Related entities (graph traversal)|
| POST   | `/v1/knowledge/search`             | Knowledge graph search            |
| GET    | `/v1/analytics/provider/:provider` | Provider analytics                |
| GET    | `/v1/analytics/debate/:debate_id`  | Debate analytics                  |
| POST   | `/v1/analytics/query`              | Custom ClickHouse query           |
| GET    | `/v1/learning/insights`            | Learning insights                 |
| GET    | `/v1/learning/patterns`            | Learned patterns                  |
| GET    | `/v1/bigdata/health`               | Component health check            |

## Configuration

Components are configured via `IntegrationConfig` or `BIGDATA_ENABLE_*` environment variables. Missing dependencies (Neo4j, ClickHouse, Kafka) cause graceful degradation -- disabled components report `"disabled"` in health checks.

```go
config := bigdata.DefaultIntegrationConfig()
config.EnableAnalytics = true
config.ClickHouseHost = "clickhouse.local"
```

## Testing

```bash
go test -v ./internal/bigdata/
make test-with-infra  # Full integration with Docker infrastructure
```

Tests requiring Kafka, ClickHouse, or Neo4j use live services started via `make test-infra-start`.
