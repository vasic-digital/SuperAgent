package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockEmbedder struct {
	embedFunc      func(ctx context.Context, texts []string) ([][]float32, error)
	embedQueryFunc func(ctx context.Context, query string) ([]float32, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, texts)
	}
	results := make([][]float32, len(texts))
	for i := range texts {
		results[i] = make([]float32, 1536)
		for j := range results[i] {
			results[i][j] = 0.1
		}
	}
	return results, nil
}

func (m *mockEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	if m.embedQueryFunc != nil {
		return m.embedQueryFunc(ctx, query)
	}
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.1
	}
	return embedding, nil
}

type mockExtractor struct {
	extractFunc              func(ctx context.Context, messages []Message, userID string) ([]*Memory, error)
	extractEntitiesFunc      func(ctx context.Context, text string) ([]*Entity, error)
	extractRelationshipsFunc func(ctx context.Context, text string, entities []*Entity) ([]*Relationship, error)
}

func (m *mockExtractor) Extract(ctx context.Context, messages []Message, userID string) ([]*Memory, error) {
	if m.extractFunc != nil {
		return m.extractFunc(ctx, messages, userID)
	}
	return []*Memory{
		{Content: "Extracted memory", UserID: userID, Type: MemoryTypeSemantic},
	}, nil
}

func (m *mockExtractor) ExtractEntities(ctx context.Context, text string) ([]*Entity, error) {
	if m.extractEntitiesFunc != nil {
		return m.extractEntitiesFunc(ctx, text)
	}
	return []*Entity{
		{Name: "Test Entity", Type: "concept"},
	}, nil
}

func (m *mockExtractor) ExtractRelationships(ctx context.Context, text string, entities []*Entity) ([]*Relationship, error) {
	if m.extractRelationshipsFunc != nil {
		return m.extractRelationshipsFunc(ctx, text, entities)
	}
	return []*Relationship{}, nil
}

type mockSummarizer struct {
	summarizeFunc            func(ctx context.Context, messages []Message) (string, error)
	summarizeProgressiveFunc func(ctx context.Context, messages []Message, existingSummary string) (string, error)
}

func (m *mockSummarizer) Summarize(ctx context.Context, messages []Message) (string, error) {
	if m.summarizeFunc != nil {
		return m.summarizeFunc(ctx, messages)
	}
	return "Test summary", nil
}

func (m *mockSummarizer) SummarizeProgressive(ctx context.Context, messages []Message, existingSummary string) (string, error) {
	if m.summarizeProgressiveFunc != nil {
		return m.summarizeProgressiveFunc(ctx, messages, existingSummary)
	}
	return existingSummary + " Updated summary", nil
}

// Tests for types.go

func TestMemoryType(t *testing.T) {
	assert.Equal(t, MemoryType("episodic"), MemoryTypeEpisodic)
	assert.Equal(t, MemoryType("semantic"), MemoryTypeSemantic)
	assert.Equal(t, MemoryType("procedural"), MemoryTypeProcedural)
	assert.Equal(t, MemoryType("working"), MemoryTypeWorking)
}

func TestDefaultSearchOptions(t *testing.T) {
	opts := DefaultSearchOptions()

	assert.Equal(t, 10, opts.TopK)
	assert.Equal(t, 0.5, opts.MinScore)
	assert.False(t, opts.IncludeGraph)
}

func TestDefaultMemoryConfig(t *testing.T) {
	config := DefaultMemoryConfig()

	assert.Equal(t, "memory", config.StorageType)
	assert.Equal(t, "helixagent_memories", config.VectorDBCollection)
	assert.Equal(t, 1536, config.EmbeddingDimension)
	assert.Equal(t, 10000, config.MaxMemoriesPerUser)
	assert.Equal(t, time.Duration(0), config.MemoryTTL)
	assert.True(t, config.EnableGraph)
	assert.True(t, config.EnableCompression)
}

