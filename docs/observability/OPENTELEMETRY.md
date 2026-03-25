# OpenTelemetry Integration

HelixAgent integrates with OpenTelemetry for distributed tracing, metrics collection, and structured logging.

## Overview

The Observability module (`digital.vasic.observability`) provides OpenTelemetry tracing, Prometheus metrics, structured logging, health checks, and ClickHouse analytics. The internal adapter (`internal/adapters/observability/`) bridges the module into HelixAgent's request pipeline.

## Tracing Setup

Configure the OTLP exporter endpoint:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
OTEL_SERVICE_NAME=helixagent
OTEL_TRACES_SAMPLER=parentbased_traceidalgo
OTEL_TRACES_SAMPLER_ARG=0.1
```

Supported backends: Jaeger, Zipkin, Langfuse, and any OTLP-compatible collector.

## Metrics

Prometheus metrics are exposed at `/metrics`. Key metrics include:

| Metric | Type | Description |
|--------|------|-------------|
| `helixagent_request_duration_seconds` | Histogram | Request latency by endpoint |
| `helixagent_provider_requests_total` | Counter | Total requests per provider |
| `helixagent_provider_errors_total` | Counter | Provider errors by type |
| `helixagent_circuit_breaker_state` | Gauge | Circuit breaker state per provider |
| `helixagent_debate_rounds_total` | Counter | Debate rounds executed |
| `helixagent_ensemble_latency_seconds` | Histogram | Ensemble response time |

## Structured Logging

Logs use structured JSON format with trace and span IDs for correlation:

```json
{
  "level": "info",
  "msg": "provider response",
  "trace_id": "abc123",
  "span_id": "def456",
  "provider": "claude",
  "duration_ms": 1234
}
```

## Grafana Dashboard

A pre-built Grafana dashboard is available at `docs/monitoring/grafana-dashboard.json`. Import it into your Grafana instance for real-time visualization.

## Related Documentation

- [Prometheus Monitoring](../monitoring/PROMETHEUS_MONITORING.md)
- [Monitoring System](../monitoring/MONITORING_SYSTEM.md)
- [Health and Monitoring](../monitoring/README.md)
