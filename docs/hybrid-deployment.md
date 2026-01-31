# Hybrid Deployment Guide

HelixAgent supports hybrid deployment where the HelixAgent binary runs locally while infrastructure services (PostgreSQL, Redis, ChromaDB) run on a remote host. This document describes the configuration and troubleshooting.

## Architecture

```
Local Machine (helixagent binary) → Remote Host (thinker.local)
    ├── PostgreSQL:5432 (remote, container)
    ├── Redis:6380 (remote, container)
    └── ChromaDB:8001 (remote, container)
```

## Configuration Steps

### 1. Remote Service Setup

On the remote host (`thinker.local`), start the infrastructure containers:

```bash
# Using docker-compose-remote.yml
docker-compose -f docker-compose-remote.yml up -d
```

Verify services are accessible:

```bash
nc -zv thinker.local 5432  # PostgreSQL
nc -zv thinker.local 6380  # Redis
nc -zv thinker.local 8001  # ChromaDB
```

### 2. Local HelixAgent Configuration

Create `.env.remote` with remote service endpoints:

```bash
# Remote services on thinker.local
DB_HOST=thinker.local
REDIS_HOST=thinker.local
CHROMADB_HOST=thinker.local

# BootManager remote flags (REQUIRED for hybrid mode)
SVC_POSTGRESQL_HOST=thinker.local
SVC_POSTGRESQL_REMOTE=true
SVC_REDIS_HOST=thinker.local
SVC_REDIS_REMOTE=true
SVC_CHROMADB_HOST=thinker.local
SVC_CHROMADB_REMOTE=true
```

### 3. Start HelixAgent

Load both `.env` (API keys) and `.env.remote` (remote config):

```bash
export $(cat .env .env.remote | grep -v '^#' | xargs)
./bin/helixagent --strict-dependencies=false --skip-mcp-preinstall
```

### 4. Verify Connectivity

Check health endpoints:

```bash
curl http://localhost:7061/health
curl http://localhost:7061/v1/startup/verification
```

## BootManager Behavior

When `SVC_*_REMOTE=true` is set:

- BootManager skips local container startup for that service
- Performs health checks against remote endpoints only
- Reports: "Service configured as remote or discovered, skipping compose start"

## Troubleshooting

### Remote Services Not Reachable

1. **Check firewall rules** on remote host:
   ```bash
   sudo ufw allow 5432/tcp  # PostgreSQL
   sudo ufw allow 6380/tcp  # Redis
   sudo ufw allow 8001/tcp  # ChromaDB
   ```

2. **Verify container ports are mapped correctly**:
   ```bash
   docker ps | grep -E "(postgres|redis|chromadb)"
   ```

3. **Test connectivity from local machine**:
   ```bash
   timeout 2 bash -c "echo > /dev/tcp/thinker.local/5432"
   ```

### HelixAgent Startup Failures

**Issue**: Provider verification loops due to API key errors (401 Unauthorized).

**Solution**:
- Disable problematic providers via environment variables:
  ```bash
  export MISTRAL_ENABLED=false
  export CEREBRAS_ENABLED=false
  export ZEN_ENABLED=false
  ```
- Use `--strict-dependencies=false` flag
- Skip MCP pre-installation with `--skip-mcp-preinstall`

**Issue**: Health endpoint not responding despite successful BootManager health checks.

**Solution**:
- Wait longer for LLMsVerifier verification to complete (may take 30-60 seconds)
- Check logs: `tail -f helixagent.log`
- Kill stuck processes: `pkill -9 helixagent`

### Auto-Start Configuration

For production deployment, configure remote containers to start automatically:

**Systemd Service** (`/etc/systemd/system/helixagent-stack.service`):

```ini
[Unit]
Description=HelixAgent Infrastructure Stack
Requires=docker.service
After=docker.service network.target

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/path/to/helixagent
ExecStart=/usr/local/bin/docker-compose -f docker-compose-remote.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose-remote.yml down

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable helixagent-stack
sudo systemctl start helixagent-stack
```

## CLI Agent Configuration

HelixAgent can generate configurations for OpenCode and Crush CLI agents:

```bash
# Generate OpenCode config
./bin/helixagent --generate-opencode-config --opencode-output=.opencode.json

# Generate Crush config
./bin/helixagent --generate-crush-config --crush-output=crush_config.json
```

These configurations point to `http://localhost:7061/v1` and expect `HELIXAGENT_API_KEY` environment variable.

## Performance Considerations

- **Network latency**: Remote database calls add ~1-10ms overhead
- **Connection pooling**: Ensure adequate PostgreSQL connection pool settings
- **Redis caching**: Remote Redis may increase cache access latency
- **Health check timeouts**: Increase timeouts for remote services

## Monitoring

Monitor hybrid deployment with:

```bash
# Check remote service health
curl http://thinker.local:5432/health  # PostgreSQL (if health endpoint exists)
curl http://thinker.local:8001/api/v1/heartbeat  # ChromaDB

# Check local HelixAgent health
curl http://localhost:7061/health
```

## Fallback Strategy

If remote services become unavailable, HelixAgent can fall back to local containers:

1. Remove `SVC_*_REMOTE=true` flags
2. Start local containers: `docker-compose up -d postgres redis chromadb`
3. Update `.env.remote` to point to `localhost`

## References

- [BootManager implementation](/internal/services/boot_manager.go)
- [Service configuration](/internal/config/config.go)
- [Remote deployer scripts](/scripts/deploy-remote.sh)
- [Docker Compose templates](/docker-compose-remote.yml)