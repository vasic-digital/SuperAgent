# Monitoring & Observability Documentation

This directory contains documentation for HelixAgent's comprehensive monitoring, observability, and health checking infrastructure.

## Overview

HelixAgent implements a multi-layered observability stack with Prometheus metrics collection, Grafana dashboards, Loki log aggregation, and a custom monitoring system for challenge execution.

## Documentation Index

| Document | Description |
|----------|-------------|
| [PROMETHEUS_MONITORING.md](PROMETHEUS_MONITORING.md) | Complete Prometheus-based monitoring stack architecture |
| [MONITORING_SYSTEM.md](MONITORING_SYSTEM.md) | Challenge execution monitoring with resource tracking and reports |
| [grafana-dashboard.json](grafana-dashboard.json) | Pre-configured Grafana dashboard for HelixAgent metrics |

## Prometheus Metrics Reference

### HelixAgent Core Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `helixagent_up` | Gauge | Service availability (1=up, 0=down) |
| `helixagent_response_time_ms` | Histogram | API response time in milliseconds |
| `helixagent_providers_total` | Gauge | Total LLM providers configured |
| `helixagent_providers_healthy` | Gauge | Number of healthy LLM providers |
| `helixagent_requests_total` | Counter | Total API requests processed |
| `helixagent_errors_total` | Counter | Total errors by type |
| `helixagent_debate_rounds` | Histogram | AI debate round counts |
| `helixagent_cache_hits` | Counter | Cache hit count |
| `helixagent_cache_misses` | Counter | Cache miss count |

### Infrastructure Metrics

| Metric | Source | Description |
|--------|--------|-------------|
| `pg_*` | PostgreSQL Exporter | Database connections, queries, locks |
| `redis_*` | Redis Exporter | Memory usage, commands, clients |
| `node_*` | Node Exporter | System CPU, memory, disk, network |
| `container_*` | cAdvisor | Container resource usage |

## Grafana Dashboards

### Pre-configured Dashboards

| Dashboard | Description |
|-----------|-------------|
| HelixAgent Overview | Core metrics, provider health, request rates |
| LLM Provider Health | Per-provider response times and availability |
| Infrastructure | PostgreSQL, Redis, system resources |
| AI Debate Analytics | Debate metrics, round counts, consensus rates |

### Accessing Dashboards

```bash
# Start monitoring stack
./scripts/monitoring.sh start

# Access Grafana
open http://localhost:3000
# Default credentials: admin/admin123
```

### Dashboard Configuration

The pre-configured dashboard is located at:
- `docs/monitoring/grafana-dashboard.json`

Import via Grafana UI or use provisioning:
```yaml
# provisioning/dashboards/helixagent.yml
apiVersion: 1
providers:
  - name: HelixAgent
    folder: HelixAgent
    type: file
    options:
      path: /var/lib/grafana/dashboards
```

## Health Checking

### Service Health Endpoints

| Endpoint | Description |
|----------|-------------|
| `/health` | Basic health check |
| `/v1/health` | Detailed health with component status |
| `/v1/monitoring/status` | Full monitoring status |
| `/v1/bigdata/health` | Big Data component health |

### Health Check Types

| Type | Method | Description |
|------|--------|-------------|
| TCP | `net.DialTimeout` | Port connectivity check |
| HTTP | `GET /health` | HTTP endpoint check |
| gRPC | gRPC Health Check | gRPC service health |

### Circuit Breaker Status

```bash
# Check circuit breaker states
make circuit-breakers

# Reset all circuit breakers
make monitoring-reset-circuits
```

## Alert Configuration

### Alertmanager Rules

Alerts are configured in `configs/prometheus/alerts.yml`:

| Alert | Condition | Severity |
|-------|-----------|----------|
| HelixAgentDown | Instance down > 1m | critical |
| HighErrorRate | Error rate > 5% | warning |
| ProviderUnhealthy | Provider health = 0 | warning |
| HighLatency | p95 > 5s | warning |
| LowCacheHitRate | Hit rate < 50% | info |

### Notification Channels

