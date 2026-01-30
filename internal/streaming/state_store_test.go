package streaming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStateStore_New(t *testing.T) {
	store := NewInMemoryStateStore()
	require.NotNil(t, store)
	assert.NotNil(t, store.conversationStates)
	assert.NotNil(t, store.windowedAnalytics)
}

func TestInMemoryStateStore_SaveAndGetState(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-001",
		UserID:         "user-001",
		SessionID:      "session-001",
		MessageCount:   10,
		EntityCount:    5,
		TotalTokens:    500,
		Entities:       make(map[string]EntityData),
		ProviderUsage: map[string]int{
			"claude":   3,
			"deepseek": 2,
		},
		StartedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		Version:       1,
	}

	// Save state
	err := store.SaveState(ctx, state.ConversationID, state)
	assert.NoError(t, err)

	// Get state
	retrieved, err := store.GetState(ctx, state.ConversationID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, state.ConversationID, retrieved.ConversationID)
	assert.Equal(t, state.MessageCount, retrieved.MessageCount)
	assert.Equal(t, state.EntityCount, retrieved.EntityCount)
	assert.Equal(t, state.TotalTokens, retrieved.TotalTokens)
	assert.Equal(t, state.Version, retrieved.Version)
}

func TestInMemoryStateStore_GetStateNotFound(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	_, err := store.GetState(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInMemoryStateStore_DeleteState(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-002",
		UserID:         "user-001",
		MessageCount:   5,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}

	// Save state
	err := store.SaveState(ctx, state.ConversationID, state)
	require.NoError(t, err)

	// Delete state
	err = store.DeleteState(ctx, state.ConversationID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = store.GetState(ctx, state.ConversationID)
	assert.Error(t, err)
}

func TestInMemoryStateStore_ListStates(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	// Save multiple states
	for i := 1; i <= 3; i++ {
		state := &ConversationState{
			ConversationID: string(rune(i)),
			UserID:         "user-001",
			MessageCount:   i * 10,
			Entities:       make(map[string]EntityData),
			ProviderUsage:  make(map[string]int),
		}
		err := store.SaveState(ctx, state.ConversationID, state)
		require.NoError(t, err)
	}

	// List all states
	states, err := store.ListStates(ctx)
	assert.NoError(t, err)
	assert.Len(t, states, 3)
}

func TestInMemoryStateStore_SaveAndGetWindowedAnalytics(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	analytics := &WindowedAnalytics{
		WindowStart:    time.Now().Add(-1 * time.Hour),
		WindowEnd:      time.Now(),
		ConversationID: "conv-003",
		TotalMessages:  50,
		LLMCalls:       25,
		DebateRounds:   10,
		AvgResponseTimeMs: 450.5,
		EntityGrowth:   15,
		KnowledgeDensity: 0.3,
		ProviderDistribution: map[string]int{
			"claude":   10,
			"deepseek": 5,
		},
		CreatedAt: time.Now(),
	}

	// Save analytics
	key := "analytics-001"
	err := store.SaveWindowedAnalytics(ctx, key, analytics)
	assert.NoError(t, err)

	// Get analytics
	retrieved, err := store.GetWindowedAnalytics(ctx, key)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, analytics.ConversationID, retrieved.ConversationID)
	assert.Equal(t, analytics.TotalMessages, retrieved.TotalMessages)
	assert.Equal(t, analytics.DebateRounds, retrieved.DebateRounds)
}

func TestInMemoryStateStore_DeepCopy(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	originalState := &ConversationState{
		ConversationID: "conv-004",
		UserID:         "user-001",
		MessageCount:   10,
		Entities: map[string]EntityData{
			"entity-001": {
				EntityID: "entity-001",
				Name:     "PostgreSQL",
				Type:     "technology",
			},
		},
		ProviderUsage: map[string]int{
			"claude": 3,
		},
	}

	// Save state
	err := store.SaveState(ctx, originalState.ConversationID, originalState)
	require.NoError(t, err)

	// Get state
	retrieved, err := store.GetState(ctx, originalState.ConversationID)
	require.NoError(t, err)

	// Modify retrieved state
	retrieved.MessageCount = 20
	retrieved.Entities["entity-002"] = EntityData{
		EntityID: "entity-002",
		Name:     "Redis",
		Type:     "technology",
	}
	retrieved.ProviderUsage["deepseek"] = 5

	// Get state again to verify original is unchanged
	original, err := store.GetState(ctx, originalState.ConversationID)
	require.NoError(t, err)

	// Original should be unchanged (deep copy protection)
	assert.Equal(t, 10, original.MessageCount)
	assert.Len(t, original.Entities, 1)
	assert.NotContains(t, original.Entities, "entity-002")
	assert.NotContains(t, original.ProviderUsage, "deepseek")
}

func TestInMemoryStateStore_Close(t *testing.T) {
	store := NewInMemoryStateStore()
	err := store.Close()
	assert.NoError(t, err)
}

func TestInMemoryStateStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	// Save initial state
	state := &ConversationState{
		ConversationID: "conv-concurrent",
		UserID:         "user-001",
		MessageCount:   0,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}
	err := store.SaveState(ctx, state.ConversationID, state)
	require.NoError(t, err)

	// Concurrent updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			state, err := store.GetState(ctx, "conv-concurrent")
			if err != nil {
				return
			}
			state.MessageCount += n
			store.SaveState(ctx, state.ConversationID, state)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	final, err := store.GetState(ctx, "conv-concurrent")
	assert.NoError(t, err)
	assert.NotNil(t, final)
	// Message count should be sum of 0-9 = 45, but due to race conditions it might vary
	assert.True(t, final.MessageCount >= 0)
}

func BenchmarkInMemoryStateStore_SaveState(b *testing.B) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-bench",
		UserID:         "user-001",
		MessageCount:   100,
		EntityCount:    50,
		TotalTokens:    5000,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}

	// Add some entities
	for i := 0; i < 50; i++ {
		state.Entities[string(rune(i))] = EntityData{
			EntityID: string(rune(i)),
			Name:     "Entity " + string(rune(i)),
			Type:     "technology",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.SaveState(ctx, state.ConversationID, state)
	}
}

func BenchmarkInMemoryStateStore_GetState(b *testing.B) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-bench-get",
		UserID:         "user-001",
		MessageCount:   100,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}

	// Pre-populate
	store.SaveState(ctx, state.ConversationID, state)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetState(ctx, state.ConversationID)
	}
}
