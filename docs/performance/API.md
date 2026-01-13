# Performance API Reference

This document describes the APIs exposed by HelixAgent's performance optimization components.

## Worker Pool API

### Creating a Worker Pool

```go
import "github.com/helixagent/internal/concurrency"

// Create with default config
pool := concurrency.NewWorkerPool(nil)

// Create with custom config
config := &concurrency.PoolConfig{
    Workers:       8,
    QueueSize:     1000,
    TaskTimeout:   30 * time.Second,
    ShutdownGrace: 5 * time.Second,
    OnError: func(taskID string, err error) {
        log.Printf("Task %s failed: %v", taskID, err)
    },
    OnComplete: func(result concurrency.Result) {
        log.Printf("Task %s completed in %v", result.TaskID, result.Duration)
    },
}
pool := concurrency.NewWorkerPool(config)
```

### Submitting Tasks

```go
// Create a task
task := concurrency.NewTaskFunc("my-task-1", func(ctx context.Context) (interface{}, error) {
    // Do work...
    return "result", nil
})

// Submit (non-blocking)
if err := pool.Submit(task); err != nil {
    log.Printf("Failed to submit: %v", err)
}

// Submit and wait for result
result, err := pool.SubmitWait(ctx, task)
if err != nil {
    log.Printf("Task failed: %v", err)
}
fmt.Printf("Result: %v\n", result.Value)
```

### Batch Operations

```go
// Submit batch of tasks
tasks := []concurrency.Task{
    concurrency.NewTaskFunc("task-1", func(ctx context.Context) (interface{}, error) { return 1, nil }),
    concurrency.NewTaskFunc("task-2", func(ctx context.Context) (interface{}, error) { return 2, nil }),
    concurrency.NewTaskFunc("task-3", func(ctx context.Context) (interface{}, error) { return 3, nil }),
}

// Get results channel
resultChan := pool.SubmitBatch(tasks)
for result := range resultChan {
    fmt.Printf("Task %s: %v\n", result.TaskID, result.Value)
}

// Or wait for all results
results, err := pool.SubmitBatchWait(ctx, tasks)
```

### Utility Functions

```go
// Parallel execute functions
results, err := concurrency.ParallelExecute(ctx, []func(context.Context) (interface{}, error){
    func(ctx context.Context) (interface{}, error) { return fetchA(ctx) },
    func(ctx context.Context) (interface{}, error) { return fetchB(ctx) },
})

// Parallel map over items
items := []string{"a", "b", "c"}
results, err := concurrency.Map(ctx, items, 4, func(ctx context.Context, item string) (int, error) {
    return len(item), nil
})
```

### Pool Management

```go
// Start the pool (auto-started on first Submit)
pool.Start()

// Check status
pool.IsRunning()      // bool
pool.QueueLength()    // int
pool.ActiveWorkers()  // int

// Get metrics
metrics := pool.Metrics()
fmt.Printf("Completed: %d, Failed: %d, Avg Latency: %v\n",
    metrics.CompletedTasks, metrics.FailedTasks, metrics.AverageLatency())

// Graceful shutdown
pool.Shutdown(5 * time.Second)

// Immediate stop
pool.Stop()

// Wait for queue to drain
pool.WaitForDrain(ctx)
```

## Event Bus API

### Creating an Event Bus

```go
import "github.com/helixagent/internal/events"

// Create with default config
bus := events.NewEventBus(nil)

// Create with custom config
config := &events.BusConfig{
    BufferSize:      1000,
    PublishTimeout:  10 * time.Millisecond,
    CleanupInterval: 30 * time.Second,
    MaxSubscribers:  100,
}
bus := events.NewEventBus(config)
```

### Publishing Events

```go
// Create an event
event := events.NewEvent(events.EventProviderHealthChanged, "provider-registry", map[string]interface{}{
    "provider": "claude",
    "healthy":  true,
})

// Add metadata
event.WithTraceID("trace-123").WithMetadata("version", "1.0")

// Publish synchronously
bus.Publish(event)

// Publish asynchronously
bus.PublishAsync(event)
```

### Subscribing to Events

