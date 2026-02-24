package memory

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// --- NewManager ---

func TestNewManager_AllNils(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	require.NotNil(t, manager)
	assert.NotNil(t, manager.config)
	assert.NotNil(t, manager.logger)
	assert.Nil(t, manager.extractor)
	assert.Nil(t, manager.summarizer)
	assert.Nil(t, manager.embedder)
}

func TestNewManager_NilConfig_UsesDefaults(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	assert.Equal(t, "memory", manager.config.StorageType)
	assert.Equal(t, 1536, manager.config.EmbeddingDimension)
	assert.True(t, manager.config.EnableGraph)
	assert.True(t, manager.config.EnableCompression)
}

func TestNewManager_NilLogger_CreatesDefault(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	assert.NotNil(t, manager.logger)
}

func TestNewManager_WithCustomConfig(t *testing.T) {
	config := &MemoryConfig{
		StorageType:        "postgres",
		MaxMemoriesPerUser: 5000,
		EmbeddingDimension: 3072,
		EnableGraph:        false,
	}

	manager := NewManager(NewInMemoryStore(), nil, nil, nil, config, nil)
	assert.Equal(t, "postgres", manager.config.StorageType)
	assert.Equal(t, 5000, manager.config.MaxMemoriesPerUser)
	assert.Equal(t, 3072, manager.config.EmbeddingDimension)
	assert.False(t, manager.config.EnableGraph)
}

func TestNewManager_WithCustomLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, logger)
	assert.Same(t, logger, manager.logger)
}

func TestNewManager_WithAllDependencies(t *testing.T) {
	store := NewInMemoryStore()
	ext := &mockExtractor{}
	sum := &mockSummarizer{}
	emb := &mockEmbedder{}
	cfg := DefaultMemoryConfig()
	log := logrus.New()

	manager := NewManager(store, ext, sum, emb, cfg, log)
	require.NotNil(t, manager)
	assert.Same(t, store, manager.store)
	assert.Same(t, cfg, manager.config)
	assert.Same(t, log, manager.logger)
}

// --- AddMemory ---

func TestManager_AddMemory_GeneratesID(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	mem := &Memory{UserID: "user1", Content: "Test"}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.NotEmpty(t, mem.ID)
}

func TestManager_AddMemory_PreservesExistingID(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	mem := &Memory{ID: "custom-id", UserID: "user1", Content: "Test"}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.Equal(t, "custom-id", mem.ID)
}

func TestManager_AddMemory_SetsTimestamps(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	before := time.Now()
	mem := &Memory{UserID: "user1", Content: "Test"}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)

	assert.True(t, mem.CreatedAt.After(before) || mem.CreatedAt.Equal(before))
	assert.True(t, mem.UpdatedAt.After(before) || mem.UpdatedAt.Equal(before))
	assert.True(t, mem.LastAccess.After(before) || mem.LastAccess.Equal(before))
}

func TestManager_AddMemory_GeneratesEmbedding(t *testing.T) {
	store := NewInMemoryStore()
	embedder := &mockEmbedder{}
	manager := NewManager(store, nil, nil, embedder, nil, nil)
	ctx := context.Background()

	mem := &Memory{UserID: "user1", Content: "Test content for embedding"}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.NotEmpty(t, mem.Embedding)
	assert.Len(t, mem.Embedding, 1536)
}

func TestManager_AddMemory_SkipsEmbeddingIfProvided(t *testing.T) {
	store := NewInMemoryStore()
	embedder := &mockEmbedder{
		embedFunc: func(_ context.Context, _ []string) ([][]float32, error) {
			t.Fatal("embedder should not be called")
			return nil, nil
		},
	}
	manager := NewManager(store, nil, nil, embedder, nil, nil)
	ctx := context.Background()

	mem := &Memory{
		UserID:    "user1",
		Content:   "Test",
		Embedding: []float32{0.5, 0.6, 0.7},
	}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.Equal(t, []float32{0.5, 0.6, 0.7}, mem.Embedding)
}

func TestManager_AddMemory_SkipsEmbeddingIfNoEmbedder(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	mem := &Memory{UserID: "user1", Content: "Test"}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.Empty(t, mem.Embedding)
}

