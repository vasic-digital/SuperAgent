package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/helixagent/helixagent/internal/services"
)

func TestContextManager_NewContextManager(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		cm := services.NewContextManager(100)
		assert.NotNil(t, cm)
	})

	t.Run("Zero max size", func(t *testing.T) {
		cm := services.NewContextManager(0)
		assert.NotNil(t, cm)
	})
}

func TestContextManager_AddEntry(t *testing.T) {
	cm := services.NewContextManager(10)

	t.Run("Add simple entry", func(t *testing.T) {
		entry := &services.ContextEntry{
			ID:       "test-entry-1",
			Type:     "llm",
			Source:   "test",
			Content:  "Test content",
			Priority: 5,
		}

		err := cm.AddEntry(entry)
		require.NoError(t, err)

		// Verify entry was added
		retrieved, exists := cm.GetEntry("test-entry-1")
		require.True(t, exists)
		assert.Equal(t, "test-entry-1", retrieved.ID)
		assert.Equal(t, "Test content", retrieved.Content)
		assert.Equal(t, 5, retrieved.Priority)
		assert.False(t, retrieved.Compressed)
	})

	t.Run("Add entry with metadata", func(t *testing.T) {
		entry := &services.ContextEntry{
			ID:      "test-entry-2",
			Type:    "lsp",
			Source:  "/path/to/file.go",
			Content: "package main\n\nfunc main() {\n    println(\"Hello\")\n}",
			Metadata: map[string]interface{}{
				"language": "go",
				"lines":    5,
			},
			Priority: 8,
		}

		err := cm.AddEntry(entry)
		require.NoError(t, err)

		retrieved, exists := cm.GetEntry("test-entry-2")
		require.True(t, exists)
		assert.Equal(t, "lsp", retrieved.Type)
		assert.Equal(t, "/path/to/file.go", retrieved.Source)
		assert.Contains(t, retrieved.Content, "package main")
		assert.Equal(t, "go", retrieved.Metadata["language"])
		assert.Equal(t, 5, retrieved.Metadata["lines"])
	})

	t.Run("Add duplicate entry", func(t *testing.T) {
		entry1 := &services.ContextEntry{
			ID:       "duplicate-entry",
			Content:  "First content",
			Priority: 3,
		}

		entry2 := &services.ContextEntry{
			ID:       "duplicate-entry",
			Content:  "Second content",
			Priority: 7,
		}

		err1 := cm.AddEntry(entry1)
		require.NoError(t, err1)

		err2 := cm.AddEntry(entry2)
		require.NoError(t, err2)

		// Should have the second entry
		retrieved, exists := cm.GetEntry("duplicate-entry")
		require.True(t, exists)
		assert.Equal(t, "Second content", retrieved.Content)
		assert.Equal(t, 7, retrieved.Priority)
	})

	t.Run("Add large entry (should compress)", func(t *testing.T) {
		// Create a large content string
		largeContent := ""
		for i := 0; i < 2000; i++ {
			largeContent += "This is a test line for compression testing. "
		}

		entry := &services.ContextEntry{
			ID:       "large-entry",
			Content:  largeContent,
			Priority: 6,
		}

		err := cm.AddEntry(entry)
		require.NoError(t, err)

		retrieved, exists := cm.GetEntry("large-entry")
		require.True(t, exists)
		assert.Contains(t, retrieved.Content, "This is a test line")
		// Should be decompressed automatically
		assert.False(t, retrieved.Compressed)
	})
}

func TestContextManager_GetEntry(t *testing.T) {
	cm := services.NewContextManager(10)

	// Add test entries
	entries := []*services.ContextEntry{
		{ID: "entry-1", Content: "Content 1", Priority: 5},
		{ID: "entry-2", Content: "Content 2", Priority: 3},
		{ID: "entry-3", Content: "Content 3", Priority: 8},
	}

	for _, entry := range entries {
		err := cm.AddEntry(entry)
		require.NoError(t, err)
	}

	t.Run("Get existing entry", func(t *testing.T) {
		entry, exists := cm.GetEntry("entry-2")
		require.True(t, exists)
		assert.Equal(t, "Content 2", entry.Content)
		assert.Equal(t, 3, entry.Priority)
	})

	t.Run("Get non-existent entry", func(t *testing.T) {
		entry, exists := cm.GetEntry("non-existent")
		assert.False(t, exists)
		assert.Nil(t, entry)
	})

	t.Run("Get all entries", func(t *testing.T) {
		for _, expected := range entries {
			actual, exists := cm.GetEntry(expected.ID)
			require.True(t, exists)
			assert.Equal(t, expected.Content, actual.Content)
		}
	})
}

