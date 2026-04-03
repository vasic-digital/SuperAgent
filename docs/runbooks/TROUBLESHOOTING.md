# Troubleshooting Runbook

## Service Won't Start

### Check ports
```bash
netstat -tlnp | grep -E '8000|8001|8283|6333|7474|6379'
```

### Check logs
```bash
podman logs helixmemory-cognee
podman logs helixmemory-mem0
podman logs helixmemory-letta
```

## Memory Operations Failing

### Check service health
```bash
curl http://localhost:8000/api/v1/health
curl http://localhost:8001/health
curl http://localhost:8283/v1/health
```

### Restart services
```bash
podman-compose -f docker-compose.memory.yml restart
```

## Database Connection Issues

### Check PostgreSQL
```bash
podman exec helixmemory-postgres pg_isready -U helixmemory
```

### Check Redis
```bash
podman exec helixmemory-redis redis-cli ping
```

## Circuit Breaker Tripped

Wait 30s for auto-recovery or restart:
```bash
podman-compose -f docker-compose.memory.yml restart cognee mem0 letta
```
