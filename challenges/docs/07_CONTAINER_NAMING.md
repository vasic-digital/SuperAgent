# Container Naming & Cleanup Conventions

## Overview

All challenge scripts that create Docker containers **MUST** follow consistent naming conventions and cleanup procedures to avoid conflicts with production deployments and ensure test isolation.

## Naming Convention

### Format

```
helixagent-<identifier>-<service>-challenge
```

### Components

| Component | Description | Examples |
|-----------|-------------|----------|
| `helixagent` | Prefix identifying HelixAgent containers | Always required |
| `<identifier>` | Challenge identifier or timestamp | `remote-deploy`, `discovery`, `20260130_1430` |
| `<service>` | Service name (lowercase, hyphenated) | `postgresql`, `redis`, `cognee` |
| `challenge` | Suffix indicating challenge container | Always required |

### Examples

```bash
# Remote deployment challenge
helixagent-remote-deploy-postgresql-challenge
helixagent-remote-deploy-redis-challenge

# Service discovery challenge  
helixagent-discovery-postgresql-challenge
helixagent-discovery-redis-challenge

# Timestamp-based (for parallel executions)
helixagent-20260130_1430-postgresql-challenge
helixagent-20260130_1430-redis-challenge

# Full system boot challenge
helixagent-full-boot-postgresql-challenge
helixagent-full-boot-redis-challenge
helixagent-full-boot-helixagent-challenge
```

## Cleanup Requirements

### Mandatory Cleanup Steps

All challenge scripts **MUST**:

1. **Stop containers**: `docker stop <challenge-container>`
2. **Remove containers**: `docker rm <challenge-container>`
3. **Remove volumes**: `docker volume rm <challenge-volume>`
4. **Remove networks**: `docker network rm <challenge-network>`
5. **Remove files**: Clean up temporary files and configurations

### Cleanup Implementation Pattern

```bash
#!/bin/bash

# Configuration
CHALLENGE_ID="remote-deploy"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
CONTAINER_PREFIX="helixagent-${CHALLENGE_ID}"

# Services to manage
SERVICES=("postgresql" "redis" "cognee")

# Cleanup function (always called on exit)
cleanup() {
    echo "Cleaning up challenge containers..."
    
    # Stop and remove containers
    for service in "${SERVICES[@]}"; do
        CONTAINER_NAME="${CONTAINER_PREFIX}-${service}-challenge"
        echo "  Stopping container: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" 2>/dev/null || true
        docker rm "$CONTAINER_NAME" 2>/dev/null || true
        
        # Remove associated volumes
        VOLUME_NAME="${CONTAINER_PREFIX}-${service}-data-challenge"
        docker volume rm "$VOLUME_NAME" 2>/dev/null || true
    done
    
    # Remove networks
    NETWORK_NAME="${CONTAINER_PREFIX}-network-challenge"
    docker network rm "$NETWORK_NAME" 2>/dev/null || true
    
    # Remove temporary files
    rm -f /tmp/${CONTAINER_PREFIX}-*.yaml
    rm -f /tmp/${CONTAINER_PREFIX}-*.log
    
    echo "Cleanup complete"
}

# Register cleanup trap
trap cleanup EXIT INT TERM

# Main execution
echo "Starting challenge with containers prefixed: $CONTAINER_PREFIX"
# ... challenge logic ...
```

### Docker Compose Naming

When using Docker Compose, set project name:

```bash
# Use project name to namespace containers
COMPOSE_PROJECT_NAME="helixagent-${CHALLENGE_ID}-challenge"
export COMPOSE_PROJECT_NAME

# Start services
docker-compose up -d

# Services will be named:
# helixagent-remote-deploy-challenge-postgresql-1
# helixagent-remote-deploy-challenge-redis-1
```

### Volume Naming

```bash
# Format: helixagent-<identifier>-<service>-<type>-challenge
helixagent-remote-deploy-postgresql-data-challenge
helixagent-discovery-redis-cache-challenge
helixagent-full-boot-cognee-storage-challenge
```

### Network Naming

```bash
# Format: helixagent-<identifier>-network-challenge
helixagent-remote-deploy-network-challenge
helixagent-discovery-network-challenge
helixagent-full-boot-network-challenge
```

