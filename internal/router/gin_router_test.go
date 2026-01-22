package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"dev.helix.agent/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// createTestGinRouter creates a GinRouter for testing without calling SetupRouter
func createTestGinRouter(cfg *config.Config, opts ...GinRouterOption) *GinRouter {
	defaultLog := logrus.New()
	defaultLog.SetLevel(logrus.InfoLevel)

	router := &GinRouter{
		config:  cfg,
		log:     defaultLog,
		running: false,
	}

	// Apply options (may override the default logger)
	for _, opt := range opts {
		opt(router)
	}

	// Create a simple Gin engine for testing
	router.engine = gin.New()
	router.engine.Use(gin.Recovery())
	router.engine.Use(router.requestCounterMiddleware())

	return router
}

// TestNewGinRouter_Structure tests that the router is created with proper structure
// This test uses createTestGinRouter to avoid route registration conflicts
// The actual NewGinRouter function is tested via integration tests
func TestNewGinRouter_Structure(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	// Test router structure using the test helper
	router := createTestGinRouter(cfg)
	require.NotNil(t, router)
	assert.NotNil(t, router.engine)
	assert.NotNil(t, router.log)
	assert.NotNil(t, router.config)
	assert.False(t, router.running)
}

func TestGinRouterStructure(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	t.Run("creates router with proper structure", func(t *testing.T) {
		router := createTestGinRouter(cfg)
		require.NotNil(t, router)
		assert.NotNil(t, router.engine)
		assert.NotNil(t, router.log)
		assert.NotNil(t, router.config)
		assert.False(t, router.running)
	})

	t.Run("applies custom logger option", func(t *testing.T) {
		customLog := logrus.New()
		customLog.SetLevel(logrus.DebugLevel)

		router := createTestGinRouter(cfg, WithLogger(customLog))
		require.NotNil(t, router)
		assert.Equal(t, customLog, router.log)
	})
}

func TestGinRouter_Engine(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)
	engine := router.Engine()

	assert.NotNil(t, engine)
	assert.IsType(t, &gin.Engine{}, engine)
}

func TestGinRouter_IsRunning(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)
	assert.False(t, router.IsRunning())
}

func TestGinRouter_GetStats(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)
	stats := router.GetStats()

	assert.False(t, stats.Running)
	assert.Equal(t, int64(0), stats.RequestCount)
	assert.Equal(t, time.Duration(0), stats.Uptime)
}

func TestGinRouter_RegisterRoutes(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Register a custom route
	router.RegisterRoutes(func(e *gin.Engine) {
		e.GET("/custom", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "custom route"})
		})
	})

	// Test the custom route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/custom", nil)
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "custom route")
}

func TestGinRouter_AddMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	middlewareCalled := false
	router.AddMiddleware(func(c *gin.Context) {
		middlewareCalled = true
		c.Next()
	})

	// Register a route to trigger the middleware
	router.engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Test that middleware is called
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.engine.ServeHTTP(w, req)

	assert.True(t, middlewareCalled)
}

func TestGinRouter_ServeHTTP(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Register a route
	router.engine.GET("/serve-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"served": true})
	})

	// Test ServeHTTP interface
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/serve-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "served")
}

func TestGinRouter_requestCounterMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Register a route
	router.engine.GET("/count-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Make several requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/count-test", nil)
		router.engine.ServeHTTP(w, req)
	}

	stats := router.GetStats()
	assert.Equal(t, int64(5), stats.RequestCount)
}

