package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superagent/superagent/internal/config"
	llm "github.com/superagent/superagent/internal/llm/cognee"
)

// CogneeHandler handles Cognee-related requests
type CogneeHandler struct {
	client *llm.Client
}

// NewCogneeHandler creates a new Cognee handler
func NewCogneeHandler(cfg *config.Config) *CogneeHandler {
	return &CogneeHandler{
		client: llm.NewClient(cfg),
	}
}

// VisualizeGraph returns graph visualization data
func (h *CogneeHandler) VisualizeGraph(c *gin.Context) {
	datasetName := c.Query("dataset")
	if datasetName == "" {
		datasetName = "default"
	}

	req := &llm.VisualizeRequest{
		DatasetName: datasetName,
		Format:      c.Query("format"), // "json", "graphml", etc.
	}

	resp, err := h.client.VisualizeGraph(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to visualize graph: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp.Graph)
}

// GetDatasets returns list of available datasets
func (h *CogneeHandler) GetDatasets(c *gin.Context) {
	resp, err := h.client.ListDatasets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list datasets: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"datasets": resp.Datasets,
		"total":    resp.Total,
	})
}
