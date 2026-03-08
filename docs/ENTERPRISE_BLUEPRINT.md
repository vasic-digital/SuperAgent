# Enterprise Deployment Blueprint

This document provides architecture guidance, security hardening, operational procedures,
and compliance considerations for deploying HelixAgent in enterprise environments.

Last updated: 2026-03-08

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [High Availability Configuration](#2-high-availability-configuration)
3. [Security Hardening Checklist](#3-security-hardening-checklist)
4. [Monitoring and Alerting](#4-monitoring-and-alerting)
5. [Backup and Disaster Recovery](#5-backup-and-disaster-recovery)
6. [Compliance Considerations](#6-compliance-considerations)
7. [Capacity Planning](#7-capacity-planning)
8. [SLA Targets and Monitoring](#8-sla-targets-and-monitoring)
9. [Operational Runbooks](#9-operational-runbooks)

---

## 1. Architecture Overview

### Production Topology

```
                        Load Balancer (HTTP/3 + TLS)
                               |
              +----------------+----------------+
              |                                 |
       HelixAgent-1                      HelixAgent-2
       (primary)                         (secondary)
              |                                 |
    +---------+---------+             +---------+---------+
    |         |         |             |         |         |
  PostgreSQL Redis  MCP Servers     PostgreSQL Redis  MCP Servers
  (primary)  (primary)              (replica)  (replica)
```

### Component Roles

| Component | Purpose | Instances |
|-----------|---------|-----------|
| HelixAgent | API server, ensemble orchestration, debate engine | 2+ |
| PostgreSQL | Debate sessions, provider scores, audit logs | 1 primary + 1 replica |
| Redis | Cache layer, rate limiting, session state | 1 primary + 1 replica |
| MCP Servers | 65+ containerized Model Context Protocol servers | Per-instance |
| LLM Providers | 22 external API providers | External |
| Prometheus | Metrics collection | 1 |
| Grafana | Dashboards and alerting | 1 |

### Network Requirements

| Connection | Protocol | Port | Direction |
|------------|----------|------|-----------|
| Client to LB | HTTP/3 (QUIC) / HTTP/2 fallback | 443 | Inbound |
| LB to HelixAgent | HTTP/2 | 7061 | Internal |
| HelixAgent to PostgreSQL | TCP | 5432 | Internal |
| HelixAgent to Redis | TCP | 6379 | Internal |
| HelixAgent to LLM APIs | HTTPS | 443 | Outbound |
| HelixAgent to MCP | HTTP | 9101-9999 | Internal |
| Prometheus scrape | HTTP | 9090 | Internal |

---

## 2. High Availability Configuration

### Minimum HA Setup

- **2 HelixAgent instances** behind a load balancer with health check routing
- **PostgreSQL streaming replication** with automatic failover (Patroni or pgBouncer)
- **Redis Sentinel** with 3 nodes for automatic master election
- **Container orchestration** via Kubernetes or Docker Swarm for service restart

### Load Balancer Configuration

```yaml
# Example: health check endpoint for load balancer
health_check:
  path: /v1/monitoring/status
  interval: 10s
  timeout: 5s
  healthy_threshold: 2
  unhealthy_threshold: 3
```

### Instance Configuration

Each instance requires:

```bash
# Environment variables for HA
PORT=7061
GIN_MODE=release

# Database -- point to primary
DB_HOST=postgres-primary.internal
DB_PORT=5432
DB_USER=helixagent
DB_NAME=helixagent_db
DB_SSLMODE=require

# Redis -- point to Sentinel
REDIS_HOST=redis-sentinel.internal
REDIS_PORT=26379
REDIS_SENTINEL_MASTER=helixagent

# Provider API keys (all instances share the same keys)
DEEPSEEK_API_KEY=...
GEMINI_API_KEY=...
# ... all 22 providers
```

### Graceful Shutdown

HelixAgent handles SIGTERM with a graceful shutdown sequence:

1. Stop accepting new requests (drain load balancer)
2. Complete in-flight requests (30s timeout)
3. Close database connections
4. Stop background workers
5. Exit

---

## 3. Security Hardening Checklist

### Authentication and Authorization

- [ ] Enable JWT authentication for all API endpoints
- [ ] Configure API key rotation schedule (90-day maximum)
- [ ] Enable OAuth2 for provider credential management
- [ ] Set rate limiting per API key (default: 100 req/min)
- [ ] Enable CORS with explicit allowed origins (no wildcards in production)

### Transport Security

- [ ] TLS 1.3 minimum for all external connections
- [ ] HTTP/3 (QUIC) enabled with HTTP/2 fallback
- [ ] Brotli compression enabled for response bodies
- [ ] HSTS headers with long max-age (1 year)
- [ ] Certificate pinning for LLM provider connections (optional)

### Data Protection

- [ ] Enable PII detection and redaction via Security module guardrails
- [ ] Configure content filtering policies for input/output
- [ ] Encrypt database at rest (PostgreSQL TDE or filesystem encryption)
- [ ] Encrypt Redis data in transit (TLS) and at rest (if supported)
- [ ] API keys stored encrypted in environment variables or secrets manager

### Container Security

- [ ] Run containers as non-root user
- [ ] Use read-only root filesystems where possible
- [ ] Apply resource limits (CPU, memory) to all containers
- [ ] Scan container images for vulnerabilities (Snyk, Trivy)
- [ ] Use signed container images

### Network Security

- [ ] Internal services not exposed to public network
- [ ] MCP servers accessible only from HelixAgent instances
- [ ] Database accessible only from application tier
- [ ] Egress filtering: allow only LLM provider domains outbound
- [ ] Network policies in Kubernetes for pod-to-pod restrictions

### Scanning and Auditing

- [ ] Run `make security-scan` (gosec) before every release
- [ ] Run Snyk dependency scanning weekly
- [ ] Review suppression list quarterly (see `docs/security/SUPPRESSIONS.md`)
- [ ] Enable audit logging for all API requests
- [ ] Debate provenance tracking enabled for reproducibility

---

## 4. Monitoring and Alerting

### Prometheus Metrics

HelixAgent exposes metrics at `/metrics` for Prometheus scraping.

Key metrics to monitor:

| Metric | Type | Alert Threshold | Description |
|--------|------|----------------|-------------|
| `helixagent_request_duration_seconds` | Histogram | p99 > 30s | API request latency |
| `helixagent_provider_errors_total` | Counter | > 50/min | LLM provider errors |
| `helixagent_circuit_breaker_state` | Gauge | state=open | Provider circuit breaker status |
| `helixagent_debate_rounds_total` | Counter | -- | Debate round throughput |
| `helixagent_cache_hit_ratio` | Gauge | < 0.5 | Cache effectiveness |
| `helixagent_db_pool_active` | Gauge | > 80% capacity | Database pool utilization |
| `helixagent_redis_latency_seconds` | Histogram | p99 > 100ms | Redis operation latency |
| `helixagent_memory_usage_bytes` | Gauge | > 80% limit | Process memory usage |

### Alerting Rules

```yaml
# Example Prometheus alerting rules
groups:
  - name: helixagent
    rules:
      - alert: HighLatency
        expr: histogram_quantile(0.99, helixagent_request_duration_seconds) > 30
        for: 5m
        labels:
          severity: warning

      - alert: ProviderDown
        expr: helixagent_circuit_breaker_state{state="open"} > 0
        for: 1m
        labels:
          severity: critical

      - alert: HighErrorRate
        expr: rate(helixagent_provider_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning

      - alert: DatabasePoolExhausted
        expr: helixagent_db_pool_active / helixagent_db_pool_max > 0.9
        for: 2m
        labels:
          severity: critical
```

### Dashboards

Recommended Grafana dashboards:

1. **Overview** -- Request rate, error rate, latency percentiles, active providers
2. **Provider Health** -- Per-provider latency, error rate, circuit breaker state
3. **Debate Engine** -- Round count, consensus rate, debate duration
4. **Infrastructure** -- DB pool, Redis latency, cache hit rate, memory usage
5. **Cost Tracking** -- Token usage per provider, estimated API cost

### Health Check Endpoints

| Endpoint | Purpose | Expected Response |
|----------|---------|-------------------|
| `GET /v1/monitoring/status` | Full system status | JSON with component states |
| `GET /health` | Simple liveness | `200 OK` |
| `GET /ready` | Readiness (DB + Redis) | `200 OK` when ready |

---

## 5. Backup and Disaster Recovery

### Backup Strategy

| Component | Method | Frequency | Retention |
|-----------|--------|-----------|-----------|
| PostgreSQL | pg_dump + WAL archiving | Continuous WAL, daily full dump | 30 days |
| Redis | RDB snapshots + AOF | Every 5 minutes (RDB), continuous (AOF) | 7 days |
| Configuration | Git repository | On every change | Indefinite |
| Container images | Registry retention | On every build | 90 days |
| LLMsVerifier scores | PostgreSQL (included in DB backup) | With DB backup | 30 days |

### Recovery Procedures

**Database Recovery (PostgreSQL)**:

1. Stop HelixAgent instances
2. Restore from latest pg_dump or point-in-time WAL recovery
3. Verify schema with `make infra-status`
4. Restart HelixAgent instances
5. Verify via `/v1/monitoring/status`

**Redis Recovery**:

1. Redis Sentinel promotes replica automatically
2. If full loss: restart Redis, HelixAgent will repopulate cache on demand
3. No persistent state loss -- Redis is used as cache only

**Full Disaster Recovery**:

1. Provision new infrastructure (containers, networking)
2. Restore PostgreSQL from backup
3. Deploy HelixAgent containers from registry
4. Apply environment configuration from Git
5. Run `make infra-start` to start supporting services
6. Verify all health checks pass
7. Switch DNS/load balancer to new instances

### Recovery Time Objectives

| Scenario | RTO | RPO |
|----------|-----|-----|
| Single instance failure | < 1 minute (auto-restart) | Zero (no data loss) |
| Database failover | < 5 minutes (automatic) | < 1 minute (WAL replication) |
| Full cluster rebuild | < 30 minutes | Last backup (max 24 hours) |

---

## 6. Compliance Considerations

### GDPR

- **PII Detection**: Enable the Security module's PII detector to scan prompts and
  responses for personal data (email, phone, SSN, credit card, IP addresses)
- **Data Minimization**: Configure prompt logging to redact PII before storage
- **Right to Erasure**: Debate sessions can be deleted via the audit API
- **Data Processing Records**: Debate provenance audit trail provides full processing records
- **Cross-Border Transfer**: Ensure LLM provider endpoints comply with data residency
  requirements (EU providers: Mistral; US providers: OpenAI, Anthropic, etc.)

### SOC 2

- **Access Control**: JWT + API key authentication with role-based access
- **Audit Logging**: All API requests logged with timestamp, user, action, and outcome
- **Change Management**: Git-based configuration with conventional commits
- **Availability Monitoring**: Prometheus metrics with SLA tracking
- **Encryption**: TLS in transit, encryption at rest for database

### Data Retention

Configure retention policies per data type:

| Data Type | Default Retention | Configurable |
|-----------|-------------------|--------------|
| Debate sessions | 90 days | Yes |
| Audit logs | 365 days | Yes |
| Provider scores | 30 days | Yes |
| Cache entries | TTL-based (1 hour max) | Yes |
| API request logs | 30 days | Yes |

---

## 7. Capacity Planning

### Resource Requirements per Instance

| Resource | Minimum | Recommended | High Volume |
|----------|---------|-------------|-------------|
| CPU | 2 cores | 4 cores | 8 cores |
| Memory | 2 GB | 4 GB | 8 GB |
| Disk | 10 GB | 50 GB | 100 GB |
| Network | 100 Mbps | 1 Gbps | 10 Gbps |

### Scaling Guidelines

| Metric | Threshold | Action |
|--------|-----------|--------|
| CPU utilization > 70% sustained | 5 minutes | Add instance |
| Memory utilization > 80% | -- | Increase instance memory |
| Request latency p99 > 10s | 5 minutes | Add instance or optimize queries |
| Database connections > 80% pool | -- | Increase pool size or add read replicas |
| Queue depth growing | 5 minutes | Add worker instances |

### Request Throughput Estimates

| Operation | Latency (p50) | Latency (p99) | Max RPS per Instance |
|-----------|---------------|---------------|----------------------|
| Simple completion | 1-5s | 10s | 50 |
| Debate (3 rounds) | 15-30s | 60s | 10 |
| Debate (5 rounds, 5 agents) | 30-90s | 180s | 5 |
| Embedding generation | 200ms | 1s | 200 |
| Format code | 50ms | 500ms | 500 |
| Health check | 5ms | 50ms | 1000 |

### Database Sizing

| Data Type | Growth Rate | Size per 1M Records |
|-----------|-------------|---------------------|
| Debate sessions | ~10 KB/session | 10 GB |
| Debate turns | ~2 KB/turn | 2 GB |
| Audit events | ~500 B/event | 500 MB |
| Provider scores | ~200 B/score | 200 MB |

---

## 8. SLA Targets and Monitoring

### Service Level Objectives (SLOs)

| SLO | Target | Measurement |
|-----|--------|-------------|
| Availability | 99.9% (43 min downtime/month) | Health check pass rate |
| API latency (simple) | p99 < 10s | Request duration histogram |
| API latency (debate) | p99 < 180s | Request duration histogram |
| Error rate | < 0.1% of requests | Error counter / total counter |
| Provider fallback success | > 95% | Fallback counter / failure counter |

### SLI Definitions

- **Availability**: Percentage of 1-minute intervals where `/health` returns 200
- **Latency**: Time from request receipt to response completion
- **Error rate**: 5xx responses divided by total responses
- **Fallback success**: Requests successfully served by a fallback provider after
  primary failure

### Error Budget

With 99.9% availability target:

| Period | Allowed Downtime |
|--------|-----------------|
| Monthly | 43 minutes |
| Quarterly | 2.2 hours |
| Annual | 8.8 hours |

### Monitoring SLA Compliance

```bash
# Check current monitoring status
make monitoring-status

# Check circuit breaker states
make circuit-breakers

# Check provider health
make provider-health

# Check fallback chain status
make fallback-chain

# Force health check on all providers
make force-health-check
```

---

## 9. Operational Runbooks

### Runbook: Provider Circuit Breaker Open

**Symptom**: Alert fires for `ProviderDown`, monitoring shows circuit breaker in open state.

**Steps**:

1. Check which provider is affected: `make circuit-breakers`
2. Verify the provider is actually down (check provider status page)
3. If provider is down: no action needed, fallback chain handles requests
4. If provider is up: reset circuit breaker via `make monitoring-reset-circuits`
5. Monitor for 5 minutes to confirm stability
6. If circuit opens again, check provider API key validity and rate limits

### Runbook: High Database Connection Usage

**Symptom**: `DatabasePoolExhausted` alert fires.

**Steps**:

1. Check active queries: `SELECT * FROM pg_stat_activity WHERE state = 'active'`
2. Identify long-running queries: `SELECT * FROM pg_stat_activity WHERE duration > interval '30s'`
3. Kill stuck queries if necessary: `SELECT pg_terminate_backend(pid)`
4. If pool is genuinely too small, increase `DB_MAX_OPEN_CONNS` and restart
5. Check for connection leaks in application logs

### Runbook: Memory Usage Growing

**Symptom**: Process memory exceeds 80% of container limit.

**Steps**:

1. Check Go heap profile: `go tool pprof http://localhost:7061/debug/pprof/heap`
2. Identify top allocators
3. Check cache sizes (in-memory cache may need lower `MaxEntries`)
4. Check for goroutine leaks: `go tool pprof http://localhost:7061/debug/pprof/goroutine`
5. If a specific debate session is consuming memory, it may need a shorter TTL

### Runbook: LLM Provider Rate Limited

**Symptom**: Provider returns 429 errors, error logs show rate limit exceeded.

**Steps**:

1. Check current rate limit headers in provider responses
2. Verify subscription tier via `internal/verifier/subscription_detector.go`
3. Reduce concurrency for that provider (lower semaphore limit)
4. Enable request queuing with backpressure
5. Consider upgrading the provider subscription tier
6. If free tier: accept rate limits and rely on fallback chain

### Runbook: Startup Verification Failure

**Symptom**: HelixAgent fails to start, logs show verification failures.

**Steps**:

1. Check verification report: `GET /v1/startup/verification`
2. Identify which providers failed verification
3. Check API key validity for failing providers
4. Check network connectivity to provider endpoints
5. Minimum 3 providers must pass for debate team selection
6. If fewer than 3 pass: fix provider configs or add new providers
7. Restart HelixAgent after fixing

### Runbook: Container Rebuild After Code Change

**Symptom**: Code changes deployed but behavior unchanged.

**Steps**:

1. Verify code changes are committed and pulled
2. Rebuild container images: `make docker-build` or `make container-build`
3. Restart containers: `make docker-run` or `make container-start`
4. If using remote distribution: re-run with `CONTAINERS_REMOTE_ENABLED=true`
5. Verify new code is running: check version endpoint or build info
6. Run health checks: `make monitoring-status`

---

## Appendix: Environment Variable Reference

See `.env.example` in the project root for the complete list of configurable environment
variables including all 22 provider API keys, database configuration, Redis configuration,
feature flags, and service overrides.

Key enterprise-relevant variables:

| Variable | Purpose | Default |
|----------|---------|---------|
| `GIN_MODE` | Server mode | `debug` (set to `release` in production) |
| `JWT_SECRET` | JWT signing secret | Required |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` (set to `require` in production) |
| `REDIS_PASSWORD` | Redis authentication | Required in production |
| `CONSTITUTION_WATCHER_ENABLED` | Auto-update constitution | `false` |
| `COGNEE_ENABLED` | Knowledge graph memory | `false` |
