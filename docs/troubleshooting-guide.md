# SuperAgent Troubleshooting Guide

Comprehensive troubleshooting guide for common issues and debugging procedures.

## Quick Diagnosis

### Health Check Commands

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/v1/health

# Provider health check
curl http://localhost:8080/v1/providers/health

# Database connectivity test
docker-compose exec postgres pg_isready -U superagent -d superagent_db

# Redis connectivity test
docker-compose exec redis redis-cli ping
```

### Log Analysis

```bash
# View application logs
docker-compose logs -f superagent

# View all service logs
docker-compose logs -f

# Filter logs by time
docker-compose logs --since "1h" superagent

# Search for specific errors
docker-compose logs superagent | grep -i error

# View logs with timestamps
docker-compose logs -f -t superagent
```

## Common Issues and Solutions

### 1. Application Won't Start

#### Symptoms
- Container exits immediately
- "CrashLoopBackOff" in Kubernetes
- Port 8080 not accessible

#### Diagnosis
```bash
# Check container status
docker-compose ps

# View startup logs
docker-compose logs superagent

# Check resource usage
docker stats

# Verify configuration
docker-compose config
```

#### Solutions

**Database Connection Failed:**
```bash
# Wait for database to be ready
docker-compose exec postgres pg_isready -U superagent -d superagent_db

# Check database logs
docker-compose logs postgres

# Verify database credentials in .env
cat .env | grep DB_
```

**Redis Connection Failed:**
```bash
# Test Redis connectivity
docker-compose exec redis redis-cli ping

# Check Redis password configuration
docker-compose exec redis redis-cli -a $REDIS_PASSWORD ping
```

**Port Already in Use:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 $(lsof -t -i :8080)

# Or use different port
echo "PORT=8081" >> .env
docker-compose up -d
```

### 2. API Returns 500 Errors

#### Symptoms
- HTTP 500 Internal Server Error
- Application logs show panics
- Requests fail intermittently

#### Diagnosis
```bash
# Enable debug logging
export LOG_LEVEL=debug
docker-compose up -d

# Monitor error logs
docker-compose logs -f superagent | grep -i error

# Check system resources
docker stats
free -h
df -h
```

#### Solutions

**Out of Memory:**
```yaml
# Increase memory limits
services:
  superagent:
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G
```

**Database Connection Pool Exhausted:**
```bash
# Check active connections
docker-compose exec postgres psql -U superagent -d superagent_db -c "SELECT count(*) FROM pg_stat_activity;"

# Increase pool size
echo "DB_MAX_CONNECTIONS=20" >> .env
```

**Provider API Rate Limits:**
```bash
# Check provider status
curl http://localhost:8080/v1/providers/health

# Implement exponential backoff
# Check logs for rate limit messages
docker-compose logs superagent | grep -i "rate limit"
```

### 3. Slow Response Times

#### Symptoms
- API responses take >5 seconds
- Timeout errors
- High latency in monitoring

#### Diagnosis
```bash
# Test response time
time curl http://localhost:8080/health

# Check Prometheus metrics
curl http://localhost:9090/metrics | grep superagent_response_time

# Monitor database performance
docker-compose exec postgres psql -U superagent -d superagent_db -c "SELECT * FROM pg_stat_activity;"

# Check Redis performance
docker-compose exec redis redis-cli info stats
```

#### Solutions

**Database Query Optimization:**
```sql
-- Check slow queries
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE state = 'active' AND now() - pg_stat_activity.query_start > interval '1 second'
ORDER BY duration DESC;

-- Create missing indexes
CREATE INDEX CONCURRENTLY idx_llm_requests_created_at ON llm_requests(created_at);
CREATE INDEX CONCURRENTLY idx_debates_status ON debates(status);
```

**Cache Configuration:**
```bash
# Check cache hit ratio
docker-compose exec redis redis-cli info stats | grep keyspace_hits

# Increase cache TTL
echo "CACHE_TTL=7200" >> .env

# Clear cache if corrupted
docker-compose exec redis redis-cli FLUSHALL
```

**Provider Timeouts:**
```bash
# Increase provider timeouts
echo "CLAUDE_TIMEOUT=60" >> .env
echo "DEEPSEEK_TIMEOUT=60" >> .env

# Switch to faster providers
curl -X POST http://localhost:8080/v1/completions \
  -d '{"model": "qwen-turbo", "prompt": "test"}'
```

### 4. Authentication Issues

#### Symptoms
- HTTP 401 Unauthorized
- JWT token rejected
- API key not working

#### Diagnosis
```bash
# Test authentication endpoint
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "test", "password": "test"}'

# Check JWT secret configuration
docker-compose exec superagent env | grep JWT

# Verify API key format
echo $SUPERAGENT_API_KEY | head -c 10
```

#### Solutions

**JWT Secret Misconfiguration:**
```bash
# Generate secure JWT secret
openssl rand -hex 32

# Update environment
echo "JWT_SECRET=your_generated_secret" >> .env
docker-compose restart superagent
```

