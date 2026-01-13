# Performance Configuration Guide

This document describes how to configure performance optimization settings in HelixAgent.

## Configuration File

Performance settings are configured in `configs/production.yaml` or through environment variables.

## Worker Pool Configuration

```yaml
performance:
  worker_pool:
    workers: 8              # Number of concurrent workers (default: CPU cores)
    queue_size: 1000        # Maximum pending tasks
    task_timeout: 30s       # Maximum time per task
    shutdown_grace: 5s      # Grace period during shutdown
```

### Environment Variables

```bash
WORKER_POOL_WORKERS=8
WORKER_POOL_QUEUE_SIZE=1000
WORKER_POOL_TASK_TIMEOUT=30s
WORKER_POOL_SHUTDOWN_GRACE=5s
```

### Tuning Guidelines

| Workload | Workers | Queue Size | Task Timeout |
|----------|---------|------------|--------------|
| CPU-bound | NumCPU | 100 | 10s |
| I/O-bound | NumCPU * 2 | 1000 | 30s |
| Mixed | NumCPU * 1.5 | 500 | 20s |

## Event Bus Configuration

```yaml
performance:
  event_bus:
    buffer_size: 1000       # Subscriber channel buffer
    publish_timeout: 10ms   # Timeout for slow subscribers
    cleanup_interval: 30s   # Dead subscriber cleanup
    max_subscribers: 100    # Max subscribers per event type
```

### Environment Variables

```bash
EVENT_BUS_BUFFER_SIZE=1000
EVENT_BUS_PUBLISH_TIMEOUT=10ms
EVENT_BUS_CLEANUP_INTERVAL=30s
EVENT_BUS_MAX_SUBSCRIBERS=100
```

### Tuning Guidelines

- **High throughput**: Increase buffer_size to 10000
- **Low latency**: Decrease publish_timeout to 1ms
- **Memory constrained**: Decrease buffer_size to 100

## Tiered Cache Configuration

```yaml
performance:
  cache:
    l1:
      max_size: 10000       # Maximum items in memory
      ttl: 5m               # Memory cache TTL
      cleanup_interval: 1m  # Expired entry cleanup
    l2:
      ttl: 30m              # Redis cache TTL
      compression: true     # Enable gzip compression
      key_prefix: "tiered:" # Redis key prefix
    negative_ttl: 30s       # Cache negative results
    enable_l1: true
    enable_l2: true
```

### Environment Variables

```bash
CACHE_L1_MAX_SIZE=10000
CACHE_L1_TTL=5m
CACHE_L1_CLEANUP_INTERVAL=1m
CACHE_L2_TTL=30m
CACHE_L2_COMPRESSION=true
CACHE_L2_KEY_PREFIX=tiered:
CACHE_NEGATIVE_TTL=30s
```

### Tuning Guidelines

| Scenario | L1 Size | L1 TTL | L2 TTL | Compression |
|----------|---------|--------|--------|-------------|
| Memory constrained | 1000 | 1m | 10m | true |
| High cache hit rate | 50000 | 10m | 1h | true |
| Low latency | 10000 | 5m | 30m | false |

## HTTP Client Pool Configuration

```yaml
performance:
  http_pool:
    max_idle_conns: 100         # Total idle connections
    max_conns_per_host: 10      # Connections per host
    idle_conn_timeout: 90s      # Idle connection lifetime
    response_header_timeout: 10s # Header read timeout
    dial_timeout: 5s            # Connection timeout

    retry:
      max_retries: 3            # Maximum retry attempts
      initial_backoff: 100ms    # First retry delay
      max_backoff: 5s           # Maximum retry delay
      backoff_multiplier: 2.0   # Exponential factor
```

### Environment Variables

```bash
HTTP_POOL_MAX_IDLE_CONNS=100
HTTP_POOL_MAX_CONNS_PER_HOST=10
HTTP_POOL_IDLE_CONN_TIMEOUT=90s
HTTP_POOL_RESPONSE_HEADER_TIMEOUT=10s
HTTP_POOL_DIAL_TIMEOUT=5s
HTTP_RETRY_MAX_RETRIES=3
HTTP_RETRY_INITIAL_BACKOFF=100ms
HTTP_RETRY_MAX_BACKOFF=5s
```

