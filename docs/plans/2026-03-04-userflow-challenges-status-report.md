# Userflow Challenges — Status Report

**Date:** 2026-03-04
**Branch:** main
**Last Commit:** `b015b0a5` (HelixAgent) / `f32b7e3` (Challenges)
**Git Status:** Clean (only LLMsVerifier submodule pointer modified, not staged)

---

## What Was Done

### Challenges Module — 9 New Testing Framework Adapters

All adapters are generic, reusable across projects, and committed to the
`Challenges/` submodule (`digital.vasic.challenges`).

| Adapter | Interface | Technology | Lines |
|---------|-----------|------------|-------|
| `SeleniumAdapter` | `BrowserAdapter` | W3C WebDriver HTTP protocol | 582 |
| `CypressCLIAdapter` | `BrowserAdapter` | Cypress CLI spec generation | 495 |
| `PuppeteerAdapter` | `BrowserAdapter` | Node.js Puppeteer scripts | 627 |
| `AppiumAdapter` | `MobileAdapter` | Appium 2.0 W3C + UiAutomator2/XCUITest | 663 |
| `MaestroCLIAdapter` | `MobileAdapter` | Maestro YAML flow generation | 434 |
| `EspressoAdapter` | `MobileAdapter` | Gradle + ADB hybrid | 504 |
| `RobolectricAdapter` | `BuildAdapter` | Android JVM tests via Gradle | 313 |
| `GRPCCLIAdapter` | `GRPCAdapter` (new) | grpcurl CLI (unary + streaming) | 351 |
| `GorillaWebSocketAdapter` | `WebSocketFlowAdapter` (new) | gorilla/websocket (thread-safe) | 314 |

**New challenge templates:** `GRPCFlowChallenge` (481 lines), `WebSocketFlowChallenge` (410 lines)

**Totals:** 11 production files (5,174 lines), 11 test files (5,601 lines), 154 test functions

### HelixAgent Integration

| File | Purpose | Lines |
|------|---------|-------|
| `internal/challenges/userflow/flows.go` | 12 API flow definitions + 12 challenge constructors | 614 |
| `internal/challenges/userflow/orchestrator.go` | Challenge orchestrator with dependency graph | 177 |
| `internal/challenges/userflow/flows_test.go` | 16 test functions (all 12 constructors verified) | 312 |
| `challenges/scripts/userflow_comprehensive_challenge.sh` | 30+ curl-based tests across 12 phases | 273 |

**Flow coverage:** health, provider discovery, chat completion, streaming,
embeddings, formatters, debate, monitoring, MCP protocol, RAG pipeline,
feature flags, full system end-to-end.

**Dependency graph:** health → providers → completion → streaming/debate,
health → monitoring, providers → embeddings → RAG

### Documentation

23 markdown files in `Challenges/docs/userflow/` covering all adapters,
architecture, challenge templates, container integration, evaluators,
framework comparison guide, and writing guides.

### Audit Fixes Applied

- 11 endpoint path mismatches fixed (matched to actual `router.go` routes)
- 4 compile-time interface checks added to production files
- 3 error wrapping issues fixed (WebSocket Send/Close, gRPC runGRPCurl)
- 1 unsafe nil return fixed (Maestro RunInstrumentedTests)
- `"userflow"` added to `knownCategories` in `cmd/helixagent/challenges.go`
- All 12 challenge constructors now tested with ID verification
- `CLAUDE.md` updated with `userflow_comprehensive_challenge.sh` listing

### Commits

**Challenges submodule (3 commits):**
```
f32b7e3 fix(userflow): update Maestro test to expect error from RunInstrumentedTests
aee9321 fix(userflow): add interface checks, error wrapping, and safety fixes
c783b34 feat(userflow): add 9 state-of-the-art testing framework adapters
```

**HelixAgent (3 commits):**
```
b015b0a5 chore: update Challenges submodule with Maestro test fix
e73e8515 fix(challenges): fix endpoint paths, expand tests, wire userflow category
b948688e feat(challenges): integrate Challenges module with comprehensive userflow testing
```

---

## What Still Needs To Be Done

### CRITICAL (must fix — blocks production use)

#### 1. Wire Go-Native Orchestrator Into Main Application

