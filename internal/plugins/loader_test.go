package plugins

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	assert.NotNil(t, loader)
	assert.Equal(t, registry, loader.registry)
}

func TestLoader_Load_FileNotFound(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	plugin, err := loader.Load("/nonexistent/path/to/plugin.so")

	assert.Error(t, err)
	assert.Nil(t, plugin)
	assert.Contains(t, err.Error(), "failed to open plugin")
}

func TestLoader_Load_InvalidPlugin(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	// Create a temp file that's not a valid Go plugin
	tmpDir := t.TempDir()
	invalidPluginPath := tmpDir + "/invalid.so"

	// Write invalid content using os.WriteFile
	err := os.WriteFile(invalidPluginPath, []byte("not a valid plugin"), 0644)
	require.NoError(t, err)

	plugin, err := loader.Load(invalidPluginPath)

	assert.Error(t, err)
	assert.Nil(t, plugin)
	assert.Contains(t, err.Error(), "failed to open plugin")
}

func TestLoader_Unload_NotFound(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	err := loader.Unload("nonexistent-plugin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unregister plugin")
}

func TestLoader_Unload_Success(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	// Register a mock plugin
	plugin := new(MockLLMPlugin)
	plugin.On("Name").Return("test-plugin")

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Unload should succeed
	err = loader.Unload("test-plugin")
	assert.NoError(t, err)

	// Plugin should no longer exist
	_, exists := registry.Get("test-plugin")
	assert.False(t, exists)
}

func TestLoader_WithNilRegistry(t *testing.T) {
	// Creating a loader with nil registry
	loader := NewLoader(nil)

	assert.NotNil(t, loader)
	assert.Nil(t, loader.registry)
}

func TestLoader_MultipleUnloads(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	// Register a mock plugin
	plugin := new(MockLLMPlugin)
	plugin.On("Name").Return("test-plugin")

	err := registry.Register(plugin)
	require.NoError(t, err)

	// First unload should succeed
	err = loader.Unload("test-plugin")
	assert.NoError(t, err)

	// Second unload should fail
	err = loader.Unload("test-plugin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unregister plugin")
}
