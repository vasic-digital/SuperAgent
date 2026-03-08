# Video Course 59: Monitoring and Observability

## Course Overview

**Duration**: 2 hours 15 minutes
**Level**: Intermediate to Advanced
**Prerequisites**: Course 01-Fundamentals, Course 09-Production-Operations, Course 49-Monitoring-Alerting, familiarity with Prometheus query language (PromQL)

This course covers the complete monitoring and observability stack for HelixAgent, including Prometheus metrics setup, OpenTelemetry tracing configuration, Grafana dashboard creation, health endpoint design, and alerting configuration.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Configure Prometheus metrics collection for HelixAgent components
2. Set up OpenTelemetry distributed tracing with Jaeger or Zipkin backends
3. Design and build Grafana dashboards for provider health and ensemble performance
4. Implement health endpoints that reflect real service readiness
5. Configure alerting rules for critical operational thresholds
6. Create a custom metric and end-to-end dashboard panel

---

## Module 1: Prometheus Metrics Setup (30 min)

### 1.1 Metrics Architecture

**Video: How HelixAgent Exposes Metrics** (10 min)

- Metrics endpoint: `GET /metrics` on the configured metrics port
- Prometheus scrapes metrics at configurable intervals
- Three metric types used: counters, histograms, gauges
- Module: `digital.vasic.observability` in `Observability/`

**Core Metric Categories:**

| Category          | Example Metric                              | Type      |
|-------------------|---------------------------------------------|-----------|
| Request volume    | `helixagent_requests_total`                 | Counter   |
| Request latency   | `helixagent_request_duration_seconds`       | Histogram |
| Provider health   | `helixagent_provider_healthy`               | Gauge     |
| Circuit breaker   | `helixagent_circuit_breaker_state`          | Gauge     |
| Ensemble          | `helixagent_ensemble_provider_count`        | Gauge     |
| Cache             | `helixagent_cache_hits_total`               | Counter   |
| Memory            | `helixmemory_store_duration_seconds`        | Histogram |
| Debate            | `helixagent_debate_rounds_total`            | Counter   |

### 1.2 Registering Custom Metrics

**Video: Adding New Metrics** (10 min)

```go
var (
    providerLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "helixagent_provider_latency_seconds",
            Help:    "Latency of provider API calls in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"provider", "model", "status"},
    )
)

func init() {
    prometheus.MustRegister(providerLatency)
}

func recordProviderCall(provider, model string, duration time.Duration, err error) {
    status := "success"
    if err != nil {
        status = "error"
    }
    providerLatency.WithLabelValues(provider, model, status).Observe(duration.Seconds())
}
```

### 1.3 Prometheus Configuration

**Video: Scrape Configuration** (10 min)

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'helixagent'
    static_configs:
      - targets: ['helixagent:9090']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'helixagent-mcp'
    static_configs:
      - targets: ['mcp-bridge:9091']

  - job_name: 'helixagent-formatters'
    static_configs:
      - targets: ['formatters:9092']
```

### Hands-On Lab 1

Set up Prometheus and verify metric collection:

```bash
# Start monitoring infrastructure
make infra-core

# Verify metrics endpoint
curl http://localhost:9090/metrics | grep helixagent

# Open Prometheus UI and run a query
# http://localhost:9090
# Query: rate(helixagent_requests_total[5m])
```

---

## Module 2: OpenTelemetry Tracing (25 min)

### 2.1 Tracing Architecture

**Video: Distributed Tracing in HelixAgent** (8 min)

- Every request creates a trace with a unique trace ID
- Spans represent individual operations within a trace
- Trace propagation follows W3C TraceContext standard
- Supported backends: Jaeger, Zipkin, OTLP-compatible collectors

### 2.2 Configuring the Tracing Provider

**Video: OpenTelemetry SDK Setup** (10 min)

```go
func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("jaeger:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("creating OTLP exporter: %w", err)
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("helixagent"),
            semconv.ServiceVersionKey.String(version.Version),
        )),
        sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

### 2.3 Instrumenting Key Paths

**Video: Adding Spans to Critical Operations** (7 min)

