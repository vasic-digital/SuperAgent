package rag

import (
	"context"
	"fmt"

	"dev.helix.agent/internal/vectordb/qdrant"
	"github.com/sirupsen/logrus"
)

// QdrantDenseRetriever implements DenseRetriever using the existing Qdrant client
type QdrantDenseRetriever struct {
	client     *qdrant.Client
	collection string
	embedder   Embedder
	logger     *logrus.Logger
}

// NewQdrantDenseRetriever creates a new Qdrant-based dense retriever
func NewQdrantDenseRetriever(client *qdrant.Client, collection string, embedder Embedder, logger *logrus.Logger) *QdrantDenseRetriever {
	if logger == nil {
		logger = logrus.New()
	}
	return &QdrantDenseRetriever{
		client:     client,
		collection: collection,
		embedder:   embedder,
		logger:     logger,
	}
}

// Retrieve performs dense retrieval using Qdrant
func (r *QdrantDenseRetriever) Retrieve(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if r.client == nil {
		return nil, fmt.Errorf("Qdrant client not initialized")
	}

	if opts == nil {
		opts = &SearchOptions{TopK: 10}
	}

	// Generate query embedding using EmbedQuery for single query
	vector, err := r.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search Qdrant
	qdrantOpts := &qdrant.SearchOptions{
		Limit:          opts.TopK,
		ScoreThreshold: float32(opts.MinScore),
		WithPayload:    true,
		WithVectors:    false,
	}

	// Add filters if specified
	if opts.Filter != nil && len(opts.Filter) > 0 {
		qdrantOpts.Filter = opts.Filter
	}

	results, err := r.client.Search(ctx, r.collection, vector, qdrantOpts)
	if err != nil {
		return nil, fmt.Errorf("Qdrant search failed: %w", err)
	}

	// Convert to SearchResult
	searchResults := make([]*SearchResult, 0, len(results))
	for _, point := range results {
		result := &SearchResult{
			Document:  pointToDocument(point),
			Score:     float64(point.Score),
			MatchType: MatchTypeDense,
		}
		searchResults = append(searchResults, result)
	}

	r.logger.WithFields(logrus.Fields{
		"query":   truncate(query, 50),
		"results": len(searchResults),
	}).Debug("Qdrant dense retrieval completed")

	return searchResults, nil
}

// GetName returns the retriever name
func (r *QdrantDenseRetriever) GetName() string {
	return "qdrant_dense"
}

// pointToDocument converts a Qdrant ScoredPoint to a Document
func pointToDocument(point qdrant.ScoredPoint) *Document {
	doc := &Document{
		ID:       fmt.Sprintf("%v", point.ID),
		Metadata: make(map[string]interface{}),
	}

	// Extract content from payload
	if content, ok := point.Payload["content"].(string); ok {
		doc.Content = content
	} else if text, ok := point.Payload["text"].(string); ok {
		doc.Content = text
	}

	// Extract title into metadata
	if title, ok := point.Payload["title"].(string); ok {
		doc.Metadata["title"] = title
	}

	// Extract source
	if source, ok := point.Payload["source"].(string); ok {
		doc.Source = source
	}

	// Copy remaining metadata
	for k, v := range point.Payload {
		if k != "content" && k != "text" && k != "title" && k != "source" {
			doc.Metadata[k] = v
		}
	}

	return doc
}

func extractMetadata(payload map[string]interface{}) map[string]interface{} {
	if payload == nil {
		return make(map[string]interface{})
	}
	return payload
}

