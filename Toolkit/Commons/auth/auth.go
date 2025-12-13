// Package auth provides authentication management for AI providers.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// AuthManager manages authentication for providers.
type AuthManager struct {
	apiKey     string
	token      string
	tokenMutex sync.RWMutex
	expiresAt  time.Time
	refresher  TokenRefresher
}

// TokenRefresher defines the interface for token refresh logic.
type TokenRefresher interface {
	RefreshToken(ctx context.Context) (*TokenResponse, error)
}

// TokenResponse represents a token refresh response.
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewAuthManager creates a new authentication manager.
func NewAuthManager(apiKey string, refresher TokenRefresher) *AuthManager {
	return &AuthManager{
		apiKey:    apiKey,
		refresher: refresher,
	}
}

// GetAuthHeader returns the appropriate authorization header value.
func (am *AuthManager) GetAuthHeader(ctx context.Context) (string, error) {
	if am.refresher != nil {
		return am.getTokenAuth(ctx)
	}
	return am.getAPIKeyAuth(), nil
}

// getAPIKeyAuth returns API key based authentication.
func (am *AuthManager) getAPIKeyAuth() string {
	return "Bearer " + am.apiKey
}

// getTokenAuth returns token-based authentication, refreshing if necessary.
func (am *AuthManager) getTokenAuth(ctx context.Context) (string, error) {
	am.tokenMutex.RLock()
	token := am.token
	expiresAt := am.expiresAt
	am.tokenMutex.RUnlock()

	// Check if token is still valid (with 5 minute buffer)
	if token != "" && time.Now().Add(5*time.Minute).Before(expiresAt) {
		return "Bearer " + token, nil
	}

	// Need to refresh token
	am.tokenMutex.Lock()
	defer am.tokenMutex.Unlock()

	// Double-check after acquiring write lock
	if am.token != "" && time.Now().Add(5*time.Minute).Before(am.expiresAt) {
		return "Bearer " + am.token, nil
	}

	// Refresh the token
	tokenResp, err := am.refresher.RefreshToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	am.token = tokenResp.Token
	am.expiresAt = tokenResp.ExpiresAt

	return "Bearer " + am.token, nil
}

// APIKeyAuth represents simple API key authentication.
type APIKeyAuth struct {
	APIKey string
}

// NewAPIKeyAuth creates a new API key authenticator.
func NewAPIKeyAuth(apiKey string) *APIKeyAuth {
	return &APIKeyAuth{APIKey: apiKey}
}

// GetAuthHeader returns the authorization header for API key auth.
func (a *APIKeyAuth) GetAuthHeader(ctx context.Context) (string, error) {
	return "Bearer " + a.APIKey, nil
}

// AuthInterceptor is an HTTP interceptor that adds authentication.
type AuthInterceptor struct {
	authManager AuthManagerInterface
}

// AuthManagerInterface defines the interface for auth managers.
type AuthManagerInterface interface {
	GetAuthHeader(ctx context.Context) (string, error)
}

// NewAuthInterceptor creates a new authentication interceptor.
func NewAuthInterceptor(authManager AuthManagerInterface) *AuthInterceptor {
	return &AuthInterceptor{authManager: authManager}
}

// Intercept adds authentication to the request.
func (ai *AuthInterceptor) Intercept(req *http.Request) error {
	authHeader, err := ai.authManager.GetAuthHeader(req.Context())
	if err != nil {
		return fmt.Errorf("failed to get auth header: %w", err)
	}

	req.Header.Set("Authorization", authHeader)
	return nil
}

// OAuth2Refresher implements OAuth2 token refresh.
type OAuth2Refresher struct {
	clientID     string
	clientSecret string
	tokenURL     string
	httpClient   *http.Client
}

// NewOAuth2Refresher creates a new OAuth2 token refresher.
func NewOAuth2Refresher(clientID, clientSecret, tokenURL string) *OAuth2Refresher {
	return &OAuth2Refresher{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// RefreshToken refreshes an OAuth2 token.
func (r *OAuth2Refresher) RefreshToken(ctx context.Context) (*TokenResponse, error) {
	// This is a simplified implementation
	// In a real implementation, you'd make an HTTP request to the token endpoint
	// For now, return a mock response
	return &TokenResponse{
		Token:     "mock_token_" + fmt.Sprintf("%d", time.Now().Unix()),
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil
}

// Middleware provides authentication middleware for HTTP clients.
type Middleware struct {
	authManager AuthManagerInterface
}

// NewMiddleware creates a new authentication middleware.
func NewMiddleware(authManager AuthManagerInterface) *Middleware {
	return &Middleware{authManager: authManager}
}

// WrapClient wraps an HTTP client with authentication.
func (m *Middleware) WrapClient(client *http.Client) *http.Client {
	// In a real implementation, this would intercept requests
	// For now, return the client as-is
	return client
}
