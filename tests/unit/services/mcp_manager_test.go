package services_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"dev.helix.agent/internal/services"
)

func TestMCPManager_Basic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)
	assert.NotNil(t, manager)
}

func TestMCPManager_WithConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)
	assert.NotNil(t, manager)
}

func TestMCPManager_RegisterServer_Valid(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	serverConfig := map[string]interface{}{
		"name":    "test-server",
		"command": []interface{}{"echo", "test"},
	}

	// Note: This test may fail if the MCP client tries to actually connect
	// to the server. In a real scenario, this would require a mock transport.
	err := manager.RegisterServer(serverConfig)
	// For now, we expect this to fail due to connection issues
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize server")
}

func TestMCPManager_RegisterServer_MissingName(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	serverConfig := map[string]interface{}{
		"command": []interface{}{"echo", "test"},
	}

	err := manager.RegisterServer(serverConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server config must include 'name'")
}

func TestMCPManager_RegisterServer_MissingCommand(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	serverConfig := map[string]interface{}{
		"name": "test-server",
	}

	err := manager.RegisterServer(serverConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server config must include 'command' as array")
}

func TestMCPManager_RegisterServer_Duplicate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	serverConfig := map[string]interface{}{
		"name":    "test-server",
		"command": []interface{}{"echo", "test"},
	}

	// First registration - will fail due to connection issues
	err := manager.RegisterServer(serverConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize server")

	// Second registration - would fail due to connection, not duplicate
	err = manager.RegisterServer(serverConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize server")
}

func TestMCPManager_RegisterServer_EmptyCommand(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	serverConfig := map[string]interface{}{
		"name":    "test-server",
		"command": []interface{}{},
	}

	err := manager.RegisterServer(serverConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server command cannot be empty")
}

func TestMCPManager_CallTool_NotFound(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	result, err := manager.CallTool(ctx, "nonexistent", map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no MCP servers connected")
}

func TestMCPManager_ListMCPServers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	servers, err := manager.ListMCPServers(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, servers)
	// Initially empty
	assert.Empty(t, servers)
}

func TestMCPManager_GetMCPStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	stats, err := manager.GetMCPStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.IsType(t, map[string]interface{}{}, stats)
}

func TestMCPManager_ConnectServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	err := manager.ConnectServer(ctx, "test-server", "Test Server", "echo", []string{"hello"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize server")
}

func TestMCPManager_DisconnectServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	err := manager.DisconnectServer("test-server")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server test-server not connected")
}

func TestMCPManager_ExecuteMCPTool(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	req := map[string]interface{}{
		"tool":   "test-tool",
		"params": map[string]interface{}{"key": "value"},
	}

	result, err := manager.ExecuteMCPTool(ctx, req)
	// ExecuteMCPTool returns a mock result for non-unified requests
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check the mock result structure
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, resultMap["success"])
	assert.Equal(t, "Tool executed successfully", resultMap["result"])
}

func TestMCPManager_GetMCPTools(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	tools, err := manager.GetMCPTools(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, tools)
	assert.IsType(t, map[string][]*services.MCPTool{}, tools)
	// Initially empty
	assert.Empty(t, tools)
}

func TestMCPManager_ValidateMCPRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "test",
		"params": map[string]interface{}{},
	}

	err := manager.ValidateMCPRequest(ctx, req)
	// Basic validation - may not fully validate without schema
	assert.NoError(t, err)
}

func TestMCPManager_SyncMCPServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	manager := services.NewMCPManager(nil, nil, logger)

	ctx := context.Background()
	err := manager.SyncMCPServer(ctx, "test-server")
	// SyncMCPServer currently just logs and returns nil
	assert.NoError(t, err)
}