func TestGinRouter_Start_AlreadyRunning(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Manually set running to true
	router.mu.Lock()
	router.running = true
	router.mu.Unlock()

	err := router.Start(":0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestGinRouter_Shutdown_NotRunning(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	ctx := context.Background()
	err := router.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestGinRouter_StartTLS_AlreadyRunning(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Manually set running to true
	router.mu.Lock()
	router.running = true
	router.mu.Unlock()

	err := router.StartTLS(":0", "cert.pem", "key.pem")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestRouterStats_JSON(t *testing.T) {
	stats := RouterStats{
		Running:      true,
		StartedAt:    time.Now(),
		Uptime:       time.Minute,
		RequestCount: 100,
	}

	assert.True(t, stats.Running)
	assert.Equal(t, int64(100), stats.RequestCount)
	assert.Equal(t, time.Minute, stats.Uptime)
}

func TestWithLogger(t *testing.T) {
	customLog := logrus.New()
	customLog.SetLevel(logrus.ErrorLevel)

	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg, WithLogger(customLog))
	assert.Equal(t, customLog, router.log)
	assert.Equal(t, logrus.ErrorLevel, router.log.Level)
}

func TestWithGinMode(t *testing.T) {
	// Test WithGinMode option directly
	gin.SetMode(gin.ReleaseMode)
	assert.Equal(t, gin.ReleaseMode, gin.Mode())

	// Reset to test mode for other tests
	gin.SetMode(gin.TestMode)
	assert.Equal(t, gin.TestMode, gin.Mode())
}

func TestWithGinMode_ViaOption(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	// Create a router and apply WithGinMode option
	router := &GinRouter{
		config:  cfg,
		log:     logrus.New(),
		running: false,
	}

	// Apply the option
	opt := WithGinMode(gin.DebugMode)
	opt(router)

	// Gin mode should be set
	assert.Equal(t, gin.DebugMode, gin.Mode())

	// Reset to test mode
	gin.SetMode(gin.TestMode)
}

func TestGinRouter_GetStats_WhileRunning(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Simulate running state
	router.mu.Lock()
	router.running = true
	router.startedAt = time.Now().Add(-time.Hour)
	router.requestCnt = 42
	router.mu.Unlock()

	stats := router.GetStats()

	assert.True(t, stats.Running)
	assert.Equal(t, int64(42), stats.RequestCount)
	assert.True(t, stats.Uptime >= time.Hour)
}

func TestGinRouter_Start_AndShutdown(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Add a test route
	router.engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Start the server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		err := router.Start(":0") // port 0 lets OS choose a free port
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Give the server time to start
	time.Sleep(50 * time.Millisecond)

	// Check that it's running
	assert.True(t, router.IsRunning())

	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := router.Shutdown(ctx)
	assert.NoError(t, err)

	// Wait for the server goroutine to exit
	select {
	case err := <-serverErr:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shutdown in time")
	}
}

func TestGinRouter_Shutdown_WithRunningServer(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Add a route
	router.engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Create a custom server and set it up
	router.server = &http.Server{
		Addr:    ":0",
		Handler: router.engine,
	}
	router.running = true

	// Shutdown should work
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := router.Shutdown(ctx)
	assert.NoError(t, err)
	assert.False(t, router.IsRunning())
}

func TestGinRouter_StartTLS_AndShutdown(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Add a test route
	router.engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Start TLS server with invalid certs will fail, but we can test the error handling
	serverErr := make(chan error, 1)
	go func() {
		err := router.StartTLS(":0", "nonexistent.pem", "nonexistent.key")
		serverErr <- err
	}()

	// Wait for the error
	select {
	case err := <-serverErr:
		// Should fail with file not found
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("StartTLS did not return error in time")
	}
}

func TestGinRouter_ConcurrentAccess(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/concurrent-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	var wg sync.WaitGroup
	numRequests := 100

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/concurrent-test", nil)
			w := httptest.NewRecorder()
			router.engine.ServeHTTP(w, req)
		}()
	}

	wg.Wait()

	stats := router.GetStats()
	assert.Equal(t, int64(numRequests), stats.RequestCount, "All requests should be counted")
}

func TestGinRouter_ConcurrentStatsAccess(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = router.GetStats()
				_ = router.IsRunning()
			}
		}()
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestGinRouter_MultipleMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	executionOrder := []string{}
	var mu sync.Mutex

	router.AddMiddleware(func(c *gin.Context) {
		mu.Lock()
		executionOrder = append(executionOrder, "first")
		mu.Unlock()
		c.Next()
	})

	router.AddMiddleware(func(c *gin.Context) {
		mu.Lock()
		executionOrder = append(executionOrder, "second")
		mu.Unlock()
		c.Next()
	})

	router.engine.GET("/multi-middleware", func(c *gin.Context) {
		mu.Lock()
		executionOrder = append(executionOrder, "handler")
		mu.Unlock()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/multi-middleware", nil)
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, []string{"first", "second", "handler"}, executionOrder,
		"Middleware should execute in order")
}

func TestGinRouter_NotFound(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent-route", nil)
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 for unknown routes")
}

func TestGinRouter_MethodNotAllowed(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/get-only", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/get-only", nil)
	router.engine.ServeHTTP(w, req)

	// Gin returns 404 by default for method not allowed
	// unless HandleMethodNotAllowed is set to true
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusMethodNotAllowed,
		"Should return 404 or 405 for wrong method")
}

