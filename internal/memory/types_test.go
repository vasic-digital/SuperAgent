package memory

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// --- MemoryType constants ---

func TestMemoryType_AllValues(t *testing.T) {
	tests := []struct {
		name     string
		mt       MemoryType
		expected string
	}{
		{"Episodic", MemoryTypeEpisodic, "episodic"},
		{"Semantic", MemoryTypeSemantic, "semantic"},
		{"Procedural", MemoryTypeProcedural, "procedural"},
		{"Working", MemoryTypeWorking, "working"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, MemoryType(tt.expected), tt.mt)
			assert.Equal(t, tt.expected, string(tt.mt))
		})
	}
}

func TestMemoryType_Uniqueness(t *testing.T) {
	types := []MemoryType{
		MemoryTypeEpisodic,
		MemoryTypeSemantic,
		MemoryTypeProcedural,
		MemoryTypeWorking,
	}
	seen := make(map[MemoryType]bool)
	for _, mt := range types {
		assert.False(t, seen[mt], "duplicate MemoryType: %s", mt)
		seen[mt] = true
	}
}

// --- Memory struct ---

func TestMemory_ZeroValue(t *testing.T) {
	var m Memory
	assert.Empty(t, m.ID)
	assert.Empty(t, m.UserID)
	assert.Empty(t, m.SessionID)
	assert.Empty(t, m.Content)
	assert.Empty(t, m.Summary)
	assert.Empty(t, m.Type)
	assert.Empty(t, m.Category)
	assert.Nil(t, m.Metadata)
	assert.Nil(t, m.Embedding)
	assert.Zero(t, m.Importance)
	assert.Zero(t, m.AccessCount)
	assert.True(t, m.LastAccess.IsZero())
	assert.True(t, m.CreatedAt.IsZero())
	assert.True(t, m.UpdatedAt.IsZero())
	assert.Nil(t, m.ExpiresAt)
}

func TestMemory_FullConstruction(t *testing.T) {
	now := time.Now()
	expires := now.Add(24 * time.Hour)
	m := &Memory{
		ID:          "test-id-123",
		UserID:      "user-456",
		SessionID:   "session-789",
		Content:     "Test content for memory",
		Summary:     "A test summary",
		Type:        MemoryTypeProcedural,
		Category:    "coding",
		Metadata:    map[string]interface{}{"key1": "val1", "key2": 42},
		Embedding:   []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		Importance:  0.95,
		AccessCount: 15,
		LastAccess:  now,
		CreatedAt:   now.Add(-time.Hour),
		UpdatedAt:   now,
		ExpiresAt:   &expires,
	}

	assert.Equal(t, "test-id-123", m.ID)
	assert.Equal(t, "user-456", m.UserID)
	assert.Equal(t, "session-789", m.SessionID)
	assert.Equal(t, "Test content for memory", m.Content)
	assert.Equal(t, "A test summary", m.Summary)
	assert.Equal(t, MemoryTypeProcedural, m.Type)
	assert.Equal(t, "coding", m.Category)
	assert.Len(t, m.Metadata, 2)
	assert.Equal(t, "val1", m.Metadata["key1"])
	assert.Equal(t, 42, m.Metadata["key2"])
	assert.Len(t, m.Embedding, 5)
	assert.InDelta(t, 0.95, m.Importance, 0.001)
	assert.Equal(t, 15, m.AccessCount)
	assert.NotNil(t, m.ExpiresAt)
	assert.True(t, m.ExpiresAt.After(now))
}

func TestMemory_NilExpiresAt(t *testing.T) {
	m := &Memory{
		ID:      "mem1",
		Content: "no expiration",
	}
	assert.Nil(t, m.ExpiresAt)
}

func TestMemory_EmptyMetadata(t *testing.T) {
	m := &Memory{
		ID:       "mem1",
		Metadata: map[string]interface{}{},
	}
	assert.NotNil(t, m.Metadata)
	assert.Empty(t, m.Metadata)
}

