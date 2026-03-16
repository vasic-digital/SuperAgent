# Video Course 61: Goroutine Safety & Lifecycle Management

## Course Overview

**Duration:** 3 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 06 (Testing), Course 19 (Concurrency Patterns)

Master goroutine lifecycle management in Go with a focus on leak prevention, graceful shutdown, race condition detection, and the WaitGroup lifecycle pattern used throughout HelixAgent.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Identify common goroutine leak patterns in Go applications
2. Implement the WaitGroup lifecycle pattern for handlers with background goroutines
3. Apply context propagation best practices for cancellation and timeouts
4. Use the Go race detector to find and fix data races
5. Write tests that validate goroutine lifecycle correctness
6. Apply deduplication patterns for concurrent refresh operations

---

## Module 1: Common Goroutine Leak Patterns (30 min)

### Video 1.1: Bare go func() Without Context (15 min)

**Topics:**
- Why bare `go func()` is dangerous in production
- Goroutines that outlive their parent scope
- The relationship between goroutine count and memory usage
- Monitoring goroutine counts with `runtime.NumGoroutine()`

**Anti-Pattern:**
```go
// BAD: No way to stop this goroutine
func (h *Handler) startBackgroundWork() {
    go func() {
        for {
            time.Sleep(time.Second)
            h.doWork() // Runs forever, leaks on shutdown
        }
    }()
}
```

**Corrected Pattern:**
```go
// GOOD: Context-aware goroutine with lifecycle tracking
func (h *Handler) startBackgroundWork(ctx context.Context) {
    h.wg.Add(1)
    go func() {
        defer h.wg.Done()
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                h.doWork()
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

### Video 1.2: Channel Operations That Block Forever (10 min)

**Topics:**
- Unbuffered channel sends with no receiver
- Buffered channels that fill up without draining
- Missing `select` with context cancellation
- How to detect blocked goroutines with pprof

**Anti-Pattern:**
```go
// BAD: If no one reads from ch, this goroutine blocks forever
ch := make(chan Result)
go func() {
    result := expensiveComputation()
    ch <- result // Blocks if parent exits
}()
```

**Corrected Pattern:**
```go
// GOOD: Use select with context to prevent blocking
ch := make(chan Result, 1) // Buffer of 1
go func() {
    result := expensiveComputation()
    select {
    case ch <- result:
    case <-ctx.Done():
        return // Exit if parent cancelled
    }
}()
```

### Video 1.3: Missing WaitGroup Tracking (5 min)

**Topics:**
- Goroutines that finish after `main()` exits
- Orphaned goroutines during graceful shutdown
- Why `sync.WaitGroup` is essential for lifecycle management

**Key Insight:**
Without WaitGroup tracking, `Shutdown()` cannot guarantee all background work has completed. This can cause data loss, incomplete writes, and resource leaks.

---

## Module 2: The WaitGroup Lifecycle Pattern (45 min)

### Video 2.1: Pattern Overview — Add/Done/Wait (15 min)

**Topics:**
- The three operations: `Add(1)` before launch, `defer Done()` inside, `Wait()` on shutdown
- Why `Add(1)` must happen before `go func()`, not inside it
- The handler struct with embedded `sync.WaitGroup`
- Shutdown sequence: cancel context, then wait

**Core Pattern:**
```go
type SSEHandler struct {
    wg     sync.WaitGroup
    cancel context.CancelFunc
    ctx    context.Context
}

func NewSSEHandler() *SSEHandler {
    ctx, cancel := context.WithCancel(context.Background())
    return &SSEHandler{ctx: ctx, cancel: cancel}
}

func (h *SSEHandler) Start() {
    h.wg.Add(1)
    go func() {
        defer h.wg.Done()
        for {
            select {
            case event := <-h.events:
                h.broadcast(event)
            case <-h.ctx.Done():
                return
            }
        }
    }()
}

