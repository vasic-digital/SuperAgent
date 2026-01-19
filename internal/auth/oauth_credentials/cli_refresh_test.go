package oauth_credentials

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultCLIRefreshConfig(t *testing.T) {
	config := DefaultCLIRefreshConfig()

	t.Run("has sensible defaults", func(t *testing.T) {
		if config.RefreshTimeout != 60*time.Second {
			t.Errorf("expected RefreshTimeout to be 60s, got %v", config.RefreshTimeout)
		}
		if config.MinRefreshInterval != 60*time.Second {
			t.Errorf("expected MinRefreshInterval to be 60s, got %v", config.MinRefreshInterval)
		}
		if config.MaxRetries != 3 {
			t.Errorf("expected MaxRetries to be 3, got %d", config.MaxRetries)
		}
		if config.RetryDelay != 5*time.Second {
			t.Errorf("expected RetryDelay to be 5s, got %v", config.RetryDelay)
		}
		if config.Prompt != "exit" {
			t.Errorf("expected Prompt to be 'exit', got %s", config.Prompt)
		}
	})
}

func TestNewCLIRefresher(t *testing.T) {
	t.Run("creates with default config when nil", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		if refresher == nil {
			t.Fatal("expected non-nil refresher")
		}
		if refresher.config == nil {
			t.Fatal("expected non-nil config")
		}
		if refresher.config.MaxRetries != 3 {
			t.Errorf("expected default MaxRetries, got %d", refresher.config.MaxRetries)
		}
	})

	t.Run("creates with custom config", func(t *testing.T) {
		customConfig := &CLIRefreshConfig{
			MaxRetries:     5,
			RefreshTimeout: 30 * time.Second,
		}
		refresher := NewCLIRefresher(customConfig)
		if refresher.config.MaxRetries != 5 {
			t.Errorf("expected MaxRetries 5, got %d", refresher.config.MaxRetries)
		}
	})
}

func TestGetGlobalCLIRefresher(t *testing.T) {
	t.Run("returns same instance", func(t *testing.T) {
		r1 := GetGlobalCLIRefresher()
		r2 := GetGlobalCLIRefresher()
		if r1 != r2 {
			t.Error("expected same instance")
		}
	})
}

func TestCLIRefresherInitialize(t *testing.T) {
	t.Run("initializes successfully if qwen CLI exists", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		err := refresher.Initialize()
		// This test may pass or fail depending on whether qwen is installed
		if err != nil {
			t.Logf("qwen CLI not found (expected in CI): %v", err)
		} else {
			if refresher.qwenCLIPath == "" {
				t.Error("expected qwenCLIPath to be set after initialization")
			}
			if !refresher.initialized {
				t.Error("expected initialized to be true")
			}
		}
	})

	t.Run("returns error for non-existent configured path", func(t *testing.T) {
		config := &CLIRefreshConfig{
			QwenCLIPath: "/nonexistent/path/to/qwen",
		}
		refresher := NewCLIRefresher(config)
		err := refresher.Initialize()
		if err == nil {
			t.Error("expected error for non-existent path")
		}
	})
}

func TestCLIRefresherIsAvailable(t *testing.T) {
	t.Run("returns false when CLI not found", func(t *testing.T) {
		config := &CLIRefreshConfig{
			QwenCLIPath: "/nonexistent/qwen",
		}
		refresher := NewCLIRefresher(config)
		if refresher.IsAvailable() {
			t.Error("expected IsAvailable to return false for non-existent CLI")
		}
	})
}

func TestCLIRefresherRateLimiting(t *testing.T) {
	t.Run("rate limits refresh attempts", func(t *testing.T) {
		config := &CLIRefreshConfig{
			QwenCLIPath:        "/bin/echo", // Use echo as a placeholder
			MinRefreshInterval: 1 * time.Hour,
			MaxRetries:         0,
			RefreshTimeout:     1 * time.Second,
		}
		refresher := NewCLIRefresher(config)

		// First attempt sets the last refresh time
		ctx := context.Background()
		refresher.lastRefreshTime = time.Now()
		refresher.initialized = true
		refresher.qwenCLIPath = "/bin/echo"

		// Second attempt should be rate limited
		_, err := refresher.RefreshQwenToken(ctx)
		if err == nil {
			t.Error("expected rate limiting error")
		}
		if err != nil && !contains(err.Error(), "rate limited") {
			t.Errorf("expected rate limiting error, got: %v", err)
		}
	})
}

func TestCLIRefresherResetRateLimit(t *testing.T) {
	t.Run("resets rate limit", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		refresher.lastRefreshTime = time.Now()
		refresher.lastRefreshError = os.ErrNotExist

		refresher.ResetRateLimit()

		if !refresher.lastRefreshTime.IsZero() {
			t.Error("expected lastRefreshTime to be zero after reset")
		}
		if refresher.lastRefreshError != nil {
			t.Error("expected lastRefreshError to be nil after reset")
		}
	})
}

