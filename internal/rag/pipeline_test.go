package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dev.helix.agent/internal/embeddings/models"
	"dev.helix.agent/internal/mcp/servers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbeddingModelForRAG implements the EmbeddingModel interface for testing
type MockEmbeddingModelForRAG struct {
	dim int
}

func (m *MockEmbeddingModelForRAG) Encode(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = make([]float32, m.dim)
		for j := 0; j < m.dim; j++ {
			embeddings[i][j] = float32(len(texts[i])%10) / 10.0
		}
	}
	return embeddings, nil
}

func (m *MockEmbeddingModelForRAG) EncodeSingle(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := m.Encode(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

func (m *MockEmbeddingModelForRAG) Name() string       { return "mock" }
func (m *MockEmbeddingModelForRAG) Dimensions() int    { return m.dim }
func (m *MockEmbeddingModelForRAG) MaxTokens() int     { return 8192 }
func (m *MockEmbeddingModelForRAG) Provider() string   { return "mock" }
func (m *MockEmbeddingModelForRAG) Health(ctx context.Context) error { return nil }
func (m *MockEmbeddingModelForRAG) Close() error       { return nil }

// Create a minimal mock embedding registry for testing
func createTestEmbeddingRegistry() *models.EmbeddingModelRegistry {
	config := models.RegistryConfig{
		FallbackChain: []string{"mock"},
	}
	registry := models.NewEmbeddingModelRegistry(config)

	// Register mock model for testing
	registry.Register("mock", &MockEmbeddingModelForRAG{dim: 384})

	return registry
}

func TestNewPipeline(t *testing.T) {
	registry := createTestEmbeddingRegistry()

	tests := []struct {
		name   string
		config PipelineConfig
	}{
		{
			name: "with_defaults",
			config: PipelineConfig{
				VectorDBType: VectorDBChroma,
			},
		},
		{
			name: "with_custom_config",
			config: PipelineConfig{
				VectorDBType:   VectorDBQdrant,
				CollectionName: "test_collection",
				ChunkingConfig: ChunkingConfig{
					ChunkSize:    1024,
					ChunkOverlap: 100,
					Separator:    "\n",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewPipeline(tt.config, registry)
			assert.NotNil(t, pipeline)
			if tt.config.CollectionName != "" {
				assert.Equal(t, tt.config.CollectionName, pipeline.config.CollectionName)
			} else {
				assert.Equal(t, "default", pipeline.config.CollectionName)
			}
		})
	}
}

func TestDefaultChunkingConfig(t *testing.T) {
	config := DefaultChunkingConfig()
	assert.Equal(t, 512, config.ChunkSize)
	assert.Equal(t, 50, config.ChunkOverlap)
	assert.Equal(t, "\n\n", config.Separator)
}

func TestPipeline_ChunkDocument(t *testing.T) {
	registry := createTestEmbeddingRegistry()

	tests := []struct {
		name          string
		config        ChunkingConfig
		docContent    string
		expectedCount int
	}{
		{
			name: "single_chunk_small_doc",
			config: ChunkingConfig{
				ChunkSize:    100,
				ChunkOverlap: 10,
				Separator:    "\n\n",
			},
			docContent:    "This is a small document.",
			expectedCount: 1,
		},
		{
			name: "multiple_chunks",
			config: ChunkingConfig{
				ChunkSize:    50,
				ChunkOverlap: 10,
				Separator:    "\n\n",
			},
			docContent:    "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.\n\nFourth paragraph.",
			expectedCount: 2, // Will be chunked into multiple parts
		},
		{
			name: "single_line_separator",
			config: ChunkingConfig{
				ChunkSize:    30,
				ChunkOverlap: 5,
				Separator:    "\n",
			},
			docContent:    "Line one\nLine two\nLine three\nLine four",
			expectedCount: 2, // Will be chunked based on line separator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewPipeline(PipelineConfig{
				VectorDBType:   VectorDBChroma,
				ChunkingConfig: tt.config,
			}, registry)

			doc := &Document{
				ID:      "test_doc",
				Content: tt.docContent,
			}

			chunks := pipeline.ChunkDocument(doc)
			assert.GreaterOrEqual(t, len(chunks), 1)

			// Verify chunk properties
			for i, chunk := range chunks {
				assert.NotEmpty(t, chunk.ID)
				assert.NotEmpty(t, chunk.Content)
				assert.Equal(t, "test_doc", chunk.DocID)
				assert.GreaterOrEqual(t, chunk.EndIdx, chunk.StartIdx)
				if i > 0 {
					// Chunks should be sequential
					assert.LessOrEqual(t, chunks[i-1].StartIdx, chunk.StartIdx)
				}
			}
		})
	}
}

