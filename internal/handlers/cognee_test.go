package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dev.helix.agent/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewCogneeHandler(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: "http://localhost:8000",
			APIKey:  "test-api-key",
		},
	}

	handler := NewCogneeHandler(cfg)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.client)
}

func TestCogneeHandler_VisualizeGraph(t *testing.T) {
	// Create a mock Cognee server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/visualize" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"graph": map[string]interface{}{
					"nodes": []interface{}{"n1", "n2"},
					"edges": []interface{}{"e1"},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: mockServer.URL,
		},
	}
	handler := NewCogneeHandler(cfg)

	t.Run("with default dataset", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/cognee/visualize", nil)

		handler.VisualizeGraph(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("with specified dataset", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/cognee/visualize?dataset=my-dataset&format=json", nil)

		handler.VisualizeGraph(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCogneeHandler_VisualizeGraph_Error(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: mockServer.URL,
		},
	}
	handler := NewCogneeHandler(cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/v1/cognee/visualize", nil)

	handler.VisualizeGraph(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to visualize graph")
}

func TestCogneeHandler_GetDatasets(t *testing.T) {
	// Create a mock Cognee server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/datasets" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"datasets": []map[string]interface{}{
					{"id": "ds-1", "name": "dataset1"},
					{"id": "ds-2", "name": "dataset2"},
				},
				"total": 2,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: mockServer.URL,
		},
	}
	handler := NewCogneeHandler(cfg)

	t.Run("successful list", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/cognee/datasets", nil)

		handler.GetDatasets(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(2), response["total"])
	})
}

func TestCogneeHandler_GetDatasets_Error(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: mockServer.URL,
		},
	}
	handler := NewCogneeHandler(cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/v1/cognee/datasets", nil)

	handler.GetDatasets(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to list datasets")
}

func BenchmarkCogneeHandler_VisualizeGraph(b *testing.B) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"graph": map[string]interface{}{}})
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Cognee: config.CogneeConfig{BaseURL: mockServer.URL},
	}
	handler := NewCogneeHandler(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/cognee/visualize", nil)
		handler.VisualizeGraph(c)
	}
}
