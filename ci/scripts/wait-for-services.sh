#!/usr/bin/env bash
set -euo pipefail

# Wait for required services to become healthy before proceeding
# Usage: wait-for-services.sh <service1:port> [service2:port] ...
# Example: wait-for-services.sh postgres:5432 redis:6379 mockllm:8090

TIMEOUT="${CI_SERVICE_TIMEOUT:-120}"
INTERVAL=3
FAILED=0

wait_for_tcp() {
  local host="$1"
  local port="$2"
  local label="${host}:${port}"
  local start=$SECONDS
  local elapsed=0

  echo "[wait] Waiting for ${label}..."
  while ! nc -z "${host}" "${port}" 2>/dev/null; do
    elapsed=$((SECONDS - start))
    if [ "${elapsed}" -ge "${TIMEOUT}" ]; then
      echo "[FAIL] ${label} not ready after ${TIMEOUT}s"
      return 1
    fi
    sleep "${INTERVAL}"
  done
  elapsed=$((SECONDS - start))
  echo "[OK]   ${label} ready (${elapsed}s)"
}

wait_for_http() {
  local url="$1"
  local label="$2"
  local start=$SECONDS
  local elapsed=0

  echo "[wait] Waiting for ${label} at ${url}..."
  while ! curl -sf "${url}" >/dev/null 2>&1; do
    elapsed=$((SECONDS - start))
    if [ "${elapsed}" -ge "${TIMEOUT}" ]; then
      echo "[FAIL] ${label} not ready after ${TIMEOUT}s"
      return 1
    fi
    sleep "${INTERVAL}"
  done
  elapsed=$((SECONDS - start))
  echo "[OK]   ${label} ready (${elapsed}s)"
}

for SERVICE in "$@"; do
  case "${SERVICE}" in
    postgres:*)
      PORT="${SERVICE#*:}"
      wait_for_tcp "postgres" "${PORT}" || FAILED=$((FAILED + 1))
      # Extra: wait for pg_isready
      if command -v pg_isready >/dev/null 2>&1; then
        timeout "${TIMEOUT}" bash -c \
          "until pg_isready -h postgres -p ${PORT} -U helixagent 2>/dev/null; do sleep ${INTERVAL}; done" \
          && echo "[OK]   PostgreSQL accepting connections" \
          || { echo "[FAIL] PostgreSQL not accepting connections"; FAILED=$((FAILED + 1)); }
      fi
      ;;
    redis:*)
      PORT="${SERVICE#*:}"
      wait_for_tcp "redis" "${PORT}" || FAILED=$((FAILED + 1))
      ;;
    mockllm:*)
      PORT="${SERVICE#*:}"
      wait_for_http "http://mockllm:${PORT}/health" "Mock LLM" || FAILED=$((FAILED + 1))
      ;;
    oauthmock:*)
      PORT="${SERVICE#*:}"
      wait_for_tcp "oauthmock" "${PORT}" || FAILED=$((FAILED + 1))
      ;;
    chromadb:*)
      PORT="${SERVICE#*:}"
      wait_for_http "http://chromadb:${PORT}/api/v1/heartbeat" "ChromaDB" || FAILED=$((FAILED + 1))
      ;;
    qdrant:*)
      PORT="${SERVICE#*:}"
      wait_for_http "http://qdrant:${PORT}/healthz" "Qdrant" || FAILED=$((FAILED + 1))
      ;;
    kafka:*)
      PORT="${SERVICE#*:}"
      wait_for_tcp "kafka" "${PORT}" || FAILED=$((FAILED + 1))
      ;;
    rabbitmq:*)
      PORT="${SERVICE#*:}"
      wait_for_tcp "rabbitmq" "${PORT}" || FAILED=$((FAILED + 1))
      ;;
    minio:*)
      PORT="${SERVICE#*:}"
      wait_for_http "http://minio:${PORT}/minio/health/ready" "MinIO" || FAILED=$((FAILED + 1))
      ;;
    emulator:*)
      PORT="${SERVICE#*:}"
      echo "[wait] Waiting for Android emulator on ${PORT}..."
      if command -v adb >/dev/null 2>&1; then
        timeout "${TIMEOUT}" bash -c \
          "until adb connect emulator:${PORT} 2>/dev/null | grep -q connected; do sleep ${INTERVAL}; done" \
          && echo "[OK]   Android emulator connected" \
          || { echo "[FAIL] Android emulator not ready"; FAILED=$((FAILED + 1)); }
      else
        echo "[SKIP] adb not available"
      fi
      ;;
    *)
      HOST="${SERVICE%%:*}"
      PORT="${SERVICE#*:}"
      wait_for_tcp "${HOST}" "${PORT}" || FAILED=$((FAILED + 1))
      ;;
  esac
done

if [ "${FAILED}" -gt 0 ]; then
  echo ""
  echo "[ERROR] ${FAILED} service(s) failed health check"
  exit 1
fi

echo ""
echo "[OK] All services ready"