func TestPipeline_ChunkDocument_LongContent(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
		ChunkingConfig: ChunkingConfig{
			ChunkSize:    100,
			ChunkOverlap: 20,
			Separator:    "\n\n",
		},
	}, registry)

	// Create long content
	var builder strings.Builder
	for i := 0; i < 20; i++ {
		builder.WriteString("This is paragraph ")
		builder.WriteString(string(rune('A' + i)))
		builder.WriteString(". It contains some text.\n\n")
	}

	doc := &Document{
		ID:      "long_doc",
		Content: builder.String(),
	}

	chunks := pipeline.ChunkDocument(doc)
	assert.Greater(t, len(chunks), 5) // Should have multiple chunks

	// Verify overlap
	for i := 1; i < len(chunks); i++ {
		// There should be some content overlap
		assert.Less(t, chunks[i].StartIdx, chunks[i-1].EndIdx+20)
	}
}

func TestPipeline_Initialize_Chroma(t *testing.T) {
	// Create mock Chroma server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/heartbeat":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]int64{"nanosecond heartbeat": 12345})
		case r.URL.Path == "/api/v1/collections" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]map[string]interface{}{})
		case r.URL.Path == "/api/v1/collections" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "coll_123",
				"name": "test_collection",
			})
		case strings.Contains(r.URL.Path, "/collections/"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "coll_123",
				"name": "test_collection",
			})
		default:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{})
		}
	}))
	defer server.Close()

	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType:   VectorDBChroma,
		CollectionName: "test_collection",
		ChromaConfig: &servers.ChromaAdapterConfig{
			BaseURL: server.URL,
		},
	}, registry)

	err := pipeline.Initialize(context.Background())
	assert.NoError(t, err)
	assert.True(t, pipeline.connected)
	assert.True(t, pipeline.initialized)

	// Close
	err = pipeline.Close()
	assert.NoError(t, err)
}

