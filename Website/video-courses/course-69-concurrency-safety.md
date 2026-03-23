# Video Course 69: Concurrency Safety Patterns in Go

## Course Overview

**Duration:** 3 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 61 (Goroutine Safety), Course 65 (Lazy Loading Patterns)

Master advanced concurrency safety patterns in Go as applied in HelixAgent. This course covers `sync.Once` for idempotent shutdown, `atomic.Bool` for lock-free coordination flags, `sync.WaitGroup` goroutine lifecycle tracking, panic recovery in goroutines, and how to use the race detector for comprehensive testing.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Use `sync.Once` for idempotent shutdown and one-time initialization
2. Apply `atomic.Bool` for lock-free boolean flags shared across goroutines
3. Implement the full WaitGroup goroutine lifecycle pattern
4. Add panic recovery to goroutines to prevent process crashes
5. Run the race detector and fix all reported data races
6. Combine these patterns into production-ready concurrent services

---

## Module 1: sync.Once for Idempotent Shutdown (35 min)

### Video 1.1: The Double-Close Problem (10 min)

**Topics:**
- What happens when `Shutdown()` is called twice on a service
- Closing an already-closed channel panics in Go
- Calling `cancel()` multiple times is safe, but cleanup code may not be
- Resource release (database connections, file handles) must be exactly-once
- The cost of missing this: panics in production during graceful restarts

**Anti-Pattern:**
```go
type Service struct {
    stopCh chan struct{}
}

func (s *Service) Shutdown() {
    close(s.stopCh) // PANIC if called twice!
    s.db.Close()    // Double-close may corrupt
}
```

### Video 1.2: sync.Once Guarantees (10 min)

**Topics:**
- `sync.Once.Do(f)` guarantees `f` is called exactly once, even from many goroutines
- All concurrent callers block until the first invocation completes
- The happens-before relationship: side effects of `f` are visible to all callers
- It is safe to call `Do()` from multiple goroutines simultaneously
- Once `Do()` has been called, subsequent calls return immediately

### Video 1.3: Idempotent Shutdown Pattern (15 min)

**Topics:**
- Wrapping all shutdown logic in `sync.Once.Do()`
- The shutdown struct: cancel function, WaitGroup, cleanup resources
- Combining `context.WithCancel` and `sync.Once` for clean teardown
- Real examples from HelixAgent: SSE handler, cache manager, debate service

**Pattern:**
```go
type Service struct {
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
    shutdownMu sync.Once
    db         *pgxpool.Pool
    redis      *redis.Client
}

func (s *Service) Shutdown() {
    s.shutdownMu.Do(func() {
        // 1. Signal all goroutines to stop
        s.cancel()

        // 2. Wait for all goroutines to finish
        s.wg.Wait()

        // 3. Release resources (exactly once)
        if s.db != nil {
            s.db.Close()
        }
        if s.redis != nil {
            s.redis.Close()
        }
    })
}
```

**Test for Idempotency:**
```go
func TestService_Shutdown_Idempotent(t *testing.T) {
    svc := NewService()
    svc.Start()

    // Call Shutdown from multiple goroutines simultaneously
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            svc.Shutdown() // Must not panic
        }()
    }
    wg.Wait()
}
```

---

## Module 2: atomic.Bool for Lock-Free Flags (30 min)

### Video 2.1: Why Mutexes Are Overkill for Booleans (10 min)

**Topics:**
- The overhead of `sync.Mutex` for a single boolean check
- Lock contention under high concurrency (many readers, rare writer)
- CPU cache line bouncing with mutex-protected booleans
- When atomic operations provide better performance

**Overhead Comparison:**
```
Benchmark                   ns/op    allocs/op
BenchmarkMutexBool-8        45.2     0
BenchmarkAtomicBool-8        3.1     0
```

### Video 2.2: atomic.Bool Operations (10 min)

**Topics:**
- `Load()` -- read the value atomically
- `Store(val)` -- write the value atomically
- `Swap(val)` -- write and return the old value atomically
- `CompareAndSwap(old, new)` -- conditional write (CAS)
- These operations are lock-free and do not block other goroutines

