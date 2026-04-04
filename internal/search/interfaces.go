// Package search provides semantic code search capabilities
package search

import "dev.helix.agent/internal/search/types"

// Re-export all types from types package for backward compatibility

// Chunk is an alias for types.Chunk
type Chunk = types.Chunk

// ChunkType is an alias for types.ChunkType
type ChunkType = types.ChunkType

// Chunker is an alias for types.Chunker
type Chunker = types.Chunker

// Embedder is an alias for types.Embedder
type Embedder = types.Embedder

// Document is an alias for types.Document
type Document = types.Document

// SearchOptions is an alias for types.SearchOptions
type SearchOptions = types.SearchOptions

// SearchResult is an alias for types.SearchResult
type SearchResult = types.SearchResult

// VectorStore is an alias for types.VectorStore
type VectorStore = types.VectorStore

// CollectionStats is an alias for types.CollectionStats
type CollectionStats = types.CollectionStats

// IndexResult is an alias for types.IndexResult
type IndexResult = types.IndexResult

// Indexer is an alias for types.Indexer
type Indexer = types.Indexer

// Searcher is an alias for types.Searcher
type Searcher = types.Searcher

// Chunk type constants
const (
	ChunkTypeFunction  = types.ChunkTypeFunction
	ChunkTypeClass     = types.ChunkTypeClass
	ChunkTypeInterface = types.ChunkTypeInterface
	ChunkTypeMethod    = types.ChunkTypeMethod
	ChunkTypeComment   = types.ChunkTypeComment
	ChunkTypeImport    = types.ChunkTypeImport
	ChunkTypeGeneral   = types.ChunkTypeGeneral
)
