# HelixMemory Setup Guide

## Overview

HelixMemory is a unified AI memory system that fuses three powerful memory backends:

- **Cognee**: Knowledge graph memory with semantic understanding
- **Mem0**: User-specific semantic memory  
- **Letta**: Agent memory with persistent state

## Deployment Modes

### Mode 1: Local Containers (Recommended, Free)

Run all three memory services locally using Docker containers.

**Requirements:**
- 4+ CPU cores (8+ recommended)
- 8GB+ RAM (16GB recommended)
- 20GB+ disk space (50GB recommended)
- Docker & Docker Compose

**Steps:**

1. **Check Hardware:**
   ```bash
   ./scripts/check_memory_hardware.sh
   ```

2. **Configure Environment:**
   ```bash
   cp .env.example .env
   # Edit .env - set HELIX_MEMORY_MODE=local
   # No API keys needed for Cognee in local mode
   ```

3. **Start Services:**
   ```bash
   docker-compose -f docker-compose.memory.yml up -d
   ```

4. **Verify:**
   ```bash
   docker-compose -f docker-compose.memory.yml ps
   ```

**Services:**
| Service | Port | Description |
|---------|------|-------------|
| Cognee | 8000 | Knowledge graph memory |
| Mem0 | 8001 | Semantic memory |
| Letta | 8283 | Agent memory |
| Qdrant | 6333 | Vector database |
| Neo4j | 7474/7687 | Graph database |
| Redis | 6379 | Cache |
| PostgreSQL | 5433 | SQL database |

### Mode 2: Cloud APIs (Mixed)

Use cloud APIs for some services while running others locally.

**When to use:**
- Limited local hardware
- Want to try specific services without full deployment
- Already have cloud API keys

**Configuration:**

```bash
# .env
HELIX_MEMORY_MODE=cloud

# Mem0 and Letta have free tiers
HELIX_MEMORY_MEM0_API_KEY=your-mem0-key
HELIX_MEMORY_LETTA_API_KEY=your-letta-key

# Cognee requires PAID subscription
# If empty, Cognee will use local container automatically
HELIX_MEMORY_COGNEE_API_KEY=      # Leave empty for local
```

### Mode 3: Full Cloud (All Paid)

Use cloud APIs for all services.

**⚠️ WARNING:** Cognee requires a **paid subscription**. This mode will fail without valid API keys for all services.

```bash
# .env
HELIX_MEMORY_MODE=cloud

HELIX_MEMORY_COGNEE_API_KEY=your-paid-cognee-key
HELIX_MEMORY_MEM0_API_KEY=your-mem0-key
HELIX_MEMORY_LETTA_API_KEY=your-letta-key
```

## API Key Setup

### Mem0 (Free tier available)

1. Sign up at https://app.mem0.ai
2. Generate API key from dashboard
3. Add to `.env`:
   ```
   HELIX_MEMORY_MEM0_API_KEY=your-key-here
   ```

### Letta (Free tier available)

1. Visit https://docs.letta.com/guides/build-with-letta/quickstart
2. Follow API key generation instructions
3. Add to `.env`:
   ```
   HELIX_MEMORY_LETTA_API_KEY=your-key-here
   ```

### Cognee (⚠️ PAID SUBSCRIPTION REQUIRED)

1. Visit https://platform.cognee.ai
2. **Subscribe to a paid plan**
3. Generate API key from dashboard
4. Add to `.env`:
   ```
   HELIX_MEMORY_COGNEE_API_KEY=your-paid-key-here
   ```

**Important:** If you don't have a Cognee API key, leave it empty and HelixMemory will automatically use the local Cognee container:
```bash
HELIX_MEMORY_COGNEE_API_KEY=      # Empty = use local container
```

## Hybrid Mode

HelixMemory intelligently handles mixed configurations:

```bash
# Example: Mem0 and Letta in cloud, Cognee local (no paid key)
HELIX_MEMORY_MODE=cloud
HELIX_MEMORY_COGNEE_API_KEY=      # Empty - uses local
HELIX_MEMORY_MEM0_API_KEY=mem0-key
HELIX_MEMORY_LETTA_API_KEY=letta-key
```

The system will:
1. Use cloud Mem0 and Letta (with provided keys)
2. Automatically use local Cognee container (no key provided)

## Endpoint Override

You can override endpoints for local instances using API keys:

```bash
# Use local endpoints but with authentication
HELIX_MEMORY_MODE=local
HELIX_MEMORY_COGNEE_ENDPOINT=http://my-cognee:8000
HELIX_MEMORY_COGNEE_API_KEY=local-api-key
```

This is useful when:
- Running services on remote servers
- Using custom deployments
- Adding authentication to local instances

## Verification

### Check Service Health

```bash
# All HelixMemory services
docker-compose -f docker-compose.memory.yml ps

# Individual service health
curl http://localhost:8000/api/v1/health   # Cognee
curl http://localhost:8001/health          # Mem0
curl http://localhost:8283/v1/health       # Letta
```

### Test Memory Operations

```bash
# Store a memory
curl -X POST http://localhost:8001/v1/memories/ \
  -H "Authorization: Token your-mem0-key" \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"Test memory"}]}'

# Search memories
curl -X POST http://localhost:8001/v1/memories/search/ \
  -H "Authorization: Token your-mem0-key" \
  -H "Content-Type: application/json" \
  -d '{"query":"test"}'
```

## Troubleshooting

### Cognee won't start (port 8000)

```bash
# Check if port is in use
sudo lsof -i :8000

# Kill process or change port in docker-compose.memory.yml
```

### Out of memory

```bash
# Check memory usage
docker stats

# Reduce memory limits in docker-compose.memory.yml
# Or switch to cloud mode for some services
```

### API key not working

1. Verify key is correct (no extra spaces)
2. Check endpoint URL matches key type (cloud vs local)
3. For cloud: verify subscription is active
4. Check network connectivity

### Services not connecting

```bash
# Check logs
docker-compose -f docker-compose.memory.yml logs -f [service]

# Restart specific service
docker-compose -f docker-compose.memory.yml restart cognee
```

## Security

### API Keys

- **NEVER** commit `.env` files to git
- `.env` is already in `.gitignore`
- Rotate keys regularly
- Use different keys for development/production

### Local Security

Default local credentials (change in production):
- Neo4j: `neo4j/helixmemory`
- PostgreSQL: `helixmemory/helixmemory`
- Redis: no password (local only)

Change these in `.env` before production deployment.

## Resource Usage

### Local Mode (All Containers)

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 4 cores | 8 cores |
| RAM | 8 GB | 16 GB |
| Disk | 20 GB | 50 GB |

### Per-Service Requirements

| Service | Memory Limit | CPU Limit |
|---------|--------------|-----------|
| Cognee | 4 GB | 2 cores |
| Mem0 | 2 GB | 1 core |
| Letta | 2 GB | 1 core |
| Neo4j | 4 GB | 2 cores |
| Qdrant | 2 GB | 1 core |
| PostgreSQL | 2 GB | 1 core |
| Redis | 512 MB | 0.5 cores |

## Migration

### From Mem0-only (Legacy)

HelixAgent previously used only Mem0. To migrate:

1. Keep existing Mem0 data
2. Add Cognee and Letta containers
3. HelixMemory will automatically fuse all three

No data migration needed - HelixMemory reads from all sources.

### To Cloud

1. Export local data (if needed)
2. Obtain cloud API keys
3. Update `.env` to `HELIX_MEMORY_MODE=cloud`
4. Restart HelixAgent

## Support

- **Cognee**: https://docs.cognee.ai
- **Mem0**: https://docs.mem0.ai
- **Letta**: https://docs.letta.com
- **HelixMemory Issues**: Create issue in HelixAgent repository
