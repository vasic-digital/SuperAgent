# HelixAgent Code Formatter System Guide

## Introduction

HelixAgent includes a comprehensive code formatting system with 32+ formatters covering 19 programming languages. The system provides a unified REST API, supports three formatter architectures (native binaries, containerized services, and language built-ins), and integrates with the AI debate pipeline for style-aware code generation. Service-based formatters run in Docker containers on ports 9210-9300, requiring zero local tool installation.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Formatter Types](#formatter-types)
3. [REST API](#rest-api)
4. [Supported Languages and Formatters](#supported-languages-and-formatters)
5. [Formatter Interface](#formatter-interface)
6. [Docker Setup for Service Formatters](#docker-setup-for-service-formatters)
7. [Configuration](#configuration)
8. [AI Debate Integration](#ai-debate-integration)
9. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

The formatter system is organized around a central registry that manages formatter instances by language. Incoming format requests are routed to the appropriate formatter based on language detection (explicit or inferred from file extension).

```
REST API (/v1/format, /v1/formatters)
        |
        v
  Formatter Registry
        |
   +----+----+----+
   |    |    |    |
   v    v    v    v
Native Service Builtin Unified
(binary) (Docker) (lang)  (multi)
```

Core files in `internal/formatters/`:

| File | Purpose |
|------|---------|
| `interface.go` | Formatter interface, request/result types |
| `registry.go` | Formatter registry and lookup |
| `executor.go` | Execution engine with timeout handling |
| `cache.go` | Result caching for repeated formats |
| `system.go` | System-level formatter management |

---

## Formatter Types

### Native Formatters

Standalone binaries installed on the host system. These offer the fastest execution with no network overhead.

| Formatter | Languages | Install Method |
|-----------|-----------|---------------|
| `clang-format` | C, C++, Java, JavaScript | apt/brew |
| `gofmt` | Go | Included with Go |
| `rustfmt` | Rust | Included with Rust |
| `black` | Python | pip |
| `prettier` | JS, TS, CSS, HTML, JSON, YAML, Markdown | npm |
| `shfmt` | Shell/Bash | binary |
| `swift-format` | Swift | brew |
| `dart format` | Dart | Included with Dart |
| `mix format` | Elixir | Included with Elixir |
| `ktlint` | Kotlin | binary |
| `stylua` | Lua | binary |

### Service Formatters

Run inside Docker containers, exposing HTTP endpoints. No local installation needed. HelixAgent connects to these via their service URL.

Service formatters run on ports 9210-9300 and are defined in `docker/formatters/docker-compose.formatters.yml`.

### Built-in Formatters

Use language-native formatting tools that ship with the language runtime (e.g., `gofmt`, `dart format`, `mix format`). These are a subset of native formatters that require no separate installation beyond the language itself.

### Unified Formatters

Multi-language formatters like `prettier` that handle several languages through a single tool with configuration-based language selection.

---

## REST API

### Format Code

```
POST /v1/format
Content-Type: application/json
```

**Request Body:**

```json
{
  "content": "def hello( ):\n  print('hello')",
  "language": "python",
  "config": { "line_length": 88 },
  "check_only": false
}
```

**Response:**

```json
{
  "content": "def hello():\n    print(\"hello\")\n",
  "changed": true,
  "formatter_name": "black",
  "success": true
}
```

**Parameters:** `content` (required), `language`, `file_path`, `config`, `line_length`, `indent_size`, `use_tabs`, `check_only`.

### List Formatters

```
GET /v1/formatters
```

Returns all registered formatters with metadata, supported languages, and health status. Use `POST /v1/format/batch` to format multiple files in a single request.

---

## Supported Languages and Formatters

| Language | Formatters Available |
|----------|---------------------|
| C / C++ | clang-format |
| Go | gofmt, goimports |
| Python | black, autopep8, yapf, isort |
| JavaScript | prettier, eslint |
| TypeScript | prettier, eslint |
| Rust | rustfmt |
| Java | clang-format, google-java-format |
| Kotlin | ktlint |
| Swift | swift-format |
| Dart | dart format |
| Elixir | mix format |
| Lua | stylua |
| Shell/Bash | shfmt |
| CSS | prettier |
| HTML | prettier |
| JSON | prettier, jq |
| YAML | prettier, yamlfmt |
| Markdown | prettier |
| SQL | sql-formatter |

---

## Formatter Interface

All formatters implement the `Formatter` interface defined in `internal/formatters/interface.go`:

```go
type Formatter interface {
    Name() string
    Version() string
    Languages() []string
    SupportsStdin() bool
    SupportsInPlace() bool
    SupportsCheck() bool
    SupportsConfig() bool
    Format(ctx context.Context, req *FormatRequest) (*FormatResult, error)
    FormatBatch(ctx context.Context, reqs []*FormatRequest) ([]*FormatResult, error)
    HealthCheck(ctx context.Context) error
    ValidateConfig(config map[string]interface{}) error
    DefaultConfig() map[string]interface{}
}
```

The `FormatResult` includes the formatted content, whether changes were made, execution duration, and optional statistics (lines changed, violations fixed).

---

## Docker Setup for Service Formatters

Service formatters run as containers. Start them with:

```bash
# Start all formatter containers
docker compose -f docker/formatters/docker-compose.formatters.yml up -d

# Or use the make target
make infra-start
```

Containers expose HTTP endpoints on ports 9210-9300. HelixAgent auto-discovers running formatter services and registers them in the formatter registry.

### Port Assignments

Service formatters are assigned sequential ports starting at 9210. Each container exposes a health endpoint at `/health` and a format endpoint at `/format`.

### Custom Formatter Container

To add a new service formatter:

1. Create a Dockerfile in `docker/formatters/`
2. Add the service to `docker-compose.formatters.yml`
3. Register the formatter in `internal/formatters/registry.go`
4. Implement the `Formatter` interface with HTTP calls to the container

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FORMATTERS_ENABLED` | `true` | Enable the formatter system |
| `FORMATTERS_CACHE_TTL` | `5m` | Cache duration for format results |
| `FORMATTERS_TIMEOUT` | `30s` | Maximum execution time per format |
| `FORMATTERS_SERVICE_BASE_URL` | `http://localhost` | Base URL for service formatters |

### Per-Formatter Configuration

Each formatter accepts language-specific configuration through the `config` field in format requests. Use `DefaultConfig()` to retrieve the default settings and `ValidateConfig()` to check custom configurations before use.

---

## AI Debate Integration

The formatter system integrates with the AI debate pipeline (`internal/services/debate_formatter_integration.go`). When code is generated during a debate, it passes through the appropriate formatter before being included in the final response. This ensures all generated code follows consistent style conventions regardless of which LLM provider produced it.

---

## Troubleshooting

**Formatter not found for language**: Check that the formatter is registered and healthy with `GET /v1/formatters`. For service formatters, verify the container is running.

**Service formatter timeout**: Increase `FORMATTERS_TIMEOUT` or check container resource limits. Large files may require more processing time.

**Inconsistent formatting**: Ensure all team members use the same formatter configuration. Store config files (`.prettierrc`, `pyproject.toml`, etc.) in the project root.

**Container fails to start**: Check Docker logs with `docker compose -f docker/formatters/docker-compose.formatters.yml logs`.
