package streaming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConversationAggregator_NewNilLogger verifies nil logger defaults to nop.
func TestConversationAggregator_NewNilLogger(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	require.NotNil(t, agg)
	require.NotNil(t, agg.logger)
	require.NotNil(t, agg.stateStore)
}

// TestConversationAggregator_AddEntity_NoState verifies AddEntity returns error
// when conversation state does not exist.
func TestConversationAggregator_AddEntity_NoState(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	entity := EntityData{
		EntityID: "ent-001",
		Name:     "PostgreSQL",
		Type:     "technology",
	}
	_, err := agg.AddEntity(ctx, "nonexistent-conv", entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get conversation state")
}

// TestConversationAggregator_AddEntity_WithState verifies AddEntity succeeds
// when conversation state exists.
func TestConversationAggregator_AddEntity_WithState(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-001",
		UserID:         "user-001",
		MessageCount:   5,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	entity := EntityData{
		EntityID: "ent-001",
		Name:     "PostgreSQL",
		Type:     "technology",
	}
	result, err := agg.AddEntity(ctx, "conv-001", entity)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.EntityCount)
	assert.Contains(t, result.Entities, "ent-001")
}

// TestConversationAggregator_AddEntity_UpdateExisting verifies that adding an
// entity with the same ID overwrites the previous entry.
func TestConversationAggregator_AddEntity_UpdateExisting(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-002",
		UserID:         "user-001",
		MessageCount:   3,
		Entities: map[string]EntityData{
			"ent-001": {EntityID: "ent-001", Name: "Old Name", Type: "technology"},
		},
		ProviderUsage: make(map[string]int),
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	updated := EntityData{
		EntityID: "ent-001",
		Name:     "New Name",
		Type:     "technology",
	}
	result, err := agg.AddEntity(ctx, "conv-002", updated)
	require.NoError(t, err)
	assert.Equal(t, 1, result.EntityCount)
	assert.Equal(t, "New Name", result.Entities["ent-001"].Name)
}

// TestConversationAggregator_UpdateProviderUsage_NoState verifies
// UpdateProviderUsage returns error when conversation state does not exist.
func TestConversationAggregator_UpdateProviderUsage_NoState(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	_, err := agg.UpdateProviderUsage(ctx, "nonexistent-conv", "claude")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get conversation state")
}

// TestConversationAggregator_UpdateProviderUsage_WithState verifies provider
// usage increments correctly.
func TestConversationAggregator_UpdateProviderUsage_WithState(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-003",
		UserID:         "user-001",
		MessageCount:   2,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  map[string]int{"claude": 3},
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	result, err := agg.UpdateProviderUsage(ctx, "conv-003", "claude")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 4, result.ProviderUsage["claude"])
}

// TestConversationAggregator_UpdateProviderUsage_NewProvider verifies a new
// provider entry is created with count = 1.
func TestConversationAggregator_UpdateProviderUsage_NewProvider(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-004",
		UserID:         "user-001",
		MessageCount:   1,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	result, err := agg.UpdateProviderUsage(ctx, "conv-004", "deepseek")
	require.NoError(t, err)
	assert.Equal(t, 1, result.ProviderUsage["deepseek"])
}

// TestConversationAggregator_GetState_Found verifies GetState returns existing
// conversation state.
func TestConversationAggregator_GetState_Found(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-005",
		UserID:         "user-001",
		MessageCount:   10,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	result, err := agg.GetState(ctx, "conv-005")
	require.NoError(t, err)
	assert.Equal(t, 10, result.MessageCount)
}

// TestConversationAggregator_GetState_NotFound verifies GetState propagates
// the not-found error.
func TestConversationAggregator_GetState_NotFound(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	_, err := agg.GetState(ctx, "no-such-conv")
	assert.Error(t, err)
}

