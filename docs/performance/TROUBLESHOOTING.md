# Performance Troubleshooting Guide

This guide helps diagnose and resolve common performance issues in HelixAgent.

## Quick Diagnostics

### Check System Health

```bash
# Get performance metrics
curl http://localhost:7061/v1/metrics/performance | jq .

# Check MCP server status
curl http://localhost:7061/v1/mcp/status | jq .

# View recent events
curl -N http://localhost:7061/v1/events/stream
```

### Run Performance Challenge

```bash
./challenges/scripts/performance_baseline_challenge.sh
```

## Common Issues

### 1. MCP Server Timeout

**Symptoms:**
- "Operation timed out" errors
- MCP tools fail to execute
- Long delays when using file system tools

**Causes:**
- NPM packages not pre-installed
- First-time `npx` download taking too long
- Network connectivity issues

**Solutions:**

1. **Enable pre-installation:**
   ```yaml
   performance:
     mcp:
       preinstall_on_startup: true
   ```

2. **Manually pre-install packages:**
   ```bash
   npm install -g @modelcontextprotocol/server-filesystem
   npm install -g @modelcontextprotocol/server-github
   npm install -g @modelcontextprotocol/server-memory
   npm install -g mcp-fetch
   npm install -g @modelcontextprotocol/server-puppeteer
   npm install -g mcp-server-sqlite
   ```

3. **Increase connection timeout:**
   ```yaml
   performance:
     mcp:
       connection_timeout: 30s
   ```

4. **Check Node.js is installed:**
   ```bash
   node --version  # Should be v18+
   npm --version
   ```

### 2. High Memory Usage

**Symptoms:**
- Memory usage growing over time
- OOM (Out of Memory) errors
- System becoming unresponsive

**Causes:**
- L1 cache too large
- Goroutine leaks
- Event bus subscribers not cleaned up

**Solutions:**

1. **Reduce L1 cache size:**
   ```yaml
   performance:
     cache:
       l1:
         max_size: 1000  # Reduce from default 10000
   ```

2. **Enable L1 cleanup:**
   ```yaml
   performance:
     cache:
       l1:
         cleanup_interval: 30s  # More frequent cleanup
   ```

3. **Check for goroutine leaks:**
   ```bash
   # Get goroutine count
   curl http://localhost:7061/debug/pprof/goroutine?debug=1 | head -1

   # Profile memory
   go tool pprof http://localhost:7061/debug/pprof/heap
   ```

4. **Verify event subscribers are unsubscribed:**
   ```go
   // Always unsubscribe when done
   ch := bus.Subscribe(events.EventCacheHit)
   defer bus.Unsubscribe(ch)
   ```

### 3. Slow Startup

**Symptoms:**
- Startup takes more than 5 seconds
- Timeouts during initialization
- Services unavailable immediately after start

**Causes:**
- Eager loading of all providers
- Database connection attempts at startup
- MCP server pre-connection

**Solutions:**

1. **Enable lazy loading:**
   ```yaml
   performance:
     lazy_loading:
       enabled: true
   ```

2. **Disable MCP pre-installation (for faster dev startup):**
   ```yaml
   performance:
     mcp:
       preinstall_on_startup: false
       lazy_connect: true
   ```

3. **Reduce preload services:**
   ```yaml
   performance:
     lazy_loading:
       preload_services: []  # Don't preload anything
   ```

4. **Profile startup:**
   ```bash
   time ./bin/helixagent --dry-run
   ```

### 4. Worker Pool Queue Full

**Symptoms:**
- "task queue is full" errors
- Tasks being rejected
- High latency on task submission

**Causes:**
- Too many concurrent requests
- Tasks taking too long to complete
- Not enough workers

**Solutions:**

1. **Increase queue size:**
   ```yaml
   performance:
     worker_pool:
       queue_size: 5000  # Increase from default 1000
   ```

2. **Increase worker count:**
   ```yaml
   performance:
     worker_pool:
       workers: 16  # Increase from default (CPU cores)
   ```

3. **Reduce task timeout (fail fast):**
   ```yaml
   performance:
     worker_pool:
       task_timeout: 10s  # Reduce from default 30s
   ```

4. **Check for slow tasks:**
   ```go
   metrics := pool.Metrics()
   fmt.Printf("Avg Latency: %v\n", metrics.AverageLatency())
   ```

### 5. Event Bus Backpressure

**Symptoms:**
- Events being dropped
- Subscribers missing events
- High `events_dropped` metric

**Causes:**
- Slow subscribers blocking the bus
- Buffer size too small
- Too many events being published

**Solutions:**

1. **Increase buffer size:**
   ```yaml
   performance:
     event_bus:
       buffer_size: 10000  # Increase from default 1000
   ```

