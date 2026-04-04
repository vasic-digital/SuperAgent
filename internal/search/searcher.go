// Package search provides semantic search capabilities
package search

import (
	"context"
	"fmt"
	"strings"

	"dev.helix.agent/internal/search/types"
)

// CodeSearcher performs semantic code searches
type CodeSearcher struct {
	embedder    types.Embedder
	vectorStore types.VectorStore
	collection  string
}

// NewCodeSearcher creates a new code searcher
func NewCodeSearcher(embedder types.Embedder, store types.VectorStore, collection string) *CodeSearcher {
	return &CodeSearcher{
		embedder:    embedder,
		vectorStore: store,
		collection:  collection,
	}
}

// Search performs semantic search with a text query
func (s *CodeSearcher) Search(ctx context.Context, query string, opts types.SearchOptions) ([]types.SearchResult, error) {
	// Generate embedding for query
	embedding, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search vector store
	results, err := s.vectorStore.Search(ctx, s.collection, embedding, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return results, nil
}

// SearchSimilar finds similar code to a given file location
func (s *CodeSearcher) SearchSimilar(ctx context.Context, filePath string, line int, opts types.SearchOptions) ([]types.SearchResult, error) {
	// For now, search using the file path as query
	// Future: Extract code at line and use that as query
	query := fmt.Sprintf("similar to %s line %d", filePath, line)
	return s.Search(ctx, query, opts)
}

// SearchWithContext performs search with additional context
func (s *CodeSearcher) SearchWithContext(ctx context.Context, query string, contextFiles []string, opts types.SearchOptions) ([]types.SearchResult, error) {
	// Enhance query with context
	enhancedQuery := s.enhanceQuery(query, contextFiles)
	return s.Search(ctx, enhancedQuery, opts)
}

// enhanceQuery adds context to the search query
func (s *CodeSearcher) enhanceQuery(query string, contextFiles []string) string {
	if len(contextFiles) == 0 {
		return query
	}

	// Add file context to query
	var fileContext strings.Builder
	fileContext.WriteString(query)
	fileContext.WriteString(" (context: ")
	for i, file := range contextFiles {
		if i > 0 {
			fileContext.WriteString(", ")
		}
		fileContext.WriteString(file)
	}
	fileContext.WriteString(")")

	return fileContext.String()
}

// RerankResults reranks search results using additional signals
// Currently returns results as-is. Future enhancements may include:
// - Recency of file modification
// - File importance (e.g., main files vs test files)
// - Query term matching
// - Code complexity
func (s *CodeSearcher) RerankResults(results []types.SearchResult, query string) []types.SearchResult {
	return results
}

// Ensure interface is implemented
var _ Searcher = (*CodeSearcher)(nil)
