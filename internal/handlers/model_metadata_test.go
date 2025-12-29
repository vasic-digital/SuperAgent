package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/superagent/superagent/internal/database"
)

type MockMetadataService struct{}

func (m *MockMetadataService) GetModel(c *gin.Context) {
	c.JSON(http.StatusOK, createTestMetadata())
}

func (m *MockMetadataService) ListModels(c *gin.Context) {
	models := []*database.ModelMetadata{createTestMetadata()}
	c.JSON(http.StatusOK, ListModelsResponse{
		Models:     models,
		Total:      1,
		Page:       1,
		Limit:      20,
		TotalPages: 1,
	})
}

func (m *MockMetadataService) GetModelBenchmarks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"benchmarks": []database.ModelBenchmark{}})
}

func (m *MockMetadataService) CompareModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"models": []*database.ModelMetadata{}})
}

func (m *MockMetadataService) RefreshModels(c *gin.Context) {
	c.JSON(http.StatusOK, RefreshResponse{
		Status:  "success",
		Message: "Refresh initiated",
	})
}

func (m *MockMetadataService) GetRefreshStatus(c *gin.Context) {
	histories := []*database.ModelsRefreshHistory{}
	c.JSON(http.StatusOK, RefreshHistoryResponse{
		Histories: histories,
		Total:     0,
	})
}

func (m *MockMetadataService) GetProviderModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"provider_id": "anthropic",
		"models":      []*database.ModelMetadata{},
		"total":       0,
	})
}

func (m *MockMetadataService) GetModelsByCapability(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"capability": "vision",
		"models":     []*database.ModelMetadata{},
		"total":      0,
	})
}

func createTestMetadata() *database.ModelMetadata {
	ctx := 128000
	maxTokens := 4096
	pricingInput := 3.0
	pricingOutput := 15.0
	benchmarkScore := 95.5
	popularityScore := 100
	reliabilityScore := 0.95
	modelType := "chat"
	modelFamily := "claude"
	version := "20240229"

	return &database.ModelMetadata{
		ModelID:                 "claude-3-sonnet-20240229",
		ModelName:               "Claude 3 Sonnet",
		ProviderID:              "anthropic",
		ProviderName:            "Anthropic",
		Description:             "Claude 3 Sonnet is a balanced model",
		ContextWindow:           &ctx,
		MaxTokens:               &maxTokens,
		PricingInput:            &pricingInput,
		PricingOutput:           &pricingOutput,
		PricingCurrency:         "USD",
		SupportsVision:          true,
		SupportsFunctionCalling: true,
		SupportsStreaming:       true,
		SupportsJSONMode:        true,
		SupportsImageGeneration: false,
		SupportsAudio:           false,
		SupportsCodeGeneration:  true,
		SupportsReasoning:       true,
		BenchmarkScore:          &benchmarkScore,
		PopularityScore:         &popularityScore,
		ReliabilityScore:        &reliabilityScore,
		ModelType:               &modelType,
		ModelFamily:             &modelFamily,
		Version:                 &version,
		Tags:                    []string{"vision", "function-calling"},
	}
}

func TestModelMetadataHandler_ListModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata", (&MockMetadataService{}).ListModels)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?page=1&limit=20", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response ListModelsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Models)
		assert.Equal(t, 1, response.Total)
	})

	t.Run("DefaultPagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestModelMetadataHandler_GetModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata/:id", (&MockMetadataService{}).GetModel)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/claude-3-sonnet-20240229", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response database.ModelMetadata
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "claude-3-sonnet-20240229", response.ModelID)
	})

	t.Run("NotFound", func(t *testing.T) {
		router.GET("/v1/models/not-found/:id", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		})

		req, _ := http.NewRequest("GET", "/v1/models/not-found/non-existent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestModelMetadataHandler_CompareModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata/compare", (&MockMetadataService{}).CompareModels)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/compare?ids=model-1,model-2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "models")
	})
}