```go
// Subscribe to specific event type
ch := bus.Subscribe(events.EventProviderHealthChanged)
go func() {
    for event := range ch {
        fmt.Printf("Received: %s\n", event.Type)
    }
}()

// Subscribe with filter
ch := bus.SubscribeWithFilter(events.EventCacheHit, func(e *events.Event) bool {
    return e.Source == "my-cache"
})

// Subscribe to multiple types
ch := bus.SubscribeMultiple(events.EventCacheHit, events.EventCacheMiss)

// Subscribe to all events
ch := bus.SubscribeAll()

// Unsubscribe
bus.Unsubscribe(ch)
```

### Waiting for Events

```go
// Wait for specific event
event, err := bus.Wait(ctx, events.EventProviderHealthChanged)

// Wait for any of multiple events
event, err := bus.WaitMultiple(ctx, events.EventDebateStarted, events.EventDebateFailed)
```

### Global Event Bus

```go
// Initialize global bus
events.InitGlobalBus(nil)

// Emit to global bus
events.Emit(events.NewEvent(events.EventSystemStartup, "main", nil))
events.EmitAsync(event)

// Subscribe to global bus
ch := events.On(events.EventSystemStartup)
```

### Event Types

```go
// Provider events
events.EventProviderRegistered      // provider.registered
events.EventProviderUnregistered    // provider.unregistered
events.EventProviderHealthChanged   // provider.health.changed
events.EventProviderScoreUpdated    // provider.score.updated

// MCP events
events.EventMCPServerConnected      // mcp.server.connected
events.EventMCPServerDisconnected   // mcp.server.disconnected
events.EventMCPServerHealthChanged  // mcp.server.health.changed
events.EventMCPToolExecuted         // mcp.tool.executed
events.EventMCPToolFailed           // mcp.tool.failed

// Debate events
events.EventDebateStarted           // debate.started
events.EventDebateRoundStarted      // debate.round.started
events.EventDebateRoundCompleted    // debate.round.completed
events.EventDebateCompleted         // debate.completed
events.EventDebateFailed            // debate.failed

// Cache events
events.EventCacheHit                // cache.hit
events.EventCacheMiss               // cache.miss
events.EventCacheInvalidated        // cache.invalidated
events.EventCacheExpired            // cache.expired

// System events
events.EventSystemStartup           // system.startup
events.EventSystemShutdown          // system.shutdown
events.EventSystemHealthCheck       // system.health.check
events.EventSystemError             // system.error

// Request events
events.EventRequestReceived         // request.received
events.EventRequestCompleted        // request.completed
events.EventRequestFailed           // request.failed
```

## Tiered Cache API

### Creating a Cache

```go
import "github.com/helixagent/internal/cache"

// Create with Redis client
config := &cache.TieredCacheConfig{
    L1MaxSize:         10000,
    L1TTL:             5 * time.Minute,
    L1CleanupInterval: time.Minute,
    L2TTL:             30 * time.Minute,
    L2Compression:     true,
    L2KeyPrefix:       "tiered:",
    NegativeTTL:       30 * time.Second,
    EnableL1:          true,
    EnableL2:          true,
}
tc := cache.NewTieredCache(redisClient, config)
```

### Basic Operations

```go
// Set a value
err := tc.Set(ctx, "user:123", userData, 10*time.Minute, "user", "active")

// Get a value
var user User
found, err := tc.Get(ctx, "user:123", &user)

// Delete a value
err := tc.Delete(ctx, "user:123")
```

### Tag-Based Invalidation

```go
// Set with tags
tc.Set(ctx, "user:123", user, 10*time.Minute, "user", "premium")
tc.Set(ctx, "user:456", user, 10*time.Minute, "user", "free")

// Invalidate by tag
count, err := tc.InvalidateByTag(ctx, "premium")  // Removes user:123
count, err := tc.InvalidateByTags(ctx, "user", "active")  // Multiple tags

// Invalidate by prefix
count, err := tc.InvalidatePrefix(ctx, "user:")  // Removes all user entries
```

### Metrics

```go
metrics := tc.Metrics()
fmt.Printf("L1 Hits: %d, Misses: %d\n", metrics.L1Hits, metrics.L1Misses)
fmt.Printf("L2 Hits: %d, Misses: %d\n", metrics.L2Hits, metrics.L2Misses)
fmt.Printf("Hit Rate: %.2f%%\n", tc.HitRate())
fmt.Printf("Compression Saved: %d bytes\n", metrics.CompressionSaved)
```

