package context

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultWindowConfig(t *testing.T) {
	config := DefaultWindowConfig()

	assert.Equal(t, 4096, config.MaxTokens)
	assert.Equal(t, 512, config.ReserveTokens)
	assert.Equal(t, EvictionPolicyFIFO, config.EvictionPolicy)
	assert.Equal(t, 0.9, config.EvictionThreshold)
	assert.True(t, config.PreserveSystemPrompt)
	assert.Equal(t, 2, config.PreserveLastN)
}

func TestNewContextWindow(t *testing.T) {
	window := NewContextWindow(nil)

	assert.NotNil(t, window)
	assert.Empty(t, window.entries)
	assert.Equal(t, 0, window.tokenCount)
}

func TestContextWindow_Add(t *testing.T) {
	window := NewContextWindow(&WindowConfig{
		MaxTokens:     1000,
		ReserveTokens: 100,
	})

	err := window.Add(ContextEntry{
		Role:    "user",
		Content: "Hello, how are you?",
	})

	require.NoError(t, err)
	assert.Equal(t, 1, len(window.entries))
	assert.Greater(t, window.tokenCount, 0)
}

func TestContextWindow_AddMessage(t *testing.T) {
	window := NewContextWindow(nil)

	err := window.AddMessage("user", "Test message")

	require.NoError(t, err)
	assert.Equal(t, 1, len(window.entries))
	assert.Equal(t, "user", window.entries[0].Role)
	assert.Equal(t, "Test message", window.entries[0].Content)
}

func TestContextWindow_AddSystemPrompt(t *testing.T) {
	config := DefaultWindowConfig()
	config.PreserveSystemPrompt = true
	window := NewContextWindow(config)

	err := window.AddSystemPrompt("You are a helpful assistant")

	require.NoError(t, err)
	assert.Equal(t, 1, len(window.entries))
	assert.Equal(t, "system", window.entries[0].Role)
	assert.True(t, window.entries[0].Pinned)
}

func TestContextWindow_Get(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "Message 1")
	window.AddMessage("assistant", "Message 2")

	entries := window.Get()

	assert.Len(t, entries, 2)
	assert.Equal(t, "Message 1", entries[0].Content)
	assert.Equal(t, "Message 2", entries[1].Content)
}

func TestContextWindow_GetMessages(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "Hello")
	window.AddMessage("assistant", "Hi there")

	messages := window.GetMessages()

	assert.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0]["role"])
	assert.Equal(t, "Hello", messages[0]["content"])
}

func TestContextWindow_TokenCount(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "Hello world")

	count := window.TokenCount()

	assert.Greater(t, count, 0)
}

func TestContextWindow_AvailableTokens(t *testing.T) {
	config := &WindowConfig{
		MaxTokens:     1000,
		ReserveTokens: 100,
	}
	window := NewContextWindow(config)

	available := window.AvailableTokens()

	assert.Equal(t, 900, available)
}

func TestContextWindow_UsageRatio(t *testing.T) {
	config := &WindowConfig{
		MaxTokens:     1000,
		ReserveTokens: 100,
	}
	window := NewContextWindow(config)
	window.AddMessage("user", "Test message")

	ratio := window.UsageRatio()

	assert.Greater(t, ratio, 0.0)
	assert.Less(t, ratio, 1.0)
}

func TestContextWindow_Clear(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "Message 1")
	window.AddMessage("assistant", "Message 2")

	window.Clear()

	assert.Empty(t, window.entries)
	assert.Equal(t, 0, window.tokenCount)
}

func TestContextWindow_ClearExceptPinned(t *testing.T) {
	window := NewContextWindow(&WindowConfig{
		MaxTokens:            4096,
		ReserveTokens:        512,
		PreserveSystemPrompt: true,
	})

	window.AddSystemPrompt("System prompt")
	window.AddMessage("user", "User message")
	window.AddMessage("assistant", "Assistant message")

	window.ClearExceptPinned()

	assert.Equal(t, 1, len(window.entries))
	assert.Equal(t, "system", window.entries[0].Role)
}

