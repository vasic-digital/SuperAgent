#!/bin/bash
set -e

# Load environment variables
# .env contains API keys, .env.remote contains remote service config
export $(cat .env | grep -v '^#' | xargs) 2>/dev/null || true
export $(cat .env.remote | grep -v '^#' | xargs) 2>/dev/null || true

# Ensure binary exists
if [[ ! -f ./bin/helixagent ]]; then
    echo "Building helixagent..."
    make build
fi

# Start helixagent in background
echo "Starting helixagent with remote configuration..."
./bin/helixagent &
HELIX_PID=$!

# Wait for server to start
sleep 5

# Check health endpoint
if curl -f http://localhost:7061/health >/dev/null 2>&1; then
    echo "HelixAgent started successfully (PID: $HELIX_PID)"
    echo "Health check passed"
else
    echo "HelixAgent failed to start"
    kill $HELIX_PID 2>/dev/null
    exit 1
fi

# Wait for user interruption
trap "kill $HELIX_PID; exit 0" INT TERM
wait $HELIX_PID