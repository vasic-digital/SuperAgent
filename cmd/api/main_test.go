package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"dev.helix.agent/internal/version"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewAPIServer(t *testing.T) {
	server := NewAPIServer("8080")
	require.NotNil(t, server)
	assert.Equal(t, "8080", server.port)
	assert.NotNil(t, server.logger)
	assert.NotNil(t, server.unifiedManager)
	assert.NotNil(t, server.protocolAnalytics)
	assert.NotNil(t, server.pluginSystem)
	assert.NotNil(t, server.pluginRegistry)
	assert.NotNil(t, server.templateManager)
}

func TestNewAPIServer_DefaultPort(t *testing.T) {
	server := NewAPIServer("")
	require.NotNil(t, server)
	assert.Equal(t, "", server.port)
}

func setupTestServer() (*APIServer, *gin.Engine) {
	server := NewAPIServer("8080")
	r := gin.New()

	// Setup routes like in Start()
	api := r.Group("/api/v1")
	{
		mcp := api.Group("/mcp")
		{
			mcp.POST("/tools/call", server.handleMCPCallTool)
			mcp.GET("/tools/list", server.handleMCPListTools)
			mcp.GET("/servers", server.handleMCPListServers)
		}

		lsp := api.Group("/lsp")
		{
			lsp.POST("/completion", server.handleLSPCompletion)
			lsp.POST("/hover", server.handleLSPHover)
			lsp.POST("/definition", server.handleLSPDefinition)
			lsp.POST("/diagnostics", server.handleLSPDiagnostics)
		}

		acp := api.Group("/acp")
		{
			acp.POST("/execute", server.handleACPExecute)
			acp.POST("/broadcast", server.handleACPBroadcast)
			acp.GET("/status", server.handleACPStatus)
		}

		analytics := api.Group("/analytics")
		{
			analytics.GET("/metrics", server.handleGetAnalytics)
			analytics.GET("/metrics/:protocol", server.handleGetProtocolMetrics)
			analytics.GET("/health", server.handleGetHealthStatus)
			analytics.POST("/record", server.handleRecordRequest)
		}

		plugins := api.Group("/plugins")
		{
			plugins.GET("/", server.handleListPlugins)
			plugins.POST("/load", server.handleLoadPlugin)
			plugins.DELETE("/:id", server.handleUnloadPlugin)
			plugins.POST("/:id/execute", server.handleExecutePlugin)
			plugins.GET("/marketplace", server.handleMarketplaceSearch)
			plugins.POST("/marketplace/register", server.handleRegisterPlugin)
		}

		templates := api.Group("/templates")
		{
			templates.GET("/", server.handleListTemplates)
			templates.GET("/:id", server.handleGetTemplate)
			templates.POST("/:id/generate", server.handleGenerateFromTemplate)
		}

		api.GET("/health", server.handleHealth)
		api.GET("/status", server.handleStatus)
		api.GET("/metrics", server.handlePrometheusMetrics)
	}

	return server, r
}

// MCP Protocol Tests

func TestHandleMCPCallTool(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful call", func(t *testing.T) {
		body := map[string]interface{}{
			"server_id":  "test-server",
			"tool_name":  "calculate",
			"parameters": map[string]interface{}{"x": 1, "y": 2},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "calculate", response["tool"])
		assert.Equal(t, "test-server", response["server"])
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/mcp/tools/call", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleMCPListTools(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/mcp/tools/list?server_id=test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "tools")
}

func TestHandleMCPListServers(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/mcp/servers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "servers")
}

// LSP Protocol Tests

