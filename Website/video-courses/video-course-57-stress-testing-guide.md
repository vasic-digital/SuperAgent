# Video Course 57: Stress Testing Guide

## Course Overview

**Duration**: 2 hours
**Level**: Intermediate to Advanced
**Prerequisites**: Course 01-Fundamentals, Course 06-Testing, Course 38-Stress-Testing, familiarity with Go testing framework and concurrency

This course covers writing and executing stress tests for HelixAgent with strict resource limits. Topics include GOMAXPROCS tuning, process priority management, concurrent access pattern testing, goroutine leak detection, and memory pressure testing.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Write stress tests that respect the mandatory 30-40% host resource limit
2. Configure GOMAXPROCS, nice, and ionice for safe test execution
3. Test concurrent access patterns with controlled parallelism
4. Detect and diagnose goroutine leaks in long-running services
5. Apply memory pressure testing to identify allocation bottlenecks
6. Interpret stress test results and correlate with production behavior

---

## Module 1: Resource-Limited Test Execution (25 min)

### 1.1 Why Resource Limits Matter

**Video: The Mandate and Its Origins** (8 min)

- Host machines run mission-critical processes alongside tests
- Unrestricted stress tests have caused system crashes and forced resets
- Constitution rule: ALL test execution limited to 30-40% of host resources
- Resource limits are non-negotiable for every test type

### 1.2 GOMAXPROCS Configuration

**Video: Controlling CPU Usage** (8 min)

```bash
# Limit Go runtime to 2 OS threads
export GOMAXPROCS=2

# Run stress tests with explicit parallelism
go test -v -p 1 -count=1 ./tests/stress/...
```

- `GOMAXPROCS=2` limits the Go scheduler to 2 threads regardless of available cores
- `-p 1` runs test packages sequentially (no parallel package execution)
- `-count=1` prevents test caching to ensure real execution

### 1.3 Process Priority with nice and ionice

**Video: OS-Level Priority Management** (9 min)

```bash
# Set lowest CPU priority and idle-only I/O class
nice -n 19 ionice -c 3 go test -v -p 1 ./tests/stress/...
```

| Setting        | Value  | Effect                                      |
|----------------|--------|---------------------------------------------|
| `nice -n 19`   | Lowest | Test yields CPU to any other process        |
| `ionice -c 3`  | Idle   | Test I/O only when disk is otherwise idle   |
| `GOMAXPROCS=2` | 2      | Maximum 2 OS threads for Go runtime         |
| `-p 1`         | 1      | Sequential package execution                |

### Hands-On Lab 1

Configure and verify resource limits:

```bash
# Run a CPU-intensive test with full limits
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -v -p 1 -run TestStress -count=1 ./tests/stress/...

# Monitor resource usage during execution
# In a separate terminal:
top -p $(pgrep -f "go test")
```

---

## Module 2: Writing Stress Tests (30 min)

### 2.1 Stress Test Structure

**Video: Anatomy of a Stress Test** (10 min)

```go
func TestStress_ConcurrentCompletions(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping stress test in short mode")
    }

    const (
        numWorkers    = 10
        numRequests   = 100
        requestTimeout = 30 * time.Second
    )

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    sem := make(chan struct{}, numWorkers)
    var wg sync.WaitGroup
    var successCount, errorCount atomic.Int64

    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func(reqID int) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()

            reqCtx, reqCancel := context.WithTimeout(ctx, requestTimeout)
            defer reqCancel()

            err := makeCompletionRequest(reqCtx, reqID)
            if err != nil {
                errorCount.Add(1)
                t.Logf("request %d failed: %v", reqID, err)
            } else {
                successCount.Add(1)
            }
        }(i)
    }

    wg.Wait()
    successRate := float64(successCount.Load()) / float64(numRequests)
    t.Logf("success rate: %.2f%% (%d/%d)", successRate*100,
        successCount.Load(), numRequests)
    assert.Greater(t, successRate, 0.95, "success rate below threshold")
}
```

### 2.2 Naming Conventions

**Video: Test Organization** (5 min)

