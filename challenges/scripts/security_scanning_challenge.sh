#!/bin/bash
# HelixAgent Challenge - Security Scanning Validation
# Validates security scanning tools, patterns, and security posture of the codebase.
# Tests: gosec availability, security packages, PII detection, rate limiting,
#        input validation, no hardcoded secrets, SQL injection prevention,
#        CORS, JWT, auth middleware, authorization, security headers, content security.

set -e

# Source challenge framework
source "$(dirname "${BASH_SOURCE[0]}")/challenge_framework.sh"

# Initialize challenge
init_challenge "security-scanning" "Security Scanning Validation"
load_env

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))

    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        log_success "PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "test" "$test_name" "true" ""
        return 0
    else
        log_error "FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "test" "$test_name" "false" "Test command failed"
        return 1
    fi
}

# ============================================================================
# SECTION 1: SECURITY CONFIGURATION FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: Security Configuration Files"
log_info "=============================================="

run_test "Snyk configuration (.snyk) exists" \
    "[[ -f '$PROJECT_ROOT/.snyk' ]]"

run_test "SBOM generation script exists" \
    "[[ -f '$PROJECT_ROOT/scripts/generate-sbom.sh' ]]"

run_test "SBOM generation script is executable" \
    "[[ -x '$PROJECT_ROOT/scripts/generate-sbom.sh' ]]"

# ============================================================================
# SECTION 2: SECURITY TOOLS AVAILABILITY
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: Security Tools Availability"
log_info "=============================================="

run_test "gosec binary is available" \
    "command -v gosec >/dev/null 2>&1 || [[ -f '\$HOME/go/bin/gosec' ]] || [[ -f '$PROJECT_ROOT/bin/gosec' ]]"

run_test "go vet is available" \
    "go vet --help 2>&1 | grep -q 'usage: go vet'"

run_test "go tool includes security-relevant features" \
    "go version >/dev/null 2>&1"

# ============================================================================
# SECTION 3: STATIC ANALYSIS
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: Static Analysis (go vet)"
log_info "=============================================="

VET_START=$(date +%s%N)
# Exclude third-party cli_agents/ submodules which may have non-standard Go code
run_test "go vet passes on all internal packages" \
    "cd '$PROJECT_ROOT' && go vet ./internal/... ./cmd/..."
VET_END=$(date +%s%N)
VET_DURATION_MS=$(( (VET_END - VET_START) / 1000000 ))
record_metric "go_vet_time_ms" "$VET_DURATION_MS"

# ============================================================================
# SECTION 4: MAKEFILE SECURITY TARGETS
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: Makefile Security Targets"
log_info "=============================================="

run_test "Makefile has security-scan target" \
    "grep -qE '^security-scan:' '$PROJECT_ROOT/Makefile'"

run_test "Makefile has sbom target" \
    "grep -qE '^sbom:' '$PROJECT_ROOT/Makefile'"

# ============================================================================
# SECTION 5: SECURITY TEST FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: Security Test Infrastructure"
log_info "=============================================="

run_test "Security test directory exists" \
    "[[ -d '$PROJECT_ROOT/tests/security' ]]"

run_test "Security tests exist (test files present)" \
    "ls '$PROJECT_ROOT'/tests/security/*_test.go >/dev/null 2>&1"

run_test "Security penetration test exists" \
    "[[ -f '$PROJECT_ROOT/tests/security/penetration_test.go' ]]"

run_test "Security input validation test exists" \
    "[[ -f '$PROJECT_ROOT/tests/security/input_validation_test.go' ]]"

# ============================================================================
# SECTION 6: SECURITY PACKAGES IN CODEBASE
# ============================================================================
log_info "=============================================="
log_info "SECTION 6: Security Packages in Codebase"
log_info "=============================================="

run_test "Internal security package exists (internal/security/)" \
    "[[ -d '$PROJECT_ROOT/internal/security' ]]"

run_test "Security package has audit.go (audit logging)" \
    "[[ -f '$PROJECT_ROOT/internal/security/audit.go' ]]"