func TestMemory_EmptyEmbedding(t *testing.T) {
	m := &Memory{
		ID:        "mem1",
		Embedding: []float32{},
	}
	assert.NotNil(t, m.Embedding)
	assert.Empty(t, m.Embedding)
}

// --- Entity struct ---

func TestEntity_ZeroValue(t *testing.T) {
	var e Entity
	assert.Empty(t, e.ID)
	assert.Empty(t, e.Name)
	assert.Empty(t, e.Type)
	assert.Nil(t, e.Properties)
	assert.Nil(t, e.Aliases)
	assert.True(t, e.CreatedAt.IsZero())
	assert.True(t, e.UpdatedAt.IsZero())
}

func TestEntity_FullConstruction(t *testing.T) {
	now := time.Now()
	e := &Entity{
		ID:         "entity-001",
		Name:       "Test Entity",
		Type:       "concept",
		Properties: map[string]interface{}{"color": "blue", "weight": 100},
		Aliases:    []string{"alias1", "alias2", "alias3"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, "entity-001", e.ID)
	assert.Equal(t, "Test Entity", e.Name)
	assert.Equal(t, "concept", e.Type)
	assert.Len(t, e.Properties, 2)
	assert.Len(t, e.Aliases, 3)
	assert.Contains(t, e.Aliases, "alias1")
	assert.Contains(t, e.Aliases, "alias2")
	assert.Contains(t, e.Aliases, "alias3")
}

func TestEntity_EntityTypes(t *testing.T) {
	types := []string{"person", "place", "thing", "concept"}
	for _, tp := range types {
		e := &Entity{ID: "e1", Name: "Test", Type: tp}
		assert.Equal(t, tp, e.Type)
	}
}

func TestEntity_NilAliases(t *testing.T) {
	e := &Entity{ID: "e1", Name: "Test"}
	assert.Nil(t, e.Aliases)
}

func TestEntity_EmptyProperties(t *testing.T) {
	e := &Entity{
		ID:         "e1",
		Properties: map[string]interface{}{},
	}
	assert.NotNil(t, e.Properties)
	assert.Empty(t, e.Properties)
}

// --- Relationship struct ---

func TestRelationship_ZeroValue(t *testing.T) {
	var r Relationship
	assert.Empty(t, r.ID)
	assert.Empty(t, r.SourceID)
	assert.Empty(t, r.TargetID)
	assert.Empty(t, r.Type)
	assert.Nil(t, r.Properties)
	assert.Zero(t, r.Strength)
	assert.True(t, r.CreatedAt.IsZero())
	assert.True(t, r.UpdatedAt.IsZero())
}

func TestRelationship_FullConstruction(t *testing.T) {
	now := time.Now()
	r := &Relationship{
		ID:         "rel-001",
		SourceID:   "entity-1",
		TargetID:   "entity-2",
		Type:       "works_at",
		Properties: map[string]interface{}{"since": "2024", "role": "engineer"},
		Strength:   0.85,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, "rel-001", r.ID)
	assert.Equal(t, "entity-1", r.SourceID)
	assert.Equal(t, "entity-2", r.TargetID)
	assert.Equal(t, "works_at", r.Type)
	assert.Len(t, r.Properties, 2)
	assert.InDelta(t, 0.85, r.Strength, 0.001)
}

func TestRelationship_StrengthBounds(t *testing.T) {
	tests := []struct {
		name     string
		strength float64
	}{
		{"Zero", 0.0},
		{"Half", 0.5},
		{"Full", 1.0},
		{"Low", 0.01},
		{"High", 0.99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Relationship{Strength: tt.strength}
			assert.Equal(t, tt.strength, r.Strength)
		})
	}
}

func TestRelationship_Types(t *testing.T) {
	relTypes := []string{"knows", "works_at", "located_in", "owns", "created_by"}
	for _, rt := range relTypes {
		r := &Relationship{ID: "r1", Type: rt}
		assert.Equal(t, rt, r.Type)
	}
}

// --- Message struct ---