func TestMemoryStruct(t *testing.T) {
	now := time.Now()
	expires := now.Add(time.Hour)

	memory := &Memory{
		ID:          "mem1",
		UserID:      "user1",
		SessionID:   "session1",
		Content:     "Test content",
		Summary:     "Test summary",
		Type:        MemoryTypeSemantic,
		Category:    "test",
		Metadata:    map[string]interface{}{"key": "value"},
		Embedding:   []float32{0.1, 0.2, 0.3},
		Importance:  0.8,
		AccessCount: 5,
		LastAccess:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   &expires,
	}

	assert.Equal(t, "mem1", memory.ID)
	assert.Equal(t, "user1", memory.UserID)
	assert.Equal(t, "session1", memory.SessionID)
	assert.Equal(t, "Test content", memory.Content)
	assert.Equal(t, MemoryTypeSemantic, memory.Type)
	assert.Equal(t, 0.8, memory.Importance)
	assert.NotNil(t, memory.ExpiresAt)
}

func TestEntityStruct(t *testing.T) {
	entity := &Entity{
		ID:         "entity1",
		Name:       "Test Entity",
		Type:       "person",
		Properties: map[string]interface{}{"role": "developer"},
		Aliases:    []string{"alias1", "alias2"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	assert.Equal(t, "entity1", entity.ID)
	assert.Equal(t, "Test Entity", entity.Name)
	assert.Equal(t, "person", entity.Type)
	assert.Contains(t, entity.Aliases, "alias1")
}

func TestRelationshipStruct(t *testing.T) {
	rel := &Relationship{
		ID:         "rel1",
		SourceID:   "entity1",
		TargetID:   "entity2",
		Type:       "knows",
		Properties: map[string]interface{}{"since": "2024"},
		Strength:   0.9,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	assert.Equal(t, "rel1", rel.ID)
	assert.Equal(t, "entity1", rel.SourceID)
	assert.Equal(t, "entity2", rel.TargetID)
	assert.Equal(t, "knows", rel.Type)
	assert.Equal(t, 0.9, rel.Strength)
}

func TestMessageStruct(t *testing.T) {
	msg := Message{
		Role:      "user",
		Content:   "Hello",
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"source": "test"},
	}

	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Hello", msg.Content)
	assert.NotNil(t, msg.Metadata)
}

func TestSearchOptions(t *testing.T) {
	now := time.Now()
	opts := &SearchOptions{
		UserID:       "user1",
		SessionID:    "session1",
		Type:         MemoryTypeSemantic,
		Category:     "test",
		TopK:         5,
		MinScore:     0.7,
		IncludeGraph: true,
		TimeRange: &TimeRange{
			Start: now.Add(-time.Hour),
			End:   now,
		},
	}

	assert.Equal(t, "user1", opts.UserID)
	assert.Equal(t, 5, opts.TopK)
	assert.True(t, opts.IncludeGraph)
	assert.NotNil(t, opts.TimeRange)
}

func TestListOptions(t *testing.T) {
	opts := &ListOptions{
		Limit:  10,
		Offset: 5,
		SortBy: "created_at",
		Order:  "desc",
	}

	assert.Equal(t, 10, opts.Limit)
	assert.Equal(t, 5, opts.Offset)
	assert.Equal(t, "created_at", opts.SortBy)
	assert.Equal(t, "desc", opts.Order)
}

func TestMemoryConfig(t *testing.T) {
	config := &MemoryConfig{
		StorageType:        "postgres",
		VectorDBEndpoint:   "localhost:6333",
		VectorDBAPIKey:     "test-key",
		VectorDBCollection: "memories",
		EmbeddingModel:     "text-embedding-3-small",
		EmbeddingEndpoint:  "https://api.openai.com",
		EmbeddingDimension: 1536,
		MaxMemoriesPerUser: 5000,
		MemoryTTL:          24 * time.Hour,
		EnableGraph:        true,
		EnableCompression:  true,
		ExtractorModel:     "gpt-4",
		ExtractorEndpoint:  "https://api.openai.com",
		ExtractorAPIKey:    "api-key",
	}

	assert.Equal(t, "postgres", config.StorageType)
	assert.Equal(t, 1536, config.EmbeddingDimension)
	assert.True(t, config.EnableGraph)
}

// Tests for manager.go

func TestNewManager(t *testing.T) {
	t.Run("WithNilConfig", func(t *testing.T) {
		store := NewInMemoryStore()
		manager := NewManager(store, nil, nil, nil, nil, nil)

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.config)
		assert.NotNil(t, manager.logger)
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		store := NewInMemoryStore()
		config := &MemoryConfig{
			StorageType:        "postgres",
			MaxMemoriesPerUser: 5000,
		}
		logger := logrus.New()

		manager := NewManager(store, &mockExtractor{}, &mockSummarizer{}, &mockEmbedder{}, config, logger)

		assert.Equal(t, "postgres", manager.config.StorageType)
		assert.Equal(t, logger, manager.logger)
	})
}

func TestManager_AddMemory(t *testing.T) {
	store := NewInMemoryStore()
	embedder := &mockEmbedder{}
	manager := NewManager(store, nil, nil, embedder, nil, nil)

	t.Run("BasicAdd", func(t *testing.T) {
		memory := &Memory{
			UserID:  "user1",
			Content: "Test memory content",
			Type:    MemoryTypeSemantic,
		}

		err := manager.AddMemory(context.Background(), memory)
		require.NoError(t, err)

		assert.NotEmpty(t, memory.ID)
		assert.NotZero(t, memory.CreatedAt)
		assert.NotEmpty(t, memory.Embedding)
		assert.Greater(t, memory.Importance, 0.0)
	})

	t.Run("WithExistingID", func(t *testing.T) {
		memory := &Memory{
			ID:      "existing-id",
			UserID:  "user1",
			Content: "Test content",
		}

		err := manager.AddMemory(context.Background(), memory)
		require.NoError(t, err)
		assert.Equal(t, "existing-id", memory.ID)
	})

	t.Run("EmbeddingError", func(t *testing.T) {
		failingEmbedder := &mockEmbedder{
			embedFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				return nil, fmt.Errorf("embedding failed")
			},
		}
		manager := NewManager(store, nil, nil, failingEmbedder, nil, nil)

		memory := &Memory{
			UserID:  "user1",
			Content: "Test content",
		}

		// Should not fail, just log warning
		err := manager.AddMemory(context.Background(), memory)
		require.NoError(t, err)
	})
}

