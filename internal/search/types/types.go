// Package types provides core search types and interfaces
package types

import (
	"context"
	"time"
)

// Chunk represents a code segment with metadata
type Chunk struct {
	ID         string
	Content    string
	FilePath   string
	StartLine  int
	EndLine    int
	Language   string
	Type       ChunkType
	Parent     string
	Imports    []string
	Embeddings []float32
}

// ChunkType represents the type of code chunk
type ChunkType string

const (
	ChunkTypeFunction  ChunkType = "function"
	ChunkTypeClass     ChunkType = "class"
	ChunkTypeInterface ChunkType = "interface"
	ChunkTypeMethod    ChunkType = "method"
	ChunkTypeComment   ChunkType = "comment"
	ChunkTypeImport    ChunkType = "import"
	ChunkTypeGeneral   ChunkType = "general"
)

// Chunker splits code into semantic chunks
type Chunker interface {
	Chunk(content string, language string) ([]Chunk, error)
}

// Embedder generates embeddings for code
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, query string) ([]float32, error)
	Dimensions() int
}

// Document represents a document in the vector store
type Document struct {
	ID       string
	Vector   []float32
	Metadata map[string]interface{}
	Content  string
}

// SearchOptions configures search behavior
type SearchOptions struct {
	TopK           int
	MinScore       float32
	Filters        map[string]interface{}
	IncludeContent bool
}

// SearchResult represents a search result
type SearchResult struct {
	Document
	Score    float32
	Distance float32
}

// VectorStore abstracts the vector database
type VectorStore interface {
	CreateCollection(ctx context.Context, name string, dims int) error
	DeleteCollection(ctx context.Context, name string) error
	Upsert(ctx context.Context, collection string, docs []Document) error
	Delete(ctx context.Context, collection string, ids []string) error
	Search(ctx context.Context, collection string, vector []float32, opts SearchOptions) ([]SearchResult, error)
	SearchByText(ctx context.Context, collection string, text string, opts SearchOptions) ([]SearchResult, error)
	GetCollectionStats(ctx context.Context, collection string) (*CollectionStats, error)
}

// CollectionStats contains statistics about a collection
type CollectionStats struct {
	Name        string
	Count       int64
	Dimensions  int
	CreatedAt   time.Time
	LastUpdated time.Time
}

// IndexResult represents the outcome of indexing
type IndexResult struct {
	FilesIndexed  int
	ChunksCreated int
	Errors        []error
	Duration      time.Duration
}

// Indexer manages the indexing pipeline
type Indexer interface {
	Index(ctx context.Context) (*IndexResult, error)
	IndexFile(ctx context.Context, path string) error
	DeleteFile(ctx context.Context, path string) error
	Watch(ctx context.Context) error
}

// Searcher performs semantic searches
type Searcher interface {
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
	SearchSimilar(ctx context.Context, filePath string, line int, opts SearchOptions) ([]SearchResult, error)
}