func TestManager_AddMemory_HandlesEmbedderError(t *testing.T) {
	store := NewInMemoryStore()
	failingEmbedder := &mockEmbedder{
		embedFunc: func(_ context.Context, _ []string) ([][]float32, error) {
			return nil, fmt.Errorf("embedding service unavailable")
		},
	}
	manager := NewManager(store, nil, nil, failingEmbedder, nil, nil)
	ctx := context.Background()

	mem := &Memory{UserID: "user1", Content: "Test"}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err) // Should not fail
	assert.Empty(t, mem.Embedding)
}

func TestManager_AddMemory_HandlesEmptyEmbedderResult(t *testing.T) {
	store := NewInMemoryStore()
	emptyEmbedder := &mockEmbedder{
		embedFunc: func(_ context.Context, _ []string) ([][]float32, error) {
			return [][]float32{}, nil
		},
	}
	manager := NewManager(store, nil, nil, emptyEmbedder, nil, nil)
	ctx := context.Background()

	mem := &Memory{UserID: "user1", Content: "Test"}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.Empty(t, mem.Embedding)
}

func TestManager_AddMemory_CalculatesImportance(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	mem := &Memory{UserID: "user1", Content: "Short", Type: MemoryTypeEpisodic}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.InDelta(t, 0.5, mem.Importance, 0.01)
}

func TestManager_AddMemory_PreservesExistingImportance(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	mem := &Memory{UserID: "user1", Content: "Test", Importance: 0.95}
	err := manager.AddMemory(ctx, mem)
	require.NoError(t, err)
	assert.InDelta(t, 0.95, mem.Importance, 0.001)
}

// --- AddFromMessages ---

func TestManager_AddFromMessages_Success(t *testing.T) {
	store := NewInMemoryStore()
	ext := &mockExtractor{}
	cfg := &MemoryConfig{EnableGraph: true}
	manager := NewManager(store, ext, nil, nil, cfg, nil)
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
		{Role: "assistant", Content: "Hi there", Timestamp: time.Now()},
	}

	memories, err := manager.AddFromMessages(ctx, messages, "user1", "session1")
	require.NoError(t, err)
	assert.NotEmpty(t, memories)
	for _, mem := range memories {
		assert.Equal(t, "session1", mem.SessionID)
	}
}

func TestManager_AddFromMessages_NoExtractor(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	_, err := manager.AddFromMessages(ctx, nil, "user1", "session1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no extractor configured")
}

func TestManager_AddFromMessages_ExtractionError(t *testing.T) {
	ext := &mockExtractor{
		extractFunc: func(_ context.Context, _ []Message, _ string) ([]*Memory, error) {
			return nil, fmt.Errorf("extraction failed")
		},
	}
	manager := NewManager(NewInMemoryStore(), ext, nil, nil, nil, nil)
	ctx := context.Background()

	_, err := manager.AddFromMessages(ctx, nil, "user1", "session1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "extraction failed")
}

func TestManager_AddFromMessages_WithGraphDisabled(t *testing.T) {
	store := NewInMemoryStore()
	ext := &mockExtractor{
		extractEntitiesFunc: func(_ context.Context, _ string) ([]*Entity, error) {
			t.Fatal("extractEntities should not be called when graph is disabled")
			return nil, nil
		},
	}
	cfg := &MemoryConfig{EnableGraph: false}
	manager := NewManager(store, ext, nil, nil, cfg, nil)
	ctx := context.Background()

	messages := []Message{{Role: "user", Content: "Hello"}}
	_, err := manager.AddFromMessages(ctx, messages, "user1", "session1")
	require.NoError(t, err)
}

func TestManager_AddFromMessages_EntityExtractionError(t *testing.T) {
	store := NewInMemoryStore()
	ext := &mockExtractor{
		extractEntitiesFunc: func(_ context.Context, _ string) ([]*Entity, error) {
			return nil, fmt.Errorf("entity extraction failed")
		},
	}
	cfg := &MemoryConfig{EnableGraph: true}
	manager := NewManager(store, ext, nil, nil, cfg, nil)
	ctx := context.Background()

	messages := []Message{{Role: "user", Content: "Hello"}}
	memories, err := manager.AddFromMessages(ctx, messages, "user1", "session1")
	require.NoError(t, err) // Should not fail, just log
	assert.NotEmpty(t, memories)
}

func TestManager_AddFromMessages_RelationshipExtractionError(t *testing.T) {
	store := NewInMemoryStore()
	ext := &mockExtractor{
		extractRelationshipsFunc: func(_ context.Context, _ string, _ []*Entity) ([]*Relationship, error) {
			return nil, fmt.Errorf("relationship extraction failed")
		},
	}
	cfg := &MemoryConfig{EnableGraph: true}
	manager := NewManager(store, ext, nil, nil, cfg, nil)
	ctx := context.Background()

	messages := []Message{{Role: "user", Content: "Hello"}}
	memories, err := manager.AddFromMessages(ctx, messages, "user1", "session1")
	require.NoError(t, err) // Should not fail, just log
	assert.NotEmpty(t, memories)
}

// --- Search ---

func TestManager_Search_WithNilOptions(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "m1", Content: "Go programming"})

	results, err := manager.Search(ctx, "programming", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestManager_Search_WithCustomOptions(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "m1", UserID: "u1", Content: "Go programming"})
	_ = store.Add(ctx, &Memory{ID: "m2", UserID: "u2", Content: "Go programming"})

	opts := &SearchOptions{UserID: "u1", MinScore: 0, TopK: 10}
	results, err := manager.Search(ctx, "programming", opts)
	require.NoError(t, err)
	for _, r := range results {
		assert.Equal(t, "u1", r.UserID)
	}
}

