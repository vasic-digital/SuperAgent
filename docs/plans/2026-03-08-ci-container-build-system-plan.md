# CI/CD Container Build System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a three-phase CI/CD system where all builds, tests, and artifact generation run inside Docker/Podman containers, producing signed artifacts and comprehensive reports.

**Architecture:** Docker Compose profiles (`go-ci`, `mobile-ci`, `web-ci`) in a single `docker-compose.ci.yml`. Each profile brings up its builder container plus required integration services. A reporter container aggregates results. Resource limits configurable via `CI_RESOURCE_LIMIT` env var.

**Tech Stack:** Docker/Podman, Go 1.24, Android SDK 34, Flutter, React Native, Node.js 20, Angular 17, Playwright, Lighthouse CI, Robolectric, gotestsum, Karma/Jasmine

**Design Doc:** `docs/plans/2026-03-08-ci-container-build-system-design.md`

---

## Task 1: Create directory structure and signing keys

**Files:**
- Create: `ci/scripts/` directory
- Create: `ci/reporter/` directory
- Create: `ci/thresholds.json`
- Create: `docker/ci/` directory
- Create: `keys/android/debug.keystore`
- Create: `keys/android/README.md`
- Modify: `.gitignore`

**Step 1: Create directories**

```bash
mkdir -p ci/scripts ci/reporter docker/ci keys/android
```

**Step 2: Generate Android debug keystore**

```bash
keytool -genkeypair -v \
  -keystore keys/android/debug.keystore \
  -alias helixagent-debug \
  -keyalg RSA -keysize 2048 \
  -validity 10000 \
  -storepass helixagent-debug \
  -keypass helixagent-debug \
  -dname "CN=HelixAgent Debug, O=HelixAgent, L=Dev, C=US"
```

**Step 3: Create keys/android/README.md**

Document keystore credentials (alias: helixagent-debug, passwords: helixagent-debug, RSA 2048, 10000 day validity). Document release signing via CI_ANDROID_RELEASE_KEYSTORE env var mount.

**Step 4: Create ci/thresholds.json**

Minimum test counts and coverage gates per phase:
- go: min 100 unit, 20 integration, 10 e2e, 5 security, 100% coverage
- mobile: flutter min 5 unit/3 widget, RN min 5 jest, robolectric min 5, 100% coverage
- web: angular min 5 karma/3 e2e/80 lighthouse, website min 3 playwright/80 lighthouse, sdk min 5 jest/100% coverage

**Step 5: Update .gitignore with CI report exclusions**

**Step 6: Commit**

```bash
git add ci/ docker/ci/ keys/android/ .gitignore
git commit -m "chore(ci): create directory structure, signing keys, and thresholds"
```

---

## Task 2: Create resource control entrypoint script

**Files:**
- Create: `ci/scripts/ci-entrypoint.sh`

Detects host CPU/memory, applies limits based on CI_RESOURCE_LIMIT (low=30%/GOMAXPROCS=2/nice19, medium=50%/GOMAXPROCS=4/nice10, high=70%/GOMAXPROCS=6/nice5). Exports CI_ALLOWED_CPUS, CI_ALLOWED_MEM_MB, GOFLAGS. Exec's the phase script passed as argument wrapped in nice/ionice.

**Commit:** `feat(ci): add resource control entrypoint script`

---

## Task 3: Create service health check wait script

**Files:**
- Create: `ci/scripts/wait-for-services.sh`

Usage: `wait-for-services.sh postgres:5432 redis:6379 mockllm:8090`

Supports TCP checks (nc -z) and HTTP checks (curl). Service-specific logic: PostgreSQL (pg_isready), Redis (redis-cli ping), Mock LLM (/health), ChromaDB (/api/v1/heartbeat), Qdrant (/healthz), MinIO (/minio/health/ready), Android emulator (adb connect). Configurable timeout via CI_SERVICE_TIMEOUT (default 120s).

**Commit:** `feat(ci): add service health check wait script`

---

## Task 4: Create false positive validation script

**Files:**
- Create: `ci/scripts/false-positive-check.sh`

6-layer validation: exit code enforcement (set -euo pipefail), test count validation (parse JUnit XML, reject 0 tests), coverage gate (parse percentages, reject below threshold), artifact validation (file exists, non-zero, correct arch), integration liveness (real HTTP requests), report cross-validation (compare XML vs stdout counts). Outputs JSON to reports/<phase>/false-positive-checks.json.

