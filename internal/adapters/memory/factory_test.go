package memory_test

import (
	"testing"

	adapter "dev.helix.agent/internal/adapters/memory"
	"github.com/stretchr/testify/assert"
)

func TestIsHelixMemoryEnabled(t *testing.T) {
	// Value depends on build tag; just verify it returns a bool
	enabled := adapter.IsHelixMemoryEnabled()
	assert.IsType(t, true, enabled)
}

func TestMemoryBackendName(t *testing.T) {
	name := adapter.MemoryBackendName()
	assert.NotEmpty(t, name)
	// Must be one of the known backends
	assert.Contains(t, []string{"digital.vasic.memory", "digital.vasic.helixmemory"}, name)
}

func TestNewOptimalStoreAdapter_Standard(t *testing.T) {
	if adapter.IsHelixMemoryEnabled() {
		t.Skip("Skipping standard test when helixmemory tag is active")
	}
	// Standard build returns nil (caller provides their own store)
	store := adapter.NewOptimalStoreAdapter()
	assert.Nil(t, store)
	assert.Equal(t, "digital.vasic.memory", adapter.MemoryBackendName())
	assert.False(t, adapter.IsHelixMemoryEnabled())
}

func TestNewOptimalStoreAdapter_HelixMemory(t *testing.T) {
	if !adapter.IsHelixMemoryEnabled() {
		t.Skip("Skipping helixmemory test when tag is not active")
	}
	// HelixMemory build returns a configured adapter
	store := adapter.NewOptimalStoreAdapter()
	assert.NotNil(t, store)
	assert.Equal(t, "digital.vasic.helixmemory", adapter.MemoryBackendName())
	assert.True(t, adapter.IsHelixMemoryEnabled())
}