func TestManager_Search_WithEmbedder(t *testing.T) {
	store := NewInMemoryStore()
	embedder := &mockEmbedder{}
	manager := NewManager(store, nil, nil, embedder, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "m1", Content: "Go programming"})

	// Should not fail even with embedder
	results, err := manager.Search(ctx, "programming", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

// --- GetContext ---

func TestManager_GetContext_ReturnsFormattedString(t *testing.T) {
	store := NewInMemoryStore()
	cfg := &MemoryConfig{EnableGraph: false}
	manager := NewManager(store, nil, nil, nil, cfg, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "m1", UserID: "u1", Content: "Go is a programming language"})
	_ = store.Add(ctx, &Memory{ID: "m2", UserID: "u1", Content: "Go was created by Google"})

	result, err := manager.GetContext(ctx, "go", "u1", 1000)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "- ")
}

func TestManager_GetContext_RespectsMaxTokens(t *testing.T) {
	store := NewInMemoryStore()
	cfg := &MemoryConfig{EnableGraph: false}
	manager := NewManager(store, nil, nil, nil, cfg, nil)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		_ = store.Add(ctx, &Memory{
			ID:      fmt.Sprintf("m%d", i),
			UserID:  "u1",
			Content: fmt.Sprintf("This is a test memory number %d with some content for context", i),
		})
	}

	// Very small token limit
	result, err := manager.GetContext(ctx, "test", "u1", 10)
	require.NoError(t, err)
	// With maxTokens=10 and approxCharsPerToken=4, that's 40 chars max
	assert.LessOrEqual(t, len(result), 100) // Allow some slack
}

func TestManager_GetContext_EmptyResults(t *testing.T) {
	store := NewInMemoryStore()
	cfg := &MemoryConfig{EnableGraph: false}
	manager := NewManager(store, nil, nil, nil, cfg, nil)
	ctx := context.Background()

	result, err := manager.GetContext(ctx, "nonexistent", "u1", 1000)
	require.NoError(t, err)
	assert.Empty(t, result)
}

// --- GetUserMemories ---

func TestManager_GetUserMemories_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = store.Add(ctx, &Memory{
			ID:      fmt.Sprintf("m%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Memory %d", i),
		})
	}

	results, err := manager.GetUserMemories(ctx, "user1", nil)
	require.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestManager_GetUserMemories_WithOptions(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		_ = store.Add(ctx, &Memory{
			ID:      fmt.Sprintf("m%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Memory %d", i),
		})
	}

	opts := &ListOptions{Limit: 3, Offset: 2}
	results, err := manager.GetUserMemories(ctx, "user1", opts)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestManager_GetUserMemories_NoMemories(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	results, err := manager.GetUserMemories(ctx, "unknown", nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

// --- GetSessionMemories ---

func TestManager_GetSessionMemories_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "m1", SessionID: "sess1", Content: "a"})
	_ = store.Add(ctx, &Memory{ID: "m2", SessionID: "sess1", Content: "b"})
	_ = store.Add(ctx, &Memory{ID: "m3", SessionID: "sess2", Content: "c"})

	results, err := manager.GetSessionMemories(ctx, "sess1")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestManager_GetSessionMemories_NoMemories(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	results, err := manager.GetSessionMemories(ctx, "unknown")
	require.NoError(t, err)
	assert.Empty(t, results)
}

