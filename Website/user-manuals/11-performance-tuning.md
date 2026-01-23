# User Manual 11: Performance Tuning Guide

## Introduction

This guide covers performance optimization techniques for HelixAgent, including caching strategies, connection pooling, batch processing, and resource management.

## Performance Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Performance Layers                            │
├─────────────────────────────────────────────────────────────────┤
│  Request Layer   │ Rate limiting, Load balancing, Compression   │
├─────────────────────────────────────────────────────────────────┤
│  Caching Layer   │ L1 (Memory), L2 (Redis), Response cache      │
├─────────────────────────────────────────────────────────────────┤
│  Processing      │ Parallel execution, Batch processing         │
├─────────────────────────────────────────────────────────────────┤
│  Provider Layer  │ Connection pooling, Circuit breakers         │
├─────────────────────────────────────────────────────────────────┤
│  Storage Layer   │ Database pooling, Query optimization         │
└─────────────────────────────────────────────────────────────────┘
```

## Caching

### Two-Tier Cache Configuration

```yaml
cache:
  enabled: true

  # L1 - In-memory cache (fast, limited size)
  l1:
    enabled: true
    max_size: 10000        # entries
    ttl: 5m
    eviction: lru

  # L2 - Redis cache (larger, shared)
  l2:
    enabled: true
    host: redis:6379
    password: ${REDIS_PASSWORD}
    db: 0
    pool_size: 20
    max_retries: 3

  # Cache settings by type
  policies:
    completions:
      ttl: 30m
      max_size: 5000
    embeddings:
      ttl: 24h
      max_size: 50000
    debates:
      ttl: 1h
      max_size: 1000
```

### Cache Key Strategies

```go
// Good: Include all relevant parameters
key := cache.CompletionKey(provider, model, prompt, temperature, maxTokens)

// Bad: Missing parameters that affect output
key := cache.CompletionKey(provider, prompt)  // Missing model, temp
```

### Cache Warming

Pre-populate cache with common queries:

```yaml
cache:
  warming:
    enabled: true
    on_startup: true
    queries:
      - prompt: "Hello, how can I help?"
        providers: ["claude", "gemini"]
      - prompt: "What is HelixAgent?"
        providers: ["claude"]
```

### Cache Invalidation

```bash
# Clear all cache
curl -X POST http://localhost:8080/v1/admin/cache/clear \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Clear by pattern
curl -X POST http://localhost:8080/v1/admin/cache/clear \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"pattern": "completion:claude:*"}'

# Clear expired entries
curl -X POST http://localhost:8080/v1/admin/cache/cleanup \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Connection Pooling

### Database Connection Pool

```yaml
database:
  host: localhost
  port: 5432
  name: helixagent

  pool:
    max_open: 50          # Max open connections
    max_idle: 10          # Max idle connections
    max_lifetime: 1h      # Connection max lifetime
    max_idle_time: 15m    # Idle timeout
    health_check: 30s     # Health check interval
```

### Redis Connection Pool

```yaml
redis:
  host: localhost
  port: 6379

  pool:
    size: 100             # Pool size
    min_idle: 10          # Min idle connections
    max_retries: 3        # Retry attempts
    timeout: 5s           # Operation timeout
    idle_timeout: 5m      # Idle connection timeout
```

### HTTP Client Pool

```yaml
http_client:
  pool:
    max_connections: 100           # Per host
    max_idle_connections: 20       # Per host
    idle_timeout: 90s
    timeout: 30s
    keep_alive: true
```

## Batch Processing

### Batch Completions

Process multiple requests together:

```bash
POST /v1/completions/batch
Content-Type: application/json

{
  "requests": [
    {"prompt": "Translate to French: Hello"},
    {"prompt": "Translate to French: Goodbye"},
    {"prompt": "Translate to French: Thank you"}
  ],
  "batch_options": {
    "parallel": true,
    "max_concurrent": 5
  }
}
```

### Batch Embeddings

```bash
POST /v1/embeddings/batch
Content-Type: application/json

{
  "texts": [
    "Document 1 content...",
    "Document 2 content...",
    "Document 3 content..."
  ],
  "model": "text-embedding-3-small",
  "batch_size": 100
}
```

### Background Processing

For large batches, use async processing:

```bash
POST /v1/tasks
Content-Type: application/json

{
  "type": "batch_completion",
  "payload": {
    "requests": [...],
    "callback_url": "https://example.com/webhook"
  },
  "priority": 5
}

Response:
{
  "task_id": "task-123",
  "status": "pending"
}
```

## Parallel Execution

### Provider Parallelism

```yaml
providers:
  parallel:
    enabled: true
    max_concurrent: 10     # Across all providers
    per_provider: 3        # Per provider limit

  # Individual provider limits
  claude:
    max_concurrent: 5
    timeout: 60s
  gemini:
    max_concurrent: 5
    timeout: 45s
```

### Ensemble Parallelism

```yaml
ensemble:
  parallel_execution: true
  max_providers: 3
  timeout: 30s
  fail_fast: false         # Continue on individual failures
```

## Streaming Optimization

### Response Streaming

