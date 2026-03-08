# Performance Cookbook

Practical recipes for optimizing HelixAgent performance. Each recipe describes a problem,
solution, code example, and guidance on when to apply it.

Last updated: 2026-03-08

---

## Table of Contents

1. [Lazy Loading with sync.Once](#1-lazy-loading-with-synconce)
2. [Connection Pooling](#2-connection-pooling)
3. [Atomic Operations for Metrics](#3-atomic-operations-for-metrics)
4. [Semaphore Patterns for Concurrency Limiting](#4-semaphore-patterns-for-concurrency-limiting)
5. [Circuit Breaker Configuration](#5-circuit-breaker-configuration)
6. [Caching Strategies](#6-caching-strategies)
7. [Streaming Optimization](#7-streaming-optimization)
8. [Database Query Optimization](#8-database-query-optimization)
9. [HTTP/3 Transport Optimization](#9-http3-transport-optimization)

---

## 1. Lazy Loading with sync.Once

### Problem

Initializing expensive resources (registries, connection pools, schema parsers) at
package load time slows startup, wastes memory when features are unused, and creates
ordering dependencies between init() functions.

### Solution

Use `sync.Once` to defer initialization to first access. The resource is created exactly
once, even under concurrent access.

### Code Example

```go
var (
    defaultRegistry     *ToolRegistry
    defaultRegistryOnce sync.Once
)

// GetDefaultToolRegistry returns the lazily-initialized global registry.
func GetDefaultToolRegistry() *ToolRegistry {
    defaultRegistryOnce.Do(func() {
        defaultRegistry = &ToolRegistry{
            tools: make(map[string]ToolDefinition),
        }
        defaultRegistry.registerBuiltins()
    })
    return defaultRegistry
}
```

### Codebase References

- `internal/tools/handler.go` -- `GetDefaultToolRegistry()` lazy tool registry
- `internal/graphql/schema.go` -- `schemaOnce` lazy GraphQL schema build
- `internal/observability/tracer.go` -- `tracerOnce` lazy tracer initialization
- `internal/observability/metrics.go` -- `metricsOnce` lazy metrics collector
- `internal/formatters/registry.go` -- `once` lazy formatter registry
- `internal/background/metrics.go` -- `globalMetricsOnce` lazy background metrics
- `internal/auth/oauth_credentials/` -- lazy global reader, refresher, CLI refresher

### When to Use

- Global singletons that may not be needed in every execution path
- Resources with expensive initialization (network connections, file parsing)
- Replacing `init()` functions that cause import side effects

### Pitfalls

- `sync.Once` cannot be reset. If initialization fails, the failed state is permanent.
  For retriable initialization, use double-checked locking with `sync.Mutex` instead
  (see `internal/http/pool.go` for the `GlobalPool` pattern).
- Do not hold the Once lock while performing operations that may deadlock.

---

## 2. Connection Pooling

### Problem

Creating a new HTTP connection or database connection for every request is expensive.
TLS handshakes, TCP setup, and authentication add latency per request.

### Solution

Maintain a pool of reusable connections. HelixAgent uses pooling at three layers:
HTTP client pool, database connection pool, and MCP client pool.

### Code Example -- HTTP Client Pool

```go
// internal/http/pool.go
type HTTPClientPool struct {
    clients    map[string]*http.Client
    mu         sync.RWMutex
    config     *PoolConfig
}

type PoolConfig struct {
    MaxIdleConns        int           // Default: 100
    MaxIdleConnsPerHost int           // Default: 10
    IdleConnTimeout     time.Duration // Default: 90s
    TLSHandshakeTimeout time.Duration // Default: 10s
}

// Get returns a client for the given base URL, creating one if needed.
func (p *HTTPClientPool) Get(baseURL string) *http.Client {
    p.mu.RLock()
    if client, ok := p.clients[baseURL]; ok {
        p.mu.RUnlock()
        return client
    }
    p.mu.RUnlock()

    p.mu.Lock()
    defer p.mu.Unlock()
    // Double-check after acquiring write lock.
    if client, ok := p.clients[baseURL]; ok {
        return client
    }
    client := p.createClient(baseURL)
    p.clients[baseURL] = client
    return client
}
```

### Code Example -- Database Connection Pool

```go
// Database module: pkg/pool
type PoolConfig struct {
    MaxOpenConns    int           // Maximum open connections
    MaxIdleConns    int           // Maximum idle connections
    ConnMaxLifetime time.Duration // Connection max lifetime
    ConnMaxIdleTime time.Duration // Connection max idle time
}

pool, _ := postgres.NewClient(postgres.Config{
    Host:            "localhost",
    Port:            5432,
    MaxOpenConns:    25,
    MaxIdleConns:    10,
    ConnMaxLifetime: 30 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
})
```

### When to Use

- Any service making repeated HTTP calls to the same host (LLM providers, MCP servers)
- Database-intensive workloads with concurrent query patterns
- Service-to-service communication within the cluster

### Configuration Guidelines

| Parameter | HTTP Pool | DB Pool | Guidance |
|-----------|-----------|---------|----------|
| Max idle connections | 100 | 10 | Set to expected concurrent request rate |
| Idle timeout | 90s | 5m | Keep below server-side timeout |
| Max lifetime | -- | 30m | Rotate connections for load balancer compatibility |
| Per-host limit | 10 | -- | Prevents overwhelming a single backend |

---

## 3. Atomic Operations for Metrics

### Problem

Using `sync.Mutex` to protect simple integer counters creates lock contention under
high-throughput paths such as request counting, message processing, and error tracking.

### Solution

Use `sync/atomic` operations for simple counters and gauges. Atomics are lock-free and
significantly faster than mutexes for single-value updates.

### Code Example

```go
type ProcessorMetrics struct {
    MessagesProcessed int64
    MessagesRetried   int64
    MessagesDiscarded int64
    ProcessingErrors  int64
}

// Increment in hot path -- no lock needed.
func (m *ProcessorMetrics) RecordProcessed() {
    atomic.AddInt64(&m.MessagesProcessed, 1)
}

// Read for reporting -- consistent snapshot not guaranteed across fields,
// but individual reads are atomic.
func (m *ProcessorMetrics) Snapshot() ProcessorMetrics {
    return ProcessorMetrics{
        MessagesProcessed: atomic.LoadInt64(&m.MessagesProcessed),
        MessagesRetried:   atomic.LoadInt64(&m.MessagesRetried),
        MessagesDiscarded: atomic.LoadInt64(&m.MessagesDiscarded),
        ProcessingErrors:  atomic.LoadInt64(&m.ProcessingErrors),
    }
}
```

### Codebase References

- `internal/messaging/replay/handler.go` -- atomic counters for replay progress
- `internal/messaging/dlq/processor.go` -- atomic DLQ processing metrics

### When to Use

- Simple counters, gauges, and flags in hot paths
- Metrics that are incremented far more often than read
- Single-value state flags (e.g., "is shutting down")

### When NOT to Use

- Protecting multiple related fields that must be updated atomically together
- Complex state transitions (use `sync.Mutex` or channels instead)
- Floating-point values (use `math.Float64bits`/`math.Float64frombits` with atomics)

---

## 4. Semaphore Patterns for Concurrency Limiting

### Problem

Unbounded goroutine creation for parallel tasks (LLM calls, formatter execution,
MCP requests) can exhaust system resources, trigger OOM kills, or overwhelm backends.

### Solution

Use a channel-based semaphore or the `Concurrency` module's weighted semaphore to
limit the number of concurrent operations.

### Code Example -- Channel Semaphore

```go
// Limit to maxConcurrent parallel LLM calls.
sem := make(chan struct{}, maxConcurrent)
var wg sync.WaitGroup

for _, provider := range providers {
    wg.Add(1)
    go func(p LLMProvider) {
        defer wg.Done()
        sem <- struct{}{}        // Acquire
        defer func() { <-sem }() // Release

        result, err := p.Complete(ctx, request)
        // ... handle result
    }(provider)
}
wg.Wait()
```

### Code Example -- Weighted Semaphore (Concurrency Module)

```go
import "digital.vasic.concurrency/pkg/semaphore"

sem := semaphore.New(4) // Max 4 concurrent

err := sem.Acquire(ctx, 1) // Acquire weight 1
if err != nil {
    return err // Context cancelled
}
defer sem.Release(1)

// ... perform bounded work
```

### Codebase References

- `internal/services/debate_performance_optimizer.go` -- semaphore limiting parallel LLM calls
- `internal/plugins/` -- `make(chan struct{})` stop channels for lifecycle control
- `Concurrency/pkg/semaphore/` -- Weighted semaphore implementation

### When to Use

- Parallel fan-out to multiple LLM providers (ensemble/debate)
- Batch formatting operations across multiple formatters
- Any operation that calls external services in parallel

### Resource Budget

Per project constitution, tests and challenges are limited to 30-40% of host resources.
Set `GOMAXPROCS=2` and limit parallel operations to 4-8 concurrent goroutines for
test workloads.

---

## 5. Circuit Breaker Configuration

### Problem

When an external service (LLM provider, database, MCP server) becomes unavailable,
continued retry attempts waste resources and increase latency for all requests.

### Solution

Use the circuit breaker pattern from the `Concurrency` module. The breaker tracks
failures and opens the circuit (fast-failing) when a threshold is exceeded.

### Code Example

```go
import "digital.vasic.concurrency/pkg/breaker"

cb := breaker.New(breaker.Config{
    MaxFailures:   5,                // Open after 5 consecutive failures
    Timeout:       30 * time.Second, // Stay open for 30s before half-open
    HalfOpenMax:   2,                // Allow 2 probe requests in half-open
})

result, err := cb.Execute(func() (interface{}, error) {
    return provider.Complete(ctx, request)
})
if err != nil {
    if errors.Is(err, breaker.ErrCircuitOpen) {
        // Fast-fail: circuit is open, provider is known-bad
        return fallbackProvider.Complete(ctx, request)
    }
    return nil, err
}
```

### Configuration Guidelines

| Parameter | Conservative | Aggressive | Guidance |
|-----------|-------------|------------|----------|
| Max failures | 5 | 3 | Lower for critical paths |
| Open timeout | 60s | 15s | Match expected recovery time |
| Half-open probes | 2 | 1 | More probes = faster recovery detection |

### When to Use

- All LLM provider HTTP calls
- Database connections in non-critical paths
- MCP server communication
- Any external dependency that may become transiently unavailable

---

## 6. Caching Strategies

### Problem

Repeated identical requests to LLM providers, embedding services, or databases waste
time and money. LLM API calls are particularly expensive.

### Solution

HelixAgent supports two caching layers: in-memory (L1) and Redis (L2), with configurable
TTL policies.

### Code Example -- Two-Level Cache

```go
import (
    "digital.vasic.cache/pkg/memory"
    "digital.vasic.cache/pkg/redis"
    "digital.vasic.cache/pkg/distributed"
)

l1 := memory.New(memory.Config{
    MaxEntries:     10000,
    EvictionPolicy: memory.LRU,
    DefaultTTL:     5 * time.Minute,
})

l2 := redis.NewClient(redis.Config{
    Addr:       "localhost:6379",
    DefaultTTL: 1 * time.Hour,
})

cache := distributed.NewTwoLevel(l1, l2)

// Set: writes to both L1 and L2
cache.Set(ctx, "prompt:hash:abc123", response, 30*time.Minute)

// Get: checks L1 first, falls back to L2, promotes to L1 on hit
result, err := cache.Get(ctx, "prompt:hash:abc123")
```

### Code Example -- Semantic Cache (GPT-Cache)

```go
import "digital.vasic.optimization/pkg/gptcache"

cache := gptcache.New(gptcache.WithSimilarityThreshold(0.95))

// Stores response keyed by embedding similarity, not exact match.
cache.Set(ctx, prompt, embedding, response)

// Returns cached response if a semantically similar prompt exists.
hit, err := cache.Get(ctx, prompt, embedding)
```

### TTL Guidelines

| Data Type | L1 TTL | L2 TTL | Rationale |
|-----------|--------|--------|-----------|
| LLM responses | 5 min | 1 hour | Responses may vary; short L1 for freshness |
| Model discovery | 15 min | 1 hour | Provider model lists change infrequently |
| Embedding vectors | 1 hour | 24 hours | Embeddings are deterministic for same input |
| Format results | 10 min | 30 min | Source code changes frequently |
| Health check status | 30 sec | 2 min | Must reflect current state |

### When to Use

- LLM completion responses for repeated or similar prompts
- Embedding generation for documents already processed
- Provider model list discovery results
- Format results for unchanged source files

---

## 7. Streaming Optimization

### Problem

Streaming LLM responses with suboptimal buffer sizes causes excessive system calls,
high CPU usage, and choppy output delivery to clients.

### Solution

Configure appropriate buffer sizes, implement backpressure, and use chunk merging to
reduce overhead.

### Code Example -- Buffered Streaming

```go
import "digital.vasic.optimization/pkg/streaming"

handler := streaming.NewBufferedHandler(streaming.Config{
    BufferSize:       4096,          // Accumulate tokens before flushing
    FlushInterval:    50 * time.Millisecond, // Max delay before flush
    MaxChunkSize:     8192,          // Maximum single chunk size
    BackpressureSize: 65536,         // Pause upstream at this buffer level
})

// Process stream with automatic buffering and backpressure.
err := handler.Process(ctx, upstreamReader, downstreamWriter)
```

### Code Example -- SSE with Heartbeat

```go
import "digital.vasic.streaming/pkg/sse"

broker := sse.NewBroker(sse.Config{
    HeartbeatInterval: 15 * time.Second,
    BufferSize:        256,  // Channel buffer per client
    WriteTimeout:      5 * time.Second,
})
```

### Configuration Guidelines

| Parameter | Value | Guidance |
|-----------|-------|----------|
| Buffer size | 4 KB | Balance latency vs. syscall overhead |
| Flush interval | 50 ms | Human-perceptible delay threshold |
| SSE heartbeat | 15 s | Keep connection alive through proxies |
| Channel buffer | 256 | Absorb bursts without blocking sender |
| Backpressure threshold | 64 KB | Pause upstream before OOM risk |

### When to Use

- All SSE streaming endpoints (`/v1/chat/completions` with `stream: true`)
- WebSocket connections for real-time updates
- gRPC streaming responses

---

## 8. Database Query Optimization

### Problem

Naive database queries with N+1 patterns, missing indexes, or unbounded result sets
degrade performance as data grows.

### Solution

Use batch operations, prepared statements, pagination, and the query builder for
type-safe queries.

### Code Example -- Batch Insert

```go
import "digital.vasic.database/pkg/postgres"

// Batch insert with a single round-trip.
tx, _ := db.Begin(ctx)
defer tx.Rollback(ctx)

stmt := `INSERT INTO debate_turns (session_id, round, agent_id, content, score)
         VALUES ($1, $2, $3, $4, $5)`

for _, turn := range turns {
    _, err := tx.Exec(ctx, stmt,
        turn.SessionID, turn.Round, turn.AgentID, turn.Content, turn.Score)
    if err != nil {
        return fmt.Errorf("insert turn: %w", err)
    }
}
return tx.Commit(ctx)
```

### Code Example -- Query Builder with Pagination

```go
import "digital.vasic.database/pkg/query"

q := query.Select("id", "content", "score", "created_at").
    From("debate_sessions").
    Where(query.Eq("status", "completed")).
    Where(query.Gte("score", 0.8)).
    OrderBy("created_at", query.Desc).
    Limit(50).
    Offset(page * 50)

sql, args := q.Build()
rows, err := db.Query(ctx, sql, args...)
```

### Guidelines

| Practice | Recommendation |
|----------|---------------|
| Batch size | 100-500 rows per batch insert |
| Connection pool | 10-25 connections for typical workloads |
| Query timeout | 30s for complex queries, 5s for simple lookups |
| Pagination | Always use LIMIT/OFFSET or cursor-based pagination |
| Indexes | Create indexes on all WHERE clause and JOIN columns |
| Prepared statements | Use parameterized queries ($1, $2) -- never string concatenation |

### When to Use

- Debate session persistence (batch turn inserts)
- Provider verification result storage
- Audit log queries with time-range filters
- Any query that could return unbounded results

---

## 9. HTTP/3 Transport Optimization

### Problem

HTTP/2 multiplexing helps but still suffers from head-of-line blocking at the TCP
layer. High-latency connections to LLM providers benefit from QUIC transport.

### Solution

HelixAgent uses HTTP/3 (QUIC) as the primary transport with automatic HTTP/2 fallback,
combined with Brotli compression for response bodies.

### Code Example

```go
import (
    "github.com/quic-go/quic-go/http3"
    "github.com/andybalholm/brotli"
)

// HTTP/3 round-tripper with fallback.
transport := &http3.RoundTripper{
    TLSClientConfig: tlsConfig,
}

client := &http.Client{
    Transport: transport,
}

// Request with Brotli accept-encoding.
req, _ := http.NewRequest("POST", endpoint, body)
req.Header.Set("Accept-Encoding", "br, gzip")
```

### Compression Priority

| Priority | Algorithm | Ratio | CPU Cost | Use Case |
|----------|-----------|-------|----------|----------|
| 1 | Brotli | Best | Higher | Default for all responses |
| 2 | gzip | Good | Lower | Fallback when Brotli unavailable |
| 3 | None | -- | None | Binary/pre-compressed content |

### When to Use

- All external HTTP communication (LLM providers, MCP servers)
- Internal service-to-service calls when crossing network boundaries
- Static asset serving from the API server

---

## General Performance Checklist

Before deploying or releasing, verify:

- [ ] All global singletons use `sync.Once` or lazy initialization
- [ ] HTTP client pool is configured with appropriate per-host limits
- [ ] Database connection pool matches expected concurrency
- [ ] Circuit breakers are configured for all external dependencies
- [ ] Caching is enabled for LLM responses and embedding vectors
- [ ] Streaming buffers are sized appropriately (4-8 KB)
- [ ] Batch database operations are used instead of row-by-row inserts
- [ ] `GOMAXPROCS` is set appropriately for the deployment environment
- [ ] Prometheus metrics are exposed for cache hit rates and pool utilization
- [ ] Resource limits are enforced for test and challenge execution (30-40% max)
