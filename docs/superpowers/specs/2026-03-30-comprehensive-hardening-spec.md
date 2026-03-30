# HelixAgent Comprehensive Hardening Specification

**Date:** 2026-03-30
**Version:** 1.0.0
**Scope:** Full project hardening — safety, dead code, tests, security, monitoring, lazy loading, documentation, challenges

## Executive Summary

Comprehensive audit of HelixAgent (677 source files, 681 test files, 41 extracted modules) identified 8 categories of work required to achieve 100% completion across all quality dimensions. This spec decomposes the work into 8 sequential phases, each independently verifiable.

**Key Metrics:**
- 53 functions marked `//nolint:unused` (dead code)
- 14 critical concurrency safety issues
- 552 flaky `time.Sleep` calls in tests
- 20+ adapter methods returning nil (stubs)
- QA/VisionEngine endpoints missing from API docs
- Security scanning infrastructure present but needs fresh scan

---

## Phase 1: Safety & Stability (CRITICAL)

**Goal:** Fix all concurrency bugs, race conditions, goroutine leaks, and deadlock risks.

### 1.1 Gemini ACP Map Race Condition
- **File:** `internal/llm/providers/gemini/gemini_acp.go`
- **Fix:** Add proper `defer respMu.RUnlock()` in `readResponses()` goroutine (line 285). Ensure lock/unlock pairing is consistent across all map access paths.

### 1.2 Unbuffered Channel Goroutine Leaks
- **File:** `internal/formatters/executor.go`
- **Fix:** Change `resultChan` and `errorChan` in `TimeoutMiddleware` (lines 176-177) to buffered channels (capacity 1). This allows the goroutine to send and exit even if the receiver has already returned due to context timeout.

### 1.3 FormatterExecutor Batch Goroutine Safety
- **File:** `internal/formatters/executor.go`
- **Fix:** Add panic recovery to goroutines in `ExecuteBatch` (line 115). Add `sync.WaitGroup` tracking. Close `resultChan` after all goroutines complete via `wg.Wait()` in a separate goroutine.

### 1.4 Messaging Hub Health Check Goroutine Leak
- **File:** `internal/messaging/hub.go`
- **Fix:** Add `sync.WaitGroup` field. Call `wg.Add(1)` before `go h.healthCheckLoop()` (line 162). Add `defer h.wg.Done()` inside the goroutine. Call `h.wg.Wait()` in `Close()`.

### 1.5 Context.Background() Anti-pattern in Goroutines
- **Files:** `internal/messaging/hub.go:278`, `internal/database/db.go` (multiple lines)
- **Fix:** Accept `context.Context` parameter in affected methods. Propagate parent context to child operations.

### 1.6 Cache Cleanup Goroutine Tracking
- **File:** `internal/cache/tiered_cache.go`
- **Fix:** Add `sync.WaitGroup` to `TieredCache`. Track L1 cleanup goroutine (line 166). Ensure `Close()` waits for cleanup completion.

### 1.7 WebSocket Lock Ordering
- **File:** `internal/notifications/websocket_server.go`
- **Fix:** Document and enforce lock acquisition order (clientsMu before globalClientsMu). Add runtime assertions in debug builds.

### 1.8 ACP Handler Cleanup Goroutine
- **File:** `internal/handlers/acp_handler.go`
- **Fix:** Add `sync.WaitGroup` tracking for `sessionCleanupWorker` goroutine (line 112). Ensure `Shutdown()` waits for completion.

### 1.9 Query Optimizer Cleanup
- **File:** `internal/database/query_optimizer.go`
- **Fix:** Add WaitGroup tracking for cleanup loop goroutine (line 92).

### Verification
- Run `go test -race ./internal/...` — zero race conditions
- Run goroutine lifecycle challenge
- Run concurrency safety challenge
- All existing tests must continue passing

---

## Phase 2: Dead Code Elimination

**Goal:** Remove all 53 unused functions. Complete all adapter stubs.

### 2.1 Remove Unused Functions
Delete functions marked `//nolint:unused` across these files:

| File | Functions to Remove | Count |
|------|-------------------|-------|
| `handlers/openai_compatible.go` | `contains`, `containsSubstring`, `generateID`, `generateDebateDialogueResponse`, `buildDebateRoleSystemPrompt`, `extractDocumentationContent` | 6 |
| `handlers/cognee_handler.go` | `getIntParam`, `getFloatParam` | 2 |
| `services/speckit_orchestrator.go` | `savePhaseToCache`, `loadPhaseFromCache`, `saveFlowToCache`, `loadFlowFromCache`, `clearFlowCache`, `resumeFlow`, `ensureDir`, `writeFile`, `readFile`, `removeDir` | 10 |
| `services/debate_service.go` | `isUserConfirmation`, `isUserRefusal`, `getUserIntentDescription` | 3 |
| `services/context_manager.go` | `isRelevant` | 1 |
| `services/provider_discovery.go` | `maskToken` | 1 |
| `services/embedding_manager.go` | `bytesToFloat64` | 1 |
| `services/lsp_manager.go` | `openDocument` | 1 |
| `verifier/scoring.go` | `calculateSpeedScore`, `calculateEfficiencyScore`, `calculateCostScore`, `calculateCapabilityScore`, `calculateRecencyScore`, `computeWeightedScore` | 6 |
| `background/stuck_detector.go` | `checkHeartbeatTimeout`, `isProcessFrozen`, `checkResourceExhaustion`, `isIOStarved`, `isNetworkHung`, `hasMemoryLeak`, `isEndlessTaskStuck`, `min3` | 8 |
| `mcp/adapters/gitlab.go` | `formatGitLabProjects` | 1 |
| `rag/qdrant_retriever.go` | `extractMetadata`, `toFloat32Slice` | 2 |
| `middleware/rate_limit.go` | `max` | 1 |
| `modelsdev/client.go` | `doPost` | 1 |
| `memory/crdt.go` | 2 functions | 2 |
| Provider files | Various helper functions | 7 |

**Process:** For each function:
1. Verify no callers exist (grep for function name across entire codebase)
2. Remove function and `//nolint:unused` comment
3. Verify compilation
4. Run tests for affected package

### 2.2 Complete Adapter Stubs
Implement adapter methods that currently return nil:

- **Streaming adapter** (`internal/adapters/streaming/adapter.go`): `RegisterClient`, `UnregisterClient`, `RegisterGlobalClient`, `UnregisterGlobalClient`, `Dispatch`, `GRPCServerOptions` — delegate to underlying Streaming module or return `fmt.Errorf("not supported")` with clear documentation
- **Database compat** (`internal/adapters/database/compat.go`): 6 compatibility methods — grep for callers; implement delegation to Database module if called, remove entirely if no callers exist
- **Formatters adapter** (`internal/adapters/formatters/adapter.go`): 4 methods — implement delegation to Formatters module

### Verification
- `go build ./...` — clean compilation
- `go vet ./...` — no issues
- Run dead code elimination challenge
- All existing tests pass

---

## Phase 3: Test Coverage Maximum

**Goal:** Achieve maximum test coverage. Fix all flaky tests. Add missing test types.

### 3.1 Fix Flaky time.Sleep Tests
Replace `time.Sleep` with synchronization primitives in test files. Priority approach:
- Use `testify/require.Eventually` for polling conditions
- Use channels with `select` + `time.After` for timeout-based waits
- Use `sync.WaitGroup` for goroutine completion
- Target: reduce 552 `time.Sleep` calls by 80%+ (some are legitimate delay tests)

### 3.2 Add Short-Mode Guards
Add `if testing.Short() { t.Skip("...") }` to all tests that require:
- Database connections (PostgreSQL, Redis)
- External API keys
- Running containers
- Network access

### 3.3 Add Missing Package-Level Tests
These root-level packages have subdirectory tests but no tests for the root package interfaces/types themselves. Add root-level test files exercising the public API surface:
- `internal/auth/auth_test.go` — test authentication type exports and utilities
- `internal/embeddings/embeddings_test.go` — test embedding provider interface and registry
- `internal/routing/routing_test.go` — test routing type exports (semantic/ subdir already tested)
- `internal/storage/storage_test.go` — test storage abstraction interface
- `internal/vectordb/vectordb_test.go` — test vector DB abstraction interface