Enable streaming for faster time-to-first-token:

```bash
POST /v1/completions
Content-Type: application/json

{
  "prompt": "Write a long story...",
  "stream": true
}

Response (SSE):
data: {"content": "Once", "done": false}
data: {"content": " upon", "done": false}
data: {"content": " a time", "done": false}
...
data: {"done": true, "usage": {...}}
```

### Streaming Configuration

```yaml
streaming:
  enabled: true
  buffer_size: 4096        # Bytes
  flush_interval: 100ms
  compression: gzip        # For large responses
```

## Resource Management

### Memory Configuration

```yaml
resources:
  memory:
    max_heap: 4G
    gc_percent: 100        # Go GC tuning
    soft_limit: 3G         # Trigger GC at this limit

  # Request memory limits
  requests:
    max_request_body: 10MB
    max_response_body: 50MB
    max_prompt_tokens: 100000
```

### CPU Configuration

```yaml
resources:
  cpu:
    gomaxprocs: 0          # 0 = auto (use all CPUs)
    worker_threads: 50     # Background worker count
```

### Timeouts

```yaml
timeouts:
  # Server timeouts
  read: 30s
  write: 60s
  idle: 120s

  # Provider timeouts
  provider_default: 60s
  provider_streaming: 300s

  # Internal timeouts
  database: 10s
  cache: 5s
```

## Query Optimization

### Database Queries

```yaml
database:
  query:
    # Enable query logging for slow queries
    slow_query_log: true
    slow_query_threshold: 100ms

    # Connection settings
    statement_cache: 1000
    prepared_statements: true
```

### Indexing Strategy

Ensure proper indexes exist:

```sql
-- Completions table
CREATE INDEX idx_completions_session ON completions(session_id);
CREATE INDEX idx_completions_created ON completions(created_at DESC);
CREATE INDEX idx_completions_provider ON completions(provider_id);

-- Background tasks
CREATE INDEX idx_tasks_pending ON background_tasks(priority DESC, created_at)
  WHERE status = 'pending';
```

## Load Balancing

### Horizontal Scaling

```yaml
# docker-compose.yml
services:
  helixagent:
    image: helixagent:latest
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2'
          memory: 4G

  nginx:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
```

### Nginx Configuration

```nginx
upstream helixagent {
    least_conn;  # Load balancing method
    server helixagent1:8080 weight=5;
    server helixagent2:8080 weight=5;
    server helixagent3:8080 weight=5;
    keepalive 32;
}

server {
    listen 80;

    location / {
        proxy_pass http://helixagent;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_connect_timeout 5s;
        proxy_read_timeout 60s;
    }
}
```

## Monitoring Performance

### Key Metrics

```yaml
metrics:
  # Request metrics
  - helixagent_request_duration_seconds
  - helixagent_request_size_bytes
  - helixagent_response_size_bytes

  # Cache metrics
  - helixagent_cache_hits_total
  - helixagent_cache_misses_total
  - helixagent_cache_latency_seconds

  # Provider metrics
  - helixagent_provider_latency_seconds
  - helixagent_provider_tokens_total
  - helixagent_provider_errors_total

  # Resource metrics
  - go_memstats_alloc_bytes
  - go_goroutines
  - process_cpu_seconds_total
```

### Performance Alerts

```yaml
alerts:
  - name: high_latency
    condition: "p99_latency > 5s"
    action: notify

  - name: cache_miss_rate
    condition: "cache_miss_rate > 0.5"
    action: investigate

  - name: memory_pressure
    condition: "memory_usage > 0.9"
    action: scale_up
```

## Performance Testing

### Load Testing

```bash
# Using hey
hey -n 1000 -c 50 -m POST \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello"}' \
  http://localhost:8080/v1/completions

# Using wrk
wrk -t12 -c400 -d30s \
  -s post.lua \
  http://localhost:8080/v1/completions
```

### Benchmark Suite

```bash
# Run benchmarks
make test-bench

# Specific benchmarks
go test -bench=BenchmarkCompletion -benchmem ./internal/handlers/...
go test -bench=BenchmarkCache -benchmem ./internal/cache/...
```

## Optimization Checklist

### Quick Wins

- [ ] Enable caching (L1 + L2)
- [ ] Configure connection pooling
- [ ] Enable response compression
- [ ] Set appropriate timeouts
- [ ] Enable streaming

### Medium Effort

- [ ] Tune database indexes
- [ ] Configure parallel execution
- [ ] Set up batch processing
- [ ] Optimize cache TTLs
- [ ] Configure rate limiting

### Advanced

- [ ] Horizontal scaling
- [ ] Custom cache warming
- [ ] Provider-specific tuning
- [ ] Memory profiling
- [ ] Query optimization

## Troubleshooting

### High Latency

1. Check provider response times
2. Verify cache hit rates
3. Review database query times
4. Check connection pool exhaustion

### Memory Issues

1. Profile memory usage
2. Check for goroutine leaks
3. Review cache sizes
4. Analyze heap allocations

### Connection Exhaustion

1. Increase pool sizes
2. Check for connection leaks
3. Review timeout settings
4. Monitor active connections

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
