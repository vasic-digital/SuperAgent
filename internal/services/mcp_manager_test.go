package services

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMCPTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	return log
}

func TestNewMCPManager(t *testing.T) {
	log := newMCPTestLogger()
	manager := NewMCPManager(nil, nil, log)

	require.NotNil(t, manager)
	assert.NotNil(t, manager.getClient())
	assert.Nil(t, manager.repo)
	assert.Nil(t, manager.cache)
	assert.NotNil(t, manager.log)
}

func TestMCPManager_ListMCPServers(t *testing.T) {
	log := newMCPTestLogger()
	manager := NewMCPManager(nil, nil, log)
	ctx := context.Background()

	servers, err := manager.ListMCPServers(ctx)
	require.NoError(t, err)
	require.NotNil(t, servers)
	assert.Len(t, servers, 0)
}

func TestMCPManager_ListTools(t *testing.T) {
	log := newMCPTestLogger()
	manager := NewMCPManager(nil, nil, log)

	tools := manager.ListTools()
	assert.Len(t, tools, 0)
}

func TestMCPToolDefinition(t *testing.T) {
	tool := MCPToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}

	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)
}

func TestToolInputSchema(t *testing.T) {
	schema := ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"param1": map[string]string{"type": "string"},
		},
		Required: []string{"param1"},
	}

	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "param1")
	assert.Contains(t, schema.Required, "param1")
}

func TestJSONRPCError(t *testing.T) {
	err := &JSONRPCError{
		Code:    -32601,
		Message: "Method not found",
	}

	assert.Equal(t, -32601, err.Code)
	assert.Equal(t, "Method not found", err.Message)
	assert.Contains(t, err.Error(), "JSON-RPC Error")
}

func TestMCPInitializeRequest(t *testing.T) {
	req := MCPInitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{},
		ClientInfo:      map[string]string{"name": "test"},
	}

	assert.Equal(t, "2024-11-05", req.ProtocolVersion)
	assert.Equal(t, "test", req.ClientInfo["name"])
}

func TestMCPInitializeResult(t *testing.T) {
	result := MCPInitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{},
		ServerInfo:      map[string]string{"name": "test-server"},
		Instructions:    "Test instructions",
	}

	assert.Equal(t, "2024-11-05", result.ProtocolVersion)
	assert.Equal(t, "test-server", result.ServerInfo["name"])
	assert.Equal(t, "Test instructions", result.Instructions)
}

func TestCapabilities(t *testing.T) {
	caps := Capabilities{
		Tools: &ToolCapabilities{
			ListChanged: true,
		},
	}

	assert.NotNil(t, caps.Tools)
	assert.True(t, caps.Tools.ListChanged)
}