2. **Reduce publish timeout:**
   ```yaml
   performance:
     event_bus:
       publish_timeout: 1ms  # Drop slow subscribers faster
   ```

3. **Use async publishing:**
   ```go
   bus.PublishAsync(event)  // Non-blocking
   ```

4. **Check subscriber performance:**
   ```go
   // Process events quickly
   for event := range ch {
       go handleEvent(event)  // Process async
   }
   ```

### 6. Cache Miss Rate High

**Symptoms:**
- Low hit rate (< 80%)
- Increased load on backend services
- Higher latency

**Causes:**
- TTL too short
- Cache size too small
- Too many unique keys

**Solutions:**

1. **Increase L1 cache size:**
   ```yaml
   performance:
     cache:
       l1:
         max_size: 50000
   ```

2. **Increase TTL:**
   ```yaml
   performance:
     cache:
       l1:
         ttl: 30m
       l2:
         ttl: 2h
   ```

3. **Use tags for efficient invalidation:**
   ```go
   // Tag related entries
   cache.Set(ctx, "user:123", user, ttl, "user", "active")
   cache.Set(ctx, "user:123:settings", settings, ttl, "user:123")

   // Invalidate all user:123 related entries
   cache.InvalidateByTag(ctx, "user:123")
   ```

4. **Check cache metrics:**
   ```bash
   curl http://localhost:7061/v1/metrics/performance | jq '.cache'
   ```

### 7. HTTP Connection Exhaustion

**Symptoms:**
- "too many open connections" errors
- Connection timeouts
- Slow external API calls

**Causes:**
- Connections not being reused
- Idle timeout too short
- Too many unique hosts

**Solutions:**

1. **Increase connection limits:**
   ```yaml
   performance:
     http_pool:
       max_idle_conns: 500
       max_conns_per_host: 50
   ```

2. **Increase idle timeout:**
   ```yaml
   performance:
     http_pool:
       idle_conn_timeout: 5m
   ```

3. **Use connection pooling:**
   ```go
   // Reuse clients per host
   client := pool.GetClient("api.example.com")
   ```

4. **Check connection count:**
   ```bash
   netstat -an | grep ESTABLISHED | wc -l
   ```

### 8. Race Conditions

**Symptoms:**
- Inconsistent behavior
- Occasional panics
- Data corruption

**Causes:**
- Concurrent access to shared state
- Missing synchronization
- Incorrect lock ordering

**Solutions:**

1. **Run race detector:**
   ```bash
   go test -race ./...
   ```

2. **Use atomic operations:**
   ```go
   atomic.AddInt64(&counter, 1)
   ```

3. **Use proper locking:**
   ```go
   mu.Lock()
   defer mu.Unlock()
   ```

4. **Run stress tests:**
   ```bash
   ./challenges/scripts/stress_resilience_challenge.sh
   ```

## Performance Monitoring

### Enable Detailed Logging

```yaml
logging:
  level: debug
  performance: true
  include_latencies: true
```

### Prometheus Metrics

```yaml
monitoring:
  prometheus:
    enabled: true
    path: /metrics
```

Key metrics to watch:
- `helixagent_worker_pool_active_workers`
- `helixagent_worker_pool_queued_tasks`
- `helixagent_event_bus_events_dropped`
- `helixagent_cache_hit_rate`
- `helixagent_http_pool_connections`
- `helixagent_mcp_connection_status`

### Grafana Dashboard

Import the HelixAgent performance dashboard:
```bash
curl -o helixagent-dashboard.json \
  https://raw.githubusercontent.com/helixagent/helixagent/main/dashboards/performance.json
```

## Diagnostic Commands

```bash
# Check all challenges
./challenges/scripts/run_all_challenges.sh

# Run specific performance challenge
./challenges/scripts/performance_baseline_challenge.sh
./challenges/scripts/parallel_execution_challenge.sh
./challenges/scripts/stress_resilience_challenge.sh

# Run Go benchmarks
go test -bench=. ./internal/concurrency/...
go test -bench=. ./internal/events/...
go test -bench=. ./internal/cache/...

# Profile CPU
go tool pprof http://localhost:7061/debug/pprof/profile?seconds=30

# Profile Memory
go tool pprof http://localhost:7061/debug/pprof/heap

# Check goroutines
curl http://localhost:7061/debug/pprof/goroutine?debug=2

# Trace execution
go tool trace <(curl http://localhost:7061/debug/pprof/trace?seconds=5)
```

## Getting Help

If you've tried the above solutions and still have issues:

1. Check the logs: `journalctl -u helixagent -f`
2. Run all diagnostics: `./challenges/scripts/run_all_challenges.sh`
3. Enable debug logging and capture output
4. Open an issue with diagnostic output at https://github.com/helixagent/helixagent/issues
