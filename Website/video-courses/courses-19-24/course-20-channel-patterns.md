# Course-20: Channel Patterns and Goroutine Management

## Course Information
- **Duration:** 50 minutes
- **Level:** Advanced
- **Prerequisites:** Course-19 (Concurrency Patterns)
- **Last Updated:** February 27, 2026

## Overview

This course covers advanced channel patterns, goroutine lifecycle management, and sophisticated concurrency techniques essential for building high-performance, reliable concurrent systems in HelixAgent.

## Learning Objectives

By the end of this course, you will:
- Master advanced channel patterns (fan-in, fan-out, pipelines)
- Implement graceful goroutine shutdown
- Prevent goroutine leaks
- Use context for cancellation propagation
- Build robust concurrent pipelines

## Module 1: Channel Fundamentals Review (8 minutes)

### 1.1 Channel Types and Properties

**Channel Types:**
```go
// Unbuffered channel - synchronous
ch1 := make(chan int)      // Blocking send/receive

// Buffered channel - asynchronous
ch2 := make(chan int, 10)  // Non-blocking until buffer full

// Receive-only channel
func receiver(ch <-chan int) { }

// Send-only channel
func sender(ch chan<- int) { }

// Bidirectional channel
func processor(ch chan int) { }
```

**Channel Properties:**
| Property | Unbuffered | Buffered |
|----------|-----------|----------|
| Send blocks | Until received | Until buffer full |
| Receive blocks | Until sent | Until buffer empty |
| Guarantee | Synchronization | Memory only |
| Use case | Coordination | Decoupling |

### 1.2 Channel Closing Semantics

**Rules:**
1. Only sender should close a channel
2. Sending on closed channel panics
3. Receiving from closed channel returns zero value
4. Use comma-ok idiom to detect closed channels

```go
// Correct closing
func sender(ch chan int) {
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)  // Sender closes
}

// Correct receiving with comma-ok
func receiver(ch <-chan int) {
    for {
        val, ok := <-ch
        if !ok {
            break  // Channel closed
        }
        fmt.Println(val)
    }
}

// Alternative: range over channel
func receiverRange(ch <-chan int) {
    for val := range ch {
        fmt.Println(val)
    }
}
```

**Exercise 1.1:**
What's wrong with this code?

```go
func main() {
    ch := make(chan int)
    go func() {
        ch <- 1
        ch <- 2
    }()
    fmt.Println(<-ch)
    close(ch)  // What's the issue?
}
```

<details>
<summary>Answer</summary>

The main goroutine closes the channel, but the sender goroutine may still be sending. This causes a panic. Only the sender should close the channel.

```go
func main() {
    ch := make(chan int)
    go func() {
        defer close(ch)  // Sender closes
        ch <- 1
        ch <- 2
    }()
    fmt.Println(<-ch)
    fmt.Println(<-ch)
}
```
</details>

## Module 2: Fan-Out, Fan-In Pattern (15 minutes)

### 2.1 Fan-Out Pattern

**Definition:** Distribute work across multiple goroutines.

**Visual:**
```
   Input Channel
        │
        ▼
   ┌─────────┐
   │ Splitter│
   └────┬────┘
        │
   ┌────┼────┐
   ▼    ▼    ▼
  W1    W2    W3  (Workers)
```

**Implementation:**
```go
// Fan-out: Distribute work to multiple workers
func fanOut(input <-chan int, numWorkers int) []<-chan int {
    workers := make([]<-chan int, numWorkers)
    
    for i := 0; i < numWorkers; i++ {
        workers[i] = worker(input, i)
    }
    
    return workers
}

func worker(input <-chan int, id int) <-chan int {
    output := make(chan int)
    go func() {
        defer close(output)
        for val := range input {
            // Process work
            result := process(val)
            output <- result
        }
    }()
    return output
}
```

### 2.2 Fan-In Pattern

**Definition:** Combine multiple channels into one.

**Visual:**
```
  W1    W2    W3  (Workers)
   │     │     │
   └─────┼─────┘
         ▼
    ┌─────────┐
    │ Merger  │
    └────┬────┘
         │
         ▼
    Output Channel
```