func TestMessage_ZeroValue(t *testing.T) {
	var m Message
	assert.Empty(t, m.Role)
	assert.Empty(t, m.Content)
	assert.True(t, m.Timestamp.IsZero())
	assert.Nil(t, m.Metadata)
}

func TestMessage_FullConstruction(t *testing.T) {
	now := time.Now()
	msg := Message{
		Role:      "assistant",
		Content:   "Hello, how can I help you?",
		Timestamp: now,
		Metadata:  map[string]interface{}{"model": "gpt-4", "tokens": 25},
	}

	assert.Equal(t, "assistant", msg.Role)
	assert.Equal(t, "Hello, how can I help you?", msg.Content)
	assert.Equal(t, now, msg.Timestamp)
	assert.Len(t, msg.Metadata, 2)
}

func TestMessage_Roles(t *testing.T) {
	roles := []string{"user", "assistant", "system"}
	for _, role := range roles {
		msg := Message{Role: role, Content: "test"}
		assert.Equal(t, role, msg.Role)
	}
}

func TestMessage_NilMetadata(t *testing.T) {
	msg := Message{Role: "user", Content: "Hello"}
	assert.Nil(t, msg.Metadata)
}

// --- SearchOptions ---

func TestSearchOptions_ZeroValue(t *testing.T) {
	var opts SearchOptions
	assert.Empty(t, opts.UserID)
	assert.Empty(t, opts.SessionID)
	assert.Empty(t, opts.Type)
	assert.Empty(t, opts.Category)
	assert.Zero(t, opts.TopK)
	assert.Zero(t, opts.MinScore)
	assert.False(t, opts.IncludeGraph)
	assert.Nil(t, opts.TimeRange)
}

func TestSearchOptions_FullConstruction(t *testing.T) {
	now := time.Now()
	opts := &SearchOptions{
		UserID:       "user-1",
		SessionID:    "session-1",
		Type:         MemoryTypeSemantic,
		Category:     "technical",
		TopK:         20,
		MinScore:     0.75,
		IncludeGraph: true,
		TimeRange: &TimeRange{
			Start: now.Add(-24 * time.Hour),
			End:   now,
		},
	}

	assert.Equal(t, "user-1", opts.UserID)
	assert.Equal(t, "session-1", opts.SessionID)
	assert.Equal(t, MemoryTypeSemantic, opts.Type)
	assert.Equal(t, "technical", opts.Category)
	assert.Equal(t, 20, opts.TopK)
	assert.InDelta(t, 0.75, opts.MinScore, 0.001)
	assert.True(t, opts.IncludeGraph)
	require.NotNil(t, opts.TimeRange)
	assert.True(t, opts.TimeRange.Start.Before(opts.TimeRange.End))
}

// --- DefaultSearchOptions ---

func TestDefaultSearchOptions_Values(t *testing.T) {
	opts := DefaultSearchOptions()
	require.NotNil(t, opts)

	assert.Equal(t, 10, opts.TopK)
	assert.InDelta(t, 0.5, opts.MinScore, 0.001)
	assert.False(t, opts.IncludeGraph)
	assert.Empty(t, opts.UserID)
	assert.Empty(t, opts.SessionID)
	assert.Empty(t, opts.Type)
	assert.Empty(t, opts.Category)
	assert.Nil(t, opts.TimeRange)
}

func TestDefaultSearchOptions_ReturnsNewInstance(t *testing.T) {
	opts1 := DefaultSearchOptions()
	opts2 := DefaultSearchOptions()
	assert.NotSame(t, opts1, opts2)

	// Mutating one should not affect the other
	opts1.TopK = 999
	assert.Equal(t, 10, opts2.TopK)
}

// --- ListOptions ---

func TestListOptions_ZeroValue(t *testing.T) {
	var opts ListOptions
	assert.Zero(t, opts.Limit)
	assert.Zero(t, opts.Offset)
	assert.Empty(t, opts.SortBy)
	assert.Empty(t, opts.Order)
}

