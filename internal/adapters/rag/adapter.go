// Package rag provides adapters between HelixAgent's internal/rag types
// and the extracted digital.vasic.rag module.
package rag

import (
	"context"
	"fmt"

	helixrag "dev.helix.agent/internal/rag"
	modchunker "digital.vasic.rag/pkg/chunker"
	modhybrid "digital.vasic.rag/pkg/hybrid"
	modpipeline "digital.vasic.rag/pkg/pipeline"
	modreranker "digital.vasic.rag/pkg/reranker"
	modretriever "digital.vasic.rag/pkg/retriever"
)

// ChunkerAdapter adapts the module's Chunker to HelixAgent usage.
type ChunkerAdapter struct {
	chunker modchunker.Chunker
}

// NewFixedSizeChunker creates a fixed-size chunker.
func NewFixedSizeChunker(chunkSize, overlap int) *ChunkerAdapter {
	config := modchunker.Config{
		ChunkSize: chunkSize,
		Overlap:   overlap,
	}
	return &ChunkerAdapter{
		chunker: modchunker.NewFixedSizeChunker(config),
	}
}

// NewRecursiveChunker creates a recursive chunker.
func NewRecursiveChunker(maxChunkSize, overlap int) *ChunkerAdapter {
	config := modchunker.Config{
		ChunkSize:  maxChunkSize,
		Overlap:    overlap,
		Separators: []string{"\n\n", "\n", ". ", " "},
	}
	return &ChunkerAdapter{
		chunker: modchunker.NewRecursiveChunker(config),
	}
}

// NewSentenceChunker creates a sentence-based chunker.
func NewSentenceChunker(maxChunkSize int) *ChunkerAdapter {
	config := modchunker.Config{
		ChunkSize: maxChunkSize,
	}
	return &ChunkerAdapter{
		chunker: modchunker.NewSentenceChunker(config),
	}
}

// Chunk splits text into chunks.
func (a *ChunkerAdapter) Chunk(text string) []helixrag.PipelineChunk {
	modChunks := a.chunker.Chunk(text)

	result := make([]helixrag.PipelineChunk, len(modChunks))
	for i, c := range modChunks {
		result[i] = helixrag.PipelineChunk{
			Content:  c.Content,
			StartIdx: c.Start,
			EndIdx:   c.End,
		}
		if c.Metadata != nil {
			result[i].Metadata = make(map[string]interface{})
			for k, v := range c.Metadata {
				result[i].Metadata[k] = v
			}
		}
	}
	return result
}

// ChunkDocument splits a document into chunks.
func (a *ChunkerAdapter) ChunkDocument(doc *helixrag.PipelineDocument) []helixrag.PipelineChunk {
	chunks := a.Chunk(doc.Content)
	// Add document context to chunks
	for i := range chunks {
		chunks[i].ID = fmt.Sprintf("%s_chunk_%d", doc.ID, i)
		chunks[i].DocID = doc.ID
		if doc.Metadata != nil {
			if chunks[i].Metadata == nil {
				chunks[i].Metadata = make(map[string]interface{})
			}
			for k, v := range doc.Metadata {
				chunks[i].Metadata[k] = v
			}
		}
	}
	return chunks
}

// PipelineAdapter adapts the module's Pipeline builder to HelixAgent's Pipeline.
type PipelineAdapter struct {
	pipeline *modpipeline.Pipeline
}

// NewPipelineAdapter creates a new pipeline adapter with a retriever.
func NewPipelineAdapter(retriever RetrieverFunc) (*PipelineAdapter, error) {
	// Create a wrapped retriever that implements the module interface
	modRetriever := &wrappedRetriever{fn: retriever}

	p, err := modpipeline.NewPipeline().
		Retrieve(modRetriever).
		Build()
	if err != nil {
		return nil, err
	}
	return &PipelineAdapter{pipeline: p}, nil
}

// RetrieverFunc is a function that retrieves documents.
type RetrieverFunc func(ctx context.Context, query string, topK int) ([]helixrag.PipelineSearchResult, error)

// EmbedderFunc is a function that generates embeddings.
type EmbedderFunc func(ctx context.Context, texts []string) ([][]float32, error)

type wrappedRetriever struct {
	fn RetrieverFunc
}

func (r *wrappedRetriever) Retrieve(ctx context.Context, query string, opts modretriever.Options) ([]modretriever.Document, error) {
	topK := 10
	if opts.TopK > 0 {
		topK = opts.TopK
	}
	results, err := r.fn(ctx, query, topK)
	if err != nil {
		return nil, err
	}
	docs := make([]modretriever.Document, len(results))
	for i, res := range results {
		docs[i] = modretriever.Document{
			ID:      res.Chunk.ID,
			Content: res.Chunk.Content,
			Score:   float64(res.Score),
		}
	}
	return docs, nil
}

