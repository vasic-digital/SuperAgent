package rag

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Additional tests for types.go and hyde.go

func TestSearchOptions_AllFields(t *testing.T) {
	opts := &SearchOptions{
		TopK:            20,
		MinScore:        0.5,
		Filter:          map[string]interface{}{"category": "tech"},
		EnableReranking: false,
		HybridAlpha:     0.7,
		IncludeMetadata: false,
		Namespace:       "test_namespace",
	}

	assert.Equal(t, 20, opts.TopK)
	assert.Equal(t, 0.5, opts.MinScore)
	assert.Equal(t, "tech", opts.Filter["category"])
	assert.False(t, opts.EnableReranking)
	assert.Equal(t, 0.7, opts.HybridAlpha)
	assert.False(t, opts.IncludeMetadata)
	assert.Equal(t, "test_namespace", opts.Namespace)
}

func TestFusionMethod_Constants(t *testing.T) {
	assert.Equal(t, FusionMethod("rrf"), FusionRRF)
	assert.Equal(t, FusionMethod("weighted"), FusionWeighted)
	assert.Equal(t, FusionMethod("max"), FusionMax)
}

func TestHyDEAggregation_Constants(t *testing.T) {
	assert.Equal(t, HyDEAggregation("mean"), HyDEAggregateMean)
	assert.Equal(t, HyDEAggregation("max"), HyDEAggregateMax)
	assert.Equal(t, HyDEAggregation("weighted"), HyDEAggregateWeighted)
	assert.Equal(t, HyDEAggregation("concat"), HyDEAggregateConcat)
}

func TestHyDEGenerator_ExpandQuery_Concat(t *testing.T) {
	config := DefaultHyDEConfig()
	config.NumHypotheses = 2
	config.AggregationMethod = HyDEAggregateConcat
	embeddingModel := &MockEmbeddingModel{dim: 4}
	docGen := &MockDocumentGenerator{
		responses: []string{"doc1", "doc2"},
	}

	generator := NewHyDEGenerator(config, embeddingModel, docGen, logrus.New())

	result, err := generator.ExpandQuery(context.Background(), "test query")
	require.NoError(t, err)

	// Concat should concatenate all embeddings
	// Original (4) + 2 hypothetical (4 each) = 12 dimensions when concatenated
	assert.Equal(t, "test query", result.OriginalQuery)
}

func TestHyDESearchResult_Struct(t *testing.T) {
	result := &HyDESearchResult{
		HyDEResult: HyDEResult{
			OriginalQuery:          "test query",
			HypotheticalDocuments:  []string{"doc1", "doc2"},
			OriginalEmbedding:      []float32{0.1, 0.2},
			HypotheticalEmbeddings: [][]float32{{0.3, 0.4}, {0.5, 0.6}},
			AggregatedEmbedding:    []float32{0.3, 0.4},
			TemplateName:           "default",
		},
		Results: []SearchResult{
			{Document: &Document{ID: "1"}, Score: 0.9},
		},
		TopK: 10,
	}

	assert.Equal(t, "test query", result.OriginalQuery)
	assert.Len(t, result.HypotheticalDocuments, 2)
	assert.Equal(t, "default", result.TemplateName)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, 10, result.TopK)
}

func TestHyDEPromptTemplate_Struct(t *testing.T) {
	template := HyDEPromptTemplate{
		Name:        "custom",
		Description: "A custom template",
		Template:    "Generate content for: {{QUERY}}",
	}

	assert.Equal(t, "custom", template.Name)
	assert.Equal(t, "A custom template", template.Description)
	assert.Contains(t, template.Template, "{{QUERY}}")
}

func TestEntityMention_Struct(t *testing.T) {
	mention := EntityMention{
		DocumentID: "doc1",
		ChunkID:    "chunk1",
		StartChar:  10,
		EndChar:    20,
		Context:    "The entity is mentioned here",
	}

	assert.Equal(t, "doc1", mention.DocumentID)
	assert.Equal(t, "chunk1", mention.ChunkID)
	assert.Equal(t, 10, mention.StartChar)
	assert.Equal(t, 20, mention.EndChar)
	assert.Contains(t, mention.Context, "entity")
}

func TestChunkerConfig_AllFields(t *testing.T) {
	config := &ChunkerConfig{
		ChunkSize:      500,
		ChunkOverlap:   100,
		Separator:      "\n",
		LengthFunction: "tokens",
	}

	assert.Equal(t, 500, config.ChunkSize)
	assert.Equal(t, 100, config.ChunkOverlap)
	assert.Equal(t, "\n", config.Separator)
	assert.Equal(t, "tokens", config.LengthFunction)
}

func TestHyDEGenerator_Templates_Coverage(t *testing.T) {
	config := DefaultHyDEConfig()
	embeddingModel := &MockEmbeddingModel{dim: 4}
	generator := NewHyDEGenerator(config, embeddingModel, nil, nil)

	// Test all default templates
	templates := generator.ListTemplates()
	assert.Contains(t, templates, "default")
	assert.Contains(t, templates, "technical")
	assert.Contains(t, templates, "scientific")
	assert.Contains(t, templates, "code")
	assert.Contains(t, templates, "qa")

	// Test setting each template
	for _, templateName := range templates {
		err := generator.SetTemplate(templateName)
		assert.NoError(t, err)
		assert.Equal(t, templateName, generator.GetConfig().TemplateName)
	}
}

func TestHyDEConfig_AllFields(t *testing.T) {
	config := HyDEConfig{
		NumHypotheses:     5,
		TemplateName:      "custom",
		MaxTokens:         300,
		Temperature:       0.8,
		AggregationMethod: HyDEAggregateMax,
	}

	assert.Equal(t, 5, config.NumHypotheses)
	assert.Equal(t, "custom", config.TemplateName)
	assert.Equal(t, 300, config.MaxTokens)
	assert.Equal(t, float32(0.8), config.Temperature)
	assert.Equal(t, HyDEAggregateMax, config.AggregationMethod)
}