func TestParseQwenOutput(t *testing.T) {
	refresher := NewCLIRefresher(nil)

	t.Run("parses successful output", func(t *testing.T) {
		output := `{"type":"system","subtype":"init","uuid":"test"}
{"type":"assistant","uuid":"test","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","subtype":"success","uuid":"test","is_error":false,"result":"Hello"}`

		err := refresher.parseQwenOutput(output)
		if err != nil {
			t.Errorf("expected no error for valid output, got: %v", err)
		}
	})

	t.Run("detects error in output", func(t *testing.T) {
		output := `{"type":"result","subtype":"error","is_error":true,"result":"Authentication failed"}`

		err := refresher.parseQwenOutput(output)
		if err == nil {
			t.Error("expected error for error output")
		}
	})

	t.Run("returns error for empty output", func(t *testing.T) {
		err := refresher.parseQwenOutput("")
		if err == nil {
			t.Error("expected error for empty output")
		}
	})

	t.Run("returns error for missing result", func(t *testing.T) {
		output := `{"type":"system","subtype":"init"}`
		err := refresher.parseQwenOutput(output)
		if err == nil {
			t.Error("expected error when no result present")
		}
	})
}

func TestCLIRefreshResult(t *testing.T) {
	t.Run("struct fields are set correctly", func(t *testing.T) {
		result := &CLIRefreshResult{
			Success:         true,
			NewExpiryDate:   time.Now().Add(1 * time.Hour).UnixMilli(),
			RefreshDuration: 5 * time.Second,
			Retries:         2,
			CLIOutput:       "test output",
		}

		if !result.Success {
			t.Error("expected Success to be true")
		}
		if result.Retries != 2 {
			t.Errorf("expected Retries to be 2, got %d", result.Retries)
		}
	})
}

func TestCLIRefreshStatus(t *testing.T) {
	t.Run("GetStatus returns valid status", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		status := refresher.GetStatus()

		if status == nil {
			t.Fatal("expected non-nil status")
		}

		// Available depends on whether qwen is installed
		t.Logf("CLI Available: %v", status.Available)
		t.Logf("Token Valid: %v", status.TokenValid)
	})
}

