# MCP Servers Reference

## Overview

HelixAgent provides 45+ MCP (Model Context Protocol) servers that CLI agents can use to extend their capabilities.

## MCP Server Categories

### 1. HelixAgent Core MCPs (Remote)

| Server | URL | Description | Tools |
|--------|-----|-------------|-------|
| **helixagent-mcp** | `http://localhost:7061/v1/mcp` | Core MCP server | 20+ tools |
| **helixagent-lsp** | `http://localhost:7061/v1/lsp` | Language Server Protocol | completions, diagnostics |
| **helixagent-acp** | `http://localhost:7061/v1/acp` | Agent Communication Protocol | agent-to-agent messaging |
| **helixagent-embeddings** | `http://localhost:7061/v1/embeddings` | Embedding generation | text embeddings, similarity |
| **helixagent-vision** | `http://localhost:7061/v1/vision` | Vision/image analysis | image understanding, OCR |
| **helixagent-rag** | `http://localhost:7061/v1/rag` | RAG retrieval | document search, context |
| **helixagent-cognee** | `http://localhost:7061/v1/cognee` | Memory/knowledge graph | persistent memory |
| **helixagent-formatters** | `http://localhost:7061/v1/formatters` | Code formatting | 32+ formatters |
| **helixagent-monitoring** | `http://localhost:7061/v1/monitoring` | Metrics and monitoring | Prometheus metrics |

### 2. File System MCPs

| Server | Type | Command | Description |
|--------|------|---------|-------------|
| **filesystem** | local | `npx -y @modelcontextprotocol/server-filesystem .` | File operations |

**Tools:**
- `read_file` - Read file contents
- `write_file` - Write file contents
- `list_directory` - List directory contents
- `create_directory` - Create new directory
- `move_file` - Move/rename files
- `search_files` - Search file contents
- `get_file_info` - Get file metadata

### 3. Web/Browser MCPs

| Server | Type | Command | Description |
|--------|------|---------|-------------|
| **fetch** | local | `npx -y mcp-fetch-server` | HTTP requests |
| **puppeteer** | local | `npx -y @modelcontextprotocol/server-puppeteer` | Browser automation |

**Tools:**
- `fetch` - HTTP GET/POST requests
- `browser_navigate` - Navigate to URL
- `browser_screenshot` - Take screenshot
- `browser_click` - Click element
- `browser_type` - Type text
- `browser_evaluate` - Execute JavaScript

### 4. Database MCPs

| Server | Type | Command | Description |
|--------|------|---------|-------------|
| **sqlite** | local | `npx -y mcp-server-sqlite-npx <db>` | SQLite database |
| **postgres-mcp** | remote | Configurable | PostgreSQL operations |

**Tools:**
- `query` - Execute SQL query
- `execute` - Execute SQL command
- `schema` - Get database schema

### 5. Memory MCPs

| Server | Type | Command | Description |
|--------|------|---------|-------------|
| **memory** | local | `npx -y @modelcontextprotocol/server-memory` | Session memory |
| **sequential-thinking** | local | `npx -y @modelcontextprotocol/server-sequential-thinking` | Chain-of-thought |

**Tools:**
- `remember` - Store information
- `recall` - Retrieve information
- `forget` - Remove information
- `think` - Chain-of-thought reasoning

### 6. External Service MCPs

| Server | Type | URL | Description |
|--------|------|-----|-------------|
| **cloudflare-docs** | remote | `https://docs.mcp.cloudflare.com/sse` | Cloudflare documentation |
| **context7** | remote | `https://mcp.context7.com/mcp` | Context7 knowledge base |
| **deepwiki** | remote | `https://mcp.deepwiki.com/mcp` | DeepWiki search |

### 7. Git MCPs

| Server | Type | Command | Description |
|--------|------|---------|-------------|
| **git** | local | `npx -y mcp-git` | Git operations |

**Tools:**
- `git_status` - Check repository status
- `git_log` - View commit history
- `git_diff` - Show changes
- `git_branch` - Branch operations
- `git_commit` - Create commits

## Configuration Examples

### Basic MCP Configuration

```json
{
  "mcp": {
    "helixagent-mcp": {
      "type": "remote",
      "url": "http://localhost:7061/v1/mcp",
      "enabled": true
    },
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem", "."],
      "enabled": true
    }
  }
}
```

### Advanced MCP Configuration

```json
{
  "mcp": {
    "helixagent-core": {
      "type": "remote",
      "url": "http://localhost:7061/v1/mcp",
      "headers": {
        "Authorization": "Bearer ${HELIXAGENT_API_KEY}"
      }
    },
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem", "/home/user/projects"],
      "env": {
        "PATH": "/usr/local/bin"
      }
    },
    "browser": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-puppeteer"],
      "timeout": 60000
    }
  }
}
```

## MCP Tool Protocol

### Tool Call Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "read_file",
    "arguments": {
      "path": "/path/to/file.txt"
    }
  }
}
```

### Tool Call Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "file contents here"
      }
    ]
  }
}
```

## Security Best Practices

1. **Local MCPs**: Run in isolated environment when possible
2. **Remote MCPs**: Use HTTPS and authentication
3. **File System**: Limit to specific directories
4. **Browser**: Use in sandboxed environment
5. **Credentials**: Never commit API keys

## Troubleshooting

### Issue: MCP Server Not Starting
**Solution:**
```bash
# Check npx is installed
npm install -g npx

# Verify MCP package exists
npx @modelcontextprotocol/server-filesystem --help
```

### Issue: Connection Refused
**Solution:**
```bash
# Verify HelixAgent is running
curl http://localhost:7061/v1/health

# Check firewall settings
sudo ufw allow 7061
```

### Issue: Permission Denied
**Solution:**
```bash
# Check file permissions
ls -la /path/to/mcp

# Run with appropriate user
sudo chown -R $USER:$USER ~/.mcp
```

---

**Last Updated:** 2026-04-02
