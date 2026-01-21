# Events Package

The events package provides a high-performance publish-subscribe event bus for HelixAgent's internal event-driven architecture.

## Overview

This package implements a type-safe event bus that enables decoupled communication between components through event publishing and subscription with support for filtering, metrics, and concurrent handling.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        EventBus                              │
│                                                              │
│  ┌─────────────┐     ┌─────────────────────────────────┐   │
│  │  Publisher  │────▶│         Event Router            │   │
│  └─────────────┘     │  ┌─────────┐ ┌─────────────────┐│   │
│                      │  │ Filters │ │ Topic Matching  ││   │
│                      │  └─────────┘ └─────────────────┘│   │
│                      └──────────────┬──────────────────┘   │
│                                     │                       │
│      ┌──────────────────────────────┼──────────────────────┐│
│      ▼                              ▼                      ▼││
│  ┌─────────┐                 ┌─────────┐            ┌─────────┐│
│  │Subscriber│                │Subscriber│           │Subscriber││
│  │ (sync)   │                │ (async)  │           │ (chan)   ││
│  └─────────┘                 └─────────┘            └─────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                      Metrics                             ││
│  │  EventsPublished | EventsDelivered | Subscribers | Lag  ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### EventBus

The central event dispatcher.

```go
type EventBus struct {
    subscribers map[string][]Subscription
    mu          sync.RWMutex
    metrics     *BusMetrics
    closed      atomic.Bool
    bufferSize  int
}
```

### Event

Represents an event in the system.

```go
type Event struct {
    ID        string                 // Unique event ID
    Type      string                 // Event type/topic
    Source    string                 // Event source component
    Timestamp time.Time              // When event was created
    Data      interface{}            // Event payload
    Metadata  map[string]string      // Additional metadata
}
```

### Subscription

Represents a subscriber's registration.

```go
type Subscription struct {
    ID       string
    Topic    string
    Handler  EventHandler
    Filter   EventFilter
    Async    bool
    Channel  chan Event
}

type EventHandler func(event Event) error
type EventFilter func(event Event) bool
```

### BusMetrics

Tracks event bus performance.

```go
type BusMetrics struct {
    EventsPublished   int64
    EventsDelivered   int64
    EventsFiltered    int64
    ActiveSubscribers int64
    DeliveryErrors    int64
    AverageLatency    time.Duration
}
```

## Event Types

HelixAgent defines standard event types:

| Event Type | Description |
|------------|-------------|
| `provider.registered` | LLM provider registered |
| `provider.health.changed` | Provider health status changed |
| `debate.started` | Debate session started |
| `debate.round.completed` | Debate round finished |
| `debate.completed` | Debate finished with consensus |
| `task.queued` | Background task queued |
| `task.started` | Task execution started |
| `task.completed` | Task finished |
| `verification.started` | Provider verification started |
| `verification.completed` | Verification finished |
| `cache.hit` | Cache hit occurred |
| `cache.miss` | Cache miss occurred |

## Usage Examples

### Basic Pub/Sub

```go
import "dev.helix.agent/internal/events"

// Create event bus
bus := events.NewEventBus(events.BusConfig{
    BufferSize: 1000,
})

// Subscribe to events
subID := bus.Subscribe("provider.health.changed", func(e events.Event) error {
    data := e.Data.(HealthChangeData)
    log.Printf("Provider %s health: %s -> %s",
        data.ProviderID, data.OldStatus, data.NewStatus)
    return nil
})

// Publish event
bus.Publish(events.Event{
    Type:   "provider.health.changed",
    Source: "health-monitor",
    Data: HealthChangeData{
        ProviderID: "claude",
        OldStatus:  "healthy",
        NewStatus:  "degraded",
    },
})

// Unsubscribe when done
bus.Unsubscribe(subID)
```

### With Filtering

```go
// Subscribe only to critical health events
bus.SubscribeWithFilter("provider.health.changed",
    func(e events.Event) error {
        // Handle critical health change
        return nil
    },
    func(e events.Event) bool {
        data := e.Data.(HealthChangeData)
        return data.NewStatus == "unhealthy"
    },
)
```

### Async Subscription

```go
// Async handler - doesn't block publisher
bus.SubscribeAsync("task.completed", func(e events.Event) error {
    // Long-running processing
    processTaskResult(e.Data)
    return nil
})
```

### Channel-Based Subscription

```go
// Get a channel of events
eventChan := bus.SubscribeChannel("debate.*", 100)

go func() {
    for event := range eventChan {
        switch event.Type {
        case "debate.started":
            handleDebateStart(event)
        case "debate.completed":
            handleDebateEnd(event)
        }
    }
}()
```

### Wildcard Topics

```go
// Subscribe to all provider events
bus.Subscribe("provider.*", func(e events.Event) error {
    log.Printf("Provider event: %s", e.Type)
    return nil
})

// Subscribe to all events
bus.Subscribe("*", func(e events.Event) error {
    metrics.RecordEvent(e.Type)
    return nil
})
```

### Event Replay

```go
// Enable event history
bus := events.NewEventBus(events.BusConfig{
    EnableHistory: true,
    HistorySize:   1000,
})

// Get recent events
recentEvents := bus.GetHistory("debate.*", 10)
for _, e := range recentEvents {
    fmt.Printf("[%s] %s: %v\n", e.Timestamp, e.Type, e.Data)
}
```

## Features

### Topic Patterns
- Exact match: `"provider.registered"`
- Wildcard suffix: `"provider.*"`
- Global wildcard: `"*"`

### Delivery Guarantees
- At-least-once delivery for sync subscribers
- Best-effort for async subscribers
- Configurable retry on handler errors

### Backpressure Handling
- Bounded channels for async delivery
- Slow subscriber detection
- Event dropping with metrics

### Thread Safety
- Concurrent publish from multiple goroutines
- Safe subscribe/unsubscribe during operation
- Lock-free fast path for publishing

## Integration with HelixAgent

The event bus is used for:

- **Provider Monitoring**: Health status change notifications
- **Debate Progress**: Real-time debate updates
- **Task Tracking**: Background task lifecycle events
- **Cache Analytics**: Hit/miss tracking for optimization
- **Audit Logging**: Security-relevant event capture

## Testing

```bash
go test -v ./internal/events/...
go test -bench=. ./internal/events/...  # Benchmark tests
go test -race ./internal/events/...     # Race detection
```

## Performance Characteristics

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Publish | O(n) | n = matching subscribers |
| Subscribe | O(1) | Amortized |
| Unsubscribe | O(n) | n = topic subscribers |
| Filter Check | O(1) | Per subscriber |

Typical throughput: 100K+ events/second on modern hardware.
