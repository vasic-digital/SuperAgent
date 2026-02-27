# User Manual 23: Observability Setup

## Overview
Setting up observability for HelixAgent.

## Components
- Metrics (Prometheus)
- Logging (Structured)
- Tracing (OpenTelemetry)
- Alerting (AlertManager)

## Quick Start
```bash
# Start observability stack
docker-compose -f docker/observability.yml up -d

# Access:
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000
# Jaeger: http://localhost:16686
```

## Configuration
```yaml
observability:
  metrics:
    enabled: true
    port: 9090
  tracing:
    enabled: true
    endpoint: http://jaeger:14268/api/traces
```

## Custom Metrics
```go
requestDuration := prometheus.NewHistogram(...)
prometheus.MustRegister(requestDuration)
```
