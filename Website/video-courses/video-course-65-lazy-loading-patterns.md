# Video Course 65: Lazy Loading Patterns & Performance Benchmarking

## Course Overview

**Duration:** 2.5 hours
**Level:** Intermediate to Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 19 (Concurrency Patterns), Course 61 (Goroutine Safety)

Master lazy initialization with `sync.Once`, semaphore-based concurrency limiting, and Go performance benchmarking. Learn the patterns HelixAgent uses to minimize startup time, bound resource usage, and detect performance regressions.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Convert `init()` functions to lazy `sync.Once` initialization
2. Implement semaphore limiting for bounded concurrent operations
3. Write effective Go benchmarks with `testing.B`
4. Use `benchstat` to detect performance regressions
5. Profile applications with pprof for memory and CPU analysis
6. Apply the performance optimization checklist to production code

---

## Module 1: The Problem with init() (25 min)

### Video 1.1: How Go Initialization Works (10 min)

**Topics:**
- Package initialization order in Go
- The `init()` function: when it runs, guarantees, and limitations
- Why init() creates hidden coupling between packages
- The startup time cost of eager initialization
- Real measurement: HelixAgent startup before and after lazy loading

**Before Lazy Loading:**
```
HelixAgent startup time: 3.2 seconds
  init() in providers/: 1.1s (22 HTTP clients created)
  init() in database/:  0.4s (connection pool opened)
  init() in cache/:     0.3s (Redis connection)
  init() in formatters/: 0.6s (32 formatters loaded)
  init() in discovery/: 0.8s (model catalog fetched)
```

**After Lazy Loading:**
```
HelixAgent startup time: 0.4 seconds
  All heavy initialization deferred to first use
```

### Video 1.2: Problems init() Causes (15 min)

**Topics:**
- **Startup overhead**: All providers initialized even if unused
- **Error swallowing**: `init()` cannot return errors
- **Test interference**: Tests import packages, triggering `init()` side effects
- **Circular dependencies**: `init()` order is fragile
- **Wasted resources**: Connections opened to services that may never be queried

**Anti-Pattern:**
```go
package providers

import "net/http"

var defaultClient *http.Client

func init() {
    // Creates HTTP client at import time
    // Cannot handle errors
    // Runs even in unit tests that don't need HTTP
    defaultClient = &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:       100,
            MaxConnsPerHost:    10,
            IdleConnTimeout:    90 * time.Second,
            TLSHandshakeTimeout: 10 * time.Second,
        },
        Timeout: 30 * time.Second,
    }
}
```

---

## Module 2: sync.Once Lazy Initialization (40 min)

### Video 2.1: The sync.Once Pattern (15 min)

**Topics:**
- `sync.Once` guarantees: exactly once, happens-before, blocking
- Basic pattern: variable + once + getter function
- Error handling with sync.Once
- Thread safety for concurrent access

**Core Pattern:**
```go
var (
    httpClient     *http.Client
    httpClientOnce sync.Once
)

func getHTTPClient() *http.Client {
    httpClientOnce.Do(func() {
        httpClient = &http.Client{
            Transport: &http.Transport{
                MaxIdleConns:    100,
                MaxConnsPerHost: 10,
            },
            Timeout: 30 * time.Second,
        }
    })
    return httpClient
}
```

### Video 2.2: Error-Aware Lazy Loading (10 min)

**Topics:**
- Capturing errors from initialization
- The dual-variable pattern (value + error)
- Retry-on-error vs fail-once semantics
- When to use each approach

**Error-Aware Pattern:**
```go
var (
    dbPool     *pgxpool.Pool
    dbPoolErr  error
    dbPoolOnce sync.Once
)

func getDBPool() (*pgxpool.Pool, error) {
    dbPoolOnce.Do(func() {
        connStr := os.Getenv("DATABASE_URL")
        config, err := pgxpool.ParseConfig(connStr)
        if err != nil {
            dbPoolErr = fmt.Errorf("parse config: %w", err)
            return
        }
        pool, err := pgxpool.NewWithConfig(context.Background(), config)
        if err != nil {
            dbPoolErr = fmt.Errorf("connect: %w", err)
            return
        }
        dbPool = pool
    })
    return dbPool, dbPoolErr
}
```

### Video 2.3: Converting init() to sync.Once (15 min)

**Topics:**
- Step-by-step conversion process
- Identifying all `init()` functions in a package
- Moving side effects into lazy getters
- Updating callers to use the getter
- Verifying with tests

**Conversion Steps:**
1. Identify the `init()` function and its side effects
2. Create a module-level `sync.Once` and the result variable
3. Move the `init()` body into a getter function wrapped in `once.Do()`
4. Replace all direct variable access with the getter call
5. Remove the `init()` function
6. Write a test verifying concurrent access returns the same instance
7. Write a benchmark comparing startup time

