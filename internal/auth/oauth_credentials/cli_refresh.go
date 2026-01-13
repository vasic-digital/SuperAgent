// Package oauth_credentials provides CLI-based OAuth2 token refresh functionality
// for Qwen Code CLI, enabling automatic token refresh by invoking the CLI.
package oauth_credentials

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CLIRefreshConfig holds configuration for CLI-based token refresh
type CLIRefreshConfig struct {
	// QwenCLIPath is the path to the qwen CLI executable (auto-detected if empty)
	QwenCLIPath string

	// RefreshTimeout is the maximum time to wait for CLI refresh
	RefreshTimeout time.Duration

	// MinRefreshInterval is the minimum time between CLI refresh attempts
	MinRefreshInterval time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration

	// Prompt is the simple prompt to send to qwen (minimal to reduce token usage)
	Prompt string
}

// DefaultCLIRefreshConfig returns sensible defaults for CLI refresh
func DefaultCLIRefreshConfig() *CLIRefreshConfig {
	return &CLIRefreshConfig{
		QwenCLIPath:        "", // Auto-detect
		RefreshTimeout:     60 * time.Second,
		MinRefreshInterval: 60 * time.Second,
		MaxRetries:         3,
		RetryDelay:         5 * time.Second,
		Prompt:             "exit", // Minimal prompt
	}
}

// CLIRefresher handles CLI-based token refresh for Qwen
type CLIRefresher struct {
	config           *CLIRefreshConfig
	mu               sync.Mutex
	lastRefreshTime  time.Time
	lastRefreshError error
	qwenCLIPath      string
	initialized      bool
}

// CLIRefreshResult holds the result of a CLI refresh operation
type CLIRefreshResult struct {
	Success         bool          `json:"success"`
	NewExpiryDate   int64         `json:"new_expiry_date,omitempty"`
	RefreshDuration time.Duration `json:"refresh_duration"`
	Error           string        `json:"error,omitempty"`
	Retries         int           `json:"retries"`
	CLIOutput       string        `json:"cli_output,omitempty"`
}

