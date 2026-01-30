# Remote Deployment & Service Discovery

HelixAgent supports **distributed deployment** across multiple hosts with **automatic service discovery**. This enables hybrid environments where services can be deployed locally, on remote machines, or discovered in your network.

## Overview

### Key Features
- **Service Discovery**: Automatic detection of PostgreSQL, Redis, ChromaDB, Cognee, and other services
- **Remote Deployment**: Deploy services to remote hosts via SSH
- **Parallel Deployment**: Deploy to multiple hosts simultaneously
- **Hybrid Environments**: Mix local, remote, and discovered services
- **BootManager Integration**: Unified service management via `BootManager`

### Use Cases
- **Distributed Testing**: Run different services on different machines
- **Resource Optimization**: Place compute-intensive services on powerful hosts
- **High Availability**: Deploy redundant services across multiple hosts
- **Development Teams**: Shared development infrastructure

## Service Discovery

HelixAgent includes a comprehensive service discovery system that can detect services using multiple strategies.

### Discovery Methods

| Method | Description | Use Case |
|--------|-------------|----------|
| **TCP** | Port scanning (22, 5432, 6379, 8000, etc.) | Fast detection of known ports |
| **HTTP** | HTTP health endpoint checks (`/health`, `/ready`) | Services with HTTP APIs |
| **DNS** | DNS SRV record lookup | Kubernetes/cloud environments |
| **mDNS** | Multicast DNS (`.local` hostnames) | Local network discovery |

### Discovery Script

The `discover-services.sh` script provides a unified interface for service discovery:

```bash
# Basic discovery
./scripts/discover-services.sh

# Use specific strategy
./scripts/discover-services.sh --strategy=tcp
./scripts/discover-services.sh --strategy=http
./scripts/discover-services.sh --strategy=dns
./scripts/discover-services.sh --strategy=mdns

# Test all strategies
./scripts/discover-services.sh --test

# Output format (JSON)
./scripts/discover-services.sh --output=json

# Save results to file
./scripts/discover-services.sh --save=discovered-services.yaml
```

### Configuration

Discovery behavior can be configured via environment variables:

```bash
# Port ranges to scan
export DISCOVERY_TCP_PORTS="22,5432,6379,8000,8080,9090,3000"
export DISCOVERY_HTTP_PORTS="8080,7061,9090,3000"

# Timeouts
export DISCOVERY_TCP_TIMEOUT=2
export DISCOVERY_HTTP_TIMEOUT=5

# Network ranges (CIDR)
export DISCOVERY_NETWORK_RANGES="192.168.1.0/24,10.0.0.0/8"
```

### Integration with BootManager

The `BootManager` automatically uses discovery results to avoid starting redundant services:

```go
// During boot, discover existing services
discovered := bootManager.DiscoverServices(ctx)

// Only start services not already available
servicesToStart := bootManager.FilterDiscoveredServices(discovered)
bootManager.StartServices(ctx, servicesToStart)
```

## Remote Deployment

### SSH Remote Deployer

The `SSHRemoteDeployer` provides remote deployment capabilities via SSH.

#### Architecture

```go
type SSHRemoteDeployer struct {
    runner SSHCommandRunner
    config RemoteDeploymentConfig
}

type SSHCommandRunner interface {
    RunCommand(ctx context.Context, host string, cmd string) (string, error)
    CopyFile(ctx context.Context, host string, src string, dst string) error
}
```

#### Deployment Process

1. **Connect Validation**: Verify SSH connectivity to remote host
2. **Docker Check**: Ensure Docker and Docker Compose are installed
3. **File Transfer**: Copy Docker Compose files to remote host
4. **Service Deployment**: Start services via `docker compose up -d`
5. **Health Verification**: Wait for services to become healthy

### Deployment Scripts

#### Single-Host Deployment

```bash
# Deploy to a single remote host
./scripts/deploy-remote.sh thinker.local

# Deploy specific services
./scripts/deploy-remote.sh thinker.local --services=postgresql,redis

# Deploy with custom compose file
./scripts/deploy-remote.sh thinker.local --compose-file=docker-compose.prod.yml

# Deploy with environment variables
./scripts/deploy-remote.sh thinker.local --env=production

# Dry run (show commands without executing)
./scripts/deploy-remote.sh thinker.local --dry-run
```

