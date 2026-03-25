# User Manual 39: HelixQA Guide

## Overview

HelixQA (`digital.vasic.helixqa`) is a QA orchestration framework for cross-platform testing with real-time crash detection, step validation, evidence collection, and automated ticket generation. It supports Android, Web, and Desktop platforms.

## Prerequisites

- Go 1.24+
- Sibling module directories: `../Challenges`, `../Containers`
- For Android testing: ADB installed, emulator or device connected
- For Web testing: Playwright or a Chromium-based browser
- For Desktop testing: X11 display available (Linux) or native window manager
- (Optional) LLM API keys for autonomous QA sessions

## Step 1: Install HelixQA

Build from source:

```bash
cd HelixQA
make build
# Binary output: bin/helixqa
```

Verify:

```bash
./bin/helixqa version
```

## Step 2: Create Test Banks in YAML

Test banks define the cases HelixQA will execute. Create a file `tests/banks/core.yaml`:

```yaml
version: "1.0"
name: "Core Feature Tests"
test_cases:
  - id: TC-001
    name: "Create new document"
    category: functional
    priority: critical
    platforms: [android, web, desktop]
    steps:
      - name: "Open app"
        action: "Launch application"
        expected: "Main editor screen visible"
      - name: "Create document"
        action: "Tap/click new document button"
        expected: "Empty document created with cursor active"
    tags: [core, smoke]
    documentation_refs:
      - type: user_guide
        section: "3.1"
        path: "docs/USER_MANUAL.md"
```

Fields: `id` (unique), `priority` (critical/high/medium/low), `platforms` (android/web/desktop), `steps` with action/expected pairs, `tags` for filtering.

## Step 3: Configure Platform Detectors

HelixQA ships 3 crash/ANR detectors:

| Detector | Platform | Mechanism |
|----------|----------|-----------|
| `android` | Android | ADB-based: `pidof`, `logcat`, `screencap` |
| `web` | Web | Browser process monitoring via `pgrep` |
| `desktop` | Desktop | JVM/process monitoring via `pgrep`, `kill` |

Configure via `.env`:

```bash
# Android
HELIXQA_ANDROID_DEVICE=emulator-5554
HELIXQA_ANDROID_PACKAGE=com.example.app

# Web
HELIXQA_WEB_URL=http://localhost:4200
HELIXQA_WEB_BROWSER=chromium

# Desktop
HELIXQA_DESKTOP_PROCESS=myapp
HELIXQA_DESKTOP_DISPLAY=:0
```

## Step 4: Run the QA Pipeline

Run tests against a specific platform:

```bash
helixqa run --banks tests/banks/ --platform android \
  --device emulator-5554 --package com.example.app
```

Run across all platforms:

```bash
helixqa run --banks tests/banks/ --platform all
```

List available test cases without executing:

```bash
helixqa list --banks tests/banks/ --platform android
```

## Step 5: Run an Autonomous QA Session

The autonomous mode uses LLM agents and VisionEngine to navigate applications without predefined steps:

```bash
cp .env.example .env
# Edit .env: set API keys and platform settings

helixqa autonomous --project /path/to/MyApp \
  --platforms android,desktop,web \
  --env .env \
  --timeout 2h \
  --coverage-target 0.9 \
  --output qa-results/ \
  --report markdown,html,json
```

Autonomous sessions run 4 phases:
1. **Setup** -- Select LLMs via LLMsVerifier, build feature map via DocProcessor, spawn agents via LLMOrchestrator
2. **Doc-Driven Verification** -- Verify every documented feature against the running app
3. **Curiosity-Driven Exploration** -- Explore undiscovered areas, test edge cases
4. **Report and Cleanup** -- Aggregate coverage, tickets, and navigation maps

## Step 6: Collect Evidence and Review Results

Evidence is captured at each test step:

```bash
ls qa-results/
# qa-report.md     -- Full report
# tickets/         -- Auto-generated issue tickets (Markdown)
# videos/          -- Screen recordings
# screenshots/     -- Per-step captures
# logcat/          -- Android logcat dumps
```

Each ticket includes: reproduction steps, evidence links, priority, and affected platform.

## Step 7: Generate Reports

Convert existing results to different formats:

```bash
helixqa report --input qa-results --format html
helixqa report --input qa-results --format json
```

Speed modes for different contexts:

| Mode | Use Case |
|------|----------|
| `slow` | Debugging, captures maximum evidence |
| `normal` | Standard QA runs |
| `fast` | CI pipelines, minimal evidence |

## Step 8: Run Tests

```bash
make test       # 235 tests
make test-race  # With race detection
make test-cover # Coverage report
```

## Troubleshooting

- **"ADB device not found"**: Ensure `adb devices` lists your target device or emulator
- **Web tests fail to connect**: Verify the application URL is reachable and the browser binary exists
- **Empty evidence directory**: Check that the platform worker has filesystem write permissions
- **Low coverage score**: Add more test banks or enable curiosity-driven exploration in autonomous mode

## Related Resources

- [User Manual 38: DocProcessor Guide](38-docprocessor-guide.md) -- Feature map extraction
- [User Manual 40: LLMOrchestrator Guide](40-llmorchestrator-guide.md) -- Agent pool management
- [User Manual 41: VisionEngine Guide](41-visionengine-guide.md) -- Screenshot analysis and navigation graphs
- Source: `HelixQA/README.md`, `HelixQA/CLAUDE.md`
