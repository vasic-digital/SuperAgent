# HelixQA API Reference

**Version**: 1.0
**Base URL**: `http://localhost:7061`
**Last Updated**: 2026-03-30

Complete reference for the HelixQA autonomous QA pipeline endpoints exposed
under `/v1/qa`. These endpoints launch and manage autonomous testing sessions,
query findings stored in the SQLite memory store, and inspect supported
platforms and project knowledge.

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Error Responses](#error-responses)
4. [POST /v1/qa/sessions](#post-v1qasessions)
5. [GET /v1/qa/findings](#get-v1qafindings)
6. [GET /v1/qa/findings/:id](#get-v1qafindingsid)
7. [PUT /v1/qa/findings/:id](#put-v1qafindingsid)
8. [GET /v1/qa/platforms](#get-v1qaplatforms)
9. [POST /v1/qa/discover](#post-v1qadiscover)
10. [Data Types](#data-types)

---

## Overview

The HelixQA subsystem provides an autonomous, multi-phase QA pipeline driven by
LLM vision analysis. A session proceeds through four stages:

1. **Learn** — reads project docs, CLAUDE.md files, and `.env` credentials.
2. **Plan** — generates a test plan for the target platforms.
3. **Execute** — drives browsers, Android emulators, or CLI tools; captures
   screenshots and crash evidence.
4. **Analyze** — classifies findings by severity, writes tickets, measures
   coverage.

All persistent data (findings, session metadata) is stored in a per-session
SQLite database at `OutputDir/memory.db` unless `memory_db_path` is overridden.

> **Service availability**: all endpoints return `503 Service Unavailable` when
> the HelixQA adapter has not been initialised (e.g. no LLM API key is present
> in the environment).

---

## Authentication

Endpoints require a Bearer token identical to the main HelixAgent API key:

```
Authorization: Bearer YOUR_API_KEY
```

---

## Error Responses

All error responses share the following shape:

```json
{
  "error": "<machine-readable code or message>"
}
```

| HTTP Status | Meaning |
|-------------|---------|
| `400 Bad Request` | Missing or malformed request body / required field |
| `404 Not Found` | Finding ID does not exist |
| `500 Internal Server Error` | Pipeline or store error |
| `503 Service Unavailable` | HelixQA adapter not initialised |

---

## POST /v1/qa/sessions

Start an autonomous QA session. The session runs synchronously through all four
pipeline phases (Learn → Plan → Execute → Analyze) and returns a result summary.
Long-running sessions should be called with an appropriate HTTP timeout.

### Request Body

`Content-Type: application/json`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project_root` | `string` | **yes** | Absolute path to the project root. Used for knowledge discovery and credential extraction. |
| `platforms` | `[]string` | **yes** | One or more target platforms. See [GET /v1/qa/platforms](#get-v1qaplatforms) for valid values. |
| `output_dir` | `string` | no | Directory for screenshots, reports, and the memory DB. Defaults to `"qa-results"`. |
| `issues_dir` | `string` | no | Directory for generated issue tickets. Defaults to `"qa-issues"`. |
| `android_device` | `string` | no | ADB device serial for single-device Android testing. |
| `android_devices` | `[]string` | no | ADB device serials for parallel multi-device Android testing. |
| `android_package` | `string` | no | Android application package name (e.g. `"com.example.app"`). Required when platform is `android` or `android_tv`. |
| `web_url` | `string` | no | Base URL for web platform testing (e.g. `"http://localhost:3000"`). Required when platform is `web`. |
| `curiosity_enabled` | `bool` | no | Enable curiosity-driven exploration that follows unexpected UI paths. Default `false`. |
| `vision_host` | `string` | no | Override the LLM vision provider host. |
| `vision_user` | `string` | no | Username for the vision provider (used with llama.cpp remotes). |
| `vision_model` | `string` | no | Model name for the vision provider (e.g. `"llava:13b"`). |
| `memory_db_path` | `string` | no | Explicit path to the SQLite memory DB. Overrides the default `output_dir/memory.db`. |

**Example request:**

```json
{
  "project_root": "/home/user/myapp",
  "platforms": ["android", "web"],
  "output_dir": "/tmp/qa-run-01",
  "android_device": "emulator-5554",
  "android_package": "com.example.myapp",
  "web_url": "http://localhost:3000",
  "curiosity_enabled": true,
  "vision_model": "gpt-4o"
}
```

### Response — 202 Accepted

The session completed (or finished with a partial error). The `status` field
indicates the outcome; `error` is populated only on failure.

| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | Session outcome: `"completed"` or `"failed"`. |
| `session_id` | `string` | Unique identifier for this session (UUID). |
| `duration` | `int64` | Wall-clock duration in nanoseconds (`time.Duration`). |
| `tests_planned` | `int` | Number of test cases generated during the Plan phase. |
| `tests_run` | `int` | Number of test cases actually executed. |
| `issues_found` | `int` | Total findings recorded, regardless of severity. |
| `tickets_created` | `int` | Number of issue ticket files written to `issues_dir`. |
| `coverage_pct` | `float64` | Estimated UI coverage percentage (0–100). |
| `error` | `string` | Error message if `status` is `"failed"`. Omitted on success. |

**Example response:**

```json
{
  "status": "completed",
  "session_id": "a3f1b2c4-dead-beef-cafe-000000000001",
  "duration": 184000000000,
  "tests_planned": 42,
  "tests_run": 38,
  "issues_found": 5,
  "tickets_created": 5,
  "coverage_pct": 74.3
}
```

**Partial failure example** (session finished but with errors):

```json
{
  "status": "failed",
  "session_id": "a3f1b2c4-dead-beef-cafe-000000000002",
  "duration": 12000000000,
  "tests_planned": 10,
  "tests_run": 3,
  "issues_found": 0,
  "tickets_created": 0,
  "coverage_pct": 0,
  "error": "android device emulator-5554 not found"
}
```

### Error Responses

| Status | `error` value | Trigger |
|--------|---------------|---------|
| `400` | binding error message | `project_root` or `platforms` missing |
| `500` | pipeline error message | Internal pipeline failure with no partial result |
| `503` | `"HelixQA adapter not initialized"` | Adapter unavailable |

### Example curl

```bash
curl -s -X POST http://localhost:7061/v1/qa/sessions \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "project_root": "/home/user/myapp",
    "platforms": ["web"],
    "web_url": "http://localhost:3000",
    "output_dir": "/tmp/qa-run-01"
  }' | jq .
```

---

## GET /v1/qa/findings

List findings recorded by previous QA sessions, optionally filtered by status.

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `status` | `string` | no | Filter findings by status. Valid values: `open`, `fixed`, `verified`. Defaults to `open` when omitted. |

### Response — 200 OK

| Field | Type | Description |
|-------|------|-------------|
| `findings` | `[]Finding` | Array of finding objects. Empty array when none match. |
| `total` | `int` | Number of findings returned. |

**Example response:**

```json
{
  "findings": [
    {
      "id": "HELIX-001",
      "session_id": "a3f1b2c4-dead-beef-cafe-000000000001",
      "severity": "critical",
      "category": "crash",
      "title": "App crashes on login with empty password",
      "description": "Tapping 'Login' with an empty password field causes a NullPointerException.",
      "repro_steps": "1. Open app\n2. Leave password blank\n3. Tap Login",
      "evidence_paths": "/tmp/qa-run-01/screenshots/login_crash.png",
      "platform": "android",
      "screen": "LoginScreen",
      "status": "open",
      "found_date": "2026-03-30T14:22:01Z"
    }
  ],
  "total": 1
}
```

### Error Responses

| Status | Trigger |
|--------|---------|
| `500` | Store read failure |
| `503` | Adapter unavailable |

### Example curl

```bash
# All open findings (default)
curl -s http://localhost:7061/v1/qa/findings \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" | jq .

# Fixed findings only
curl -s "http://localhost:7061/v1/qa/findings?status=fixed" \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" | jq .
```

---

## GET /v1/qa/findings/:id

Retrieve a single finding by its ID.

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | `string` | Finding identifier (e.g. `HELIX-001`). |

### Response — 200 OK

Returns a `Finding` object.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Unique finding identifier (e.g. `HELIX-001`). |
| `session_id` | `string` | ID of the session that produced this finding. |
| `severity` | `string` | Severity level: `critical`, `high`, `medium`, `low`, or `info`. |
| `category` | `string` | Finding category: `crash`, `ui`, `performance`, `security`, `functional`, or `accessibility`. |
| `title` | `string` | Short human-readable summary. |
| `description` | `string` | Full description of the defect. |
| `repro_steps` | `string` | Step-by-step reproduction instructions. |
| `evidence_paths` | `string` | Comma-separated paths to screenshot or video evidence. |
| `platform` | `string` | Platform where the finding was observed. |
| `screen` | `string` | Screen or page name where the finding occurred. |
| `status` | `string` | Current lifecycle status: `open`, `fixed`, or `verified`. |
| `found_date` | `string` | RFC 3339 timestamp when the finding was recorded. |

**Example response:**

```json
{
  "id": "HELIX-001",
  "session_id": "a3f1b2c4-dead-beef-cafe-000000000001",
  "severity": "critical",
  "category": "crash",
  "title": "App crashes on login with empty password",
  "description": "Tapping 'Login' with an empty password field causes a NullPointerException in AuthViewModel.validate().",
  "repro_steps": "1. Open app\n2. Leave password blank\n3. Tap Login",
  "evidence_paths": "/tmp/qa-run-01/screenshots/login_crash.png,/tmp/qa-run-01/videos/login_crash.mp4",
  "platform": "android",
  "screen": "LoginScreen",
  "status": "open",
  "found_date": "2026-03-30T14:22:01Z"
}
```

### Error Responses

| Status | Trigger |
|--------|---------|
| `404` | Finding ID not found in the store |
| `503` | Adapter unavailable |

### Example curl

```bash
curl -s http://localhost:7061/v1/qa/findings/HELIX-001 \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" | jq .
```

---

## PUT /v1/qa/findings/:id

Update the lifecycle status of a finding (e.g. mark it as fixed after a
developer resolves the underlying defect).

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | `string` | Finding identifier (e.g. `HELIX-001`). |

### Request Body

`Content-Type: application/json`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | `string` | **yes** | New status. Valid values: `open`, `fixed`, `verified`. |

**Example request:**

```json
{
  "status": "fixed"
}
```

### Response — 200 OK

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Finding identifier that was updated. |
| `status` | `string` | The new status value as persisted. |

**Example response:**

```json
{
  "id": "HELIX-001",
  "status": "fixed"
}
```

### Error Responses

| Status | Trigger |
|--------|---------|
| `400` | `status` field missing from request body |
| `500` | Store write failure |
| `503` | Adapter unavailable |

### Example curl

```bash
curl -s -X PUT http://localhost:7061/v1/qa/findings/HELIX-001 \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"status": "fixed"}' | jq .
```

---

## GET /v1/qa/platforms

List the platform identifiers supported by the HelixQA pipeline.

### Response — 200 OK

| Field | Type | Description |
|-------|------|-------------|
| `platforms` | `[]string` | Array of supported platform names. |
| `total` | `int` | Number of supported platforms. |

**Example response:**

```json
{
  "platforms": [
    "android",
    "android_tv",
    "web",
    "desktop",
    "cli",
    "api"
  ],
  "total": 6
}
```

### Platform Descriptions

| Platform | Description |
|----------|-------------|
| `android` | Android mobile application, driven via ADB and UI Automator. |
| `android_tv` | Android TV application, driven via ADB with D-pad navigation. |
| `web` | Web application, driven via headless Chromium/Playwright. |
| `desktop` | Desktop application (Linux/macOS), driven via accessibility APIs. |
| `cli` | Command-line application, driven by spawning child processes and analysing stdout/stderr. |
| `api` | REST/gRPC API surface, driven by HTTP request generation and response validation. |

### Error Responses

| Status | Trigger |
|--------|---------|
| `503` | Adapter unavailable |

### Example curl

```bash
curl -s http://localhost:7061/v1/qa/platforms \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" | jq .
```

---

## POST /v1/qa/discover

Scan a project directory for documentation, CLAUDE.md files, and environment
configuration to build a knowledge base used during the Learn phase. Credential
values discovered in `.env` files are **redacted** in the response (replaced
with `"***"`); only the credential keys are exposed.

### Request Body

`Content-Type: application/json`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project_root` | `string` | **yes** | Absolute path to the project root to scan. |

**Example request:**

```json
{
  "project_root": "/home/user/myapp"
}
```

### Response — 200 OK

| Field | Type | Description |
|-------|------|-------------|
| `docs_count` | `int` | Number of documentation files discovered (README, CLAUDE.md, etc.). |
| `constraints_count` | `int` | Number of constraint rules extracted from documentation. |
| `credentials` | `map[string]string` | Map of credential keys to `"***"`. Values are always redacted. Omitted when no credentials are found. |

**Example response:**

```json
{
  "docs_count": 14,
  "constraints_count": 26,
  "credentials": {
    "ANDROID_TEST_USER": "***",
    "ANDROID_TEST_PASSWORD": "***",
    "WEB_TEST_USER": "***"
  }
}
```

**Example response (no credentials found):**

```json
{
  "docs_count": 3,
  "constraints_count": 5
}
```

### Error Responses

| Status | Trigger |
|--------|---------|
| `400` | `project_root` field missing |
| `500` | Directory read failure |
| `503` | Adapter unavailable |

### Example curl

```bash
curl -s -X POST http://localhost:7061/v1/qa/discover \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"project_root": "/home/user/myapp"}' | jq .
```

---

## Data Types

### SessionResult

Returned by [POST /v1/qa/sessions](#post-v1qasessions).

```json
{
  "status": "completed",
  "session_id": "a3f1b2c4-dead-beef-cafe-000000000001",
  "duration": 184000000000,
  "tests_planned": 42,
  "tests_run": 38,
  "issues_found": 5,
  "tickets_created": 5,
  "coverage_pct": 74.3,
  "error": ""
}
```

**`status` values:**

| Value | Description |
|-------|-------------|
| `pending` | Session created but not yet started (future async mode). |
| `running` | Session is actively executing (future async mode). |
| `completed` | All pipeline phases finished without fatal error. |
| `failed` | Session terminated early due to an unrecoverable error. |

### Finding

Returned by [GET /v1/qa/findings](#get-v1qafindings) and
[GET /v1/qa/findings/:id](#get-v1qafindingsid).

```json
{
  "id": "HELIX-001",
  "session_id": "a3f1b2c4-dead-beef-cafe-000000000001",
  "severity": "critical",
  "category": "crash",
  "title": "App crashes on login with empty password",
  "description": "Full description of the defect.",
  "repro_steps": "Step-by-step instructions.",
  "evidence_paths": "/tmp/qa-run-01/screenshots/crash.png",
  "platform": "android",
  "screen": "LoginScreen",
  "status": "open",
  "found_date": "2026-03-30T14:22:01Z"
}
```

**`severity` values:** `critical`, `high`, `medium`, `low`, `info`

**`category` values:** `crash`, `ui`, `performance`, `security`, `functional`,
`accessibility`

**`status` values:** `open`, `fixed`, `verified`

### KnowledgeSummary

Returned by [POST /v1/qa/discover](#post-v1qadiscover).

```json
{
  "docs_count": 14,
  "constraints_count": 26,
  "credentials": {
    "ANDROID_TEST_USER": "***"
  }
}
```

---

*For the main HelixAgent API reference see [API_REFERENCE.md](API_REFERENCE.md).*