**Usage:**
```go
var isShuttingDown atomic.Bool

// Writer (called once)
func (s *Service) Shutdown() {
    if isShuttingDown.CompareAndSwap(false, true) {
        // First caller wins, perform shutdown
        s.doShutdown()
    }
    // Subsequent callers: no-op
}

// Reader (called from many goroutines)
func (s *Service) HandleRequest(ctx context.Context, req *Request) {
    if isShuttingDown.Load() {
        return // Reject new requests during shutdown
    }
    s.processRequest(ctx, req)
}
```

### Video 2.3: Common Use Cases in HelixAgent (10 min)

**Topics:**
- Shutdown flag: reject new work during graceful shutdown
- Initialization flag: track whether a service has been initialized
- Health status: atomically set healthy/unhealthy
- Refresh-in-progress flag: prevent duplicate refresh goroutines
- Circuit breaker open/closed flag

**Refresh Deduplication:**
```go
type ModelRefresher struct {
    refreshing atomic.Bool
    wg         sync.WaitGroup
}

func (r *ModelRefresher) TryRefresh(ctx context.Context) {
    // CompareAndSwap: only one goroutine enters
    if !r.refreshing.CompareAndSwap(false, true) {
        return // Already refreshing
    }

    r.wg.Add(1)
    go func() {
        defer r.wg.Done()
        defer r.refreshing.Store(false)
        r.doRefresh(ctx)
    }()
}
```

---

## Module 3: WaitGroup Goroutine Lifecycle Tracking (35 min)

### Video 3.1: The Complete Lifecycle Pattern (15 min)

**Topics:**
- Three rules: `Add(1)` before the goroutine, `defer Done()` inside, `Wait()` on shutdown
- Why `Add(1)` must happen before `go func()`, not inside it
- The sequence: Add, launch goroutine, goroutine does work, context cancelled, goroutine exits, Done, Wait returns
- Combining WaitGroup with context.WithCancel for two-phase shutdown

**Complete Pattern:**
```go
type Handler struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func NewHandler() *Handler {
    ctx, cancel := context.WithCancel(context.Background())
    return &Handler{ctx: ctx, cancel: cancel}
}

func (h *Handler) StartWorker(name string) {
    h.wg.Add(1) // BEFORE go func()
    go func() {
        defer h.wg.Done() // ALWAYS defer
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                h.doWork(name)
            case <-h.ctx.Done():
                return // Clean exit
            }
        }
    }()
}

func (h *Handler) Shutdown() {
    h.cancel()  // Phase 1: Signal goroutines to stop
    h.wg.Wait() // Phase 2: Wait for all to finish
}
```

### Video 3.2: Multiple Workers and Shutdown Ordering (10 min)

**Topics:**
- Starting multiple workers with the same WaitGroup
- All workers stop when the shared context is cancelled
- `Wait()` blocks until ALL workers have called `Done()`
- Ordering: cancel first, then wait -- never the reverse
- Timeout on shutdown to prevent hanging forever

**Timeout Pattern:**
```go
func (h *Handler) ShutdownWithTimeout(timeout time.Duration) error {
    h.cancel()

    done := make(chan struct{})
    go func() {
        h.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return nil // All goroutines exited cleanly
    case <-time.After(timeout):
        return fmt.Errorf("shutdown timed out after %v", timeout)
    }
}
```

### Video 3.3: Testing Goroutine Lifecycle (10 min)

**Topics:**
- Measure `runtime.NumGoroutine()` before and after
- Use `require.Eventually` for async assertions (goroutines may take a moment to exit)
- The `goleak` library for automated leak detection
- Testing concurrent Start/Shutdown sequences

**Test:**
```go
func TestHandler_NoGoroutineLeak(t *testing.T) {
    baseline := runtime.NumGoroutine()

    h := NewHandler()
    h.StartWorker("worker-1")
    h.StartWorker("worker-2")
    h.StartWorker("worker-3")

    assert.GreaterOrEqual(t, runtime.NumGoroutine(), baseline+3)

    h.Shutdown()

    require.Eventually(t, func() bool {
        return runtime.NumGoroutine() <= baseline+1
    }, 5*time.Second, 100*time.Millisecond,
        "goroutine leak detected after shutdown")
}
```

---

## Module 4: Panic Recovery in Goroutines (30 min)

### Video 4.1: Why Goroutine Panics Are Dangerous (10 min)

