# SP1: Critical Fixes & Dead Code Elimination — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every line of production code compilable, reachable, and connected — zero compiler errors, zero nil-panics, zero dead code.

**Architecture:** Fix a duplicate method declaration that prevents compilation, move misplaced route registrations out of request handlers, gate nil-service handlers behind availability checks, and remove or wire 5 dead adapter packages plus 2 dead directories.

**Tech Stack:** Go 1.25.3, Gin v1.12.0, testify v1.11.1

**Spec:** `docs/superpowers/specs/2026-03-25-comprehensive-completion-design.md` (SP1 section)

---

### Task 1: Fix Duplicate GetAgentPool() Method (Compiler Error)

**Files:**
- Modify: `internal/debate/comprehensive/integration.go:403-406`

- [ ] **Step 1: Verify the duplicate**

Run: `grep -n 'func (m \*IntegrationManager) GetAgentPool' internal/debate/comprehensive/integration.go`
Expected: Two matches at lines 197 and 404

- [ ] **Step 2: Remove the duplicate declaration (lines 403-406)**

Delete the second `GetAgentPool()` at line 403-406. Keep the first one at line 197.

```go
// DELETE these lines (403-406):
// // GetAgentPool returns the agent pool
// func (m *IntegrationManager) GetAgentPool() *AgentPool {
// 	return m.pool
// }
```

- [ ] **Step 3: Verify compilation**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && go build ./internal/debate/comprehensive/`
Expected: Clean build, no errors

- [ ] **Step 4: Run existing tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && go test ./internal/debate/comprehensive/ -short -count=1`
Expected: All tests pass

- [ ] **Step 5: Commit**

```bash
git add internal/debate/comprehensive/integration.go
git commit -m "fix(debate): remove duplicate GetAgentPool method declaration"
```

---

### Task 2: Move Skills Routes Out of Request Handler Closure

**Files:**
- Modify: `internal/router/router.go:774-792`

The skills route registration (`/v1/skills/*`) is currently inside a per-request handler closure for `GET /providers/:id/health`. Every health check request for a healthy provider with a circuit breaker re-registers these routes. This must be moved outside the closure.

- [ ] **Step 1: Read the surrounding context**

Read `internal/router/router.go` lines 750-800 to understand the closure structure. The skills registration block (lines 781-791) is nested inside:
1. A `GET("/:id/health", func(c *gin.Context) {` handler
2. An `if cb := providerRegistry.GetCircuitBreaker(name); cb != nil {` block
3. An `else { response["healthy"] = true` block

- [ ] **Step 2: Cut the skills block from inside the closure**

Remove lines 781-791 from inside the handler closure:

```go
// REMOVE from inside the closure (lines 781-791):
// 							// Skills endpoints
// 							skillsHandler := handlers.NewSkillsHandler(skillsIntegration)
// 							skillsHandler.SetLogger(logger)
// 							skillsGroup := protected.Group("/skills")
// 							{
// 								skillsGroup.GET("", skillsHandler.ListSkills)
// 								skillsGroup.GET("/categories", skillsHandler.ListCategories)
// 								skillsGroup.GET("/:category", skillsHandler.GetSkillsByCategory)
// 								skillsGroup.POST("/match", skillsHandler.MatchSkills)
// 							}
// 							logger.Info("Skills endpoints registered at /v1/skills/*")
```

- [ ] **Step 3: Add skills registration at the top level (after other endpoint groups)**

Find where other endpoint groups are registered (near line 1114, after `protocolSSEHandler.RegisterSSERoutes(protected)`) and add:

```go
// Skills endpoints (skill registry integration)
if skillsIntegration != nil {
    skillsHandler := handlers.NewSkillsHandler(skillsIntegration)
    skillsHandler.SetLogger(logger)
    skillsGroup := protected.Group("/skills")
    {
        skillsGroup.GET("", skillsHandler.ListSkills)
        skillsGroup.GET("/categories", skillsHandler.ListCategories)
        skillsGroup.GET("/:category", skillsHandler.GetSkillsByCategory)
        skillsGroup.POST("/match", skillsHandler.MatchSkills)
    }
    logger.Info("Skills endpoints registered at /v1/skills/*")
}
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./internal/router/`
Expected: Clean build

- [ ] **Step 5: Run router tests**

Run: `go test ./internal/router/ -short -count=1`
Expected: All pass

- [ ] **Step 6: Commit**

```bash
git add internal/router/router.go
git commit -m "fix(router): move skills route registration out of per-request handler closure"
```

---

### Task 3: Gate Nil-Service Handlers Behind Availability Checks

**Files:**
- Modify: `internal/router/router.go:1117-1204`

