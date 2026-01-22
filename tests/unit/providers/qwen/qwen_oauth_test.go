package qwen_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dev.helix.agent/internal/auth/oauth_credentials"
	"dev.helix.agent/internal/llm/providers/qwen"
	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQwenOAuth_APIEndpointSelection tests that the correct API endpoint is selected
// based on whether compatible-mode is used
func TestQwenOAuth_APIEndpointSelection(t *testing.T) {
	tests := []struct {
		name             string
		baseURL          string
		expectedEndpoint string
	}{
		{
			name:             "compatible-mode uses chat/completions endpoint",
			baseURL:          "/compatible-mode/v1",
			expectedEndpoint: "/compatible-mode/v1/chat/completions",
		},
		{
			name:             "standard mode uses generation endpoint",
			baseURL:          "/api/v1",
			expectedEndpoint: "/api/v1/services/aigc/text-generation/generation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path

				response := map[string]interface{}{
					"id":      "test-id",
					"object":  "chat.completion",
					"created": time.Now().Unix(),
					"model":   "qwen-plus",
					"choices": []map[string]interface{}{
						{
							"index": 0,
							"message": map[string]interface{}{
								"role":    "assistant",
								"content": "Test response",
							},
							"finish_reason": "stop",
						},
					},
					"usage": map[string]interface{}{
						"total_tokens": 10,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			provider := qwen.NewQwenProvider("test-api-key", server.URL+tt.baseURL, "qwen-plus")

			req := &models.LLMRequest{
				ID:     "test-req",
				Prompt: "Test prompt",
			}

			_, err := provider.Complete(context.Background(), req)
			require.NoError(t, err)

			// Verify the correct endpoint was called
			assert.Contains(t, receivedPath, tt.expectedEndpoint,
				"Expected endpoint %s but got %s", tt.expectedEndpoint, receivedPath)
		})
	}
}

// TestQwenOAuth_CredentialFileStructure tests parsing of Qwen OAuth credential files
func TestQwenOAuth_CredentialFileStructure(t *testing.T) {
	// Create a temporary credentials file
	tempDir := t.TempDir()
	credPath := filepath.Join(tempDir, "oauth_creds.json")

	// Test valid credentials
	t.Run("valid credentials", func(t *testing.T) {
		validCreds := oauth_credentials.QwenOAuthCredentials{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			IDToken:      "test-id-token",
			ExpiryDate:   time.Now().Add(time.Hour).UnixMilli(),
			TokenType:    "Bearer",
			ResourceURL:  "https://dashscope.aliyuncs.com",
		}

		data, err := json.Marshal(validCreds)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(credPath, data, 0600))

		// Read and verify
		readData, err := os.ReadFile(credPath)
		require.NoError(t, err)

		var parsed oauth_credentials.QwenOAuthCredentials
		require.NoError(t, json.Unmarshal(readData, &parsed))

		assert.Equal(t, validCreds.AccessToken, parsed.AccessToken)
		assert.Equal(t, validCreds.RefreshToken, parsed.RefreshToken)
		assert.Equal(t, validCreds.IDToken, parsed.IDToken)
		assert.Equal(t, validCreds.ExpiryDate, parsed.ExpiryDate)
		assert.Equal(t, validCreds.TokenType, parsed.TokenType)
	})

	// Test expired credentials detection
	t.Run("expired credentials detection", func(t *testing.T) {
		expiredCreds := oauth_credentials.QwenOAuthCredentials{
			AccessToken:  "expired-token",
			RefreshToken: "refresh-token",
			ExpiryDate:   time.Now().Add(-time.Hour).UnixMilli(), // Expired 1 hour ago
			TokenType:    "Bearer",
		}

		data, err := json.Marshal(expiredCreds)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(credPath, data, 0600))

		assert.True(t, oauth_credentials.IsExpired(expiredCreds.ExpiryDate))
		assert.True(t, oauth_credentials.NeedsRefresh(expiredCreds.ExpiryDate))
	})

	// Test credentials needing refresh
	t.Run("credentials needing refresh", func(t *testing.T) {
		needsRefreshCreds := oauth_credentials.QwenOAuthCredentials{
			AccessToken:  "expiring-soon-token",
			RefreshToken: "refresh-token",
			ExpiryDate:   time.Now().Add(5 * time.Minute).UnixMilli(), // Expires in 5 minutes
			TokenType:    "Bearer",
		}

		assert.False(t, oauth_credentials.IsExpired(needsRefreshCreds.ExpiryDate))
		assert.True(t, oauth_credentials.NeedsRefresh(needsRefreshCreds.ExpiryDate))
	})
}