#### Multi-Host Parallel Deployment

```bash
# Deploy to all configured hosts in parallel
./scripts/deploy-all-remote.sh

# Use custom configuration
./scripts/deploy-all-remote.sh --config=configs/remote-hosts.yaml

# Limit concurrent deployments
./scripts/deploy-all-remote.sh --max-concurrent=3

# Skip health checks for faster deployment
./scripts/deploy-all-remote.sh --skip-health-check
```

### Configuration

#### Remote Hosts Configuration (`configs/remote-hosts.yaml`)

```yaml
remote_deployment:
  enabled: true
  default_user: "ubuntu"
  ssh_key_path: "~/.ssh/id_rsa"
  ssh_timeout: 30
  
  hosts:
    thinker.local:
      user: "milos"
      services:
        - postgresql
        - redis
        - cognee
      compose_file: "docker-compose.thinker.yml"
      environment: "production"
      
    raspberrypi.local:
      user: "pi"
      services:
        - ollama
        - zen
      compose_file: "docker-compose.raspberrypi.yml"
      environment: "development"
      
    gpu-server.local:
      user: "gpu"
      services:
        - sglang
        - llamaindex
        - cognee
      compose_file: "docker-compose.gpu.yml"
      environment: "production"
      gpu_enabled: true
```

#### Environment-Specific Configuration

```yaml
# configs/remote-deployment/development.yaml
remote_deployment:
  hosts:
    localhost:
      services:
        - postgresql
        - redis
      compose_file: "docker-compose.dev.yml"

# configs/remote-deployment/production.yaml
remote_deployment:
  hosts:
    db01.prod.example.com:
      services:
        - postgresql
      compose_file: "docker-compose.db.yml"
    
    cache01.prod.example.com:
      services:
        - redis
      compose_file: "docker-compose.cache.yml"
    
    llm01.prod.example.com:
      services:
        - claude
        - deepseek
        - gemini
      compose_file: "docker-compose.llm.yml"
```

### BootManager Integration

The `BootManager` has been extended to support remote deployment:

```go
// Deploy services to remote hosts
err := bootManager.DeployRemoteServices(ctx)
if err != nil {
    log.Printf("Remote deployment failed: %v", err)
    // Fall back to local deployment
    err = bootManager.StartLocalServices(ctx)
}

// Health check remote services
status := bootManager.HealthCheckRemoteServices(ctx)
if !status.Healthy {
    log.Printf("Remote services unhealthy: %v", status.Errors)
}

// Combined health check (local + remote)
combinedStatus := bootManager.HealthCheckAllServices(ctx)
```

#### New BootManager Methods

| Method | Description |
|--------|-------------|
| `DeployRemoteServices(ctx)` | Deploy services to configured remote hosts |
| `HealthCheckRemoteServices(ctx)` | Health check remote services only |
| `DiscoverServices(ctx)` | Discover services in network |
| `FilterDiscoveredServices(discovered)` | Filter out already discovered services |

## Hybrid Environments

### Service Placement Strategies

| Strategy | Description | Example |
|----------|-------------|---------|
| **Local First** | Run all services locally, use remote only if needed | Development environment |
| **Remote Compute** | Run LLM providers on GPU servers, databases locally | AI development |
| **Distributed HA** | Run redundant services across multiple hosts | Production deployment |
| **Dynamic Discovery** | Discover existing services, deploy missing ones | Team development |

### Configuration Examples

#### Development (Local + Remote LLMs)
```yaml
services:
  postgresql: local
  redis: local
  chromadb: local
  claude: remote@thinker.local
  deepseek: remote@gpu-server.local
  cognee: remote@thinker.local
```

#### Production (Fully Distributed)
```yaml
services:
  postgresql: remote@db01.prod.example.com
  redis: remote@cache01.prod.example.com
  chromadb: remote@vector01.prod.example.com
  claude: remote@llm01.prod.example.com
  deepseek: remote@llm02.prod.example.com
  gemini: remote@llm03.prod.example.com
  cognee: remote@knowledge01.prod.example.com
```

