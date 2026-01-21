// Package oauth_credentials provides OAuth2 token refresh functionality
// for Claude Code and Qwen Code CLI agents.
package oauth_credentials

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Token refresh configuration
const (
	// RefreshThreshold is the time before expiration when we proactively refresh
	RefreshThreshold = 10 * time.Minute

	// Claude OAuth endpoints - public endpoint URLs, not credentials
	ClaudeTokenEndpoint = "https://console.anthropic.com/v1/oauth/token" // #nosec G101 - public OAuth endpoint URL

	// Qwen OAuth endpoints (Qwen Code / Tongyi Lingma) - public endpoint URLs, not credentials
	QwenTokenEndpoint = "https://chat.qwen.ai/api/v1/oauth2/token" // #nosec G101 - public OAuth endpoint URL

	// Qwen OAuth client ID (public client, used with PKCE)
	QwenOAuthClientID = "f0304373b74a44d2b584a3fb70ca9e56"

	// HTTP client timeout for token refresh
	RefreshTimeout = 30 * time.Second
)

// TokenRefresher handles automatic token refresh for OAuth credentials
type TokenRefresher struct {
	mu         sync.Mutex
	httpClient *http.Client
	// Track last refresh attempt to avoid hammering endpoints
	lastClaudeRefresh time.Time
	lastQwenRefresh   time.Time
	// Minimum interval between refresh attempts
	minRefreshInterval time.Duration
}

// NewTokenRefresher creates a new token refresher
func NewTokenRefresher() *TokenRefresher {
	return &TokenRefresher{
		httpClient: &http.Client{
			Timeout: RefreshTimeout,
		},
		minRefreshInterval: 30 * time.Second,
	}
}

// ClaudeRefreshResponse represents the response from Claude token refresh
type ClaudeRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
}

// QwenRefreshResponse represents the response from Qwen token refresh
type QwenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
}

// NeedsRefresh checks if a token needs refreshing based on expiration time
func NeedsRefresh(expiresAt int64) bool {
	if expiresAt == 0 {
		return false // No expiration set, assume valid
	}
	expirationTime := time.UnixMilli(expiresAt)
	return time.Until(expirationTime) < RefreshThreshold
}

// IsExpired checks if a token is already expired
func IsExpired(expiresAt int64) bool {
	if expiresAt == 0 {
		return false // No expiration set, assume valid
	}
	return time.Now().UnixMilli() >= expiresAt
}

// RefreshClaudeToken attempts to refresh the Claude OAuth token
func (tr *TokenRefresher) RefreshClaudeToken(refreshToken string) (*ClaudeRefreshResponse, error) {
	tr.mu.Lock()
	if time.Since(tr.lastClaudeRefresh) < tr.minRefreshInterval {
		tr.mu.Unlock()
		return nil, fmt.Errorf("refresh rate limited: last attempt was %v ago", time.Since(tr.lastClaudeRefresh))
	}
	tr.lastClaudeRefresh = time.Now()
	tr.mu.Unlock()

	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Prepare refresh request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", ClaudeTokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := tr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var refreshResp ClaudeRefreshResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	return &refreshResp, nil
}

// RefreshQwenToken attempts to refresh the Qwen OAuth token
func (tr *TokenRefresher) RefreshQwenToken(refreshToken string, resourceURL string) (*QwenRefreshResponse, error) {
	tr.mu.Lock()
	if time.Since(tr.lastQwenRefresh) < tr.minRefreshInterval {
		tr.mu.Unlock()
		return nil, fmt.Errorf("refresh rate limited: last attempt was %v ago", time.Since(tr.lastQwenRefresh))
	}
	tr.lastQwenRefresh = time.Now()
	tr.mu.Unlock()

	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Use the standard Qwen OAuth endpoint
	// Note: resourceURL is used by Qwen for the DashScope API, not for token refresh
	tokenEndpoint := QwenTokenEndpoint

	// Prepare refresh request - Qwen requires client_id for token refresh
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", QwenOAuthClientID)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := tr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	// Handle HTTP 400 specifically - indicates refresh token is expired/invalid
	if resp.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("refresh token expired or invalid (HTTP 400): re-authentication required via Qwen Code CLI")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var refreshResp QwenRefreshResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	return &refreshResp, nil
}

