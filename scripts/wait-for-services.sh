#!/bin/bash

# ============================================================================
# Wait for Services to be Healthy
# ============================================================================
# Waits for all big data services to pass health checks before proceeding
# Usage: ./scripts/wait-for-services.sh [timeout_seconds]
# ============================================================================

TIMEOUT=${1:-300}  # Default 5 minutes
ELAPSED=0
INTERVAL=5

echo "Waiting for all services to be healthy (timeout: ${TIMEOUT}s)..."

while [ $ELAPSED -lt $TIMEOUT ]; do
    # Count healthy services
    HEALTHY=$(docker ps --filter "health=healthy" --filter "name=helixagent-*" | wc -l)
    HEALTHY=$((HEALTHY - 1))  # Subtract header

    TOTAL=$(docker ps --filter "name=helixagent-*" | wc -l)
    TOTAL=$((TOTAL - 1))

    echo -n "Progress: $HEALTHY/$TOTAL healthy... "

    if [ $HEALTHY -eq $TOTAL ] && [ $TOTAL -gt 0 ]; then
        echo "✓ All services healthy!"
        exit 0
    fi

    echo "waiting..."
    sleep $INTERVAL
    ELAPSED=$((ELAPSED + INTERVAL))
done

echo "✗ Timeout reached. Some services may not be healthy."
./scripts/check-bigdata-services.sh
exit 1
