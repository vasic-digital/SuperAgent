// Package chunker provides code chunking capabilities
package chunker

import "dev.helix.agent/internal/search/types"

// Re-export types from types package

// Chunk is an alias for types.Chunk
type Chunk = types.Chunk

// ChunkType is an alias for types.ChunkType
type ChunkType = types.ChunkType

// Chunker is an alias for types.Chunker
type Chunker = types.Chunker

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
