# HelixAgent Service Management

Complete guide for managing all HelixAgent services, containers, and infrastructure.

## Table of Contents

- [Quick Start](#quick-start)
- [Service Scripts](#service-scripts)
- [Service Architecture](#service-architecture)
- [Container Management](#container-management)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)

---

## Quick Start

### Start All Services

```bash
# Start all core and optional services
./scripts/start-all-services.sh

# Start with Big Data services
START_BIGDATA=true ./scripts/start-all-services.sh

# Start with Security services
START_SECURITY=true ./scripts/start-all-services.sh

# Start everything
START_BIGDATA=true START_SECURITY=true ./scripts/start-all-services.sh
```

### Verify Services

```bash
# Run comprehensive service verification
./scripts/verify-services.sh
```

### Stop All Services

```bash
# Gracefully stop all services
./scripts/stop-all-services.sh
```

### Start HelixAgent Server

```bash
# After services are running
make run

# Or in development mode
make run-dev
```

---

## Service Scripts

### start-all-services.sh

Comprehensive startup script that brings up all HelixAgent services in the correct order.

**Features:**
- Automatic runtime detection (Docker/Podman)
- Health check verification
- Phased startup (Core → Messaging → Monitoring → Protocols → Integration)
- Optional Big Data and Security services
- Detailed status output

**Usage:**
```bash
./scripts/start-all-services.sh

# With optional services
START_BIGDATA=true START_SECURITY=true ./scripts/start-all-services.sh
```

**Startup Phases:**
1. **Phase 1: Core Infrastructure** - PostgreSQL, Redis, Cognee, ChromaDB
2. **Phase 2: Messaging & Queuing** - Kafka, RabbitMQ, Zookeeper
3. **Phase 3: Monitoring & Observability** - Prometheus, Grafana, Loki
4. **Phase 4: Protocol Servers** - MCP, LSP, ACP
5. **Phase 5: Integration Services** - Weaviate, Neo4j, LangChain, LlamaIndex
6. **Phase 6: Analytics & Big Data** (optional) - Flink, MinIO, Qdrant, Iceberg, Spark
7. **Phase 7: Security Services** (optional) - Vault, OWASP ZAP

### stop-all-services.sh

Gracefully shuts down all HelixAgent services in reverse order.

**Features:**
- Reverse order shutdown (prevents dependency issues)
- Cleanup of orphaned containers
- Status verification

**Usage:**
```bash
./scripts/stop-all-services.sh
```

### verify-services.sh

Comprehensive service health verification.

**Features:**
- Container status checks
- Connectivity tests
- Health endpoint verification
- Detailed failure reporting

**Usage:**
```bash
./scripts/verify-services.sh
```

**Checks Performed:**
- Container existence and status
- PostgreSQL connectivity (`pg_isready`)
- Redis connectivity (`PING`)
- Cognee health endpoint
- Prometheus health endpoint
- Grafana health endpoint
- HelixAgent API endpoints

---

## Service Architecture

### Core Services (Always Required)

| Service | Container | Port | Purpose |
|---------|-----------|------|---------|
| **PostgreSQL** | helixagent-postgres | 5432 | Primary database |
| **Redis** | helixagent-redis | 6379 | Cache and session store |
| **Cognee** | helixagent-cognee | 8000 | Knowledge graph and RAG |
| **ChromaDB** | helixagent-chromadb | 8000 | Vector database |

### Messaging Services

| Service | Container | Port | Purpose |
|---------|-----------|------|---------|
| **Kafka** | helixagent-kafka | 9092 | Event streaming |
| **RabbitMQ** | helixagent-rabbitmq | 5672, 15672 | Message queue |
| **Zookeeper** | helixagent-zookeeper | 2181 | Kafka coordination |

### Monitoring Services

| Service | Container | Port | Purpose |
|---------|-----------|------|---------|
| **Prometheus** | helixagent-prometheus | 9090 | Metrics collection |
| **Grafana** | helixagent-grafana | 3000 | Metrics visualization |
| **Loki** | helixagent-loki | 3100 | Log aggregation |
| **Alertmanager** | helixagent-alertmanager | 9093 | Alert management |

### Integration Services

| Service | Container | Port | Purpose |
|---------|-----------|------|---------|
| **Weaviate** | helixagent-weaviate | 8080 | Vector search |
| **Neo4j** | helixagent-neo4j | 7474, 7687 | Graph database |
| **LangChain** | langchain-server | 8001 | LangChain integration |
| **LlamaIndex** | llamaindex-server | 8002 | LlamaIndex integration |
| **Guidance** | guidance-server | 8003 | Guidance integration |
| **LMQL** | lmql-server | 8004 | LMQL integration |

### Big Data Services (Optional)

| Service | Container | Port | Purpose |
|---------|-----------|------|---------|
| **Flink JobManager** | flink-jobmanager | 8081 | Stream processing |
| **Flink TaskManager** | flink-taskmanager | - | Flink workers |
| **MinIO** | helixagent-minio | 9000, 9001 | Object storage |
| **Qdrant** | helixagent-qdrant | 6333 | Vector search |
| **Iceberg REST** | iceberg-rest | 8181 | Table format |
| **Spark Master** | spark-master | 7077, 8080 | Batch processing |

### Security Services (Optional)

| Service | Container | Port | Purpose |
|---------|-----------|------|---------|
| **Vault** | helixagent-vault | 8200 | Secrets management |
| **OWASP ZAP** | helixagent-zap | 8090 | Security testing |

---

## Container Management

### Check Running Containers

```bash
# Using Podman
podman ps --filter name=helixagent

# Using Docker
docker ps --filter name=helixagent
```

### View Container Logs

```bash
# All services
podman-compose logs -f

# Specific service
podman-compose logs -f helixagent-postgres
podman-compose logs -f helixagent-redis

# Last 100 lines
podman logs -n 100 helixagent-postgres
```

### Restart a Specific Service

```bash
# Using compose
podman-compose restart helixagent-postgres

# Direct container restart
podman restart helixagent-postgres
```

### Check Service Health

```bash
# PostgreSQL
podman exec helixagent-postgres pg_isready

# Redis
podman exec helixagent-redis redis-cli ping

# Cognee
curl http://localhost:8000/health

# HelixAgent
curl http://localhost:7061/health
curl http://localhost:7061/v1/startup/verification
```

### Access Service Shells

```bash
# PostgreSQL
podman exec -it helixagent-postgres psql -U helixagent -d helixagent_db

# Redis
podman exec -it helixagent-redis redis-cli

# Kafka
podman exec -it helixagent-kafka kafka-topics --list --bootstrap-server localhost:9092
```

---

## Troubleshooting

### Services Won't Start

**Check container runtime:**
```bash
# Verify Podman/Docker is running
podman --version
systemctl status podman.socket  # Podman
systemctl status docker          # Docker
```

**Check port conflicts:**
```bash
# Check if ports are already in use
ss -tulpn | grep -E '5432|6379|8000|9090|3000'
```

**Check logs:**
```bash
podman-compose logs helixagent-postgres
podman-compose logs helixagent-redis
```

### Service Health Check Failures

**PostgreSQL not ready:**
```bash
# Check if PostgreSQL is accepting connections
podman exec helixagent-postgres pg_isready -U helixagent

# Check logs
podman logs helixagent-postgres
```

**Redis not responding:**
```bash
# Test Redis connection
podman exec helixagent-redis redis-cli ping

# Check logs
podman logs helixagent-redis
```

**Cognee not healthy:**
```bash
# Check health endpoint
curl -v http://localhost:8000/health

# Check logs
podman logs helixagent-cognee
```

### Container Crashes

**View crash logs:**
```bash
podman logs --tail 100 helixagent-postgres
```

**Check container resources:**
```bash
podman stats helixagent-postgres
```

**Restart crashed container:**
```bash
podman restart helixagent-postgres
```

### Cleanup Orphaned Resources

**Remove stopped containers:**
```bash
podman ps -a --filter name=helixagent --filter status=exited -q | xargs podman rm
```

**Remove unused volumes:**
```bash
podman volume prune
```

**Remove unused networks:**
```bash
podman network prune
```

### Complete Reset

**WARNING: This will delete all data!**

```bash
# Stop everything
./scripts/stop-all-services.sh

# Remove all containers
podman ps -a --filter name=helixagent -q | xargs podman rm -f

# Remove all volumes (DATA LOSS!)
podman volume ls --filter name=helixagent -q | xargs podman volume rm

# Remove networks
podman network ls --filter name=helixagent -q | xargs podman network rm

# Start fresh
./scripts/start-all-services.sh
```

---

## Advanced Configuration

### Environment Variables

Set these before starting services:

```bash
# Big Data services
export START_BIGDATA=true

# Security services
export START_SECURITY=true

# Custom PostgreSQL configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=helixagent
export DB_PASSWORD=helixagent123
export DB_NAME=helixagent_db

# Custom Redis configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=helixagent123

# Cognee configuration
export COGNEE_AUTH_EMAIL=admin@helixagent.ai
export COGNEE_AUTH_PASSWORD=HelixAgentPass123
```

### Persistent Data

All data is stored in named volumes:

```bash
# List volumes
podman volume ls --filter name=helixagent

# Inspect volume
podman volume inspect helixagent-postgres-data

# Backup volume
podman run --rm -v helixagent-postgres-data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz /data
```

### Custom Compose Files

Create custom compose overrides:

```yaml
# docker-compose.override.yml
version: '3.8'

services:
  postgres:
    environment:
      - POSTGRES_MAX_CONNECTIONS=200
    resources:
      limits:
        memory: 2G
```

Then start with:
```bash
podman-compose -f docker-compose.yml -f docker-compose.override.yml up -d
```

### Production Deployment

For production, use:

```bash
podman-compose -f docker-compose.production.yml up -d
```

This includes:
- Resource limits
- Health checks
- Restart policies
- Security hardening
- TLS configuration

---

## Service Dependencies

### Dependency Graph

```
HelixAgent Server
    ├── PostgreSQL (required)
    ├── Redis (required)
    ├── Cognee (required)
    ├── ChromaDB (required)
    ├── Kafka (optional, for events)
    │   └── Zookeeper (required by Kafka)
    ├── Prometheus (optional, for metrics)
    └── Integration Services (optional)
        ├── Weaviate
        ├── Neo4j
        ├── LangChain
        └── LlamaIndex
```

### Startup Order

1. **Core Database Layer:** PostgreSQL, Redis
2. **Application Services:** Cognee, ChromaDB
3. **Messaging Layer:** Zookeeper → Kafka → RabbitMQ
4. **Monitoring Layer:** Prometheus, Grafana, Loki
5. **Protocol Layer:** MCP, LSP, ACP servers
6. **Integration Layer:** Weaviate, Neo4j, LangChain, etc.
7. **HelixAgent Server:** Main application

---

## Quick Reference

### Essential Commands

```bash
# Start everything
./scripts/start-all-services.sh

# Verify all services
./scripts/verify-services.sh

# Start HelixAgent
make run

# View logs
podman-compose logs -f

# Stop everything
./scripts/stop-all-services.sh

# Status check
podman ps --filter name=helixagent
```

### Health Checks

```bash
# PostgreSQL
podman exec helixagent-postgres pg_isready

# Redis
podman exec helixagent-redis redis-cli ping

# Cognee
curl http://localhost:8000/health

# HelixAgent
curl http://localhost:7061/health
curl http://localhost:7061/v1/startup/verification

# Prometheus
curl http://localhost:9090/-/healthy

# Grafana
curl http://localhost:3000/api/health
```

---

## See Also

- [Deployment Guide](./deployment/DEPLOYMENT_GUIDE.md)
- [Service Boot Configuration](./architecture/SERVICE_ARCHITECTURE.md)
- [Docker Compose Reference](../docker-compose.yml)
- [Troubleshooting Guide](./troubleshooting/TROUBLESHOOTING.md)
