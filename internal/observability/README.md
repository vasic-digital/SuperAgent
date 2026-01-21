# Package: observability

## Overview

The `observability` package provides comprehensive observability capabilities including distributed tracing, metrics, and logging integration using OpenTelemetry. It supports multiple exporters including Jaeger, Zipkin, and Langfuse.

## Architecture

```
observability/
├── tracing.go          # Distributed tracing with OpenTelemetry
├── metrics.go          # Prometheus metrics
├── exporters.go        # Trace exporters (Jaeger, Zipkin, Langfuse)
└── observability_test.go # Unit tests (81.1% coverage)
```

## Features

- **Distributed Tracing**: End-to-end request tracing
- **Metrics**: Prometheus-compatible metrics
- **Multi-Exporter**: Jaeger, Zipkin, Langfuse support
- **Context Propagation**: W3C Trace Context

## Key Types

### TracerConfig

```go
type TracerConfig struct {
    ServiceName    string
    Environment    string
    Endpoint       string
    SampleRate     float64
    ExporterType   ExporterType
}
```

### ExporterType

```go
const (
    ExporterTypeJaeger   ExporterType = "jaeger"
    ExporterTypeZipkin   ExporterType = "zipkin"
    ExporterTypeLangfuse ExporterType = "langfuse"
    ExporterTypeOTLP     ExporterType = "otlp"
)
```

## Usage

### Initialize Tracing

```go
import "dev.helix.agent/internal/observability"

config := &observability.TracerConfig{
    ServiceName:  "helixagent",
    Environment:  "production",
    Endpoint:     "http://jaeger:14268/api/traces",
    SampleRate:   0.1,  // 10% sampling
    ExporterType: observability.ExporterTypeJaeger,
}

tracer, shutdown, err := observability.InitTracer(config)
defer shutdown(ctx)
```

### Creating Spans

```go
ctx, span := tracer.Start(ctx, "process-request")
defer span.End()

// Add attributes
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.Int("tokens.used", tokens),
)

// Record events
span.AddEvent("model-response-received")

// Handle errors
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

### Metrics

```go
// Create metrics
requestCounter := observability.NewCounter("requests_total", "Total requests")
latencyHistogram := observability.NewHistogram("request_latency_seconds", "Request latency")

// Record metrics
requestCounter.Inc()
latencyHistogram.Observe(duration.Seconds())
```

### Langfuse Integration

```go
config := &observability.TracerConfig{
    ServiceName:  "helixagent",
    ExporterType: observability.ExporterTypeLangfuse,
    Endpoint:     "https://cloud.langfuse.com",
    // Langfuse-specific config
    LangfusePublicKey: os.Getenv("LANGFUSE_PUBLIC_KEY"),
    LangfuseSecretKey: os.Getenv("LANGFUSE_SECRET_KEY"),
}
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| ServiceName | string | "helixagent" | Service identifier |
| SampleRate | float64 | 1.0 | Trace sampling rate |
| ExporterType | string | "otlp" | Exporter type |
| BatchTimeout | time.Duration | 5s | Batch export timeout |

## Environment Variables

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
OTEL_SERVICE_NAME=helixagent
JAEGER_AGENT_HOST=localhost
JAEGER_AGENT_PORT=6831
LANGFUSE_PUBLIC_KEY=pk-xxx
LANGFUSE_SECRET_KEY=sk-xxx
```

## Testing

```bash
go test -v ./internal/observability/...
go test -cover ./internal/observability/...  # 81.1% coverage
```

## Dependencies

### External
- `go.opentelemetry.io/otel` - OpenTelemetry SDK
- `go.opentelemetry.io/otel/exporters/*` - Trace exporters

## See Also

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Langfuse Documentation](https://langfuse.com/docs)
