#!/usr/bin/env bash
set -euo pipefail

# Report aggregation orchestrator
# Called after all phases complete

WORKSPACE="${WORKSPACE:-/workspace}"

echo "========================================"
echo "CI Report Aggregation"
echo "========================================"

cd /opt/reporter
node aggregate.js

echo ""
echo "Reports generated:"
echo "  HTML Dashboard: ${WORKSPACE}/reports/summary.html"
echo "  JSON Results:   ${WORKSPACE}/reports/results.json"
echo ""

# Validate results
FAILED=$(jq -r '.totals.tests_failed' "${WORKSPACE}/reports/results.json" 2>/dev/null || echo "-1")
if [ "${FAILED}" = "0" ]; then
  echo "[OK] All tests passed"
elif [ "${FAILED}" -gt 0 ] 2>/dev/null; then
  echo "[FAIL] ${FAILED} test(s) failed across all phases"
  exit 1
else
  echo "[WARN] Could not parse results"
  exit 1
fi