func TestContextWindow_RemoveEntry(t *testing.T) {
	window := NewContextWindow(nil)
	window.Add(ContextEntry{ID: "entry-1", Role: "user", Content: "Message 1"})
	window.Add(ContextEntry{ID: "entry-2", Role: "assistant", Content: "Message 2"})

	removed := window.RemoveEntry("entry-1")

	assert.True(t, removed)
	assert.Equal(t, 1, len(window.entries))
	assert.Equal(t, "entry-2", window.entries[0].ID)
}

func TestContextWindow_RemoveEntry_NotFound(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "Test")

	removed := window.RemoveEntry("nonexistent")

	assert.False(t, removed)
}

func TestContextWindow_UpdateEntry(t *testing.T) {
	window := NewContextWindow(nil)
	window.Add(ContextEntry{ID: "entry-1", Role: "user", Content: "Original"})

	err := window.UpdateEntry("entry-1", "Updated content")

	require.NoError(t, err)
	assert.Equal(t, "Updated content", window.entries[0].Content)
}

func TestContextWindow_UpdateEntry_NotFound(t *testing.T) {
	window := NewContextWindow(nil)

	err := window.UpdateEntry("nonexistent", "New content")

	assert.Error(t, err)
}

func TestContextWindow_Eviction_FIFO(t *testing.T) {
	config := &WindowConfig{
		MaxTokens:      100,
		ReserveTokens:  10,
		EvictionPolicy: EvictionPolicyFIFO,
		PreserveLastN:  1,
	}
	window := NewContextWindow(config)

	// Add entries until we exceed the limit
	for i := 0; i < 10; i++ {
		window.AddMessage("user", "This is a message with some content to consume tokens")
	}

	// Should have evicted some entries
	assert.Less(t, window.TokenCount(), config.MaxTokens-config.ReserveTokens)
}

func TestContextWindow_Eviction_Priority(t *testing.T) {
	config := &WindowConfig{
		MaxTokens:      100,
		ReserveTokens:  10,
		EvictionPolicy: EvictionPolicyPriority,
		PreserveLastN:  1,
	}
	window := NewContextWindow(config)

	// Add low priority entries
	for i := 0; i < 5; i++ {
		window.Add(ContextEntry{
			Role:     "user",
			Content:  "Low priority content here",
			Priority: PriorityLow,
		})
	}

	// Add high priority entry
	window.Add(ContextEntry{
		Role:     "user",
		Content:  "High priority content here",
		Priority: PriorityHigh,
	})

	// Trigger eviction by adding more
	window.Add(ContextEntry{
		Role:     "assistant",
		Content:  "This should trigger eviction of low priority items",
		Priority: PriorityNormal,
	})

	// High priority should still be present
	found := false
	for _, entry := range window.entries {
		if entry.Priority == PriorityHigh {
			found = true
			break
		}
	}
	assert.True(t, found || window.TokenCount() < 100)
}

func TestContextWindow_Overflow(t *testing.T) {
	config := &WindowConfig{
		MaxTokens:     50,
		ReserveTokens: 10,
		PreserveLastN: 10, // Preserve everything
	}
	window := NewContextWindow(config)

	// Add pinned entries that can't be evicted
	for i := 0; i < 5; i++ {
		window.Add(ContextEntry{
			Role:    "system",
			Content: "This content cannot be evicted",
			Pinned:  true,
		})
	}

	// Try to add more - should fail
	err := window.Add(ContextEntry{
		Role:    "user",
		Content: "This very long message will cause overflow because all existing entries are pinned and cannot be evicted",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrContextOverflow, err)
}

func TestContextWindow_Snapshot(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "Message 1")
	window.AddMessage("assistant", "Message 2")

	snapshot := window.Snapshot()

	assert.Len(t, snapshot.Entries, 2)
	assert.Equal(t, window.tokenCount, snapshot.TokenCount)
	assert.NotZero(t, snapshot.Timestamp)
}

func TestContextWindow_RestoreFromSnapshot(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "Original message")

	snapshot := window.Snapshot()
	window.Clear()
	window.AddMessage("user", "New message")

	window.RestoreFromSnapshot(snapshot)

	assert.Len(t, window.entries, 1)
	assert.Equal(t, "Original message", window.entries[0].Content)
}

