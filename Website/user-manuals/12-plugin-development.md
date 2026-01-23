# User Manual 12: Plugin Development Guide

## Introduction

HelixAgent supports a powerful plugin system that allows extending functionality through custom plugins. This guide covers plugin development, lifecycle management, and best practices.

## Plugin Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Plugin System                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Plugin    │  │   Plugin    │  │      Plugin             │ │
│  │   Loader    │  │  Registry   │  │     Manager             │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘ │
│         │                │                      │               │
│         ▼                ▼                      ▼               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Plugin Interface                       │  │
│  │  Initialize() │ Execute() │ Shutdown() │ HealthCheck()   │  │
│  └──────────────────────────────────────────────────────────┘  │
│         │                │                      │               │
│         ▼                ▼                      ▼               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Custom    │  │   Custom    │  │        Custom           │ │
│  │   Plugin A  │  │   Plugin B  │  │        Plugin C         │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Create Plugin Structure

```bash
mkdir -p plugins/my-plugin
cd plugins/my-plugin
```

```
my-plugin/
├── plugin.yaml          # Plugin manifest
├── main.go             # Plugin entry point
├── handlers.go         # Request handlers
├── config.go          # Configuration
└── README.md          # Documentation
```

### 2. Define Plugin Manifest

```yaml
# plugin.yaml
name: my-plugin
version: 1.0.0
description: My custom plugin
author: Your Name

# Plugin type
type: extension

# Dependencies
dependencies:
  helixagent: ">=1.0.0"

# Configuration schema
config:
  api_key:
    type: string
    required: true
    secret: true
  timeout:
    type: duration
    default: 30s

# Hooks to implement
hooks:
  - pre_completion
  - post_completion

# API endpoints
endpoints:
  - path: /v1/my-plugin/action
    method: POST
    handler: HandleAction
```

### 3. Implement Plugin Interface

```go
// main.go
package main

import (
    "context"
    "dev.helix.agent/internal/plugins"
)

// MyPlugin implements the Plugin interface
type MyPlugin struct {
    config *Config
    logger plugins.Logger
}

// New creates a new plugin instance
func New() plugins.Plugin {
    return &MyPlugin{}
}

// Info returns plugin metadata
func (p *MyPlugin) Info() plugins.PluginInfo {
    return plugins.PluginInfo{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "My custom plugin",
        Author:      "Your Name",
    }
}

// Initialize sets up the plugin
func (p *MyPlugin) Initialize(ctx context.Context, cfg plugins.Config) error {
    p.config = &Config{
        APIKey:  cfg.GetString("api_key"),
        Timeout: cfg.GetDuration("timeout"),
    }
    p.logger = cfg.Logger()
    p.logger.Info("Plugin initialized")
    return nil
}

// Shutdown cleans up resources
func (p *MyPlugin) Shutdown(ctx context.Context) error {
    p.logger.Info("Plugin shutting down")
    return nil
}

// HealthCheck verifies plugin is working
func (p *MyPlugin) HealthCheck(ctx context.Context) error {
    // Verify dependencies, connections, etc.
    return nil
}

// Export the plugin
var Plugin MyPlugin
```

### 4. Build and Install

```bash
# Build plugin
go build -buildmode=plugin -o my-plugin.so .

# Install plugin
cp my-plugin.so /path/to/helixagent/plugins/

# Or via CLI
./helixagent plugins install ./my-plugin.so
```

## Plugin Types

### Extension Plugin

Adds new functionality:

```go
// Implements custom endpoints
func (p *MyPlugin) RegisterRoutes(router plugins.Router) {
    router.POST("/v1/my-plugin/action", p.HandleAction)
    router.GET("/v1/my-plugin/status", p.HandleStatus)
}

func (p *MyPlugin) HandleAction(ctx plugins.Context) error {
    var req ActionRequest
    if err := ctx.Bind(&req); err != nil {
        return err
    }

    result, err := p.processAction(ctx.Context(), req)
    if err != nil {
        return ctx.Error(500, err)
    }

    return ctx.JSON(200, result)
}
```

### Hook Plugin

Intercepts request/response flow:

```go
// PreCompletion hook - runs before LLM completion
func (p *MyPlugin) PreCompletion(ctx context.Context, req *plugins.CompletionRequest) (*plugins.CompletionRequest, error) {
    // Modify request
    req.Prompt = p.preprocessPrompt(req.Prompt)

    // Add metadata
    req.Metadata["plugin"] = "my-plugin"

    return req, nil
}

// PostCompletion hook - runs after LLM completion
func (p *MyPlugin) PostCompletion(ctx context.Context, req *plugins.CompletionRequest, resp *plugins.CompletionResponse) (*plugins.CompletionResponse, error) {
    // Modify response
    resp.Content = p.postprocessContent(resp.Content)

    // Log metrics
    p.logger.Info("Completion processed",
        "tokens", resp.Usage.TotalTokens,
        "latency", resp.Latency)

    return resp, nil
}
```

