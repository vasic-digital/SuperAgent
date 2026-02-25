# HelixAgent Installation Guide

This directory contains comprehensive installation guides for deploying HelixAgent in various environments.

## Quick Start

For the fastest way to get started:

```bash
# Clone the repository
git clone git@github.com:anomaly/helixagent.git
cd helixagent

# Start with Docker
make docker-run

# Or run locally
make build && make run
```

## Installation Methods

| Method | Guide | Use Case |
|--------|-------|----------|
| Docker | [docker.md](./docker.md) | Recommended for most users |
| Podman | [podman.md](./podman.md) | Alternative to Docker |
| Kubernetes | [kubernetes.md](./kubernetes.md) | Production deployments |
| Bare Metal | [bare-metal.md](./bare-metal.md) | Direct installation |

## Prerequisites

### Required
- Go 1.24+ (for building from source)
- Docker or Podman (for containerized deployment)
- Git

### Optional
- PostgreSQL 15+ (for persistence)
- Redis 7+ (for caching)
- Make (for build automation)

## System Requirements

### Minimum
- CPU: 2 cores
- RAM: 4 GB
- Disk: 10 GB

### Recommended
- CPU: 4+ cores
- RAM: 8+ GB
- Disk: 50+ GB (with vector storage)

## Environment Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `7061` |
| `GIN_MODE` | Gin mode (debug/release) | `release` |
| `JWT_SECRET` | JWT signing secret | (required) |
| `DB_*` | PostgreSQL connection | localhost |
| `REDIS_*` | Redis connection | localhost |

## Next Steps

1. Configure your LLM providers (see [Provider Configuration](../user-guides/PROTOCOLS_COMPREHENSIVE_GUIDE.md))
2. Set up MCP servers (see [MCP Integration](../mcp/))
3. Configure monitoring (see [Deployment Guide](../deployment/DEPLOYMENT_GUIDE.md))

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Find and kill process using port
   lsof -i :7061
   kill -9 <PID>
   ```

2. **Database connection failed**
   ```bash
   # Verify PostgreSQL is running
   docker ps | grep postgres
   
   # Check connection
   psql -h localhost -p 15432 -U helixagent -d helixagent_db
   ```

3. **Redis connection refused**
   ```bash
   # Verify Redis is running
   docker ps | grep redis
   
   # Test connection
   redis-cli -h localhost -p 16379 -a helixagent123 ping
   ```

For more help, see [Troubleshooting Guide](../guides/troubleshooting-guide.md).