func TestGinRouter_Panic_Recovery(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Should not panic - Gin has built-in recovery
	require.NotPanics(t, func() {
		router.engine.ServeHTTP(w, req)
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Should return 500 on panic")
}

func TestGinRouter_LargeRequestCount(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/stress", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	var wg sync.WaitGroup
	numRequests := 1000

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/stress", nil)
			w := httptest.NewRecorder()
			router.engine.ServeHTTP(w, req)
		}()
	}

	wg.Wait()

	stats := router.GetStats()
	assert.Equal(t, int64(numRequests), stats.RequestCount,
		"All %d requests should be counted", numRequests)
}

func TestGinRouter_UptimeCalculation(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	startTime := time.Now().Add(-1 * time.Hour) // Started 1 hour ago

	router.mu.Lock()
	router.running = true
	router.startedAt = startTime
	router.mu.Unlock()

	stats := router.GetStats()

	// Uptime should be approximately 1 hour
	assert.True(t, stats.Uptime >= 59*time.Minute, "Uptime should be at least 59 minutes")
	assert.True(t, stats.Uptime <= 61*time.Minute, "Uptime should be at most 61 minutes")

	// Cleanup
	router.mu.Lock()
	router.running = false
	router.mu.Unlock()
}

func TestGinRouter_HTTPMethods(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	// Register routes for different HTTP methods
	router.engine.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "GET"})
	})
	router.engine.POST("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "POST"})
	})
	router.engine.PUT("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "PUT"})
	})
	router.engine.DELETE("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "DELETE"})
	})
	router.engine.PATCH("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "PATCH"})
	})

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, "/api", nil)
			router.engine.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), method)
		})
	}
}

func TestGinRouter_QueryParams(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		page := c.DefaultQuery("page", "1")
		c.JSON(http.StatusOK, gin.H{"query": query, "page": page})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?q=test&page=2", nil)
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
	assert.Contains(t, w.Body.String(), "2")
}

func TestGinRouter_PathParams(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"user_id": id})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/123", nil)
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "123")
}

func TestGinRouter_HeadersHandling(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/headers", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		c.JSON(http.StatusOK, gin.H{"auth": authHeader})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-token")
}

func TestGinRouter_ResponseHeaders(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	router.engine.GET("/custom-headers", func(c *gin.Context) {
		c.Header("X-Custom-Header", "custom-value")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/custom-headers", nil)
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
}

func TestGinRouter_StatusCodes(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	statusCodes := map[string]int{
		"/ok":           http.StatusOK,
		"/created":      http.StatusCreated,
		"/no-content":   http.StatusNoContent,
		"/bad-request":  http.StatusBadRequest,
		"/unauthorized": http.StatusUnauthorized,
		"/forbidden":    http.StatusForbidden,
		"/internal":     http.StatusInternalServerError,
	}

	for path, code := range statusCodes {
		statusCode := code // capture for closure
		router.engine.GET(path, func(c *gin.Context) {
			c.Status(statusCode)
		})
	}

	for path, expectedCode := range statusCodes {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", path, nil)
			router.engine.ServeHTTP(w, req)
			assert.Equal(t, expectedCode, w.Code)
		})
	}
}

func TestGinRouter_GroupRoutes(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)

	v1 := router.engine.Group("/v1")
	{
		v1.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"version": "v1", "resource": "users"})
		})
	}

	v2 := router.engine.Group("/v2")
	{
		v2.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"version": "v2", "resource": "users"})
		})
	}

	t.Run("v1 users", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/users", nil)
		router.engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "v1")
	})

	t.Run("v2 users", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v2/users", nil)
		router.engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "v2")
	})
}

func TestGinRouter_AbortMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: "test-secret-key-1234567890",
		},
	}

	router := createTestGinRouter(cfg)
	handlerCalled := false

	router.AddMiddleware(func(c *gin.Context) {
		if c.GetHeader("X-Block") == "true" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "blocked"})
			return
		}
		c.Next()
	})

	router.engine.GET("/protected", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.Run("blocked request", func(t *testing.T) {
		handlerCalled = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-Block", "true")
		router.engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.False(t, handlerCalled, "Handler should not be called when middleware aborts")
	})

	t.Run("allowed request", func(t *testing.T) {
		handlerCalled = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		router.engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, handlerCalled, "Handler should be called when middleware passes")
	})
}