// TestConversationAggregator_AggregateWindow_NoState verifies AggregateWindow
// returns an error when conversation state does not exist.
func TestConversationAggregator_AggregateWindow_NoState(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	now := time.Now()
	_, err := agg.AggregateWindow(ctx, "nonexistent", now.Add(-time.Hour), now)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get conversation state")
}

// TestConversationAggregator_AggregateWindow_Basic verifies AggregateWindow
// produces correct analytics for a basic state.
func TestConversationAggregator_AggregateWindow_Basic(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-006",
		UserID:         "user-001",
		MessageCount:   10,
		EntityCount:    5,
		DebateRoundCount: 0,
		Entities:       make(map[string]EntityData),
		ProviderUsage: map[string]int{
			"claude":   6,
			"deepseek": 4,
		},
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	now := time.Now()
	window, err := agg.AggregateWindow(ctx, "conv-006", now.Add(-time.Hour), now)
	require.NoError(t, err)
	require.NotNil(t, window)

	assert.Equal(t, "conv-006", window.ConversationID)
	assert.Equal(t, 10, window.TotalMessages)
	assert.Equal(t, 5, window.EntityGrowth)
	assert.InDelta(t, 0.5, window.KnowledgeDensity, 0.001)
	assert.Equal(t, float64(0), window.AvgResponseTimeMs)
	assert.Equal(t, 6, window.ProviderDistribution["claude"])
	assert.Equal(t, 4, window.ProviderDistribution["deepseek"])
}

// TestConversationAggregator_AggregateWindow_WithDebateRounds verifies that
// AvgResponseTimeMs is set when DebateRoundCount > 0.
func TestConversationAggregator_AggregateWindow_WithDebateRounds(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID:   "conv-007",
		MessageCount:     4,
		EntityCount:      2,
		DebateRoundCount: 3,
		Entities:         make(map[string]EntityData),
		ProviderUsage:    make(map[string]int),
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	now := time.Now()
	window, err := agg.AggregateWindow(ctx, "conv-007", now.Add(-time.Hour), now)
	require.NoError(t, err)
	assert.Equal(t, 500.0, window.AvgResponseTimeMs) // Placeholder value
	assert.Equal(t, 3, window.DebateRounds)
}

// TestConversationAggregator_AggregateWindow_ZeroMessages verifies
// KnowledgeDensity is 0 when MessageCount is 0.
func TestConversationAggregator_AggregateWindow_ZeroMessages(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-008",
		MessageCount:   0,
		EntityCount:    3,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}
	require.NoError(t, store.SaveState(ctx, state.ConversationID, state))

	now := time.Now()
	window, err := agg.AggregateWindow(ctx, "conv-008", now.Add(-time.Hour), now)
	require.NoError(t, err)
	assert.Equal(t, float64(0), window.KnowledgeDensity)
}

// TestConversationAggregator_MergeStates_Empty verifies MergeStates returns
// nil when given no states.
func TestConversationAggregator_MergeStates_Empty(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	result := agg.MergeStates()
	assert.Nil(t, result)
}

// TestConversationAggregator_MergeStates_Single verifies MergeStates returns
// the single state unchanged.
func TestConversationAggregator_MergeStates_Single(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	state := &ConversationState{
		ConversationID: "conv-single",
		MessageCount:   7,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}
	result := agg.MergeStates(state)
	assert.Equal(t, state, result)
}