#### Team Development (Shared Infrastructure)
```yaml
services:
  postgresql: discovered  # Use existing team DB
  redis: discovered       # Use existing team cache
  chromadb: local         # Personal vector store
  claude: remote@shared-gpu.local
  cognee: local          # Personal knowledge graph
```

## Challenge Scripts

Remote deployment and discovery are validated through comprehensive challenge scripts.

### Remote Deployment Challenge

```bash
# Run the remote deployment challenge
./challenges/scripts/remote_deployment_challenge.sh

# Test with specific host
./challenges/scripts/remote_deployment_challenge.sh --host=thinker.local

# Test discovery only
./challenges/scripts/remote_deployment_challenge.sh --discovery-only
```

### Challenge Container Naming Convention

Challenge scripts use a specific naming convention to avoid conflicts with production containers:

```
helixagent-<identifier>-<service>-challenge
```

Examples:
- `helixagent-remote-deploy-postgresql-challenge`
- `helixagent-discovery-redis-challenge`
- `helixagent-challenge-<timestamp>-<service>`

### Cleanup Requirements

All challenge scripts **must** clean up after themselves:

1. **Stop containers**: `docker stop <challenge-container>`
2. **Remove containers**: `docker rm <challenge-container>`
3. **Remove volumes**: `docker volume rm <challenge-volume>`
4. **Remove networks**: `docker network rm <challenge-network>`
5. **Remove files**: Clean up temporary files and configurations

Example cleanup:
```bash
# Always run cleanup, even on failure
cleanup() {
    docker stop helixagent-challenge-postgresql 2>/dev/null || true
    docker rm helixagent-challenge-postgresql 2>/dev/null || true
    docker volume rm helixagent-challenge-postgresql-data 2>/dev/null || true
    rm -f /tmp/helixagent-challenge-*.yaml
}
trap cleanup EXIT
```

## Testing

### Unit Tests

```bash
# Run remote deployment unit tests
go test -v ./internal/services -run TestSSHRemoteDeployer

# Run discovery unit tests
go test -v ./internal/services -run TestServiceDiscovery

# Run BootManager extension tests
go test -v ./internal/services -run TestBootManagerRemote
```

### Integration Tests

Integration tests require SSH access to test hosts:

```bash
# Set up test environment
export TEST_REMOTE_HOST=thinker.local
export TEST_SSH_USER=milos
export TEST_SSH_KEY=~/.ssh/id_rsa

# Run integration tests
go test -v ./tests/integration -run TestRemoteDeployment
```

### Manual Testing

1. **Discovery Test**:
   ```bash
   ./scripts/discover-services.sh --test
   ```

2. **Single Host Deployment Test**:
   ```bash
   ./scripts/deploy-remote.sh thinker.local --services=postgresql --dry-run
   ```

3. **Multi-Host Deployment Test**:
   ```bash
   ./scripts/deploy-all-remote.sh --dry-run
   ```

## Troubleshooting

### Common Issues

#### SSH Connection Failed
```bash
# Test SSH connectivity
ssh thinker.local "echo Connected"

# Check SSH key permissions
chmod 600 ~/.ssh/id_rsa

# Verify SSH configuration
cat ~/.ssh/config
```

#### Docker Not Installed on Remote Host
```bash
# Check Docker installation
ssh thinker.local "docker --version"

# Install Docker (Ubuntu)
ssh thinker.local "curl -fsSL https://get.docker.com | sh"
```

#### Port Already In Use
```bash
# Check ports on remote host
ssh thinker.local "netstat -tuln | grep :5432"

# Use different ports in compose file
services:
  postgresql:
    ports:
      - "5433:5432"  # Map host port 5433 to container 5432
```

#### Permission Denied
```bash
# Add user to docker group
ssh thinker.local "sudo usermod -aG docker $USER"

# Re-login to apply group changes
ssh thinker.local "newgrp docker"
```

### Debug Mode

Enable debug logging for detailed output:

```bash
export REMOTE_DEPLOYMENT_DEBUG=true
export DISCOVERY_DEBUG=true
./scripts/deploy-remote.sh thinker.local
```

## Security Considerations

### SSH Key Management
- Use dedicated deployment keys with limited permissions
- Store keys in a secure location (not in repository)
- Rotate keys regularly
- Use SSH agents for key management

