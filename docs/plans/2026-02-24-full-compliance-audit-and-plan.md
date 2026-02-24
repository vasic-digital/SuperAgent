# HelixAgent Full Compliance Audit Report & Phased Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Date:** 2026-02-24
**Version:** 1.0.0
**Scope:** Complete audit of all modules, tests, documentation, security, performance, and operational readiness

---

## PART 1: COMPREHENSIVE AUDIT REPORT

### 1.1 Executive Summary

HelixAgent is a mature, production-ready AI ensemble system with 28 extracted modules, 1,686+ test files, 413 challenge scripts, 50+ Prometheus metrics, and comprehensive infrastructure. However, the audit identified **5 critical gaps** (zero-test modules), **3 placeholder implementations**, **7 undocumented env vars**, **8 GraphQL placeholder resolvers**, and **40+ internal files without dedicated tests**. This plan addresses every finding systematically.

---

### 1.2 CRITICAL: Modules With Zero Test Files

These 5 extracted modules have production code but **zero test files of any type**:

| # | Module | Package | Production Files | Lines (est.) | Severity |
|---|--------|---------|-----------------|-------------|----------|
| 1 | **Planning** | `digital.vasic.planning` | `hiplan.go`, `mcts.go`, `tree_of_thoughts.go` | ~900 | CRITICAL |
| 2 | **LLMOps** | `digital.vasic.llmops` | `evaluator.go`, `experiments.go`, `integration.go`, `prompts.go`, `types.go` | ~1200 | CRITICAL |
| 3 | **Benchmark** | `digital.vasic.benchmark` | `runner.go`, `types.go`, `integration.go` | ~800 | CRITICAL |
| 4 | **SelfImprove** | `digital.vasic.selfimprove` | `feedback.go`, `integration.go`, `optimizer.go`, `reward.go`, `types.go` | ~1100 | CRITICAL |
| 5 | **Agentic** | `digital.vasic.agentic` | `workflow.go` | ~400 | CRITICAL |

**Constitution Violation:** CONST-001 "100% Test Coverage" — every component MUST have unit, integration, E2E, security, stress, benchmark, and automation tests.

---

### 1.3 HIGH: Placeholder / Unfinished Implementations

| # | File | Line | Issue | Impact |
|---|------|------|-------|--------|
| 1 | `internal/streaming/conversation_aggregator.go` | 123 | Hardcoded `avgResponseTime = 500.0 // Placeholder` | Analytics accuracy |
| 2 | `internal/llmops/evaluator.go` | 254 | Hardcoded `result.Scores[metric] = 0.8 // Placeholder` | Evaluation accuracy |
| 3 | `internal/services/discovery/discoverer.go` | 272 | mDNS discovery not implemented, TCP fallback | Feature gap |
| 4 | `internal/graphql/schema.go` | Multiple | 8 resolver functions return nil (placeholder) | GraphQL incomplete |

---

### 1.4 HIGH: Internal Files Without Dedicated Tests

#### Handlers (4 files without test coverage):
| File | Lines | Notes |
|------|-------|-------|
| `internal/handlers/vision_handler.go` | ~150 | No vision_handler_test.go |
| `internal/handlers/skills_handler.go` | ~100 | No skills_handler_test.go |
| `internal/handlers/verifier_types.go` | ~40 | Type definitions untested |

#### Services (7 critical files):
| File | Lines | Impact |
|------|-------|--------|
| `internal/services/debate_formatter_integration.go` | ~120 | Debate/formatter bridge |
| `internal/services/llm_intent_classifier.go` | ~180 | Core intent classification |
| `internal/services/protocol_cache_manager.go` | ~140 | Protocol caching |
| `internal/services/concurrency_monitor.go` | ~100 | Resource monitoring |
| `internal/services/security_adapters.go` | ~160 | Security integration |
| `internal/services/ssh_command_runner.go` | ~90 | SSH execution |
| `internal/services/cli_agent_config_exporter.go` | ~150 | CLI config generation |

#### Memory Package (3 core files):
| File | Lines | Impact |
|------|-------|--------|
| `internal/memory/manager.go` | ~180 | Core memory manager |
| `internal/memory/store_memory.go` | ~150 | Memory storage |
| `internal/memory/types.go` | ~60 | Type definitions |

#### Config Package (2 files):
| File | Impact |
|------|--------|
| `internal/config/ai_debate_loader.go` | Debate config loading |
| `internal/config/multi_provider.go` | Multi-provider config |

