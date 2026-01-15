# HelixAgent Remediation Tracker
**Created**: 2026-01-15
**Based on**: COMPREHENSIVE_AUDIT_2026_01_15.md
**Status**: IN PROGRESS

---

## Quick Navigation

- [Critical Issues](#critical-issues)
- [High Priority Issues](#high-priority-issues)
- [Test Coverage Tasks](#test-coverage-tasks)
- [Documentation Updates](#documentation-updates)
- [Progress Summary](#progress-summary)

---

## Critical Issues

### CRIT-001: Memory Database QueryRow Not Implemented
- **Status**: [ ] NOT STARTED
- **File**: `internal/database/memory.go:98`
- **Impact**: Standalone mode fails for any row queries
- **Fix**: Implement proper QueryRow() method or document limitation
- **Tests Required**: Unit tests for memory database operations
- **Verification**: Run `make test-unit` + standalone mode test

### CRIT-002: Auth Endpoints Missing from Router
- **Status**: [ ] NOT STARTED
- **Files**:
  - `internal/router/router.go`
  - `docs/api/openapi.yaml`
- **Endpoints to Add**:
  - [ ] `/v1/auth/refresh` (POST)
  - [ ] `/v1/auth/logout` (POST)
  - [ ] `/v1/auth/me` (GET)
- **Tests Required**: Handler tests for each endpoint
- **Verification**: API integration tests

### CRIT-003: Streaming Endpoints Not Registered
- **Status**: [ ] NOT STARTED
- **File**: `internal/router/router.go`
- **Endpoints to Register**:
  - [ ] `/v1/completions/stream` → `h.CompletionsStream`
  - [ ] `/v1/chat/completions/stream` → `h.ChatCompletionsStream`
- **Tests Required**: Streaming integration tests
- **Verification**: Challenge script: `streaming_types_challenge.sh`

### CRIT-004: gRPC Service Methods Unimplemented
- **Status**: [ ] NOT STARTED
- **File**: `pkg/api/llm-facade_grpc.pb.go:244-277`
- **Options**:
  - [ ] Option A: Implement all 17 gRPC methods
  - [ ] Option B: Document as REST-only API
- **Tests Required**: gRPC integration tests (if Option A)
- **Verification**: gRPC client test

### CRIT-005: Grep Embedded Calls Returns Mock Response
- **Status**: [ ] NOT STARTED
- **File**: `internal/handlers/openai_compatible.go:5220`
- **Current**: Returns `"Search pattern registered: %s (grep not fully implemented)"`
- **Fix**: Implement actual grep functionality for embedded tool calls
- **Tests Required**: Tool execution tests
- **Verification**: `tool_execution_challenge.sh`

---

## High Priority Issues

### HIGH-001: Swallowed Errors in Protocol Manager
- **Status**: [ ] NOT STARTED
- **File**: `internal/services/unified_protocol_manager.go:344-353`
- **Issue**: Errors assigned to `_` and ignored
- **Fix**: Add proper error handling and logging
- **Tests Required**: Protocol manager tests with error injection
- **Verification**: Error handling tests

### HIGH-002: Redis Cache Clear Not Implemented
- **Status**: [ ] NOT STARTED
- **File**: `internal/services/model_metadata_redis_cache.go:84`
- **Current**: Logs "not fully implemented" and returns nil
- **Fix**: Implement FLUSHDB or key pattern deletion
- **Tests Required**: Cache invalidation tests
- **Verification**: Cache integration tests

### HIGH-003: Streaming Support Check
- **Status**: [ ] NOT STARTED
- **File**: `internal/streaming/types.go:104`
- **Issue**: Returns error if `http.Flusher` not implemented
- **Fix**: Add graceful fallback for non-streaming environments
- **Tests Required**: Streaming tests in test environments
- **Verification**: Integration tests

### HIGH-004: Plugin Interface Validation
- **Status**: [ ] NOT STARTED
- **File**: `internal/plugins/loader.go:37`
- **Issue**: Silently fails to load plugins without proper interface
- **Fix**: Add better error messages and logging
- **Tests Required**: Plugin loading tests with invalid plugins
- **Verification**: Plugin system tests

### HIGH-005: Provider Import Not Implemented
- **Status**: [ ] NOT STARTED
- **File**: `LLMsVerifier/llm-verifier/cmd/main.go:866-867`
- **Issue**: CLI provider import feature is stub
- **Fix**: Implement JSON import or remove feature from help
- **Tests Required**: CLI import tests
- **Verification**: LLMsVerifier CLI tests

### HIGH-006: Batch Verification Incomplete
- **Status**: [ ] NOT STARTED
- **File**: `LLMsVerifier/llm-verifier/cmd/main.go:1604`
- **Issue**: Prints "not yet fully implemented"
- **Fix**: Complete implementation or document as planned feature
- **Tests Required**: Batch verification tests
- **Verification**: LLMsVerifier CLI tests

---

## Test Coverage Tasks

### Package: database (12% → 80%)

| Test File | Functions to Test | Status |
|-----------|-------------------|--------|
| `db_test.go` | Connection, Pool, Health | [ ] |
| `background_task_repository_test.go` | Create, Update, Get, List, Delete | [ ] |
| `cognee_memory_repository_test.go` | All CRUD operations | [ ] |
| `memory_test.go` | QueryRow, Exec, Query | [ ] |

**Estimated Tests**: ~50
**Progress**: 0/50

### Package: tools (18% → 80%)

| Test File | Functions to Test | Status |
|-----------|-------------------|--------|
| `handler_test.go` | All 21 tool executors | [ ] |
| `schema_test.go` | Schema validation | [ ] |

**Estimated Tests**: ~40
**Progress**: 0/40

### Package: router (20% → 80%)

| Test File | Functions to Test | Status |
|-----------|-------------------|--------|
| `router_test.go` | All route registrations | [ ] |
| `middleware_integration_test.go` | Auth, CORS, Rate limit | [ ] |

**Estimated Tests**: ~30
**Progress**: 0/30

### Package: cache (46% → 80%)

| Test File | Functions to Test | Status |
|-----------|-------------------|--------|
| `tiered_cache_test.go` | L1/L2 sync, eviction | [ ] |
| `metrics_test.go` | Hit/miss tracking | [ ] |

**Estimated Tests**: ~20
**Progress**: 0/20

### Package: handlers (55% → 80%)

| Test File | Functions to Test | Status |
|-----------|-------------------|--------|
| `cognee_handler_test.go` | All 19 methods | [ ] |
| `monitoring_handler_test.go` | Metrics, Health | [ ] |
| `agent_handler_test.go` | Agent operations | [ ] |

**Estimated Tests**: ~30
**Progress**: 0/30

---

## Documentation Updates

### DOC-001: Update Provider Count
- **Status**: [ ] NOT STARTED
- **File**: `CLAUDE.md`
- **Change**: "18+ providers" → "10 providers"
- **Section**: Project Overview

### DOC-002: Document Groq Routing
- **Status**: [ ] NOT STARTED
- **File**: `CLAUDE.md` or `docs/providers/openrouter.md`
- **Add**: Note that Groq is available via OpenRouter, not standalone

### DOC-003: Add Ollama Deprecation
- **Status**: [ ] NOT STARTED
- **Files**:
  - `internal/llm/providers/ollama/ollama.go`
  - `CLAUDE.md`
- **Add**: `// Deprecated: Ollama has score 5.0, use as fallback only`

### DOC-004: Document Fallback Scoring
- **Status**: [ ] NOT STARTED
- **File**: `CLAUDE.md` or new `docs/verifier/FALLBACK_BEHAVIOR.md`
- **Add**: Document that system works without LLMsVerifier (fallback mode)

### DOC-005: Update OpenAPI Spec
- **Status**: [ ] NOT STARTED
- **File**: `docs/api/openapi.yaml`
- **Add**:
  - [ ] `/v1/monitoring/*` endpoints
  - [ ] `/v1/debates/team` endpoint
  - [ ] `/v1/tasks/queue/stats` endpoint

---

## Progress Summary

### Issues Fixed

| Priority | Total | Fixed | Remaining |
|----------|-------|-------|-----------|
| CRITICAL | 5 | 0 | 5 |
| HIGH | 6 | 0 | 6 |
| MEDIUM | 5 | 0 | 5 |
| LOW | 5 | 0 | 5 |
| **TOTAL** | **21** | **0** | **21** |

### Test Coverage Progress

| Target | Current Tests | Added | Total | Coverage |
|--------|---------------|-------|-------|----------|
| 170 | 0 | 0 | 0 | 67% |

### Documentation Updates

| Total | Completed | Remaining |
|-------|-----------|-----------|
| 5 | 0 | 5 |

---

## How to Use This Tracker

1. **Starting a Task**: Change `[ ] NOT STARTED` to `[~] IN PROGRESS`
2. **Completing a Task**: Change `[~] IN PROGRESS` to `[x] COMPLETED`
3. **Add Verification**: After each fix, run:
   ```bash
   make test-unit
   make test-integration
   ./challenges/scripts/run_all_challenges.sh
   ```
4. **Update Coverage**: After adding tests:
   ```bash
   go test -coverprofile=coverage.out ./internal/...
   go tool cover -func=coverage.out
   ```

---

## Session Checkpoints

### Checkpoint 1 (Start)
- **Date**: 2026-01-15
- **Status**: Audit complete, tracker created
- **Next**: Begin Phase 1 Critical Fixes

### Checkpoint 2
- **Date**: [TBD]
- **Status**: [TBD]
- **Next**: [TBD]

### Checkpoint 3
- **Date**: [TBD]
- **Status**: [TBD]
- **Next**: [TBD]

---

*Update this tracker after each work session to maintain continuity.*