## HTTP Client Pool API

### Creating a Pool

```go
import "github.com/helixagent/internal/http"

config := &http.PoolConfig{
    MaxIdleConns:          100,
    MaxConnsPerHost:       10,
    IdleConnTimeout:       90 * time.Second,
    ResponseHeaderTimeout: 10 * time.Second,
    DialTimeout:           5 * time.Second,
}
pool := http.NewHTTPClientPool(config)
```

### Getting Clients

```go
// Get client for host (reuses existing)
client := pool.GetClient("api.example.com")

// Make request
resp, err := client.Get("https://api.example.com/data")
```

### Retry Client

```go
retryConfig := &http.RetryConfig{
    MaxRetries:        3,
    InitialBackoff:    100 * time.Millisecond,
    MaxBackoff:        5 * time.Second,
    BackoffMultiplier: 2.0,
}
retryClient := http.NewRetryClient(pool.GetClient("api.example.com"), retryConfig)

// Requests automatically retry on failure
resp, err := retryClient.Do(req)
```

### Pool Management

```go
// Get metrics
metrics := pool.Metrics()
fmt.Printf("Connections: %d, Requests: %d\n",
    metrics.ActiveConnections, metrics.TotalRequests)

// Close pool
pool.Close()
```

## Lazy Provider API

### Creating a Lazy Provider

```go
import "github.com/helixagent/internal/llm"

provider := llm.NewLazyProvider(func() (llm.LLMProvider, error) {
    // Heavy initialization here
    return claude.NewClaudeProvider(config)
})
```

### Using the Provider

```go
// Get the provider (initializes on first call)
p, err := provider.Get()
if err != nil {
    log.Fatal(err)
}

// Use the provider
response, err := p.Complete(ctx, request)

// Check initialization status
if provider.IsInitialized() {
    fmt.Printf("Initialized in %v\n", provider.InitTime())
}
```

## REST API Endpoints

### Performance Metrics

```http
GET /v1/metrics/performance
```

Response:
```json
{
  "worker_pool": {
    "active_workers": 4,
    "queued_tasks": 12,
    "completed_tasks": 1523,
    "failed_tasks": 3,
    "average_latency_ms": 45
  },
  "event_bus": {
    "events_published": 5432,
    "events_delivered": 5430,
    "events_dropped": 2,
    "active_subscribers": 15
  },
  "cache": {
    "l1_hits": 8765,
    "l1_misses": 234,
    "l2_hits": 200,
    "l2_misses": 34,
    "hit_rate": 97.8,
    "l1_size": 4532,
    "compression_saved_bytes": 1234567
  },
  "http_pool": {
    "active_connections": 45,
    "total_requests": 12345,
    "average_latency_ms": 23
  },
  "mcp": {
    "active_connections": 8,
    "pending_connections": 2,
    "failed_connections": 1
  }
}
```

### MCP Server Status

```http
GET /v1/mcp/status
```

Response:
```json
{
  "servers": [
    {
      "name": "filesystem",
      "status": "connected",
      "last_heartbeat": "2024-01-15T10:30:00Z",
      "latency_ms": 5
    },
    {
      "name": "github",
      "status": "connected",
      "last_heartbeat": "2024-01-15T10:30:00Z",
      "latency_ms": 12
    }
  ],
  "total": 12,
  "connected": 10,
  "pending": 1,
  "failed": 1
}
```

### Event Stream (SSE)

```http
GET /v1/events/stream
Accept: text/event-stream
```

Response:
```
event: provider.health.changed
data: {"type":"provider.health.changed","source":"provider-registry","payload":{"provider":"claude","healthy":true},"timestamp":"2024-01-15T10:30:00Z"}

event: cache.hit
data: {"type":"cache.hit","source":"tiered-cache","payload":{"key":"user:123","tier":"l1"},"timestamp":"2024-01-15T10:30:01Z"}
```

### Cache Management

```http
DELETE /v1/cache/invalidate
Content-Type: application/json

{
  "pattern": "user:*"
}
```

```http
DELETE /v1/cache/tag/:tag
```