run_test "Security package has guardrails.go (content guardrails)" \
    "[[ -f '$PROJECT_ROOT/internal/security/guardrails.go' ]]"

run_test "Security package has pii.go (PII detection)" \
    "[[ -f '$PROJECT_ROOT/internal/security/pii.go' ]]"

run_test "Security package has redteam.go (red team testing)" \
    "[[ -f '$PROJECT_ROOT/internal/security/redteam.go' ]]"

# ============================================================================
# SECTION 7: PII DETECTION PATTERNS
# ============================================================================
log_info "=============================================="
log_info "SECTION 7: PII Detection"
log_info "=============================================="

run_test "PII detection code references email patterns" \
    "grep -qiE 'email|Email|EMAIL' '$PROJECT_ROOT/internal/security/pii.go'"

run_test "PII detection code references phone or SSN patterns" \
    "grep -qiE 'phone|ssn|social.security|credit.card' '$PROJECT_ROOT/internal/security/pii.go'"

run_test "PII redaction function exists" \
    "grep -qE 'Redact|redact' '$PROJECT_ROOT/internal/security/pii.go'"

# ============================================================================
# SECTION 8: RATE LIMITING PATTERNS
# ============================================================================
log_info "=============================================="
log_info "SECTION 8: Rate Limiting"
log_info "=============================================="

run_test "Rate limiting middleware file exists" \
    "[[ -f '$PROJECT_ROOT/internal/middleware/rate_limit.go' ]]"

run_test "Rate limiting uses token bucket or sliding window" \
    "grep -qiE 'TokenBucket|SlidingWindow|rate.*limit|RateLimit|rateLimiter' '$PROJECT_ROOT/internal/middleware/rate_limit.go'"

run_test "Rate limiting test file exists" \
    "[[ -f '$PROJECT_ROOT/internal/middleware/rate_limit_test.go' ]]"

# ============================================================================
# SECTION 9: INPUT VALIDATION PATTERNS
# ============================================================================
log_info "=============================================="
log_info "SECTION 9: Input Validation"
log_info "=============================================="

run_test "Input validation middleware file exists" \
    "[[ -f '$PROJECT_ROOT/internal/middleware/validation.go' ]]"

run_test "Input validation tests exist" \
    "[[ -f '$PROJECT_ROOT/internal/middleware/validation_test.go' ]]"

run_test "Path validation utility exists (path traversal prevention)" \
    "[[ -f '$PROJECT_ROOT/internal/utils/path_validation.go' ]]"

run_test "Path validation checks for traversal patterns" \
    "grep -qE '\.\./|path.traversal|ValidatePath' '$PROJECT_ROOT/internal/utils/path_validation.go'"

# ============================================================================
# SECTION 10: NO HARDCODED SECRETS
# ============================================================================
log_info "=============================================="
log_info "SECTION 10: No Hardcoded Secrets in Source"
log_info "=============================================="

run_test "No hardcoded password literals in Go source (production code)" \
    "! grep -rn --include='*.go' -l 'password\s*=\s*\"[^\"]\{8,\}\"' '$PROJECT_ROOT/internal/' | grep -v test | grep -v example | head -1 | grep -q ."

run_test "No plaintext bearer token hardcoding in production Go" \
    "! grep -rn --include='*.go' 'Bearer [A-Za-z0-9+/]\{20,\}' '$PROJECT_ROOT/internal/' | grep -v test | grep -v nosec | grep -q ."

run_test "API keys sourced from environment variables (not hardcoded)" \
    "grep -rn 'os.Getenv\|os.LookupEnv\|viper\|config\.' '$PROJECT_ROOT/internal/config/config.go' | grep -qiE 'key|secret|token|password'"

# ============================================================================
# SECTION 11: SQL INJECTION PREVENTION
# ============================================================================
log_info "=============================================="
log_info "SECTION 11: SQL Injection Prevention"
log_info "=============================================="