// toFloat32Slice converts float64 slice to float32 slice for embeddings
func toFloat32Slice(embedding []float64) []float32 {
	result := make([]float32, len(embedding))
	for i, v := range embedding {
		result[i] = float32(v)
	}
	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// QdrantDocumentStore provides document storage using Qdrant
type QdrantDocumentStore struct {
	client     *qdrant.Client
	collection string
	embedder   Embedder
	logger     *logrus.Logger
}

// NewQdrantDocumentStore creates a new Qdrant document store
func NewQdrantDocumentStore(client *qdrant.Client, collection string, embedder Embedder, logger *logrus.Logger) *QdrantDocumentStore {
	if logger == nil {
		logger = logrus.New()
	}
	return &QdrantDocumentStore{
		client:     client,
		collection: collection,
		embedder:   embedder,
		logger:     logger,
	}
}

// AddDocument adds a document to the store
func (s *QdrantDocumentStore) AddDocument(ctx context.Context, doc *Document) error {
	if s.client == nil {
		return fmt.Errorf("Qdrant client not initialized")
	}

	// Generate embedding using Embed with single-item slice
	embeddings, err := s.embedder.Embed(ctx, []string{doc.Content})
	if err != nil {
		return fmt.Errorf("failed to embed document: %w", err)
	}
	if len(embeddings) == 0 {
		return fmt.Errorf("no embedding returned")
	}

	// Create point - get title from metadata if present
	title := ""
	if t, ok := doc.Metadata["title"].(string); ok {
		title = t
	}

	point := qdrant.Point{
		ID:     doc.ID,
		Vector: embeddings[0],
		Payload: map[string]interface{}{
			"content": doc.Content,
			"title":   title,
			"source":  doc.Source,
		},
	}

	// Add metadata to payload
	for k, v := range doc.Metadata {
		point.Payload[k] = v
	}

	// Upsert to Qdrant
	if err := s.client.UpsertPoints(ctx, s.collection, []qdrant.Point{point}); err != nil {
		return fmt.Errorf("failed to upsert document: %w", err)
	}

	s.logger.WithField("doc_id", doc.ID).Debug("Document added to Qdrant")
	return nil
}

// AddDocuments adds multiple documents to the store
func (s *QdrantDocumentStore) AddDocuments(ctx context.Context, docs []*Document) error {
	if len(docs) == 0 {
		return nil
	}

	// Generate embeddings for all documents using Embed (which is batch)
	contents := make([]string, len(docs))
	for i, doc := range docs {
		contents[i] = doc.Content
	}

	embeddings, err := s.embedder.Embed(ctx, contents)
	if err != nil {
		return fmt.Errorf("failed to embed documents: %w", err)
	}

	// Create points
	points := make([]qdrant.Point, len(docs))
	for i, doc := range docs {
		// Get title from metadata if present
		title := ""
		if t, ok := doc.Metadata["title"].(string); ok {
			title = t
		}

		points[i] = qdrant.Point{
			ID:     doc.ID,
			Vector: embeddings[i],
			Payload: map[string]interface{}{
				"content": doc.Content,
				"title":   title,
				"source":  doc.Source,
			},
		}
		for k, v := range doc.Metadata {
			points[i].Payload[k] = v
		}
	}

	// Upsert to Qdrant
	if err := s.client.UpsertPoints(ctx, s.collection, points); err != nil {
		return fmt.Errorf("failed to upsert documents: %w", err)
	}

	s.logger.WithField("count", len(docs)).Debug("Documents added to Qdrant")
	return nil
}

// DeleteDocument deletes a document from the store
func (s *QdrantDocumentStore) DeleteDocument(ctx context.Context, id string) error {
	if err := s.client.DeletePoints(ctx, s.collection, []string{id}); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// GetDocument retrieves a document by ID
func (s *QdrantDocumentStore) GetDocument(ctx context.Context, id string) (*Document, error) {
	points, err := s.client.GetPoints(ctx, s.collection, []string{id})
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("document not found: %s", id)
	}

	return payloadToDocument(id, points[0].Payload), nil
}

func payloadToDocument(id string, payload map[string]interface{}) *Document {
	doc := &Document{
		ID:       id,
		Metadata: make(map[string]interface{}),
	}

	if content, ok := payload["content"].(string); ok {
		doc.Content = content
	}
	// Store title in metadata since Document doesn't have a Title field
	if title, ok := payload["title"].(string); ok {
		doc.Metadata["title"] = title
	}
	if source, ok := payload["source"].(string); ok {
		doc.Source = source
	}

	for k, v := range payload {
		if k != "content" && k != "title" && k != "source" {
			doc.Metadata[k] = v
		}
	}

	return doc
}

// EnsureCollection creates the collection if it doesn't exist
func (s *QdrantDocumentStore) EnsureCollection(ctx context.Context, vectorSize int) error {
	exists, err := s.client.CollectionExists(ctx, s.collection)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if !exists {
		config := &qdrant.CollectionConfig{
			Name:       s.collection,
			VectorSize: vectorSize,
			Distance:   qdrant.DistanceCosine,
		}
		if err := s.client.CreateCollection(ctx, config); err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		s.logger.WithField("collection", s.collection).Info("Created Qdrant collection")
	}

	return nil
}