**Before:**
```go
var registry *ProviderRegistry

func init() {
    registry = NewProviderRegistry()
    registry.Register("deepseek", NewDeepSeekProvider())
    registry.Register("claude", NewClaudeProvider())
    // ... 20 more providers
}
```

**After:**
```go
var (
    registry     *ProviderRegistry
    registryOnce sync.Once
)

func getRegistry() *ProviderRegistry {
    registryOnce.Do(func() {
        registry = NewProviderRegistry()
        registry.Register("deepseek", NewDeepSeekProvider())
        registry.Register("claude", NewClaudeProvider())
    })
    return registry
}
```

---

## Module 3: Semaphore Limiting (30 min)

### Video 3.1: Why Limit Concurrency? (10 min)

**Topics:**
- Unbounded goroutines exhaust connections, memory, and file descriptors
- LLM providers enforce rate limits -- concurrent calls trigger 429 errors
- The host machine's 30-40% resource limit mandate
- Semaphore as a bounded concurrency primitive

**Problem Without Limiting:**
```go
// BAD: 22 providers called simultaneously
for _, provider := range providers {
    go func(p Provider) {
        results <- p.Complete(ctx, prompt) // 22 concurrent HTTP calls!
    }(provider)
}
```

### Video 3.2: Channel-Based Semaphore (10 min)

**Topics:**
- Using a buffered channel as a counting semaphore
- Acquire (send) and release (receive)
- Context-aware acquisition with `select`
- The HelixAgent ensemble semaphore implementation

**Implementation:**
```go
type Semaphore struct {
    slots chan struct{}
}

func NewSemaphore(max int) *Semaphore {
    return &Semaphore{slots: make(chan struct{}, max)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
    select {
    case s.slots <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (s *Semaphore) Release() {
    <-s.slots
}
```

### Video 3.3: Semaphore in the Ensemble Orchestrator (10 min)

**Topics:**
- How HelixAgent limits concurrent LLM calls to 5 (configurable)
- Context cancellation when consensus is reached early
- Metrics: semaphore wait time in Prometheus
- Upgrading from a simple mutex to a semaphore for higher throughput

**Ensemble Pattern:**
```go
orchestrator := NewEnsembleOrchestrator(5) // Max 5 concurrent calls

for _, provider := range selectedProviders {
    wg.Add(1)
    go func(p Provider) {
        defer wg.Done()

        // Wait for a semaphore slot
        if err := orchestrator.semaphore.Acquire(ctx); err != nil {
            return // Context cancelled
        }
        defer orchestrator.semaphore.Release()

        resp, err := p.Complete(ctx, prompt)
        if err == nil {
            results <- resp
        }
    }(provider)
}
```

---

## Module 4: Go Performance Benchmarking (30 min)

### Video 4.1: Writing Benchmarks with testing.B (10 min)

**Topics:**
- The `Benchmark` prefix convention
- `b.N` loop and auto-calibration
- `b.ResetTimer()` for setup exclusion
- `b.ReportAllocs()` for allocation tracking
- `b.RunParallel()` for concurrent benchmarks

**Basic Benchmark:**
```go
func BenchmarkProviderRegistry_Lookup(b *testing.B) {
    registry := setupRegistryWith100Providers()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        registry.Get("provider-50")
    }
}
```

**Parallel Benchmark:**
```go
func BenchmarkResponseCache_Get(b *testing.B) {
    cache := setupCacheWith1000Entries()
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            cache.Get("key-500")
        }
    })
}
```

### Video 4.2: Running and Interpreting Benchmarks (10 min)

**Topics:**
- `go test -bench=. -benchmem` output format
- Understanding ns/op, B/op, allocs/op
- Running with `-count` for statistical significance
- Resource-limited benchmark execution

**Example Output:**
```
BenchmarkProviderRegistry_Lookup-8    28504123    42.1 ns/op    0 B/op    0 allocs/op
BenchmarkResponseCache_Get-8          15234567    78.6 ns/op   16 B/op    1 allocs/op
BenchmarkEnsembleVoting_Majority-8     1456789   823 ns/op    512 B/op    8 allocs/op
```

Reading the output:
- `28504123`: Number of iterations (higher = more precise)
- `42.1 ns/op`: Time per operation (lower = faster)
- `0 B/op`: Bytes allocated per operation (lower = less GC pressure)
- `0 allocs/op`: Allocations per operation (lower = less GC pressure)

### Video 4.3: Comparing Benchmarks with benchstat (10 min)

