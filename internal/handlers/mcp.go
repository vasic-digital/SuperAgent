package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/mcp/adapters"
	"dev.helix.agent/internal/services"
	"dev.helix.agent/internal/tools"
)

// MCPHandler handles Model Context Protocol (MCP) endpoints
type MCPHandler struct {
	config           *config.MCPConfig
	providerRegistry *services.ProviderRegistry
	mcpManager       *services.MCPManager
	logger           *logrus.Logger
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(providerRegistry *services.ProviderRegistry, cfg *config.MCPConfig) *MCPHandler {
	return &MCPHandler{
		config:           cfg,
		providerRegistry: providerRegistry,
		mcpManager:       services.NewMCPManager(nil, nil, logrus.New()),
		logger:           logrus.New(),
	}
}

// MCPCapabilities returns MCP server capabilities
func (h *MCPHandler) MCPCapabilities(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	capabilities := map[string]interface{}{
		"version": "1.0.0",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
			"prompts": map[string]interface{}{
				"listChanged": true,
			},
			"resources": map[string]interface{}{
				"listChanged": true,
			},
		},
		"providers":   []string{},
		"mcp_servers": []string{},
	}

	if h.providerRegistry != nil {
		providers := h.providerRegistry.ListProviders()
		capabilities["providers"] = providers
	}

	if h.mcpManager != nil {
		servers, err := h.mcpManager.ListMCPServers(c.Request.Context())
		if err == nil {
			serverNames := make([]string, len(servers))
			for i, server := range servers {
				serverNames[i] = server.Name
			}
			capabilities["mcp_servers"] = serverNames
		}
	}

	c.JSON(http.StatusOK, capabilities)
}

// MCPTools returns available MCP tools
func (h *MCPHandler) MCPTools(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	tools := []map[string]interface{}{}
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

// MCPToolsCall executes an MCP tool
func (h *MCPHandler) MCPToolsCall(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if h.providerRegistry == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Provider registry not available",
		})
		return
	}

	// Process tool call
	name, ok := req["name"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tool name is required",
		})
		return
	}

	arguments, _ := req["arguments"].(map[string]interface{})

	// Handle unified namespace if enabled
	if h.config.UnifiedToolNamespace {
		if idx := findUnderscoreIndex(name); idx > 0 {
			providerName := name[:idx]
			toolName := name[idx+1:]

			h.logger.WithFields(logrus.Fields{
				"provider": providerName,
				"tool":     toolName,
			}).Info("Executing tool with unified namespace")

			// Try to execute tool on specific provider
			// This is a simplified implementation
			c.JSON(http.StatusOK, gin.H{
				"result":    fmt.Sprintf("Tool %s executed on provider %s", toolName, providerName),
				"arguments": arguments,
			})
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Invalid tool format or provider not found",
	})
}