**Commit:** `feat(ci): add false positive validation framework`

---

## Task 5: Create artifact validation script

**Files:**
- Create: `ci/scripts/validate-artifacts.sh`

Phase-specific validation:
- go: 7 apps x 5 platforms — check binary exists, non-zero, build-info.json valid, native binary runs --version
- mobile: APK/AAB exist, apksigner verify, emulator install check
- web: Angular dist/index.html, website build, SDK dist bundles

**Commit:** `feat(ci): add artifact validation script`

---

## Task 6: Create Phase 1 Go CI Dockerfile

**Files:**
- Create: `docker/ci/Dockerfile.ci-go`

Based on golang:1.24-alpine. Installs: bash, make, git, gcc, musl-dev, curl, jq, netcat, postgresql15-client, redis, gotestsum, golangci-lint, gosec, goimports. Copies CI scripts. Entrypoint: ci-entrypoint.sh.

**Commit:** `feat(ci): add Go CI builder Dockerfile`

---

## Task 7: Create Phase 1 Go CI pipeline script

**Files:**
- Create: `ci/scripts/ci-go.sh`

Pipeline steps:
1. Wait for services (postgres, redis, mockllm, oauthmock)
2. Code quality (fmt, vet, lint, gosec)
3. Unit tests (gotestsum → JUnit XML + coverage)
4. Integration tests
5. E2E tests
6. Security tests
7. Stress tests
8. Benchmarks
9. Race detection
10. Build all 7 apps x 5 platforms (using version-manager.sh, go build with ldflags)
11. Challenge scripts
12. Artifact validation
13. False positive checks

All output to reports/go/. All binaries to releases/.

**Commit:** `feat(ci): add Phase 1 Go CI pipeline script`

---

## Task 8: Create Phase 2 Mobile CI Dockerfile

**Files:**
- Create: `docker/ci/Dockerfile.ci-mobile`

Based on ubuntu:22.04. Installs: Android SDK 34, build-tools, platform-tools, NDK 26, Flutter SDK (stable), Node.js 20, OpenJDK 17, Kotlin compiler, ADB. Copies CI scripts.

**Commit:** `feat(ci): add Mobile CI builder Dockerfile`

---

## Task 9: Create Android emulator container

**Files:**
- Create: `docker/ci/Dockerfile.ci-emulator`
- Create: `docker/ci/emulator-start.sh`

Based on ubuntu:22.04. Installs Android SDK + emulator + system-images;android-34;google_apis;x86_64. Creates AVD "ci-device" (pixel_6). Emulator startup: headless, no-audio, no-boot-anim, swiftshader GPU, wipe-data. Enables ADB over TCP on port 5555. Health check: `adb shell getprop sys.boot_completed`. Requires /dev/kvm passthrough.

**Commit:** `feat(ci): add Android emulator container`

---

## Task 10: Create Phase 2 Mobile CI pipeline script

**Files:**
- Create: `ci/scripts/ci-mobile.sh`

Pipeline:
1. Wait for emulator, ADB connect
2. Flutter: pub get, test (coverage), build apk --release (signed), build appbundle, Robolectric (gradlew testReleaseUnitTest), install on emulator
3. React Native: npm ci, jest (coverage), gradlew assembleRelease (signed), Robolectric, install on emulator
4. Native SDKs: kotlinc compile check, swiftc syntax check (skip on Linux)
5. Artifact validation, signing verification

**Commit:** `feat(ci): add Phase 2 Mobile CI pipeline script`

---

## Task 11: Create Phase 3 Web CI Dockerfile

**Files:**
- Create: `docker/ci/Dockerfile.ci-web`

Based on node:20-bookworm. Installs: Angular CLI 17, Playwright (chromium/firefox/webkit), Lighthouse CI, http-server, Python 3, Playwright system deps. Copies CI scripts.

**Commit:** `feat(ci): add Web CI builder Dockerfile`

---

## Task 12: Create Phase 3 Web CI pipeline script

**Files:**
- Create: `ci/scripts/ci-web.sh`

Pipeline:
1. Angular: npm ci, ng test (Karma/ChromeHeadless/coverage), ng build --production, Playwright E2E, Lighthouse audit
2. Website: npm ci, postcss+uglifyjs build, Playwright (homepage load, console errors, responsive), Lighthouse audit
3. JS SDK: npm ci, jest (coverage), npm run build (CJS/ESM/browser), bundle size check
4. Artifact validation