func (h *SSEHandler) Shutdown() {
    h.cancel()
    h.wg.Wait() // Blocks until goroutine exits
}
```

### Video 2.2: Handler Implementation Examples (15 min)

**Topics:**
- SSE streaming handler lifecycle
- Cache invalidation loop lifecycle
- Model refresh deduplication lifecycle
- Debate log tracking lifecycle

**Real-World Example (from HelixAgent):**
```go
// Cache invalidation with WaitGroup lifecycle
type CacheManager struct {
    wg     sync.WaitGroup
    cancel context.CancelFunc
    ctx    context.Context
    cache  *redis.Client
}

func (c *CacheManager) StartInvalidationLoop() {
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                c.invalidateExpired()
            case <-c.ctx.Done():
                return
            }
        }
    }()
}

func (c *CacheManager) Stop() {
    c.cancel()
    c.wg.Wait()
}
```

### Video 2.3: Testing Goroutine Count with runtime.NumGoroutine() (15 min)

**Topics:**
- Baseline goroutine count measurement
- Asserting no goroutine leaks after shutdown
- Using `require.Eventually` for async assertions
- The `goleak` library for automated leak detection

**Test Example:**
```go
func TestSSEHandler_NoGoroutineLeak(t *testing.T) {
    baseline := runtime.NumGoroutine()

    handler := NewSSEHandler()
    handler.Start()

    // Goroutine count should increase
    assert.Greater(t, runtime.NumGoroutine(), baseline)

    handler.Shutdown()

    // After shutdown, goroutine count should return to baseline
    require.Eventually(t, func() bool {
        return runtime.NumGoroutine() <= baseline+1
    }, 5*time.Second, 100*time.Millisecond)
}
```

---

## Module 3: Context Propagation Best Practices (30 min)

### Video 3.1: context.Background() vs Derived Contexts (10 min)

**Topics:**
- When to use `context.Background()`: only at the top-level entry point
- When to use `context.WithCancel()`: for goroutines you need to stop
- When to use `context.WithTimeout()`: for operations with deadlines
- Never store contexts in structs for request-scoped operations

**Guidelines:**
```go
// Top-level service initialization
ctx, cancel := context.WithCancel(context.Background())

// Request handler — derive from request context
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // Use request context, not Background
    result, err := h.service.Execute(ctx, req)
}

// Background task — derive with timeout
func (s *Service) StartBackgroundTask() {
    ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
    defer cancel()
    s.processQueue(ctx)
}
```

### Video 3.2: Timeout Contexts for Background Operations (10 min)

**Topics:**
- Setting appropriate timeouts for LLM calls (30-60 seconds)
- Timeout for database operations (5-10 seconds)
- Timeout for health checks (3-5 seconds)
- Cascading timeouts in nested calls

### Video 3.3: Deduplication with sync.Mutex (10 min)

**Topics:**
- The refresh deduplication pattern
- Preventing duplicate goroutines for the same operation
- `tryStartRefresh` / `finishRefresh` pattern
- Atomic flags for goroutine-safe state tracking

**Pattern:**
```go
type ModelRefresher struct {
    mu         sync.Mutex
    refreshing bool
    wg         sync.WaitGroup
}

func (r *ModelRefresher) TryStartRefresh() bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    if r.refreshing {
        return false // Already refreshing, skip
    }
    r.refreshing = true
    r.wg.Add(1)
    return true
}

func (r *ModelRefresher) FinishRefresh() {
    r.mu.Lock()
    r.refreshing = false
    r.mu.Unlock()
    r.wg.Done()
}

func (r *ModelRefresher) Refresh(ctx context.Context) {
    if !r.TryStartRefresh() {
        return // Deduplicated
    }
    go func() {
        defer r.FinishRefresh()
        r.doRefresh(ctx)
    }()
}
```

---

## Module 4: Race Condition Detection (30 min)

### Video 4.1: Using go test -race (10 min)

**Topics:**
- How the Go race detector works (ThreadSanitizer)
- Running tests with `-race` flag
- Interpreting race detector output
- Performance impact of race detection (~10x slower)

**Commands:**
```bash
# Run all tests with race detection
make test-race

# Run specific test with race detection
go test -race -v -run TestEnsembleParallel ./internal/services/...