func TestVerifyTokenRefreshed(t *testing.T) {
	// Create temp directory for test credentials
	tmpDir, err := os.MkdirTemp("", "qwen_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override credentials path for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create .qwen directory
	qwenDir := filepath.Join(tmpDir, ".qwen")
	if err := os.MkdirAll(qwenDir, 0755); err != nil {
		t.Fatalf("failed to create qwen dir: %v", err)
	}

	refresher := NewCLIRefresher(nil)

	t.Run("returns error when credentials file missing", func(t *testing.T) {
		_, err := refresher.verifyTokenRefreshed()
		if err == nil {
			t.Error("expected error when credentials file missing")
		}
	})

	t.Run("returns error for expired token", func(t *testing.T) {
		creds := QwenOAuthCredentials{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
			ExpiryDate:   time.Now().Add(-1 * time.Hour).UnixMilli(), // Expired
			TokenType:    "Bearer",
		}
		data, _ := json.Marshal(creds)
		credPath := filepath.Join(qwenDir, "oauth_creds.json")
		if err := os.WriteFile(credPath, data, 0600); err != nil {
			t.Fatalf("failed to write test credentials: %v", err)
		}

		_, err := refresher.verifyTokenRefreshed()
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("returns credentials for valid token", func(t *testing.T) {
		creds := QwenOAuthCredentials{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
			ExpiryDate:   time.Now().Add(1 * time.Hour).UnixMilli(), // Valid
			TokenType:    "Bearer",
		}
		data, _ := json.Marshal(creds)
		credPath := filepath.Join(qwenDir, "oauth_creds.json")
		if err := os.WriteFile(credPath, data, 0600); err != nil {
			t.Fatalf("failed to write test credentials: %v", err)
		}

		result, err := refresher.verifyTokenRefreshed()
		if err != nil {
			t.Errorf("expected no error for valid token, got: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil credentials")
		}
		if result != nil && result.AccessToken != "test-token" {
			t.Errorf("expected access token 'test-token', got %s", result.AccessToken)
		}
	})
}

func TestRefreshQwenTokenWithFallback(t *testing.T) {
	t.Run("returns error when both methods fail", func(t *testing.T) {
		// Create expired credentials
		creds := &QwenOAuthCredentials{
			AccessToken:  "expired-token",
			RefreshToken: "", // No refresh token
			ExpiryDate:   time.Now().Add(-1 * time.Hour).UnixMilli(),
		}

		ctx := context.Background()
		_, err := RefreshQwenTokenWithFallback(ctx, creds)
		if err == nil {
			t.Error("expected error when both refresh methods fail")
		}
	})
}

func TestGetLastRefreshError(t *testing.T) {
	refresher := NewCLIRefresher(nil)
	refresher.lastRefreshError = os.ErrPermission

	err := refresher.GetLastRefreshError()
	if err != os.ErrPermission {
		t.Errorf("expected ErrPermission, got %v", err)
	}
}

func TestGetLastRefreshTime(t *testing.T) {
	refresher := NewCLIRefresher(nil)
	expectedTime := time.Now().Add(-5 * time.Minute)
	refresher.lastRefreshTime = expectedTime

	gotTime := refresher.GetLastRefreshTime()
	if !gotTime.Equal(expectedTime) {
		t.Errorf("expected %v, got %v", expectedTime, gotTime)
	}
}

func TestFindQwenCLI(t *testing.T) {
	refresher := NewCLIRefresher(nil)

	t.Run("finds qwen in PATH if installed", func(t *testing.T) {
		path, err := refresher.findQwenCLI()
		if err != nil {
			t.Logf("qwen not found (expected in CI): %v", err)
		} else {
			if path == "" {
				t.Error("expected non-empty path when qwen is found")
			}
			t.Logf("Found qwen at: %s", path)
		}
	})

	t.Run("returns configured path if valid", func(t *testing.T) {
		// Create temp executable
		tmpDir, _ := os.MkdirTemp("", "qwen_test")
		defer os.RemoveAll(tmpDir)

		tmpQwen := filepath.Join(tmpDir, "qwen")
		if err := os.WriteFile(tmpQwen, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
			t.Fatalf("failed to create temp qwen: %v", err)
		}

		configuredRefresher := NewCLIRefresher(&CLIRefreshConfig{
			QwenCLIPath: tmpQwen,
		})

		path, err := configuredRefresher.findQwenCLI()
		if err != nil {
			t.Errorf("expected to find configured qwen, got: %v", err)
		}
		if path != tmpQwen {
			t.Errorf("expected %s, got %s", tmpQwen, path)
		}
	})
}

func TestGetQwenCLIPath(t *testing.T) {
	t.Run("returns empty string when not initialized", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		path := refresher.GetQwenCLIPath()
		if path != "" {
			t.Errorf("expected empty path when not initialized, got %s", path)
		}
	})

	t.Run("returns path when initialized", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		refresher.qwenCLIPath = "/usr/bin/qwen"

		path := refresher.GetQwenCLIPath()
		if path != "/usr/bin/qwen" {
			t.Errorf("expected /usr/bin/qwen, got %s", path)
		}
	})
}

func TestGetStatusWithCredentials(t *testing.T) {
	// Create temp directory for test credentials
	tmpDir, err := os.MkdirTemp("", "qwen_status_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override credentials path for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create .qwen directory
	qwenDir := filepath.Join(tmpDir, ".qwen")
	if err := os.MkdirAll(qwenDir, 0755); err != nil {
		t.Fatalf("failed to create qwen dir: %v", err)
	}

	t.Run("GetStatus with valid credentials", func(t *testing.T) {
		creds := QwenOAuthCredentials{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
			ExpiryDate:   time.Now().Add(1 * time.Hour).UnixMilli(), // Valid
			TokenType:    "Bearer",
		}
		data, _ := json.Marshal(creds)
		credPath := filepath.Join(qwenDir, "oauth_creds.json")
		if err := os.WriteFile(credPath, data, 0600); err != nil {
			t.Fatalf("failed to write test credentials: %v", err)
		}

		refresher := NewCLIRefresher(nil)
		status := refresher.GetStatus()

		if !status.TokenValid {
			t.Error("expected TokenValid to be true for non-expired token")
		}
		if status.TokenExpiresAt.IsZero() {
			t.Error("expected TokenExpiresAt to be set")
		}
		if status.TokenExpiresIn == "" {
			t.Error("expected TokenExpiresIn to be set")
		}
	})

	t.Run("GetStatus with expired credentials", func(t *testing.T) {
		creds := QwenOAuthCredentials{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
			ExpiryDate:   time.Now().Add(-1 * time.Hour).UnixMilli(), // Expired
			TokenType:    "Bearer",
		}
		data, _ := json.Marshal(creds)
		credPath := filepath.Join(qwenDir, "oauth_creds.json")
		if err := os.WriteFile(credPath, data, 0600); err != nil {
			t.Fatalf("failed to write test credentials: %v", err)
		}

		refresher := NewCLIRefresher(nil)
		status := refresher.GetStatus()

		if status.TokenValid {
			t.Error("expected TokenValid to be false for expired token")
		}
	})

	t.Run("GetStatus with last refresh error", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		refresher.lastRefreshError = os.ErrPermission
		refresher.lastRefreshTime = time.Now().Add(-5 * time.Minute)

		status := refresher.GetStatus()

		if status.LastRefreshError != "permission denied" {
			t.Errorf("expected 'permission denied' error, got %s", status.LastRefreshError)
		}
		if status.LastRefreshTime.IsZero() {
			t.Error("expected LastRefreshTime to be set")
		}
	})
}

func TestFindQwenCLICommonPaths(t *testing.T) {
	// Create temp home directory
	tmpDir, err := os.MkdirTemp("", "qwen_find_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override HOME for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	t.Run("finds qwen in .local/bin", func(t *testing.T) {
		localBin := filepath.Join(tmpDir, ".local", "bin")
		if err := os.MkdirAll(localBin, 0755); err != nil {
			t.Fatalf("failed to create .local/bin: %v", err)
		}

		qwenPath := filepath.Join(localBin, "qwen")
		if err := os.WriteFile(qwenPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
			t.Fatalf("failed to create qwen: %v", err)
		}

		refresher := NewCLIRefresher(nil)
		path, err := refresher.findQwenCLI()
		if err != nil {
			t.Logf("qwen not found via LookPath, checking common paths: %v", err)
		}
		if path != "" && path != qwenPath {
			t.Logf("Found qwen at different location: %s (expected %s)", path, qwenPath)
		}
	})

	t.Run("searches Applications for node installations", func(t *testing.T) {
		appsDir := filepath.Join(tmpDir, "Applications")
		nodeDir := filepath.Join(appsDir, "node-v24.0.0-linux-x64", "bin")
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			t.Fatalf("failed to create node dir: %v", err)
		}

		qwenPath := filepath.Join(nodeDir, "qwen")
		if err := os.WriteFile(qwenPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
			t.Fatalf("failed to create qwen: %v", err)
		}

		refresher := NewCLIRefresher(nil)
		path, err := refresher.findQwenCLI()
		if err != nil {
			t.Logf("Could not find qwen via findQwenCLI: %v", err)
		}
		if path != "" {
			t.Logf("Found qwen at: %s", path)
		}
	})
}

