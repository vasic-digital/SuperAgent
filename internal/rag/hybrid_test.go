package rag

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDenseRetriever implements DenseRetriever for testing
type MockDenseRetriever struct {
	results []*SearchResult
	err     error
}

func (m *MockDenseRetriever) Retrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.results, nil
}

func (m *MockDenseRetriever) Index(ctx context.Context, docs []*Document) error {
	return m.err
}

func (m *MockDenseRetriever) Delete(ctx context.Context, ids []string) error {
	return m.err
}

func (m *MockDenseRetriever) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return mock embeddings
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = []float32{0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

// MockSparseRetriever implements SparseRetriever for testing
type MockSparseRetriever struct {
	results []*SearchResult
	err     error
}

func (m *MockSparseRetriever) Retrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.results, nil
}

func (m *MockSparseRetriever) Index(ctx context.Context, docs []*Document) error {
	return m.err
}

func (m *MockSparseRetriever) Delete(ctx context.Context, ids []string) error {
	return m.err
}

func (m *MockSparseRetriever) GetTermFrequencies(ctx context.Context, docID string) (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return mock term frequencies
	return map[string]float64{"term1": 0.5, "term2": 0.3}, nil
}

// MockReranker implements Reranker for testing
type MockReranker struct {
	results []*SearchResult
	err     error
}

func (m *MockReranker) Rerank(ctx context.Context, query string, results []*SearchResult, topK int) ([]*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.results != nil {
		return m.results, nil
	}
	// Return original results limited to topK
	if len(results) > topK {
		return results[:topK], nil
	}
	return results, nil
}

func TestDefaultHybridConfig(t *testing.T) {
	config := DefaultHybridConfig()

	assert.Equal(t, 0.5, config.Alpha)
	assert.Equal(t, FusionRRF, config.FusionMethod)
	assert.Equal(t, 60, config.RRFK)
	assert.True(t, config.EnableReranking)
	assert.Equal(t, 50, config.RerankTopK)
	assert.Equal(t, 3, config.PreRetrieveMultiplier)
}

func TestNewHybridRetriever(t *testing.T) {
	denseRetriever := &MockDenseRetriever{}
	sparseRetriever := &MockSparseRetriever{}
	reranker := &MockReranker{}
	logger := logrus.New()

	t.Run("with all parameters", func(t *testing.T) {
		config := &HybridConfig{
			Alpha:        0.7,
			FusionMethod: FusionWeighted,
		}

		retriever := NewHybridRetriever(denseRetriever, sparseRetriever, reranker, config, logger)

		assert.NotNil(t, retriever)
		assert.Equal(t, 0.7, retriever.config.Alpha)
		assert.Equal(t, FusionWeighted, retriever.config.FusionMethod)
	})

	t.Run("with nil config uses defaults", func(t *testing.T) {
		retriever := NewHybridRetriever(denseRetriever, sparseRetriever, reranker, nil, nil)

		assert.NotNil(t, retriever)
		assert.Equal(t, 0.5, retriever.config.Alpha)
		assert.Equal(t, FusionRRF, retriever.config.FusionMethod)
	})
}