**Implementation:**
```go
// Fan-in: Merge multiple channels into one
func fanIn(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)
    
    // Start a goroutine for each input channel
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                merged <- val
            }
        }(ch)
    }
    
    // Close merged channel when all inputs closed
    go func() {
        wg.Wait()
        close(merged)
    }()
    
    return merged
}
```

### 2.3 Complete Fan-Out/Fan-In Example

```go
func main() {
    // Generate input
    input := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
    
    // Fan-out to 3 workers
    workers := fanOut(input, 3)
    
    // Fan-in results
    results := fanIn(workers...)
    
    // Consume results
    for result := range results {
        fmt.Println(result)
    }
}

func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}
```

**Performance Characteristics:**
```
Sequential Processing:
- Time: 1000ms
- Throughput: 100 items/sec

Fan-Out (3 workers):
- Time: 350ms
- Throughput: 285 items/sec
- Speedup: 2.85x

Fan-Out (10 workers):
- Time: 120ms
- Throughput: 833 items/sec
- Speedup: 8.3x

Note: Diminishing returns after CPU core count
```

## Module 3: Pipeline Pattern (12 minutes)

### 3.1 Pipeline Architecture

**Definition:** Chain of processing stages connected by channels.

**Visual:**
```
Input → [Stage 1] → [Stage 2] → [Stage 3] → Output
           │            │            │
         Parse       Validate      Transform
```

**Implementation:**
```go
// Stage 1: Parse input
func parse(input <-chan string) <-chan Data {
    output := make(chan Data)
    go func() {
        defer close(output)
        for text := range input {
            data, err := parseData(text)
            if err == nil {
                output <- data
            }
        }
    }()
    return output
}

// Stage 2: Validate data
func validate(input <-chan Data) <-chan Data {
    output := make(chan Data)
    go func() {
        defer close(output)
        for data := range input {
            if isValid(data) {
                output <- data
            }
        }
    }()
    return output
}

// Stage 3: Transform data
func transform(input <-chan Data) <-chan Result {
    output := make(chan Result)
    go func() {
        defer close(output)
        for data := range input {
            result := transformData(data)
            output <- result
        }
    }()
    return output
}

// Compose pipeline
func pipeline(input <-chan string) <-chan Result {
    return transform(validate(parse(input)))
}
```

### 3.2 Pipeline with Cancellation

```go
func parseWithCancel(ctx context.Context, input <-chan string) <-chan Data {
    output := make(chan Data)
    go func() {
        defer close(output)
        for {
            select {
            case text, ok := <-input:
                if !ok {
                    return
                }
                data, err := parseData(text)
                if err == nil {
                    select {
                    case output <- data:
                    case <-ctx.Done():
                        return
                    }
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    return output
}
```

### 3.3 Parallel Pipeline Stages

```go
// Parallel stage with fan-out/fan-in
func parallelStage(input <-chan Data, numWorkers int) <-chan Result {
    // Fan-out
    workers := make([]<-chan Result, numWorkers)
    for i := 0; i < numWorkers; i++ {
        workers[i] = worker(input, i)
    }
    
    // Fan-in
    return fanIn(workers...)
}
```

**Lab Exercise 3.1:**
Build a log processing pipeline with 4 stages:
1. Parse log lines
2. Filter by severity
3. Enrich with metadata
4. Batch write to database

## Module 4: Graceful Shutdown (10 minutes)

### 4.1 The Problem

**Goroutine Leak Scenario:**
```go
func leaky() {
    ch := make(chan int)
    
    go func() {
        // This goroutine may run forever
        val := <-ch  // Blocks if no sender
        fmt.Println(val)
    }()
    
    // Function returns, but goroutine still blocked!
}
```

### 4.2 Using Quit Channels

```go
func graceful(done <-chan struct{}) {
    ch := make(chan int)
    
    go func() {
        for {
            select {
            case val := <-ch:
                fmt.Println(val)
            case <-done:
                // Cleanup and exit
                return
            }
        }
    }()
}

// Usage
func main() {
    done := make(chan struct{})
    graceful(done)
    
    // Later, signal shutdown
    close(done)
}
```

### 4.3 Using Context

