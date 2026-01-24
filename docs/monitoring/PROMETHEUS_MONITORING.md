# HelixAgent Prometheus Monitoring Stack

## Overview

HelixAgent includes a comprehensive Prometheus-based monitoring stack that provides real-time observability for all services in the ecosystem.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                     HELIXAGENT MONITORING ARCHITECTURE                        │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ HelixAgent  │  │  ChromaDB   │  │   Cognee    │  │LLMsVerifier │         │
│  │  API:7061   │  │   :8001     │  │   :8000     │  │   :8180     │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
│         │                │                │                │                 │
│         └────────────────┴────────────────┴────────────────┘                 │
│                                   │                                          │
│                    ┌──────────────▼──────────────┐                          │
│                    │  Custom HelixAgent Exporter │                          │
│                    │         :9200               │                          │
│                    └──────────────┬──────────────┘                          │
│                                   │                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                          PROMETHEUS :9090                            │    │
│  │  Scrape Targets:                                                     │    │
│  │  • HelixAgent metrics     • PostgreSQL exporter    • Redis exporter │    │
│  │  • ChromaDB health        • Cognee health          • Node exporter  │    │
│  │  • LLMsVerifier metrics   • cAdvisor               • Blackbox prober│    │
│  └────────────────────────────────┬────────────────────────────────────┘    │
│                                   │                                          │
│         ┌─────────────────────────┼─────────────────────────┐               │
│         ▼                         ▼                         ▼               │
│  ┌─────────────┐          ┌─────────────┐          ┌─────────────┐         │
│  │  Grafana    │          │ Alertmanager│          │    Loki     │         │
│  │   :3000     │          │   :9093     │          │   :3100     │         │
│  └─────────────┘          └─────────────┘          └─────────────┘         │
│                                                            ▲                │
│                                                            │                │
│                                                     ┌─────────────┐         │
│                                                     │  Promtail   │         │
│                                                     │ Log Shipper │         │
│                                                     └─────────────┘         │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| **Prometheus** | 9090 | Metrics collection, alerting, and time-series database |
| **Grafana** | 3000 | Visualization dashboards and alerting UI |
| **Alertmanager** | 9093 | Alert routing, deduplication, and notifications |
| **Loki** | 3100 | Log aggregation and querying |
| **Promtail** | - | Log collection agent |
| **Node Exporter** | 9100 | System metrics (CPU, memory, disk, network) |
| **cAdvisor** | 8081 | Container metrics |
| **Redis Exporter** | 9121 | Redis metrics |
| **PostgreSQL Exporter** | 9187 | PostgreSQL metrics |
| **Blackbox Exporter** | 9115 | HTTP/TCP/ICMP probing |
| **HelixAgent Exporter** | 9200 | Custom metrics for HelixAgent ecosystem |

## Quick Start

### Start the Monitoring Stack

```bash
# Start all monitoring services
cd /path/to/HelixAgent
podman-compose -f docker-compose.monitoring.yml up -d

# Or use the convenience script
./scripts/monitoring.sh start
```

### Access the UIs

- **Grafana**: http://localhost:3000 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093
- **Loki**: http://localhost:3100

### Check Status

```bash
./scripts/monitoring.sh status
```

## Monitored Services

### HelixAgent Core (Port 7061)

| Metric | Description |
|--------|-------------|
| `helixagent_up` | Service availability (1=up, 0=down) |
| `helixagent_response_time_ms` | API response time |
| `helixagent_providers_total` | Total LLM providers |
| `helixagent_providers_healthy` | Healthy LLM providers |
| `helixagent_mcp_servers_total` | MCP servers count |
| `helixagent_tools_total` | Available tools count |
| `helixagent_llm_request_duration_seconds` | LLM request latency histogram |
| `helixagent_llm_request_errors_total` | Error count by provider |
| `helixagent_debate_consensus_score` | AI debate consensus score |
| `helixagent_token_usage_total` | Token usage counter |
| `helixagent_llm_cost_total_usd` | Cost in USD |

### ChromaDB (Port 8001)

| Metric | Description |
|--------|-------------|
| `chromadb_up` | Service availability |
| `chromadb_response_time_ms` | Response time |
| `chromadb_collections_total` | Number of collections |

### Cognee (Port 8000)