**Problem:** `internal/challenges/userflow/orchestrator.go` defines a fully
functional Go-native orchestrator with 12 registered challenges, but it is
**never imported or invoked** by any other part of the codebase. It is dead code.

**Where to fix:** `internal/challenges/orchestrator.go` — the main orchestrator's
`RegisterAll()` method. It currently only calls `RegisterShellChallengesEnhanced()`.
It needs to also import and invoke the userflow orchestrator to register the
12 Go-native challenges.

**Key files:**
- `internal/challenges/orchestrator.go` — main orchestrator (`RegisterAll()`)
- `internal/challenges/userflow/orchestrator.go` — userflow orchestrator
- `cmd/helixagent/challenges.go` — CLI entry point

**Approach:** In `RegisterAll()`, after registering shell challenges, create a
`userflow.NewOrchestrator(baseURL)` and register its challenges into the main
registry. The `baseURL` can come from config or default to `http://localhost:7061`.

#### 2. Add `"userflow_"` to `detectCategory()` Prefix List

**Problem:** The `detectCategory()` function in `internal/challenges/orchestrator.go`
does not have `"userflow_"` in its prefix list. The shell script
`userflow_comprehensive_challenge.sh` gets categorized as `"shell"` instead of
`"userflow"`. Running `--run-challenges=userflow` executes zero challenges.

**Where to fix:** `internal/challenges/orchestrator.go`, line ~224, the `prefixes`
slice. Add `"userflow_"` to the list.

**Current prefixes:**
```go
prefixes := []string{
    "provider_", "security_", "debate_", "cli_",
    "mcp_", "bigdata_", "memory_", "performance_",
    "grpc_", "release_", "speckit_", "subscription_",
    "verification_", "fallback_", "semantic_",
    "integration_", "full_system_", "constitution_",
    "challenge_module_",
}
```

### IMPORTANT (should be done)

#### 3. Missing Test Coverage — 3 Files With Zero Tests

| File | Size | What Needs Testing |
|------|------|--------------------|
| `Challenges/pkg/userflow/playwright_http_adapter.go` | 6,698 bytes | 19 exported methods (HTTP-based Playwright adapter) |
| `Challenges/pkg/userflow/options.go` | 1,207 bytes | 3 option functions + `resolveChallengeConfig()` |
| `Challenges/pkg/userflow/flow_ipc.go` | 670 bytes | `IPCCommand` struct and flow types |

#### 4. Missing Everyday Use Case Challenges

These real-world scenarios are NOT covered:

| Scenario | Priority | Notes |
|----------|----------|-------|
| Authentication/login flow (JWT) | HIGH | Token acquisition, refresh, auth-gated endpoints |
| Error handling scenarios | HIGH | Invalid models, bad JSON, auth failures, 4xx/5xx |
| Concurrent user load | HIGH | Parallel requests, race conditions |
| Multi-turn conversation | MEDIUM | Conversation continuity with context |
| Tool/function calling | MEDIUM | OpenAI-compatible tool use |
| Provider failover | MEDIUM | Primary down, fallback succeeds |
| WebSocket streaming challenge | MEDIUM | SSE/WS real-time data flow |
| gRPC service challenge | MEDIUM | gRPC endpoint testing via grpcurl |
| Rate limiting | LOW | Burst request testing |
| Pagination | LOW | Paginated endpoint testing |

#### 5. Missing Integration Tests

No integration tests exist in `tests/integration/` for the userflow system.
Need tests that spin up the server and execute userflow challenges against
real running infrastructure.

#### 6. Missing Benchmark Tests

Zero `func Benchmark*` functions in any userflow test file. Need benchmarks
for: HTTP API adapter request execution, flow step evaluation, assertion
evaluation, orchestrator registration.

#### 7. Missing Security Tests

No security tests for: input validation/sanitization, injection attacks via
flow step parameters, credential handling, WebSocket origin validation.

#### 8. Missing Stress Tests

No stress tests for: concurrent flow execution, adapter pool exhaustion,
WebSocket connection storms, registry contention under load.

#### 9. `Challenges/README.md` Has No Userflow Mention

Zero mentions of "userflow" or any of the new adapters in the main
`Challenges/README.md` file.