#### Plugins Package (2 core files):
| File | Lines |
|------|-------|
| `internal/plugins/lifecycle.go` | 144 |
| `internal/plugins/registry.go` | 62 |

#### Background Package:
| File | Impact |
|------|--------|
| `internal/background/stuck_detector.go` | Stuck task detection |

#### Infrastructure Packages:
| Package | Files Without Tests |
|---------|-------------------|
| `internal/http` | `pool.go`, `quic_client.go` |
| `internal/mcp` | `connection_pool.go`, `preinstaller.go` |
| `internal/observability` | `exporter.go`, `llm_middleware.go`, `metrics.go`, `tracer.go` |
| `internal/security` | `secure_fix_agent.go` |
| `internal/streaming` | `analytics_sink.go`, `conversation_aggregator.go` |
| `internal/verifier` | `metrics.go`, `provider_types.go` |

**Total internal files needing tests: ~40**

---

### 1.5 MEDIUM: Documentation Gaps

| # | Issue | Location | Impact |
|---|-------|----------|--------|
| 1 | Toolkit missing CLAUDE.md, AGENTS.md, docs/ | `Toolkit/` | Module standards violation |
| 2 | 7 undocumented env variables | `.env.example` | Operator confusion |
| 3 | Video course needs HelixSpecifier content | `docs/video-course/` | Incomplete curriculum |
| 4 | Website needs HelixSpecifier page | `docs/website/` | Marketing gap |
| 5 | HelixSpecifier challenge not in CLAUDE.md challenges list | `CLAUDE.md` | Documentation sync |

**Missing .env.example entries:**
- `CRUSH_CLI`, `CRUSH_VERSION`
- `HELIXCODE`, `HELIXCODE_VERSION`
- `KILOCODE`, `KILOCODE_VERSION`
- `ENABLE_PPROF`

---

### 1.6 MEDIUM: Race Condition & Memory Safety

| # | Issue | File | Severity |
|---|-------|------|----------|
| 1 | Goroutine without context cancellation | `internal/handlers/discovery_handler.go:~180` | MEDIUM |
| 2 | Global singletons without graceful shutdown | `internal/services/protocol_monitor.go`, `internal/services/request_service.go` | LOW |
| 3 | SonarQube quality gate disabled | `sonar-project.properties` | LOW |

---

### 1.7 MEDIUM: Missing Challenge Scripts

Modules/features without dedicated challenge scripts:
- Agentic workflows
- LLMOps operations
- SelfImprove feedback loops
- Planning algorithms
- Benchmark execution
- Adapter layer (20+ files)
- HTTP/3 QUIC transport
- Lazy loading verification
- GraphQL resolvers

---

### 1.8 LOW: Lazy Loading Opportunities

| Component | Current | Recommendation |
|-----------|---------|---------------|
| MCP Adapters (45+) | Eager via registry | Defer until first use |
| Debate Topology Builders | Eager loading | Load on-demand by type |
| Format Providers (32+) | Eager via registry | Lazy formatter discovery |
| GraphQL Schema | Eager init | Defer until first GraphQL request |

---

### 1.9 Positive Findings (No Action Required)

- **1,686+ test files** across the project
- **413 challenge scripts** with 900+ tests
- **50+ Prometheus metrics** fully instrumented
- **Full OpenTelemetry** integration with Jaeger/Zipkin/Langfuse
- **11 docker-compose files** all complete and functional
- **108+ Makefile targets** all operational
- **18 SQL schema files** comprehensive and migrated
- **Security scanning** (Snyk, SonarQube, Trivy, Gosec) all containerized
- **All 27 extracted modules** have README.md, CLAUDE.md, AGENTS.md, docs/
- **Constitution framework** with 26 mandatory rules, synchronized
- **Build-tag conditional compilation** properly implemented
- **sync.Once** lazy patterns used correctly in 5+ locations
- **Resource cleanup** (defer Close, ticker.Stop, ctx.Done) applied throughout
- **Race detection** configured in Makefile (`make test-race`)

---

## PART 2: PHASED IMPLEMENTATION PLAN

### Phase Overview

