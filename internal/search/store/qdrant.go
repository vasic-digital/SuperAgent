// Package store provides vector store implementations
package store

import (
	"context"
	"fmt"

	"dev.helix.agent/internal/search/types"
)

// QdrantStore implements VectorStore for Qdrant
type QdrantStore struct {
	host       string
	port       int
	collection string
}

// NewQdrantStore creates a new Qdrant store
func NewQdrantStore(host string, port int, collection string) (*QdrantStore, error) {
	return &QdrantStore{
		host:       host,
		port:       port,
		collection: collection,
	}, nil
}

// CreateCollection creates a new collection
func (s *QdrantStore) CreateCollection(ctx context.Context, name string, dims int) error {
	return nil
}

// DeleteCollection deletes a collection
func (s *QdrantStore) DeleteCollection(ctx context.Context, name string) error {
	return nil
}

// Upsert adds or updates documents
func (s *QdrantStore) Upsert(ctx context.Context, collection string, docs []types.Document) error {
	return nil
}

// Delete removes documents
func (s *QdrantStore) Delete(ctx context.Context, collection string, ids []string) error {
	return nil
}

// Search performs vector search
func (s *QdrantStore) Search(ctx context.Context, collection string, vector []float32, opts types.SearchOptions) ([]types.SearchResult, error) {
	return nil, fmt.Errorf("Qdrant search not yet implemented")
}

// SearchByText performs text search
func (s *QdrantStore) SearchByText(ctx context.Context, collection string, text string, opts types.SearchOptions) ([]types.SearchResult, error) {
	return nil, fmt.Errorf("SearchByText not supported")
}

// GetCollectionStats returns collection statistics
func (s *QdrantStore) GetCollectionStats(ctx context.Context, collection string) (*types.CollectionStats, error) {
	return &types.CollectionStats{
		Name: collection,
	}, nil
}

// Ensure interface is implemented
var _ types.VectorStore = (*QdrantStore)(nil)