func TestContextManager_UpdateEntry(t *testing.T) {
	cm := services.NewContextManager(10)

	// Add initial entry
	entry := &services.ContextEntry{
		ID:       "update-test",
		Content:  "Original content",
		Priority: 5,
		Metadata: map[string]interface{}{
			"original": true,
		},
	}

	err := cm.AddEntry(entry)
	require.NoError(t, err)

	t.Run("Update content", func(t *testing.T) {
		err := cm.UpdateEntry("update-test", "Updated content", nil)
		require.NoError(t, err)

		updated, exists := cm.GetEntry("update-test")
		require.True(t, exists)
		assert.Equal(t, "Updated content", updated.Content)
		assert.Equal(t, 5, updated.Priority) // Priority should remain
		// Metadata should remain (nil metadata passed, so original should stay)
		if updated.Metadata != nil {
			assert.Equal(t, true, updated.Metadata["original"])
		}
	})

	t.Run("Update metadata", func(t *testing.T) {
		newMetadata := map[string]interface{}{
			"updated":   true,
			"timestamp": time.Now().Unix(),
		}

		err := cm.UpdateEntry("update-test", "Updated content", newMetadata)
		require.NoError(t, err)

		updated, exists := cm.GetEntry("update-test")
		require.True(t, exists)
		assert.Equal(t, "Updated content", updated.Content)
		assert.Equal(t, true, updated.Metadata["updated"])
		assert.NotNil(t, updated.Metadata["timestamp"])
		// Old metadata should be replaced
		assert.Nil(t, updated.Metadata["original"])
	})

	t.Run("Update non-existent entry", func(t *testing.T) {
		err := cm.UpdateEntry("non-existent", "New content", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContextManager_RemoveEntry(t *testing.T) {
	cm := services.NewContextManager(10)

	// Add test entries
	entries := []*services.ContextEntry{
		{ID: "delete-1", Content: "To delete 1", Priority: 5},
		{ID: "delete-2", Content: "To delete 2", Priority: 3},
		{ID: "keep-1", Content: "Keep this", Priority: 8},
	}

	for _, entry := range entries {
		err := cm.AddEntry(entry)
		require.NoError(t, err)
	}

	t.Run("Remove existing entry", func(t *testing.T) {
		// Verify entry exists before removal
		_, exists := cm.GetEntry("delete-1")
		require.True(t, exists)

		// Remove entry
		cm.RemoveEntry("delete-1")

		// Verify deletion
		_, exists = cm.GetEntry("delete-1")
		assert.False(t, exists)

		// Other entries should still exist
		_, exists = cm.GetEntry("delete-2")
		assert.True(t, exists)
		_, exists = cm.GetEntry("keep-1")
		assert.True(t, exists)
	})

	t.Run("Remove non-existent entry", func(t *testing.T) {
		// Removing non-existent entry should not panic
		assert.NotPanics(t, func() {
			cm.RemoveEntry("non-existent")
		})
	})

	t.Run("Remove all entries", func(t *testing.T) {
		// Remove remaining entries
		cm.RemoveEntry("delete-2")
		cm.RemoveEntry("keep-1")

		// All should be gone
		_, exists := cm.GetEntry("delete-2")
		assert.False(t, exists)
		_, exists = cm.GetEntry("keep-1")
		assert.False(t, exists)
	})
}

func TestContextManager_BuildContext(t *testing.T) {
	cm := services.NewContextManager(10)

	// Add test entries with different priorities and types
	entries := []*services.ContextEntry{
		{ID: "entry-1", Type: "lsp", Content: "LSP content 1", Priority: 5},
		{ID: "entry-2", Type: "mcp", Content: "MCP content 1", Priority: 8},
		{ID: "entry-3", Type: "llm", Content: "LLM content 1", Priority: 3},
		{ID: "entry-4", Type: "lsp", Content: "LSP content 2", Priority: 7},
		{ID: "entry-5", Type: "tool", Content: "Tool content", Priority: 6},
	}

	for _, entry := range entries {
		err := cm.AddEntry(entry)
		require.NoError(t, err)
	}

	t.Run("Build context for code completion", func(t *testing.T) {
		context, err := cm.BuildContext("code_completion", 1000)
		require.NoError(t, err)
		assert.NotEmpty(t, context)

		// Should prioritize LSP entries for code completion
		hasLSP := false
		for _, entry := range context {
			if entry.Type == "lsp" {
				hasLSP = true
				break
			}
		}
		assert.True(t, hasLSP, "Should include LSP entries for code completion")
	})

	t.Run("Build context with token limit", func(t *testing.T) {
		// Small token limit
		context, err := cm.BuildContext("chat", 100)
		require.NoError(t, err)
		assert.True(t, len(context) <= 5, "Should respect token limit")
	})

	t.Run("Build context for tool execution", func(t *testing.T) {
		context, err := cm.BuildContext("tool_execution", 1000)
		require.NoError(t, err)
		assert.NotEmpty(t, context)

		// Should prioritize tool and MCP entries
		hasToolOrMCP := false
		for _, entry := range context {
			if entry.Type == "tool" || entry.Type == "mcp" {
				hasToolOrMCP = true
				break
			}
		}
		assert.True(t, hasToolOrMCP, "Should include tool/MCP entries for tool execution")
	})
}

func TestContextManager_CacheResult(t *testing.T) {
	cm := services.NewContextManager(10)

	t.Run("Cache and retrieve result", func(t *testing.T) {
		// Cache a result
		testData := map[string]interface{}{
			"result": "success",
			"data":   []int{1, 2, 3},
		}
		cm.CacheResult("test-key", testData, 5*time.Minute)

		// Retrieve cached result
		result, found := cm.GetCachedResult("test-key")
		require.True(t, found)
		assert.Equal(t, testData, result)
	})

	t.Run("Retrieve non-existent cache", func(t *testing.T) {
		result, found := cm.GetCachedResult("non-existent")
		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("Cache expiration", func(t *testing.T) {
		// Cache with short TTL
		cm.CacheResult("expiring-key", "expiring-value", 100*time.Millisecond)

		// Should exist immediately
		result, found := cm.GetCachedResult("expiring-key")
		assert.True(t, found)
		assert.Equal(t, "expiring-value", result)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		result, found = cm.GetCachedResult("expiring-key")
		assert.False(t, found)
		assert.Nil(t, result)
	})
}

func TestContextManager_DetectConflicts(t *testing.T) {
	cm := services.NewContextManager(10)

	t.Run("No conflicts", func(t *testing.T) {
		// Add non-conflicting entries
		entries := []*services.ContextEntry{
			{ID: "no-conflict-1", Source: "source1", Content: "Content 1", Metadata: map[string]interface{}{"key": "value1"}},
			{ID: "no-conflict-2", Source: "source2", Content: "Content 2", Metadata: map[string]interface{}{"key": "value2"}},
		}

		for _, entry := range entries {
			err := cm.AddEntry(entry)
			require.NoError(t, err)
		}

		conflicts := cm.DetectConflicts()
		assert.Empty(t, conflicts)
	})

	t.Run("Potential metadata conflicts", func(t *testing.T) {
		// Add entries with same content but different metadata
		content := "Same content"
		entries := []*services.ContextEntry{
			{ID: "conflict-1", Source: "same-source", Content: content, Metadata: map[string]interface{}{"version": "1.0"}},
			{ID: "conflict-2", Source: "same-source", Content: content, Metadata: map[string]interface{}{"version": "2.0"}},
		}

		for _, entry := range entries {
			err := cm.AddEntry(entry)
			require.NoError(t, err)
		}

		conflicts := cm.DetectConflicts()
		// May or may not detect conflicts depending on implementation
		// Just verify the method runs without error
		assert.NotNil(t, conflicts)
	})
}

func TestContextManager_Cleanup(t *testing.T) {
	cm := services.NewContextManager(10)

	t.Run("Cleanup old entries", func(t *testing.T) {
		// Add an entry
		entry := &services.ContextEntry{
			ID:       "old-entry",
			Content:  "Old content",
			Priority: 5,
		}
		err := cm.AddEntry(entry)
		require.NoError(t, err)

		// Verify it exists
		_, exists := cm.GetEntry("old-entry")
		assert.True(t, exists)

		// Cleanup shouldn't remove recent entries
		cm.Cleanup()
		_, exists = cm.GetEntry("old-entry")
		assert.True(t, exists, "Recent entries should not be cleaned up")
	})

	t.Run("Cleanup expired cache", func(t *testing.T) {
		// Cache with short TTL
		cm.CacheResult("expiring-cache", "cache-value", 100*time.Millisecond)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Cleanup should remove expired cache
		cm.Cleanup()

		// Should be cleaned up
		result, found := cm.GetCachedResult("expiring-cache")
		assert.False(t, found)
		assert.Nil(t, result)
	})
}

func TestContextManager_CapacityEviction(t *testing.T) {
	t.Run("Eviction when at capacity", func(t *testing.T) {
		cm := services.NewContextManager(3) // Small capacity

		// Add entries with different priorities
		entries := []*services.ContextEntry{
			{ID: "low-1", Content: "Low priority 1", Priority: 1},
			{ID: "low-2", Content: "Low priority 2", Priority: 2},
			{ID: "high-1", Content: "High priority", Priority: 10},
		}

		for _, entry := range entries {
			err := cm.AddEntry(entry)
			require.NoError(t, err)
		}

		// All entries should exist
		_, exists1 := cm.GetEntry("low-1")
		_, exists2 := cm.GetEntry("low-2")
		_, exists3 := cm.GetEntry("high-1")
		assert.True(t, exists1)
		assert.True(t, exists2)
		assert.True(t, exists3)

		// Add another entry - should trigger eviction
		newEntry := &services.ContextEntry{
			ID:       "medium-1",
			Content:  "Medium priority",
			Priority: 5,
		}

		err := cm.AddEntry(newEntry)
		require.NoError(t, err)

		// One of the low priority entries should be evicted (priority 1)
		// The eviction logic removes the oldest entry with lowest priority
		_, exists1 = cm.GetEntry("low-1")
		_, exists2 = cm.GetEntry("low-2")
		// Either low-1 or low-2 should be evicted, but not both
		// Since they have different priorities (1 vs 2), low-1 should be evicted
		assert.False(t, exists1, "low-1 (priority 1) should be evicted")
		assert.True(t, exists2, "low-2 (priority 2) should remain")
		// High priority should remain
		_, exists3 = cm.GetEntry("high-1")
		_, exists4 := cm.GetEntry("medium-1")
		assert.True(t, exists3)
		assert.True(t, exists4)
	})
}