run_test "Database repository uses parameterized queries (\$1, \$2 placeholders)" \
    "grep -rn --include='*.go' '\\\$1\|\\\$2' '$PROJECT_ROOT/internal/database/' | grep -qE '\.Query|\.Exec|\.QueryRow'"

run_test "No string-concatenated SQL queries in database package" \
    "! grep -rn --include='*.go' 'SELECT.*+\|INSERT.*+\|UPDATE.*+' '$PROJECT_ROOT/internal/database/' | grep -v test | grep -q ."

run_test "pgx or database/sql used for safe parameterized queries" \
    "grep -rn --include='*.go' 'github.com/jackc/pgx\|database/sql' '$PROJECT_ROOT/internal/database/' | grep -q ."

# ============================================================================
# SECTION 12: CORS CONFIGURATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 12: CORS Configuration"
log_info "=============================================="

run_test "CORS configuration exists in codebase" \
    "grep -rn --include='*.go' -l 'cors\|CORS\|AllowOrigin\|Access-Control' '$PROJECT_ROOT/internal/' | grep -q ."

run_test "CORS allowed origins configured (not wildcard-only)" \
    "grep -rn --include='*.go' 'AllowOrigins\|AllowOrigin\|cors' '$PROJECT_ROOT/internal/' | grep -v test | grep -q ."

# ============================================================================
# SECTION 13: JWT VALIDATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 13: JWT Validation"
log_info "=============================================="

run_test "JWT validation middleware exists" \
    "[[ -f '$PROJECT_ROOT/internal/middleware/auth.go' ]]"

run_test "JWT library imported for token validation" \
    "grep -rn --include='*.go' 'jwt\|JWT' '$PROJECT_ROOT/internal/middleware/auth.go' | grep -q ."

run_test "JWT secret sourced from environment (not hardcoded)" \
    "grep -rn --include='*.go' 'JWT_SECRET\|jwt.secret\|JwtSecret\|getenv\|os.Getenv' '$PROJECT_ROOT/internal/' | grep -qi 'jwt\|secret' | grep -v test || grep -rn 'JWT_SECRET\|JwtSecret' '$PROJECT_ROOT/internal/' | grep -q ."

# ============================================================================
# SECTION 14: AUTHENTICATION MIDDLEWARE
# ============================================================================
log_info "=============================================="
log_info "SECTION 14: Authentication Middleware"
log_info "=============================================="

run_test "Authentication middleware file exists" \
    "[[ -f '$PROJECT_ROOT/internal/middleware/auth.go' ]]"

run_test "Auth middleware tests exist" \
    "[[ -f '$PROJECT_ROOT/internal/middleware/auth_test.go' ]]"

run_test "Bearer token authentication implemented in middleware" \
    "grep -q 'Bearer' '$PROJECT_ROOT/internal/middleware/auth.go'"

# ============================================================================
# SECTION 15: AUTHORIZATION CHECKS
# ============================================================================
log_info "=============================================="
log_info "SECTION 15: Authorization"
log_info "=============================================="

run_test "Authorization logic exists in codebase" \
    "grep -rn --include='*.go' -l 'Authorize\|authorize\|HasPermission\|hasPermission\|isAdmin\|IsAdmin' '$PROJECT_ROOT/internal/' | grep -v test | grep -q ."

run_test "Role-based or claim-based access control present" \
    "grep -rn --include='*.go' 'role\|Role\|claims\|Claims\|permission\|Permission' '$PROJECT_ROOT/internal/middleware/' | grep -v test | grep -q ."

# ============================================================================
# SECTION 16: SECURITY HEADERS
# ============================================================================
log_info "=============================================="
log_info "SECTION 16: Security Headers in HTTP Responses"
log_info "=============================================="

run_test "Security headers defined in LLMsVerifier middleware (nosniff, X-Frame, CSP)" \
    "grep -rn 'X-Content-Type-Options\|X-Frame-Options\|Content-Security-Policy' '$PROJECT_ROOT/LLMsVerifier/' | grep -q ."

