# Big Data Optimization Guide

**Version**: 1.0
**Last Updated**: 2026-01-30
**Target**: Production deployments handling 10K+ req/sec

---

## Table of Contents

1. [Overview](#overview)
2. [Kafka Optimization](#kafka-optimization)
3. [ClickHouse Optimization](#clickhouse-optimization)
4. [Neo4j Optimization](#neo4j-optimization)
5. [Context Compression Optimization](#context-compression-optimization)
6. [Memory Sync Optimization](#memory-sync-optimization)
7. [Batch Processing Optimization](#batch-processing-optimization)
8. [Monitoring & Profiling](#monitoring--profiling)
9. [Troubleshooting](#troubleshooting)

---

## Overview

This guide provides step-by-step instructions for optimizing each component of the HelixAgent big data system for production workloads.

### Performance Targets

| Component | Metric | Target | Acceptable | Poor |
|-----------|--------|--------|------------|------|
| **Kafka** | Throughput | >10K msg/sec | >5K msg/sec | <5K msg/sec |
| **Kafka** | Latency (p95) | <10ms | <50ms | >50ms |
| **ClickHouse** | Insert | >50K rows/sec | >20K rows/sec | <20K rows/sec |
| **ClickHouse** | Query (p95) | <50ms | <100ms | >100ms |
| **Neo4j** | Write | >5K nodes/sec | >2K nodes/sec | <2K nodes/sec |
| **Neo4j** | Query (p95) | <100ms | <200ms | >200ms |
| **Context Replay** | 10K messages | <5s | <10s | >10s |
| **Memory Sync** | Lag | <1s | <5s | >5s |

---

## Kafka Optimization

### 1. Partition Tuning

**Determine Optimal Partition Count**:

```bash
# Rule of thumb: (desired throughput / per-partition throughput)
# Per-partition throughput: ~10MB/sec or ~10K msg/sec

# For memory events (100K msg/sec target):
Partitions = 100000 / 10000 = 10 (use 12 for headroom)

# For conversations (high volume):
Partitions = 16+ (for 160K msg/sec)
```

**Rebalance Partitions**:

```bash
# Increase partitions for existing topic
kafka-topics.sh --bootstrap-server localhost:9092 \
  --alter --topic helixagent.memory.events \
  --partitions 12
```

### 2. Producer Optimization

**Key Settings** (`configs/bigdata_performance.yaml`):

```yaml
kafka:
  producer:
    compression_type: lz4              # Best balance (or zstd for higher compression)
    batch_size: 32768                  # 32KB batches (increase to 65536 for higher throughput)
    linger_ms: 10                      # Wait up to 10ms to batch messages
    buffer_memory: 67108864            # 64MB buffer (increase for bursty workloads)
    max_in_flight_requests: 5          # Pipeline 5 requests (increase to 10 for high bandwidth)
    acks: 1                            # Leader ack (use 'all' for critical data)
```

**Benchmark Command**:

```bash
./scripts/benchmark-bigdata.sh kafka
# Expected: >10K msg/sec, p95 latency <10ms
```

### 3. Consumer Optimization

**Key Settings**:

```yaml
kafka:
  consumer:
    fetch_min_bytes: 1024              # Fetch at least 1KB (increase to reduce requests)
    fetch_max_wait_ms: 500             # Wait up to 500ms (decrease for lower latency)
    max_partition_fetch_bytes: 1048576 # 1MB per partition (increase for large messages)
    max_poll_records: 500              # Process 500 records (tune based on processing speed)
```

**Consumer Group Scaling**:

```bash
# Scale consumers to match partition count
# For 12 partitions, run 12 consumers (1 per partition)

# Check consumer lag
kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group helixagent-memory-consumer

# If LAG > 10K, add more consumers or increase max_poll_records
```

### 4. Topic Configuration

**Optimize Retention**:

```bash
# Short retention for analytics (7 days)
kafka-configs.sh --bootstrap-server localhost:9092 \
  --alter --entity-type topics \
  --entity-name helixagent.analytics.providers \
  --add-config retention.ms=604800000

# Long retention for conversations (1 year)
kafka-configs.sh --bootstrap-server localhost:9092 \
  --alter --entity-type topics \
  --entity-name helixagent.conversations \
  --add-config retention.ms=31536000000
```

**Enable Compression**:

```bash
kafka-configs.sh --bootstrap-server localhost:9092 \
  --alter --entity-type topics \
  --entity-name helixagent.memory.events \
  --add-config compression.type=lz4
```

### 5. Broker Tuning

**OS-Level Tuning** (`/etc/sysctl.conf`):

```bash
# Increase network buffer sizes
net.core.rmem_max = 134217728          # 128MB
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 87380 134217728
net.ipv4.tcp_wmem = 4096 65536 134217728

# Increase file descriptors
fs.file-max = 100000

# Apply changes
sudo sysctl -p
```

**JVM Settings** (`docker-compose.bigdata.yml`):

```yaml
kafka:
  environment:
    KAFKA_HEAP_OPTS: "-Xms4g -Xmx4g"   # 4GB heap
    KAFKA_JVM_PERFORMANCE_OPTS: "-XX:+UseG1GC -XX:MaxGCPauseMillis=20"
```

---

## ClickHouse Optimization

### 1. Table Engine Configuration

**MergeTree Settings**:

```sql
CREATE TABLE conversation_metrics (
    timestamp DateTime,
    conversation_id String,
    message_count UInt32,
    total_tokens UInt64
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, conversation_id)
SETTINGS
    index_granularity = 8192,          -- 8K rows per index granule
    merge_max_block_size = 8192,       -- 8K rows per merge block
    max_parts_in_total = 1000,         -- Force merge after 1000 parts
    parts_to_throw_insert = 300;       -- Reject inserts after 300 parts
```

**Partitioning Strategy**:

```sql
-- Monthly partitions for high-volume tables
PARTITION BY toYYYYMM(timestamp)

-- Daily partitions for very high volume
PARTITION BY toYYYYMMDD(timestamp)

-- No partitioning for low-volume tables
-- (partitioning overhead not worth it)
```

### 2. Materialized Views

**Provider Performance Aggregates**:

```sql
-- Hourly aggregates (refresh every minute)
CREATE MATERIALIZED VIEW provider_metrics_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, provider)
POPULATE AS
SELECT
    toStartOfHour(timestamp) AS hour,
    provider,
    COUNT(*) AS total_requests,
    AVG(response_time_ms) AS avg_response_time,
    quantile(0.95)(response_time_ms) AS p95_response_time,
    SUM(tokens_used) AS total_tokens
FROM provider_metrics
GROUP BY hour, provider;
```

**Debate Aggregates**:

```sql
CREATE MATERIALIZED VIEW debate_metrics_daily
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (day, winner_provider)
POPULATE AS
SELECT
    toDate(timestamp) AS day,
    winner_provider,
    COUNT(*) AS total_debates,
    AVG(total_rounds) AS avg_rounds,
    AVG(confidence) AS avg_confidence
FROM debate_metrics
GROUP BY day, winner_provider;
```

### 3. Query Optimization

**Use Proper Indexes**:

```sql
-- Primary key ORDER BY should match common query patterns
ORDER BY (timestamp, provider)  -- Good for time-series queries filtered by provider

-- Add secondary indexes for other columns
ALTER TABLE provider_metrics ADD INDEX idx_model (model) TYPE minmax GRANULARITY 4;
```

**Optimize WHERE Clauses**:

```sql
-- GOOD: Uses primary key
SELECT * FROM provider_metrics
WHERE timestamp >= now() - INTERVAL 1 HOUR
AND provider = 'claude'
ORDER BY timestamp;

-- BAD: Full table scan
SELECT * FROM provider_metrics
WHERE response_time_ms > 1000;  -- Not in primary key

-- BETTER: Add secondary index
ALTER TABLE provider_metrics ADD INDEX idx_response_time (response_time_ms) TYPE minmax GRANULARITY 4;
```

**Use PREWHERE**:

```sql
-- Filter early with PREWHERE (evaluated before reading all columns)
SELECT provider, AVG(response_time_ms)
FROM provider_metrics
PREWHERE timestamp >= now() - INTERVAL 1 DAY
WHERE provider IN ('claude', 'gemini')
GROUP BY provider;
```

### 4. Insert Optimization

**Batch Inserts**:

```go
// Batch size: 10K-100K rows
const batchSize = 10000

batch := make([]ProviderMetrics, 0, batchSize)
for metric := range metricsChannel {
    batch = append(batch, metric)

    if len(batch) >= batchSize {
        insertBatch(batch)
        batch = batch[:0]
    }
}
```

**Async Inserts** (ClickHouse 21.11+):

```sql
SET async_insert = 1;
SET wait_for_async_insert = 0;

INSERT INTO provider_metrics VALUES (...);
-- Returns immediately, batched automatically
```

### 5. Memory & CPU Tuning

**Configuration** (`/etc/clickhouse-server/config.xml`):

```xml
<max_memory_usage>10737418240</max_memory_usage>  <!-- 10GB per query -->
<max_threads>8</max_threads>                       <!-- 8 threads per query -->
<max_concurrent_queries>100</max_concurrent_queries>

<!-- Increase cache sizes -->
<mark_cache_size>5368709120</mark_cache_size>      <!-- 5GB mark cache -->
<uncompressed_cache_size>8589934592</uncompressed_cache_size>  <!-- 8GB uncompressed cache -->
```

**Benchmark**:

```bash
./scripts/benchmark-bigdata.sh clickhouse
# Expected: >50K rows/sec insert, <50ms p95 query
```

---

## Neo4j Optimization

### 1. Memory Configuration

**Heap Size** (`docker-compose.bigdata.yml`):

```yaml
neo4j:
  environment:
    NEO4J_server_memory_heap_initial__size: 4G
    NEO4J_server_memory_heap_max__size: 6G
    NEO4J_server_memory_pagecache_size: 8G
```

**Formula**:
- Heap: 2-4GB (minimum), up to 12GB (maximum)
- Page Cache: As large as possible (ideally hold entire graph)
- Total: Heap + Page Cache < 90% of available RAM

### 2. Index Creation

**Entity Indexes**:

```cypher
// Index on Entity.id (most common lookup)
CREATE INDEX entity_id_idx FOR (e:Entity) ON (e.id);

// Index on Entity.type (filter by type)
CREATE INDEX entity_type_idx FOR (e:Entity) ON (e.type);

// Index on Entity.name (search by name)
CREATE INDEX entity_name_idx FOR (e:Entity) ON (e.name);

// Full-text search index
CREATE FULLTEXT INDEX entity_fulltext FOR (e:Entity) ON EACH [e.name, e.description];
```

**Relationship Indexes**:

```cypher
// Index on relationship types
CREATE INDEX rel_type_idx FOR ()-[r:RELATED_TO]-() ON (r.type);
```

**Verify Indexes**:

```cypher
SHOW INDEXES;
```

### 3. Query Optimization

**Use MATCH with Indexes**:

```cypher
// GOOD: Uses index on Entity.id
MATCH (e:Entity {id: $entity_id})
RETURN e;

// BAD: Full node scan
MATCH (e:Entity)
WHERE e.importance > 0.8
RETURN e;

// BETTER: Add index on importance
CREATE INDEX entity_importance_idx FOR (e:Entity) ON (e.importance);
```

**Limit Result Sets**:

```cypher
// Always use LIMIT for unbounded queries
MATCH (e:Entity)-[:RELATED_TO]->(related)
RETURN e, related
LIMIT 100;  -- Prevent returning millions of results
```

**Use PROFILE/EXPLAIN**:

```cypher
// Analyze query performance
PROFILE
MATCH (e:Entity {type: 'person'})-[:RELATED_TO]-(r)
RETURN e, r
LIMIT 10;

// Shows:
// - Node scans vs. index lookups
// - Rows processed at each step
// - Database hits
```

### 4. Batch Operations

**Batch Node Creation**:

```cypher
// Create nodes in batches of 1000
UNWIND $batch AS entity
MERGE (e:Entity {id: entity.id})
SET e.name = entity.name,
    e.type = entity.type,
    e.properties = entity.properties;
```

**Batch Relationship Creation**:

```cypher
// Create relationships in batches
UNWIND $batch AS rel
MATCH (source:Entity {id: rel.source_id})
MATCH (target:Entity {id: rel.target_id})
MERGE (source)-[r:RELATED_TO]->(target)
SET r.type = rel.type,
    r.strength = rel.strength;
```

### 5. Transaction Optimization

**Use Explicit Transactions**:

```go
// Single transaction for batch operations
tx, err := driver.NewSession().BeginTransaction()
defer tx.Close()

for _, entity := range entities {
    tx.Run(createEntityQuery, map[string]interface{}{
        "entity": entity,
    })
}

tx.Commit()
```

**Benchmark**:

```bash
./scripts/benchmark-bigdata.sh neo4j
# Expected: >5K nodes/sec write, <100ms p95 query
```

---

## Context Compression Optimization

### 1. Compression Strategies

**Strategy Selection**:

| Strategy | Use Case | Ratio | Quality | Speed |
|----------|----------|-------|---------|-------|
| **Window Summary** | Short conversations (<1K msg) | 0.40 | High | Fast |
| **Entity Graph** | Entity-heavy conversations | 0.35 | Medium | Medium |
| **Full** | Long conversations (>5K msg) | 0.25 | Low | Slow |
| **Hybrid** | General purpose | 0.30 | High | Medium |

**Configuration**:

```yaml
infinite_context:
  compression:
    default_strategy: hybrid
    target_ratio: 0.30              # Target 30% of original
    quality_threshold: 0.90         # Minimum 90% quality
    preserve_recent_count: 50       # Always keep last 50 messages
```

### 2. Cache Optimization

**LRU Cache Settings**:

```yaml
infinite_context:
  cache:
    size: 1000                      # Cache 1000 conversations
    ttl: 1800                       # 30 minutes TTL
    eviction_policy: lru
```

**Cache Hit Ratio Monitoring**:

```go
// Track cache performance
type CacheMetrics struct {
    Hits   int64
    Misses int64
}

func (m *CacheMetrics) HitRatio() float64 {
    total := m.Hits + m.Misses
    if total == 0 {
        return 0
    }
    return float64(m.Hits) / float64(total)
}

// Target: >80% cache hit ratio
```

### 3. Parallel Compression

**Concurrent Processing**:

```go
// Compress multiple conversations in parallel
const maxConcurrent = 10

semaphore := make(chan struct{}, maxConcurrent)
var wg sync.WaitGroup

for _, conversationID := range conversations {
    wg.Add(1)
    semaphore <- struct{}{} // Acquire

    go func(id string) {
        defer wg.Done()
        defer func() { <-semaphore }() // Release

        compressed, err := compressor.Compress(id)
        // ...
    }(conversationID)
}

wg.Wait()
```

### 4. Benchmark

**Test Compression Performance**:

```bash
./scripts/benchmark-bigdata.sh context

# Expected results:
# - 1K messages: <1s
# - 5K messages: <3s
# - 10K messages: <5s
# - Compression ratio: 0.25-0.35
# - Quality score: >0.90
```

---

## Memory Sync Optimization

### 1. Batch Synchronization

**Batch Settings**:

```yaml
distributed_memory:
  sync:
    batch_size: 100                 # Sync 100 memories per batch
    interval: 1s                    # Sync every second
    max_lag: 5s                     # Alert if lag > 5s
```

**Adaptive Batching**:

```go
// Adjust batch size based on lag
func (dm *DistributedMemory) adjustBatchSize(lag time.Duration) int {
    if lag > 10*time.Second {
        return 500  // Larger batches to catch up
    } else if lag > 5*time.Second {
        return 200
    }
    return 100      // Normal batch size
}
```

### 2. CRDT Conflict Resolution

**Strategy Selection**:

```yaml
distributed_memory:
  crdt:
    strategy: merge_all             # merge_all or last_write_wins
```

**Performance Comparison**:

| Strategy | Latency | Conflicts | Data Loss |
|----------|---------|-----------|-----------|
| **LastWriteWins** | Low | Low | Possible |
| **MergeAll** | Medium | None | None |

**Benchmark Conflict Resolution**:

```bash
# Test conflict resolution performance
# Create 100 concurrent updates to same memory

./scripts/test-memory-conflicts.sh
# Expected: <100ms resolution time, 0% data loss with MergeAll
```

### 3. Snapshot Optimization

**Snapshot Settings**:

```yaml
distributed_memory:
  snapshot:
    interval: 300s                  # Snapshot every 5 minutes
    retention_count: 12             # Keep last hour
    compression: true
```

**Snapshot Size Reduction**:

```go
// Only snapshot changed memories
type Snapshot struct {
    Timestamp time.Time
    ChangedMemories []Memory      // Only memories modified since last snapshot
    DeletedIDs      []string
}
```

---

## Batch Processing Optimization

### 1. Spark Configuration

**Executor Settings** (`docker-compose.bigdata.yml`):

```yaml
spark-worker:
  environment:
    SPARK_WORKER_CORES: 8           # 8 cores per worker
    SPARK_WORKER_MEMORY: 16g        # 16GB per worker
    SPARK_EXECUTOR_CORES: 4         # 4 cores per executor
    SPARK_EXECUTOR_MEMORY: 8g       # 8GB per executor
```

**Job Submission**:

```bash
spark-submit \
  --master spark://spark-master:7077 \
  --deploy-mode client \
  --executor-cores 4 \
  --executor-memory 8g \
  --num-executors 4 \
  --conf spark.sql.shuffle.partitions=200 \
  entity_extraction.py
```

### 2. Partition Tuning

**Optimal Partition Count**:

```python
# Rule: 2-4x number of cores
num_cores = 32  # 4 workers * 8 cores
partitions = num_cores * 3  # 96 partitions

df = spark.read.parquet("s3://bucket/conversations/")
df = df.repartition(partitions)
```

**Avoid Small Files**:

```python
# Coalesce to reduce number of output files
df.coalesce(10).write.parquet("output/")
```

### 3. Data Lake Optimization

**Parquet Settings**:

```python
df.write \
  .mode("append") \
  .option("compression", "snappy") \
  .option("parquet.block.size", 134217728) \  # 128MB blocks
  .partitionBy("year", "month", "day") \
  .parquet("s3://bucket/conversations/")
```

**Query Optimization**:

```sql
-- Partition pruning (only read relevant partitions)
SELECT * FROM conversations
WHERE year = 2026 AND month = 1 AND day = 30;

-- Predicate pushdown (filter at read time)
SELECT conversation_id, message_count
FROM conversations
WHERE message_count > 100;
```

---

## Monitoring & Profiling

### 1. Metrics Collection

**Prometheus Metrics**:

```yaml
# scrape_configs in prometheus.yml
- job_name: 'helixagent-bigdata'
  static_configs:
    - targets: ['localhost:7061']
  metrics_path: '/metrics'
  scrape_interval: 15s
```

**Key Metrics**:

```go
// Kafka metrics
kafka_producer_record_send_total
kafka_consumer_lag_records
kafka_producer_batch_size_avg

// ClickHouse metrics
clickhouse_query_duration_seconds
clickhouse_insert_rows_per_second
clickhouse_memory_usage_bytes

// Neo4j metrics
neo4j_transaction_committed_total
neo4j_page_cache_hit_ratio
neo4j_store_size_bytes

// Context replay metrics
context_replay_duration_seconds
context_compression_ratio
context_cache_hit_ratio

// Memory sync metrics
memory_sync_lag_seconds
memory_conflict_resolution_duration_seconds
```

### 2. Profiling

**Go pprof**:

```bash
# CPU profile
go tool pprof http://localhost:7061/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:7061/debug/pprof/heap

# Analyze
(pprof) top10
(pprof) list compressConversation
```

**ClickHouse Query Log**:

```sql
-- Enable query log
SET log_queries = 1;

-- View slow queries
SELECT
    query,
    query_duration_ms,
    read_rows,
    memory_usage
FROM system.query_log
WHERE query_duration_ms > 100
ORDER BY query_duration_ms DESC
LIMIT 10;
```

**Neo4j Query Log**:

```cypher
// Enable query logging in neo4j.conf
dbms.logs.query.enabled=true
dbms.logs.query.threshold=100ms

// View logs
tail -f /var/log/neo4j/query.log
```

---

## Troubleshooting

### Issue: High Kafka Consumer Lag

**Symptoms**:
- Consumer lag > 10K messages
- Increasing lag over time

**Diagnosis**:
```bash
kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group helixagent-memory-consumer

# Check LAG column
```

**Solutions**:
1. **Scale consumers**: Add more consumers (up to partition count)
2. **Increase max_poll_records**: Process more records per poll
3. **Optimize processing**: Profile consumer code, reduce processing time
4. **Increase partitions**: More parallelism

### Issue: Slow ClickHouse Queries

**Symptoms**:
- Query duration > 1s
- High memory usage

**Diagnosis**:
```sql
EXPLAIN indexes = 1
SELECT * FROM provider_metrics
WHERE timestamp >= now() - INTERVAL 1 DAY;

-- Check if index is used
```

**Solutions**:
1. **Add indexes**: On frequently filtered columns
2. **Use materialized views**: For common aggregations
3. **Optimize ORDER BY**: Match primary key order
4. **Use PREWHERE**: Filter early
5. **Increase memory**: max_memory_usage

### Issue: Neo4j Out of Memory

**Symptoms**:
- OutOfMemoryError in logs
- Slow query performance

**Diagnosis**:
```cypher
// Check memory usage
CALL dbms.queryJmx('org.neo4j:*') YIELD attributes;
```

**Solutions**:
1. **Increase heap**: NEO4J_server_memory_heap_max__size
2. **Increase page cache**: NEO4J_server_memory_pagecache_size
3. **Add indexes**: Reduce memory needed for scans
4. **Limit result sets**: Always use LIMIT in queries

### Issue: High Context Replay Latency

**Symptoms**:
- Replay taking >10s for 10K messages

**Diagnosis**:
```bash
./scripts/benchmark-bigdata.sh context
```

**Solutions**:
1. **Increase cache size**: Cache more conversations
2. **Optimize compression**: Use faster strategy (window vs. full)
3. **Parallel replay**: Process messages in parallel
4. **Reduce Kafka fetch latency**: Tune fetch_max_wait_ms

---

## Next Steps

After optimization:

1. **Run Benchmarks**: `./scripts/benchmark-bigdata.sh all`
2. **Monitor Metrics**: Set up Grafana dashboards
3. **Load Test**: Test with production-like workloads
4. **Tune Iteratively**: Monitor, identify bottlenecks, optimize, repeat

---

**Optimization Complete!** 🚀

The system should now handle production workloads with optimal performance.
