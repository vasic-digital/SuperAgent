#!/usr/bin/env bash
set -euo pipefail

# False Positive Prevention Framework
# Validates test results are genuine, not silent passes
#
# Can be sourced (to use functions) or run standalone:
#   source false-positive-check.sh
#   validate_junit_xml "path/to/tests.xml" "label" 10
#
# Standalone: false-positive-check.sh <phase> <reports-dir> [thresholds-file]

CHECKS=()
FAILURES=0

add_check() {
  local name="$1"
  local expected="$2"
  local actual="$3"
  local verdict="$4"
  CHECKS+=("{\"name\":\"${name}\",\"expected\":\"${expected}\",\"actual\":\"${actual}\",\"verdict\":\"${verdict}\"}")
  if [ "${verdict}" = "FAIL" ]; then
    FAILURES=$((FAILURES + 1))
    echo "[FP-FAIL] ${name}: expected=${expected}, actual=${actual}"
  else
    echo "[FP-OK]   ${name}: ${actual}"
  fi
}

# Validate JUnit XML files have actual tests
validate_junit_xml() {
  local xml_file="$1"
  local label="$2"
  local min_tests="${3:-1}"

  if [ ! -f "${xml_file}" ]; then
    add_check "${label}_report_exists" "file exists" "missing" "FAIL"
    return
  fi

  local tests_count
  tests_count=$(grep -oE 'tests="[0-9]+"' "${xml_file}" | head -1 | grep -oE '[0-9]+' || echo "0")
  local failures_count
  failures_count=$(grep -oE 'failures="[0-9]+"' "${xml_file}" | head -1 | grep -oE '[0-9]+' || echo "0")

  if [ "${tests_count}" -lt "${min_tests}" ]; then
    add_check "${label}_test_count" ">=${min_tests}" "${tests_count}" "FAIL"
  else
    add_check "${label}_test_count" ">=${min_tests}" "${tests_count}" "PASS"
  fi

  if [ "${failures_count}" -gt 0 ]; then
    add_check "${label}_no_failures" "0 failures" "${failures_count} failures" "FAIL"
  else
    add_check "${label}_no_failures" "0 failures" "0 failures" "PASS"
  fi
}

# Validate coverage percentage from summary file
validate_coverage() {
  local coverage_file="$1"
  local label="$2"
  local min_percent="${3:-100}"

  if [ ! -f "${coverage_file}" ]; then
    add_check "${label}_coverage_exists" "file exists" "missing" "FAIL"
    return
  fi

  local coverage
  coverage=$(grep -oE '[0-9]+\.[0-9]+%' "${coverage_file}" | tail -1 | sed 's/%//' || echo "0")
  local coverage_int=${coverage%.*}

  if [ "${coverage_int}" -lt "${min_percent}" ]; then
    add_check "${label}_coverage" ">=${min_percent}%" "${coverage}%" "FAIL"
  else
    add_check "${label}_coverage" ">=${min_percent}%" "${coverage}%" "PASS"
  fi
}

# Validate binary artifact exists and has correct properties
validate_binary() {
  local binary_path="$1"
  local label="$2"
  local expected_arch="${3:-}"

  if [ ! -f "${binary_path}" ]; then
    add_check "${label}_exists" "file exists" "missing" "FAIL"
    return
  fi

  local size
  size=$(stat -c%s "${binary_path}" 2>/dev/null || echo "0")
  if [ "${size}" -lt 1000 ]; then
    add_check "${label}_size" ">1KB" "${size} bytes" "FAIL"
  else
    add_check "${label}_size" ">1KB" "${size} bytes" "PASS"
  fi

  if [ -n "${expected_arch}" ]; then
    local file_type
    file_type=$(file "${binary_path}" 2>/dev/null || echo "unknown")
    if echo "${file_type}" | grep -qi "${expected_arch}"; then
      add_check "${label}_arch" "${expected_arch}" "matched" "PASS"
    else
      add_check "${label}_arch" "${expected_arch}" "${file_type}" "FAIL"
    fi
  fi
}

# Validate APK is signed
validate_apk_signing() {
  local apk_path="$1"
  local label="$2"

  if [ ! -f "${apk_path}" ]; then
    add_check "${label}_exists" "file exists" "missing" "FAIL"
    return
  fi

  if command -v apksigner >/dev/null 2>&1; then
    if apksigner verify --print-certs "${apk_path}" 2>/dev/null | grep -q "Signer"; then
      add_check "${label}_signed" "signed" "verified" "PASS"
    else
      add_check "${label}_signed" "signed" "not signed" "FAIL"
    fi
  else
    add_check "${label}_signer_available" "apksigner present" "not installed" "FAIL"
  fi
}

# Write JSON report of all checks
write_fp_report() {
  local output_file="$1"

  {
    echo "{"
    echo "  \"phase\": \"${FP_PHASE:-unknown}\","
    echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
    echo "  \"total_checks\": ${#CHECKS[@]},"
    echo "  \"failures\": ${FAILURES},"
    echo "  \"checks\": ["
    local first=true
    for check in "${CHECKS[@]}"; do
      if [ "${first}" = true ]; then
        first=false
      else
        echo ","
      fi
      echo -n "    ${check}"
    done
    echo ""
    echo "  ]"
    echo "}"
  } > "${output_file}"

  echo ""
  echo "========================================"
  echo "False Positive Results: ${#CHECKS[@]} checks, ${FAILURES} failures"
  echo "Written to: ${output_file}"
  echo "========================================"
}

# Standalone mode: called directly with phase and reports dir
if [ "${BASH_SOURCE[0]}" = "$0" ] && [ $# -ge 2 ]; then
  FP_PHASE="$1"
  REPORTS_DIR="$2"
  THRESHOLDS_FILE="${3:-/workspace/ci/thresholds.json}"

  echo "========================================"
  echo "False Positive Validation: ${FP_PHASE}"
  echo "========================================"

  write_fp_report "${REPORTS_DIR}/false-positive-checks.json"

  if [ "${FAILURES}" -gt 0 ]; then
    exit 1
  fi
fi
