# SuperAgent Plugin Development Guide

## Overview

SuperAgent's plugin system enables extensibility through hot-reloadable plugins that can add new LLM providers, custom tools, and specialized functionality. This guide covers everything you need to know to develop, test, and deploy plugins for SuperAgent.

## Plugin Architecture

### Core Concepts

SuperAgent plugins are Go packages compiled as shared libraries (`.so` files) that implement predefined interfaces. The plugin system provides:

- **Hot Reloading**: Plugins can be loaded, unloaded, and reloaded at runtime without restarting the server
- **Dependency Resolution**: Automatic handling of plugin dependencies with cycle detection
- **Health Monitoring**: Built-in health checks and metrics collection
- **Security Contexts**: Sandboxed execution with configurable permissions
- **Versioning**: Semantic versioning support for compatibility management

### System Architecture

```
+-------------------+     +--------------------+     +------------------+
|   Plugin Loader   |---->|  Plugin Registry   |---->|  Plugin Manager  |
+-------------------+     +--------------------+     +------------------+
         |                        |                         |
         v                        v                         v
+-------------------+     +--------------------+     +------------------+
|  .so Shared Libs  |     |  Plugin Metadata   |     |  Health Monitor  |
+-------------------+     +--------------------+     +------------------+
```

## Interface Implementation Guide

### LLMPlugin Interface

The primary interface for LLM provider plugins:

```go
package plugins

import (
    "context"
    "github.com/superagent/superagent/internal/models"
)

// LLMPlugin defines the interface for LLM provider plugins
type LLMPlugin interface {
    // Name returns the unique identifier for this plugin
    Name() string

    // Version returns the semantic version of the plugin
    Version() string

    // Capabilities returns the plugin's supported features and limits
    Capabilities() *models.ProviderCapabilities

    // Init initializes the plugin with configuration
    Init(config map[string]interface{}) error

    // Shutdown gracefully stops the plugin
    Shutdown(ctx context.Context) error

    // HealthCheck verifies the plugin is functioning correctly
    HealthCheck(ctx context.Context) error

    // SetSecurityContext configures the security context for plugin operations
    SetSecurityContext(ctx *PluginSecurityContext) error

    // Complete processes a completion request
    Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)

    // CompleteStream processes a streaming completion request
    CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error)
}
```

### PluginSecurityContext

The security context provides sandboxed access controls:

```go
type PluginSecurityContext struct {
    // PluginID uniquely identifies the plugin instance
    PluginID string

    // AllowedPaths defines filesystem paths the plugin can access
    AllowedPaths []string

    // AllowedHosts defines network hosts the plugin can connect to
    AllowedHosts []string

    // MaxMemoryMB limits memory usage
    MaxMemoryMB int

    // MaxCPUPercent limits CPU usage
    MaxCPUPercent int

    // Timeout defines maximum execution time
    Timeout time.Duration

    // Permissions defines specific capability flags
    Permissions PluginPermissions
}

type PluginPermissions struct {
    AllowNetworkAccess  bool
    AllowFileSystem     bool
    AllowExec           bool
    AllowEnvironment    bool
}
```

### ProviderCapabilities Structure

Define your plugin's capabilities accurately:

```go
type ProviderCapabilities struct {
    // SupportedModels lists all model identifiers
    SupportedModels []string

    // SupportedFeatures lists feature flags
    SupportedFeatures []string

    // SupportedRequestTypes defines supported request types
    SupportedRequestTypes []string

    // Feature flags
    SupportsStreaming       bool
    SupportsFunctionCalling bool
    SupportsVision          bool

    // Limits defines operational constraints
    Limits ModelLimits

    // Metadata contains additional plugin information
    Metadata map[string]string
}

type ModelLimits struct {
    MaxTokens             int
    MaxInputLength        int
    MaxOutputLength       int
    MaxConcurrentRequests int
}
```

## Hot Reload System

### How It Works

The hot reload system monitors the plugin directory and automatically handles plugin lifecycle events:

1. **Discovery**: The plugin loader scans configured directories for `.so` files
2. **Loading**: Plugins are loaded using Go's `plugin` package
3. **Initialization**: The `Init()` method is called with configuration
4. **Registration**: Plugins are registered in the plugin registry
5. **Monitoring**: File system watchers detect changes
6. **Reloading**: Modified plugins are gracefully reloaded

### Configuration

Configure hot reloading in your configuration file:

```yaml
plugins:
  enabled: true
  directory: "./plugins"
  hot_reload: true
  scan_interval: 5s
  max_load_time: 30s
  health_check_interval: 60s

  # Security settings
  security:
    sandbox_enabled: true
    default_timeout: 30s
    max_memory_mb: 512
    max_cpu_percent: 50
```

