# MCP Servers for HelixAgent

This directory contains Model Context Protocol (MCP) server implementations used by HelixAgent.
The servers are included as git submodules from the official MCP repositories.

## Overview

MCP (Model Context Protocol) servers provide tools and capabilities that can be used by AI assistants.
HelixAgent integrates these servers to provide comprehensive functionality to CLI agents like OpenCode.

**Important**: MCP servers communicate via **stdio** (Standard Input/Output) using JSON-RPC, not HTTP.
They are designed to be launched by MCP clients (like OpenCode) on demand, not run as persistent services.
The container infrastructure provided here builds and installs all dependencies, making the servers ready
for use by MCP clients.

## Git Submodules

### Active Servers
**Repository**: [modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)

These servers are actively maintained by the MCP steering group:

| Server | Port | Description |
|--------|------|-------------|
| **fetch** | 3001 | HTTP fetch operations for retrieving web content |
| **filesystem** | 3002 | File system access for reading/writing files |
| **git** | 3003 | Git repository operations (status, commit, push, etc.) |
| **memory** | 3004 | Persistent memory/notes storage and retrieval |
| **time** | 3005 | Time and timezone operations |
| **sequential-thinking** | 3006 | Step-by-step reasoning and analysis |
| **everything** | 3007 | Local file search using Everything search engine |

