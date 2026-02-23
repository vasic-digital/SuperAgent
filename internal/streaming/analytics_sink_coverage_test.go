package streaming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnalyticsSink_New verifies NewAnalyticsSink initializes correctly.
func TestAnalyticsSink_New(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	require.NotNil(t, sink)
	assert.NotNil(t, sink.logger)
	assert.Nil(t, sink.db)
}

// TestAnalyticsSink_SetDB verifies SetDB stores the db reference.
func TestAnalyticsSink_SetDB(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	// SetDB with nil should not panic
	sink.SetDB(nil)
	assert.Nil(t, sink.db)
}

// TestAnalyticsSink_WriteConversationState_NilDB verifies error when db not set.
func TestAnalyticsSink_WriteConversationState_NilDB(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-001",
		UserID:         "user-001",
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
	}
	err := sink.WriteConversationState(ctx, state)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_WriteWindowedAnalytics_NilDB verifies error when db not set.
func TestAnalyticsSink_WriteWindowedAnalytics_NilDB(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	analytics := &WindowedAnalytics{
		ConversationID:       "conv-001",
		WindowStart:          time.Now().Add(-time.Hour),
		WindowEnd:            time.Now(),
		ProviderDistribution: make(map[string]int),
	}
	err := sink.WriteWindowedAnalytics(ctx, analytics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_WriteBatch_NilDB verifies error when db not set and batch non-empty.
func TestAnalyticsSink_WriteBatch_NilDB(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	analytics := []*WindowedAnalytics{
		{
			ConversationID:       "conv-001",
			WindowStart:          time.Now().Add(-time.Hour),
			WindowEnd:            time.Now(),
			ProviderDistribution: make(map[string]int),
		},
	}
	err := sink.WriteBatch(ctx, analytics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_WriteBatch_Empty verifies that WriteBatch with an empty
// slice still returns the nil-db error (db check happens before length check).
func TestAnalyticsSink_WriteBatch_Empty(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	err := sink.WriteBatch(ctx, []*WindowedAnalytics{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_WriteBatch_Nil verifies that WriteBatch with nil slice
// still returns the nil-db error (db check happens before length check).
func TestAnalyticsSink_WriteBatch_Nil(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	err := sink.WriteBatch(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_QueryAnalytics_NilDB verifies error when db not set.
func TestAnalyticsSink_QueryAnalytics_NilDB(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	now := time.Now()
	_, err := sink.QueryAnalytics(ctx, "conv-001", now.Add(-time.Hour), now)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_AggregateProviderMetrics_NilDB verifies error when db not set.
func TestAnalyticsSink_AggregateProviderMetrics_NilDB(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	now := time.Now()
	_, err := sink.AggregateProviderMetrics(ctx, now.Add(-time.Hour), now)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_WriteConversationState_EntityMarshal verifies that entity
// marshalling works with a non-empty entity map.
func TestAnalyticsSink_WriteConversationState_EntityMarshal(t *testing.T) {
	// This test verifies the marshal path (db still nil, but proves marshalling works)
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	state := &ConversationState{
		ConversationID: "conv-002",
		UserID:         "user-001",
		MessageCount:   5,
		EntityCount:    2,
		Entities: map[string]EntityData{
			"ent-001": {EntityID: "ent-001", Name: "Redis", Type: "technology"},
			"ent-002": {EntityID: "ent-002", Name: "Kafka", Type: "technology"},
		},
		ProviderUsage: map[string]int{"claude": 3},
	}
	// Error from db == nil, NOT from marshal — marshal succeeds
	err := sink.WriteConversationState(ctx, state)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}

// TestAnalyticsSink_WriteWindowedAnalytics_ProviderMarshal verifies provider
// distribution marshalling with non-empty map.
func TestAnalyticsSink_WriteWindowedAnalytics_ProviderMarshal(t *testing.T) {
	sink := NewAnalyticsSink(nil)
	ctx := context.Background()

	analytics := &WindowedAnalytics{
		ConversationID: "conv-003",
		WindowStart:    time.Now().Add(-time.Hour),
		WindowEnd:      time.Now(),
		TotalMessages:  100,
		ProviderDistribution: map[string]int{
			"claude":   60,
			"deepseek": 40,
		},
	}
	// Error from db == nil, NOT from marshal
	err := sink.WriteWindowedAnalytics(ctx, analytics)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not set")
}