func TestManager_AddFromMessages(t *testing.T) {
	store := NewInMemoryStore()
	extractor := &mockExtractor{}
	config := &MemoryConfig{EnableGraph: true}
	manager := NewManager(store, extractor, nil, nil, config, nil)

	t.Run("Success", func(t *testing.T) {
		messages := []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
		}

		memories, err := manager.AddFromMessages(context.Background(), messages, "user1", "session1")
		require.NoError(t, err)
		assert.NotEmpty(t, memories)
		assert.Equal(t, "session1", memories[0].SessionID)
	})

	t.Run("NoExtractor", func(t *testing.T) {
		manager := NewManager(store, nil, nil, nil, nil, nil)

		_, err := manager.AddFromMessages(context.Background(), nil, "user1", "session1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no extractor configured")
	})

	t.Run("ExtractionError", func(t *testing.T) {
		failingExtractor := &mockExtractor{
			extractFunc: func(ctx context.Context, messages []Message, userID string) ([]*Memory, error) {
				return nil, fmt.Errorf("extraction failed")
			},
		}
		manager := NewManager(store, failingExtractor, nil, nil, nil, nil)

		_, err := manager.AddFromMessages(context.Background(), nil, "user1", "session1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "extraction failed")
	})
}

func TestManager_Search(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	// Add test memories
	memories := []*Memory{
		{ID: "1", UserID: "user1", Content: "Go programming language basics", Type: MemoryTypeSemantic, Importance: 0.8},
		{ID: "2", UserID: "user1", Content: "Python machine learning tutorial", Type: MemoryTypeSemantic, Importance: 0.7},
		{ID: "3", UserID: "user2", Content: "JavaScript web development", Type: MemoryTypeSemantic, Importance: 0.6},
	}

	for _, mem := range memories {
		_ = store.Add(context.Background(), mem)
	}

	t.Run("BasicSearch", func(t *testing.T) {
		results, err := manager.Search(context.Background(), "programming", nil)
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("WithOptions", func(t *testing.T) {
		opts := &SearchOptions{
			UserID:   "user1",
			TopK:     1,
			MinScore: 0.1,
		}

		results, err := manager.Search(context.Background(), "programming", opts)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 1)
	})
}

func TestManager_GetContext(t *testing.T) {
	store := NewInMemoryStore()
	config := &MemoryConfig{EnableGraph: false}
	manager := NewManager(store, nil, nil, nil, config, nil)

	// Add test memories
	for i := 0; i < 5; i++ {
		_ = store.Add(context.Background(), &Memory{
			ID:      fmt.Sprintf("mem%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Test memory content %d for context building", i),
		})
	}

	context_str, err := manager.GetContext(context.Background(), "test", "user1", 100)
	require.NoError(t, err)
	assert.NotEmpty(t, context_str)
}

func TestManager_GetUserMemories(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	// Add test memories
	for i := 0; i < 3; i++ {
		_ = store.Add(context.Background(), &Memory{
			ID:      fmt.Sprintf("mem%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Memory %d", i),
		})
	}

	memories, err := manager.GetUserMemories(context.Background(), "user1", nil)
	require.NoError(t, err)
	assert.Len(t, memories, 3)
}

func TestManager_GetSessionMemories(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	// Add test memories
	_ = store.Add(context.Background(), &Memory{ID: "1", SessionID: "session1", Content: "Memory 1"})
	_ = store.Add(context.Background(), &Memory{ID: "2", SessionID: "session1", Content: "Memory 2"})
	_ = store.Add(context.Background(), &Memory{ID: "3", SessionID: "session2", Content: "Memory 3"})

	memories, err := manager.GetSessionMemories(context.Background(), "session1")
	require.NoError(t, err)
	assert.Len(t, memories, 2)
}

func TestManager_DeleteMemory(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	_ = store.Add(context.Background(), &Memory{ID: "mem1", UserID: "user1", Content: "Test"})

	err := manager.DeleteMemory(context.Background(), "mem1")
	require.NoError(t, err)

	// Verify deleted
	_, err = store.Get(context.Background(), "mem1")
	require.Error(t, err)
}

func TestManager_DeleteUserMemories(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	// Add test memories
	for i := 0; i < 3; i++ {
		_ = store.Add(context.Background(), &Memory{
			ID:      fmt.Sprintf("mem%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Memory %d", i),
		})
	}

	err := manager.DeleteUserMemories(context.Background(), "user1")
	require.NoError(t, err)

	// Verify deleted
	memories, _ := store.GetByUser(context.Background(), "user1", nil)
	assert.Empty(t, memories)
}

func TestManager_SummarizeHistory(t *testing.T) {
	store := NewInMemoryStore()
	summarizer := &mockSummarizer{}
	manager := NewManager(store, nil, summarizer, nil, nil, nil)

	t.Run("Success", func(t *testing.T) {
		messages := []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi"},
		}

		summary, err := manager.SummarizeHistory(context.Background(), messages)
		require.NoError(t, err)
		assert.Equal(t, "Test summary", summary)
	})

	t.Run("NoSummarizer", func(t *testing.T) {
		manager := NewManager(store, nil, nil, nil, nil, nil)

		_, err := manager.SummarizeHistory(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no summarizer configured")
	})
}

func TestManager_SummarizeProgressively(t *testing.T) {
	store := NewInMemoryStore()
	summarizer := &mockSummarizer{}
	manager := NewManager(store, nil, summarizer, nil, nil, nil)

	t.Run("Success", func(t *testing.T) {
		messages := []Message{{Role: "user", Content: "New message"}}

		summary, err := manager.SummarizeProgressively(context.Background(), messages, "Existing summary")
		require.NoError(t, err)
		assert.Contains(t, summary, "Existing summary")
	})

	t.Run("NoSummarizer", func(t *testing.T) {
		manager := NewManager(store, nil, nil, nil, nil, nil)

		_, err := manager.SummarizeProgressively(context.Background(), nil, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no summarizer configured")
	})
}

func TestManager_GetRelatedEntities(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	// Add test entities
	_ = store.AddEntity(context.Background(), &Entity{ID: "e1", Name: "Test Entity"})
	_ = store.AddEntity(context.Background(), &Entity{ID: "e2", Name: "Another Entity"})

	entities, err := manager.GetRelatedEntities(context.Background(), "test", 10)
	require.NoError(t, err)
	assert.NotEmpty(t, entities)
}

func TestManager_GetEntityRelationships(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	// Add test relationship
	_ = store.AddRelationship(context.Background(), &Relationship{
		ID:       "r1",
		SourceID: "entity1",
		TargetID: "entity2",
		Type:     "knows",
	})

	relationships, err := manager.GetEntityRelationships(context.Background(), "entity1")
	require.NoError(t, err)
	assert.NotEmpty(t, relationships)
}

func TestManager_calculateImportance(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)

	t.Run("BaseImportance", func(t *testing.T) {
		memory := &Memory{
			Content: "Short",
			Type:    MemoryTypeEpisodic,
		}
		importance := manager.calculateImportance(memory)
		assert.Equal(t, 0.5, importance)
	})

	t.Run("SemanticBoost", func(t *testing.T) {
		memory := &Memory{
			Content: "Short",
			Type:    MemoryTypeSemantic,
		}
		importance := manager.calculateImportance(memory)
		assert.Equal(t, 0.7, importance)
	})

	t.Run("LongContentBoost", func(t *testing.T) {
		longContent := ""
		for i := 0; i < 50; i++ {
			longContent += "word "
		}
		memory := &Memory{
			Content: longContent,
			Type:    MemoryTypeEpisodic,
		}
		importance := manager.calculateImportance(memory)
		assert.Equal(t, 0.6, importance)
	})

	t.Run("EntityBoost", func(t *testing.T) {
		memory := &Memory{
			Content: "Short",
			Type:    MemoryTypeEpisodic,
			Metadata: map[string]interface{}{
				"entities": []interface{}{"entity1"},
			},
		}
		importance := manager.calculateImportance(memory)
		assert.Equal(t, 0.6, importance)
	})

	t.Run("MaxCap", func(t *testing.T) {
		longContent := ""
		for i := 0; i < 50; i++ {
			longContent += "word "
		}
		memory := &Memory{
			Content: longContent,
			Type:    MemoryTypeSemantic,
			Metadata: map[string]interface{}{
				"entities": []interface{}{"entity1"},
			},
		}
		importance := manager.calculateImportance(memory)
		// 0.5 (base) + 0.2 (semantic) + 0.1 (long) + 0.1 (entities) = 0.9
		assert.InDelta(t, 0.9, importance, 0.01)
	})
}

func TestManager_Cleanup(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	err := manager.Cleanup(context.Background())
	require.NoError(t, err)
}

func TestManager_GetStats(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)

	// Add test memories
	now := time.Now()
	memories := []*Memory{
		{ID: "1", UserID: "user1", Content: "Memory 1", Type: MemoryTypeSemantic, CreatedAt: now.Add(-time.Hour)},
		{ID: "2", UserID: "user1", Content: "Memory 2", Type: MemoryTypeEpisodic, CreatedAt: now},
	}
	for _, mem := range memories {
		_ = store.Add(context.Background(), mem)
	}

	stats, err := manager.GetStats(context.Background(), "user1")
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalMemories)
	assert.Equal(t, 1, stats.MemoriesByType["semantic"])
	assert.Equal(t, 1, stats.MemoriesByType["episodic"])
	assert.NotNil(t, stats.OldestMemory)
	assert.NotNil(t, stats.NewestMemory)
}