| Phase | Name | Tasks | Priority | Estimated Effort |
|-------|------|-------|----------|-----------------|
| **1** | Critical Module Tests | 5 modules × 7 test types | CRITICAL | Large |
| **2** | Placeholder Fixes & Dead Code | 4 fixes + GraphQL resolvers | HIGH | Medium |
| **3** | Internal Package Tests | 40+ files need tests | HIGH | Large |
| **4** | Safety & Security | Race fixes, Snyk/SonarQube scan, memory safety | HIGH | Medium |
| **5** | Performance & Optimization | Lazy loading, semaphores, non-blocking, monitoring tests | MEDIUM | Medium |
| **6** | Challenge Scripts | 9 new challenge scripts | MEDIUM | Medium |
| **7** | Stress & Integration Tests | Comprehensive stress/load tests | MEDIUM | Large |
| **8** | Documentation & Media | Toolkit docs, env vars, video courses, website, manuals | MEDIUM | Medium |
| **9** | Final Validation & Polish | Full CI run, constitution sync, coverage verification | LOW | Small |

---

### Phase 1: Critical Module Tests (CRITICAL)

**Goal:** Bring 5 zero-test modules to full test coverage across all 7 required test types.

Each module requires: unit tests, integration tests, E2E tests, security tests, stress tests, benchmark tests, automation tests.

All test files MUST include `runtime.GOMAXPROCS(2)` in `init()` per resource limits policy.

#### Task 1.1: Planning Module Tests

**Files to create:**
- `Planning/planning/hiplan_test.go` — Unit tests for HiPlan algorithm
- `Planning/planning/mcts_test.go` — Unit tests for Monte Carlo Tree Search
- `Planning/planning/tree_of_thoughts_test.go` — Unit tests for Tree of Thoughts
- `Planning/tests/integration/integration_test.go` — Cross-algorithm pipeline tests
- `Planning/tests/e2e/e2e_test.go` — Full planning workflow simulation
- `Planning/tests/security/security_test.go` — Input validation, resource exhaustion
- `Planning/tests/stress/stress_test.go` — Concurrent planning, saturation
- `Planning/tests/benchmark/benchmark_test.go` — Performance measurement
- `Planning/tests/automation/automation_test.go` — API contract, structure verification

**Read first:** `Planning/planning/hiplan.go`, `mcts.go`, `tree_of_thoughts.go`

**Test patterns:**
- Table-driven tests with `testify`
- Naming: `Test<Struct>_<Method>_<Scenario>`
- Resource limits: `runtime.GOMAXPROCS(2)` in `init()`
- Benchmark: `func Benchmark<Component>(b *testing.B)`

**Verification:**
```bash
cd Planning && GOMAXPROCS=2 go test -count=1 -p 1 ./...
cd Planning && GOMAXPROCS=2 go test -count=1 -race -p 1 ./...
cd Planning && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

#### Task 1.2: LLMOps Module Tests

**Files to create:**
- `LLMOps/llmops/evaluator_test.go` — Evaluator unit tests
- `LLMOps/llmops/experiments_test.go` — A/B experiment tests
- `LLMOps/llmops/integration_test.go` — Integration adapter tests
- `LLMOps/llmops/prompts_test.go` — Prompt versioning tests
- `LLMOps/llmops/types_test.go` — Type validation tests
- `LLMOps/tests/integration/integration_test.go` — Cross-component pipeline
- `LLMOps/tests/e2e/e2e_test.go` — Full LLMOps workflow
- `LLMOps/tests/security/security_test.go` — Input validation, data isolation
- `LLMOps/tests/stress/stress_test.go` — Concurrent evaluation, experiment load
- `LLMOps/tests/benchmark/benchmark_test.go` — Performance baselines
- `LLMOps/tests/automation/automation_test.go` — API contract verification

**Read first:** `LLMOps/llmops/evaluator.go`, `experiments.go`, `integration.go`, `prompts.go`, `types.go`

#### Task 1.3: Benchmark Module Tests

**Files to create:**
- `Benchmark/benchmark/runner_test.go` — Runner unit tests
- `Benchmark/benchmark/types_test.go` — Type validation tests
- `Benchmark/benchmark/integration_test.go` — Integration adapter tests
- `Benchmark/tests/integration/integration_test.go` — Cross-component tests
- `Benchmark/tests/e2e/e2e_test.go` — Full benchmark workflow
- `Benchmark/tests/security/security_test.go` — Input validation
- `Benchmark/tests/stress/stress_test.go` — Concurrent benchmarking
- `Benchmark/tests/benchmark/benchmark_test.go` — Meta-benchmarks
- `Benchmark/tests/automation/automation_test.go` — API contract verification

**Read first:** `Benchmark/benchmark/runner.go`, `types.go`, `integration.go`

#### Task 1.4: SelfImprove Module Tests

**Files to create:**
- `SelfImprove/selfimprove/feedback_test.go` — RLHF feedback tests
- `SelfImprove/selfimprove/optimizer_test.go` — Optimizer unit tests
- `SelfImprove/selfimprove/reward_test.go` — Reward modelling tests
- `SelfImprove/selfimprove/integration_test.go` — Integration adapter tests
- `SelfImprove/selfimprove/types_test.go` — Type validation
- `SelfImprove/tests/integration/integration_test.go` — Cross-component pipeline
- `SelfImprove/tests/e2e/e2e_test.go` — Full self-improve workflow
- `SelfImprove/tests/security/security_test.go` — Input validation, safety
- `SelfImprove/tests/stress/stress_test.go` — Concurrent optimization
- `SelfImprove/tests/benchmark/benchmark_test.go` — Performance measurement
- `SelfImprove/tests/automation/automation_test.go` — API contract verification

**Read first:** `SelfImprove/selfimprove/feedback.go`, `optimizer.go`, `reward.go`, `integration.go`, `types.go`

#### Task 1.5: Agentic Module Tests

**Files to create:**
- `Agentic/agentic/workflow_test.go` — Workflow unit tests (graph nodes, edges, execution)
- `Agentic/tests/integration/integration_test.go` — Multi-step workflow tests
- `Agentic/tests/e2e/e2e_test.go` — Full agentic workflow simulation
- `Agentic/tests/security/security_test.go` — Input validation, sandbox testing
- `Agentic/tests/stress/stress_test.go` — Concurrent workflow execution
- `Agentic/tests/benchmark/benchmark_test.go` — Workflow execution performance
- `Agentic/tests/automation/automation_test.go` — API contract verification

**Read first:** `Agentic/agentic/workflow.go`

---

### Phase 2: Placeholder Fixes & Dead Code Elimination (HIGH)

**Goal:** Replace all placeholder values with proper implementations and complete GraphQL resolvers.

#### Task 2.1: Fix Streaming Analytics Placeholder

**File:** `internal/streaming/conversation_aggregator.go:123`

**Step 1:** Read the file to understand the full context around line 123
**Step 2:** Replace `avgResponseTime = 500.0 // Placeholder` with proper calculation from available response data, or use a sentinel value (-1) when no data available
**Step 3:** Write test in `internal/streaming/conversation_aggregator_test.go` covering both paths (data available / no data)
**Step 4:** Run `go test -v ./internal/streaming/...`