### Network Security
- Limit SSH access to specific IP ranges
- Use VPN for remote host access
- Encrypt traffic with SSH tunneling
- Firewall rules for service ports

### Container Security
- Run containers as non-root users
- Use read-only filesystems where possible
- Limit container capabilities
- Regular security updates

## Performance Optimization

### Parallel Deployment
- Deploy to multiple hosts simultaneously
- Use connection pooling for SSH
- Batch file transfers
- Async health checks

### Caching
- Cache discovery results (TTL: 5 minutes)
- Cache remote host status
- Cache Docker image pulls
- Cache configuration files

### Resource Management
- Limit concurrent deployments
- Timeout long-running operations
- Retry with exponential backoff
- Circuit breaker pattern

## API Reference

### Remote Deployment Service

```go
// Create deployer
deployer := services.NewSSHRemoteDeployer(config)

// Deploy services
err := deployer.DeployServices(ctx, hostConfig)

// Check deployment status
status := deployer.GetDeploymentStatus(ctx, host)

// Cleanup deployment
err := deployer.Cleanup(ctx, host)
```

### Service Discovery Service

```go
// Create discoverer
discoverer := services.NewServiceDiscoverer(config)

// Discover services
services, err := discoverer.Discover(ctx)

// Filter by type
dbServices := discoverer.FilterByType(services, "database")

// Health check discovered services
health := discoverer.HealthCheck(ctx, services)
```

## Examples

### Complete Deployment Workflow

```bash
#!/bin/bash
# deploy-helixagent.sh

# Discover existing services
echo "Discovering existing services..."
DISCOVERED=$(./scripts/discover-services.sh --output=json)

# Parse discovered services
POSTGRES_DISCOVERED=$(echo "$DISCOVERED" | jq '.services[] | select(.type=="postgresql")')

if [ -z "$POSTGRES_DISCOVERED" ]; then
    echo "No PostgreSQL found, deploying..."
    ./scripts/deploy-remote.sh db-host.local --services=postgresql
else
    echo "Using existing PostgreSQL: $POSTGRES_DISCOVERED"
fi

# Deploy remaining services
echo "Deploying Redis..."
./scripts/deploy-remote.sh cache-host.local --services=redis

echo "Deploying LLM providers..."
./scripts/deploy-remote.sh llm-host.local --services=claude,deepseek,gemini

echo "Deployment complete!"
```

### Hybrid Environment Setup

```go
// hybrid_setup.go
package main

import (
    "context"
    "dev.helix.agent/internal/services"
)

func main() {
    ctx := context.Background()
    
    // Discover existing services
    bootManager := services.NewBootManager()
    discovered := bootManager.DiscoverServices(ctx)
    
    // Configure hybrid deployment
    config := services.HybridConfig{
        LocalServices: []string{"chromadb", "cognee"},
        RemoteServices: map[string][]string{
            "thinker.local": {"postgresql", "redis"},
            "gpu-server.local": {"claude", "deepseek", "gemini"},
        },
        UseDiscovered: true,
    }
    
    // Deploy hybrid environment
    err := bootManager.DeployHybrid(ctx, config, discovered)
    if err != nil {
        panic(err)
    }
    
    // Start HelixAgent
    err = bootManager.StartHelixAgent(ctx)
    if err != nil {
        panic(err)
    }
}
```

## Conclusion

HelixAgent's remote deployment and service discovery system enables flexible, distributed architectures. By combining discovery, remote deployment, and hybrid environments, you can optimize resource usage, improve availability, and simplify infrastructure management.

### Next Steps
1. Configure remote hosts in `configs/remote-hosts.yaml`
2. Test discovery with `./scripts/discover-services.sh --test`
3. Deploy to a test host with `./scripts/deploy-remote.sh <host>`
4. Run the challenge script: `./challenges/scripts/remote_deployment_challenge.sh`

### Additional Resources
- [BootManager Documentation](SERVICE_MANAGEMENT.md)
- [Docker Compose Configuration](../docker/README.md)
- [Security Best Practices](../security/SANDBOXING.md)
- [Monitoring and Observability](../monitoring/MONITORING_SYSTEM.md)