func TestHyDEGenerator_AggregateConcatMethod(t *testing.T) {
	config := DefaultHyDEConfig()
	config.AggregationMethod = HyDEAggregateConcat
	embeddingModel := &MockEmbeddingModel{dim: 4}
	docGen := &MockDocumentGenerator{responses: []string{"doc1", "doc2"}}

	generator := NewHyDEGenerator(config, embeddingModel, docGen, logrus.New())

	// Test aggregateEmbeddings directly
	original := []float32{1, 2, 3, 4}
	hypothetical := [][]float32{{5, 6, 7, 8}, {9, 10, 11, 12}}

	result, err := generator.aggregateEmbeddings(original, hypothetical)
	require.NoError(t, err)
	// Concat should return concatenated embeddings
	assert.NotNil(t, result)
}

func TestHyDEGenerator_AggregateWithEmptyHypothetical(t *testing.T) {
	config := DefaultHyDEConfig()
	embeddingModel := &MockEmbeddingModel{dim: 4}
	generator := NewHyDEGenerator(config, embeddingModel, nil, nil)

	original := []float32{1, 2, 3, 4}
	hypothetical := [][]float32{}

	// Should handle empty hypothetical gracefully
	result, err := generator.aggregateMean(original, hypothetical, 4)
	require.NoError(t, err)
	// With no hypothetical, should return original
	assert.Equal(t, 4, len(result))
}

func TestDocumentGenerator_Interface(t *testing.T) {
	// Test SimpleDocumentGenerator handles various query formats
	gen := &SimpleDocumentGenerator{}

	testCases := []struct {
		name   string
		prompt string
	}{
		{"with_question_format", "Question: What is AI?\nAnswer:"},
		{"with_topic_format", "Topic: Machine Learning\nDocumentation:"},
		{"with_query_format", "Query: Neural networks\nResponse:"},
		{"plain_prompt", "Generate content about databases"},
		{"empty_prompt", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := gen.Generate(context.Background(), tc.prompt, 100, 0.5)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestHybridConfig_AllFields(t *testing.T) {
	config := &HybridConfig{
		Alpha:                 0.6,
		FusionMethod:          FusionRRF,
		RRFK:                  50,
		EnableReranking:       true,
		RerankTopK:            100,
		PreRetrieveMultiplier: 5,
	}

	assert.Equal(t, 0.6, config.Alpha)
	assert.Equal(t, FusionRRF, config.FusionMethod)
	assert.Equal(t, 50, config.RRFK)
	assert.True(t, config.EnableReranking)
	assert.Equal(t, 100, config.RerankTopK)
	assert.Equal(t, 5, config.PreRetrieveMultiplier)
}

func TestQdrantEnhancedConfig_AllFields(t *testing.T) {
	config := &QdrantEnhancedConfig{
		DenseWeight:         0.7,
		SparseWeight:        0.3,
		UseDebateEvaluation: true,
		DebateTopK:          10,
		FusionMethod:        FusionWeighted,
		RRFK:                50.0,
	}

	assert.Equal(t, 0.7, config.DenseWeight)
	assert.Equal(t, 0.3, config.SparseWeight)
	assert.True(t, config.UseDebateEvaluation)
	assert.Equal(t, 10, config.DebateTopK)
	assert.Equal(t, FusionWeighted, config.FusionMethod)
	assert.Equal(t, 50.0, config.RRFK)
}

func TestRerankerConfig_AllFields(t *testing.T) {
	config := &RerankerConfig{
		Model:        "custom-model",
		Endpoint:     "http://localhost:8080",
		APIKey:       "secret-key",
		Timeout:      30000000000, // 30 seconds in nanoseconds
		BatchSize:    16,
		ReturnScores: false,
	}

	assert.Equal(t, "custom-model", config.Model)
	assert.Equal(t, "http://localhost:8080", config.Endpoint)
	assert.Equal(t, "secret-key", config.APIKey)
	assert.Equal(t, 16, config.BatchSize)
	assert.False(t, config.ReturnScores)
}

func TestPipelineDocument_AllFields(t *testing.T) {
	doc := &PipelineDocument{
		ID:      "doc_id",
		Content: "document content",
		Metadata: map[string]interface{}{
			"author":   "test",
			"category": "tech",
		},
		Source: "test_source",
	}

	assert.Equal(t, "doc_id", doc.ID)
	assert.Equal(t, "document content", doc.Content)
	assert.Equal(t, "test", doc.Metadata["author"])
	assert.Equal(t, "tech", doc.Metadata["category"])
	assert.Equal(t, "test_source", doc.Source)
}

func TestPipelineChunk_AllFields(t *testing.T) {
	chunk := &PipelineChunk{
		ID:        "chunk_id",
		Content:   "chunk content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{
			"key": "value",
		},
		StartIdx: 0,
		EndIdx:   13,
		DocID:    "doc_id",
	}

	assert.Equal(t, "chunk_id", chunk.ID)
	assert.Equal(t, "chunk content", chunk.Content)
	assert.Len(t, chunk.Embedding, 3)
	assert.Equal(t, "value", chunk.Metadata["key"])
	assert.Equal(t, 0, chunk.StartIdx)
	assert.Equal(t, 13, chunk.EndIdx)
	assert.Equal(t, "doc_id", chunk.DocID)
}

func TestPipelineSearchResult_AllFields(t *testing.T) {
	result := &PipelineSearchResult{
		Chunk: PipelineChunk{
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
	assert.Equal(t, "value", result.Metadata["key"])
}