func TestStats(t *testing.T) {
	now := time.Now()
	stats := &Stats{
		TotalMemories:  100,
		MemoriesByType: map[string]int{"semantic": 50, "episodic": 50},
		TotalEntities:  10,
		TotalRelations: 5,
		OldestMemory:   &now,
		NewestMemory:   &now,
	}

	assert.Equal(t, 100, stats.TotalMemories)
	assert.Equal(t, 50, stats.MemoriesByType["semantic"])
}

// Tests for store_memory.go

func TestNewInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()

	assert.NotNil(t, store.memories)
	assert.NotNil(t, store.entities)
	assert.NotNil(t, store.relationships)
	assert.NotNil(t, store.userIndex)
	assert.NotNil(t, store.sessionIndex)
	assert.NotNil(t, store.entityIndex)
}

func TestInMemoryStore_CRUD(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		memory := &Memory{
			UserID:    "user1",
			SessionID: "session1",
			Content:   "Test content",
		}

		err := store.Add(ctx, memory)
		require.NoError(t, err)
		assert.NotEmpty(t, memory.ID)
	})

	t.Run("Get", func(t *testing.T) {
		_ = store.Add(ctx, &Memory{ID: "mem1", Content: "Test"})

		memory, err := store.Get(ctx, "mem1")
		require.NoError(t, err)
		assert.Equal(t, "mem1", memory.ID)
		assert.Equal(t, 1, memory.AccessCount)
	})

	t.Run("GetNotFound", func(t *testing.T) {
		_, err := store.Get(ctx, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "memory not found")
	})

	t.Run("Update", func(t *testing.T) {
		_ = store.Add(ctx, &Memory{ID: "mem2", Content: "Original"})

		err := store.Update(ctx, &Memory{ID: "mem2", Content: "Updated"})
		require.NoError(t, err)

		memory, _ := store.Get(ctx, "mem2")
		assert.Equal(t, "Updated", memory.Content)
	})

	t.Run("UpdateNotFound", func(t *testing.T) {
		err := store.Update(ctx, &Memory{ID: "nonexistent"})
		require.Error(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		_ = store.Add(ctx, &Memory{ID: "mem3", UserID: "user1", SessionID: "session1", Content: "To delete"})

		err := store.Delete(ctx, "mem3")
		require.NoError(t, err)

		_, err = store.Get(ctx, "mem3")
		require.Error(t, err)
	})

	t.Run("DeleteNotFound", func(t *testing.T) {
		err := store.Delete(ctx, "nonexistent")
		require.Error(t, err)
	})
}

