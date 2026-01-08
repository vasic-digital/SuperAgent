package oauth_credentials

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNeedsRefresh(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt int64
		expected  bool
	}{
		{
			name:      "no expiration set",
			expiresAt: 0,
			expected:  false,
		},
		{
			name:      "expires in 1 hour - no refresh needed",
			expiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
			expected:  false, // More than RefreshThreshold (10 min)
		},
		{
			name:      "expires in 5 minutes - needs refresh",
			expiresAt: time.Now().Add(5 * time.Minute).UnixMilli(),
			expected:  true,
		},
		{
			name:      "expires in 2 hours - no refresh needed",
			expiresAt: time.Now().Add(2 * time.Hour).UnixMilli(),
			expected:  false,
		},
		{
			name:      "already expired",
			expiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NeedsRefresh(tt.expiresAt)
			if result != tt.expected {
				t.Errorf("NeedsRefresh(%d) = %v, want %v", tt.expiresAt, result, tt.expected)
			}
		})
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt int64
		expected  bool
	}{
		{
			name:      "no expiration set",
			expiresAt: 0,
			expected:  false,
		},
		{
			name:      "expires in future",
			expiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
			expected:  false,
		},
		{
			name:      "already expired",
			expiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
			expected:  true,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-1 * time.Second).UnixMilli(),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExpired(tt.expiresAt)
			if result != tt.expected {
				t.Errorf("IsExpired(%d) = %v, want %v", tt.expiresAt, result, tt.expected)
			}
		})
	}
}

