// Package handlers provides HTTP handlers
package handlers

import (
	"net/http"
	"time"

	"dev.helix.agent/internal/checkpoints"
	"github.com/gin-gonic/gin"
)

// CheckpointHandler handles checkpoint-related HTTP requests
type CheckpointHandler struct {
	manager *checkpoints.Manager
}

// NewCheckpointHandler creates a new checkpoint handler
func NewCheckpointHandler(manager *checkpoints.Manager) *CheckpointHandler {
	return &CheckpointHandler{manager: manager}
}

// CreateCheckpointRequest represents a checkpoint creation request
type CreateCheckpointRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// CreateCheckpointResponse represents the result of creating a checkpoint
type CreateCheckpointResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	GitRef    string    `json:"git_ref,omitempty"`
	GitBranch string    `json:"git_branch,omitempty"`
	FileCount int       `json:"file_count"`
}

// CheckpointInfo represents checkpoint metadata
type CheckpointInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	GitRef      string    `json:"git_ref,omitempty"`
	GitBranch   string    `json:"git_branch,omitempty"`
	Tags        []string  `json:"tags"`
	FileCount   int       `json:"file_count"`
}

// ListCheckpointsResponse represents the list of checkpoints
type ListCheckpointsResponse struct {
	Checkpoints []CheckpointInfo `json:"checkpoints"`
}

// CreateCheckpoint creates a new checkpoint
func (h *CheckpointHandler) CreateCheckpoint(c *gin.Context) {
	var req CreateCheckpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checkpoint, err := h.manager.Create(req.Name, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := CreateCheckpointResponse{
		ID:        checkpoint.ID,
		Name:      checkpoint.Name,
		CreatedAt: checkpoint.CreatedAt,
		GitRef:    checkpoint.GitRef,
		GitBranch: checkpoint.GitBranch,
		FileCount: len(checkpoint.Files),
	}

	c.JSON(http.StatusCreated, resp)
}

// ListCheckpoints lists all checkpoints
func (h *CheckpointHandler) ListCheckpoints(c *gin.Context) {
	checkpointList, err := h.manager.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	infos := make([]CheckpointInfo, len(checkpointList))
	for i, cp := range checkpointList {
		infos[i] = CheckpointInfo{
			ID:          cp.ID,
			Name:        cp.Name,
			Description: cp.Description,
			CreatedAt:   cp.CreatedAt,
			GitRef:      cp.GitRef,
			GitBranch:   cp.GitBranch,
			Tags:        cp.Tags,
			FileCount:   len(cp.Files),
		}
	}

	c.JSON(http.StatusOK, ListCheckpointsResponse{Checkpoints: infos})
}

// RestoreCheckpoint restores a checkpoint
func (h *CheckpointHandler) RestoreCheckpoint(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Restore(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "checkpoint restored successfully"})
}

// DeleteCheckpoint deletes a checkpoint
func (h *CheckpointHandler) DeleteCheckpoint(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "checkpoint deleted successfully"})
}

// RegisterRoutes registers the checkpoint routes
func (h *CheckpointHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/v1/checkpoints")
	{
		v1.GET("", h.ListCheckpoints)
		v1.POST("", h.CreateCheckpoint)
		v1.POST("/:id/restore", h.RestoreCheckpoint)
		v1.DELETE("/:id", h.DeleteCheckpoint)
	}
}