### Provider Plugin

Adds new LLM providers:

```go
// Implements LLMProvider interface
type CustomProvider struct {
    config *ProviderConfig
}

func (p *CustomProvider) Complete(ctx context.Context, req *llm.Request) (*llm.Response, error) {
    // Implement completion logic
    return p.client.Complete(ctx, req)
}

func (p *CustomProvider) CompleteStream(ctx context.Context, req *llm.Request) (<-chan *llm.StreamResponse, error) {
    // Implement streaming
    return p.client.Stream(ctx, req)
}

func (p *CustomProvider) GetCapabilities() llm.Capabilities {
    return llm.Capabilities{
        Streaming:    true,
        ToolUse:      true,
        Vision:       false,
        MaxTokens:    100000,
    }
}

// Register provider
func (p *MyPlugin) RegisterProviders(registry plugins.ProviderRegistry) {
    registry.Register("custom-provider", &CustomProvider{})
}
```

### Tool Plugin

Adds new MCP tools:

```go
// Implements Tool interface
type CustomTool struct {
    client *http.Client
}

func (t *CustomTool) Info() plugins.ToolInfo {
    return plugins.ToolInfo{
        Name:        "custom_search",
        Description: "Search using custom API",
        Parameters: map[string]plugins.Parameter{
            "query": {
                Type:        "string",
                Description: "Search query",
                Required:    true,
            },
            "limit": {
                Type:        "integer",
                Description: "Max results",
                Default:     10,
            },
        },
    }
}

func (t *CustomTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    query := params["query"].(string)
    limit := params["limit"].(int)

    results, err := t.search(ctx, query, limit)
    if err != nil {
        return nil, err
    }

    return results, nil
}

// Register tool
func (p *MyPlugin) RegisterTools(registry plugins.ToolRegistry) {
    registry.Register(&CustomTool{})
}
```

## Plugin Configuration

### Configuration Schema

```yaml
# plugin.yaml
config:
  # Simple types
  api_key:
    type: string
    required: true
    secret: true
    env: MY_PLUGIN_API_KEY

  timeout:
    type: duration
    default: 30s

  max_retries:
    type: integer
    default: 3
    min: 1
    max: 10

  # Complex types
  endpoints:
    type: array
    items:
      type: object
      properties:
        url:
          type: string
        priority:
          type: integer

  # Nested config
  auth:
    type: object
    properties:
      type:
        type: string
        enum: ["api_key", "oauth", "basic"]
      credentials:
        type: object
```

### Accessing Configuration

```go
func (p *MyPlugin) Initialize(ctx context.Context, cfg plugins.Config) error {
    // Get simple values
    apiKey := cfg.GetString("api_key")
    timeout := cfg.GetDuration("timeout")
    maxRetries := cfg.GetInt("max_retries")

    // Get nested values
    authType := cfg.GetString("auth.type")
    credentials := cfg.GetStringMap("auth.credentials")

    // Get with defaults
    verbose := cfg.GetBool("verbose", false)

    return nil
}
```

## Plugin Lifecycle

### Startup Sequence

```
1. Load plugin manifest
2. Validate dependencies
3. Load plugin binary
4. Call Initialize()
5. Register routes/hooks/providers
6. Run HealthCheck()
7. Plugin ready
```

### Shutdown Sequence

```
1. Stop accepting requests
2. Wait for active requests
3. Call Shutdown()
4. Unregister all components
5. Unload binary
```

### Hot Reload

```yaml
# Enable hot reload in development
plugins:
  hot_reload:
    enabled: true
    watch_dir: ./plugins
    debounce: 1s
```

```bash
# Manually reload plugin
./helixagent plugins reload my-plugin
```

## Plugin Communication

### Inter-Plugin Communication

```go
// Get another plugin's API
func (p *MyPlugin) Initialize(ctx context.Context, cfg plugins.Config) error {
    // Get plugin registry
    registry := cfg.PluginRegistry()

    // Get another plugin
    otherPlugin, err := registry.Get("other-plugin")
    if err != nil {
        return err
    }

    // Call plugin API
    if callable, ok := otherPlugin.(CallablePlugin); ok {
        result, err := callable.Call(ctx, "method", params)
    }

    return nil
}
```

### Event System