7 handlers are constructed with nil services. The handlers already return 503 when services are nil (graceful degradation). In SP1, we document this pattern. **Actual lazy service wiring is deferred to SP4 Task 1** (LazyServiceProvider), which will initialize services on first request. SP1 ensures the current behavior is safe and documented; SP4 makes the endpoints functional.

- [ ] **Step 1: Wrap BackgroundTaskHandler registration (line 1117-1121)**

The handler receives 11 params (10 nil + logger). Keep the current pattern — the handler already returns 503 gracefully. Add a comment documenting the nil-guard behavior:

```go
// Background Task endpoints — services wired lazily; handler returns 503 if services unavailable
backgroundTaskHandler := handlers.NewBackgroundTaskHandler(
    nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, logger,
)
backgroundTaskHandler.RegisterRoutes(protected)
logger.Info("Background task endpoints registered at /v1/tasks/* (services pending)")
```

- [ ] **Step 2: Add nil-guard comments to remaining 6 handlers**

For DiscoveryHandler (line 1124), ScoringHandler (line 1137), VerificationHandler (line 1153), HealthHandler (line 1168), LLMOpsHandler (line 1197), BenchmarkHandler (line 1202) — add a `// services pending` comment to each log line. Verify each handler's constructor accepts nil without panic.

- [ ] **Step 3: Verify compilation and tests**

Run: `go build ./internal/router/ && go test ./internal/router/ -short -count=1`
Expected: Clean build and all tests pass

- [ ] **Step 4: Commit**

```bash
git add internal/router/router.go
git commit -m "docs(router): document nil-guard pattern for lazily-wired handler services"
```

---

### Task 4: Add OAuth Manager Nil-Safety

**Files:**
- Modify: `internal/router/router.go:380-389`

- [ ] **Step 1: Read current initialization pattern**

Line 380-389: `oauthManager` is declared, conditionally created. Line 969: assigned to `rc.oauthCredentialManager`. The Warn log at line 386 already handles the error case, and `oauthManager` will be nil if creation fails. Check that all code using `rc.oauthCredentialManager` handles nil.

- [ ] **Step 2: Search for usages of oauthCredentialManager**

Run: `grep -rn 'oauthCredentialManager' internal/router/`
Verify every usage has a nil check.

- [ ] **Step 3: Add nil check if missing**

If any usage lacks a nil check, add one:

```go
if rc.oauthCredentialManager != nil {
    // use oauthCredentialManager
}
```

- [ ] **Step 4: Verify and commit**

Run: `go build ./internal/router/ && go test ./internal/router/ -short -count=1`

```bash
git add internal/router/router.go
git commit -m "fix(router): add nil-safety guard for OAuth credential manager"
```

---

### Task 5: Remove Dead Code — background/backup/ Directory

**Files:**
- Remove: `internal/background/backup/` (21 files, 364KB — full package duplication)

- [ ] **Step 1: Verify it's a duplicate**

Run: `diff <(ls internal/background/backup/ | sort) <(ls internal/background/*.go | xargs -I{} basename {} | sort)`
Expected: Significant overlap. The backup directory is a stale copy.

- [ ] **Step 2: Verify no imports reference it**

Run: `grep -rn 'internal/background/backup' internal/ --include='*.go'`
Expected: Zero matches

- [ ] **Step 3: Remove the directory**

```bash
rm -rf internal/background/backup/
```

- [ ] **Step 4: Verify build**

Run: `go build ./internal/background/`
Expected: Clean

- [ ] **Step 5: Commit**

```bash
git add -A internal/background/backup/
git commit -m "chore(background): remove stale backup directory (duplicate package, 364KB)"
```

---

### Task 6: Remove Dead Code — challenges/userflow/results/ Directory

**Files:**
- Remove or gitignore: `internal/challenges/userflow/results/` (empty output directory)

- [ ] **Step 1: Verify it contains no Go files**

Run: `find internal/challenges/userflow/results/ -name '*.go' | wc -l`
Expected: 0

- [ ] **Step 2: Add to .gitignore**

Append to `.gitignore`:
```
internal/challenges/userflow/results/
```

- [ ] **Step 3: Remove tracked empty dirs**

```bash
git rm -r --cached internal/challenges/userflow/results/ 2>/dev/null || true
```

- [ ] **Step 4: Commit**

```bash
git add .gitignore
git add -A internal/challenges/userflow/results/
git commit -m "chore(challenges): gitignore userflow results output directory"
```

---

### Task 7: Audit Dead Adapter Packages

**Files:**
- Audit: `internal/adapters/background/`, `internal/adapters/observability/`, `internal/adapters/events/`, `internal/adapters/http/`, `internal/adapters/helixqa/`

Each has an `adapter.go` + `adapter_test.go`. They were written to bridge internal types to extracted modules but never wired in.

- [ ] **Step 1: Check each adapter's purpose**

