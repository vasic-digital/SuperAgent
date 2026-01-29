# Services Configuration Guide

HelixAgent manages all infrastructure services (databases, caches, monitoring, vector stores) through a unified configuration system. Each service can run locally via Docker Compose or connect to remote instances.

## Overview

All services are configured under the `services:` section in your YAML config file (`configs/development.yaml` or `configs/production.yaml`).

### Supported Services

| Service | Default Port | Health Check | Required by Default |
|---------|-------------|-------------|-------------------|
| PostgreSQL | 5432 | TCP | Yes |
| Redis | 6379 | TCP | Yes |
| Cognee | 8000 | HTTP `/health` | Yes |
| ChromaDB | 8100 | HTTP `/api/v1/heartbeat` | Yes |
| Prometheus | 9090 | HTTP `/-/healthy` | No |
| Grafana | 3000 | HTTP `/api/health` | No |
| Neo4j | 7474 | HTTP `/` | No |
| Kafka | 9092 | TCP | No |
| RabbitMQ | 5672 | TCP | No |
| Qdrant | 6333 | HTTP `/health` | No |
| Weaviate | 8080 | HTTP `/v1/.well-known/ready` | No |
| LangChain | 8200 | HTTP `/health` | No |
| LlamaIndex | 8300 | HTTP `/health` | No |

## YAML Configuration

Each service supports these fields:

```yaml
services:
  postgresql:
    host: "localhost"           # Service hostname
    port: "5432"                # Service port
    url: ""                     # Full URL override (takes precedence over host:port)
    enabled: true               # Whether to manage this service
    required: true              # If true, boot fails when service is unavailable
    remote: false               # If true, skip Docker Compose start (health check only)
    health_type: "tcp"          # Health check type: "tcp", "http"
    health_path: ""             # HTTP health check endpoint path
    timeout: "30s"              # Health check timeout
    retry_count: 3              # Number of health check retries
    compose_file: "docker-compose.yml"  # Docker Compose file
    service_name: "postgres"    # Service name in compose file
```

### Example: Development Config

```yaml
services:
  postgresql:
    host: "localhost"
    port: "5432"
    enabled: true
    required: true
    health_type: "tcp"
    timeout: "30s"
    retry_count: 3
    compose_file: "docker-compose.yml"
    service_name: "postgres"

  redis:
    host: "localhost"
    port: "6379"
    enabled: true
    required: true
    health_type: "tcp"
    timeout: "30s"
    retry_count: 3
    compose_file: "docker-compose.yml"
    service_name: "redis"

  cognee:
    host: "localhost"
    port: "8000"
    enabled: true
    required: true
    health_type: "http"
    health_path: "/health"
    timeout: "60s"
    retry_count: 5
    compose_file: "docker-compose.yml"
    service_name: "cognee"

  prometheus:
    host: "localhost"
    port: "9090"
    enabled: false
    required: false
    health_type: "http"
    health_path: "/-/healthy"
```

## Remote Services

When deploying with external/managed services, set `remote: true` to skip Docker Compose startup while still performing health checks:

```yaml
services:
  postgresql:
    host: "db.production.example.com"
    port: "5432"
    enabled: true
    required: true
    remote: true          # Don't start via Docker Compose
    health_type: "tcp"
    timeout: "30s"
    retry_count: 5

  redis:
    host: "redis.production.example.com"
    port: "6379"
    enabled: true
    required: true
    remote: true
    health_type: "tcp"
```

See `configs/remote-services-example.yaml` for a complete remote configuration example.

## Environment Variable Overrides

Every service field can be overridden via environment variables with the prefix `SVC_<SERVICE>_<FIELD>`:

| Variable | Description |
|----------|-------------|
| `SVC_POSTGRESQL_HOST` | PostgreSQL hostname |
| `SVC_POSTGRESQL_PORT` | PostgreSQL port |
| `SVC_POSTGRESQL_REMOTE` | Set to `true` for remote mode |
| `SVC_POSTGRESQL_ENABLED` | Set to `false` to disable |
| `SVC_POSTGRESQL_URL` | Full URL override |
| `SVC_REDIS_HOST` | Redis hostname |
| `SVC_REDIS_PORT` | Redis port |
| `SVC_REDIS_REMOTE` | Redis remote mode |
| `SVC_COGNEE_HOST` | Cognee hostname |
| `SVC_CHROMADB_HOST` | ChromaDB hostname |
| `SVC_PROMETHEUS_ENABLED` | Enable/disable Prometheus |
| `SVC_GRAFANA_ENABLED` | Enable/disable Grafana |
| `SVC_QDRANT_HOST` | Qdrant hostname |
| `SVC_WEAVIATE_HOST` | Weaviate hostname |
| `SVC_LANGCHAIN_HOST` | LangChain hostname |
| `SVC_LLAMAINDEX_HOST` | LlamaIndex hostname |

Example:
```bash
SVC_POSTGRESQL_HOST=db.example.com SVC_POSTGRESQL_REMOTE=true ./bin/helixagent
```

## Boot Behavior

### Startup Sequence

1. Load `ServicesConfig` from YAML + environment variables
2. `BootManager.BootAll()` iterates all enabled services
3. For local services: start via `docker compose -f <file> up -d <service>`
4. For remote services (`remote: true`): skip compose start
5. Health check all enabled services with retries
6. Required services that fail health check cause boot abort
7. Optional services that fail are logged as warnings

### Shutdown Sequence

1. HTTP server receives shutdown signal (SIGTERM/SIGINT)
2. HTTP server graceful shutdown (drain connections)
3. `BootManager.ShutdownAll()` stops all locally-started compose services
4. Process exits

## Troubleshooting

### Service Won't Start
- Check Docker/Podman is running: `docker ps` or `podman ps`
- Check compose file exists at the configured `compose_file` path
- Check service name matches the compose file definition

### Health Check Fails
- Increase `timeout` and `retry_count` for slow-starting services
- Verify the service is actually running: `docker compose ps`
- For HTTP health checks, verify the `health_path` is correct

### Remote Service Unreachable
- Verify network connectivity: `curl http://<host>:<port><health_path>`
- Check firewall rules allow the connection
- Ensure `remote: true` is set (otherwise HelixAgent tries to start it locally)

## Key Files

| File | Description |
|------|-------------|
| `internal/config/config.go` | `ServiceEndpoint`, `ServicesConfig` types and helpers |
| `internal/services/boot_manager.go` | Service startup/shutdown orchestration |
| `internal/services/health_checker.go` | TCP and HTTP health checking with retries |
| `configs/development.yaml` | Development configuration |
| `configs/production.yaml` | Production configuration |
| `configs/remote-services-example.yaml` | Remote services example |
