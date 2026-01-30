# HelixAgent Big Data Integration - User Guide

**Version**: 1.0
**Last Updated**: 2026-01-30

---

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Infinite Context Engine](#infinite-context-engine)
4. [Distributed Memory](#distributed-memory)
5. [Knowledge Graph](#knowledge-graph)
6. [Analytics & Insights](#analytics--insights)
7. [Data Lake & Batch Processing](#data-lake--batch-processing)
8. [Cross-Session Learning](#cross-session-learning)
9. [Configuration](#configuration)
10. [Troubleshooting](#troubleshooting)

---

## Introduction

HelixAgent's Big Data Integration transforms the platform into an enterprise-grade streaming AI system with:

- **Infinite Context**: Unlimited conversation history via Kafka event sourcing
- **Distributed Memory**: Multi-node memory synchronization with CRDT conflict resolution
- **Real-Time Knowledge Graph**: Streaming entity updates to Neo4j
- **Sub-100ms Analytics**: Time-series metrics in ClickHouse
- **Batch Processing**: Apache Spark for large-scale data analysis
- **Cross-Session Learning**: Continuous knowledge accumulation

### Architecture Overview

```
User → REST API → AI Debate → Kafka Streams → Storage (PostgreSQL, Neo4j, ClickHouse)
                                ↓
                    Infinite Context Engine
                    Distributed Memory Sync
                    Knowledge Graph Updates
                    Cross-Session Learning
```

---

## Quick Start

### Prerequisites

- Docker & Docker Compose
- 8GB RAM minimum (16GB recommended)
- 20GB free disk space

### 1. Start Big Data Services

```bash
# Start all services
docker-compose -f docker-compose.bigdata.yml up -d

# Verify services are healthy
docker-compose -f docker-compose.bigdata.yml ps
```

**Services Started**:
- Zookeeper (port 2181)
- Kafka (port 9092)
- PostgreSQL (port 5432)
- Redis (port 6379)
- Neo4j (ports 7474, 7687)
- ClickHouse (ports 8123, 9000)
- MinIO (ports 9000, 9001)

### 2. Start HelixAgent

```bash
# Build and run
make build
./bin/helixagent

# Or use Docker
docker-compose up -d helixagent
```

### 3. Verify Integration

```bash
# Check health endpoint
curl http://localhost:7061/health

# Expected response:
{
  "status": "healthy",
  "services": {
    "kafka": "connected",
    "neo4j": "connected",
    "clickhouse": "connected",
    "minio": "connected"
  }
}
```

### 4. Run Your First Long Conversation

```bash
# Start a conversation
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Explain microservices architecture",
    "conversation_id": "my-first-conversation"
  }'

# Continue the conversation (unlimited history!)
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "What about service discovery?",
    "conversation_id": "my-first-conversation"
  }'
```

---

## Infinite Context Engine

### Overview

Traditional LLMs have token limits (4K-128K). HelixAgent eliminates this limitation by:
1. Storing all conversation events in Kafka (unlimited retention)
2. Replaying full history when needed
3. Intelligently compressing when token limits are exceeded

### How It Works

```
User Message → Kafka Event → Event Log (Permanent Storage)
                                ↓
Context Needed? → Replay from Kafka → Full History Retrieved
                                ↓
Too Many Tokens? → LLM Compression → Compressed Context (30% ratio)
```

### Usage

**Automatic (Recommended)**:
```bash
# Just continue your conversation - context is preserved automatically
curl -X POST http://localhost:7061/v1/debates \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Your question",
    "conversation_id": "conv-123"
  }'
```

**Manual Replay**:
```bash
# Replay specific conversation
curl -X POST http://localhost:7061/v1/context/replay \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "conv-123",
    "compression_strategy": "hybrid"
  }'
```

### Compression Strategies

| Strategy | Use Case | Quality | Speed |
|----------|----------|---------|-------|
| **Window Summary** | Short conversations (<500 msg) | High | Fast |
| **Entity Graph** | Entity-heavy conversations | Very High | Medium |
| **Full Summary** | General purpose | Medium | Slow |
| **Hybrid** (Default) | All conversations | Very High | Medium |

**Configure Strategy**:
```bash
# Set compression strategy
export CONTEXT_COMPRESSION_STRATEGY=hybrid  # window, entity, full, hybrid
export CONTEXT_COMPRESSION_RATIO=0.30      # Target 30% compression
export CONTEXT_MIN_QUALITY_SCORE=0.95      # Minimum quality threshold
```

### Compression Example

**Before Compression (10,000 messages, 40,000 tokens)**:
```json
{
  "messages": [
    {"role": "user", "content": "Message 1..."},
    {"role": "assistant", "content": "Response 1..."},
    ...
    {"role": "user", "content": "Message 10000..."}
  ]
}
```

**After Hybrid Compression (30% ratio, 12,000 tokens)**:
```json
{
  "summary": "User discussed microservices architecture over 3 sessions...",
  "key_entities": ["Docker", "Kubernetes", "API Gateway", "Service Mesh"],
  "key_decisions": ["Use gRPC for inter-service communication"],
  "recent_messages": [
    {"role": "user", "content": "Message 9950..."},
    ...
    {"role": "user", "content": "Message 10000..."}
  ],
  "compression_stats": {
    "original_tokens": 40000,
    "compressed_tokens": 12000,
    "ratio": 0.30,
    "quality_score": 0.96
  }
}
```

### Benefits

✅ **No Token Limit**: Store unlimited conversation history
✅ **Perfect Recall**: Access any message from any point in time
✅ **Quality Preserved**: LLM-based compression maintains coherence
✅ **Automatic**: No manual intervention required
✅ **Fast**: Cached results for <100ms retrieval

---

## Distributed Memory

### Overview

Run multiple HelixAgent nodes with synchronized memory using Kafka-based event sourcing and CRDT conflict resolution.

### Architecture

```
Node 1 (US-East) ←→ Kafka ←→ Node 2 (US-West) ←→ Node 3 (EU-Central)
     ↓                            ↓                      ↓
   PostgreSQL                PostgreSQL              PostgreSQL
```

### Setup Multi-Node Deployment

**Node 1 (Primary)**:
```bash
# .env.node1
NODE_ID=node-us-east-1
KAFKA_BOOTSTRAP_SERVERS=kafka-cluster:9092
DB_HOST=postgres-node1
MEMORY_EVENT_SOURCING_ENABLED=true
```

**Node 2 (Secondary)**:
```bash
# .env.node2
NODE_ID=node-us-west-1
KAFKA_BOOTSTRAP_SERVERS=kafka-cluster:9092
DB_HOST=postgres-node2
MEMORY_EVENT_SOURCING_ENABLED=true
```

**Start Nodes**:
```bash
# Node 1
docker-compose -f docker-compose.node1.yml up -d

# Node 2
docker-compose -f docker-compose.node2.yml up -d
```

### How Memory Sync Works

1. **User creates memory on Node 1**:
   ```bash
   curl -X POST http://node1:7061/v1/memory \
     -H "Content-Type: application/json" \
     -d '{"content": "User prefers OAuth2 for authentication"}'
   ```

2. **Node 1 publishes event to Kafka**:
   ```
   Topic: helixagent.memory.events
   Event: {type: "Created", node_id: "node-us-east-1", memory_id: "mem-123", ...}
   ```

3. **Node 2 receives event and replicates**:
   ```
   Node 2 → Consume Event → Store Locally → Update Cache
   ```

4. **Memory is now available on both nodes**:
   ```bash
   # Query from Node 2
   curl http://node2:7061/v1/memory/mem-123
   # Returns the same memory
   ```

### Conflict Resolution (CRDT)

**Scenario**: Both Node 1 and Node 2 update the same memory simultaneously.

**Without CRDT** (❌ Bad):
```
Node 1: memory.content = "Prefer OAuth2"
Node 2: memory.content = "Prefer JWT"
Result: One update is lost!
```

**With CRDT** (✅ Good):
```
Node 1: memory.content = "Prefer OAuth2" (timestamp: 10:00:00, version: 2)
Node 2: memory.content = "Prefer JWT" (timestamp: 10:00:01, version: 2)

CRDT Resolver detects conflict (version 2 vs 2)
Applies merge strategy (MergeAll):
  Merged: memory.content = "Prefer OAuth2 and JWT"
  New version: 3

Both nodes receive merged result
```

**Conflict Strategies**:
```bash
# LastWriteWins (default)
export CRDT_STRATEGY=last_write_wins

# MergeAll (merge content)
export CRDT_STRATEGY=merge_all

# Custom logic (implement your own)
export CRDT_STRATEGY=custom
```

### Monitoring Memory Sync

```bash
# Check sync status
curl http://localhost:7061/v1/memory/sync/status

{
  "node_id": "node-us-east-1",
  "events_published": 1234,
  "events_consumed": 5678,
  "conflicts_resolved": 12,
  "last_sync_timestamp": "2026-01-30T10:00:00Z",
  "sync_lag_ms": 45
}
```

---

## Knowledge Graph

### Overview

HelixAgent automatically builds and maintains a knowledge graph in Neo4j with real-time streaming updates.

### Graph Schema

**Nodes**:
- `(:Entity)` - Extracted entities (people, technologies, concepts)
- `(:Conversation)` - Conversation metadata
- `(:User)` - User profiles

**Relationships**:
- `[:RELATED_TO]` - Entities that are related
- `[:MENTIONED_IN]` - Entity mentioned in conversation
- `[:CO_OCCURS_WITH]` - Entities that appear together

### Automatic Entity Extraction

```bash
# Start a conversation
curl -X POST http://localhost:7061/v1/debates \
  -d '{"topic": "Tell me about Docker and Kubernetes"}'

# System automatically:
# 1. Extracts entities: Docker (TECH), Kubernetes (TECH)
# 2. Creates nodes in Neo4j
# 3. Creates relationship: Docker -[:RELATED_TO]-> Kubernetes
```

### Query the Knowledge Graph

**Via Cypher (Direct Neo4j)**:
```cypher
// Find all entities related to "Docker"
MATCH (e1:Entity {name: "Docker"})-[:RELATED_TO]-(e2:Entity)
RETURN e2.name, e2.type
LIMIT 10;

// Find entity co-occurrences
MATCH (e1:Entity)-[r:CO_OCCURS_WITH]->(e2:Entity)
WHERE r.frequency > 5
RETURN e1.name, e2.name, r.frequency
ORDER BY r.frequency DESC;
```

**Via REST API**:
```bash
# Get related entities
curl http://localhost:7061/v1/knowledge/related?entity=Docker

{
  "entity": "Docker",
  "related": [
    {"name": "Kubernetes", "type": "TECH", "strength": 0.92},
    {"name": "Container", "type": "CONCEPT", "strength": 0.88},
    {"name": "Microservices", "type": "CONCEPT", "strength": 0.75}
  ]
}
```

### Visualize the Graph

**Neo4j Browser** (http://localhost:7474):
```cypher
// Show entire graph (limited)
MATCH (n)-[r]->(m)
RETURN n, r, m
LIMIT 100;

// Show Docker ecosystem
MATCH path = (e1:Entity {name: "Docker"})-[:RELATED_TO*1..2]-(e2)
RETURN path;
```

### Use Cases

1. **Recommendation**: "You asked about Docker, you might also want to learn about Kubernetes"
2. **Context Enrichment**: Automatically fetch related concepts during conversations
3. **Knowledge Discovery**: Find hidden relationships between entities
4. **Topic Clustering**: Group conversations by related entities

---

## Analytics & Insights

### Overview

ClickHouse provides sub-100ms analytics queries for real-time insights.

### Available Metrics

| Metric Type | Description | Latency |
|-------------|-------------|---------|
| **Provider Performance** | Response time, confidence, win rate | <50ms |
| **Conversation Metrics** | Message count, duration, entity density | <30ms |
| **Debate Analytics** | Round count, winner distribution | <40ms |
| **Entity Metrics** | Extraction rate, unique count | <20ms |
| **System Health** | CPU, memory, request rate | <10ms |

### Query Examples

**Provider Performance (Last 24 Hours)**:
```bash
curl "http://localhost:7061/v1/analytics/providers?window=24h"

{
  "providers": [
    {
      "provider": "claude",
      "total_requests": 1234,
      "avg_response_time_ms": 850,
      "p95_response_time_ms": 1200,
      "avg_confidence": 0.92,
      "win_rate": 0.45
    },
    {
      "provider": "deepseek",
      "total_requests": 987,
      "avg_response_time_ms": 650,
      "p95_response_time_ms": 900,
      "avg_confidence": 0.87,
      "win_rate": 0.38
    }
  ]
}
```

**Conversation Trends**:
```bash
curl "http://localhost:7061/v1/analytics/conversations/trends?days=7"

{
  "daily_stats": [
    {"date": "2026-01-24", "count": 245, "avg_messages": 12, "avg_duration_min": 8},
    {"date": "2026-01-25", "count": 287, "avg_messages": 15, "avg_duration_min": 10},
    ...
  ]
}
```

**Debate Winners**:
```bash
curl "http://localhost:7061/v1/analytics/debates/winners?limit=10"

{
  "top_winners": [
    {"provider": "claude", "model": "claude-3-5-sonnet", "wins": 156, "percentage": 45.2},
    {"provider": "deepseek", "model": "deepseek-chat", "wins": 132, "percentage": 38.3},
    ...
  ]
}
```

### Direct ClickHouse Queries

```sql
-- Provider performance over time
SELECT
    toStartOfHour(timestamp) AS hour,
    provider,
    COUNT(*) AS requests,
    AVG(response_time_ms) AS avg_response_time,
    quantile(0.95)(response_time_ms) AS p95_response_time
FROM debate_metrics
WHERE timestamp >= now() - INTERVAL 24 HOUR
GROUP BY hour, provider
ORDER BY hour DESC, avg_response_time ASC;

-- Entity extraction trends
SELECT
    toDate(timestamp) AS date,
    COUNT(*) AS total_extractions,
    COUNT(DISTINCT entity_id) AS unique_entities,
    AVG(entities_per_message) AS avg_entities_per_msg
FROM entity_extraction_metrics
WHERE timestamp >= now() - INTERVAL 7 DAY
GROUP BY date
ORDER BY date DESC;
```

### Grafana Dashboard (Optional)

```bash
# Start Grafana
docker-compose -f docker-compose.monitoring.yml up -d grafana

# Access dashboard
open http://localhost:3000
# Login: admin / helixagent123
```

**Pre-built Dashboards**:
- Provider Performance
- Conversation Analytics
- Entity Growth
- System Health

---

## Data Lake & Batch Processing

### Overview

HelixAgent archives all conversations to a data lake (MinIO/S3) in Parquet format for large-scale batch processing with Apache Spark.

### Data Lake Structure

```
s3://helixagent-datalake/
├── conversations/
│   └── year=2026/month=01/day=30/
│       ├── conversation_conv-123.parquet
│       └── conversation_conv-456.parquet
├── debates/
│   └── year=2026/month=01/day=30/
│       └── debate_debate-789.parquet
├── entities/
│   └── snapshot_2026-01-30T10-00-00.parquet
└── analytics/
    └── daily_aggregates_2026-01-30.parquet
```

### Manual Archive

```bash
# Archive specific conversation
curl -X POST http://localhost:7061/v1/datalake/archive \
  -d '{"conversation_id": "conv-123"}'

# Archive all conversations from yesterday
curl -X POST http://localhost:7061/v1/datalake/archive \
  -d '{"date": "2026-01-29"}'
```

### Run Spark Batch Jobs

**Entity Extraction Job**:
```bash
# Submit Spark job
curl -X POST http://localhost:7061/v1/spark/jobs \
  -d '{
    "job_type": "EntityExtraction",
    "input_path": "s3://helixagent-datalake/conversations/year=2026/month=01/",
    "output_path": "s3://helixagent-datalake/entities/"
  }'

# Monitor job
curl http://localhost:7061/v1/spark/jobs/job-123/status

{
  "job_id": "job-123",
  "status": "running",
  "progress": 0.45,
  "rows_processed": 45000,
  "entities_extracted": 12345
}
```

**Available Job Types**:
- `EntityExtraction` - Extract entities from conversations
- `RelationshipMining` - Find entity co-occurrences
- `TopicModeling` - LDA/NMF topic detection
- `ProviderPerformance` - Aggregate provider statistics
- `DebateAnalysis` - Multi-round debate pattern detection

### Query Data Lake with SQL

**Using Presto/Trino**:
```sql
-- Count conversations by date
SELECT
    year,
    month,
    day,
    COUNT(*) AS conversation_count
FROM helixagent.conversations
WHERE year = 2026 AND month = 1
GROUP BY year, month, day
ORDER BY day DESC;

-- Find most mentioned entities
SELECT
    entity_name,
    entity_type,
    COUNT(*) AS mention_count
FROM helixagent.entities
WHERE year = 2026 AND month = 1
GROUP BY entity_name, entity_type
ORDER BY mention_count DESC
LIMIT 10;
```

---

## Cross-Session Learning

### Overview

HelixAgent continuously learns from all conversations to improve future interactions.

### Learning Types

| Pattern Type | Description | Example |
|--------------|-------------|---------|
| **User Intent** | What the user wants to accomplish | help_seeking, implementation, comparison |
| **User Preference** | Communication style, format | concise, detailed, code-heavy |
| **Entity Co-occurrence** | Entities that appear together | Docker + Kubernetes, OAuth2 + JWT |
| **Debate Strategy** | Which providers win which positions | Claude for researcher, DeepSeek for critic |
| **Conversation Flow** | Timing and pacing | rapid (3s/msg), thoughtful (30s/msg) |
| **Provider Performance** | Historical success rates | Claude: 92% confidence, DeepSeek: 87% |

### How Learning Works

1. **Conversation Completes**:
   ```
   User finishes conversation → Kafka event published
   ```

2. **Pattern Extraction**:
   ```
   Cross-Session Learner → Extracts 6 pattern types → Stores in PostgreSQL
   ```

3. **Insight Generation**:
   ```
   Pattern frequency > threshold → Generate insight → Publish to Kafka
   ```

4. **Future Conversations**:
   ```
   User starts new conversation → Query learned patterns → Adapt behavior
   ```

### View Learned Patterns

```bash
# Get user preferences
curl "http://localhost:7061/v1/learning/preferences?user_id=user-123"

{
  "user_id": "user-123",
  "preferences": [
    {
      "type": "communication_style",
      "value": "concise",
      "confidence": 0.85,
      "frequency": 12
    },
    {
      "type": "response_format",
      "value": "markdown",
      "confidence": 0.78,
      "frequency": 10
    }
  ]
}

# Get top patterns
curl "http://localhost:7061/v1/learning/patterns/top?limit=10"

{
  "patterns": [
    {
      "pattern_type": "entity_cooccurrence",
      "description": "Docker and Kubernetes often discussed together",
      "frequency": 45,
      "confidence": 0.92
    },
    {
      "pattern_type": "debate_strategy",
      "description": "Claude wins researcher position 45% of time",
      "frequency": 156,
      "confidence": 0.89
    }
  ]
}
```

### PostgreSQL Queries

```sql
-- Get user preferences
SELECT * FROM get_user_preferences_summary('user-123');

-- Get entity co-occurrences
SELECT * FROM get_entity_cooccurrence_network('Docker', 10);

-- Get best debate strategies
SELECT * FROM get_best_debate_strategies('researcher', 5);

-- Get learning progress
SELECT * FROM get_learning_progress(30);  -- Last 30 days
```

### Benefits

✅ **Personalized Responses**: Adapt to individual user preferences
✅ **Optimized Debate Teams**: Select best providers based on historical performance
✅ **Proactive Suggestions**: Recommend related topics based on entity relationships
✅ **Continuous Improvement**: System gets smarter with every conversation

---

## Configuration

### Environment Variables

**Kafka Configuration**:
```bash
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_CONSUMER_GROUP=helixagent
KAFKA_AUTO_OFFSET_RESET=earliest
```

**Neo4j Configuration**:
```bash
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=helixagent123
NEO4J_DATABASE=neo4j
```

**ClickHouse Configuration**:
```bash
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=helixagent
CLICKHOUSE_USERNAME=default
CLICKHOUSE_PASSWORD=
```

**MinIO/S3 Configuration**:
```bash
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=helixagent
MINIO_SECRET_KEY=helixagent123
MINIO_BUCKET=helixagent-datalake
MINIO_USE_SSL=false
```

**Context Engine Configuration**:
```bash
CONTEXT_COMPRESSION_STRATEGY=hybrid
CONTEXT_COMPRESSION_RATIO=0.30
CONTEXT_MIN_QUALITY_SCORE=0.95
CONTEXT_CACHE_SIZE=100
CONTEXT_CACHE_TTL=1800  # 30 minutes
```

**Memory Sync Configuration**:
```bash
MEMORY_EVENT_SOURCING_ENABLED=true
MEMORY_CRDT_STRATEGY=merge_all
MEMORY_SNAPSHOT_INTERVAL=300  # 5 minutes
```

**Learning Configuration**:
```bash
LEARNING_ENABLED=true
LEARNING_MIN_CONFIDENCE=0.70
LEARNING_MIN_FREQUENCY=3
LEARNING_INSIGHT_THRESHOLD=0.80
```

### Configuration Files

**Development** (`configs/development.yaml`):
```yaml
bigdata:
  kafka:
    bootstrap_servers: "localhost:9092"
  neo4j:
    uri: "bolt://localhost:7687"
  clickhouse:
    host: "localhost"
    port: 9000
  minio:
    endpoint: "localhost:9000"
    use_ssl: false
```

**Production** (`configs/production.yaml`):
```yaml
bigdata:
  kafka:
    bootstrap_servers: "${KAFKA_BOOTSTRAP_SERVERS}"
  neo4j:
    uri: "${NEO4J_URI}"
  clickhouse:
    host: "${CLICKHOUSE_HOST}"
    port: ${CLICKHOUSE_PORT}
  minio:
    endpoint: "${MINIO_ENDPOINT}"
    use_ssl: ${MINIO_USE_SSL}
```

---

## Troubleshooting

### Kafka Connection Issues

**Problem**: "Failed to connect to Kafka broker"

**Solution**:
```bash
# Check if Kafka is running
docker-compose ps kafka

# Check Kafka logs
docker-compose logs kafka

# Test connection
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Restart Kafka
docker-compose restart kafka
```

### Neo4j Query Timeout

**Problem**: "Neo4j query took too long"

**Solution**:
```cypher
// Create missing indexes
CREATE INDEX ON :Entity(id);
CREATE INDEX ON :Entity(type);
CREATE INDEX ON :Entity(name);

// Check query performance
PROFILE MATCH (e:Entity {name: "Docker"})
RETURN e;
```

### Context Compression Quality Low

**Problem**: "Compression quality score < 0.95"

**Solution**:
```bash
# Try different strategy
export CONTEXT_COMPRESSION_STRATEGY=entity  # Instead of window

# Adjust compression ratio
export CONTEXT_COMPRESSION_RATIO=0.40  # Less aggressive compression

# Check ClickHouse metrics
curl http://localhost:7061/v1/analytics/compression/stats
```

### Data Lake Archive Failure

**Problem**: "Failed to upload to MinIO"

**Solution**:
```bash
# Check MinIO is running
curl http://localhost:9000/minio/health/live

# Test MinIO connection
mc alias set myminio http://localhost:9000 helixagent helixagent123
mc ls myminio/helixagent-datalake

# Check bucket exists
curl -X PUT http://localhost:9000/helixagent-datalake \
  -H "Authorization: AWS4-HMAC-SHA256 ..."
```

### High Memory Usage

**Problem**: "HelixAgent using too much memory"

**Solution**:
```bash
# Reduce cache sizes
export CONTEXT_CACHE_SIZE=50  # Reduce from 100
export LEARNING_INSIGHT_CACHE_SIZE=500  # Reduce from 1000

# Enable garbage collection
export GOGC=50  # More aggressive GC

# Restart HelixAgent
docker-compose restart helixagent
```

### Slow Query Performance

**Problem**: "ClickHouse queries taking >100ms"

**Solution**:
```sql
-- Check table sizes
SELECT
    table,
    formatReadableSize(sum(bytes)) AS size,
    count() AS parts
FROM system.parts
WHERE database = 'helixagent'
GROUP BY table;

-- Optimize tables
OPTIMIZE TABLE debate_metrics;
OPTIMIZE TABLE conversation_metrics;

-- Check materialized views
SHOW TABLES FROM helixagent LIKE '%_hourly';
```

---

## Performance Tuning

### Kafka Tuning

```bash
# Increase partitions for high-throughput topics
kafka-topics.sh --bootstrap-server localhost:9092 \
  --alter --topic helixagent.conversations \
  --partitions 24

# Increase consumer parallelism
export KAFKA_CONSUMER_THREADS=16
```

### Neo4j Tuning

```cypher
// Increase query memory
CALL dbms.setConfigValue('dbms.memory.heap.max_size', '4G');

// Enable query caching
CALL dbms.setConfigValue('dbms.query_cache_size', '1000');
```

### ClickHouse Tuning

```sql
-- Increase max threads
SET max_threads = 16;

-- Use sampling for large aggregations
SELECT ... FROM table SAMPLE 0.1;  -- Sample 10%

-- Materialize common queries
CREATE MATERIALIZED VIEW IF NOT EXISTS provider_stats_minutely
ENGINE = SummingMergeTree()
...
```

---

## Additional Resources

- **API Reference**: `/docs/api/BIG_DATA_API.md`
- **SQL Schemas**: `/sql/schema/`
- **Architecture Diagrams**: `/docs/diagrams/`
- **Challenge Scripts**: `/challenges/scripts/bigdata/`
- **Video Course**: See `VIDEO_COURSE_OUTLINE.md`

---

**Need Help?** File an issue at https://github.com/anthropics/helixagent/issues
