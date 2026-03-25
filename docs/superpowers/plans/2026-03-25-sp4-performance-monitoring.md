# SP4: Performance & Monitoring — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every service instrumented with metrics, monitoring tests validate collection accuracy, lazy loading everywhere possible, benchmarks produce documented baselines.

**Architecture:** Add sync.Once lazy initialization to 6 subsystems (handlers, MCP adapters, formatters, VectorDB, embeddings, BigData), verify and improve 5 non-blocking operations, create 10 monitoring validation tests, establish 10 benchmark baselines, add backpressure mechanisms to 4 hot paths.

**Tech Stack:** Go 1.25.3, Prometheus client_golang, OpenTelemetry, sync.Once, testify v1.11.1

**Spec:** `docs/superpowers/specs/2026-03-25-comprehensive-completion-design.md` (SP4 section)

**Depends on:** SP1 complete. Can run concurrently with SP2 and SP3.

---

### Task 1: Lazy-Load Handler Service Initialization

**Files:**
- Modify: `internal/router/router.go` (lines 1117-1204)

The 7 handlers currently initialized with nil services (SP1 documents them, doesn't wire them). Add lazy initialization so services are created on first request.

- [ ] **Step 1: Read how services are created elsewhere in the codebase**

Search for existing service constructors:
```bash
grep -rn 'NewDiscoveryService\|NewScoringService\|NewHealthService' internal/services/ --include='*.go' | head -10
```

- [ ] **Step 2: Create a lazy service provider using sync.Once per service**

```go
// internal/router/lazy_services.go
package router

import "sync"

// LazyService wraps a factory function with sync.Once for thread-safe lazy init.
type LazyService struct {
    once    sync.Once
    service interface{}
    factory func() interface{}
}

func NewLazyService(factory func() interface{}) *LazyService {
    return &LazyService{factory: factory}
}

func (ls *LazyService) Get() interface{} {
    ls.once.Do(func() {
        ls.service = ls.factory()
    })
    return ls.service
}

// LazyServiceProvider manages named lazy services.
type LazyServiceProvider struct {
    mu       sync.RWMutex
    services map[string]*LazyService
}

func NewLazyServiceProvider() *LazyServiceProvider {
    return &LazyServiceProvider{
        services: make(map[string]*LazyService),
    }
}

func (p *LazyServiceProvider) Register(name string, factory func() interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.services[name] = NewLazyService(factory)
}

func (p *LazyServiceProvider) Get(name string) interface{} {
    p.mu.RLock()
    svc, ok := p.services[name]
    p.mu.RUnlock()
    if !ok {
        return nil
    }
    return svc.Get()
}
```

- [ ] **Step 3: Write test for lazy service provider**

```go
// internal/router/lazy_services_test.go
func TestLazyServiceProvider_InitializesOnce(t *testing.T) {
    callCount := 0
    p := NewLazyServiceProvider()
    p.Register("test", func() interface{} {
        callCount++
        return "service"
    })

    result1 := p.Get("test")
    result2 := p.Get("test")
    assert.Equal(t, "service", result1)
    assert.Equal(t, "service", result2)
    assert.Equal(t, 1, callCount) // Factory called only once
}
```

- [ ] **Step 4: Run tests and commit**

```bash
go test ./internal/router/ -run TestLazyServiceProvider -v -count=1
git add internal/router/lazy_services.go internal/router/lazy_services_test.go
git commit -m "feat(router): add lazy service provider for on-demand handler initialization"
```

---

### Task 2: Lazy-Load MCP Adapters

**Files:**
- Modify: `internal/mcp/adapters/` (registry or initialization code)

- [ ] **Step 1: Find MCP adapter registration**

```bash
grep -rn 'Register\|Init\|New.*Adapter' internal/mcp/adapters/ --include='*.go' | head -20
```

- [ ] **Step 2: Wrap each adapter initialization in sync.Once**

Each adapter gets a `sync.Once` guard so it's only initialized when first accessed via `/v1/mcp`:

```go
type LazyAdapter struct {
    once    sync.Once
    adapter MCPAdapter
    factory func() (MCPAdapter, error)
    err     error
}

func (la *LazyAdapter) Get() (MCPAdapter, error) {
    la.once.Do(func() {
        la.adapter, la.err = la.factory()
    })
    return la.adapter, la.err
}
```

- [ ] **Step 3: Write test verifying lazy behavior**

- [ ] **Step 4: Run tests and commit**

```bash
go test ./internal/mcp/... -short -count=1
git add internal/mcp/
git commit -m "perf(mcp): lazy-load adapters on first access instead of startup"
```

---

### Task 3: Lazy-Load Formatter Registry

**Files:**
- Modify: `internal/formatters/registry.go` (or equivalent)

- [ ] **Step 1: Read current registration pattern**

- [ ] **Step 2: Wrap formatter initialization in sync.Once per language**

Formatters are only initialized when `POST /v1/format` requests formatting for that specific language.

- [ ] **Step 3: Write test and commit**

```bash
go test ./internal/formatters/ -short -count=1
git add internal/formatters/
git commit -m "perf(formatters): lazy-load formatters per language on first format request"
```

---

### Task 4: Lazy-Load VectorDB and Embedding Connections

**Files:**
- Modify: `internal/vectordb/` (connection initialization)
- Modify: `internal/embedding/` (provider initialization)

- [ ] **Step 1: Add sync.Once to VectorDB connection per store**

```go
type LazyVectorStore struct {
    once  sync.Once
    store VectorStore
    err   error
    cfg   VectorStoreConfig
}
```

- [ ] **Step 2: Add sync.Once to each embedding provider**

Same pattern — each of 6 providers (OpenAI, Cohere, Voyage, Jina, Google, Bedrock) initializes on first call.

- [ ] **Step 3: Write tests and commit**

```bash
go test ./internal/vectordb/ -short -count=1
go test ./internal/embedding/ -short -count=1
git add internal/vectordb/ internal/embedding/
git commit -m "perf(vectordb,embedding): lazy-load connections and provider clients"
```

---

### Task 5: Guard BigData Components Behind Env Checks

**Files:**
- Modify: `internal/bigdata/integration.go`

- [ ] **Step 1: Read current initialization**

Check if `BIGDATA_ENABLE_*` env vars are already checked. The spec notes graceful degradation exists.

- [ ] **Step 2: Add sync.Once per component**

Each BigData component (Neo4j, ClickHouse, Kafka) gets a `sync.Once` that only fires when the component is actually used AND the env var is true.

- [ ] **Step 3: Write test and commit**

```bash
go test ./internal/bigdata/ -short -count=1
git add internal/bigdata/
git commit -m "perf(bigdata): add sync.Once lazy initialization per component"
```

---

### Task 6: Create Monitoring Validation Tests

**Files:**
- Create: `tests/monitoring/metrics_validation_test.go`

- [ ] **Step 1: Write TestPrometheusMetricsRegistered**

```go
//go:build integration

package monitoring

import (
    "testing"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/stretchr/testify/assert"
)

func TestPrometheusMetricsRegistered(t *testing.T) {
    // Gather all registered metrics
    families, err := prometheus.DefaultGatherer.Gather()
    assert.NoError(t, err)

    metricNames := make(map[string]bool)
    for _, f := range families {
        metricNames[*f.Name] = true
    }

    // Verify expected metrics exist
    expected := []string{
        "helixagent_circuit_breaker_open_total",
        "helixagent_provider_health",
        "helixagent_provider_health_check_duration_seconds",
    }
    for _, name := range expected {
        assert.True(t, metricNames[name], "metric %s should be registered", name)
    }
}
```

- [ ] **Step 2: Write remaining 9 monitoring tests**

Create test functions for:
- `TestMetricsIncrementOnRequest`
- `TestCircuitBreakerMetricsOnFailure`
- `TestLatencyHistogramBuckets`
- `TestTokenCountAccuracy`
- `TestCacheHitRatioMetric`
- `TestProviderHealthMetricTransitions`
- `TestConcurrencyMonitorSaturation`
- `TestDebateMetricsRoundCounting`
- `TestResourceMonitorMemoryAccuracy`

Each test exercises the corresponding code path and verifies the metric value changes.

- [ ] **Step 3: Run tests**

Run: `go test ./tests/monitoring/ -tags integration -v -count=1`

- [ ] **Step 4: Commit**

```bash
git add tests/monitoring/metrics_validation_test.go
git commit -m "test(monitoring): add 10 monitoring validation tests for metric accuracy"
```

---

### Task 7: Create Benchmark Suite

**Files:**
- Create: `tests/performance/core_benchmarks_test.go`

- [ ] **Step 1: Write 10 benchmarks**

```go
//go:build performance

package performance

import "testing"

func BenchmarkEnsembleVoting(b *testing.B) {
    // Setup ensemble with 5 mock providers
    for i := 0; i < b.N; i++ {
        // Execute voting
    }
}

func BenchmarkProviderSelection(b *testing.B) { /* 43 providers */ }
func BenchmarkHTTPPoolAcquire(b *testing.B) { /* pool.GetClient() */ }
func BenchmarkCacheReadWrite(b *testing.B) { /* cache set + get */ }
func BenchmarkCircuitBreakerState(b *testing.B) { /* atomic state read */ }
func BenchmarkToolSchemaValidation(b *testing.B) { /* full schema validation */ }
func BenchmarkFormatterLookup(b *testing.B) { /* registry lookup */ }
func BenchmarkSkillsMatch(b *testing.B) { /* match across 48 agents */ }
func BenchmarkMCPAdapterResolve(b *testing.B) { /* adapter resolution */ }
func BenchmarkConsensusDetection(b *testing.B) { /* 25 responses */ }
```

- [ ] **Step 2: Run benchmarks and capture baselines**

Run: `GOMAXPROCS=2 nice -n 19 go test ./tests/performance/ -tags performance -bench=. -benchmem -count=3 | tee /tmp/benchmark-baselines.txt`

- [ ] **Step 3: Document baselines**

Create `docs/performance/BENCHMARKS.md` with results table.

- [ ] **Step 4: Commit**

```bash
git add tests/performance/core_benchmarks_test.go docs/performance/BENCHMARKS.md
git commit -m "perf(benchmark): add 10 core benchmarks with documented baselines"
```

---

### Task 8: Non-Blocking Improvements (Spec Section 4.4)

**Files:**
- Audit/Modify: `internal/services/debate_service.go` (debate log persistence)
- Audit/Modify: `internal/services/constitution_watcher.go` (fsnotify vs polling)
- Audit/Modify: `internal/services/provider_health_monitor.go` (goroutine pool)
- Audit/Modify: `internal/notifications/sse_manager.go` (non-blocking add/remove)
- Audit/Modify: `internal/cache/` (cache warming)

- [ ] **Step 1: Verify debate log persistence write-behind buffer is bounded**

Read `internal/services/debate_service.go` or `internal/database/debate_log_repository.go`. Find the log write channel. Verify it has a bounded buffer. If unbounded, add a cap.

- [ ] **Step 2: Verify Constitution watcher uses fsnotify**

Read `internal/services/constitution_watcher.go`. Check if it uses `fsnotify` (inotify-based) or periodic `os.Stat`. If periodic, switch to `fsnotify` for event-driven file watching:

```go
import "github.com/fsnotify/fsnotify"
```

- [ ] **Step 3: Verify provider health checks use dedicated goroutine pool**

Read `internal/services/provider_health_monitor.go`. Confirm health checks run in a dedicated goroutine pool (not on the request path). If they block request handlers, move to a background worker.

- [ ] **Step 4: Verify SSE client registration is non-blocking**

Read `internal/notifications/sse_manager.go`. Ensure client add/remove uses select-with-default pattern to avoid blocking the event loop:

```go
select {
case m.addClient <- client:
default:
    // Log warning: client registration queue full
}
```

- [ ] **Step 5: Add bounded concurrency to cache warming**

If cache warming exists, ensure it uses a semaphore to limit concurrent prefetch operations. If it doesn't exist yet, skip (not a regression).

- [ ] **Step 6: Run tests and commit**

```bash
go test ./internal/services/ ./internal/notifications/ ./internal/cache/ -short -count=1
git add internal/services/ internal/notifications/ internal/cache/
git commit -m "perf(non-blocking): verify and improve non-blocking patterns in 5 subsystems"
```

---

### Task 9: Add Backpressure Mechanisms

**Files:**
- Modify: `internal/services/debate_performance_optimizer.go` (exponential backoff)
- Modify: `internal/notifications/sse_manager.go` (connection cap)
- Modify: `internal/background/worker_pool.go` (queue depth metric)

- [ ] **Step 1: Add exponential backoff to debate optimizer**

When semaphore acquisition fails (channel full), use exponential backoff with jitter instead of immediate retry:

```go
func (o *PerformanceOptimizer) acquireSemaphoreWithBackoff(ctx context.Context) error {
    backoff := 10 * time.Millisecond
    for {
        select {
        case o.semaphore <- struct{}{}:
            return nil
        case <-ctx.Done():
            return ctx.Err()
        default:
            jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
            time.Sleep(backoff + jitter)
            backoff = min(backoff*2, 5*time.Second)
        }
    }
}
```

- [ ] **Step 2: Add server-wide HTTP request limiter**

Create a Gin middleware that limits max in-flight requests using a buffered channel semaphore:

```go
// internal/middleware/concurrency_limiter.go
func ConcurrencyLimiter(maxInFlight int) gin.HandlerFunc {
    sem := make(chan struct{}, maxInFlight)
    return func(c *gin.Context) {
        select {
        case sem <- struct{}{}:
            defer func() { <-sem }()
            c.Next()
        default:
            c.AbortWithStatusJSON(503, gin.H{"error": "server at capacity"})
        }
    }
}
```

Register in router before protected routes. Default `maxInFlight` configurable via env var `MAX_IN_FLIGHT_REQUESTS` (default: 1000).

- [ ] **Step 3: Add max concurrent SSE/WebSocket connections per client**

In `sse_manager.go`, track connections per client IP with a configurable cap (default: 10):

```go
const MaxConnectionsPerClient = 10
```

- [ ] **Step 3: Add queue depth metric to worker pool**

In `worker_pool.go`, expose queue depth as a Prometheus gauge:

```go
queueDepth := prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "helixagent_background_queue_depth",
    Help: "Current depth of background task queue",
})
```

Log a warning when queue exceeds 80% capacity.

- [ ] **Step 4: Write tests for each**

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/services/ ./internal/notifications/ ./internal/background/ -short -count=1
git add internal/services/ internal/notifications/ internal/background/
git commit -m "perf(backpressure): add exponential backoff, connection caps, queue depth metrics"
```

---

### Task 10: Create Performance Challenge Scripts

**Files:**
- Create: `challenges/scripts/monitoring_metrics_accuracy_challenge.sh`
- Create: `challenges/scripts/lazy_loading_comprehensive_challenge.sh`
- Create: `challenges/scripts/benchmark_regression_challenge.sh`

- [ ] **Step 1: Write monitoring_metrics_accuracy_challenge.sh**

Validates: all expected Prometheus metrics registered, sync.Once metrics patterns, OpenTelemetry spans.

- [ ] **Step 2: Write lazy_loading_comprehensive_challenge.sh**

Validates: sync.Once in MCP adapters, formatters, VectorDB, embedding providers, BigData components. Checks startup time does not include heavy initialization.

- [ ] **Step 3: Write benchmark_regression_challenge.sh**

Validates: benchmarks run successfully, results within acceptable thresholds vs documented baselines.

- [ ] **Step 4: Make executable and commit**

```bash
chmod +x challenges/scripts/monitoring_metrics_accuracy_challenge.sh
chmod +x challenges/scripts/lazy_loading_comprehensive_challenge.sh
chmod +x challenges/scripts/benchmark_regression_challenge.sh
git add challenges/scripts/
git commit -m "test(challenges): add monitoring, lazy loading, and benchmark regression challenges"
```

---

### Task 11: Final SP4 Validation

- [ ] **Step 1: Run monitoring tests**

Run: `go test ./tests/monitoring/ -tags integration -v -count=1`

- [ ] **Step 2: Run benchmarks**

Run: `go test ./tests/performance/ -tags performance -bench=. -benchmem -count=1`

- [ ] **Step 3: Verify lazy loading**

Run: `./challenges/scripts/lazy_loading_comprehensive_challenge.sh`

- [ ] **Step 4: Verify pprof**

Run: `./challenges/scripts/pprof_memory_profiling_challenge.sh`

- [ ] **Step 5: Tag completion**

```bash
git tag sp4-complete
```