// UpdateClaudeCredentialsFile updates the Claude credentials file with new tokens
func UpdateClaudeCredentialsFile(newAccessToken, newRefreshToken string, expiresIn int64) error {
	credPath := GetClaudeCredentialsPath()
	if credPath == "" {
		return fmt.Errorf("unable to determine credentials path")
	}

	// Read existing file
	data, err := os.ReadFile(credPath)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds ClaudeOAuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if creds.ClaudeAiOauth == nil {
		return fmt.Errorf("no OAuth credentials in file")
	}

	// Update tokens
	creds.ClaudeAiOauth.AccessToken = newAccessToken
	if newRefreshToken != "" {
		creds.ClaudeAiOauth.RefreshToken = newRefreshToken
	}
	// ExpiresIn is in seconds, convert to milliseconds timestamp
	creds.ClaudeAiOauth.ExpiresAt = time.Now().UnixMilli() + (expiresIn * 1000)

	// Write back to file
	updatedData, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated credentials: %w", err)
	}

	if err := os.WriteFile(credPath, updatedData, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// UpdateQwenCredentialsFile updates the Qwen credentials file with new tokens
func UpdateQwenCredentialsFile(newAccessToken, newRefreshToken, newIDToken string, expiresIn int64) error {
	credPath := GetQwenCredentialsPath()
	if credPath == "" {
		return fmt.Errorf("unable to determine credentials path")
	}

	// Read existing file
	data, err := os.ReadFile(credPath)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds QwenOAuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return fmt.Errorf("failed to parse credentials file: %w", err)
	}

	// Update tokens
	creds.AccessToken = newAccessToken
	if newRefreshToken != "" {
		creds.RefreshToken = newRefreshToken
	}
	if newIDToken != "" {
		creds.IDToken = newIDToken
	}
	// ExpiresIn is in seconds, convert to milliseconds timestamp
	creds.ExpiryDate = time.Now().UnixMilli() + (expiresIn * 1000)

	// Write back to file
	updatedData, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated credentials: %w", err)
	}

	if err := os.WriteFile(credPath, updatedData, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// Global token refresher singleton
var (
	globalRefresher     *TokenRefresher
	globalRefresherOnce sync.Once
)

// GetGlobalRefresher returns the global token refresher instance
func GetGlobalRefresher() *TokenRefresher {
	globalRefresherOnce.Do(func() {
		globalRefresher = NewTokenRefresher()
	})
	return globalRefresher
}

// AutoRefreshClaudeToken checks if Claude token needs refresh and refreshes if necessary
func AutoRefreshClaudeToken(creds *ClaudeOAuthCredentials) (*ClaudeOAuthCredentials, error) {
	if creds == nil || creds.ClaudeAiOauth == nil {
		return creds, nil
	}

	// Check if refresh is needed
	if !NeedsRefresh(creds.ClaudeAiOauth.ExpiresAt) {
		return creds, nil
	}

	// Check if we have a refresh token
	if creds.ClaudeAiOauth.RefreshToken == "" {
		if IsExpired(creds.ClaudeAiOauth.ExpiresAt) {
			return nil, fmt.Errorf("token expired and no refresh token available")
		}
		return creds, nil // Token expiring soon but no refresh token, return as-is
	}

	// Attempt refresh
	refresher := GetGlobalRefresher()
	resp, err := refresher.RefreshClaudeToken(creds.ClaudeAiOauth.RefreshToken)
	if err != nil {
		// If refresh fails but token isn't expired yet, return existing token
		if !IsExpired(creds.ClaudeAiOauth.ExpiresAt) {
			return creds, nil
		}
		return nil, fmt.Errorf("failed to refresh expired token: %w", err)
	}

	// Update credentials file
	newRefreshToken := resp.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = creds.ClaudeAiOauth.RefreshToken // Keep existing if not returned
	}

	if err := UpdateClaudeCredentialsFile(resp.AccessToken, newRefreshToken, resp.ExpiresIn); err != nil {
		// Log error but continue with new token in memory
		fmt.Fprintf(os.Stderr, "Warning: failed to update Claude credentials file: %v\n", err)
	}

	// Update in-memory credentials
	creds.ClaudeAiOauth.AccessToken = resp.AccessToken
	if resp.RefreshToken != "" {
		creds.ClaudeAiOauth.RefreshToken = resp.RefreshToken
	}
	creds.ClaudeAiOauth.ExpiresAt = time.Now().UnixMilli() + (resp.ExpiresIn * 1000)

	return creds, nil
}