### 3.4 Stress Tests
Add stress tests to `tests/stress/` for:
- QA handler concurrent session creation
- VisionEngine remote pool under load
- Adapter initialization race conditions
- Memory store concurrent findings CRUD

### 3.5 Integration Tests
Add integration tests to `tests/integration/` for:
- HelixQA adapter → handler → router chain
- VisionEngine remote pool lifecycle
- Full QA session pipeline (with mock LLM)

### 3.6 Benchmark Tests
Add benchmarks for:
- Adapter type conversion performance
- VisionPool slot assignment scaling
- Finding deduplication throughput

### Verification
- `go test -short ./internal/...` — all pass
- `go test -race -short ./internal/...` — zero races
- Coverage report shows improvement

---

## Phase 4: Security Scanning

**Goal:** Run all security scanners, analyze findings, resolve all HIGH/CRITICAL issues.

### 4.1 Run Security Scans
Execute via containerized scanners (Docker/Podman):
```bash
make security-scan-gosec      # Go-specific security
make security-scan-snyk       # Dependency vulnerabilities
make security-scan-trivy      # Container + code scanning
make security-scan-sonarqube  # Code quality + security analysis
```

### 4.2 Analyze Findings
- Triage all HIGH/CRITICAL findings
- Categorize: real vulnerability vs. false positive
- Document false positives in `.gosec.yml` exclusions with justification

### 4.3 Remediate
- Fix all real HIGH/CRITICAL vulnerabilities
- Update dependency versions for known CVEs
- Add input validation where missing
- Ensure no secrets in code

### 4.4 Security Tests
- Add tests validating each remediation
- Update security challenge scripts

### Verification
- `make security-scan-gosec` — zero HIGH issues
- Security scanning challenge passes
- All security tests pass

---

## Phase 5: Monitoring & Metrics

**Goal:** Create tests that exercise monitoring and metrics collection for optimization insights.

### 5.1 Prometheus Metrics Tests
Create tests in `tests/performance/` that:
- Start HelixAgent with Prometheus endpoint
- Exercise key code paths (debate, QA, adapters)
- Collect metrics via HTTP
- Assert metric values (counters, histograms, gauges)

### 5.2 Memory Profiling Tests
Create pprof-based tests:
- Allocations during adapter lifecycle
- Memory growth during sustained load
- GC pressure under concurrent operations

### 5.3 Latency Benchmarks
Create benchmarks measuring:
- Handler response time at various concurrency levels
- Adapter initialization latency
- Provider selection latency
- Cache hit/miss latency

### Verification
- Monitoring dashboard challenge passes
- pprof memory profiling challenge passes
- All benchmark tests run without OOM

---

## Phase 6: Lazy Loading & Semaphores

**Goal:** Implement lazy initialization and non-blocking patterns for maximum responsiveness.

### 6.1 Lazy Adapter Initialization
Apply `sync.Once` pattern to all adapters that don't already use it:
- Check each adapter's `Initialize()` or `New()` for eager resource allocation
- Defer expensive operations (DB connections, network clients) to first use
- Use `sync.Once` for thread-safe lazy init

### 6.2 Semaphore for Concurrent Provider Calls
Add semaphore limiting for:
- LLM provider concurrent requests (prevent overloading)
- Debate parallel execution (cap goroutines)
- QA pipeline multi-device orchestration
Use `golang.org/x/sync/semaphore` (already in project dependencies)

### 6.3 Non-Blocking Health Checks
Ensure all health check operations are non-blocking:
- Use context with short timeouts (500ms)
- Return cached last-known status if check is in progress
- Never block startup on slow health checks

### Verification
- Lazy loading validation challenge passes
- No startup time regression
- All existing tests pass

---

## Phase 7: Documentation Complete

**Goal:** Complete all documentation to match the Constitution's requirements.

### 7.1 API Documentation
- Create `docs/api/QA_API_REFERENCE.md` — full specs for `/v1/qa/*` endpoints
- Create `docs/api/VISION_API_REFERENCE.md` — specs for vision endpoints
- Update `docs/api/API_REFERENCE.md` — add QA and vision sections