```go
func (s *EnsembleService) Execute(ctx context.Context, req *Request) (*Response, error) {
    ctx, span := tracer.Start(ctx, "ensemble.execute",
        trace.WithAttributes(
            attribute.String("strategy", req.Strategy),
            attribute.Int("provider_count", len(req.Providers)),
        ),
    )
    defer span.End()

    // Each provider call gets its own child span
    for _, provider := range req.Providers {
        provCtx, provSpan := tracer.Start(ctx, "provider.complete",
            trace.WithAttributes(
                attribute.String("provider", provider.Name()),
            ),
        )
        resp, err := provider.Complete(provCtx, req)
        provSpan.End()
        // ...
    }
}
```

### Hands-On Lab 2

Configure tracing and view a distributed trace:

```bash
# Start Jaeger
docker run -d --name jaeger \
  -p 16686:16686 -p 4317:4317 \
  jaegertracing/all-in-one:latest

# Set tracing environment variables
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# Make a request to HelixAgent
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"auto","messages":[{"role":"user","content":"Hello"}]}'

# Open Jaeger UI: http://localhost:16686
# Search for service "helixagent"
```

---

## Module 3: Grafana Dashboard Creation (30 min)

### 3.1 Dashboard Design Principles

**Video: Effective Dashboard Layout** (8 min)

- Top row: high-level service health (request rate, error rate, latency)
- Second row: provider-specific metrics (per-provider latency, health status)
- Third row: infrastructure metrics (DB connections, cache hit rate, memory)
- Bottom row: detailed breakdown (debate metrics, ensemble statistics)

### 3.2 Essential Panels

**Video: Building Core Panels** (12 min)

**Request Rate Panel (Graph):**
```promql
sum(rate(helixagent_requests_total[5m])) by (endpoint)
```

**Error Rate Panel (Graph):**
```promql
sum(rate(helixagent_requests_total{status="error"}[5m]))
/ sum(rate(helixagent_requests_total[5m]))
```

**P95 Latency Panel (Graph):**
```promql
histogram_quantile(0.95,
  sum(rate(helixagent_request_duration_seconds_bucket[5m])) by (le)
)
```

**Provider Health Panel (Stat):**
```promql
helixagent_provider_healthy
```

**Circuit Breaker State Panel (State Timeline):**
```promql
helixagent_circuit_breaker_state
```

### 3.3 Advanced Panels

**Video: Specialized Visualizations** (10 min)

**Debate Performance Panel:**
```promql
histogram_quantile(0.95,
  sum(rate(helixagent_debate_round_duration_seconds_bucket[5m])) by (le, topology)
)
```

**Cache Efficiency Panel:**
```promql
sum(rate(helixagent_cache_hits_total[5m]))
/ (sum(rate(helixagent_cache_hits_total[5m])) + sum(rate(helixagent_cache_misses_total[5m])))
```

**HelixMemory Backend Latency Panel:**
```promql
histogram_quantile(0.95,
  sum(rate(helixmemory_store_duration_seconds_bucket[5m])) by (le, backend)
)
```

### Hands-On Lab 3

Build a Grafana dashboard from scratch:

1. Create a new dashboard in Grafana
2. Add a request rate panel with per-endpoint breakdown
3. Add an error rate panel with threshold coloring (green < 1%, yellow < 5%, red >= 5%)
4. Add a provider health status panel showing all providers
5. Add a P95 latency graph with provider breakdown
6. Save and export the dashboard JSON

---

## Module 4: Health Endpoint Design (20 min)

### 4.1 Health Check Levels

**Video: Shallow vs Deep Health Checks** (8 min)

| Level    | Endpoint            | Checks                                | Use Case           |
|----------|---------------------|---------------------------------------|---------------------|
| Liveness | `GET /health`       | Process is running                    | K8s liveness probe  |
| Readiness| `GET /health/ready` | All required dependencies reachable   | K8s readiness probe |
| Deep     | `GET /v1/monitoring/status` | Full component status with metrics | Operations dashboard |

### 4.2 Implementing Readiness Checks

**Video: Dependency-Aware Readiness** (7 min)

