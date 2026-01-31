package bigdata

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler provides HTTP endpoints for big data operations
type Handler struct {
	integration       *BigDataIntegration
	debateIntegration *DebateIntegration
	logger            *logrus.Logger
}

// NewHandler creates a new big data handler
func NewHandler(
	integration *BigDataIntegration,
	debateIntegration *DebateIntegration,
	logger *logrus.Logger,
) *Handler {
	return &Handler{
		integration:       integration,
		debateIntegration: debateIntegration,
		logger:            logger,
	}
}

// RegisterRoutes registers all big data routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		// Context endpoints
		context := v1.Group("/context")
		{
			context.POST("/replay", h.ReplayConversation)
			context.GET("/stats/:conversation_id", h.GetContextStats)
		}

		// Memory synchronization endpoints
		memory := v1.Group("/memory")
		{
			memory.GET("/sync/status", h.GetMemorySyncStatus)
			memory.POST("/sync/force", h.ForceMemorySync)
		}

		// Knowledge graph endpoints
		knowledge := v1.Group("/knowledge")
		{
			knowledge.GET("/related/:entity_id", h.GetRelatedEntities)
			knowledge.POST("/search", h.SearchKnowledgeGraph)
		}

		// Analytics endpoints
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/provider/:provider", h.GetProviderAnalytics)
			analytics.GET("/debate/:debate_id", h.GetDebateAnalytics)
			analytics.POST("/query", h.QueryAnalytics)
		}

		// Learning endpoints
		learning := v1.Group("/learning")
		{
			learning.GET("/insights", h.GetLearningInsights)
			learning.GET("/patterns", h.GetLearnedPatterns)
		}

		// Health check
		v1.GET("/bigdata/health", h.HealthCheck)
	}
}

// ReplayConversationRequest represents a context replay request
type ReplayConversationRequest struct {
	ConversationID      string `json:"conversation_id" binding:"required"`
	MaxTokens           int    `json:"max_tokens,omitempty"`
	CompressionStrategy string `json:"compression_strategy,omitempty"` // "window", "entity", "full", "hybrid"
}

// ReplayConversation replays a conversation from Kafka with optional compression
func (h *Handler) ReplayConversation(c *gin.Context) {
	var req ReplayConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MaxTokens == 0 {
		req.MaxTokens = 4000 // Default context window
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if h.debateIntegration == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "debate integration not available (infinite context disabled)"})
		return
	}

	conversationCtx, err := h.debateIntegration.GetConversationContext(
		ctx,
		req.ConversationID,
		req.MaxTokens,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to replay conversation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "context replay failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation_id":   conversationCtx.ConversationID,
		"messages":          conversationCtx.Messages,
		"entities":          conversationCtx.Entities,
		"total_tokens":      conversationCtx.TotalTokens,
		"compressed":        conversationCtx.Compressed,
		"compression_stats": conversationCtx.CompressionStats,
	})
}

// GetContextStats returns statistics about a conversation's context
func (h *Handler) GetContextStats(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	// Get infinite context engine
	infiniteContext := h.integration.GetInfiniteContext()
	if infiniteContext == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "infinite context not enabled"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Get conversation snapshot
	snapshot, err := infiniteContext.GetConversationSnapshot(ctx, conversationID)
	if err != nil {
		h.logger.WithError(err).WithField("conversation_id", conversationID).Error("Failed to get conversation snapshot")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation snapshot"})
		return
	}

	// Extract statistics
	messageCount := 0
	entityCount := 0
	totalTokens := int64(0)
	compressed := false

	if snapshot.Context != nil {
		messageCount = snapshot.Context.MessageCount
		entityCount = snapshot.Context.EntityCount
		totalTokens = snapshot.Context.TotalTokens
		compressed = snapshot.Context.CompressedCount > 0
	} else {
		// Fallback to counting messages and entities directly
		messageCount = len(snapshot.Messages)
		entityCount = len(snapshot.Entities)
		// Estimate tokens (rough average)
		totalTokens = int64(messageCount * 100)
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation_id": conversationID,
		"message_count":   messageCount,
		"entity_count":    entityCount,
		"total_tokens":    totalTokens,
		"compressed":      compressed,
		"snapshot_id":     snapshot.SnapshotID,
		"timestamp":       snapshot.Timestamp,
	})
}