// TestConversationAggregator_MergeStates_Multiple verifies MergeStates
// correctly sums messages, tokens, entities, and provider usage.
func TestConversationAggregator_MergeStates_Multiple(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	earlier := time.Now().Add(-2 * time.Hour)
	later := time.Now().Add(-1 * time.Hour)

	s1 := &ConversationState{
		ConversationID:   "conv-multi",
		UserID:           "user-001",
		SessionID:        "session-001",
		MessageCount:     10,
		TotalTokens:      1000,
		DebateRoundCount: 2,
		EntityCount:      2,
		Entities: map[string]EntityData{
			"ent-001": {EntityID: "ent-001", Name: "Redis", Type: "tech"},
		},
		ProviderUsage: map[string]int{"claude": 5, "deepseek": 5},
		StartedAt:     earlier,
		LastUpdatedAt: earlier,
	}
	s2 := &ConversationState{
		ConversationID:   "conv-multi",
		UserID:           "user-001",
		SessionID:        "session-001",
		MessageCount:     5,
		TotalTokens:      500,
		DebateRoundCount: 1,
		EntityCount:      1,
		Entities: map[string]EntityData{
			"ent-002": {EntityID: "ent-002", Name: "Postgres", Type: "tech"},
		},
		ProviderUsage: map[string]int{"claude": 3},
		StartedAt:     later,
		LastUpdatedAt: later,
	}

	merged := agg.MergeStates(s1, s2)
	require.NotNil(t, merged)

	assert.Equal(t, 15, merged.MessageCount)
	assert.Equal(t, int64(1500), merged.TotalTokens)
	assert.Equal(t, 3, merged.DebateRoundCount)
	assert.Equal(t, 2, len(merged.Entities))
	assert.Equal(t, 8, merged.ProviderUsage["claude"])
	assert.Equal(t, 5, merged.ProviderUsage["deepseek"])
	// StartedAt should be the earliest
	assert.Equal(t, earlier.Unix(), merged.StartedAt.Unix())
	// EntityCount should reflect merged entities
	assert.Equal(t, 2, merged.EntityCount)
}

// TestConversationAggregator_CalculateKnowledgeDensity_ZeroMessages verifies
// that KnowledgeDensity is 0 when MessageCount is 0.
func TestConversationAggregator_CalculateKnowledgeDensity_ZeroMessages(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	state := &ConversationState{
		MessageCount: 0,
		EntityCount:  5,
	}
	density := agg.CalculateKnowledgeDensity(state)
	assert.Equal(t, float64(0), density)
}

// TestConversationAggregator_CalculateKnowledgeDensity_NonZero verifies
// the formula entities/messages.
func TestConversationAggregator_CalculateKnowledgeDensity_NonZero(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	state := &ConversationState{
		MessageCount: 10,
		EntityCount:  4,
	}
	density := agg.CalculateKnowledgeDensity(state)
	assert.InDelta(t, 0.4, density, 0.001)
}

// TestConversationAggregator_CalculateProviderDistribution_Empty verifies
// empty result when no provider usage.
func TestConversationAggregator_CalculateProviderDistribution_Empty(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	state := &ConversationState{
		ProviderUsage: make(map[string]int),
	}
	dist := agg.CalculateProviderDistribution(state)
	assert.Empty(t, dist)
}

// TestConversationAggregator_CalculateProviderDistribution_Single verifies
// a single provider gets 100%.
func TestConversationAggregator_CalculateProviderDistribution_Single(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	state := &ConversationState{
		ProviderUsage: map[string]int{"claude": 10},
	}
	dist := agg.CalculateProviderDistribution(state)
	assert.InDelta(t, 100.0, dist["claude"], 0.001)
}

// TestConversationAggregator_CalculateProviderDistribution_Multiple verifies
// percentages sum to 100%.
func TestConversationAggregator_CalculateProviderDistribution_Multiple(t *testing.T) {
	store := NewInMemoryStateStore()
	agg := NewConversationAggregator(store, nil)

	state := &ConversationState{
		ProviderUsage: map[string]int{
			"claude":   6,
			"deepseek": 4,
		},
	}
	dist := agg.CalculateProviderDistribution(state)

	total := 0.0
	for _, pct := range dist {
		total += pct
	}
	assert.InDelta(t, 100.0, total, 0.001)
	assert.InDelta(t, 60.0, dist["claude"], 0.001)
	assert.InDelta(t, 40.0, dist["deepseek"], 0.001)
}