func TestHybridRetriever_Retrieve(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	createDoc := func(id string, score float64) *SearchResult {
		return &SearchResult{
			Document: &Document{ID: id, Content: "Content for " + id},
			Score:    score,
		}
	}

	t.Run("RRF fusion", func(t *testing.T) {
		denseResults := []*SearchResult{
			createDoc("doc1", 0.9),
			createDoc("doc2", 0.8),
			createDoc("doc3", 0.7),
		}
		sparseResults := []*SearchResult{
			createDoc("doc2", 0.95),
			createDoc("doc3", 0.85),
			createDoc("doc4", 0.75),
		}

		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			RRFK:                  60,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: sparseResults},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.NoError(t, err)
		assert.NotEmpty(t, results)
		// doc2 and doc3 appear in both, should have higher fused scores
	})

	t.Run("weighted fusion", func(t *testing.T) {
		denseResults := []*SearchResult{
			createDoc("doc1", 0.9),
			createDoc("doc2", 0.8),
		}
		sparseResults := []*SearchResult{
			createDoc("doc2", 0.95),
			createDoc("doc3", 0.85),
		}

		config := &HybridConfig{
			Alpha:                 0.7, // Favor dense
			FusionMethod:          FusionWeighted,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: sparseResults},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("max fusion", func(t *testing.T) {
		denseResults := []*SearchResult{
			createDoc("doc1", 0.9),
			createDoc("doc2", 0.5),
		}
		sparseResults := []*SearchResult{
			createDoc("doc2", 0.8),
			createDoc("doc3", 0.7),
		}

		config := &HybridConfig{
			FusionMethod:          FusionMax,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: sparseResults},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.NoError(t, err)
		assert.NotEmpty(t, results)

		// doc2 should have score 0.8 (max of 0.5 and 0.8)
		var doc2Result *SearchResult
		for _, r := range results {
			if r.Document.ID == "doc2" {
				doc2Result = r
				break
			}
		}
		require.NotNil(t, doc2Result)
		assert.Equal(t, 0.8, doc2Result.Score)
	})

	t.Run("with reranking enabled", func(t *testing.T) {
		denseResults := []*SearchResult{createDoc("doc1", 0.9)}
		sparseResults := []*SearchResult{createDoc("doc2", 0.8)}

		reranker := &MockReranker{}

		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			RRFK:                  60,
			EnableReranking:       true,
			RerankTopK:            50,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: sparseResults},
			reranker,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("reranking failure falls back to fused results", func(t *testing.T) {
		denseResults := []*SearchResult{createDoc("doc1", 0.9)}
		sparseResults := []*SearchResult{createDoc("doc2", 0.8)}

		reranker := &MockReranker{err: errors.New("reranking failed")}

		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			EnableReranking:       true,
			RerankTopK:            50,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: sparseResults},
			reranker,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("both retrievers fail", func(t *testing.T) {
		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{err: errors.New("dense failed")},
			&MockSparseRetriever{err: errors.New("sparse failed")},
			nil,
			config,
			logger,
		)

		_, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "both retrievers failed")
	})

	t.Run("only dense fails", func(t *testing.T) {
		sparseResults := []*SearchResult{createDoc("doc1", 0.9)}

		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{err: errors.New("dense failed")},
			&MockSparseRetriever{results: sparseResults},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("only sparse fails", func(t *testing.T) {
		denseResults := []*SearchResult{createDoc("doc1", 0.9)}

		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{err: errors.New("sparse failed")},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{TopK: 10})

		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("with min score filter", func(t *testing.T) {
		denseResults := []*SearchResult{
			createDoc("doc1", 0.9),
			createDoc("doc2", 0.3),
		}

		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: []*SearchResult{}},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", &SearchOptions{
			TopK:     10,
			MinScore: 0.02, // RRF scores are typically small
		})

		require.NoError(t, err)
		// Should filter out low score results
		for _, r := range results {
			assert.GreaterOrEqual(t, r.Score, 0.02)
		}
	})

	t.Run("with nil options uses defaults", func(t *testing.T) {
		denseResults := []*SearchResult{createDoc("doc1", 0.9)}

		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: []*SearchResult{}},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "test query", nil)

		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})
}

