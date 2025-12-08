package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/superagent/superagent/internal/models"
	"github.com/superagent/superagent/internal/services"
	"github.com/superagent/superagent/internal/utils"
)

type CompletionHandler struct {
	requestService *services.RequestService
}

func NewCompletionHandler(requestService *services.RequestService) *CompletionHandler {
	return &CompletionHandler{
		requestService: requestService,
	}
}

func (h *CompletionHandler) Complete(c *gin.Context) {
	var req models.LLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, utils.NewAppError("INVALID_REQUEST", "Invalid request format", http.StatusBadRequest, err))
		return
	}

	// Generate request ID if not provided
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// Process the request
	response, err := h.requestService.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		utils.HandleError(c, utils.NewAppError("PROCESSING_ERROR", "Failed to process request", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, response)
}
