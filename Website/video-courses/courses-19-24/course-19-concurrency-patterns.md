# Course-19: Advanced Concurrency Patterns and Mutex Best Practices

## Course Information
- **Duration:** 45 minutes
- **Level:** Advanced
- **Prerequisites:** Course-01-Fundamentals, Go programming basics
- **Last Updated:** February 27, 2026

## Overview

This course covers advanced concurrency patterns in HelixAgent, focusing on mutex usage, race condition prevention, and thread-safe programming techniques essential for building robust concurrent systems.

## Learning Objectives

By the end of this course, you will:
- Understand different mutex types and their use cases
- Implement thread-safe data structures
- Detect and prevent race conditions
- Apply deadlock prevention techniques
- Optimize concurrent performance

## Module 1: Understanding Go's Concurrency Model (12 minutes)

### 1.1 Goroutines and the Go Scheduler

**Key Concepts:**
- Goroutines are lightweight threads managed by the Go runtime
- The Go scheduler uses M:N scheduling (M goroutines on N OS threads)
- Goroutines start with small stacks (2KB) that grow dynamically

**Visual Explanation:**
```
┌─────────────────────────────────────────────┐
│           Go Runtime Scheduler              │
├─────────────────────────────────────────────┤
│  Goroutine 1  │  Goroutine 2  │  Goroutine N│
│  (Stack: 2KB) │  (Stack: 4KB) │  (Growing)  │
├─────────────────────────────────────────────┤
│     M OS Threads (Managed by Runtime)       │
├─────────────────────────────────────────────┤
│        Operating System (N CPU Cores)       │
└─────────────────────────────────────────────┘
```

**Code Example:**
```go
// Creating goroutines
func main() {
    var wg sync.WaitGroup
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            processTask(id)
        }(i)
    }
    
    wg.Wait()
}
```

### 1.2 The sync Package Essentials

**Mutex Types:**
1. **sync.Mutex** - Exclusive lock
2. **sync.RWMutex** - Read/Write lock
3. **sync.WaitGroup** - Goroutine synchronization
4. **sync.Once** - One-time initialization
5. **sync.Pool** - Object reuse

**Comparison Table:**
| Type | Reads | Writes | Use Case |
|------|-------|--------|----------|
| Mutex | Exclusive | Exclusive | Simple protection |
| RWMutex | Concurrent | Exclusive | Read-heavy workloads |

**Exercise 1.1:**
Create a goroutine that increments a counter 1000 times. Run 10 goroutines concurrently. What's the final count? Why?

<details>
<summary>Answer</summary>

The count will be less than 10000 due to race conditions. The increment operation is not atomic.

```go
// Race condition version
var counter int
for i := 0; i < 10; i++ {
    go func() {
        for j := 0; j < 1000; j++ {
            counter++ // Race here!
        }
    }()
}

// Correct version
var counter int
var mu sync.Mutex
for i := 0; i < 10; i++ {
    go func() {
        for j := 0; j < 1000; j++ {
            mu.Lock()
            counter++
            mu.Unlock()
        }
    }()
}
```
</details>

## Module 2: Mutex Patterns and Best Practices (15 minutes)

### 2.1 Proper Mutex Usage

**Golden Rules:**
1. Always pair Lock() with Unlock()
2. Use defer for Unlock() when possible
3. Keep critical sections small
4. Never call Lock() twice (deadlock)
5. Avoid calling external code while holding locks

**Anti-Patterns:**
```go
// ❌ Bad: Missing unlock on error path
func badExample() {
    mu.Lock()
    if err := doSomething(); err != nil {
        return // Lock never released!
    }
    mu.Unlock()
}

// ✅ Good: Using defer
func goodExample() {
    mu.Lock()
    defer mu.Unlock()
    
    if err := doSomething(); err != nil {
        return // Still unlocked via defer
    }
}
```

### 2.2 RWMutex for Read-Heavy Workloads

**When to Use:**
- Read operations >> Write operations
- Data changes infrequently
- Multiple readers, single writer