- Prefix: `TestStress_` for all stress tests
- Location: `tests/stress/` package
- Naming: `TestStress_<Component>_<Scenario>`
- Examples: `TestStress_ProviderRegistry_ConcurrentAccess`, `TestStress_Cache_HighThroughput`

### 2.3 Timeout and Cancellation

**Video: Graceful Termination** (8 min)

- Every stress test must have a top-level context timeout
- Per-request timeouts prevent individual hangs from blocking the suite
- `t.Deadline()` integration for Go test framework timeout awareness
- Cleanup functions registered with `t.Cleanup()` for resource release

### 2.4 Metrics Collection During Stress

**Video: Measuring Under Load** (7 min)

```go
type StressMetrics struct {
    latencies    []time.Duration
    mu           sync.Mutex
}

func (m *StressMetrics) Record(d time.Duration) {
    m.mu.Lock()
    m.latencies = append(m.latencies, d)
    m.mu.Unlock()
}

func (m *StressMetrics) Report(t *testing.T) {
    m.mu.Lock()
    defer m.mu.Unlock()
    sort.Slice(m.latencies, func(i, j int) bool {
        return m.latencies[i] < m.latencies[j]
    })
    n := len(m.latencies)
    t.Logf("p50: %v, p95: %v, p99: %v, max: %v",
        m.latencies[n/2], m.latencies[n*95/100],
        m.latencies[n*99/100], m.latencies[n-1])
}
```

### Hands-On Lab 2

Write a stress test for a concurrent map access pattern:

1. Create `TestStress_ConfigCache_ConcurrentReadWrite`
2. Spawn 20 readers and 5 writers hitting a shared config cache
3. Run for 30 seconds with proper resource limits
4. Collect and report latency percentiles
5. Verify no data races with `-race` flag

---

## Module 3: Concurrent Access Pattern Testing (25 min)

### 3.1 Read-Write Contention

**Video: Testing Shared State** (8 min)

- Test patterns: many-readers-one-writer, many-writers, read-modify-write
- Use `sync.RWMutex` contention as a baseline measurement
- Compare with lock-free alternatives (atomic, sync.Map)

### 3.2 Provider Registry Stress

**Video: Registry Under Load** (8 min)

- Concurrent provider registration and lookup
- Health check updates while queries are in flight
- Provider removal during active request routing
- Circuit breaker state transitions under concurrent failures

### 3.3 Cache Stampede Testing

**Video: Thundering Herd Prevention** (9 min)

```go
func TestStress_Cache_Stampede(t *testing.T) {
    cache := NewCache(WithTTL(1 * time.Second))
    var fetchCount atomic.Int64

    // Simulate cache stampede: TTL expires, 100 goroutines request simultaneously
    cache.Set("key", "value")
    time.Sleep(1100 * time.Millisecond) // Let TTL expire

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            cache.GetOrFetch("key", func() (string, error) {
                fetchCount.Add(1)
                time.Sleep(100 * time.Millisecond) // Simulate slow fetch
                return "refreshed", nil
            })
        }()
    }
    wg.Wait()

    // With singleflight, only 1 fetch should occur
    assert.LessOrEqual(t, fetchCount.Load(), int64(2),
        "cache stampede not prevented")
}
```

### Hands-On Lab 3

Test the provider registry under concurrent stress:

1. Write a test with 10 goroutines registering providers
2. Write 20 goroutines performing lookups concurrently
3. Write 5 goroutines updating health status
4. Run for 60 seconds and check for races or panics
5. Measure lookup latency degradation under contention

---

## Module 4: Goroutine Leak Detection (20 min)

### 4.1 Detecting Leaks

**Video: Finding Orphaned Goroutines** (10 min)

```go
func TestStress_NoGoroutineLeaks(t *testing.T) {
    before := runtime.NumGoroutine()

    // Execute the operation under test
    svc := NewService()
    svc.Start()
    // ... perform operations ...
    svc.Stop()

    // Allow goroutines to clean up
    time.Sleep(500 * time.Millisecond)

    after := runtime.NumGoroutine()
    leaked := after - before
    assert.LessOrEqual(t, leaked, 2,
        "goroutine leak detected: %d goroutines leaked", leaked)
}
```

