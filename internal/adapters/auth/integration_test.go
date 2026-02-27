package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserService struct {
	users map[string]*User
}

func newMockUserService() *mockUserService {
	return &mockUserService{
		users: map[string]*User{
			"test-api-key-123": {
				ID:       1,
				Username: "testuser",
				Role:     "user",
			},
			"admin-api-key-456": {
				ID:       2,
				Username: "admin",
				Role:     "admin",
			},
		},
	}
}

func (m *mockUserService) AuthenticateByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	if user, ok := m.users[apiKey]; ok {
		return user, nil
	}
	return nil, errors.New("invalid api key")
}

func TestAPIKeyValidator_ValidateAPIKey(t *testing.T) {
	logger := logrus.New()
	userService := newMockUserService()
	validator := NewAPIKeyValidator(userService, logger)

	tests := []struct {
		name      string
		apiKey    string
		wantValid bool
		wantUser  string
	}{
		{
			name:      "valid user api key",
			apiKey:    "test-api-key-123",
			wantValid: true,
			wantUser:  "testuser",
		},
		{
			name:      "valid admin api key",
			apiKey:    "admin-api-key-456",
			wantValid: true,
			wantUser:  "admin",
		},
		{
			name:      "invalid api key",
			apiKey:    "invalid-key",
			wantValid: false,
		},
		{
			name:      "empty api key",
			apiKey:    "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, claims := validator.ValidateAPIKey(tt.apiKey)
			assert.Equal(t, tt.wantValid, valid)

			if tt.wantValid {
				assert.Equal(t, tt.wantUser, claims["username"])
				assert.NotNil(t, claims["user_id"])
				assert.NotNil(t, claims["role"])
			}
		})
	}
}

func TestAPIKeyValidator_ValidateAPIKey_NilService(t *testing.T) {
	logger := logrus.New()
	validator := NewAPIKeyValidator(nil, logger)

	valid, claims := validator.ValidateAPIKey("any-key")
	assert.False(t, valid)
	assert.Nil(t, claims)
}

func TestAPIKeyAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	userService := newMockUserService()
	validator := NewAPIKeyValidator(userService, logger)

	tests := []struct {
		name       string
		apiKey     string
		wantStatus int
		wantAuth   bool
	}{
		{
			name:       "valid api key",
			apiKey:     "test-api-key-123",
			wantStatus: http.StatusOK,
			wantAuth:   true,
		},
		{
			name:       "invalid api key",
			apiKey:     "invalid-key",
			wantStatus: http.StatusUnauthorized,
			wantAuth:   false,
		},
		{
			name:       "no api key",
			apiKey:     "",
			wantStatus: http.StatusOK,
			wantAuth:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(APIKeyAuthMiddleware(validator, "X-API-Key"))
			router.GET("/test", func(c *gin.Context) {
				auth, _ := c.Get("authenticated")
				c.JSON(http.StatusOK, gin.H{"authenticated": auth})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{
			name:    "valid bearer token",
			header:  "Bearer test-token-123",
			want:    "test-token-123",
			wantErr: false,
		},
		{
			name:    "missing bearer prefix",
			header:  "test-token-123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty header",
			header:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "lowercase bearer",
			header:  "bearer test-token",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := extractBearerToken(tt.header)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, token)
			}
		})
	}
}

func TestBearerTokenAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := NewBearerTokenValidator("test-secret", "helixagent")

	router := gin.New()
	router.Use(BearerTokenAuthMiddleware(validator))
	router.GET("/test", func(c *gin.Context) {
		auth, _ := c.Get("authenticated")
		c.JSON(http.StatusOK, gin.H{"authenticated": auth})
	})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "no authorization header",
			authHeader: "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid format",
			authHeader: "invalid-format",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestRequireScopes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		auth       bool
		role       string
		scopes     []string
		wantStatus int
	}{
		{
			name:       "not authenticated",
			auth:       false,
			role:       "",
			scopes:     []string{"read"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "authenticated with correct scope",
			auth:       true,
			role:       "user",
			scopes:     []string{"user"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "authenticated as admin bypasses scopes",
			auth:       true,
			role:       "admin",
			scopes:     []string{"admin", "user"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("authenticated", tt.auth)
				c.Set("role", tt.role)
				c.Next()
			})
			router.Use(RequireScopes(tt.scopes...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetOAuthCredentialPaths(t *testing.T) {
	paths := GetOAuthCredentialPaths()

	assert.NotNil(t, paths)
	assert.IsType(t, map[string]string{}, paths)
}

func TestOAuthCredentialManager(t *testing.T) {
	logger := logrus.New()

	t.Run("new manager with no paths fails", func(t *testing.T) {
		manager, err := NewOAuthCredentialManager(map[string]string{}, "client-id", logger)
		assert.Error(t, err)
		assert.Nil(t, manager)
	})

	t.Run("new manager with paths succeeds", func(t *testing.T) {
		paths := map[string]string{
			"claude": "/tmp/test-claude-creds.json",
		}

		manager, err := NewOAuthCredentialManager(paths, "client-id", logger)
		require.NoError(t, err)
		assert.NotNil(t, manager)
	})
}

func TestGetTokenEndpoint(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{
			provider: "claude",
			want:     "https://api.anthropic.com/oauth/token",
		},
		{
			provider: "qwen",
			want:     "https://dashscope.aliyuncs.com/api/token",
		},
		{
			provider: "unknown",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			endpoint := getTokenEndpoint(tt.provider)
			assert.Equal(t, tt.want, endpoint)
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		auth     interface{}
		expected bool
	}{
		{
			name:     "authenticated true",
			auth:     true,
			expected: true,
		},
		{
			name:     "authenticated false",
			auth:     false,
			expected: false,
		},
		{
			name:     "not set",
			auth:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.auth != nil {
				c.Set("authenticated", tt.auth)
			}
			assert.Equal(t, tt.expected, isAuthenticated(c))
		})
	}
}

func TestGetUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		role     interface{}
		expected string
	}{
		{
			name:     "admin role",
			role:     "admin",
			expected: "admin",
		},
		{
			name:     "user role",
			role:     "user",
			expected: "user",
		},
		{
			name:     "not string",
			role:     123,
			expected: "",
		},
		{
			name:     "not set",
			role:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.role != nil {
				c.Set("role", tt.role)
			}
			assert.Equal(t, tt.expected, getUserRole(c))
		})
	}
}

func TestHasScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		role     string
		scope    string
		expected bool
	}{
		{
			name:     "matches role",
			role:     "user",
			scope:    "user",
			expected: true,
		},
		{
			name:     "admin has all scopes",
			role:     "admin",
			scope:    "anything",
			expected: true,
		},
		{
			name:     "different role",
			role:     "user",
			scope:    "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("role", tt.role)
			assert.Equal(t, tt.expected, hasScope(c, tt.scope))
		})
	}
}

func TestInitializeAuthIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	userService := newMockUserService()

	router := gin.New()
	err := InitializeAuthIntegration(router, userService, "test-secret", logger)
	require.NoError(t, err)

	assert.NotNil(t, router)
}

func TestOAuthCredentialManager_Integration(t *testing.T) {
	if os.Getenv("SKIP_OAUTH_TESTS") != "" {
		t.Skip("Skipping OAuth integration tests")
	}

	logger := logrus.New()
	tempDir := t.TempDir()

	credsFile := tempDir + "/test-creds.json"
	credsContent := `{
		"access_token": "test-token",
		"refresh_token": "test-refresh",
		"expires_at": "2099-01-01T00:00:00Z"
	}`
	err := os.WriteFile(credsFile, []byte(credsContent), 0644)
	require.NoError(t, err)

	paths := map[string]string{
		"test": credsFile,
	}

	manager, err := NewOAuthCredentialManager(paths, "test-client", logger)
	require.NoError(t, err)

	t.Run("get access token", func(t *testing.T) {
		token, err := manager.GetAccessToken("test")
		require.NoError(t, err)
		assert.Equal(t, "test-token", token)
	})

	t.Run("has valid credentials", func(t *testing.T) {
		valid := manager.HasValidCredentials("test")
		assert.True(t, valid)
	})

	t.Run("invalid provider", func(t *testing.T) {
		_, err := manager.GetAccessToken("nonexistent")
		assert.Error(t, err)
	})

	t.Run("start and refresh", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		manager.Start(ctx)
		time.Sleep(50 * time.Millisecond)

		err := manager.RefreshAll(ctx)
		assert.NoError(t, err)
	})
}
