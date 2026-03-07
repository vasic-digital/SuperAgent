# Writing Challenges

This guide explains how to write new challenge scripts for HelixAgent, covering both shell script challenges and Go-native challenges.

---

## Shell Script Challenges

### Required Structure

Every shell challenge script must follow this structure:

```bash
#!/bin/bash
# <Challenge Name>
# Tests: ~<N> tests across <M> sections
# Validates: <brief description of what is validated>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    # Inline fallback if common.sh is unavailable
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="<Your Challenge Name>"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."
HELIXAGENT_URL="${HELIXAGENT_URL:-http://localhost:7061}"

# --- Test sections go here ---

# --- Summary ---
echo ""
echo "=========================================="
echo "Challenge Summary: $CHALLENGE_NAME"
echo "=========================================="
echo "Total Tests: $TOTAL"
echo "Passed: $PASSED"
echo "Failed: $FAILED"

if [ $FAILED -eq 0 ]; then
    echo "ALL TESTS PASSED!"
    exit 0
else
    echo "SOME TESTS FAILED!"
    exit 1
fi
```

### Naming Convention

Challenge scripts must be named `<descriptive_name>_challenge.sh` and placed in `challenges/scripts/`. The name should clearly indicate the subsystem being validated:

```
release_build_challenge.sh
fallback_mechanism_challenge.sh
debate_orchestrator_challenge.sh
helixmemory_challenge.sh
```

### Pass/Fail Helper Functions

There are two common patterns for recording test results. Use whichever matches the existing challenges in the area you are working on.

**Pattern A: Simple pass/fail functions (used by release_build_challenge.sh):**

```bash
pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}[FAIL]${NC} $1"
}
```

**Pattern B: Inline counter updates with log functions (used by fallback_mechanism_challenge.sh):**

```bash
TOTAL=$((TOTAL + 1))
log_info "Test 1: FallbackConfig type defined"
if grep -q "type FallbackConfig struct" "$PROJECT_ROOT/internal/services/types.go" 2>/dev/null; then
    log_success "FallbackConfig type defined"
    PASSED=$((PASSED + 1))
else
    log_error "FallbackConfig type NOT defined!"
    FAILED=$((FAILED + 1))
fi
```

### Organizing Tests into Sections

Group related tests into clearly labeled sections:

```bash
# ============================================================================
# Section 1: Code-Level Implementation
# ============================================================================

log_info "Section 1: Code-Level Implementation"

# Test 1: ...
# Test 2: ...

# ============================================================================
# Section 2: Unit Test Coverage
# ============================================================================

log_info "Section 2: Unit Test Coverage"

# Test 3: ...
# Test 4: ...
```

### Types of Assertions

Use these assertion patterns to validate actual behavior:

**1. File existence:**
```bash
if [ -f "$PROJECT_ROOT/internal/services/my_service.go" ]; then
    pass "my_service.go exists"
else
    fail "my_service.go missing"
fi
```

**2. Code structure (grep for types, functions, imports):**
```bash
if grep -q "type MyService struct" "$PROJECT_ROOT/internal/services/my_service.go" 2>/dev/null; then
    pass "MyService type defined"
else
    fail "MyService type NOT defined"
fi
```

**3. Compilation check:**
```bash
if (cd "$PROJECT_ROOT" && go build ./internal/services/... >/dev/null 2>&1); then
    pass "Package compiles"
else
    fail "Package does not compile"
fi
```

**4. Test execution:**
```bash
if (cd "$PROJECT_ROOT" && go test ./internal/services/... -short >/dev/null 2>&1); then
    pass "Tests pass"
else
    fail "Tests fail"
fi
```

**5. HTTP endpoint validation (requires running server):**
```bash
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$HELIXAGENT_URL/v1/health" 2>/dev/null)
if [ "$RESPONSE" = "200" ]; then
    pass "Health endpoint returns 200"
else
    fail "Health endpoint returned $RESPONSE"
fi
```

**6. Response content validation:**
```bash
BODY=$(curl -s "$HELIXAGENT_URL/v1/models" 2>/dev/null)
if echo "$BODY" | grep -q '"data"'; then
    pass "Models endpoint returns data array"
else
    fail "Models endpoint missing data array"
fi
```

### Summary Section

Every challenge must end with a summary that prints totals and exits with the correct code:

```bash
echo ""
echo "=========================================="
echo "Challenge Summary: $CHALLENGE_NAME"
echo "=========================================="
echo "Total Tests: $TOTAL"
echo "Passed: $PASSED"
if [ $FAILED -gt 0 ]; then
    echo "Failed: $FAILED"
    exit 1
else
    echo "Failed: 0"
    echo "ALL TESTS PASSED!"
    exit 0
fi
```

The exit code is critical: `0` means all tests passed, `1` means at least one failed. The `run_all_challenges.sh` runner uses exit codes to determine overall success.

---

## Go-Native Challenges

### Structure

Go-native challenges use the `digital.vasic.challenges` module framework. They live in `internal/challenges/` and implement the challenge interfaces.

### Writing an API User Flow

User flow challenges define step-by-step API interactions in `internal/challenges/userflow/flows.go`:

```go
func MyFeatureFlow(token string) uf.APIFlow {
    return uf.APIFlow{
        Steps: []uf.APIStep{
            {
                Name:           "step_one",
                Method:         "GET",
                Path:           "/v1/my-endpoint",
                ExpectedStatus: 200,
                Assertions: []uf.StepAssertion{
                    {Type: "not_empty", Target: "body"},
                    {Type: "response_contains", Target: "field", Value: "expected"},
                },
            },
            {
                Name:           "step_two",
                Method:         "POST",
                Path:           "/v1/my-endpoint",
                Body:           `{"key": "value"}`,
                ExpectedStatus: 201,
                Headers: map[string]string{
                    "Authorization": "Bearer " + token,
                    "Content-Type":  "application/json",
                },
            },
        },
    }
}
```

### Registering a New Challenge

Register your challenge in the orchestrator setup so it is discovered and executed:

```go
// In internal/challenges/orchestrator.go or a new file
challenge := challenge.New(
    challenge.ID("my-feature-check"),
    "My Feature Check",
    "Validates my feature works correctly",
    "integration",
    []challenge.ID{},  // dependencies (other challenge IDs that must pass first)
    myCheckFunc,
)
reg.Register(challenge)
```

### Assertion Types for API Flows

| Type                | Description                                    |
|---------------------|------------------------------------------------|
| `not_empty`         | Response body is not empty                     |
| `response_contains` | Body contains the specified value              |
| `status_code`       | HTTP status code matches                       |
| `json_path`         | JSON field at path matches expected value      |

---

## Best Practices

### 1. No False Positives

The most important rule: a passing challenge must mean the capability genuinely works. Never write a test that trivially passes.

Bad (always passes):
```bash
TOTAL=$((TOTAL + 1))
if true; then
    PASSED=$((PASSED + 1))
fi
```

Bad (checks return code but not content):
```bash
curl -s "$HELIXAGENT_URL/v1/endpoint" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    pass "Endpoint works"  # curl succeeds even on 404/500
fi
```

Good (validates actual response):
```bash
RESPONSE=$(curl -s -w "\n%{http_code}" "$HELIXAGENT_URL/v1/endpoint" 2>/dev/null)
HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)
if [ "$HTTP_CODE" = "200" ] && echo "$BODY" | grep -q '"status":"healthy"'; then
    pass "Endpoint returns healthy status"
else
    fail "Endpoint did not return healthy status (HTTP $HTTP_CODE)"
fi
```

### 2. Validate Actual Behavior, Not Return Codes

Grep-only validation is never sufficient for runtime checks. When validating that code works at runtime, make actual HTTP requests and inspect response bodies.

### 3. Use Specific Grep Patterns

When checking code structure, use patterns specific enough to avoid false matches:

Bad:
```bash
grep -q "Fallback" "$FILE"  # Matches comments, variable names, anything
```

Good:
```bash
grep -q "type FallbackConfig struct" "$FILE"  # Matches only the type definition
```

### 4. Suppress Stderr When Appropriate

Use `2>/dev/null` on grep and curl commands to keep output clean:

```bash
if grep -q "pattern" "$FILE" 2>/dev/null; then
```

### 5. Always Provide Meaningful Failure Messages

The failure message should tell the developer exactly what is wrong and where to look:

Bad:
```bash
fail "Test failed"
```

Good:
```bash
fail "FallbackConfig type NOT defined in internal/services/debate_support_types.go"
```

### 6. Handle Missing Files Gracefully

Files might not exist. Always guard with `2>/dev/null` or explicit existence checks:

```bash
if [ -f "$FILE" ] && grep -q "pattern" "$FILE"; then
    pass "Pattern found"
else
    fail "Pattern not found (file may not exist: $FILE)"
fi
```

### 7. Number Your Tests

Number tests sequentially within each section for easy reference in failure reports:

```bash
# Test 1: ...
# Test 2: ...
# ...
# Test 14: ...
```

### 8. Document Test Count in the Header

The script header comment should list the approximate number of tests so reviewers know the scope:

```bash
# Tests: ~25 tests across 6 sections
```

### 9. Respect Resource Limits

Challenges must not exceed 30-40% of host resources. For `go test` invocations inside challenges:

```bash
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -p 1 -short ./internal/services/... >/dev/null 2>&1
```

### 10. Make Challenges Idempotent

Challenges should be runnable multiple times in succession without side effects. Do not create files, modify databases, or change state that would cause the next run to behave differently.

---

## Shell Challenge Template

Use this as a starting point for new shell challenges:

```bash
#!/bin/bash
# <Feature Name> Challenge
# Tests: ~<N> tests across <M> sections
# Validates: <what this challenge validates>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="<Feature Name> Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."
HELIXAGENT_URL="${HELIXAGENT_URL:-http://localhost:7061}"

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="

# ============================================================================
# Section 1: Code Structure
# ============================================================================

log_info ""
log_info "Section 1: Code Structure"
log_info ""

# Test 1: Required type exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: <TypeName> type defined"
if grep -q "type <TypeName> struct" "$PROJECT_ROOT/internal/<package>/<file>.go" 2>/dev/null; then
    log_success "<TypeName> type defined"
    PASSED=$((PASSED + 1))
else
    log_error "<TypeName> type NOT defined"
    FAILED=$((FAILED + 1))
fi

# Test 2: Required function exists
TOTAL=$((TOTAL + 1))
log_info "Test 2: <FunctionName> function exists"
if grep -q "func.*<FunctionName>" "$PROJECT_ROOT/internal/<package>/<file>.go" 2>/dev/null; then
    log_success "<FunctionName> function exists"
    PASSED=$((PASSED + 1))
else
    log_error "<FunctionName> function NOT found"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Test Coverage
# ============================================================================

log_info ""
log_info "Section 2: Test Coverage"
log_info ""

# Test 3: Tests exist and pass
TOTAL=$((TOTAL + 1))
log_info "Test 3: Package tests pass"
if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 go test -p 1 -short ./internal/<package>/... >/dev/null 2>&1); then
    log_success "Package tests pass"
    PASSED=$((PASSED + 1))
else
    log_error "Package tests FAIL"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Summary
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Challenge Summary: $CHALLENGE_NAME"
log_info "=============================================="
log_info "Total Tests: $TOTAL"
log_success "Passed: $PASSED"
if [ $FAILED -gt 0 ]; then
    log_error "Failed: $FAILED"
fi

if [ $FAILED -eq 0 ]; then
    log_success "ALL TESTS PASSED!"
    exit 0
else
    log_error "SOME TESTS FAILED!"
    exit 1
fi
```

## Go-Native Challenge Template

Use this as a starting point for new Go-native user flow challenges:

```go
// In internal/challenges/userflow/flows.go

// MyFeatureFlow returns a flow that validates the <feature>
// endpoints respond correctly with expected data.
func MyFeatureFlow(token string) uf.APIFlow {
    return uf.APIFlow{
        Steps: []uf.APIStep{
            {
                Name:           "feature_list",
                Method:         "GET",
                Path:           "/v1/my-feature",
                ExpectedStatus: 200,
                Headers: map[string]string{
                    "Authorization": "Bearer " + token,
                },
                Assertions: []uf.StepAssertion{
                    {Type: "not_empty", Target: "body"},
                },
            },
            {
                Name:           "feature_create",
                Method:         "POST",
                Path:           "/v1/my-feature",
                Body:           `{"name": "test-item"}`,
                ExpectedStatus: 201,
                Headers: map[string]string{
                    "Authorization": "Bearer " + token,
                    "Content-Type":  "application/json",
                },
                Assertions: []uf.StepAssertion{
                    {Type: "response_contains", Target: "name", Value: "test-item"},
                },
            },
        },
    }
}
```

Register the flow in the orchestrator by adding it to the challenge list, then run with `--run-challenges=userflow`.