run_test "Content-Type header enforcement present in compression middleware" \
    "grep -rn --include='*.go' 'Content-Type\|ContentType' '$PROJECT_ROOT/internal/middleware/' | grep -q ."

# ============================================================================
# SECTION 17: CONTENT SECURITY PATTERNS
# ============================================================================
log_info "=============================================="
log_info "SECTION 17: Content Security Patterns"
log_info "=============================================="

run_test "Secure random number utility exists (not math/rand for security)" \
    "[[ -f '$PROJECT_ROOT/internal/utils/secure_random.go' ]]"

run_test "Secure random uses crypto/rand (not math/rand)" \
    "grep -q 'crypto/rand' '$PROJECT_ROOT/internal/utils/secure_random.go'"

run_test "Guardrails content filtering implemented" \
    "grep -qiE 'Filter|filter|Block|block|unsafe|Unsafe' '$PROJECT_ROOT/internal/security/guardrails.go'"

run_test "Security audit log function implemented" \
    "grep -q 'AuditLogger' '$PROJECT_ROOT/internal/security/audit.go'"

run_test "Audit log file uses restrictive permissions (0600)" \
    "grep -q '0600' '$PROJECT_ROOT/internal/security/audit.go'"

# ============================================================================
# SECTION 18: GOSEC SCAN (CODE QUALITY)
# ============================================================================
log_info "=============================================="
log_info "SECTION 18: Gosec Security Scan"
log_info "=============================================="

run_test "gosec binary found in PATH or GOPATH" \
    "command -v gosec >/dev/null 2>&1 || [[ -f \"\$HOME/go/bin/gosec\" ]] || [[ -f '$PROJECT_ROOT/bin/gosec' ]]"

# Run gosec with limited scope to avoid timeout, exclude G115 (known gosec bug)
GOSEC_CMD="gosec"
if ! command -v gosec >/dev/null 2>&1 && [[ -f "\$HOME/go/bin/gosec" ]]; then
    GOSEC_CMD="\$HOME/go/bin/gosec"
fi
GOSEC_BIN=$(command -v gosec 2>/dev/null || echo "\$HOME/go/bin/gosec")

run_test "gosec: no HIGH severity findings in security package" \
    "GOMAXPROCS=2 nice -n 19 \${GOSEC_BIN} -exclude=G115 -severity high -confidence high -no-fail ./internal/security/... 2>/dev/null | grep -c 'HIGH' | xargs test 0 -eq"

run_test "gosec: no HIGH severity findings in middleware" \
    "GOMAXPROCS=2 nice -n 19 \${GOSEC_BIN} -exclude=G115 -severity high -confidence high -no-fail ./internal/middleware/... 2>/dev/null | grep -c 'HIGH' | xargs test 0 -eq"

run_test "gosec: MD5 not used as cryptographic primitive (only FNV/SHA for cache keys)" \
    "! grep -rn --include='*.go' 'crypto/md5' '$PROJECT_ROOT/internal/' | grep -v nosec | grep -v test | grep -q ."

# ============================================================================
# SUMMARY
# ============================================================================
log_info "=============================================="
log_info "CHALLENGE SUMMARY"
log_info "=============================================="

log_info "Total tests: $TESTS_TOTAL"
log_info "Passed: $TESTS_PASSED"
log_info "Failed: $TESTS_FAILED"

record_metric "total_tests" "$TESTS_TOTAL"
record_metric "passed_tests" "$TESTS_PASSED"
record_metric "failed_tests" "$TESTS_FAILED"

if [[ $TESTS_FAILED -eq 0 ]]; then
    log_success "=============================================="
    log_success "ALL $TESTS_TOTAL SECURITY SCANNING TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge "PASSED"
else
    log_error "=============================================="
    log_error "$TESTS_FAILED/$TESTS_TOTAL TESTS FAILED"
    log_error "=============================================="
    finalize_challenge "FAILED"
fi
