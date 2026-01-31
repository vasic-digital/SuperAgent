#!/bin/bash
set -e

cd "$(dirname "$0")"

# Kill any existing helixagent
pkill -f helixagent 2>/dev/null || true
sleep 1

# Set environment variables for standalone mode
export DB_HOST=""
export REDIS_HOST=""
export COGNEE_ENABLED=false
export PORT=7061

# Start server in background
./bin/helixagent > /tmp/helixagent.log 2>&1 &
SERVER_PID=$!
echo "Started HelixAgent with PID $SERVER_PID"

# Wait for server to be ready (max 30 seconds)
for i in {1..30}; do
    if curl -s -f http://localhost:7061/v1/health > /dev/null 2>&1; then
        echo "Server is ready"
        break
    fi
    sleep 1
done

# Test skills endpoints
echo "Testing /v1/skills"
curl -s -f http://localhost:7061/v1/skills > /dev/null
echo "  OK - endpoint exists"

echo "Testing /v1/skills/categories"
curl -s -f http://localhost:7061/v1/skills/categories > /dev/null
echo "  OK - endpoint exists"

echo "Testing /v1/skills/match (POST)"
curl -s -f -X POST http://localhost:7061/v1/skills/match \
    -H "Content-Type: application/json" \
    -d '{"input": "hello"}' > /dev/null
echo "  OK - endpoint exists"

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
echo "Skills endpoints test passed"