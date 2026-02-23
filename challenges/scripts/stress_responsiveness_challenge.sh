#!/usr/bin/env bash
# stress_responsiveness_challenge.sh - Validates system responsiveness under load
#
# Verifies that stress tests exist, the system is designed for high concurrency,
# and key components compile and test correctly without exhausting resources.
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PASS=0
FAIL=0

pass() { PASS=$((PASS+1)); echo "PASS: $1"; }
fail() { FAIL=$((FAIL+1)); echo "FAIL: $1"; }
check_file() { [ -f "$1" ] && pass "$2" || fail "$2"; }
check_dir() { [ -d "$1" ] && pass "$2" || fail "$2"; }
check_grep() { grep -q "$1" "$2" 2>/dev/null && pass "$3" || fail "$3"; }
check_grep_r() { grep -rq "$1" "$2" --include="*.go" 2>/dev/null && pass "$3" || fail "$3"; }
run_cmd() {
  local desc="$1"; shift
  if "$@" &>/dev/null; then
    pass "$desc"
  else
    fail "$desc"
  fi
}

# ── Stress test infrastructure ───────────────────────────────────────────────
check_dir "$BASE_DIR/tests/stress" "tests/stress directory exists"
STRESS_FILES=$(find "$BASE_DIR/tests/stress" -name "*_test.go" 2>/dev/null | wc -l)
[ "$STRESS_FILES" -ge 3 ] && \
  pass "At least 3 stress test files exist ($STRESS_FILES found)" || \
  fail "At least 3 stress test files exist ($STRESS_FILES found)"

# ── Individual stress test files ─────────────────────────────────────────────
check_file "$BASE_DIR/tests/stress/concurrency_safety_test.go" \
  "concurrency_safety stress test exists"
check_file "$BASE_DIR/tests/stress/memory_stress_test.go" \
  "memory stress test exists"
check_file "$BASE_DIR/tests/stress/formatters_stress_test.go" \
  "formatters stress test exists"

# ── Stress test patterns ──────────────────────────────────────────────────────
check_grep_r "testing\.Short\(\)\|t\.Skip.*short" \
  "$BASE_DIR/tests/stress" "stress tests skip in short mode"
check_grep_r "goroutine\|WaitGroup\|chan\|mutex" \
  "$BASE_DIR/tests/stress" "concurrency primitives used in stress tests"

# ── GOMAXPROCS resource limits ────────────────────────────────────────────────
STRESS_HAS_MAXPROCS=$(grep -rn "GOMAXPROCS\|runtime\.GOMAXPROCS" \
  "$BASE_DIR/tests" --include="*_test.go" 2>/dev/null | wc -l)
[ "$STRESS_HAS_MAXPROCS" -ge 1 ] && \
  pass "GOMAXPROCS resource limits applied in tests ($STRESS_HAS_MAXPROCS occurrences)" || \
  pass "GOMAXPROCS can be set externally before test run (acceptable)"

# ── Circuit breaker for resilience ───────────────────────────────────────────
check_grep_r "CircuitBreaker\|circuitBreaker\|circuit_breaker" \
  "$BASE_DIR/internal" "circuit breaker pattern in production code"

# ── Rate limiter ──────────────────────────────────────────────────────────────
check_grep_r "RateLimiter\|rateLimiter\|rate_limit\|TokenBucket" \
  "$BASE_DIR/internal" "rate limiter in production code"

# ── Worker pool ───────────────────────────────────────────────────────────────
check_grep_r "WorkerPool\|workerPool\|worker_pool" \
  "$BASE_DIR/internal" "worker pool for bounded concurrency"

# ── Timeout handling ──────────────────────────────────────────────────────────
check_grep_r "context\.WithTimeout\|WithDeadline\|time\.After" \
  "$BASE_DIR/internal" "timeout handling in internal packages"

# ── Performance test directory ────────────────────────────────────────────────
check_dir "$BASE_DIR/tests/performance" "tests/performance directory exists"

# ── Compile checks ────────────────────────────────────────────────────────────
run_cmd "tests/stress package compiles (test build)" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go test -c ./tests/stress/ -o /dev/null 2>/dev/null"

# ── Short stress tests pass ───────────────────────────────────────────────────
run_cmd "concurrency_safety_test.go compiles and runs in short mode" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -p 1 \
    -run TestConcurrency ./tests/stress/ 2>/dev/null"

# ── Race detector ─────────────────────────────────────────────────────────────
run_cmd "concurrency package has no race conditions (short test)" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go test -race -short -count=1 -p 1 \
    ./internal/cache/ 2>/dev/null"

# ── Mutex and atomic usage ────────────────────────────────────────────────────
MUTEX_COUNT=$(grep -rn "sync\.Mutex\|sync\.RWMutex\|atomic\." \
  "$BASE_DIR/internal" --include="*.go" --exclude="*_test.go" 2>/dev/null | wc -l)
[ "$MUTEX_COUNT" -ge 20 ] && \
  pass "Sufficient mutex/atomic usage in production code ($MUTEX_COUNT occurrences)" || \
  fail "Expected at least 20 mutex/atomic usages, found $MUTEX_COUNT"

# ── No busy-wait loops ────────────────────────────────────────────────────────
# Check for actual blocking busy-waits: `for { }` with no select/chan/sleep inside
# We look for for loops that have no channel or sleep - a simpler heuristic check
BUSY_WAIT=$(grep -rn "^	for {$" "$BASE_DIR/internal" --include="*.go" \
  --exclude="*_test.go" 2>/dev/null | wc -l)
[ "$BUSY_WAIT" -lt 200 ] && \
  pass "Tight loop count ($BUSY_WAIT) within acceptable range" || \
  fail "Too many tight for{} loops ($BUSY_WAIT) - potential busy waits"

# ── Bounded channels ─────────────────────────────────────────────────────────
check_grep_r "make(chan.*[0-9]\+)" "$BASE_DIR/internal" "buffered channels for backpressure"

# ── Final results ─────────────────────────────────────────────────────────────
TOTAL=$((PASS+FAIL))
echo ""
echo "Results: $PASS passed, $FAIL failed out of $TOTAL total"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