#### 10. Orchestrator `RunAll()` and `RunByID()` Untested

The `internal/challenges/userflow/flows_test.go` tests constructor, list,
summary, and count — but `RunAll()` and `RunByID()` are never tested.
These require a running server or mock HTTP server.

#### 11. Vendor Directory

Run `go mod vendor` to refresh the vendor directory after any changes.
Known issue: `go mod verify` shows `llm-verifier` ziphash missing.

### PRE-EXISTING (not caused by our work)

#### 12. Playwright CLI Adapter Test Failures

2 test failures in `Challenges/pkg/userflow/playwright_cli_adapter_test.go`:
- `TestPlaywrightCLIAdapter_Constructor`
- `TestPlaywrightCLIAdapter_ConfigVariants` (3 sub-tests)

Root cause: `cdpToHTTP()` converts `ws://` to `http://` but tests expect
unconverted WebSocket URLs. This is a test expectations bug, not a
production bug.

---

## Key File Locations

### Challenges Module (Submodule)
```
Challenges/pkg/userflow/selenium_adapter.go          # W3C WebDriver
Challenges/pkg/userflow/appium_adapter.go             # Appium 2.0
Challenges/pkg/userflow/cypress_adapter.go            # Cypress CLI
Challenges/pkg/userflow/puppeteer_adapter.go          # Puppeteer Node.js
Challenges/pkg/userflow/maestro_adapter.go            # Maestro YAML
Challenges/pkg/userflow/espresso_adapter.go           # Espresso Gradle+ADB
Challenges/pkg/userflow/robolectric_adapter.go        # Robolectric JVM
Challenges/pkg/userflow/adapter_grpc.go               # gRPC via grpcurl
Challenges/pkg/userflow/adapter_websocket_flow.go     # WebSocket gorilla
Challenges/pkg/userflow/challenge_grpc_flow.go        # gRPC challenge template
Challenges/pkg/userflow/challenge_websocket_flow.go   # WebSocket challenge template
Challenges/docs/userflow/                             # 23 documentation files
```

### HelixAgent
```
internal/challenges/userflow/flows.go                 # 12 API flow definitions
internal/challenges/userflow/orchestrator.go          # Go-native orchestrator
internal/challenges/userflow/flows_test.go            # 16 test functions
challenges/scripts/userflow_comprehensive_challenge.sh # Shell-based challenge
cmd/helixagent/challenges.go                          # CLI challenge runner
internal/challenges/orchestrator.go                   # Main orchestrator (needs wiring)
```

---

## Recommended Order of Work for Tomorrow

1. **Fix Critical #1:** Wire `userflow.NewOrchestrator()` into
   `internal/challenges/orchestrator.go:RegisterAll()`
2. **Fix Critical #2:** Add `"userflow_"` to `detectCategory()` prefixes
3. **Add authentication/login challenge** — most impactful everyday scenario
4. **Add error handling challenge** — validates robustness
5. **Add concurrent user challenge** — validates stability
6. **Write integration tests** in `tests/integration/userflow_test.go`
7. **Write benchmark tests** for adapter performance
8. **Write security tests** for input validation
9. **Write stress tests** for concurrency
10. **Add remaining missing test coverage** (playwright_http, options, flow_ipc)
11. **Update Challenges/README.md** with userflow section
12. **Fix pre-existing Playwright test expectations** (optional)

---

## Test Verification Commands

```bash
# HelixAgent userflow tests (should all pass)
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
GOMAXPROCS=2 go test -count=1 -short -p 1 -v ./internal/challenges/userflow/

# Challenges module new adapter tests (should all pass)
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent/Challenges
GOMAXPROCS=2 go test -count=1 -short -p 1 \
  -run "TestSelenium|TestAppium|TestCypress|TestPuppeteer|TestMaestro|TestRobolectric|TestEspresso|TestNewGRPC|TestGRPCCLI|TestGRPCOption|TestGorillaWebSocket|TestNewWebSocket|TestGRPCFlow|TestWebSocketFlow" \
  ./pkg/userflow/...

# Full compilation check
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
go vet ./internal/challenges/... ./cmd/helixagent/

# Shell challenge (requires running server)
./challenges/scripts/userflow_comprehensive_challenge.sh
```
