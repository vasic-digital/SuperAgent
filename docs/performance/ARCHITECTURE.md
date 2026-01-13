# Performance Architecture

This document describes the performance optimization architecture implemented in HelixAgent.

## Overview

HelixAgent implements a multi-layered performance optimization strategy:

1. **MCP Pre-Installation & Lazy Loading** - NPM packages pre-installed at startup, connections lazily established
2. **Worker Pool** - Bounded concurrency with configurable workers
3. **Event Bus** - Async pub/sub communication between components
4. **Tiered Cache** - L1 (memory) + L2 (Redis) with compression
5. **HTTP Client Pool** - Connection reuse with retry logic
6. **Lazy Provider Initialization** - Providers initialized on first use

## Component Diagrams

### MCP Pre-Installation Flow

```
┌─────────────────┐
│  Startup Hook   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Preinstaller   │──────────────────┐
│  (Background)   │                  │
└────────┬────────┘                  │
         │                           │
         ▼                           ▼
┌─────────────────┐        ┌─────────────────┐
│ filesystem pkg  │        │   github pkg    │
└─────────────────┘        └─────────────────┘
         │                           │
         ▼                           ▼
┌─────────────────┐        ┌─────────────────┐
│  memory pkg     │        │   fetch pkg     │
└─────────────────┘        └─────────────────┘
         │                           │
         ▼                           ▼
┌─────────────────┐        ┌─────────────────┐
│  puppeteer pkg  │        │   sqlite pkg    │
└─────────────────┘        └─────────────────┘
```

### Lazy Connection Pool

```
┌─────────────────────────────────────────────────┐
│            MCPConnectionPool                     │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ pending  │  │connecting│  │connected │       │
│  │ servers  │──│  pool    │──│  pool    │       │
│  └──────────┘  └──────────┘  └──────────┘       │
│        │             │             │             │
│        ▼             ▼             ▼             │
│  GetConnection() creates on-demand              │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Worker Pool Architecture

```
┌─────────────────────────────────────────────────┐
│               WorkerPool                         │
├─────────────────────────────────────────────────┤
│                                                  │
│   Submit()                                       │
│      │                                           │
│      ▼                                           │
│  ┌────────────────────────────────┐             │
│  │         Task Queue             │             │
│  │   (Buffered Channel)           │             │
│  └────────────┬───────────────────┘             │
│               │                                  │
│       ┌───────┼───────┬───────┬───────┐        │
│       ▼       ▼       ▼       ▼       ▼        │
│    ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐    │
│    │ W1  │ │ W2  │ │ W3  │ │ W4  │ │ Wn  │    │
│    └──┬──┘ └──┬──┘ └──┬──┘ └──┬──┘ └──┬──┘    │
│       │       │       │       │       │        │
│       └───────┴───────┴───────┴───────┘        │
│                       │                         │
│                       ▼                         │
│              ┌─────────────────┐               │
│              │  Results Chan   │               │
│              └─────────────────┘               │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Event Bus Topology

