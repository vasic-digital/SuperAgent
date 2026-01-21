# Concurrency Package

The concurrency package provides a high-performance worker pool implementation for parallel task execution in HelixAgent.

## Overview

This package implements a task-based worker pool that enables efficient concurrent processing of workloads with built-in metrics collection, batch processing support, and graceful shutdown capabilities.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      WorkerPool                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  Task Queue │──│   Workers   │──│  Result Collector   │ │
│  │  (buffered) │  │  (N conc.)  │  │  (callbacks/chan)   │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Metrics                               ││
│  │  TasksSubmitted | TasksCompleted | TasksFailed | AvgTime││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### WorkerPool

The main worker pool that manages concurrent task execution.

```go
type WorkerPool struct {
    workers      int
    taskQueue    chan Task
    results      chan TaskResult
    wg           sync.WaitGroup
    ctx          context.Context
    cancel       context.CancelFunc
    metrics      *PoolMetrics
    running      atomic.Bool
}
```

### Task

Represents a unit of work to be executed.

```go
type Task struct {
    ID       string
    Execute  func(ctx context.Context) (interface{}, error)
    Callback func(result TaskResult)
    Priority int
    Timeout  time.Duration
}
```

### TaskResult

Contains the result of task execution.

```go
type TaskResult struct {
    TaskID   string
    Result   interface{}
    Error    error
    Duration time.Duration
}
```

### PoolMetrics

Tracks pool performance statistics.

```go
type PoolMetrics struct {
    TasksSubmitted   int64
    TasksCompleted   int64
    TasksFailed      int64
    TotalDuration    time.Duration
    AverageTaskTime  time.Duration
}
```

## Configuration

```go
type PoolConfig struct {
    Workers         int           // Number of concurrent workers
    QueueSize       int           // Task queue buffer size
    TaskTimeout     time.Duration // Default timeout per task
    ShutdownTimeout time.Duration // Graceful shutdown timeout
    EnableMetrics   bool          // Enable metrics collection
}
```

## Usage Examples

### Basic Worker Pool

```go
import "dev.helix.agent/internal/concurrency"

// Create a pool with 10 workers
pool := concurrency.NewWorkerPool(concurrency.PoolConfig{
    Workers:     10,
    QueueSize:   100,
    TaskTimeout: 30 * time.Second,
})

// Start the pool
pool.Start()

// Submit a task
task := concurrency.Task{
    ID: "task-1",
    Execute: func(ctx context.Context) (interface{}, error) {
        // Perform work
        return processData(), nil
    },
}
pool.Submit(task)

// Graceful shutdown
pool.Shutdown()
```

### With Callbacks

```go
task := concurrency.Task{
    ID: "task-with-callback",
    Execute: func(ctx context.Context) (interface{}, error) {
        return fetchData(ctx)
    },
    Callback: func(result concurrency.TaskResult) {
        if result.Error != nil {
            log.Printf("Task %s failed: %v", result.TaskID, result.Error)
            return
        }
        log.Printf("Task %s completed in %v", result.TaskID, result.Duration)
    },
}
pool.Submit(task)
```

### Batch Processing

```go
// Submit multiple tasks as a batch
tasks := []concurrency.Task{
    {ID: "batch-1", Execute: task1Func},
    {ID: "batch-2", Execute: task2Func},
    {ID: "batch-3", Execute: task3Func},
}

results := pool.SubmitBatch(tasks)
for result := range results {
    fmt.Printf("Task %s: %v\n", result.TaskID, result.Result)
}
```

### With Metrics

```go
pool := concurrency.NewWorkerPool(concurrency.PoolConfig{
    Workers:       20,
    EnableMetrics: true,
})
pool.Start()

// ... submit tasks ...

// Get metrics
metrics := pool.GetMetrics()
fmt.Printf("Submitted: %d, Completed: %d, Failed: %d\n",
    metrics.TasksSubmitted,
    metrics.TasksCompleted,
    metrics.TasksFailed)
fmt.Printf("Average task time: %v\n", metrics.AverageTaskTime)
```

## Features

### Graceful Shutdown
- Waits for in-flight tasks to complete
- Configurable shutdown timeout
- Drains task queue before stopping

### Priority Scheduling
- Higher priority tasks execute first
- Priority levels: Low (0), Normal (1), High (2), Critical (3)

### Context Propagation
- Tasks receive context with timeout
- Supports cancellation propagation
- Parent context cancellation stops all tasks

### Error Handling
- Task errors captured in TaskResult
- Failed tasks don't affect other tasks
- Panic recovery within workers

## Integration with HelixAgent

The worker pool is used throughout HelixAgent for:

- **Parallel LLM Calls**: Execute requests to multiple providers concurrently
- **Batch Verification**: Run verification tests in parallel
- **Background Tasks**: Process queued tasks asynchronously
- **Debate Orchestration**: Parallel agent execution in debates

## Testing

```bash
go test -v ./internal/concurrency/...
go test -bench=. ./internal/concurrency/...  # Benchmark tests
go test -race ./internal/concurrency/...     # Race detection
```

## Performance Considerations

1. **Worker Count**: Set based on workload characteristics (I/O vs CPU bound)
2. **Queue Size**: Large enough to handle burst traffic without blocking
3. **Task Timeout**: Prevent stuck tasks from blocking workers
4. **Metrics Overhead**: Minimal when enabled, atomic operations only
