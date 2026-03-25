# User Manual 43: Performance Optimization Guide

## Overview

This guide covers the operational workflow for profiling, measuring, and optimizing HelixAgent performance: running benchmarks, enabling pprof, reading Prometheus dashboards, applying lazy loading, tuning the HTTP pool, and monitoring with Grafana.

## Prerequisites

- HelixAgent built and running (`make build && make run`)
- Prometheus and Grafana running (`docker compose --profile full up -d`)
- Go toolchain for `go tool pprof`
- `curl` and a web browser for dashboard access

## Step 1: Run Benchmarks

Establish a performance baseline before making changes:

```bash
# Run all benchmarks
make test-bench

# Run specific benchmarks with memory allocation stats
go test -bench=BenchmarkEnsemble -benchmem ./internal/services/...

# Compare before/after using benchstat
go install golang.org/x/perf/cmd/benchstat@latest
go test -bench=. -count=5 ./internal/services/... > before.txt
# (apply optimization)
go test -bench=. -count=5 ./internal/services/... > after.txt
benchstat before.txt after.txt
```

Key metrics to capture: ns/op (latency), B/op (memory per operation), allocs/op (allocation count).

## Step 2: Enable pprof Profiling

HelixAgent exposes pprof endpoints when running:

```bash
# CPU profile (30-second sample)
curl -o cpu.prof http://localhost:7061/debug/pprof/profile?seconds=30

# Heap memory profile
curl -o heap.prof http://localhost:7061/debug/pprof/heap

# Goroutine profile
curl -o goroutine.prof http://localhost:7061/debug/pprof/goroutine

# Block profile (contention)
curl -o block.prof http://localhost:7061/debug/pprof/block
```

Analyze with the interactive pprof tool:

```bash
go tool pprof -http=:6060 cpu.prof
```

This opens a web UI at `http://localhost:6060` with flame graphs, call graphs, and top functions by CPU/memory.

## Step 3: Read Prometheus Dashboards

HelixAgent exports metrics at `/metrics`. Key metrics to monitor:

| Metric | Type | Description |
|--------|------|-------------|
| `helixagent_http_requests_total` | Counter | Total HTTP requests by method, path, status |
| `helixagent_http_request_duration_seconds` | Histogram | Request latency distribution |
| `helixagent_provider_latency_seconds` | Histogram | Per-provider LLM response time |
| `helixagent_provider_errors_total` | Counter | Provider errors by type |
| `helixagent_goroutines` | Gauge | Current goroutine count |
| `helixagent_ensemble_voting_duration_seconds` | Histogram | Ensemble aggregation time |
| `helixagent_circuit_breaker_state` | Gauge | Circuit breaker state per provider (0=closed, 1=open) |

Query examples in Prometheus (`http://localhost:9090`):

```promql
# P99 request latency over 5 minutes
histogram_quantile(0.99, rate(helixagent_http_request_duration_seconds_bucket[5m]))

# Error rate by provider
rate(helixagent_provider_errors_total[5m])

# Goroutine trend
helixagent_goroutines
```

## Step 4: Apply Lazy Loading Patterns

HelixAgent uses `sync.Once` for thread-safe deferred initialization. Apply this pattern to expensive resources:

```go
type ExpensiveService struct {
    initOnce sync.Once
    client   *http.Client
}

func (s *ExpensiveService) getClient() *http.Client {
    s.initOnce.Do(func() {
        s.client = &http.Client{
            Timeout:   30 * time.Second,
            Transport: buildTransport(),
        }
    })
    return s.client
}
```

Where to apply lazy loading:
- LLM provider clients (initialized on first request, not at boot)
- Database connection pools (created when first query arrives)
- Cache connections (deferred until first cache operation)
- MCP adapter initialization (loaded when adapter is first invoked)

Validate with the challenge:

```bash
./challenges/scripts/lazy_loading_validation_challenge.sh
```

## Step 5: Tune the HTTP Connection Pool

HelixAgent's HTTP/3 client pool is configured in the HTTP adapter:

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 20,
    MaxConnsPerHost:     50,
    IdleConnTimeout:     90 * time.Second,
}
```

Tuning guidelines:

| Parameter | Low Traffic | High Traffic |
|-----------|-------------|--------------|
| `MaxIdleConns` | 50 | 200 |
| `MaxIdleConnsPerHost` | 10 | 40 |
| `MaxConnsPerHost` | 20 | 100 |
| `IdleConnTimeout` | 90s | 120s |

Monitor connection reuse via Prometheus: rising `helixagent_http_requests_total` without rising goroutine count indicates healthy pooling.

## Step 6: Monitor with Grafana

Access Grafana at `http://localhost:3000` (default credentials: `admin`/`admin`).

Recommended dashboard panels:

1. **Request Rate** -- `rate(helixagent_http_requests_total[1m])` as a time series
2. **Latency Heatmap** -- `helixagent_http_request_duration_seconds_bucket` as a heatmap
3. **Provider Health** -- `helixagent_circuit_breaker_state` as a stat panel per provider
4. **Goroutine Count** -- `helixagent_goroutines` as a time series (watch for leaks)
5. **Memory Usage** -- `process_resident_memory_bytes` as a time series
6. **Error Budget** -- `rate(helixagent_provider_errors_total[5m]) / rate(helixagent_http_requests_total[5m])` as a gauge

Set alerts for:
- P99 latency exceeding 5 seconds
- Goroutine count exceeding 10,000
- Error rate exceeding 5%
- Circuit breaker open for more than 5 minutes

## Step 7: Validate Optimizations

After applying changes, run the full validation suite:

```bash
# Benchmark comparison
make test-bench

# Memory profiling challenge
./challenges/scripts/pprof_memory_profiling_challenge.sh

# Monitoring dashboard challenge
./challenges/scripts/monitoring_dashboard_challenge.sh

# Resource limits enforcement
./challenges/scripts/resource_limits_challenge.sh
```

## Troubleshooting

- **pprof endpoint returns 404**: Ensure HelixAgent is running in debug mode (`make run-dev`) or pprof is explicitly enabled
- **Prometheus shows no data**: Verify the scrape target is configured in `prometheus.yml` with the correct port
- **Grafana shows "No Data"**: Check the Prometheus data source URL in Grafana settings
- **High goroutine count**: Use `go tool pprof goroutine.prof` to identify the source of leaked goroutines

## Related Resources

- [User Manual 11: Performance Tuning](11-performance-tuning.md) -- Server and database tuning
- [User Manual 18: Performance Monitoring](18-performance-monitoring.md) -- Metrics collection setup
- [User Manual 33: Performance Optimization Guide](33-performance-optimization-guide.md) -- Internal Go patterns (sync.Once, semaphores)
- Challenge scripts: `challenges/scripts/pprof_memory_profiling_challenge.sh`, `challenges/scripts/monitoring_dashboard_challenge.sh`
