# MCP Systems Documentation

This directory contains comprehensive documentation for HelixAgent's Model Context Protocol (MCP) infrastructure, including adapter registry, server containerization, configuration, and validation guides.

## Overview

HelixAgent provides 45+ MCP adapters for integrating with external services, with 65+ containerized MCP servers eliminating all npm/npx dependencies. The MCP infrastructure enables AI agents to interact with productivity tools, databases, cloud services, and more.

## Documentation Index

### Core Documents

| Document | Description |
|----------|-------------|
| [adapters-registry.md](adapters-registry.md) | Complete catalog of 45+ MCP adapters with configuration and tool listings |
| [CONTAINERIZATION.md](CONTAINERIZATION.md) | Containerized MCP server infrastructure with port allocation scheme |
| [MCP_CONFIGURATION_REQUIREMENTS.md](MCP_CONFIGURATION_REQUIREMENTS.md) | Prerequisites and configuration for all 43+ MCPs |
| [VERIFIED_MCP_SERVERS.md](VERIFIED_MCP_SERVERS.md) | 92 verified MCP servers from Git submodules and npm packages |
| [MCP_VALIDATION_GUIDE.md](MCP_VALIDATION_GUIDE.md) | Validation system, API key setup, and testing procedures |
| [MCP_STATUS_REPORT.md](MCP_STATUS_REPORT.md) | Current status of 72 configured MCPs with working/setup requirements |
| [MCP_SERVERS_COMPREHENSIVE.md](MCP_SERVERS_COMPREHENSIVE.md) | Complete server documentation with quick start guide |

## Adapter Registry Status

### By Category

| Category | Count | Description |
|----------|-------|-------------|
| Productivity | 8 | Linear, Asana, Jira, Notion, Trello, Todoist, Monday, Airtable |
| Communication | 5 | Slack, Discord, Telegram, Microsoft Teams, Email |
| Development | 8 | GitHub, GitLab, Sentry, Heroku, Vercel, Bitbucket |
| Data | 10 | MongoDB, Redis, PostgreSQL, Elasticsearch, Qdrant, Pinecone |
| Cloud | 6 | AWS, GCP, Azure, Cloudflare, Kubernetes, Docker |
| Specialized | 8+ | Figma, Obsidian, YouTube, Google services |

### By Status

| Status | Count | Description |
|--------|-------|-------------|
| No Config Required | 14 | Works immediately (filesystem, memory, git, time, etc.) |
| Local Service Required | 6 | Requires local database/service |
| API Key Required | 14+ | Requires external API credentials |
| HelixAgent Remote | 9 | Requires HelixAgent running for protocol tools |

## Server Containerization

### Architecture

All MCP servers run in Docker/Podman containers, eliminating npm/npx dependencies:

```
+-------------------------------------------------------------------+
|                       CLI Agents                                    |
|              (OpenCode, Crush, ClaudeCode, etc.)                   |
+-------------------------------------------------------------------+
                              | HTTP/SSE
                              v
+-------------------------------------------------------------------+
|                        HelixAgent                                   |
|  +---------------------------------------------------------------+  |
|  |              ContainerMCPConfigGenerator                       |  |
|  |  - Generates URLs for all 65 containerized MCPs               |  |
|  |  - Zero npx commands                                          |  |
|  |  - Environment-based enable/disable                           |  |
|  +---------------------------------------------------------------+  |
+-------------------------------------------------------------------+
                              | HTTP/SSE (ports 9101-9999)
                              v
+-------------------------------------------------------------------+
|              Docker/Podman Container Network                        |
|                  (helixagent-mcp-network)                          |
|  +----------+  +----------+  +----------+                          |
|  | mcp-fetch|  | mcp-git  |  | mcp-time |  ... 65+ containers     |
|  |  :9101   |  |  :9102   |  |  :9103   |                          |
|  +----------+  +----------+  +----------+                          |
+-------------------------------------------------------------------+
```

### Port Allocation Scheme

