// Package rag provides RAG (Retrieval-Augmented Generation) capabilities.
package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"dev.helix.agent/internal/embeddings/models"
	"dev.helix.agent/internal/mcp/servers"
)

// VectorDBType represents the type of vector database.
type VectorDBType string

const (
	VectorDBChroma   VectorDBType = "chroma"
	VectorDBQdrant   VectorDBType = "qdrant"
	VectorDBWeaviate VectorDBType = "weaviate"
)

// PipelineDocument represents a document for RAG pipeline processing.
type PipelineDocument struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Chunks    []PipelineChunk        `json:"chunks,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// PipelineChunk represents a chunk of a document in the pipeline.
type PipelineChunk struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Embedding []float32              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	StartIdx  int                    `json:"start_idx"`
	EndIdx    int                    `json:"end_idx"`
	DocID     string                 `json:"doc_id"`
}

// PipelineSearchResult represents a search result from the pipeline.
type PipelineSearchResult struct {
	Chunk    PipelineChunk          `json:"chunk"`
	Score    float32                `json:"score"`
	Distance float32                `json:"distance"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChunkingConfig contains configuration for document chunking.
type ChunkingConfig struct {
	ChunkSize    int    `json:"chunk_size"`
	ChunkOverlap int    `json:"chunk_overlap"`
	Separator    string `json:"separator"`
}

// PipelineConfig contains configuration for the RAG pipeline.
type PipelineConfig struct {
	VectorDBType     VectorDBType                   `json:"vector_db_type"`
	CollectionName   string                         `json:"collection_name"`
	EmbeddingModel   string                         `json:"embedding_model"`
	ChunkingConfig   ChunkingConfig                 `json:"chunking_config"`
	ChromaConfig     *servers.ChromaAdapterConfig   `json:"chroma_config,omitempty"`
	QdrantConfig     *servers.QdrantAdapterConfig   `json:"qdrant_config,omitempty"`
	WeaviateConfig   *servers.WeaviateAdapterConfig `json:"weaviate_config,omitempty"`
	EnableCache      bool                           `json:"enable_cache"`
	CacheTTL         time.Duration                  `json:"cache_ttl"`
}

// DefaultChunkingConfig returns default chunking configuration.
func DefaultChunkingConfig() ChunkingConfig {
	return ChunkingConfig{
		ChunkSize:    512,
		ChunkOverlap: 50,
		Separator:    "\n\n",
	}
}

// Pipeline represents a RAG pipeline.
type Pipeline struct {
	mu                sync.RWMutex
	config            PipelineConfig
	embeddingRegistry *models.EmbeddingModelRegistry
	chromaAdapter     *servers.ChromaAdapter
	qdrantAdapter     *servers.QdrantAdapter
	weaviateAdapter   *servers.WeaviateAdapter
	connected         bool
	initialized       bool
}

// NewPipeline creates a new RAG pipeline.
func NewPipeline(config PipelineConfig, embeddingRegistry *models.EmbeddingModelRegistry) *Pipeline {
	// Set defaults
	if config.ChunkingConfig.ChunkSize == 0 {
		config.ChunkingConfig = DefaultChunkingConfig()
	}
	if config.CollectionName == "" {
		config.CollectionName = "default"
	}

	return &Pipeline{
		config:            config,
		embeddingRegistry: embeddingRegistry,
	}
}

// Initialize initializes the pipeline by connecting to the vector database.
func (p *Pipeline) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	var err error
	switch p.config.VectorDBType {
	case VectorDBChroma:
		if p.config.ChromaConfig == nil {
			return fmt.Errorf("chroma config is required")
		}
		p.chromaAdapter = servers.NewChromaAdapter(*p.config.ChromaConfig)
		err = p.chromaAdapter.Connect(ctx)
	case VectorDBQdrant:
		if p.config.QdrantConfig == nil {
			return fmt.Errorf("qdrant config is required")
		}
		p.qdrantAdapter = servers.NewQdrantAdapter(*p.config.QdrantConfig)
		err = p.qdrantAdapter.Connect(ctx)
	case VectorDBWeaviate:
		if p.config.WeaviateConfig == nil {
			return fmt.Errorf("weaviate config is required")
		}
		p.weaviateAdapter = servers.NewWeaviateAdapter(*p.config.WeaviateConfig)
		err = p.weaviateAdapter.Connect(ctx)
	default:
		return fmt.Errorf("unsupported vector database type: %s", p.config.VectorDBType)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to vector database: %w", err)
	}

	// Ensure collection/index exists
	if err := p.ensureCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	p.connected = true
	p.initialized = true
	return nil
}