func TestHybridRetriever_Index(t *testing.T) {
	logger := logrus.New()

	docs := []*Document{
		{ID: "doc1", Content: "Content 1"},
		{ID: "doc2", Content: "Content 2"},
	}

	t.Run("successful indexing", func(t *testing.T) {
		retriever := NewHybridRetriever(
			&MockDenseRetriever{},
			&MockSparseRetriever{},
			nil,
			nil,
			logger,
		)

		err := retriever.Index(context.Background(), docs)

		require.NoError(t, err)
	})

	t.Run("dense indexing fails", func(t *testing.T) {
		retriever := NewHybridRetriever(
			&MockDenseRetriever{err: errors.New("dense indexing failed")},
			&MockSparseRetriever{},
			nil,
			nil,
			logger,
		)

		err := retriever.Index(context.Background(), docs)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "dense indexing failed")
	})

	t.Run("sparse indexing fails", func(t *testing.T) {
		retriever := NewHybridRetriever(
			&MockDenseRetriever{},
			&MockSparseRetriever{err: errors.New("sparse indexing failed")},
			nil,
			nil,
			logger,
		)

		err := retriever.Index(context.Background(), docs)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "sparse indexing failed")
	})
}

func TestHybridRetriever_Delete(t *testing.T) {
	logger := logrus.New()

	ids := []string{"doc1", "doc2"}

	t.Run("successful deletion", func(t *testing.T) {
		retriever := NewHybridRetriever(
			&MockDenseRetriever{},
			&MockSparseRetriever{},
			nil,
			nil,
			logger,
		)

		err := retriever.Delete(context.Background(), ids)

		require.NoError(t, err)
	})

	t.Run("dense deletion fails", func(t *testing.T) {
		retriever := NewHybridRetriever(
			&MockDenseRetriever{err: errors.New("dense deletion failed")},
			&MockSparseRetriever{},
			nil,
			nil,
			logger,
		)

		err := retriever.Delete(context.Background(), ids)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "dense deletion failed")
	})

	t.Run("sparse deletion fails", func(t *testing.T) {
		retriever := NewHybridRetriever(
			&MockDenseRetriever{},
			&MockSparseRetriever{err: errors.New("sparse deletion failed")},
			nil,
			nil,
			logger,
		)

		err := retriever.Delete(context.Background(), ids)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "sparse deletion failed")
	})
}

func TestFusionMethods(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	createResults := func(ids []string, scores []float64) []*SearchResult {
		results := make([]*SearchResult, len(ids))
		for i, id := range ids {
			results[i] = &SearchResult{
				Document: &Document{ID: id, Content: "Content for " + id},
				Score:    scores[i],
			}
		}
		return results
	}

	t.Run("RRF fusion combines ranks", func(t *testing.T) {
		config := &HybridConfig{
			FusionMethod:          FusionRRF,
			RRFK:                  60,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		denseResults := createResults([]string{"a", "b", "c"}, []float64{0.9, 0.8, 0.7})
		sparseResults := createResults([]string{"b", "c", "d"}, []float64{0.95, 0.85, 0.75})

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: sparseResults},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "query", &SearchOptions{TopK: 10})

		require.NoError(t, err)

		// Documents appearing in both lists should have higher scores
		scoreMap := make(map[string]float64)
		for _, r := range results {
			scoreMap[r.Document.ID] = r.Score
		}

		// b and c appear in both, should have higher combined scores
		assert.Greater(t, scoreMap["b"], scoreMap["a"])
		assert.Greater(t, scoreMap["c"], scoreMap["d"])
	})

	t.Run("Weighted fusion with high alpha favors dense", func(t *testing.T) {
		config := &HybridConfig{
			Alpha:                 0.9, // Strongly favor dense
			FusionMethod:          FusionWeighted,
			EnableReranking:       false,
			PreRetrieveMultiplier: 1,
		}

		denseResults := createResults([]string{"a"}, []float64{1.0})
		sparseResults := createResults([]string{"b"}, []float64{1.0})

		retriever := NewHybridRetriever(
			&MockDenseRetriever{results: denseResults},
			&MockSparseRetriever{results: sparseResults},
			nil,
			config,
			logger,
		)

		results, err := retriever.Retrieve(context.Background(), "query", &SearchOptions{TopK: 10})

		require.NoError(t, err)

		// Dense result should have higher score
		scoreMap := make(map[string]float64)
		for _, r := range results {
			scoreMap[r.Document.ID] = r.Score
		}

		assert.Greater(t, scoreMap["a"], scoreMap["b"])
	})
}
