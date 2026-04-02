package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMCPTestRouter(handler *MCPHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/v1/mcp/capabilities", handler.MCPCapabilities)
	r.GET("/v1/mcp/tools", handler.MCPTools)
	r.POST("/v1/mcp/tools/call", handler.MCPToolsCall)
	r.GET("/v1/mcp/prompts", handler.MCPPrompts)
	r.GET("/v1/mcp/resources", handler.MCPResources)
	r.GET("/v1/mcp/tools/search", handler.MCPToolSearch)
	r.POST("/v1/mcp/tools/search", handler.MCPToolSearch)
	r.GET("/v1/mcp/adapters/search", handler.MCPAdapterSearch)
	r.POST("/v1/mcp/adapters/search", handler.MCPAdapterSearch)
	r.GET("/v1/mcp/tools/suggestions", handler.MCPToolSuggestions)
	r.GET("/v1/mcp/categories", handler.MCPCategories)
	r.GET("/v1/mcp/stats", handler.MCPStats)

	return r
}

func TestNewMCPHandler(t *testing.T) {
	t.Run("creates handler with config", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		registry := &services.ProviderRegistry{}

		handler := NewMCPHandler(registry, cfg)

		assert.NotNil(t, handler)
		assert.Equal(t, cfg, handler.config)
		assert.Equal(t, registry, handler.providerRegistry)
	})
}

func TestMCPHandler_MCPCapabilities(t *testing.T) {
	t.Run("returns capabilities when enabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/capabilities", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", response["version"])
		assert.NotNil(t, response["capabilities"])
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/capabilities", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "MCP is not enabled", response["error"])
	})
}

