# SP3: Safety & Security Hardening — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Zero race conditions, zero memory leaks, zero goroutine leaks, all 7 security scanners green with zero unresolved HIGH/CRITICAL findings.

**Architecture:** Audit and fix channel leaks in ACP providers, add context-based cancellation to cleanup goroutines, add LRU eviction to unbounded sync.Map, run all 7 containerized security scanners, triage and fix findings, expand stress tests for 6 critical paths.

**Tech Stack:** Go 1.25.3, go.uber.org/goleak, Docker/Podman (security scanner containers), testify v1.11.1

**Spec:** `docs/superpowers/specs/2026-03-25-comprehensive-completion-design.md` (SP3 section)

**Depends on:** SP1 complete. Can run concurrently with SP2 and SP4.

---

### Task 1: Fix Channel Leak in Gemini ACP Provider

**Files:**
- Modify: `internal/llm/providers/gemini/gemini_acp.go` (around line 655)

- [ ] **Step 1: Read the function containing the channel creation**

Read `gemini_acp.go` lines 640-680. Identify the `done := make(chan bool, 1)` and trace all code paths to verify the channel is always consumed.

- [ ] **Step 2: Identify uncovered code paths**

Look for `return` statements or error branches where the goroutine writing to the channel is still running but the reading goroutine has exited.

- [ ] **Step 3: Add context cancellation to the goroutine**

Wrap the goroutine with a `select` on both `ctx.Done()` and the channel:

```go
select {
case <-ctx.Done():
    return
case result := <-done:
    // process result
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/llm/providers/gemini/ -v -short -count=1`
Expected: All pass

- [ ] **Step 5: Commit**

```bash
git add internal/llm/providers/gemini/gemini_acp.go
git commit -m "fix(gemini): prevent channel leak in ACP provider by adding context cancellation"
```

---

### Task 2: Fix Channel Leak in Qwen ACP Provider

**Files:**
- Modify: `internal/llm/providers/qwen/qwen_acp.go` (around line 639)

- [ ] **Step 1: Read the function containing doneCh**

Read `qwen_acp.go` lines 625-660. Identify `doneCh := make(chan completeResult, 1)`.

- [ ] **Step 2: Apply the same context-select pattern as Task 1**

- [ ] **Step 3: Run tests and commit**

```bash
go test ./internal/llm/providers/qwen/ -v -short -count=1
git add internal/llm/providers/qwen/qwen_acp.go
git commit -m "fix(qwen): prevent channel leak in ACP provider by adding context cancellation"
```

---

### Task 3: Add Context Select to query_optimizer Cleanup Loop

**Files:**
- Modify: `internal/database/query_optimizer.go` (around line 87)

- [ ] **Step 1: Read the cleanupLoop function**

Read `query_optimizer.go` to find the `go qc.cleanupLoop()` call and the loop implementation. Determine if it uses `time.Ticker` or `time.Sleep` and whether it selects on any context.

- [ ] **Step 2: Add context.Context field if missing**

```go
type QueryOptimizer struct {
    // ... existing fields ...
    ctx    context.Context
    cancel context.CancelFunc
}
```

- [ ] **Step 3: Update cleanupLoop to select on ctx.Done()**

```go
func (qc *QueryOptimizer) cleanupLoop() {
    ticker := time.NewTicker(qc.cleanupInterval)
    defer ticker.Stop()
    for {
        select {
        case <-qc.ctx.Done():
            return
        case <-ticker.C:
            qc.cleanup()
        }
    }
}
```

- [ ] **Step 4: Add Shutdown method**

```go
func (qc *QueryOptimizer) Shutdown() {
    qc.cancel()
}
```

- [ ] **Step 5: Verify tiered_cache.go cleanup loop**

Read `internal/cache/tiered_cache.go` `l1CleanupLoop()` method. Verify it selects on `tc.ctx.Done()`. If not, add the select pattern. The struct already has `ctx` and `cancel` fields (lines 160-161).

- [ ] **Step 6: Run tests and commit**

```bash
go test ./internal/database/ -v -short -count=1
go test ./internal/cache/ -v -short -count=1
git add internal/database/query_optimizer.go internal/cache/tiered_cache.go
git commit -m "fix(safety): add context cancellation to cleanup goroutines in query_optimizer and tiered_cache"
```

---

### Task 4: Add LRU Eviction to Debate Performance Optimizer sync.Map

**Files:**
- Modify: `internal/services/debate_performance_optimizer.go` (around line 52)

- [ ] **Step 1: Read current cache implementation**

The optimizer uses `sync.Map` for response caching with TTL. It has no max size limit.

- [ ] **Step 2: Add max entries configuration**

```go
type PerformanceOptimizerConfig struct {
    // ... existing fields ...
    MaxCacheEntries int // Maximum cache entries before eviction (default: 10000)
}
```

