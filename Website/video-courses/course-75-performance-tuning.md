# Video Course 75: Performance Tuning and Profiling

## Course Overview

**Duration:** 3 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 65 (Lazy Loading Patterns), Course 69 (Concurrency Safety)

Master performance tuning for HelixAgent. This course covers lazy loading and initialization patterns, benchmark methodology, monitoring-driven optimization, semaphore and backpressure mechanisms, HTTP connection pool tuning, and pprof-based memory and CPU profiling for identifying and resolving bottlenecks.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Apply lazy loading and `sync.Once` initialization across services
2. Write reliable Go benchmarks and interpret results correctly
3. Use Prometheus metrics to identify performance bottlenecks
4. Implement semaphore limiting and backpressure for load management
5. Tune HTTP/3 connection pools for optimal throughput
6. Profile applications with pprof to find memory leaks and CPU hotspots

---

## Module 1: Lazy Loading Patterns (30 min)

### Video 1.1: sync.Once Initialization (15 min)

**Topics:**
- `sync.Once` guarantees one-time initialization even under concurrent access
- Pattern: deferred initialization of expensive resources (database pools, HTTP clients)
- Avoiding startup latency by initializing on first use
- Thread safety: all concurrent callers block until initialization completes

**Lazy Loading Pattern:**
```go
type ProviderRegistry struct {
    providers map[string]LLMProvider
    initOnce  sync.Once
    initErr   error
}

func (r *ProviderRegistry) Get(name string) (LLMProvider, error) {
    r.initOnce.Do(func() {
        r.initErr = r.loadProviders()
    })
    if r.initErr != nil {
        return nil, r.initErr
    }
    return r.providers[name], nil
}
```

### Video 1.2: Lazy Loading in HelixAgent (15 min)

**Topics:**
- Provider discovery: models loaded on first request, not at boot
- MCP adapter initialization: adapters connect when first accessed
- Formatter registry: formatters initialized lazily per language
- Validation challenge: `lazy_loading_validation_challenge.sh`
- Avoiding the thundering herd: `sync.Once` serializes initialization

**HelixAgent Lazy Loading Sites:**
```
Startup (fast):
  - HTTP server binds port
  - Router registered
  - Health endpoint ready

First Request (on-demand):
  - Provider models discovered
  - MCP adapters connected
  - Formatter processes started
  - Vector DB connections opened
```

---

## Module 2: Benchmark Methodology (30 min)

### Video 2.1: Writing Go Benchmarks (15 min)

**Topics:**
- `testing.B` and the `b.N` loop for stable measurements
- `b.ResetTimer()` to exclude setup from timing
- `b.ReportAllocs()` for allocation tracking
- Sub-benchmarks with `b.Run()` for parameterized testing
- Resource limits: `GOMAXPROCS=2`, `nice -n 19` for reproducible results

**Benchmark Example:**
```go
func BenchmarkEnsembleVoting(b *testing.B) {
    ensemble := setupEnsemble(b)
    responses := generateMockResponses(10)
    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _, err := ensemble.Vote(responses)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Video 2.2: Interpreting Benchmark Results (15 min)

**Topics:**
- Reading ns/op, B/op, allocs/op columns
- `benchstat` for statistical comparison between runs
- Avoiding common pitfalls: compiler optimization, GC interference
- Running benchmarks: `make test-bench` or `go test -bench=. -benchmem`
- Comparing before/after with `benchstat old.txt new.txt`

**Benchmark Output:**
```
BenchmarkEnsembleVoting-2    50000    24300 ns/op    4096 B/op    12 allocs/op
BenchmarkCacheGet-2         500000     2450 ns/op       0 B/op     0 allocs/op
BenchmarkProviderSelect-2   200000     6100 ns/op    1024 B/op     3 allocs/op
```

---

## Module 3: Monitoring-Driven Optimization (25 min)

### Video 3.1: Prometheus Metrics for Performance (15 min)

**Topics:**
- Key metrics: request latency (histogram), throughput (counter), error rate (counter)
- Provider-level metrics: per-provider response time and success rate
- Debate metrics: round count, consensus time, cache hit ratio
- Goroutine count: `runtime.NumGoroutine()` as a gauge
- Memory usage: `runtime.MemStats` exported as gauges

**Key Metrics:**
```
helixagent_request_duration_seconds{handler="/v1/chat/completions"}
helixagent_provider_latency_seconds{provider="deepseek"}
helixagent_debate_rounds_total{topology="mesh"}
helixagent_cache_hit_ratio
helixagent_goroutine_count
helixagent_memory_alloc_bytes
```

### Video 3.2: Identifying Bottlenecks from Dashboards (10 min)

**Topics:**
- P99 latency spikes: correlate with provider health and cache miss rates
- Goroutine growth: indicates leak or backlog
- Memory growth: indicates leak or unbounded caching
- Systematic approach: measure, hypothesize, profile, fix, verify
- Dashboard challenge: `monitoring_dashboard_challenge.sh`

---

## Module 4: Semaphore and Backpressure (25 min)

### Video 4.1: Semaphore Limiting (15 min)

**Topics:**
- Bounding concurrent operations to protect downstream services
- `semaphore.Weighted` from `golang.org/x/sync` for fine-grained control
- Channel-based semaphore for simpler use cases
- Debate performance optimizer: semaphore limits parallel LLM calls
- Configuring limits based on provider rate limits and host resources

**Semaphore Pattern:**
```go
type RateLimitedPool struct {
    sem chan struct{}
}