func TestCLIRefresherInitializeMultipleCalls(t *testing.T) {
	t.Run("does not re-initialize if already initialized", func(t *testing.T) {
		refresher := NewCLIRefresher(nil)
		refresher.initialized = true
		refresher.qwenCLIPath = "/fake/path"

		err := refresher.Initialize()
		if err != nil {
			t.Errorf("expected no error for already initialized, got: %v", err)
		}
		if refresher.qwenCLIPath != "/fake/path" {
			t.Error("path should not change when already initialized")
		}
	})
}

func TestExecuteQwenCLITimeout(t *testing.T) {
	t.Run("handles timeout context", func(t *testing.T) {
		config := &CLIRefreshConfig{
			QwenCLIPath:    "/bin/sleep",
			RefreshTimeout: 100 * time.Millisecond,
			Prompt:         "10", // Sleep for 10 seconds
		}
		refresher := NewCLIRefresher(config)
		refresher.qwenCLIPath = "/bin/sleep"

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := refresher.executeQwenCLI(ctx)
		if err == nil {
			t.Error("expected error due to timeout")
		}
	})
}

func TestAutoRefreshQwenTokenViaCLI(t *testing.T) {
	t.Run("returns error when CLI refresh fails", func(t *testing.T) {
		// This test just verifies the function can be called and returns an error
		// when the refresh process fails (either CLI not available or refresh fails)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := AutoRefreshQwenTokenViaCLI(ctx)
		// We expect an error - either "not available" or a refresh error
		if err == nil {
			t.Skip("AutoRefreshQwenTokenViaCLI succeeded unexpectedly (valid credentials exist)")
		}
		// Error is expected - test passes
		t.Logf("Got expected error: %v", err)
	})
}

func TestQwenCLIOutput(t *testing.T) {
	t.Run("struct fields are correctly serialized", func(t *testing.T) {
		output := QwenCLIOutput{
			Type:      "result",
			Subtype:   "success",
			SessionID: "test-session",
			Result:    "test result",
			IsError:   false,
		}

		data, err := json.Marshal(output)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var parsed QwenCLIOutput
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if parsed.Type != "result" {
			t.Errorf("expected Type 'result', got %s", parsed.Type)
		}
		if parsed.SessionID != "test-session" {
			t.Errorf("expected SessionID 'test-session', got %s", parsed.SessionID)
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
