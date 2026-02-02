# formatters

Package `formatters` implements a universal code formatting system supporting 32+ formatters across 19 languages. Formatters are categorized as native (standalone binaries), service (Docker HTTP/RPC), or built-in (language-native tools). The package provides a registry, executor, caching layer, health monitoring, metrics, and factory for formatter lifecycle management.

## Architecture

### Key Types

- **`Formatter`** -- Core interface: `Format`, `FormatBatch`, `HealthCheck`, `ValidateConfig`, capability queries.
- **`BaseFormatter`** -- Shared implementation for identity, capabilities, and default config.
- **`FormatterRegistry`** -- Thread-safe registry mapping formatters by name and language, with auto-detection.
- **`FormatterMetadata`** -- Describes a formatter's type, architecture, installation, performance, and config format.
- **`FormatRequest`** -- Input: content, language, config overrides, line length, indent, check-only mode.
- **`FormatResult`** -- Output: formatted content, changed flag, duration, warnings, statistics.
- **`FormatStats`** -- Lines/bytes changed, style violations fixed.

### Formatter Types

| Type       | Constant              | Description                          | Examples                    |
|------------|-----------------------|--------------------------------------|-----------------------------|
| Native     | `FormatterTypeNative` | Standalone binary                    | clang-format, rustfmt, gofmt|
| Service    | `FormatterTypeService`| Dockerized HTTP/RPC formatter        | prettier, black, ktlint     |
| Built-in   | `FormatterTypeBuiltin`| Language-bundled tool                 | go fmt, dart format         |
| Unified    | `FormatterTypeUnified`| Multi-language formatter             | prettier                    |

### Package Files

| File           | Responsibility                                    |
|----------------|---------------------------------------------------|
| `interface.go` | `Formatter` interface, request/result types, `BaseFormatter` |
| `registry.go`  | `FormatterRegistry` with language detection (50+ extensions) |
| `executor.go`  | Concurrent format execution with timeouts         |
| `cache.go`     | Content-based format result caching                |
| `factory.go`   | Formatter construction from metadata               |
| `health.go`    | Health check orchestration                         |
| `metrics.go`   | Prometheus metrics collection                      |
| `config.go`    | Registry and formatter configuration               |
| `system.go`    | System-level formatter management                  |
| `init.go`      | Package initialization and default registration    |
| `versions.go`  | Version detection and compatibility                |

## Public API

```go
// Registry
NewFormatterRegistry(config *RegistryConfig, logger *logrus.Logger) *FormatterRegistry
Register(formatter Formatter, metadata *FormatterMetadata) error
Unregister(name string) error
Get(name string) (Formatter, error)
GetByLanguage(language string) []Formatter
GetPreferredFormatter(language string, preferences map[string]string) (Formatter, error)
DetectFormatter(filePath string, content string) (Formatter, error)
DetectLanguageFromPath(filePath string) string
List() []string
ListByType(ftype FormatterType) []string
HealthCheckAll(ctx context.Context) map[string]error
Start(ctx context.Context) error
Stop(ctx context.Context) error
```

## Configuration

```go
config := &formatters.RegistryConfig{
    SubmodulesPath:     "formatters/",
    BinariesPath:       "bin/formatters/",
    ServicesComposeFile: "docker/formatters/docker-compose.yml",
    ServicesEnabled:     true,
    EnableCaching:       true,
    CacheTTL:            10 * time.Minute,
    DefaultTimeout:      30 * time.Second,
    MaxConcurrent:       8,
}
```

Service-based formatters run in Docker containers on ports 9210-9300. REST API endpoints: `POST /v1/format`, `GET /v1/formatters`.

## Usage

```go
registry := formatters.NewFormatterRegistry(config, logger)
registry.Start(ctx)

formatter, err := registry.DetectFormatter("main.go", content)
result, err := formatter.Format(ctx, &formatters.FormatRequest{
    Content:  content,
    FilePath: "main.go",
    Language: "go",
})
// result.Content => formatted code
// result.Changed => true/false
```

## Testing

```bash
go test -v ./internal/formatters/
go test -bench=. ./internal/formatters/          # Benchmarks (cache)
go test -v -run TestRegistry ./internal/formatters/
go test -v -run TestExecutor ./internal/formatters/
```

Service formatter tests require Docker containers started via `make infra-start`.
