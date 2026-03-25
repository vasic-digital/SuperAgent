# HelixAgent Comprehensive Completion Design Spec

**Date:** 2026-03-25
**Version:** 1.1.0
**Status:** Approved (v1.1 — corrected per spec review)
**Scope:** Full project completion — code fixes, test coverage, security, performance, documentation, website

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Assessment](#current-state-assessment)
3. [Sub-Project 1: Critical Fixes & Dead Code Elimination](#sp1-critical-fixes--dead-code-elimination)
4. [Sub-Project 2: Test Coverage Maximization](#sp2-test-coverage-maximization)
5. [Sub-Project 3: Safety & Security Hardening](#sp3-safety--security-hardening)
6. [Sub-Project 4: Performance & Monitoring](#sp4-performance--monitoring)
7. [Sub-Project 5: Documentation Completeness](#sp5-documentation-completeness)
8. [Sub-Project 6: Website & Course Updates](#sp6-website--course-updates)
9. [Execution Order & Dependencies](#execution-order--dependencies)
10. [Constraints & Compliance](#constraints--compliance)

---

## Executive Summary

HelixAgent is a production-grade AI ensemble LLM service with 8,904 Go source files, 2,424 test files, 492 challenge scripts, 41 top-level submodules, and 43 LLM providers. While the project is fundamentally strong, comprehensive research identified gaps across 6 dimensions that must be resolved to achieve full Constitution compliance:

- **12 code-level issues** (1 compiler error, 7 nil handlers, 5 dead code packages)
- **~428 unorchestrated challenge scripts** (87% of 492)
- **12 packages below 50% test file coverage**
- **32 broken documentation links**, 6 undocumented modules
- **Security scanner findings** requiring triage and resolution
- **Lazy loading gaps** in 6 subsystems

This spec decomposes the work into 6 sub-projects with clear dependencies, validation criteria, and Constitution compliance checkpoints.

---

## Current State Assessment

### Project Statistics

| Metric | Value |
|--------|-------|
| Total Go source files | 8,904 |
| Go files in `internal/` | 687 (source) + 680 (test) |
| Test files (`_test.go`) | 2,424 |
| Test file ratio | 0.99 in `internal/` (680/687) |
| Challenge scripts | 492 (64 orchestrated, ~428 not) |
| Top-level submodules | 41 (35 in MODULES.md, 6 missing) |
| LLM providers | 43 |
| Applications (cmd/) | 7 |
| Makefile targets | 202 |
| Docker compose files | 71 |
| Security scanners | 7 containerized |
| CLI agents | 48 |
| MCP adapters | 45+ |
| Code formatters | 32+ |

### Critical Issues Found

| # | Severity | Issue | File | Impact |
|---|----------|-------|------|--------|
| 1 | CRITICAL | Duplicate `GetAgentPool()` method | `internal/debate/comprehensive/integration.go:404` | Compiler error |
| 2 | HIGH | 7 handlers initialized with nil services | `internal/router/router.go:1118-1202` | Runtime 503s on 7 endpoint groups |
| 3 | HIGH | Skills routes registered inside per-request handler closure | `internal/router/router.go:780-792` | Routes re-registered on every GET /providers/:id/health request |
| 4 | HIGH | 5 adapter packages never imported | `internal/adapters/{background,observability,events,http,helixqa}/` | Dead code |
| 5 | HIGH | 87% challenges unorchestrated (~428 of 492) | `challenges/scripts/run_all_challenges.sh` | Incomplete validation |
| 6 | MEDIUM | 12 packages below 50% test coverage | Various `internal/` packages | Coverage gaps |
| 7 | MEDIUM | 32 broken documentation links | Various `docs/` files | Broken navigation |
| 8 | MEDIUM | 6 modules missing from MODULES.md | DocProcessor, HelixQA, LLMOrchestrator, VisionEngine, LLMsVerifier, MCP-Servers | Incomplete catalog |
| 9 | MEDIUM | Channel leak risks in ACP providers | `gemini_acp.go:655`, `qwen_acp.go:639` | Potential goroutine leaks |
| 10 | MEDIUM | sync.Map unbounded growth | `services/debate_performance_optimizer.go:52` | Memory pressure risk |
| 11 | LOW | 2 PlantUML sources missing rendered output | `docs/diagrams/output/` has 7 of 9 sources rendered | 2 diagrams not generated |
| 12 | LOW | GraphQL endpoint off by default | `internal/router/router.go:1207` | Feature not discoverable |
| 13 | LOW | Stale docs/README.md | `docs/README.md` | Last updated Feb 10, 2026 |

### Strengths (No Changes Needed)

- All 41 top-level submodules properly initialized and compiling
- All 43 LLM providers implemented with lazy initialization
- Comprehensive security scanning infrastructure (7 scanners, all containerized)
- Excellent resource management (sync.Once in 18+ files, semaphores in 83+ files)
- Proper goroutine lifecycle in 30+ files (WaitGroup + context patterns)
- HTTP pool properly configured (MaxIdleConns=100, timeouts configured)
- Circuit breakers with listener cap (MaxCircuitBreakerListeners=100)
- All init() functions are lightweight (deferred initialization)
- Prometheus metrics with sync.Once registration (no duplicate registration)
- OpenTelemetry integration for tracing and metrics

---

## SP1: Critical Fixes & Dead Code Elimination

**Goal:** Every line of production code compilable, reachable, and connected. Zero compiler errors, zero nil-panics, zero dead code.

**Depends on:** Nothing (execute first)

### 1.1 Compiler & Runtime Fixes

| # | Issue | File | Fix |
|---|-------|------|-----|
| 1 | Duplicate `GetAgentPool()` method | `internal/debate/comprehensive/integration.go:404` | Remove duplicate declaration (keep line 197) |
| 2 | BackgroundTaskHandler — all 10 params nil | `internal/router/router.go:1118` | Wire real services from BootManager or gate registration behind service availability check |
| 3 | DiscoveryHandler — nil service | `internal/router/router.go:1124` | Wire discovery service or gate |
| 4 | ScoringHandler — nil service | `internal/router/router.go:1137` | Wire scoring service or gate |
| 5 | VerificationHandler — all 4 params nil | `internal/router/router.go:1153` | Wire verification services or gate |
| 6 | HealthHandler — nil service | `internal/router/router.go:1168` | Wire health service or gate |
| 7 | LLMOpsHandler — nil service | `internal/router/router.go:1197` | Wire LLMOps service or gate |
| 8 | BenchmarkHandler — nil service | `internal/router/router.go:1202` | Wire benchmark service or gate |
| 9 | Skills routes inside per-request handler closure | `internal/router/router.go:780-792` | Move skills group registration outside the provider health handler, to the same level as other endpoint group registrations |

**Risk note for items #2-8:** Handler wiring must be gated behind service availability checks (nil-guard pattern). If a service is unavailable during boot, the handler should return 503 with a descriptive message rather than panic. Consider using existing lazy initialization patterns (`sync.Once`) for handler service dependencies. Incorrect wiring could introduce runtime panics on previously-working (if 503-returning) endpoints.

### 1.2 Dead Code Removal

| # | Target | Action |
|---|--------|--------|
| 10 | `internal/adapters/background/` (6 files, ~88KB, never imported) | Determine if wiring into BootManager is feasible; if not, remove |
| 11 | `internal/adapters/observability/` (never imported) | Wire into observability pipeline or remove |
| 12 | `internal/adapters/events/` (never imported) | Wire into EventBus integration or remove |
| 13 | `internal/adapters/http/` (never imported) | Wire into HTTP pool or remove |
| 14 | `internal/adapters/helixqa/` (never imported) | Wire into HelixQA integration or remove |
| 15 | `internal/challenges/userflow/results/` (empty output dirs, no Go files) | Add to `.gitignore` or remove from version control |
| 16 | `internal/background/backup/` (full package duplication, ~365KB) | Verify `internal/background/` is the canonical version, then remove the backup directory |

### 1.3 Unconnected Features

| # | Feature | Action |
|---|---------|--------|
| 17 | OAuth manager nil-safety (`router/router.go:381`) | Add nil-check guard after conditional initialization |
| 18 | GraphQL endpoint | Keep feature-flagged; document in API reference; ensure tests cover both `GRAPHQL_ENABLED=true` and `false` states |

### 1.4 Validation Criteria

- `go build ./...` — zero errors
- `go vet ./...` — zero warnings
- All existing tests pass (`go test ./internal/...`)
- No `nil` panics on any registered endpoint (HTTP smoke test)
- New challenge: `dead_code_verification_challenge.sh`

---

## SP2: Test Coverage Maximization

**Goal:** Every package at 100% test file coverage, all 492 challenges orchestrated, skip ratios minimized, fuzz testing expanded.

**Depends on:** SP1

### 2.1 Challenge Orchestration

| # | Action | Detail |
|---|--------|--------|
| 1 | Update `run_all_challenges.sh` | Add all ~428 missing challenge scripts (currently 64 orchestrated of 492 total) |
| 2 | Fix stale comments in `run_all_challenges.sh` | Lines 8 ("50 challenges") and 197 ("42 challenges") contradict actual count of 64 |
| 3 | Categorize challenges into tiers | Critical (blocks release), Standard (validates features), Extended (stress/edge) |
| 4 | Add parallel execution | Challenges within same tier run concurrently (respecting 30-40% resource limits per Constitution) |
| 5 | Add dependency ordering | Ensure infrastructure challenges run before feature challenges |

### 2.2 Package Test Coverage

12 packages below 50% test file ratio:

| Package | Current Ratio | Files Missing Tests |
|---------|---------------|---------------------|
| `agents` | 33% | `registry.go` |
| `formatters` | 33% | `executor.go`, `cache.go`, `factory.go`, `health.go`, `interface.go` |
| `tools` | 42% | `schema.go`, `handler.go`, `validation.go` |
| `notifications` | 43% | `webhook_dispatcher.go`, `hub.go`, `kafka_transport.go`, `polling_store.go`, `sse_manager.go` |
| `optimization` | 43% | `config.go`, `metrics.go`, `pipeline.go`, `optimizer.go` |
| `debate` | 44% | `lesson_bank.go` + 11 other files |
| `middleware` | 44% | `rate_limit.go`, `compression.go`, `auth.go`, `validation.go` |
| `streaming` | 47% | `entity_extractor.go`, `analytics_sink.go`, `conversation_aggregator.go`, `state_store.go` |
| `adapters` | 47% | `eventbus.go` + 4 files |
| `skills` | 47% | `registry.go`, `matcher.go`, `tracker.go`, `loader.go`, `service.go` |
| `messaging` | 47% | `middleware.go`, `task_queue.go`, `hub.go`, `broker.go`, `metrics.go` |
| `challenges` | 48% | `types.go`, `plugin.go`, `infra_provider.go`, `provider_verification.go`, `debate_formation.go` |

**Per missing file:** Create unit test (table-driven), integration test (where applicable), benchmark test.

### 2.3 Fuzz Testing Expansion

| Target | Rationale |
|--------|-----------|
| JSON request parsing (handlers) | Untrusted input from HTTP clients |
| Tool schema validation | Complex nested structures |
| MCP/ACP/LSP protocol parsing | External protocol messages |
| Prompt template rendering | User-provided template variables |
| Configuration parsing | YAML/env var edge cases |

### 2.4 Skip Ratio Reduction

106 test files with 25%+ skip ratios.

| Strategy | Action |
|----------|--------|
| Infrastructure preconditions | Expand `tests/precondition/` — add DB, Redis, API health checks |
| Consistent build tags | Group infra-dependent tests under `//go:build integration` |
| Mock fallback (unit tests ONLY) | Where tests skip due to missing server, provide in-process mock fallback for unit coverage ONLY (tagged `//go:build !integration`). Integration, E2E, and challenge tests must continue to require live infrastructure per Constitution rule 1 |

### 2.5 Test Type Completeness

Ensure all critical packages have: unit, integration, E2E, security, stress, benchmark, fuzz, chaos tests where applicable.

### 2.6 Validation Criteria

- `go test ./...` — all pass
- Every `.go` source file has a corresponding `_test.go`
- `run_all_challenges.sh` executes all 492 scripts
- Fuzz corpus replay clean
- New challenge: `test_coverage_completeness_challenge.sh`

---

## SP3: Safety & Security Hardening

**Goal:** Zero race conditions, zero memory leaks, zero goroutine leaks, all security scanners green, findings resolved.

**Depends on:** SP1 (can run concurrently with SP2, SP4)

### 3.1 Race Condition & Memory Safety

| # | Risk | File(s) | Fix |
|---|------|---------|-----|
| 1 | Channel leak in ACP providers | `gemini_acp.go:655`, `qwen_acp.go:639` | Audit all code paths; ensure channels consumed or goroutines cancelled via context in every branch |
| 2 | Goroutine cleanup loop without context select | `database/query_optimizer.go:87` | Add `context.Context` select in cleanup loop; wire into `Shutdown()`. Note: `cache/tiered_cache.go` already stores ctx/cancel (lines 160-161) but verify its `l1CleanupLoop()` selects on `ctx.Done()` |
| 3 | OAuth manager nil dereference | `router/router.go:381` | Add nil-check guard after conditional initialization |
| 4 | sync.Map unbounded growth | `services/debate_performance_optimizer.go:52` | Add LRU eviction wrapper with configurable max entries |
| 5 | Circuit breaker near-cap warning | `llm/circuit_breaker.go:76` | Add monitoring metric when listener count > 80% of MaxCircuitBreakerListeners |

### 3.2 Goroutine Lifecycle Audit

| Action | Scope |
|--------|-------|
| Audit all `go func()` calls | Verify every goroutine has: WaitGroup tracking OR context cancellation, defer recovery, reachable termination |
| Verify all `Shutdown()` methods | Every service with background goroutines: `cancel()` + `wg.Wait()` |
| Add `goleak` detector | Use `go.uber.org/goleak` in `TestMain` for critical packages |
| Expand goroutine challenge | `goroutine_lifecycle_challenge.sh` validates all new code |

### 3.3 Race Condition Detection

| Action | Detail |
|--------|--------|
| Full `-race` test run | `go test -race ./...` across all packages; fix any detected races |
| Expand race challenge | `race_condition_challenge.sh` covers debate optimizer, streaming, notifications |
| Make `-race` mandatory | `make test-race` covers all packages |

### 3.4 Security Scanning Execution & Resolution

For each of the 7 scanners:

| Scanner | Make Target | Report Location |
|---------|-------------|-----------------|
| Snyk | `make security-scan-snyk` | `reports/security/snyk-*.json` |
| SonarQube | `make security-scan-sonarqube` | `reports/security/SONARQUBE_REPORT.md` |
| Trivy | `make security-scan-trivy` | `reports/security/trivy-report.json` |
| Gosec | `make security-scan-gosec` | `reports/security/gosec-report.json` |
| Semgrep | `make security-scan-semgrep` | `reports/security/semgrep-report.json` |
| KICS | `make security-scan-kics` | `reports/security/kics-report.json` |
| Grype | `make security-scan-grype` | `reports/security/grype-report.json` |

**Triage process per scanner:**
1. Execute scan, capture report
2. Categorize: Critical (fix immediately), High (fix in SP3), Medium (document + plan), Low (accept with justification)
3. Apply fixes
4. Re-scan to confirm resolution
5. Update `SECURITY_SCAN_SUMMARY.md`

### 3.5 Stress Testing Expansion

| Target | Test Description |
|--------|-----------------|
| Rate limiter under load | 10K concurrent requests, verify no bypass |
| Ensemble with all providers failing | Graceful degradation, no goroutine leak |
| Debate maximum participants | 25 concurrent LLM calls with semaphore saturation |
| WebSocket/SSE connection storm | 1K simultaneous streaming connections |
| Cache stampede | Cold cache with 100 concurrent identical requests |
| Database connection pool exhaustion | All connections busy, verify timeout behavior |

### 3.6 Validation Criteria

- `go test -race ./...` — clean
- All 7 security scanners produce clean reports (zero HIGH/CRITICAL unresolved)
- `goleak` passes in all critical packages
- Stress tests pass at 30-40% resource limits
- New challenges: `safety_comprehensive_challenge.sh`, `security_scan_resolution_challenge.sh`

---

## SP4: Performance & Monitoring

**Goal:** Every service instrumented, monitoring tests validate collection, lazy loading everywhere, benchmarks prove optimization.

**Depends on:** SP1 (can run concurrently with SP2, SP3)

### 4.1 Lazy Loading Gaps

| # | Target | Fix |
|---|--------|-----|
| 1 | Handler service initialization (router) | Wrap in `sync.Once` — initialize on first request per handler group |
| 2 | MCP adapter loading (45+ at startup) | Lazy-load each adapter on first `/v1/mcp` call for that adapter |
| 3 | Formatter registry (32+ at boot) | Load formatters on-demand when `POST /v1/format` specifies language |
| 4 | VectorDB connections | Connect on first vector operation per store |
| 5 | Embedding provider clients (6 at init) | Lazy-init each provider on first embedding request |
| 6 | BigData components | Guard behind env check + `sync.Once` per component |

### 4.2 Monitoring Tests

| Test | Validates |
|------|-----------|
| `TestPrometheusMetricsRegistered` | All expected metric names exist in registry |
| `TestMetricsIncrementOnRequest` | HTTP request increments `helixagent_requests_total` |
| `TestCircuitBreakerMetricsOnFailure` | Provider failure updates circuit breaker gauge |
| `TestLatencyHistogramBuckets` | Response time lands in correct histogram bucket |
| `TestTokenCountAccuracy` | Token metrics match actual LLM response token counts |
| `TestCacheHitRatioMetric` | Cache operations update hit/miss counters |
| `TestProviderHealthMetricTransitions` | Health gauge transitions on failure/recovery |
| `TestConcurrencyMonitorSaturation` | Saturation metric reflects goroutine pressure |
| `TestDebateMetricsRoundCounting` | Debate rounds metric matches actual round count |
| `TestResourceMonitorMemoryAccuracy` | Memory metrics correlate with `runtime.MemStats` |

### 4.3 Benchmark-Driven Optimization

| Hot Path | Benchmark Name | Target |
|----------|---------------|--------|
| Ensemble voting | `BenchmarkEnsembleVoting` | < 1ms for 5 providers |
| Provider selection | `BenchmarkProviderSelection` | < 100us for 43 providers |
| HTTP pool acquire | `BenchmarkHTTPPoolAcquire` | < 10us per client |
| Cache read/write | `BenchmarkCacheReadWrite` | < 50us including serialization |
| Circuit breaker check | `BenchmarkCircuitBreakerState` | < 100ns (atomic read) |
| Tool schema validation | `BenchmarkToolSchemaValidation` | < 500us for full schema |
| Formatter registry lookup | `BenchmarkFormatterLookup` | < 1us per lookup |
| Skills matching | `BenchmarkSkillsMatch` | < 100us for 48 agents |
| MCP adapter resolution | `BenchmarkMCPAdapterResolve` | < 10us per adapter |
| Debate consensus detection | `BenchmarkConsensusDetection` | < 1ms for 25 responses |

### 4.4 Non-Blocking Improvements

| Target | Change |
|--------|--------|
| Debate log persistence | Verify write-behind buffer is bounded |
| Constitution watcher | Use `fsnotify` (inotify) instead of periodic stat if not already |
| Provider health checks | Verify dedicated goroutine pool, not blocking request path |
| SSE client registration | Non-blocking add/remove with select-default pattern |
| Cache warming | Background prefetch with bounded concurrency |

### 4.5 Semaphore & Backpressure

| Target | Improvement |
|--------|-------------|
| Debate optimizer | Add exponential backoff when semaphore saturated |
| HTTP handler concurrency | Add server-wide request limiter (configurable max in-flight) |
| Streaming connections | Cap max concurrent SSE/WebSocket per client |
| Background task queue | Queue depth metric + backpressure signal at 80% full |

### 4.6 Validation Criteria

- All monitoring tests pass and verify metric accuracy
- Benchmarks produce baseline numbers in `docs/performance/BENCHMARKS.md`
- Lazy loading verified: reduced startup time, acceptable first-request latency
- `pprof_memory_profiling_challenge.sh` passes
- New challenges: `monitoring_metrics_accuracy_challenge.sh`, `lazy_loading_comprehensive_challenge.sh`, `benchmark_regression_challenge.sh`

---

## SP5: Documentation Completeness

**Goal:** Zero broken links, every module documented, all diagrams rendered, SQL indexed, governance docs synchronized.

**Depends on:** SP1, SP2, SP3, SP4 (documents final state)

### 5.1 Broken Link Resolution (32 links)

| Area | Count | Fix Strategy |
|------|-------|--------------|
| `docs/deployment/` | 5 | Create missing files or redirect to correct paths |
| `docs/api/` | 4 | Fix relative paths, correct anchors |
| `docs/guides/` | 8 | Create messaging sub-docs, fix absolute-to-relative paths |
| `docs/sdk/` | 2 | Create `docs/security/AUTHENTICATION.md`, `docs/operations/RATE_LIMITING.md` |
| `docs/monitoring/` | 1 | Create `docs/observability/OPENTELEMETRY.md` |
| `docs/user/` | 1 | Create or redirect |
| `docs/development/` | 1 | Fix path |
| Other | 10 | Fix individually |

### 5.2 Module Documentation (6 modules)

For DocProcessor, HelixQA, LLMOrchestrator, VisionEngine, LLMsVerifier, MCP-Servers:

- Create `docs/` directory in modules that lack one (DocProcessor, HelixQA, LLMOrchestrator, VisionEngine)
- Add all 6 to `docs/MODULES.md` (35 -> 41 modules)
- Update MODULES.md header text from "33 independent Go modules" to match actual count (41)
- Update CLAUDE.md Extracted Modules section
- Update AGENTS.md with module descriptions
- Sync CONSTITUTION.md if needed

### 5.3 Diagram Generation

- Render 2 missing PlantUML sources to SVG/PNG/PDF in `docs/diagrams/output/` (7 of 9 already rendered; missing: `debate-orchestration-flow`, `goroutine-lifecycle`, `lazy-loading-architecture`, `module-dependency-graph`, `security-scanning-pipeline`, `test-pyramid` — verify which 2 are actually missing)
- Sync `docs/diagrams/rendered/` (25 SVGs) with `docs/diagrams/output/` (7 per format)
- Add module dependency diagrams for all 41 modules

### 5.4 SQL Documentation

- Create `sql/README.md` — index of all 21 schema files
- Create `sql/SCHEMA_GUIDE.md` — ER diagram, migration order, naming conventions

### 5.5 Documentation Cleanup

- Consolidate 5 duplicate API reference files into single `docs/api/API_REFERENCE.md`
- Remove `MODULES.md.backup` and `.backup2`
- Update stale `docs/README.md` (Feb 10 -> current)
- Designate `Website/` as canonical for user manuals and video courses
- Archive old completion reports in `docs/reports/`

### 5.6 User Manual Updates (37 -> 44)

New manuals: DocProcessor, HelixQA, LLMOrchestrator, VisionEngine, Security Scanning, Performance Optimization, Full Module Integration.

### 5.7 Video Course Extensions (37 -> 43)

New courses: Course 70 (DocProcessor), 71 (HelixQA), 72 (LLMOrchestrator), 73 (VisionEngine), 74 (Security Scanning), 75 (Performance Tuning).

### 5.8 Synchronization Verification

- CLAUDE.md <-> AGENTS.md <-> CONSTITUTION.md — all 41 modules listed consistently
- Expand `documentation_completeness_challenge.sh`
- Create `docs_sync_challenge.sh`

### 5.9 Validation Criteria

- Zero broken links (automated link checker)
- All 41 modules in MODULES.md
- All diagram output directories populated
- SQL schema fully indexed
- All 3 governance docs synchronized
- New challenge: `documentation_completeness_v2_challenge.sh`

---

## SP6: Website & Course Updates

**Goal:** Website reflects current state, courses cover every module, manuals are complete step-by-step, changelog current.

**Depends on:** SP5

### 6.1 Website Content Updates

| Page | Updates |
|------|---------|
| `Website/public/index.html` | Module count 35->41, feature highlights for 6 new modules, date |
| `Website/public/features.html` | Sections for 6 new modules, security scanning (7 scanners), test stats |
| `Website/public/pricing.html` | Verify feature matrix |
| `Website/public/changelog.html` | Entries for all SP1-SP5 work |
| `Website/public/contact.html` | Verify links |

### 6.2 Website Technical

- `npm run build` — CSS/JS minification
- Responsive design verification
- SEO metadata update
- Internal link validation

### 6.3 Video Course Production (6 new)

- Course 70: DocProcessor Deep Dive
- Course 71: HelixQA Autonomous Testing
- Course 72: LLMOrchestrator Mastery
- Course 73: VisionEngine & Navigation
- Course 74: Security Scanning Pipeline
- Course 75: Performance Tuning

Update `VIDEO_METADATA.md`, `COURSE_OUTLINE.md`, `INSTRUCTOR_GUIDE.md`.

### 6.4 Existing Course Updates

- Courses 01-05: Add 4 new module references
- Course 07 (debate): Performance optimizer, reflexion framework
- Course 12 (MCP): Updated adapter count
- Course 15 (memory): HelixMemory fusion pipeline
- Course 17 (security): All 7 scanners

### 6.5 User Manual Production (7 new)

- Manual 38: DocProcessor
- Manual 39: HelixQA
- Manual 40: LLMOrchestrator
- Manual 41: VisionEngine
- Manual 42: Security Scanning
- Manual 43: Performance Optimization
- Manual 44: Full Module Integration

### 6.6 Changelog

Comprehensive `CHANGELOG.md` entry covering SP1-SP6.

### 6.7 Validation Criteria

- All pages render correctly (`npm run preview`)
- 43 video course files with consistent format
- 44 user manuals with numbered step-by-step instructions
- `npm run build` produces minified assets
- New challenge: `website_content_completeness_challenge.sh`

---

## Execution Order & Dependencies

```
SP1 (Critical Fixes) ──┬── SP2 (Test Coverage)    ──┐
                        ├── SP3 (Safety & Security)  ├── SP5 (Documentation) ── SP6 (Website & Courses)
                        └── SP4 (Performance)        ┘
```

- **SP1** must complete first (foundation for all other work)
- **SP2, SP3, SP4** can execute concurrently after SP1
- **SP5** waits for SP2+SP3+SP4 (documents final state)
- **SP6** follows SP5 (website reflects documented state)

### Estimated Scope

| Sub-Project | New Files | Modified Files | New Tests | New Challenges |
|-------------|-----------|----------------|-----------|----------------|
| SP1 | 1 | ~15 | ~5 | 1 |
| SP2 | ~60 | ~5 | ~60 | 2 |
| SP3 | ~10 | ~15 | ~20 | 2 |
| SP4 | ~15 | ~20 | ~25 | 3 |
| SP5 | ~30 | ~50 | 0 | 2 |
| SP6 | ~20 | ~15 | 0 | 1 |
| **Total** | **~136** | **~120** | **~110** | **11** |

---

## Constraints & Compliance

All work must adhere to:

### Constitution Rules (26 mandatory)

- **No CI/CD Pipelines** — no GitHub Actions, GitLab CI, or any automated pipeline
- **No Git Hooks** — no pre-commit, pre-push hooks
- **SSH Only** for all Git operations
- **30-40% Resource Limits** — GOMAXPROCS=2, nice -n 19, ionice -c 3 for all tests/challenges
- **Container-Based Builds** — all releases inside Docker/Podman
- **Mandatory Container Orchestration Flow** — HelixAgent binary handles all container lifecycle
- **HTTP/3 (QUIC) Primary** — Brotli compression, HTTP/2 fallback only
- **Non-Interactive Execution** — no interactive prompts, no sudo password requests

### Code Style

- Standard Go conventions, `gofmt` formatting
- Imports: stdlib / third-party / internal (blank line separated)
- Line length <= 100 chars
- Table-driven tests, `testify`, naming `Test<Struct>_<Method>_<Scenario>`
- Errors: always check, wrap with `fmt.Errorf("...: %w", err)`

### Safety

- No mocks in production code
- Changes must not break existing functionality
- All changes verified from multiple angles: compile, test, runtime, structure

### Documentation

- Every new component: README.md, CLAUDE.md, AGENTS.md, docs/
- CLAUDE.md <-> AGENTS.md <-> CONSTITUTION.md synchronized
- Conventional Commits: `<type>(<scope>): <description>`

---

*This design spec was collaboratively developed through brainstorming and approved section-by-section on 2026-03-25.*