```go
func withContext(ctx context.Context) {
    ch := make(chan int)
    
    go func() {
        for {
            select {
            case val := <-ch:
                fmt.Println(val)
            case <-ctx.Done():
                fmt.Println("Shutting down:", ctx.Err())
                return
            }
        }
    }()
}

// Usage
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    withContext(ctx)
    
    // Later, signal shutdown
    cancel()
}
```

### 4.4 Graceful HTTP Server Shutdown

```go
func main() {
    srv := &http.Server{Addr: ":8080"}
    
    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    <-quit
    log.Println("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}
```

### 4.5 Worker Pool with Graceful Shutdown

```go
type WorkerPool struct {
    workers  int
    jobs     chan Job
    results  chan Result
    quit     chan struct{}
    wg       sync.WaitGroup
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    
    for {
        select {
        case job, ok := <-wp.jobs:
            if !ok {
                return  // Channel closed
            }
            result := wp.process(job)
            wp.results <- result
        case <-wp.quit:
            return
        }
    }
}

func (wp *WorkerPool) Stop() {
    close(wp.quit)
    wp.wg.Wait()
}
```

## Module 5: Advanced Patterns (5 minutes)

### 5.1 Select with Default (Non-Blocking)

```go
select {
case val := <-ch:
    fmt.Println("Received:", val)
default:
    fmt.Println("No data available")
}
```

### 5.2 Select with Timeout

```go
select {
case val := <-ch:
    fmt.Println("Received:", val)
case <-time.After(5 * time.Second):
    fmt.Println("Timeout!")
}
```

### 5.3 Tee Pattern (Broadcast)

```go
func tee(input <-chan int) (<-chan int, <-chan int) {
    out1, out2 := make(chan int), make(chan int)
    
    go func() {
        defer close(out1)
        defer close(out2)
        
        for val := range input {
            out1 <- val
            out2 <- val
        }
    }()
    
    return out1, out2
}
```

### 5.4 Bridge Pattern (Channel of Channels)

```go
func bridge(input <-chan <-chan int) <-chan int {
    output := make(chan int)
    
    go func() {
        defer close(output)
        for ch := range input {
            for val := range ch {
                output <- val
            }
        }
    }()
    
    return output
}
```

## Assessment

### Quiz Questions

1. **What happens when you send to a closed channel?**
   - A) Returns an error
   - B) Blocks forever
   - C) Panic
   - D) No-op

2. **In Fan-Out pattern, what distributes work to workers?**
   - A) Fan-In function
   - B) Splitter/Multiplexer
   - C) Merge function
   - D) Pipeline stage

3. **Which is the best way to stop a goroutine gracefully?**
   - A) Use runtime.Goexit()
   - B) Use a quit channel or context cancellation
   - C) Let it run forever
   - D) Use panic/recover

### Practical Assignment

Implement a concurrent file processor:
1. Read file paths from a channel
2. Fan-out to 4 workers to read files
3. Process files through a 3-stage pipeline (parse, validate, transform)
4. Fan-in results and write to output
5. Implement graceful shutdown
6. No goroutine leaks

**Requirements:**
- Use fan-out/fan-in patterns
- Implement pipeline stages
- Handle errors gracefully
- Support cancellation
- Pass race detector

## Resources

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Advanced Go Concurrency](https://github.com/golang/go/wiki/ConcurrencyPatterns)
- [Context Package](https://pkg.go.dev/context)
- HelixAgent Pipeline Module: `internal/debate/pipeline.go`

## Troubleshooting

**Issue:** Deadlock in pipeline

**Solution:**
```go
// ❌ Deadlock: Unbuffered channel blocks
ch := make(chan int)
ch <- 1  // Blocks if no receiver

// ✅ Fixed: Use goroutine or buffered channel
go func() { ch <- 1 }()
// or
ch := make(chan int, 1)
ch <- 1
```

**Issue:** Goroutine leak

**Solution:**
```go
// ❌ Leak: Goroutine may block forever
go func() {
    val := <-ch  // No way to unblock
    fmt.Println(val)
}()

// ✅ Fixed: Use select with quit channel
quit := make(chan struct{})
go func() {
    select {
    case val := <-ch:
        fmt.Println(val)
    case <-quit:
        return
    }
}()
```

---

**Next Course:** Course-21: Race Condition Detection and Prevention
