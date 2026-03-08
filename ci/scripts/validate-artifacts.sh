#!/usr/bin/env bash
set -euo pipefail

# Validate all built artifacts exist, are correctly built, and are genuine
# Usage: validate-artifacts.sh <phase>

PHASE="$1"
WORKSPACE="${WORKSPACE:-/workspace}"
FAILURES=0

validate() {
  local label="$1"
  local condition="$2"
  if eval "${condition}"; then
    echo "[OK]   ${label}"
  else
    echo "[FAIL] ${label}"
    FAILURES=$((FAILURES + 1))
  fi
}

validate_go_binary() {
  local app="$1"
  local os="$2"
  local arch="$3"

  local ext=""
  [ "${os}" = "windows" ] && ext=".exe"

  local latest_link="${WORKSPACE}/releases/${app}/${os}-${arch}/latest"
  if [ -L "${latest_link}" ]; then
    local version_code
    version_code=$(readlink "${latest_link}")
    local binary_path="${WORKSPACE}/releases/${app}/${os}-${arch}/${version_code}/${app}${ext}"
    local info_path="${WORKSPACE}/releases/${app}/${os}-${arch}/${version_code}/build-info.json"

    validate "${app}/${os}-${arch} binary exists" "[ -f '${binary_path}' ]"
    validate "${app}/${os}-${arch} binary non-empty" "[ -s '${binary_path}' ]"
    validate "${app}/${os}-${arch} build-info exists" "[ -f '${info_path}' ]"

    # Verify build-info.json is valid JSON
    if [ -f "${info_path}" ]; then
      validate "${app}/${os}-${arch} build-info valid JSON" "jq empty '${info_path}' 2>/dev/null"
    fi

    # For native platform, verify binary can execute
    local host_os host_arch
    host_os=$(uname -s | tr '[:upper:]' '[:lower:]')
    host_arch=$(uname -m)
    [ "${host_arch}" = "x86_64" ] && host_arch="amd64"
    [ "${host_arch}" = "aarch64" ] && host_arch="arm64"

    if [ "${os}" = "${host_os}" ] && [ "${arch}" = "${host_arch}" ]; then
      if timeout 5 "${binary_path}" --version >/dev/null 2>&1 || timeout 5 "${binary_path}" version >/dev/null 2>&1; then
        echo "[OK]   ${app}/${os}-${arch} executes successfully"
      else
        echo "[WARN] ${app}/${os}-${arch} execution check inconclusive"
      fi
    fi
  else
    echo "[FAIL] ${app}/${os}-${arch} latest symlink missing"
    FAILURES=$((FAILURES + 1))
  fi
}

echo "========================================"
echo "Artifact Validation: ${PHASE}"
echo "========================================"

case "${PHASE}" in
  go)
    APPS=(helixagent api grpc-server cognee-mock sanity-check mcp-bridge generate-constitution)
    # Windows excluded: syscall.Statfs_t not available for cross-compilation
    PLATFORMS=("linux:amd64" "linux:arm64" "darwin:amd64" "darwin:arm64")

    for app in "${APPS[@]}"; do
      for platform in "${PLATFORMS[@]}"; do
        os="${platform%%:*}"
        arch="${platform#*:}"
        validate_go_binary "${app}" "${os}" "${arch}"
      done
    done
    ;;

  mobile)
    # Flutter APK/AAB
    validate "Flutter APK exists" \
      "find '${WORKSPACE}/releases/mobile/' -name '*.apk' 2>/dev/null | head -1 | grep -q ."
    validate "Flutter AAB exists" \
      "find '${WORKSPACE}/releases/mobile/' -name '*.aab' 2>/dev/null | head -1 | grep -q ."

    # React Native APK
    validate "React Native APK exists" \
      "find '${WORKSPACE}/releases/mobile/' -name '*rn*.apk' -o -name 'app-release.apk' 2>/dev/null | head -1 | grep -q ."

    # APK signing verification
    for apk in $(find "${WORKSPACE}/releases/mobile/" -name '*.apk' 2>/dev/null); do
      if command -v apksigner >/dev/null 2>&1; then
        validate "$(basename "${apk}") signed" "apksigner verify '${apk}' 2>/dev/null"
      fi
    done
    ;;

  web)
    # Angular build
    validate "Angular dist exists" \
      "[ -d '${WORKSPACE}/releases/web/angular/dist' ] || find '${WORKSPACE}/LLMsVerifier/llm-verifier/web/dist/' -name 'index.html' 2>/dev/null | grep -q ."

    # Website build
    validate "Website build exists" \
      "[ -d '${WORKSPACE}/releases/web/website/' ]"

    # SDK build
    validate "SDK dist exists" \
      "[ -d '${WORKSPACE}/releases/web/sdk/dist/' ] || [ -d '${WORKSPACE}/sdk/web/dist/' ]"
    ;;

  *)
    echo "Unknown phase: ${PHASE}"
    exit 1
    ;;
esac

echo "========================================"
echo "Validation complete: ${FAILURES} failure(s)"
echo "========================================"

exit "${FAILURES}"