### File System Monitoring

The system watches for:
- New `.so` files: Triggers automatic loading
- Modified `.so` files: Triggers graceful reload
- Deleted `.so` files: Triggers unloading
- Configuration changes: Triggers reinitialization

## Step-by-Step Tutorial

### Step 1: Create Plugin Structure

Create your plugin directory:

```bash
mkdir -p plugins/myprovider
cd plugins/myprovider
```

Create the main plugin file `plugin.go`:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/superagent/superagent/internal/models"
    "github.com/superagent/superagent/internal/plugins"
)

// Plugin is the exported plugin instance - REQUIRED
var Plugin plugins.LLMPlugin = &MyProviderPlugin{}

// MyProviderPlugin implements the LLMPlugin interface
type MyProviderPlugin struct {
    config     map[string]interface{}
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

func (p *MyProviderPlugin) Name() string {
    return "myprovider"
}

func (p *MyProviderPlugin) Version() string {
    return "1.0.0"
}

func (p *MyProviderPlugin) Capabilities() *models.ProviderCapabilities {
    return &models.ProviderCapabilities{
        SupportedModels:         []string{"myprovider-base", "myprovider-pro"},
        SupportedFeatures:       []string{"streaming", "function_calling"},
        SupportedRequestTypes:   []string{"chat", "completion"},
        SupportsStreaming:       true,
        SupportsFunctionCalling: true,
        SupportsVision:          false,
        Limits: models.ModelLimits{
            MaxTokens:             8192,
            MaxInputLength:        4096,
            MaxOutputLength:       4096,
            MaxConcurrentRequests: 20,
        },
        Metadata: map[string]string{
            "author":      "Your Name",
            "license":     "MIT",
            "description": "Custom LLM provider integration",
        },
    }
}

func (p *MyProviderPlugin) Init(config map[string]interface{}) error {
    p.config = config

    // Extract API key from config
    if apiKey, ok := config["api_key"].(string); ok {
        p.apiKey = apiKey
    } else {
        return fmt.Errorf("api_key is required")
    }

    // Extract base URL with default
    p.baseURL = "https://api.myprovider.com/v1"
    if url, ok := config["base_url"].(string); ok {
        p.baseURL = url
    }

    // Initialize HTTP client
    p.httpClient = &http.Client{
        Timeout: 60 * time.Second,
    }

    return nil
}

func (p *MyProviderPlugin) Shutdown(ctx context.Context) error {
    // Cleanup resources
    if p.httpClient != nil {
        p.httpClient.CloseIdleConnections()
    }
    return nil
}

func (p *MyProviderPlugin) HealthCheck(ctx context.Context) error {
    // Implement provider-specific health check
    req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/health", nil)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+p.apiKey)

    resp, err := p.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check returned status %d", resp.StatusCode)
    }

    return nil
}

