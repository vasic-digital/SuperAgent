# Lab 4: MCP Protocol Integration

## Overview

In this lab, you will learn how to integrate the Model Context Protocol (MCP) with HelixAgent to enable seamless tool and resource sharing between AI agents and external systems.

**Duration**: 2.5 hours

**Prerequisites**:
- Completed Labs 1-3
- HelixAgent running locally
- Docker installed
- Basic understanding of HTTP protocols

## Learning Objectives

By the end of this lab, you will be able to:
- Understand the MCP protocol architecture
- Configure MCP servers in HelixAgent
- Register and implement custom MCP tools
- Connect external MCP clients
- Test and debug MCP integrations

---

## Part 1: Understanding MCP (30 minutes)

### What is MCP?

The Model Context Protocol (MCP) is a standardized protocol for connecting AI models with external tools, data sources, and services. It provides a unified interface for:

- **Tools**: Functions that AI can invoke
- **Resources**: Data sources AI can access
- **Prompts**: Reusable prompt templates

### MCP Architecture

```
┌─────────────────┐         ┌─────────────────┐
│   MCP Client    │ ◄─────► │   MCP Server    │
│  (HelixAgent)   │  JSON   │   (External)    │
└─────────────────┘  RPC    └─────────────────┘
        │                           │
        ▼                           ▼
┌─────────────────┐         ┌─────────────────┐
│   AI Model      │         │  Tools/Resources│
└─────────────────┘         └─────────────────┘
```

### Exercise 1.1: Review MCP Configuration

Examine HelixAgent's MCP configuration:

```bash
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent

# View MCP package structure
ls -la internal/mcp/

# Read MCP types
cat internal/mcp/types.go | head -100
```

**Questions to Answer**:
1. What are the main components of an MCP tool definition?
2. How does HelixAgent handle MCP authentication?
3. What transport protocols are supported?

---

## Part 2: Setting Up MCP Server (30 minutes)

### Exercise 2.1: Start the Built-in MCP Server

HelixAgent includes a built-in MCP server:

```bash
# Build HelixAgent if not already built
make build

# Start the MCP server (default port 8081)
./bin/helixagent mcp-server --port 8081 &

# Verify server is running
curl http://localhost:8081/health
```

Expected output:
```json
{"status": "healthy", "version": "1.0.0"}
```

### Exercise 2.2: Configure MCP Client Connection

Create a configuration file for MCP clients:

```yaml
# configs/mcp-client.yaml
mcp:
  servers:
    - name: "helixagent-mcp"
      url: "http://localhost:8081"
      transport: "http"
      auth:
        type: "bearer"
        token: "${MCP_AUTH_TOKEN}"
      tools:
        enabled: true
        whitelist:
          - read_file
          - write_file
          - execute_command
          - search_codebase
      resources:
        enabled: true
      prompts:
        enabled: true
```

### Exercise 2.3: Test Basic Connection

```bash
# Test MCP server info endpoint
curl http://localhost:8081/mcp/info

# List available tools
curl http://localhost:8081/mcp/tools

# List available resources
curl http://localhost:8081/mcp/resources
```

---

## Part 3: Registering Custom Tools (45 minutes)

### Exercise 3.1: Define a Custom Tool

Create a new MCP tool definition:

```go
// Example: internal/mcp/tools/search_codebase.go

package tools

import (
    "context"
    "dev.helix.agent/internal/mcp"
)

// SearchCodebaseTool searches the codebase for patterns
var SearchCodebaseTool = &mcp.Tool{
    Name:        "search_codebase",
    Description: "Search the codebase for files matching a pattern or containing specific text",
    InputSchema: mcp.Schema{
        Type: "object",
        Properties: map[string]mcp.Property{
            "pattern": {
                Type:        "string",
                Description: "Search pattern (glob for files, regex for content)",
            },
            "path": {
                Type:        "string",
                Description: "Directory to search in",
                Default:     ".",
            },
            "type": {
                Type:        "string",
                Enum:        []string{"file", "content"},
                Description: "Search type: 'file' for filename matching, 'content' for text search",
                Default:     "content",
            },
        },
        Required: []string{"pattern"},
    },
}
```

### Exercise 3.2: Implement Tool Handler

```go
// Example handler implementation

func HandleSearchCodebase(ctx context.Context, params map[string]interface{}) (*mcp.ToolResult, error) {
    pattern, ok := params["pattern"].(string)
    if !ok {
        return nil, fmt.Errorf("pattern is required")
    }
    
    path, _ := params["path"].(string)
    if path == "" {
        path = "."
    }
    
    searchType, _ := params["type"].(string)
    if searchType == "" {
        searchType = "content"
    }
    
    var results []SearchResult
    
    if searchType == "file" {
        results, err = searchFiles(path, pattern)
    } else {
        results, err = searchContent(path, pattern)
    }
    
    if err != nil {
        return &mcp.ToolResult{
            IsError: true,
            Content: []mcp.Content{
                {Type: "text", Text: fmt.Sprintf("Search failed: %v", err)},
            },
        }, nil
    }
    
    return &mcp.ToolResult{
        Content: []mcp.Content{
            {Type: "text", Text: formatResults(results)},
        },
    }, nil
}
```

### Exercise 3.3: Register the Tool

