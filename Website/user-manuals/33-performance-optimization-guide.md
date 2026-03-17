# User Manual 33: Performance Optimization Guide

**Version:** 1.0
**Last Updated:** March 17, 2026
**Audience:** Developers, Performance Engineers, System Architects

---

## Table of Contents

1. [Overview](#overview)
2. [Lazy Loading Patterns](#lazy-loading-patterns)
3. [sync.Once Initialization](#synconce-initialization)
4. [Semaphore Limiting](#semaphore-limiting)
5. [Ensemble Performance Optimization](#ensemble-performance-optimization)
6. [Memory Profiling with pprof](#memory-profiling-with-pprof)
7. [Benchmark Suite](#benchmark-suite)
8. [Performance Monitoring](#performance-monitoring)
9. [Optimization Checklist](#optimization-checklist)
10. [Troubleshooting](#troubleshooting)

---

## Overview

HelixAgent processes requests through 22 LLM providers, an ensemble debate system, and multiple middleware layers. Performance optimization is critical for maintaining low latency and high throughput while respecting the 30-40% host resource limit.

This guide covers the performance patterns used throughout HelixAgent:
- **Lazy loading** to defer expensive initialization until first use
- **sync.Once** for thread-safe one-time initialization
- **Semaphore limiting** for bounded concurrent LLM calls
- **pprof profiling** for memory leak detection
- **Benchmarking** for regression detection

---

## Lazy Loading Patterns

Lazy loading defers initialization of expensive resources (database connections, HTTP clients, provider registrations) until they are actually needed. This reduces startup time and avoids initializing resources that may never be used.

### The Problem with init()

Go's `init()` functions execute during package initialization, before `main()`. They run in dependency order and cannot be controlled:

```go
// BEFORE: Eager initialization in init()
// This runs at import time, even if the provider is never used
func init() {
    defaultClient = &http.Client{
        Timeout:   30 * time.Second,
        Transport: buildComplexTransport(), // Expensive!
    }
    registry.Register("provider", defaultClient) // Side effect!
}
```

**Problems with init():**
- Slows down application startup
- Cannot handle initialization errors gracefully
- Creates hidden dependencies between packages
- Makes testing difficult (cannot control initialization order)
- Wastes resources if the initialized component is never used

### The Lazy Loading Solution

Replace `init()` with lazy initialization using `sync.Once`:

```go
// AFTER: Lazy initialization with sync.Once
var (
    defaultClient     *http.Client
    defaultClientOnce sync.Once
)

func getClient() *http.Client {
    defaultClientOnce.Do(func() {
        defaultClient = &http.Client{
            Timeout:   30 * time.Second,
            Transport: buildComplexTransport(),
        }
    })
    return defaultClient
}
```

**Benefits:**
- Zero startup cost -- client is created on first use
- Thread-safe -- `sync.Once` guarantees exactly one execution
- Error handling possible -- can return errors from the initializer
- Testable -- can control when initialization happens

### Lazy Loading with Error Handling

For initialization that can fail:

```go
var (
    dbPool     *pgxpool.Pool
    dbPoolOnce sync.Once
    dbPoolErr  error
)

func getDBPool() (*pgxpool.Pool, error) {
    dbPoolOnce.Do(func() {
        config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
        if err != nil {
            dbPoolErr = fmt.Errorf("parse database config: %w", err)
            return
        }
        dbPool, dbPoolErr = pgxpool.NewWithConfig(context.Background(), config)
    })
    return dbPool, dbPoolErr
}
```

### Where HelixAgent Uses Lazy Loading

| Component | Before | After | Savings |
|-----------|--------|-------|---------|
| HTTP clients per provider | `init()` creates 22 clients | `sync.Once` per provider | ~500ms startup |
| Database connection pool | `init()` connects | Lazy on first query | ~200ms startup |
| Redis client | `init()` connects | Lazy on first cache access | ~100ms startup |
| Provider registry | `init()` registers all | Lazy registration | ~300ms startup |
| Model discovery cache | `init()` fetches all models | Lazy on first discovery | ~2s startup |
| Formatter registry | `init()` loads all formatters | Lazy per formatter | ~400ms startup |

---

## sync.Once Initialization

`sync.Once` is the foundation of thread-safe lazy loading in Go. It guarantees that a function executes exactly once, regardless of how many goroutines call it concurrently.

### Basic Pattern

```go
var (
    instance     *Service
    instanceOnce sync.Once
)

func GetService() *Service {
    instanceOnce.Do(func() {
        instance = &Service{
            client: http.DefaultClient,
            cache:  make(map[string]interface{}),
        }
    })
    return instance
}
```

### sync.Once Guarantees

1. **Exactly once**: The function passed to `Do()` executes at most once
2. **Happens-before**: All goroutines see the initialized value after `Do()` returns
3. **Blocking**: If goroutine A is inside `Do()`, goroutine B blocks until A finishes
4. **Panic safety**: If the function panics, `Do()` considers it executed (subsequent calls are no-ops)

### Common Mistakes

**Mistake 1: Reassigning the Once variable**

```go
// BAD: Resetting Once allows re-execution
var once sync.Once
once = sync.Once{} // This re-enables Do()!
```

**Mistake 2: Recursive Do() calls**

```go
// BAD: Deadlock -- Do() inside Do()
var once sync.Once
once.Do(func() {
    once.Do(func() { // DEADLOCK
        // ...
    })
})
```

**Mistake 3: Ignoring initialization errors**

```go
// BAD: No way to detect initialization failure
var once sync.Once
var client *Client
once.Do(func() {
    client, _ = NewClient() // Error silently ignored
})
// client might be nil!
```

### Testing sync.Once Components

```go
func TestLazyService_ConcurrentAccess(t *testing.T) {
    const goroutines = 100
    var wg sync.WaitGroup

    services := make([]*Service, goroutines)
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            services[idx] = GetService()
        }(i)
    }
    wg.Wait()

    // All goroutines must get the same instance
    for i := 1; i < goroutines; i++ {
        assert.Same(t, services[0], services[i],
            "all goroutines should receive the same instance")
    }
}
```

---

## Semaphore Limiting

HelixAgent's ensemble system calls multiple LLM providers in parallel. Without limiting, this can exhaust connections, trigger rate limits, and overload the host. Semaphore limiting bounds the number of concurrent operations.

### The Semaphore Pattern

```go
type EnsembleOrchestrator struct {
    semaphore chan struct{}
}

func NewEnsembleOrchestrator(maxConcurrent int) *EnsembleOrchestrator {
    return &EnsembleOrchestrator{
        semaphore: make(chan struct{}, maxConcurrent),
    }
}

func (e *EnsembleOrchestrator) CallProvider(ctx context.Context, provider Provider) (*Response, error) {
    // Acquire semaphore slot
    select {
    case e.semaphore <- struct{}{}:
        // Slot acquired
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    // Release slot when done
    defer func() { <-e.semaphore }()

    return provider.Complete(ctx)
}
```

### Weighted Semaphore

For providers with different resource costs:

```go
type WeightedSemaphore struct {
    sem *semaphore.Weighted
}

func NewWeightedSemaphore(maxWeight int64) *WeightedSemaphore {
    return &WeightedSemaphore{
        sem: semaphore.NewWeighted(maxWeight),
    }
}

func (w *WeightedSemaphore) Acquire(ctx context.Context, weight int64) error {
    return w.sem.Acquire(ctx, weight)
}

func (w *WeightedSemaphore) Release(weight int64) {
    w.sem.Release(weight)
}
```

### HelixAgent's Ensemble Semaphore

The ensemble orchestrator uses a semaphore to limit concurrent LLM calls during debate:

- **Default limit**: 5 concurrent provider calls
- **Configurable**: Via `ENSEMBLE_MAX_CONCURRENT` environment variable
- **Context-aware**: Respects request timeouts and cancellation
- **Metrics**: Tracks semaphore wait time in Prometheus

```go
// Configuration
maxConcurrent := 5
if v := os.Getenv("ENSEMBLE_MAX_CONCURRENT"); v != "" {
    maxConcurrent, _ = strconv.Atoi(v)
}

orchestrator := NewEnsembleOrchestrator(maxConcurrent)
```

### Early Termination on Consensus

The performance optimizer can terminate early when enough providers agree:

```go
func (o *Optimizer) ExecuteWithEarlyTermination(ctx context.Context, providers []Provider) (*Result, error) {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    results := make(chan *Response, len(providers))
    consensusThreshold := len(providers)/2 + 1

    for _, p := range providers {
        go func(provider Provider) {
            resp, err := o.orchestrator.CallProvider(ctx, provider)
            if err == nil {
                results <- resp
            }
        }(p)
    }

    var collected []*Response
    for resp := range results {
        collected = append(collected, resp)
        if checkConsensus(collected) >= consensusThreshold {
            cancel() // Stop remaining providers
            return aggregate(collected), nil
        }
    }
    return aggregate(collected), nil
}
```

---

## Ensemble Performance Optimization

The debate performance optimizer (`internal/services/debate_performance_optimizer.go`) implements several strategies:

### Parallel LLM Execution with Semaphore

All provider calls execute in parallel, bounded by the semaphore:

```
Provider 1: [==========]      (1.2s)
Provider 2: [========]        (1.0s)
Provider 3: [============]    (1.4s)  <- Semaphore slot freed for Provider 4
Provider 4:          [=======] (0.8s)
Provider 5:          [=====]   (0.6s)
            |--------|---------|
            0s       1s        2s

Total time: ~2s (not 5s serial)
```

### Response Caching with TTL

Identical prompts within the TTL window return cached responses:

```go
type ResponseCache struct {
    mu      sync.RWMutex
    entries map[string]*CacheEntry
    ttl     time.Duration
}

type CacheEntry struct {
    response  *Response
    createdAt time.Time
}
```

- Default TTL: 5 minutes
- Cache key: SHA256 of (model + prompt + temperature)
- Thread-safe with `sync.RWMutex`

### Smart Fallback Chain

When a provider fails, the optimizer traverses the fallback chain based on verification scores:

```
Primary (score: 8.5) -> Fail
  Fallback 1 (score: 8.2) -> Fail
    Fallback 2 (score: 7.9) -> Success
```

The chain skips providers with open circuit breakers, avoiding known-failing providers.

### Performance Statistics

The optimizer tracks:
- Total execution time per debate
- Per-provider latency (P50, P95, P99)
- Cache hit/miss ratio
- Semaphore wait time
- Early termination rate

---

## Memory Profiling with pprof

HelixAgent includes pprof memory leak tests that verify no memory growth under sustained load.

### Enabling pprof

pprof is available at runtime:

```go
import _ "net/http/pprof"

// Start pprof server on a separate port
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

### Analyzing Memory

```bash
# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Allocs profile (total allocations)
go tool pprof http://localhost:6060/debug/pprof/allocs
```

### Automated Leak Detection Tests

The pprof memory profiling tests compare memory before and after sustained load:

```go
func TestMemoryLeak_SustainedLoad(t *testing.T) {
    runtime.GC()
    var before runtime.MemStats
    runtime.ReadMemStats(&before)

    // Generate sustained load
    for i := 0; i < 10000; i++ {
        processRequest()
    }

    runtime.GC()
    var after runtime.MemStats
    runtime.ReadMemStats(&after)

    // Memory growth should be bounded
    growth := after.HeapAlloc - before.HeapAlloc
    assert.Less(t, growth, uint64(50*1024*1024),
        "heap growth should be less than 50MB after 10k requests")
}
```

### Common Memory Leaks

| Leak Pattern | Symptom | Fix |
|-------------|---------|-----|
| Goroutine leak | Goroutine count grows | WaitGroup lifecycle pattern |
| Map growth | Heap grows unbounded | TTL-based eviction |
| Slice retain | Large allocations not freed | Copy relevant data, release backing array |
| Channel buffer | Channels fill up | Drain channels on shutdown |
| Context leak | Context tree grows | Always call cancel() |

---

## Benchmark Suite

HelixAgent includes 14 performance benchmarks covering critical paths.

### Running Benchmarks

```bash
# All benchmarks
make test-bench

# Specific benchmark
go test -bench=BenchmarkEnsembleVoting -benchmem ./internal/services/...

# With CPU profiling
go test -bench=BenchmarkEnsembleVoting -cpuprofile=cpu.out ./internal/services/...
go tool pprof cpu.out

# Resource-limited
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -bench=. -benchmem -p 1 ./internal/...
```

### Available Benchmarks

| Benchmark | What It Measures | Target |
|-----------|-----------------|--------|
| BenchmarkEnsembleVoting_MajorityVote | Majority vote aggregation | < 1us |
| BenchmarkEnsembleVoting_ConfidenceWeighted | Weighted voting | < 5us |
| BenchmarkEnsembleVoting_BordaCount | Borda count tabulation | < 10us |
| BenchmarkProviderRegistry_Lookup | Provider lookup by name | < 100ns |
| BenchmarkProviderRegistry_GetAll | List all providers | < 500ns |
| BenchmarkCircuitBreaker_Allow | Circuit breaker check | < 50ns |
| BenchmarkResponseCache_Get | Cache lookup | < 200ns |
| BenchmarkResponseCache_Set | Cache insertion | < 500ns |
| BenchmarkJSONParsing_ChatRequest | Request deserialization | < 5us |
| BenchmarkJSONParsing_ToolSchema | Tool schema parsing | < 3us |
| BenchmarkModelIDParsing | Model ID format parsing | < 100ns |
| BenchmarkScoringWeights_Calculate | Score calculation | < 1us |
| BenchmarkSSEEncoder_Encode | SSE event encoding | < 500ns |
| BenchmarkSemaphore_AcquireRelease | Semaphore throughput | < 100ns |

### Comparing Benchmark Results

Use `benchstat` to compare benchmark runs across commits:

```bash
# Save baseline
go test -bench=. -count=10 ./internal/... > old.txt

# Make changes, then:
go test -bench=. -count=10 ./internal/... > new.txt

# Compare
benchstat old.txt new.txt
```

Output:
```
name                          old time/op  new time/op  delta
EnsembleVoting_MajorityVote   845ns        792ns        -6.27%
ProviderRegistry_Lookup       89.2ns       45.1ns       -49.44%
```

---

## Performance Monitoring

### Prometheus Metrics

HelixAgent exports performance metrics:

```
# Request latency histogram
helixagent_request_duration_seconds{endpoint,status}

# Provider latency
helixagent_provider_latency_seconds{provider}

# Semaphore wait time
helixagent_semaphore_wait_seconds{pool}

# Cache hit ratio
helixagent_cache_hits_total / helixagent_cache_requests_total

# Active goroutines
helixagent_goroutines

# Memory usage
helixagent_memory_heap_bytes
```

### Grafana Dashboard

The monitoring dashboard (`docker/grafana/dashboards/`) provides:
- Request latency P50/P95/P99
- Provider health heatmap
- Cache hit ratio over time
- Goroutine count trend
- Memory usage trend
- Semaphore utilization

---

## Optimization Checklist

Before deploying performance-sensitive changes:

- [ ] All `init()` functions converted to lazy loading where applicable
- [ ] `sync.Once` used for one-time initialization (no raw mutexes for init)
- [ ] Semaphore limits configured for concurrent operations
- [ ] pprof memory leak tests pass under sustained load
- [ ] Benchmarks show no regression (< 5% slowdown)
- [ ] Goroutine count returns to baseline after load
- [ ] Cache TTLs configured appropriately
- [ ] Circuit breakers configured for all external dependencies
- [ ] Resource limits enforced (GOMAXPROCS, nice, ionice)

---

## Troubleshooting

### High latency under load

**Symptom:** P99 latency spikes during peak traffic.

**Solutions:**
1. Check semaphore utilization -- increase limit if wait time is high
2. Verify cache hit ratio -- low ratio means cache TTL may be too short
3. Check provider health -- circuit breaker may be cycling
4. Profile with pprof to find hot paths

### Memory growth over time

**Symptom:** Heap usage grows continuously.

**Solutions:**
1. Run pprof heap profile: `go tool pprof http://localhost:6060/debug/pprof/heap`
2. Look for goroutine leaks: `go tool pprof http://localhost:6060/debug/pprof/goroutine`
3. Check map sizes in cache and registry components
4. Verify context cancel functions are called

### Benchmark regression

**Symptom:** Benchmark shows > 10% slowdown after a change.

**Solutions:**
1. Use `benchstat` to verify statistical significance
2. Profile the specific benchmark: `go test -bench=BenchmarkName -cpuprofile=cpu.out`
3. Check if the change added allocations: compare `-benchmem` output
4. Revert and A/B test to isolate the cause

---

## Related Resources

- [User Manual 18: Performance Monitoring](18-performance-monitoring.md) -- Prometheus and Grafana setup
- [User Manual 19: Concurrency Patterns](19-concurrency-patterns.md) -- Concurrency primitives
- [User Manual 20: Testing Strategies](20-testing-strategies.md) -- Benchmark test patterns
- [Video Course 65: Lazy Loading Patterns](../video-courses/video-course-65-lazy-loading-patterns.md) -- Video walkthrough
- [Go Performance Wiki](https://go.dev/wiki/Performance)
- [pprof Documentation](https://pkg.go.dev/net/http/pprof)

---

**Next Manual:** User Manual 34 - (Reserved)