```go
// Subscribe to events
func (p *MyPlugin) Initialize(ctx context.Context, cfg plugins.Config) error {
    events := cfg.EventBus()

    // Subscribe
    events.Subscribe("completion.created", p.onCompletionCreated)
    events.Subscribe("user.authenticated", p.onUserAuthenticated)

    return nil
}

// Handle events
func (p *MyPlugin) onCompletionCreated(event plugins.Event) {
    data := event.Data.(CompletionEvent)
    p.logger.Info("Completion created", "id", data.ID)
}

// Publish events
func (p *MyPlugin) DoSomething(ctx context.Context) {
    p.events.Publish("my-plugin.action", ActionEvent{
        Type:      "process",
        Timestamp: time.Now(),
    })
}
```

## Testing Plugins

### Unit Tests

```go
// plugin_test.go
func TestMyPlugin_Initialize(t *testing.T) {
    plugin := New()

    cfg := plugins.NewMockConfig(map[string]interface{}{
        "api_key":  "test-key",
        "timeout":  "30s",
    })

    err := plugin.Initialize(context.Background(), cfg)
    assert.NoError(t, err)
}

func TestMyPlugin_PreCompletion(t *testing.T) {
    plugin := setupPlugin(t)

    req := &plugins.CompletionRequest{
        Prompt: "Hello",
    }

    result, err := plugin.PreCompletion(context.Background(), req)
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Prompt)
}
```

### Integration Tests

```go
func TestMyPlugin_Integration(t *testing.T) {
    // Start test server with plugin
    srv := plugins.NewTestServer(t)
    srv.LoadPlugin("./my-plugin.so")
    srv.Start()
    defer srv.Stop()

    // Test plugin endpoint
    resp, err := http.Post(
        srv.URL+"/v1/my-plugin/action",
        "application/json",
        strings.NewReader(`{"input": "test"}`),
    )
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Best Practices

### 1. Error Handling

```go
// Good: Wrap errors with context
func (p *MyPlugin) process(ctx context.Context, data string) error {
    result, err := p.client.Call(ctx, data)
    if err != nil {
        return fmt.Errorf("my-plugin: failed to process: %w", err)
    }
    return nil
}

// Good: Use typed errors
var (
    ErrInvalidInput = errors.New("my-plugin: invalid input")
    ErrAPIFailure   = errors.New("my-plugin: API failure")
)
```

### 2. Logging

```go
// Good: Structured logging with context
func (p *MyPlugin) HandleRequest(ctx plugins.Context) error {
    p.logger.Info("Processing request",
        "request_id", ctx.RequestID(),
        "user_id", ctx.UserID(),
        "action", "process")

    // ... processing ...

    p.logger.Debug("Request completed",
        "request_id", ctx.RequestID(),
        "duration_ms", time.Since(start).Milliseconds())

    return nil
}
```

### 3. Resource Management

```go
// Good: Clean up resources
func (p *MyPlugin) Initialize(ctx context.Context, cfg plugins.Config) error {
    // Create connection pool
    p.pool = NewConnectionPool(10)

    // Start background worker
    p.worker = NewWorker()
    go p.worker.Run(ctx)

    return nil
}

func (p *MyPlugin) Shutdown(ctx context.Context) error {
    // Stop worker gracefully
    p.worker.Stop()

    // Close connections
    p.pool.Close()

    return nil
}
```

### 4. Configuration Validation

```go
func (p *MyPlugin) Initialize(ctx context.Context, cfg plugins.Config) error {
    // Validate required config
    apiKey := cfg.GetString("api_key")
    if apiKey == "" {
        return errors.New("my-plugin: api_key is required")
    }

    timeout := cfg.GetDuration("timeout")
    if timeout < time.Second {
        return errors.New("my-plugin: timeout must be at least 1s")
    }

    return nil
}
```

## Plugin Distribution

### Packaging

```bash
# Build for distribution
GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o my-plugin-linux-amd64.so .
GOOS=darwin GOARCH=amd64 go build -buildmode=plugin -o my-plugin-darwin-amd64.so .

# Create release archive
tar -czvf my-plugin-1.0.0.tar.gz \
    my-plugin-*.so \
    plugin.yaml \
    README.md
```

### Publishing

```bash
# Publish to plugin registry
./helixagent plugins publish ./my-plugin-1.0.0.tar.gz

# Install from registry
./helixagent plugins install my-plugin@1.0.0
```

## CLI Commands

```bash
# List installed plugins
./helixagent plugins list

# Install plugin
./helixagent plugins install <path-or-url>

# Uninstall plugin
./helixagent plugins uninstall <name>

# Reload plugin
./helixagent plugins reload <name>

# View plugin info
./helixagent plugins info <name>

# Plugin health check
./helixagent plugins health <name>
```

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