// Search performs a search using the pipeline.
func (a *PipelineAdapter) Search(ctx context.Context, query string, topK int) ([]helixrag.PipelineSearchResult, error) {
	result, err := a.pipeline.Execute(ctx, query)
	if err != nil {
		return nil, err
	}
	docs := result.Documents
	if topK > 0 && len(docs) > topK {
		docs = docs[:topK]
	}
	results := make([]helixrag.PipelineSearchResult, len(docs))
	for i, d := range docs {
		results[i] = helixrag.PipelineSearchResult{
			Chunk: helixrag.PipelineChunk{
				ID:      d.ID,
				Content: d.Content,
			},
			Score: float32(d.Score),
		}
	}
	return results, nil
}

// RerankerAdapter adapts the module's Reranker to HelixAgent usage.
type RerankerAdapter struct {
	reranker modreranker.Reranker
}

// NewScoreReranker creates a score-based reranker.
func NewScoreReranker(topK int) *RerankerAdapter {
	config := modreranker.Config{
		TopK: topK,
	}
	return &RerankerAdapter{
		reranker: modreranker.NewScoreReranker(config),
	}
}

// NewMMRReranker creates an MMR (Maximal Marginal Relevance) reranker.
func NewMMRReranker(lambda float64, topK int) *RerankerAdapter {
	config := modreranker.Config{
		Lambda: lambda,
		TopK:   topK,
	}
	return &RerankerAdapter{
		reranker: modreranker.NewMMRReranker(config),
	}
}

// Rerank reorders search results.
func (a *RerankerAdapter) Rerank(ctx context.Context, query string, results []helixrag.PipelineSearchResult, topK int) ([]helixrag.PipelineSearchResult, error) {
	docs := make([]modretriever.Document, len(results))
	for i, r := range results {
		docs[i] = modretriever.Document{
			ID:      r.Chunk.ID,
			Content: r.Chunk.Content,
			Score:   float64(r.Score),
		}
	}

	reranked, err := a.reranker.Rerank(ctx, query, docs)
	if err != nil {
		return nil, err
	}

	if topK > 0 && len(reranked) > topK {
		reranked = reranked[:topK]
	}

	output := make([]helixrag.PipelineSearchResult, len(reranked))
	for i, d := range reranked {
		output[i] = helixrag.PipelineSearchResult{
			Chunk: helixrag.PipelineChunk{
				ID:      d.ID,
				Content: d.Content,
			},
			Score: float32(d.Score),
		}
	}
	return output, nil
}

// HybridRetrieverAdapter adapts the module's HybridRetriever.
type HybridRetrieverAdapter struct {
	hybrid *modhybrid.HybridRetriever
}

// NewHybridRetrieverAdapter creates a hybrid retriever.
func NewHybridRetrieverAdapter(
	semanticRetriever modretriever.Retriever,
	fusionMethod string,
) *HybridRetrieverAdapter {
	var fusion modhybrid.FusionStrategy
	switch fusionMethod {
	case "rrf":
		fusion = modhybrid.NewRRFStrategy(60)
	case "linear":
		fusion = modhybrid.NewLinearStrategy(0.6, 0.4)
	default:
		fusion = modhybrid.NewRRFStrategy(60)
	}

	// Create a keyword retriever for BM25 search
	keyword := modhybrid.NewKeywordRetriever()

	// Wrap the semantic retriever
	semantic := modhybrid.NewSemanticRetriever(semanticRetriever)

	hybrid := modhybrid.NewHybridRetriever(
		semantic,
		keyword,
		fusion,
		modhybrid.DefaultHybridConfig(),
	)
	return &HybridRetrieverAdapter{hybrid: hybrid}
}

// Search performs hybrid retrieval.
func (a *HybridRetrieverAdapter) Search(ctx context.Context, query string, topK int) ([]helixrag.PipelineSearchResult, error) {
	opts := modretriever.Options{TopK: topK}
	docs, err := a.hybrid.Retrieve(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	results := make([]helixrag.PipelineSearchResult, len(docs))
	for i, d := range docs {
		results[i] = helixrag.PipelineSearchResult{
			Chunk: helixrag.PipelineChunk{
				ID:      d.ID,
				Content: d.Content,
			},
			Score: float32(d.Score),
		}
	}
	return results, nil
}

// ToModuleDocument converts a HelixAgent PipelineDocument to module Document.
func ToModuleDocument(h *helixrag.PipelineDocument) modretriever.Document {
	if h == nil {
		return modretriever.Document{}
	}
	return modretriever.Document{
		ID:      h.ID,
		Content: h.Content,
	}
}

// ToHelixSearchResult converts a module Document to HelixAgent PipelineSearchResult.
func ToHelixSearchResult(d modretriever.Document) helixrag.PipelineSearchResult {
	return helixrag.PipelineSearchResult{
		Chunk: helixrag.PipelineChunk{
			ID:      d.ID,
			Content: d.Content,
		},
		Score: float32(d.Score),
	}
}

// FusionMethod constants
const (
	FusionRRF    = "rrf"
	FusionLinear = "linear"
)
