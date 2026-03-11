#!/usr/bin/env bash
set -euo pipefail

# HelixAgent Signing Keys Generator
# Generates all default signing keys for development and CI

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="$(dirname "${SCRIPT_DIR}")"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

generate_android_debug() {
    local ANDROID_DIR="${KEYS_DIR}/android"
    local KEYSTORE="${ANDROID_DIR}/debug.keystore"

    log_info "Generating Android debug keystore..."

    if [ -f "${KEYSTORE}" ]; then
        log_warn "Debug keystore already exists at ${KEYSTORE}"
        read -rp "Overwrite? (y/N): " choice
        case "${choice:-n}" in
            y|Y) rm -f "${KEYSTORE}" ;;
            *) return 0 ;;
        esac
    fi

    mkdir -p "${ANDROID_DIR}"

    keytool -genkeypair \
        -alias helixagent-debug \
        -keyalg RSA \
        -keysize 2048 \
        -validity 10000 \
        -keystore "${KEYSTORE}" \
        -storepass helixagent-debug \
        -keypass helixagent-debug \
        -dname "CN=HelixAgent Debug, O=HelixAgent, L=Development, C=US" \
        2>/dev/null

    if [ -f "${KEYSTORE}" ]; then
        log_info "✅ Android debug keystore created: ${KEYSTORE}"
        log_info "   Alias: helixagent-debug"
        log_info "   Password: helixagent-debug"
    else
        log_error "Failed to create Android debug keystore"
        return 1
    fi
}

generate_android_release() {
    local ANDROID_DIR="${KEYS_DIR}/android"
    local KEYSTORE="${ANDROID_DIR}/release.keystore"

    log_info "Generating Android release keystore..."

    if [ -f "${KEYSTORE}" ]; then
        log_warn "Release keystore already exists at ${KEYSTORE}"
        return 0
    fi

    read -rp "Generate release keystore for Play Store? (y/N): " choice
    case "${choice:-n}" in
        y|Y) ;;
        *) return 0 ;;
    esac

    read -rp "Keystore alias (default: helixagent-release): " alias
    alias="${alias:-helixagent-release}"

    read -rsp "Keystore password: " store_pass
    echo
    read -rsp "Confirm password: " store_pass_confirm
    echo

    if [ "${store_pass}" != "${store_pass_confirm}" ]; then
        log_error "Passwords do not match"
        return 1
    fi

    read -rp "Organization name: " org
    read -rp "Organization unit (default: Development): " org_unit
    org_unit="${org_unit:-Development}"
    read -rp "City/Locality: " city
    read -rp "State/Province: " state
    read -rp "Country code (2 letters, e.g., US): " country

    mkdir -p "${ANDROID_DIR}"

    keytool -genkeypair \
        -alias "${alias}" \
        -keyalg RSA \
        -keysize 2048 \
        -validity 10000 \
        -keystore "${KEYSTORE}" \
        -storepass "${store_pass}" \
        -keypass "${store_pass}" \
        -dname "CN=HelixAgent Release, OU=${org_unit}, O=${org}, L=${city}, ST=${state}, C=${country}" \
        2>/dev/null

    if [ -f "${KEYSTORE}" ]; then
        log_info "✅ Android release keystore created: ${KEYSTORE}"
        log_warn "⚠️  DO NOT COMMIT THIS FILE TO VERSION CONTROL"
        log_info "   Add to .gitignore: keys/android/release.keystore"
        echo ""
        log_info "Set environment variables for CI:"
        echo "  export CI_ANDROID_RELEASE_KEYSTORE=${KEYSTORE}"
        echo "  export CI_KEYSTORE_PASSWORD=<your-password>"
        echo "  export CI_KEY_ALIAS=${alias}"
        echo "  export CI_KEY_PASSWORD=<your-password>"
    else
        log_error "Failed to create Android release keystore"
        return 1
    fi
}