**Commit:** `feat(ci): add Phase 3 Web CI pipeline script`

---

## Task 13: Create report aggregator

**Files:**
- Create: `ci/reporter/package.json` (deps: fast-xml-parser, glob)
- Create: `ci/reporter/aggregate.js` (parses JUnit XML, coverage files, build-info.json, false-positive-checks.json; generates results.json and summary.html)
- Create: `ci/reporter/dashboard-template.html` (dark theme dashboard with cards, tables, phase results, FP checks, raw JSON)
- Create: `docker/ci/Dockerfile.ci-reporter` (node:20-alpine, copies reporter, npm ci)

Uses `execFileSync` (not `exec`) for git metadata. Calculates totals across phases. Exits non-zero if any failures.

**Commit:** `feat(ci): add report aggregator with HTML dashboard`

---

## Task 14: Create docker-compose.ci.yml

**Files:**
- Create: `docker-compose.ci.yml`

Profiles: go-ci, mobile-ci, web-ci, report, infra.

Integration services (go-ci + infra profiles): postgres (15-alpine), redis (7-alpine), mockllm (built from tests/mock-llm-server), oauthmock (navikt/mock-oauth2-server), chromadb, qdrant, kafka (bitnami, KRaft mode), rabbitmq (3-management-alpine), minio. All with health checks on ci-network.

CI containers: ci-go (go-ci profile, depends on postgres+redis+mockllm healthy), ci-mobile (mobile-ci profile, depends on emulator healthy), emulator (mobile-ci, /dev/kvm device), ci-web (web-ci, no deps), ci-reporter (report profile).

Volumes: go-mod-cache, go-build-cache, gradle-cache, flutter-cache, npm-cache, npm-cache-web.

Resource limits via CI_CPUS/CI_MEMORY env vars.

**Commit:** `feat(ci): add docker-compose.ci.yml with all phases and integration services`

---

## Task 15: Create report orchestrator script

**Files:**
- Create: `ci/scripts/ci-report.sh`

Runs aggregate.js, prints report paths, validates results.json for 0 failures.

**Commit:** `feat(ci): add report aggregation orchestrator`

---

## Task 16: Add Makefile CI targets

**Files:**
- Modify: `Makefile`

Add targets: ci-all, ci-go, ci-mobile, ci-web, ci-report, ci-build-images, ci-clean. Auto-detect docker compose vs podman-compose. Each phase: compose up --build --abort-on-container-exit --exit-code-from, then down. ci-all chains: ci-go, ci-mobile, ci-web, ci-report.

**Commit:** `feat(ci): add Makefile CI targets`

---

## Task 17: Create CI Build Guide documentation

**Files:**
- Create: `docs/CI_BUILD_GUIDE.md`

Covers: overview, prerequisites, quick start, phase-by-phase usage, resource config, signing config, report interpretation, troubleshooting, adding new apps/tests.

**Commit:** `docs(ci): add comprehensive CI Build Guide`

---

## Task 18: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md` — add CI/CD Container Build System section
- Modify: `.gitignore` — ensure CI reports excluded

**Commit:** `docs(ci): update CLAUDE.md with CI system references`

---

## Tasks 19-22: Integration verification

**Task 19:** Build CI images (`make ci-build-images`), run Phase 1 (`make ci-go`), verify Go artifacts and reports.

**Task 20:** Verify /dev/kvm, run Phase 2 (`make ci-mobile`), verify APKs signed and installed on emulator.

**Task 21:** Run Phase 3 (`make ci-web`), verify Angular/website/SDK builds and Lighthouse reports.

**Task 22:** Run full pipeline (`make ci-all`), verify reports/results.json shows 0 failures, summary.html renders correctly, all artifacts in releases/.

---

## Execution Order

Tasks 1-18: implementation (sequential).
Tasks 19-22: verification (sequential, each phase then full).

Key dependencies:
- Tasks 2-5 need Task 1 (directories)
- Tasks 6-12 need Tasks 2-5 (scripts for COPY in Dockerfiles)
- Task 14 needs Tasks 6-13 (all Dockerfiles)
- Task 16 needs Task 14 (compose file)
- Tasks 19-22 need Task 18 (everything ready)
