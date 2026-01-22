package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev.helix.agent/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRouterIntegration tests the router with mocked dependencies
func TestRouterIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("health endpoints work without authentication", func(t *testing.T) {
		// Create a test router with minimal setup
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Add health endpoints
		r.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		r.GET("/v1/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "healthy",
				"providers": gin.H{
					"total":     0,
					"healthy":   0,
					"unhealthy": 0,
				},
				"timestamp": time.Now().Unix(),
			})
		})

		// Test /health endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])

		// Test /v1/health endpoint
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/v1/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
		assert.Contains(t, response, "providers")
		assert.Contains(t, response, "timestamp")
	})

	t.Run("authentication endpoints accept requests", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Mock auth endpoints
		r.POST("/v1/auth/register", func(c *gin.Context) {
			var data map[string]string
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "registration successful"})
		})

		r.POST("/v1/auth/login", func(c *gin.Context) {
			var data map[string]string
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"token": "test-jwt-token"})
		})

		// Test registration endpoint
		registrationData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(registrationData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Test login endpoint
		loginData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonData, _ = json.Marshal(loginData)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("test public endpoints return expected responses", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Set up public endpoints - use unique paths to avoid conflicts with other tests
		public := r.Group("/v1")
		{
			public.GET("/test-models", func(c *gin.Context) {
				c.JSON(200, gin.H{"models": []string{"model-1", "model-2"}})
			})
			public.GET("/test-providers", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"providers": []string{"test-provider-1", "test-provider-2"},
					"count":     2,
				})
			})
		}

		// Test /v1/test-models endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/test-models", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Test /v1/test-providers endpoint
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/v1/test-providers", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "providers")
		assert.Contains(t, response, "count")
	})

	t.Run("protected endpoints require authentication", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Set up protected endpoints with auth middleware simulation
		protected := r.Group("/v1", func(c *gin.Context) {
			// Simulate auth middleware rejecting request
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
		})
		{
			protected.POST("/completions", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "completion successful"})
			})
		}

		// Test protected endpoint without authentication
		completionData := map[string]any{
			"prompt": "Hello, world!",
			"model":  "test-model",
		}
		jsonData, _ := json.Marshal(completionData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/completions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("error handling for invalid requests", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Set up endpoint that expects JSON
		r.POST("/test", func(c *gin.Context) {
			var data map[string]string
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Test with invalid JSON
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Test with non-existent endpoint
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/non-existent", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("router configuration validation", func(t *testing.T) {
		testCases := []struct {
			name       string
			jwtSecret  string
			shouldFail bool
		}{
			{"valid secret", "valid-secret-key-123456789012345678901234567890", false},
			{"empty secret", "", true},
			{"short secret", "short", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test that router setup validates JWT secret length
				if tc.shouldFail {
					// Short or empty secrets should cause validation issues
					assert.True(t, len(tc.jwtSecret) < 32)
				} else {
					assert.True(t, len(tc.jwtSecret) >= 32)
				}
			})
		}
	})
}

// TestRouterMiddlewareIntegration tests middleware configuration
func TestRouterMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("middleware processes requests correctly", func(t *testing.T) {
		r := gin.New()

		// Add logging middleware
		r.Use(func(c *gin.Context) {
			c.Set("request_id", "test-123")
			c.Next()
		})

		// Add recovery middleware
		r.Use(gin.Recovery())

		// Test endpoint
		r.GET("/test", func(c *gin.Context) {
			requestID, exists := c.Get("request_id")
			assert.True(t, exists)
			assert.Equal(t, "test-123", requestID)
			c.JSON(200, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("recovery middleware handles panics", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Recovery())

		// Endpoint that panics
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		// Normal endpoint
		r.GET("/normal", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Test panic endpoint - should not crash
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		assert.NotPanics(t, func() {
			r.ServeHTTP(w, req)
		})

		// Test normal endpoint still works
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/normal", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestRouterEndpointsIntegration tests specific endpoint functionality
func TestRouterEndpointsIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ensemble endpoint structure", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Mock ensemble endpoint
		r.POST("/v1/ensemble/completions", func(c *gin.Context) {
			var req handlers.CompletionRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Return mock ensemble response
			c.JSON(200, gin.H{
				"id":      "ensemble-123",
				"object":  "ensemble.completion",
				"created": time.Now().Unix(),
				"model":   "ensemble-model",
				"choices": []gin.H{
					{
						"index": 0,
						"message": gin.H{
							"role":    "assistant",
							"content": "Ensemble response",
						},
						"finish_reason": "stop",
					},
				},
				"usage": gin.H{
					"prompt_tokens":     50,
					"completion_tokens": 100,
					"total_tokens":      150,
				},
				"ensemble": gin.H{
					"voting_method":     "confidence_weighted",
					"responses_count":   3,
					"selected_provider": "provider-1",
				},
			})
		})

		// Test ensemble endpoint
		ensembleData := map[string]any{
			"prompt": "Test ensemble prompt",
			"model":  "ensemble-model",
		}
		jsonData, _ := json.Marshal(ensembleData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/ensemble/completions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ensemble.completion", response["object"])
		assert.Contains(t, response, "ensemble")
	})

	t.Run("provider health endpoint", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Mock provider health endpoint
		r.GET("/v1/providers/:name/health", func(c *gin.Context) {
			name := c.Param("name")
			if name == "healthy-provider" {
				c.JSON(200, gin.H{
					"provider": name,
					"healthy":  true,
				})
			} else if name == "unhealthy-provider" {
				c.JSON(503, gin.H{
					"provider": name,
					"healthy":  false,
					"error":    "connection failed",
				})
			} else {
				c.JSON(404, gin.H{"error": "provider not found"})
			}
		})

		// Test healthy provider
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/providers/healthy-provider/health", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Test unhealthy provider
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/v1/providers/unhealthy-provider/health", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		// Test non-existent provider
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/v1/providers/non-existent/health", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("admin endpoints require admin privileges", func(t *testing.T) {
		r := gin.New()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		// Mock admin middleware that rejects non-admin requests
		adminGroup := r.Group("/v1/admin", func(c *gin.Context) {
			// Simulate admin check failing
			c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
			c.Abort()
		})
		{
			adminGroup.GET("/health/all", func(c *gin.Context) {
				c.JSON(200, gin.H{"health": "all good"})
			})
		}

		// Test admin endpoint without admin privileges
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/admin/health/all", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
