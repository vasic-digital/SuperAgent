#!/usr/bin/env bash
set -euo pipefail

# Phase 4: Desktop CI Pipeline
# Electron, Tauri desktop apps

WORKSPACE="${WORKSPACE:-/workspace}"
REPORTS_DIR="${WORKSPACE}/reports/desktop"
RELEASES_DIR="${WORKSPACE}/releases/desktop"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PHASE_START=$SECONDS
PHASE_FAILURES=0

mkdir -p "${REPORTS_DIR}" "${RELEASES_DIR}/electron" "${RELEASES_DIR}/tauri"

cd "${WORKSPACE}"

echo "========================================"
echo "Phase 4: Desktop CI Pipeline"
echo "Started: ${TIMESTAMP}"
echo "========================================"

# ============================================================
# ELECTRON APP
# ============================================================
ELECTRON_DIR="${WORKSPACE}/desktop/electron"

if [ -d "${ELECTRON_DIR}" ] && [ -f "${ELECTRON_DIR}/package.json" ]; then
  echo ""
  echo "========================================"
  echo "Electron App"
  echo "========================================"

  cd "${ELECTRON_DIR}"

  # Install dependencies
  echo "--- npm install ---"
  if [ -f "package-lock.json" ]; then
    npm ci 2>&1 | tee "${REPORTS_DIR}/electron-npm-install.log" || {
      echo "[WARN] Electron npm install failed"; PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
  else
    npm install 2>&1 | tee "${REPORTS_DIR}/electron-npm-install.log" || {
      echo "[WARN] Electron npm install failed"; PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
  fi

  # Unit tests (Jest)
  echo "--- Electron unit tests ---"
  npx jest --ci --coverage \
    --reporters=default --reporters=jest-junit \
    2>&1 | tee "${REPORTS_DIR}/electron-jest.log" || {
    echo "[WARN] Some Electron unit tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy coverage and results
  if [ -d "coverage" ]; then
    cp -r coverage "${REPORTS_DIR}/electron-coverage" 2>/dev/null || true
  fi
  [ -f "junit.xml" ] && cp junit.xml "${REPORTS_DIR}/electron-jest-tests.xml" || true

  # E2E tests with Playwright (if configured)
  echo "--- Electron E2E tests ---"
  if [ -f "playwright.config.js" ] || [ -f "playwright.config.ts" ]; then
    npx playwright test \
      2>&1 | tee "${REPORTS_DIR}/electron-playwright.log" || {
      echo "[WARN] Some Electron E2E tests failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
    find . -name "results.xml" -path "*test-results*" \
      -exec cp {} "${REPORTS_DIR}/electron-playwright-tests.xml" \; 2>/dev/null || true
  fi

  # Production build
  echo "--- Electron build ---"
  npm run build 2>&1 | tee "${REPORTS_DIR}/electron-build.log" || {
    echo "[FAIL] Electron build failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy build output
  if [ -d "dist" ]; then
    cp -r dist "${RELEASES_DIR}/electron/dist"
    echo "[OK] Electron build: ${RELEASES_DIR}/electron/dist/"
  fi

  # Signing (if keys available)
  if [ -f "${WORKSPACE}/keys/desktop/code-signing.p12" ]; then
    echo "--- Electron signing ---"
    # Placeholder: invoke signing script
    echo "[INFO] Code signing key present (signing not implemented)"
  fi
else
  echo "[SKIP] Electron app directory not found"
fi

# ============================================================
# TAURI APP
# ============================================================
TAURI_DIR="${WORKSPACE}/desktop/tauri"

if [ -d "${TAURI_DIR}" ] && [ -f "${TAURI_DIR}/package.json" ]; then
  echo ""
  echo "========================================"
  echo "Tauri App"
  echo "========================================"

  cd "${TAURI_DIR}"

  # Install dependencies
  echo "--- npm install ---"
  if [ -f "package-lock.json" ]; then
    npm ci 2>&1 | tee "${REPORTS_DIR}/tauri-npm-install.log" || {
      echo "[WARN] Tauri npm install failed"; PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
  else
    npm install 2>&1 | tee "${REPORTS_DIR}/tauri-npm-install.log" || {
      echo "[WARN] Tauri npm install failed"; PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
  fi

  # Unit tests (Jest)
  echo "--- Tauri unit tests ---"
  npx jest --ci --coverage \
    --reporters=default --reporters=jest-junit \
    2>&1 | tee "${REPORTS_DIR}/tauri-jest.log" || {
    echo "[WARN] Some Tauri unit tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy coverage and results
  if [ -d "coverage" ]; then
    cp -r coverage "${REPORTS_DIR}/tauri-coverage" 2>/dev/null || true
  fi
  [ -f "junit.xml" ] && cp junit.xml "${REPORTS_DIR}/tauri-jest-tests.xml" || true

  # Production build
  echo "--- Tauri build ---"
  npm run tauri build 2>&1 | tee "${REPORTS_DIR}/tauri-build.log" || {
    echo "[FAIL] Tauri build failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy build output
  if [ -d "src-tauri/target/release" ]; then
    find "src-tauri/target/release" -name "*.app" -o -name "*.exe" -o -name "*.AppImage" \
      2>/dev/null | while read -r artifact; do
      cp "${artifact}" "${RELEASES_DIR}/tauri/" 2>/dev/null || true
    done
    echo "[OK] Tauri build artifacts copied to ${RELEASES_DIR}/tauri/"
  fi

  # Signing (if keys available)
  if [ -f "${WORKSPACE}/keys/desktop/mac-developer.p12" ]; then
    echo "--- Tauri signing ---"
    # Placeholder: invoke signing script
    echo "[INFO] macOS signing key present (signing not implemented)"
  fi
else
  echo "[SKIP] Tauri app directory not found"
fi

# --- Artifact validation ---
echo ""
echo "--- Artifact validation ---"
/usr/local/bin/validate-artifacts.sh desktop \
  2>&1 | tee "${REPORTS_DIR}/artifact-validation.log" || {
  PHASE_FAILURES=$((PHASE_FAILURES + 1))
}

echo ""
echo "--- False positive validation ---"

# Source the function library
source /usr/local/bin/false-positive-check.sh
FP_PHASE="desktop"

THRESHOLDS="${WORKSPACE}/ci/thresholds.json"
if [ -f "${THRESHOLDS}" ]; then
  # Electron unit tests
  if [ -f "${REPORTS_DIR}/electron-jest-tests.xml" ]; then
    validate_junit_xml "${REPORTS_DIR}/electron-jest-tests.xml" "electron_jest" \
      "$(jq -r '.desktop.electron.min_unit_tests' "${THRESHOLDS}")"
  fi
  # Electron Jest coverage
  if [ -f "${REPORTS_DIR}/electron-coverage/coverage-summary.json" ]; then
    validate_jest_coverage "${REPORTS_DIR}/electron-coverage/coverage-summary.json" "electron_jest" \
      "$(jq -r '.desktop.electron.min_coverage_percent' "${THRESHOLDS}")"
  fi
  # Electron E2E tests
  if [ -f "${REPORTS_DIR}/electron-playwright-tests.xml" ]; then
    validate_junit_xml "${REPORTS_DIR}/electron-playwright-tests.xml" "electron_playwright" \
      "$(jq -r '.desktop.electron.min_e2e_tests' "${THRESHOLDS}")"
  fi
  # Tauri unit tests
  if [ -f "${REPORTS_DIR}/tauri-jest-tests.xml" ]; then
    validate_junit_xml "${REPORTS_DIR}/tauri-jest-tests.xml" "tauri_jest" \
      "$(jq -r '.desktop.tauri.min_unit_tests' "${THRESHOLDS}")"
  fi
  # Tauri Jest coverage
  if [ -f "${REPORTS_DIR}/tauri-coverage/coverage-summary.json" ]; then
    validate_jest_coverage "${REPORTS_DIR}/tauri-coverage/coverage-summary.json" "tauri_jest" \
      "$(jq -r '.desktop.tauri.min_coverage_percent' "${THRESHOLDS}")"
  fi
fi

write_fp_report "${REPORTS_DIR}/false-positive-checks.json"
PHASE_FAILURES=$((PHASE_FAILURES + FAILURES))

PHASE_DURATION=$((SECONDS - PHASE_START))

echo ""
echo "========================================"
echo "Phase 4 Complete"
echo "Duration: ${PHASE_DURATION}s"
echo "Failures: ${PHASE_FAILURES}"
echo "Reports:  ${REPORTS_DIR}/"
echo "Releases: ${RELEASES_DIR}/"
echo "========================================"

exit "${PHASE_FAILURES}"