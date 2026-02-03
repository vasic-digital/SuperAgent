// Package auth provides adapters that bridge HelixAgent-specific authentication
// operations with the generic digital.vasic.auth module.
//
// This adapter layer maintains backward compatibility with existing code while
// allowing gradual migration to the extracted auth module. The internal auth
// package has HelixAgent-specific OAuth credential handling (Claude, Qwen specific)
// that extends the generic module's functionality.
package auth

import (
	"context"
	"net/http"
	"time"

	genericjwt "digital.vasic.auth/pkg/jwt"
	genericmw "digital.vasic.auth/pkg/middleware"
	genericoauth "digital.vasic.auth/pkg/oauth"
	generictoken "digital.vasic.auth/pkg/token"
)

// TypeAliases for re-exporting generic module types that can be used directly.

// GenericCredentials is the generic OAuth credentials type from the extracted module.
type GenericCredentials = genericoauth.Credentials

// GenericCredentialReader is the generic credential reader interface from the extracted module.
type GenericCredentialReader = genericoauth.CredentialReader

// GenericFileCredentialReader is the file-based credential reader from the extracted module.
type GenericFileCredentialReader = genericoauth.FileCredentialReader

// GenericTokenRefresher is the generic token refresher interface from the extracted module.
type GenericTokenRefresher = genericoauth.TokenRefresher

// GenericHTTPTokenRefresher is the HTTP-based token refresher from the extracted module.
type GenericHTTPTokenRefresher = genericoauth.HTTPTokenRefresher

// GenericAutoRefresher is the auto-refreshing credential manager from the extracted module.
type GenericAutoRefresher = genericoauth.AutoRefresher

// GenericOAuthConfig is the OAuth configuration from the extracted module.
type GenericOAuthConfig = genericoauth.Config

// GenericJWTConfig is the JWT configuration from the extracted module.
type GenericJWTConfig = genericjwt.Config

// GenericJWTManager is the JWT manager from the extracted module.
type GenericJWTManager = genericjwt.Manager

// GenericJWTToken is the parsed JWT token from the extracted module.
type GenericJWTToken = genericjwt.Token

// GenericMiddleware is the middleware function type from the extracted module.
type GenericMiddleware = genericmw.Middleware

// GenericTokenValidator is the token validator interface from the extracted module.
type GenericTokenValidator = genericmw.TokenValidator

// GenericAPIKeyValidator is the API key validator interface from the extracted module.
type GenericAPIKeyValidator = genericmw.APIKeyValidator

// GenericTokenStore is the token store interface from the extracted module.
type GenericTokenStore = generictoken.Store

// NewFileCredentialReader creates a new file-based credential reader using the generic module.
func NewFileCredentialReader(paths map[string]string) *genericoauth.FileCredentialReader {
	return genericoauth.NewFileCredentialReader(paths)
}

// NewHTTPTokenRefresher creates a new HTTP-based token refresher using the generic module.
func NewHTTPTokenRefresher(client *http.Client, clientID string, extraParams map[string]string) *genericoauth.HTTPTokenRefresher {
	return genericoauth.NewHTTPTokenRefresher(client, clientID, extraParams)
}

// NewAutoRefresher creates a new auto-refreshing credential manager using the generic module.
func NewAutoRefresher(
	reader genericoauth.CredentialReader,
	refresher genericoauth.TokenRefresher,
	config *genericoauth.Config,
	endpoints map[string]string,
) *genericoauth.AutoRefresher {
	return genericoauth.NewAutoRefresher(reader, refresher, config, endpoints)
}

// DefaultOAuthConfig returns the default OAuth configuration from the generic module.
func DefaultOAuthConfig() *genericoauth.Config {
	return genericoauth.DefaultConfig()
}

// NewJWTConfig creates a new JWT configuration using the generic module.
func NewJWTConfig(secret string) *genericjwt.Config {
	return genericjwt.DefaultConfig(secret)
}

// NewJWTManager creates a new JWT manager using the generic module.
func NewJWTManager(config *genericjwt.Config) *genericjwt.Manager {
	return genericjwt.NewManager(config)
}

// IsExpired checks if a token expiration time has passed.
func IsExpired(expiresAt time.Time) bool {
	return genericoauth.IsExpired(expiresAt)
}