func TestMCPHandler_MCPTools(t *testing.T) {
	t.Run("returns tools list when enabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response["tools"])
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_MCPToolsCall(t *testing.T) {
	t.Run("returns error when MCP disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		body := map[string]interface{}{"name": "test_tool"}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/call", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error when provider registry not available", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		body := map[string]interface{}{"name": "test_tool"}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("returns error when tool name missing", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		registry := &services.ProviderRegistry{}
		handler := NewMCPHandler(registry, cfg)
		router := setupMCPTestRouter(handler)

		body := map[string]interface{}{}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("handles unified namespace tool", func(t *testing.T) {
		cfg := &config.MCPConfig{
			Enabled:              true,
			UnifiedToolNamespace: true,
		}
		registry := &services.ProviderRegistry{}
		handler := NewMCPHandler(registry, cfg)
		router := setupMCPTestRouter(handler)

		body := map[string]interface{}{
			"name": "claude_test_tool",
			"arguments": map[string]interface{}{
				"param1": "value1",
			},
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["result"], "test_tool")
		assert.Contains(t, response["result"], "claude")
	})

	t.Run("returns error for invalid tool format", func(t *testing.T) {
		cfg := &config.MCPConfig{
			Enabled:              true,
			UnifiedToolNamespace: true,
		}
		registry := &services.ProviderRegistry{}
		handler := NewMCPHandler(registry, cfg)
		router := setupMCPTestRouter(handler)

		body := map[string]interface{}{
			"name": "invalidtoolformat", // No underscore
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/call", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestMCPHandler_MCPPrompts(t *testing.T) {
	t.Run("returns prompts when enabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/prompts", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		prompts, ok := response["prompts"].([]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(prompts), 2)
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/prompts", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_MCPResources(t *testing.T) {
	t.Run("returns resources when enabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/resources", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		resources, ok := response["resources"].([]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(resources), 2)
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/resources", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_MCPToolSearch(t *testing.T) {
	t.Run("searches tools via GET", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools/search?q=test&categories=core,web&include_params=true&fuzzy=true&max_results=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "test", response["query"])
		assert.NotNil(t, response["count"])
		assert.NotNil(t, response["results"])
	})

	t.Run("searches tools via POST", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		body := MCPToolSearchRequest{
			Query:         "search query",
			Categories:    []string{"core"},
			IncludeParams: true,
			FuzzyMatch:    true,
			MaxResults:    5,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "search query", response["query"])
	})

	t.Run("returns error for empty query", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools/search", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for invalid JSON in POST", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/tools/search", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools/search?q=test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_MCPAdapterSearch(t *testing.T) {
	t.Run("searches adapters via GET", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/adapters/search?q=git&categories=vcs&official=true", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "git", response["query"])
		assert.NotNil(t, response["count"])
		assert.NotNil(t, response["results"])
	})

	t.Run("searches adapters via POST", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		body := MCPAdapterSearchRequest{
			Query:      "database",
			Categories: []string{"database"},
			MaxResults: 10,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/adapters/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("handles empty POST body", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/mcp/adapters/search", nil)
		router.ServeHTTP(w, req)

		// Should succeed with empty search
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/adapters/search?q=test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_MCPToolSuggestions(t *testing.T) {
	t.Run("returns suggestions when enabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools/suggestions?prefix=git&max=5", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "git", response["prefix"])
		assert.NotNil(t, response["count"])
		assert.NotNil(t, response["suggestions"])
	})

	t.Run("returns error for empty prefix", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools/suggestions", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/tools/suggestions?prefix=test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_MCPCategories(t *testing.T) {
	t.Run("returns categories when enabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/categories", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response["tool_categories"])
		assert.NotNil(t, response["adapter_categories"])
		assert.NotNil(t, response["auth_types"])
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/categories", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_MCPStats(t *testing.T) {
	t.Run("returns stats when enabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/stats", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response["tools"])
		assert.NotNil(t, response["adapters"])

		tools := response["tools"].(map[string]interface{})
		adapters := response["adapters"].(map[string]interface{})

		assert.NotNil(t, tools["total"])
		assert.NotNil(t, tools["by_category"])
		assert.NotNil(t, adapters["total"])
		assert.NotNil(t, adapters["by_category"])
	})

	t.Run("returns service unavailable when disabled", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: false}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/mcp/stats", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestMCPHandler_GetMCPManager(t *testing.T) {
	t.Run("returns manager instance", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)

		manager := handler.GetMCPManager()

		assert.NotNil(t, manager)
	})

	t.Run("returns same manager on multiple calls", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)

		manager1 := handler.GetMCPManager()
		manager2 := handler.GetMCPManager()

		assert.Equal(t, manager1, manager2)
	})
}

func TestMCPHandler_RegisterMCPServer(t *testing.T) {
	t.Run("registers server successfully", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)

		// Initialize manager first
		_ = handler.GetMCPManager()

		config := map[string]interface{}{
			"name": "test-server",
			"url":  "http://localhost:8080",
		}

		err := handler.RegisterMCPServer(config)
		assert.NoError(t, err)
	})

	t.Run("returns error when manager not initialized", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)

		config := map[string]interface{}{
			"name": "test-server",
		}

		err := handler.RegisterMCPServer(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

func TestFindUnderscoreIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"claude_test_tool", 6},
		{"openai_generate", 6},
		{"provider_name", 8},
		{"nounderscore", -1},
		{"_leading", -1},
		{"trailing_", -1},
		{"a_b", 1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := findUnderscoreIndex(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPHandler_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent requests", func(t *testing.T) {
		cfg := &config.MCPConfig{Enabled: true}
		handler := NewMCPHandler(nil, cfg)
		router := setupMCPTestRouter(handler)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/mcp/capabilities", nil)
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
			}()
		}

		wg.Wait()
	})
}

// MockMCPManager is a mock implementation for testing
type MockMCPManager struct {
	ListMCPServersFunc func(ctx context.Context) ([]*services.MCPServer, error)
}

func (m *MockMCPManager) ListMCPServers(ctx context.Context) ([]*services.MCPServer, error) {
	if m.ListMCPServersFunc != nil {
		return m.ListMCPServersFunc(ctx)
	}
	return nil, nil
}
