# HelixAgent Operations Runbook

## Overview

This runbook provides step-by-step procedures for common operational tasks, incident response, and troubleshooting for HelixAgent deployments.

## Table of Contents

1. [Startup & Shutdown](#startup--shutdown)
2. [Health Checks](#health-checks)
3. [Monitoring & Alerting](#monitoring--alerting)
4. [Common Issues](#common-issues)
5. [Incident Response](#incident-response)
6. [Maintenance Procedures](#maintenance-procedures)
7. [Backup & Recovery](#backup--recovery)
8. [Scaling Operations](#scaling-operations)

---

## Startup & Shutdown

### Starting HelixAgent

#### Docker Compose

```bash
# Start all services
docker-compose up -d

# Verify services are running
docker-compose ps

# Check logs
docker-compose logs -f helixagent
```

#### Kubernetes

```bash
# Apply deployment
kubectl apply -f k8s/deployment.yaml

# Verify pods
kubectl get pods -l app=helixagent

# Check logs
kubectl logs -f deployment/helixagent
```

#### Systemd

```bash
# Start service
sudo systemctl start helixagent

# Check status
sudo systemctl status helixagent

# View logs
journalctl -u helixagent -f
```

### Graceful Shutdown

```bash
# Docker
docker-compose down

# Kubernetes
kubectl scale deployment helixagent --replicas=0

# Systemd
sudo systemctl stop helixagent
```

**Important**: Allow at least 30 seconds for graceful shutdown to complete in-flight requests.

---

## Health Checks

### Endpoint Health

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed

# Ready check (for k8s)
curl http://localhost:8080/ready
```

### Expected Response

```json
{
  "status": "healthy",
  "components": {
    "database": {"status": "healthy", "latency_ms": 5},
    "redis": {"status": "healthy", "latency_ms": 2},
    "providers": {"status": "healthy", "active": 8}
  },
  "uptime": "48h32m15s",
  "version": "1.0.0"
}
```

### Provider Health

```bash
# Check all providers
curl http://localhost:8080/v1/providers/health

# Check specific provider
curl http://localhost:8080/v1/providers/claude/health
```

### Database Health

```bash
# Check PostgreSQL
psql -h localhost -U helixagent -c "SELECT 1"

# Check connection pool
curl http://localhost:8080/metrics | grep db_pool
```

### Cache Health

```bash
# Check Redis
redis-cli ping

# Check cache stats
curl http://localhost:8080/metrics | grep cache
```

---

## Monitoring & Alerting

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `helixagent_request_duration_p99` | 99th percentile latency | > 5s |
| `helixagent_error_rate` | Error rate | > 5% |
| `helixagent_provider_health` | Provider availability | < 0.5 |
| `helixagent_cache_hit_rate` | Cache effectiveness | < 0.3 |
| `helixagent_db_pool_waiting` | DB connection contention | > 10 |

### Prometheus Queries

```promql
# Request rate
rate(helixagent_requests_total[5m])

# Error rate
rate(helixagent_errors_total[5m]) / rate(helixagent_requests_total[5m])

# P99 latency
histogram_quantile(0.99, rate(helixagent_request_duration_bucket[5m]))

# Provider health
avg(helixagent_provider_health) by (provider)
```

### Alert Rules

```yaml
groups:
  - name: helixagent
    rules:
      - alert: HighErrorRate
        expr: rate(helixagent_errors_total[5m]) / rate(helixagent_requests_total[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanize }}"

      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(helixagent_request_duration_bucket[5m])) > 5
        for: 5m
        annotations:
          summary: "High latency detected"
          description: "P99 latency is {{ $value | humanize }}s"

      - alert: ProviderUnhealthy
        expr: helixagent_provider_health < 0.5
        for: 2m
        annotations:
          summary: "Provider unhealthy"
          description: "Provider {{ $labels.provider }} is unhealthy"
```

---

## Common Issues

### Issue: High Latency

**Symptoms:**
- P99 latency > 5 seconds
- Users reporting slow responses

**Investigation:**
```bash
# Check provider latencies
curl http://localhost:8080/v1/providers | jq '.[] | {name, latency_ms}'

# Check database latency
curl http://localhost:8080/metrics | grep db_query_duration

# Check cache hit rate
curl http://localhost:8080/metrics | grep cache_hit
```

**Resolution:**
1. If provider latency high → Consider switching providers or enabling fallbacks
2. If database latency high → Check connection pool, run VACUUM ANALYZE
3. If cache hit rate low → Review cache TTLs, increase cache size

### Issue: Provider Failures

**Symptoms:**
- Specific provider returning errors
- Fallback providers being used

**Investigation:**
```bash
# Check provider status
curl http://localhost:8080/v1/providers/claude/health

# Check error logs
grep "provider error" /var/log/helixagent/app.log | tail -100

# Check rate limits
curl http://localhost:8080/metrics | grep rate_limit
```

**Resolution:**
1. API key expired → Rotate API key
2. Rate limited → Reduce request rate or add providers
3. Service outage → Monitor provider status page, use fallbacks

### Issue: Database Connection Issues

**Symptoms:**
- "connection refused" or "too many connections" errors
- Database health check failing

**Investigation:**
```bash
# Check PostgreSQL
psql -h localhost -U helixagent -c "SELECT count(*) FROM pg_stat_activity"

# Check connection pool
curl http://localhost:8080/metrics | grep db_pool

# Check for long-running queries
psql -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query
         FROM pg_stat_activity
         WHERE state = 'active'
         ORDER BY duration DESC LIMIT 10"
```

**Resolution:**
1. Too many connections → Increase pool size or kill idle connections
2. Long-running queries → Identify and optimize or kill
3. Connection refused → Restart PostgreSQL, check disk space

### Issue: Memory Issues

**Symptoms:**
- OOM kills
- High memory usage alerts

**Investigation:**
```bash
# Check Go memory stats
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Check container memory
docker stats helixagent

# Check for goroutine leaks
curl http://localhost:8080/debug/pprof/goroutine?debug=1 | head -100
```

**Resolution:**
1. Goroutine leak → Identify and fix code, restart service
2. Cache too large → Reduce cache size
3. Memory pressure → Scale horizontally or increase memory limits

---

## Incident Response

### Severity Levels

| Level | Description | Response Time | Example |
|-------|-------------|---------------|---------|
| P1 | Complete outage | Immediate | Service unreachable |
| P2 | Major degradation | < 15 min | High error rate |
| P3 | Minor degradation | < 1 hour | Single provider down |
| P4 | Cosmetic/minor | < 24 hours | Slow non-critical endpoint |

### P1 Incident Procedure

1. **Acknowledge** (< 5 min)
   - Acknowledge alert
   - Join incident channel
   - Assign incident commander

2. **Assess** (< 10 min)
   ```bash
   # Quick health check
   curl http://localhost:8080/health

   # Check recent deploys
   kubectl get pods -o wide

   # Check error logs
   kubectl logs -l app=helixagent --since=10m | grep ERROR
   ```

3. **Mitigate** (< 30 min)
   - Rollback if recent deploy: `kubectl rollout undo deployment/helixagent`
   - Enable maintenance mode if needed
   - Communicate status to stakeholders

4. **Resolve**
   - Fix root cause
   - Verify fix
   - Monitor for recurrence

5. **Post-mortem** (within 48 hours)
   - Document timeline
   - Identify root cause
   - Define action items

### Communication Templates

**Initial Notification:**
```
[P1 INCIDENT] HelixAgent Service Degradation

Status: Investigating
Impact: API requests failing
Started: HH:MM UTC
Updates: #incident-channel
```

**Resolution:**
```
[RESOLVED] HelixAgent Service Degradation

Duration: X hours Y minutes
Root Cause: [Brief description]
Resolution: [What was done]
Post-mortem: [Link]
```

---

## Maintenance Procedures

### Deploying Updates

```bash
# Build new image
docker build -t helixagent:v1.2.0 .

# Push to registry
docker push registry.example.com/helixagent:v1.2.0

# Update Kubernetes
kubectl set image deployment/helixagent helixagent=registry.example.com/helixagent:v1.2.0

# Watch rollout
kubectl rollout status deployment/helixagent
```

### Database Migrations

```bash
# Run migrations
make migrate-up

# Verify migration
psql -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5"

# Rollback if needed
make migrate-down
```

### Certificate Rotation

```bash
# Generate new certificates
./scripts/generate-certs.sh

# Update secrets
kubectl create secret tls helixagent-tls \
  --cert=certs/server.crt \
  --key=certs/server.key \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods to pick up new certs
kubectl rollout restart deployment/helixagent
```

### API Key Rotation

```bash
# Generate new key
./helixagent keys rotate --name production-app

# Update client configuration
# Old key remains valid for 24 hours

# Verify new key works
curl -H "X-API-Key: new-key" http://localhost:8080/v1/completions
```

---

## Backup & Recovery

### Database Backup

```bash
# Manual backup
pg_dump -Fc helixagent_db > backup_$(date +%Y%m%d_%H%M%S).dump

# Automated backup (via cron)
0 2 * * * /usr/local/bin/backup-helixagent.sh
```

### Database Restore

```bash
# Stop application
kubectl scale deployment/helixagent --replicas=0

# Restore database
pg_restore -d helixagent_db backup.dump

# Verify data
psql -c "SELECT count(*) FROM completions"

# Start application
kubectl scale deployment/helixagent --replicas=3
```

### Configuration Backup

```bash
# Export secrets
kubectl get secrets -o yaml > secrets-backup.yaml

# Export configmaps
kubectl get configmaps -o yaml > configmaps-backup.yaml

# Store securely (encrypted)
gpg -c secrets-backup.yaml
```

---

## Scaling Operations

### Horizontal Scaling

```bash
# Manual scale
kubectl scale deployment/helixagent --replicas=5

# Auto-scale based on CPU
kubectl autoscale deployment/helixagent --min=3 --max=10 --cpu-percent=70
```

### Vertical Scaling

```yaml
# Update resource limits
resources:
  requests:
    memory: "2Gi"
    cpu: "1000m"
  limits:
    memory: "4Gi"
    cpu: "2000m"
```

### Database Scaling

```bash
# Add read replica
# 1. Create replica
# 2. Configure connection string
# 3. Update config to use replica for reads

# Scale connection pool
# Update config: database.pool.max_open: 100
```

---

## Contacts

| Role | Contact | Escalation |
|------|---------|------------|
| On-call Engineer | oncall@example.com | PagerDuty |
| Database Admin | dba@example.com | Slack #db-help |
| Security Team | security@example.com | Slack #security |
| Management | manager@example.com | Phone |

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
