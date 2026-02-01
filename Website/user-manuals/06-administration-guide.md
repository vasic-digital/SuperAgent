# HelixAgent Administration Guide

## Overview

This guide covers administrative tasks for managing HelixAgent in production environments.

## User Management

### API Key Administration

```bash
# Generate new API key
curl -X POST http://localhost:7061/admin/api-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name": "production-app", "permissions": ["read", "write"]}'

# List API keys
curl http://localhost:7061/admin/api-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Revoke API key
curl -X DELETE http://localhost:7061/admin/api-keys/{key_id} \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Permission Levels

| Level | Permissions |
|-------|-------------|
| Read | View completions, list models |
| Write | Create completions, run debates |
| Admin | Manage keys, view metrics, configure providers |

## Provider Management

### Adding Providers

```yaml
# configs/providers.yaml
providers:
  new_provider:
    enabled: true
    api_key: ${NEW_PROVIDER_API_KEY}
    models:
      - model-a
      - model-b
    weight: 1.0
    timeout: 30s
```

### Provider Health Monitoring

```bash
# Check all providers
curl http://localhost:7061/v1/providers/health

# Check specific provider
curl http://localhost:7061/v1/providers/claude/health

# Trigger manual verification
curl -X POST http://localhost:7061/v1/providers/verify \
  -d '{"provider": "claude"}'
```

### Provider Failover Configuration

```yaml
fallback:
  enabled: true
  chain:
    - claude    # Primary
    - deepseek  # First fallback
    - gemini    # Second fallback
  max_retries: 3
  retry_delay: 1s
```

## Performance Tuning

### Resource Configuration

```yaml
# configs/performance.yaml
server:
  workers: ${CPU_CORES * 2}
  max_connections: 1000
  request_timeout: 30s
  idle_timeout: 120s

database:
  pool:
    max_conns: 25
    min_conns: 5
    max_conn_lifetime: 1h

cache:
  l1:
    max_size: 10000
    ttl: 5m
  l2:
    enabled: true
    ttl: 1h
```

### Memory Optimization

```yaml
optimization:
  gc_percent: 100  # Go garbage collection percentage
  memory_limit: "2GB"  # Max memory usage
  request_buffer: 1024  # Request buffer size KB
```

### Connection Pooling

```yaml
http_pool:
  max_idle_conns: 100
  max_conns_per_host: 10
  idle_conn_timeout: 90s
```

## Security Configuration

### Authentication Setup

```yaml
auth:
  jwt:
    enabled: true
    secret: ${JWT_SECRET}
    expiration: 24h
  api_key:
    enabled: true
    header: "Authorization"
```

### Rate Limiting

```yaml
rate_limit:
  enabled: true
  requests_per_minute: 100
  tokens_per_minute: 10000
  burst: 20
```

### CORS Configuration

```yaml
cors:
  allowed_origins:
    - "https://your-domain.com"
  allowed_methods:
    - GET
    - POST
    - DELETE
  allowed_headers:
    - Content-Type
    - Authorization
  max_age: 3600
```

### TLS Configuration

```yaml
tls:
  enabled: true
  cert_file: /path/to/cert.pem
  key_file: /path/to/key.pem
  min_version: "TLS1.2"
```

## Backup and Recovery

### Database Backup

```bash
# Backup PostgreSQL
pg_dump -h localhost -U helixagent -d helixagent > backup_$(date +%Y%m%d).sql

# Restore from backup
psql -h localhost -U helixagent -d helixagent < backup_20260113.sql
```

### Configuration Backup

```bash
# Backup all configs
tar -czvf configs_backup_$(date +%Y%m%d).tar.gz configs/

# Backup with secrets (encrypted)
tar -cvf - configs/ | gpg -c > configs_backup_$(date +%Y%m%d).tar.gpg
```

### Cache Backup (Redis)

```bash
# Create RDB snapshot
redis-cli BGSAVE

# Copy RDB file
cp /var/lib/redis/dump.rdb /backup/redis_$(date +%Y%m%d).rdb
```

## Monitoring Setup

### Prometheus Metrics

Enable metrics endpoint:
```yaml
metrics:
  enabled: true
  endpoint: /metrics
  auth_required: false