// --- DeleteMemory ---

func TestManager_DeleteMemory_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "m1", UserID: "u1", Content: "To Delete"})

	err := manager.DeleteMemory(ctx, "m1")
	require.NoError(t, err)

	_, err = store.Get(ctx, "m1")
	require.Error(t, err)
}

func TestManager_DeleteMemory_NotFound(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	err := manager.DeleteMemory(ctx, "nonexistent")
	require.Error(t, err)
}

// --- GetMemory ---

func TestManager_GetMemory_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "gm-m1", Content: "Test Content"})

	mem, err := manager.GetMemory(ctx, "gm-m1")
	require.NoError(t, err)
	assert.Equal(t, "gm-m1", mem.ID)
	assert.Equal(t, "Test Content", mem.Content)
}

func TestManager_GetMemory_NotFound(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	_, err := manager.GetMemory(ctx, "nonexistent")
	require.Error(t, err)
}

// --- UpdateMemory ---

func TestManager_UpdateMemory_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "um-m1", Content: "Original"})

	before := time.Now()
	err := manager.UpdateMemory(ctx, &Memory{ID: "um-m1", Content: "Updated"})
	require.NoError(t, err)

	mem, _ := store.Get(ctx, "um-m1")
	assert.Equal(t, "Updated", mem.Content)
	assert.True(t, mem.UpdatedAt.After(before) || mem.UpdatedAt.Equal(before))
}

func TestManager_UpdateMemory_NotFound(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	err := manager.UpdateMemory(ctx, &Memory{ID: "nonexistent", Content: "test"})
	require.Error(t, err)
}

func TestManager_UpdateMemory_SetsUpdatedAt(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "um-m2", Content: "Original"})

	before := time.Now()
	mem := &Memory{ID: "um-m2", Content: "Updated"}
	err := manager.UpdateMemory(ctx, mem)
	require.NoError(t, err)
	assert.True(t, mem.UpdatedAt.After(before) || mem.UpdatedAt.Equal(before))
}

// --- DeleteUserMemories ---

func TestManager_DeleteUserMemories_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = store.Add(ctx, &Memory{
			ID:      fmt.Sprintf("dum-m%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Memory %d", i),
		})
	}

	err := manager.DeleteUserMemories(ctx, "user1")
	require.NoError(t, err)

	results, _ := store.GetByUser(ctx, "user1", nil)
	assert.Empty(t, results)
}

func TestManager_DeleteUserMemories_NoMemories(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	err := manager.DeleteUserMemories(ctx, "unknown")
	require.NoError(t, err)
}

func TestManager_DeleteUserMemories_DoesNotAffectOtherUsers(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.Add(ctx, &Memory{ID: "m1", UserID: "user1", Content: "a"})
	_ = store.Add(ctx, &Memory{ID: "m2", UserID: "user2", Content: "b"})

	err := manager.DeleteUserMemories(ctx, "user1")
	require.NoError(t, err)

	// user2 memories should still exist
	results, _ := store.GetByUser(ctx, "user2", nil)
	assert.Len(t, results, 1)
}

// --- SummarizeHistory ---

func TestManager_SummarizeHistory_Success(t *testing.T) {
	summarizer := &mockSummarizer{
		summarizeFunc: func(_ context.Context, _ []Message) (string, error) {
			return "Summary of conversation", nil
		},
	}
	manager := NewManager(NewInMemoryStore(), nil, summarizer, nil, nil, nil)
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}

	summary, err := manager.SummarizeHistory(ctx, messages)
	require.NoError(t, err)
	assert.Equal(t, "Summary of conversation", summary)
}

