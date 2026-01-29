# Formatter Services Docker Containers

This directory contains Docker containers for **14 service-type formatters** that run as HTTP services.

## Overview

Service formatters run as isolated HTTP services that accept formatting requests via REST API. This provides several benefits:

- **Isolation**: Each formatter runs in its own container with specific dependencies
- **Scalability**: Services can be scaled independently
- **Language Independence**: Any language can call the HTTP API
- **Reliability**: Container failures don't affect other formatters
- **Consistency**: Same environment across dev/staging/production

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│             HelixAgent Formatters API                    │
│            (POST /v1/format)                             │
└──────────────┬──────────────────────────────────────────┘
               │
               ├─ Native Formatters (Direct Binary Execution)
               │  ├─ black, ruff (Python)
               │  ├─ prettier, biome (JavaScript/TypeScript)
               │  ├─ gofmt (Go)
               │  ├─ rustfmt (Rust)
               │  ├─ clang-format (C/C++)
               │  ├─ shfmt (Shell)
               │  ├─ yamlfmt (YAML)
               │  ├─ taplo (TOML)
               │  └─ stylua (Lua)
               │
               └─ Service Formatters (HTTP Services)
                  ├─ autopep8 (Python) - Port 9211
                  ├─ yapf (Python) - Port 9210
                  ├─ sqlfluff (SQL) - Port 9220
                  ├─ rubocop (Ruby) - Port 9230
                  ├─ standardrb (Ruby) - Port 9231
                  ├─ php-cs-fixer (PHP) - Port 9240
                  ├─ laravel-pint (PHP) - Port 9241
                  ├─ perltidy (Perl) - Port 9250
                  ├─ cljfmt (Clojure) - Port 9260
                  ├─ spotless (Java/Kotlin) - Port 9270
                  ├─ groovy-lint (Groovy) - Port 9280
                  ├─ styler (R) - Port 9290
                  ├─ air (R, 300x faster) - Port 9291
                  └─ psscriptanalyzer (PowerShell) - Port 9300
```

## Port Allocation

| Port | Formatter | Language | Version |
|------|-----------|----------|---------|
| 9210 | yapf | Python | 0.40.2 |
| 9211 | autopep8 | Python | 2.0.4 |
| 9220 | sqlfluff | SQL | 3.4.1 |
| 9230 | rubocop | Ruby | 1.72.0 |
| 9231 | standardrb | Ruby | 1.42.1 |
| 9240 | php-cs-fixer | PHP | 3.68.0 |
| 9241 | laravel-pint | PHP | 1.19.0 |
| 9250 | perltidy | Perl | 20260109.01 |
| 9260 | cljfmt | Clojure | 0.12.0 |
| 9270 | spotless | Java/Kotlin | 7.0.0.BETA4 |
| 9280 | groovy-lint | Groovy | 15.0.4 |
| 9290 | styler | R | 1.10.3 |
| 9291 | air | R | 0.2.0 |
| 9300 | psscriptanalyzer | PowerShell | 1.23.0 |

## Quick Start

### Build All Containers

```bash
./build-all.sh
```

This builds 14 Docker images with all service formatters.

### Start Services

```bash
# Start all formatters
docker-compose -f docker-compose.formatters.yml up -d

# Start specific formatters
docker-compose -f docker-compose.formatters.yml up -d autopep8 yapf sqlfluff

# View logs
docker-compose -f docker-compose.formatters.yml logs -f

# Check status
docker-compose -f docker-compose.formatters.yml ps
```

### Stop Services

```bash
# Stop all formatters
docker-compose -f docker-compose.formatters.yml down

# Stop and remove volumes
docker-compose -f docker-compose.formatters.yml down -v
```

## Service API

Each formatter service exposes two endpoints:

### Health Check

```bash
GET http://localhost:{PORT}/health

Response:
{
  "status": "healthy",
  "formatter": "autopep8",
  "version": "2.0.4"
}
```

### Format Code

```bash
POST http://localhost:{PORT}/format
Content-Type: application/json

{
  "content": "def hello(  x,y ):\n  return x+y",
  "options": {}
}

Response:
{
  "success": true,
  "content": "def hello(x, y):\n    return x + y\n",
  "changed": true,
  "formatter": "autopep8"
}
```

## Examples

### Python (autopep8)

```bash
curl -X POST http://localhost:9211/format \
  -H 'Content-Type: application/json' \
  -d '{
    "content": "def hello(  x,y ):\n  return x+y"
  }'
