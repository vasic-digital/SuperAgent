# Service Architecture

HelixAgent uses a unified service management architecture to orchestrate all infrastructure dependencies (databases, caches, monitoring, vector stores) through a single boot manager.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     cmd/helixagent/main.go                   │
│                                                              │
│  1. Load Config (YAML + env vars)                           │
│  2. bootMgr := services.NewBootManager(&cfg.Services)       │
│  3. bootMgr.BootAll()                                       │
│  4. ... run server ...                                      │
│  5. bootMgr.ShutdownAll()                                   │
└──────────┬──────────────────────────────────────────────────┘
           │
┌──────────▼──────────────────────────────────────────────────┐
│                      BootManager                             │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐  │
│  │ ServicesConfig  │  │ HealthChecker  │  │   Results     │  │
│  │ (13 services)  │  │ (TCP + HTTP)   │  │ map[string]*  │  │
│  └────────┬───────┘  └───────┬────────┘  └──────────────┘  │
│           │                  │                               │
│  ┌────────▼──────────────────▼──────────────────────────┐   │
│  │              Per-Service Boot Logic                    │   │
│  │                                                       │   │
│  │  Disabled? → skip                                     │   │
│  │  Remote?   → health check only                        │   │
│  │  Local?    → docker compose up → health check         │   │
│  │  Required? → fail boot on health check failure        │   │
│  │  Optional? → warn and continue                        │   │
│  └───────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
```

## Service Dependency Graph

```
                    ┌──────────────┐
                    │  HelixAgent  │
                    └──────┬───────┘
           ┌───────────────┼───────────────┐
           │               │               │
    ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐
    │ PostgreSQL  │ │    Redis    │ │   Cognee    │
    │ (required)  │ │ (required)  │ │ (required)  │
    └─────────────┘ └─────────────┘ └──────┬──────┘
                                           │
                                    ┌──────▼──────┐
                                    │  ChromaDB   │
                                    │ (required)  │
                                    └─────────────┘

    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │ Prometheus  │ │   Grafana   │ │   Qdrant    │
    │ (optional)  │ │ (optional)  │ │ (optional)  │
    └─────────────┘ └─────────────┘ └─────────────┘

    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │  Weaviate   │ │  LangChain  │ │ LlamaIndex  │
    │ (optional)  │ │ (optional)  │ │ (optional)  │
    └─────────────┘ └─────────────┘ └─────────────┘
```

## Component Details

### ServicesConfig (`internal/config/config.go`)

Defines all managed services with their connection details, health check configuration, and boot behavior:

- **ServiceEndpoint**: Individual service definition (host, port, health check type, timeouts, compose file reference)
- **ServicesConfig**: Collection of all 13+ services
- **DefaultServicesConfig()**: Sensible defaults for development
- **LoadServicesFromEnv()**: Environment variable overrides (`SVC_<SERVICE>_<FIELD>`)
- **AllEndpoints()**: Returns all services as `map[string]ServiceEndpoint`
- **RequiredEndpoints()**: Returns only required, enabled services

### BootManager (`internal/services/boot_manager.go`)

Orchestrates the full service lifecycle:

- **BootAll()**: Groups local services by compose file, starts them via `docker compose up -d`, then health checks all enabled services
- **HealthCheckAll()**: Runs health checks on all enabled services, returns results map
- **ShutdownAll()**: Stops all compose services that were started during boot
- **detectComposeCmd()**: Auto-detects Docker Compose V2, docker-compose V1, Podman Compose, or Podman Compose plugin

### HealthChecker (`internal/services/health_checker.go`)

Performs service health checks:

- **Check()**: Dispatches to TCP or HTTP check based on `HealthType`
- **CheckWithRetry()**: Retries failed checks with configurable count and 2-second delays
- TCP check: `net.DialTimeout` to `host:port`
- HTTP check: GET request to `url + health_path`, expects status < 500

## Boot Sequence Diagram

```
main.go          BootManager       HealthChecker     Docker Compose
   │                  │                  │                  │
   │─BootAll()───────►│                  │                  │
   │                  │                  │                  │
   │                  │ [for each enabled local service]    │
   │                  │──compose up ─────────────────────►  │
   │                  │◄─────────────────────────── ok ──── │
   │                  │                  │                  │
   │                  │ [for each enabled service]          │
   │                  │──CheckWithRetry()│                  │
   │                  │─────────────────►│                  │
   │                  │                  │──TCP/HTTP check──►
   │                  │◄─────── result ──│                  │
   │                  │                  │                  │
   │                  │ [if required && failed]             │
   │◄── error ────────│                  │                  │
   │                  │                  │                  │
   │                  │ [if optional && failed]             │
   │                  │  log warning     │                  │
   │◄── nil ──────────│                  │                  │
```

## Shutdown Sequence

```
Signal (SIGTERM)    main.go          BootManager       Docker Compose
      │                │                  │                  │
      │───────────────►│                  │                  │
      │                │─server.Shutdown()│                  │
      │                │                  │                  │
      │                │─ShutdownAll()───►│                  │
      │                │                  │                  │
      │                │                  │ [for each compose group]
      │                │                  │──compose down──► │
      │                │                  │◄────── ok ────── │
      │                │◄──── nil ────────│                  │
      │                │                  │                  │
      │                │─ os.Exit(0) ─────│                  │
```

## Remote Deployment Patterns

### All Services Remote (Cloud)

All services are managed externally (AWS RDS, ElastiCache, etc.):

```yaml
services:
  postgresql:
    host: "rds-instance.amazonaws.com"
    port: "5432"
    remote: true
    required: true
  redis:
    host: "elasticache-cluster.amazonaws.com"
    port: "6379"
    remote: true
    required: true
```

### Mixed Local/Remote

Some services local (development databases), others remote (production monitoring):

```yaml
services:
  postgresql:
    host: "localhost"
    port: "5432"
    remote: false        # Started locally via Docker
    required: true
  prometheus:
    host: "monitoring.internal"
    port: "9090"
    remote: true         # Remote monitoring stack
    required: false
```

### Environment-Based Override

Use environment variables for deployment flexibility without changing YAML:

```bash
# Override for staging
export SVC_POSTGRESQL_HOST=staging-db.internal
export SVC_POSTGRESQL_REMOTE=true
export SVC_REDIS_HOST=staging-redis.internal
export SVC_REDIS_REMOTE=true
./bin/helixagent
```

## Diagrams

For visual architecture diagrams, see:

- `docs/diagrams/src/architecture-overview.mmd` - System architecture (Mermaid)
- `docs/diagrams/src/service-dependencies.mmd` - Service dependency graph (Mermaid)
- `docs/diagrams/src/boot-sequence.mmd` - Boot sequence diagram (Mermaid)
- `docs/diagrams/src/shutdown-sequence.mmd` - Shutdown sequence (Mermaid)
- `docs/diagrams/src/architecture.puml` - UML component diagram (PlantUML)
- `docs/diagrams/src/boot-sequence.puml` - UML sequence diagram (PlantUML)

Generate rendered diagrams:
```bash
./scripts/generate-diagrams.sh
```

Output in `docs/diagrams/output/{svg,png,pdf}/`.
