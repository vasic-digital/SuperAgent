// Package handlers provides HTTP handlers
package handlers

import (
	"net/http"

	"dev.helix.agent/internal/templates"
	"github.com/gin-gonic/gin"
)

// TemplateHandler handles template-related HTTP requests
type TemplateHandler struct {
	manager *templates.TemplateManager
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(manager *templates.TemplateManager) *TemplateHandler {
	return &TemplateHandler{manager: manager}
}

// ListTemplatesRequest represents a request to list templates
type ListTemplatesRequest struct {
	Tag string `json:"tag,omitempty" form:"tag"`
}

// ListTemplatesResponse represents the list of templates
type ListTemplatesResponse struct {
	Templates []TemplateInfo `json:"templates"`
}

// TemplateInfo represents template metadata
type TemplateInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// ContextTemplate is an alias for templates.ContextTemplate
type ContextTemplate = templates.ContextTemplate

// ApplyTemplateRequest represents a template application request
type ApplyTemplateRequest struct {
	TemplateID string            `json:"template_id" binding:"required"`
	Variables  map[string]string `json:"variables,omitempty"`
}

// ApplyTemplateResponse represents the result of applying a template
type ApplyTemplateResponse struct {
	FilesLoaded   int    `json:"files_loaded"`
	Instructions  string `json:"instructions"`
	ContextLength int    `json:"context_length"`
}

// ListTemplates lists all available templates
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	var req ListTemplatesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templateList := h.manager.List()

	// Filter by tag if specified
	var filtered []*templates.ContextTemplate
	if req.Tag != "" {
		for _, t := range templateList {
			for _, tag := range t.Metadata.Tags {
				if tag == req.Tag {
					filtered = append(filtered, t)
					break
				}
			}
		}
	} else {
		filtered = templateList
	}

	// Convert to response format
	infos := make([]TemplateInfo, len(filtered))
	for i, t := range filtered {
		infos[i] = TemplateInfo{
			ID:          t.Metadata.ID,
			Name:        t.Metadata.Name,
			Description: t.Metadata.Description,
			Tags:        t.Metadata.Tags,
		}
	}

	c.JSON(http.StatusOK, ListTemplatesResponse{Templates: infos})
}

// GetTemplate returns a specific template
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	template, err := h.manager.GetTemplate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ApplyTemplate applies a template with variables
func (h *TemplateHandler) ApplyTemplate(c *gin.Context) {
	var req ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.manager.ApplyTemplate(req.TemplateID, req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := ApplyTemplateResponse{
		FilesLoaded:   len(result.Files),
		Instructions:  result.Instructions,
		ContextLength: len(result.FormatContext()),
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes registers the template routes
func (h *TemplateHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/v1/templates")
	{
		v1.GET("", h.ListTemplates)
		v1.GET("/:id", h.GetTemplate)
		v1.POST("/apply", h.ApplyTemplate)
	}
}
