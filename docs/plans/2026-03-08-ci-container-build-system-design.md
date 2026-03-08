# CI/CD Container Build System Design

**Date:** 2026-03-08
**Status:** Approved
**Scope:** Full CI/CD pipeline running inside Docker/Podman containers

## Overview

Three-phase CI/CD container system where ALL builds, tests, and artifact generation run inside Docker/Podman compose containers. Each phase is independently functional and produces validated, signed artifacts with comprehensive reports.

**Phases:**
- **Phase 1 (Go CI):** 7 Go binaries x 5 platforms + full test suite + integration services
- **Phase 2 (Mobile CI):** Flutter + React Native apps, Robolectric tests, Android emulator E2E
- **Phase 3 (Web CI):** Angular dashboard + static website + JS SDK, Playwright E2E, Lighthouse audits

## Architecture

```
docker-compose.ci.yml
├── Profile: go-ci        (Phase 1)
├── Profile: mobile-ci    (Phase 2)
└── Profile: web-ci       (Phase 3)
```

**Entry points:** `make ci-all` (all phases), `make ci-go`, `make ci-mobile`, `make ci-web`

**Output:**
```
releases/          <- all built artifacts
reports/           <- all test/coverage/audit reports
├── go/
├── mobile/
├── web/
├── summary.html   <- aggregated dashboard
├── results.json   <- machine-readable summary
└── false-positive-checks.json
keys/              <- default signing keystores
```

## Phase 1: Go CI Container

**Image:** `helixagent-ci-go` based on `golang:1.24-alpine`

**Installed tools:**
- Build: make, git, gcc, musl-dev
- Quality: golangci-lint, gosec, goimports
- Test: gotestsum (JUnit XML), go tool cover
- Runtime: curl, jq, bash

**Integration services** (sibling containers):
- PostgreSQL 15 (port 5432)
- Redis 7 (port 6379)
- Mock LLM Server (port 18081)
- OAuth Mock Server (port 18091)
- ChromaDB, Qdrant (vector DB)
- Kafka, RabbitMQ (messaging)
- MinIO (object storage)

**Pipeline:**
1. `make fmt vet lint security-scan` - quality gates
2. `make test-unit` - unit tests with coverage
3. `make test-integration` - integration tests (live services)
4. `make test-e2e` - end-to-end tests
5. `make test-security` - security tests
6. `make test-stress` - stress tests (resource-limited)
7. `make test-bench` - benchmarks
8. `make test-race` - race detection
9. `make release-all` - build all 7 apps x 5 platforms
10. `run_all_challenges.sh` - challenge scripts
11. Report generation

**Volumes:**
- Project root -> `/workspace` (read-write)
- `go-mod-cache` named volume
- `go-build-cache` named volume

## Phase 2: Mobile CI Container

**Image:** `helixagent-ci-mobile` based on `ubuntu:22.04`

**Installed tools:**
- Android SDK (API 34), build-tools, platform-tools, cmdline-tools
- Android NDK
- Flutter SDK (stable)
- Node.js 20 LTS + npm (React Native)
- OpenJDK 17 (Gradle, Robolectric)
- Kotlin compiler

**Android Emulator:** Separate container `helixagent-ci-emulator`
- API 34, x86_64 system image
- KVM passthrough (`/dev/kvm`)
- ADB on port 5555
- Headless, GPU swiftshader

**Pipeline:**

### Flutter app (`LLMsVerifier/llm-verifier/mobile/flutter_app/`):
- `flutter test` - unit/widget tests
- `flutter build apk --release` - signed APK
- `flutter build appbundle` - signed AAB
- `./gradlew testReleaseUnitTest` - Robolectric
- ADB install + instrumentation tests on emulator

### React Native app (`LLMsVerifier/llm-verifier/mobile/react-native/`):
- `npm test` - Jest unit tests
- `./gradlew assembleRelease` - signed APK
- `./gradlew testReleaseUnitTest` - Robolectric
- ADB install + Detox/Appium E2E on emulator

### Native SDKs:
- Kotlin SDK: `kotlinc` compilation + ktlint
- Swift SDK: syntax validation (flagged for macOS CI)

**Volumes:** `gradle-cache`, `npm-cache`, `flutter-cache` named volumes

## Phase 3: Web CI Container

**Image:** `helixagent-ci-web` based on `node:20-bookworm`

**Installed tools:**
- Angular CLI 17
- Playwright (Chromium, Firefox, WebKit)
- Lighthouse CI
- PostCSS, UglifyJS
- Python 3 (static site server)

**Pipeline:**

### Angular web app (`LLMsVerifier/llm-verifier/web/`):
- `npm ci`
- `ng test --watch=false --browsers=ChromeHeadless --code-coverage` - Karma/Jasmine
- `ng build --configuration production`
- `npx playwright test` - E2E against production build
- `lhci autorun` - Lighthouse audit

### Static website (`Website/`):
- PostCSS + UglifyJS minification
- Playwright tests (links, pages, responsive)
- Lighthouse audit

### JavaScript SDK (`sdk/web/`):
- `npm test` - Jest with coverage
- `npm run build` - CJS, ESM, browser bundles
- Bundle size check

## Signing Infrastructure

### Android (committed defaults):
- `keys/android/debug.keystore`
  - Alias: `helixagent-debug`
  - Passwords: `helixagent-debug`
  - Validity: 10000 days