// MCPPrompts returns available MCP prompts
func (h *MCPHandler) MCPPrompts(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	prompts := []map[string]interface{}{
		{
			"name":        "summarize",
			"description": "Summarize text content",
			"arguments": []map[string]interface{}{
				{
					"name":        "text",
					"description": "Text to summarize",
					"required":    true,
				},
			},
		},
		{
			"name":        "analyze",
			"description": "Analyze content for insights",
			"arguments": []map[string]interface{}{
				{
					"name":        "content",
					"description": "Content to analyze",
					"required":    true,
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"prompts": prompts})
}

// MCPResources returns available MCP resources
func (h *MCPHandler) MCPResources(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	resources := []map[string]interface{}{
		{
			"uri":         "helixagent://providers",
			"name":        "Provider Information",
			"description": "Information about configured LLM providers",
			"mimeType":    "application/json",
		},
		{
			"uri":         "helixagent://models",
			"name":        "Model Metadata",
			"description": "Metadata about available LLM models",
			"mimeType":    "application/json",
		},
	}

	c.JSON(http.StatusOK, gin.H{"resources": resources})
}

// GetMCPManager returns the MCP manager instance
func (h *MCPHandler) GetMCPManager() *services.MCPManager {
	return h.mcpManager
}

// RegisterMCPServer registers an MCP server
func (h *MCPHandler) RegisterMCPServer(config map[string]interface{}) error {
	if h.mcpManager == nil {
		return fmt.Errorf("MCP manager not initialized")
	}

	// This is a simplified implementation
	// In a real implementation, this would connect to the actual MCP server
	h.logger.WithField("config", config).Info("Registering MCP server")

	return nil
}

// findUnderscoreIndex finds the first underscore in a string (not at start or end)
func findUnderscoreIndex(s string) int {
	for i, char := range s {
		if char == '_' && i > 0 && i < len(s)-1 {
			return i
		}
	}
	return -1
}

// MCPToolSearchRequest represents a tool search request
type MCPToolSearchRequest struct {
	Query         string   `json:"query" binding:"required"`
	Categories    []string `json:"categories,omitempty"`
	IncludeParams bool     `json:"include_params,omitempty"`
	FuzzyMatch    bool     `json:"fuzzy_match,omitempty"`
	MaxResults    int      `json:"max_results,omitempty"`
}

// MCPToolSearch searches for tools by query
func (h *MCPHandler) MCPToolSearch(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	// Support both GET with query params and POST with JSON body
	var req MCPToolSearchRequest
	if c.Request.Method == "GET" {
		req.Query = c.Query("q")
		if req.Query == "" {
			req.Query = c.Query("query")
		}
		if categories := c.Query("categories"); categories != "" {
			req.Categories = strings.Split(categories, ",")
		}
		req.IncludeParams = c.Query("include_params") == "true"
		req.FuzzyMatch = c.Query("fuzzy") == "true"
		if maxStr := c.Query("max_results"); maxStr != "" {
			if max, err := strconv.Atoi(maxStr); err == nil {
				req.MaxResults = max
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter is required",
		})
		return
	}

	// Search tools
	opts := tools.SearchOptions{
		Query:         req.Query,
		Categories:    req.Categories,
		IncludeParams: req.IncludeParams,
		FuzzyMatch:    req.FuzzyMatch,
		MaxResults:    req.MaxResults,
	}

	results := tools.SearchTools(opts)

	// Format response
	toolResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		toolResults[i] = map[string]interface{}{
			"name":        result.Tool.Name,
			"description": result.Tool.Description,
			"category":    result.Tool.Category,
			"score":       result.Score,
			"match_type":  result.MatchType,
			"parameters":  result.Tool.Parameters,
			"required":    result.Tool.RequiredFields,
			"aliases":     result.Tool.Aliases,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"count":   len(results),
		"results": toolResults,
	})
}

// MCPAdapterSearchRequest represents an adapter search request
type MCPAdapterSearchRequest struct {
	Query      string   `json:"query"`
	Categories []string `json:"categories,omitempty"`
	AuthTypes  []string `json:"auth_types,omitempty"`
	Official   *bool    `json:"official,omitempty"`
	Supported  *bool    `json:"supported,omitempty"`
	MaxResults int      `json:"max_results,omitempty"`
}

// MCPAdapterSearch searches for MCP adapters
func (h *MCPHandler) MCPAdapterSearch(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	// Support both GET with query params and POST with JSON body
	var req MCPAdapterSearchRequest
	if c.Request.Method == "GET" {
		req.Query = c.Query("q")
		if req.Query == "" {
			req.Query = c.Query("query")
		}
		if categories := c.Query("categories"); categories != "" {
			req.Categories = strings.Split(categories, ",")
		}
		if authTypes := c.Query("auth_types"); authTypes != "" {
			req.AuthTypes = strings.Split(authTypes, ",")
		}
		if official := c.Query("official"); official != "" {
			val := official == "true"
			req.Official = &val
		}
		if supported := c.Query("supported"); supported != "" {
			val := supported == "true"
			req.Supported = &val
		}
		if maxStr := c.Query("max_results"); maxStr != "" {
			if max, err := strconv.Atoi(maxStr); err == nil {
				req.MaxResults = max
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}
	}

	// Convert category strings to AdapterCategory
	var categories []adapters.AdapterCategory
	for _, cat := range req.Categories {
		categories = append(categories, adapters.AdapterCategory(cat))
	}

	// Search adapters
	opts := adapters.AdapterSearchOptions{
		Query:      req.Query,
		Categories: categories,
		AuthTypes:  req.AuthTypes,
		Official:   req.Official,
		Supported:  req.Supported,
		MaxResults: req.MaxResults,
	}

	results := adapters.DefaultRegistry.Search(opts)

	// Format response
	adapterResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		adapterResults[i] = map[string]interface{}{
			"name":        result.Adapter.Name,
			"description": result.Adapter.Description,
			"category":    result.Adapter.Category,
			"auth_type":   result.Adapter.AuthType,
			"official":    result.Adapter.Official,
			"supported":   result.Adapter.Supported,
			"docs_url":    result.Adapter.DocsURL,
			"score":       result.Score,
			"match_type":  result.MatchType,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"count":   len(results),
		"results": adapterResults,
	})
}

// MCPToolSuggestions returns tool suggestions for autocomplete
func (h *MCPHandler) MCPToolSuggestions(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	prefix := c.Query("prefix")
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Prefix parameter is required",
		})
		return
	}

	maxSuggestions := 10
	if maxStr := c.Query("max"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil {
			maxSuggestions = max
		}
	}

	suggestions := tools.GetToolSuggestions(prefix, maxSuggestions)

	result := make([]map[string]interface{}, len(suggestions))
	for i, tool := range suggestions {
		result[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"category":    tool.Category,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"prefix":      prefix,
		"count":       len(suggestions),
		"suggestions": result,
	})
}

// MCPCategories returns available tool categories
func (h *MCPHandler) MCPCategories(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	// Tool categories
	toolCategories := []string{
		tools.CategoryCore,
		tools.CategoryFileSystem,
		tools.CategoryVersionControl,
		tools.CategoryCodeIntel,
		tools.CategoryWorkflow,
		tools.CategoryWeb,
	}

	// Adapter categories
	adapterCategories := adapters.GetAllCategories()
	adapterCatStrings := make([]string, len(adapterCategories))
	for i, cat := range adapterCategories {
		adapterCatStrings[i] = string(cat)
	}

	c.JSON(http.StatusOK, gin.H{
		"tool_categories":    toolCategories,
		"adapter_categories": adapterCatStrings,
		"auth_types":         adapters.GetAllAuthTypes(),
	})
}

// MCPStats returns MCP statistics
func (h *MCPHandler) MCPStats(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP is not enabled",
		})
		return
	}

	// Count tools by category
	toolsByCategory := make(map[string]int)
	for _, tool := range tools.ToolSchemaRegistry {
		toolsByCategory[tool.Category]++
	}

	// Count adapters by category
	adaptersByCategory := make(map[string]int)
	for _, adapter := range adapters.AvailableAdapters {
		adaptersByCategory[string(adapter.Category)]++
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": map[string]interface{}{
			"total":       len(tools.ToolSchemaRegistry),
			"by_category": toolsByCategory,
		},
		"adapters": map[string]interface{}{
			"total":       len(adapters.AvailableAdapters),
			"by_category": adaptersByCategory,
			"official":    len(adapters.GetOfficialAdapters()),
			"supported":   len(adapters.GetSupportedAdapters()),
		},
	})
}
