# CI/CD Container Build System Guide

## Overview

HelixAgent uses a five-phase CI/CD system where **all builds, tests, and artifact generation run inside Docker/Podman containers**. This ensures reproducible builds, consistent environments, and complete isolation from the host system.

## Architecture

```
docker-compose.ci.yml
├── Profile: go-ci        Phase 1: Go builds + full test suite
├── Profile: mobile-ci    Phase 2: Flutter/RN + Robolectric + emulator
├── Profile: web-ci       Phase 3: Angular + Website + Playwright + Lighthouse
├── Profile: desktop-ci   Phase 4: Electron/Tauri desktop apps
├── Profile: integration  Phase 5: Full-stack integration tests
└── Profile: report       Report aggregation
```

All phases share a `ci-network` bridge network. Integration services (PostgreSQL, Redis, etc.) are started automatically as part of the Go CI profile.

## Prerequisites

- **Docker** (with compose plugin) or **Podman** (with podman-compose)
- **KVM** support for Android emulator (Phase 2 only): `ls /dev/kvm`
- **Disk space**: ~20GB for all CI images and caches

## Quick Start

```bash
# Full CI run (all phases + report)
make ci-all

# Single phase
make ci-go          # Go builds + tests
make ci-mobile      # Mobile builds + tests
make ci-web         # Web builds + tests
make ci-desktop     # Desktop apps (Electron/Tauri)
make ci-integration # Full-stack integration tests

# Report only (after phases have run)
make ci-report

# Build images without running
make ci-build-images

# Cleanup
make ci-clean
```

## Resource Configuration

Control resource usage via `CI_RESOURCE_LIMIT` environment variable:

| Setting    | CPU    | Memory | GOMAXPROCS | nice | ionice  |
|------------|--------|--------|------------|------|---------|
| `low`      | 30%    | 30%    | 2          | 19   | class 3 |
| `medium`   | 50%    | 50%    | 4          | 10   | class 2 |
| `high`     | 70%    | 70%    | 6          | 5    | class 2 |

```bash
CI_RESOURCE_LIMIT=medium make ci-all
```

Default is `low` (30-40% of host resources) per project constitution.

You can also set explicit container limits:

```bash
CI_CPUS=4 CI_MEMORY=8G make ci-go
```

## Phase 1: Go CI

**Container:** `helixagent-ci-go` (golang:1.24-alpine)

**Integration services started:** PostgreSQL 15, Redis 7, Mock LLM Server, OAuth Mock, ChromaDB, Qdrant, Kafka, RabbitMQ, MinIO

**Pipeline:**
1. Code quality gates (fmt, vet, lint, gosec)
2. Unit tests with coverage
3. Integration tests
4. E2E tests
5. Security tests
6. Stress tests
7. Benchmarks
8. Race detection
9. Build all 7 apps x 5 platforms
10. Challenge scripts
11. Artifact validation
12. False positive checks

**Output:**
- `releases/<app>/<os>-<arch>/<version>/` - Built binaries with build-info.json
- `reports/go/` - JUnit XML, coverage HTML, lint/security JSON

## Phase 2: Mobile CI

**Containers:** `helixagent-ci-mobile` (Ubuntu 22.04) + `helixagent-ci-emulator`

**Requirements:** `/dev/kvm` for Android emulator

**Pipeline:**
1. Flutter: tests, APK build (signed), AAB build, Robolectric, emulator install
2. React Native: Jest tests, APK build (signed), Robolectric, emulator install
3. Native SDK validation: Kotlin compile check, Swift syntax check

**Output:**
- `releases/mobile/flutter/` - Signed APK and AAB
- `releases/mobile/react-native/` - Signed APK
- `reports/mobile/` - Test results, coverage, signing verification

## Phase 3: Web CI

**Container:** `helixagent-ci-web` (node:20-bookworm with Playwright browsers)