func NewRateLimitedPool(maxConcurrent int) *RateLimitedPool {
    return &RateLimitedPool{
        sem: make(chan struct{}, maxConcurrent),
    }
}

func (p *RateLimitedPool) Execute(ctx context.Context, fn func() error) error {
    select {
    case p.sem <- struct{}{}:
        defer func() { <-p.sem }()
        return fn()
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Video 4.2: Backpressure Strategies (10 min)

**Topics:**
- Rejecting requests when queue depth exceeds threshold (HTTP 429)
- Adaptive rate limiting: adjust limits based on downstream latency
- Circuit breakers as backpressure: open circuit rejects fast
- Background task queue depth monitoring via resource monitor
- Combining semaphore + circuit breaker + queue depth for layered protection

---

## Module 5: HTTP Pool Tuning (25 min)

### Video 5.1: Connection Pool Configuration (15 min)

**Topics:**
- `http.Transport` settings: MaxIdleConns, MaxIdleConnsPerHost, IdleConnTimeout
- HTTP/3 (QUIC) pool: `quic-go` transport with connection reuse
- Brotli compression: reducing payload size for faster transfers
- Keep-alive settings: balancing connection reuse vs resource consumption
- Per-provider pool tuning: high-traffic providers get larger pools

**Transport Configuration:**
```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    TLSHandshakeTimeout: 10 * time.Second,
    ResponseHeaderTimeout: 30 * time.Second,
}
```

### Video 5.2: HTTP/3 and Compression Tuning (10 min)

**Topics:**
- HTTP/3 (QUIC) advantages: 0-RTT connection setup, multiplexing, no head-of-line blocking
- Fallback chain: HTTP/3 primary, HTTP/2 fallback
- Brotli compression: 15-25% better than gzip for text payloads
- Compression negotiation: Accept-Encoding header handling
- Adapter layer: `internal/adapters/http/` centralizes pool management

---

## Module 6: pprof Profiling (25 min)

### Video 6.1: CPU Profiling (10 min)

**Topics:**
- Enabling pprof: `import _ "net/http/pprof"` and the debug endpoint
- Collecting CPU profiles: `go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30`
- Reading flame graphs: identifying hot functions
- `top` command in pprof: sorting by cumulative vs flat time
- Comparing profiles before and after optimization

**pprof Commands:**
```bash
# Collect 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Interactive mode
(pprof) top 10
(pprof) list functionName
(pprof) web  # Opens flame graph in browser
```

### Video 6.2: Memory Profiling and Leak Detection (15 min)

**Topics:**
- Heap profile: `go tool pprof http://localhost:6060/debug/pprof/heap`
- `alloc_space` vs `inuse_space`: total allocations vs current live memory
- Identifying leaks: growing `inuse_space` over time
- Common leak patterns: unclosed channels, forgotten goroutines, unbounded caches
- Challenge: `pprof_memory_profiling_challenge.sh`

**Memory Profile Analysis:**
```bash
# Collect heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Show allocations by function
(pprof) top 10 -inuse_space

# Compare two snapshots for leak detection
go tool pprof -base profile1.pb.gz profile2.pb.gz
(pprof) top 10 -inuse_space
```

**Common Leak Fix:**
```go
// BEFORE: goroutine leak (channel never drained)
ch := make(chan Result)
go func() { ch <- doWork() }()
// ch abandoned if caller returns early

// AFTER: context cancellation stops goroutine
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()
go func() {
    select {
    case ch <- doWork():
    case <-ctx.Done():
    }
}()
```

---

## Assessment

### Quiz (10 questions)

1. How does `sync.Once` prevent thundering herd during lazy initialization?
2. What does `b.ReportAllocs()` add to benchmark output?
3. Which Prometheus metric type is best for tracking request latency?
4. How does a channel-based semaphore limit concurrency?
5. What HTTP Transport setting controls max idle connections per host?
6. What is the difference between `alloc_space` and `inuse_space` in pprof?
7. What is HelixAgent's primary HTTP transport protocol?
8. How do you compare two pprof heap snapshots for leak detection?
9. What backpressure signal does HTTP 429 communicate?
10. What resource limits should benchmark runs enforce on HelixAgent's host?

### Practical Assessment

Optimize a slow HelixAgent operation:
1. Write a benchmark for the target operation and record baseline results
2. Collect Prometheus metrics showing the performance bottleneck
3. Run a CPU profile and identify the top 3 hot functions
4. Run a heap profile and verify no memory leaks
5. Apply an optimization (lazy loading, semaphore, or pool tuning)
6. Re-benchmark and compare with `benchstat`

Deliverables:
1. Baseline and optimized benchmark results with `benchstat` comparison
2. CPU profile flame graph highlighting the bottleneck
3. Heap profile showing stable memory after optimization
4. Code diff with the optimization applied
5. Prometheus dashboard screenshot showing improved latency

---

## Resources

- [Lazy Loading Challenge](../../challenges/scripts/lazy_loading_validation_challenge.sh)
- [pprof Profiling Challenge](../../challenges/scripts/pprof_memory_profiling_challenge.sh)
- [Monitoring Dashboard Challenge](../../challenges/scripts/monitoring_dashboard_challenge.sh)
- [Debate Performance Optimizer](../../internal/services/debate_performance_optimizer.go)
- [HTTP Adapter](../../internal/adapters/http/)
- [Course 65: Lazy Loading Patterns](course-65-lazy-loading-patterns.md)
- [Course 69: Concurrency Safety Patterns](course-69-concurrency-safety.md)
- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)