#### Task 2.2: Fix LLMOps Evaluator Placeholder

**File:** `internal/llmops/evaluator.go:254`

**Step 1:** Read the file to understand the evaluation pipeline
**Step 2:** Replace `result.Scores[metric] = 0.8 // Placeholder` with proper default scoring using signal-based heuristics (response length, keyword presence, structure analysis)
**Step 3:** Write test covering nil-evaluator fallback path
**Step 4:** Run `go test -v ./internal/llmops/...`

#### Task 2.3: Implement mDNS Discovery

**File:** `internal/services/discovery/discoverer.go:272`

**Step 1:** Read discoverer.go to understand the discovery interface
**Step 2:** Implement mDNS discovery using `net` package or `hashicorp/mdns`
**Step 3:** Keep TCP as fallback when mDNS unavailable
**Step 4:** Write tests for both mDNS and TCP paths
**Step 5:** Run `go test -v ./internal/services/discovery/...`

#### Task 2.4: Complete GraphQL Resolvers

**File:** `internal/graphql/schema.go` (416 lines)

**Step 1:** Read schema.go to identify all nil-returning resolvers
**Step 2:** Implement proper resolvers for: Debates, Tasks, VerificationResults, SubmitDebateResponse, CreateTask, CancelTask, RefreshProvider
**Step 3:** Connect resolvers to actual services (DebateService, TaskQueue, ProviderRegistry)
**Step 4:** Update `internal/graphql/schema_test.go` — change `assert.Nil(t, result)` to proper assertions
**Step 5:** Run `go test -v ./internal/graphql/...`

#### Task 2.5: Document Environment Variables

**File:** `.env.example`

Add entries for:
```
# CLI Agent Detection
CRUSH_CLI=                          # Path to Crush CLI binary
CRUSH_VERSION=                      # Crush version override
HELIXCODE=                          # Path to HelixCode binary
HELIXCODE_VERSION=                  # HelixCode version override
KILOCODE=                           # Path to KiloCode binary
KILOCODE_VERSION=                   # KiloCode version override

# Profiling
ENABLE_PPROF=false                  # Enable pprof endpoint on /debug/pprof
```

---

### Phase 3: Internal Package Tests (HIGH)

**Goal:** Add dedicated tests for all 40+ internal files identified without coverage.

