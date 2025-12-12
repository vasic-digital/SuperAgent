package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/database"
	"github.com/superagent/superagent/internal/services"
)

// MockDatabase implements database interface for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) NewPostgresDB(cfg *config.Config) (*database.DB, error) {
	args := m.Called(cfg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.DB), args.Error(1)
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockUserService implements user service for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) NewUserService(db *database.DB, secret string, expiry time.Duration) *services.UserService {
	args := m.Called(db, secret, expiry)
	return args.Get(0).(*services.UserService)
}

func (m *MockUserService) Register(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) ValidateToken(token string) (gin.H, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(gin.H), args.Error(1)
}

// MockProviderRegistry implements provider registry for testing
type MockProviderRegistry struct {
	mock.Mock
}

func (m *MockProviderRegistry) NewProviderRegistry(config *services.RegistryConfig, memoryService *services.MemoryService) *services.ProviderRegistry {
	args := m.Called(config, memoryService)
	return args.Get(0).(*services.ProviderRegistry)
}

func (m *MockProviderRegistry) GetRequestService() *services.RequestService {
	args := m.Called()
	return args.Get(0).(*services.RequestService)
}

func (m *MockProviderRegistry) GetEnsembleService() *services.EnsembleService {
	args := m.Called()
	return args.Get(0).(*services.EnsembleService)
}

func (m *MockProviderRegistry) HealthCheck() map[string]error {
	args := m.Called()
	return args.Get(0).(map[string]error)
}

func (m *MockProviderRegistry) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockProviderRegistry) GetProvider(name string) (any, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func TestSetupRouter(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("router setup with valid config", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)
		assert.NotNil(t, router)
		assert.IsType(t, &gin.Engine{}, router)
	})
}

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET /health returns healthy status", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})

	t.Run("GET /v1/health returns enhanced health check", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
		assert.Contains(t, response, "providers")
		assert.Contains(t, response, "timestamp")
	})
}

func TestPublicEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET /v1/models returns model list", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/models", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GET /v1/providers returns provider list", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/providers", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "providers")
		assert.Contains(t, response, "count")
	})
}

func TestAuthenticationEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("POST /v1/auth/register accepts registration request", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		registrationData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(registrationData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Registration endpoint exists and accepts requests
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("POST /v1/auth/login accepts login request", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		loginData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(loginData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Login endpoint exists and accepts requests
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})
}

func TestProtectedEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("POST /v1/completions requires authentication", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		completionData := map[string]any{
			"prompt": "Hello, world!",
			"model":  "test-model",
		}
		jsonData, _ := json.Marshal(completionData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/completions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should return 401 or 403 without authentication
		assert.NotEqual(t, http.StatusOK, w.Code)
		assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, w.Code)
	})

	t.Run("POST /v1/chat/completions requires authentication", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		chatData := map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": "Hello!"},
			},
			"model": "test-model",
		}
		jsonData, _ := json.Marshal(chatData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should return 401 or 403 without authentication
		assert.NotEqual(t, http.StatusOK, w.Code)
		assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, w.Code)
	})
}

func TestEnsembleEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("POST /v1/ensemble/completions endpoint exists", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		ensembleData := map[string]any{
			"prompt": "Test ensemble prompt",
			"model":  "ensemble-model",
		}
		jsonData, _ := json.Marshal(ensembleData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/ensemble/completions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should return 401 or 403 without authentication
		assert.NotEqual(t, http.StatusOK, w.Code)
		assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, w.Code)
	})
}

func TestProviderManagementEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET /v1/providers/:name/health endpoint exists", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/providers/test-provider/health", nil)
		router.ServeHTTP(w, req)

		// Should return 401 or 403 without authentication
		assert.NotEqual(t, http.StatusOK, w.Code)
		assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, w.Code)
	})
}

func TestAdminEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET /v1/admin/health/all endpoint exists", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/admin/health/all", nil)
		router.ServeHTTP(w, req)

		// Should return 401 or 403 without authentication
		assert.NotEqual(t, http.StatusOK, w.Code)
		assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, w.Code)
	})
}

func TestRouterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("router includes required middleware", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		// Test that health endpoint works (no auth required)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Test that metrics endpoint exists
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/metrics", nil)
		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})
}

func TestRouterErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("router handles invalid JSON gracefully", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		// Test with invalid JSON
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/auth/register", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("router handles non-existent endpoints", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				JWTSecret: "test-secret-key-1234567890",
			},
		}

		router := SetupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/non-existent-endpoint", nil)
		router.ServeHTTP(w, req)

		// Should return 404 Not Found
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRouterConfiguration(t *testing.T) {
	t.Run("router configuration with different JWT secrets", func(t *testing.T) {
		testCases := []struct {
			name       string
			jwtSecret  string
			shouldFail bool
		}{
			{"valid secret", "valid-secret-key-1234567890", false},
			{"empty secret", "", true},      // Empty secret should cause issues
			{"short secret", "short", true}, // Too short secret should cause issues
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &config.Config{
					Server: config.ServerConfig{
						JWTSecret: tc.jwtSecret,
					},
				}

				// This will panic if JWT secret is invalid
				if tc.shouldFail {
					assert.Panics(t, func() {
						SetupRouter(cfg)
					})
				} else {
					assert.NotPanics(t, func() {
						router := SetupRouter(cfg)
						assert.NotNil(t, router)
					})
				}
			})
		}
	})
}