```
┌─────────────────────────────────────────────────┐
│                  EventBus                        │
├─────────────────────────────────────────────────┤
│                                                  │
│  Publishers                  Subscribers         │
│  ──────────                  ───────────         │
│                                                  │
│  ┌─────────┐    ┌───────┐   ┌─────────────┐    │
│  │Provider │───▶│ Topic │──▶│  Cache      │    │
│  │Registry │    │ Router│   │  Invalidator│    │
│  └─────────┘    │       │   └─────────────┘    │
│                 │       │                       │
│  ┌─────────┐    │       │   ┌─────────────┐    │
│  │  MCP    │───▶│       │──▶│  Metrics    │    │
│  │ Pool    │    │       │   │  Collector  │    │
│  └─────────┘    │       │   └─────────────┘    │
│                 │       │                       │
│  ┌─────────┐    │       │   ┌─────────────┐    │
│  │ Worker  │───▶│       │──▶│  Health     │    │
│  │ Pool    │    │       │   │  Monitor    │    │
│  └─────────┘    └───────┘   └─────────────┘    │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Tiered Cache System

```
┌─────────────────────────────────────────────────┐
│              TieredCache                         │
├─────────────────────────────────────────────────┤
│                                                  │
│  GET Request                                     │
│      │                                           │
│      ▼                                           │
│  ┌─────────────────┐                            │
│  │   L1 (Memory)   │──HIT──▶ Return Value       │
│  │   sync.Map      │                            │
│  └────────┬────────┘                            │
│           │ MISS                                 │
│           ▼                                      │
│  ┌─────────────────┐                            │
│  │   L2 (Redis)    │──HIT──▶ Promote to L1     │
│  │   + Compression │         Return Value       │
│  └────────┬────────┘                            │
│           │ MISS                                 │
│           ▼                                      │
│       Return nil                                 │
│                                                  │
│  SET Request                                     │
│      │                                           │
│      ├──────────────────────────┐               │
│      ▼                          ▼               │
│  ┌─────────────────┐  ┌─────────────────┐       │
│  │   L1 (Memory)   │  │   L2 (Redis)    │       │
│  │   TTL: short    │  │   TTL: long     │       │
│  └─────────────────┘  │   + gzip        │       │
│                       └─────────────────┘       │
│                                                  │
└─────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Lazy Initialization

All heavy resources are lazily initialized to minimize startup time:

- **LLM Providers**: Created on first API call using `sync.Once`
- **MCP Connections**: Established on first tool execution
- **Database Pools**: Connected on first query
- **HTTP Clients**: Created per-host on demand

### 2. Bounded Concurrency

Worker pool prevents resource exhaustion:

- Configurable number of workers (default: `runtime.NumCPU()`)
- Bounded task queue prevents memory bloat
- Graceful shutdown with drain timeout
- Per-task timeout prevents hanging

### 3. Event-Driven Communication

Loose coupling through events:

- Components publish events without knowing subscribers
- Subscribers filter events by type
- Async publishing prevents blocking
- Metrics tracked for observability

### 4. Multi-Tier Caching

Optimized cache hierarchy:

- L1 (memory): Sub-millisecond access, limited size
- L2 (Redis): Millisecond access, persistent, compressed
- Tag-based invalidation for related entries
- Event-driven invalidation on state changes

## Performance Characteristics

| Component | Latency | Throughput | Memory |
|-----------|---------|------------|--------|
| Worker Pool Submit | ~100ns | 1M+ ops/sec | O(queue size) |
| Event Bus Publish | ~1us | 100K+ events/sec | O(subscribers) |
| L1 Cache Get | ~50ns | 10M+ ops/sec | Configurable |
| L2 Cache Get | ~1ms | 10K+ ops/sec | Unbounded |
| HTTP Pool Get | ~10us | 100K+ ops/sec | O(hosts) |

## Thread Safety

All components are thread-safe:

- `sync.RWMutex` for read-heavy data structures
- `sync.atomic` for counters and metrics
- `sync.Once` for lazy initialization
- Channel-based communication for task/event passing

## Error Handling

- **Circuit Breaker**: Failed operations tracked, circuit opens on threshold
- **Retry with Backoff**: Transient failures retried with exponential backoff
- **Graceful Degradation**: System continues with reduced functionality
- **Event Notification**: Errors published as events for monitoring

## Monitoring Integration

All components expose metrics:

- Worker pool: active workers, queued tasks, completed/failed counts
- Event bus: published/delivered/dropped counts
- Cache: L1/L2 hits/misses, evictions, compression savings
- HTTP pool: connection count, request latency, error rates

## File Locations

| Component | Location |
|-----------|----------|
| Worker Pool | `internal/concurrency/worker_pool.go` |
| Event Bus | `internal/events/bus.go` |
| Tiered Cache | `internal/cache/tiered_cache.go` |
| HTTP Pool | `internal/http/pool.go` |
| Lazy Provider | `internal/llm/lazy_provider.go` |
| MCP Preinstaller | `internal/mcp/preinstaller.go` |
| MCP Connection Pool | `internal/mcp/connection_pool.go` |
