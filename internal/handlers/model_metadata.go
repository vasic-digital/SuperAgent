package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superagent/superagent/internal/database"
	"github.com/superagent/superagent/internal/services"
)

type ModelMetadataHandler struct {
	service *services.ModelMetadataService
}

func NewModelMetadataHandler(service *services.ModelMetadataService) *ModelMetadataHandler {
	return &ModelMetadataHandler{
		service: service,
	}
}

type ListModelsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Provider   string `form:"provider" binding:"omitempty"`
	ModelType  string `form:"type" binding:"omitempty"`
	Search     string `form:"search" binding:"omitempty"`
	Capability string `form:"capability" binding:"omitempty"`
}

type ListModelsResponse struct {
	Models     []*database.ModelMetadata `json:"models"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	Limit      int                       `json:"limit"`
	TotalPages int                       `json:"total_pages"`
}

type ModelDetailsResponse struct {
	*database.ModelMetadata
	Benchmarks []*database.ModelBenchmark `json:"benchmarks,omitempty"`
}

type RefreshResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RefreshHistoryResponse struct {
	Histories []*database.ModelsRefreshHistory `json:"histories"`
	Total     int                              `json:"total"`
}

func (h *ModelMetadataHandler) ListModels(c *gin.Context) {
	var req ListModelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}

	var models []*database.ModelMetadata
	var total int
	var err error

	ctx := c.Request.Context()

	if req.Search != "" {
		models, total, err = h.service.SearchModels(ctx, req.Search, req.Page, req.Limit)
	} else if req.Capability != "" {
		models, err = h.service.GetModelsByCapability(ctx, req.Capability)
		total = len(models)
	} else {
		models, total, err = h.service.ListModels(ctx, req.Provider, req.ModelType, req.Page, req.Limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list models", "details": err.Error()})
		return
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	c.JSON(http.StatusOK, ListModelsResponse{
		Models:     models,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	})
}

func (h *ModelMetadataHandler) GetModel(c *gin.Context) {
	modelID := c.Param("id")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
		return
	}

	ctx := c.Request.Context()
	metadata, err := h.service.GetModel(ctx, modelID)
	if err != nil {
		if err.Error() == "model metadata not found: "+modelID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get model", "details": err.Error()})
		return
	}

	repo, ok := c.Get("repository")
	if !ok {
		c.JSON(http.StatusOK, metadata)
		return
	}

	modelMetadataRepo := repo.(*database.ModelMetadataRepository)
	benchmarks, err := modelMetadataRepo.GetBenchmarks(ctx, modelID)
	if err != nil {
		c.JSON(http.StatusOK, metadata)
		return
	}

	response := ModelDetailsResponse{
		ModelMetadata: metadata,
		Benchmarks:    benchmarks,
	}

	c.JSON(http.StatusOK, response)
}

func (h *ModelMetadataHandler) GetModelBenchmarks(c *gin.Context) {
	modelID := c.Param("id")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
		return
	}

	repo, ok := c.Get("repository")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	ctx := c.Request.Context()
	modelMetadataRepo := repo.(*database.ModelMetadataRepository)
	benchmarks, err := modelMetadataRepo.GetBenchmarks(ctx, modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get benchmarks", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"benchmarks": benchmarks})
}

func (h *ModelMetadataHandler) CompareModels(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least 2 model IDs required for comparison"})
		return
	}
	if len(ids) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 10 models can be compared at once"})
		return
	}

	ctx := c.Request.Context()
	models, err := h.service.CompareModels(ctx, ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compare models", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (h *ModelMetadataHandler) RefreshModels(c *gin.Context) {
	providerID := c.Query("provider")

	ctx := c.Request.Context()

	if providerID != "" {
		err := h.service.RefreshProviderModels(ctx, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh provider models", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, RefreshResponse{
			Status:  "success",
			Message: "Provider models refresh initiated",
		})
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if err := h.service.RefreshModels(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh models", "details": err.Error()})
			return
		}
	}()

	c.JSON(http.StatusAccepted, RefreshResponse{
		Status:  "accepted",
		Message: "Full models refresh initiated",
	})
}

func (h *ModelMetadataHandler) GetRefreshStatus(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	ctx := c.Request.Context()
	histories, err := h.service.GetRefreshHistory(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get refresh history", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RefreshHistoryResponse{
		Histories: histories,
		Total:     len(histories),
	})
}

func (h *ModelMetadataHandler) GetProviderModels(c *gin.Context) {
	providerID := c.Param("provider_id")
	if providerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider ID is required"})
		return
	}

	ctx := c.Request.Context()
	models, err := h.service.GetProviderModels(ctx, providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider models", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider_id": providerID,
		"models":      models,
		"total":       len(models),
	})
}

func (h *ModelMetadataHandler) GetModelsByCapability(c *gin.Context) {
	capability := c.Param("capability")
	if capability == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Capability is required"})
		return
	}

	validCapabilities := map[string]bool{
		"vision":           true,
		"function_calling": true,
		"streaming":        true,
		"json_mode":        true,
		"image_generation": true,
		"audio":            true,
		"code_generation":  true,
		"reasoning":        true,
	}

	if !validCapabilities[capability] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid capability"})
		return
	}

	ctx := c.Request.Context()
	models, err := h.service.GetModelsByCapability(ctx, capability)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get models by capability", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"capability": capability,
		"models":     models,
		"total":      len(models),
	})
}