- Release keystore: mounted via `CI_ANDROID_RELEASE_KEYSTORE` env var

### Gradle signing config:
```groovy
signingConfigs {
    debug {
        storeFile file(System.env.CI_KEYSTORE_PATH ?: '/workspace/keys/android/debug.keystore')
        storePassword System.env.CI_KEYSTORE_PASSWORD ?: 'helixagent-debug'
        keyAlias System.env.CI_KEY_ALIAS ?: 'helixagent-debug'
        keyPassword System.env.CI_KEY_PASSWORD ?: 'helixagent-debug'
    }
}
```

### Go binary integrity:
- `build-info.json` with SHA256 per artifact
- `--version` execution test
- `file` command architecture validation

### Future mount points (ready, not implemented):
- `CI_IOS_SIGNING_IDENTITY`, `CI_GPG_KEY`, `CI_COSIGN_KEY`

## Resource Control

Environment variable `CI_RESOURCE_LIMIT` (default: `low`):

| Setting  | CPU Limit       | Memory Limit    | GOMAXPROCS | nice | ionice  |
|----------|----------------|-----------------|------------|------|---------|
| `low`    | 30% host cores | 30% host RAM   | 2          | 19   | class 3 |
| `medium` | 50% host cores | 50% host RAM   | 4          | 10   | class 2 |
| `high`   | 70% host cores | 70% host RAM   | 6          | 5    | class 2 |

Android emulator: fixed 2 CPU, 4GB RAM regardless of setting.

## Container Networking

Single `ci-network` (bridge mode). All services addressable by name:
- `postgres:5432`, `redis:6379`, `mockllm:18081`, `oauthmock:18091`
- `chromadb:8000`, `qdrant:6333`, `kafka:9092`, `rabbitmq:5672`, `minio:9000`
- `emulator:5555` (ADB)

**Health gates:** Each phase waits for required services before starting.

## False Positive Prevention

**6 layers:**

1. **Exit code enforcement:** `set -euo pipefail`, `PIPESTATUS` checks
2. **Test count validation:** reject if 0 tests ran or below expected minimum
3. **Coverage gate:** parse reports programmatically, reject if below 100%
4. **Artifact validation:** verify exists, non-zero, correct arch, runs `--version`, `apksigner verify`
5. **Integration liveness:** real HTTP requests to each service after tests pass
6. **Report cross-validation:** compare JUnit XML counts vs stdout, coverage XML vs HTML

All results logged to `reports/false-positive-checks.json`.

## Report Aggregation

**Reporter container** (`helixagent-ci-reporter`, `node:20-alpine`) runs after all phases.

**`reports/summary.html`:**
- Build metadata (git commit, branch, timestamp, resource limit)
- Phase status table (pass/fail, duration)
- Test breakdown per category
- Coverage matrix with color coding
- Artifact inventory (name, platform, version, size, signed status)
- Lighthouse scores
- False positive validation log
- Failure details

**`reports/results.json`:**
```json
{
  "timestamp": "...",
  "git": { "commit": "...", "branch": "...", "dirty": false },
  "resource_limit": "low",
  "phases": {
    "go": { "status": "pass", "duration_s": 0, "tests": {}, "coverage": {}, "artifacts": [] },
    "mobile": { "status": "pass", "...": "..." },
    "web": { "status": "pass", "...": "..." }
  },
  "totals": { "tests_total": 0, "tests_passed": 0, "tests_failed": 0, "coverage_avg": 0 },
  "false_positive_checks": [],
  "signing": { "android_apk": "verified", "android_aab": "verified" },
  "lighthouse": { "angular": {}, "website": {} }
}
```

**Exit code:** Non-zero if ANY test failed, coverage below 100%, artifact missing, signing failed, false positive detected, or Lighthouse below threshold.

## File Structure

```
docker/ci/
├── Dockerfile.ci-go
├── Dockerfile.ci-mobile
├── Dockerfile.ci-web
├── Dockerfile.ci-reporter
├── Dockerfile.ci-emulator
docker-compose.ci.yml
keys/
├── android/
│   ├── debug.keystore
│   └── README.md
ci/
├── thresholds.json
├── scripts/
│   ├── ci-entrypoint.sh
│   ├── ci-go.sh
│   ├── ci-mobile.sh
│   ├── ci-web.sh
│   ├── ci-report.sh
│   ├── wait-for-services.sh
│   ├── validate-artifacts.sh
│   └── false-positive-check.sh
├── reporter/
│   ├── package.json
│   ├── aggregate.js
│   └── dashboard-template.html
docs/
└── CI_BUILD_GUIDE.md
```

## Makefile Targets

```makefile
ci-all              # All three phases + report
ci-go               # Phase 1 only
ci-mobile           # Phase 2 only
ci-web              # Phase 3 only
ci-report           # Aggregate reports only
ci-build-images     # Build all CI container images
ci-clean            # Remove CI containers, networks, caches
```

## Usage Examples

```bash
# Full CI, default (low) resources
make ci-all

# Go phase only, medium resources
CI_RESOURCE_LIMIT=medium make ci-go

# Full CI with release signing
CI_ANDROID_RELEASE_KEYSTORE=/path/to/release.keystore \
CI_KEYSTORE_PASSWORD=secret make ci-all

# Rebuild CI images
make ci-build-images
```
