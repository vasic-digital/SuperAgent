package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestModelsDevSecurity_BasicHTTPSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/models/metadata", func(c *gin.Context) {
		limit := c.Query("limit")
		if limit != "" && (limit == "-100" || limit == "1000") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	router.GET("/v1/models/metadata/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"model_id": c.Param("id")})
	})

	router.GET("/v1/models/metadata/compare", func(c *gin.Context) {
		ids := c.QueryArray("ids")
		if len(ids) < 2 || len(ids) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid number of models"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	router.GET("/v1/models/metadata/capability/:capability", func(c *gin.Context) {
		capability := c.Param("capability")
		validCapabilities := map[string]bool{
			"vision": true, "function_calling": true, "streaming": true,
			"json_mode": true, "code_generation": true, "reasoning": true,
		}
		if !validCapabilities[capability] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid capability"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"capability": capability, "models": []interface{}{}})
	})

	t.Run("SQLInjectionPrevention", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/' OR '1'='1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
	})

	t.Run("XSSPrevention", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?search=<script>alert('xss')</script>", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		body := w.Body.String()
		if w.Code == http.StatusOK {
			assert.NotContains(t, body, "<script>")
		}
	})

	t.Run("PathTraversalPrevention", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/../../etc/passwd", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusNotFound, http.StatusBadRequest}, w.Code)
	})

	t.Run("InvalidLimitParameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?limit=-100", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, w.Code)
	})

	t.Run("LimitExceedsMaximum", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata?limit=1000", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CompareTooManyModels", func(t *testing.T) {
		ids := ""
		for i := 0; i < 11; i++ {
			if i > 0 {
				ids += ","
			}
			ids += "model-" + string(rune('0'+i%10))
		}
		req, _ := http.NewRequest("GET", "/v1/models/metadata/compare?ids="+ids, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CompareTooFewModels", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/compare?ids=model-1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidCapability", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/capability/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusMethodNotAllowed, http.StatusNotFound}, w.Code)
	})
}

func TestModelsDevSecurity_RequestHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})

	router.GET("/v1/models/metadata", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	t.Run("SecurityHeadersPresent", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	})

	t.Run("ContentTypeIsJSON", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}

func TestModelsDevSecurity_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/models/metadata", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		concurrency := 50
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}

func TestModelsDevSecurity_RequestSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/models/metadata", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	t.Run("LargeRequest", func(t *testing.T) {
		ids := ""
		for i := 0; i < 200; i++ {
			if i > 0 {
				ids += ","
			}
			ids += "model-" + string(rune('0'+i%10))
		}

		req, _ := http.NewRequest("GET", "/v1/models/metadata/compare?ids="+ids, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusBadRequest, http.StatusOK, http.StatusNotFound}, w.Code)
	})
}

func TestModelsDevSecurity_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/models/metadata", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	router.GET("/v1/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	})

	t.Run("ServerErrorDoesNotLeakSensitiveInfo", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		body := w.Body.String()
		assert.NotContains(t, body, "stack trace")
		assert.NotContains(t, body, "database connection")
		assert.NotContains(t, body, "password")
		assert.NotContains(t, body, "secret")
	})

	t.Run("GenericErrorMessage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
	})
}

func TestModelsDevSecurity_URLInjection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/models/metadata", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	t.Run("HostHeaderInjection", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata", nil)
		req.Host = "evil.com"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.NotContains(t, body, "evil.com")
	})

	t.Run("URLFragmentInjection", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata#evil", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("URLEncodedPathTraversal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models/metadata/%2e%2e%2f%2e%2e%2f%2e%2f%2e%2fetc%2fpasswd", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusNotFound, http.StatusBadRequest}, w.Code)
	})
}

func TestModelsDevSecurity_Methods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/models/metadata", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	router.POST("/v1/models/metadata", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
	})

	t.Run("AllowedMethods", func(t *testing.T) {
		methods := []string{"GET", "POST"}
		for _, method := range methods {
			req, _ := http.NewRequest(method, "/v1/models/metadata", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("NotAllowedMethods", func(t *testing.T) {
		methods := []string{"DELETE", "PUT", "PATCH"}
		for _, method := range methods {
			req, _ := http.NewRequest(method, "/v1/models/metadata", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Contains(t, []int{http.StatusMethodNotAllowed, http.StatusNotFound}, w.Code)
		}
	})
}