// TestQwenOAuth_ProviderWithOAuth tests creating a provider with OAuth credentials
func TestQwenOAuth_ProviderWithOAuth(t *testing.T) {
	// Create mock OAuth server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify OAuth Bearer token
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"message": "Missing Bearer token"}}`))
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"message": "Empty token"}}`))
			return
		}

		response := map[string]interface{}{
			"id":      "oauth-test-id",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "qwen-plus",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "OAuth authenticated response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"total_tokens": 15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with OAuth token (simulating NewQwenProviderWithOAuth)
	provider := qwen.NewQwenProvider(
		"oauth-access-token",
		server.URL+"/compatible-mode/v1",
		"qwen-plus",
	)

	req := &models.LLMRequest{
		ID:     "oauth-test",
		Prompt: "Test OAuth authentication",
	}

	resp, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "OAuth authenticated response", resp.Content)
}

// TestQwenOAuth_AutoDetection tests the auto-detection of OAuth credentials
func TestQwenOAuth_AutoDetection(t *testing.T) {
	// Test when env var is set
	t.Run("env var explicitly enabled", func(t *testing.T) {
		os.Setenv("QWEN_CODE_USE_OAUTH_CREDENTIALS", "true")
		defer os.Unsetenv("QWEN_CODE_USE_OAUTH_CREDENTIALS")

		assert.True(t, oauth_credentials.IsQwenOAuthEnabled())
	})

	t.Run("env var explicitly disabled", func(t *testing.T) {
		os.Setenv("QWEN_CODE_USE_OAUTH_CREDENTIALS", "false")
		defer os.Unsetenv("QWEN_CODE_USE_OAUTH_CREDENTIALS")

		assert.False(t, oauth_credentials.IsQwenOAuthEnabled())
	})

	// Test different value formats
	t.Run("env var value formats", func(t *testing.T) {
		testCases := []struct {
			value    string
			expected bool
		}{
			{"true", true},
			{"1", true},
			{"yes", true},
			{"false", false},
			{"0", false},
			{"no", false},
		}

		for _, tc := range testCases {
			os.Setenv("QWEN_CODE_USE_OAUTH_CREDENTIALS", tc.value)
			assert.Equal(t, tc.expected, oauth_credentials.IsQwenOAuthEnabled(),
				"Expected %v for value '%s'", tc.expected, tc.value)
			os.Unsetenv("QWEN_CODE_USE_OAUTH_CREDENTIALS")
		}
	})
}

// TestQwenOAuth_StreamingWithOAuth tests streaming responses with OAuth authentication
func TestQwenOAuth_StreamingWithOAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify OAuth header
		authHeader := r.Header.Get("Authorization")
		assert.True(t, strings.HasPrefix(authHeader, "Bearer "), "Should have Bearer token")

		// Send SSE streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("Expected http.Flusher")
		}

		chunks := []string{"Hello", " from", " OAuth", " stream", "!"}
		for i, content := range chunks {
			chunk := map[string]interface{}{
				"id":      fmt.Sprintf("stream-chunk-%d", i),
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   "qwen-plus",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"delta": map[string]interface{}{
							"content": content,
						},
					},
				},
			}
			jsonData, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	provider := qwen.NewQwenProvider(
		"oauth-streaming-token",
		server.URL+"/compatible-mode/v1",
		"qwen-plus",
	)

	req := &models.LLMRequest{
		ID:     "oauth-stream-test",
		Prompt: "Test OAuth streaming",
	}

	responseChan, err := provider.CompleteStream(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, responseChan)

	var fullContent string
	for chunk := range responseChan {
		fullContent += chunk.Content
	}

	assert.Equal(t, "Hello from OAuth stream!", fullContent)
}