**Implementation:**
```go
type Cache struct {
    mu     sync.RWMutex
    data   map[string]string
    hits   int64
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    val, ok := c.data[key]
    if ok {
        atomic.AddInt64(&c.hits, 1)
    }
    return val, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[key] = value
}
```

**Performance Comparison:**
```
Scenario: 90% reads, 10% writes, 1000 goroutines

Mutex:
- Throughput: ~50,000 ops/sec
- Latency: 20ms p99

RWMutex:
- Throughput: ~450,000 ops/sec
- Latency: 2ms p99

Improvement: 9x throughput, 10x lower latency
```

### 2.3 Avoiding Common Deadlocks

**Deadlock Recipe:**
1. Goroutine A locks Resource 1, waits for Resource 2
2. Goroutine B locks Resource 2, waits for Resource 1

**Visual Representation:**
```
Goroutine A          Goroutine B
     │                    │
     ▼                    ▼
 Lock(X)               Lock(Y)
     │                    │
     ▼                    ▼
 Wait(Y) ────────────► Wait(X)
   (blocked)            (blocked)
   
   DEADLOCK!
```

**Prevention Strategies:**

1. **Lock Ordering:** Always acquire locks in the same order
```go
// Global lock order
const (
    OrderUsers = iota
    OrderAccounts
    OrderTransactions
)

func transfer(from, to *Account) {
    // Always lock in order of account ID
    if from.ID < to.ID {
        from.mu.Lock()
        to.mu.Lock()
    } else {
        to.mu.Lock()
        from.mu.Lock()
    }
    defer from.mu.Unlock()
    defer to.mu.Unlock()
    
    // Perform transfer
}
```

2. **Timeouts:** Never wait indefinitely
```go
func tryLockWithTimeout(mu *sync.Mutex, timeout time.Duration) bool {
    ch := make(chan struct{}, 1)
    
    go func() {
        mu.Lock()
        ch <- struct{}{}
    }()
    
    select {
    case <-ch:
        return true
    case <-time.After(timeout):
        return false
    }
}
```

**Lab Exercise 2.1:**
Implement a thread-safe LRU cache using sync.RWMutex. Measure its performance under different read/write ratios.

## Module 3: Race Condition Detection and Prevention (12 minutes)

### 3.1 Understanding Race Conditions

**Definition:** A race condition occurs when multiple goroutines access shared data concurrently, and at least one access is a write.

**Common Sources:**
1. Unprotected shared variables
2. Unsynchronized map access
3. Concurrent slice operations
4. Iterator invalidation

### 3.2 Using the Race Detector

**Running Tests:**
```bash
# Run with race detector
go test -race ./...

# Run specific package
go test -race ./internal/cache/...

# Run benchmark with race detection
go test -race -bench=. ./...
```

**Interpreting Output:**
```
WARNING: DATA RACE
Read at 0x00c0000... by goroutine 8:
  main.incrementCounter()
      main.go:15 +0x3e

Previous write at 0x00c0000... by goroutine 7:
  main.incrementCounter()
      main.go:15 +0x5a

Goroutine 8 (running) created at:
  main.main()
      main.go:23 +0x8c

Goroutine 7 (running) created at:
  main.main()
      main.go:22 +0x7b
```

**Key Information:**
- Location of conflicting access
- Which goroutines involved
- Where goroutines created

### 3.3 Atomic Operations

**When to Use:**
- Simple counter increments
- Flag setting/checking
- Single value updates

**Example:**
```go
var counter int64

// Concurrent increments (atomic)
for i := 0; i < 1000; i++ {
    go func() {
        atomic.AddInt64(&counter, 1)
    }()
}

// Atomic read
value := atomic.LoadInt64(&counter)

// Compare and swap
if atomic.CompareAndSwapInt64(&counter, oldVal, newVal) {
    // Successfully updated
}
```

**Atomic vs Mutex Performance:**
```
Operation: 1,000,000 increments

Atomic AddInt64:
- Time: 23ms
- No contention

Mutex Lock/Unlock:
- Time: 67ms
- Higher overhead

Atomic is ~3x faster for simple operations
```