- [ ] **Step 3: Add size tracking and eviction**

Track cache size with an atomic counter. On insert, if size exceeds `MaxCacheEntries`, evict oldest entries (by TTL timestamp).

```go
func (o *PerformanceOptimizer) cacheResponse(key string, resp interface{}) {
    currentSize := atomic.LoadInt64(&o.cacheSize)
    if currentSize >= int64(o.config.MaxCacheEntries) {
        o.evictOldest()
    }
    o.cache.Store(key, &cachedResponse{
        response:  resp,
        timestamp: time.Now(),
    })
    atomic.AddInt64(&o.cacheSize, 1)
}
```

- [ ] **Step 4: Write test for eviction behavior**

```go
func TestPerformanceOptimizer_CacheEviction(t *testing.T) {
    config := PerformanceOptimizerConfig{MaxCacheEntries: 10, CacheTTL: time.Minute}
    opt := NewPerformanceOptimizer(config)

    // Fill cache beyond limit
    for i := 0; i < 20; i++ {
        opt.cacheResponse(fmt.Sprintf("key-%d", i), "value")
    }

    // Verify size is bounded
    assert.LessOrEqual(t, opt.CacheSize(), int64(10))
}
```

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/services/ -run TestPerformanceOptimizer -v -count=1
git add internal/services/debate_performance_optimizer.go internal/services/debate_performance_optimizer_test.go
git commit -m "fix(debate): add LRU eviction to performance optimizer cache (max 10000 entries)"
```

---

### Task 5: Add Circuit Breaker Near-Cap Warning Metric

**Files:**
- Modify: `internal/llm/circuit_breaker.go`

- [ ] **Step 1: Add Prometheus metric for listener count**

In the `AddListener` method, after adding a listener, check if count > 80% of `MaxCircuitBreakerListeners` and emit a warning log + increment a metric.

- [ ] **Step 2: Run tests and commit**

```bash
go test ./internal/llm/ -run TestCircuitBreaker -v -count=1
git add internal/llm/circuit_breaker.go
git commit -m "feat(llm): add near-cap warning metric for circuit breaker listeners"
```

---

### Task 6: Add goleak to Critical Package TestMain

**Files:**
- Modify or create `TestMain` in:
  - `internal/llm/main_test.go`
  - `internal/services/main_test.go`
  - `internal/handlers/main_test.go`
  - `internal/cache/main_test.go`
  - `internal/background/main_test.go`

- [ ] **Step 1: Add goleak dependency if not present**

Check `go.mod` for `go.uber.org/goleak`. If missing:
```bash
go get go.uber.org/goleak
```

- [ ] **Step 2: Add TestMain with goleak to each package**

```go
package llm

import (
    "testing"
    "go.uber.org/goleak"
)

