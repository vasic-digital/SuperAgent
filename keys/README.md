# HelixAgent Signing Keys Management

## Overview

All signing keys for CI/CD builds are stored in the `keys/` directory. This document describes the key structure, usage, and security considerations.

## Directory Structure

```
keys/
├── android/
│   ├── debug.keystore          # Debug signing key (safe to commit)
│   ├── release.keystore        # Release signing key (DO NOT COMMIT - .gitignored)
│   └── README.md
├── ios/
│   ├── development.p12         # iOS development certificate
│   ├── distribution.p12        # iOS distribution certificate
│   ├── provisioning/           # Provisioning profiles
│   └── README.md
├── desktop/
│   ├── code-signing.p12        # Windows code signing
│   ├── mac-developer.p12       # macOS developer certificate
│   └── README.md
├── web/
│   ├── code-signing.key        # Web extension signing
│   └── README.md
└── README.md                   # This file
```

## Android Signing

### Debug Keystore (Included)

The debug keystore is included in the repository for development and CI builds:

- **File**: `keys/android/debug.keystore`
- **Alias**: `helixagent-debug`
- **Store Password**: `helixagent-debug`
- **Key Password**: `helixagent-debug`
- **Validity**: 10000 days

Apps signed with this key can be installed on devices but **cannot** be published to Google Play Store.

### Release Keystore (Required for Production)

For Play Store distribution, you must provide a release keystore:

```bash
# Generate release keystore
keytool -genkeypair \
  -alias helixagent-release \
  -keyalg RSA \
  -keysize 2048 \
  -validity 10000 \
  -keystore keys/android/release.keystore

# Set environment variables for CI
export CI_ANDROID_RELEASE_KEYSTORE=/path/to/release.keystore
export CI_KEYSTORE_PASSWORD=<your-store-password>
export CI_KEY_ALIAS=helixagent-release
export CI_KEY_PASSWORD=<your-key-password>
```

### CI/CD Integration

The CI pipeline automatically selects the appropriate keystore:

1. If `CI_ANDROID_RELEASE_KEYSTORE` is set → uses release keystore
2. Otherwise → falls back to `keys/android/debug.keystore`

## iOS Signing

### Prerequisites

1. Apple Developer Account
2. Development and Distribution certificates
3. Provisioning profiles for your app

### Setup

```bash
# Export certificates from Keychain
security find-identity -v -p codesigning

# Create .p12 files
# (Use Keychain Access on macOS to export)

# Copy provisioning profiles
cp ~/Library/MobileDevice/Provisioning\ Profiles/*.mobileprovision keys/ios/provisioning/
```

### CI/CD Environment Variables

```bash
export CI_IOS_DEV_CERTIFICATE=keys/ios/development.p12
export CI_IOS_DIST_CERTIFICATE=keys/ios/distribution.p12
export CI_IOS_CERT_PASSWORD=<your-p12-password>
export CI_IOS_PROVISIONING_PROFILE=keys/ios/provisioning/AppStore.mobileprovision
```

## Desktop Signing

### Windows Code Signing

```bash
# Obtain code signing certificate from CA (DigiCert, Sectigo, etc.)
# Convert to PFX format

export CI_WINDOWS_CERT=keys/desktop/code-signing.p12
export CI_WINDOWS_CERT_PASSWORD=<password>
```

### macOS Code Signing

```bash
# Requires Apple Developer certificate
export CI_MAC_CERT=keys/desktop/mac-developer.p12
export CI_MAC_CERT_PASSWORD=<password>
export CI_MAC_IDENTITY="Developer ID Application: Your Name (TEAMID)"
```

## Web Extension Signing

```bash
# Generate key for browser extension signing
openssl genrsa -out keys/web/code-signing.key 2048

export CI_WEB_SIGNING_KEY=keys/web/code-signing.key
```

## Security Guidelines

### DO
- ✅ Use environment variables for passwords
- ✅ Store release keystores securely (outside repo)
- ✅ Rotate keys periodically
- ✅ Use different keys for debug/release
- ✅ Enable 2FA on signing certificate providers

### DO NOT
- ❌ Commit release keystores to version control
- ❌ Hardcode passwords in scripts
- ❌ Share keystore files via email/chat
- ❌ Use debug keystores for production
- ❌ Skip signing for release builds

## Git Configuration

The `.gitignore` is configured to exclude:

```gitignore
keys/android/release.keystore
keys/android/*.jks
keys/ios/*.p12
keys/ios/provisioning/
keys/desktop/*.p12
keys/desktop/*.key
keys/web/*.key
keys/**/*.env
```

## CI/CD Key Injection

For CI/CD systems, inject keys via secure environment variables:

### GitHub Actions (if enabled)

```yaml
env:
  CI_ANDROID_RELEASE_KEYSTORE: ${{ secrets.ANDROID_KEYSTORE_BASE64 }}
  CI_KEYSTORE_PASSWORD: ${{ secrets.ANDROID_KEYSTORE_PASSWORD }}
```

### Manual CI Execution

```bash
# Encode keystore as base64 for environment variable
export CI_ANDROID_KEYSTORE_B64=$(base64 -w0 keys/android/release.keystore)

# In CI, decode and use
echo "$CI_ANDROID_KEYSTORE_B64" | base64 -d > /tmp/release.keystore
export CI_ANDROID_RELEASE_KEYSTORE=/tmp/release.keystore
```

## Troubleshooting

### "Keystore was tampered with"
- Verify password is correct
- Check file wasn't corrupted during transfer

### "Certificate expired"
- Generate new keystore with longer validity
- Or obtain new signing certificate from CA

### "Unable to sign APK"
- Verify all four signing parameters are set
- Check keystore file path is accessible
- Ensure key alias matches what's in keystore
