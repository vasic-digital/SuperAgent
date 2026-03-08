#!/usr/bin/env bash
set -euo pipefail

# Phase 3: Web CI Pipeline
# Angular dashboard, static website, JS SDK

WORKSPACE="${WORKSPACE:-/workspace}"
REPORTS_DIR="${WORKSPACE}/reports/web"
RELEASES_DIR="${WORKSPACE}/releases/web"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PHASE_START=$SECONDS
PHASE_FAILURES=0

mkdir -p "${REPORTS_DIR}" "${RELEASES_DIR}/angular" "${RELEASES_DIR}/website" "${RELEASES_DIR}/sdk"

cd "${WORKSPACE}"

echo "========================================"
echo "Phase 3: Web CI Pipeline"
echo "Started: ${TIMESTAMP}"
echo "========================================"

# ============================================================
# ANGULAR WEB APP
# ============================================================
ANGULAR_DIR="${WORKSPACE}/LLMsVerifier/llm-verifier/web"

if [ -d "${ANGULAR_DIR}" ] && [ -f "${ANGULAR_DIR}/package.json" ]; then
  echo ""
  echo "========================================"
  echo "Angular Web App"
  echo "========================================"

  cd "${ANGULAR_DIR}"

  # Install dependencies
  echo "--- npm ci ---"
  npm ci 2>&1 | tee "${REPORTS_DIR}/angular-npm-install.log"

  # Karma/Jasmine unit tests
  echo "--- Karma unit tests ---"
  if [ -f "karma.conf.js" ]; then
    npx ng test --watch=false --browsers=ChromeHeadless --code-coverage \
      2>&1 | tee "${REPORTS_DIR}/angular-karma.log" || {
      echo "[WARN] Some Karma tests failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }

    # Copy coverage
    if [ -d "coverage" ]; then
      cp -r coverage "${REPORTS_DIR}/angular-coverage" 2>/dev/null || true
    fi
    # Copy JUnit results
    find . -name "TESTS-*.xml" \
      -exec cp {} "${REPORTS_DIR}/angular-karma-tests.xml" \; 2>/dev/null || true
  fi

  # Production build
  echo "--- Angular production build ---"
  npx ng build --configuration production \
    2>&1 | tee "${REPORTS_DIR}/angular-build.log" || {
    echo "[FAIL] Angular build failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy build output
  DIST_DIR=$(find . -path "*/dist/*" -name "index.html" \
    -exec dirname {} \; 2>/dev/null | head -1)
  if [ -n "${DIST_DIR}" ]; then
    cp -r "${DIST_DIR}" "${RELEASES_DIR}/angular/dist"
    echo "[OK] Angular build: ${RELEASES_DIR}/angular/dist/"
  fi

  # Playwright E2E tests
  echo "--- Playwright E2E tests ---"
  if [ -f "playwright.config.ts" ] || [ -d "e2e" ]; then
    if [ -d "${RELEASES_DIR}/angular/dist" ]; then
      http-server "${RELEASES_DIR}/angular/dist" -p 4200 -s &
      SERVER_PID=$!
      sleep 2

      npx playwright test \
        2>&1 | tee "${REPORTS_DIR}/angular-playwright.log" || {
        echo "[WARN] Some Playwright tests failed"
        PHASE_FAILURES=$((PHASE_FAILURES + 1))
      }

      # Copy playwright results
      find . -name "results.xml" -path "*test-results*" \
        -exec cp {} "${REPORTS_DIR}/angular-playwright-tests.xml" \; 2>/dev/null || true

      kill "${SERVER_PID}" 2>/dev/null || true
    fi
  fi

  # Lighthouse audit
  echo "--- Lighthouse audit ---"
  if [ -d "${RELEASES_DIR}/angular/dist" ]; then
    http-server "${RELEASES_DIR}/angular/dist" -p 4201 -s &
    SERVER_PID=$!
    sleep 2

    lhci collect --url="http://localhost:4201" \
      --settings.chromeFlags="--headless --no-sandbox" \
      2>&1 | tee "${REPORTS_DIR}/angular-lighthouse.log" || true

    find .lighthouseci -name "*.json" \
      -exec cp {} "${REPORTS_DIR}/angular-lighthouse.json" \; 2>/dev/null || true
    find .lighthouseci -name "*.html" \
      -exec cp {} "${REPORTS_DIR}/angular-lighthouse.html" \; 2>/dev/null || true

    kill "${SERVER_PID}" 2>/dev/null || true
    rm -rf .lighthouseci 2>/dev/null || true
  fi
else
  echo "[SKIP] Angular app directory not found"
fi

# ============================================================
# STATIC WEBSITE
# ============================================================
WEBSITE_DIR="${WORKSPACE}/Website"

if [ -d "${WEBSITE_DIR}" ] && [ -f "${WEBSITE_DIR}/package.json" ]; then
  echo ""
  echo "========================================"
  echo "Static Website"
  echo "========================================"

  cd "${WEBSITE_DIR}"

  # Install dependencies
  echo "--- npm ci ---"
  npm ci 2>&1 | tee "${REPORTS_DIR}/website-npm-install.log"

  # Build (PostCSS + UglifyJS)
  echo "--- Website build ---"
  npm run build 2>&1 | tee "${REPORTS_DIR}/website-build.log" || {
    echo "[WARN] Website build had issues"
  }

  # Copy build output
  if [ -d "public" ]; then
    cp -r public "${RELEASES_DIR}/website/"
    echo "[OK] Website build: ${RELEASES_DIR}/website/"
  fi

  # Serve and test with Playwright
  echo "--- Website Playwright tests ---"
  python3 -m http.server 8080 --directory "${RELEASES_DIR}/website/public" &
  SERVER_PID=$!
  sleep 2

  # Inline Playwright test spec
  mkdir -p /tmp/pw-tests
  cat > /tmp/pw-tests/website.spec.js <<'PWEOF'
const { test, expect } = require('@playwright/test');

test('homepage loads', async ({ page }) => {
  const response = await page.goto('http://localhost:8080');
  expect(response.status()).toBe(200);
  await expect(page.locator('body')).toBeVisible();
});

test('no console errors', async ({ page }) => {
  const errors = [];
  page.on('console', msg => {
    if (msg.type() === 'error') errors.push(msg.text());
  });
  await page.goto('http://localhost:8080');
  await page.waitForLoadState('networkidle');
  expect(errors).toHaveLength(0);
});

test('responsive layout', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 667 });
  await page.goto('http://localhost:8080');
  await expect(page.locator('body')).toBeVisible();

  await page.setViewportSize({ width: 1920, height: 1080 });
  await page.goto('http://localhost:8080');
  await expect(page.locator('body')).toBeVisible();
});
PWEOF

  npx playwright test /tmp/pw-tests/website.spec.js \
    2>&1 | tee "${REPORTS_DIR}/website-playwright.log" || {
    echo "[WARN] Website Playwright tests had issues"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Lighthouse audit
  echo "--- Website Lighthouse audit ---"
  lhci collect --url="http://localhost:8080" \
    --settings.chromeFlags="--headless --no-sandbox" \
    2>&1 | tee "${REPORTS_DIR}/website-lighthouse.log" || true

  find .lighthouseci -name "*.json" \
    -exec cp {} "${REPORTS_DIR}/website-lighthouse.json" \; 2>/dev/null || true
  find .lighthouseci -name "*.html" \
    -exec cp {} "${REPORTS_DIR}/website-lighthouse.html" \; 2>/dev/null || true

  kill "${SERVER_PID}" 2>/dev/null || true
  rm -rf .lighthouseci /tmp/pw-tests 2>/dev/null || true
else
  echo "[SKIP] Website directory not found"
fi

# ============================================================
# JAVASCRIPT SDK
# ============================================================
SDK_DIR="${WORKSPACE}/sdk/web"

if [ -d "${SDK_DIR}" ] && [ -f "${SDK_DIR}/package.json" ]; then
  echo ""
  echo "========================================"
  echo "JavaScript SDK"
  echo "========================================"

  cd "${SDK_DIR}"

  # Install dependencies
  echo "--- npm ci ---"
  npm ci 2>&1 | tee "${REPORTS_DIR}/sdk-npm-install.log"

  # Jest tests
  echo "--- Jest tests ---"
  npx jest --ci --coverage \
    --reporters=default --reporters=jest-junit \
    2>&1 | tee "${REPORTS_DIR}/sdk-jest.log" || {
    echo "[WARN] SDK Jest tests had issues"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  if [ -d "coverage" ]; then
    cp -r coverage "${REPORTS_DIR}/sdk-coverage" 2>/dev/null || true
  fi
  [ -f "junit.xml" ] && cp junit.xml "${REPORTS_DIR}/sdk-jest-tests.xml" || true

  # Build all targets
  echo "--- SDK build ---"
  npm run build 2>&1 | tee "${REPORTS_DIR}/sdk-build.log" || {
    echo "[WARN] SDK build had issues"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy build output
  if [ -d "dist" ]; then
    cp -r dist "${RELEASES_DIR}/sdk/dist"
    echo "[OK] SDK build: ${RELEASES_DIR}/sdk/dist/"

    # Bundle size check
    echo "--- Bundle sizes ---"
    for f in dist/*; do
      if [ -f "${f}" ]; then
        SIZE=$(stat -c%s "${f}" 2>/dev/null || echo 0)
        echo "  $(basename "${f}"): ${SIZE} bytes"
      fi
    done
  fi
else
  echo "[SKIP] JavaScript SDK directory not found"
fi

# --- Artifact validation ---
echo ""
echo "--- Artifact validation ---"
/usr/local/bin/validate-artifacts.sh web \
  2>&1 | tee "${REPORTS_DIR}/artifact-validation.log" || {
  PHASE_FAILURES=$((PHASE_FAILURES + 1))
}

PHASE_DURATION=$((SECONDS - PHASE_START))

echo ""
echo "========================================"
echo "Phase 3 Complete"
echo "Duration: ${PHASE_DURATION}s"
echo "Failures: ${PHASE_FAILURES}"
echo "Reports:  ${REPORTS_DIR}/"
echo "Releases: ${RELEASES_DIR}/"
echo "========================================"

exit "${PHASE_FAILURES}"