```go
// In internal/mcp/server.go

func (s *Server) RegisterTools() {
    s.RegisterTool(tools.SearchCodebaseTool, tools.HandleSearchCodebase)
    s.RegisterTool(tools.ReadFileTool, tools.HandleReadFile)
    s.RegisterTool(tools.WriteFileTool, tools.HandleWriteFile)
    // Add more tools...
}
```

---

## Part 4: Testing MCP Integration (45 minutes)

### Exercise 4.1: Test Tool Execution via HTTP

```bash
# Execute the search_codebase tool
curl -X POST http://localhost:8081/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "search_codebase",
    "arguments": {
      "pattern": "func Test",
      "path": "./internal",
      "type": "content"
    }
  }'
```

Expected response:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Found 150 matches in 45 files:\n- internal/llm/ensemble_test.go:15\n- internal/services/debate_test.go:23\n..."
    }
  ],
  "isError": false
}
```

### Exercise 4.2: Write Integration Test

Create a comprehensive integration test:

```go
// tests/integration/mcp_integration_test.go

//go:build integration

package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "dev.helix.agent/internal/mcp"
)

func TestMCPIntegration(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Create MCP client
    client, err := mcp.NewClient("http://localhost:8081")
    require.NoError(t, err)
    
    t.Run("list tools", func(t *testing.T) {
        tools, err := client.ListTools(ctx)
        require.NoError(t, err)
        assert.NotEmpty(t, tools)
        
        // Verify search_codebase tool exists
        found := false
        for _, tool := range tools {
            if tool.Name == "search_codebase" {
                found = true
                break
            }
        }
        assert.True(t, found, "search_codebase tool should be available")
    })
    
    t.Run("execute tool", func(t *testing.T) {
        result, err := client.CallTool(ctx, "search_codebase", map[string]interface{}{
            "pattern": "package main",
            "path":    "./cmd",
            "type":    "content",
        })
        require.NoError(t, err)
        assert.False(t, result.IsError)
        assert.NotEmpty(t, result.Content)
    })
    
    t.Run("tool error handling", func(t *testing.T) {
        result, err := client.CallTool(ctx, "nonexistent_tool", nil)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "tool not found")
    })
}
```

### Exercise 4.3: Run Integration Tests

```bash
# Start test infrastructure
make test-infra-start

# Run MCP integration tests
go test -v -tags=integration ./tests/integration/mcp_integration_test.go

# Stop infrastructure
make test-infra-stop
```

---

## Part 5: Advanced MCP Features (30 minutes)

### Exercise 5.1: Resource Provider

Implement a resource provider for database access:

```go
// MCP Resource for database queries
var DatabaseResource = &mcp.Resource{
    URI:         "db://helixagent/queries",
    Name:        "Database Queries",
    Description: "Execute read-only database queries",
    MimeType:    "application/json",
}

func HandleDatabaseResource(ctx context.Context, uri string) (*mcp.ResourceContent, error) {
    // Parse query from URI
    query := extractQuery(uri)
    
    // Execute read-only query
    results, err := db.QueryReadOnly(ctx, query)
    if err != nil {
        return nil, err
    }
    
    return &mcp.ResourceContent{
        URI:      uri,
        MimeType: "application/json",
        Text:     string(results),
    }, nil
}
```

### Exercise 5.2: Prompt Templates

Create reusable prompt templates:

```go
var CodeReviewPrompt = &mcp.Prompt{
    Name:        "code_review",
    Description: "Review code for issues and improvements",
    Arguments: []mcp.PromptArgument{
        {Name: "code", Description: "Code to review", Required: true},
        {Name: "language", Description: "Programming language", Required: false},
    },
}

func HandleCodeReviewPrompt(args map[string]string) (*mcp.PromptResult, error) {
    code := args["code"]
    language := args["language"]
    
    template := fmt.Sprintf(`Review the following %s code for:
1. Bugs and errors
2. Performance issues
3. Security vulnerabilities
4. Code style improvements

Code:
%s

Provide specific, actionable feedback.`, language, code)
    
    return &mcp.PromptResult{
        Messages: []mcp.Message{
            {Role: "user", Content: mcp.TextContent{Text: template}},
        },
    }, nil
}
```

---

## Verification Checklist

Before completing this lab, verify:

- [ ] MCP server starts and responds to health checks
- [ ] Can list available tools via HTTP API
- [ ] Custom tool (search_codebase) is registered and functional
- [ ] Tool execution returns expected results
- [ ] Error handling works correctly for invalid inputs
- [ ] Integration tests pass
- [ ] Resources can be accessed via MCP
- [ ] Prompt templates work correctly

## Summary

In this lab, you learned:

1. **MCP Architecture**: Understanding the protocol structure and components
2. **Server Setup**: Configuring and starting an MCP server
3. **Tool Registration**: Creating and registering custom MCP tools
4. **Testing**: Writing and running MCP integration tests
5. **Advanced Features**: Resources and prompt templates

## Next Steps

- **Lab 5**: Production Deployment
- Explore MCP server federation
- Build custom MCP clients
- Implement streaming tool responses

## Resources

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [HelixAgent MCP Documentation](../../api/mcp.md)
- [MCP Server Examples](../../examples/mcp/)