// ensureCollection ensures the collection/index exists.
func (p *Pipeline) ensureCollection(ctx context.Context) error {
	// Get embedding dimension
	dim, err := p.getEmbeddingDimension(ctx)
	if err != nil {
		return err
	}

	switch p.config.VectorDBType {
	case VectorDBChroma:
		_, err := p.chromaAdapter.GetCollection(ctx, p.config.CollectionName)
		if err != nil {
			// Create collection
			_, err = p.chromaAdapter.CreateCollection(ctx, p.config.CollectionName, nil)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	case VectorDBQdrant:
		exists, _ := p.qdrantAdapter.CollectionExists(ctx, p.config.CollectionName)
		if !exists {
			// Create collection
			err = p.qdrantAdapter.CreateCollection(ctx, p.config.CollectionName, uint64(dim), "Cosine")
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	case VectorDBWeaviate:
		_, err := p.weaviateAdapter.GetClass(ctx, p.config.CollectionName)
		if err != nil {
			// Create class
			class := &servers.WeaviateClass{
				Class:       p.config.CollectionName,
				Description: "RAG collection",
				Properties: []servers.WeaviateProperty{
					{Name: "content", DataType: []string{"text"}},
					{Name: "doc_id", DataType: []string{"text"}},
					{Name: "chunk_id", DataType: []string{"text"}},
					{Name: "start_idx", DataType: []string{"int"}},
					{Name: "end_idx", DataType: []string{"int"}},
				},
			}
			err = p.weaviateAdapter.CreateClass(ctx, class)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}

	return nil
}

// getEmbeddingDimension returns the embedding dimension for the configured model.
func (p *Pipeline) getEmbeddingDimension(ctx context.Context) (int, error) {
	// Encode a test string to get dimension
	embeddings, _, err := p.embeddingRegistry.EncodeWithFallback(ctx, []string{"test"})
	if err != nil {
		// Default dimension for common models
		return 1536, nil
	}
	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return 1536, nil
	}
	return len(embeddings[0]), nil
}

// Close closes the pipeline connections.
func (p *Pipeline) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var err error
	switch p.config.VectorDBType {
	case VectorDBChroma:
		if p.chromaAdapter != nil {
			err = p.chromaAdapter.Close()
		}
	case VectorDBQdrant:
		if p.qdrantAdapter != nil {
			err = p.qdrantAdapter.Close()
		}
	case VectorDBWeaviate:
		if p.weaviateAdapter != nil {
			err = p.weaviateAdapter.Close()
		}
	}

	p.connected = false
	return err
}

// Health checks the health of the pipeline.
func (p *Pipeline) Health(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return fmt.Errorf("pipeline not connected")
	}

	switch p.config.VectorDBType {
	case VectorDBChroma:
		return p.chromaAdapter.Health(ctx)
	case VectorDBQdrant:
		return p.qdrantAdapter.Health(ctx)
	case VectorDBWeaviate:
		return p.weaviateAdapter.Health(ctx)
	}

	return nil
}