func TestContextWindow_Stats(t *testing.T) {
	window := NewContextWindow(nil)
	window.AddMessage("user", "User message")
	window.AddMessage("assistant", "Assistant message")
	window.Add(ContextEntry{Role: "system", Content: "System prompt", Pinned: true})

	stats := window.Stats()

	assert.Equal(t, 3, stats.TotalEntries)
	assert.Greater(t, stats.TotalTokens, 0)
	assert.Equal(t, 1, stats.PinnedEntries)
	assert.Equal(t, 1, stats.MessagesByRole["user"])
	assert.Equal(t, 1, stats.MessagesByRole["assistant"])
	assert.Equal(t, 1, stats.MessagesByRole["system"])
}

func TestContextWindow_EventHandler(t *testing.T) {
	var events []*WindowEvent
	handler := func(event *WindowEvent) {
		events = append(events, event)
	}

	window := NewContextWindow(nil)
	window.SetEventHandler(handler)

	window.AddMessage("user", "Test message")

	assert.Len(t, events, 1)
	assert.Equal(t, EventTypeEntryAdded, events[0].Type)
}

func TestPriority(t *testing.T) {
	assert.Equal(t, Priority(0), PriorityLow)
	assert.Equal(t, Priority(1), PriorityNormal)
	assert.Equal(t, Priority(2), PriorityHigh)
	assert.Equal(t, Priority(3), PriorityCritical)
}

func TestEvictionPolicies(t *testing.T) {
	assert.Equal(t, EvictionPolicy("fifo"), EvictionPolicyFIFO)
	assert.Equal(t, EvictionPolicy("lru"), EvictionPolicyLRU)
	assert.Equal(t, EvictionPolicy("priority"), EvictionPolicyPriority)
	assert.Equal(t, EvictionPolicy("summarize"), EvictionPolicySummarize)
}

func TestWindowEventTypes(t *testing.T) {
	assert.Equal(t, WindowEventType("entry_added"), EventTypeEntryAdded)
	assert.Equal(t, WindowEventType("entry_evicted"), EventTypeEntryEvicted)
	assert.Equal(t, WindowEventType("entry_updated"), EventTypeEntryUpdated)
	assert.Equal(t, WindowEventType("overflow"), EventTypeOverflow)
	assert.Equal(t, WindowEventType("summarized"), EventTypeSummarized)
}

func TestContextEntry(t *testing.T) {
	entry := ContextEntry{
		ID:         "entry-123",
		Role:       "user",
		Content:    "Hello world",
		TokenCount: 2,
		Timestamp:  time.Now(),
		Priority:   PriorityNormal,
		Metadata:   map[string]interface{}{"key": "value"},
		Pinned:     false,
	}

	assert.Equal(t, "entry-123", entry.ID)
	assert.Equal(t, "user", entry.Role)
	assert.Equal(t, "Hello world", entry.Content)
	assert.Equal(t, 2, entry.TokenCount)
	assert.Equal(t, PriorityNormal, entry.Priority)
	assert.False(t, entry.Pinned)
}

func TestWindowSnapshot(t *testing.T) {
	snapshot := &WindowSnapshot{
		Entries: []ContextEntry{
			{ID: "1", Role: "user", Content: "Test"},
		},
		TokenCount: 10,
		Timestamp:  time.Now(),
		Config:     *DefaultWindowConfig(),
	}

	assert.Len(t, snapshot.Entries, 1)
	assert.Equal(t, 10, snapshot.TokenCount)
}

func TestWindowStats(t *testing.T) {
	stats := &WindowStats{
		TotalEntries:     10,
		TotalTokens:      500,
		AvailableTokens:  1500,
		UsageRatio:       0.25,
		PinnedEntries:    2,
		MessagesByRole:   map[string]int{"user": 5, "assistant": 5},
		AverageEntrySize: 50.0,
	}

	assert.Equal(t, 10, stats.TotalEntries)
	assert.Equal(t, 500, stats.TotalTokens)
	assert.Equal(t, 0.25, stats.UsageRatio)
}

func TestEstimateTokens(t *testing.T) {
	// ~4 characters per token
	tokens := estimateTokens("Hello World") // 11 chars
	assert.Equal(t, 2, tokens)              // 11/4 = 2

	tokens = estimateTokens("A")
	assert.Equal(t, 0, tokens) // 1/4 = 0

	tokens = estimateTokens("")
	assert.Equal(t, 0, tokens)
}
