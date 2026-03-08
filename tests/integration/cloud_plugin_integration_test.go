package integration

import (
	"context"
	"testing"
	"time"

	"dev.helix.agent/internal/plugins"
	"dev.helix.agent/internal/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPluginSystemIntegration tests the plugin system integration
func TestPluginSystemIntegration(t *testing.T) {
	// Integration test — no external deps required

	t.Run("Registry and Loader integration", func(t *testing.T) {
		registry := plugins.NewRegistry()
		loader := plugins.NewLoader(registry)

		// Verify loader is connected to registry
		assert.NotNil(t, loader)
		assert.NotNil(t, registry)

		// Get registered plugins (should be empty initially)
		registeredPlugins := registry.List()
		assert.Empty(t, registeredPlugins)
	})

	t.Run("Registry plugin operations", func(t *testing.T) {
		registry := plugins.NewRegistry()

		// Test that registry is empty initially
		pluginList := registry.List()
		assert.Empty(t, pluginList)

		// Test getting a non-existent plugin
		plugin, found := registry.Get("non-existent")
		assert.Nil(t, plugin)
		assert.False(t, found)
	})

	t.Run("Loader initialization", func(t *testing.T) {
		registry := plugins.NewRegistry()
		loader := plugins.NewLoader(registry)

		// Verify loader can be created
		assert.NotNil(t, loader)

		// Test loading a non-existent plugin path
		_, err := loader.Load("/nonexistent/path/plugin.so")
		assert.Error(t, err)
	})
}

// TestServiceToolIntegration tests service and tool integration
func TestServiceToolIntegration(t *testing.T) {
	// Integration test — no external deps required

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("MCP and LSP with ToolRegistry", func(t *testing.T) {
		// Create MCP manager
		mcpManager := services.NewMCPManager(nil, nil, nil)

		// Create LSP client
		lspClient := services.NewLSPClient(logger)

		// Create tool registry
		toolRegistry := services.NewToolRegistry(mcpManager, lspClient)

		// Verify components are connected
		assert.NotNil(t, toolRegistry)

		// Register a custom tool
		customTool := &IntegrationTestTool{
			name:        "integration-tool",
			description: "Tool for integration testing",
			parameters:  map[string]interface{}{"input": map[string]interface{}{"type": "string"}},
		}

		err := toolRegistry.RegisterCustomTool(customTool)
		require.NoError(t, err)

		// List tools and check if ours is registered
		tools := toolRegistry.ListTools()
		found := false
		for _, tool := range tools {
			if tool.Name() == "integration-tool" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find integration-tool in registered tools")
	})

	t.Run("ContextManager with multiple sources", func(t *testing.T) {
		contextManager := services.NewContextManager(100)

		// Add entries from different sources
		sources := []struct {
			id       string
			typ      string
			source   string
			priority int
		}{
			{"lsp-1", "lsp", "gopls", 9},
			{"mcp-1", "mcp", "filesystem", 8},
			{"tool-1", "tool", "grep", 7},
			{"memory-1", "memory", "conversation", 6},
			{"system-1", "system", "config", 5},
		}

		for _, src := range sources {
			err := contextManager.AddEntry(&services.ContextEntry{
				ID:       src.id,
				Type:     src.typ,
				Source:   src.source,
				Content:  "Context from " + src.source,
				Priority: src.priority,
			})
			require.NoError(t, err)
		}

		// Build context for different request types
		requestTypes := []string{"code_completion", "tool_execution", "chat"}

		for _, reqType := range requestTypes {
			ctx, err := contextManager.BuildContext(reqType, 1000)
			require.NoError(t, err)
			assert.NotEmpty(t, ctx, "Context for %s should not be empty", reqType)
		}
	})

	t.Run("SecuritySandbox with tool validation", func(t *testing.T) {
		sandbox := services.NewSecuritySandbox()

		// Test safe parameters
		safeParams := map[string]interface{}{
			"file":   "/tmp/test.txt",
			"action": "read",
			"query":  "search term",
		}

		err := sandbox.ValidateToolExecution("file-reader", safeParams)
		require.NoError(t, err)

		// Test dangerous parameters
		dangerousParams := map[string]interface{}{
			"command": "rm -rf /",
			"script":  "; DROP TABLE users;--",
		}

		err = sandbox.ValidateToolExecution("executor", dangerousParams)
		assert.Error(t, err, "Should reject dangerous parameters")
	})
}

// TestContextAndCacheIntegration tests context manager caching integration
func TestContextAndCacheIntegration(t *testing.T) {
	// Integration test — no external deps required

	t.Run("Cache results with context", func(t *testing.T) {
		contextManager := services.NewContextManager(100)

		// Cache some results
		testData := map[string]interface{}{
			"model":    "test-model",
			"response": "cached response",
			"tokens":   100,
		}

		contextManager.CacheResult("test-request-1", testData, 5*time.Minute)

		// Retrieve cached result
		cached, found := contextManager.GetCachedResult("test-request-1")
		assert.True(t, found)
		assert.Equal(t, testData, cached)

		// Non-existent key
		_, found = contextManager.GetCachedResult("non-existent")
		assert.False(t, found)
	})

	t.Run("Context with priority sorting", func(t *testing.T) {
		contextManager := services.NewContextManager(100)

		// Add entries with different priorities
		entries := []struct {
			id       string
			priority int
		}{
			{"low-priority", 1},
			{"high-priority", 10},
			{"medium-priority", 5},
		}

		for _, e := range entries {
			err := contextManager.AddEntry(&services.ContextEntry{
				ID:       e.id,
				Type:     "test",
				Source:   "test",
				Content:  "Content for " + e.id,
				Priority: e.priority,
			})
			require.NoError(t, err)
		}

		// Build context and verify ordering
		ctx, err := contextManager.BuildContext("chat", 1000)
		require.NoError(t, err)
		assert.NotEmpty(t, ctx)

		// Higher priority entries should appear first
		if len(ctx) > 1 {
			// First entry should have higher or equal priority than second
			assert.GreaterOrEqual(t, ctx[0].Priority, ctx[1].Priority)
		}
	})
}

// IntegrationTestTool is a mock tool for integration testing
type IntegrationTestTool struct {
	name        string
	description string
	parameters  map[string]interface{}
}

func (t *IntegrationTestTool) Name() string {
	return t.name
}

func (t *IntegrationTestTool) Description() string {
	return t.description
}

func (t *IntegrationTestTool) Parameters() map[string]interface{} {
	return t.parameters
}

func (t *IntegrationTestTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"tool":   t.name,
		"params": params,
		"result": "executed",
	}, nil
}

func (t *IntegrationTestTool) Source() string {
	return "integration-test"
}