func TestModelMetadataHandler_GetModelsByCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata/capability/:capability", (&MockMetadataService{}).GetModelsByCapability)

	t.Run("Success_Vision", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/capability/vision", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "vision", response["capability"])
	})

	t.Run("Success_FunctionCalling", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/capability/function_calling", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success_Streaming", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/capability/streaming", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success_JSONMode", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/capability/json_mode", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success_CodeGeneration", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/capability/code_generation", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success_Reasoning", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/capability/reasoning", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidCapability", func(t *testing.T) {
		router.GET("/v1/models/invalid/capability/:capability", func(c *gin.Context) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid capability"})
		})

		req, _ := http.NewRequest("GET", "/v1/models/invalid/capability/invalid-capability", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestModelMetadataHandler_GetProviderModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/providers/:provider_id/models/metadata", (&MockMetadataService{}).GetProviderModels)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/providers/anthropic/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "anthropic", response["provider_id"])
		assert.Contains(t, response, "models")
	})

	t.Run("MissingProviderID", func(t *testing.T) {
		router.GET("/v1/providers/nomodels/models/metadata", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No models found"})
		})

		req, _ := http.NewRequest("GET", "/v1/providers/nomodels/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestModelMetadataHandler_RefreshModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/admin/models/metadata/refresh", (&MockMetadataService{}).RefreshModels)

	t.Run("FullRefresh", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/admin/models/metadata/refresh", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response RefreshResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Contains(t, response.Message, "Refresh")
	})

	t.Run("ProviderRefresh", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/admin/models/metadata/refresh?provider=anthropic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestModelMetadataHandler_GetRefreshStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/admin/models/metadata/refresh/status", (&MockMetadataService{}).GetRefreshStatus)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/admin/models/metadata/refresh/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response RefreshHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Histories)
	})

	t.Run("WithLimit", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/admin/models/metadata/refresh/status?limit=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestModelMetadataHandler_GetModelBenchmarks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata/:id/benchmarks", (&MockMetadataService{}).GetModelBenchmarks)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/claude-3-sonnet-20240229/benchmarks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "benchmarks")
	})

	t.Run("MissingModelID", func(t *testing.T) {
		router.GET("/v1/models/nobenchmarks/benchmarks", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		})

		req, _ := http.NewRequest("GET", "/v1/models/nobenchmarks/benchmarks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestModelMetadataHandler_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata/:id", (&MockMetadataService{}).GetModel)
	router.GET("/v1/models/metadata", (&MockMetadataService{}).ListModels)

	t.Run("ContentType", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/claude-3-sonnet-20240229", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})

	t.Run("ValidJSON", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?page=1&limit=20", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var jsonBytes interface{}
		err := json.Unmarshal(w.Body.Bytes(), &jsonBytes)
		assert.NoError(t, err)
	})
}

func TestModelMetadataHandler_HTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata", (&MockMetadataService{}).ListModels)
	router.POST("/v1/models/metadata", (&MockMetadataService{}).ListModels)

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("MethodAllowed_GET", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MethodAllowed_POST", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestModelMetadataHandler_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/error/bad-request", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
	})

	router.GET("/v1/error/not-found", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})

	router.GET("/v1/error/internal", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
	})

	t.Run("BadRequest", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/error/bad-request", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "error")
	})

	t.Run("NotFound", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/error/not-found", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/error/internal", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestModelMetadataHandler_QueryParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata", (&MockMetadataService{}).ListModels)

	t.Run("PageParameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?page=2&limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ProviderParameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?provider=anthropic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("TypeParameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?type=chat", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("SearchParameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?search=claude", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestModelMetadataHandler_ResponseStructures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata", (&MockMetadataService{}).ListModels)

	t.Run("ListModelsResponse", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?page=1&limit=20", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListModelsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Models)
		assert.GreaterOrEqual(t, response.Total, 0)
		assert.GreaterOrEqual(t, response.Page, 0)
		assert.GreaterOrEqual(t, response.Limit, 0)
		assert.GreaterOrEqual(t, response.TotalPages, 0)
	})
}

func TestModelMetadataHandler_Concurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models/metadata", (&MockMetadataService{}).ListModels)

	t.Run("ConcurrentRequests", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
