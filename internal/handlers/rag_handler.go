package handlers

import (
	"net/http"
	"time"

	"dev.helix.agent/internal/embeddings/models"
	"dev.helix.agent/internal/rag"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RAGHandler handles all RAG API endpoints
type RAGHandler struct {
	pipeline    *rag.Pipeline
	advancedRAG *rag.AdvancedRAG
	logger      *logrus.Logger
}

// RAGHandlerConfig configures the RAG handler
type RAGHandlerConfig struct {
	Pipeline          *rag.Pipeline
	EmbeddingRegistry *models.EmbeddingModelRegistry
	Logger            *logrus.Logger
}

// NewRAGHandler creates a new RAG API handler
func NewRAGHandler(config RAGHandlerConfig) *RAGHandler {
	if config.Logger == nil {
		config.Logger = logrus.New()
	}

	var advancedRAG *rag.AdvancedRAG
	if config.Pipeline != nil {
		advancedConfig := rag.DefaultAdvancedRAGConfig()
		advancedRAG = rag.NewAdvancedRAG(advancedConfig, config.Pipeline)
	}

	return &RAGHandler{
		pipeline:    config.Pipeline,
		advancedRAG: advancedRAG,
		logger:      config.Logger,
	}
}

// =====================================================
// HEALTH & STATUS ENDPOINTS
// =====================================================

// Health checks RAG pipeline health
// GET /v1/rag/health
func (h *RAGHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()

	status := "healthy"
	code := http.StatusOK
	details := gin.H{}

	if h.pipeline == nil {
		status = "not_configured"
		code = http.StatusServiceUnavailable
		details["error"] = "RAG pipeline not initialized"
	} else {
		health := h.pipeline.Health(ctx)
		if health != nil {
			status = "unhealthy"
			code = http.StatusServiceUnavailable
			details["error"] = health.Error()
		}
	}

	c.JSON(code, gin.H{
		"status":  status,
		"details": details,
	})
}

// Stats returns RAG pipeline statistics
// GET /v1/rag/stats
func (h *RAGHandler) Stats(c *gin.Context) {
	if h.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	ctx := c.Request.Context()
	stats, err := h.pipeline.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// =====================================================
// DOCUMENT ENDPOINTS
// =====================================================

// IngestDocumentRequest represents a request to ingest a document
type IngestDocumentRequest struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
	Source   string                 `json:"source"`
}

// IngestDocument ingests a document into the RAG pipeline
// POST /v1/rag/documents
func (h *RAGHandler) IngestDocument(c *gin.Context) {
	if h.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	var req IngestDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	doc := &rag.PipelineDocument{
		ID:        req.ID,
		Content:   req.Content,
		Metadata:  req.Metadata,
		Source:    req.Source,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.pipeline.IngestDocument(ctx, doc); err != nil {
		h.logger.WithError(err).Error("Failed to ingest document")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Document ingested successfully",
		"document_id": doc.ID,
		"chunks":      len(doc.Chunks),
	})
}

// IngestDocumentsRequest represents a batch ingest request
type IngestDocumentsRequest struct {
	Documents []IngestDocumentRequest `json:"documents" binding:"required"`
}

// IngestDocuments ingests multiple documents
// POST /v1/rag/documents/batch
func (h *RAGHandler) IngestDocuments(c *gin.Context) {
	if h.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	var req IngestDocumentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	docs := make([]*rag.PipelineDocument, len(req.Documents))
	for i, d := range req.Documents {
		docs[i] = &rag.PipelineDocument{
			ID:        d.ID,
			Content:   d.Content,
			Metadata:  d.Metadata,
			Source:    d.Source,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	if err := h.pipeline.IngestDocuments(ctx, docs); err != nil {
		h.logger.WithError(err).Error("Failed to ingest documents")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalChunks := 0
	for _, doc := range docs {
		totalChunks += len(doc.Chunks)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Documents ingested successfully",
		"documents":    len(docs),
		"total_chunks": totalChunks,
	})
}

// DeleteDocument deletes a document from the RAG pipeline
// DELETE /v1/rag/documents/:id
func (h *RAGHandler) DeleteDocument(c *gin.Context) {
	if h.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document ID required"})
		return
	}

	ctx := c.Request.Context()

	if err := h.pipeline.DeleteDocument(ctx, documentID); err != nil {
		h.logger.WithError(err).Error("Failed to delete document")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Document deleted successfully",
		"document_id": documentID,
	})
}

// =====================================================
// SEARCH ENDPOINTS
// =====================================================

// SearchRequest represents a search request
type SearchRequest struct {
	Query   string                 `json:"query" binding:"required"`
	TopK    int                    `json:"top_k"`
	Filters map[string]interface{} `json:"filters"`
}

// Search performs a basic vector search
// POST /v1/rag/search
func (h *RAGHandler) Search(c *gin.Context) {
	if h.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 10
	}

	ctx := c.Request.Context()

	var results []rag.PipelineSearchResult
	var err error

	if len(req.Filters) > 0 {
		results, err = h.pipeline.SearchWithFilter(ctx, req.Query, req.TopK, req.Filters)
	} else {
		results, err = h.pipeline.Search(ctx, req.Query, req.TopK)
	}

	if err != nil {
		h.logger.WithError(err).Error("Search failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": results,
		"count":   len(results),
	})
}

// HybridSearchRequest represents a hybrid search request
type HybridSearchRequest struct {
	Query         string  `json:"query" binding:"required"`
	TopK          int     `json:"top_k"`
	VectorWeight  float64 `json:"vector_weight"`
	KeywordWeight float64 `json:"keyword_weight"`
}