## Challenge Script Requirements

### 1. Pre-Execution Validation

Check for existing containers with same names:

```bash
validate_no_conflicts() {
    CONFLICTS=$(docker ps -a --filter "name=helixagent-.*-challenge" --format "{{.Names}}")
    if [[ -n "$CONFLICTS" ]]; then
        echo "ERROR: Existing challenge containers found:"
        echo "$CONFLICTS"
        echo "Run cleanup script or remove manually:"
        echo "  ./challenges/scripts/cleanup_results.sh"
        exit 1
    fi
}
```

### 2. Container Creation

Use explicit names:

```bash
# Create network
docker network create helixagent-${CHALLENGE_ID}-network-challenge

# Start PostgreSQL container
docker run -d \
  --name helixagent-${CHALLENGE_ID}-postgresql-challenge \
  --network helixagent-${CHALLENGE_ID}-network-challenge \
  -e POSTGRES_PASSWORD=helixagent123 \
  -v helixagent-${CHALLENGE_ID}-postgresql-data-challenge:/var/lib/postgresql/data \
  postgres:15
```

### 3. Health Checking

Check containers with their full names:

```bash
check_container_health() {
    CONTAINER_NAME="helixagent-${CHALLENGE_ID}-postgresql-challenge"
    if docker exec "$CONTAINER_NAME" pg_isready -U postgres; then
        echo "✓ $CONTAINER_NAME is healthy"
    else
        echo "✗ $CONTAINER_NAME is unhealthy"
        return 1
    fi
}
```

### 4. Post-Execution Cleanup

Ensure cleanup runs even on failure:

```bash
# Setup trap first
trap cleanup EXIT INT TERM

# Execute challenge
main() {
    # ... challenge logic ...
}

# Run main, capture exit code
main "$@"
EXIT_CODE=$?

# Cleanup will run via trap
exit $EXIT_CODE
```

## Example Script Template

```bash
#!/bin/bash
# challenges/scripts/example_challenge.sh

set -e

#===============================================================================
# CONFIGURATION
#===============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

CHALLENGE_ID="example"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
CONTAINER_PREFIX="helixagent-${CHALLENGE_ID}"
NETWORK_NAME="${CONTAINER_PREFIX}-network-challenge"

# Services to manage
SERVICES=("postgresql" "redis")

#===============================================================================
# CLEANUP FUNCTION
#===============================================================================
cleanup() {
    echo "🧹 Cleaning up challenge containers..."
    
    # Stop and remove containers
    for service in "${SERVICES[@]}"; do
        CONTAINER_NAME="${CONTAINER_PREFIX}-${service}-challenge"
        echo "  Stopping: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" 2>/dev/null || true
        docker rm "$CONTAINER_NAME" 2>/dev/null || true
        
        # Remove volumes
        VOLUME_NAME="${CONTAINER_PREFIX}-${service}-data-challenge"
        docker volume rm "$VOLUME_NAME" 2>/dev/null || true
    done
    
    # Remove network
    docker network rm "$NETWORK_NAME" 2>/dev/null || true
    
    echo "✅ Cleanup complete"
}

# Register cleanup trap
trap cleanup EXIT INT TERM

#===============================================================================
# VALIDATION
#===============================================================================
validate_no_conflicts() {
    echo "🔍 Checking for existing challenge containers..."
    CONFLICTS=$(docker ps -a --filter "name=${CONTAINER_PREFIX}.*-challenge" --format "{{.Names}}" 2>/dev/null || true)
    if [[ -n "$CONFLICTS" ]]; then
        echo "❌ Existing challenge containers found:"
        echo "$CONFLICTS"
        echo "Run cleanup script or remove manually."
        exit 1
    fi
    echo "✅ No conflicts found"
}

#===============================================================================
# MAIN EXECUTION
#===============================================================================
main() {
    echo "🚀 Starting ${CHALLENGE_ID} challenge"
    
    # Validate no conflicts
    validate_no_conflicts
    
    # Create network
    echo "🌐 Creating network: $NETWORK_NAME"
    docker network create "$NETWORK_NAME"
    
    # Start services
    for service in "${SERVICES[@]}"; do
        CONTAINER_NAME="${CONTAINER_PREFIX}-${service}-challenge"
        echo "🐳 Starting container: $CONTAINER_NAME"
        
        case $service in
            postgresql)
                docker run -d \
                  --name "$CONTAINER_NAME" \
                  --network "$NETWORK_NAME" \
                  -e POSTGRES_PASSWORD=helixagent123 \
                  -v "${CONTAINER_PREFIX}-${service}-data-challenge":/var/lib/postgresql/data \
                  postgres:15
                ;;
            redis)
                docker run -d \
                  --name "$CONTAINER_NAME" \
                  --network "$NETWORK_NAME" \
                  -v "${CONTAINER_PREFIX}-${service}-data-challenge":/data \
                  redis:7-alpine \
                  redis-server --appendonly yes
                ;;
        esac
    done
    
    # Wait for services to be ready
    echo "⏳ Waiting for services to be ready..."
    sleep 10
    
    # Validate services
    echo "✅ Challenge setup complete"
    echo "📊 Containers running:"
    docker ps --filter "name=${CONTAINER_PREFIX}.*-challenge" --format "table {{.Names}}\t{{.Status}}"
}

#===============================================================================
# EXECUTION
#===============================================================================
main "$@"
```

