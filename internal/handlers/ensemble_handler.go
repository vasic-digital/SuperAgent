// Package handlers provides HTTP handlers for HelixAgent.
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"dev.helix.agent/internal/clis"
	"dev.helix.agent/internal/ensemble/multi_instance"
)

// EnsembleHandler handles ensemble-related endpoints.
type EnsembleHandler struct {
	coordinator *multi_instance.Coordinator
}

// NewEnsembleHandler creates a new ensemble handler.
func NewEnsembleHandler(coordinator *multi_instance.Coordinator) *EnsembleHandler {
	return &EnsembleHandler{coordinator: coordinator}
}

// RegisterRoutes registers the ensemble routes.
func (h *EnsembleHandler) RegisterRoutes(r *gin.RouterGroup) {
	sessions := r.Group("/ensemble/sessions")
	{
		sessions.POST("", h.CreateSession)
		sessions.GET("", h.ListSessions)
		sessions.GET("/:id", h.GetSession)
		sessions.POST("/:id/execute", h.ExecuteSession)
		sessions.POST("/:id/cancel", h.CancelSession)
	}
}

// CreateSessionRequest represents a create session request.
type CreateSessionRequest struct {
	Strategy     string   `json:"strategy" binding:"required"`
	Participants ParticipantRequest `json:"participants" binding:"required"`
}

// ParticipantRequest represents participant configuration.
type ParticipantRequest struct {
	Primary   *InstanceRequest   `json:"primary,omitempty"`
	Critiques []InstanceRequest  `json:"critiques,omitempty"`
	Verifiers []InstanceRequest  `json:"verifiers,omitempty"`
	Fallbacks []InstanceRequest  `json:"fallbacks,omitempty"`
}

// InstanceRequest represents instance configuration.
type InstanceRequest struct {
	Type     string                 `json:"type" binding:"required"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Provider map[string]interface{} `json:"provider,omitempty"`
}

// CreateSession creates a new ensemble session.
func (h *EnsembleHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to coordinator format
	participants := multi_instance.ParticipantConfig{}
	
	if req.Participants.Primary != nil {
		participants.Primary = multi_instance.InstanceConfig{
			Type: clis.AgentType(req.Participants.Primary.Type),
		}
	}
	
	for _, ic := range req.Participants.Critiques {
		participants.Critiques = append(participants.Critiques, multi_instance.InstanceConfig{
			Type: clis.AgentType(ic.Type),
		})
	}
	
	for _, ic := range req.Participants.Verifiers {
		participants.Verifiers = append(participants.Verifiers, multi_instance.InstanceConfig{
			Type: clis.AgentType(ic.Type),
		})
	}

	strategy := multi_instance.EnsembleStrategy(req.Strategy)
	config := multi_instance.DefaultEnsembleConfig()

	session, err := h.coordinator.CreateSession(c.Request.Context(), strategy, config, participants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       session.ID,
		"strategy": session.Strategy,
		"status":   session.Status,
	})
}

// ListSessions lists all ensemble sessions.
func (h *EnsembleHandler) ListSessions(c *gin.Context) {
	status := c.Query("status")
	sessions := h.coordinator.ListSessions(multi_instance.SessionStatus(status))
	
	var response []gin.H
	for _, s := range sessions {
		response = append(response, gin.H{
			"id":         s.ID,
			"strategy":   s.Strategy,
			"status":     s.Status,
			"created_at": s.CreatedAt,
		})
	}
	
	c.JSON(http.StatusOK, response)
}

// GetSession gets a session by ID.
func (h *EnsembleHandler) GetSession(c *gin.Context) {
	id := c.Param("id")
	
	session, err := h.coordinator.GetSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"id":         session.ID,
		"strategy":   session.Strategy,
		"status":     session.Status,
		"created_at": session.CreatedAt,
		"started_at": session.StartedAt,
	})
}

// ExecuteRequest represents an execute request.
type ExecuteRequest struct {
	Content string `json:"content" binding:"required"`
	Timeout int    `json:"timeout,omitempty"`
}

// ExecuteSession executes a task in a session.
func (h *EnsembleHandler) ExecuteSession(c *gin.Context) {
	id := c.Param("id")
	
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	task := multi_instance.Task{
		Type:    "completion",
		Content: req.Content,
	}
	
	if req.Timeout > 0 {
		task.Timeout = time.Duration(req.Timeout) * time.Second
	}
	
	result, err := h.coordinator.ExecuteSession(c.Request.Context(), id, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"consensus_reached": result.Reached,
		"confidence":        result.Confidence,
		"rounds":            result.Rounds,
		"results":           result.AllResults,
	})
}

// CancelSession cancels a session.
func (h *EnsembleHandler) CancelSession(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.coordinator.CancelSession(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}