func (p *MyProviderPlugin) SetSecurityContext(ctx *plugins.PluginSecurityContext) error {
    // Validate security constraints
    if !ctx.Permissions.AllowNetworkAccess {
        return fmt.Errorf("network access is required for this plugin")
    }
    return nil
}
```

### Step 2: Implement Completion Methods

Add the `Complete` method:

```go
func (p *MyProviderPlugin) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
    startTime := time.Now()

    // Build API request
    apiReq := map[string]interface{}{
        "model":       req.Model,
        "prompt":      req.Prompt,
        "max_tokens":  req.MaxTokens,
        "temperature": req.Temperature,
    }

    jsonData, err := json.Marshal(apiReq)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/completions", bytes.NewReader(jsonData))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := p.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("API request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
    }

    var apiResp struct {
        ID      string `json:"id"`
        Content string `json:"content"`
        Usage   struct {
            TotalTokens int `json:"total_tokens"`
        } `json:"usage"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return &models.LLMResponse{
        RequestID:    req.ID,
        ProviderID:   "myprovider",
        ProviderName: "myprovider",
        Content:      apiResp.Content,
        Confidence:   0.85,
        TokensUsed:   apiResp.Usage.TotalTokens,
        ResponseTime: time.Since(startTime).Milliseconds(),
        FinishReason: "stop",
        CreatedAt:    time.Now(),
    }, nil
}
```

Add the `CompleteStream` method:

```go
func (p *MyProviderPlugin) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
    ch := make(chan *models.LLMResponse)

    go func() {
        defer close(ch)

        // Build streaming request
        apiReq := map[string]interface{}{
            "model":       req.Model,
            "prompt":      req.Prompt,
            "max_tokens":  req.MaxTokens,
            "temperature": req.Temperature,
            "stream":      true,
        }

        jsonData, err := json.Marshal(apiReq)
        if err != nil {
            return
        }

        httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/completions", bytes.NewReader(jsonData))
        if err != nil {
            return
        }

        httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Set("Accept", "text/event-stream")

        resp, err := p.httpClient.Do(httpReq)
        if err != nil {
            return
        }
        defer resp.Body.Close()

        reader := bufio.NewReader(resp.Body)
        var tokenCount int
        startTime := time.Now()

        for {
            select {
            case <-ctx.Done():
                return
            default:
                line, err := reader.ReadString('\n')
                if err != nil {
                    // Send final response
                    ch <- &models.LLMResponse{
                        RequestID:    req.ID,
                        ProviderID:   "myprovider",
                        ProviderName: "myprovider",
                        Content:      "",
                        TokensUsed:   tokenCount,
                        ResponseTime: time.Since(startTime).Milliseconds(),
                        FinishReason: "stop",
                        CreatedAt:    time.Now(),
                    }
                    return
                }

                // Parse SSE data
                if strings.HasPrefix(line, "data: ") {
                    data := strings.TrimPrefix(line, "data: ")
                    data = strings.TrimSpace(data)

                    if data == "[DONE]" {
                        continue
                    }

                    var chunk struct {
                        Content string `json:"content"`
                    }
                    if err := json.Unmarshal([]byte(data), &chunk); err != nil {
                        continue
                    }

                    tokenCount++
                    ch <- &models.LLMResponse{
                        RequestID:    req.ID,
                        ProviderID:   "myprovider",
                        ProviderName: "myprovider",
                        Content:      chunk.Content,
                        TokensUsed:   tokenCount,
                        ResponseTime: time.Since(startTime).Milliseconds(),
                        FinishReason: "",
                        CreatedAt:    time.Now(),
                    }
                }
            }
        }
    }()

    return ch, nil
}

// Required main function for plugin compilation
func main() {
    // Plugin entry point
    select {}
}
```

### Step 3: Build the Plugin

Create a `Makefile` for your plugin:

```makefile
.PHONY: build clean test

PLUGIN_NAME=myprovider
PLUGIN_VERSION=1.0.0

build:
	go build -buildmode=plugin -o $(PLUGIN_NAME).so plugin.go

clean:
	rm -f $(PLUGIN_NAME).so

test:
	go test -v ./...

install: build
	cp $(PLUGIN_NAME).so ../../plugins/
```

Build and install:

```bash
make build
make install
```

### Step 4: Configure Plugin Loading

Add your plugin configuration to SuperAgent:

```yaml
# configs/development.yaml
plugins:
  enabled: true
  directory: "./plugins"

  providers:
    myprovider:
      enabled: true
      api_key: "${MYPROVIDER_API_KEY}"
      base_url: "https://api.myprovider.com/v1"
      priority: 50
```

## Testing Plugins

### Unit Testing

Create comprehensive unit tests:

```go
package main

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/superagent/superagent/internal/models"
)

func TestPluginInit(t *testing.T) {
    plugin := &MyProviderPlugin{}

    // Test successful initialization
    config := map[string]interface{}{
        "api_key":  "test-key",
        "base_url": "https://test.api.com",
    }

    err := plugin.Init(config)
    require.NoError(t, err)
    assert.Equal(t, "test-key", plugin.apiKey)
    assert.Equal(t, "https://test.api.com", plugin.baseURL)
}

func TestPluginInitMissingAPIKey(t *testing.T) {
    plugin := &MyProviderPlugin{}

    config := map[string]interface{}{
        "base_url": "https://test.api.com",
    }

    err := plugin.Init(config)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "api_key is required")
}

func TestPluginCapabilities(t *testing.T) {
    plugin := &MyProviderPlugin{}
    caps := plugin.Capabilities()

    assert.Contains(t, caps.SupportedModels, "myprovider-base")
    assert.True(t, caps.SupportsStreaming)
    assert.Equal(t, 8192, caps.Limits.MaxTokens)
}

func TestPluginComplete(t *testing.T) {
    // Use httptest for mocking API responses
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/completions", r.URL.Path)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id":      "test-123",
            "content": "Test response",
            "usage": map[string]interface{}{
                "total_tokens": 10,
            },
        })
    }))
    defer server.Close()

    plugin := &MyProviderPlugin{}
    plugin.Init(map[string]interface{}{
        "api_key":  "test-key",
        "base_url": server.URL,
    })

    req := &models.LLMRequest{
        ID:          "req-1",
        Model:       "myprovider-base",
        Prompt:      "Hello",
        MaxTokens:   100,
        Temperature: 0.7,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := plugin.Complete(ctx, req)
    require.NoError(t, err)
    assert.Equal(t, "Test response", resp.Content)
    assert.Equal(t, 10, resp.TokensUsed)
}
```

### Integration Testing

Test your plugin with the actual plugin loader:

```go
func TestPluginIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    loader := plugins.NewPluginLoader("./test_plugins", logrus.New())

    // Load plugin
    err := loader.LoadPlugin("myprovider.so")
    require.NoError(t, err)

    // Get plugin instance
    plugin, err := loader.GetPlugin("myprovider")
    require.NoError(t, err)

    // Verify capabilities
    caps := plugin.Capabilities()
    assert.NotEmpty(t, caps.SupportedModels)

    // Test health check
    ctx := context.Background()
    err = plugin.HealthCheck(ctx)
    assert.NoError(t, err)
}
```

## Best Practices

### Error Handling

1. **Return Meaningful Errors**: Include context in error messages
   ```go
   return nil, fmt.Errorf("failed to connect to %s: %w", p.baseURL, err)
   ```

2. **Handle Context Cancellation**: Always respect context deadlines
   ```go
   select {
   case <-ctx.Done():
       return nil, ctx.Err()
   default:
       // Continue processing
   }
   ```

3. **Implement Graceful Degradation**: Handle partial failures
   ```go
   if err != nil {
       log.Warn("Primary endpoint failed, trying fallback")
       return p.fallbackComplete(ctx, req)
   }
   ```

### Performance Optimization

1. **Connection Pooling**: Reuse HTTP clients
   ```go
   // Initialize once in Init()
   p.httpClient = &http.Client{
       Transport: &http.Transport{
           MaxIdleConns:        100,
           MaxIdleConnsPerHost: 100,
           IdleConnTimeout:     90 * time.Second,
       },
   }
   ```

2. **Request Batching**: Batch small requests when possible

3. **Caching**: Implement response caching for repeated requests

### Security Considerations

1. **Validate Input**: Always validate request parameters
   ```go
   if req.MaxTokens > p.Capabilities().Limits.MaxTokens {
       return nil, fmt.Errorf("max_tokens exceeds limit of %d", p.Capabilities().Limits.MaxTokens)
   }
   ```

2. **Secure Credentials**: Never log or expose API keys
   ```go
   // BAD: log.Printf("Using API key: %s", p.apiKey)
   // GOOD: log.Printf("Using API key: %s***", p.apiKey[:4])
   ```

3. **Honor Security Context**: Respect all security constraints
   ```go
   if !secCtx.Permissions.AllowNetworkAccess {
       return nil, fmt.Errorf("network access not permitted")
   }
   ```

### Versioning

1. **Follow Semantic Versioning**: Use MAJOR.MINOR.PATCH
2. **Document Breaking Changes**: Update changelog for major versions
3. **Maintain Backwards Compatibility**: When possible, support older API versions

### Logging

1. **Use Structured Logging**: Include relevant fields
   ```go
   log.WithFields(logrus.Fields{
       "plugin":   p.Name(),
       "request":  req.ID,
       "model":    req.Model,
   }).Info("Processing request")
   ```

2. **Log Levels**: Use appropriate log levels
   - DEBUG: Detailed debugging information
   - INFO: General operational information
   - WARN: Recoverable issues
   - ERROR: Failures requiring attention

## Troubleshooting

### Common Issues

1. **Plugin Not Loading**
   - Verify the `.so` file exists in the plugins directory
   - Check file permissions
   - Ensure Go version compatibility
   - Verify the `Plugin` variable is exported

2. **Initialization Failures**
   - Check configuration values
   - Verify API credentials
   - Review log output for specific errors

3. **Health Check Failures**
   - Verify network connectivity
   - Check API endpoint availability
   - Review timeout settings

4. **Memory Issues**
   - Monitor memory usage with metrics
   - Implement proper resource cleanup in `Shutdown()`
   - Use connection pooling

### Debug Mode

Enable debug logging for plugins:

```yaml
logging:
  level: debug

plugins:
  debug: true
  verbose_errors: true
```

## Reference

### Plugin Registry API

```go
// Register a plugin
registry.Register(plugin)

// Get a plugin by name
plugin, err := registry.Get("myprovider")

// List all plugins
plugins := registry.List()

// Unregister a plugin
err := registry.Unregister("myprovider")
```

### Plugin Loader API

```go
// Create loader
loader := plugins.NewPluginLoader(directory, logger)

// Load all plugins
err := loader.LoadAll()

// Load specific plugin
err := loader.LoadPlugin("myprovider.so")

// Reload plugin
err := loader.ReloadPlugin("myprovider")

// Unload plugin
err := loader.UnloadPlugin("myprovider")
```

---

For additional support, see the [SuperAgent GitHub repository](https://github.com/superagent/superagent) or join our community Discord.
