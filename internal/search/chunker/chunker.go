// Package chunker provides code chunking capabilities
package chunker

import (
	"fmt"
	"strings"
)

// SimpleChunker implements basic line-based chunking
type SimpleChunker struct {
	chunkSize    int
	chunkOverlap int
}

// NewSimpleChunker creates a new simple chunker
func NewSimpleChunker(chunkSize, chunkOverlap int) *SimpleChunker {
	return &SimpleChunker{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

// Chunk splits content into chunks
func (c *SimpleChunker) Chunk(content string, language string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []Chunk

	for i := 0; i < len(lines); i += c.chunkSize - c.chunkOverlap {
		end := i + c.chunkSize
		if end > len(lines) {
			end = len(lines)
		}

		chunkContent := strings.Join(lines[i:end], "\n")
		chunk := Chunk{
			ID:        fmt.Sprintf("chunk_%d_%d", i, end),
			Content:   chunkContent,
			StartLine: i + 1,
			EndLine:   end,
			Language:  language,
			Type:      ChunkTypeGeneral,
		}
		chunks = append(chunks, chunk)

		if end == len(lines) {
			break
		}
	}

	return chunks, nil
}

// LanguageBasedChunker chunks code based on language constructs
type LanguageBasedChunker struct {
	maxChunkSize int
}

// NewLanguageBasedChunker creates a language-aware chunker
func NewLanguageBasedChunker(maxChunkSize int) *LanguageBasedChunker {
	return &LanguageBasedChunker{maxChunkSize: maxChunkSize}
}

// Chunk splits content into semantic chunks based on language
// Uses language-specific chunking when available, falls back to simple chunking.
func (c *LanguageBasedChunker) Chunk(content string, language string) ([]Chunk, error) {
	switch language {
	case "go", "python", "javascript", "typescript", "rust":
		return c.chunkByFunctions(content, language)
	default:
		return NewSimpleChunker(50, 10).Chunk(content, language)
	}
}

// chunkByFunctions attempts to chunk by function boundaries
// Currently falls back to simple chunking. Future enhancement: tree-sitter based detection.
func (c *LanguageBasedChunker) chunkByFunctions(content string, language string) ([]Chunk, error) {
	return NewSimpleChunker(50, 10).Chunk(content, language)
}

// ChunkFile chunks a file into multiple chunks
func ChunkFile(filePath string, content string, language string, chunker Chunker) ([]Chunk, error) {
	chunks, err := chunker.Chunk(content, language)
	if err != nil {
		return nil, err
	}

	// Update chunk IDs to include file path
	for i := range chunks {
		chunks[i].ID = fmt.Sprintf("%s:%d-%d", filePath, chunks[i].StartLine, chunks[i].EndLine)
		chunks[i].FilePath = filePath
	}

	return chunks, nil
}
