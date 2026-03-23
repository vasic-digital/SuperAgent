package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dev.helix.agent/internal/llmops"
)

// LLMOpsHandler provides HTTP endpoints for LLM operations management.
// Endpoints:
//
//	POST /v1/llmops/experiments     - create an A/B experiment
//	GET  /v1/llmops/experiments     - list experiments
//	GET  /v1/llmops/experiments/:id - get experiment details
//	POST /v1/llmops/evaluate        - run continuous evaluation
//	GET  /v1/llmops/prompts         - list prompt versions
//	POST /v1/llmops/prompts         - create prompt version
type LLMOpsHandler struct {
	system *llmops.LLMOpsSystem
}

// NewLLMOpsHandler creates a new LLMOps handler
func NewLLMOpsHandler(system *llmops.LLMOpsSystem) *LLMOpsHandler {
	return &LLMOpsHandler{
		system: system,
	}
}

// checkSystem returns false and writes a 503 response if the LLMOps system
// has not been initialised.
func (h *LLMOpsHandler) checkSystem(c *gin.Context) bool {
	if h.system == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "LLMOps system not initialized",
		})
		return false
	}
	return true
}

// --- request / response types ---

// CreateExperimentRequest represents a request to create an A/B experiment
type CreateExperimentRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description,omitempty"`
	Variants     []*llmops.Variant      `json:"variants" binding:"required,min=2"`
	TrafficSplit map[string]float64     `json:"traffic_split,omitempty"`
	Metrics      []string               `json:"metrics,omitempty"`
	TargetMetric string                 `json:"target_metric,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ExperimentResponse represents an experiment in API responses
type ExperimentResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Variants     []*llmops.Variant      `json:"variants"`
	TrafficSplit map[string]float64     `json:"traffic_split"`
	Status       llmops.ExperimentStatus `json:"status"`
	Metrics      []string               `json:"metrics"`
	TargetMetric string                 `json:"target_metric"`
	Winner       string                 `json:"winner,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// ListExperimentsResponse represents the list experiments response
type ListExperimentsResponse struct {
	Experiments []ExperimentResponse `json:"experiments"`
	Total       int                  `json:"total"`
}

// CreateEvaluationRequest represents a request to create an evaluation run
type CreateEvaluationRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description,omitempty"`
	Dataset       string                 `json:"dataset" binding:"required"`
	PromptName    string                 `json:"prompt_name,omitempty"`
	PromptVersion string                 `json:"prompt_version,omitempty"`
	ModelName     string                 `json:"model_name,omitempty"`
	Metrics       []string               `json:"metrics,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// EvaluationRunResponse represents an evaluation run in API responses
type EvaluationRunResponse struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description,omitempty"`
	Dataset       string                  `json:"dataset"`
	PromptName    string                  `json:"prompt_name,omitempty"`
	PromptVersion string                  `json:"prompt_version,omitempty"`
	ModelName     string                  `json:"model_name,omitempty"`
	Metrics       []string                `json:"metrics,omitempty"`
	Status        llmops.EvaluationStatus `json:"status"`
	Results       *llmops.EvaluationResults `json:"results,omitempty"`
	CreatedAt     string                  `json:"created_at"`
}