### 3.4 Thread-Safe Data Structures

**sync.Map:**
```go
var cache sync.Map

// Store
cache.Store("key", "value")

// Load
if val, ok := cache.Load("key"); ok {
    fmt.Println(val)
}

// Load or store
val, loaded := cache.LoadOrStore("key", "default")

// Delete
cache.Delete("key")

// Range
cache.Range(func(key, value interface{}) bool {
    fmt.Printf("%s: %s\n", key, value)
    return true // continue iteration
})
```

**When to Use sync.Map:**
- Cache with frequent reads and infrequent writes
- Multiple goroutines reading/writing different keys
- Dynamic set of keys

**When NOT to use:**
- Need strong typing
- Require complex operations
- Small, fixed data sets

## Module 4: Advanced Patterns (6 minutes)

### 4.1 Double-Checked Locking

**Pattern:**
```go
type Singleton struct {
    mu     sync.Mutex
    value  *ExpensiveResource
}

func (s *Singleton) Get() *ExpensiveResource {
    // First check (no lock)
    if s.value != nil {
        return s.value
    }
    
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Second check (with lock)
    if s.value != nil {
        return s.value
    }
    
    s.value = createExpensiveResource()
    return s.value
}
```

**Modern Alternative (sync.Once):**
```go
type Singleton struct {
    once  sync.Once
    value *ExpensiveResource
}

func (s *Singleton) Get() *ExpensiveResource {
    s.once.Do(func() {
        s.value = createExpensiveResource()
    })
    return s.value
}
```

### 4.2 Errgroup Pattern

**Use Case:** Concurrent operations with error handling

```go
import "golang.org/x/sync/errgroup"

func processBatch(items []Item) error {
    g, ctx := errgroup.WithContext(context.Background())
    
    for _, item := range items {
        item := item // capture loop variable
        g.Go(func() error {
            return processItem(ctx, item)
        })
    }
    
    return g.Wait() // Returns first error or nil
}
```

## Assessment

### Quiz Questions

1. **What is the primary difference between sync.Mutex and sync.RWMutex?**
   - A) Mutex is faster for all operations
   - B) RWMutex allows concurrent reads
   - C) Mutex uses less memory
   - D) RWMutex is deprecated

2. **Which tool detects race conditions in Go programs?**
   - A) go vet
   - B) go test -race
   - C) go fmt
   - D) go build -race

3. **What is a deadlock?**
   - A) A goroutine that runs forever
   - B) A situation where goroutines wait for each other indefinitely
   - C) A memory leak
   - D) A race condition

### Practical Assignment

Implement a thread-safe rate limiter using token bucket algorithm with:
- sync.RWMutex for read-heavy operations
- Proper cleanup on shutdown
- Race-condition free
- Benchmark showing throughput

**Submission:**
- Source code
- Benchmark results
- Race detector output showing no races

## Resources

- [Go Memory Model](https://golang.org/ref/mem)
- [Sync Package Documentation](https://pkg.go.dev/sync)
- [Race Detector Blog Post](https://blog.golang.org/race-detector)
- HelixAgent Concurrency Module: `internal/concurrency/`

## Troubleshooting

**Common Issue:** "fatal error: all goroutines are asleep - deadlock!"

**Solution:**
```go
// ❌ Deadlock: Goroutine waits for itself
func deadlockExample() {
    ch := make(chan int)
    ch <- 1 // Blocks forever - no receiver
    <-ch
}

// ✅ Fixed: Use goroutine or buffered channel
func fixedExample() {
    ch := make(chan int, 1) // Buffered
    ch <- 1
    <-ch
}
```

**Common Issue:** Data race with range loop

**Solution:**
```go
// ❌ Race: Sharing loop variable
for _, item := range items {
    go func() {
        process(item) // All goroutines see last item!
    }()
}

// ✅ Fixed: Capture loop variable
for _, item := range items {
    item := item // Shadow variable
    go func() {
        process(item)
    }()
}
```

---

**Next Course:** Course-20: Channel Patterns and Goroutine Management