**Topics:**
- A panic in any goroutine crashes the entire process
- Unlike thread-based systems, Go has no per-goroutine exception isolation
- Common panic sources: nil pointer dereference, index out of range, closed channel send
- In HelixAgent, a provider returning unexpected data must not crash the server

### Video 4.2: The Recovery Pattern (10 min)

**Topics:**
- `recover()` only works inside a deferred function
- Place `defer recoverPanic()` at the top of every goroutine body
- Log the panic with full stack trace for debugging
- Convert the panic into a handled error

**Pattern:**
```go
func safeGo(f func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Errorf("goroutine panic recovered: %v\n%s",
                    r, debug.Stack())
            }
        }()
        f()
    }()
}

// Usage
safeGo(func() {
    result := provider.Complete(ctx, prompt)
    results <- result
})
```

### Video 4.3: Recovery with WaitGroup Integration (10 min)

**Topics:**
- Combining panic recovery with `defer wg.Done()`
- The deferred function order: recovery first, then Done
- Ensuring WaitGroup is decremented even after a panic
- Propagating panic information through error channels

**Combined Pattern:**
```go
func (h *Handler) StartSafeWorker(name string) {
    h.wg.Add(1)
    go func() {
        defer h.wg.Done() // Always decremented, even on panic
        defer func() {
            if r := recover(); r != nil {
                h.logger.Errorf("worker %s panicked: %v\n%s",
                    name, r, debug.Stack())
            }
        }()

        // Actual work
        for {
            select {
            case task := <-h.tasks:
                h.processTask(task) // May panic
            case <-h.ctx.Done():
                return
            }
        }
    }()
}
```

---

## Module 5: Race Detector and Testing (30 min)

### Video 5.1: Running the Race Detector (10 min)

**Topics:**
- `go test -race` enables ThreadSanitizer-based race detection
- Performance impact: approximately 10x slower, 5-10x more memory
- The race detector is not a static analysis tool -- it requires execution
- Resource-limited execution for HelixAgent's host environment

**Commands:**
```bash
# Run all tests with race detection
make test-race

# Run specific package with race detection
GOMAXPROCS=2 nice -n 19 go test -race -v -p 1 ./internal/services/...

# Run a single test with race detection
go test -race -v -run TestEnsembleParallel ./internal/llm/...
```

### Video 5.2: Reading Race Detector Output (10 min)

**Topics:**
- Two goroutine stacks: the "previous write" and the "read"
- File, line number, and function for each access
- How to trace the shared variable from the stack
- Common race patterns: map read/write, struct field, slice append, counter

**Example Output:**
```
WARNING: DATA RACE
Read at 0x00c000123456 by goroutine 7:
  mypackage.(*Service).GetStatus()
      /path/to/service.go:42 +0x64

Previous write at 0x00c000123456 by goroutine 12:
  mypackage.(*Service).UpdateStatus()
      /path/to/service.go:58 +0x70

Goroutine 7 (running) created at:
  mypackage.TestConcurrentAccess()
      /path/to/service_test.go:15 +0x130
```

### Video 5.3: Fixing Common Race Conditions (10 min)

**Topics:**
- Map access: use `sync.RWMutex` or `sync.Map`
- Struct field access: protect with mutex, or use atomic types
- Counter increment: use `atomic.AddInt64`
- Slice append: protect with mutex, or use channel-based collection
- Boolean flag: use `atomic.Bool`

**Fix Examples:**
```go
// RACE: concurrent map access
// Fix: sync.RWMutex
type Cache struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

func (c *Cache) Get(key string) interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

func (c *Cache) Set(key string, val interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = val
}

// RACE: concurrent counter
// Fix: atomic.Int64
var requestCount atomic.Int64

func handleRequest() {
    requestCount.Add(1) // Lock-free, race-free
}
```

---

## Module 6: Putting It All Together (15 min)

### Video 6.1: Production Concurrency Checklist (15 min)

**The HelixAgent Concurrency Safety Checklist:**

1. **Every service with goroutines** must have a `sync.WaitGroup` for lifecycle tracking
2. **Every Shutdown method** must use `sync.Once` for idempotent cleanup
3. **Every boolean flag** shared across goroutines must use `atomic.Bool`
4. **Every goroutine** must have `defer recover()` to prevent process crashes
5. **Every shared data structure** (map, slice, struct) must be protected with `sync.RWMutex`
6. **Every context** must be propagated to goroutines for cancellation
7. **Every ticker/timer** must be stopped in deferred cleanup
8. **All tests** must pass `go test -race` without warnings

