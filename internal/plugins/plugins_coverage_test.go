package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestConfigManager_LoadPluginConfig_InvalidPath verifies that LoadPluginConfig
// returns an error when the plugin name fails path validation (e.g., contains "..").
func TestConfigManager_LoadPluginConfig_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewConfigManager(tmpDir)

	// "../etc/passwd" contains ".." → filepath.Clean keeps ".." → ValidatePath returns false
	_, err := cm.LoadPluginConfig("../etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plugin name")
}

// TestConfigManager_SavePluginConfig_MarshalError verifies that SavePluginConfig
// returns an error when config contains a non-JSON-serializable value (channel).
func TestConfigManager_SavePluginConfig_MarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewConfigManager(tmpDir)

	// Channels are not JSON-serializable — json.MarshalIndent must fail
	config := map[string]interface{}{
		"name":    "test-plugin",
		"version": "1.0.0",
		"invalid": make(chan int), // causes marshal failure
	}

	err := cm.SavePluginConfig("test-plugin", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal config")
}

// TestConfigManager_UpdatePluginConfig_InvalidPath verifies that UpdatePluginConfig
// propagates the error from LoadPluginConfig when the plugin name is invalid.
func TestConfigManager_UpdatePluginConfig_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewConfigManager(tmpDir)

	// "../bad" fails ValidatePath → LoadPluginConfig returns error → UpdatePluginConfig returns it
	updates := map[string]interface{}{"key": "value"}
	err := cm.UpdatePluginConfig("../bad", updates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plugin name")
}

// TestReloader_ReloadPluginConfig_LoadConfigError verifies that ReloadPluginConfig
// propagates the error when LoadPluginConfig returns a non-nil error.
//
// We achieve this by creating a directory named "my-plugin.json" inside the
// config dir — ReadFile on a directory fails with a read error.
func TestReloader_ReloadPluginConfig_LoadConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry()

	// Register a plugin so the registry lookup succeeds
	plugin := new(MockLLMPlugin)
	plugin.On("Name").Return("my-plugin")
	plugin.On("Init", mock.Anything).Return(nil)
	require.NoError(t, registry.Register(plugin))

	// Create a directory where the config file would be — ReadFile on a dir fails
	configDir := filepath.Join(tmpDir, "my-plugin.json")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configMgr := NewConfigManager(tmpDir)
	loader := NewLoader(registry)
	health := NewHealthMonitor(registry, 0, 0)
	lifecycle := NewLifecycleManager(registry, loader, health)
	reloader := NewReloader(registry, configMgr, lifecycle)

	ctx := context.Background()
	err := reloader.ReloadPluginConfig(ctx, "my-plugin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load new config")
}
