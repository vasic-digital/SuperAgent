# MCP Server Validation Guide

This document provides comprehensive information about MCP (Model Context Protocol) server validation in HelixAgent.

## Overview

HelixAgent includes an MCP validation system that:
- Automatically validates MCP server requirements
- Only enables MCPs that have all required dependencies
- Generates reports about MCP status
- Prevents configuration of non-working MCPs

## MCP Categories

### 1. Core MCPs (Always Work)

These MCPs require no external API keys and work out of the box:

| MCP | Package | Description |
|-----|---------|-------------|
| `filesystem` | `@modelcontextprotocol/server-filesystem` | File system access - read, write, list files |
| `fetch` | `mcp-fetch-server` | HTTP fetch - make web requests |
| `memory` | `@modelcontextprotocol/server-memory` | In-memory key-value storage |
| `time` | `@theo.foobar/mcp-time` | Time and timezone utilities |
| `git` | `mcp-git` | Git repository operations |
| `sequential-thinking` | `@modelcontextprotocol/server-sequential-thinking` | Sequential thinking and reasoning |
| `everything` | `@anthropic-ai/mcp-server-everything` | Test MCP server with all features |

### 2. Database MCPs (Local Services)

These MCPs require local database services:

| MCP | Package | Requirements |
|-----|---------|--------------|
| `sqlite` | `@modelcontextprotocol/server-sqlite` | None (creates local DB) |
| `postgres` | `@modelcontextprotocol/server-postgres` | PostgreSQL running on port 15432 |
| `redis` | `mcp-server-redis` | Redis running on port 16379 |
| `mongodb` | `mcp-server-mongodb` | MongoDB running locally |

### 3. DevOps MCPs (Local Tools)

| MCP | Package | Requirements |
|-----|---------|--------------|
| `docker` | `mcp-server-docker` | Docker or Podman installed |
| `puppeteer` | `@modelcontextprotocol/server-puppeteer` | Chrome/Chromium available |
| `kubernetes` | `mcp-server-kubernetes` | kubectl configured |

### 4. API-Required MCPs

These MCPs require external API keys:

| MCP | Package | Required Environment Variables |
|-----|---------|-------------------------------|
| `github` | `@modelcontextprotocol/server-github` | `GITHUB_TOKEN` |
| `gitlab` | `@modelcontextprotocol/server-gitlab` | `GITLAB_TOKEN` |
| `slack` | `@modelcontextprotocol/server-slack` | `SLACK_BOT_TOKEN`, `SLACK_TEAM_ID` |
| `discord` | `mcp-server-discord` | `DISCORD_TOKEN` |
| `notion` | `@notionhq/notion-mcp-server` | `NOTION_API_KEY` |
| `linear` | `@modelcontextprotocol/server-linear` | `LINEAR_API_KEY` |
| `brave-search` | `@modelcontextprotocol/server-brave-search` | `BRAVE_API_KEY` |
| `sentry` | `@modelcontextprotocol/server-sentry` | `SENTRY_AUTH_TOKEN`, `SENTRY_ORG` |
| `cloudflare` | `@cloudflare/mcp-server-cloudflare` | `CLOUDFLARE_API_TOKEN` |
| `exa` | `exa-mcp-server` | `EXA_API_KEY` |

## Setting Up API Keys

### 1. Create/Edit `.env` file

```bash
# In your project root
cp .env.example .env
nano .env
```

### 2. Add Required Keys

```bash
# GitHub (required for github MCP)
GITHUB_TOKEN=ghp_your_github_token

# Slack (required for slack MCP)
SLACK_BOT_TOKEN=xoxb-your-slack-token
SLACK_TEAM_ID=T12345678

# Notion (required for notion MCP)
NOTION_API_KEY=secret_your_notion_key

# Brave Search (required for brave-search MCP)
BRAVE_API_KEY=your_brave_api_key

# HelixAgent (required for helixagent MCP)
HELIXAGENT_API_KEY=your_helixagent_key
```

### 3. Regenerate Configuration

```bash
./bin/helixagent --generate-opencode-config --opencode-output ~/.config/opencode/.opencode.json
./bin/helixagent --generate-crush-config --crush-output ~/.config/crush/crush.json
```

## Validation Process

### How Validation Works

1. **Environment Check**: Validator reads `.env` and environment variables
2. **Requirement Check**: For each MCP, checks if required env vars are set
3. **Service Check**: Checks if required local services are running
4. **Enable/Disable**: Only MCPs with all requirements met are enabled

### Running Validation

```bash
# Run validation challenge
./challenges/scripts/mcp_validation_challenge.sh

# Or use the Go validator directly
go test -v ./internal/mcp/validation/...
```

### Validation Report Format

```
================================================================================
MCP VALIDATION REPORT
================================================================================

Summary: 24 total, 13 working, 11 disabled, 0 failed

WORKING MCPs (Enabled):
--------------------------------------------------------------------------------
  ✓ helixagent               [helixagent] HelixAgent MCP plugin
  ✓ filesystem               [core] File system access
  ✓ fetch                    [core] HTTP fetch
  ✓ memory                   [core] In-memory storage
  ...

DISABLED MCPs (Missing Requirements):
--------------------------------------------------------------------------------
  ✗ slack                    [communication] Missing: SLACK_BOT_TOKEN, SLACK_TEAM_ID
  ✗ notion                   [productivity] Missing: NOTION_API_KEY
  ✗ brave-search             [search] Missing: BRAVE_API_KEY
  ...
```

## HelixAgent MCP Plugin

### Installation

The HelixAgent MCP plugin is automatically installed to:
```
~/.helixagent/plugins/mcp-server/dist/index.js
```

### Manual Installation

```bash
# Build the plugin
cd plugins/mcp-server
npm install
npm run build

# Install to user directory
mkdir -p ~/.helixagent/plugins/mcp-server/dist
cp -r dist/* ~/.helixagent/plugins/mcp-server/dist/
```

### Testing

```bash
# Test the plugin
node ~/.helixagent/plugins/mcp-server/dist/index.js --help
```

## Troubleshooting

### MCP Connection Errors

If you see "Connection closed" errors:

1. **Check if MCP requires API key**: Look at the table above
2. **Verify environment variable is set**: `echo $VARIABLE_NAME`
3. **Regenerate configuration**: Only validated MCPs will be included

### Missing Dependencies

```bash
# Install Node.js (for npx)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Verify npx
npx --version
```

### Database MCPs Not Working

```bash
# Start PostgreSQL
podman start helixagent-postgres

# Start Redis
podman start helixagent-redis

# Or use docker-compose
docker-compose up -d postgres redis
```

## Best Practices

1. **Only enable MCPs you need**: Fewer MCPs = faster startup
2. **Keep API keys secure**: Use `.env` file, don't commit to git
3. **Run validation after changes**: Ensures config is correct
4. **Monitor MCP health**: Check OpenCode status regularly

## Related Documentation

- [MCP Configuration Requirements](MCP_CONFIGURATION_REQUIREMENTS.md)
- [Verified MCP Servers](VERIFIED_MCP_SERVERS.md)
- [CLI Agent Configuration](../cli-agents/README.md)