func TestInMemoryStore_Search(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Add test data
	now := time.Now()
	memories := []*Memory{
		{ID: "1", UserID: "user1", Content: "Go programming language", Type: MemoryTypeSemantic, Category: "tech", CreatedAt: now},
		{ID: "2", UserID: "user1", Content: "Python machine learning", Type: MemoryTypeEpisodic, Category: "tech", CreatedAt: now.Add(-time.Hour)},
		{ID: "3", UserID: "user2", Content: "JavaScript web development", Type: MemoryTypeSemantic, Category: "web", CreatedAt: now},
	}
	for _, mem := range memories {
		_ = store.Add(ctx, mem)
	}

	t.Run("BasicSearch", func(t *testing.T) {
		results, err := store.Search(ctx, "programming", nil)
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("SearchByUser", func(t *testing.T) {
		opts := &SearchOptions{UserID: "user1", MinScore: 0}
		results, err := store.Search(ctx, "programming", opts)
		require.NoError(t, err)
		for _, r := range results {
			assert.Equal(t, "user1", r.UserID)
		}
	})

	t.Run("SearchByType", func(t *testing.T) {
		opts := &SearchOptions{Type: MemoryTypeSemantic, MinScore: 0}
		results, err := store.Search(ctx, "programming", opts)
		require.NoError(t, err)
		for _, r := range results {
			assert.Equal(t, MemoryTypeSemantic, r.Type)
		}
	})

	t.Run("SearchByCategory", func(t *testing.T) {
		opts := &SearchOptions{Category: "tech", MinScore: 0}
		results, err := store.Search(ctx, "programming", opts)
		require.NoError(t, err)
		for _, r := range results {
			assert.Equal(t, "tech", r.Category)
		}
	})

	t.Run("SearchByTimeRange", func(t *testing.T) {
		opts := &SearchOptions{
			TimeRange: &TimeRange{
				Start: now.Add(-30 * time.Minute),
				End:   now.Add(time.Minute),
			},
			MinScore: 0,
		}
		results, err := store.Search(ctx, "programming", opts)
		require.NoError(t, err)
		// Should exclude the memory from 1 hour ago
		for _, r := range results {
			assert.True(t, r.CreatedAt.After(now.Add(-30*time.Minute)))
		}
	})

	t.Run("SearchWithTopK", func(t *testing.T) {
		opts := &SearchOptions{TopK: 1, MinScore: 0}
		results, err := store.Search(ctx, "programming", opts)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 1)
	})

	t.Run("SearchNoResults", func(t *testing.T) {
		results, err := store.Search(ctx, "nonexistent term xyz", nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestInMemoryStore_GetByUser(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Add test data
	for i := 0; i < 5; i++ {
		_ = store.Add(ctx, &Memory{
			ID:         fmt.Sprintf("mem%d", i),
			UserID:     "user1",
			Content:    fmt.Sprintf("Memory %d", i),
			Importance: float64(i) / 10.0,
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Minute),
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		results, err := store.GetByUser(ctx, "user1", nil)
		require.NoError(t, err)
		assert.Len(t, results, 5)
	})

	t.Run("WithPagination", func(t *testing.T) {
		opts := &ListOptions{Limit: 2, Offset: 1}
		results, err := store.GetByUser(ctx, "user1", opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("WithSorting", func(t *testing.T) {
		opts := &ListOptions{SortBy: "importance", Order: "desc"}
		results, err := store.GetByUser(ctx, "user1", opts)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, results[0].Importance, results[1].Importance)
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		results, err := store.GetByUser(ctx, "nonexistent", nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("OffsetExceedsLength", func(t *testing.T) {
		opts := &ListOptions{Offset: 100}
		results, err := store.GetByUser(ctx, "user1", opts)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestInMemoryStore_GetBySession(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Add test data
	_ = store.Add(ctx, &Memory{ID: "1", SessionID: "session1", CreatedAt: time.Now().Add(-time.Minute)})
	_ = store.Add(ctx, &Memory{ID: "2", SessionID: "session1", CreatedAt: time.Now()})
	_ = store.Add(ctx, &Memory{ID: "3", SessionID: "session2"})

	t.Run("GetSession", func(t *testing.T) {
		results, err := store.GetBySession(ctx, "session1")
		require.NoError(t, err)
		assert.Len(t, results, 2)
		// Should be sorted by creation time
		assert.True(t, results[0].CreatedAt.Before(results[1].CreatedAt))
	})

	t.Run("NonexistentSession", func(t *testing.T) {
		results, err := store.GetBySession(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestInMemoryStore_EntityOperations(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	t.Run("AddEntity", func(t *testing.T) {
		entity := &Entity{Name: "Test Entity", Type: "concept"}
		err := store.AddEntity(ctx, entity)
		require.NoError(t, err)
		assert.NotEmpty(t, entity.ID)
		assert.NotZero(t, entity.CreatedAt)
	})

	t.Run("GetEntity", func(t *testing.T) {
		_ = store.AddEntity(ctx, &Entity{ID: "e1", Name: "Entity 1"})

		entity, err := store.GetEntity(ctx, "e1")
		require.NoError(t, err)
		assert.Equal(t, "Entity 1", entity.Name)
	})

	t.Run("GetEntityNotFound", func(t *testing.T) {
		_, err := store.GetEntity(ctx, "nonexistent")
		require.Error(t, err)
	})

	t.Run("SearchEntities", func(t *testing.T) {
		_ = store.AddEntity(ctx, &Entity{ID: "e2", Name: "Test Entity"})
		_ = store.AddEntity(ctx, &Entity{ID: "e3", Name: "Another Entity"})

		results, err := store.SearchEntities(ctx, "test", 10)
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("SearchEntitiesWithLimit", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			_ = store.AddEntity(ctx, &Entity{Name: fmt.Sprintf("Entity %d", i)})
		}

		results, err := store.SearchEntities(ctx, "entity", 2)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 2)
	})
}

func TestInMemoryStore_RelationshipOperations(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	t.Run("AddRelationship", func(t *testing.T) {
		rel := &Relationship{
			SourceID: "entity1",
			TargetID: "entity2",
			Type:     "knows",
		}
		err := store.AddRelationship(ctx, rel)
		require.NoError(t, err)
		assert.NotEmpty(t, rel.ID)
		assert.NotZero(t, rel.CreatedAt)
	})

	t.Run("GetRelationships", func(t *testing.T) {
		_ = store.AddRelationship(ctx, &Relationship{
			ID:       "r1",
			SourceID: "e1",
			TargetID: "e2",
		})
		_ = store.AddRelationship(ctx, &Relationship{
			ID:       "r2",
			SourceID: "e1",
			TargetID: "e3",
		})

		results, err := store.GetRelationships(ctx, "e1")
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("GetRelationshipsEmpty", func(t *testing.T) {
		results, err := store.GetRelationships(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestRemoveFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected []string
	}{
		{
			name:     "RemoveFirst",
			slice:    []string{"a", "b", "c"},
			item:     "a",
			expected: []string{"b", "c"},
		},
		{
			name:     "RemoveMiddle",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: []string{"a", "c"},
		},
		{
			name:     "RemoveLast",
			slice:    []string{"a", "b", "c"},
			item:     "c",
			expected: []string{"a", "b"},
		},
		{
			name:     "ItemNotFound",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "EmptySlice",
			slice:    []string{},
			item:     "a",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeFromSlice(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInMemoryStore_calculateMatchScore(t *testing.T) {
	store := NewInMemoryStore()

	t.Run("FullMatch", func(t *testing.T) {
		memory := &Memory{Content: "Go programming language"}
		score := store.calculateMatchScore("go programming", memory)
		assert.Equal(t, 1.0, score)
	})

	t.Run("PartialMatch", func(t *testing.T) {
		memory := &Memory{Content: "Go programming language"}
		score := store.calculateMatchScore("go python", memory)
		assert.Equal(t, 0.5, score)
	})

	t.Run("NoMatch", func(t *testing.T) {
		memory := &Memory{Content: "Go programming language"}
		score := store.calculateMatchScore("python rust", memory)
		assert.Equal(t, 0.0, score)
	})

	t.Run("EmptyQuery", func(t *testing.T) {
		memory := &Memory{Content: "Go programming language"}
		score := store.calculateMatchScore("", memory)
		assert.Equal(t, 0.0, score)
	})
}

func TestInMemoryStore_sortMemories(t *testing.T) {
	store := NewInMemoryStore()
	now := time.Now()

	memories := []*Memory{
		{ID: "1", CreatedAt: now, UpdatedAt: now, Importance: 0.5, AccessCount: 10},
		{ID: "2", CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour), Importance: 0.8, AccessCount: 5},
		{ID: "3", CreatedAt: now.Add(time.Hour), UpdatedAt: now.Add(time.Hour), Importance: 0.3, AccessCount: 15},
	}

	t.Run("SortByCreatedAtAsc", func(t *testing.T) {
		mems := make([]*Memory, len(memories))
		copy(mems, memories)
		store.sortMemories(mems, "created_at", "asc")
		assert.Equal(t, "2", mems[0].ID)
	})

	t.Run("SortByCreatedAtDesc", func(t *testing.T) {
		mems := make([]*Memory, len(memories))
		copy(mems, memories)
		store.sortMemories(mems, "created_at", "desc")
		assert.Equal(t, "3", mems[0].ID)
	})

	t.Run("SortByImportanceDesc", func(t *testing.T) {
		mems := make([]*Memory, len(memories))
		copy(mems, memories)
		store.sortMemories(mems, "importance", "desc")
		assert.Equal(t, "2", mems[0].ID)
	})

	t.Run("SortByAccessCountDesc", func(t *testing.T) {
		mems := make([]*Memory, len(memories))
		copy(mems, memories)
		store.sortMemories(mems, "access_count", "desc")
		assert.Equal(t, "3", mems[0].ID)
	})

	t.Run("SortByUpdatedAt", func(t *testing.T) {
		mems := make([]*Memory, len(memories))
		copy(mems, memories)
		store.sortMemories(mems, "updated_at", "asc")
		assert.Equal(t, "2", mems[0].ID)
	})

	t.Run("SortByDefault", func(t *testing.T) {
		mems := make([]*Memory, len(memories))
		copy(mems, memories)
		store.sortMemories(mems, "unknown", "asc")
		// Should default to created_at
		assert.Equal(t, "2", mems[0].ID)
	})
}