#### Task 3.1: Handler Tests

**Create:**
- `internal/handlers/vision_handler_test.go`
- `internal/handlers/skills_handler_test.go`
- `internal/handlers/verifier_types_test.go`

**Pattern:** Table-driven tests with httptest.NewRecorder, gin.CreateTestContext

#### Task 3.2: Service Tests

**Create:**
- `internal/services/debate_formatter_integration_test.go`
- `internal/services/llm_intent_classifier_test.go`
- `internal/services/protocol_cache_manager_test.go`
- `internal/services/concurrency_monitor_test.go`
- `internal/services/security_adapters_test.go`
- `internal/services/ssh_command_runner_test.go`
- `internal/services/cli_agent_config_exporter_test.go`

**Pattern:** Mock interfaces for external deps in unit tests, real connections in integration tests

#### Task 3.3: Memory Package Tests

**Create:**
- `internal/memory/manager_test.go`
- `internal/memory/store_memory_test.go`
- `internal/memory/types_test.go`

#### Task 3.4: Config Package Tests

**Create:**
- `internal/config/ai_debate_loader_test.go`
- `internal/config/multi_provider_test.go`

#### Task 3.5: Plugin Package Tests

**Create:**
- `internal/plugins/lifecycle_test.go`
- `internal/plugins/registry_test.go`

#### Task 3.6: Background Package Tests

**Create:**
- `internal/background/stuck_detector_test.go`

#### Task 3.7: Infrastructure Package Tests

**Create:**
- `internal/http/pool_test.go`
- `internal/http/quic_client_test.go`
- `internal/mcp/connection_pool_test.go`
- `internal/mcp/preinstaller_test.go`
- `internal/observability/exporter_test.go`
- `internal/observability/llm_middleware_test.go`
- `internal/security/secure_fix_agent_test.go`
- `internal/streaming/analytics_sink_test.go`
- `internal/streaming/conversation_aggregator_test.go`
- `internal/verifier/metrics_test.go`
- `internal/verifier/provider_types_test.go`

**Verification for all Phase 3:**
```bash
GOMAXPROCS=2 go test -count=1 -p 1 ./internal/...
go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out
```

---

### Phase 4: Safety & Security (HIGH)

**Goal:** Fix race conditions, run security scans, apply memory safety improvements.

#### Task 4.1: Fix Discovery Handler Race Condition

**File:** `internal/handlers/discovery_handler.go:~180`

**Step 1:** Read the file to find the goroutine launch
**Step 2:** Add context parameter to the goroutine
**Step 3:** Ensure graceful shutdown on context cancellation
**Step 4:** Write test verifying goroutine terminates on context cancel

#### Task 4.2: Add Graceful Shutdown to Global Singletons

**Files:**
- `internal/services/protocol_monitor.go` — Add `Shutdown()` method to metricsCollectorInstance
- `internal/services/request_service.go` — Add `Shutdown()` method to GlobalMetricsRegistry

**Step 1:** Read both files
**Step 2:** Add shutdown hooks that flush pending data and release resources
**Step 3:** Wire shutdown into the main application shutdown sequence (cmd/helixagent/main.go)
**Step 4:** Write tests verifying clean shutdown

#### Task 4.3: Run Security Scans

**Step 1:** Start security scanning infrastructure
```bash
make security-scan-all
```

**Step 2:** Analyze results from `reports/security/`
**Step 3:** Prioritize and fix Critical/High findings
**Step 4:** Re-run scans to verify fixes
**Step 5:** Document findings and resolutions

**Note:** Use `docker-compose.security.yml` for containerized scanning. Do NOT use sudo or interactive commands.

#### Task 4.4: Enable SonarQube Quality Gate

**File:** `sonar-project.properties`

Change `sonar.qualitygate.wait=false` to `sonar.qualitygate.wait=true`

#### Task 4.5: Race Condition Testing

**Step 1:** Run full race detector:
```bash
GOMAXPROCS=2 go test -count=1 -race -p 1 ./...
```
**Step 2:** Fix any detected races
**Step 3:** Add dedicated race condition tests in `tests/stress/race_test.go`

#### Task 4.6: Memory Safety Audit

**Step 1:** Add pprof-based memory tests:
- Create `tests/stress/memory_leak_test.go`
- Test goroutine count before/after operations
- Test heap allocation growth under sustained load
- Verify all channels are closed and goroutines terminate

**Step 2:** Add deadlock detection tests:
- Create `tests/stress/deadlock_test.go`
- Test concurrent access to all shared state
- Test lock ordering under contention

