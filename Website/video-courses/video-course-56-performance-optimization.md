# Video Course 56: Performance Optimization

## Course Overview

**Duration**: 2 hours
**Level**: Intermediate to Advanced
**Prerequisites**: Course 01-Fundamentals, Course 24-Profiling, Course 25-Lazy-Loading, familiarity with Go concurrency primitives

This course covers performance optimization techniques used throughout HelixAgent, including lazy loading with `sync.Once`, atomic operations for metrics, connection pooling, semaphore-limited parallel execution, and circuit breaker patterns. Includes hands-on profiling and optimization of a hot path.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Apply lazy loading with `sync.Once` and double-checked locking patterns
2. Use atomic operations for high-throughput metrics collection
3. Configure and tune connection pools for HTTP, database, and MCP connections
4. Implement semaphore-limited parallel execution for concurrent LLM calls
5. Design circuit breaker configurations for external dependency protection
6. Profile a hot path using `pprof` and apply targeted optimizations

---

## Module 1: Lazy Loading Patterns (25 min)

### 1.1 sync.Once for Singleton Initialization

**Video: One-Time Initialization** (8 min)

```go
var (
    registry     *ToolRegistry
    registryOnce sync.Once
)

func GetDefaultToolRegistry() *ToolRegistry {
    registryOnce.Do(func() {
        registry = buildToolRegistry()
    })
    return registry
}
```

- `sync.Once` guarantees exactly one execution across all goroutines
- Zero contention after initialization completes
- Used in HelixAgent for tool registry, provider registry, and global pools

### 1.2 Double-Checked Locking

**Video: When sync.Once Is Not Enough** (8 min)

```go
type GlobalPool struct {
    mu       sync.Mutex
    pool     *http.Client
    initDone bool
}

func (g *GlobalPool) Get() *http.Client {
    if g.initDone {
        return g.pool
    }
    g.mu.Lock()
    defer g.mu.Unlock()
    if !g.initDone {
        g.pool = buildHTTPClient()
        g.initDone = true
    }
    return g.pool
}
```

- Fast path: single atomic read when already initialized
- Slow path: mutex-protected initialization runs once
- Used for `GlobalPool` in `internal/http/` (commit 94a2f9ae)

### 1.3 Lazy Registry Pattern

**Video: Converting Eager to Lazy Initialization** (9 min)

- Identify registries initialized at import time
- Wrap with getter function using `sync.Once`
- Measure startup time improvement with benchmarks
- Example: `DefaultToolRegistry` conversion (commit ad0f0f0d)

### Hands-On Lab 1

Convert an eager initialization to lazy loading:

1. Identify a package-level variable initialized at `init()` time
2. Wrap it with a `sync.Once`-based getter function
3. Benchmark startup time before and after
4. Verify thread safety with `-race` flag

---

## Module 2: Atomic Operations for Metrics (20 min)

### 2.1 Why Atomics Over Mutexes for Counters

**Video: Lock-Free Metrics** (8 min)

```go
type Metrics struct {
    requestCount  atomic.Int64
    errorCount    atomic.Int64
    totalLatencyNs atomic.Int64
}

func (m *Metrics) RecordRequest(latency time.Duration) {
    m.requestCount.Add(1)
    m.totalLatencyNs.Add(int64(latency))
}
```

- Atomic operations are 5-10x faster than mutex-protected increments
- No risk of deadlock or priority inversion
- Suitable for counters, gauges, and simple aggregates

### 2.2 Avoiding Metric Underflows

**Video: Safe Counter Subtraction** (6 min)

```go
// Unsafe: can underflow if concurrent decrements exceed current value
func (m *Metrics) Decrement() {
    m.gauge.Add(-1)
}

// Safe: compare-and-swap loop prevents underflow
func (m *Metrics) SafeDecrement() {
    for {
        current := m.gauge.Load()
        if current <= 0 {
            return
        }
        if m.gauge.CompareAndSwap(current, current-1) {
            return
        }
    }
}
```

- Fixed in commit 762c33fd for stress test memory metrics
- Always validate before subtraction when metrics represent non-negative quantities

### 2.3 Atomic vs sync.Map for Hot-Path Data

**Video: Choosing the Right Primitive** (6 min)

- `sync.Map` for read-heavy maps with infrequent writes
- `atomic.Value` for read-heavy configuration snapshots
- Plain atomics for counters and timestamps
- `sync.RWMutex` when write frequency is moderate

### Hands-On Lab 2

Replace a mutex-protected counter with atomic operations:

1. Identify a metrics counter guarded by `sync.Mutex`
2. Convert to `atomic.Int64`
3. Run benchmark comparison: mutex vs atomic
4. Verify correctness under race detector

---

## Module 3: Connection Pooling (25 min)

### 3.1 HTTP Client Pooling