**API Key Format Issues:**
```bash
# Check API key format (should start with sk-)
echo $CLAUDE_API_KEY | grep "^sk-"

# Test provider API directly
curl -H "Authorization: Bearer $CLAUDE_API_KEY" \
  https://api.anthropic.com/v1/messages
```

**CORS Issues:**
```bash
# Check CORS headers
curl -I http://localhost:8080/v1/models

# Add allowed origins
echo "CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com" >> .env
```

### 5. AI Debate Issues

#### Symptoms
- Debates fail to start
- Participants don't respond
- Consensus never reached

#### Diagnosis
```bash
# Check debate status
curl http://localhost:8080/v1/debates/your-debate-id/status

# View debate logs
docker-compose logs superagent | grep "debate-id"

# Test individual LLM calls
curl -X POST http://localhost:8080/v1/completions \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"model": "claude-3-haiku", "prompt": "Hello"}'
```

#### Solutions

**Participant LLM Configuration:**
```json
{
  "participants": [
    {
      "name": "Expert",
      "llms": [
        {
          "provider": "claude",
          "model": "claude-3-5-sonnet-20241022",
          "api_key": "sk-ant-api03-..."
        }
      ]
    }
  ]
}
```

**Cognee Configuration Issues:**
```bash
# Check Cognee settings
curl http://localhost:8080/v1/debates/your-debate-id

# Verify dataset exists
echo "COGNEE_DATASET=your_dataset_name" >> .env
```

**Consensus Threshold Too High:**
```json
{
  "consensus_threshold": 0.6,
  "maximal_repeat_rounds": 5
}
```

### 6. Model Context Protocol (MCP) Issues

#### Symptoms
- MCP capabilities return empty
- Tool calls fail
- Server connections fail

#### Diagnosis
```bash
# Check MCP configuration
curl http://localhost:8080/mcp/capabilities

# Test tool listing
curl http://localhost:8080/mcp/tools

# View MCP logs
docker-compose logs superagent | grep -i mcp
```

#### Solutions

**MCP Not Enabled:**
```bash
echo "MCP_ENABLED=true" >> .env
docker-compose restart superagent
```

**Server Connection Issues:**
```bash
# Check MCP server configurations
curl http://localhost:8080/v1/admin/mcp/servers

# Test individual server connections
docker-compose exec superagent ./superagent mcp test-server
```

### 7. Database Issues

#### Symptoms
- "Connection refused" errors
- Slow queries
- Data corruption

#### Diagnosis
```bash
# Check database status
docker-compose ps postgres

# View database logs
docker-compose logs postgres

# Test database connectivity
docker-compose exec postgres pg_isready -U superagent -d superagent_db

# Check disk space
df -h /var/lib/postgresql/data

# Monitor active connections
docker-compose exec postgres psql -U superagent -d superagent_db -c "SELECT count(*) FROM pg_stat_activity;"
```

#### Solutions

**Database Full:**
```bash
# Check disk usage
docker-compose exec postgres du -sh /var/lib/postgresql/data

# Clean up old data
docker-compose exec postgres psql -U superagent -d superagent_db -c "DELETE FROM llm_requests WHERE created_at < now() - interval '30 days';"

# Vacuum database
docker-compose exec postgres psql -U superagent -d superagent_db -c "VACUUM ANALYZE;"
```

**Connection Pool Exhaustion:**
```bash
# Increase connection pool
echo "DB_MAX_CONNECTIONS=20" >> .env

# Check for connection leaks
docker-compose exec postgres psql -U superagent -d superagent_db -c "
SELECT pid, usename, application_name, client_addr, state, state_change
FROM pg_stat_activity
WHERE state = 'idle in transaction'
ORDER BY state_change ASC;
"
```

**Replication Issues:**
```bash
# Check replication status
docker-compose exec postgres psql -U superagent -d superagent_db -c "SELECT * FROM pg_stat_replication;"

# Reinitialize replica
docker-compose exec postgres pg_ctl reload
```

### 8. Redis Issues

#### Symptoms
- Cache misses increase
- "Connection refused" errors
- Memory usage high

#### Diagnosis
```bash
# Check Redis status
docker-compose ps redis

# Test connectivity
docker-compose exec redis redis-cli ping

# View Redis info
docker-compose exec redis redis-cli info

# Monitor memory usage
docker-compose exec redis redis-cli info memory

# Check persistence
docker-compose exec redis redis-cli lastsave
```

#### Solutions

**Memory Issues:**
```bash
# Configure Redis memory limits
docker-compose exec redis redis-cli config set maxmemory 512mb
docker-compose exec redis redis-cli config set maxmemory-policy allkeys-lru
```

**Persistence Issues:**
```bash
# Enable AOF persistence
docker-compose exec redis redis-cli config set appendonly yes

# Manual save
docker-compose exec redis redis-cli bgsave
```

**Cluster Issues:**
```bash
# Check cluster status
docker-compose exec redis redis-cli cluster nodes

# Reshard if needed
docker-compose exec redis redis-cli cluster reshard
```

### 9. Monitoring Issues

#### Symptoms
- Metrics not appearing
- Grafana dashboards empty
- Alerts not firing