For each of the 5 packages, read `adapter.go` to understand what it bridges. Determine if the corresponding extracted module is actively used.

- [ ] **Step 2: Search for any imports**

Run for each:
```bash
grep -rn '"dev.helix.agent/internal/adapters/background"' internal/ --include='*.go'
grep -rn '"dev.helix.agent/internal/adapters/observability"' internal/ --include='*.go'
grep -rn '"dev.helix.agent/internal/adapters/events"' internal/ --include='*.go'
grep -rn '"dev.helix.agent/internal/adapters/http"' internal/ --include='*.go'
grep -rn '"dev.helix.agent/internal/adapters/helixqa"' internal/ --include='*.go'
```
Expected: Zero matches for each

- [ ] **Step 3: Decision per adapter**

For each adapter, choose one of:
- **Wire it in**: If the extracted module is actively used and the adapter provides value, import it in the appropriate service/cmd file
- **Remove it**: If the extracted module is not used or the adapter is redundant

Document the decision for each.

- [ ] **Step 4: Apply changes**

Remove dead adapters or wire them in based on Step 3 decisions.

- [ ] **Step 5: Verify build**

Run: `go build ./...`
Expected: Clean

- [ ] **Step 6: Commit**

```bash
git add -A internal/adapters/
git commit -m "chore(adapters): remove unused adapter packages (background, observability, events, http, helixqa)"
```

---

### Task 8: Create dead_code_verification_challenge.sh

**Files:**
- Create: `challenges/scripts/dead_code_verification_challenge.sh`

- [ ] **Step 1: Write the challenge script**

```bash
#!/bin/bash
# Dead Code Verification Challenge
# Validates that no dead code packages exist in the codebase

set -euo pipefail

PASS=0
FAIL=0
TOTAL=0

check() {
    local desc="$1"
    local result="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$result" = "0" ]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Dead Code Verification Challenge ==="
echo ""

# 1. No duplicate method declarations
echo "--- Compiler Integrity ---"
DUP=$(go vet ./internal/debate/comprehensive/ 2>&1 | grep -c "already declared" || true)
check "No duplicate method declarations in debate/comprehensive" "$DUP"

# 2. No dead backup directories
echo "--- Dead Directories ---"
BACKUP=$(test -d internal/background/backup && echo 1 || echo 0)
check "No internal/background/backup/ directory" "$BACKUP"

# 3. No unimported adapter packages
echo "--- Dead Adapter Packages ---"
for pkg in background observability events http helixqa; do
    if [ -d "internal/adapters/$pkg" ]; then
        IMPORTS=$(grep -rn "\"dev.helix.agent/internal/adapters/$pkg\"" internal/ --include='*.go' 2>/dev/null | grep -v '_test.go' | wc -l)
        if [ "$IMPORTS" = "0" ]; then
            check "internal/adapters/$pkg/ is imported or removed" "1"
        else
            check "internal/adapters/$pkg/ is imported or removed" "0"
        fi
    else
        check "internal/adapters/$pkg/ is imported or removed" "0"
    fi
done

# 4. Skills routes not inside handler closure
echo "--- Route Registration ---"
SKILLS_IN_HANDLER=$(grep -A2 'response\["healthy"\] = true' internal/router/router.go | grep -c "Skills" || true)
check "Skills routes not inside health handler closure" "$SKILLS_IN_HANDLER"

# 5. Build succeeds
echo "--- Build Verification ---"
if go build ./... 2>/dev/null; then
    check "Full project builds cleanly" "0"
else
    check "Full project builds cleanly" "1"
fi

# 6. go vet passes
if go vet ./internal/... 2>/dev/null; then
    check "go vet passes on internal/" "0"
else
    check "go vet passes on internal/" "1"
fi

echo ""
echo "=== Results: $PASS/$TOTAL passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
```

- [ ] **Step 2: Make executable**

```bash
chmod +x challenges/scripts/dead_code_verification_challenge.sh
```

- [ ] **Step 3: Commit**

```bash
git add challenges/scripts/dead_code_verification_challenge.sh
git commit -m "test(challenges): add dead code verification challenge"
```

---

### Task 9: Final SP1 Validation

- [ ] **Step 1: Full build**

Run: `go build ./...`
Expected: Zero errors

- [ ] **Step 2: Full vet**

Run: `go vet ./internal/...`
Expected: Zero warnings

- [ ] **Step 3: Unit tests**

Run: `GOMAXPROCS=2 nice -n 19 go test ./internal/... -short -count=1 -p 1`
Expected: All pass

- [ ] **Step 4: Run dead code challenge**

Run: `./challenges/scripts/dead_code_verification_challenge.sh`
Expected: All checks pass

- [ ] **Step 5: Tag completion**

```bash
git tag sp1-complete
```