func TestPipeline_Health(t *testing.T) {
	registry := createTestEmbeddingRegistry()

	t.Run("not_connected", func(t *testing.T) {
		pipeline := NewPipeline(PipelineConfig{
			VectorDBType: VectorDBChroma,
		}, registry)

		err := pipeline.Health(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestPipeline_IngestDocument_NotConnected(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	doc := &Document{
		ID:      "test_doc",
		Content: "Test content",
	}

	err := pipeline.IngestDocument(context.Background(), doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestPipeline_Search_NotConnected(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	results, err := pipeline.Search(context.Background(), "test query", 10)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestPipeline_Close_Multiple(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	// Should not error even if not connected
	err := pipeline.Close()
	assert.NoError(t, err)

	// Should be safe to call multiple times
	err = pipeline.Close()
	assert.NoError(t, err)
}

func TestGenerateDocID(t *testing.T) {
	id1 := generateDocID("content1")
	id2 := generateDocID("content2")
	id3 := generateDocID("content1")

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Equal(t, id1, id3) // Same content should produce same ID
}

func TestGenerateChunkID(t *testing.T) {
	id1 := generateChunkID("doc1", 0)
	id2 := generateChunkID("doc1", 1)
	id3 := generateChunkID("doc2", 0)

	assert.Equal(t, "doc1_chunk_0", id1)
	assert.Equal(t, "doc1_chunk_1", id2)
	assert.Equal(t, "doc2_chunk_0", id3)
}

func TestCopyMetadata(t *testing.T) {
	t.Run("nil_metadata", func(t *testing.T) {
		result := copyMetadata(nil)
		assert.Nil(t, result)
	})

	t.Run("non_nil_metadata", func(t *testing.T) {
		original := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		result := copyMetadata(original)
		assert.Equal(t, original, result)

		// Modify copy, original should be unchanged
		result["key3"] = "value3"
		_, exists := original["key3"]
		assert.False(t, exists)
	})
}

func TestPipeline_GetStats_NotConnected(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	stats, err := pipeline.GetStats(context.Background())
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestPipeline_DeleteDocument_NotConnected(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	err := pipeline.DeleteDocument(context.Background(), "test_doc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestPipeline_VectorDBTypes(t *testing.T) {
	assert.Equal(t, VectorDBType("chroma"), VectorDBChroma)
	assert.Equal(t, VectorDBType("qdrant"), VectorDBQdrant)
	assert.Equal(t, VectorDBType("weaviate"), VectorDBWeaviate)
}

func TestDocument_Struct(t *testing.T) {
	doc := Document{
		ID:      "test_id",
		Content: "test content",
		Metadata: map[string]interface{}{
			"author": "test",
		},
		Source: "test_source",
	}

	assert.Equal(t, "test_id", doc.ID)
	assert.Equal(t, "test content", doc.Content)
	assert.Equal(t, "test", doc.Metadata["author"])
	assert.Equal(t, "test_source", doc.Source)
}

func TestChunk_Struct(t *testing.T) {
	chunk := Chunk{
		ID:        "chunk_id",
		Content:   "chunk content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Metadata:  map[string]interface{}{"key": "value"},
		StartIdx:  0,
		EndIdx:    13,
		DocID:     "doc_id",
	}

	assert.Equal(t, "chunk_id", chunk.ID)
	assert.Equal(t, "chunk content", chunk.Content)
	assert.Len(t, chunk.Embedding, 3)
	assert.Equal(t, 0, chunk.StartIdx)
	assert.Equal(t, 13, chunk.EndIdx)
	assert.Equal(t, "doc_id", chunk.DocID)
}

func TestSearchResult_Struct(t *testing.T) {
	result := SearchResult{
		Chunk: Chunk{
			ID:      "chunk_id",
			Content: "content",
		},
		Score:    0.95,
		Distance: 0.05,
		Metadata: map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "chunk_id", result.Chunk.ID)
	assert.Equal(t, float32(0.95), result.Score)
	assert.Equal(t, float32(0.05), result.Distance)
}

func TestPipeline_Initialize_UnsupportedDB(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBType("unknown"),
	}, registry)

	err := pipeline.Initialize(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestPipeline_Initialize_MissingConfig(t *testing.T) {
	registry := createTestEmbeddingRegistry()

	tests := []struct {
		name   string
		dbType VectorDBType
	}{
		{"chroma_no_config", VectorDBChroma},
		{"qdrant_no_config", VectorDBQdrant},
		{"weaviate_no_config", VectorDBWeaviate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewPipeline(PipelineConfig{
				VectorDBType: tt.dbType,
			}, registry)

			err := pipeline.Initialize(context.Background())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "config is required")
		})
	}
}

func TestPipeline_IngestDocuments(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	docs := []*Document{
		{ID: "doc1", Content: "Content 1"},
		{ID: "doc2", Content: "Content 2"},
	}

	// Should fail because not connected
	err := pipeline.IngestDocuments(context.Background(), docs)
	assert.Error(t, err)
}

func TestPipeline_SearchWithFilter_NotConnected(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	filter := map[string]interface{}{"source": "test"}
	results, err := pipeline.SearchWithFilter(context.Background(), "query", 10, filter)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestPipeline_ConvertResults(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	t.Run("convert_chroma_results", func(t *testing.T) {
		chromaResults := &servers.ChromaQueryResult{
			IDs:       [][]string{{"id1", "id2"}},
			Documents: [][]string{{"content1", "content2"}},
			Distances: [][]float32{{0.1, 0.2}},
			Metadatas: [][]map[string]interface{}{
				{{"key": "value1"}, {"key": "value2"}},
			},
		}

		results := pipeline.convertChromaResults(chromaResults)
		require.Len(t, results, 2)
		assert.Equal(t, "id1", results[0].Chunk.ID)
		assert.Equal(t, "content1", results[0].Chunk.Content)
		assert.Equal(t, float32(0.9), results[0].Score)
	})

	t.Run("convert_qdrant_results", func(t *testing.T) {
		qdrantResults := []servers.QdrantSearchResult{
			{
				ID:      "id1",
				Score:   0.9,
				Payload: map[string]interface{}{"content": "content1"},
			},
		}

		results := pipeline.convertQdrantResults(qdrantResults)
		require.Len(t, results, 1)
		assert.Equal(t, "id1", results[0].Chunk.ID)
		assert.Equal(t, float32(0.9), results[0].Score)
	})

	t.Run("convert_weaviate_results", func(t *testing.T) {
		weaviateResults := []servers.WeaviateSearchResult{
			{
				ID:         "id1",
				Distance:   0.2,
				Certainty:  0.8,
				Properties: map[string]interface{}{"content": "content1"},
			},
		}

		results := pipeline.convertWeaviateResults(weaviateResults)
		require.Len(t, results, 1)
		assert.Equal(t, "id1", results[0].Chunk.ID)
		assert.Equal(t, float32(0.8), results[0].Score)
	})
}

func TestPipeline_ChunkDocument_EmptyContent(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
		ChunkingConfig: ChunkingConfig{
			ChunkSize:    100,
			ChunkOverlap: 10,
			Separator:    "\n\n",
		},
	}, registry)

	doc := &Document{
		ID:      "empty_doc",
		Content: "",
	}

	chunks := pipeline.ChunkDocument(doc)
	assert.Len(t, chunks, 1)
	assert.Empty(t, chunks[0].Content)
}

func TestPipeline_ChunkDocument_WithMetadata(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
		ChunkingConfig: ChunkingConfig{
			ChunkSize:    100,
			ChunkOverlap: 10,
			Separator:    "\n\n",
		},
	}, registry)

	metadata := map[string]interface{}{
		"source":   "test_file.txt",
		"category": "test",
	}

	doc := &Document{
		ID:       "meta_doc",
		Content:  "First paragraph.\n\nSecond paragraph.",
		Metadata: metadata,
	}

	chunks := pipeline.ChunkDocument(doc)
	for _, chunk := range chunks {
		assert.NotNil(t, chunk.Metadata)
		assert.Equal(t, "test_file.txt", chunk.Metadata["source"])
		assert.Equal(t, "test", chunk.Metadata["category"])
	}
}

func TestPipeline_ConvertChromaResults_Empty(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBChroma,
	}, registry)

	// Test nil result
	results := pipeline.convertChromaResults(nil)
	assert.Empty(t, results)

	// Test empty result
	emptyResult := &servers.ChromaQueryResult{
		IDs: [][]string{},
	}
	results = pipeline.convertChromaResults(emptyResult)
	assert.Empty(t, results)
}

func TestPipeline_ConvertQdrantResults_IDTypes(t *testing.T) {
	registry := createTestEmbeddingRegistry()
	pipeline := NewPipeline(PipelineConfig{
		VectorDBType: VectorDBQdrant,
	}, registry)

	// Test string ID
	stringIDResults := []servers.QdrantSearchResult{
		{ID: "string_id", Score: 0.9, Payload: map[string]interface{}{"content": "test"}},
	}
	results := pipeline.convertQdrantResults(stringIDResults)
	assert.Equal(t, "string_id", results[0].Chunk.ID)

	// Test numeric ID
	numericIDResults := []servers.QdrantSearchResult{
		{ID: float64(12345), Score: 0.9, Payload: map[string]interface{}{"content": "test"}},
	}
	results = pipeline.convertQdrantResults(numericIDResults)
	assert.Equal(t, "12345", results[0].Chunk.ID)
}