# Run race detection with resource limits
GOMAXPROCS=2 nice -n 19 go test -race -v -p 1 ./internal/...
```

### Video 4.2: Common Race Patterns in Go (10 min)

**Topics:**
- Map concurrent read/write (most common)
- Slice append from multiple goroutines
- Struct field access without synchronization
- Counter increment without atomic operations

**Examples:**
```go
// RACE: concurrent map write
// Fix: use sync.Mutex or sync.Map
var m = make(map[string]int)
go func() { m["a"] = 1 }()
go func() { m["b"] = 2 }() // DATA RACE

// RACE: concurrent slice append
// Fix: use mutex around append, or channel
var s []int
go func() { s = append(s, 1) }()
go func() { s = append(s, 2) }() // DATA RACE

// RACE: non-atomic counter
// Fix: use atomic.AddInt64 or mutex
var count int64
go func() { count++ }()
go func() { count++ }() // DATA RACE
```

### Video 4.3: sync.Mutex vs sync.RWMutex (10 min)

**Topics:**
- When to use `Mutex` (read and write equally frequent)
- When to use `RWMutex` (reads far outnumber writes)
- Lock contention and performance implications
- Deadlock prevention: always `defer Unlock()`

**Benchmark comparison:**
```go
func BenchmarkMutex_ReadHeavy(b *testing.B) {
    var mu sync.Mutex
    data := make(map[string]int)
    data["key"] = 42

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            mu.Lock()
            _ = data["key"]
            mu.Unlock()
        }
    })
}

func BenchmarkRWMutex_ReadHeavy(b *testing.B) {
    var mu sync.RWMutex
    data := make(map[string]int)
    data["key"] = 42

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            mu.RLock()
            _ = data["key"]
            mu.RUnlock()
        }
    })
}
```

---

## Module 5: Hands-On Lab (45 min)

### Lab 1: Fix a Goroutine Leak in a Real Handler (15 min)

**Objective:** Given a handler with a bare `go func()` that leaks on shutdown, refactor it to use the WaitGroup lifecycle pattern.

**Steps:**
1. Identify the goroutine that lacks lifecycle tracking
2. Add a `sync.WaitGroup` field to the handler struct
3. Add `context.WithCancel` for shutdown signaling
4. Wrap the goroutine with `Add(1)` / `defer Done()` / `select ctx.Done()`
5. Implement a `Shutdown()` method that calls `cancel()` + `wg.Wait()`

### Lab 2: Add WaitGroup Lifecycle Tracking (15 min)

**Objective:** Add lifecycle tracking to a cache invalidation loop that runs in the background.

**Steps:**
1. Create a `CacheInvalidator` struct with `wg`, `ctx`, and `cancel` fields
2. Implement `Start()` that launches a ticker-based loop goroutine
3. Implement `Stop()` that cancels context and waits for completion
4. Write a test that verifies no goroutine leak after `Stop()`

### Lab 3: Run Race Detector and Fix Findings (15 min)

**Objective:** Run the race detector against a concurrent service and fix all findings.

**Steps:**
1. Run `go test -race -v ./path/to/package`
2. Analyze race detector output to identify shared state
3. Apply appropriate synchronization (mutex, atomic, or channel)
4. Re-run race detector to verify all races are resolved

---

## Resources

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [sync Package Documentation](https://pkg.go.dev/sync)
- [context Package Documentation](https://pkg.go.dev/context)
- [HelixAgent Concurrency Manual](../user-manuals/19-concurrency-patterns.md)
- [Goroutine Lifecycle Diagram](../../docs/diagrams/src/goroutine-lifecycle.puml)
- [goleak Library](https://github.com/uber-go/goleak)

---

## Course Completion

Congratulations! You have completed the Goroutine Safety & Lifecycle Management course. You should now be able to:

- Identify and fix goroutine leaks in production handlers
- Implement the WaitGroup lifecycle pattern (Add/Done/Wait)
- Apply context propagation and deduplication patterns
- Use the Go race detector to find and eliminate data races
- Write tests that validate goroutine lifecycle correctness