---

### Phase 5: Performance & Optimization (MEDIUM)

**Goal:** Implement lazy loading, semaphore mechanisms, non-blocking patterns, and monitoring tests.

#### Task 5.1: Lazy Loading for MCP Adapters

**File:** `internal/mcp/adapters/` (45+ adapters)

**Step 1:** Read the adapter registry initialization
**Step 2:** Convert eager loading to lazy loading with sync.Once per adapter
**Step 3:** First access triggers initialization
**Step 4:** Write benchmark comparing eager vs lazy boot time

#### Task 5.2: Lazy Loading for Format Providers

**File:** `internal/formatters/` (32+ formatters)

**Step 1:** Read formatter registry
**Step 2:** Implement lazy formatter initialization — load on first format request
**Step 3:** Write benchmark measuring startup improvement

#### Task 5.3: Lazy Loading for GraphQL Schema

**File:** `internal/graphql/schema.go`

**Step 1:** Wrap schema initialization in sync.Once
**Step 2:** First GraphQL request triggers schema build
**Step 3:** Write test verifying lazy initialization

#### Task 5.4: Semaphore Mechanisms

Add bounded concurrency where missing:

**Files to audit and improve:**
- `internal/llm/ensemble.go` — Ensure bounded provider dispatch
- `internal/services/debate_service.go` — Bound concurrent debate rounds
- `internal/background/worker_pool.go` — Verify semaphore implementation
- `internal/http/pool.go` — Verify connection pool limits

**Pattern:** Use `golang.org/x/sync/semaphore` or channel-based semaphores

#### Task 5.5: Non-Blocking Patterns

**Audit and improve:**
- `internal/cache/` — Ensure cache misses don't block
- `internal/notifications/` — Ensure notification sends are async
- `internal/streaming/` — Verify stream processing is non-blocking
- `internal/background/` — Verify task submission never blocks

**Pattern:** Use buffered channels, select with default, context timeouts

#### Task 5.6: Monitoring & Metrics Tests

**Create:**
- `tests/integration/monitoring_metrics_test.go` — Verify Prometheus metrics are properly recorded
- `tests/integration/opentelemetry_spans_test.go` — Verify spans are created correctly
- `tests/stress/metrics_under_load_test.go` — Verify metrics accuracy under high concurrency

**Verify all 50+ metrics are:**
1. Registered correctly
2. Recorded on every operation
3. Accurate under concurrent access
4. Exportable via Prometheus endpoint

---

### Phase 6: Challenge Scripts (MEDIUM)

**Goal:** Create dedicated challenge scripts for all uncovered modules and features.

#### Task 6.1: AI/ML Module Challenges

**Create:**
- `challenges/scripts/planning_challenge.sh` — HiPlan, MCTS, Tree of Thoughts validation
- `challenges/scripts/llmops_challenge.sh` — Evaluator, experiments, prompt versioning
- `challenges/scripts/selfimprove_challenge.sh` — Feedback, reward, optimizer validation
- `challenges/scripts/benchmark_module_challenge.sh` — Runner, comparison, leaderboard
- `challenges/scripts/agentic_workflow_challenge.sh` — Graph execution, branching, state

**Pattern:** Follow existing challenge script structure from `helixspecifier_challenge.sh`

Each script must:
- Source `common.sh`
- Define section-based tests
- Use `assert_*` functions
- Track pass/fail counts
- Exit with proper code

#### Task 6.2: Infrastructure Challenges

**Create:**
- `challenges/scripts/adapter_layer_challenge.sh` — All 20+ adapter files
- `challenges/scripts/http3_quic_challenge.sh` — HTTP/3 transport, Brotli compression
- `challenges/scripts/graphql_challenge.sh` — All GraphQL resolvers
- `challenges/scripts/lazy_loading_challenge.sh` — Verify lazy init patterns work

#### Task 6.3: Update run_all_challenges.sh

Add all new challenges to the master orchestrator.

#### Task 6.4: Update CLAUDE.md Challenge List

Add all new challenges with test counts.

---

### Phase 7: Stress & Integration Tests (MEDIUM)

**Goal:** Comprehensive stress and integration tests validating system responsiveness and resilience.

#### Task 7.1: Flash Responsiveness Tests

**Create:** `tests/stress/responsiveness_test.go`

Tests:
- HTTP endpoint response time under 100ms for health checks
- HTTP endpoint response time under 500ms for simple completions
- WebSocket connection setup under 200ms
- SSE stream first byte under 300ms
- GraphQL query response under 200ms
- Rate limiter response (429) under 10ms
- Circuit breaker trip under 50ms