// ChunkDocument chunks a document into smaller pieces.
func (p *Pipeline) ChunkDocument(doc *PipelineDocument) []PipelineChunk {
	cfg := p.config.ChunkingConfig
	content := doc.Content
	chunks := []PipelineChunk{}

	if len(content) <= cfg.ChunkSize {
		// Single chunk
		chunks = append(chunks, PipelineChunk{
			ID:       generateChunkID(doc.ID, 0),
			Content:  content,
			StartIdx: 0,
			EndIdx:   len(content),
			DocID:    doc.ID,
			Metadata: doc.Metadata,
		})
		return chunks
	}

	// Split by separator first
	parts := strings.Split(content, cfg.Separator)

	currentChunk := ""
	startIdx := 0
	chunkIdx := 0

	for _, part := range parts {
		if len(currentChunk)+len(part)+len(cfg.Separator) > cfg.ChunkSize {
			// Save current chunk
			if currentChunk != "" {
				chunks = append(chunks, PipelineChunk{
					ID:       generateChunkID(doc.ID, chunkIdx),
					Content:  strings.TrimSpace(currentChunk),
					StartIdx: startIdx,
					EndIdx:   startIdx + len(currentChunk),
					DocID:    doc.ID,
					Metadata: copyMetadata(doc.Metadata),
				})
				chunkIdx++

				// Overlap: keep some of the current chunk
				if cfg.ChunkOverlap > 0 && len(currentChunk) > cfg.ChunkOverlap {
					overlapStart := len(currentChunk) - cfg.ChunkOverlap
					startIdx = startIdx + overlapStart
					currentChunk = currentChunk[overlapStart:]
				} else {
					startIdx = startIdx + len(currentChunk)
					currentChunk = ""
				}
			}
		}

		if currentChunk == "" {
			currentChunk = part
		} else {
			currentChunk += cfg.Separator + part
		}
	}

	// Add remaining content
	if currentChunk != "" {
		chunks = append(chunks, PipelineChunk{
			ID:       generateChunkID(doc.ID, chunkIdx),
			Content:  strings.TrimSpace(currentChunk),
			StartIdx: startIdx,
			EndIdx:   startIdx + len(currentChunk),
			DocID:    doc.ID,
			Metadata: copyMetadata(doc.Metadata),
		})
	}

	return chunks
}

// IngestDocument ingests a document into the pipeline.
func (p *Pipeline) IngestDocument(ctx context.Context, doc *PipelineDocument) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return fmt.Errorf("pipeline not connected")
	}

	// Generate ID if not provided
	if doc.ID == "" {
		doc.ID = generateDocID(doc.Content)
	}

	// Chunk the document
	chunks := p.ChunkDocument(doc)
	doc.Chunks = chunks

	// Generate embeddings for chunks
	contents := make([]string, len(chunks))
	for i, chunk := range chunks {
		contents[i] = chunk.Content
	}

	embeddings, _, err := p.embeddingRegistry.EncodeWithFallback(ctx, contents)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Store in vector database
	for i := range chunks {
		if i < len(embeddings) {
			chunks[i].Embedding = embeddings[i]
		}
	}

	return p.storeChunks(ctx, chunks)
}

// storeChunks stores chunks in the vector database.
func (p *Pipeline) storeChunks(ctx context.Context, chunks []PipelineChunk) error {
	switch p.config.VectorDBType {
	case VectorDBChroma:
		return p.storeChunksChroma(ctx, chunks)
	case VectorDBQdrant:
		return p.storeChunksQdrant(ctx, chunks)
	case VectorDBWeaviate:
		return p.storeChunksWeaviate(ctx, chunks)
	}
	return nil
}

func (p *Pipeline) storeChunksChroma(ctx context.Context, chunks []PipelineChunk) error {
	docs := make([]servers.ChromaDocument, len(chunks))

	for i, chunk := range chunks {
		metadata := chunk.Metadata
		if metadata == nil {
			metadata = map[string]interface{}{}
		}
		metadata["doc_id"] = chunk.DocID
		metadata["start_idx"] = chunk.StartIdx
		metadata["end_idx"] = chunk.EndIdx

		docs[i] = servers.ChromaDocument{
			ID:        chunk.ID,
			Document:  chunk.Content,
			Embedding: chunk.Embedding,
			Metadata:  metadata,
		}
	}

	return p.chromaAdapter.AddDocuments(ctx, p.config.CollectionName, docs)
}

func (p *Pipeline) storeChunksQdrant(ctx context.Context, chunks []PipelineChunk) error {
	points := make([]servers.QdrantPoint, len(chunks))

	for i, chunk := range chunks {
		payload := map[string]interface{}{
			"content":   chunk.Content,
			"doc_id":    chunk.DocID,
			"start_idx": chunk.StartIdx,
			"end_idx":   chunk.EndIdx,
		}
		for k, v := range chunk.Metadata {
			payload[k] = v
		}

		points[i] = servers.QdrantPoint{
			ID:      chunk.ID,
			Vector:  chunk.Embedding,
			Payload: payload,
		}
	}

	return p.qdrantAdapter.UpsertPoints(ctx, p.config.CollectionName, points)
}

