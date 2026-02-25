# Docker Installation Guide

This guide covers installing and running HelixAgent using Docker.

## Prerequisites

- Docker 20.10+ or Docker Desktop
- Docker Compose v2+
- 8 GB RAM minimum
- 20 GB disk space

## Quick Start

```bash
# Clone and start
git clone git@github.com:anomaly/helixagent.git
cd helixagent
make docker-run
```

## Docker Compose Files

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Production deployment |
| `docker-compose.test.yml` | Test infrastructure |
| `docker-compose.security.yml` | Security scanning |
| `docker-compose.mcp-full.yml` | MCP server containers |

## Container Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Docker Network                      │
├─────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │ HelixAgent  │  │ PostgreSQL  │  │   Redis     │  │
│  │   :7061     │  │  :15432     │  │   :16379    │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  │
│                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │  ChromaDB   │  │   Mock LLM  │  │ MCP Servers │  │
│  │   :8000     │  │   :18081    │  │  :9101-9999 │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Building Images

```bash
# Build main image
make docker-build

# Build with specific tag
docker build -t helixagent:v1.0.0 -f docker/build/Dockerfile .

# Build for multiple platforms
make build-all
```

## Running Containers

### Development Mode

```bash
# Start all services
make docker-run

# View logs
make docker-logs

# Stop services
make docker-stop
```

### Production Mode

```bash
# Start with production compose
docker compose -f docker-compose.yml up -d

# Scale horizontally
docker compose up -d --scale helixagent=3
```

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
# Server
PORT=7061
GIN_MODE=release

# Database
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=helixagent
POSTGRES_PASSWORD=secure_password
POSTGRES_DB=helixagent_db

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=secure_password

# LLM Providers
CLAUDE_API_KEY=your-key
DEEPSEEK_API_KEY=your-key
GEMINI_API_KEY=your-key
```

### Volume Mounts

```yaml
volumes:
  - ./configs:/app/configs:ro
  - ./data:/app/data
  - ./logs:/app/logs
```

## Health Checks

```bash
# Check container health
docker ps --format "table {{.Names}}\t{{.Status}}"

# Detailed health check
curl http://localhost:7061/v1/health
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker compose logs helixagent

# Check for port conflicts
netstat -tulpn | grep 7061
```

### Database Connection Issues

```bash
# Verify PostgreSQL is healthy
docker compose exec postgres pg_isready

# Check connection from HelixAgent
docker compose exec helixagent nc -zv postgres 5432
```

### Memory Issues

```bash
# Check container resource usage
docker stats

# Increase memory limits in compose
services:
  helixagent:
    deploy:
      resources:
        limits:
          memory: 4G
```

## Cleanup

```bash
# Stop and remove containers
make docker-stop

# Remove volumes
docker compose down -v

# Remove images
docker rmi helixagent:latest
```