#### Task 7.2: Overload Protection Tests

**Create:** `tests/stress/overload_protection_test.go`

Tests:
- 1000 concurrent HTTP requests — no crashes, proper 429s
- 100 concurrent debate sessions — bounded by semaphore
- 500 concurrent cache operations — no deadlocks
- Memory stays within 2x baseline under sustained load
- Goroutine count bounded (no leaks)
- Background task queue backpressure works
- Provider circuit breakers trip under cascade failures

#### Task 7.3: Extended Integration Tests

**Create:** `tests/integration/full_pipeline_test.go`

Tests:
- Request → intent classification → provider selection → completion → response
- Request → debate → consensus → formatted output
- Memory store → search → recall → consolidation cycle
- RAG retrieve → rerank → augment → generate pipeline
- MCP adapter → tool execution → result formatting
- Agentic workflow → node execution → state propagation

#### Task 7.4: Saturation Tests

**Create:** `tests/stress/saturation_test.go`

Tests:
- Gradual load increase from 1 to 1000 RPS — measure degradation curve
- Sustained 500 RPS for 60 seconds — no OOM, no deadlocks
- Provider failure injection during load — fallback chain works
- Redis unavailable during load — in-memory cache takes over
- Database connection exhaustion — proper error handling

---

### Phase 8: Documentation & Media (MEDIUM)

**Goal:** Complete all documentation gaps, extend video courses, update website.

#### Task 8.1: Toolkit Module Documentation

**Create:**
- `Toolkit/CLAUDE.md` — Module overview, architecture, build, test, code style
- `Toolkit/AGENTS.md` — Agent integration points
- `Toolkit/docs/README.md` — Detailed user guide

**Read first:** `Toolkit/README.md` for existing content

#### Task 8.2: Video Course Updates

**File:** `docs/video-course/MODULE_SCRIPTS.md`

**Add sections for:**
- HelixSpecifier: 3-pillar fusion, DebateFunc injection, intent classifier
- Planning: HiPlan, MCTS, Tree of Thoughts algorithms
- LLMOps: Continuous evaluation, A/B experiments, prompt versioning
- Benchmark: SWE-bench, HumanEval, MMLU, custom benchmarks
- SelfImprove: RLHF feedback, reward modelling, optimizer
- Agentic: Graph-based workflows, conditional branching, state management
- HTTP/3 QUIC transport and Brotli compression
- Security scanning with Snyk, SonarQube, Trivy
- Lazy loading and performance optimization patterns

**File:** `docs/video-course/VIDEO_METADATA.md`

**Add metadata entries for all new video scripts**

#### Task 8.3: Website Content Updates

**File:** `docs/website/` (10 files)

**Update:**
- Features page: Add HelixSpecifier, Planning, LLMOps, Benchmark, SelfImprove, Agentic
- Architecture page: Add 3-pillar fusion diagram
- Integration page: Add HTTP/3 QUIC section
- Security page: Add Snyk/SonarQube/Trivy scanning details

#### Task 8.4: Course Outline Updates

**File:** `docs/courses/COURSE_OUTLINE.md`

**Add modules:**
- Module: HelixSpecifier — SDD Fusion Engine
- Module: AI Planning Algorithms
- Module: LLM Operations and Evaluation
- Module: Self-Improvement and RLHF
- Module: Agentic Workflows
- Module: Benchmarking and Comparison
- Lab exercises for each new module

#### Task 8.5: User Manual Updates

**Files:** `docs/manuals/`, `docs/guides/`

**Add/Update:**
- HelixSpecifier setup and configuration guide
- Security scanning step-by-step manual
- Performance tuning guide (lazy loading, semaphores)
- New module getting-started guides

#### Task 8.6: Architecture Diagram Updates

**File:** `docs/diagrams/`

**Add:**
- HelixSpecifier 3-pillar fusion flow (Mermaid)
- Lazy loading initialization sequence
- Security scanning pipeline
- Stress test architecture

#### Task 8.7: SQL Definition Updates

**File:** `sql/schema/`

**Verify and update:**
- Any new tables needed for Planning, LLMOps, Benchmark, SelfImprove, Agentic
- Index optimization based on stress test findings

---

### Phase 9: Final Validation & Polish (LOW)

**Goal:** Full CI validation, constitution sync, and coverage verification.

#### Task 9.1: Full Test Suite

```bash
make test-infra-start
GOMAXPROCS=2 go test -count=1 -p 1 ./...
GOMAXPROCS=2 go test -count=1 -race -p 1 ./...
```

