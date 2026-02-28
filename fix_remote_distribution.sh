#!/bin/bash
# Emergency Remote Distribution Fix Script
# Forces redistribution of all containers to thinker.local

set -e

echo "========================================"
echo "EMERGENCY: Remote Distribution Fix"
echo "========================================"

# 1. Stop local helixagent containers (except keep postgres/redis for now)
echo "[1/5] Stopping local HelixAgent containers..."
podman stop helixagent-chromadb 2>/dev/null || true
podman rm helixagent-chromadb 2>/dev/null || true

# 2. Verify remote host is ready
echo "[2/5] Verifying thinker.local is ready..."
if ! ssh milosvasic@thinker.local 'echo OK' >/dev/null 2>&1; then
    echo "ERROR: Cannot connect to thinker.local"
    exit 1
fi
echo "✓ thinker.local accessible"

# 3. Clean up any existing containers on remote
echo "[3/5] Cleaning up remote containers..."
ssh milosvasic@thinker.local 'podman stop -a 2>/dev/null; podman rm -a 2>/dev/null; echo "Remote containers cleaned"' || true

# 4. Start HelixAgent with remote distribution enabled
echo "[4/5] Starting HelixAgent with remote distribution..."
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
export CONTAINERS_REMOTE_ENABLED=true
./bin/helixagent > /tmp/helixagent-remote.log 2>&1 &
HELIXAGENT_PID=$!
echo $HELIXAGENT_PID > /tmp/helixagent.pid
echo "✓ HelixAgent started (PID: $HELIXAGENT_PID)"

# 5. Wait for startup and verify distribution
echo "[5/5] Waiting for remote distribution..."
sleep 10

MAX_WAIT=60
WAITED=0
while [ $WAITED -lt $MAX_WAIT ]; do
    REMOTE_CONTAINERS=$(ssh milosvasic@thinker.local 'podman ps --format "{{.Names}}" 2>/dev/null | wc -l')
    if [ "$REMOTE_CONTAINERS" -gt 1 ]; then
        echo "✓ SUCCESS: $REMOTE_CONTAINERS containers running on thinker.local"
        break
    fi
    sleep 2
    WAITED=$((WAITED + 2))
    echo "  Waiting... ($WAITED/$MAX_WAIT seconds)"
done

if [ $WAITED -ge $MAX_WAIT ]; then
    echo "✗ TIMEOUT: No containers distributed to thinker.local"
    echo "Check logs: tail -50 /tmp/helixagent-remote.log"
    exit 1
fi

echo ""
echo "========================================"
echo "Remote Distribution Complete!"
echo "========================================"
echo "Check status: ssh milosvasic@thinker.local 'podman ps'"
echo "Logs: tail -f /tmp/helixagent-remote.log"