### 4.2 Using goleak

**Video: Automated Leak Detection with goleak** (5 min)

```go
import "go.uber.org/goleak"

func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

- `goleak` checks for unexpected goroutines at test suite completion
- Exclude known background goroutines with `goleak.IgnoreTopFunction()`
- Integrates with `TestMain` for suite-level detection

### 4.3 Diagnosing Leak Sources

**Video: Tracing Goroutine Origins** (5 min)

```bash
# Dump all goroutine stacks
curl http://localhost:6060/debug/pprof/goroutine?debug=2 > goroutines.txt

# Count goroutines by creation site
grep "^goroutine" goroutines.txt | wc -l
```

- Group goroutines by stack trace to identify accumulating patterns
- Common sources: unclosed channels, missing context cancellation, abandoned timers

### Hands-On Lab 4

Detect and fix a goroutine leak:

1. Write a service that intentionally leaks goroutines (timer without stop)
2. Write a test that detects the leak using goroutine count comparison
3. Fix the leak by adding proper cleanup
4. Re-run the test to confirm the fix

---

## Module 5: Memory Pressure Testing (20 min)

### 5.1 Allocation Profiling Under Load

**Video: Finding Allocation Hotspots** (8 min)

```bash
# Run stress test with memory profiling
go test -v -run TestStress -memprofile mem.prof ./tests/stress/

# Analyze allocations
go tool pprof -alloc_space mem.prof
```

### 5.2 GC Pressure Testing

**Video: Garbage Collector Behavior Under Load** (7 min)

```bash
# Enable GC tracing during stress test
GODEBUG=gctrace=1 go test -v -run TestStress ./tests/stress/ 2>&1 | \
  grep "gc " | tail -20
```

- Monitor GC pause duration during high-throughput scenarios
- `GOGC` tuning to trade memory for lower GC frequency
- Verify no unbounded memory growth over extended runs

### 5.3 Memory Limit Enforcement

**Video: Capping Memory Usage** (5 min)

```go
func TestStress_MemoryBound(t *testing.T) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    baseAlloc := memStats.Alloc

    // Run heavy operation
    heavyOperation()

    runtime.ReadMemStats(&memStats)
    growth := memStats.Alloc - baseAlloc
    maxAllowed := uint64(100 * 1024 * 1024) // 100MB
    assert.Less(t, growth, maxAllowed,
        "memory growth %dMB exceeds limit", growth/1024/1024)
}
```

### Hands-On Lab 5

Write and execute a complete stress test:

1. Create `TestStress_Ensemble_MemoryPressure`
2. Send 500 ensemble requests with 3 providers
3. Enforce resource limits (GOMAXPROCS=2, nice -n 19)
4. Measure peak memory usage and GC pause times
5. Assert memory growth stays under a defined threshold
6. Report latency percentiles (p50, p95, p99)

---

## Course Summary

### Key Takeaways

1. Resource limits (GOMAXPROCS=2, nice -n 19, ionice -c 3, -p 1) are mandatory for all stress tests
2. Stress tests must have top-level timeouts, per-request timeouts, and proper cleanup
3. Concurrent access pattern tests expose race conditions and contention bottlenecks
4. Goroutine leak detection prevents long-term resource exhaustion in production
5. Memory pressure tests identify allocation hotspots and GC pressure issues
6. Always run stress tests with `-race` to catch data races

### Assessment Questions

1. Why is GOMAXPROCS=2 mandatory for stress tests in this project?
2. How does nice -n 19 differ from GOMAXPROCS in controlling resource usage?
3. Describe three techniques for detecting goroutine leaks.
4. What is a cache stampede and how is it prevented?
5. How would you measure GC pause impact during a stress test?

### Related Courses

- Course 06: Testing
- Course 38: Stress Testing
- Course 56: Performance Optimization
- Course 58: Chaos Engineering

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
