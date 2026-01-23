#!/bin/bash
# Start all MCP servers for HelixAgent
# Each server runs on a dedicated port in SSE mode

set -e

LOG_DIR="/var/log/mcp-servers"
mkdir -p "$LOG_DIR"

echo "Starting MCP servers..."

# Function to start a server
start_server() {
    local name=$1
    local port=$2
    local dir=$3
    local script=${4:-"index.js"}

    echo "Starting $name on port $port..."
    cd "$dir"

    if [ -f "dist/$script" ]; then
        MCP_PORT=$port node "dist/$script" >> "$LOG_DIR/$name.log" 2>&1 &
    elif [ -f "$script" ]; then
        MCP_PORT=$port node "$script" >> "$LOG_DIR/$name.log" 2>&1 &
    else
        echo "Warning: Could not find entry point for $name"
        return 1
    fi

    echo "$!" > "/var/run/mcp-$name.pid"
    echo "$name started with PID $(cat /var/run/mcp-$name.pid)"
}

# Start active servers (from servers repo)
start_server "fetch" 3001 "/app/servers/src/fetch"
start_server "filesystem" 3002 "/app/servers/src/filesystem"
start_server "git" 3003 "/app/servers/src/git"
start_server "memory" 3004 "/app/servers/src/memory"
start_server "time" 3005 "/app/servers/src/time"
start_server "sequential-thinking" 3006 "/app/servers/src/sequentialthinking"
start_server "everything" 3007 "/app/servers/src/everything"

# Start archived servers (from servers-archived repo)
start_server "postgres" 3008 "/app/servers-archived/postgres"
start_server "sqlite" 3009 "/app/servers-archived/sqlite"
start_server "slack" 3010 "/app/servers-archived/slack"
start_server "github" 3011 "/app/servers-archived/github"
start_server "gitlab" 3012 "/app/servers-archived/gitlab"
start_server "google-maps" 3013 "/app/servers-archived/google-maps"
start_server "brave-search" 3014 "/app/servers-archived/brave-search"
start_server "puppeteer" 3015 "/app/servers-archived/puppeteer"
start_server "redis" 3016 "/app/servers-archived/redis"
start_server "sentry" 3017 "/app/servers-archived/sentry"
start_server "gdrive" 3018 "/app/servers-archived/gdrive"
start_server "everart" 3019 "/app/servers-archived/everart"
start_server "aws-kb-retrieval" 3020 "/app/servers-archived/aws-kb-retrieval-server"

echo ""
echo "All MCP servers started. Monitoring..."

# Keep container running and monitor servers
while true; do
    sleep 30

    # Check if servers are still running
    for pid_file in /var/run/mcp-*.pid; do
        if [ -f "$pid_file" ]; then
            name=$(basename "$pid_file" .pid | sed 's/mcp-//')
            pid=$(cat "$pid_file")
            if ! kill -0 "$pid" 2>/dev/null; then
                echo "Warning: $name (PID $pid) has stopped"
            fi
        fi
    done
done
