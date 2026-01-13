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

## Support

- **Documentation**: https://helixagent.ai/docs
- **GitHub Issues**: https://github.com/helixagent/helixagent/issues
- **Email Support**: support@helixagent.ai

---

*Administration Guide Version: 1.0.0*
*Last Updated: January 2026*
