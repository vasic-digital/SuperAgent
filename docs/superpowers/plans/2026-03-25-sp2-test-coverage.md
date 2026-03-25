# SP2: Test Coverage Maximization — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every package at 100% test file coverage, all 492 challenges orchestrated, skip ratios minimized, fuzz testing expanded to 6+ targets.

**Architecture:** Update the challenge orchestrator to include all 492 scripts with tiered execution, add missing test files for 12 under-covered packages (60+ new test files), expand fuzz testing from 1 to 6 targets, and reduce skip ratios by improving precondition infrastructure.

**Tech Stack:** Go 1.25.3, testify v1.11.1, `testing.F` (native fuzz)

**Spec:** `docs/superpowers/specs/2026-03-25-comprehensive-completion-design.md` (SP2 section)

**Depends on:** SP1 complete

---

### Task 1: Fix Stale Comments in run_all_challenges.sh

**Files:**
- Modify: `challenges/scripts/run_all_challenges.sh`

- [ ] **Step 1: Find and fix stale comments**

Search for lines referencing incorrect challenge counts:
```bash
grep -n 'challenges' challenges/scripts/run_all_challenges.sh | head -20
```

Update line 8 ("50 challenges") and line 197 ("42 challenges") to match the actual CHALLENGES array count (64 currently, will grow).

- [ ] **Step 2: Commit**

```bash
git add challenges/scripts/run_all_challenges.sh
git commit -m "fix(challenges): correct stale challenge count comments in orchestrator"
```

---

### Task 2: Add All Missing Challenges to Orchestrator

**Files:**
- Modify: `challenges/scripts/run_all_challenges.sh`

- [ ] **Step 1: Generate list of all challenge scripts**

```bash
find challenges/scripts/ -name '*_challenge.sh' -o -name '*_challenge*.sh' | sort > /tmp/all_challenges.txt
```

- [ ] **Step 2: Generate list of already-orchestrated challenges**

Extract the CHALLENGES array from `run_all_challenges.sh` into `/tmp/orchestrated.txt`.

- [ ] **Step 3: Find the difference**

```bash
comm -23 /tmp/all_challenges.txt /tmp/orchestrated.txt > /tmp/missing.txt
```

- [ ] **Step 4: Categorize missing challenges into tiers**

Read each missing challenge script's header/description and categorize:
- **Tier 1 (Critical):** Infrastructure, boot, provider verification, build
- **Tier 2 (Standard):** Feature validation, protocol, debate, memory
- **Tier 3 (Extended):** Stress, performance, security scanning, advanced AI

- [ ] **Step 5: Add all missing challenges to the CHALLENGES array**

Group by tier within the array using comments:

```bash
CHALLENGES=(
    # === Tier 1: Critical (Infrastructure & Boot) ===
    "full_system_boot_challenge"
    # ... existing + new tier 1 ...

    # === Tier 2: Standard (Feature Validation) ===
    "debate_orchestrator_challenge"
    # ... existing + new tier 2 ...

    # === Tier 3: Extended (Stress & Security) ===
    "stress_test_challenge"
    # ... existing + new tier 3 ...
)
```

- [ ] **Step 6: Update the total count comment at the top of the file**

- [ ] **Step 7: Verify script syntax**

Run: `bash -n challenges/scripts/run_all_challenges.sh`
Expected: No syntax errors

- [ ] **Step 8: Commit**

```bash
git add challenges/scripts/run_all_challenges.sh
git commit -m "feat(challenges): orchestrate all 492 challenge scripts in tiered order"
```

---

### Task 3: Add Tests for agents Package (33% coverage)

**Files:**
- Create: `internal/agents/registry_test.go`
- Reference: `internal/agents/registry.go`

- [ ] **Step 1: Read registry.go to understand exported API**

- [ ] **Step 2: Write table-driven unit tests**

Test all exported functions: agent registration, lookup by name, listing all agents, capability matching. Use `testify/assert`.

