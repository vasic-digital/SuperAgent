# HelixAgent BigData Integration Guide

## Introduction

HelixAgent includes a BigData integration layer that connects five large-scale data processing components into the ensemble LLM pipeline. These components handle infinite conversation context, distributed memory synchronization, knowledge graph streaming, analytics, and cross-session learning. Each component can be independently enabled or disabled, and missing infrastructure dependencies (Neo4j, ClickHouse, Kafka) degrade gracefully without blocking startup.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Components](#components)
3. [Configuration](#configuration)
4. [Environment Variables](#environment-variables)
5. [Health Monitoring](#health-monitoring)
6. [Usage Examples](#usage-examples)
7. [Infrastructure Requirements](#infrastructure-requirements)
8. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

The `BigDataIntegration` manager (`internal/bigdata/integration.go`) orchestrates five core components through a unified lifecycle: Initialize, Start, Stop, and HealthCheck. Components communicate through a Kafka message broker for event-driven coordination.

```
BigDataIntegration
  |-- InfiniteContextEngine    (conversation management)
  |-- DistributedMemoryManager (cross-node memory sync)
  |-- StreamingKnowledgeGraph  (Neo4j-backed graph)
  |-- ClickHouseAnalytics      (query analytics)
  |-- CrossSessionLearner      (pattern extraction)
  |
  +-- Kafka MessageBroker      (event bus)
```

---

## Components

### Infinite Context Engine

Manages conversation windows that exceed model token limits. Uses LLM-based compression (temperature 0.1 for consistency) to summarize older context, maintaining a cache of recent sessions.

| Setting | Default | Description |
|---------|---------|-------------|
| Cache Size | 100 | Number of contexts held in memory |
| Cache TTL | 30 minutes | Time before cached context expires |
| Compression Type | hybrid | Summarization strategy (hybrid, extractive, abstractive) |

### Distributed Memory

Synchronizes memory state across multiple HelixAgent nodes using CRDT conflict resolution and event sourcing over Kafka. See the [Memory System Guide](15-memory-system.md) for details on the underlying memory model.

### Knowledge Graph Streaming

Maintains a streaming knowledge graph backed by Neo4j. Entities and relationships extracted during conversations are ingested in real time, enabling graph-based retrieval and reasoning.

### ClickHouse Analytics

Stores detailed request analytics (latencies, token usage, provider performance) in ClickHouse for high-throughput analytical queries. Useful for monitoring provider cost-effectiveness over time.

### Cross-Session Learning

Extracts patterns from completed sessions and applies learned preferences to future interactions. Requires a minimum confidence threshold and frequency count before patterns are promoted.

| Setting | Default | Description |
|---------|---------|-------------|
| Min Confidence | 0.7 | Threshold for pattern promotion |
| Min Frequency | 3 | Minimum occurrences before learning |

---

## Configuration

### Environment Variables

All BigData components are configured via `BIGDATA_ENABLE_*` environment variables and infrastructure connection strings.

| Variable | Default | Description |
|----------|---------|-------------|
| `BIGDATA_ENABLE_INFINITE_CONTEXT` | `true` | Enable infinite context engine |
| `BIGDATA_ENABLE_DISTRIBUTED_MEMORY` | `false` | Enable distributed memory sync |
| `BIGDATA_ENABLE_KNOWLEDGE_GRAPH` | `false` | Enable Neo4j knowledge graph |
| `BIGDATA_ENABLE_ANALYTICS` | `false` | Enable ClickHouse analytics |
| `BIGDATA_ENABLE_CROSS_LEARNING` | `true` | Enable cross-session learning |

### Kafka Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `KAFKA_BOOTSTRAP_SERVERS` | `localhost:9092` | Kafka broker addresses |
| `KAFKA_CONSUMER_GROUP` | `helixagent-bigdata` | Consumer group ID |

### ClickHouse Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CLICKHOUSE_HOST` | `localhost` | ClickHouse server host |
| `CLICKHOUSE_PORT` | `9000` | ClickHouse native port |
| `CLICKHOUSE_DATABASE` | `helixagent_analytics` | Database name |
| `CLICKHOUSE_USER` | `default` | Username |
| `CLICKHOUSE_PASSWORD` | (empty) | Password |

### Neo4j Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NEO4J_URI` | `bolt://localhost:7687` | Neo4j Bolt protocol URI |
| `NEO4J_USERNAME` | `neo4j` | Username |
| `NEO4J_PASSWORD` | `helixagent123` | Password |
| `NEO4J_DATABASE` | `helixagent` | Database name |

---

## Health Monitoring

The BigData integration exposes health status through the `/v1/bigdata/health` endpoint, which returns per-component status:

```bash
curl http://localhost:8080/v1/bigdata/health
```

```json
{
  "infinite_context": "healthy",
  "distributed_memory": "disabled",
  "knowledge_graph": "disabled",
  "analytics": "disabled",
  "cross_learning": "healthy",
  "overall": "healthy"
}
```

Components that are disabled report as `"disabled"` rather than unhealthy. The `overall` status is `"healthy"` if all enabled components are healthy.

You can also check BigData status through the unified monitoring endpoint:

```bash
make monitoring-status
```

---

## Usage Examples

### Enabling All Components

Set the following in your `.env` file or environment:

```bash
BIGDATA_ENABLE_INFINITE_CONTEXT=true
BIGDATA_ENABLE_DISTRIBUTED_MEMORY=true
BIGDATA_ENABLE_KNOWLEDGE_GRAPH=true
BIGDATA_ENABLE_ANALYTICS=true
BIGDATA_ENABLE_CROSS_LEARNING=true

KAFKA_BOOTSTRAP_SERVERS=kafka:9092
CLICKHOUSE_HOST=clickhouse
NEO4J_URI=bolt://neo4j:7687
```

### Starting Infrastructure

Use Docker Compose to bring up the required backing services:

```bash
make infra-start
```

Or start individual services:

```bash
docker compose -f docker/bigdata/docker-compose.bigdata.yml up -d
```

### Programmatic Access

The BigData integration exposes accessor methods for each component:

```go
bdi := server.GetBigDataIntegration()

// Access infinite context
engine := bdi.GetInfiniteContext()

// Access knowledge graph
graph := bdi.GetKnowledgeGraph()

// Check running state
if bdi.IsRunning() {
    status := bdi.HealthCheck(ctx)
}
```

---

## Infrastructure Requirements

| Component | Requires | Port |
|-----------|----------|------|
| Infinite Context | None (in-process) | N/A |
| Distributed Memory | Kafka | 9092 |
| Knowledge Graph | Neo4j | 7687 |
| Analytics | ClickHouse | 9000 |
| Cross-Session Learning | None (in-process) | N/A |

All external dependencies are optional. If a required service is unavailable when its component is enabled, initialization will fail with a descriptive error and HelixAgent will continue operating without that component.

---

## Troubleshooting

**Component fails to initialize**: Check that the backing service (Kafka, Neo4j, ClickHouse) is reachable. Verify connection strings in environment variables.

**High memory usage with infinite context**: Reduce `ContextCacheSize` or lower `ContextCacheTTL` to expire cached contexts sooner.

**Knowledge graph queries are slow**: Ensure Neo4j has adequate memory and that indexes are created on entity types used in queries.

**Analytics data not appearing**: Confirm ClickHouse is running and the database exists. Check that the `BIGDATA_ENABLE_ANALYTICS` variable is set to `true`.
