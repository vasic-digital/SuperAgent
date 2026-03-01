// Package auth provides authentication and authorization adapters for HelixAgent.
//
// This package bridges HelixAgent's authentication needs with the extracted
// digital.vasic.auth module, providing JWT validation, API key authentication,
// and OAuth integration.
//
// # Overview
//
// The auth package provides:
//
//   - JWT token validation and generation
//   - API key authentication
//   - OAuth 2.0 flow support
//   - Role-based access control (RBAC)
//   - Middleware for HTTP authentication
//
// # JWT Authentication
//
// Validate JWT tokens:
//
//	adapter := auth.NewAdapter(cfg)
//	claims, err := adapter.ValidateJWT(tokenString)
//	if err != nil {
//	    // Handle invalid token
//	}
//
// Generate JWT tokens:
//
//	token, err := adapter.GenerateJWT(userID, roles, expiration)
//
// # API Key Authentication
//
// Validate API keys:
//
//	valid, err := adapter.ValidateAPIKey(ctx, apiKey)
//
// # OAuth Integration
//
// Handle OAuth callbacks:
//
//	token, err := adapter.ExchangeOAuthCode(ctx, code, provider)
//
// # HTTP Middleware
//
// Protect routes with authentication:
//
//	router.Use(auth.JWTMiddleware(adapter))
//	router.Use(auth.APIKeyMiddleware(adapter))
//
// # Role-Based Access Control
//
// Check permissions:
//
//	hasRole := adapter.HasRole(claims, "admin")
//	hasPermission := adapter.HasPermission(claims, "debate:create")
//
// # Configuration
//
//	cfg := &auth.Config{
//	    JWTSecret:     os.Getenv("JWT_SECRET"),
//	    TokenDuration: 24 * time.Hour,
//	    APIKeyHeader:  "X-API-Key",
//	}
//
// # Environment Variables
//
//	JWT_SECRET         - Secret key for JWT signing
//	JWT_DURATION       - Token duration (default: 24h)
//	API_KEY_HEADER     - Header name for API keys (default: X-API-Key)
//
// # Key Files
//
//   - adapter.go: Core authentication adapter
//   - integration.go: Integration with extracted auth module
//   - adapter_test.go: Unit tests
//   - integration_test.go: Integration tests
//
// # Module Integration
//
// This package uses digital.vasic.auth for:
//   - JWT implementation (pkg/jwt)
//   - API key management (pkg/apikey)
//   - OAuth flows (pkg/oauth)
//   - RBAC (pkg/rbac)
package auth