```go
//go:build !integration

package agents

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
    // Table-driven tests for register/get/list operations
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/agents/ -v -short -count=1`
Expected: All pass

- [ ] **Step 4: Commit**

```bash
git add internal/agents/registry_test.go
git commit -m "test(agents): add unit tests for agent registry"
```

---

### Task 4: Add Tests for formatters Package (33% coverage)

**Files:**
- Create: `internal/formatters/executor_test.go`
- Create: `internal/formatters/cache_test.go`
- Create: `internal/formatters/factory_test.go`
- Create: `internal/formatters/health_test.go`
- Create: `internal/formatters/interface_test.go`

- [ ] **Step 1: Read each source file to understand the API**

- [ ] **Step 2: Write tests for executor.go**

Test: execution pipeline, error handling, timeout behavior, concurrent execution.

- [ ] **Step 3: Write tests for cache.go**

Test: cache hit/miss, TTL expiration, cache invalidation, concurrent access.

- [ ] **Step 4: Write tests for factory.go**

Test: formatter creation by language, unknown language handling, registry population.

- [ ] **Step 5: Write tests for health.go**

Test: health check pass/fail, formatter availability reporting.

- [ ] **Step 6: Write tests for interface.go**

Test: interface compliance, type assertions, capability checking.

- [ ] **Step 7: Run all formatter tests**

Run: `go test ./internal/formatters/ -v -short -count=1`
Expected: All pass

- [ ] **Step 8: Commit**

```bash
git add internal/formatters/*_test.go
git commit -m "test(formatters): add unit tests for executor, cache, factory, health, interface"
```

---

### Task 5: Add Tests for tools Package (42% coverage)

**Files:**
- Create: `internal/tools/schema_test.go`
- Create: `internal/tools/handler_test.go`
- Create: `internal/tools/validation_test.go`

- [ ] **Step 1: Read source files**

- [ ] **Step 2: Write tests for schema.go**

Test: schema definition, parameter types (snake_case validation), tool registration.

- [ ] **Step 3: Write tests for handler.go**

Test: tool dispatch, unknown tool handling, parameter passing.

- [ ] **Step 4: Write tests for validation.go**

Test: required params, type validation, nested object validation, edge cases.

- [ ] **Step 5: Run and commit**

```bash
go test ./internal/tools/ -v -short -count=1
git add internal/tools/*_test.go
git commit -m "test(tools): add unit tests for schema, handler, validation"
```

---

### Task 6: Add Tests for middleware Package (44% coverage)

**Files:**
- Create: `internal/middleware/rate_limit_test.go`
- Create: `internal/middleware/compression_test.go`
- Create: `internal/middleware/auth_test.go`
- Create: `internal/middleware/validation_test.go`

- [ ] **Step 1: Read source files**

- [ ] **Step 2: Write rate_limit_test.go**

Test: request counting, window sliding, limit enforcement, concurrent requests, bypass prevention.

- [ ] **Step 3: Write compression_test.go**

Test: Brotli compression, gzip fallback, Accept-Encoding header parsing, minimum size threshold.

- [ ] **Step 4: Write auth_test.go**

Test: JWT validation, API key validation, missing auth header, expired tokens.

- [ ] **Step 5: Write validation_test.go**

Test: request body validation, content-type enforcement, size limits.

- [ ] **Step 6: Run and commit**

```bash
go test ./internal/middleware/ -v -short -count=1
git add internal/middleware/*_test.go
git commit -m "test(middleware): add unit tests for rate_limit, compression, auth, validation"
```

---

### Task 7: Add Tests for notifications Package (43% coverage)

**Files:**
- Create: `internal/notifications/webhook_dispatcher_test.go`
- Create: `internal/notifications/hub_test.go`
- Create: `internal/notifications/kafka_transport_test.go`
- Create: `internal/notifications/polling_store_test.go`
- Create: `internal/notifications/sse_manager_test.go`

- [ ] **Step 1-5: Read each source file and write corresponding tests**

Focus on: dispatch logic, subscription management, message routing, SSE streaming, polling behavior.

