package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	helixqaadapter "dev.helix.agent/internal/adapters/helixqa"
)

// QAHandler provides HTTP endpoints for HelixQA autonomous QA
// operations.
// Endpoints:
//
//	POST /v1/qa/sessions          - start autonomous QA session
//	GET  /v1/qa/findings          - list findings by status
//	GET  /v1/qa/findings/:id      - get specific finding
//	PUT  /v1/qa/findings/:id      - update finding status
//	GET  /v1/qa/platforms         - list supported platforms
//	POST /v1/qa/discover          - discover project knowledge
type QAHandler struct {
	adapter *helixqaadapter.Adapter
}

// NewQAHandler creates a new QA handler with the given adapter.
// If adapter is nil, all endpoints return 503.
func NewQAHandler(adapter *helixqaadapter.Adapter) *QAHandler {
	return &QAHandler{adapter: adapter}
}

// checkAdapter returns false and writes a 503 response if the
// adapter has not been initialised.
func (h *QAHandler) checkAdapter(c *gin.Context) bool {
	if h.adapter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "HelixQA adapter not initialized",
		})
		return false
	}
	return true
}

// --- request / response types ---

// StartQASessionRequest represents a request to start an
// autonomous QA session.
type StartQASessionRequest struct {
	ProjectRoot      string   `json:"project_root" binding:"required"`
	Platforms        []string `json:"platforms" binding:"required"`
	OutputDir        string   `json:"output_dir,omitempty"`
	IssuesDir        string   `json:"issues_dir,omitempty"`
	AndroidDevice    string   `json:"android_device,omitempty"`
	AndroidDevices   []string `json:"android_devices,omitempty"`
	AndroidPackage   string   `json:"android_package,omitempty"`
	WebURL           string   `json:"web_url,omitempty"`
	CuriosityEnabled bool     `json:"curiosity_enabled,omitempty"`
	VisionHost       string   `json:"vision_host,omitempty"`
	VisionUser       string   `json:"vision_user,omitempty"`
	VisionModel      string   `json:"vision_model,omitempty"`
	MemoryDBPath     string   `json:"memory_db_path,omitempty"`
}

// UpdateFindingRequest represents a request to update a
// finding's status.
type UpdateFindingRequest struct {
	Status string `json:"status" binding:"required"`
}

// DiscoverKnowledgeRequest represents a request to discover
// project knowledge.
type DiscoverKnowledgeRequest struct {
	ProjectRoot string `json:"project_root" binding:"required"`
}

// KnowledgeSummary holds a summary of discovered project
// knowledge.
type KnowledgeSummary struct {
	DocsCount        int               `json:"docs_count"`
	ConstraintsCount int               `json:"constraints_count"`
	Credentials      map[string]string `json:"credentials,omitempty"`
}

// --- handlers ---

// StartSession starts an autonomous QA session.
// @Summary Start autonomous QA session
// @Description Launch a full autonomous QA pipeline
// @Tags qa
// @Accept json
// @Produce json
// @Param request body StartQASessionRequest true "Session config"
// @Success 202 {object} helixqaadapter.SessionResult
// @Failure 400 {object} VerifierErrorResponse
// @Failure 503 {object} VerifierErrorResponse
// @Router /v1/qa/sessions [post]
func (h *QAHandler) StartSession(c *gin.Context) {
	if !h.checkAdapter(c) {
		return
	}

	var req StartQASessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	cfg := &helixqaadapter.SessionConfig{
		ProjectRoot:      req.ProjectRoot,
		Platforms:        req.Platforms,
		OutputDir:        req.OutputDir,
		IssuesDir:        req.IssuesDir,
		AndroidDevice:    req.AndroidDevice,
		AndroidDevices:   req.AndroidDevices,
		AndroidPackage:   req.AndroidPackage,
		WebURL:           req.WebURL,
		CuriosityEnabled: req.CuriosityEnabled,
		VisionHost:       req.VisionHost,
		VisionUser:       req.VisionUser,
		VisionModel:      req.VisionModel,
		MemoryDBPath:     req.MemoryDBPath,
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = "qa-results"
	}
	if cfg.IssuesDir == "" {
		cfg.IssuesDir = "qa-issues"
	}

	result, err := h.adapter.RunAutonomousSession(
		c.Request.Context(), cfg)
	if err != nil {
		// Return partial result with error info.
		if result != nil {
			c.JSON(http.StatusAccepted, result)
			return
		}
		c.JSON(http.StatusInternalServerError,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, result)
}

// ListFindings returns findings filtered by status.
// @Summary List QA findings
// @Description List findings from autonomous QA sessions
// @Tags qa
// @Produce json
// @Param status query string false "Filter by status"
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} VerifierErrorResponse
// @Router /v1/qa/findings [get]
func (h *QAHandler) ListFindings(c *gin.Context) {
	if !h.checkAdapter(c) {
		return
	}

	status := c.Query("status")
	findings, err := h.adapter.GetFindings(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"findings": findings,
		"total":    len(findings),
	})
}

