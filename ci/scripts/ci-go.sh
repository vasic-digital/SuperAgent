#!/usr/bin/env bash
set -euo pipefail

# Phase 1: Go CI Pipeline
# Runs all Go tests, builds all binaries, generates reports

WORKSPACE="${WORKSPACE:-/workspace}"
REPORTS_DIR="${WORKSPACE}/reports/go"
RELEASES_DIR="${WORKSPACE}/releases"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PHASE_START=$SECONDS
PHASE_FAILURES=0

mkdir -p "${REPORTS_DIR}"

cd "${WORKSPACE}"

echo "========================================"
echo "Phase 1: Go CI Pipeline"
echo "Started: ${TIMESTAMP}"
echo "========================================"

# --- Step 1: Wait for integration services ---
echo ""
echo "--- Step 1: Waiting for services ---"
/usr/local/bin/wait-for-services.sh \
  postgres:5432 \
  redis:6379 \
  mockllm:8090

# --- Step 2: Code quality gates ---
echo ""
echo "--- Step 2: Code quality ---"

echo "Running go fmt check..."
UNFMT=$(gofmt -l ./internal/ ./cmd/ 2>/dev/null || true)
if [ -n "${UNFMT}" ]; then
  echo "[WARN] Unformatted files:"
  echo "${UNFMT}"
fi

echo "Running go vet..."
go vet ./... > "${REPORTS_DIR}/vet-output.txt" 2>&1 && echo "[OK] go vet passed" || {
  echo "[WARN] go vet had findings (see reports/go/vet-output.txt)"
}

echo "Running golangci-lint..."
golangci-lint run --timeout 5m --out-format json \
  > "${REPORTS_DIR}/lint-results.json" 2>/dev/null && echo "[OK] lint passed" || {
  echo "[WARN] lint issues found (see reports/go/lint-results.json)"
}

if command -v gosec >/dev/null 2>&1; then
  echo "Running gosec..."
  gosec -fmt json -out "${REPORTS_DIR}/security-scan.json" ./... 2>/dev/null && echo "[OK] gosec passed" || {
    echo "[WARN] security issues found (see reports/go/security-scan.json)"
  }
fi

# --- Step 3: Unit tests ---
echo ""
echo "--- Step 3: Unit tests ---"

export DB_HOST="${DB_HOST:-postgres}" DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-helixagent}" DB_PASSWORD="${DB_PASSWORD:-helixagent123}"
export DB_NAME="${DB_NAME:-helixagent_db}"
export REDIS_HOST="${REDIS_HOST:-redis}" REDIS_PORT="${REDIS_PORT:-6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-helixagent123}"
export MOCK_LLM_URL="${MOCK_LLM_URL:-http://mockllm:8090}"
export MOCK_LLM_ENABLED="${MOCK_LLM_ENABLED:-true}"
export JWT_SECRET="${JWT_SECRET:-ci-test-secret}"
export CI=true FULL_TEST_MODE=true

gotestsum --format standard-verbose \
  --junitfile "${REPORTS_DIR}/unit-tests.xml" \
  -- -short -coverprofile="${REPORTS_DIR}/unit-coverage.out" \
  -covermode=atomic -timeout 300s \
  ./internal/... \
  2>&1 | tee "${REPORTS_DIR}/unit-tests.log" || {
  echo "[WARN] Some unit tests failed"
  PHASE_FAILURES=$((PHASE_FAILURES + 1))
}

go tool cover -func="${REPORTS_DIR}/unit-coverage.out" \
  > "${REPORTS_DIR}/unit-coverage-summary.txt" 2>/dev/null || true
go tool cover -html="${REPORTS_DIR}/unit-coverage.out" \
  -o "${REPORTS_DIR}/unit-coverage.html" 2>/dev/null || true

# --- Step 4: Integration tests ---
echo ""
echo "--- Step 4: Integration tests ---"