```go
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
    checks := map[string]HealthStatus{
        "database": h.checkDatabase(),
        "redis":    h.checkRedis(),
        "providers": h.checkMinProviders(2),
    }

    allHealthy := true
    for _, status := range checks {
        if status != Healthy {
            allHealthy = false
            break
        }
    }

    code := http.StatusOK
    if !allHealthy {
        code = http.StatusServiceUnavailable
    }

    c.JSON(code, gin.H{
        "status":    boolToStatus(allHealthy),
        "checks":    checks,
        "timestamp": time.Now().UTC(),
    })
}
```

### 4.3 Status Monitoring Endpoint

**Video: The /v1/monitoring/status Endpoint** (5 min)

```bash
curl http://localhost:7061/v1/monitoring/status
```

Response includes:
- Overall system status
- Per-provider health with last check timestamp
- Circuit breaker states
- Database and Redis connection pool statistics
- Active debate session count
- HelixMemory backend statuses

### Hands-On Lab 4

Implement and test health endpoints:

1. Verify liveness endpoint returns 200 when process is running
2. Stop Redis and verify readiness endpoint returns 503
3. Restart Redis and verify readiness recovers to 200
4. Check the full monitoring status endpoint output
5. Configure Kubernetes liveness and readiness probes

---

## Module 5: Alerting Configuration (20 min)

### 5.1 Alert Rule Design

**Video: Meaningful Alert Rules** (8 min)

```yaml
# alertmanager/rules/helixagent.yml
groups:
  - name: helixagent
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(helixagent_requests_total{status="error"}[5m]))
          / sum(rate(helixagent_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Error rate above 5% for 5 minutes"
          description: "Current error rate: {{ $value | humanizePercentage }}"

      - alert: ProviderDown
        expr: helixagent_provider_healthy == 0
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Provider {{ $labels.provider }} is unhealthy"

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            sum(rate(helixagent_request_duration_seconds_bucket[5m])) by (le)
          ) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "P95 latency above 5 seconds"

      - alert: CircuitBreakerOpen
        expr: helixagent_circuit_breaker_state == 2
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Circuit breaker open for {{ $labels.provider }}"

      - alert: AllProvidersDown
        expr: sum(helixagent_provider_healthy) == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "All LLM providers are unhealthy"
```

### 5.2 Notification Channels

**Video: Configuring Alert Delivery** (7 min)

- Email notifications for warning-level alerts
- Slack/webhook notifications for critical alerts
- PagerDuty integration for out-of-hours escalation
- Silencing and inhibition rules to prevent alert fatigue

### 5.3 Alert Testing

**Video: Verifying Alerts Fire Correctly** (5 min)

```bash
# Use Prometheus alerting rule unit testing
promtool test rules tests/alerting-rules-test.yml

# Or trigger a test alert manually
curl -X POST http://localhost:9093/api/v2/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {"alertname": "TestAlert", "severity": "info"},
    "annotations": {"summary": "Test alert from course exercise"}
  }]'
```

### Hands-On Lab 5

Create a custom metric and complete dashboard panel:

1. Define a new Prometheus histogram for a specific operation
2. Register the metric and instrument the code path
3. Deploy and verify the metric appears in `/metrics`
4. Write a PromQL query for P95 latency of the new metric
5. Create a Grafana panel displaying the metric
6. Write an alerting rule for when P95 exceeds a threshold
7. Test the alert fires by injecting artificial latency

---

## Course Summary

### Key Takeaways

1. Prometheus metrics provide real-time visibility into request rates, latencies, and provider health
2. OpenTelemetry tracing with W3C TraceContext enables end-to-end request path analysis
3. Grafana dashboards should follow a top-down layout: service health, providers, infrastructure, details
4. Health endpoints must distinguish between liveness (process alive) and readiness (dependencies available)
5. Alert rules should have meaningful thresholds, duration windows, and clear annotations
6. Every custom metric needs an associated dashboard panel and alerting rule

### Assessment Questions

1. What are the three Prometheus metric types used in HelixAgent and when is each appropriate?
2. How does trace sampling rate affect observability and performance?
3. What PromQL expression calculates the P95 latency from a histogram?
4. What is the difference between liveness and readiness health checks?
5. Why should alert rules include a `for` duration clause?

### Related Courses

- Course 09: Production Operations
- Course 28: Resource Monitoring
- Course 49: Monitoring and Alerting
- Course 53: HelixMemory Deep Dive
- Course 60: Enterprise Deployment

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
