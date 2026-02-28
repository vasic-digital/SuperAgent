#!/bin/bash
# Complete Remote Distribution Reset Script
# Kills everything and restarts with remote distribution enabled

set -e

echo "========================================"
echo "COMPLETE REMOTE DISTRIBUTION RESET"
echo "========================================"
echo ""

# 1. Stop HelixAgent
echo "[1/8] Stopping HelixAgent..."
pkill -f "helixagent" 2>/dev/null || true
sleep 2

# 2. Stop all local containers
echo "[2/8] Stopping all local containers..."
podman stop -a 2>/dev/null || true
sleep 2

# 3. Remove all local helixagent containers
echo "[3/8] Removing local helixagent containers..."
podman rm -f helixagent-postgres helixagent-redis helixagent-chromadb 2>/dev/null || true
podman rm -f $(podman ps -aq --filter "name=helixagent" 2>/dev/null) 2>/dev/null || true

# 4. Clean up remote host
echo "[4/8] Cleaning up thinker.local..."
ssh milosvasic@thinker.local 'podman stop -a 2>/dev/null; podman rm -f $(podman ps -aq 2>/dev/null) 2>/dev/null; echo "Remote cleaned"' 2>&1 || echo "Warning: Could not clean remote"

# 5. Set environment variable explicitly
echo "[5/8] Setting remote distribution environment..."
export CONTAINERS_REMOTE_ENABLED=true
export CONTAINERS_REMOTE_HOST_1_NAME=thinker
export CONTAINERS_REMOTE_HOST_1_ADDRESS=thinker.local
export CONTAINERS_REMOTE_HOST_1_PORT=22
export CONTAINERS_REMOTE_HOST_1_USER=milosvasic
export CONTAINERS_REMOTE_HOST_1_RUNTIME=podman

# 6. Rebuild HelixAgent binary with latest code
echo "[6/8] Rebuilding HelixAgent..."
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
make build > /tmp/helixagent-build.log 2>&1 || {
    echo "Build failed, using existing binary"
}

# 7. Start HelixAgent with remote distribution
echo "[7/8] Starting HelixAgent with remote distribution..."
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
./bin/helixagent > /tmp/helixagent-distributed.log 2>&1 &
HELIXAGENT_PID=$!
echo $HELIXAGENT_PID > /tmp/helixagent.pid
echo "✓ HelixAgent started (PID: $HELIXAGENT_PID)"

# 8. Wait and verify distribution
echo "[8/8] Waiting for remote distribution (this may take 2-3 minutes)..."
sleep 5

MAX_WAIT=180
WAITED=0
while [ $WAITED -lt $MAX_WAIT ]; do
    # Check if containers are running on remote
    REMOTE_COUNT=$(ssh milosvasic@thinker.local 'podman ps -q 2>/dev/null | wc -l' 2>&1 || echo "0")
    
    if [ "$REMOTE_COUNT" -gt 0 ]; then
        echo ""
        echo "✅ SUCCESS! Containers distributed to thinker.local:"
        ssh milosvasic@thinker.local 'podman ps --format "table {{.Names}}\t{{.Status}}"' 2>&1
        echo ""
        echo "========================================"
        echo "REMOTE DISTRIBUTION COMPLETE!"
        echo "========================================"
        echo "Check status: ssh milosvasic@thinker.local 'podman ps'"
        echo "Logs: tail -f /tmp/helixagent-distributed.log"
        exit 0
    fi
    
    sleep 3
    WAITED=$((WAITED + 3))
    
    # Show progress every 30 seconds
    if [ $((WAITED % 30)) -eq 0 ]; then
        echo "  Still waiting... ($WAITED/$MAX_WAIT seconds)"
        echo "  Remote containers: $REMOTE_COUNT"
    fi
done

echo ""
echo "========================================"
echo "⚠️  TIMEOUT: Distribution may be in progress"
echo "========================================"
echo "Remote containers: $REMOTE_COUNT"
echo "Check logs: tail -100 /tmp/helixagent-distributed.log"
echo ""
echo "Current local containers:"
podman ps --format "table {{.Names}}\t{{.Status}}" | grep helixagent || echo "  (none)"
