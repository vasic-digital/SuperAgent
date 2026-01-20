// Package rag provides Retrieval-Augmented Generation capabilities with hybrid search,
// reranking, and knowledge graph integration.
package rag

import (
	"context"
	"time"
)

// Document represents a document in the RAG system
type Document struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Embedding []float32              `json:"embedding,omitempty"`
	Score     float64                `json:"score,omitempty"`
	Source    string                 `json:"source,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Chunk represents a text chunk for embedding
type Chunk struct {
	ID         string                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	Content    string                 `json:"content"`
	Index      int                    `json:"index"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Embedding  []float32              `json:"embedding,omitempty"`
	StartChar  int                    `json:"start_char"`
	EndChar    int                    `json:"end_char"`
}

// SearchResult represents a search result with relevance scoring
type SearchResult struct {
	Document      *Document `json:"document"`
	Score         float64   `json:"score"`
	RerankedScore float64   `json:"reranked_score,omitempty"`
	Highlights    []string  `json:"highlights,omitempty"`
	MatchType     MatchType `json:"match_type"`
}

// MatchType indicates how a document was matched
type MatchType string

const (
	MatchTypeDense  MatchType = "dense"  // Semantic/embedding match
	MatchTypeSparse MatchType = "sparse" // Keyword/BM25 match
	MatchTypeHybrid MatchType = "hybrid" // Combined match
)

// SearchOptions configures search behavior
type SearchOptions struct {
	TopK            int                    `json:"top_k"`
	MinScore        float64                `json:"min_score"`
	Filter          map[string]interface{} `json:"filter,omitempty"`
	EnableReranking bool                   `json:"enable_reranking"`
	HybridAlpha     float64                `json:"hybrid_alpha"` // 0=sparse only, 1=dense only, 0.5=balanced
	IncludeMetadata bool                   `json:"include_metadata"`
	Namespace       string                 `json:"namespace,omitempty"`
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		TopK:            10,
		MinScore:        0.0,
		EnableReranking: true,
		HybridAlpha:     0.5,
		IncludeMetadata: true,
	}
}

// Retriever defines the interface for document retrieval
type Retriever interface {
	// Retrieve searches for relevant documents
	Retrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error)
	// Index adds documents to the retriever
	Index(ctx context.Context, docs []*Document) error
	// Delete removes documents by ID
	Delete(ctx context.Context, ids []string) error
}

// DenseRetriever uses embeddings for semantic search
type DenseRetriever interface {
	Retriever
	// Embed generates embeddings for text
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// SparseRetriever uses keyword matching (BM25, TF-IDF)
type SparseRetriever interface {
	Retriever
	// GetTermFrequencies returns term frequencies for a document
	GetTermFrequencies(ctx context.Context, docID string) (map[string]float64, error)
}

// Reranker reorders search results for better relevance
type Reranker interface {
	// Rerank reorders results based on query relevance
	Rerank(ctx context.Context, query string, results []*SearchResult, topK int) ([]*SearchResult, error)
}

// Embedder generates vector embeddings
type Embedder interface {
	// Embed generates embeddings for texts
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	// EmbedQuery generates an embedding optimized for queries
	EmbedQuery(ctx context.Context, query string) ([]float32, error)
	// GetDimension returns the embedding dimension
	GetDimension() int
	// GetModelName returns the model name
	GetModelName() string
}

// Chunker splits documents into smaller pieces
type Chunker interface {
	// Chunk splits a document into chunks
	Chunk(ctx context.Context, doc *Document) ([]*Chunk, error)
}

// ChunkerConfig configures chunking behavior
type ChunkerConfig struct {
	ChunkSize      int    `json:"chunk_size"`
	ChunkOverlap   int    `json:"chunk_overlap"`
	Separator      string `json:"separator"`
	LengthFunction string `json:"length_function"` // "chars" or "tokens"
}

// DefaultChunkerConfig returns default chunker configuration
func DefaultChunkerConfig() *ChunkerConfig {
	return &ChunkerConfig{
		ChunkSize:      1000,
		ChunkOverlap:   200,
		Separator:      "\n\n",
		LengthFunction: "chars",
	}
}

// Entity represents an extracted entity for knowledge graphs
type Entity struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Mentions   []EntityMention        `json:"mentions,omitempty"`
}

// EntityMention represents where an entity appears in text
type EntityMention struct {
	DocumentID string `json:"document_id"`
	ChunkID    string `json:"chunk_id"`
	StartChar  int    `json:"start_char"`
	EndChar    int    `json:"end_char"`
	Context    string `json:"context"`
}

// Relation represents a relationship between entities
type Relation struct {
	ID         string                 `json:"id"`
	SourceID   string                 `json:"source_id"`
	TargetID   string                 `json:"target_id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Confidence float64                `json:"confidence"`
}

// KnowledgeGraph provides graph-based retrieval
type KnowledgeGraph interface {
	// AddEntity adds an entity to the graph
	AddEntity(ctx context.Context, entity *Entity) error
	// AddRelation adds a relation between entities
	AddRelation(ctx context.Context, relation *Relation) error
	// GetRelatedEntities finds entities related to the query
	GetRelatedEntities(ctx context.Context, entityID string, depth int) ([]*Entity, error)
	// TraverseForAnswer traverses the graph to find answers
	TraverseForAnswer(ctx context.Context, query string, startEntities []*Entity) ([]*SearchResult, error)
}