- [ ] **Step 6: Run and commit**

```bash
go test ./internal/notifications/ -v -short -count=1
git add internal/notifications/*_test.go
git commit -m "test(notifications): add unit tests for webhook, hub, kafka, polling, sse"
```

---

### Task 8: Add Tests for optimization Package (43% coverage)

**Files:**
- Create: `internal/optimization/config_test.go`
- Create: `internal/optimization/metrics_test.go`
- Create: `internal/optimization/pipeline_test.go`
- Create: `internal/optimization/optimizer_test.go`

- [ ] **Step 1-4: Read and write tests for each file**

- [ ] **Step 5: Run and commit**

```bash
go test ./internal/optimization/ -v -short -count=1
git add internal/optimization/*_test.go
git commit -m "test(optimization): add unit tests for config, metrics, pipeline, optimizer"
```

---

### Task 9: Add Tests for Remaining 6 Packages

**Files:**
- Create tests for: `streaming/` (4 files), `adapters/` (5 files), `skills/` (5 files), `messaging/` (5 files), `challenges/` (5 files), `debate/` (12 files)

- [ ] **Step 1: streaming package — 4 test files**

Create: `entity_extractor_test.go`, `analytics_sink_test.go`, `conversation_aggregator_test.go`, `state_store_test.go`

- [ ] **Step 2: adapters package — 5 test files**

Create tests for each untested adapter file (including `eventbus.go`).

- [ ] **Step 3: skills package — 5 test files**

Create: `registry_test.go`, `matcher_test.go`, `tracker_test.go`, `loader_test.go`, `service_test.go`

- [ ] **Step 4: messaging package — 5 test files**

Create: `middleware_test.go`, `task_queue_test.go`, `hub_test.go`, `broker_test.go`, `metrics_test.go`

- [ ] **Step 5: challenges package — 5 test files**

Create: `types_test.go`, `plugin_test.go`, `infra_provider_test.go`, `provider_verification_test.go`, `debate_formation_test.go`

- [ ] **Step 6: debate package — 12 test files**

Create tests for `lesson_bank.go` and the other 11 untested files in the debate package.

- [ ] **Step 7: Run all tests**

Run: `GOMAXPROCS=2 nice -n 19 go test ./internal/... -short -count=1 -p 1`
Expected: All pass

- [ ] **Step 8: Commit per package**

One commit per package, e.g.:
```bash
git commit -m "test(streaming): add unit tests for entity_extractor, analytics_sink, conversation_aggregator, state_store"
```

---

### Task 10: Expand Fuzz Testing (1 -> 6 targets)

**Files:**
- Create: `tests/fuzz/json_parsing_fuzz_test.go`
- Create: `tests/fuzz/tool_schema_fuzz_test.go`
- Create: `tests/fuzz/protocol_parsing_fuzz_test.go`
- Create: `tests/fuzz/prompt_template_fuzz_test.go`
- Create: `tests/fuzz/config_parsing_fuzz_test.go`

- [ ] **Step 1: Write JSON parsing fuzz test**

```go
//go:build fuzz

package fuzz

import (
    "testing"
)

func FuzzJSONRequestParsing(f *testing.F) {
    f.Add([]byte(`{"model":"test","messages":[{"role":"user","content":"hello"}]}`))
    f.Add([]byte(`{}`))
    f.Add([]byte(`{"messages":null}`))

    f.Fuzz(func(t *testing.T, data []byte) {
        // Parse request — should never panic
        // ... import and call the handler's parse function
    })
}
```

- [ ] **Step 2-5: Write remaining 4 fuzz tests**

Each follows the same pattern: seed corpus, fuzz function that must not panic.

- [ ] **Step 6: Run corpus replay**

Run: `go test ./tests/fuzz/ -fuzz=. -fuzztime=10s`
Expected: No crashes

- [ ] **Step 7: Commit**

```bash
git add tests/fuzz/
git commit -m "test(fuzz): expand fuzz testing to 6 targets (JSON, schema, protocol, template, config)"
```