func TestListOptions_FullConstruction(t *testing.T) {
	opts := &ListOptions{
		Limit:  50,
		Offset: 100,
		SortBy: "updated_at",
		Order:  "asc",
	}

	assert.Equal(t, 50, opts.Limit)
	assert.Equal(t, 100, opts.Offset)
	assert.Equal(t, "updated_at", opts.SortBy)
	assert.Equal(t, "asc", opts.Order)
}

func TestListOptions_SortByValues(t *testing.T) {
	sortFields := []string{"created_at", "updated_at", "importance", "access_count"}
	for _, field := range sortFields {
		opts := &ListOptions{SortBy: field}
		assert.Equal(t, field, opts.SortBy)
	}
}

func TestListOptions_OrderValues(t *testing.T) {
	orders := []string{"asc", "desc"}
	for _, order := range orders {
		opts := &ListOptions{Order: order}
		assert.Equal(t, order, opts.Order)
	}
}

// --- TimeRange ---

func TestTimeRange_ZeroValue(t *testing.T) {
	var tr TimeRange
	assert.True(t, tr.Start.IsZero())
	assert.True(t, tr.End.IsZero())
}

func TestTimeRange_Construction(t *testing.T) {
	now := time.Now()
	tr := &TimeRange{
		Start: now.Add(-time.Hour),
		End:   now,
	}

	assert.True(t, tr.Start.Before(tr.End))
	assert.Equal(t, time.Hour, tr.End.Sub(tr.Start))
}

func TestTimeRange_SameStartAndEnd(t *testing.T) {
	now := time.Now()
	tr := &TimeRange{Start: now, End: now}
	assert.Equal(t, tr.Start, tr.End)
}

func TestTimeRange_LargeSpan(t *testing.T) {
	now := time.Now()
	tr := &TimeRange{
		Start: now.Add(-365 * 24 * time.Hour),
		End:   now,
	}
	assert.True(t, tr.End.Sub(tr.Start) > 364*24*time.Hour)
}

// --- MemoryConfig ---

func TestMemoryConfig_ZeroValue(t *testing.T) {
	var cfg MemoryConfig
	assert.Empty(t, cfg.StorageType)
	assert.Empty(t, cfg.VectorDBEndpoint)
	assert.Empty(t, cfg.VectorDBAPIKey)
	assert.Empty(t, cfg.VectorDBCollection)
	assert.Empty(t, cfg.EmbeddingModel)
	assert.Empty(t, cfg.EmbeddingEndpoint)
	assert.Zero(t, cfg.EmbeddingDimension)
	assert.Zero(t, cfg.MaxMemoriesPerUser)
	assert.Zero(t, cfg.MemoryTTL)
	assert.False(t, cfg.EnableGraph)
	assert.False(t, cfg.EnableCompression)
	assert.Empty(t, cfg.ExtractorModel)
	assert.Empty(t, cfg.ExtractorEndpoint)
	assert.Empty(t, cfg.ExtractorAPIKey)
}

func TestMemoryConfig_FullConstruction(t *testing.T) {
	cfg := &MemoryConfig{
		StorageType:        "qdrant",
		VectorDBEndpoint:   "http://localhost:6333",
		VectorDBAPIKey:     "api-key-123",
		VectorDBCollection: "test_memories",
		EmbeddingModel:     "text-embedding-3-large",
		EmbeddingEndpoint:  "https://api.openai.com/v1",
		EmbeddingDimension: 3072,
		MaxMemoriesPerUser: 20000,
		MemoryTTL:          48 * time.Hour,
		EnableGraph:        true,
		EnableCompression:  false,
		ExtractorModel:     "claude-3-opus",
		ExtractorEndpoint:  "https://api.anthropic.com",
		ExtractorAPIKey:    "sk-ant-123",
	}

	assert.Equal(t, "qdrant", cfg.StorageType)
	assert.Equal(t, "http://localhost:6333", cfg.VectorDBEndpoint)
	assert.Equal(t, "api-key-123", cfg.VectorDBAPIKey)
	assert.Equal(t, "test_memories", cfg.VectorDBCollection)
	assert.Equal(t, "text-embedding-3-large", cfg.EmbeddingModel)
	assert.Equal(t, 3072, cfg.EmbeddingDimension)
	assert.Equal(t, 20000, cfg.MaxMemoriesPerUser)
	assert.Equal(t, 48*time.Hour, cfg.MemoryTTL)
	assert.True(t, cfg.EnableGraph)
	assert.False(t, cfg.EnableCompression)
	assert.Equal(t, "claude-3-opus", cfg.ExtractorModel)
	assert.Equal(t, "sk-ant-123", cfg.ExtractorAPIKey)
}