func (p *Pipeline) storeChunksWeaviate(ctx context.Context, chunks []PipelineChunk) error {
	objects := make([]servers.WeaviateObject, len(chunks))

	for i, chunk := range chunks {
		properties := map[string]interface{}{
			"content":   chunk.Content,
			"doc_id":    chunk.DocID,
			"start_idx": chunk.StartIdx,
			"end_idx":   chunk.EndIdx,
			"chunk_id":  chunk.ID,
		}
		for k, v := range chunk.Metadata {
			properties[k] = v
		}

		objects[i] = servers.WeaviateObject{
			Class:      p.config.CollectionName,
			Properties: properties,
			Vector:     chunk.Embedding,
		}
	}

	return p.weaviateAdapter.BatchCreateObjects(ctx, objects)
}

// IngestDocuments ingests multiple documents.
func (p *Pipeline) IngestDocuments(ctx context.Context, docs []*PipelineDocument) error {
	for _, doc := range docs {
		if err := p.IngestDocument(ctx, doc); err != nil {
			return fmt.Errorf("failed to ingest document %s: %w", doc.ID, err)
		}
	}
	return nil
}

// Search performs semantic search.
func (p *Pipeline) Search(ctx context.Context, query string, topK int) ([]PipelineSearchResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, fmt.Errorf("pipeline not connected")
	}

	if topK <= 0 {
		topK = 10
	}

	// Generate query embedding
	embeddings, _, err := p.embeddingRegistry.EncodeWithFallback(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, fmt.Errorf("failed to generate query embedding")
	}

	queryEmbedding := embeddings[0]

	switch p.config.VectorDBType {
	case VectorDBChroma:
		return p.searchChroma(ctx, queryEmbedding, topK)
	case VectorDBQdrant:
		return p.searchQdrant(ctx, queryEmbedding, topK)
	case VectorDBWeaviate:
		return p.searchWeaviate(ctx, queryEmbedding, topK)
	}

	return nil, fmt.Errorf("unsupported vector database type")
}

func (p *Pipeline) searchChroma(ctx context.Context, queryEmbedding []float32, topK int) ([]PipelineSearchResult, error) {
	result, err := p.chromaAdapter.Query(ctx, p.config.CollectionName, [][]float32{queryEmbedding}, topK, nil)
	if err != nil {
		return nil, err
	}

	searchResults := make([]PipelineSearchResult, 0)
	if result != nil && len(result.IDs) > 0 && len(result.IDs[0]) > 0 {
		for i, id := range result.IDs[0] {
			var content string
			var metadata map[string]interface{}
			var distance float32

			if len(result.Documents) > 0 && len(result.Documents[0]) > i {
				content = result.Documents[0][i]
			}
			if len(result.Metadatas) > 0 && len(result.Metadatas[0]) > i {
				metadata = result.Metadatas[0][i]
			}
			if len(result.Distances) > 0 && len(result.Distances[0]) > i {
				distance = result.Distances[0][i]
			}

			chunk := PipelineChunk{
				ID:       id,
				Content:  content,
				Metadata: metadata,
			}
			if docID, ok := metadata["doc_id"].(string); ok {
				chunk.DocID = docID
			}
			if startIdx, ok := metadata["start_idx"].(float64); ok {
				chunk.StartIdx = int(startIdx)
			}
			if endIdx, ok := metadata["end_idx"].(float64); ok {
				chunk.EndIdx = int(endIdx)
			}

			searchResults = append(searchResults, PipelineSearchResult{
				Chunk:    chunk,
				Score:    1.0 - distance,
				Distance: distance,
				Metadata: metadata,
			})
		}
	}

	return searchResults, nil
}