| Port Range | Category | Servers |
|------------|----------|---------|
| 9101-9120 | Core | fetch, git, time, filesystem, memory, everything, sequential-thinking |
| 9201-9220 | Database | MongoDB, Redis, MySQL, Elasticsearch, Supabase |
| 9301-9320 | Vector | Qdrant, Chroma, Pinecone, Weaviate |
| 9401-9440 | DevOps | GitHub, GitLab, Sentry, Kubernetes, Docker, AWS, GCP |
| 9501-9520 | Browser | Playwright, Browserbase, Firecrawl, Crawl4AI |
| 9601-9620 | Communication | Slack, Discord, Telegram |
| 9701-9740 | Productivity | Notion, Linear, Jira, Asana, Trello |
| 9801-9840 | Search/AI | Brave Search, Exa, Tavily, Perplexity |
| 9901-9920 | Google | Google Drive, Calendar, Maps, YouTube |
| 9921-9960 | Monitoring/Finance | Datadog, Grafana, Stripe, HubSpot |
| 9961-9999 | Design | Figma |

## Configuration Overview

### Quick Start

```bash
# Build core MCP server images
./scripts/mcp/build-core-mcp-images.sh

# Start core servers
podman-compose -f docker/mcp/docker-compose.mcp-core.yml up -d

# Start all MCP backend services
./scripts/mcp/start-mcp-services.sh --all

# Validate servers
./challenges/scripts/mcp_validation_comprehensive.sh --quick
```

### Environment Variables

Required API keys are configured via environment variables:

```bash
# Development Platforms
GITHUB_TOKEN=ghp_...
GITLAB_TOKEN=glpat_...

# Communication
SLACK_BOT_TOKEN=xoxb-...
DISCORD_TOKEN=...

# Productivity
NOTION_API_KEY=secret_...
LINEAR_API_KEY=lin_...
JIRA_API_TOKEN=...

# Search
BRAVE_API_KEY=...
EXA_API_KEY=...
```

### MCPs Without Configuration

These MCPs work immediately:

| MCP | Description |
|-----|-------------|
| filesystem | File system operations |
| memory | In-memory key-value storage |
| sequential-thinking | Chain-of-thought reasoning |
| everything | Demo server with all capabilities |
| puppeteer | Browser automation |
| docker | Container management |
| kubernetes | K8s cluster management |
| git | Repository operations |
| time | Timezone utilities |
| sqlite | Local database |

## Key Files

| File | Purpose |
|------|---------|
| `internal/mcp/adapters/*.go` | 45+ adapter implementations |
| `internal/mcp/config/generator_container.go` | Container config generator |
| `docker/mcp/docker-compose.mcp-full.yml` | Full MCP compose file |
| `docker/mcp/docker-compose.mcp-core.yml` | Core MCP servers |
| `.env.example` | Environment variable template |

## Validation

### Automatic Validation

The MCP validation system:
- Checks server requirements
- Only enables MCPs with all dependencies
- Generates status reports
- Prevents configuration of non-working MCPs

### Manual Validation

```bash
# Validate all MCPs
./challenges/scripts/mcp_validation_comprehensive.sh

# Quick validation
./challenges/scripts/mcp_validation_comprehensive.sh --quick

# Generate configuration
./bin/helixagent --generate-opencode-config
```

## Related Documentation

- [Architecture Overview](../architecture/README.md)
- [Protocol Support](../architecture/PROTOCOL_SUPPORT_DOCUMENTATION.md)
- [TOON Protocol](../architecture/toon-protocol.md)
- [Service Architecture](../architecture/SERVICE_ARCHITECTURE.md)

## Submodules

MCP servers are managed as Git submodules under `MCP/submodules/`:

| Category | Examples |
|----------|----------|
| Official | github-mcp-server, notion-mcp-server, sentry-mcp, heroku-mcp |
| Databases | redis-mcp, mongodb-mcp, qdrant-mcp, elasticsearch-mcp |
| Cloud | aws-mcp, kubernetes-mcp, cloudflare-mcp |
| Browser | playwright-mcp, browserbase-mcp |
| SDKs | python-sdk, typescript-sdk |

Update submodules:
```bash
git submodule update --remote MCP/submodules/*
```
