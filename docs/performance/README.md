# Performance Optimization Guide

This directory contains documentation for performance optimization, benchmarking, and profiling in HelixAgent.

## Overview

HelixAgent implements a multi-layered performance optimization strategy designed for high-throughput AI workloads. This guide covers architecture decisions, configuration options, benchmarking methodologies, and optimization techniques.

## Documentation Index

| Document | Description |
|----------|-------------|
| [Architecture](./ARCHITECTURE.md) | Performance optimization architecture and component diagrams |
| [Configuration](./CONFIGURATION.md) | Performance tuning configuration guide |
| [API Reference](./API.md) | Performance-related API documentation |
| [Troubleshooting](./TROUBLESHOOTING.md) | Diagnosing and resolving performance issues |

## Performance Architecture

HelixAgent's performance optimization system consists of several key components:

### 1. MCP Pre-Installation and Lazy Loading

- NPM packages pre-installed at startup
- Connections established lazily on first use
- Background preinstallation of common packages

### 2. Worker Pool

- Bounded concurrency with configurable workers
- Task queue with overflow management
- Graceful shutdown support

### 3. Event Bus

- Asynchronous pub/sub communication
- Configurable buffer sizes
- Dead subscriber cleanup

### 4. Tiered Cache

- L1: In-memory cache for hot data
- L2: Redis cache for distributed caching
- Compression for large payloads

### 5. HTTP Client Pool

- Connection reuse across requests
- Configurable retry logic
- Circuit breaker integration

### 6. Lazy Provider Initialization

- Providers initialized on first use
- Reduced startup time
- Memory-efficient operation

## Benchmarking

### Running Benchmarks

```bash
# Run all benchmarks
make test-bench

# Run specific benchmark
go test -bench=BenchmarkEnsemble -benchmem ./internal/llm/...

# Run with profiling
go test -bench=BenchmarkEnsemble -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/llm/...
```

### Benchmark Categories

| Category | Location | Description |
|----------|----------|-------------|
| LLM Provider | `internal/llm/providers/*/` | Provider response times |
| Ensemble | `internal/llm/` | Ensemble strategy performance |
| Cache | `internal/cache/` | Cache hit/miss rates |
| Database | `internal/database/` | Query performance |
| HTTP | `internal/handlers/` | API endpoint latency |

### Benchmark Results

Typical benchmark results on reference hardware (8-core CPU, 32GB RAM):

| Operation | p50 | p95 | p99 |
|-----------|-----|-----|-----|
| Chat completion (single provider) | 150ms | 450ms | 800ms |
| Ensemble voting (3 providers) | 500ms | 1200ms | 2000ms |
| Cache hit | 0.1ms | 0.5ms | 1ms |
| Database query | 5ms | 15ms | 50ms |
| MCP tool execution | 50ms | 200ms | 500ms |

## Profiling Guide

### CPU Profiling

```bash
# Start server with profiling enabled
./bin/helixagent --enable-profiling

# Collect CPU profile
curl http://localhost:7061/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze with pprof
go tool pprof -http=:8080 cpu.prof
```

### Memory Profiling

```bash
# Collect heap profile
curl http://localhost:7061/debug/pprof/heap > heap.prof

# Analyze memory allocations
go tool pprof -http=:8080 heap.prof
```

### Trace Analysis

```bash
# Collect execution trace
curl http://localhost:7061/debug/pprof/trace?seconds=5 > trace.out

# View trace
go tool trace trace.out
```

## Optimization Techniques

### 1. Connection Pooling

Configure HTTP client pools for external services:

```yaml
performance:
  http_client:
    max_idle_conns: 100
    max_conns_per_host: 10
    idle_timeout: 90s
```

### 2. Caching Strategy

Implement multi-tier caching:

```yaml
performance:
  cache:
    l1_size: 10000        # In-memory entries
    l1_ttl: 5m            # L1 TTL
    l2_enabled: true      # Redis L2 cache
    l2_ttl: 30m           # L2 TTL
    compression: true     # Compress large values
```

### 3. Worker Pool Tuning

Optimize worker pool for your workload:

```yaml
performance:
  worker_pool:
    workers: 8            # Match CPU cores for CPU-bound
    queue_size: 1000      # Buffer for burst traffic
    task_timeout: 30s     # Prevent stuck tasks
```

### 4. Rate Limiting

Protect against overload:

```yaml
rate_limiting:
  requests_per_second: 100
  burst_size: 200
  per_client_limit: 10
```

### 5. Database Optimization

- Use connection pooling
- Index frequently queried columns
- Implement query result caching
- Use prepared statements

### 6. Provider Timeout Configuration

Set appropriate timeouts per provider:

```yaml
providers:
  claude:
    timeout: 60s
    retry_count: 2
  deepseek:
    timeout: 45s
    retry_count: 3
```

## Monitoring Performance

### Metrics Endpoints

```bash
# Get performance metrics
curl http://localhost:7061/v1/metrics/performance | jq .

# Check system health
curl http://localhost:7061/v1/monitoring/status | jq .

# View circuit breaker state
curl http://localhost:7061/v1/monitoring/circuit-breakers | jq .
```

### Prometheus Metrics

Key metrics exposed for monitoring:

| Metric | Type | Description |
|--------|------|-------------|
| `helixagent_request_duration_seconds` | Histogram | Request latency |
| `helixagent_cache_hits_total` | Counter | Cache hit count |
| `helixagent_provider_errors_total` | Counter | Provider errors |
| `helixagent_worker_pool_queue_size` | Gauge | Current queue depth |
| `helixagent_active_connections` | Gauge | Active connections |

### Performance Alerts

Configure alerts for performance degradation:

```yaml
alerts:
  - name: high_latency
    condition: "p99_latency > 2s"
    severity: warning
  - name: error_rate
    condition: "error_rate > 5%"
    severity: critical
```

## Performance Challenge

Run the performance baseline challenge to validate optimization:

```bash
./challenges/scripts/performance_baseline_challenge.sh
```

## Related Documentation

- [Architecture Overview](../architecture/README.md)
- [Configuration Guide](../configuration/README.md)
- [Monitoring Guide](../operations/MONITORING.md)