// QwenCLIOutput represents the JSON output from qwen CLI
type QwenCLIOutput struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Result    string `json:"result,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// NewCLIRefresher creates a new CLI-based token refresher
func NewCLIRefresher(config *CLIRefreshConfig) *CLIRefresher {
	if config == nil {
		config = DefaultCLIRefreshConfig()
	}
	return &CLIRefresher{
		config: config,
	}
}

// Global CLI refresher singleton
var (
	globalCLIRefresher     *CLIRefresher
	globalCLIRefresherOnce sync.Once
)

// GetGlobalCLIRefresher returns the global CLI refresher instance
func GetGlobalCLIRefresher() *CLIRefresher {
	globalCLIRefresherOnce.Do(func() {
		globalCLIRefresher = NewCLIRefresher(nil)
	})
	return globalCLIRefresher
}

// Initialize discovers the qwen CLI path and validates it
func (cr *CLIRefresher) Initialize() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.initialized {
		return nil
	}

	// Find qwen CLI
	cliPath, err := cr.findQwenCLI()
	if err != nil {
		return fmt.Errorf("failed to find qwen CLI: %w", err)
	}

	cr.qwenCLIPath = cliPath
	cr.initialized = true
	return nil
}

// findQwenCLI searches for the qwen CLI executable
func (cr *CLIRefresher) findQwenCLI() (string, error) {
	// Use configured path if provided
	if cr.config.QwenCLIPath != "" {
		if _, err := os.Stat(cr.config.QwenCLIPath); err == nil {
			return cr.config.QwenCLIPath, nil
		}
		return "", fmt.Errorf("configured qwen CLI path not found: %s", cr.config.QwenCLIPath)
	}

	// Try to find qwen in PATH
	path, err := exec.LookPath("qwen")
	if err == nil {
		return path, nil
	}

	// Check common installation locations
	homeDir, _ := os.UserHomeDir()
	commonPaths := []string{
		filepath.Join(homeDir, ".local", "bin", "qwen"),
		filepath.Join(homeDir, "Applications", "node-v24.12.0-linux-x64", "bin", "qwen"),
		"/usr/local/bin/qwen",
		"/usr/bin/qwen",
	}

	// Also check for node-based installations
	if homeDir != "" {
		// Check for any node installation in Applications
		appDir := filepath.Join(homeDir, "Applications")
		if entries, err := os.ReadDir(appDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "node-") {
					nodePath := filepath.Join(appDir, entry.Name(), "bin", "qwen")
					commonPaths = append(commonPaths, nodePath)
				}
			}
		}
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("qwen CLI not found in PATH or common locations")
}

// GetQwenCLIPath returns the detected qwen CLI path
func (cr *CLIRefresher) GetQwenCLIPath() string {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.qwenCLIPath
}

// IsAvailable checks if qwen CLI is available for use
func (cr *CLIRefresher) IsAvailable() bool {
	if err := cr.Initialize(); err != nil {
		return false
	}
	return cr.qwenCLIPath != ""
}

// RefreshQwenToken refreshes the Qwen OAuth token by invoking the qwen CLI
func (cr *CLIRefresher) RefreshQwenToken(ctx context.Context) (*CLIRefreshResult, error) {
	startTime := time.Now()
	result := &CLIRefreshResult{}

	// Initialize if needed
	if err := cr.Initialize(); err != nil {
		result.Error = err.Error()
		result.RefreshDuration = time.Since(startTime)
		return result, err
	}

	// Check rate limiting
	cr.mu.Lock()
	if time.Since(cr.lastRefreshTime) < cr.config.MinRefreshInterval {
		cr.mu.Unlock()
		err := fmt.Errorf("refresh rate limited: last attempt was %v ago (min interval: %v)",
			time.Since(cr.lastRefreshTime), cr.config.MinRefreshInterval)
		result.Error = err.Error()
		result.RefreshDuration = time.Since(startTime)
		return result, err
	}
	cr.lastRefreshTime = time.Now()
	cr.mu.Unlock()

	// Attempt refresh with retries
	var lastErr error
	for attempt := 0; attempt <= cr.config.MaxRetries; attempt++ {
		result.Retries = attempt

		if attempt > 0 {
			select {
			case <-ctx.Done():
				result.Error = ctx.Err().Error()
				result.RefreshDuration = time.Since(startTime)
				return result, ctx.Err()
			case <-time.After(cr.config.RetryDelay):
			}
		}

		output, err := cr.executeQwenCLI(ctx)
		result.CLIOutput = output

		if err != nil {
			lastErr = err
			continue
		}

		// Verify token was refreshed by checking the credentials file
		creds, err := cr.verifyTokenRefreshed()
		if err != nil {
			lastErr = err
			continue
		}

		// Success
		result.Success = true
		result.NewExpiryDate = creds.ExpiryDate
		result.RefreshDuration = time.Since(startTime)
		return result, nil
	}

	result.Error = fmt.Sprintf("all %d attempts failed: %v", cr.config.MaxRetries+1, lastErr)
	result.RefreshDuration = time.Since(startTime)
	cr.mu.Lock()
	cr.lastRefreshError = lastErr
	cr.mu.Unlock()
	return result, lastErr
}

// executeQwenCLI runs the qwen CLI with a minimal prompt
func (cr *CLIRefresher) executeQwenCLI(ctx context.Context) (string, error) {
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, cr.config.RefreshTimeout)
	defer cancel()

	// Build command: qwen "exit" -o json --max-session-turns 1
	cmd := exec.CommandContext(execCtx, cr.qwenCLIPath,
		cr.config.Prompt,
		"-o", "json",
		"--max-session-turns", "1",
	)

	// Set up environment
	cmd.Env = append(os.Environ(),
		"TERM=dumb",            // Disable terminal features
		"NO_COLOR=1",           // Disable colors
		"QWEN_NO_TELEMETRY=1",  // Disable telemetry
	)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	// Combine output for logging
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR: " + stderr.String()
	}

	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			return output, fmt.Errorf("qwen CLI timed out after %v", cr.config.RefreshTimeout)
		}
		return output, fmt.Errorf("qwen CLI failed: %w (output: %s)", err, output)
	}

	// Parse JSON output to verify success
	if err := cr.parseQwenOutput(stdout.String()); err != nil {
		return output, fmt.Errorf("failed to parse qwen output: %w", err)
	}

	return output, nil
}

// parseQwenOutput parses the JSON output from qwen CLI
func (cr *CLIRefresher) parseQwenOutput(output string) error {
	// The output is NDJSON (newline-delimited JSON)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("empty output from qwen CLI")
	}

	var hasResult bool
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var msg QwenCLIOutput
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // Skip unparseable lines
		}

		// Check for errors
		if msg.IsError {
			return fmt.Errorf("qwen CLI reported error: %s", msg.Result)
		}

		// Check for successful result
		if msg.Type == "result" && msg.Subtype == "success" {
			hasResult = true
		}
	}

	if !hasResult {
		return fmt.Errorf("no successful result in qwen CLI output")
	}

	return nil
}

// verifyTokenRefreshed reads the credentials file and verifies the token is valid
func (cr *CLIRefresher) verifyTokenRefreshed() (*QwenOAuthCredentials, error) {
	credPath := GetQwenCredentialsPath()
	if credPath == "" {
		return nil, fmt.Errorf("unable to determine credentials path")
	}

	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds QwenOAuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	// Verify token is valid
	if creds.AccessToken == "" {
		return nil, fmt.Errorf("no access token in credentials file")
	}

	if IsExpired(creds.ExpiryDate) {
		return nil, fmt.Errorf("token still expired after CLI refresh")
	}

	return &creds, nil
}

// AutoRefreshQwenTokenViaCLI attempts CLI-based refresh when standard refresh fails
func AutoRefreshQwenTokenViaCLI(ctx context.Context) (*QwenOAuthCredentials, error) {
	refresher := GetGlobalCLIRefresher()

	if !refresher.IsAvailable() {
		return nil, fmt.Errorf("qwen CLI not available for token refresh")
	}

	result, err := refresher.RefreshQwenToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("CLI refresh failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("CLI refresh unsuccessful: %s", result.Error)
	}

	// Read and return the refreshed credentials
	credPath := GetQwenCredentialsPath()
	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read refreshed credentials: %w", err)
	}

	var creds QwenOAuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse refreshed credentials: %w", err)
	}

	return &creds, nil
}

// RefreshQwenTokenWithFallback attempts standard refresh first, then falls back to CLI
func RefreshQwenTokenWithFallback(ctx context.Context, creds *QwenOAuthCredentials) (*QwenOAuthCredentials, error) {
	// Try standard refresh first
	refreshedCreds, err := AutoRefreshQwenToken(creds)
	if err == nil && refreshedCreds != nil && !IsExpired(refreshedCreds.ExpiryDate) {
		return refreshedCreds, nil
	}

	// Standard refresh failed, try CLI refresh
	cliCreds, cliErr := AutoRefreshQwenTokenViaCLI(ctx)
	if cliErr != nil {
		// Return original error if CLI also fails
		if err != nil {
			return nil, fmt.Errorf("both standard and CLI refresh failed: standard: %v, CLI: %v", err, cliErr)
		}
		return nil, fmt.Errorf("CLI refresh failed: %w", cliErr)
	}

	return cliCreds, nil
}

// GetLastRefreshError returns the last refresh error
func (cr *CLIRefresher) GetLastRefreshError() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.lastRefreshError
}

// GetLastRefreshTime returns the last refresh attempt time
func (cr *CLIRefresher) GetLastRefreshTime() time.Time {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.lastRefreshTime
}

// ResetRateLimit resets the rate limiting (useful for testing)
func (cr *CLIRefresher) ResetRateLimit() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.lastRefreshTime = time.Time{}
	cr.lastRefreshError = nil
}

// CLIRefreshStatus provides status information about CLI refresh capability
type CLIRefreshStatus struct {
	Available        bool      `json:"available"`
	QwenCLIPath      string    `json:"qwen_cli_path,omitempty"`
	LastRefreshTime  time.Time `json:"last_refresh_time,omitempty"`
	LastRefreshError string    `json:"last_refresh_error,omitempty"`
	TokenValid       bool      `json:"token_valid"`
	TokenExpiresAt   time.Time `json:"token_expires_at,omitempty"`
	TokenExpiresIn   string    `json:"token_expires_in,omitempty"`
}

// GetStatus returns the current status of the CLI refresher
func (cr *CLIRefresher) GetStatus() *CLIRefreshStatus {
	status := &CLIRefreshStatus{
		Available: cr.IsAvailable(),
	}

	cr.mu.Lock()
	status.QwenCLIPath = cr.qwenCLIPath
	status.LastRefreshTime = cr.lastRefreshTime
	if cr.lastRefreshError != nil {
		status.LastRefreshError = cr.lastRefreshError.Error()
	}
	cr.mu.Unlock()

	// Check current token status
	credPath := GetQwenCredentialsPath()
	if data, err := os.ReadFile(credPath); err == nil {
		var creds QwenOAuthCredentials
		if json.Unmarshal(data, &creds) == nil {
			status.TokenValid = !IsExpired(creds.ExpiryDate)
			if creds.ExpiryDate > 0 {
				expiresAt := time.UnixMilli(creds.ExpiryDate)
				status.TokenExpiresAt = expiresAt
				status.TokenExpiresIn = time.Until(expiresAt).Round(time.Second).String()
			}
		}
	}

	return status
}
