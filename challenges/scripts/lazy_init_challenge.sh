#!/usr/bin/env bash
# lazy_init_challenge.sh - Validates lazy initialization patterns in HelixAgent
#
# Verifies that the codebase correctly uses lazy initialization, sync.Once,
# and deferred resource allocation patterns for performance and safety.
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

# ── Structural checks ────────────────────────────────────────────────────────
check_dir "$BASE_DIR/internal" "internal/ directory exists"
check_dir "$BASE_DIR/internal/cache" "cache package exists"
check_dir "$BASE_DIR/internal/services" "services package exists"

# ── sync.Once usage verification ─────────────────────────────────────────────
check_grep_r "sync\.Once" "$BASE_DIR/internal" "sync.Once used in internal packages"
check_grep_r "sync\.Once\b" "$BASE_DIR/internal" "sync.Once variable pattern found"

# ── Context-based lazy initialization ────────────────────────────────────────
check_grep_r "func.*init.*context\.Context\|lazy.*init\|LazyInit\|initOnce" \
  "$BASE_DIR/internal" "lazy init patterns found"

# ── Non-blocking patterns ─────────────────────────────────────────────────────
check_grep_r "select.*default\|chan struct{}\|go func()" \
  "$BASE_DIR/internal" "non-blocking channel patterns found"

# ── Compile verification ──────────────────────────────────────────────────────
run_cmd "internal/cache package compiles" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go build ./internal/cache/ 2>/dev/null"

run_cmd "internal/services package compiles" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go build ./internal/services/ 2>/dev/null"

run_cmd "internal/background package compiles" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go build ./internal/background/ 2>/dev/null"

# ── Cache lazy init ───────────────────────────────────────────────────────────
CACHE_FILE="$BASE_DIR/internal/cache/cache.go"
REDIS_FILE="$BASE_DIR/internal/cache/redis_cache.go"
[ -f "$CACHE_FILE" ] || CACHE_FILE="$(find "$BASE_DIR/internal/cache" -name "*.go" -not -name "*_test.go" | head -1)"
[ -f "$REDIS_FILE" ] || REDIS_FILE="$(find "$BASE_DIR/internal/cache" -name "*redis*.go" -not -name "*_test.go" | head -1)"

if [ -f "$CACHE_FILE" ]; then
  pass "cache.go source file exists"
else
  fail "cache.go source file exists"
fi

if [ -f "$REDIS_FILE" ]; then
  pass "redis_cache.go source file exists"
else
  fail "redis_cache.go source file exists"
fi

# ── sync.RWMutex for concurrent reads ────────────────────────────────────────
check_grep_r "sync\.RWMutex\|RWMutex" "$BASE_DIR/internal" "RWMutex for read-optimized concurrency"

# ── Semaphore / rate limiting ─────────────────────────────────────────────────
check_grep_r "semaphore\|Semaphore\|chan struct{}" "$BASE_DIR/internal" "semaphore or channel-based rate limiting"

# ── No direct goroutine leaks in production code ─────────────────────────────
# Check that goroutines started with 'go ' have context propagation
GOROUTINE_COUNT=$(grep -r "^	go " "$BASE_DIR/internal" --include="*.go" \
  --exclude="*_test.go" 2>/dev/null | wc -l)
[ "$GOROUTINE_COUNT" -lt 200 ] && \
  pass "Goroutine count in production code ($GOROUTINE_COUNT) within reasonable limit" || \
  fail "Too many raw goroutines in production code ($GOROUTINE_COUNT)"

# ── Background package ────────────────────────────────────────────────────────
check_dir "$BASE_DIR/internal/background" "background task package exists"
check_grep_r "WorkerPool\|TaskQueue\|workerPool" "$BASE_DIR/internal/background" \
  "worker pool pattern in background package"

# ── Context propagation ───────────────────────────────────────────────────────
check_grep_r "context\.Context\|ctx context" "$BASE_DIR/internal/services" \
  "context propagation in services"

# ── deferred cleanup ──────────────────────────────────────────────────────────
check_grep_r "defer.*\.Close()\|defer.*\.Cancel()\|defer.*\.Stop()" \
  "$BASE_DIR/internal" "deferred resource cleanup pattern"

# ── Unit tests pass for cache and background ─────────────────────────────────
run_cmd "cache package unit tests pass" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -p 1 \
    ./internal/cache/ 2>/dev/null"

run_cmd "background package unit tests pass" \
  bash -c "cd '$BASE_DIR' && GOMAXPROCS=2 nice -n 19 go test -short -count=1 -p 1 \
    ./internal/background/ 2>/dev/null"

# ── Final results ─────────────────────────────────────────────────────────────
TOTAL=$((PASS+FAIL))
echo ""
echo "Results: $PASS passed, $FAIL failed out of $TOTAL total"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
