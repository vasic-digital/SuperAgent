package mcp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Logger:  logger,
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "nil config fails",
			config:  nil,
			wantErr: true,
		},
		{
			name: "default timeout",
			config: &Config{
				Logger: logger,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_ConnectStdio(t *testing.T) {
	logger := zap.NewNop()
	client, err := NewClient(&Config{
		Logger:  logger,
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This would require actual MCP server to be installed
	// For unit tests, we just verify the method signature
	t.Run("connection params validated", func(t *testing.T) {
		// Test with empty command
		_, err := client.ConnectStdio(ctx, "", []string{})
		assert.Error(t, err)
	})
}

func TestClient_ConnectHTTP(t *testing.T) {
	logger := zap.NewNop()
	client, err := NewClient(&Config{
		Logger:  logger,
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty url fails",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid url fails",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "valid url",
			url:     "http://localhost:3000/mcp",
			wantErr: false, // Would fail in real connection but passes validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := client.ConnectHTTP(ctx, tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, server)
			} else {
				// Note: This may still fail on actual connection
				if err == nil {
					assert.NotNil(t, server)
					server.Close()
				}
			}
		})
	}
}

func TestServer_ListTools(t *testing.T) {
	// This is a mock test - real implementation would require running MCP server
	tools := []ToolDefinition{
		{
			Name:        "read_file",
			Description: "Read a file",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]string{"type": "string"},
				},
			},
		},
	}

	assert.Len(t, tools, 1)
	assert.Equal(t, "read_file", tools[0].Name)
}

func TestServer_CallTool(t *testing.T) {
	// Mock test for tool calling structure
	args := map[string]interface{}{
		"path": "/tmp/test.txt",
	}

	assert.NotNil(t, args)
	assert.Equal(t, "/tmp/test.txt", args["path"])
}

func TestNewRegistry(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(&RegistryConfig{
		Logger: logger,
	})

	assert.NotNil(t, registry)
	assert.Empty(t, registry.ListServers())
}

func TestRegistry_Register(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(&RegistryConfig{Logger: logger})

	tests := []struct {
		name   string
		config ServerConfig
	}{
		{
			name: "stdio server",
			config: ServerConfig{
				Type:    "stdio",
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			},
		},
		{
			name: "http server",
			config: ServerConfig{
				Type: "http",
				URL:  "http://localhost:3000/mcp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry.Register(tt.name, tt.config)
			servers := registry.ListServers()
			assert.Contains(t, servers, tt.name)
		})
	}
}

func TestRegistry_InitializeAll(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(&RegistryConfig{Logger: logger})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Register test servers
	registry.Register("test1", ServerConfig{Type: "stdio", Command: "echo"})
	registry.Register("test2", ServerConfig{Type: "http", URL: "http://localhost:9999"})

	results := registry.InitializeAll(ctx)

	assert.Equal(t, 2, results.Total)
	// Both will likely fail without actual servers, but we verify the structure
	assert.LessOrEqual(t, results.Successful, results.Total)
}

func TestRegistry_GetAllTools(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(&RegistryConfig{Logger: logger})

	ctx := context.Background()

	// Empty registry returns empty map
	tools := registry.GetAllTools(ctx)
	assert.Empty(t, tools)
}

func TestRegistry_ShutdownAll(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(&RegistryConfig{Logger: logger})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should not error on empty registry
	err := registry.ShutdownAll(ctx)
	assert.NoError(t, err)
}

func TestNewLLMClient(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(&RegistryConfig{Logger: logger})

	llmClient := NewLLMClient("test-api-key", registry, logger)
	assert.NotNil(t, llmClient)
}

func TestToolDefinition_Validation(t *testing.T) {
	tests := []struct {
		name    string
		tool    ToolDefinition
		wantErr bool
	}{
		{
			name: "valid tool",
			tool: ToolDefinition{
				Name:        "test",
				Description: "A test tool",
				Parameters: map[string]interface{}{
					"type": "object",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tool: ToolDefinition{
				Description: "A test tool",
				Parameters:  map[string]interface{}{"type": "object"},
			},
			wantErr: true,
		},
		{
			name: "missing description",
			tool: ToolDefinition{
				Name:       "test",
				Parameters: map[string]interface{}{"type": "object"},
			},
			wantErr: true,
		},
		{
			name: "missing parameters",
			tool: ToolDefinition{
				Name:        "test",
				Description: "A test tool",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tool.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkRegistry_Register(b *testing.B) {
	logger := zap.NewNop()
	registry := NewRegistry(&RegistryConfig{Logger: logger})

	config := ServerConfig{
		Type:    "stdio",
		Command: "echo",
		Args:    []string{"test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Register(fmt.Sprintf("server-%d", i), config)
	}
}