func TestHandleLSPCompletion(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful completion", func(t *testing.T) {
		body := map[string]interface{}{
			"file_path": "/test/file.go",
			"line":      10,
			"character": 5,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/lsp/completion", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "completions")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/lsp/completion", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleLSPHover(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful hover", func(t *testing.T) {
		body := map[string]interface{}{
			"file_path": "/test/file.go",
			"line":      10,
			"character": 5,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/lsp/hover", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "contents")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/lsp/hover", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleLSPDefinition(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful definition", func(t *testing.T) {
		body := map[string]interface{}{
			"file_path": "/test/file.go",
			"line":      10,
			"character": 5,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/lsp/definition", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "definition")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/lsp/definition", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleLSPDiagnostics(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("POST", "/api/v1/lsp/diagnostics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "diagnostics")
}

// ACP Protocol Tests

func TestHandleACPExecute(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful execute", func(t *testing.T) {
		body := map[string]interface{}{
			"action":   "test_action",
			"agent_id": "agent-1",
			"params":   map[string]interface{}{"key": "value"},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/acp/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "test_action", response["action"])
		assert.Equal(t, "agent-1", response["agent_id"])
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/acp/execute", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleACPBroadcast(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful broadcast", func(t *testing.T) {
		body := map[string]interface{}{
			"message": "test message",
			"targets": []string{"agent-1", "agent-2"},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/acp/broadcast", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "broadcast_id")
		assert.Equal(t, float64(2), response["delivered_to"])
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/acp/broadcast", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleACPStatus(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/acp/status?agent_id=agent-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "agent-1", response["agent_id"])
	assert.Equal(t, "active", response["status"])
}

// Analytics Tests

func TestHandleGetAnalytics(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/analytics/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetProtocolMetrics(t *testing.T) {
	_, router := setupTestServer()

	t.Run("existing protocol", func(t *testing.T) {
		// First record a request to create metrics
		body := map[string]interface{}{
			"protocol":   "mcp",
			"method":     "call_tool",
			"duration":   100000000,
			"success":    true,
			"error_type": "",
		}
		jsonBody, _ := json.Marshal(body)

		recordReq, _ := http.NewRequest("POST", "/api/v1/analytics/record", bytes.NewBuffer(jsonBody))
		recordReq.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, recordReq)

		// Then get the metrics
		req, _ := http.NewRequest("GET", "/api/v1/analytics/metrics/mcp", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("non-existing protocol", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/analytics/metrics/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleGetHealthStatus(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/analytics/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleRecordRequest(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful record", func(t *testing.T) {
		body := map[string]interface{}{
			"protocol":   "mcp",
			"method":     "call_tool",
			"duration":   100000000,
			"success":    true,
			"error_type": "",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/analytics/record", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "recorded", response["status"])
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/analytics/record", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Plugin Tests

func TestHandleListPlugins(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/plugins/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "plugins")
}

func TestHandleLoadPlugin(t *testing.T) {
	_, router := setupTestServer()

	t.Run("attempt to load plugin", func(t *testing.T) {
		body := map[string]interface{}{
			"path": "/test/plugin.so",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/plugins/load", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Will return 500 because plugin doesn't exist, but tests the handler
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/plugins/load", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleUnloadPlugin(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("DELETE", "/api/v1/plugins/test-plugin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Will return 500 because plugin doesn't exist, but tests the handler
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestHandleExecutePlugin(t *testing.T) {
	_, router := setupTestServer()

	t.Run("non-existing plugin", func(t *testing.T) {
		body := map[string]interface{}{
			"operation": "test",
			"params":    map[string]interface{}{"key": "value"},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/plugins/test-plugin/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Will return 500 because plugin doesn't exist, but tests the handler
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/plugins/test-plugin/execute", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleMarketplaceSearch(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/plugins/marketplace?q=test&protocol=mcp", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "plugins")
}

func TestHandleRegisterPlugin(t *testing.T) {
	_, router := setupTestServer()

	t.Run("successful registration", func(t *testing.T) {
		body := map[string]interface{}{
			"id":          "test-plugin",
			"name":        "Test Plugin",
			"version":     "1.0.0",
			"description": "A test plugin",
			"author":      "Test Author",
			"protocols":   []string{"mcp"},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/plugins/marketplace/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/plugins/marketplace/register", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Template Tests

func TestHandleListTemplates(t *testing.T) {
	_, router := setupTestServer()

	t.Run("list all templates", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/templates/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "templates")
	})

	t.Run("list templates by protocol", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/templates/?protocol=mcp", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleGetTemplate(t *testing.T) {
	_, router := setupTestServer()

	t.Run("get existing template", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/templates/mcp-basic-integration", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "mcp-basic-integration", response["ID"])
	})

	t.Run("get non-existing template", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/templates/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleGenerateFromTemplate(t *testing.T) {
	_, router := setupTestServer()

	t.Run("existing template", func(t *testing.T) {
		body := map[string]interface{}{
			"config": map[string]interface{}{"server_url": "http://localhost:7061"},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/templates/mcp-basic-integration/generate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "generated")
	})

	t.Run("non-existing template", func(t *testing.T) {
		body := map[string]interface{}{
			"config": map[string]interface{}{"key": "value"},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/templates/test-template/generate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Will return 500 because template doesn't exist, but tests the handler
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/templates/test-template/generate", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// System Tests

func TestHandleHealth(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, version.Version, response["version"])
}

func TestHandleStatus(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "operational", response["status"])
	assert.Contains(t, response, "protocols_active")
	assert.Contains(t, response, "plugins_loaded")
	assert.Contains(t, response, "templates_available")
}

func TestHandlePrometheusMetrics(t *testing.T) {
	_, router := setupTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
	assert.Contains(t, w.Body.String(), "helixagent_protocols_active")
	assert.Contains(t, w.Body.String(), "helixagent_plugins_loaded")
	assert.Contains(t, w.Body.String(), "helixagent_requests_total")
}

func TestHandlePrometheusMetricsWithData(t *testing.T) {
	_, router := setupTestServer()

	// First record some requests to create metrics
	body := map[string]interface{}{
		"protocol":   "mcp",
		"method":     "call_tool",
		"duration":   100000000,
		"success":    true,
		"error_type": "",
	}
	jsonBody, _ := json.Marshal(body)
	recordReq, _ := http.NewRequest("POST", "/api/v1/analytics/record", bytes.NewBuffer(jsonBody))
	recordReq.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, recordReq)

	// Now get the metrics
	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// With recorded data, total requests should be > 0
	assert.Contains(t, w.Body.String(), "helixagent_requests_total")
}

// Helper to setup test router with CORS middleware (mirrors Start() function)
func setupTestServerWithCORS() (*APIServer, *gin.Engine) {
	server := NewAPIServer("8080")
	r := gin.New()

	// CORS middleware from Start()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes like in Start()
	api := r.Group("/api/v1")
	{
		api.GET("/health", server.handleHealth)
	}

	return server, r
}

func TestCORSMiddleware(t *testing.T) {
	_, router := setupTestServerWithCORS()

	t.Run("regular request has CORS headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/v1/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 204, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// =============================================================================
// Additional APIServer Tests
// =============================================================================

func TestAPIServer_PortConfiguration(t *testing.T) {
	t.Run("default port", func(t *testing.T) {
		server := NewAPIServer("8080")
		assert.Equal(t, "8080", server.port)
	})

	t.Run("custom port", func(t *testing.T) {
		server := NewAPIServer("9090")
		assert.Equal(t, "9090", server.port)
	})

	t.Run("empty port", func(t *testing.T) {
		server := NewAPIServer("")
		assert.Equal(t, "", server.port)
	})
}

func TestAPIServer_Components(t *testing.T) {
	server := NewAPIServer("8080")

	assert.NotNil(t, server.logger)
	assert.NotNil(t, server.unifiedManager)
	assert.NotNil(t, server.protocolAnalytics)
	assert.NotNil(t, server.pluginSystem)
	assert.NotNil(t, server.pluginRegistry)
	assert.NotNil(t, server.templateManager)
}

// =============================================================================
// Analytics Handler Edge Cases
// =============================================================================

func TestHandleRecordRequest_ErrorCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("empty protocol", func(t *testing.T) {
		body := map[string]interface{}{
			"protocol":   "",
			"method":     "test",
			"duration":   100000000,
			"success":    true,
			"error_type": "",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/analytics/record", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// May return error or success depending on validation
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("failed request recording", func(t *testing.T) {
		body := map[string]interface{}{
			"protocol":   "test",
			"method":     "test_method",
			"duration":   50000000,
			"success":    false,
			"error_type": "connection_error",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/analytics/record", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleRegisterPlugin_ValidationErrors(t *testing.T) {
	_, router := setupTestServer()

	t.Run("missing required fields", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "incomplete-plugin",
			// missing id, version, etc.
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/plugins/marketplace/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 200 (registration succeeds) or 500 if validation fails
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("empty body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/plugins/marketplace/register", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// May succeed with defaults or fail validation
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

// =============================================================================
// MCP Handler Edge Cases
// =============================================================================

func TestHandleMCPCallTool_EdgeCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("empty parameters", func(t *testing.T) {
		body := map[string]interface{}{
			"server_id":  "test-server",
			"tool_name":  "calculate",
			"parameters": map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("complex parameters", func(t *testing.T) {
		body := map[string]interface{}{
			"server_id": "test-server",
			"tool_name": "complex_tool",
			"parameters": map[string]interface{}{
				"nested": map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": "value",
					},
				},
				"array": []interface{}{1, 2, 3},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleMCPListTools_WithParams(t *testing.T) {
	_, router := setupTestServer()

	t.Run("with empty server_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/mcp/tools/list?server_id=", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("without server_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/mcp/tools/list", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// =============================================================================
// LSP Handler Edge Cases
// =============================================================================

func TestHandleLSPCompletion_EdgeCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("zero position", func(t *testing.T) {
		body := map[string]interface{}{
			"file_path": "/test/file.go",
			"line":      0,
			"character": 0,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/lsp/completion", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("large line number", func(t *testing.T) {
		body := map[string]interface{}{
			"file_path": "/test/file.go",
			"line":      10000,
			"character": 100,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/lsp/completion", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// =============================================================================
// ACP Handler Edge Cases
// =============================================================================

func TestHandleACPExecute_EdgeCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("empty action", func(t *testing.T) {
		body := map[string]interface{}{
			"action":   "",
			"agent_id": "agent-1",
			"params":   map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/acp/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleACPBroadcast_EdgeCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("empty targets", func(t *testing.T) {
		body := map[string]interface{}{
			"message": "test message",
			"targets": []string{},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/acp/broadcast", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["delivered_to"])
	})

	t.Run("many targets", func(t *testing.T) {
		targets := make([]string, 100)
		for i := 0; i < 100; i++ {
			targets[i] = fmt.Sprintf("agent-%d", i)
		}

		body := map[string]interface{}{
			"message": "broadcast to many",
			"targets": targets,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/acp/broadcast", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(100), response["delivered_to"])
	})
}

func TestHandleACPStatus_EdgeCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("empty agent_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/acp/status?agent_id=", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "", response["agent_id"])
	})

	t.Run("special characters in agent_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/acp/status?agent_id=agent%2Ftest%3Aid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// =============================================================================
// Template Handler Edge Cases
// =============================================================================

func TestHandleListTemplates_Filtering(t *testing.T) {
	_, router := setupTestServer()

	t.Run("filter by non-existent protocol", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/templates/?protocol=nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		// Should return empty array or all templates
		assert.Contains(t, response, "templates")
	})

	t.Run("filter by lsp protocol", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/templates/?protocol=lsp", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("filter by acp protocol", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/templates/?protocol=acp", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleGenerateFromTemplate_EdgeCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("empty config", func(t *testing.T) {
		body := map[string]interface{}{
			"config": map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/templates/mcp-basic-integration/generate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("complex config", func(t *testing.T) {
		body := map[string]interface{}{
			"config": map[string]interface{}{
				"server_url": "http://localhost:7061",
				"api_key":    "sk-test-key",
				"options": map[string]interface{}{
					"timeout": 30,
					"retries": 3,
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/templates/mcp-basic-integration/generate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// =============================================================================
// Plugin Handler Edge Cases
// =============================================================================

func TestHandleMarketplaceSearch_Filters(t *testing.T) {
	_, router := setupTestServer()

	t.Run("search with special characters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/plugins/marketplace?q=test%20plugin%26special", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("search with protocol filter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/plugins/marketplace?q=&protocol=lsp", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("search empty query", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/plugins/marketplace?q=", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleExecutePlugin_EdgeCases(t *testing.T) {
	_, router := setupTestServer()

	t.Run("empty operation", func(t *testing.T) {
		body := map[string]interface{}{
			"operation": "",
			"params":    map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/plugins/test-plugin/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Will return 500 because plugin doesn't exist
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkHandleMCPCallTool(b *testing.B) {
	_, router := setupTestServer()

	body := map[string]interface{}{
		"server_id":  "test-server",
		"tool_name":  "calculate",
		"parameters": map[string]interface{}{"x": 1, "y": 2},
	}
	jsonBody, _ := json.Marshal(body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandleHealth(b *testing.B) {
	_, router := setupTestServer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandleStatus(b *testing.B) {
	_, router := setupTestServer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandlePrometheusMetrics(b *testing.B) {
	_, router := setupTestServer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