func (p *Pipeline) searchQdrant(ctx context.Context, queryEmbedding []float32, topK int) ([]PipelineSearchResult, error) {
	results, err := p.qdrantAdapter.Search(ctx, p.config.CollectionName, queryEmbedding, topK, nil, true, false)
	if err != nil {
		return nil, err
	}

	searchResults := make([]PipelineSearchResult, 0, len(results))
	for _, result := range results {
		content, _ := result.Payload["content"].(string)
		docID, _ := result.Payload["doc_id"].(string)
		startIdx, _ := result.Payload["start_idx"].(float64)
		endIdx, _ := result.Payload["end_idx"].(float64)

		var id string
		switch v := result.ID.(type) {
		case string:
			id = v
		case float64:
			id = fmt.Sprintf("%.0f", v)
		}

		chunk := PipelineChunk{
			ID:       id,
			Content:  content,
			DocID:    docID,
			StartIdx: int(startIdx),
			EndIdx:   int(endIdx),
			Metadata: result.Payload,
		}

		searchResults = append(searchResults, PipelineSearchResult{
			Chunk:    chunk,
			Score:    result.Score,
			Distance: 1.0 - result.Score,
			Metadata: result.Payload,
		})
	}

	return searchResults, nil
}

func (p *Pipeline) searchWeaviate(ctx context.Context, queryEmbedding []float32, topK int) ([]PipelineSearchResult, error) {
	results, err := p.weaviateAdapter.VectorSearch(ctx, p.config.CollectionName, queryEmbedding, topK, 0.0, nil)
	if err != nil {
		return nil, err
	}

	searchResults := make([]PipelineSearchResult, 0, len(results))
	for _, result := range results {
		content, _ := result.Properties["content"].(string)
		docID, _ := result.Properties["doc_id"].(string)
		chunkID, _ := result.Properties["chunk_id"].(string)
		startIdx, _ := result.Properties["start_idx"].(float64)
		endIdx, _ := result.Properties["end_idx"].(float64)

		chunk := PipelineChunk{
			ID:       chunkID,
			Content:  content,
			DocID:    docID,
			StartIdx: int(startIdx),
			EndIdx:   int(endIdx),
			Metadata: result.Properties,
		}

		searchResults = append(searchResults, PipelineSearchResult{
			Chunk:    chunk,
			Score:    result.Certainty,
			Distance: result.Distance,
			Metadata: result.Properties,
		})
	}

	return searchResults, nil
}

// SearchWithFilter performs semantic search with metadata filtering.
func (p *Pipeline) SearchWithFilter(ctx context.Context, query string, topK int, filter map[string]interface{}) ([]PipelineSearchResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, fmt.Errorf("pipeline not connected")
	}

	// Generate query embedding
	embeddings, _, err := p.embeddingRegistry.EncodeWithFallback(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}

	queryEmbedding := embeddings[0]

	switch p.config.VectorDBType {
	case VectorDBChroma:
		result, err := p.chromaAdapter.Query(ctx, p.config.CollectionName, [][]float32{queryEmbedding}, topK, filter)
		if err != nil {
			return nil, err
		}
		return p.convertChromaResults(result), nil
	case VectorDBQdrant:
		results, err := p.qdrantAdapter.Search(ctx, p.config.CollectionName, queryEmbedding, topK, filter, true, false)
		if err != nil {
			return nil, err
		}
		return p.convertQdrantResults(results), nil
	case VectorDBWeaviate:
		results, err := p.weaviateAdapter.VectorSearch(ctx, p.config.CollectionName, queryEmbedding, topK, 0.0, nil)
		if err != nil {
			return nil, err
		}
		return p.convertWeaviateResults(results), nil
	}

	return nil, fmt.Errorf("unsupported vector database type")
}

func (p *Pipeline) convertChromaResults(result *servers.ChromaQueryResult) []PipelineSearchResult {
	searchResults := make([]PipelineSearchResult, 0)
	if result != nil && len(result.IDs) > 0 && len(result.IDs[0]) > 0 {
		for i, id := range result.IDs[0] {
			var content string
			var metadata map[string]interface{}
			var distance float32

			if len(result.Documents) > 0 && len(result.Documents[0]) > i {
				content = result.Documents[0][i]
			}
			if len(result.Metadatas) > 0 && len(result.Metadatas[0]) > i {
				metadata = result.Metadatas[0][i]
			}
			if len(result.Distances) > 0 && len(result.Distances[0]) > i {
				distance = result.Distances[0][i]
			}

			chunk := PipelineChunk{
				ID:       id,
				Content:  content,
				Metadata: metadata,
			}
			searchResults = append(searchResults, PipelineSearchResult{
				Chunk:    chunk,
				Score:    1.0 - distance,
				Distance: distance,
				Metadata: metadata,
			})
		}
	}
	return searchResults
}