| Metric | Description |
|--------|-------------|
| `cognee_up` | Service availability |
| `cognee_response_time_ms` | Response time |

### LLMsVerifier (Port 8180)

| Metric | Description |
|--------|-------------|
| `llmsverifier_up` | Service availability |
| `llmsverifier_response_time_ms` | Response time |
| `llmsverifier_verifications_total` | Total verifications |
| `llmsverifier_providers_verified` | Verified providers |

### PostgreSQL

| Metric | Description |
|--------|-------------|
| `pg_up` | Database availability |
| `pg_stat_activity_count` | Active connections |
| `pg_stat_database_deadlocks` | Deadlock count |
| `pg_stat_activity_max_tx_duration` | Longest transaction |

### Redis

| Metric | Description |
|--------|-------------|
| `redis_up` | Cache availability |
| `redis_memory_used_bytes` | Memory usage |
| `redis_evicted_keys_total` | Evicted keys |
| `redis_rejected_connections_total` | Rejected connections |

## Alert Rules

### Alert Categories

| Category | Alerts | Severity |
|----------|--------|----------|
| **Availability** | HelixAgentDown, ChromaDBDown, CogneeDown, PostgreSQLDown, RedisDown | Critical |
| **Performance** | HighLatency, HighErrorRate, HighResponseTime | Warning |
| **Providers** | AllProvidersDown, LowProviderCount, ProviderHighRateLimit | Warning-Critical |
| **Resources** | HighMemoryUsage, HighCPUUsage, LowDiskSpace | Warning-Critical |
| **Debate** | LowDebateConsensus, DebateTimeout | Warning |
| **Cost** | HighCostPerHour, TokenBudgetExceeded | Warning-Info |
| **MCP** | NoMCPServers, LowMCPServerCount | Warning-Critical |
| **Infrastructure** | HighNodeCPU, HighNodeMemory, ContainerRestarting | Warning |

### Example Alerts

```yaml
# Critical: All LLM providers down
- alert: AllProvidersDown
  expr: helixagent_providers_healthy == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "All LLM providers are down"

# Warning: High response time
- alert: HelixAgentHighResponseTime
  expr: helixagent_response_time_ms > 5000
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "HelixAgent response time above 5s"
```

## Grafana Dashboards

### HelixAgent Overview Dashboard

Panels:
- Service Health Status (all services)
- LLM Provider Status Grid
- Request Rate & Latency Graphs
- Error Rate Trend
- Token Usage & Cost Tracking
- AI Debate Consensus Score
- MCP Server Count

### Messaging Dashboard

Panels:
- Kafka Consumer Lag
- RabbitMQ Queue Depth
- Message Processing Rate
- Error Rates by Topic

## Custom Exporter

The custom HelixAgent exporter (`monitoring/helixagent-exporter.py`) collects metrics from services that don't natively expose Prometheus metrics.

### Endpoints

| Endpoint | Description |
|----------|-------------|
| `/metrics` | Prometheus metrics |
| `/health` | Exporter health check |
| `/` | HTML info page |

### Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `HELIXAGENT_URL` | HelixAgent API URL | http://localhost:7061 |
| `CHROMADB_URL` | ChromaDB URL | http://localhost:8001 |
| `COGNEE_URL` | Cognee URL | http://localhost:8000 |
| `LLMSVERIFIER_URL` | LLMsVerifier URL | http://localhost:8180 |
| `EXPORTER_PORT` | Exporter port | 9200 |

## Challenge Script

The monitoring challenge validates the entire monitoring stack:

```bash
./challenges/scripts/comprehensive_monitoring_challenge.sh
```

### Tests (50+)

1. **Configuration Files** (10 tests)
   - Prometheus config exists
   - All required scrape jobs configured
   - Alert rules file exists
   - Sufficient alert rules (30+)
   - Service-specific alerts configured

2. **Service Health** (6 tests)
   - HelixAgent health
   - ChromaDB health
   - Cognee health
   - PostgreSQL health
   - Redis health
   - Metrics endpoint

3. **Prometheus Integration** (5 tests)
   - Prometheus availability
   - Target scraping
   - Alert rules loaded
   - No firing alerts

4. **Grafana Integration** (3 tests)
   - Grafana availability
   - Datasources configured
   - Dashboards loaded