// --- DefaultMemoryConfig ---

func TestDefaultMemoryConfig_Values(t *testing.T) {
	cfg := DefaultMemoryConfig()
	require.NotNil(t, cfg)

	assert.Equal(t, "memory", cfg.StorageType)
	assert.Equal(t, "helixagent_memories", cfg.VectorDBCollection)
	assert.Equal(t, 1536, cfg.EmbeddingDimension)
	assert.Equal(t, 10000, cfg.MaxMemoriesPerUser)
	assert.Equal(t, time.Duration(0), cfg.MemoryTTL)
	assert.True(t, cfg.EnableGraph)
	assert.True(t, cfg.EnableCompression)
	assert.Empty(t, cfg.VectorDBEndpoint)
	assert.Empty(t, cfg.VectorDBAPIKey)
	assert.Empty(t, cfg.EmbeddingModel)
	assert.Empty(t, cfg.EmbeddingEndpoint)
	assert.Empty(t, cfg.ExtractorModel)
	assert.Empty(t, cfg.ExtractorEndpoint)
	assert.Empty(t, cfg.ExtractorAPIKey)
}

func TestDefaultMemoryConfig_ReturnsNewInstance(t *testing.T) {
	cfg1 := DefaultMemoryConfig()
	cfg2 := DefaultMemoryConfig()
	assert.NotSame(t, cfg1, cfg2)

	// Mutating one should not affect the other
	cfg1.StorageType = "postgres"
	assert.Equal(t, "memory", cfg2.StorageType)
}

func TestDefaultMemoryConfig_StorageTypes(t *testing.T) {
	storageTypes := []string{"memory", "postgres", "redis", "qdrant"}
	for _, st := range storageTypes {
		cfg := DefaultMemoryConfig()
		cfg.StorageType = st
		assert.Equal(t, st, cfg.StorageType)
	}
}

// --- MemoryStore interface compliance ---

func TestInMemoryStore_ImplementsMemoryStore(t *testing.T) {
	var _ MemoryStore = (*InMemoryStore)(nil)
}

// --- Concurrent access to types ---

func TestMemory_ConcurrentMetadataAccess(t *testing.T) {
	m := &Memory{
		ID:       "mem1",
		Metadata: make(map[string]interface{}),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mu.Lock()
			m.Metadata[string(rune('a'+idx))] = idx
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	mu.Lock()
	assert.Len(t, m.Metadata, 10)
	mu.Unlock()
}

func TestSearchOptions_ConcurrentCreation(t *testing.T) {
	var wg sync.WaitGroup
	results := make([]*SearchOptions, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = DefaultSearchOptions()
		}(i)
	}

	wg.Wait()

	for i, opts := range results {
		require.NotNil(t, opts, "index %d was nil", i)
		assert.Equal(t, 10, opts.TopK)
		assert.InDelta(t, 0.5, opts.MinScore, 0.001)
	}
}

func TestDefaultMemoryConfig_ConcurrentCreation(t *testing.T) {
	var wg sync.WaitGroup
	results := make([]*MemoryConfig, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = DefaultMemoryConfig()
		}(i)
	}

	wg.Wait()

	for i, cfg := range results {
		require.NotNil(t, cfg, "index %d was nil", i)
		assert.Equal(t, "memory", cfg.StorageType)
		assert.Equal(t, 1536, cfg.EmbeddingDimension)
		assert.True(t, cfg.EnableGraph)
	}
}
