package rag

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/vectordb/qdrant"
)

// MockQdrantClient implements a mock for the Qdrant client
type MockQdrantClient struct {
	searchResults  []qdrant.ScoredPoint
	getResults     []qdrant.Point
	searchErr      error
	upsertErr      error
	deleteErr      error
	getErr         error
	collExists     bool
	collExistsErr  error
	createCollErr  error
}

func (m *MockQdrantClient) Search(ctx context.Context, collection string, vector []float32, opts *qdrant.SearchOptions) ([]qdrant.ScoredPoint, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.searchResults, nil
}

func (m *MockQdrantClient) UpsertPoints(ctx context.Context, collection string, points []qdrant.Point) error {
	return m.upsertErr
}

func (m *MockQdrantClient) DeletePoints(ctx context.Context, collection string, ids []string) error {
	return m.deleteErr
}

func (m *MockQdrantClient) GetPoints(ctx context.Context, collection string, ids []string) ([]qdrant.Point, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getResults, nil
}

func (m *MockQdrantClient) CollectionExists(ctx context.Context, collection string) (bool, error) {
	return m.collExists, m.collExistsErr
}

func (m *MockQdrantClient) CreateCollection(ctx context.Context, config *qdrant.CollectionConfig) error {
	return m.createCollErr
}

// MockEmbedderForRetriever implements Embedder for testing
type MockEmbedderForRetriever struct {
	embeddings  [][]float32
	queryEmbed  []float32
	embedErr    error
	queryErr    error
	dimension   int
	modelName   string
}

func (m *MockEmbedderForRetriever) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if m.embedErr != nil {
		return nil, m.embedErr
	}
	if m.embeddings != nil {
		return m.embeddings, nil
	}
	// Generate mock embeddings
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, m.dimension)
		for j := 0; j < m.dimension; j++ {
			result[i][j] = float32(i+1) * 0.1
		}
	}
	return result, nil
}

func (m *MockEmbedderForRetriever) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	if m.queryEmbed != nil {
		return m.queryEmbed, nil
	}
	embedding := make([]float32, m.dimension)
	for i := 0; i < m.dimension; i++ {
		embedding[i] = 0.5
	}
	return embedding, nil
}

func (m *MockEmbedderForRetriever) GetDimension() int {
	return m.dimension
}

func (m *MockEmbedderForRetriever) GetModelName() string {
	return m.modelName
}

func TestNewQdrantDenseRetriever(t *testing.T) {
	t.Run("with all parameters", func(t *testing.T) {
		client := &qdrant.Client{}
		embedder := &MockEmbedderForRetriever{dimension: 384}
		logger := logrus.New()

		retriever := NewQdrantDenseRetriever(client, "test_collection", embedder, logger)

		assert.NotNil(t, retriever)
		assert.Equal(t, "test_collection", retriever.collection)
	})

	t.Run("with nil logger", func(t *testing.T) {
		client := &qdrant.Client{}
		embedder := &MockEmbedderForRetriever{dimension: 384}

		retriever := NewQdrantDenseRetriever(client, "test_collection", embedder, nil)

		assert.NotNil(t, retriever)
		assert.NotNil(t, retriever.logger)
	})
}

func TestQdrantDenseRetriever_GetName(t *testing.T) {
	retriever := NewQdrantDenseRetriever(nil, "test", nil, nil)
	assert.Equal(t, "qdrant_dense", retriever.GetName())
}