## Integration with Existing Challenges

### Update Existing Scripts

All existing challenge scripts that create containers must be updated to follow this convention. Key scripts include:

1. `full_system_boot_challenge.sh`
2. `unified_service_boot_challenge.sh`  
3. `cognee_full_integration_challenge.sh`
4. `mcp_containerized_challenge.sh`
5. `remote_services_challenge.sh`

### Common Functions

Consider creating a shared library for container management:

```bash
# challenges/scripts/lib/container_utils.sh

container_prefix() {
    local challenge_id="$1"
    echo "helixagent-${challenge_id}"
}

container_name() {
    local challenge_id="$1"
    local service="$2"
    echo "$(container_prefix "$challenge_id")-${service}-challenge"
}

volume_name() {
    local challenge_id="$1"
    local service="$2"
    local type="${3:-data}"
    echo "$(container_prefix "$challenge_id")-${service}-${type}-challenge"
}

network_name() {
    local challenge_id="$1"
    echo "$(container_prefix "$challenge_id")-network-challenge"
}

cleanup_containers() {
    local challenge_id="$1"
    shift
    local services=("$@")
    
    local prefix=$(container_prefix "$challenge_id")
    
    for service in "${services[@]}"; do
        docker stop "${prefix}-${service}-challenge" 2>/dev/null || true
        docker rm "${prefix}-${service}-challenge" 2>/dev/null || true
        docker volume rm "${prefix}-${service}-data-challenge" 2>/dev/null || true
    done
    
    docker network rm "${prefix}-network-challenge" 2>/dev/null || true
}
```

## Validation

### Pre-Commit Hook

Consider adding a pre-commit hook to validate container naming:

```bash
#!/bin/bash
# .githooks/pre-commit/check-container-names.sh

# Check for hardcoded container names in challenge scripts
PATTERN="docker.*--name.*helixagent.*challenge"
if grep -r "$PATTERN" challenges/scripts/ --include="*.sh" | grep -v "helixagent-.*-.*-challenge"; then
    echo "ERROR: Challenge scripts with non-standard container names found"
    grep -r "$PATTERN" challenges/scripts/ --include="*.sh" | grep -v "helixagent-.*-.*-challenge"
    exit 1
fi
```

### CI/CD Validation

Add to CI pipeline:

```yaml
# .github/workflows/validate-challenges.yml
jobs:
  validate-container-names:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate container naming
        run: |
          ./challenges/scripts/validate_container_naming.sh
```

## Summary

Following these conventions ensures:

1. **Isolation**: Challenge containers don't interfere with production
2. **Cleanup**: No leftover containers after tests
3. **Consistency**: Predictable naming across all challenges
4. **Debugging**: Easy identification of challenge containers
5. **Parallel Execution**: Multiple challenges can run simultaneously

All new challenge scripts **MUST** follow this convention, and existing scripts should be updated during their next modification cycle.
