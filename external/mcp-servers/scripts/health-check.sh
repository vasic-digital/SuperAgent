#!/bin/bash
# Health check script for MCP servers

HEALTHY=true

# Define expected servers and their ports
declare -A SERVERS=(
    ["fetch"]=3001
    ["filesystem"]=3002
    ["git"]=3003
    ["memory"]=3004
    ["time"]=3005
    ["sequential-thinking"]=3006
    ["everything"]=3007
    ["postgres"]=3008
    ["sqlite"]=3009
    ["slack"]=3010
    ["github"]=3011
    ["gitlab"]=3012
    ["google-maps"]=3013
    ["brave-search"]=3014
    ["puppeteer"]=3015
    ["redis"]=3016
    ["sentry"]=3017
    ["gdrive"]=3018
    ["everart"]=3019
    ["aws-kb-retrieval"]=3020
)

echo "MCP Servers Health Check"
echo "========================"

for name in "${!SERVERS[@]}"; do
    port=${SERVERS[$name]}
    pid_file="/var/run/mcp-$name.pid"

    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            echo "[OK] $name (port $port, PID $pid)"
        else
            echo "[FAIL] $name - process not running"
            HEALTHY=false
        fi
    else
        echo "[FAIL] $name - PID file not found"
        HEALTHY=false
    fi
done

echo ""
if [ "$HEALTHY" = true ]; then
    echo "All servers healthy"
    exit 0
else
    echo "Some servers are unhealthy"
    exit 1
fi
