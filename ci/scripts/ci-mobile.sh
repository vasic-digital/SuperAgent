#!/usr/bin/env bash
set -euo pipefail

# Phase 2: Mobile CI Pipeline
# Builds Flutter/RN apps, runs Robolectric tests, E2E on emulator

WORKSPACE="${WORKSPACE:-/workspace}"
REPORTS_DIR="${WORKSPACE}/reports/mobile"
RELEASES_DIR="${WORKSPACE}/releases/mobile"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PHASE_START=$SECONDS
PHASE_FAILURES=0

mkdir -p "${REPORTS_DIR}" "${RELEASES_DIR}/flutter" "${RELEASES_DIR}/react-native"

cd "${WORKSPACE}"

echo "========================================"
echo "Phase 2: Mobile CI Pipeline"
echo "Started: ${TIMESTAMP}"
echo "========================================"

# --- Wait for emulator ---
echo ""
echo "--- Waiting for Android emulator ---"
EMULATOR_HOST="${EMULATOR_HOST:-localhost}"
/usr/local/bin/wait-for-services.sh "${EMULATOR_HOST}:5555"

# Connect to emulator
adb connect "${EMULATOR_HOST}:5555" 2>/dev/null || true
adb -s "${EMULATOR_HOST}:5555" wait-for-device
echo "[OK] Emulator connected"

# ============================================================
# FLUTTER APP
# ============================================================
FLUTTER_DIR="${WORKSPACE}/LLMsVerifier/llm-verifier/mobile/flutter_app"