**Video: Tuning the HTTP Transport** (8 min)

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    TLSHandshakeTimeout: 10 * time.Second,
    DisableKeepAlives:   false,
}
```

- Connection reuse eliminates TCP+TLS handshake overhead per request
- `MaxIdleConnsPerHost` should match expected concurrency per provider
- Monitor idle connection eviction with custom round-tripper metrics

### 3.2 Database Connection Pooling

**Video: PostgreSQL pgx Pool Configuration** (8 min)

```go
poolConfig, _ := pgxpool.ParseConfig(connString)
poolConfig.MaxConns = 20
poolConfig.MinConns = 5
poolConfig.MaxConnLifetime = 30 * time.Minute
poolConfig.MaxConnIdleTime = 5 * time.Minute
poolConfig.HealthCheckPeriod = 1 * time.Minute
```

- `MaxConns` limits total connections to prevent database saturation
- `MinConns` pre-warms the pool for burst handling
- `HealthCheckPeriod` detects and replaces stale connections

### 3.3 MCP Connection Pooling

**Video: MCP Adapter Connection Management** (9 min)

- 45+ MCP adapters each maintain a connection pool
- Idle connection timeouts prevent resource exhaustion
- Health checks verify MCP server availability
- Pool size scales based on adapter usage frequency

### Hands-On Lab 3

Tune HTTP connection pool settings for a provider:

1. Set up a benchmark that sends 100 concurrent requests
2. Measure latency with default pool settings
3. Adjust `MaxIdleConnsPerHost` and `IdleConnTimeout`
4. Re-run benchmark and compare p50, p95, p99 latencies

---

## Module 4: Semaphore-Limited Parallel Execution (20 min)

### 4.1 Semaphore Pattern for LLM Calls

**Video: Bounded Concurrency** (10 min)

```go
type ParallelExecutor struct {
    sem chan struct{}
}

func NewParallelExecutor(maxConcurrent int) *ParallelExecutor {
    return &ParallelExecutor{
        sem: make(chan struct{}, maxConcurrent),
    }
}

func (e *ParallelExecutor) Execute(ctx context.Context, fn func() error) error {
    select {
    case e.sem <- struct{}{}:
        defer func() { <-e.sem }()
        return fn()
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

- Prevents overwhelming LLM providers with too many simultaneous requests
- Context cancellation respects caller timeouts
- Used in debate performance optimizer for parallel LLM execution

### 4.2 Early Termination on Consensus

**Video: Short-Circuiting Parallel Work** (5 min)

- When debate participants reach consensus early, cancel remaining calls
- Saves provider API costs and reduces overall latency
- Implemented via context cancellation propagation

### 4.3 Fallback Chain with Semaphore

**Video: Smart Fallback Traversal** (5 min)

- Try providers in score-ranked order with semaphore limiting
- On failure, release semaphore slot and try next provider
- Track attempt statistics for optimization feedback

### Hands-On Lab 4

Implement and benchmark a semaphore-limited executor:

1. Create a parallel executor with configurable concurrency
2. Run 50 simulated LLM calls with max concurrency of 5
3. Measure total wall-clock time vs sequential execution
4. Add early termination and measure improvement

---

## Module 5: Circuit Breaker Patterns (15 min)

### 5.1 Circuit Breaker Configuration

**Video: Protecting External Dependencies** (8 min)

```go
type CircuitBreaker struct {
    failureThreshold int
    successThreshold int
    recoveryTimeout  time.Duration
    state            atomic.Int32 // 0=closed, 1=open, 2=half-open
    failures         atomic.Int32
    successes        atomic.Int32
}
```

- Closed: all requests pass through, failures counted
- Open: requests fail immediately, no external calls
- Half-Open: limited probe requests test recovery

### 5.2 Tuning for LLM Providers

**Video: Provider-Specific Settings** (7 min)

| Provider     | Failure Threshold | Recovery Timeout | Rationale                    |
|--------------|-------------------|------------------|------------------------------|
| Claude       | 3                 | 30s              | Low tolerance, fast recovery |
| DeepSeek     | 5                 | 60s              | Higher tolerance, slower API |
| Ollama       | 2                 | 15s              | Local, should recover fast   |
| OpenRouter   | 5                 | 45s              | Gateway, transient failures  |

### Hands-On Lab 5

Profile and optimize a hot path end to end:

```bash
# Enable pprof
export HELIX_PPROF_ENABLED=true

# Start HelixAgent
./bin/helixagent &

# Generate load
go test -bench=BenchmarkCompletionEndpoint -benchtime=30s ./tests/bench/

# Collect CPU profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze
go tool pprof -http=:8081 cpu.prof
```

1. Identify the top 3 CPU-consuming functions
2. Apply one optimization (lazy init, pool tuning, or atomic conversion)
3. Re-profile and measure improvement

---

## Course Summary

### Key Takeaways

1. `sync.Once` and double-checked locking eliminate repeated initialization overhead
2. Atomic operations provide 5-10x faster metrics collection than mutex-based approaches
3. Connection pool tuning (HTTP, DB, MCP) directly impacts request latency
4. Semaphore-limited parallel execution prevents provider saturation while maximizing throughput
5. Circuit breakers must be tuned per-provider based on expected failure patterns
6. Always profile before optimizing -- measure, change, re-measure

### Assessment Questions

1. When would you use double-checked locking instead of `sync.Once`?
2. How do you prevent metric underflows with atomic operations?
3. What is the impact of `MaxIdleConnsPerHost` on HTTP request latency?
4. How does the semaphore pattern interact with context cancellation?
5. Describe the three states of a circuit breaker and when transitions occur.

### Related Courses

- Course 24: Profiling
- Course 25: Lazy Loading
- Course 29: Optimization Techniques
- Course 30: Performance Tuning
- Course 57: Stress Testing Guide

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