// NeedsRefresh checks if a token needs refreshing based on threshold.
func NeedsRefresh(expiresAt time.Time, threshold time.Duration) bool {
	return genericoauth.NeedsRefresh(expiresAt, threshold)
}

// BearerTokenMiddleware creates middleware that validates Bearer tokens using the generic module.
func BearerTokenMiddleware(validator genericmw.TokenValidator) genericmw.Middleware {
	return genericmw.BearerToken(validator)
}

// APIKeyHeaderMiddleware creates middleware that validates API keys from headers using the generic module.
func APIKeyHeaderMiddleware(validator genericmw.APIKeyValidator, headerName string) genericmw.Middleware {
	return genericmw.APIKeyHeader(validator, headerName)
}

// RequireScopesMiddleware creates middleware that requires specific scopes using the generic module.
func RequireScopesMiddleware(scopes ...string) genericmw.Middleware {
	return genericmw.RequireScopes(scopes...)
}

// ChainMiddleware chains multiple middleware together using the generic module.
func ChainMiddleware(middlewares ...genericmw.Middleware) genericmw.Middleware {
	return genericmw.Chain(middlewares...)
}

// ClaimsFromContext extracts claims from request context using the generic module.
func ClaimsFromContext(ctx context.Context) map[string]interface{} {
	return genericmw.ClaimsFromContext(ctx)
}

// ScopesFromContext extracts scopes from request context using the generic module.
func ScopesFromContext(ctx context.Context) []string {
	return genericmw.ScopesFromContext(ctx)
}

// APIKeyFromContext extracts the API key from request context using the generic module.
func APIKeyFromContext(ctx context.Context) string {
	return genericmw.APIKeyFromContext(ctx)
}

// JWTValidatorAdapter adapts a JWT manager to the TokenValidator interface.
type JWTValidatorAdapter struct {
	manager *genericjwt.Manager
}

// NewJWTValidatorAdapter creates a new JWT validator adapter.
func NewJWTValidatorAdapter(manager *genericjwt.Manager) *JWTValidatorAdapter {
	return &JWTValidatorAdapter{manager: manager}
}

// ValidateToken validates a JWT token and returns claims.
func (a *JWTValidatorAdapter) ValidateToken(token string) (map[string]interface{}, error) {
	parsed, err := a.manager.Validate(token)
	if err != nil {
		return nil, err
	}
	return parsed.Claims, nil
}

// CredentialReaderAdapter adapts the generic CredentialReader to work with
// HelixAgent-specific credential types.
type CredentialReaderAdapter struct {
	reader genericoauth.CredentialReader
}

// NewCredentialReaderAdapter creates a new credential reader adapter.
func NewCredentialReaderAdapter(reader genericoauth.CredentialReader) *CredentialReaderAdapter {
	return &CredentialReaderAdapter{reader: reader}
}

// ReadCredentials reads credentials for the named provider.
func (a *CredentialReaderAdapter) ReadCredentials(providerName string) (*genericoauth.Credentials, error) {
	return a.reader.ReadCredentials(providerName)
}

// GetAccessToken retrieves the access token for a provider.
func (a *CredentialReaderAdapter) GetAccessToken(providerName string) (string, error) {
	creds, err := a.reader.ReadCredentials(providerName)
	if err != nil {
		return "", err
	}
	return creds.AccessToken, nil
}

// HasValidCredentials checks if valid credentials are available for a provider.
func (a *CredentialReaderAdapter) HasValidCredentials(providerName string) bool {
	creds, err := a.reader.ReadCredentials(providerName)
	if err != nil {
		return false
	}
	return creds.AccessToken != "" && !creds.IsExpired()
}

// GetCredentialInfo returns information about credentials for a provider.
func (a *CredentialReaderAdapter) GetCredentialInfo(providerName string) map[string]interface{} {
	creds, err := a.reader.ReadCredentials(providerName)
	if err != nil {
		return map[string]interface{}{
			"available": false,
			"error":     err.Error(),
		}
	}

	expiresIn := time.Duration(0)
	if !creds.ExpiresAt.IsZero() {
		expiresIn = time.Until(creds.ExpiresAt)
	}

	return map[string]interface{}{
		"available":         true,
		"scopes":            creds.Scopes,
		"expires_in":        expiresIn.String(),
		"has_refresh_token": creds.RefreshToken != "",
		"metadata":          creds.Metadata,
	}
}
