# User Manual 23: Observability Setup

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Observability Architecture](#observability-architecture)
4. [Quick Start](#quick-start)
5. [OpenTelemetry Configuration](#opentelemetry-configuration)
6. [Distributed Tracing](#distributed-tracing)
7. [Jaeger Setup](#jaeger-setup)
8. [Zipkin Setup](#zipkin-setup)
9. [Langfuse Integration](#langfuse-integration)
10. [Structured Logging](#structured-logging)
11. [Prometheus Metrics](#prometheus-metrics)
12. [Health Endpoints](#health-endpoints)
13. [Trace Propagation](#trace-propagation)
14. [Configuration Reference](#configuration-reference)
15. [Troubleshooting](#troubleshooting)
16. [Related Resources](#related-resources)

## Overview

HelixAgent provides a comprehensive observability stack covering three pillars: metrics (Prometheus), tracing (OpenTelemetry with Jaeger/Zipkin), and logging (structured JSON logs). Additionally, Langfuse integration provides LLM-specific observability including prompt tracking, token usage analytics, and model comparison.

The Observability module (`digital.vasic.observability`) provides the reusable primitives: OpenTelemetry tracing, Prometheus metrics, structured logging, health checks, and ClickHouse analytics.

## Prerequisites

- Docker or Podman for running the observability stack
- HelixAgent running on port 7061
- Network access between HelixAgent and observability services
- 4 GB additional RAM for the full observability stack

## Observability Architecture

```
+------------------+    +------------------+    +------------------+
|   HelixAgent     |    |   HelixAgent     |    |   HelixAgent     |
|   Instance 1     |    |   Instance 2     |    |   Instance N     |
+--------+---------+    +--------+---------+    +--------+---------+
         |                       |                       |
         |  OTLP/gRPC           |  OTLP/gRPC           |  OTLP/gRPC
         |                       |                       |
+--------v-----------------------v-----------------------v---------+
|                    OpenTelemetry Collector                        |
|                    (receives, processes, exports)                 |
+-----+----------------+----------------+----------------+---------+
      |                |                |                |
      v                v                v                v
+----------+   +----------+   +------------+   +------------+
| Jaeger   |   | Zipkin   |   | Prometheus |   | ClickHouse |
| (traces) |   | (traces) |   | (metrics)  |   | (analytics)|
+----------+   +----------+   +-----+------+   +------------+
                                     |
                               +-----v------+
                               |  Grafana   |
                               | (dashboards)|
                               +------------+
```

## Quick Start

### Start the Full Observability Stack

```bash
# Start all observability services
docker-compose -f docker/observability.yml up -d

# Verify services
docker-compose -f docker/observability.yml ps
```

### Access Points

| Service | URL | Purpose |
|---|---|---|
| Prometheus | http://localhost:9090 | Metrics queries and alerting |
| Grafana | http://localhost:3000 | Dashboards and visualization |
| Jaeger | http://localhost:16686 | Distributed trace viewer |
| Zipkin | http://localhost:9411 | Alternative trace viewer |
| Langfuse | http://localhost:3001 | LLM observability platform |
| HelixAgent Metrics | http://localhost:7061/metrics | Raw Prometheus metrics |
| HelixAgent Health | http://localhost:7061/v1/monitoring/status | System health status |

### Minimal Setup (Tracing Only)

```bash
# Start just Jaeger for trace collection
docker run -d --name jaeger \
    -p 16686:16686 \
    -p 14268:14268 \
    -p 4317:4317 \
    jaegertracing/all-in-one:1.55
```

Configure HelixAgent to send traces:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_SERVICE_NAME=helixagent
```

## OpenTelemetry Configuration

### Environment Variables

```bash
# OpenTelemetry SDK configuration
export OTEL_SERVICE_NAME=helixagent
export OTEL_SERVICE_VERSION=1.0.0
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_TRACES_SAMPLER=parentbased_traceidratio
export OTEL_TRACES_SAMPLER_ARG=0.1  # Sample 10% of traces
export OTEL_RESOURCE_ATTRIBUTES=deployment.environment=production,host.name=helixagent-1
```

### Application Configuration

```yaml
# configs/production.yaml
observability:
  tracing:
    enabled: true
    exporter: otlp
    endpoint: http://otel-collector:4317
    sampler: parentbased_traceidratio
    sample_rate: 0.1
    propagators:
      - tracecontext
      - baggage
  metrics:
    enabled: true
    exporter: prometheus
    port: 7061
    path: /metrics
  logging:
    level: info
    format: json
    output: stdout
```

### Initializing the Tracer

```go
package observability

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/sdk/resource"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func InitTracer(ctx context.Context, serviceName, endpoint string) (*trace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("create OTLP exporter: %w", err)
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
        )),
        trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.1))),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

## Distributed Tracing

### Creating Spans

```go
import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("helixagent.services")

func (s *EnsembleService) Execute(ctx context.Context, req *EnsembleRequest) (*EnsembleResponse, error) {
    ctx, span := tracer.Start(ctx, "ensemble.execute")
    defer span.End()

    span.SetAttributes(
        attribute.String("ensemble.topology", req.Topology),
        attribute.Int("ensemble.providers", len(req.Providers)),
    )

    // Child span for each provider call
    for _, provider := range req.Providers {
        func() {
            ctx, providerSpan := tracer.Start(ctx, "provider.complete",
                trace.WithAttributes(
                    attribute.String("provider.name", provider.Name()),
                ))
            defer providerSpan.End()

            resp, err := provider.Complete(ctx, req.LLMRequest)
            if err != nil {
                providerSpan.RecordError(err)
                providerSpan.SetStatus(codes.Error, err.Error())
                return
            }
            providerSpan.SetAttributes(
                attribute.Int("response.tokens", resp.Usage.TotalTokens),
            )
        }()
    }

    return response, nil
}
```

### Trace Context in HTTP Handlers

```go
import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

func SetupRouter() *gin.Engine {
    r := gin.New()
    r.Use(otelgin.Middleware("helixagent"))
    // ... routes ...
    return r
}
```

## Jaeger Setup

### Docker Compose Configuration

```yaml
services:
  jaeger:
    image: jaegertracing/all-in-one:1.55
    ports:
      - "16686:16686"   # Jaeger UI
      - "14268:14268"   # HTTP collector
      - "4317:4317"     # OTLP gRPC
      - "4318:4318"     # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - SPAN_STORAGE_TYPE=badger
      - BADGER_EPHEMERAL=false
      - BADGER_DIRECTORY_VALUE=/badger/data
      - BADGER_DIRECTORY_KEY=/badger/key
    volumes:
      - jaeger_data:/badger
```

### Searching Traces in Jaeger

1. Open http://localhost:16686
2. Select service: `helixagent`
3. Filter by operation: `ensemble.execute`, `provider.complete`, `debate.round`
4. Set time range and click "Find Traces"
5. Click a trace to see the full span tree with timing

## Zipkin Setup

```yaml
services:
  zipkin:
    image: openzipkin/zipkin:3
    ports:
      - "9411:9411"
    environment:
      - STORAGE_TYPE=mem
```

Configure HelixAgent to export to Zipkin:

```bash
export OTEL_EXPORTER_ZIPKIN_ENDPOINT=http://localhost:9411/api/v2/spans
```

## Langfuse Integration

Langfuse provides LLM-specific observability: prompt version tracking, token usage, cost analysis, and model comparison.

### Configuration

```bash
export LANGFUSE_HOST=http://localhost:3001
export LANGFUSE_PUBLIC_KEY=pk-lf-...
export LANGFUSE_SECRET_KEY=sk-lf-...
```

### Tracking LLM Calls

```go
func (p *Provider) Complete(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
    // Start Langfuse generation span
    generation := langfuse.StartGeneration(ctx, langfuse.GenerationParams{
        Name:  "chat-completion",
        Model: p.model,
        Input: req.Messages,
    })
    defer generation.End()

    resp, err := p.doComplete(ctx, req)
    if err != nil {
        generation.SetError(err)
        return nil, err
    }

    generation.SetOutput(resp.Content)
    generation.SetUsage(langfuse.Usage{
        PromptTokens:     resp.Usage.PromptTokens,
        CompletionTokens: resp.Usage.CompletionTokens,
    })

    return resp, nil
}
```

## Structured Logging

HelixAgent uses structured JSON logging for machine-parseable log output:

```go
import "log/slog"

var logger = slog.Default()

func (s *Service) ProcessRequest(ctx context.Context, req *Request) {
    logger.InfoContext(ctx, "processing request",
        slog.String("request_id", req.ID),
        slog.String("provider", req.Provider),
        slog.Int("message_count", len(req.Messages)),
    )

    // ... processing ...

    if err != nil {
        logger.ErrorContext(ctx, "request failed",
            slog.String("request_id", req.ID),
            slog.String("error", err.Error()),
            slog.Duration("duration", time.Since(start)),
        )
    }
}
```

### Log Levels

| Level | Use Case |
|---|---|
| `DEBUG` | Detailed diagnostic information (disabled in production) |
| `INFO` | Routine operational events (request received, provider selected) |
| `WARN` | Recoverable issues (circuit breaker tripped, fallback activated) |
| `ERROR` | Failures that need attention (provider error, database connection lost) |

### Log Format

```json
{
    "time": "2026-03-08T10:15:30.123Z",
    "level": "INFO",
    "msg": "processing request",
    "request_id": "req-abc123",
    "provider": "deepseek",
    "message_count": 3,
    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

## Prometheus Metrics

See [User Manual 18: Performance Monitoring](18-performance-monitoring.md) for detailed Prometheus setup. Key points:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Expose metrics endpoint
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

## Health Endpoints

Every HelixAgent service exposes health check endpoints:

```bash
# Full system status (all components)
curl -s http://localhost:7061/v1/monitoring/status | jq .

# Startup verification report
curl -s http://localhost:7061/v1/startup/verification | jq .

# BigData subsystem health
curl -s http://localhost:7061/v1/bigdata/health | jq .
```

### Health Check Response Format

```json
{
    "status": "healthy",
    "timestamp": "2026-03-08T10:15:30Z",
    "components": {
        "postgresql": {"status": "healthy", "latency_ms": 2},
        "redis": {"status": "healthy", "latency_ms": 1},
        "providers": {
            "healthy": 18,
            "unhealthy": 4,
            "total": 22
        },
        "circuit_breakers": {
            "closed": 18,
            "open": 3,
            "half_open": 1
        }
    },
    "uptime_seconds": 86400
}
```

## Trace Propagation

HelixAgent propagates trace context across service boundaries using the W3C Trace Context standard:

### HTTP Headers

```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
tracestate: helixagent=provider:deepseek;debate_id:d123
```

### Cross-Service Propagation

When HelixAgent calls LLM providers, MCP servers, or internal services, trace context is automatically injected into outgoing HTTP headers:

```go
import "go.opentelemetry.io/otel/propagation"

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Inject trace context into outgoing request
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
    return c.httpClient.Do(req)
}
```

## Configuration Reference

| Environment Variable | Default | Description |
|---|---|---|
| `OTEL_SERVICE_NAME` | `helixagent` | Service name in traces |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `""` | OTLP collector endpoint |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` | OTLP protocol (grpc or http) |
| `OTEL_TRACES_SAMPLER` | `always_on` | Sampling strategy |
| `OTEL_TRACES_SAMPLER_ARG` | `1.0` | Sampling rate (0.0-1.0) |
| `LANGFUSE_HOST` | `""` | Langfuse server URL |
| `LANGFUSE_PUBLIC_KEY` | `""` | Langfuse public key |
| `LANGFUSE_SECRET_KEY` | `""` | Langfuse secret key |
| `LOG_LEVEL` | `info` | Logging level |
| `LOG_FORMAT` | `json` | Log format (json or text) |

## Troubleshooting

### No Traces Appearing in Jaeger

**Symptom:** Jaeger UI shows no traces for the `helixagent` service.

**Solutions:**
1. Verify `OTEL_EXPORTER_OTLP_ENDPOINT` is set correctly
2. Check Jaeger is running: `curl http://localhost:16686`
3. Check OTLP port is accessible: `curl http://localhost:4317`
4. Verify sampling is not set to 0: check `OTEL_TRACES_SAMPLER_ARG`
5. Look for OTLP export errors in HelixAgent logs

### Metrics Endpoint Returns 404

**Symptom:** `curl http://localhost:7061/metrics` returns 404.

**Solutions:**
1. Verify the Prometheus HTTP handler is registered on the router
2. Check if metrics are exposed on a different port
3. Confirm HelixAgent started successfully (check startup logs)

### High Cardinality Warning

**Symptom:** Prometheus warns about high cardinality metrics.

**Solutions:**
1. Avoid using request IDs or user IDs as metric labels
2. Use bounded label values (provider names, status codes, error categories)
3. Use `metric_relabel_configs` in Prometheus to drop high-cardinality series
4. Aggregate at the application level before exposing

### Log Output is Unstructured

**Symptom:** Logs appear as plain text instead of JSON.

**Solutions:**
1. Set `LOG_FORMAT=json` in environment
2. Ensure `slog` is configured with a JSON handler, not the default text handler
3. Check for third-party libraries writing directly to `fmt.Println` or `log.Println`

## Related Resources

- [User Manual 18: Performance Monitoring](18-performance-monitoring.md) -- Prometheus metrics and Grafana dashboards
- [User Manual 29: Disaster Recovery](29-disaster-recovery.md) -- Monitoring during DR events
- Observability module: `Observability/`
- Internal observability: `internal/observability/`
- OpenTelemetry Go SDK: https://opentelemetry.io/docs/languages/go/
- Jaeger documentation: https://www.jaegertracing.io/docs/
- Langfuse documentation: https://langfuse.com/docs