func TestManager_SummarizeHistory_NoSummarizer(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	_, err := manager.SummarizeHistory(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no summarizer configured")
}

func TestManager_SummarizeHistory_SummarizerError(t *testing.T) {
	summarizer := &mockSummarizer{
		summarizeFunc: func(_ context.Context, _ []Message) (string, error) {
			return "", fmt.Errorf("summarization failed")
		},
	}
	manager := NewManager(NewInMemoryStore(), nil, summarizer, nil, nil, nil)
	ctx := context.Background()

	_, err := manager.SummarizeHistory(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "summarization failed")
}

// --- SummarizeProgressively ---

func TestManager_SummarizeProgressively_Success(t *testing.T) {
	summarizer := &mockSummarizer{
		summarizeProgressiveFunc: func(_ context.Context, _ []Message, existing string) (string, error) {
			return existing + " + new info", nil
		},
	}
	manager := NewManager(NewInMemoryStore(), nil, summarizer, nil, nil, nil)
	ctx := context.Background()

	messages := []Message{{Role: "user", Content: "New message"}}
	summary, err := manager.SummarizeProgressively(ctx, messages, "Existing")
	require.NoError(t, err)
	assert.Equal(t, "Existing + new info", summary)
}

func TestManager_SummarizeProgressively_NoSummarizer(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	_, err := manager.SummarizeProgressively(ctx, nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no summarizer configured")
}

// --- GetRelatedEntities ---

func TestManager_GetRelatedEntities_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.AddEntity(ctx, &Entity{ID: "e1", Name: "Test Entity"})
	_ = store.AddEntity(ctx, &Entity{ID: "e2", Name: "Another Test"})

	entities, err := manager.GetRelatedEntities(ctx, "test", 10)
	require.NoError(t, err)
	assert.NotEmpty(t, entities)
}

func TestManager_GetRelatedEntities_NoResults(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	entities, err := manager.GetRelatedEntities(ctx, "nonexistent", 10)
	require.NoError(t, err)
	assert.Empty(t, entities)
}

// --- GetEntityRelationships ---

func TestManager_GetEntityRelationships_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	_ = store.AddRelationship(ctx, &Relationship{
		ID: "r1", SourceID: "e1", TargetID: "e2", Type: "knows",
	})

	rels, err := manager.GetEntityRelationships(ctx, "e1")
	require.NoError(t, err)
	assert.Len(t, rels, 1)
}

func TestManager_GetEntityRelationships_NoResults(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	rels, err := manager.GetEntityRelationships(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, rels)
}

// --- calculateImportance ---

func TestManager_CalculateImportance_TableDriven(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)

	longContent := strings.Repeat("word ", 25) // > 100 chars

	tests := []struct {
		name     string
		memory   *Memory
		expected float64
	}{
		{
			name:     "BaseImportance_Episodic",
			memory:   &Memory{Content: "Short", Type: MemoryTypeEpisodic},
			expected: 0.5,
		},
		{
			name:     "BaseImportance_Working",
			memory:   &Memory{Content: "Short", Type: MemoryTypeWorking},
			expected: 0.5,
		},
		{
			name:     "BaseImportance_Procedural",
			memory:   &Memory{Content: "Short", Type: MemoryTypeProcedural},
			expected: 0.5,
		},
		{
			name:     "SemanticBoost",
			memory:   &Memory{Content: "Short", Type: MemoryTypeSemantic},
			expected: 0.7,
		},
		{
			name:     "LongContentBoost",
			memory:   &Memory{Content: longContent, Type: MemoryTypeEpisodic},
			expected: 0.6,
		},
		{
			name: "EntityBoost",
			memory: &Memory{
				Content: "Short",
				Type:    MemoryTypeEpisodic,
				Metadata: map[string]interface{}{
					"entities": []interface{}{"entity1"},
				},
			},
			expected: 0.6,
		},
		{
			name: "SemanticPlusLong",
			memory: &Memory{
				Content: longContent,
				Type:    MemoryTypeSemantic,
			},
			expected: 0.8,
		},
		{
			name: "AllBoosts",
			memory: &Memory{
				Content: longContent,
				Type:    MemoryTypeSemantic,
				Metadata: map[string]interface{}{
					"entities": []interface{}{"entity1"},
				},
			},
			expected: 0.9,
		},
		{
			name:     "NilMetadata",
			memory:   &Memory{Content: "Short", Type: MemoryTypeEpisodic, Metadata: nil},
			expected: 0.5,
		},
		{
			name: "EmptyEntities",
			memory: &Memory{
				Content: "Short",
				Type:    MemoryTypeEpisodic,
				Metadata: map[string]interface{}{
					"entities": []interface{}{},
				},
			},
			expected: 0.5,
		},
		{
			name: "EntitiesWrongType",
			memory: &Memory{
				Content: "Short",
				Type:    MemoryTypeEpisodic,
				Metadata: map[string]interface{}{
					"entities": "not-a-slice",
				},
			},
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importance := manager.calculateImportance(tt.memory)
			assert.InDelta(t, tt.expected, importance, 0.01)
		})
	}
}