if [ -d "${FLUTTER_DIR}" ] && [ -f "${FLUTTER_DIR}/pubspec.yaml" ]; then
  echo ""
  echo "========================================"
  echo "Flutter App"
  echo "========================================"

  cd "${FLUTTER_DIR}"

  # Install dependencies
  echo "--- Flutter pub get ---"
  flutter pub get 2>&1 | tee "${REPORTS_DIR}/flutter-pub-get.log" || {
    echo "[WARN] Flutter pub get failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Unit/widget tests
  echo "--- Flutter tests ---"
  flutter test --coverage --machine \
    > "${REPORTS_DIR}/flutter-tests.json" 2>&1 || {
    echo "[WARN] Some Flutter tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy coverage
  if [ -f "coverage/lcov.info" ]; then
    cp "coverage/lcov.info" "${REPORTS_DIR}/flutter-coverage.lcov"
  fi

  # Build signed APK
  echo "--- Flutter build APK ---"
  flutter build apk --release \
    2>&1 | tee "${REPORTS_DIR}/flutter-build-apk.log" || {
    echo "[WARN] Flutter APK build failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy APK to releases
  APK_PATH=$(find build/app/outputs -name "*.apk" -type f 2>/dev/null | head -1 || true)
  if [ -n "${APK_PATH}" ]; then
    cp "${APK_PATH}" "${RELEASES_DIR}/flutter/llm-verifier.apk"
    echo "[OK] Flutter APK: ${RELEASES_DIR}/flutter/llm-verifier.apk"

    # Verify signing
    if command -v apksigner >/dev/null 2>&1; then
      apksigner verify --print-certs "${RELEASES_DIR}/flutter/llm-verifier.apk" \
        > "${REPORTS_DIR}/flutter-signing.txt" 2>&1 || true
    fi
  fi

  # Build AAB
  echo "--- Flutter build AAB ---"
  flutter build appbundle --release \
    2>&1 | tee "${REPORTS_DIR}/flutter-build-aab.log" || {
    echo "[WARN] Flutter AAB build failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  AAB_PATH=$(find build/app/outputs -name "*.aab" -type f 2>/dev/null | head -1 || true)
  if [ -n "${AAB_PATH}" ]; then
    cp "${AAB_PATH}" "${RELEASES_DIR}/flutter/llm-verifier.aab"
    echo "[OK] Flutter AAB: ${RELEASES_DIR}/flutter/llm-verifier.aab"
  fi

  # Robolectric tests (via Gradle)
  echo "--- Robolectric tests ---"
  if [ -f "android/gradlew" ]; then
    cd android
    chmod +x gradlew
    ./gradlew testReleaseUnitTest --no-daemon \
      2>&1 | tee "${REPORTS_DIR}/flutter-robolectric.log" || {
      echo "[WARN] Robolectric tests failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
    # Copy test results
    find . -path "*/test-results/*.xml" \
      -exec cp {} "${REPORTS_DIR}/" \; 2>/dev/null || true
    cd "${FLUTTER_DIR}"
  fi

  # Install on emulator and verify
  echo "--- Emulator install test ---"
  if [ -f "${RELEASES_DIR}/flutter/llm-verifier.apk" ]; then
    adb -s ${EMULATOR_HOST}:5555 install -r \
      "${RELEASES_DIR}/flutter/llm-verifier.apk" \
      2>&1 | tee "${REPORTS_DIR}/flutter-install.log" || {
      echo "[WARN] Flutter APK install failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }

    # Verify package is installed
    if adb -s ${EMULATOR_HOST}:5555 shell pm list packages 2>/dev/null \
       | grep -q "llm_verifier"; then
      echo "[OK] Flutter app installed on emulator"
    else
      echo "[WARN] Flutter app package not found on emulator"
    fi
  fi
else
  echo "[SKIP] Flutter app directory not found"
fi

# ============================================================
# REACT NATIVE APP
# ============================================================
RN_DIR="${WORKSPACE}/LLMsVerifier/llm-verifier/mobile/react-native"

if [ -d "${RN_DIR}" ] && [ -f "${RN_DIR}/package.json" ]; then
  echo ""
  echo "========================================"
  echo "React Native App"
  echo "========================================"

  cd "${RN_DIR}"

  # Install dependencies
  echo "--- npm install ---"
  if [ -f "package-lock.json" ]; then
    npm ci 2>&1 | tee "${REPORTS_DIR}/rn-npm-install.log" || {
      echo "[WARN] React Native npm ci failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
  else
    npm install 2>&1 | tee "${REPORTS_DIR}/rn-npm-install.log" || {
      echo "[WARN] React Native npm install failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
  fi

  # Jest tests
  echo "--- Jest tests ---"
  npx jest --ci --coverage \
    --reporters=default --reporters=jest-junit \
    2>&1 | tee "${REPORTS_DIR}/rn-jest.log" || {
    echo "[WARN] Some Jest tests failed"
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
  }

  # Copy coverage and results
  if [ -d "coverage" ]; then
    cp -r coverage "${REPORTS_DIR}/rn-coverage" 2>/dev/null || true
  fi
  [ -f "junit.xml" ] && cp junit.xml "${REPORTS_DIR}/rn-jest-tests.xml" || true

  # Build signed APK
  echo "--- React Native build APK ---"
  if [ -d "android" ] && [ -f "android/gradlew" ]; then
    cd android
    chmod +x gradlew
    ./gradlew assembleRelease --no-daemon \
      -Pandroid.injected.signing.store.file="${WORKSPACE}/keys/android/debug.keystore" \
      -Pandroid.injected.signing.store.password="helixagent-debug" \
      -Pandroid.injected.signing.key.alias="helixagent-debug" \
      -Pandroid.injected.signing.key.password="helixagent-debug" \
      2>&1 | tee "${REPORTS_DIR}/rn-build-apk.log" || {
      echo "[WARN] React Native APK build failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }

    RN_APK=$(find . -path "*/release/*.apk" -type f 2>/dev/null | head -1 || true)
    if [ -n "${RN_APK}" ]; then
      cp "${RN_APK}" "${RELEASES_DIR}/react-native/llm-verifier-rn.apk"
      echo "[OK] RN APK: ${RELEASES_DIR}/react-native/llm-verifier-rn.apk"

      if command -v apksigner >/dev/null 2>&1; then
        apksigner verify --print-certs \
          "${RELEASES_DIR}/react-native/llm-verifier-rn.apk" \
          > "${REPORTS_DIR}/rn-signing.txt" 2>&1 || true
      fi
    fi

    # Robolectric tests
    echo "--- RN Robolectric tests ---"
    ./gradlew testReleaseUnitTest --no-daemon \
      2>&1 | tee "${REPORTS_DIR}/rn-robolectric.log" || {
      echo "[WARN] RN Robolectric tests failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
    find . -path "*/test-results/*.xml" \
      -exec cp {} "${REPORTS_DIR}/" \; 2>/dev/null || true

    cd "${RN_DIR}"
  fi

  # Install on emulator
  echo "--- Emulator install test ---"
  if [ -f "${RELEASES_DIR}/react-native/llm-verifier-rn.apk" ]; then
    adb -s ${EMULATOR_HOST}:5555 install -r \
      "${RELEASES_DIR}/react-native/llm-verifier-rn.apk" \
      2>&1 | tee "${REPORTS_DIR}/rn-install.log" || {
      echo "[WARN] RN APK install failed"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    }
  fi
else
  echo "[SKIP] React Native app directory not found"
fi

# ============================================================
# NATIVE SDK VALIDATION
# ============================================================
echo ""
echo "========================================"
echo "Native SDK Validation"
echo "========================================"

# Kotlin SDK
KOTLIN_SDK="${WORKSPACE}/sdk/android/SuperAgent.kt"
if [ -f "${KOTLIN_SDK}" ]; then
  echo "--- Kotlin SDK compile check ---"
  if command -v kotlinc >/dev/null 2>&1; then
    kotlinc "${KOTLIN_SDK}" -d /tmp/kotlin-out 2>"${REPORTS_DIR}/kotlin-compile.log" && {
      echo "[OK] Kotlin SDK compiles"
    } || {
      echo "[WARN] Kotlin SDK compilation issues (see reports/mobile/kotlin-compile.log)"
    }
  else
    echo "[SKIP] kotlinc not available"
  fi
fi

# Swift SDK (syntax validation only - no Xcode on Linux)
SWIFT_SDK="${WORKSPACE}/sdk/ios/SuperAgent.swift"
if [ -f "${SWIFT_SDK}" ]; then
  echo "--- Swift SDK syntax check ---"
  if command -v swiftc >/dev/null 2>&1; then
    swiftc -parse "${SWIFT_SDK}" 2>"${REPORTS_DIR}/swift-check.log" && {
      echo "[OK] Swift SDK syntax valid"
    } || {
      echo "[WARN] Swift SDK syntax issues (flagged for macOS CI)"
    }
  else
    echo "[SKIP] Swift compiler not available (Linux) - flagged for macOS CI"
  fi
fi

# --- Artifact validation ---
echo ""
echo "--- Artifact validation ---"
/usr/local/bin/validate-artifacts.sh mobile \
  2>&1 | tee "${REPORTS_DIR}/artifact-validation.log" || {
  PHASE_FAILURES=$((PHASE_FAILURES + 1))
}

# --- Signing verification summary ---
echo ""
echo "--- Signing verification ---"

FLUTTER_SIG="not_checked"
RN_SIG="not_checked"
[ -f "${REPORTS_DIR}/flutter-signing.txt" ] && \
  grep -q "Signer" "${REPORTS_DIR}/flutter-signing.txt" 2>/dev/null && \
  FLUTTER_SIG="verified"
[ -f "${REPORTS_DIR}/rn-signing.txt" ] && \
  grep -q "Signer" "${REPORTS_DIR}/rn-signing.txt" 2>/dev/null && \
  RN_SIG="verified"

cat > "${REPORTS_DIR}/signing-verification.json" <<SIGEOF
{
  "flutter_apk": "${FLUTTER_SIG}",
  "react_native_apk": "${RN_SIG}",
  "keystore": "debug.keystore",
  "timestamp": "${TIMESTAMP}"
}
SIGEOF

# --- False positive validation ---
echo ""
echo "--- False positive validation ---"

# Source the function library
source /usr/local/bin/false-positive-check.sh
FP_PHASE="mobile"

THRESHOLDS="${WORKSPACE}/ci/thresholds.json"
if [ -f "${THRESHOLDS}" ]; then
  # Flutter unit tests
  if [ -f "${REPORTS_DIR}/flutter-tests.json" ]; then
    # parse count from json
    FLUTTER_TESTS=$(jq -r '.testCount // 0' "${REPORTS_DIR}/flutter-tests.json" 2>/dev/null || echo 0)
    if [ "${FLUTTER_TESTS}" -lt "$(jq -r '.mobile.flutter.min_unit_tests' "${THRESHOLDS}")" ]; then
      echo "[FAIL] Flutter unit tests below threshold"
      PHASE_FAILURES=$((PHASE_FAILURES + 1))
    fi
  fi
  # Flutter coverage
  if [ -f "${REPORTS_DIR}/flutter-coverage.lcov" ]; then
    validate_flutter_coverage "${REPORTS_DIR}/flutter-coverage.lcov" "flutter" \
      "$(jq -r '.mobile.flutter.min_coverage_percent' "${THRESHOLDS}")"
  fi
  # React Native Jest tests
  if [ -f "${REPORTS_DIR}/rn-jest-tests.xml" ]; then
    validate_junit_xml "${REPORTS_DIR}/rn-jest-tests.xml" "rn_jest" \
      "$(jq -r '.mobile.react_native.min_jest_tests' "${THRESHOLDS}")"
  fi
  # React Native Jest coverage
  if [ -f "${REPORTS_DIR}/rn-coverage/coverage-summary.json" ]; then
    validate_jest_coverage "${REPORTS_DIR}/rn-coverage/coverage-summary.json" "rn_jest" \
      "$(jq -r '.mobile.react_native.min_coverage_percent' "${THRESHOLDS}")"
  fi
  # Robolectric tests (flutter)
  ROBOTEST_FILES=$(find "${REPORTS_DIR}" -name "*.xml" -path "*test-results*" 2>/dev/null | head -1)
  if [ -n "${ROBOTEST_FILES}" ]; then
    validate_junit_xml "${ROBOTEST_FILES}" "robolectric" \
      "$(jq -r '.mobile.robolectric.min_tests' "${THRESHOLDS}")"
  fi
  # Robolectric coverage (if available)
  ROBOTEST_COVERAGE=$(find "${REPORTS_DIR}" -name "coverage*.ec" -o -name "coverage.xml" 2>/dev/null | head -1)
  if [ -n "${ROBOTEST_COVERAGE}" ]; then
    # Placeholder: Robolectric coverage validation not yet implemented
    echo "[INFO] Robolectric coverage file found but validation not implemented"
  fi
fi

write_fp_report "${REPORTS_DIR}/false-positive-checks.json"
PHASE_FAILURES=$((PHASE_FAILURES + FAILURES))

PHASE_DURATION=$((SECONDS - PHASE_START))

echo ""
echo "========================================"
echo "Phase 2 Complete"
echo "Duration: ${PHASE_DURATION}s"
echo "Failures: ${PHASE_FAILURES}"
echo "Reports:  ${REPORTS_DIR}/"
echo "Releases: ${RELEASES_DIR}/"
echo "========================================"

exit "${PHASE_FAILURES}"