func TestMain(m *testing.M) {
    // NOTE: goleak.VerifyTestMain calls m.Run() internally — do NOT also call os.Exit(m.Run())
    goleak.VerifyTestMain(m,
        goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
        goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
    )
}
```

If `TestMain` already exists, add `goleak.VerifyTestMain` wrapping.

- [ ] **Step 3: Run tests to verify no leaks**

Run: `go test ./internal/llm/ -v -count=1`
If goleak reports leaks, fix them before proceeding.

- [ ] **Step 4: Commit**

```bash
git add internal/*/main_test.go go.mod go.sum
git commit -m "test(safety): add goleak goroutine leak detection to critical packages"
```

---

### Task 7: Run Full Race Condition Detection

- [ ] **Step 1: Run race detector on all internal packages**

Run: `GOMAXPROCS=2 nice -n 19 go test -race ./internal/... -short -count=1 -p 1 2>&1 | tee /tmp/race-results.txt`

- [ ] **Step 2: Analyze findings**

```bash
grep -c "DATA RACE" /tmp/race-results.txt
```

- [ ] **Step 3: Fix each detected race**

For each race, add proper synchronization (mutex, atomic, or channel-based coordination).

- [ ] **Step 4: Re-run to confirm fixes**

Run: `go test -race ./internal/... -short -count=1 -p 1`
Expected: Zero DATA RACE detections

- [ ] **Step 5: Commit fixes**

Stage only the specific files that were modified to fix races (do NOT use `git add -A`):

```bash
git add internal/path/to/fixed_file1.go internal/path/to/fixed_file2.go
git commit -m "fix(safety): resolve all race conditions detected by -race flag"
```

---

### Task 8: Run Security Scanners — Gosec

**No code files modified. Scanner execution and report.**

- [ ] **Step 1: Run Gosec**

Run: `make security-scan-gosec`
Expected: Report generated at `reports/security/gosec-report.json`

- [ ] **Step 2: Analyze findings**

```bash
cat reports/security/gosec-report.json | python3 -m json.tool | grep -c '"severity"'
```

- [ ] **Step 3: Fix HIGH/CRITICAL findings**

Address each finding. Common Gosec issues: G104 (unhandled errors), G304 (file path injection), G401 (weak crypto).

- [ ] **Step 4: Re-scan to confirm**

Run: `make security-scan-gosec`

- [ ] **Step 5: Commit fixes**

```bash
git add -A
git commit -m "fix(security): resolve Gosec security findings"
```

---

### Task 9: Run Security Scanners — Trivy

- [ ] **Step 1: Run Trivy**

Run: `make security-scan-trivy`

- [ ] **Step 2: Fix HIGH/CRITICAL vulnerabilities**

Typically: upgrade vulnerable dependencies in `go.mod`.

- [ ] **Step 3: Re-scan and commit**

```bash
go mod tidy
git add go.mod go.sum reports/security/
git commit -m "fix(security): resolve Trivy vulnerability findings"
```

---

### Task 10: Run Security Scanners — Snyk, SonarQube, Semgrep, KICS, Grype

- [ ] **Step 1: Run each scanner**

```bash
make security-scan-snyk
make security-scan-sonarqube
make security-scan-semgrep
make security-scan-kics
make security-scan-grype
```

- [ ] **Step 2: Triage findings per scanner**

For each: categorize as Critical/High/Medium/Low. Fix Critical and High.

- [ ] **Step 3: Fix findings and re-scan**

- [ ] **Step 4: Update SECURITY_SCAN_SUMMARY.md**

Write consolidated report at `reports/security/SECURITY_SCAN_SUMMARY.md` with status per scanner.

- [ ] **Step 5: Commit**

```bash
git add -A reports/security/
git add -A internal/ go.mod go.sum
git commit -m "fix(security): resolve findings from Snyk, SonarQube, Semgrep, KICS, Grype"
```

---

### Task 11: Expand Stress Tests

**Files:**
- Create: `tests/stress/rate_limiter_stress_test.go`
- Create: `tests/stress/ensemble_failure_stress_test.go`
- Create: `tests/stress/debate_concurrency_stress_test.go`
- Create: `tests/stress/streaming_storm_stress_test.go`
- Create: `tests/stress/cache_stampede_stress_test.go`
- Create: `tests/stress/db_pool_exhaustion_stress_test.go`

- [ ] **Step 1: Write rate limiter stress test**

```go
//go:build stress

package stress

func TestRateLimiter_10KConcurrentRequests(t *testing.T) {
    // Launch 10K goroutines hitting the rate limiter
    // Verify no requests bypass the limit
    // Verify no goroutine leaks after completion
}
```

- [ ] **Step 2-6: Write remaining 5 stress tests**

Each tests a different failure mode at 30-40% resource limits (GOMAXPROCS=2).

- [ ] **Step 7: Run stress tests**

Run: `GOMAXPROCS=2 nice -n 19 ionice -c 3 go test ./tests/stress/ -tags stress -v -count=1 -p 1 -timeout 5m`

- [ ] **Step 8: Commit**

```bash
git add tests/stress/
git commit -m "test(stress): add 6 stress tests for rate limiter, ensemble, debate, streaming, cache, db pool"
```

---

### Task 12: Create Safety Challenge Scripts

**Files:**
- Create: `challenges/scripts/safety_comprehensive_challenge.sh`
- Create: `challenges/scripts/security_scan_resolution_challenge.sh`

- [ ] **Step 1: Write safety_comprehensive_challenge.sh**

Validates: no race conditions (-race clean), goleak passes, channel leaks fixed, cleanup loops have context.

- [ ] **Step 2: Write security_scan_resolution_challenge.sh**

Validates: all 7 scanner configs exist, reports directory has recent reports, no HIGH/CRITICAL unresolved.

- [ ] **Step 3: Make executable and commit**

```bash
chmod +x challenges/scripts/safety_comprehensive_challenge.sh
chmod +x challenges/scripts/security_scan_resolution_challenge.sh
git add challenges/scripts/
git commit -m "test(challenges): add safety and security scan resolution challenges"
```

---

### Task 13: Final SP3 Validation

- [ ] **Step 1: Race detection clean**

Run: `go test -race ./internal/... -short -count=1 -p 1`

- [ ] **Step 2: goleak clean**

Run: `go test ./internal/llm/ ./internal/services/ ./internal/handlers/ ./internal/cache/ ./internal/background/ -v -count=1`

- [ ] **Step 3: All security reports generated**

Verify: `ls reports/security/*.json reports/security/SECURITY_SCAN_SUMMARY.md`

- [ ] **Step 4: Run safety challenge**

Run: `./challenges/scripts/safety_comprehensive_challenge.sh`

- [ ] **Step 5: Tag completion**

```bash
git tag sp3-complete
```