#### Diagnosis
```bash
# Check Prometheus targets
curl http://localhost:9090/targets

# Test metrics endpoint
curl http://localhost:9090/metrics | head -20

# Check Grafana logs
docker-compose logs grafana

# Verify alert rules
curl http://localhost:9090/api/v1/rules
```

#### Solutions

**Metrics Not Exposed:**
```bash
# Enable metrics in configuration
echo "PROMETHEUS_ENABLED=true" >> .env
echo "METRICS_ENABLED=true" >> .env
```

**Grafana Configuration:**
```bash
# Reset Grafana admin password
docker-compose exec grafana grafana-cli admin reset-admin-password yournewpassword

# Import dashboards
curl -X POST http://localhost:3000/api/dashboards/import \
  -H "Content-Type: application/json" \
  -d @monitoring/grafana/dashboards/superagent.json
```

### 10. SSL/TLS Issues

#### Symptoms
- HTTPS not working
- Certificate errors
- Mixed content warnings

#### Diagnosis
```bash
# Test SSL connection
openssl s_client -connect localhost:443 -servername localhost

# Check certificate validity
openssl x509 -in /etc/nginx/ssl/superagent.crt -text -noout

# Verify nginx configuration
docker-compose exec nginx nginx -t
```

#### Solutions

**Certificate Expired:**
```bash
# Generate new certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout superagent.key -out superagent.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Update nginx
docker-compose exec nginx nginx -s reload
```

**SSL Configuration Issues:**
```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/nginx/ssl/superagent.crt;
    ssl_certificate_key /etc/nginx/ssl/superagent.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;

    location / {
        proxy_pass http://superagent:8080;
    }
}
```

## Advanced Debugging

### Enable Debug Mode

```bash
# Environment variables for debugging
export LOG_LEVEL=debug
export GIN_MODE=debug
export REQUEST_LOGGING=true
export DB_LOG_QUERIES=true

# Restart with debug mode
docker-compose up -d superagent
```

### Performance Profiling

```go
// Enable Go profiling
import _ "net/http/pprof"

// Access profiling endpoints
curl http://localhost:8080/debug/pprof/profile > cpu.prof
curl http://localhost:8080/debug/pprof/heap > heap.prof

// Analyze profiles
go tool pprof cpu.prof
go tool pprof heap.prof
```

### Database Query Analysis

```sql
-- Find slow queries
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE state = 'active'
ORDER BY duration DESC
LIMIT 10;

-- Check index usage
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public'
ORDER BY n_distinct DESC;

-- Monitor lock waits
SELECT blocked_locks.pid AS blocked_pid,
       blocking_locks.pid AS blocking_pid,
       blocked_activity.usename AS blocked_user,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_query,
       blocking_activity.query AS blocking_query
FROM pg_locks blocked_locks
JOIN pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

### Memory Leak Detection

```bash
# Monitor memory usage over time
docker stats superagent

# Check Go memory statistics
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof -top heap.prof

# Enable garbage collection logging
export GODEBUG=gctrace=1
docker-compose up -d superagent
```

## Emergency Procedures

### Service Restart

```bash
# Graceful restart
docker-compose restart superagent

# Force restart
docker-compose kill superagent
docker-compose up -d superagent

# Rolling restart (Kubernetes)
kubectl rollout restart deployment/superagent
```

### Database Recovery

```bash
# Stop application
docker-compose stop superagent

# Backup current database
docker-compose exec postgres pg_dump -U superagent superagent_db > emergency_backup.sql

# Restore from backup
docker-compose exec -T postgres psql -U superagent superagent_db < backup.sql

# Start application
docker-compose start superagent
```

### Full System Reset

```bash
# WARNING: This will delete all data
docker-compose down -v
docker-compose up -d

# Or reset specific services
docker-compose down postgres redis
docker-compose up -d postgres redis
```

## Support Resources

### Log File Locations

```
/var/log/superagent/application.log
/var/log/postgresql/postgresql.log
/var/log/redis/redis.log
/var/log/nginx/access.log
/var/log/nginx/error.log
```

### Configuration Files

```
/etc/superagent/config.yaml
/etc/nginx/nginx.conf
/etc/postgresql/postgresql.conf
/etc/redis/redis.conf
```

### Useful Commands

```bash
# System information
uname -a
docker --version
docker-compose --version

# Network diagnostics
netstat -tlnp
ss -tlnp
curl -I http://localhost:8080/health

# Disk usage
du -sh /var/lib/*
df -h

# Process information
ps aux | grep superagent
top -p $(pgrep superagent)
```

### Getting Help

1. **Check Documentation**: Review this troubleshooting guide and API docs
2. **Search Issues**: Check GitHub issues for similar problems
3. **Enable Debug Logging**: Set LOG_LEVEL=debug and reproduce the issue
4. **Collect Diagnostics**:
   - Application logs
   - System resource usage
   - Configuration files (without secrets)
   - Error messages and stack traces
5. **Contact Support**: enterprise@superagent.ai with diagnostic information

---

**Quick Reference:**
- Health Check: `curl http://localhost:8080/health`
- View Logs: `docker-compose logs -f superagent`
- Restart Service: `docker-compose restart superagent`
- Check Resources: `docker stats`