#!/bin/sh
# Health check script for MCP servers

HEALTHY=true

echo "MCP Servers Health Check"
echo "========================"

# Check each server by its PID file
check_server() {
    name=$1
    port=$2
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
}

# Active servers
check_server "fetch" 3001
check_server "filesystem" 3002
check_server "git" 3003
check_server "memory" 3004
check_server "time" 3005
check_server "sequential-thinking" 3006
check_server "everything" 3007

# Archived servers
check_server "postgres" 3008
check_server "sqlite" 3009
check_server "slack" 3010
check_server "github" 3011
check_server "gitlab" 3012
check_server "google-maps" 3013
check_server "brave-search" 3014
check_server "puppeteer" 3015
check_server "redis" 3016
check_server "sentry" 3017
check_server "gdrive" 3018
check_server "everart" 3019
check_server "aws-kb-retrieval" 3020

echo ""
if [ "$HEALTHY" = true ]; then
    echo "All servers healthy"
    exit 0
else
    echo "Some servers are unhealthy"
    exit 1
fi