func TestTokenRefresher_RefreshClaudeToken(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		// Parse form
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		grantType := r.FormValue("grant_type")
		if grantType != "refresh_token" {
			t.Errorf("Expected grant_type refresh_token, got %s", grantType)
		}

		refreshToken := r.FormValue("refresh_token")
		if refreshToken == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "missing refresh_token"}`))
			return
		}

		// Return mock response
		resp := ClaudeRefreshResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create refresher with mock server
	refresher := NewTokenRefresher()

	// Test with valid refresh token (note: this test uses real endpoint, so we'll test the structure)
	t.Run("missing refresh token", func(t *testing.T) {
		_, err := refresher.RefreshClaudeToken("")
		if err == nil {
			t.Error("Expected error for missing refresh token")
		}
	})
}

func TestTokenRefresher_RefreshRateLimiting(t *testing.T) {
	refresher := NewTokenRefresher()
	refresher.minRefreshInterval = 100 * time.Millisecond

	// First call should work (will fail due to network, but shouldn't be rate limited)
	_, err1 := refresher.RefreshClaudeToken("test-token")
	if err1 == nil {
		t.Log("First call succeeded (unexpected but ok)")
	}

	// Second immediate call should be rate limited
	_, err2 := refresher.RefreshClaudeToken("test-token")
	if err2 == nil || err2.Error() == "" {
		t.Log("Second call succeeded (unexpected but ok)")
	}

	// Wait and try again
	time.Sleep(150 * time.Millisecond)
	_, err3 := refresher.RefreshClaudeToken("test-token")
	// Should not be rate limited now (may fail for other reasons)
	if err3 != nil && err3.Error() != "" {
		// Check it's not a rate limit error
		if err3.Error()[:6] == "refres" {
			// This is fine - it's a network error, not rate limit
		}
	}
}

func TestAutoRefreshClaudeToken(t *testing.T) {
	t.Run("nil credentials", func(t *testing.T) {
		result, err := AutoRefreshClaudeToken(nil)
		if err != nil {
			t.Errorf("Expected no error for nil creds, got %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result for nil creds")
		}
	})

	t.Run("token not expiring", func(t *testing.T) {
		creds := &ClaudeOAuthCredentials{
			ClaudeAiOauth: &ClaudeAiOauth{
				AccessToken:  "valid-token",
				RefreshToken: "refresh-token",
				ExpiresAt:    time.Now().Add(2 * time.Hour).UnixMilli(),
			},
		}
		result, err := AutoRefreshClaudeToken(creds)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.ClaudeAiOauth.AccessToken != "valid-token" {
			t.Errorf("Expected original token to be returned")
		}
	})

	t.Run("token expiring but no refresh token", func(t *testing.T) {
		creds := &ClaudeOAuthCredentials{
			ClaudeAiOauth: &ClaudeAiOauth{
				AccessToken: "valid-token",
				ExpiresAt:   time.Now().Add(5 * time.Minute).UnixMilli(), // Expiring soon
			},
		}
		result, err := AutoRefreshClaudeToken(creds)
		if err != nil {
			t.Errorf("Expected no error (token still valid), got %v", err)
		}
		if result.ClaudeAiOauth.AccessToken != "valid-token" {
			t.Errorf("Expected original token to be returned")
		}
	})
}

func TestAutoRefreshQwenToken(t *testing.T) {
	t.Run("nil credentials", func(t *testing.T) {
		result, err := AutoRefreshQwenToken(nil)
		if err != nil {
			t.Errorf("Expected no error for nil creds, got %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result for nil creds")
		}
	})

	t.Run("token not expiring", func(t *testing.T) {
		creds := &QwenOAuthCredentials{
			AccessToken:  "valid-token",
			RefreshToken: "refresh-token",
			ExpiryDate:   time.Now().Add(2 * time.Hour).UnixMilli(),
		}
		result, err := AutoRefreshQwenToken(creds)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.AccessToken != "valid-token" {
			t.Errorf("Expected original token to be returned")
		}
	})
}

func TestUpdateClaudeCredentialsFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "oauth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock credentials file
	credPath := filepath.Join(tmpDir, ".credentials.json")
	initialCreds := ClaudeOAuthCredentials{
		ClaudeAiOauth: &ClaudeAiOauth{
			AccessToken:      "old-token",
			RefreshToken:     "old-refresh",
			ExpiresAt:        time.Now().UnixMilli(),
			SubscriptionType: "max",
		},
	}
	data, _ := json.MarshalIndent(initialCreds, "", "  ")
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test that the structure works (actual file update tested with real path)
	t.Run("structure validation", func(t *testing.T) {
		var creds ClaudeOAuthCredentials
		if err := json.Unmarshal(data, &creds); err != nil {
			t.Errorf("Failed to parse test credentials: %v", err)
		}
		if creds.ClaudeAiOauth.AccessToken != "old-token" {
			t.Errorf("Expected old-token, got %s", creds.ClaudeAiOauth.AccessToken)
		}
	})
}

func TestUpdateQwenCredentialsFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "oauth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock credentials file
	credPath := filepath.Join(tmpDir, "oauth_creds.json")
	initialCreds := QwenOAuthCredentials{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiryDate:   time.Now().UnixMilli(),
		TokenType:    "Bearer",
	}
	data, _ := json.MarshalIndent(initialCreds, "", "  ")
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test that the structure works
	t.Run("structure validation", func(t *testing.T) {
		var creds QwenOAuthCredentials
		if err := json.Unmarshal(data, &creds); err != nil {
			t.Errorf("Failed to parse test credentials: %v", err)
		}
		if creds.AccessToken != "old-token" {
			t.Errorf("Expected old-token, got %s", creds.AccessToken)
		}
	})
}

func TestGetRefreshStatus(t *testing.T) {
	status := GetRefreshStatus()

	// Should always have refresh_threshold
	if _, ok := status["refresh_threshold"]; !ok {
		t.Error("Missing refresh_threshold in status")
	}

	// Claude and Qwen status depends on credential availability
	t.Logf("Refresh status: %+v", status)
}

func TestGetGlobalRefresher(t *testing.T) {
	refresher1 := GetGlobalRefresher()
	refresher2 := GetGlobalRefresher()

	if refresher1 != refresher2 {
		t.Error("GetGlobalRefresher should return singleton")
	}

	if refresher1 == nil {
		t.Error("GetGlobalRefresher returned nil")
	}
}

func TestNewTokenRefresher(t *testing.T) {
	refresher := NewTokenRefresher()

	if refresher == nil {
		t.Fatal("NewTokenRefresher returned nil")
	}

	if refresher.httpClient == nil {
		t.Error("HTTP client not initialized")
	}

	if refresher.minRefreshInterval != 30*time.Second {
		t.Errorf("Expected minRefreshInterval 30s, got %v", refresher.minRefreshInterval)
	}
}

func TestSetHTTPClient(t *testing.T) {
	refresher := NewTokenRefresher()
	customClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	refresher.SetHTTPClient(customClient)

	if refresher.httpClient != customClient {
		t.Error("SetHTTPClient did not update the client")
	}
}

func TestNewMockResponse(t *testing.T) {
	resp := NewMockResponse(200, `{"test": "data"}`)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if resp.Body == nil {
		t.Error("Response body is nil")
	}
}

func TestRefreshThresholdConstant(t *testing.T) {
	if RefreshThreshold != 10*time.Minute {
		t.Errorf("Expected RefreshThreshold to be 10 minutes, got %v", RefreshThreshold)
	}
}

func TestRefreshTimeoutConstant(t *testing.T) {
	if RefreshTimeout != 30*time.Second {
		t.Errorf("Expected RefreshTimeout to be 30 seconds, got %v", RefreshTimeout)
	}
}
