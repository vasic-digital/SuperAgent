# Course-23: Memory Leak Prevention

## Course Information
- **Duration:** 45 minutes
- **Level:** Advanced
- **Prerequisites:** Course-22

## Module 1: Types of Memory Leaks (12 min)

**Type 1: Goroutine Leaks**
```go
func leaky() {
    ch := make(chan int)
    go func() {
        <-ch // Blocks forever
    }()
}
```

**Type 2: Slice Backing Array**
```go
big := make([]byte, 1000000)
small := big[:10] // Keeps reference to entire array!
```

**Type 3: Map Growth**
```go
m := make(map[string][]byte)
for {
    m[uuid()] = bigData // Never deleted
}
```

**Type 4: Circular References**
```go
type Node struct {
    next *Node
}
// A -> B -> A (circular)
```

## Module 2: Detection Tools (15 min)

**pprof Integration:**
```go
import _ "net/http/pprof"

func main() {
    go http.ListenAndServe("localhost:6060", nil)
}
```

**Usage:**
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

**Runtime Metrics:**
```go
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Alloc = %v MB", m.Alloc/1024/1024)
```

## Module 3: Prevention Patterns (15 min)

**Pattern 1: Proper Cleanup**
```go
func worker(ctx context.Context) {
    defer cleanup()
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-ch:
            process(msg)
        }
    }
}
```

**Pattern 2: Slice Truncation**
```go
// BAD
small := big[:10]

// GOOD
small := make([]byte, 10)
copy(small, big)
```

**Pattern 3: Bounded Maps**
```go
type LRUCache struct {
    maxSize int
    items   map[string]interface{}
}
```

## Module 4: Testing for Leaks (3 min)

**Leak Detector:**
```go
func TestNoLeak(t *testing.T) {
    before := runtime.NumGoroutine()
    functionUnderTest()
    time.Sleep(100 * time.Millisecond)
    after := runtime.NumGoroutine()
    assert.Equal(t, before, after)
}
```

## Assessment

**Lab:** Fix memory leaks in 3 real-world scenarios.
