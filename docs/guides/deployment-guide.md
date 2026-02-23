# HelixAgent Deployment Guide

**Version:** 1.2.0  
**Last Updated:** 2026-02-23

## Prerequisites

- Go 1.24+
- Docker or Podman
- PostgreSQL 15+
- Redis 7+

## Quick Start

### 1. Install Dependencies

```bash
make install-deps
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your settings
```

### 3. Start Infrastructure

```bash
make infra-start
```

### 4. Build and Run

```bash
make build
make run
```

## Container Deployment

### Docker Compose

```bash
make docker-build
docker-compose up -d
```

### Podman

```bash
make container-build
podman-compose up -d
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `7061` |
| `GIN_MODE` | Gin mode | `release` |
| `JWT_SECRET` | JWT signing key | Required |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |

### LLM Provider API Keys

| Provider | Environment Variable |
|----------|---------------------|
| Claude | `CLAUDE_API_KEY` |
| DeepSeek | `DEEPSEEK_API_KEY` |
| Gemini | `GEMINI_API_KEY` |
| Mistral | `MISTRAL_API_KEY` |
| OpenRouter | `OPENROUTER_API_KEY` |
| ZAI | `ZAI_API_KEY` |
| Cerebras | `CEREBRAS_API_KEY` |

## Health Checks

```bash
curl http://localhost:7061/v1/health
curl http://localhost:7061/v1/monitoring/status
```

## Monitoring

```bash
make monitoring-status
make monitoring-circuit-breakers
make monitoring-provider-health
```

## Scaling

### Horizontal Scaling

1. Use external PostgreSQL and Redis
2. Configure `SVC_*_REMOTE=true` in `.env`
3. Deploy multiple instances behind load balancer

### Resource Limits

All containers support resource limits via Docker Compose:

```yaml
services:
  helixagent:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

## Troubleshooting

### Common Issues

1. **Port already in use**: Change `PORT` in `.env`
2. **Database connection failed**: Check `DB_*` settings
3. **Redis connection failed**: Check `REDIS_*` settings
4. **Provider unavailable**: Verify API keys

### Logs

```bash
make container-logs
```

## See Also

- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guide
- [API_REFERENCE.md](api/API_REFERENCE.md) - API documentation
