# MCP Servers Comprehensive Documentation

## Overview

HelixAgent supports **45+ MCP (Model Context Protocol) servers** across multiple categories, providing extensive tool capabilities for AI agents and LLM interactions.

## Quick Start

```bash
# Build core MCP server images
./scripts/mcp/build-core-mcp-images.sh

# Start core servers (fetch, git, time, filesystem, memory, everything, sequentialthinking)
podman-compose -f docker/mcp/docker-compose.mcp-core.yml up -d

# Validate servers
./challenges/scripts/mcp_validation_comprehensive.sh --quick
```

## Server Categories

### 1. Core Servers (from MCP-Servers repo)

| Server | Port | Type | Description |
|--------|------|------|-------------|
| fetch | 9101 | Python | HTTP fetch and web content retrieval |
| git | 9102 | Python | Git repository operations |
| time | 9103 | Python | Time and timezone utilities |
| filesystem | 9104 | TypeScript | File and directory operations |
| memory | 9105 | TypeScript | Knowledge graph storage (entities, relations) |
| everything | 9106 | TypeScript | Reference server with all MCP capabilities |
| sequentialthinking | 9107 | TypeScript | Step-by-step reasoning tools |

### 2. Database & Storage

| Server | Port | Type | API Key Required |
|--------|------|------|-----------------|
| mongodb | 9201 | TypeScript | No (connection string) |
| redis | 9202 | Python | No (connection string) |
| qdrant | 9301 | Python | Optional |
| supabase | 9302 | TypeScript | Yes |

### 3. DevOps & Infrastructure

| Server | Port | Type | API Key Required |
|--------|------|------|-----------------|
| github | 9203 | Go | Yes (GitHub PAT) |
| kubernetes | 9207 | Go | No (kubeconfig) |
| k8s | 9208 | Python | No (kubeconfig) |
| heroku | 9701 | TypeScript | Yes |
| cloudflare | 9702 | TypeScript | Yes |
| workers | 9703 | TypeScript | No |
| sentry | 9901 | TypeScript | Yes |

### 4. Productivity & Communication

| Server | Port | Type | API Key Required |
|--------|------|------|-----------------|
| slack | 9204 | Go | Yes (Bot token) |
| notion | 9205 | TypeScript | Yes |
| trello | 9206 | TypeScript (Bun) | Yes |
| telegram | 9501 | Python | Yes (Bot token) |
| airtable | 9601 | TypeScript | Yes |
| obsidian | 9602 | TypeScript | No (local API) |
| atlassian | 9303 | Python | Yes (Jira/Confluence) |

### 5. Browser & Web

| Server | Port | Type | API Key Required |
|--------|------|------|-----------------|
| browserbase | 9401 | TypeScript | Yes |
| firecrawl | 9402 | TypeScript | Yes |
| brave-search | 9403 | TypeScript | Yes |
| playwright | 9404 | TypeScript | No |

### 6. AI & Search

| Server | Port | Type | API Key Required |
|--------|------|------|-----------------|
| perplexity | 9801 | TypeScript | Yes |
| omnisearch | 9802 | TypeScript | Yes (multiple) |
| context7 | 9803 | TypeScript | Optional |
| llamaindex | 9804 | TypeScript | Optional |
| langchain | 9805 | Python | Optional |

## Port Allocation Scheme

```
9101-9199: Core MCP Servers (MCP-Servers repo)
9201-9299: Enterprise Integration (databases, VCS)
9301-9399: Data & Vector (Qdrant, Supabase)
9401-9499: Browser & Web (Browserbase, Playwright)
9501-9599: Communication (Telegram)
9601-9699: Productivity (Airtable, Obsidian)
9701-9799: Cloud & Deployment (Heroku, Cloudflare)
9801-9899: AI & Search (Perplexity, LLM tools)
9901-9999: DevOps & Monitoring (Sentry, Microsoft)
```

## Protocol Details

### Transport: NDJSON over TCP

MCP servers use **Newline-Delimited JSON (NDJSON)** over stdio, which we expose via TCP using socat:

```
[Client] --TCP--> [socat] --stdio--> [MCP Server]
```

**Critical Configuration**:
- Use `SYSTEM:` mode in socat, NOT `pty`
- PTY mode corrupts the NDJSON framing

### Session Handshake

MCP requires a proper session handshake:

```json
// 1. Client sends initialize
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{
  "protocolVersion":"2024-11-05",
  "capabilities":{},
  "clientInfo":{"name":"client","version":"1.0"}
}}

// 2. Server responds with capabilities
{"jsonrpc":"2.0","id":1,"result":{
  "protocolVersion":"2024-11-05",
  "capabilities":{"tools":{}},
  "serverInfo":{"name":"server","version":"1.0"}
}}

// 3. Client sends initialized notification
{"jsonrpc":"2.0","method":"notifications/initialized"}

// 4. Now client can call tools/list, tools/call, etc.
```

