# Protocol Advanced Examples

This document provides comprehensive advanced examples for the three protocols supported by HelixAgent: Language Server Protocol (LSP), Model Context Protocol (MCP), and Agent Client Protocol (ACP).

## Table of Contents

- [LSP Advanced Examples](#lsp-advanced-examples)
- [MCP Advanced Examples](#mcp-advanced-examples)
- [ACP Advanced Examples](#acp-advanced-examples)

---

## LSP Advanced Examples

HelixAgent's LSP integration enables advanced code intelligence capabilities through Language Server Protocol servers.

### Example 1: Multi-Language Workspace Setup

Connect to multiple language servers for a polyglot project:

```go
package main

import (
    "context"
    "dev.helix.agent/internal/services"
    "github.com/sirupsen/logrus"
)

func setupMultiLanguageWorkspace(ctx context.Context) error {
    logger := logrus.New()
    lspClient := services.NewLSPClient(logger)

    // Connect Go language server
    err := lspClient.ConnectServer(ctx,
        "go-server",        // Server ID
        "gopls",            // Server name
        "go",               // Language
        "gopls",            // Command
        []string{},         // Args
        "/workspace/go",    // Workspace root
    )
    if err != nil {
        return fmt.Errorf("failed to connect gopls: %w", err)
    }

    // Connect TypeScript language server
    err = lspClient.ConnectServer(ctx,
        "ts-server",
        "typescript-language-server",
        "typescript",
        "typescript-language-server",
        []string{"--stdio"},
        "/workspace/frontend",
    )
    if err != nil {
        return fmt.Errorf("failed to connect ts server: %w", err)
    }

    // Connect Python language server
    err = lspClient.ConnectServer(ctx,
        "python-server",
        "pylsp",
        "python",
        "pylsp",
        []string{},
        "/workspace/ml",
    )
    if err != nil {
        return fmt.Errorf("failed to connect pylsp: %w", err)
    }

    logger.Info("Multi-language workspace initialized")
    return nil
}
```

### Example 2: Advanced Code Completion with Context

Get context-aware code completions with surrounding code analysis:

```go
func getAdvancedCompletion(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"
    fileURI := "file:///workspace/go/internal/handlers/api.go"

    // Read file content
    content, err := os.ReadFile("/workspace/go/internal/handlers/api.go")
    if err != nil {
        return err
    }

    // Open document
    err = lspClient.OpenFile(ctx, serverID, fileURI, "go", string(content))
    if err != nil {
        return err
    }

    // Request completion at cursor position (line 42, character 15)
    completions, err := lspClient.GetCompletion(ctx, serverID, fileURI, 42, 15)
    if err != nil {
        return err
    }

    // Filter and rank completions
    for _, item := range completions.Items {
        fmt.Printf("Label: %s\n", item.Label)
        fmt.Printf("Kind: %d\n", item.Kind)
        fmt.Printf("Detail: %s\n", item.Detail)
        fmt.Printf("Documentation: %s\n\n", item.Documentation)
    }

    return nil
}
```

### Example 3: Go-to-Definition with Cross-File Navigation

Navigate to symbol definitions across files:

```go
func navigateToDefinition(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"
    fileURI := "file:///workspace/go/cmd/main.go"

    // Get definition for symbol at position
    location, err := lspClient.GetDefinition(ctx, serverID, fileURI, 25, 10)
    if err != nil {
        return err
    }

    fmt.Printf("Definition found at:\n")
    fmt.Printf("  File: %s\n", location.URI)
    fmt.Printf("  Line: %d-%d\n", location.Range.Start.Line, location.Range.End.Line)
    fmt.Printf("  Character: %d-%d\n", location.Range.Start.Character, location.Range.End.Character)

    // Open the definition file
    definitionContent, _ := readFileFromURI(location.URI)
    err = lspClient.OpenFile(ctx, serverID, location.URI, "go", definitionContent)

    return err
}
```

### Example 4: Semantic Hover Information

Get detailed type and documentation information on hover:

```go
func getSemanticHover(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "ts-server"
    fileURI := "file:///workspace/frontend/src/components/Dashboard.tsx"

    // Get hover information
    hover, err := lspClient.GetHover(ctx, serverID, fileURI, 55, 22)
    if err != nil {
        return err
    }

    fmt.Printf("Hover Information:\n")
    fmt.Printf("  Content Type: %s\n", hover.Contents.Kind)
    fmt.Printf("  Content:\n%s\n", hover.Contents.Value)

    if hover.Range != nil {
        fmt.Printf("  Range: L%d:C%d - L%d:C%d\n",
            hover.Range.Start.Line, hover.Range.Start.Character,
            hover.Range.End.Line, hover.Range.End.Character)
    }

    return nil
}
```

### Example 5: Real-Time Document Synchronization

Keep documents synchronized with live editing:

```go
func syncDocumentEdits(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"
    fileURI := "file:///workspace/go/internal/service.go"

    // Initial document open
    originalContent := `package internal

func ProcessData(data []byte) error {
    // TODO: implement
    return nil
}
`
    err := lspClient.OpenFile(ctx, serverID, fileURI, "go", originalContent)
    if err != nil {
        return err
    }

    // Simulate editing - add implementation
    updatedContent := `package internal

import "encoding/json"

func ProcessData(data []byte) error {
    var result map[string]interface{}
    if err := json.Unmarshal(data, &result); err != nil {
        return fmt.Errorf("failed to parse data: %w", err)
    }
    return nil
}
`
    // Send incremental update
    err = lspClient.UpdateFile(ctx, serverID, fileURI, updatedContent)
    if err != nil {
        return err
    }

    // Get completions in updated context
    completions, _ := lspClient.GetCompletion(ctx, serverID, fileURI, 6, 20)
    fmt.Printf("Available completions after edit: %d\n", len(completions.Items))

    return nil
}
```

### Example 6: Find All References

Locate all usages of a symbol across the codebase:

```go
func findAllReferences(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"
    fileURI := "file:///workspace/go/internal/models/user.go"

    // Find references for the User struct at line 10
    referencesReq := services.LSPMessage{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "textDocument/references",
        Params: map[string]interface{}{
            "textDocument": map[string]string{"uri": fileURI},
            "position":     map[string]int{"line": 10, "character": 6},
            "context":      map[string]bool{"includeDeclaration": true},
        },
    }

    // Send request through transport
    connection, _ := lspClient.GetServerCapabilities(serverID)
    fmt.Printf("Finding all references for User struct...\n")
    fmt.Printf("Server capabilities: %+v\n", connection)

    return nil
}
```

### Example 7: Workspace Symbol Search

Search for symbols across the entire workspace:

```go
func searchWorkspaceSymbols(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"

    // Search for all Handler types
    symbolReq := services.LSPMessage{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "workspace/symbol",
        Params: map[string]interface{}{
            "query": "Handler",
        },
    }

    fmt.Printf("Searching for Handler symbols in workspace...\n")
    fmt.Printf("Request: %+v\n", symbolReq)

    return nil
}
```

### Example 8: Code Action Requests

Get available refactoring actions:

```go
func getCodeActions(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"
    fileURI := "file:///workspace/go/internal/api/handler.go"

    codeActionReq := services.LSPMessage{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "textDocument/codeAction",
        Params: map[string]interface{}{
            "textDocument": map[string]string{"uri": fileURI},
            "range": map[string]interface{}{
                "start": map[string]int{"line": 20, "character": 0},
                "end":   map[string]int{"line": 30, "character": 0},
            },
            "context": map[string]interface{}{
                "diagnostics": []interface{}{},
                "only":        []string{"quickfix", "refactor"},
            },
        },
    }

    fmt.Printf("Available code actions: %+v\n", codeActionReq)
    return nil
}
```

### Example 9: Document Formatting

Format code according to language standards:

```go
func formatDocument(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"
    fileURI := "file:///workspace/go/internal/unformatted.go"

    formatReq := services.LSPMessage{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "textDocument/formatting",
        Params: map[string]interface{}{
            "textDocument": map[string]string{"uri": fileURI},
            "options": map[string]interface{}{
                "tabSize":      4,
                "insertSpaces": false,
            },
        },
    }

    fmt.Printf("Formatting document: %+v\n", formatReq)
    return nil
}
```

### Example 10: Rename Symbol Refactoring

Safely rename symbols across the codebase:

```go
func renameSymbol(ctx context.Context, lspClient *services.LSPClient) error {
    serverID := "go-server"
    fileURI := "file:///workspace/go/internal/models/user.go"

    renameReq := services.LSPMessage{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "textDocument/rename",
        Params: map[string]interface{}{
            "textDocument": map[string]string{"uri": fileURI},
            "position":     map[string]int{"line": 15, "character": 10},
            "newName":      "Account",
        },
    }

    fmt.Printf("Rename refactoring request: %+v\n", renameReq)
    return nil
}
```

---

## MCP Advanced Examples

HelixAgent's MCP integration enables interaction with external tools and services.

### Example 1: Multi-Server MCP Configuration

Configure and connect to multiple MCP servers:

```go
package main

import (
    "context"
    "dev.helix.agent/internal/services"
    "github.com/sirupsen/logrus"
)

func setupMCPServers(ctx context.Context) (*services.MCPClient, error) {
    logger := logrus.New()
    mcpClient := services.NewMCPClient(logger)

    // Connect to filesystem MCP server
    err := mcpClient.ConnectServer(ctx, services.MCPServerConfig{
        ID:        "filesystem",
        Name:      "Filesystem Tools",
        Transport: "stdio",
        Command:   "npx",
        Args:      []string{"-y", "@modelcontextprotocol/server-filesystem", "/workspace"},
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect filesystem server: %w", err)
    }

    // Connect to database MCP server
    err = mcpClient.ConnectServer(ctx, services.MCPServerConfig{
        ID:        "postgres",
        Name:      "PostgreSQL Tools",
        Transport: "stdio",
        Command:   "npx",
        Args:      []string{"-y", "@modelcontextprotocol/server-postgres"},
        Env: map[string]string{
            "POSTGRES_URL": "postgresql://user:pass@localhost:5432/mydb",
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect postgres server: %w", err)
    }

    // Connect to GitHub MCP server
    err = mcpClient.ConnectServer(ctx, services.MCPServerConfig{
        ID:        "github",
        Name:      "GitHub Tools",
        Transport: "stdio",
        Command:   "npx",
        Args:      []string{"-y", "@modelcontextprotocol/server-github"},
        Env: map[string]string{
            "GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect github server: %w", err)
    }

    return mcpClient, nil
}
```

### Example 2: Tool Discovery and Invocation

Discover available tools and invoke them dynamically:

```go
func discoverAndInvokeTools(ctx context.Context, mcpClient *services.MCPClient) error {
    // List all available tools
    tools, err := mcpClient.ListTools(ctx)
    if err != nil {
        return err
    }

    fmt.Println("Available MCP Tools:")
    for _, tool := range tools {
        fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
        fmt.Printf("    Schema: %v\n", tool.InputSchema)
    }

    // Invoke a specific tool
    result, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "filesystem",
        Name:     "read_file",
        Arguments: map[string]interface{}{
            "path": "/workspace/README.md",
        },
    })
    if err != nil {
        return err
    }

    fmt.Printf("Tool result:\n%v\n", result)
    return nil
}
```

### Example 3: Resource Management

List and read MCP resources:

```go
func manageResources(ctx context.Context, mcpClient *services.MCPClient) error {
    // List available resources
    resources, err := mcpClient.ListResources(ctx)
    if err != nil {
        return err
    }

    fmt.Println("Available Resources:")
    for _, resource := range resources {
        fmt.Printf("  URI: %s\n", resource.URI)
        fmt.Printf("  Name: %s\n", resource.Name)
        fmt.Printf("  MIME Type: %s\n", resource.MimeType)
        fmt.Printf("  Description: %s\n\n", resource.Description)
    }

    // Read a specific resource
    content, err := mcpClient.ReadResource(ctx, "helixagent://providers")
    if err != nil {
        return err
    }

    fmt.Printf("Resource content:\n%v\n", content)
    return nil
}
```

### Example 4: Prompt Template Usage

Use MCP prompt templates for consistent interactions:

```go
func usePromptTemplates(ctx context.Context, mcpClient *services.MCPClient) error {
    // List available prompts
    prompts, err := mcpClient.ListPrompts(ctx)
    if err != nil {
        return err
    }

    fmt.Println("Available Prompts:")
    for _, prompt := range prompts {
        fmt.Printf("  - %s: %s\n", prompt.Name, prompt.Description)
    }

    // Get a specific prompt with arguments
    messages, err := mcpClient.GetPrompt(ctx, services.MCPPromptRequest{
        Name: "analyze_code",
        Arguments: map[string]string{
            "language":  "go",
            "file_path": "/workspace/main.go",
            "focus":     "security",
        },
    })
    if err != nil {
        return err
    }

    fmt.Println("Generated prompt messages:")
    for _, msg := range messages {
        fmt.Printf("  Role: %s\n", msg.Role)
        fmt.Printf("  Content: %s\n\n", msg.Content)
    }

    return nil
}
```

### Example 5: Database Query Tool

Execute database queries through MCP:

```go
func executeDatabaseQuery(ctx context.Context, mcpClient *services.MCPClient) error {
    // Query users table
    result, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "postgres",
        Name:     "query",
        Arguments: map[string]interface{}{
            "sql": "SELECT id, email, created_at FROM users WHERE active = true LIMIT 10",
        },
    })
    if err != nil {
        return err
    }

    fmt.Printf("Query results:\n%v\n", result)

    // Describe table schema
    schema, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "postgres",
        Name:     "describe_table",
        Arguments: map[string]interface{}{
            "table": "users",
        },
    })
    if err != nil {
        return err
    }

    fmt.Printf("Table schema:\n%v\n", schema)
    return nil
}
```

### Example 6: GitHub Repository Operations

Interact with GitHub through MCP tools:

```go
func githubOperations(ctx context.Context, mcpClient *services.MCPClient) error {
    // Create a pull request
    pr, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "github",
        Name:     "create_pull_request",
        Arguments: map[string]interface{}{
            "owner":  "helixagent",
            "repo":   "helixagent",
            "title":  "Add new feature",
            "body":   "This PR adds a new feature for...",
            "head":   "feature-branch",
            "base":   "main",
        },
    })
    if err != nil {
        return err
    }
    fmt.Printf("Created PR: %v\n", pr)

    // List issues
    issues, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "github",
        Name:     "list_issues",
        Arguments: map[string]interface{}{
            "owner":  "helixagent",
            "repo":   "helixagent",
            "state":  "open",
            "labels": []string{"bug", "high-priority"},
        },
    })
    if err != nil {
        return err
    }
    fmt.Printf("Open issues: %v\n", issues)

    return nil
}
```

### Example 7: File System Operations

Perform filesystem operations through MCP:

```go
func filesystemOperations(ctx context.Context, mcpClient *services.MCPClient) error {
    // List directory contents
    listing, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "filesystem",
        Name:     "list_directory",
        Arguments: map[string]interface{}{
            "path": "/workspace/src",
        },
    })
    if err != nil {
        return err
    }
    fmt.Printf("Directory listing:\n%v\n", listing)

    // Search for files
    files, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "filesystem",
        Name:     "search_files",
        Arguments: map[string]interface{}{
            "path":    "/workspace",
            "pattern": "*.go",
        },
    })
    if err != nil {
        return err
    }
    fmt.Printf("Go files found:\n%v\n", files)

    // Write a file
    _, err = mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "filesystem",
        Name:     "write_file",
        Arguments: map[string]interface{}{
            "path":    "/workspace/output/report.json",
            "content": `{"status": "complete", "items": 42}`,
        },
    })
    if err != nil {
        return err
    }
    fmt.Println("File written successfully")

    return nil
}
```

### Example 8: Tool Chaining

Chain multiple MCP tools together:

```go
func chainedToolExecution(ctx context.Context, mcpClient *services.MCPClient) error {
    // Step 1: Read a configuration file
    configContent, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "filesystem",
        Name:     "read_file",
        Arguments: map[string]interface{}{
            "path": "/workspace/config.yaml",
        },
    })
    if err != nil {
        return err
    }

    // Step 2: Parse and extract database URL from config
    var config map[string]interface{}
    yaml.Unmarshal([]byte(configContent.(string)), &config)
    dbURL := config["database"].(map[string]interface{})["url"].(string)

    // Step 3: Query the database
    users, err := mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "postgres",
        Name:     "query",
        Arguments: map[string]interface{}{
            "sql": "SELECT COUNT(*) as count FROM users",
        },
    })
    if err != nil {
        return err
    }

    // Step 4: Write results to a file
    report := fmt.Sprintf("Database: %s\nUser count: %v\n", dbURL, users)
    _, err = mcpClient.CallTool(ctx, services.MCPToolCall{
        ServerID: "filesystem",
        Name:     "write_file",
        Arguments: map[string]interface{}{
            "path":    "/workspace/reports/user_count.txt",
            "content": report,
        },
    })

    return err
}
```

### Example 9: MCP Server Health Monitoring

Monitor MCP server health and status:

```go
func monitorMCPServers(ctx context.Context, mcpClient *services.MCPClient) error {
    servers := mcpClient.ListServers()

    for _, server := range servers {
        fmt.Printf("Server: %s\n", server.ID)
        fmt.Printf("  Name: %s\n", server.Name)
        fmt.Printf("  Connected: %v\n", server.Connected)
        fmt.Printf("  Last Used: %v\n", server.LastUsed)

        // Check capabilities
        caps, err := mcpClient.GetCapabilities(ctx, server.ID)
        if err != nil {
            fmt.Printf("  Status: ERROR - %v\n", err)
            continue
        }

        fmt.Printf("  Capabilities:\n")
        fmt.Printf("    Tools: %v\n", caps.Tools.ListChanged)
        fmt.Printf("    Prompts: %v\n", caps.Prompts.ListChanged)
        fmt.Printf("    Resources: %v\n", caps.Resources.ListChanged)
    }

    return nil
}
```

### Example 10: Custom MCP Tool Creation

Create and register custom MCP tools:

```go
func registerCustomTools(ctx context.Context, mcpClient *services.MCPClient) error {
    // Define a custom tool
    customTool := services.MCPTool{
        Name:        "analyze_sentiment",
        Description: "Analyze sentiment of text using LLM",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "text": map[string]interface{}{
                    "type":        "string",
                    "description": "Text to analyze",
                },
                "language": map[string]interface{}{
                    "type":        "string",
                    "description": "Language of the text",
                    "default":     "en",
                },
            },
            "required": []string{"text"},
        },
    }

    // Register the tool handler
    mcpClient.RegisterToolHandler("analyze_sentiment", func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        text := args["text"].(string)
        language := args["language"].(string)

        // Call LLM for sentiment analysis
        result, err := analyzeSentimentWithLLM(ctx, text, language)
        if err != nil {
            return nil, err
        }

        return map[string]interface{}{
            "sentiment": result.Sentiment,
            "score":     result.Score,
            "keywords":  result.Keywords,
        }, nil
    })

    fmt.Printf("Registered custom tool: %s\n", customTool.Name)
    return nil
}
```

---

## ACP Advanced Examples

HelixAgent's ACP (Agent Client Protocol) enables communication with external agent servers.

### Example 1: ACP Server Registration and Discovery

Register and discover ACP servers dynamically:

```go
package main

import (
    "context"
    "dev.helix.agent/internal/services"
    "github.com/sirupsen/logrus"
)

func setupACPServers(ctx context.Context) (*services.ACPManager, error) {
    logger := logrus.New()
    acpManager := services.NewACPManager(nil, nil, logger)

    // Register HTTP ACP server
    err := acpManager.RegisterServer(&services.ACPServer{
        ID:      "code-agent",
        Name:    "Code Analysis Agent",
        URL:     "http://localhost:8090/acp",
        Enabled: true,
        Capabilities: []services.ACPCapability{
            {Name: "analyze_code", Description: "Analyze code for issues"},
            {Name: "generate_tests", Description: "Generate unit tests"},
            {Name: "refactor", Description: "Suggest refactoring"},
        },
    })
    if err != nil {
        return nil, err
    }

    // Register WebSocket ACP server
    err = acpManager.RegisterServer(&services.ACPServer{
        ID:      "realtime-agent",
        Name:    "Realtime Collaboration Agent",
        URL:     "ws://localhost:8091/acp",
        Enabled: true,
        Capabilities: []services.ACPCapability{
            {Name: "collaborative_edit", Description: "Real-time collaborative editing"},
            {Name: "presence", Description: "User presence tracking"},
        },
    })
    if err != nil {
        return nil, err
    }

    return acpManager, nil
}
```

### Example 2: Execute ACP Actions

Execute various actions on ACP servers:

```go
func executeACPActions(ctx context.Context, acpManager *services.ACPManager) error {
    // Execute code analysis
    result, err := acpManager.ExecuteACPAction(ctx, services.ACPRequest{
        ServerID: "code-agent",
        Action:   "analyze_code",
        Parameters: map[string]interface{}{
            "file_path": "/workspace/src/main.go",
            "checks":    []string{"security", "performance", "style"},
        },
    })
    if err != nil {
        return err
    }

    fmt.Printf("Analysis result:\n")
    fmt.Printf("  Success: %v\n", result.Success)
    fmt.Printf("  Data: %v\n", result.Data)
    fmt.Printf("  Timestamp: %v\n", result.Timestamp)

    return nil
}
```

### Example 3: WebSocket Real-Time Communication

Establish real-time communication with ACP servers:

```go
func realtimeCommunication(ctx context.Context) error {
    acpClient := services.NewACPClient(30*time.Second, 3, logrus.New())

    // Connect via WebSocket
    serverURL := "ws://localhost:8091/acp"

    // Subscribe to events
    subscribeReq := services.ACPProtocolRequest{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "subscribe",
        Params: map[string]interface{}{
            "events": []string{"code_change", "cursor_move", "user_join"},
        },
    }

    resp, err := acpClient.ExecuteWS(ctx, serverURL, subscribeReq)
    if err != nil {
        return err
    }

    fmt.Printf("Subscription response: %v\n", resp)

    // Listen for events
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Process incoming events
            eventReq := services.ACPProtocolRequest{
                JSONRPC: "2.0",
                ID:      2,
                Method:  "poll_events",
            }
            events, err := acpClient.ExecuteWS(ctx, serverURL, eventReq)
            if err != nil {
                continue
            }
            fmt.Printf("Received events: %v\n", events)
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

### Example 4: ACP Server Synchronization

Synchronize state with ACP servers:

```go
func syncACPServer(ctx context.Context, acpManager *services.ACPManager) error {
    serverID := "code-agent"

    // Sync server configuration
    err := acpManager.SyncACPServer(ctx, serverID)
    if err != nil {
        return fmt.Errorf("sync failed: %w", err)
    }

    // Get updated server info
    server, err := acpManager.GetACPServer(ctx, serverID)
    if err != nil {
        return err
    }

    fmt.Printf("Server synced:\n")
    fmt.Printf("  ID: %s\n", server.ID)
    fmt.Printf("  Name: %s\n", server.Name)
    fmt.Printf("  Version: %s\n", server.Version)
    fmt.Printf("  Last Sync: %v\n", server.LastSync)
    fmt.Printf("  Capabilities: %d\n", len(server.Capabilities))

    for _, cap := range server.Capabilities {
        fmt.Printf("    - %s: %s\n", cap.Name, cap.Description)
    }

    return nil
}
```

### Example 5: Batch ACP Requests

Execute multiple ACP requests efficiently:

```go
func batchACPRequests(ctx context.Context, acpManager *services.ACPManager) error {
    // Prepare batch of requests
    requests := []services.ACPRequest{
        {
            ServerID: "code-agent",
            Action:   "analyze_file",
            Parameters: map[string]interface{}{
                "path": "/workspace/src/handler.go",
            },
        },
        {
            ServerID: "code-agent",
            Action:   "analyze_file",
            Parameters: map[string]interface{}{
                "path": "/workspace/src/service.go",
            },
        },
        {
            ServerID: "code-agent",
            Action:   "analyze_file",
            Parameters: map[string]interface{}{
                "path": "/workspace/src/repository.go",
            },
        },
    }

    // Execute in parallel
    var wg sync.WaitGroup
    results := make([]*services.ACPResponse, len(requests))
    errors := make([]error, len(requests))

    for i, req := range requests {
        wg.Add(1)
        go func(idx int, r services.ACPRequest) {
            defer wg.Done()
            result, err := acpManager.ExecuteACPAction(ctx, r)
            results[idx] = result
            errors[idx] = err
        }(i, req)
    }

    wg.Wait()

    // Process results
    for i, result := range results {
        if errors[i] != nil {
            fmt.Printf("Request %d failed: %v\n", i, errors[i])
            continue
        }
        fmt.Printf("Request %d succeeded: %v\n", i, result.Data)
    }

    return nil
}
```

### Example 6: ACP Error Handling and Retry

Implement robust error handling with retries:

```go
func executeWithRetry(ctx context.Context, acpManager *services.ACPManager, req services.ACPRequest) (*services.ACPResponse, error) {
    maxRetries := 3
    backoff := 100 * time.Millisecond

    var lastErr error
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(backoff * time.Duration(attempt)):
            }
        }

        result, err := acpManager.ExecuteACPAction(ctx, req)
        if err == nil && result.Success {
            return result, nil
        }

        if err != nil {
            lastErr = err
            fmt.Printf("Attempt %d failed: %v\n", attempt+1, err)
        } else if !result.Success {
            lastErr = fmt.Errorf("action failed: %s", result.Error)
            fmt.Printf("Attempt %d failed: %s\n", attempt+1, result.Error)
        }
    }

    return nil, fmt.Errorf("all retries failed: %w", lastErr)
}
```

### Example 7: ACP Statistics and Monitoring

Monitor ACP server usage and statistics:

```go
func monitorACPStats(ctx context.Context, acpManager *services.ACPManager) error {
    // Get overall statistics
    stats, err := acpManager.GetACPStats(ctx)
    if err != nil {
        return err
    }

    fmt.Println("ACP Statistics:")
    fmt.Printf("  Total Servers: %v\n", stats["totalServers"])
    fmt.Printf("  Enabled Servers: %v\n", stats["enabledServers"])
    fmt.Printf("  Total Capabilities: %v\n", stats["totalCapabilities"])
    fmt.Printf("  Last Sync: %v\n", stats["lastSync"])

    // List all servers with status
    servers, err := acpManager.ListACPServers(ctx)
    if err != nil {
        return err
    }

    fmt.Println("\nServer Status:")
    for _, server := range servers {
        status := "disabled"
        if server.Enabled {
            status = "enabled"
        }
        fmt.Printf("  %s (%s): %s\n", server.Name, server.ID, status)
        if server.LastSync != nil {
            fmt.Printf("    Last sync: %v\n", server.LastSync)
        }
    }

    return nil
}
```

### Example 8: Test Generation Through ACP

Use ACP to generate unit tests:

```go
func generateTests(ctx context.Context, acpManager *services.ACPManager) error {
    // Request test generation
    result, err := acpManager.ExecuteACPAction(ctx, services.ACPRequest{
        ServerID: "code-agent",
        Action:   "generate_tests",
        Parameters: map[string]interface{}{
            "file_path":  "/workspace/src/calculator.go",
            "test_type":  "unit",
            "coverage":   0.8,
            "framework":  "testify",
            "mock_style": "gomock",
        },
    })
    if err != nil {
        return err
    }

    if !result.Success {
        return fmt.Errorf("test generation failed: %s", result.Error)
    }

    // Extract generated tests
    tests := result.Data.(map[string]interface{})["tests"].(string)
    fmt.Printf("Generated tests:\n%s\n", tests)

    // Write to file
    testPath := "/workspace/src/calculator_test.go"
    err = os.WriteFile(testPath, []byte(tests), 0644)
    if err != nil {
        return err
    }

    fmt.Printf("Tests written to: %s\n", testPath)
    return nil
}
```

### Example 9: Code Refactoring Through ACP

Request code refactoring suggestions:

```go
func refactorCode(ctx context.Context, acpManager *services.ACPManager) error {
    result, err := acpManager.ExecuteACPAction(ctx, services.ACPRequest{
        ServerID: "code-agent",
        Action:   "refactor",
        Parameters: map[string]interface{}{
            "file_path": "/workspace/src/legacy_handler.go",
            "refactoring_types": []string{
                "extract_method",
                "rename_variable",
                "simplify_conditionals",
                "remove_duplication",
            },
            "preserve_behavior": true,
        },
    })
    if err != nil {
        return err
    }

    suggestions := result.Data.([]interface{})
    fmt.Printf("Refactoring suggestions (%d):\n", len(suggestions))

    for i, s := range suggestions {
        suggestion := s.(map[string]interface{})
        fmt.Printf("\n%d. %s\n", i+1, suggestion["type"])
        fmt.Printf("   Description: %s\n", suggestion["description"])
        fmt.Printf("   Impact: %s\n", suggestion["impact"])
        fmt.Printf("   Diff:\n%s\n", suggestion["diff"])
    }

    return nil
}
```

### Example 10: Multi-Agent Coordination

Coordinate multiple ACP agents for complex tasks:

```go
func coordinateAgents(ctx context.Context, acpManager *services.ACPManager) error {
    // Step 1: Code analysis agent identifies issues
    analysisResult, err := acpManager.ExecuteACPAction(ctx, services.ACPRequest{
        ServerID: "code-agent",
        Action:   "analyze_code",
        Parameters: map[string]interface{}{
            "directory": "/workspace/src",
            "recursive": true,
        },
    })
    if err != nil {
        return err
    }

    issues := analysisResult.Data.(map[string]interface{})["issues"].([]interface{})
    fmt.Printf("Found %d issues\n", len(issues))

    // Step 2: For each issue, get fix suggestions
    for _, issue := range issues {
        issueData := issue.(map[string]interface{})

        fixResult, err := acpManager.ExecuteACPAction(ctx, services.ACPRequest{
            ServerID: "code-agent",
            Action:   "suggest_fix",
            Parameters: map[string]interface{}{
                "issue_type": issueData["type"],
                "file_path":  issueData["file"],
                "line":       issueData["line"],
            },
        })
        if err != nil {
            fmt.Printf("Failed to get fix for issue: %v\n", err)
            continue
        }

        fmt.Printf("Fix for %s at %s:%v:\n", issueData["type"], issueData["file"], issueData["line"])
        fmt.Printf("  %v\n", fixResult.Data)
    }

    // Step 3: Generate tests for modified files
    modifiedFiles := extractModifiedFiles(issues)
    for _, file := range modifiedFiles {
        testResult, err := acpManager.ExecuteACPAction(ctx, services.ACPRequest{
            ServerID: "code-agent",
            Action:   "generate_tests",
            Parameters: map[string]interface{}{
                "file_path": file,
                "test_type": "unit",
            },
        })
        if err != nil {
            fmt.Printf("Failed to generate tests for %s: %v\n", file, err)
            continue
        }

        fmt.Printf("Generated tests for %s\n", file)
        fmt.Printf("  %v\n", testResult.Data)
    }

    return nil
}

func extractModifiedFiles(issues []interface{}) []string {
    files := make(map[string]bool)
    for _, issue := range issues {
        issueData := issue.(map[string]interface{})
        files[issueData["file"].(string)] = true
    }

    result := make([]string, 0, len(files))
    for file := range files {
        result = append(result, file)
    }
    return result
}
```

---

## Summary

This document provided 30+ advanced examples across the three protocols supported by HelixAgent:

- **LSP Examples**: Multi-language workspaces, advanced completions, go-to-definition, hover information, document synchronization, references, symbol search, code actions, formatting, and renaming.

- **MCP Examples**: Multi-server configuration, tool discovery and invocation, resource management, prompt templates, database queries, GitHub operations, filesystem operations, tool chaining, health monitoring, and custom tool creation.

- **ACP Examples**: Server registration and discovery, action execution, real-time WebSocket communication, server synchronization, batch requests, error handling with retry, statistics and monitoring, test generation, code refactoring, and multi-agent coordination.

For more information, see the [HelixAgent documentation](https://dev.helix.agent).