func (p *Pipeline) convertQdrantResults(results []servers.QdrantSearchResult) []PipelineSearchResult {
	searchResults := make([]PipelineSearchResult, 0, len(results))
	for _, result := range results {
		content, _ := result.Payload["content"].(string)
		var id string
		switch v := result.ID.(type) {
		case string:
			id = v
		case float64:
			id = fmt.Sprintf("%.0f", v)
		}
		chunk := PipelineChunk{
			ID:       id,
			Content:  content,
			Metadata: result.Payload,
		}
		searchResults = append(searchResults, PipelineSearchResult{
			Chunk:    chunk,
			Score:    result.Score,
			Distance: 1.0 - result.Score,
			Metadata: result.Payload,
		})
	}
	return searchResults
}

func (p *Pipeline) convertWeaviateResults(results []servers.WeaviateSearchResult) []PipelineSearchResult {
	searchResults := make([]PipelineSearchResult, 0, len(results))
	for _, result := range results {
		content, _ := result.Properties["content"].(string)
		chunk := PipelineChunk{
			ID:       result.ID,
			Content:  content,
			Metadata: result.Properties,
		}
		searchResults = append(searchResults, PipelineSearchResult{
			Chunk:    chunk,
			Score:    result.Certainty,
			Distance: result.Distance,
			Metadata: result.Properties,
		})
	}
	return searchResults
}

// DeleteDocument deletes a document and its chunks.
func (p *Pipeline) DeleteDocument(ctx context.Context, docID string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return fmt.Errorf("pipeline not connected")
	}

	switch p.config.VectorDBType {
	case VectorDBChroma:
		// Delete by filter - Chroma needs IDs, so we need to search first
		return fmt.Errorf("delete by filter not directly supported in Chroma, use DeleteDocuments with IDs")
	case VectorDBQdrant:
		// Use scroll to find matching points and delete them
		return fmt.Errorf("delete document requires listing and deleting points individually")
	case VectorDBWeaviate:
		// Delete by query not directly supported
		return fmt.Errorf("delete document requires finding and deleting objects individually")
	}

	return nil
}

// GetStats returns pipeline statistics.
func (p *Pipeline) GetStats(ctx context.Context) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, fmt.Errorf("pipeline not connected")
	}

	stats := map[string]interface{}{
		"vector_db_type":  p.config.VectorDBType,
		"collection_name": p.config.CollectionName,
		"embedding_model": p.config.EmbeddingModel,
		"connected":       p.connected,
	}

	switch p.config.VectorDBType {
	case VectorDBChroma:
		coll, err := p.chromaAdapter.GetCollection(ctx, p.config.CollectionName)
		if err == nil {
			stats["collection_info"] = coll
		}
		count, err := p.chromaAdapter.Count(ctx, p.config.CollectionName)
		if err == nil {
			stats["document_count"] = count
		}
	case VectorDBQdrant:
		info, err := p.qdrantAdapter.GetCollectionInfo(ctx, p.config.CollectionName)
		if err == nil {
			stats["collection_info"] = info
		}
		count, err := p.qdrantAdapter.CountPoints(ctx, p.config.CollectionName)
		if err == nil {
			stats["point_count"] = count
		}
	case VectorDBWeaviate:
		class, err := p.weaviateAdapter.GetClass(ctx, p.config.CollectionName)
		if err == nil {
			stats["class"] = class
		}
	}

	return stats, nil
}

// Helper functions

func generateDocID(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:16])
}

func generateChunkID(docID string, idx int) string {
	return fmt.Sprintf("%s_chunk_%d", docID, idx)
}

func copyMetadata(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	cpy := make(map[string]interface{}, len(m))
	for k, v := range m {
		cpy[k] = v
	}
	return cpy
}