```

### SQL (sqlfluff)

```bash
curl -X POST http://localhost:9220/format \
  -H 'Content-Type: application/json' \
  -d '{
    "content": "SELECT * FROM users WHERE id=1;"
  }'
```

### Ruby (rubocop)

```bash
curl -X POST http://localhost:9230/format \
  -H 'Content-Type: application/json' \
  -d '{
    "content": "def hello\nputs \"hello\"\nend"
  }'
```

### PHP (php-cs-fixer)

```bash
curl -X POST http://localhost:9240/format \
  -H 'Content-Type: application/json' \
  -d '{
    "content": "<?php\n$x=1+2;\necho $x;"
  }'
```

## Integration with HelixAgent

HelixAgent automatically detects and uses service formatters when available:

1. **Auto-detection**: HelixAgent checks for service formatters at startup
2. **Fallback**: If service is unavailable, falls back to native formatters
3. **Health Monitoring**: Regular health checks ensure services are available
4. **Load Balancing**: Distributes requests across available formatters

### Configuration

Add to `configs/development.yaml`:

```yaml
formatters:
  service_formatters:
    enabled: true
    base_url: "http://localhost"
    health_check_interval: 30s
    timeout: 30s
    retry_count: 3

    endpoints:
      autopep8:
        port: 9211
        enabled: true
      yapf:
        port: 9210
        enabled: true
      sqlfluff:
        port: 9220
        enabled: true
      rubocop:
        port: 9230
        enabled: true
      php-cs-fixer:
        port: 9240
        enabled: true
      # ... etc
```

## Troubleshooting

### Container Won't Start

Check logs:
```bash
docker-compose -f docker-compose.formatters.yml logs autopep8
```

### Health Check Fails

Test manually:
```bash
curl http://localhost:9211/health
```

### Formatting Fails

Check formatter-specific logs:
```bash
docker logs formatter-autopep8
```

### Port Conflicts

If ports are already in use, modify `docker-compose.formatters.yml`:
```yaml
ports:
  - "19211:9211"  # Use different external port
```

## Development

### Adding a New Formatter

1. Create Dockerfile:
```dockerfile
# Dockerfile.myformatter
FROM language:version
RUN install-formatter
COPY formatter-service.py /app/
CMD ["python", "/app/formatter-service.py", "--formatter", "myformatter", "--port", "9999"]
```

2. Add to `docker-compose.formatters.yml`:
```yaml
myformatter:
  build:
    context: .
    dockerfile: Dockerfile.myformatter
  ports:
    - "9999:9999"
  networks:
    - formatters
```

3. Add to `build-all.sh`:
```bash
FORMATTERS=(
    # ... existing formatters
    "myformatter"
)
```

### HTTP Service Wrapper

The `formatter-service.py` script provides a universal HTTP wrapper for CLI formatters:

- **Input**: JSON with `content` and `options` fields
- **Execution**: Pipes content through formatter CLI
- **Output**: JSON with `success`, `content`, `changed` fields
- **Health**: `/health` endpoint checks formatter availability

## Performance

Service formatters add ~10-50ms overhead compared to native execution due to HTTP roundtrip. However, benefits include:

- **Isolation**: Failures don't crash HelixAgent
- **Scalability**: Run multiple instances for high load
- **Caching**: HTTP layer can cache results
- **Monitoring**: Easier to monitor and debug

## Security

- **Non-root**: All containers run as non-root user (UID 1000)
- **Read-only**: Containers have read-only filesystem (except /workspace)
- **Network isolation**: Services run on isolated Docker network
- **Resource limits**: CPU and memory limits configured

## Monitoring

Health checks run every 30 seconds:

```bash
# Check all services
docker-compose -f docker-compose.formatters.yml ps

# View health status
for port in 9210 9211 9220 9230 9240 9250 9260 9270 9280 9290 9291 9300; do
  echo "Port $port: $(curl -s http://localhost:$port/health | jq -r '.status')"
done
```

## Production Deployment

For production, consider:

1. **Orchestration**: Use Kubernetes or Docker Swarm
2. **Scaling**: Run multiple replicas per formatter
3. **Load Balancing**: Add nginx/HAProxy in front
4. **Monitoring**: Integrate with Prometheus/Grafana
5. **Logging**: Centralized logging with ELK/Loki
6. **Security**: Add authentication and rate limiting

## License

Copyright © 2026 HelixAgent