func TestManager_CalculateImportance_CappedAt1(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)

	// Even with all boosts, should not exceed 1.0
	longContent := strings.Repeat("word ", 25)
	mem := &Memory{
		Content: longContent,
		Type:    MemoryTypeSemantic,
		Metadata: map[string]interface{}{
			"entities": []interface{}{"e1", "e2", "e3"},
		},
	}
	importance := manager.calculateImportance(mem)
	assert.LessOrEqual(t, importance, 1.0)
}

// --- Cleanup ---

func TestManager_Cleanup_Success(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	err := manager.Cleanup(ctx)
	require.NoError(t, err)
}

// --- GetStats ---

func TestManager_GetStats_Success(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()
	now := time.Now()

	memories := []*Memory{
		{ID: "s1", UserID: "u1", Content: "a", Type: MemoryTypeSemantic, CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "s2", UserID: "u1", Content: "b", Type: MemoryTypeSemantic, CreatedAt: now.Add(-time.Hour)},
		{ID: "s3", UserID: "u1", Content: "c", Type: MemoryTypeEpisodic, CreatedAt: now},
		{ID: "s4", UserID: "u1", Content: "d", Type: MemoryTypeProcedural, CreatedAt: now.Add(-30 * time.Minute)},
	}

	for _, mem := range memories {
		_ = store.Add(ctx, mem)
	}

	stats, err := manager.GetStats(ctx, "u1")
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, 4, stats.TotalMemories)
	assert.Equal(t, 2, stats.MemoriesByType["semantic"])
	assert.Equal(t, 1, stats.MemoriesByType["episodic"])
	assert.Equal(t, 1, stats.MemoriesByType["procedural"])
	require.NotNil(t, stats.OldestMemory)
	require.NotNil(t, stats.NewestMemory)
	assert.True(t, stats.OldestMemory.Before(*stats.NewestMemory))
}

func TestManager_GetStats_EmptyUser(t *testing.T) {
	manager := NewManager(NewInMemoryStore(), nil, nil, nil, nil, nil)
	ctx := context.Background()

	stats, err := manager.GetStats(ctx, "unknown")
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalMemories)
	assert.Empty(t, stats.MemoriesByType)
	assert.Nil(t, stats.OldestMemory)
	assert.Nil(t, stats.NewestMemory)
}

func TestManager_GetStats_SingleMemory(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()
	now := time.Now()

	_ = store.Add(ctx, &Memory{
		ID: "m1", UserID: "u1", Content: "a",
		Type: MemoryTypeSemantic, CreatedAt: now,
	})

	stats, err := manager.GetStats(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalMemories)
	require.NotNil(t, stats.OldestMemory)
	require.NotNil(t, stats.NewestMemory)
	assert.Equal(t, stats.OldestMemory, stats.NewestMemory)
}

// --- Stats struct ---

func TestStats_ZeroValue(t *testing.T) {
	var stats Stats
	assert.Zero(t, stats.TotalMemories)
	assert.Nil(t, stats.MemoriesByType)
	assert.Zero(t, stats.TotalEntities)
	assert.Zero(t, stats.TotalRelations)
	assert.Nil(t, stats.OldestMemory)
	assert.Nil(t, stats.NewestMemory)
}

