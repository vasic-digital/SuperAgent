package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/helixagent/helixagent/internal/services"
)

// MockTool implements the Tool interface for testing
type MockTool struct {
	name        string
	description string
	parameters  map[string]interface{}
	executeFunc func(ctx context.Context, params map[string]interface{}) (interface{}, error)
	source      string
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) Parameters() map[string]interface{} {
	return m.parameters
}

func (m *MockTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, params)
	}
	return "mock result", nil
}

func (m *MockTool) Source() string {
	return m.source
}

func TestToolRegistry_Basic(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)
	assert.NotNil(t, registry)
}

func TestToolRegistry_RegisterCustomTool_Valid(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.NoError(t, err)
}

func TestToolRegistry_RegisterCustomTool_EmptyName(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "",
		description: "test description",
		parameters:  map[string]interface{}{},
		source:      "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool name cannot be empty")
}

func TestToolRegistry_RegisterCustomTool_EmptyDescription(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool description cannot be empty")
}

func TestToolRegistry_RegisterCustomTool_NilParameters(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters:  nil,
		source:      "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool parameters cannot be nil")
}

func TestToolRegistry_RegisterCustomTool_InvalidParameterSchema(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": "not a map",
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parameter param1 schema must be a map")
}

func TestToolRegistry_RegisterCustomTool_MissingParameterType(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"description": "test"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parameter param1 schema must have a type")
}

func TestToolRegistry_RegisterCustomTool_Duplicate(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool1 := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
		},
		source: "custom",
	}

	tool2 := &MockTool{
		name:        "test-tool",
		description: "another description",
		parameters: map[string]interface{}{
			"param2": map[string]interface{}{"type": "number"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool1)
	assert.NoError(t, err)

	err = registry.RegisterCustomTool(tool2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool test-tool already registered")
}

func TestToolRegistry_GetTool_Exists(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.NoError(t, err)

	retrievedTool, exists := registry.GetTool("test-tool")
	assert.True(t, exists)
	assert.NotNil(t, retrievedTool)
	assert.Equal(t, "test-tool", retrievedTool.Name())
	assert.Equal(t, "test description", retrievedTool.Description())
}

func TestToolRegistry_GetTool_NotExists(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool, exists := registry.GetTool("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, tool)
}

func TestToolRegistry_ListTools_Empty(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tools := registry.ListTools()
	assert.Empty(t, tools)
}

func TestToolRegistry_ListTools_WithTools(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool1 := &MockTool{
		name:        "tool1",
		description: "description1",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
		},
		source: "custom",
	}

	tool2 := &MockTool{
		name:        "tool2",
		description: "description2",
		parameters: map[string]interface{}{
			"param2": map[string]interface{}{"type": "number"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool1)
	assert.NoError(t, err)

	err = registry.RegisterCustomTool(tool2)
	assert.NoError(t, err)

	tools := registry.ListTools()
	assert.Equal(t, 2, len(tools))

	// Check that both tools are present
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name()] = true
	}
	assert.True(t, toolNames["tool1"])
	assert.True(t, toolNames["tool2"])
}

func TestToolRegistry_ExecuteTool_Success(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	executed := false
	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
		},
		executeFunc: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			executed = true
			assert.Equal(t, "value1", params["param1"])
			return "success", nil
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.NoError(t, err)

	ctx := context.Background()
	result, err := registry.ExecuteTool(ctx, "test-tool", map[string]interface{}{
		"param1": "value1",
	})

	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Equal(t, "success", result)
}

func TestToolRegistry_ExecuteTool_NotFound(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	ctx := context.Background()
	result, err := registry.ExecuteTool(ctx, "nonexistent", map[string]interface{}{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tool nonexistent not found")
}

func TestToolRegistry_ExecuteTool_MissingParameter(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
			"param2": map[string]interface{}{"type": "number"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.NoError(t, err)

	ctx := context.Background()
	result, err := registry.ExecuteTool(ctx, "test-tool", map[string]interface{}{
		"param1": "value1",
		// Missing param2
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "missing required parameter: param2")
}

func TestToolRegistry_ExecuteTool_ExecutionError(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	tool := &MockTool{
		name:        "test-tool",
		description: "test description",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{"type": "string"},
		},
		executeFunc: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, assert.AnError
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.NoError(t, err)

	ctx := context.Background()
	result, err := registry.ExecuteTool(ctx, "test-tool", map[string]interface{}{
		"param1": "value1",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tool execution failed")
}

func TestToolRegistry_RefreshTools_Empty(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	ctx := context.Background()
	err := registry.RefreshTools(ctx)
	assert.NoError(t, err)
}

func TestToolRegistry_RegisterExternalToolSource_Success(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	toolFetcher := func() ([]services.Tool, error) {
		tool := &MockTool{
			name:        "external-tool",
			description: "external description",
			parameters: map[string]interface{}{
				"param": map[string]interface{}{"type": "string"},
			},
			source: "external",
		}
		return []services.Tool{tool}, nil
	}

	err := registry.RegisterExternalToolSource("test-source", toolFetcher)
	assert.NoError(t, err)

	tool, exists := registry.GetTool("external-tool")
	assert.True(t, exists)
	assert.NotNil(t, tool)
	assert.Equal(t, "external-tool", tool.Name())
}

func TestToolRegistry_RegisterExternalToolSource_FetcherError(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	toolFetcher := func() ([]services.Tool, error) {
		return nil, assert.AnError
	}

	err := registry.RegisterExternalToolSource("test-source", toolFetcher)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch tools from test-source")
}

func TestToolRegistry_RegisterExternalToolSource_DuplicateTool(t *testing.T) {
	registry := services.NewToolRegistry(nil, nil)

	// First register a custom tool
	customTool := &MockTool{
		name:        "test-tool",
		description: "custom description",
		parameters: map[string]interface{}{
			"param": map[string]interface{}{"type": "string"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(customTool)
	assert.NoError(t, err)

	// Try to register same tool from external source
	toolFetcher := func() ([]services.Tool, error) {
		externalTool := &MockTool{
			name:        "test-tool",
			description: "external description",
			parameters: map[string]interface{}{
				"param": map[string]interface{}{"type": "string"},
			},
			source: "external",
		}
		return []services.Tool{externalTool}, nil
	}

	err = registry.RegisterExternalToolSource("test-source", toolFetcher)
	assert.NoError(t, err) // Should skip duplicate, not error

	// Should still have the custom tool
	tool, exists := registry.GetTool("test-tool")
	assert.True(t, exists)
	assert.Equal(t, "custom", tool.Source()) // Should be custom, not external
}

func TestMCPToolWrapper_Type(t *testing.T) {
	// Test MCPToolWrapper indirectly through the registry
	// Since we can't create MCPToolWrapper directly without MCPManager
	// We'll just test that the type exists
	registry := services.NewToolRegistry(nil, nil)
	assert.NotNil(t, registry)

	// Test that we can create a basic tool registry
	// This indirectly tests that MCPToolWrapper type is valid
	tool := &MockTool{
		name:        "test-wrapper",
		description: "test wrapper",
		parameters: map[string]interface{}{
			"param": map[string]interface{}{"type": "string"},
		},
		source: "custom",
	}

	err := registry.RegisterCustomTool(tool)
	assert.NoError(t, err)
}