// GetFinding returns a specific finding by ID.
// @Summary Get specific QA finding
// @Description Get details of a specific finding by ID
// @Tags qa
// @Produce json
// @Param id path string true "Finding ID"
// @Success 200 {object} helixqaadapter.Finding
// @Failure 404 {object} VerifierErrorResponse
// @Failure 503 {object} VerifierErrorResponse
// @Router /v1/qa/findings/{id} [get]
func (h *QAHandler) GetFinding(c *gin.Context) {
	if !h.checkAdapter(c) {
		return
	}

	id := c.Param("id")
	finding, err := h.adapter.GetFinding(id)
	if err != nil {
		c.JSON(http.StatusNotFound,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, finding)
}

// UpdateFinding updates a finding's status.
// @Summary Update QA finding status
// @Description Update the status of a finding
// @Tags qa
// @Accept json
// @Produce json
// @Param id path string true "Finding ID"
// @Param request body UpdateFindingRequest true "Status update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} VerifierErrorResponse
// @Failure 503 {object} VerifierErrorResponse
// @Router /v1/qa/findings/{id} [put]
func (h *QAHandler) UpdateFinding(c *gin.Context) {
	if !h.checkAdapter(c) {
		return
	}

	id := c.Param("id")
	var req UpdateFindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	if err := h.adapter.UpdateFindingStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     id,
		"status": req.Status,
	})
}

// ListPlatforms returns supported QA platforms.
// @Summary List supported QA platforms
// @Description Get the list of platforms supported by HelixQA
// @Tags qa
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} VerifierErrorResponse
// @Router /v1/qa/platforms [get]
func (h *QAHandler) ListPlatforms(c *gin.Context) {
	if !h.checkAdapter(c) {
		return
	}

	platforms := h.adapter.SupportedPlatforms()
	c.JSON(http.StatusOK, gin.H{
		"platforms": platforms,
		"total":     len(platforms),
	})
}

// DiscoverKnowledge discovers project knowledge from docs
// and configuration files.
// @Summary Discover project knowledge
// @Description Scan project for docs, constraints, credentials
// @Tags qa
// @Accept json
// @Produce json
// @Param request body DiscoverKnowledgeRequest true "Project root"
// @Success 200 {object} KnowledgeSummary
// @Failure 400 {object} VerifierErrorResponse
// @Failure 503 {object} VerifierErrorResponse
// @Router /v1/qa/discover [post]
func (h *QAHandler) DiscoverKnowledge(c *gin.Context) {
	if !h.checkAdapter(c) {
		return
	}

	var req DiscoverKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	kb, err := h.adapter.DiscoverKnowledge(req.ProjectRoot)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			VerifierErrorResponse{Error: err.Error()})
		return
	}

	// Redact credential values for API response.
	redactedCreds := make(map[string]string, len(kb.Credentials))
	for k := range kb.Credentials {
		redactedCreds[k] = "***"
	}

	c.JSON(http.StatusOK, KnowledgeSummary{
		DocsCount:        len(kb.Docs),
		ConstraintsCount: len(kb.Constraints),
		Credentials:      redactedCreds,
	})
}

// RegisterQARoutes registers QA routes on the given router
// group.
func RegisterQARoutes(r *gin.RouterGroup, h *QAHandler) {
	qa := r.Group("/qa")
	{
		qa.POST("/sessions", h.StartSession)
		qa.GET("/findings", h.ListFindings)
		qa.GET("/findings/:id", h.GetFinding)
		qa.PUT("/findings/:id", h.UpdateFinding)
		qa.GET("/platforms", h.ListPlatforms)
		qa.POST("/discover", h.DiscoverKnowledge)
	}
}