generate_ios_dev() {
    local IOS_DIR="${KEYS_DIR}/ios"

    log_warn "iOS certificates must be obtained from Apple Developer Portal"
    log_info "Steps:"
    echo "  1. Log in to https://developer.apple.com"
    echo "  2. Create development certificate in Certificates, Identifiers & Profiles"
    echo "  3. Export certificate from Keychain Access as .p12"
    echo "  4. Place in: ${IOS_DIR}/development.p12"
    echo ""
    log_info "Environment variables for CI:"
    echo "  export CI_IOS_DEV_CERTIFICATE=${IOS_DIR}/development.p12"
    echo "  export CI_IOS_CERT_PASSWORD=<your-p12-password>"
}

generate_desktop_codesign() {
    local DESKTOP_DIR="${KEYS_DIR}/desktop"

    log_warn "Desktop code signing certificates must be obtained from a CA"
    log_info ""
    log_info "Windows: Obtain from DigiCert, Sectigo, or other CA"
    log_info "  Place in: ${DESKTOP_DIR}/code-signing.p12"
    echo ""
    log_info "macOS: Obtain from Apple Developer Portal"
    log_info "  Place in: ${DESKTOP_DIR}/mac-developer.p12"
}

generate_web_key() {
    local WEB_DIR="${KEYS_DIR}/web"
    local KEY_FILE="${WEB_DIR}/code-signing.key"

    log_info "Generating web extension signing key..."

    if [ -f "${KEY_FILE}" ]; then
        log_warn "Web signing key already exists"
        return 0
    fi

    mkdir -p "${WEB_DIR}"

    openssl genrsa -out "${KEY_FILE}" 2048 2>/dev/null

    if [ -f "${KEY_FILE}" ]; then
        log_info "✅ Web signing key created: ${KEY_FILE}"
        log_warn "⚠️  DO NOT COMMIT THIS FILE TO VERSION CONTROL"
    else
        log_error "Failed to create web signing key"
        return 1
    fi
}

verify_keys() {
    log_info "Verifying signing keys..."
    echo ""

    local all_ok=true

    # Android debug
    if [ -f "${KEYS_DIR}/android/debug.keystore" ]; then
        log_info "✅ Android debug keystore: OK"
    else
        log_warn "❌ Android debug keystore: MISSING"
        all_ok=false
    fi

    # Android release
    if [ -f "${KEYS_DIR}/android/release.keystore" ]; then
        log_info "✅ Android release keystore: OK (remember: DO NOT COMMIT)"
    else
        log_info "ℹ️  Android release keystore: Not present (optional for Play Store)"
    fi

    # iOS
    if [ -f "${KEYS_DIR}/ios/development.p12" ]; then
        log_info "✅ iOS development certificate: OK"
    else
        log_info "ℹ️  iOS development certificate: Not present (required for iOS builds)"
    fi

    # Desktop
    if [ -f "${KEYS_DIR}/desktop/code-signing.p12" ]; then
        log_info "✅ Windows code signing: OK"
    else
        log_info "ℹ️  Windows code signing: Not present (optional)"
    fi

    echo ""
    if [ "${all_ok}" = true ]; then
        log_info "All required signing keys are present"
    else
        log_warn "Some signing keys are missing - run generation steps above"
    fi
}

show_usage() {
    echo "HelixAgent Signing Keys Generator"
    echo ""
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  all             Generate all default keys (debug only)"
    echo "  android-debug   Generate Android debug keystore"
    echo "  android-release Generate Android release keystore (interactive)"
    echo "  ios             Show iOS certificate setup instructions"
    echo "  desktop         Show desktop signing certificate instructions"
    echo "  web             Generate web extension signing key"
    echo "  verify          Verify all signing keys"
    echo ""
}

case "${1:-}" in
    all)
        generate_android_debug
        generate_web_key
        verify_keys
        ;;
    android-debug)
        generate_android_debug
        ;;
    android-release)
        generate_android_release
        ;;
    ios)
        generate_ios_dev
        ;;
    desktop)
        generate_desktop_codesign
        ;;
    web)
        generate_web_key
        ;;
    verify)
        verify_keys
        ;;
    *)
        show_usage
        exit 1
        ;;
esac
