# Production Deployment Guide

## Quick Start

```bash
# 1. Environment
cp .env.example .env
# Edit .env with production values

# 2. Infrastructure
podman-compose -f docker-compose.memory.yml up -d

# 3. HelixAgent
make build
./bin/helixagent
```

## Requirements

- 16GB RAM, 8 CPU cores
- Docker/Podman
- SSL certificates

## Verification

```bash
curl http://localhost:7061/v1/health
curl http://localhost:8000/api/v1/health  # Cognee
curl http://localhost:8001/health         # Mem0
curl http://localhost:8283/v1/health      # Letta
```