func TestStats_FullConstruction(t *testing.T) {
	now := time.Now()
	oldest := now.Add(-24 * time.Hour)
	stats := &Stats{
		TotalMemories: 200,
		MemoriesByType: map[string]int{
			"semantic":   100,
			"episodic":   50,
			"procedural": 30,
			"working":    20,
		},
		TotalEntities:  50,
		TotalRelations: 75,
		OldestMemory:   &oldest,
		NewestMemory:   &now,
	}

	assert.Equal(t, 200, stats.TotalMemories)
	assert.Len(t, stats.MemoriesByType, 4)
	assert.Equal(t, 100, stats.MemoriesByType["semantic"])
	assert.Equal(t, 50, stats.TotalEntities)
	assert.Equal(t, 75, stats.TotalRelations)
	assert.True(t, stats.OldestMemory.Before(*stats.NewestMemory))
}

// --- Embedder interface ---

func TestEmbedder_MockEmbed(t *testing.T) {
	emb := &mockEmbedder{}

	results, err := emb.Embed(context.Background(), []string{"hello", "world"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Len(t, results[0], 1536)
	assert.Len(t, results[1], 1536)
}

func TestEmbedder_MockEmbedQuery(t *testing.T) {
	emb := &mockEmbedder{}

	result, err := emb.EmbedQuery(context.Background(), "hello")
	require.NoError(t, err)
	assert.Len(t, result, 1536)
}

func TestEmbedder_CustomEmbed(t *testing.T) {
	emb := &mockEmbedder{
		embedFunc: func(_ context.Context, texts []string) ([][]float32, error) {
			results := make([][]float32, len(texts))
			for i := range texts {
				results[i] = []float32{float32(i) * 0.1}
			}
			return results, nil
		},
	}

	results, err := emb.Embed(context.Background(), []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.InDelta(t, 0.0, results[0][0], 0.001)
	assert.InDelta(t, 0.1, results[1][0], 0.001)
	assert.InDelta(t, 0.2, results[2][0], 0.001)
}

// --- Concurrent Manager operations ---

func TestManager_ConcurrentAddMemory(t *testing.T) {
	store := NewInMemoryStore()
	embedder := &mockEmbedder{}
	manager := NewManager(store, nil, nil, embedder, nil, nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mem := &Memory{
				UserID:  "user1",
				Content: fmt.Sprintf("Concurrent memory %d", idx),
				Type:    MemoryTypeSemantic,
			}
			err := manager.AddMemory(ctx, mem)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
	results, _ := store.GetByUser(ctx, "user1", nil)
	assert.Len(t, results, 100)
}

func TestManager_ConcurrentSearchAndAdd(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 20; i++ {
		_ = store.Add(ctx, &Memory{
			ID:      fmt.Sprintf("pre-m%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Existing memory about programming %d", i),
		})
	}

	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mem := &Memory{
				UserID:  "user1",
				Content: fmt.Sprintf("New concurrent memory %d about programming", idx),
			}
			_ = manager.AddMemory(ctx, mem)
		}(i)
	}

	// Concurrent searches
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manager.Search(ctx, "programming", nil)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestManager_ConcurrentGetStats(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 10; i++ {
		_ = store.Add(ctx, &Memory{
			ID:      fmt.Sprintf("stats-m%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Memory %d", i),
			Type:    MemoryTypeSemantic,
		})
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats, err := manager.GetStats(ctx, "user1")
			assert.NoError(t, err)
			assert.Equal(t, 10, stats.TotalMemories)
		}()
	}

	wg.Wait()
}

func TestManager_ConcurrentDeleteAndGet(t *testing.T) {
	store := NewInMemoryStore()
	manager := NewManager(store, nil, nil, nil, nil, nil)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 50; i++ {
		_ = store.Add(ctx, &Memory{
			ID:      fmt.Sprintf("cdg-m%d", i),
			UserID:  "user1",
			Content: fmt.Sprintf("Memory %d", i),
		})
	}

	var wg sync.WaitGroup

	// Concurrent deletes
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = manager.DeleteMemory(ctx, fmt.Sprintf("cdg-m%d", idx))
		}(i)
	}

	// Concurrent gets
	for i := 25; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mem, err := manager.GetMemory(ctx, fmt.Sprintf("cdg-m%d", idx))
			if err == nil {
				assert.NotNil(t, mem)
			}
		}(i)
	}

	wg.Wait()
}
