#!/usr/bin/env bash
set -euo pipefail

# Phase 5: Integration Tests (Full Stack)
# Runs comprehensive integration tests across all services

WORKSPACE="${WORKSPACE:-/workspace}"
REPORTS_DIR="${WORKSPACE}/reports/integration"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PHASE_START=$SECONDS
PHASE_FAILURES=0

mkdir -p "${REPORTS_DIR}"

cd "${WORKSPACE}"

echo "========================================"
echo "Phase 5: Integration Tests"
echo "Started: ${TIMESTAMP}"
echo "========================================"

# --- Wait for all required services ---
echo ""
echo "--- Waiting for services ---"
/usr/local/bin/wait-for-services.sh \
  postgres:5432 \
  redis:6379 \
  mockllm:8090 \
  chromadb:8000 \
  qdrant:6333 \
  kafka:9092 \
  rabbitmq:5672 \
  minio:9000

# Set environment for integration tests
export DB_HOST="${DB_HOST:-postgres}" DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-helixagent}" DB_PASSWORD="${DB_PASSWORD:-helixagent123}"
export DB_NAME="${DB_NAME:-helixagent_db}"
export REDIS_HOST="${REDIS_HOST:-redis}" REDIS_PORT="${REDIS_PORT:-6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-helixagent123}"
export MOCK_LLM_URL="${MOCK_LLM_URL:-http://mockllm:8090}"
export MOCK_LLM_ENABLED="${MOCK_LLM_ENABLED:-true}"
export JWT_SECRET="${JWT_SECRET:-ci-test-secret}"
export CHROMADB_URL="${CHROMADB_URL:-http://chromadb:8000}"
export QDRANT_URL="${QDRANT_URL:-http://qdrant:6333}"
export KAFKA_BROKERS="${KAFKA_BROKERS:-kafka:9092}"
export RABBITMQ_URL="${RABBITMQ_URL:-amqp://helixagent:helixagent123@rabbitmq:5672/}"
export MINIO_URL="${MINIO_URL:-http://minio:9000}"
export MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY:-helixagent}"
export MINIO_SECRET_KEY="${MINIO_SECRET_KEY:-helixagent123}"
export CI=true FULL_TEST_MODE=true FULL_INTEGRATION_MODE=true

# --- Integration test suites ---
echo ""
echo "--- Integration test suite ---"

if [ -d "./tests/integration" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/integration-tests.xml" \
    -- -coverprofile="${REPORTS_DIR}/integration-coverage.out" \
    -covermode=atomic -timeout 1800s \
    ./tests/integration/... \
    2>&1 | tee "${REPORTS_DIR}/integration-tests.log" || {
    echo "[WARN] Some integration tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  go tool cover -func="${REPORTS_DIR}/integration-coverage.out" \
    > "${REPORTS_DIR}/integration-coverage-summary.txt" 2>/dev/null || true
  go tool cover -html="${REPORTS_DIR}/integration-coverage.out" \
    -o "${REPORTS_DIR}/integration-coverage.html" 2>/dev/null || true
else
  echo "[SKIP] No integration tests directory"
fi

# --- Cross-service orchestration tests ---
echo ""
echo "--- Cross-service orchestration tests ---"

if [ -d "./tests/orchestration" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/orchestration-tests.xml" \
    -- -coverprofile="${REPORTS_DIR}/orchestration-coverage.out" \
    -covermode=atomic -timeout 1200s \
    ./tests/orchestration/... \
    2>&1 | tee "${REPORTS_DIR}/orchestration-tests.log" || {
    echo "[WARN] Some orchestration tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }
else
  echo "[SKIP] No orchestration tests directory"
fi

# --- End-to-end workflow tests ---
echo ""
echo "--- End-to-end workflow tests ---"

if [ -d "./tests/workflow" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/workflow-tests.xml" \
    -- -coverprofile="${REPORTS_DIR}/workflow-coverage.out" \
    -covermode=atomic -timeout 2400s \
    ./tests/workflow/... \
    2>&1 | tee "${REPORTS_DIR}/workflow-tests.log" || {
    echo "[WARN] Some workflow tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }
else
  echo "[SKIP] No workflow tests directory"
fi

# --- Performance under load tests ---
echo ""
echo "--- Performance under load tests ---"

if [ -d "./tests/load" ]; then
  gotestsum --format standard-verbose \
    --junitfile "${REPORTS_DIR}/load-tests.xml" \
    -- -timeout 3600s \
    ./tests/load/... \
    2>&1 | tee "${REPORTS_DIR}/load-tests.log" || {
    echo "[WARN] Some load tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }
else
  echo "[SKIP] No load tests directory"
fi

# --- False positive validation ---
echo ""
echo "--- False positive validation ---"

source /usr/local/bin/false-positive-check.sh
FP_PHASE="integration"

THRESHOLDS="${WORKSPACE}/ci/thresholds.json"
if [ -f "${THRESHOLDS}" ]; then
  # Integration test counts
  if [ -f "${REPORTS_DIR}/integration-tests.xml" ]; then
    validate_junit_xml "${REPORTS_DIR}/integration-tests.xml" "integration" \
      "$(jq -r '.integration.min_tests' "${THRESHOLDS}")"
  fi
  # Integration coverage
  if [ -f "${REPORTS_DIR}/integration-coverage-summary.txt" ]; then
    validate_coverage "${REPORTS_DIR}/integration-coverage-summary.txt" "integration" \
      "$(jq -r '.integration.min_coverage_percent' "${THRESHOLDS}")"
  fi
fi

write_fp_report "${REPORTS_DIR}/false-positive-checks.json"
PHASE_FAILURES=$((PHASE_FAILURES + FAILURES))

PHASE_DURATION=$((SECONDS - PHASE_START))

echo ""
echo "========================================"
echo "Phase 5 Complete"
echo "Duration: ${PHASE_DURATION}s"
echo "Failures: ${PHASE_FAILURES}"
echo "Reports:  ${REPORTS_DIR}/"
echo "========================================"

exit "${PHASE_FAILURES}"