**Complete Service Example:**
```go
type ProductionService struct {
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup
    shutdownOnce sync.Once
    isRunning    atomic.Bool
    data         map[string]interface{}
    mu           sync.RWMutex
    logger       *logrus.Logger
}

func (s *ProductionService) Start() {
    s.isRunning.Store(true)

    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        defer func() {
            if r := recover(); r != nil {
                s.logger.Errorf("panic: %v\n%s", r, debug.Stack())
            }
        }()

        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                s.mu.Lock()
                s.data["last_tick"] = time.Now()
                s.mu.Unlock()
            case <-s.ctx.Done():
                return
            }
        }
    }()
}

func (s *ProductionService) Shutdown() {
    s.shutdownOnce.Do(func() {
        s.isRunning.Store(false)
        s.cancel()
        s.wg.Wait()
    })
}
```

---

## Module 7: Hands-On Labs (25 min)

### Lab 1: Make a Service Shutdown-Safe (10 min)

**Objective:** Take a service with double-close bugs and make it idempotent.

**Steps:**
1. Identify the `Shutdown()` method that panics on second call
2. Add `sync.Once` to wrap the shutdown logic
3. Write a test calling `Shutdown()` 10 times concurrently
4. Verify no panics occur

### Lab 2: Replace Mutex Booleans with atomic.Bool (5 min)

**Objective:** Replace mutex-protected boolean flags with atomic operations.

**Steps:**
1. Find all `sync.Mutex`-protected `bool` fields in a service
2. Replace with `atomic.Bool` and update Load/Store calls
3. Run benchmarks to compare performance
4. Verify with race detector

### Lab 3: Add Panic Recovery to a Worker Pool (5 min)

**Objective:** Prevent a worker panic from crashing the process.

**Steps:**
1. Identify the goroutine launch site in a worker pool
2. Add `defer recover()` with logging
3. Ensure `wg.Done()` is called even after panic
4. Write a test that triggers a panic and verifies the pool continues

### Lab 4: Fix All Race Conditions (5 min)

**Objective:** Run the race detector and fix all findings.

**Steps:**
1. Run `go test -race ./path/to/package`
2. Analyze each race detector report
3. Apply the appropriate fix (mutex, atomic, channel)
4. Re-run and verify zero race conditions

---

## Assessment

### Quiz (10 questions)

1. What guarantee does `sync.Once.Do(f)` provide?
2. Why is `atomic.Bool` preferred over `sync.Mutex` for boolean flags?
3. What is the correct order: `Add(1)` then `go func()`, or `go func()` then `Add(1)`?
4. Where must `recover()` be called to catch a goroutine panic?
5. What tool detects data races in Go programs?
6. How do you shut down a service with a timeout using WaitGroup?
7. What does `CompareAndSwap(false, true)` do on an `atomic.Bool`?
8. Why must deferred functions be ordered: recovery first, then `wg.Done()`?
9. What is the performance impact of the Go race detector?
10. What is the HelixAgent resource limit for running tests with `-race`?

### Practical Assessment

Refactor a provided concurrent service to apply all five patterns:

1. Add `sync.Once` to `Shutdown()` for idempotent cleanup
2. Replace mutex-protected booleans with `atomic.Bool`
3. Add `sync.WaitGroup` lifecycle tracking to all goroutines
4. Add panic recovery to every goroutine
5. Fix all data races detected by `go test -race`

Deliverables:
1. Refactored code with all five patterns applied
2. Test suite with idempotent shutdown, goroutine leak, and race detection tests
3. Benchmark comparison before and after atomic.Bool conversion
4. Race detector output showing zero warnings

---

## Resources

- [Course 61: Goroutine Safety & Lifecycle Management](video-course-61-goroutine-safety.md)
- [Course 65: Lazy Loading Patterns](video-course-65-lazy-loading-patterns.md)
- [sync Package Documentation](https://pkg.go.dev/sync)
- [sync/atomic Package Documentation](https://pkg.go.dev/sync/atomic)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Goroutine Lifecycle Diagram](../../docs/diagrams/src/goroutine-lifecycle.puml)
