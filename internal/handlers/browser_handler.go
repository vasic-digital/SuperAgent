// Package handlers provides HTTP handlers
package handlers

import (
	"context"
	"net/http"
	"time"

	"dev.helix.agent/internal/browser"
	"github.com/gin-gonic/gin"
)

// BrowserHandler handles browser automation HTTP requests
type BrowserHandler struct {
	manager *browser.Manager
}

// NewBrowserHandler creates a new browser handler
func NewBrowserHandler(manager *browser.Manager) *BrowserHandler {
	return &BrowserHandler{manager: manager}
}

// NavigateRequest represents a navigation request
type NavigateRequest struct {
	URL     string `json:"url" binding:"required"`
	WaitFor string `json:"wait_for,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// NavigateResponse represents navigation result
type NavigateResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	Title   string `json:"title"`
}

// ClickRequest represents a click action
type ClickRequest struct {
	Selector string `json:"selector" binding:"required"`
	Button   string `json:"button,omitempty"`
}

// TypeRequest represents a type action
type TypeRequest struct {
	Selector string `json:"selector" binding:"required"`
	Text     string `json:"text" binding:"required"`
	Clear    bool   `json:"clear,omitempty"`
}

// ScreenshotRequest represents a screenshot request
type ScreenshotRequest struct {
	Selector string `json:"selector,omitempty"`
	FullPage bool   `json:"full_page,omitempty"`
}

// ScreenshotResponse represents screenshot result
type ScreenshotResponse struct {
	Data     []byte `json:"data"`
	MimeType string `json:"mime_type"`
}

// ExtractRequest represents content extraction request
type ExtractRequest struct {
	Selector  string `json:"selector" binding:"required"`
	Type      string `json:"type,omitempty"` // "text", "html", "attribute"
	Attribute string `json:"attribute,omitempty"`
}

// ExtractResponse represents extraction result
type ExtractResponse struct {
	Content  string `json:"content"`
	Selector string `json:"selector"`
}

// EvaluateRequest represents JavaScript evaluation request
type EvaluateRequest struct {
	Script string `json:"script" binding:"required"`
}

// EvaluateResponse represents evaluation result
type EvaluateResponse struct {
	Result interface{} `json:"result"`
}

// Navigate navigates to a URL
func (h *BrowserHandler) Navigate(c *gin.Context) {
	var req NavigateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeout := 30 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	actions := []browser.Action{
		&browser.NavigationAction{URL: req.URL},
	}

	if req.WaitFor != "" {
		actions = append(actions, &browser.WaitAction{
			Type:     "selector",
			Selector: req.WaitFor,
			Timeout:  timeout,
		})
	}

	result, err := h.manager.Execute(ctx, actions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, NavigateResponse{
		Success: result.Success,
		URL:     result.URL,
		Title:   result.Title,
	})
}

// Click clicks an element
func (h *BrowserHandler) Click(c *gin.Context) {
	var req ClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	actions := []browser.Action{
		&browser.ClickAction{Selector: req.Selector, Button: req.Button},
	}

	_, err := h.manager.Execute(ctx, actions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Type types text into an element
func (h *BrowserHandler) Type(c *gin.Context) {
	var req TypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	actions := []browser.Action{
		&browser.TypeAction{Selector: req.Selector, Text: req.Text, Clear: req.Clear},
	}

	_, err := h.manager.Execute(ctx, actions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Screenshot captures a screenshot
func (h *BrowserHandler) Screenshot(c *gin.Context) {
	var req ScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	actions := []browser.Action{
		&browser.ScreenshotAction{Selector: req.Selector, FullPage: req.FullPage},
	}

	result, err := h.manager.Execute(ctx, actions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.Screenshot == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "screenshot failed"})
		return
	}

	c.Data(http.StatusOK, "image/png", result.Screenshot.Data)
}

// Extract extracts content from the page
func (h *BrowserHandler) Extract(c *gin.Context) {
	var req ExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	actions := []browser.Action{
		&browser.ExtractAction{Selector: req.Selector, Type: req.Type},
	}

	result, err := h.manager.Execute(ctx, actions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.Extracted == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "extraction failed"})
		return
	}

	c.JSON(http.StatusOK, ExtractResponse{
		Content:  result.Extracted.Content,
		Selector: result.Extracted.Selector,
	})
}

// Evaluate executes JavaScript
func (h *BrowserHandler) Evaluate(c *gin.Context) {
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	actions := []browser.Action{
		&browser.EvaluateAction{Script: req.Script},
	}

	result, err := h.manager.Execute(ctx, actions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, EvaluateResponse{
		Result: result.Evaluated.Result,
	})
}

// RegisterRoutes registers the browser routes
func (h *BrowserHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/v1/browser")
	{
		v1.POST("/navigate", h.Navigate)
		v1.POST("/click", h.Click)
		v1.POST("/type", h.Type)
		v1.POST("/screenshot", h.Screenshot)
		v1.POST("/extract", h.Extract)
		v1.POST("/evaluate", h.Evaluate)
	}
}