// AutoRefreshQwenToken checks if Qwen token needs refresh and refreshes if necessary
func AutoRefreshQwenToken(creds *QwenOAuthCredentials) (*QwenOAuthCredentials, error) {
	if creds == nil {
		return creds, nil
	}

	// Check if refresh is needed
	if !NeedsRefresh(creds.ExpiryDate) {
		return creds, nil
	}

	// Check if we have a refresh token
	if creds.RefreshToken == "" {
		if IsExpired(creds.ExpiryDate) {
			return nil, fmt.Errorf("token expired and no refresh token available")
		}
		return creds, nil // Token expiring soon but no refresh token, return as-is
	}

	// Attempt refresh
	refresher := GetGlobalRefresher()
	resp, err := refresher.RefreshQwenToken(creds.RefreshToken, creds.ResourceURL)
	if err != nil {
		// If refresh fails but token isn't expired yet, return existing token
		if !IsExpired(creds.ExpiryDate) {
			return creds, nil
		}
		return nil, fmt.Errorf("failed to refresh expired token: %w", err)
	}

	// Update credentials file
	newRefreshToken := resp.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = creds.RefreshToken // Keep existing if not returned
	}

	if err := UpdateQwenCredentialsFile(resp.AccessToken, newRefreshToken, resp.IDToken, resp.ExpiresIn); err != nil {
		// Log error but continue with new token in memory
		fmt.Fprintf(os.Stderr, "Warning: failed to update Qwen credentials file: %v\n", err)
	}

	// Update in-memory credentials
	creds.AccessToken = resp.AccessToken
	if resp.RefreshToken != "" {
		creds.RefreshToken = resp.RefreshToken
	}
	if resp.IDToken != "" {
		creds.IDToken = resp.IDToken
	}
	creds.ExpiryDate = time.Now().UnixMilli() + (resp.ExpiresIn * 1000)

	return creds, nil
}

// RefreshResult contains the result of an auto-refresh operation
type RefreshResult struct {
	Refreshed    bool
	NewExpiresAt int64
	Error        error
}

// GetRefreshStatus returns the refresh status for both providers
func GetRefreshStatus() map[string]interface{} {
	reader := GetGlobalReader()

	status := map[string]interface{}{
		"refresh_threshold": RefreshThreshold.String(),
	}

	// Claude status
	claudeCreds, err := reader.ReadClaudeCredentials()
	if err == nil && claudeCreds.ClaudeAiOauth != nil {
		status["claude"] = map[string]interface{}{
			"needs_refresh":     NeedsRefresh(claudeCreds.ClaudeAiOauth.ExpiresAt),
			"is_expired":        IsExpired(claudeCreds.ClaudeAiOauth.ExpiresAt),
			"has_refresh_token": claudeCreds.ClaudeAiOauth.RefreshToken != "",
			"expires_at":        time.UnixMilli(claudeCreds.ClaudeAiOauth.ExpiresAt).Format(time.RFC3339),
		}
	}

	// Qwen status
	qwenCreds, err := reader.ReadQwenCredentials()
	if err == nil {
		status["qwen"] = map[string]interface{}{
			"needs_refresh":     NeedsRefresh(qwenCreds.ExpiryDate),
			"is_expired":        IsExpired(qwenCreds.ExpiryDate),
			"has_refresh_token": qwenCreds.RefreshToken != "",
			"expires_at":        time.UnixMilli(qwenCreds.ExpiryDate).Format(time.RFC3339),
		}
	}

	return status
}

// StartBackgroundRefresh starts a background goroutine that periodically checks
// and refreshes tokens before they expire
func StartBackgroundRefresh(stopChan <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				reader := GetGlobalReader()

				// Try to refresh Claude token if needed
				if IsClaudeOAuthEnabled() {
					creds, err := reader.ReadClaudeCredentials()
					if err == nil && creds.ClaudeAiOauth != nil {
						if NeedsRefresh(creds.ClaudeAiOauth.ExpiresAt) && creds.ClaudeAiOauth.RefreshToken != "" {
							_, err := AutoRefreshClaudeToken(creds)
							if err != nil {
								fmt.Fprintf(os.Stderr, "Background Claude token refresh failed: %v\n", err)
							} else {
								reader.ClearCache() // Clear cache to force re-read of updated file
							}
						}
					}
				}

				// Try to refresh Qwen token if needed
				if IsQwenOAuthEnabled() {
					creds, err := reader.ReadQwenCredentials()
					if err == nil {
						if NeedsRefresh(creds.ExpiryDate) && creds.RefreshToken != "" {
							_, err := AutoRefreshQwenToken(creds)
							if err != nil {
								fmt.Fprintf(os.Stderr, "Background Qwen token refresh failed: %v\n", err)
							} else {
								reader.ClearCache() // Clear cache to force re-read of updated file
							}
						}
					}
				}
			}
		}
	}()
}

// SetHTTPClient allows setting a custom HTTP client (useful for testing)
func (tr *TokenRefresher) SetHTTPClient(client *http.Client) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.httpClient = client
}