if [ -d "./tests/integration" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/integration-tests.xml" \
    -- -coverprofile="${REPORTS_DIR}/integration-coverage.out" \
    -covermode=atomic -timeout 600s \
    ./tests/integration/... \
    2>&1 | tee "${REPORTS_DIR}/integration-tests.log" || {
    echo "[WARN] Some integration tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }
else
  echo "[SKIP] No integration tests directory"
fi

# --- Step 5: E2E tests ---
echo ""
echo "--- Step 5: E2E tests ---"

if [ -d "./tests/e2e" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/e2e-tests.xml" \
    -- -coverprofile="${REPORTS_DIR}/e2e-coverage.out" \
    -covermode=atomic -timeout 600s \
    ./tests/e2e/... \
    2>&1 | tee "${REPORTS_DIR}/e2e-tests.log" || {
    echo "[WARN] Some E2E tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }
else
  echo "[SKIP] No E2E tests directory"
fi

# --- Step 6: Security tests ---
echo ""
echo "--- Step 6: Security tests ---"

if [ -d "./tests/security" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/security-tests.xml" \
    -- -timeout 300s \
    ./tests/security/... \
    2>&1 | tee "${REPORTS_DIR}/security-tests.log" || {
    echo "[WARN] Some security tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }
else
  echo "[SKIP] No security tests directory"
fi

# --- Step 7: Stress tests ---
echo ""
echo "--- Step 7: Stress tests ---"

if [ -d "./tests/stress" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/stress-tests.xml" \
    -- -timeout 600s \
    ./tests/stress/... \
    2>&1 | tee "${REPORTS_DIR}/stress-tests.log" || {
    echo "[WARN] Some stress tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }
else
  echo "[SKIP] No stress tests directory"
fi

# --- Step 8: Benchmarks ---
echo ""
echo "--- Step 8: Benchmarks ---"

# Limit total benchmark time to 20 minutes (some benchmarks are pathologically slow)
timeout 1200 go test -bench=. -benchmem -benchtime=1s -timeout 120s -short \
  ./internal/... \
  > "${REPORTS_DIR}/benchmark-results.txt" 2>&1 || {
  echo "[WARN] Some benchmarks failed or timed out"
}

# --- Step 9: Race detection ---
echo ""
echo "--- Step 9: Race detection ---"

gotestsum --format standard-verbose \
  --junitfile "${REPORTS_DIR}/race-tests.xml" \
  -- -race -short -timeout 300s \
  ./internal/... \
  2>&1 | tee "${REPORTS_DIR}/race-tests.log" || {
  echo "[WARN] Race detection found issues"
  PHASE_FAILURES=$((PHASE_FAILURES + 1))
}

# --- Step 10: Build all releases ---
echo ""
echo "--- Step 10: Building releases ---"

if [ -f "${WORKSPACE}/scripts/build/version-manager.sh" ]; then
  # shellcheck source=/dev/null
  source "${WORKSPACE}/scripts/build/version-manager.sh"

  APPS=(helixagent api grpc-server cognee-mock sanity-check mcp-bridge generate-constitution)
  # Windows excluded: syscall.Statfs_t not available for cross-compilation
  PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

  SEMANTIC_VERSION=$(get_semantic_version)

  for app in "${APPS[@]}"; do
    CMD_PATH="${APP_REGISTRY[${app}]:-./cmd/${app}}"
    VERSION_CODE=$(increment_version_code "${app}")
    SOURCE_HASH=$(compute_source_hash "${app}")
    LDFLAGS=$(get_ldflags "${app}" "${VERSION_CODE}" "${SOURCE_HASH}" "ci-go")

    for platform in "${PLATFORMS[@]}"; do
      GOOS="${platform%%/*}"
      GOARCH="${platform#*/}"
      EXT=""
      [ "${GOOS}" = "windows" ] && EXT=".exe"

      OUTPUT_DIR="${RELEASES_DIR}/${app}/${GOOS}-${GOARCH}/${VERSION_CODE}"
      mkdir -p "${OUTPUT_DIR}"
      OUTPUT="${OUTPUT_DIR}/${app}${EXT}"

      echo "Building ${app} for ${GOOS}/${GOARCH}..."
      CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" \
        go build -ldflags "${LDFLAGS}" -o "${OUTPUT}" "./${CMD_PATH}" 2>&1 || {
        echo "[FAIL] Build failed: ${app} ${GOOS}/${GOARCH}"
        PHASE_FAILURES=$((PHASE_FAILURES + 1))
        continue
      }

      # Generate build-info.json
      HASH=$(sha256sum "${OUTPUT}" | awk '{print $1}')
      cat > "${OUTPUT_DIR}/build-info.json" <<BIEOF
{
  "app": "${app}",
  "version": "${SEMANTIC_VERSION}",
  "version_code": ${VERSION_CODE},
  "os": "${GOOS}",
  "arch": "${GOARCH}",
  "sha256": "${HASH}",
  "git_commit": "$(git rev-parse HEAD 2>/dev/null || echo unknown)",
  "git_branch": "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)",
  "build_date": "${TIMESTAMP}",
  "go_version": "$(go version | awk '{print $3}')",
  "builder": "ci-go"
}
BIEOF

      # Update latest symlink
      ln -sfn "${VERSION_CODE}" "${RELEASES_DIR}/${app}/${GOOS}-${GOARCH}/latest"
      echo "[OK] ${app} ${GOOS}/${GOARCH}"
    done

    save_hash "${app}" "${SOURCE_HASH}"
  done
else
  echo "[WARN] version-manager.sh not found, skipping release builds"
  PHASE_FAILURES=$((PHASE_FAILURES + 1))
fi

# --- Step 11: Artifact validation ---
echo ""
echo "--- Step 11: Artifact validation ---"

/usr/local/bin/validate-artifacts.sh go 2>&1 | tee "${REPORTS_DIR}/artifact-validation.log" || {
  PHASE_FAILURES=$((PHASE_FAILURES + 1))
}

# --- Step 12: False positive checks ---
echo ""
echo "--- Step 12: False positive validation ---"

# Source the function library
# shellcheck source=/dev/null
source /usr/local/bin/false-positive-check.sh
FP_PHASE="go"

THRESHOLDS="${WORKSPACE}/ci/thresholds.json"
if [ -f "${THRESHOLDS}" ]; then
  validate_junit_xml "${REPORTS_DIR}/unit-tests.xml" "unit" \
    "$(jq -r '.go.min_unit_tests' "${THRESHOLDS}")"
  validate_junit_xml "${REPORTS_DIR}/integration-tests.xml" "integration" \
    "$(jq -r '.go.min_integration_tests' "${THRESHOLDS}")"
  validate_junit_xml "${REPORTS_DIR}/e2e-tests.xml" "e2e" \
    "$(jq -r '.go.min_e2e_tests' "${THRESHOLDS}")"
  validate_junit_xml "${REPORTS_DIR}/security-tests.xml" "security" \
    "$(jq -r '.go.min_security_tests' "${THRESHOLDS}")"
  validate_coverage "${REPORTS_DIR}/unit-coverage-summary.txt" "unit" \
    "$(jq -r '.go.min_coverage_percent' "${THRESHOLDS}")"
fi

write_fp_report "${REPORTS_DIR}/false-positive-checks.json"

PHASE_DURATION=$((SECONDS - PHASE_START))

echo ""
echo "========================================"
echo "Phase 1 Complete"
echo "Duration: ${PHASE_DURATION}s"
echo "Failures: ${PHASE_FAILURES}"
echo "Reports:  ${REPORTS_DIR}/"
echo "Releases: ${RELEASES_DIR}/"
echo "========================================"

exit "${PHASE_FAILURES}"