**Pipeline:**
1. Angular: Karma unit tests, production build, Playwright E2E, Lighthouse audit
2. Static website: PostCSS/UglifyJS build, Playwright tests, Lighthouse audit
3. JS SDK: Jest tests, CJS/ESM/browser builds, bundle size check

**Output:**
- `releases/web/angular/dist/` - Production Angular build
- `releases/web/website/` - Minified website
- `releases/web/sdk/dist/` - SDK bundles
- `reports/web/` - Test results, coverage, Lighthouse reports

## Phase 4: Desktop CI

**Container:** `helixagent-ci-desktop` (node:20-bookworm with Rust, Electron, Tauri tooling)

**Pipeline:**
1. Electron: Jest unit tests, Playwright E2E, production build, optional signing
2. Tauri: Jest unit tests, production build, optional signing

**Output:**
- `releases/desktop/electron/dist/` - Electron build output
- `releases/desktop/tauri/` - Tauri executables (.app, .exe, .AppImage)
- `reports/desktop/` - Test results, coverage, validation logs

## Phase 5: Integration CI

**Container:** `helixagent-ci-go` (shared with Phase 1)

**Pipeline:**
1. Wait for all integration services (PostgreSQL, Redis, ChromaDB, Qdrant, Kafka, RabbitMQ, MinIO)
2. Integration test suites (`./tests/integration/`)
3. Cross-service orchestration tests (`./tests/orchestration/`)
4. End-to-end workflow tests (`./tests/workflow/`)
5. Performance under load tests (`./tests/load/`)

**Output:**
- `reports/integration/` - Test results, coverage, performance metrics

## Android Signing

**Default (debug) keystore** at `keys/android/debug.keystore`:
- Alias: `helixagent-debug`
- Passwords: `helixagent-debug`
- Suitable for device installation and testing

**Release signing** via environment variables:

```bash
CI_ANDROID_RELEASE_KEYSTORE=/path/to/release.keystore \
CI_KEYSTORE_PASSWORD=secret \
CI_KEY_ALIAS=release-alias \
CI_KEY_PASSWORD=secret \
make ci-mobile
```

For iOS, desktop, and web extension signing, see `keys/README.md`.

## Reports

After `make ci-all` or `make ci-report`:

- **`reports/summary.html`** - Interactive dashboard with all results
- **`reports/results.json`** - Machine-readable aggregate data
- **`reports/<phase>/`** - Per-phase JUnit XML, coverage, logs

### Results JSON structure

```json
{
  "timestamp": "2026-03-08T...",
  "git": { "commit": "abc123", "branch": "main" },
  "resource_limit": "low",
  "phases": {
    "go": { "status": "pass", "tests": { "total": 500, "failed": 0 }, ... },
    "mobile": { ... },
    "web": { ... },
    "desktop": { ... },
    "integration": { ... }
  },
  "totals": { "tests_total": 800, "tests_failed": 0, "coverage_avg": 95.5 },
  "false_positive_checks": [ ... ],
  "signing": { "flutter_apk": "verified", ... }
}
```

## False Positive Prevention

Every phase validates:
1. **Exit codes** - `set -euo pipefail`, PIPESTATUS checks
2. **Test counts** - Rejects if fewer tests than threshold (ci/thresholds.json)
3. **Coverage** - Rejects if below configured minimum
4. **Artifacts** - Verifies binaries exist, are non-zero, correct architecture
5. **Integration liveness** - Real HTTP requests to services
6. **Report cross-validation** - Compares JUnit XML vs stdout counts

Results in `reports/<phase>/false-positive-checks.json`.

## Troubleshooting

### "No container runtime found"
Install Docker or Podman and ensure it's in PATH.

### KVM not available (Phase 2)
```bash
# Check KVM support
ls -la /dev/kvm

# If missing, load module
sudo modprobe kvm kvm_intel  # or kvm_amd
```