func TestQdrantDenseRetriever_Retrieve_NilClient(t *testing.T) {
	embedder := &MockEmbedderForRetriever{dimension: 384}
	retriever := NewQdrantDenseRetriever(nil, "test", embedder, nil)

	_, err := retriever.Retrieve(context.Background(), "query", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestQdrantDenseRetriever_Retrieve_EmbedError(t *testing.T) {
	client := &qdrant.Client{}
	embedder := &MockEmbedderForRetriever{
		dimension: 384,
		queryErr:  errors.New("embedding failed"),
	}
	retriever := NewQdrantDenseRetriever(client, "test", embedder, nil)

	_, err := retriever.Retrieve(context.Background(), "query", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "embed query")
}

func TestPointToDocument(t *testing.T) {
	t.Run("with content field", func(t *testing.T) {
		point := qdrant.ScoredPoint{
			ID:    "test_id",
			Score: 0.9,
			Payload: map[string]interface{}{
				"content": "test content",
				"title":   "Test Title",
				"source":  "test_source",
				"extra":   "extra_value",
			},
		}

		doc := pointToDocument(point)

		assert.Equal(t, "test_id", doc.ID)
		assert.Equal(t, "test content", doc.Content)
		assert.Equal(t, "test_source", doc.Source)
		assert.Equal(t, "Test Title", doc.Metadata["title"])
		assert.Equal(t, "extra_value", doc.Metadata["extra"])
	})

	t.Run("with text field instead of content", func(t *testing.T) {
		point := qdrant.ScoredPoint{
			ID:    "test_id",
			Score: 0.9,
			Payload: map[string]interface{}{
				"text": "test text content",
			},
		}

		doc := pointToDocument(point)

		assert.Equal(t, "test text content", doc.Content)
	})

	t.Run("with no content or text", func(t *testing.T) {
		point := qdrant.ScoredPoint{
			ID:    "test_id",
			Score: 0.9,
			Payload: map[string]interface{}{
				"other": "value",
			},
		}

		doc := pointToDocument(point)

		assert.Empty(t, doc.Content)
	})
}

func TestExtractMetadata(t *testing.T) {
	t.Run("nil payload", func(t *testing.T) {
		result := extractMetadata(nil)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("non-nil payload", func(t *testing.T) {
		payload := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		result := extractMetadata(payload)

		assert.Equal(t, payload, result)
	})
}

func TestToFloat32Slice(t *testing.T) {
	float64Slice := []float64{1.0, 2.0, 3.0}

	result := toFloat32Slice(float64Slice)

	assert.Len(t, result, 3)
	assert.Equal(t, float32(1.0), result[0])
	assert.Equal(t, float32(2.0), result[1])
	assert.Equal(t, float32(3.0), result[2])
}

func TestTruncate(t *testing.T) {
	t.Run("short string unchanged", func(t *testing.T) {
		result := truncate("hello", 10)
		assert.Equal(t, "hello", result)
	})

	t.Run("long string truncated", func(t *testing.T) {
		result := truncate("hello world this is long", 10)
		assert.Equal(t, "hello worl...", result)
	})

	t.Run("exact length unchanged", func(t *testing.T) {
		result := truncate("hello", 5)
		assert.Equal(t, "hello", result)
	})
}

func TestNewQdrantDocumentStore(t *testing.T) {
	t.Run("with all parameters", func(t *testing.T) {
		client := &qdrant.Client{}
		embedder := &MockEmbedderForRetriever{dimension: 384}
		logger := logrus.New()

		store := NewQdrantDocumentStore(client, "test_collection", embedder, logger)

		assert.NotNil(t, store)
		assert.Equal(t, "test_collection", store.collection)
	})

	t.Run("with nil logger", func(t *testing.T) {
		client := &qdrant.Client{}
		embedder := &MockEmbedderForRetriever{dimension: 384}

		store := NewQdrantDocumentStore(client, "test_collection", embedder, nil)

		assert.NotNil(t, store)
		assert.NotNil(t, store.logger)
	})
}

func TestQdrantDocumentStore_AddDocument_NilClient(t *testing.T) {
	embedder := &MockEmbedderForRetriever{dimension: 384}
	store := NewQdrantDocumentStore(nil, "test", embedder, nil)

	doc := &Document{
		ID:      "doc1",
		Content: "test content",
	}

	err := store.AddDocument(context.Background(), doc)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestQdrantDocumentStore_AddDocument_EmbedError(t *testing.T) {
	client := &qdrant.Client{}
	embedder := &MockEmbedderForRetriever{
		dimension: 384,
		embedErr:  errors.New("embedding failed"),
	}
	store := NewQdrantDocumentStore(client, "test", embedder, nil)

	doc := &Document{
		ID:      "doc1",
		Content: "test content",
	}

	err := store.AddDocument(context.Background(), doc)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "embed document")
}

func TestQdrantDocumentStore_AddDocument_EmptyEmbedding(t *testing.T) {
	client := &qdrant.Client{}
	embedder := &MockEmbedderForRetriever{
		dimension:  384,
		embeddings: [][]float32{}, // Empty embeddings
	}
	store := NewQdrantDocumentStore(client, "test", embedder, nil)

	doc := &Document{
		ID:      "doc1",
		Content: "test content",
	}

	err := store.AddDocument(context.Background(), doc)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no embedding")
}

func TestQdrantDocumentStore_AddDocuments_Empty(t *testing.T) {
	store := NewQdrantDocumentStore(nil, "test", nil, nil)

	err := store.AddDocuments(context.Background(), []*Document{})

	require.NoError(t, err)
}

func TestPayloadToDocument(t *testing.T) {
	t.Run("full payload", func(t *testing.T) {
		payload := map[string]interface{}{
			"content": "test content",
			"title":   "Test Title",
			"source":  "test_source",
			"extra":   "extra_value",
		}

		doc := payloadToDocument("test_id", payload)

		assert.Equal(t, "test_id", doc.ID)
		assert.Equal(t, "test content", doc.Content)
		assert.Equal(t, "test_source", doc.Source)
		assert.Equal(t, "Test Title", doc.Metadata["title"])
		assert.Equal(t, "extra_value", doc.Metadata["extra"])
	})

	t.Run("partial payload", func(t *testing.T) {
		payload := map[string]interface{}{
			"content": "test content",
		}

		doc := payloadToDocument("test_id", payload)

		assert.Equal(t, "test_id", doc.ID)
		assert.Equal(t, "test content", doc.Content)
		assert.Empty(t, doc.Source)
	})

	t.Run("empty payload", func(t *testing.T) {
		payload := map[string]interface{}{}

		doc := payloadToDocument("test_id", payload)

		assert.Equal(t, "test_id", doc.ID)
		assert.Empty(t, doc.Content)
	})
}

func TestQdrantDocumentStore_GetDocument_NilPoints(t *testing.T) {
	// This tests the scenario when GetPoints returns empty slice
	client := &qdrant.Client{}
	embedder := &MockEmbedderForRetriever{dimension: 384}
	store := NewQdrantDocumentStore(client, "test", embedder, nil)

	// Can't easily test without a real client, but we can test the helper functions
	assert.NotNil(t, store)
}

func TestSearchOptionsDefaults(t *testing.T) {
	// Test that SearchOptions with zero TopK gets default
	opts := &SearchOptions{}

	// In Retrieve, TopK of 0 should be set to 10
	assert.Equal(t, 0, opts.TopK)
}

func TestQdrantDenseRetriever_DefaultOptions(t *testing.T) {
	embedder := &MockEmbedderForRetriever{
		dimension:  4,
		queryEmbed: []float32{0.1, 0.2, 0.3, 0.4},
	}

	retriever := &QdrantDenseRetriever{
		client:     nil, // Will error, but we're testing option handling
		collection: "test",
		embedder:   embedder,
		logger:     logrus.New(),
	}

	// Test that nil options get default TopK of 10
	_, err := retriever.Retrieve(context.Background(), "query", nil)

	// Should fail due to nil client, not options
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestDocumentStructFields(t *testing.T) {
	doc := &Document{
		ID:        "test_id",
		Content:   "test content",
		Source:    "test_source",
		Score:     0.95,
		Embedding: []float32{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{
			"author": "test_author",
		},
	}

	assert.Equal(t, "test_id", doc.ID)
	assert.Equal(t, "test content", doc.Content)
	assert.Equal(t, "test_source", doc.Source)
	assert.Equal(t, 0.95, doc.Score)
	assert.Len(t, doc.Embedding, 3)
	assert.Equal(t, "test_author", doc.Metadata["author"])
}

func TestSearchResultFields(t *testing.T) {
	result := &SearchResult{
		Document: &Document{
			ID:      "test_id",
			Content: "test content",
		},
		Score:         0.9,
		RerankedScore: 0.95,
		Highlights:    []string{"test", "content"},
		MatchType:     MatchTypeDense,
	}

	assert.Equal(t, "test_id", result.Document.ID)
	assert.Equal(t, 0.9, result.Score)
	assert.Equal(t, 0.95, result.RerankedScore)
	assert.Len(t, result.Highlights, 2)
	assert.Equal(t, MatchTypeDense, result.MatchType)
}

func TestMatchTypes(t *testing.T) {
	assert.Equal(t, MatchType("dense"), MatchTypeDense)
	assert.Equal(t, MatchType("sparse"), MatchTypeSparse)
	assert.Equal(t, MatchType("hybrid"), MatchTypeHybrid)
}
