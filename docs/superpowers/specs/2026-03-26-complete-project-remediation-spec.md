# HelixAgent Complete Project Remediation & Perfection Specification

**Version:** 1.1.0
**Date:** 2026-03-26 (updated during execution)
**Status:** In Progress — Phases 1, 2, 4, 9 COMPLETE; Phase 3 SKIPPED (not needed)
**Scope:** Full project audit remediation — dead code, memory safety, test coverage, security scanning, documentation, website, video courses, monitoring, lazy loading, stress testing

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Audit Findings](#2-audit-findings)
3. [Phase 1 — Dead Code Elimination](#3-phase-1--dead-code-elimination)
4. [Phase 2 — Memory Safety & Concurrency Fixes](#4-phase-2--memory-safety--concurrency-fixes)
5. [Phase 3 — Test Re-enablement](#5-phase-3--test-re-enablement)
6. [Phase 4 — Coverage Gap Closure](#6-phase-4--coverage-gap-closure)
7. [Phase 5 — New Stress & Integration Tests](#7-phase-5--new-stress--integration-tests)
8. [Phase 6 — Security Scanning Execution & Resolution](#8-phase-6--security-scanning-execution--resolution)
9. [Phase 7 — Monitoring & Metrics Tests](#9-phase-7--monitoring--metrics-tests)
10. [Phase 8 — Lazy Loading & Performance Optimization](#10-phase-8--lazy-loading--performance-optimization)
11. [Phase 9 — Documentation & Content Completion](#11-phase-9--documentation--content-completion)
12. [Phase 10 — Website & Final Validation](#12-phase-10--website--final-validation)
13. [Verification Strategy](#13-verification-strategy)
14. [Risk Mitigation](#14-risk-mitigation)
15. [Success Criteria](#15-success-criteria)

---

## 1. Executive Summary

This specification defines a 10-phase remediation plan to bring the HelixAgent project to 100% completion across all dimensions: code quality, test coverage, documentation, security, performance, and content. The plan is ordered to maximize safety — each phase is independently verifiable and must not break existing functionality.

### Key Metrics

| Dimension | Before | After | Status |
|-----------|--------|-------|--------|
| Dead code | 9,156 lines | 0 lines | DONE (Phase 1) |
| Race conditions | 2 critical + others | 0 critical, regression tests added | DONE (Phase 2) |
| Memory leak risks | 1 unbounded cache | 0 unbounded (eviction added) | DONE (Phase 2) |
| Skipped tests | All conditional (testing.Short) | Already correct — standard Go practice | SKIPPED (Phase 3 not needed) |
| Test coverage (utils/) | 86.9% (not 25%) | 86.9% + 53 new tests | DONE (Phase 4) |
| Fuzz tests | 7 files | 15 files (+8 new) | DONE (Phase 4) |
| Race tests | 1 file | 11 files (+10 new, 30 tests) | DONE (Phase 4) |
| Duplicate docs | 2 pairs of manuals | 0 duplicates (merged) | DONE (Phase 9) |
| Security docs | Scattered | Centralized docs/security/ (5 files) | DONE (Phase 9) |
| Video course org | Empty dir, no README index | Cleaned, full 76-course catalog | DONE (Phase 9) |
| Stress tests | Pending | Pending | Phase 5 |
| Security scan findings | Not yet run | Pending | Phase 6 |
| Monitoring tests | 8 files | Pending | Phase 7 |
| Lazy loading | 30+ sync.Once | Pending | Phase 8 |
| Website updates | Pending | Pending | Phase 10 |

---

## 2. Audit Findings

### 2.1 Dead Code Inventory

#### Completely Unused Service Files (2,672 lines)

| File | Lines | Why Dead |
|------|-------|----------|
| `internal/services/high_availability.go` | 665 | ProtocolFederation, EventBus, MCPFederatedProtocol, LSPFederatedProtocol — constructors never called |
| `internal/services/protocol_analytics.go` | 486 | ProtocolAnalyticsService — never instantiated |
| `internal/services/security_sandbox.go` | 311 | SecuritySandbox — never wired into handler pipeline |
| `internal/services/protocol_plugin_system.go` | 840 | ProtocolPluginSystem + ProtocolPluginRegistry — never called |
| `internal/services/debate_resilience_service.go` | 370 | DebateResilienceService — debate system uses different resilience approach |

#### Backup/Vestigial Files (3,600 lines)

| File | ~Lines |
|------|--------|
| `internal/background/task_queue.go.backup` | 400 |
| `internal/background/stuck_detector.go.backup` | 350 |
| `internal/background/resource_monitor.go.backup` | 280 |
| `internal/background/messaging_adapter.go.backup` | 250 |
| `internal/config/ai_debate_integration_test.go.backup` | 300 |
| `internal/handlers/openai_compatible_test.go.backup` | 400 |
| `internal/mcp/adapters/brave_search.go.bak` | 200 |
| `internal/mcp/config/generator_full.go.bak` | 180 |
| `internal/services/concurrency_alert_manager_test.go.bak` | 150 |
| `internal/services/provider_registry.go.bak` | 800 |
| `internal/streaming/flink/client.go.bak` | 300 |

#### Unused Helper Methods (//nolint:unused markers)

| File | Method/Field | Type |
|------|-------------|------|
| `handlers/openai_compatible.go` | `contains()`, `containsSubstring()`, `generateID()`, `generateDebateDialogueResponse()`, `buildDebateRoleSystemPrompt()`, `extractDocumentationContent()` | Functions |
| `handlers/cognee_handler.go` | `getIntParam()`, `getFloatParam()` | Functions |
| `services/plugin_system.go` | `mu`, `current`, `weights` | Fields |
| `services/lsp_manager.go` | `openDocument()` | Method |
| `services/embedding_manager.go` | `bytesToFloat64()` | Function |
| `services/debate_performance_service.go` | `lastMemStats` | Field |
| `services/request_service.go` | `mu` | Field |
| `middleware/rate_limit.go` | `max()` | Function |
| `modelsdev/client.go` | `doPost()` | Method |
| `database/db.go` | `migrations` | Slice |
| `background/stuck_detector.go` | `checkHeartbeatTimeout()`, `isProcessFrozen()`, `checkResourceExhaustion()`, `isIOStarved()`, `isNetworkHung()`, `hasMemoryLeak()`, `isEndlessTaskStuck()` | Methods |
| `background/events.go` | `wg` | Field |
| `background/task_queue.go` | `mu` | Field |
| `handlers/verifier_types.go` | `VerifierErrorResponse` | Type |

### 2.2 Memory Safety Issues

| # | Severity | Location | Issue |
|---|----------|----------|-------|
| 1 | CRITICAL | `internal/llm/circuit_breaker.go:298-319` | Circuit breaker listener notification spawns inner goroutine without lifecycle tracking; under timeout, inner goroutine is orphaned |
| 2 | CRITICAL | `internal/llm/circuit_breaker.go:164-178` | CompleteStream wraps response channel in goroutine without WaitGroup; leaks if caller never drains channel |
| 3 | HIGH | `internal/services/debate_service.go:48-49` | `intentCache` map grows indefinitely without eviction policy |
| 4 | HIGH | `internal/notifications/sse_manager.go:108-121` | Channel closing race — concurrent Send() may write while Stop() is closing channels |
| 5 | HIGH | `internal/http/pool.go:104-109` | HTTPClientPool `clients` map grows unbounded per unique host |
| 6 | HIGH | `internal/mcp/connection_pool.go:73,83-90` | MCPConnectionPool `connections` map unbounded; MaxConnections config not enforced |
| 7 | MEDIUM-HIGH | `internal/services/provider_registry.go:85` | `initOnce` map accessed without synchronization for initial writes |
| 8 | MEDIUM | `internal/plugins/hot_reload.go:95-99` | watcher.Close() races with watchLoop; should be inside stopOnce.Do() |
| 9 | MEDIUM | `internal/services/agent_worker_pool.go:83-89` | Context monitoring goroutine without WaitGroup tracking |
| 10 | MEDIUM | `internal/notifications/websocket_server.go:26-28` | Lock ordering documented but not enforced |
| 11 | MEDIUM | `internal/cache/expiration.go:84-96` | Start() returns before goroutine actually starts; Stop() could race |
| 12 | LOW-MEDIUM | `internal/services/provider_health_monitor.go:60-120` | Health check goroutines need lifecycle verification |
| 13 | LOW-MEDIUM | `internal/services/debate_service.go:1368` | `context.Background()` used instead of request context |

### 2.3 Skipped Tests Summary

| Category | Skipped Files | Skipped Tests |
|----------|--------------|---------------|
| Stress tests | 40 | ~120 |
| Submodule test suites (integration/e2e/security/stress) | 120+ | ~480 |
| Protocol functional tests | 7 | ~37 |
| LLMsVerifier tests | 8 | ~43 |
| Monitoring tests | 4 | ~16 |
| Performance tests | 2 | ~11 |
| Chaos tests | 3 | ~15 |
| Other (challenge, CLI, handlers, etc.) | ~10 | ~50 |
| **Total** | **183** | **900+** |

### 2.4 Coverage Gaps

| Package | Files | Tests | Coverage | Gap |
|---------|-------|-------|----------|-----|
| `internal/utils/` | 8 | 2 | 25% | 6 files untested |
| `internal/benchmark/` | 3 | 1 | 33% | 2 files untested |
| `internal/llmops/` | 5 | 3 | 60% | 2 files untested |
| `internal/selfimprove/` | 5 | 3 | 60% | 2 files untested |
| `internal/observability/` | 5 | 3 | 60% | 2 files untested |
| `internal/planning/` | 3 | 2 | 66% | 1 file untested |
| `internal/knowledge/` | 3 | 2 | 66% | 1 file untested |

### 2.5 Documentation Issues

| Issue | Details |
|-------|---------|
| Duplicate user manuals | 32/42 (security scanning), 33/43 (performance optimization) |
| Video course numbering | 3 naming schemes: `course-N`, `video-course-N`, subdirectories |
| Duplicate video courses | 4+ topics with multiple versions |
| Empty directory | `Website/video-courses/courses-21-50/` is empty |
| Missing docs/security/ | Security docs scattered across multiple locations |
| Missing MODULES.md TOC | 730-line document without navigation |
| Missing FAQ.md | No standalone FAQ document |
| Broken reference | docs/README.md references `docs/api/grpc-api.md` but file is `docs/api/grpc.md` |
| MCP-Servers docs | Missing CLAUDE.md, AGENTS.md, docs/ (third-party repo — low priority) |

### 2.6 Disabled Features

| Feature | Location | Status |
|---------|----------|--------|
| Cognee memory system | `internal/config/config.go` | Disabled (replaced by Mem0) |
| 8-phase orchestrator (streaming) | `internal/handlers/openai_compatible.go:549-554` | Disabled for streaming (timeout issue) |
| GraphQL endpoint | `internal/features/config.go` | Feature-flagged disabled by default |

---

## 3. Phase 1 — Dead Code Elimination

**Goal:** Remove all dead code, achieving 0 lines of unreachable/unused production code.

### 3.1 Delete Unused Service Files

**Action:** Delete each file, run `go build ./cmd/helixagent` after each to verify no breakage.

1. Delete `internal/services/high_availability.go` (665 lines)
2. Delete `internal/services/protocol_analytics.go` (486 lines)
3. Delete `internal/services/security_sandbox.go` (311 lines)
4. Delete `internal/services/protocol_plugin_system.go` (840 lines)
5. Delete `internal/services/debate_resilience_service.go` (370 lines)

**Pre-condition check:** Grep for each constructor name (NewProtocolFederation, NewProtocolAnalyticsService, NewSecuritySandbox, NewProtocolPluginSystem, NewDebateResilienceService) to confirm zero callers.

### 3.2 Delete Backup Files

**Action:** Delete all `.backup` and `.bak` files.

Files to delete:
- `internal/background/task_queue.go.backup`
- `internal/background/stuck_detector.go.backup`
- `internal/background/resource_monitor.go.backup`
- `internal/background/messaging_adapter.go.backup`
- `internal/config/ai_debate_integration_test.go.backup`
- `internal/handlers/openai_compatible_test.go.backup`
- `internal/mcp/adapters/brave_search.go.bak`
- `internal/mcp/config/generator_full.go.bak`
- `internal/services/concurrency_alert_manager_test.go.bak`
- `internal/services/provider_registry.go.bak`
- `internal/streaming/flink/client.go.bak`

### 3.3 Remove Unused Helper Methods

**Action:** For each `//nolint:unused` item listed in Section 2.1, remove the function/field/method. Run `go build` and `go vet` after each batch removal.

**Special cases:**
- `background/stuck_detector.go` has 7 unused methods — verify none are called via interface dispatch before removing
- `handlers/openai_compatible.go` has 6 unused functions — remove all
- `handlers/verifier_types.go` — delete entire file if `VerifierErrorResponse` is unused

### 3.4 Verification

```bash
go build -mod=vendor ./cmd/helixagent
go build -mod=vendor ./cmd/api
go build -mod=vendor ./cmd/grpc-server
go vet -mod=vendor ./internal/...
make fmt
```

### 3.5 Challenge

Create `challenges/scripts/dead_code_verification_challenge.sh` that:
- Greps for `//nolint:unused` and fails if count exceeds threshold
- Checks no `.backup` or `.bak` files exist in `internal/`
- Verifies all 5 dead service files are gone
- Builds all 7 apps successfully

---

## 4. Phase 2 — Memory Safety & Concurrency Fixes

**Goal:** Zero race conditions, zero goroutine leaks, zero unbounded memory growth.

### 4.1 Fix Circuit Breaker Goroutine Leaks (CRITICAL)

**File:** `internal/llm/circuit_breaker.go`

**Fix 1 — Listener notification (lines 298-319):**
- Add `context.Context` to `notifyListeners()` method
- Replace inner goroutine pattern with `context.WithTimeout` + direct call
- If listener blocks past timeout, log warning and continue (no orphaned goroutine)

```go
// Before: orphaned goroutine on timeout
go func(l CircuitBreakerListener) {
    done := make(chan struct{})
    go func() { defer close(done); l(providerID, old, new) }()
    select { case <-done: case <-time.After(timeout): }
}(listener)

// After: context-controlled, no leak
go func(l CircuitBreakerListener) {
    defer wg.Done()
    ctx, cancel := context.WithTimeout(parentCtx, timeout)
    defer cancel()
    done := make(chan struct{})
    go func() { defer close(done); l(providerID, old, new) }()
    select {
    case <-done:
    case <-ctx.Done():
        // Log timeout, goroutine will complete naturally but is tracked
    }
}(listener)
```

**Fix 2 — CompleteStream wrapping (lines 164-178):**
- Add WaitGroup to CircuitBreaker struct
- Track the stream-wrapping goroutine
- Add cleanup in Shutdown/Close method

### 4.2 Fix Unbounded Caches (HIGH)

**File:** `internal/services/debate_service.go`
- Add LRU eviction to `intentCache` with max 10,000 entries
- Add TTL-based cleanup (1-hour TTL for intent classifications)
- Implementation: Use bounded map with eviction goroutine (WaitGroup-tracked)

**File:** `internal/http/pool.go`
- Enforce max pool size in HTTPClientPool (default: 100 unique hosts)
- Add LRU eviction when pool is full
- Close evicted clients gracefully

**File:** `internal/mcp/connection_pool.go`
- Enforce `MaxConnections` config (currently defined but not checked)
- Add eviction of least-recently-used connections
- Close evicted connections gracefully

### 4.3 Fix SSE Manager Race (HIGH)

**File:** `internal/notifications/sse_manager.go`
- Add atomic `stopping` flag checked before sends
- Change Stop() to set flag before closing channels
- Add recovery for send-on-closed-channel panics as safety net

### 4.4 Fix Provider Registry Race (MEDIUM-HIGH)

**File:** `internal/services/provider_registry.go`
- Replace `map[string]*sync.Once` with `sync.Map` for `initOnce`
- Or add RWMutex protection around map access in GetProvider()

### 4.5 Fix Plugin Hot Reload Race (MEDIUM)

**File:** `internal/plugins/hot_reload.go`
- Move `h.watcher.Close()` inside the `stopOnce.Do()` closure
- This ensures watcher is only closed once and before WaitGroup.Wait()

### 4.6 Fix Agent Worker Pool Context Leak (MEDIUM)

**File:** `internal/services/agent_worker_pool.go`
- Add context monitoring goroutine to pool's WaitGroup
- Ensure goroutine exits when both contexts are done

### 4.7 Fix Context.Background() Misuse (LOW-MEDIUM)

**File:** `internal/services/debate_service.go:1368`
- Change `context.WithTimeout(context.Background(), 15*time.Second)` to use request context as parent
- Pattern: `context.WithTimeout(ctx, 15*time.Second)` where `ctx` is passed from handler

### 4.8 Verification

```bash
go build -race -mod=vendor ./cmd/helixagent
go test -race -mod=vendor -count=1 -short ./internal/llm/...
go test -race -mod=vendor -count=1 -short ./internal/services/...
go test -race -mod=vendor -count=1 -short ./internal/notifications/...
go test -race -mod=vendor -count=1 -short ./internal/http/...
go test -race -mod=vendor -count=1 -short ./internal/mcp/...
go test -race -mod=vendor -count=1 -short ./internal/plugins/...
go test -race -mod=vendor -count=1 -short ./internal/cache/...
```

### 4.9 New Tests

For each fix, add corresponding test:
- `internal/llm/circuit_breaker_lifecycle_test.go` — goroutine leak detection
- `internal/services/debate_service_cache_eviction_test.go` — cache bounds
- `internal/http/pool_eviction_test.go` — pool bounds
- `internal/mcp/connection_pool_bounds_test.go` — connection limits
- `internal/notifications/sse_manager_race_test.go` — concurrent send+stop
- `internal/services/provider_registry_concurrent_test.go` — concurrent GetProvider

### 4.10 Challenge

Create `challenges/scripts/memory_safety_comprehensive_challenge.sh` that:
- Runs race detector on all packages
- Verifies no unbounded maps in critical services
- Checks WaitGroup usage in goroutine-spawning functions
- Validates context propagation patterns

---

## 5. Phase 3 — Test Re-enablement

**Goal:** Zero blanket-skipped test files. Tests that need infrastructure get conditional skips (not `t.Skip("skip")` at the top).

### 5.1 Strategy

Tests fall into 3 categories:

1. **Infrastructure-dependent** — Need PostgreSQL, Redis, Mock LLM, etc.
   - Change from `t.Skip("reason")` to `skipIfNoInfra(t)` helper that checks env vars
   - These tests run when `make test-with-infra` is used

2. **Provider-dependent** — Need real API keys
   - Change to `skipIfNoAPIKey(t, "PROVIDER_API_KEY")` helper
   - These tests run in full verification mode only

3. **Incorrectly skipped** — Can run without any infrastructure
   - Remove skip entirely and fix any issues

### 5.2 Test Helper Functions

Add to `tests/testutils/test_helpers.go`:

```go
func SkipIfNoInfra(t *testing.T) {
    t.Helper()
    if os.Getenv("DB_HOST") == "" && os.Getenv("TEST_INFRA_RUNNING") == "" {
        t.Skip("Infrastructure not running (use make test-with-infra)")
    }
}

func SkipIfNoRedis(t *testing.T) {
    t.Helper()
    if os.Getenv("REDIS_HOST") == "" {
        t.Skip("Redis not available")
    }
}

func SkipIfNoAPIKey(t *testing.T, envVar string) {
    t.Helper()
    if os.Getenv(envVar) == "" {
        t.Skipf("API key %s not set", envVar)
    }
}
```

### 5.3 Files to Un-skip by Category

#### Stress Tests (40 files) — Category 1 (Infrastructure-dependent)
All files in `tests/stress/` — replace blanket skip with conditional infra check.

#### Submodule Tests (120+ files) — Category 1 or 3
Each submodule's integration/e2e/security/stress tests:
- EventBus, Concurrency, Observability, Auth, Storage, Streaming (24 files)
- Security, Embeddings, VectorDB, Database, Cache (20 files)
- MCP_Module, Formatters, Plugins (12 files)
- RAG, Memory, Optimization (12 files)
- Agentic, LLMOps, SelfImprove, Planning, Benchmark (20 files)
- DebateOrchestrator (1 file)
- HelixSpecifier (3 files)

For each: assess whether tests actually need infrastructure or were blanket-skipped during extraction. Remove skip if tests can run standalone.

#### Protocol Tests (7 files) — Category 1
- `internal/testing/mcp/functional_test.go`
- `internal/testing/lsp/functional_test.go`
- `internal/testing/acp/functional_test.go`
- `internal/testing/embeddings/functional_test.go`
- `internal/testing/vision/functional_test.go`
- `internal/testing/integration/mcp_debate_integration_test.go`
- `internal/debate/comprehensive/e2e_test.go`

#### Monitoring Tests (4 files) — Category 1
- `tests/monitoring/cache_hit_ratio_test.go`
- `tests/monitoring/circuit_breaker_transitions_test.go`
- `tests/monitoring/database_query_performance_test.go`
- `tests/monitoring/provider_latency_tracking_test.go`

#### LLMsVerifier Tests (8 files) — Category 2
Replace with conditional API key checks.

#### Other (remaining files) — Mixed categories
Each file assessed individually.

### 5.4 Verification

```bash
# Quick smoke test (no infra)
go test -mod=vendor -short -count=1 ./internal/... 2>&1 | grep -c "SKIP"
# Should show significantly fewer skips than before

# Full test (with infra)
make test-with-infra
```

### 5.5 Challenge

Update `challenges/scripts/test_skip_audit_challenge.sh`:
- Count total `t.Skip(` calls across all test files
- Verify no files have ALL tests skipped (blanket skip)
- Verify every skip uses one of the approved helper functions
- Report skip ratio per package

---

## 6. Phase 4 — Coverage Gap Closure

**Goal:** Every internal/ package has test coverage matching its source file count.

### 6.1 utils/ Package (25% -> 100%)

Create new test files:

| File to Test | New Test File | Tests to Add |
|-------------|--------------|--------------|
| `errors.go` | `errors_test.go` | Error wrapping, error types, error message formatting |
| `fibonacci.go` | `fibonacci_test.go` | Sequence generation, edge cases (0, 1, negative), performance |
| `logger.go` | `logger_test.go` | Log levels, output formatting, structured logging |
| `math.go` | `math_test.go` | All math utility functions, boundary values, overflow |
| `path_validation.go` | `path_validation_test.go` | Path traversal prevention, allowed paths, edge cases |
| `string.go` | `string_test.go` | String manipulation, encoding, sanitization |
| `testing.go` | (test helper — validate via usage) | Ensure helpers are used and work correctly |

### 6.2 benchmark/ Package (33% -> 100%)

- Add `benchmark_runner_test.go` — Benchmark orchestration
- Add `benchmark_results_test.go` — Result aggregation and leaderboard

### 6.3 Other Packages (60-66% -> 100%)

| Package | Files Needed |
|---------|-------------|
| `planning/` | 1 additional test file for uncovered planning algorithm |
| `knowledge/` | 1 additional test file for knowledge graph operations |
| `llmops/` | 2 additional test files for experiments and prompt versioning |
| `selfimprove/` | 2 additional test files for RLHF and reward modeling |
| `observability/` | 2 additional test files for tracing and metrics export |

### 6.4 Expand Fuzz Tests (7 -> 17)

New fuzz tests to create:

| Test File | Target |
|-----------|--------|
| `streaming_data_fuzz_test.go` | SSE/WebSocket message parsing |
| `debate_message_fuzz_test.go` | Debate protocol message structures |
| `mcp_protocol_fuzz_test.go` | MCP JSON-RPC message parsing |
| `memory_operation_fuzz_test.go` | Memory store/retrieve operations |
| `cache_key_fuzz_test.go` | Cache key generation and retrieval |
| `vector_query_fuzz_test.go` | Vector database query construction |
| `embedding_input_fuzz_test.go` | Embedding input sanitization |
| `rate_limit_fuzz_test.go` | Rate limiter key extraction |
| `health_check_fuzz_test.go` | Health check response parsing |
| `router_path_fuzz_test.go` | Router path matching |

### 6.5 Expand Race Condition Tests (1 -> 12)

New race test files:

| Test File | Focus |
|-----------|-------|
| `circuit_breaker_race_test.go` | Concurrent state transitions |
| `cache_race_test.go` | Concurrent read/write/eviction |
| `provider_registry_race_test.go` | Concurrent provider initialization |
| `ensemble_race_test.go` | Concurrent voting and aggregation |
| `debate_service_race_test.go` | Concurrent debate sessions |
| `sse_manager_race_test.go` | Concurrent subscribe/unsubscribe/send |
| `websocket_server_race_test.go` | Concurrent client management |
| `mcp_pool_race_test.go` | Concurrent connection acquisition |
| `http_pool_race_test.go` | Concurrent HTTP client creation |
| `background_task_race_test.go` | Concurrent task submission/completion |
| `plugin_hot_reload_race_test.go` | Concurrent plugin load/unload |

### 6.6 Expand Automation Tests (3 -> 15)

New automation test files covering:
- Agentic workflow execution end-to-end
- Tool execution pipeline automation
- Provider fallback chain automation
- Debate team selection automation
- Background task lifecycle automation
- MCP adapter registration automation
- Health check automation
- Circuit breaker recovery automation
- Cache warming automation
- Model discovery automation
- Config generation automation
- Challenge execution automation

### 6.7 Verification

```bash
# Per-package coverage check
go test -mod=vendor -coverprofile=coverage.out ./internal/utils/...
go tool cover -func=coverage.out | grep total
# Repeat for each package
```

### 6.8 Challenge

Create `challenges/scripts/coverage_completeness_challenge.sh`:
- For every internal/ package, verify test file count >= source file count
- Verify no package has less than 80% line coverage
- Verify fuzz test count >= 15
- Verify race test count >= 10
- Verify automation test count >= 12

---

## 7. Phase 5 — New Stress & Integration Tests

**Goal:** Validate that the system is "responsive like the flash and not possible to overload or break."

### 7.1 Handler Stress Tests

For every handler in `internal/handlers/`, create stress tests:

| Test | Scenario | Assertions |
|------|----------|------------|
| Concurrent chat completions | 100 concurrent requests | All complete within 30s, no panics |
| Concurrent streaming | 50 concurrent SSE streams | All streams deliver data, clean close |
| Concurrent health checks | 200 concurrent /health | All return 200, p99 < 100ms |
| Concurrent monitoring | 100 concurrent /v1/monitoring/* | All return valid JSON |
| Rate limit saturation | 10x rate limit requests | Proper 429 responses, no crash |
| Large payload | 1MB request body | Proper rejection or processing |
| Malformed requests | 1000 invalid JSON bodies | All return 400, no panic |

### 7.2 Provider Stress Tests

| Test | Scenario | Assertions |
|------|----------|------------|
| All providers concurrent | Hit all 43 providers simultaneously | Circuit breakers trigger appropriately |
| Provider failover storm | All primary providers fail | Fallback chain engages, no deadlock |
| Discovery cache stampede | Expire model cache, 100 concurrent requests | Single fetch, all served from cache |
| Rate limit backpressure | Provider returns 429 | Proper backoff, no retry storm |

### 7.3 Memory Pressure Tests

| Test | Scenario | Assertions |
|------|----------|------------|
| Heap growth under load | 10,000 requests | Heap stabilizes (no leak) |
| Goroutine count stability | 10,000 requests | Goroutine count returns to baseline |
| Cache memory bounds | Fill caches to limit | Eviction works, memory bounded |
| Connection pool bounds | Exhaust all connections | Proper queuing, no panic |

### 7.4 Database Integration Tests

| Test | Scenario | Assertions |
|------|----------|------------|
| Connection pool exhaustion | Use all connections | Proper timeout, no deadlock |
| Concurrent migrations | Parallel schema updates | Proper locking, no corruption |
| Transaction isolation | Concurrent read/write | No phantom reads, proper isolation |
| Recovery after disconnect | Kill DB, reconnect | Auto-reconnect, circuit breaker |

### 7.5 Network Partition Tests

| Test | Scenario | Assertions |
|------|----------|------------|
| Redis disconnect | Kill Redis mid-operation | Graceful degradation to in-memory |
| Provider timeout | Simulate 30s response time | Proper timeout, fallback triggered |
| DNS failure | Simulate DNS resolution failure | Proper error, no hang |
| TLS handshake failure | Simulate cert issue | Proper error message |

### 7.6 Verification

```bash
# Resource-limited execution per Constitution
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -mod=vendor -v -p 1 -timeout 300s ./tests/stress/...
```

### 7.7 Challenge

Create `challenges/scripts/stress_comprehensive_challenge.sh`:
- Run all stress tests with resource limits
- Verify p99 latency under load
- Verify zero panics
- Verify memory doesn't grow unbounded
- Verify goroutine count returns to baseline

---

## 8. Phase 6 — Security Scanning Execution & Resolution

**Goal:** Run all security scanners, analyze findings, resolve all actionable items.

### 8.1 Execution Plan

#### Step 1: Start Scanning Infrastructure

```bash
# Start SonarQube (containerized)
cd docker/security/sonarqube
docker compose up -d
# Wait for SonarQube to be healthy (60s startup)

# Verify Snyk containers
cd docker/security/snyk
docker compose --profile all up
```

#### Step 2: Run All Scanners

```bash
make security-scan-all
# This runs: gosec, snyk, sonarqube, trivy
```

#### Step 3: Collect Reports

Reports output to:
- `reports/gosec-report.json`
- `reports/snyk-report.json`
- `reports/sonarqube/` (web UI at localhost:9000)
- `reports/trivy-report.json`

### 8.2 Resolution Process

For each finding:
1. **Classify**: Critical / High / Medium / Low / False Positive
2. **Assess**: Is it actionable? Does fix risk breaking functionality?
3. **Fix or Suppress**: Fix if actionable; document suppression reason if false positive
4. **Test**: Verify fix doesn't break anything
5. **Document**: Record finding and resolution in `docs/security/scan-resolutions.md`

### 8.3 Expected Finding Categories

| Category | Expected Action |
|----------|----------------|
| G404 (weak random in retry jitter) | Already suppressed in .gosec.yml — acceptable |
| Dependency vulnerabilities | Update dependencies if safe |
| SQL injection risks | Verify parameterized queries |
| Hardcoded credentials | Verify these are test-only |
| File permission issues | Fix permissions |
| Certificate validation | Verify TLS configuration |

### 8.4 Challenge

Existing challenges validate scanning:
- `challenges/scripts/snyk_automated_scanning_challenge.sh` (38 tests)
- `challenges/scripts/sonarqube_automated_scanning_challenge.sh` (45 tests)

Create new `challenges/scripts/security_scan_resolution_verification_challenge.sh`:
- Verify all critical/high findings resolved
- Verify suppression documentation exists for each suppressed finding
- Verify no new critical findings introduced

---

## 9. Phase 7 — Monitoring & Metrics Tests

**Goal:** Comprehensive tests validating monitoring infrastructure works end-to-end.

### 9.1 Prometheus Metrics Tests

| Test File | Validates |
|-----------|----------|
| `prometheus_metric_registration_test.go` | All expected metrics registered |
| `prometheus_metric_emission_test.go` | Metrics emit correct values under load |
| `prometheus_label_consistency_test.go` | Label names consistent across metrics |
| `prometheus_scrape_endpoint_test.go` | /metrics endpoint returns valid exposition format |

### 9.2 Health Endpoint Tests

| Test File | Validates |
|-----------|----------|
| `health_endpoint_comprehensive_test.go` | All health endpoints return correct status |
| `health_degraded_state_test.go` | Health reports degraded when services fail |
| `health_recovery_test.go` | Health recovers after service restoration |

### 9.3 Circuit Breaker Metric Tests

| Test File | Validates |
|-----------|----------|
| `circuit_breaker_metric_test.go` | State transitions emit correct metrics |
| `circuit_breaker_alert_test.go` | Alerts fire at correct thresholds |
| `circuit_breaker_recovery_metric_test.go` | Recovery metrics accurate |

### 9.4 pprof-Based Tests

| Test File | Validates |
|-----------|----------|
| `pprof_goroutine_monitoring_test.go` | Goroutine count tracked via pprof |
| `pprof_heap_monitoring_test.go` | Heap growth monitored and bounded |
| `pprof_cpu_profile_test.go` | CPU profiling captures expected functions |
| `pprof_mutex_contention_test.go` | Mutex contention visible in profiles |

### 9.5 Dashboard Validation Tests

| Test File | Validates |
|-----------|----------|
| `grafana_dashboard_json_test.go` | Dashboard JSON is valid, all panels reference existing metrics |
| `prometheus_config_test.go` | prometheus.yml references valid targets |

### 9.6 Challenge

Create `challenges/scripts/monitoring_comprehensive_challenge.sh`:
- Verify all Prometheus metrics registered
- Verify health endpoints functional
- Verify Grafana dashboard JSON valid
- Verify prometheus.yml config valid
- Verify pprof endpoints accessible when enabled

---

## 10. Phase 8 — Lazy Loading & Performance Optimization

**Goal:** Maximum lazy loading, minimum blocking, flawless responsiveness.

### 10.1 Audit Eager Initialization

Review all `init()` functions and constructor calls for opportunities to defer work:

| Location | Current | Proposed |
|----------|---------|----------|
| Provider initialization | Eager in registry | Already lazy via LazyProvider |
| MCP adapter loading | Eager at startup | Convert to lazy (load on first use) |
| Formatter registry | Eager at startup | Convert to lazy (discover on first format request) |
| Database migrations | Eager at boot | Keep eager (must be ready before serving) |
| Config loading | Eager at boot | Keep eager (required for all operations) |

### 10.2 Add Semaphore Protection

Add `golang.org/x/sync/semaphore` to operations currently unbounded:

| Operation | Current | Proposed |
|-----------|---------|----------|
| Concurrent provider health checks | Unbounded goroutines | Semaphore-limited (10 concurrent) |
| MCP adapter preinstall | Limited by semaphore | Verify limit is appropriate |
| Background task execution | Worker pool bounded | Verify pool size config |
| Database connection pool | pgx pool bounded | Verify max connections |

### 10.3 Non-Blocking Improvements

| Location | Current Issue | Fix |
|----------|--------------|-----|
| Provider health monitoring | Blocking health checks | Add context timeout + select with default |
| MCP connection establishment | Blocking TCP connect | Add dial timeout |
| Cache write-through | Blocking on Redis write | Async write with buffered channel |

### 10.4 Benchmarks

Create benchmark suite in `tests/performance/`:

| Benchmark | Measures |
|-----------|---------|
| `BenchmarkLazyProviderInit` | First-access initialization latency |
| `BenchmarkLazyProviderCachedAccess` | Subsequent access latency (should be ~0) |
| `BenchmarkEnsembleVoting` | Voting strategy throughput |
| `BenchmarkCacheReadWrite` | Cache operations throughput |
| `BenchmarkRouterDispatch` | Request routing overhead |
| `BenchmarkMCPAdapterLoad` | MCP adapter lazy loading time |
| `BenchmarkFormatterExecution` | Formatter execution time |
| `BenchmarkHealthCheck` | Health check response time |

### 10.5 Challenge

Update `challenges/scripts/lazy_loading_validation_challenge.sh`:
- Verify sync.Once count >= 35 (up from 30)
- Verify no eager initialization of providers
- Verify semaphore protection on all concurrent operations
- Verify benchmark suite has 8+ benchmarks
- Verify non-blocking patterns on all external calls

---

## 11. Phase 9 — Documentation & Content Completion

**Goal:** Zero documentation gaps, zero duplicates, unified content organization.

### 11.1 Create Missing Documentation

#### docs/security/ Directory

Create centralized security documentation:

| File | Content |
|------|---------|
| `docs/security/README.md` | Security documentation index |
| `docs/security/scanning-guide.md` | Consolidated scanning procedures (merge from scattered docs) |
| `docs/security/vulnerability-disclosure.md` | How to report vulnerabilities |
| `docs/security/best-practices.md` | Go security best practices applied in HelixAgent |
| `docs/security/scan-resolutions.md` | Log of all scan findings and resolutions |
| `docs/security/threat-model.md` | System threat model and mitigations |

#### Other Missing Docs

| File | Content |
|------|---------|
| `docs/FAQ.md` | Frequently asked questions |
| Add TOC to `docs/MODULES.md` | Clickable navigation for 730-line document |
| Fix `docs/README.md` | Change `docs/api/grpc-api.md` reference to `docs/api/grpc.md` |

### 11.2 Consolidate Duplicate User Manuals

| Duplicate Pair | Action |
|---------------|--------|
| 32-automated-security-scanning.md (511 lines) + 42-security-scanning-guide.md (180 lines) | Merge into 32, delete 42, renumber 43-45 to 42-44 |
| 33-performance-optimization-guide.md (630 lines) + 43-performance-optimization-guide.md (194 lines) | Merge into 33, delete 43, renumber remaining |

### 11.3 Reorganize Video Courses

**Current state:** 76 files across 3 naming schemes and 6 directories.

**Target state:** All courses in single flat directory with unified naming.

**Action plan:**
1. Move all courses from subdirectories to `Website/video-courses/`
2. Rename all to unified format: `course-NN-topic.md` (zero-padded 2-digit)
3. Renumber sequentially 01-76 (preserving logical topic ordering)
4. Delete empty `courses-21-50/` directory
5. Eliminate duplicate courses (keep the most comprehensive version):
   - Security scanning: Keep best of 4 versions
   - Performance: Keep best of 2 versions
   - Concurrency: Keep best of 2 versions
   - Lazy loading: Keep best of 2 versions
6. Update `Website/video-courses/README.md` with new index

### 11.4 Extend Existing Documentation

For each newly implemented feature (AgenticEnsemble, latest debate improvements):
- Update `docs/FEATURES.md` with new feature descriptions
- Update `docs/API_REFERENCE.md` with new endpoints
- Update `docs/API_DOCUMENTATION.md` with usage examples
- Update relevant guides in `docs/guides/`

### 11.5 Extend User Manuals

Create new user manuals for features not yet covered:
- `45-agentic-ensemble-advanced.md` — Advanced agentic ensemble configuration
- `46-memory-safety-guide.md` — Understanding HelixAgent's memory safety guarantees
- `47-monitoring-metrics-guide.md` — Setting up and using monitoring dashboards

### 11.6 Extend Video Courses

Create new video course scripts for:
- AgenticEnsemble deep dive (configuration, debugging, optimization)
- Memory safety patterns in HelixAgent
- Monitoring and metrics collection
- Security scanning walkthrough
- Complete deployment from scratch

### 11.7 Update Diagrams

Create new diagrams in `docs/diagrams/src/`:
- `agentic-ensemble-flow.mmd` — AgenticEnsemble request processing
- `memory-safety-patterns.mmd` — Thread safety architecture
- `security-scanning-pipeline.mmd` — Scanning infrastructure flow (if not exists)
- `test-pyramid-updated.mmd` — Updated test architecture

### 11.8 Synchronize Governance Documents

After all changes:
- Update CLAUDE.md with new module counts, test counts, challenge counts
- Update AGENTS.md to match
- Update CONSTITUTION.md if any rules change
- Regenerate CONSTITUTION.json

### 11.9 Challenge

Create `challenges/scripts/documentation_sync_comprehensive_challenge.sh`:
- Verify no duplicate user manuals
- Verify video course numbering is sequential
- Verify docs/security/ directory exists with all required files
- Verify MODULES.md has Table of Contents
- Verify FAQ.md exists
- Verify all governance docs are synchronized

---

## 12. Phase 10 — Website & Final Validation

**Goal:** Website reflects current feature set. All validations pass.

### 12.1 Website Updates

| Page | Updates Needed |
|------|---------------|
| `features.html` | Add AgenticEnsemble, latest module count |
| `public/docs/index.html` | Add links to new documentation sections |
| `public/docs/security.html` | Link to new docs/security/ content |
| `public/docs/api.html` | Add new agentic/planning/benchmark endpoints |
| `changelog.html` | Add recent releases |

### 12.2 Build & Verify Website

```bash
cd Website
npm run build
# Verify build succeeds
# Check all internal links resolve
```

### 12.3 Final Validation Suite

Execute the complete validation pipeline:

```bash
# 1. Build all apps
make build-all

# 2. Run all tests (with infra)
make test-with-infra

# 3. Run all challenges
./challenges/scripts/run_all_challenges.sh

# 4. Run security scans
make security-scan-all

# 5. Run CI validation
make ci-validate-all

# 6. Run specific validation challenges
./challenges/scripts/dead_code_verification_challenge.sh
./challenges/scripts/memory_safety_comprehensive_challenge.sh
./challenges/scripts/coverage_completeness_challenge.sh
./challenges/scripts/stress_comprehensive_challenge.sh
./challenges/scripts/monitoring_comprehensive_challenge.sh
./challenges/scripts/documentation_sync_comprehensive_challenge.sh
./challenges/scripts/security_scan_resolution_verification_challenge.sh

# 7. Build website
cd Website && npm run build && cd ..

# 8. Final fmt/vet/lint
make fmt vet lint
```

### 12.4 Success Criteria Checklist

- [ ] `go build ./cmd/helixagent` succeeds
- [ ] All 7 apps build successfully
- [ ] `make test` passes with 0 failures
- [ ] `make test-with-infra` passes with 0 failures
- [ ] Zero blanket-skipped test files
- [ ] Zero `//nolint:unused` in production code
- [ ] Zero `.backup`/`.bak` files in internal/
- [ ] Zero unbounded maps/caches in production code
- [ ] `go test -race` passes on all packages
- [ ] All security scanner findings resolved or documented
- [ ] All challenge scripts pass
- [ ] No duplicate user manuals
- [ ] Video courses sequentially numbered
- [ ] docs/security/ directory complete
- [ ] Website builds successfully
- [ ] CLAUDE.md, AGENTS.md, CONSTITUTION.md synchronized

---

## 13. Verification Strategy

Each phase has its own verification step. Additionally, after every phase:

1. **Build verification**: `go build -mod=vendor ./cmd/helixagent`
2. **Unit test verification**: `go test -mod=vendor -short ./internal/...`
3. **Vet verification**: `go vet -mod=vendor ./internal/...`
4. **Format verification**: `gofmt -l ./internal/` (should produce no output)

Phases are committed independently with conventional commit messages:
- Phase 1: `refactor(cleanup): remove dead code and backup files`
- Phase 2: `fix(concurrency): resolve race conditions and memory leaks`
- Phase 3: `test(enablement): re-enable 183 skipped test files`
- Phase 4: `test(coverage): close coverage gaps in utils, benchmark, and other packages`
- Phase 5: `test(stress): add comprehensive stress and integration tests`
- Phase 6: `fix(security): resolve security scanner findings`
- Phase 7: `test(monitoring): add monitoring and metrics validation tests`
- Phase 8: `perf(lazy-loading): extend lazy loading and add semaphore protection`
- Phase 9: `docs(completion): consolidate documentation and content`
- Phase 10: `chore(validation): final website updates and validation`

---

## 14. Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Deleting code that's actually used | Grep for all callers before deletion; build verification after each file |
| Concurrency fixes introducing new races | Run with `-race` flag; add dedicated race tests |
| Un-skipping tests that fail | Fix underlying issues; don't just un-skip |
| Security scan false positives | Document each suppression with justification |
| Documentation drift | Final sync pass in Phase 10 |
| Breaking the build | Each phase independently verified with full build |
| Resource limit violation | All tests run with GOMAXPROCS=2, nice -n 19, ionice -c 3 |

---

## 15. Success Criteria

### Quantitative

| Metric | Target |
|--------|--------|
| Dead code lines | 0 |
| Blanket-skipped test files | 0 |
| Race conditions | 0 (verified by -race flag) |
| Unbounded caches | 0 |
| Security scan critical/high findings | 0 unresolved |
| Test coverage (all packages) | >= 80% line coverage |
| Duplicate documentation | 0 |
| Broken internal links | 0 |
| Build success (all 7 apps) | 100% |
| Challenge pass rate | 100% |

### Qualitative

- System remains responsive under 10x normal load
- No goroutine leaks under sustained operation
- Memory usage stabilizes (no growth) under steady-state load
- Documentation is comprehensive, accurate, and synchronized
- Video courses cover all major features
- Website accurately reflects current capabilities

---

*Specification authored: 2026-03-26*
*Phases: 10 sequential, independently verifiable*
*Estimated total changes: ~15,000 lines removed, ~25,000 lines added*