Without KVM, skip Phase 2: `make ci-go && make ci-web && make ci-report`

### Build cache issues
```bash
make ci-clean  # Remove all CI volumes and containers
```

### Slow builds
Use higher resource limits: `CI_RESOURCE_LIMIT=high make ci-all`

### npm/network failures in Podman
Podman rootless bridge networks cannot reach external registries (npm, pub.dev, etc.). The compose file uses `network_mode: host` for web, mobile, emulator, and reporter containers. Go CI retains bridge networking for integration service access (postgres, redis, etc.).

If you still have issues:
```bash
# Test connectivity
podman run --rm node:20-bookworm npm ping

# Build images with host network
podman build --network=host -f docker/ci/Dockerfile.ci-web -t helixagent_ci-web:latest .

# Or configure Podman DNS
echo "[containers]" >> ~/.config/containers/containers.conf
echo "dns_servers = [\"8.8.8.8\", \"1.1.1.1\"]" >> ~/.config/containers/containers.conf
```

### Flutter tar ownership errors in Podman (Phase 2)
Flutter's gradle wrapper extraction fails in Podman rootless due to `tar` ownership changes. The Dockerfile sets `ENV TAR_OPTIONS="--no-same-owner"` to resolve this.

### Android emulator multi-device errors (Phase 2)
With host networking, ADB may see multiple devices. The emulator start script uses `adb -s emulator-5554` to target the specific emulator instance.

### Windows cross-compilation failures
Windows builds fail due to `syscall.Statfs_t` not being available. This is a known Go limitation. The CI excludes Windows builds by default.

## Adding New Components

### New Go app
1. Add entry to `scripts/build/version-manager.sh` APP_REGISTRY
2. The CI pipeline auto-discovers apps from the registry

### New mobile app
1. Add build steps to `ci/scripts/ci-mobile.sh`
2. Update thresholds in `ci/thresholds.json`

### New web app
1. Add build steps to `ci/scripts/ci-web.sh`
2. Update thresholds in `ci/thresholds.json`

### New desktop app
1. Add build steps to `ci/scripts/ci-desktop.sh`
2. Update thresholds in `ci/thresholds.json`

### New integration tests
1. Add test directories under `tests/integration/`, `tests/orchestration/`, `tests/workflow/`, or `tests/load/`
2. Update thresholds in `ci/thresholds.json``

## File Reference

| File | Purpose |
|------|---------|
| `docker-compose.ci.yml` | CI compose with all profiles and services |
| `docker/ci/Dockerfile.ci-go` | Go CI builder image |
| `docker/ci/Dockerfile.ci-mobile` | Mobile CI builder image |
| `docker/ci/Dockerfile.ci-emulator` | Android emulator image |
| `docker/ci/Dockerfile.ci-web` | Web CI builder image |
| `docker/ci/Dockerfile.ci-desktop` | Desktop CI builder image (Electron/Tauri) |
| `docker/ci/Dockerfile.ci-reporter` | Report aggregator image |
| `ci/scripts/ci-entrypoint.sh` | Resource control entrypoint |
| `ci/scripts/ci-go.sh` | Phase 1 pipeline |
| `ci/scripts/ci-mobile.sh` | Phase 2 pipeline |
| `ci/scripts/ci-web.sh` | Phase 3 pipeline |
| `ci/scripts/ci-desktop.sh` | Phase 4 pipeline |
| `ci/scripts/ci-integration.sh` | Phase 5 pipeline |
| `ci/scripts/ci-report.sh` | Report orchestrator |
| `ci/scripts/wait-for-services.sh` | Service health gate |
| `ci/scripts/false-positive-check.sh` | FP validation framework |
| `ci/scripts/validate-artifacts.sh` | Artifact integrity checks |
| `ci/thresholds.json` | Test count and coverage thresholds |
| `ci/reporter/` | Node.js report aggregator |
| `keys/android/debug.keystore` | Default Android signing key |