// GetMemorySyncStatus returns the status of distributed memory synchronization
func (h *Handler) GetMemorySyncStatus(c *gin.Context) {
	if h.integration.distributedMemory == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "distributed memory not enabled"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	status := h.integration.distributedMemory.GetSyncStatus(ctx)
	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// ForceMemorySync forces synchronization of all memory nodes
func (h *Handler) ForceMemorySync(c *gin.Context) {
	if h.integration.distributedMemory == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "distributed memory not enabled"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.integration.distributedMemory.ForceSync(ctx); err != nil {
		h.logger.WithError(err).Error("Failed to force memory sync")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to force memory sync",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "sync_request_sent",
		"message": "Sync request published to Kafka",
	})
}

// GetRelatedEntities finds entities related to a given entity
func (h *Handler) GetRelatedEntities(c *gin.Context) {
	entityID := c.Param("entity_id")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id is required"})
		return
	}

	if h.integration.graphStreaming == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "knowledge graph not enabled"})
		return
	}

	maxDepth := 1
	if depthStr := c.DefaultQuery("max_depth", "1"); depthStr != "" {
		if depth, err := strconv.Atoi(depthStr); err == nil && depth > 0 && depth <= 5 {
			maxDepth = depth
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	entities, err := h.integration.graphStreaming.GetRelatedEntities(ctx, entityID, maxDepth)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to get related entities")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query related entities"})
		return
	}

	related := make([]map[string]interface{}, len(entities))
	for i, entity := range entities {
		related[i] = map[string]interface{}{
			"id":         entity.ID,
			"type":       entity.Type,
			"name":       entity.Name,
			"value":      entity.Value,
			"confidence": entity.Confidence,
			"importance": entity.Importance,
			"properties": entity.Properties,
			"created_at": entity.CreatedAt,
			"updated_at": entity.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id": entityID,
		"max_depth": maxDepth,
		"related":   related,
	})
}

// SearchKnowledgeGraphRequest represents a knowledge graph search request
type SearchKnowledgeGraphRequest struct {
	Query      string   `json:"query" binding:"required"`
	EntityType string   `json:"entity_type,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Filters    []string `json:"filters,omitempty"`
}

// SearchKnowledgeGraph searches the knowledge graph
func (h *Handler) SearchKnowledgeGraph(c *gin.Context) {
	var req SearchKnowledgeGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.integration.graphStreaming == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "knowledge graph not enabled"})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	entities, err := h.integration.graphStreaming.SearchKnowledgeGraph(ctx, req.Query, req.EntityType, req.Limit)
	if err != nil {
		h.logger.WithError(err).WithField("query", req.Query).Error("Failed to search knowledge graph")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search knowledge graph"})
		return
	}

	results := make([]map[string]interface{}, len(entities))
	for i, entity := range entities {
		results[i] = map[string]interface{}{
			"id":         entity.ID,
			"type":       entity.Type,
			"name":       entity.Name,
			"value":      entity.Value,
			"confidence": entity.Confidence,
			"importance": entity.Importance,
			"properties": entity.Properties,
			"created_at": entity.CreatedAt,
			"updated_at": entity.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// GetProviderAnalytics returns analytics for a specific provider
func (h *Handler) GetProviderAnalytics(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	if h.integration.clickhouseAnalytics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "analytics not enabled"})
		return
	}

	window := c.DefaultQuery("window", "24h")
	duration, err := time.ParseDuration(window)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid window duration"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.integration.clickhouseAnalytics.GetProviderAnalytics(ctx, provider, duration)
	if err != nil {
		h.logger.WithError(err).WithField("provider", provider).Error("Failed to get provider analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query provider analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider":           stats.Provider,
		"window":             stats.Period,
		"total_requests":     stats.TotalRequests,
		"avg_response_time":  stats.AvgResponseTime,
		"p95_response_time":  stats.P95ResponseTime,
		"p99_response_time":  stats.P99ResponseTime,
		"avg_confidence":     stats.AvgConfidence,
		"total_tokens":       stats.TotalTokens,
		"avg_tokens_per_req": stats.AvgTokensPerReq,
		"error_rate":         stats.ErrorRate,
		"win_rate":           stats.WinRate,
	})
}

// GetDebateAnalytics returns analytics for a specific debate
func (h *Handler) GetDebateAnalytics(c *gin.Context) {
	debateID := c.Param("debate_id")
	if debateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "debate_id is required"})
		return
	}

	if h.integration.clickhouseAnalytics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "analytics not enabled"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	analytics, err := h.integration.clickhouseAnalytics.GetDebateAnalytics(ctx, debateID)
	if err != nil {
		h.logger.WithError(err).WithField("debate_id", debateID).Error("Failed to get debate analytics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query debate analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"debate_id":         analytics["debate_id"],
		"rounds":            analytics["total_rounds"],
		"participants":      analytics["participants"],
		"avg_response_time": analytics["avg_response_time"],
		"total_tokens":      analytics["total_tokens"],
		"winner":            analytics["winner"],
		"confidence":        analytics["avg_confidence"],
	})
}

// QueryAnalyticsRequest represents a custom analytics query
type QueryAnalyticsRequest struct {
	Query      string                 `json:"query" binding:"required"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Format     string                 `json:"format,omitempty"` // "json", "csv"
}