Configure in `configs/alertmanager/alertmanager.yml`:
- Email
- Slack
- PagerDuty
- Webhook

## Monitoring Stack Architecture

```
+----------------------------------------------------------------------+
|                   HELIXAGENT MONITORING ARCHITECTURE                   |
+----------------------------------------------------------------------+
|                                                                        |
|  +-------------+  +-------------+  +-------------+  +--------------+  |
|  | HelixAgent  |  |  ChromaDB   |  |   Cognee    |  | LLMsVerifier |  |
|  |  API:7061   |  |   :8001     |  |   :8000     |  |    :8180     |  |
|  +------+------+  +------+------+  +------+------+  +------+-------+  |
|         |                |                |                |          |
|         +----------------+----------------+----------------+          |
|                                   |                                   |
|                    +--------------v--------------+                    |
|                    | Custom HelixAgent Exporter  |                    |
|                    |          :9200              |                    |
|                    +--------------+--------------+                    |
|                                   |                                   |
|  +----------------------------------------------------------------+  |
|  |                      PROMETHEUS :9090                           |  |
|  |  Scrape Targets:                                                |  |
|  |  - HelixAgent metrics    - PostgreSQL exporter   - Redis exp   |  |
|  |  - ChromaDB health       - Cognee health         - Node exp    |  |
|  |  - LLMsVerifier          - cAdvisor              - Blackbox    |  |
|  +----------------------------------------------------------------+  |
|                                   |                                   |
|         +-------------------------+-------------------------+         |
|         v                         v                         v         |
|  +-----------+             +-----------+             +-----------+    |
|  |  Grafana  |             |Alertmanager|           |   Loki    |    |
|  |   :3000   |             |   :9093   |             |   :3100   |    |
|  +-----------+             +-----------+             +-----------+    |
|                                                            ^          |
|                                                     +-----------+     |
|                                                     | Promtail  |     |
|                                                     +-----------+     |
+----------------------------------------------------------------------+
```

## Services Summary

| Service | Port | Description |
|---------|------|-------------|
| Prometheus | 9090 | Metrics collection and alerting |
| Grafana | 3000 | Visualization dashboards |
| Alertmanager | 9093 | Alert routing and notifications |
| Loki | 3100 | Log aggregation |
| Promtail | - | Log collection agent |
| Node Exporter | 9100 | System metrics |
| cAdvisor | 8081 | Container metrics |
| Redis Exporter | 9121 | Redis metrics |
| PostgreSQL Exporter | 9187 | PostgreSQL metrics |
| Blackbox Exporter | 9115 | HTTP/TCP/ICMP probing |
| HelixAgent Exporter | 9200 | Custom HelixAgent metrics |

## Quick Start

```bash
# Start all monitoring services
podman-compose -f docker-compose.monitoring.yml up -d

# Or use convenience script
./scripts/monitoring.sh start

# Check status
./scripts/monitoring.sh status

# Stop monitoring
./scripts/monitoring.sh stop
```

## Challenge Monitoring

The challenge monitoring system provides:
- Real-time resource monitoring (CPU, memory, disk, network)
- Log collection from all components
- Memory leak detection
- Warning/error analysis
- Automatic issue investigation
- HTML/JSON report generation

```bash
# Run monitored challenges
./challenges/monitoring/run_monitored_challenges.sh

# Run specific challenges
./challenges/monitoring/run_monitored_challenges.sh \
  --challenges "health_monitoring,provider_verification"
```

## Related Documentation

- [Architecture Overview](../architecture/README.md)
- [Service Architecture](../architecture/SERVICE_ARCHITECTURE.md)
- [Security Scanning](../SECURITY_SCANNING.md)
- [Circuit Breaker](../architecture/CIRCUIT_BREAKER.md)

## Make Targets

```bash
make monitoring-status          # Check monitoring status
make circuit-breakers           # View circuit breaker states
make provider-health            # Check provider health
make fallback-chain             # View fallback chain
make monitoring-reset-circuits  # Reset circuit breakers
make force-health-check         # Force health check
```