#### Task 9.2: Full Challenge Suite

```bash
./challenges/scripts/run_all_challenges.sh
```

#### Task 9.3: Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Target:** Maximum achievable coverage (aim for >90% on all internal packages)

#### Task 9.4: Security Scan

```bash
make security-scan-all
```

**Verify:** All Critical/High findings resolved

#### Task 9.5: Constitution Sync

**Verify:**
- CLAUDE.md, AGENTS.md, CONSTITUTION.json all synchronized
- All 26+ mandatory rules reflected
- New challenge counts updated
- Module descriptions current

#### Task 9.6: Documentation Review

**Verify:**
- All module README.md, CLAUDE.md, AGENTS.md current
- docs/MODULES.md reflects all changes
- Video course scripts cover all modules
- Website content mentions all features
- No broken markdown links

#### Task 9.7: Commit and Push

```bash
# Stage all changes
git add <specific files>
git commit -m "feat: complete full compliance — tests, security, performance, documentation"

# Push all submodules
for sub in Planning LLMOps Benchmark SelfImprove Agentic Toolkit; do
    cd $sub && git push origin main && cd ..
done

# Push main repo
git push githubhelixdevelopment main
git push upstream main
```

---

## PART 3: DEPENDENCY GRAPH

```
Phase 1 (Critical Module Tests)
    ↓
Phase 2 (Placeholder Fixes) ← can run parallel with Phase 1
    ↓
Phase 3 (Internal Package Tests) ← depends on Phase 2 for GraphQL
    ↓
Phase 4 (Safety & Security) ← can run parallel with Phase 3
    ↓
Phase 5 (Performance & Optimization) ← depends on Phase 4 for race fixes
    ↓
Phase 6 (Challenge Scripts) ← depends on Phases 1-5 for testable code
    ↓
Phase 7 (Stress & Integration Tests) ← depends on Phases 3-5
    ↓
Phase 8 (Documentation & Media) ← depends on all code phases
    ↓
Phase 9 (Final Validation) ← depends on ALL phases
```

**Parallelizable pairs:**
- Phase 1 + Phase 2 (independent modules)
- Phase 3 + Phase 4 (different areas)
- Phase 6 + Phase 8 (challenges + documentation)

---

## PART 4: RISK ASSESSMENT

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| GraphQL resolver changes break existing endpoints | Low | High | Backward-compatible additions only |
| Lazy loading introduces startup race conditions | Medium | Medium | sync.Once + integration tests |
| Security scan findings require architectural changes | Low | High | Prioritize, defer large refactors |
| Stress tests expose real concurrency bugs | Medium | High | Fix bugs, add regression tests |
| Module test creation reveals production bugs | Medium | Medium | Fix bugs before adding more tests |

---

## PART 5: SUCCESS CRITERIA

1. **All 28 modules** have unit, integration, E2E, security, stress, benchmark, and automation tests
2. **Zero placeholder** values in production code
3. **All GraphQL resolvers** return real data (not nil)
4. **Zero race conditions** detected by `go test -race`
5. **Security scan** shows 0 Critical, 0 High findings
6. **All 40+ internal files** have dedicated test coverage
7. **Coverage >90%** on all internal packages
8. **All challenge scripts** pass (including 9+ new ones)
9. **Stress tests** show system handles 500+ concurrent requests
10. **All documentation** updated: README, CLAUDE.md, AGENTS.md, video courses, website, manuals
11. **Constitution synchronized** across CLAUDE.md, AGENTS.md, CONSTITUTION.json
12. **7 undocumented env vars** added to .env.example
13. **Lazy loading** implemented for MCP adapters, formatters, GraphQL schema
14. **Monitoring tests** validate all 50+ Prometheus metrics
15. **Toolkit module** has CLAUDE.md, AGENTS.md, docs/

---

## PART 6: ESTIMATED TOTALS

| Metric | Count |
|--------|-------|
| New test files to create | ~80 |
| New challenge scripts | 9 |
| Files to modify (fixes) | ~15 |
| Documentation files to create/update | ~25 |
| New Mermaid diagrams | 4 |
| Video course sections to add | 9 |
| Website pages to update | 4 |

---

*This plan respects all constraints from CLAUDE.md, AGENTS.md, and CONSTITUTION.json. All changes must be rock-solid, safe, non-error-prone, and MUST NOT BREAK any existing working functionality. Resource limits (GOMAXPROCS=2, nice -n 19, ionice -c 3) apply to all test execution.*