**Topics:**
- Installing benchstat: `go install golang.org/x/perf/cmd/benchstat@latest`
- Capturing baseline and new results
- Statistical comparison with confidence intervals
- Detecting significant regressions

**Workflow:**
```bash
# Capture baseline (10 runs for statistical significance)
go test -bench=. -count=10 -benchmem ./internal/... > baseline.txt

# Make code changes...

# Capture new results
go test -bench=. -count=10 -benchmem ./internal/... > new.txt

# Compare
benchstat baseline.txt new.txt
```

**Output:**
```
name                        old time/op  new time/op  delta
ProviderRegistry/Lookup-8   42.1ns       21.3ns       -49.4% (p=0.000 n=10+10)
ResponseCache/Get-8         78.6ns       79.1ns        ~     (p=0.482 n=10+10)
EnsembleVoting/Majority-8    823ns        845ns        ~     (p=0.142 n=10+10)
```

`~` means no statistically significant difference. A percentage with p < 0.05 is significant.

---

## Module 5: pprof Memory and CPU Profiling (15 min)

### Video 5.1: Profiling in Practice (15 min)

**Topics:**
- Enabling pprof in HelixAgent
- Heap profiling: finding memory leaks
- CPU profiling: finding hot paths
- Goroutine profiling: finding leaks
- Flame graphs for visual analysis

**Commands:**
```bash
# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Goroutine dump
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Interactive commands in pprof
(pprof) top 10        # Top 10 by memory/CPU
(pprof) web           # Open flame graph in browser
(pprof) list funcName # Show source-level profile
```

**Automated Leak Detection:**
```go
func TestNoMemoryLeak_SustainedLoad(t *testing.T) {
    runtime.GC()
    var before runtime.MemStats
    runtime.ReadMemStats(&before)

    for i := 0; i < 10000; i++ {
        processRequest()
    }

    runtime.GC()
    var after runtime.MemStats
    runtime.ReadMemStats(&after)

    growth := after.HeapAlloc - before.HeapAlloc
    assert.Less(t, growth, uint64(50*1024*1024),
        "heap should not grow more than 50MB under sustained load")
}
```

---

## Module 6: Hands-On Labs (30 min)

### Lab 1: Convert init() to sync.Once (10 min)

**Objective:** Convert an `init()` function to lazy loading.

**Steps:**
1. Read the provided `init()` function
2. Create `sync.Once` and variable declarations
3. Write the getter function
4. Update all callers
5. Remove `init()`
6. Write a concurrent access test
7. Benchmark startup time before and after

### Lab 2: Add Semaphore Limiting to a Worker Pool (10 min)

**Objective:** Add bounded concurrency to an unbounded worker pool.

**Steps:**
1. Identify the goroutine launch site
2. Create a channel-based semaphore
3. Add acquire/release around the goroutine body
4. Add context-aware cancellation
5. Benchmark throughput with different semaphore sizes

### Lab 3: Write and Compare Benchmarks (10 min)

**Objective:** Write benchmarks, make an optimization, and prove the improvement.

**Steps:**
1. Write a benchmark for a provided function
2. Run 10 times and save as baseline
3. Apply an optimization (e.g., pre-allocate slice, use sync.Pool)
4. Run 10 times and save as new
5. Use benchstat to compare and verify improvement

---

## Assessment

### Quiz (10 questions)

1. What guarantees does `sync.Once` provide?
2. Why is `init()` problematic for application startup time?
3. How does a channel-based semaphore limit concurrency?
4. What does `b.N` represent in a Go benchmark?
5. What tool compares benchmark results statistically?
6. What is the default semaphore limit for HelixAgent's ensemble?
7. How do you enable pprof profiling in a Go application?
8. What does `B/op` mean in benchmark output?
9. How do you handle errors in a sync.Once initialization?
10. What is the resource limit for running benchmarks on the HelixAgent host?

### Practical Assessment

Optimize a provided service that currently:
- Uses `init()` for 5 resources
- Has no concurrency limits on 10 goroutines
- Has no benchmarks

Deliverables:
1. Convert all 5 `init()` to `sync.Once` with tests
2. Add semaphore limiting (max 3 concurrent)
3. Write 5 benchmarks covering key paths
4. Demonstrate improvement with benchstat
5. Run pprof to verify no memory leak under load

---

## Resources

- [User Manual 33: Performance Optimization Guide](../user-manuals/33-performance-optimization-guide.md)
- [User Manual 19: Concurrency Patterns](../user-manuals/19-concurrency-patterns.md)
- [sync Package Documentation](https://pkg.go.dev/sync)
- [testing.B Documentation](https://pkg.go.dev/testing#B)
- [benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Go Performance Wiki](https://go.dev/wiki/Performance)