5. **Alertmanager Integration** (2 tests)
   - Alertmanager availability
   - Cluster status

6. **Loki Integration** (2 tests)
   - Loki availability
   - Labels available

7. **Custom Exporter** (5 tests)
   - Exporter health
   - HelixAgent metrics
   - ChromaDB metrics
   - Cognee metrics

8. **Documentation** (3 tests)
   - Docs exist
   - Comprehensive content
   - Dashboard files exist

9. **Test Coverage** (3 tests)
   - Integration tests exist
   - Unit tests exist
   - Tests pass

10. **LLMsVerifier** (3 tests)
    - Directory exists
    - Prometheus integration
    - Configured in scraper

## Integration Tests

```bash
# Run all monitoring integration tests
go test -v ./tests/integration/comprehensive_monitoring_test.go

# Run specific test
go test -v -run TestPrometheusTargets ./tests/integration/...
```

### Test Coverage

| Test | Description |
|------|-------------|
| `TestHelixAgentHealth` | API health check |
| `TestHelixAgentMetrics` | Metrics endpoint |
| `TestHelixAgentProviders` | Provider monitoring |
| `TestChromaDBHealth` | ChromaDB health |
| `TestCogneeHealth` | Cognee health |
| `TestPrometheusHealth` | Prometheus health |
| `TestPrometheusTargets` | Scrape targets |
| `TestPrometheusAlertRules` | Alert rules |
| `TestGrafanaHealth` | Grafana health |
| `TestAlertmanagerHealth` | Alertmanager health |
| `TestLokiHealth` | Loki health |
| `TestCustomExporterMetrics` | Custom exporter |
| `TestMCPToolSearch` | MCP functionality |
| `TestMonitoringEndpointsLatency` | Response times |
| `TestMonitoringStackIntegration` | Full stack test |

## Troubleshooting

### Common Issues

**1. Prometheus can't scrape HelixAgent**
```bash
# Check if HelixAgent exposes metrics
curl http://localhost:7061/metrics

# Verify Prometheus config
cat monitoring/prometheus.yml | grep helixagent
```

**2. Grafana can't connect to datasources**
```bash
# Check datasource config
cat monitoring/grafana-datasources.yml

# Verify Prometheus is accessible
curl http://localhost:9090/-/healthy
```

**3. Alert rules not loading**
```bash
# Check Prometheus logs
podman logs helixagent-prometheus

# Validate alert rules
promtool check rules monitoring/alert-rules.yml
```

**4. Custom exporter not collecting metrics**
```bash
# Check exporter logs
podman logs helixagent-service-exporter

# Test metrics endpoint
curl http://localhost:9200/metrics
```

## Configuration Files

| File | Description |
|------|-------------|
| `docker-compose.monitoring.yml` | Container orchestration |
| `monitoring/prometheus.yml` | Prometheus scrape configuration |
| `monitoring/alert-rules.yml` | Alert definitions (45+ rules) |
| `monitoring/alertmanager.yml` | Alert routing configuration |
| `monitoring/loki-config.yml` | Log aggregation configuration |
| `monitoring/promtail-config.yml` | Log collection configuration |
| `monitoring/blackbox.yml` | HTTP/TCP probe configuration |
| `monitoring/helixagent-exporter.py` | Custom metrics exporter |
| `monitoring/grafana-datasources.yml` | Grafana datasource provisioning |
| `monitoring/grafana-dashboards.yml` | Dashboard provisioning |
| `monitoring/dashboards/*.json` | Grafana dashboard definitions |

## Best Practices

1. **Alerting**: Set appropriate thresholds based on your SLOs
2. **Retention**: Configure Prometheus retention based on storage capacity
3. **Dashboards**: Create service-specific dashboards for detailed analysis
4. **Labels**: Use consistent labels for filtering and aggregation
5. **Silence**: Use Alertmanager silences during maintenance windows
6. **Recording Rules**: Create recording rules for expensive queries

## Related Documentation

- [MONITORING_SYSTEM.md](MONITORING_SYSTEM.md) - Shell-based monitoring library
- [OpenTelemetry Integration](../observability/OPENTELEMETRY.md) - Tracing setup
- [CLAUDE.md](../../CLAUDE.md) - Project overview