// HybridSearch performs hybrid search combining vector and keyword matching
// POST /v1/rag/search/hybrid
func (h *RAGHandler) HybridSearch(c *gin.Context) {
	if h.pipeline == nil || h.advancedRAG == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	var req HybridSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 10
	}

	ctx := c.Request.Context()

	// Initialize advanced RAG if needed
	if err := h.advancedRAG.Initialize(ctx); err != nil {
		h.logger.WithError(err).Error("Failed to initialize advanced RAG")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	results, err := h.advancedRAG.HybridSearch(ctx, req.Query, req.TopK)
	if err != nil {
		h.logger.WithError(err).Error("Hybrid search failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": results,
		"count":   len(results),
		"type":    "hybrid",
	})
}

// ExpandedSearchRequest represents a search with query expansion
type ExpandedSearchRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k"`
}

// SearchWithExpansion performs search with query expansion
// POST /v1/rag/search/expanded
func (h *RAGHandler) SearchWithExpansion(c *gin.Context) {
	if h.pipeline == nil || h.advancedRAG == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	var req ExpandedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 10
	}

	ctx := c.Request.Context()

	// Initialize advanced RAG if needed
	if err := h.advancedRAG.Initialize(ctx); err != nil {
		h.logger.WithError(err).Error("Failed to initialize advanced RAG")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get query expansions
	expansions := h.advancedRAG.ExpandQuery(ctx, req.Query)

	// Search with expansion
	results, err := h.advancedRAG.SearchWithExpansion(ctx, req.Query, req.TopK)
	if err != nil {
		h.logger.WithError(err).Error("Expanded search failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":      req.Query,
		"expansions": expansions,
		"results":    results,
		"count":      len(results),
		"type":       "expanded",
	})
}

// =====================================================
// RE-RANKING & COMPRESSION ENDPOINTS
// =====================================================

// ReRankRequest represents a re-ranking request
type ReRankRequest struct {
	Query   string                    `json:"query" binding:"required"`
	Results []rag.PipelineSearchResult `json:"results" binding:"required"`
}

// ReRank re-ranks search results
// POST /v1/rag/rerank
func (h *RAGHandler) ReRank(c *gin.Context) {
	if h.advancedRAG == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Advanced RAG not initialized"})
		return
	}

	var req ReRankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Initialize advanced RAG if needed
	if err := h.advancedRAG.Initialize(ctx); err != nil {
		h.logger.WithError(err).Error("Failed to initialize advanced RAG")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reranked, err := h.advancedRAG.ReRank(ctx, req.Query, req.Results)
	if err != nil {
		h.logger.WithError(err).Error("Re-ranking failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": reranked,
		"count":   len(reranked),
	})
}

// CompressRequest represents a context compression request
type CompressRequest struct {
	Query   string                    `json:"query" binding:"required"`
	Results []rag.PipelineSearchResult `json:"results" binding:"required"`
}

// CompressContext compresses search results into condensed context
// POST /v1/rag/compress
func (h *RAGHandler) CompressContext(c *gin.Context) {
	if h.advancedRAG == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Advanced RAG not initialized"})
		return
	}

	var req CompressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Initialize advanced RAG if needed
	if err := h.advancedRAG.Initialize(ctx); err != nil {
		h.logger.WithError(err).Error("Failed to initialize advanced RAG")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	compressed, err := h.advancedRAG.CompressContext(ctx, req.Query, req.Results)
	if err != nil {
		h.logger.WithError(err).Error("Compression failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, compressed)
}

// =====================================================
// QUERY EXPANSION ENDPOINT
// =====================================================

// ExpandQueryRequest represents a query expansion request
type ExpandQueryRequest struct {
	Query string `json:"query" binding:"required"`
}

// ExpandQuery returns query expansions without searching
// POST /v1/rag/expand
func (h *RAGHandler) ExpandQuery(c *gin.Context) {
	if h.advancedRAG == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Advanced RAG not initialized"})
		return
	}

	var req ExpandQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Initialize advanced RAG if needed
	if err := h.advancedRAG.Initialize(ctx); err != nil {
		h.logger.WithError(err).Error("Failed to initialize advanced RAG")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	expansions := h.advancedRAG.ExpandQuery(ctx, req.Query)

	c.JSON(http.StatusOK, gin.H{
		"query":      req.Query,
		"expansions": expansions,
		"count":      len(expansions),
	})
}

// =====================================================
// CHUNKING ENDPOINT
// =====================================================

// ChunkDocumentRequest represents a chunking request
type ChunkDocumentRequest struct {
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ChunkDocument chunks a document without storing
// POST /v1/rag/chunk
func (h *RAGHandler) ChunkDocument(c *gin.Context) {
	if h.pipeline == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RAG pipeline not initialized"})
		return
	}

	var req ChunkDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc := &rag.PipelineDocument{
		Content:  req.Content,
		Metadata: req.Metadata,
	}

	chunks := h.pipeline.ChunkDocument(doc)

	// Convert to response format
	chunkResponses := make([]gin.H, len(chunks))
	for i, chunk := range chunks {
		chunkResponses[i] = gin.H{
			"id":        chunk.ID,
			"content":   chunk.Content,
			"start_idx": chunk.StartIdx,
			"end_idx":   chunk.EndIdx,
			"metadata":  chunk.Metadata,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"chunks": chunkResponses,
		"count":  len(chunks),
	})
}