```

Key metrics to monitor:
- `helixagent_requests_total`
- `helixagent_request_duration_seconds`
- `helixagent_provider_health`
- `helixagent_cache_hit_rate`
- `helixagent_active_debates`

### Alert Configuration

```yaml
# alerts/rules.yml
groups:
  - name: helixagent
    rules:
      - alert: ProviderDown
        expr: helixagent_provider_health == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Provider {{ $labels.provider }} is down"

      - alert: HighErrorRate
        expr: rate(helixagent_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
```

### Log Management

```yaml
logging:
  level: info
  format: json
  output: stdout
  file:
    enabled: true
    path: /var/log/helixagent/app.log
    max_size: 100MB
    max_backups: 5
```

## Health Checks

### Endpoints

| Endpoint | Purpose |
|----------|---------|
| `/healthz/live` | Liveness check |
| `/healthz/ready` | Readiness check |
| `/health` | Full health status |

### Kubernetes Probes

```yaml
livenessProbe:
  httpGet:
    path: /healthz/live
    port: 7061
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /healthz/ready
    port: 7061
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Troubleshooting

### Common Issues

#### Provider Authentication Fails
```bash
# Check API key
curl -v http://localhost:7061/v1/providers/claude/health

# Verify environment variable
echo $CLAUDE_API_KEY
```

#### High Memory Usage
```bash
# Check memory stats
curl http://localhost:7061/debug/vars

# Force garbage collection
curl -X POST http://localhost:7061/admin/gc
```

#### Slow Responses
```bash
# Check provider latencies
curl http://localhost:7061/v1/providers/health | jq '.providers[].latency_ms'

# Enable debug logging
export HELIXAGENT_DEBUG=true
```

### Debug Mode

```bash
# Enable debug mode
export HELIXAGENT_DEBUG=true
./bin/helixagent

# Debug specific component
export HELIXAGENT_DEBUG_PROVIDERS=true
export HELIXAGENT_DEBUG_CACHE=true
```

## Maintenance Tasks

### Regular Maintenance Checklist

Daily:
- [ ] Check provider health
- [ ] Review error logs
- [ ] Monitor response times

Weekly:
- [ ] Backup database
- [ ] Review metrics trends
- [ ] Update API keys if needed

Monthly:
- [ ] Update dependencies
- [ ] Review security logs
- [ ] Performance audit
- [ ] Backup verification

### Upgrading HelixAgent

```bash
# Pull latest version
git pull origin main

# Rebuild
make build

# Test in staging
./bin/helixagent --config configs/staging.yaml

# Restart production
systemctl restart helixagent
```

## Role-Based Access Control (RBAC)

### Role Definitions

| Role | Description | Permissions |
|------|-------------|-------------|
| `admin` | Full system access | All operations |
| `operator` | Operations management | Provider management, monitoring |
| `developer` | API access | Completions, debates, embeddings |
| `viewer` | Read-only access | View models, health checks |

### Creating Roles

```yaml
# configs/rbac.yaml
roles:
  custom_analyst:
    description: "Data analysis role"
    permissions:
      - "completions:read"
      - "completions:write"
      - "debates:read"
      - "debates:write"
      - "embeddings:read"
      - "embeddings:write"
    denied:
      - "admin:*"
      - "providers:write"
```

### Assigning Roles to Users

```bash
# Assign role to user
curl -X POST http://localhost:7061/admin/users/{user_id}/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"role": "developer"}'

# List user roles
curl http://localhost:7061/admin/users/{user_id}/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Remove role
curl -X DELETE http://localhost:7061/admin/users/{user_id}/roles/developer \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Permission Inheritance

```yaml
rbac:
  inheritance:
    admin:
      inherits: [operator, developer, viewer]
    operator:
      inherits: [viewer]
    developer:
      inherits: [viewer]
```

## Audit Logging

### Configuration

```yaml
audit:
  enabled: true
  log_path: /var/log/helixagent/audit.log
  format: json
  events:
    - authentication
    - authorization
    - api_calls
    - admin_actions
    - provider_changes
  retention_days: 90
  include_request_body: false  # Privacy consideration
  include_response: false
```

### Audit Log Format

```json
{
  "timestamp": "2026-01-22T10:30:00Z",
  "event_type": "api_call",
  "user_id": "user-123",
  "action": "POST /v1/completions",
  "ip_address": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "result": "success",
  "duration_ms": 1250,
  "provider_used": "claude",
  "tokens_in": 100,
  "tokens_out": 500
}
```

### Querying Audit Logs

```bash
# View recent authentication events
curl "http://localhost:7061/admin/audit?event_type=authentication&limit=100" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Filter by user
curl "http://localhost:7061/admin/audit?user_id=user-123&since=2026-01-01" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Export audit log
curl "http://localhost:7061/admin/audit/export?format=csv&from=2026-01-01&to=2026-01-31" \
  -H "Authorization: Bearer $ADMIN_TOKEN" > audit_january.csv
```

### Audit Log Retention

```bash
# Archive old logs
./scripts/archive-audit-logs.sh --older-than 90d --compress

# Verify log integrity
./scripts/verify-audit-logs.sh --from 2026-01-01 --to 2026-01-31
```

## API Key Rotation

### Rotation Procedure

```bash
# 1. Generate new key with same permissions
NEW_KEY=$(curl -X POST http://localhost:7061/admin/api-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name": "production-app-v2", "permissions": ["read", "write"]}' | jq -r '.key')

# 2. Update applications with new key (grace period)
# Allow both old and new keys to work

# 3. Monitor for old key usage
curl "http://localhost:7061/admin/api-keys/{old_key_id}/usage" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# 4. Revoke old key after migration complete
curl -X DELETE http://localhost:7061/admin/api-keys/{old_key_id} \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Automated Key Rotation

```yaml
# configs/key-rotation.yaml
key_rotation:
  enabled: true
  schedule: "0 0 1 * *"  # Monthly
  grace_period: 7d
  notify_before: 14d
  notifications:
    slack: https://hooks.slack.com/services/xxx
    email: admin@company.com
```

### Key Expiration Policies

```yaml
api_keys:
  policies:
    default:
      max_age: 90d
      require_rotation: true
    service_account:
      max_age: 365d
      require_rotation: true
    temporary:
      max_age: 24h
      require_rotation: false
```

## Disaster Recovery

### Recovery Point Objective (RPO)

| Data Type | RPO | Backup Frequency |
|-----------|-----|------------------|
| Database | 1 hour | Hourly snapshots |
| Configuration | 24 hours | Daily |
| Audit Logs | 24 hours | Daily |
| Redis Cache | N/A | Ephemeral |

### Recovery Time Objective (RTO)

| Scenario | RTO | Procedure |
|----------|-----|-----------|
| Single node failure | 5 minutes | Auto-failover |
| Database corruption | 30 minutes | Restore from snapshot |
| Complete site failure | 4 hours | DR site activation |

### Disaster Recovery Procedure

```bash
# 1. Assess damage
./scripts/dr-assess.sh

# 2. Activate DR site (if needed)
./scripts/dr-activate.sh --site us-west-2

# 3. Restore database
./scripts/db-restore.sh --snapshot latest --verify

# 4. Restore configuration
./scripts/config-restore.sh --backup configs_backup_latest.tar.gpg

# 5. Verify services
./scripts/health-check.sh --all

# 6. Update DNS (if site switch)
./scripts/dns-failover.sh --activate dr-site
```

### DR Testing Schedule

```yaml
disaster_recovery:
  testing:
    database_restore: monthly
    config_restore: monthly
    full_dr_drill: quarterly
    document_review: annually
```

## Scaling and High Availability

### Horizontal Scaling

```yaml
# configs/scaling.yaml
scaling:
  horizontal:
    enabled: true
    min_replicas: 2
    max_replicas: 10
    metrics:
      - name: cpu_utilization
        target: 70%
      - name: memory_utilization
        target: 80%
      - name: requests_per_second
        target: 100
```

### Load Balancing

```yaml
load_balancing:
  strategy: "round_robin"
  health_check:
    path: /healthz/ready
    interval: 10s
    timeout: 2s
    healthy_threshold: 2
    unhealthy_threshold: 3
  session_affinity: false
```

### High Availability Configuration

```yaml
high_availability:
  enabled: true
  node_count: 3
  quorum: 2
  auto_failover: true
  failover_timeout: 30s
  data_replication:
    mode: "async"
    consistency: "eventual"
```

### Multi-region Deployment

```yaml
multi_region:
  enabled: false
  regions:
    - name: us-east-1
      weight: 1.0
      primary: true
    - name: eu-west-1
      weight: 0.7
      primary: false
  routing:
    strategy: "latency_based"
    failover_enabled: true
```

### Kubernetes Deployment

```yaml
# helm/values.yaml
replicaCount: 3
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchLabels:
            app: helixagent
        topologyKey: kubernetes.io/hostname
resources:
  limits:
    cpu: 2
    memory: 4Gi
  requests:
    cpu: 1
    memory: 2Gi
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

## Security Hardening

### Network Security

```yaml
# configs/security.yaml
network:
  bind_address: "127.0.0.1"  # Bind to localhost, use reverse proxy
  allowed_ips:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
  blocked_ips: []
  rate_limit:
    enabled: true
    per_ip: 100/minute
    per_user: 1000/minute
```

### API Security

```yaml
security:
  api:
    require_https: true
    min_tls_version: "1.2"
    hsts:
      enabled: true
      max_age: 31536000
      include_subdomains: true
    content_security_policy: "default-src 'self'"
    x_frame_options: "DENY"
    x_content_type_options: "nosniff"
```

### Secret Management

```yaml
secrets:
  provider: "vault"  # or "aws-secrets-manager", "env"
  vault:
    address: "https://vault.internal:8200"
    auth_method: "kubernetes"
    secret_path: "secret/data/helixagent"
  refresh_interval: 5m
```

### Security Scanning

```bash
# Run security scan
make security-scan

# Check for vulnerabilities
govulncheck ./...

# Scan container image
trivy image helixagent:latest

# Static analysis
gosec ./...
```

### Hardening Checklist

- [ ] Disable debug mode in production
- [ ] Use strong JWT secrets (256+ bits)
- [ ] Enable TLS for all connections
- [ ] Configure firewall rules
- [ ] Set up intrusion detection
- [ ] Enable audit logging
- [ ] Implement rate limiting
- [ ] Use read-only filesystems where possible
- [ ] Run as non-root user
- [ ] Keep dependencies updated

## Compliance

### SOC 2 Requirements

| Control | Implementation | Evidence |
|---------|----------------|----------|
| Access Control | RBAC with audit logging | Audit logs, role configs |
| Data Encryption | TLS 1.2+, AES-256 at rest | TLS configs, encryption settings |
| Monitoring | Prometheus metrics, alerting | Dashboards, alert rules |
| Incident Response | Documented procedures | Runbooks, incident reports |
| Change Management | Git-based config management | Git history, PR reviews |

### GDPR Compliance

```yaml
gdpr:
  data_retention:
    request_logs: 30d
    user_data: "until_deletion_request"
    audit_logs: 90d
  data_subject_rights:
    export_endpoint: /api/v1/users/{id}/export
    delete_endpoint: /api/v1/users/{id}/delete
  consent_management:
    enabled: true
    purposes:
      - service_provision
      - analytics
      - marketing
```

### Data Subject Requests

```bash
# Export user data (GDPR Article 20)
curl -X POST http://localhost:7061/admin/gdpr/export \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"user_id": "user-123", "format": "json"}'

# Delete user data (GDPR Article 17)
curl -X POST http://localhost:7061/admin/gdpr/delete \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"user_id": "user-123", "confirm": true}'
```

### Compliance Reporting

```bash
# Generate compliance report
curl -X POST http://localhost:7061/admin/compliance/report \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"type": "soc2", "period": "2026-Q1"}' > soc2_q1_2026.pdf

# List compliance status
curl http://localhost:7061/admin/compliance/status \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Support

- **Documentation**: https://helixagent.ai/docs
- **GitHub Issues**: https://github.com/helixagent/helixagent/issues
- **Email Support**: support@helixagent.ai
- **Security Issues**: security@helixagent.ai

---

*Administration Guide Version: 2.0.0*
*Last Updated: January 2026*