### Archived Servers
**Repository**: [modelcontextprotocol/servers-archived](https://github.com/modelcontextprotocol/servers-archived)

These servers are from the archived repository (no longer actively maintained but fully functional):

| Server | Port | Required Environment | Description |
|--------|------|---------------------|-------------|
| **postgres** | 3008 | `POSTGRES_URL` | PostgreSQL database operations |
| **sqlite** | 3009 | - | SQLite database operations |
| **slack** | 3010 | `SLACK_BOT_TOKEN`, `SLACK_TEAM_ID` | Slack messaging integration |
| **github** | 3011 | `GITHUB_TOKEN` | GitHub API operations |
| **gitlab** | 3012 | `GITLAB_TOKEN` | GitLab API operations |
| **google-maps** | 3013 | `GOOGLE_MAPS_API_KEY` | Google Maps API integration |
| **brave-search** | 3014 | `BRAVE_API_KEY` | Brave Search API integration |
| **puppeteer** | 3015 | - | Browser automation with Puppeteer |
| **redis** | 3016 | `REDIS_URL` | Redis cache operations |
| **sentry** | 3017 | `SENTRY_AUTH_TOKEN`, `SENTRY_ORG` | Sentry error tracking |
| **gdrive** | 3018 | `GOOGLE_CREDENTIALS_PATH` | Google Drive operations |
| **everart** | 3019 | `EVERART_API_KEY` | Everart image generation |
| **aws-kb-retrieval** | 3020 | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | AWS Knowledge Base retrieval |

## Installation

### Initialize Submodules

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/your-org/HelixAgent.git

# Or initialize submodules after clone
git submodule update --init --recursive
```

### Using Docker/Podman

```bash
# Build the MCP servers container
cd external/mcp-servers
docker build -t helixagent-mcp-servers .

# Run with docker-compose
docker-compose up -d
```

### Using Podman

```bash
# Build
podman build -t helixagent-mcp-servers external/mcp-servers

# Run
podman-compose -f external/mcp-servers/docker-compose.yml up -d
```

## Configuration

### Environment Variables

Create a `.env` file in the `external/mcp-servers` directory:

```env
# GitHub
GITHUB_TOKEN=ghp_your_token_here

# GitLab
GITLAB_TOKEN=glpat_your_token_here

# Slack
SLACK_BOT_TOKEN=xoxb-your-token
SLACK_TEAM_ID=T12345678

# Brave Search
BRAVE_API_KEY=your_brave_api_key

# Google Maps
GOOGLE_MAPS_API_KEY=your_google_maps_key

# Google Drive
GOOGLE_CREDENTIALS_PATH=/path/to/credentials.json

# Sentry
SENTRY_AUTH_TOKEN=your_sentry_token
SENTRY_ORG=your-org

# Everart
EVERART_API_KEY=your_everart_key

# AWS
AWS_ACCESS_KEY_ID=your_aws_key
AWS_SECRET_ACCESS_KEY=your_aws_secret

# Database connections
POSTGRES_URL=postgresql://user:pass@host:5432/db
REDIS_URL=redis://:password@host:6379
```

### OpenCode Integration

The MCP servers are automatically configured in OpenCode when you generate the configuration:

```bash
# Generate OpenCode configuration
LOCAL_ENDPOINT=http://localhost:7061 ./bin/helixagent --generate-opencode-config > ~/.config/opencode/opencode.json
```

## Health Check

Check the health of all MCP servers:

```bash
# If running in container
docker exec helixagent-mcp-servers /app/scripts/health-check.sh

# Or run the challenge
./challenges/scripts/external_mcp_servers_challenge.sh
```

## Testing

Run the integration tests:

```bash
# Run MCP server tests
go test -v ./tests/integration/... -run "MCP"

# Run the full MCP challenge
./challenges/scripts/mcp_servers_challenge.sh
./challenges/scripts/external_mcp_servers_challenge.sh
```

## Architecture

```
external/mcp-servers/
├── servers/                    # Active MCP servers (git submodule)
│   └── src/
│       ├── fetch/              # Python (mcp_server_fetch)
│       ├── filesystem/         # Node.js
│       ├── git/                # Python (mcp_server_git)
│       ├── memory/             # Node.js
│       ├── time/               # Python (mcp_server_time)
│       ├── sequentialthinking/ # Node.js
│       └── everything/         # Node.js
├── servers-archived/           # Archived MCP servers (git submodule)
│   └── src/
│       ├── postgres/           # Node.js
│       ├── sqlite/             # Python (mcp_server_sqlite)
│       ├── slack/              # Node.js
│       ├── github/             # Node.js
│       ├── gitlab/             # Node.js
│       ├── google-maps/        # Node.js
│       ├── brave-search/       # Node.js
│       ├── puppeteer/          # Node.js
│       ├── redis/              # Node.js
│       ├── sentry/             # Python (mcp_server_sentry)
│       ├── gdrive/             # Node.js
│       ├── everart/            # Node.js
│       └── aws-kb-retrieval-server/  # Node.js
├── scripts/
│   ├── build.sh               # Build script with network handling
│   ├── start-all.sh           # Startup script for all servers
│   └── health-check.sh        # Health check script
├── Dockerfile                  # Container build configuration
├── docker-compose.yml          # Container orchestration
└── README.md                   # This file
```

**Note**: The MCP servers are a mix of Node.js and Python implementations.
The Dockerfile installs both runtimes and the startup scripts handle both types.

## Updating Submodules

To update the MCP servers to the latest version:

```bash
# Update all submodules
git submodule update --remote

# Or update specific submodule
git submodule update --remote external/mcp-servers/servers
```

## Troubleshooting

### Container Build Fails with Alpine Package Errors

**Symptom**: Build fails with errors like:
```
WARNING: fetching https://dl-cdn.alpinelinux.org/alpine/v3.23/main/x86_64/APKINDEX.tar.gz: I/O error
ERROR: unable to select packages: git (no such package)
```

**Cause**: Container DNS resolution issues, especially common with Podman.

**Solution**: Use the build script which includes `--network=host`:
```bash
# Use the build script (recommended)
./scripts/build.sh

# Or manually with --network=host
podman build --network=host -t helixagent-mcp-servers:latest .
docker build --network=host -t helixagent-mcp-servers:latest .
```

### Container DNS Resolution Issues

**Symptom**: Container can't reach external URLs during build or runtime.

**Diagnosis**:
```bash
# Test DNS inside container with default network
podman run --rm alpine:latest sh -c "apk update"

# Test with host network (should work)
podman run --rm --network=host alpine:latest sh -c "apk update"
```

**Solution**: The Dockerfile and docker-compose.yml are configured to use host network for builds.
If running manually, always use `--network=host` for the build step.

### Scripts Fail with "not found" Error

**Symptom**: Container starts but immediately exits with:
```
/usr/local/bin/docker-entrypoint.sh: exec: line 11: /app/scripts/start-all.sh: not found
```

**Cause**: Scripts use `#!/bin/bash` but Alpine Linux doesn't have bash by default.

**Solution**: All scripts in this project use `#!/bin/sh` for Alpine compatibility.
If you've modified scripts, ensure they use `/bin/sh` not `/bin/bash`.

### Server not starting

1. Check if the required environment variables are set
2. Check the logs: `docker logs helixagent-mcp-servers`
3. Verify the container is running: `docker ps`

### Connection refused

1. Ensure the MCP servers container is running
2. Check that the port is not blocked by firewall
3. Verify the correct host/port in your configuration

### Authentication errors

1. Verify API keys are correct and not expired
2. Check that tokens have the required permissions
3. Ensure environment variables are properly passed to the container

### Network Pre-flight Checks

Before building, ensure these URLs are reachable from your host:
```bash
# Alpine package repository
curl -I https://dl-cdn.alpinelinux.org/alpine/v3.23/main/x86_64/APKINDEX.tar.gz

# npm registry
curl -I https://registry.npmjs.org/

# PyPI (for Python servers)
curl -I https://pypi.org/simple/
```

## Contributing

1. Fork the repository
2. Make your changes
3. Run the tests: `./challenges/scripts/external_mcp_servers_challenge.sh`
4. Submit a pull request

## License

The MCP servers are licensed under their respective licenses:
- Active servers: MIT License (modelcontextprotocol/servers)
- Archived servers: MIT License (modelcontextprotocol/servers-archived)