## MCP Server Configuration

```yaml
performance:
  mcp:
    preinstall_on_startup: true  # Pre-install NPM packages
    connection_pool_size: 12     # Max concurrent MCP connections
    connection_timeout: 5s       # Connection establishment timeout
    lazy_connect: true           # Connect on first use
    warmup_servers:              # Pre-connect these servers
      - filesystem
      - github
```

### Environment Variables

```bash
MCP_PREINSTALL_ON_STARTUP=true
MCP_CONNECTION_POOL_SIZE=12
MCP_CONNECTION_TIMEOUT=5s
MCP_LAZY_CONNECT=true
MCP_WARMUP_SERVERS=filesystem,github
```

### Standard MCP Packages

| Server | NPM Package | Purpose |
|--------|-------------|---------|
| filesystem | @modelcontextprotocol/server-filesystem | File operations |
| github | @modelcontextprotocol/server-github | GitHub API |
| memory | @modelcontextprotocol/server-memory | Persistent memory |
| fetch | mcp-fetch | HTTP requests |
| puppeteer | @modelcontextprotocol/server-puppeteer | Browser automation |
| sqlite | mcp-server-sqlite | SQLite database |

## Lazy Loading Configuration

```yaml
performance:
  lazy_loading:
    enabled: true
    preload_services:           # Services to initialize early
      - provider_registry
      - mcp_pool
    init_timeout: 10s           # Maximum initialization time
```

### Environment Variables

```bash
LAZY_LOADING_ENABLED=true
LAZY_LOADING_PRELOAD_SERVICES=provider_registry,mcp_pool
LAZY_LOADING_INIT_TIMEOUT=10s
```

## Complete Example Configuration

```yaml
# configs/performance.yaml
performance:
  worker_pool:
    workers: 8
    queue_size: 1000
    task_timeout: 30s
    shutdown_grace: 5s

  event_bus:
    buffer_size: 1000
    publish_timeout: 10ms
    cleanup_interval: 30s
    max_subscribers: 100

  cache:
    l1:
      max_size: 10000
      ttl: 5m
      cleanup_interval: 1m
    l2:
      ttl: 30m
      compression: true
      key_prefix: "tiered:"
    negative_ttl: 30s
    enable_l1: true
    enable_l2: true

  http_pool:
    max_idle_conns: 100
    max_conns_per_host: 10
    idle_conn_timeout: 90s
    response_header_timeout: 10s
    dial_timeout: 5s
    retry:
      max_retries: 3
      initial_backoff: 100ms
      max_backoff: 5s
      backoff_multiplier: 2.0

  mcp:
    preinstall_on_startup: true
    connection_pool_size: 12
    connection_timeout: 5s
    lazy_connect: true
    warmup_servers:
      - filesystem
      - github

  lazy_loading:
    enabled: true
    preload_services:
      - provider_registry
      - mcp_pool
    init_timeout: 10s
```

## Profile-Based Configuration

### Development Profile

```yaml
performance:
  worker_pool:
    workers: 2
    queue_size: 100
  cache:
    l1:
      max_size: 100
    enable_l2: false
  mcp:
    preinstall_on_startup: false
    lazy_connect: true
```

### Production Profile

```yaml
performance:
  worker_pool:
    workers: 16
    queue_size: 5000
  cache:
    l1:
      max_size: 50000
    l2:
      compression: true
  mcp:
    preinstall_on_startup: true
    lazy_connect: true
```

### High-Throughput Profile

```yaml
performance:
  worker_pool:
    workers: 32
    queue_size: 10000
  event_bus:
    buffer_size: 10000
  cache:
    l1:
      max_size: 100000
  http_pool:
    max_idle_conns: 500
    max_conns_per_host: 50
```

## Monitoring Configuration

Enable performance metrics:

```yaml
monitoring:
  enabled: true
  metrics_path: /metrics
  include_performance: true
  performance_metrics:
    - worker_pool_active_workers
    - worker_pool_queued_tasks
    - worker_pool_completed_tasks
    - event_bus_events_published
    - event_bus_events_delivered
    - cache_l1_hits
    - cache_l1_misses
    - cache_l2_hits
    - cache_l2_misses
    - http_pool_connections
    - mcp_active_connections
```