---

### Task 11: Expand Precondition Tests

**Files:**
- Create: `tests/precondition/database_health_test.go`
- Create: `tests/precondition/redis_health_test.go`
- Create: `tests/precondition/api_health_test.go`

- [ ] **Step 1: Write database precondition test**

```go
//go:build integration

package precondition

import (
    "net"
    "os"
    "testing"
    "time"
)

func TestDatabaseConnectivity(t *testing.T) {
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    if host == "" { host = "localhost" }
    if port == "" { port = "15432" }

    conn, err := net.DialTimeout("tcp", host+":"+port, 5*time.Second)
    if err != nil {
        t.Skipf("Database not available at %s:%s: %v", host, port, err)
    }
    conn.Close()
}
```

- [ ] **Step 2-3: Write Redis and API health precondition tests**

- [ ] **Step 4: Commit**

```bash
git add tests/precondition/
git commit -m "test(precondition): add database, redis, API health precondition tests"
```

---

### Task 12: Ensure Test Type Completeness (Spec Section 2.5)

For each of the 12 under-covered packages, verify they have representation across all applicable test types. Unit tests are added in Tasks 3-9. This task adds the remaining types.

**Files:**
- Create: `tests/integration/formatters_integration_test.go`
- Create: `tests/integration/middleware_integration_test.go`
- Create: `tests/integration/skills_integration_test.go`
- Create: `tests/stress/formatters_stress_test.go`
- Create: `tests/stress/middleware_stress_test.go`
- Create benchmark tests for packages that lack them

- [ ] **Step 1: Add integration tests for critical packages**

For `formatters`, `middleware`, `skills` — write integration tests that test against real (non-mocked) dependencies. Tag with `//go:build integration`.

- [ ] **Step 2: Add benchmark tests for packages that lack them**

For `formatters`, `skills`, `tools`, `middleware` — write `Benchmark*` functions in `tests/performance/` or in-package `_bench_test.go` files.

- [ ] **Step 3: Add stress tests for middleware and formatters**

Tag with `//go:build stress`. Test concurrent access patterns at 30-40% resource limits.

- [ ] **Step 4: Run all new tests**

```bash
GOMAXPROCS=2 nice -n 19 go test ./tests/integration/ -tags integration -short -count=1 -p 1
GOMAXPROCS=2 nice -n 19 go test ./tests/stress/ -tags stress -short -count=1 -p 1
```

- [ ] **Step 5: Commit**

```bash
git add tests/integration/ tests/stress/ tests/performance/ internal/
git commit -m "test: add integration, stress, and benchmark tests for under-covered packages (spec 2.5)"
```

---

### Task 13: Create test_coverage_completeness_challenge.sh

**Files:**
- Create: `challenges/scripts/test_coverage_completeness_challenge.sh`

- [ ] **Step 1: Write challenge that verifies every .go file has a _test.go**

The challenge scans `internal/` for `.go` files (excluding `doc.go`) and checks each has a corresponding `_test.go` either in-package or in the `tests/` directory.

- [ ] **Step 2: Make executable and commit**

```bash
chmod +x challenges/scripts/test_coverage_completeness_challenge.sh
git add challenges/scripts/test_coverage_completeness_challenge.sh
git commit -m "test(challenges): add test coverage completeness challenge"
```

---

### Task 14: Final SP2 Validation

- [ ] **Step 1: Verify all tests pass**

Run: `GOMAXPROCS=2 nice -n 19 go test ./internal/... -short -count=1 -p 1`

- [ ] **Step 2: Verify fuzz corpus replays**

Run: `go test ./tests/fuzz/ -short`

- [ ] **Step 3: Run coverage completeness challenge**

Run: `./challenges/scripts/test_coverage_completeness_challenge.sh`

- [ ] **Step 4: Verify orchestrator syntax**

Run: `bash -n challenges/scripts/run_all_challenges.sh`

- [ ] **Step 5: Tag completion**

```bash
git tag sp2-complete
```
