package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/database"
	"github.com/superagent/superagent/internal/handlers"
	"github.com/superagent/superagent/internal/middleware"
	"github.com/superagent/superagent/internal/services"
)

func TestMultiProviderIntegration(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8081", // Use different port for testing
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			Name:     "test_superagent",
			User:     "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
	}

	// Initialize database connection
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Initialize provider registry with test configuration
	registryConfig := &services.RegistryConfig{
		Providers: map[string]*services.ProviderConfig{
			"test-provider": {
				Name:    "Test Provider",
				Type:    "openrouter",
				Enabled: true,
				APIKey:  "test-key",
				Models: []services.ModelConfig{
					{
						ID:      "test-model",
						Name:    "Test Model",
						Enabled: true,
					},
				},
			},
		},
	}

	// Create memory service
	memoryService := services.NewMemoryService(cfg)

	// Initialize provider registry
	providerRegistry := services.NewProviderRegistry(registryConfig, memoryService)

	// Initialize handlers
	unifiedHandler := handlers.NewUnifiedHandler(providerRegistry, cfg)

	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Recovery())

	// Register routes
	api := router.Group("/v1")
	auth := func(c *gin.Context) { /* Simple auth for tests */ }
	unifiedHandler.RegisterOpenAIRoutes(api, auth)

	// Test models endpoint
	t.Run("ModelsEndpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/models", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}

		if data, ok := response["data"].([]interface{}); !ok {
			t.Error("Expected data array in response")
		} else {
			if len(data) == 0 {
				t.Error("Expected at least one model in response")
			}
		}
	})

	// Test chat completions endpoint
	t.Run("ChatCompletionsEndpoint", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "superagent-ensemble",
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": "Hello, this is a test.",
				},
			},
			"max_tokens": 10,
		}

		jsonData, _ := json.Marshal(request)
		req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// In test mode, should return mock response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}

		if response["model"] == nil {
			t.Error("Expected model in response")
		}

		if response["choices"] == nil {
			t.Error("Expected choices in response")
		}
	})
}

func TestProviderHealthCheck(t *testing.T) {
	// Test health check functionality
	cfg := &config.Config{}
	
	// Create memory service
	memoryService := services.NewMemoryService(cfg)

	// Initialize provider registry
	providerRegistry := services.NewProviderRegistry(&services.RegistryConfig{}, memoryService)

	// Test provider list
	providers := providerRegistry.ListProviders()
	if providers == nil {
		t.Error("Expected providers list, got nil")
	}

	// Test ensemble service status
	ensemble := providerRegistry.GetEnsembleService()
	if ensemble == nil {
		t.Error("Expected ensemble service, got nil")
	}
}

func TestAuthenticationFlow(t *testing.T) {
	// Test authentication middleware
	authConfig := middleware.AuthConfig{
		SecretKey:   "test-secret-key",
		TokenExpiry: time.Hour,
		Issuer:      "test-superagent",
		SkipPaths:   []string{"/health", "/v1/models"},
	}

	userService := services.NewUserService(nil, authConfig.SecretKey, authConfig.TokenExpiry)
	authMiddleware := middleware.NewAuthMiddleware(authConfig, userService)

	// Test token generation
	token, err := authMiddleware.GenerateToken("test-user", "testuser", "user")
	if err != nil {
		t.Errorf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Test token validation through middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware.Middleware(authConfig.SkipPaths))
	
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "authenticated"})
	})

	// Test with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with valid token, got %d", w.Code)
	}

	// Test with invalid token
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with invalid token, got %d", w.Code)
	}
}