// QueryAnalytics executes a custom analytics query
func (h *Handler) QueryAnalytics(c *gin.Context) {
	var req QueryAnalyticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.integration.clickhouseAnalytics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "analytics not enabled"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	results, err := h.integration.clickhouseAnalytics.ExecuteQuery(ctx, req.Query, req.Parameters)
	if err != nil {
		h.logger.WithError(err).WithField("query", req.Query).Error("Failed to execute analytics query")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query execution failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// GetLearningInsights returns recent learning insights
func (h *Handler) GetLearningInsights(c *gin.Context) {
	crossLearner := h.integration.GetCrossLearner()
	if crossLearner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cross-session learning not enabled"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}
	if limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}

	// Get insights from cross-session learner
	insights := crossLearner.GetInsights(limit)

	// Convert to JSON-friendly format
	insightMaps := make([]map[string]interface{}, len(insights))
	for i, insight := range insights {
		insightMaps[i] = map[string]interface{}{
			"insight_id":   insight.InsightID,
			"user_id":      insight.UserID,
			"insight_type": insight.InsightType,
			"title":        insight.Title,
			"description":  insight.Description,
			"confidence":   insight.Confidence,
			"impact":       insight.Impact,
			"created_at":   insight.CreatedAt,
			"metadata":     insight.Metadata,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"insights": insightMaps,
		"count":    len(insights),
		"limit":    limit,
	})
}

// GetLearnedPatterns returns learned patterns across conversations
func (h *Handler) GetLearnedPatterns(c *gin.Context) {
	crossLearner := h.integration.GetCrossLearner()
	if crossLearner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cross-session learning not enabled"})
		return
	}

	patternType := c.DefaultQuery("type", "all")

	// Get patterns from cross-session learner
	patterns := crossLearner.GetPatterns(patternType)

	// Convert to JSON-friendly format
	patternMaps := make([]map[string]interface{}, len(patterns))
	for i, pattern := range patterns {
		patternMaps[i] = map[string]interface{}{
			"pattern_id":   pattern.PatternID,
			"pattern_type": pattern.PatternType,
			"description":  pattern.Description,
			"frequency":    pattern.Frequency,
			"confidence":   pattern.Confidence,
			"first_seen":   pattern.FirstSeen,
			"last_seen":    pattern.LastSeen,
			"examples":     pattern.Examples,
			"metadata":     pattern.Metadata,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"patterns": patternMaps,
		"type":     patternType,
		"count":    len(patterns),
	})
}

// HealthCheck returns the health status of all big data components
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := h.integration.HealthCheck(ctx)

	// Determine overall status
	status := "healthy"
	for component, componentStatus := range health {
		if componentStatus == "unhealthy" || componentStatus == "not_initialized" {
			status = "degraded"
			h.logger.WithField("component", component).
				WithField("status", componentStatus).
				Warn("Big data component unhealthy")
		}
	}

	statusCode := http.StatusOK
	if status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":     status,
		"components": health,
		"running":    h.integration.IsRunning(),
	})
}