### 7.2 User Manuals
- Update `Website/user-manuals/39-helixqa-guide.md` — add autonomous pipeline, multi-device, credential discovery
- Update `Website/user-manuals/41-visionengine-guide.md` — add remote pool, llama.cpp deployment
- Create `Website/user-manuals/44-qa-api-guide.md` — REST API usage for QA
- Create `Website/user-manuals/45-challenge-framework-guide.md` — challenge development

### 7.3 Video Courses
- Update `Website/video-courses/course-71-helixqa.md` — add autonomous pipeline lessons
- Update `Website/video-courses/course-73-visionengine.md` — add remote pool lessons
- Create `Website/video-courses/course-77-qa-api-integration.md` — QA API course

### 7.4 Website Module Documentation
- Create `Website/README.md`
- Create `Website/CLAUDE.md`
- Create `Website/AGENTS.md`
- Create `Website/docs/` directory with architecture

### 7.5 Architecture Diagrams
- Create `docs/diagrams/src/qa-pipeline-flow.puml` — QA pipeline phases
- Create `docs/diagrams/src/vision-pool-architecture.puml` — VisionPool slot management
- Create `docs/diagrams/src/adapter-helixqa.puml` — HelixQA adapter integration

### 7.6 SQL Definitions
- Update `sql/schema/complete_schema.sql` with any new tables

### Verification
- Documentation completeness challenge passes
- All referenced files exist
- No broken links in documentation

---

## Phase 8: Challenge Coverage

**Goal:** Create comprehensive challenges validating all new and existing functionality.

### 8.1 New Challenge Scripts
- `challenges/scripts/concurrency_fixes_challenge.sh` — validates Phase 1 fixes
- `challenges/scripts/dead_code_audit_challenge.sh` — validates no unused functions remain
- `challenges/scripts/test_coverage_gap_challenge.sh` — validates all packages have tests
- `challenges/scripts/security_scan_results_challenge.sh` — validates zero HIGH findings
- `challenges/scripts/monitoring_metrics_challenge.sh` — validates metrics collection
- `challenges/scripts/lazy_init_comprehensive_challenge.sh` — validates lazy loading patterns
- `challenges/scripts/documentation_api_challenge.sh` — validates API docs completeness
- `challenges/scripts/flaky_test_detection_challenge.sh` — validates no time.Sleep abuse

### 8.2 Update Existing Challenges
- Update `run_all_challenges.sh` to include new challenges
- Update challenge count in CLAUDE.md

### Verification
- `./challenges/scripts/run_all_challenges.sh` — all pass
- Each new challenge passes independently

---

## Execution Order & Dependencies

```
Phase 1 (Safety) ─────┐
                       ├──> Phase 3 (Tests) ──> Phase 5 (Monitoring)
Phase 2 (Dead Code) ──┘                    ──> Phase 6 (Lazy Loading)

Phase 4 (Security) ──> independent

Phase 7 (Docs) ──> Phase 8 (Challenges) ──> final validation
```

Phases 1 & 2 must complete before Phase 3 (tests need stable code).
Phase 4 is independent (security scanning).
Phases 5 & 6 depend on Phase 3 (need test infrastructure).
Phases 7 & 8 run last (document completed work).

---

## Constraints

1. **No CI/CD pipelines** — all validation is manual via Makefile targets
2. **No interactive processes** — no sudo, no password prompts
3. **Resource limits** — GOMAXPROCS=2, nice -n 19, ionice -c 3
4. **SSH only for git** — no HTTPS URLs
5. **Container builds** — Docker/Podman for scanning infrastructure
6. **No mocks in production** — only in unit tests
7. **Backward compatibility** — all existing tests must continue passing
8. **Non-blocking** — all operations respect context cancellation

## Success Criteria

- `go build ./...` — clean compilation
- `go test -short -race ./internal/...` — all pass, zero races
- `make security-scan-gosec` — zero HIGH issues
- All challenge scripts pass
- All documentation files exist and are non-empty
- No `//nolint:unused` markers remain in production code
- No goroutine leaks detectable by stress tests