// CreatePromptVersionRequest represents a request to create a prompt version
type CreatePromptVersionRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Version     string                 `json:"version" binding:"required"`
	Content     string                 `json:"content" binding:"required"`
	Variables   []llmops.PromptVariable `json:"variables,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Author      string                 `json:"author,omitempty"`
	Description string                 `json:"description,omitempty"`
	IsActive    bool                   `json:"is_active,omitempty"`
}

// PromptVersionResponse represents a prompt version in API responses
type PromptVersionResponse struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Version     string                  `json:"version"`
	Content     string                  `json:"content"`
	Variables   []llmops.PromptVariable `json:"variables,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
	Author      string                  `json:"author,omitempty"`
	Description string                  `json:"description,omitempty"`
	IsActive    bool                    `json:"is_active"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

// ListPromptsResponse represents the list prompts response
type ListPromptsResponse struct {
	Prompts []PromptVersionResponse `json:"prompts"`
	Total   int                     `json:"total"`
}

// --- handlers ---

// CreateExperiment godoc
// @Summary Create an A/B experiment
// @Description Create a new A/B testing experiment for LLM evaluation
// @Tags llmops
// @Accept json
// @Produce json
// @Param request body CreateExperimentRequest true "Experiment details"
// @Success 201 {object} ExperimentResponse
// @Failure 400 {object} VerifierErrorResponse
// @Failure 500 {object} VerifierErrorResponse
// @Router /v1/llmops/experiments [post]
func (h *LLMOpsHandler) CreateExperiment(c *gin.Context) {
	if !h.checkSystem(c) {
		return
	}
	var req CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, VerifierErrorResponse{Error: err.Error()})
		return
	}

	exp := &llmops.Experiment{
		Name:         req.Name,
		Description:  req.Description,
		Variants:     req.Variants,
		TrafficSplit: req.TrafficSplit,
		Metrics:      req.Metrics,
		TargetMetric: req.TargetMetric,
		Metadata:     req.Metadata,
	}

	mgr := h.system.GetExperimentManager()
	if mgr == nil {
		c.JSON(
			http.StatusInternalServerError,
			VerifierErrorResponse{Error: "experiment manager not initialized"},
		)
		return
	}

	if err := mgr.Create(c.Request.Context(), exp); err != nil {
		c.JSON(http.StatusBadRequest, VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, experimentToResponse(exp))
}

// ListExperiments godoc
// @Summary List experiments
// @Description List all A/B testing experiments, optionally filtered by status
// @Tags llmops
// @Produce json
// @Param status query string false "Filter by status (draft, running, paused, completed, cancelled)"
// @Success 200 {object} ListExperimentsResponse
// @Failure 500 {object} VerifierErrorResponse
// @Router /v1/llmops/experiments [get]
func (h *LLMOpsHandler) ListExperiments(c *gin.Context) {

	if !h.checkSystem(c) {
		return
	}
	mgr := h.system.GetExperimentManager()
	if mgr == nil {
		c.JSON(
			http.StatusInternalServerError,
			VerifierErrorResponse{Error: "experiment manager not initialized"},
		)
		return
	}

	status := llmops.ExperimentStatus(c.Query("status"))

	experiments, err := mgr.List(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, VerifierErrorResponse{Error: err.Error()})
		return
	}

	resp := ListExperimentsResponse{
		Experiments: make([]ExperimentResponse, len(experiments)),
		Total:       len(experiments),
	}
	for i, exp := range experiments {
		resp.Experiments[i] = experimentToResponse(exp)
	}

	c.JSON(http.StatusOK, resp)
}

// GetExperiment godoc
// @Summary Get experiment details
// @Description Get details of a specific experiment by ID
// @Tags llmops
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} ExperimentResponse
// @Failure 404 {object} VerifierErrorResponse
// @Router /v1/llmops/experiments/{id} [get]
func (h *LLMOpsHandler) GetExperiment(c *gin.Context) {

	if !h.checkSystem(c) {
		return
	}
	mgr := h.system.GetExperimentManager()
	if mgr == nil {
		c.JSON(
			http.StatusInternalServerError,
			VerifierErrorResponse{Error: "experiment manager not initialized"},
		)
		return
	}

	id := c.Param("id")

	exp, err := mgr.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, experimentToResponse(exp))
}

// CreateEvaluation godoc
// @Summary Run continuous evaluation
// @Description Create and start a continuous evaluation run
// @Tags llmops
// @Accept json
// @Produce json
// @Param request body CreateEvaluationRequest true "Evaluation run details"
// @Success 201 {object} EvaluationRunResponse
// @Failure 400 {object} VerifierErrorResponse
// @Failure 500 {object} VerifierErrorResponse
// @Router /v1/llmops/evaluate [post]
func (h *LLMOpsHandler) CreateEvaluation(c *gin.Context) {

	if !h.checkSystem(c) {
		return
	}
	var req CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, VerifierErrorResponse{Error: err.Error()})
		return
	}

	eval := h.system.GetEvaluator()
	if eval == nil {
		c.JSON(
			http.StatusInternalServerError,
			VerifierErrorResponse{Error: "evaluator not initialized"},
		)
		return
	}

	run := &llmops.EvaluationRun{
		Name:          req.Name,
		Description:   req.Description,
		Dataset:       req.Dataset,
		PromptName:    req.PromptName,
		PromptVersion: req.PromptVersion,
		ModelName:     req.ModelName,
		Metrics:       req.Metrics,
		Metadata:      req.Metadata,
	}

	if err := eval.CreateRun(c.Request.Context(), run); err != nil {
		c.JSON(http.StatusBadRequest, VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, evaluationRunToResponse(run))
}

// ListPrompts godoc
// @Summary List prompt versions
// @Description List all prompt versions
// @Tags llmops
// @Produce json
// @Success 200 {object} ListPromptsResponse
// @Failure 500 {object} VerifierErrorResponse
// @Router /v1/llmops/prompts [get]
func (h *LLMOpsHandler) ListPrompts(c *gin.Context) {

	if !h.checkSystem(c) {
		return
	}
	registry := h.system.GetPromptRegistry()
	if registry == nil {
		c.JSON(
			http.StatusInternalServerError,
			VerifierErrorResponse{Error: "prompt registry not initialized"},
		)
		return
	}

	prompts, err := registry.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, VerifierErrorResponse{Error: err.Error()})
		return
	}

	resp := ListPromptsResponse{
		Prompts: make([]PromptVersionResponse, len(prompts)),
		Total:   len(prompts),
	}
	for i, p := range prompts {
		resp.Prompts[i] = promptToResponse(p)
	}

	c.JSON(http.StatusOK, resp)
}

// CreatePromptVersion godoc
// @Summary Create a prompt version
// @Description Create a new versioned prompt template
// @Tags llmops
// @Accept json
// @Produce json
// @Param request body CreatePromptVersionRequest true "Prompt version details"
// @Success 201 {object} PromptVersionResponse
// @Failure 400 {object} VerifierErrorResponse
// @Failure 500 {object} VerifierErrorResponse
// @Router /v1/llmops/prompts [post]
func (h *LLMOpsHandler) CreatePromptVersion(c *gin.Context) {

	if !h.checkSystem(c) {
		return
	}
	var req CreatePromptVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, VerifierErrorResponse{Error: err.Error()})
		return
	}

	registry := h.system.GetPromptRegistry()
	if registry == nil {
		c.JSON(
			http.StatusInternalServerError,
			VerifierErrorResponse{Error: "prompt registry not initialized"},
		)
		return
	}

	prompt := &llmops.PromptVersion{
		Name:        req.Name,
		Version:     req.Version,
		Content:     req.Content,
		Variables:   req.Variables,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		Author:      req.Author,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := registry.Create(c.Request.Context(), prompt); err != nil {
		c.JSON(http.StatusBadRequest, VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, promptToResponse(prompt))
}

// --- helpers ---

func experimentToResponse(exp *llmops.Experiment) ExperimentResponse {
	return ExperimentResponse{
		ID:           exp.ID,
		Name:         exp.Name,
		Description:  exp.Description,
		Variants:     exp.Variants,
		TrafficSplit: exp.TrafficSplit,
		Status:       exp.Status,
		Metrics:      exp.Metrics,
		TargetMetric: exp.TargetMetric,
		Winner:       exp.Winner,
		Metadata:     exp.Metadata,
		CreatedAt:    exp.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    exp.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func evaluationRunToResponse(run *llmops.EvaluationRun) EvaluationRunResponse {
	return EvaluationRunResponse{
		ID:            run.ID,
		Name:          run.Name,
		Description:   run.Description,
		Dataset:       run.Dataset,
		PromptName:    run.PromptName,
		PromptVersion: run.PromptVersion,
		ModelName:     run.ModelName,
		Metrics:       run.Metrics,
		Status:        run.Status,
		Results:       run.Results,
		CreatedAt:     run.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func promptToResponse(p *llmops.PromptVersion) PromptVersionResponse {
	return PromptVersionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Version:     p.Version,
		Content:     p.Content,
		Variables:   p.Variables,
		Tags:        p.Tags,
		Author:      p.Author,
		Description: p.Description,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// RegisterLLMOpsRoutes registers LLMOps routes
func RegisterLLMOpsRoutes(r *gin.RouterGroup, h *LLMOpsHandler) {
	ops := r.Group("/llmops")
	{
		ops.POST("/experiments", h.CreateExperiment)
		ops.GET("/experiments", h.ListExperiments)
		ops.GET("/experiments/:id", h.GetExperiment)
		ops.POST("/evaluate", h.CreateEvaluation)
		ops.GET("/prompts", h.ListPrompts)
		ops.POST("/prompts", h.CreatePromptVersion)
	}
}
