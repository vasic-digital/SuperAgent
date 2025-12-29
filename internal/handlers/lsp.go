package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superagent/superagent/internal/services"
)

// LSPHandler handles LSP (Language Server Protocol) requests
type LSPHandler struct {
	lspService *services.LSPManager
	log        *logrus.Logger
}

// NewLSPHandler creates a new LSP handler
func NewLSPHandler(lspService *services.LSPManager, log *logrus.Logger) *LSPHandler {
	return &LSPHandler{
		lspService: lspService,
		log:        log,
	}
}

// ListLSPServers handles GET /v1/lsp/servers
func (h *LSPHandler) ListLSPServers(c *gin.Context) {
	servers, err := h.lspService.ListLSPServers(c.Request.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to list LSP servers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, servers)
}

// ExecuteLSPRequest handles POST /v1/lsp/execute
func (h *LSPHandler) ExecuteLSPRequest(c *gin.Context) {
	var req services.UnifiedProtocolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Failed to bind LSP request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For LSP, the tool name represents the operation (completion, hover, etc.)
	// This would be expanded in a full implementation

	h.log.WithFields(logrus.Fields{
		"serverId":  req.ServerID,
		"operation": req.ToolName,
	}).Info("LSP request executed (placeholder)")

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"serverId":  req.ServerID,
		"operation": req.ToolName,
		"result":    "LSP operation completed successfully",
	})
}

// SyncLSPServer handles POST /v1/lsp/servers/:id/sync
func (h *LSPHandler) SyncLSPServer(c *gin.Context) {
	serverID := c.Param("id")

	// Placeholder implementation
	h.log.WithField("serverId", serverID).Info("LSP server sync completed")

	c.JSON(http.StatusOK, gin.H{"message": "LSP server synced successfully"})
}

// GetLSPStats handles GET /v1/lsp/stats
func (h *LSPHandler) GetLSPStats(c *gin.Context) {
	stats, err := h.lspService.GetLSPStats(c.Request.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to get LSP stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
