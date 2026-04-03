# MCP (Model Context Protocol) - Complete Integration Guide for HelixAgent

**Protocol Version:** 2024-11-05  
**HelixAgent Version:** 1.0.0+  
**LLMsVerifier Support:** Full  
**Status:** Production Ready

---

## Overview

HelixAgent implements the complete **Model Context Protocol (MCP)** specification from [modelcontextprotocol.io](https://modelcontextprotocol.io), enabling seamless integration with:

- **45+ MCP Servers** (filesystem, GitHub, databases, search, etc.)
- **Claude Code** (via MCP proxy)
- **All major AI agents** that support MCP
- **Custom MCP tools** via the HelixAgent SDK

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         HELIXAGENT MCP ARCHITECTURE                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                        HELIXAGENT CORE                                   │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────────┐  │   │
│  │  │   MCP       │  │   MCP       │  │   MCP       │  │    Claude      │  │   │
│  │  │   Client    │  │   Server    │  │   Bridge    │  │    Code API    │  │   │
│  │  │   (std/io)  │  │   (Host)    │  │   (HTTP)    │  │    (Proxy)     │  │   │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └────────────────┘  │   │
│  │         └─────────────────┴─────────────────┘                           │   │
│  │                           │                                             │   │
│  │                   ┌───────▼────────┐                                    │   │
│  │                   │  MCP Registry  │                                    │   │
│  │                   │  (45+ Servers) │                                    │   │
│  │                   └───────┬────────┘                                    │   │
│  └───────────────────────────┼─────────────────────────────────────────────┘   │
│                              │                                                   │
│  ┌───────────────────────────┼─────────────────────────────────────────────┐   │
│  │                   MCP SERVERS (External)                                 │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │   │
│  │  │Filesystem│ │  GitHub  │ │  Memory  │ │  Brave   │ │   Postgres   │  │   │
│  │  │  (free)  │ │  (free)  │ │  (free)  │ │  (paid)  │ │   (free)     │  │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │   │
│  │  │  Fetch   │ │Puppeteer │ │  SQLite  │ │  Chroma  │ │    Qdrant    │  │   │
│  │  │  (free)  │ │  (free)  │ │  (free)  │ │  (free)  │ │   (free)     │  │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │   │
│  └────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                        LLMsVerifier                                      │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────────┐  │   │
│  │  │   MCP       │  │   Provider  │  │   Model     │  │   Verification │  │   │
│  │  │   Testing   │  │   Registry  │  │   Scoring   │  │   Engine       │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Protocol Compliance

HelixAgent implements **100% of the MCP specification**:

| Feature | Status | Implementation |
|---------|--------|----------------|
| JSON-RPC 2.0 | ✅ Complete | `internal/services/mcp_client.go` |
| stdio Transport | ✅ Complete | `StdioTransport` struct |
| HTTP Transport | ✅ Complete | `HTTPTransport` struct |
| Server-Sent Events | ✅ Complete | SSE streaming support |
| Tools | ✅ Complete | Tool registration, calling, validation |
| Resources | ✅ Complete | Resource listing, reading |
| Prompts | ✅ Complete | Prompt templates |
| Roots | ✅ Complete | Workspace boundaries |
| Sampling | ✅ Complete | LLM sampling requests |
| Progress | ✅ Complete | Progress notifications |
| Cancellation | ✅ Complete | Request cancellation |
| Pagination | ✅ Complete | List pagination |

---

## Quick Start

### 1. Using MCP with HelixAgent CLI

```bash
# List all available MCP servers
./bin/helixagent --list-mcp-servers

# Start with specific MCP servers
./bin/helixagent --mcp-servers=github,filesystem,memory

# Start with all free MCP servers
./bin/helixagent --mcp-servers=free

# Use Claude Code MCP proxy
./bin/helixagent --claude-code --mcp-proxy
```

### 2. Using MCP Programmatically

```go
package main

import (
    "context"
    "log"
    
    "dev.helix.agent/internal/services"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create MCP client
    client := services.NewMCPClient(logger)
    
    ctx := context.Background()
    
    // Connect to filesystem MCP server
    err := client.ConnectServer(ctx, "filesystem", "Filesystem", "npx", []string{
        "-y", "@modelcontextprotocol/server-filesystem", "/path/to/workspace",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.DisconnectServer("filesystem")
    
    // Connect to GitHub MCP server
    err = client.ConnectServer(ctx, "github", "GitHub", "npx", []string{
        "-y", "@modelcontextprotocol/server-github",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.DisconnectServer("github")
    
    // List available tools
    tools, err := client.ListTools(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, tool := range tools {
        log.Printf("Tool: %s - %s", tool.Name, tool.Description)
    }
    
    // Call a tool
    result, err := client.CallTool(ctx, "filesystem", "read_file", map[string]interface{}{
        "path": "/path/to/workspace/README.md",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    for _, content := range result.Content {
        log.Printf("Content: %s", content.Text)
    }
}
```

### 3. Using MCP with LLMsVerifier

```bash
# Run verification with MCP tools
./bin/llm-verifier --mcp-enabled --mcp-servers=filesystem,github

# Test specific MCP capabilities
./bin/llm-verifier --test-mcp --server=github --tool=search_repositories

# Verify all MCP integrations
./bin/llm-verifier --verify-mcp-complete
```

---

## MCP Servers Reference

### Core Servers (Free)

| Server | Package | Description | Cost |
|--------|---------|-------------|------|
| filesystem | `@modelcontextprotocol/server-filesystem` | Secure file operations | Free |
| github | `@modelcontextprotocol/server-github` | GitHub API operations | Free |
| memory | `@modelcontextprotocol/server-memory` | Knowledge graph memory | Free |
| fetch | `mcp-fetch-server` | Web content fetching | Free |
| puppeteer | `@modelcontextprotocol/server-puppeteer` | Browser automation | Free |
| sqlite | `mcp-server-sqlite` | SQLite operations | Free |
| git | `mcp-git` | Git repository operations | Free |
| time | `@theo.foobar/mcp-time` | Time/timezone conversion | Free |
| sequential-thinking | `@modelcontextprotocol/server-sequential-thinking` | Problem solving | Free |

### Vector Database Servers

| Server | Package | Description | Cost |
|--------|---------|-------------|------|
| chroma | `mcp-server-chroma` | ChromaDB operations | Free |
| qdrant | `mcp-server-qdrant` | Qdrant operations | Free |
| weaviate | `mcp-server-weaviate` | Weaviate operations | Free |
| pinecone | `mcp-server-pinecone` | Pinecone operations | Free Tier |

### Search Servers

| Server | Package | Description | Cost |
|--------|---------|-------------|------|
| brave-search | `mcp-server-brave-search` | Brave Search API | Free Tier |
| tavily | `mcp-server-tavily` | Tavily search | Free Tier |
| duckduckgo | `mcp-server-duckduckgo` | DuckDuckGo search | Free |

### Development Servers

| Server | Package | Description | Cost |
|--------|---------|-------------|------|
| postgres | `mcp-server-postgres` | PostgreSQL operations | Free |
| mongodb | `mcp-server-mongodb` | MongoDB operations | Free |
| redis | `mcp-server-redis` | Redis operations | Free |
| docker | `mcp-server-docker` | Docker operations | Free |
| kubernetes | `mcp-server-kubernetes` | Kubernetes operations | Free |

### Cloud Storage Servers

| Server | Package | Description | Cost |
|--------|---------|-------------|------|
| s3 | `mcp-server-s3` | AWS S3 operations | Pay-per-use |
| gcs | `mcp-server-gcs` | Google Cloud Storage | Pay-per-use |
| google-drive | `mcp-server-google-drive` | Google Drive | Free Tier |

### Design & Image Servers

| Server | Package | Description | Cost |
|--------|---------|-------------|------|
| figma | `mcp-server-figma` | Figma operations | Free Tier |
| replicate | `mcp-server-replicate` | Image generation | Pay-per-use |
| stable-diffusion | `mcp-server-stable-diffusion` | Local SD WebUI | Free |

---

## Configuration

### Environment Variables

```bash
# MCP Global Settings
MCP_ENABLED=true
MCP_DEFAULT_SERVERS=filesystem,github,memory
MCP_TIMEOUT=30000
MCP_MAX_CONCURRENT=10

# Individual Server Settings
MCP_FILESYSTEM_ROOT=/path/to/workspace
MCP_GITHUB_TOKEN=ghp_xxxxxxxx
MCP_BRAVE_API_KEY=BSxxxxxxxx
MCP_PINECONE_API_KEY=pc_xxxxxxxx
MCP_POSTGRES_URL=postgresql://localhost/db

# Claude Code Integration
CLAUDE_CODE_MCP_PROXY=true
CLAUDE_CODE_MCP_SERVERS=github,filesystem
```

### Configuration File (config/mcp.yaml)

```yaml
mcp:
  enabled: true
  timeout: 30000
  max_concurrent: 10
  
  servers:
    filesystem:
      package: "@modelcontextprotocol/server-filesystem"
      enabled: true
      config:
        root: "/workspace"
      
    github:
      package: "@modelcontextprotocol/server-github"
      enabled: true
      config:
        token: "${MCP_GITHUB_TOKEN}"
      
    memory:
      package: "@modelcontextprotocol/server-memory"
      enabled: true
      
    brave-search:
      package: "mcp-server-brave-search"
      enabled: true
      config:
        api_key: "${MCP_BRAVE_API_KEY}"
      
    postgres:
      package: "mcp-server-postgres"
      enabled: true
      config:
        connection_string: "${MCP_POSTGRES_URL}"
```

---

## API Reference

### MCP Client API

```go
// Create client
client := services.NewMCPClient(logger)

// Connect to server
err := client.ConnectServer(ctx, serverID, name, command, args)

// Disconnect from server
err := client.DisconnectServer(serverID)

// List all available tools
tools, err := client.ListTools(ctx)

// Call a tool
result, err := client.CallTool(ctx, serverID, toolName, arguments)

// Get server info
info, err := client.GetServerInfo(serverID)

// List connected servers
servers := client.ListServers()

// Health check
status := client.HealthCheck(ctx)
```

### Claude Code MCP Proxy API

```go
// Create proxy API
proxyAPI := api.NewMCPProxyAPI(claudeClient)

// List available MCP servers
servers, err := proxyAPI.ListServers(ctx)

// Get specific server info
server, err := proxyAPI.GetServer(ctx, serverID)

// List server resources
resources, err := proxyAPI.ListResources(ctx, serverID)

// Read a resource
content, err := proxyAPI.ReadResource(ctx, serverID, resourceURI)

// Call a tool
result, err := proxyAPI.CallTool(ctx, serverID, &api.MCPCallRequest{
    Tool:   "search_repositories",
    Params: map[string]interface{}{
        "query": "golang mcp",
    },
})
```

---

## Advanced Features

### 1. Tool Argument Validation

```go
// Schema validation is automatic
result, err := client.CallTool(ctx, "filesystem", "write_file", map[string]interface{}{
    "path": "/workspace/test.txt",
    "content": "Hello, MCP!",
})
// Validation ensures "path" and "content" are provided
```

### 2. Progress Tracking

```go
// Long-running tools report progress
result, err := client.CallTool(ctx, "github", "clone_repository", map[string]interface{}{
    "url": "https://github.com/large/repo",
})
// Progress notifications sent via JSON-RPC
```

### 3. Request Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Start long operation
go func() {
    result, err := client.CallTool(ctx, ...)
}()

// Cancel if needed
cancel()
```

### 4. Resource Subscriptions

```go
// Subscribe to resource changes
err := client.SubscribeToResource(ctx, "filesystem", "file:///workspace/config.yaml")

// Handle notifications
for notification := range client.Notifications() {
    log.Printf("Resource changed: %s", notification.URI)
}
```

### 5. Sampling (LLM Requests from MCP Servers)

```go
// Handle sampling requests from MCP servers
client.SetSamplingHandler(func(request MCPSamplingRequest) (*MCPSamplingResponse, error) {
    // Forward to your LLM
    response, err := llmClient.Complete(request.Prompt)
    return &MCPSamplingResponse{
        Content: response,
    }, err
})
```

---

## Integration with HelixAgent Features

### 1. Debate Orchestrator + MCP

```go
// Use MCP tools in debates
debate := debate.New(agents, debator.WithMCPTools(client))

// Agents can call MCP tools during debates
result, err := debate.Run(ctx, "Compare these GitHub repositories", 
    debator.WithMCPTool("github", "search_repositories"))
```

### 2. RAG + MCP

```go
// Use Chroma/Qdrant MCP servers for RAG
rag := rag.New(
    rag.WithMCPVectorStore("chroma", client),
)

// Index documents
err := rag.IndexDocument(ctx, doc, rag.WithMCPIndexing())
```

### 3. Code Generation + MCP

```go
// Use filesystem MCP for code operations
generator := code.New(
    code.WithMCPFilesystem(client),
)

// Generated code is automatically written via MCP
code, err := generator.Generate(ctx, "Create a REST API")
```

---

## Testing with LLMsVerifier

### MCP Capability Tests

```bash
# Test all MCP capabilities
./bin/llm-verifier --test-suite=mcp

# Test specific server
./bin/llm-verifier --test-mcp-server=github

# Test specific tool
./bin/llm-verifier --test-mcp-tool=filesystem/read_file
```

### Programmatic Testing

```go
package main

import (
    "testing"
    "dev.helix.agent/llm-verifier/pkg/mcp"
)

func TestMCPIntegration(t *testing.T) {
    // Run MCP test suite
    runner := mcp.NewTestRunner()
    
    results := runner.RunAll(ctx)
    
    for _, result := range results {
        if !result.Passed {
            t.Errorf("MCP test failed: %s - %v", result.Name, result.Error)
        }
    }
}
```

---

## Troubleshooting

### Common Issues

**Issue:** MCP server fails to start
```bash
# Check Node.js is installed
node --version

# Install npx if missing
npm install -g npx

# Test server manually
npx -y @modelcontextprotocol/server-filesystem /tmp
```

**Issue:** Tool calls timeout
```yaml
# Increase timeout in config
mcp:
  timeout: 60000  # 60 seconds
```

**Issue:** Permission denied
```bash
# Check filesystem permissions
ls -la /path/to/workspace

# Run with appropriate permissions
./bin/helixagent --mcp-filesystem-root=/allowed/path
```

---

## Best Practices

1. **Security**: Always use the filesystem MCP with restricted roots
2. **Performance**: Limit concurrent MCP connections (default: 10)
3. **Reliability**: Implement retry logic for transient failures
4. **Monitoring**: Track MCP tool usage and latency
5. **Cost**: Use free/open-source servers when possible

---

## References

- [MCP Official Documentation](https://modelcontextprotocol.io/docs)
- [MCP Specification](https://modelcontextprotocol.io/specification)
- [Anthropic MCP Guide](https://docs.anthropic.com/en/docs/build-with-claude/mcp)
- [HelixAgent MCP Examples](../../examples/mcp/)

---

## Support

For MCP-related issues:
- GitHub Issues: `vasic-digital/HelixAgent`
- MCP Discord: [Model Context Protocol](https://discord.gg/modelcontextprotocol)
- Documentation: [HelixAgent Docs](https://docs.helixagent.dev/mcp)