### Tool Discovery

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}

// Response
{"jsonrpc":"2.0","id":2,"result":{
  "tools":[
    {"name":"get_current_time","description":"Get current time","inputSchema":{...}}
  ]
}}
```

### Tool Execution

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{
  "name":"get_current_time",
  "arguments":{"timezone":"UTC"}
}}

// Response
{"jsonrpc":"2.0","id":3,"result":{
  "content":[{"type":"text","text":"2026-01-27T00:00:00Z"}]
}}
```

## Docker Configuration

### Building Images

**TypeScript Servers** (Node.js based):
```bash
podman build --network=host -t mcp-memory:latest \
  -f docker/mcp/Dockerfile.mcp-server \
  --build-arg SERVER_NAME=memory \
  --build-arg SOURCE_DIR=MCP-Servers \
  .
```

**Python Servers**:
```bash
podman build --network=host -t mcp-time:latest \
  -f docker/mcp/Dockerfile.mcp-server-python \
  --build-arg SERVER_NAME=time \
  --build-arg SOURCE_DIR=MCP-Servers \
  .
```

### Running Containers

```bash
# Single server
podman run -d --rm --name mcp-time -p 9103:9000 mcp-time:latest

# All core servers
podman-compose -f docker/mcp/docker-compose.mcp-core.yml up -d
```

## Testing & Validation

### Quick Validation (26 tests)

```bash
./challenges/scripts/mcp_validation_comprehensive.sh --quick
```

Tests:
1. **TCP Connectivity** (7 tests) - Port reachability
2. **Protocol Compliance** (7 tests) - JSON-RPC initialize
3. **Tool Discovery** (7 tests) - tools/list call
4. **Tool Execution** (5 tests) - Real tool calls

### Go Functional Tests

```bash
go test -v ./internal/testing/mcp/...
```

Tests:
- TestMCPTimeServerFunctional - Time tools
- TestMCPMemoryServerFunctional - Knowledge graph operations
- TestMCPFilesystemServerFunctional - Directory operations
- TestMCPFetchServerFunctional - HTTP fetch
- TestMCPGitServerFunctional - Git operations

### Full Validation (with LLM integration)

```bash
./challenges/scripts/mcp_validation_comprehensive.sh --full
```

Additional tests:
- LLM provider integration
- AI Debate system integration

## Integration with HelixAgent

### API Endpoint

```
POST /v1/chat/completions
{
  "model": "debate",
  "messages": [...],
  "tools": [
    {"type": "mcp", "mcp": {"server": "time"}}
  ]
}
```

### AI Debate Integration

MCP tools are available in debates:

```
POST /v1/debates
{
  "topic": "What time is it in different timezones?",
  "participants": [...],
  "mcp_servers": ["time", "memory"],
  "enable_mcp_tools": true
}
```

## Troubleshooting

### Server Not Responding

1. Check container logs: `podman logs mcp-servername`
2. Verify port binding: `podman port mcp-servername`
3. Test TCP: `timeout 2 bash -c "echo '' > /dev/tcp/localhost/PORT"`

### Protocol Errors

Common issues:
- **PTY mode in socat** - Use `SYSTEM:` not `pty`
- **Missing handshake** - Send initialize before tools/list
- **Wrong NDJSON framing** - Each message on one line, terminated by `\n`

### Python Server Issues

- **pydantic_core errors** - Use Python-specific Dockerfile
- **Virtual env not activating** - Use `. .venv/bin/activate` not `source`

## Environment Variables

### Core Servers (no API keys required)

```env
# Memory server knowledge graph path
MEMORY_DATA_PATH=/data/knowledge-graph

# Filesystem allowed directories
FS_ALLOWED_DIRS=/home/user,/tmp

# Fetch timeout
FETCH_TIMEOUT=30000
```

### External Service Servers

```env
# GitHub
GITHUB_TOKEN=ghp_xxxxxxxxxxxx

# Slack
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxx

# Notion
NOTION_API_KEY=secret_xxxxxxxxxxxx

# Telegram
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# Brave Search
BRAVE_API_KEY=BSAxxxxxxxxxxxx
```

## See Also

- [MCP Protocol Specification](https://modelcontextprotocol.io/specification)
- [MCP TypeScript SDK](../MCP/submodules/typescript-sdk/CLAUDE.md)
- [Docker Compose Files](../docker/mcp/)
- [Challenge Scripts](../challenges/scripts/)