// TestQwenOAuth_ErrorHandling tests error handling for OAuth-related errors
func TestQwenOAuth_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		errorResponse string
		expectedError string
	}{
		{
			name:       "expired token",
			statusCode: http.StatusUnauthorized,
			errorResponse: `{
				"error": {
					"message": "Token has expired",
					"type": "authentication_error",
					"code": "401"
				}
			}`,
			expectedError: "Token has expired",
		},
		{
			name:       "invalid token",
			statusCode: http.StatusUnauthorized,
			errorResponse: `{
				"error": {
					"message": "Invalid access token",
					"type": "authentication_error",
					"code": "401"
				}
			}`,
			expectedError: "Invalid access token",
		},
		{
			name:       "rate limited",
			statusCode: http.StatusTooManyRequests,
			errorResponse: `{
				"error": {
					"message": "Rate limit exceeded",
					"type": "rate_limit_error",
					"code": "429"
				}
			}`,
			expectedError: "Rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.errorResponse))
			}))
			defer server.Close()

			provider := qwen.NewQwenProviderWithRetry(
				"test-token",
				server.URL+"/compatible-mode/v1",
				"qwen-plus",
				qwen.RetryConfig{MaxRetries: 0},
			)

			req := &models.LLMRequest{
				ID:     "error-test",
				Prompt: "Test error handling",
			}

			_, err := provider.Complete(context.Background(), req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestQwenOAuth_CompatibleModeEndpoint tests that compatible-mode properly uses
// the OpenAI-compatible /chat/completions endpoint
func TestQwenOAuth_CompatibleModeEndpoint(t *testing.T) {
	var capturedPath string
	var capturedMethod string
	var capturedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		capturedContentType = r.Header.Get("Content-Type")

		response := map[string]interface{}{
			"id":      "compat-test",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "qwen-plus",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Compatible mode response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"total_tokens": 10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := qwen.NewQwenProvider(
		"test-token",
		server.URL+"/compatible-mode/v1",
		"qwen-plus",
	)

	req := &models.LLMRequest{
		ID:     "compat-test",
		Prompt: "Test compatible mode",
	}

	_, err := provider.Complete(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, "POST", capturedMethod)
	assert.Contains(t, capturedPath, "/chat/completions")
	assert.Equal(t, "application/json", capturedContentType)
}

// TestQwenOAuth_AllModels tests that all Qwen models work with OAuth
func TestQwenOAuth_AllModels(t *testing.T) {
	qwenModels := []string{
		"qwen-max",
		"qwen-plus",
		"qwen-turbo",
		"qwen-coder-turbo",
		"qwen-long",
		"qwen-max-longcontext",
	}

	for _, model := range qwenModels {
		t.Run(model, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"id":      fmt.Sprintf("%s-test", model),
					"object":  "chat.completion",
					"created": time.Now().Unix(),
					"model":   model,
					"choices": []map[string]interface{}{
						{
							"index": 0,
							"message": map[string]interface{}{
								"role":    "assistant",
								"content": fmt.Sprintf("Response from %s", model),
							},
							"finish_reason": "stop",
						},
					},
					"usage": map[string]interface{}{
						"total_tokens": 10,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			provider := qwen.NewQwenProvider(
				"oauth-token",
				server.URL+"/compatible-mode/v1",
				model,
			)

			req := &models.LLMRequest{
				ID:     fmt.Sprintf("%s-req", model),
				Prompt: "Test model",
			}

			resp, err := provider.Complete(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Contains(t, resp.Content, model)
		})
	}
}

// TestQwenOAuth_BackgroundRefresh tests the background token refresh functionality
func TestQwenOAuth_BackgroundRefresh(t *testing.T) {
	// Test that StartBackgroundRefresh can be called without panicking
	stopChan := make(chan struct{})

	// This should not panic
	oauth_credentials.StartBackgroundRefresh(stopChan)

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop the background refresh
	close(stopChan)

	// Give it time to clean up
	time.Sleep(100 * time.Millisecond)
}

// TestQwenOAuth_RefreshStatus tests the refresh status reporting
func TestQwenOAuth_RefreshStatus(t *testing.T) {
	status := oauth_credentials.GetRefreshStatus()

	// Should always have refresh_threshold
	assert.Contains(t, status, "refresh_threshold")
	assert.NotEmpty(t, status["refresh_threshold"])
}

// TestQwenOAuth_NeedsRefreshThreshold tests the refresh threshold timing
func TestQwenOAuth_NeedsRefreshThreshold(t *testing.T) {
	// Token expiring in 5 minutes should need refresh
	expiringIn5Min := time.Now().Add(5 * time.Minute).UnixMilli()
	assert.True(t, oauth_credentials.NeedsRefresh(expiringIn5Min))

	// Token expiring in 15 minutes should NOT need refresh
	expiringIn15Min := time.Now().Add(15 * time.Minute).UnixMilli()
	assert.False(t, oauth_credentials.NeedsRefresh(expiringIn15Min))

	// Token expiring in 1 hour should NOT need refresh
	expiringIn1Hour := time.Now().Add(time.Hour).UnixMilli()
	assert.False(t, oauth_credentials.NeedsRefresh(expiringIn1Hour))

	// Zero expiry should NOT need refresh (no expiration set)
	assert.False(t, oauth_credentials.NeedsRefresh(0))
}

// TestQwenOAuth_IsExpired tests the expiration check
func TestQwenOAuth_IsExpired(t *testing.T) {
	// Already expired token
	expiredToken := time.Now().Add(-time.Hour).UnixMilli()
	assert.True(t, oauth_credentials.IsExpired(expiredToken))

	// Token expiring in future
	futureToken := time.Now().Add(time.Hour).UnixMilli()
	assert.False(t, oauth_credentials.IsExpired(futureToken))

	// Zero expiry means no expiration
	assert.False(t, oauth_credentials.IsExpired(0))
}

// TestQwenOAuth_TokenRefresher tests the token refresher
func TestQwenOAuth_TokenRefresher(t *testing.T) {
	refresher := oauth_credentials.NewTokenRefresher()
	assert.NotNil(t, refresher)
}

// TestQwenOAuth_GlobalReader tests the global credential reader
func TestQwenOAuth_GlobalReader(t *testing.T) {
	reader := oauth_credentials.GetGlobalReader()
	assert.NotNil(t, reader)

	// Should return the same instance
	reader2 := oauth_credentials.GetGlobalReader()
	assert.Equal(t, reader, reader2)
}

// TestQwenOAuth_CredentialsPath tests getting the credentials path
func TestQwenOAuth_CredentialsPath(t *testing.T) {
	path := oauth_credentials.GetQwenCredentialsPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".qwen")
	assert.Contains(t, path, "oauth_creds.json")
}
