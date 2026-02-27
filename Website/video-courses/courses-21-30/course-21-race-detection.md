# Course-21: Race Condition Detection and Prevention

## Course Information
- **Duration:** 45 minutes
- **Level:** Advanced
- **Prerequisites:** Course-19, Course-20

## Module 1: Understanding Race Conditions (10 min)

**Definition:** A race condition occurs when the outcome depends on the non-deterministic ordering of events.

**Common Causes:**
1. Unprotected shared variables
2. Concurrent map writes
3. Unsynchronized counter increments
4. Iterator invalidation

**Example Race:**
```go
var counter int

func increment() {
    counter++ // Not atomic! Read-modify-write
}
```

## Module 2: Go Race Detector (15 min)

**Usage:**
```bash
go test -race ./...
go run -race main.go
```

**Interpreting Output:**
```
WARNING: DATA RACE
Read at 0x00c000... by goroutine 8:
  main.increment()
      main.go:15 +0x3e

Previous write at 0x00c000... by goroutine 7:
  main.increment()
      main.go:15 +0x5a
```

## Module 3: Prevention Patterns (15 min)

**Pattern 1: Mutex Protection**
```go
var (
    counter int
    mu      sync.Mutex
)

func safeIncrement() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}
```

**Pattern 2: Atomic Operations**
```go
var counter int64

func atomicIncrement() {
    atomic.AddInt64(&counter, 1)
}
```

**Pattern 3: Channel Communication**
```go
func worker(ch chan int) {
    for val := range ch {
        process(val) // No shared state
    }
}
```

## Module 4: Testing for Races (5 min)

**Stress Testing:**
```go
func TestConcurrentAccess(t *testing.T) {
    for i := 0; i < 1000; i++ {
        go accessSharedResource()
    }
}
```

Run with: `go test -race -count=100`

## Assessment

**Quiz:**
1. What tool detects race conditions in Go?
2. Why is `counter++` not atomic?
3. What's the difference between mutex and atomic?

**Lab:** Fix 5 race conditions in provided code.
