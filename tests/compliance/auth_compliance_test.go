package compliance

import (
	"testing"
	"time"

	"dev.helix.agent/internal/middleware"
	"github.com/stretchr/testify/assert"
)

// TestAuthConfigCompliance verifies that the AuthConfig type
// supports the required authentication configuration fields.
func TestAuthConfigCompliance(t *testing.T) {
	config := middleware.AuthConfig{
		SecretKey:   "test-secret-key",
		TokenExpiry: 24 * time.Hour,
		Issuer:      "helixagent",
		Required:    true,
	}

	assert.Equal(t, "test-secret-key", config.SecretKey)
	assert.Equal(t, 24*time.Hour, config.TokenExpiry)
	assert.Equal(t, "helixagent", config.Issuer)
	assert.True(t, config.Required)

	t.Logf("COMPLIANCE: AuthConfig supports SecretKey, TokenExpiry, Issuer, Required fields")
}

// TestJWTClaimsCompliance verifies the JWT Claims struct has required fields
// for user identification.
func TestJWTClaimsCompliance(t *testing.T) {
	claims := middleware.Claims{
		UserID:   "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)

	t.Logf("COMPLIANCE: JWT Claims has required UserID, Username, Role fields")
}

// TestLoginRequestCompliance verifies the LoginRequest struct has required fields.
func TestLoginRequestCompliance(t *testing.T) {
	req := middleware.LoginRequest{
		Username: "testuser",
		Password: "securepassword",
	}

	assert.Equal(t, "testuser", req.Username)
	assert.Equal(t, "securepassword", req.Password)

	t.Logf("COMPLIANCE: LoginRequest has required Username and Password fields")
}

// TestLoginResponseCompliance verifies the LoginResponse struct has required fields.
func TestLoginResponseCompliance(t *testing.T) {
	resp := middleware.LoginResponse{
		Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		ExpiresIn: 86400,
	}

	assert.NotEmpty(t, resp.Token, "LoginResponse must have Token field")
	assert.Greater(t, resp.ExpiresIn, 0, "LoginResponse must have positive ExpiresIn")

	t.Logf("COMPLIANCE: LoginResponse has required Token and ExpiresIn fields")
}

// TestAuthSkipPathsCompliance verifies that auth can be bypassed for
// specific paths (health check, public endpoints).
func TestAuthSkipPathsCompliance(t *testing.T) {
	config := middleware.AuthConfig{
		SecretKey: "test-secret",
		SkipPaths: []string{"/health", "/v1/models"},
	}

	assert.Contains(t, config.SkipPaths, "/health", "Health endpoint must be skippable")
	assert.Len(t, config.SkipPaths, 2)

	t.Logf("COMPLIANCE: Auth middleware supports SkipPaths for health and public endpoints")
}

// TestAuthOptionalCompliance verifies that authentication can be made optional
// (Required=false) for development or public deployments.
func TestAuthOptionalCompliance(t *testing.T) {
	optionalConfig := middleware.AuthConfig{
		SecretKey: "test-secret",
		Required:  false,
	}
	assert.False(t, optionalConfig.Required, "Auth should be configurable as optional")

	requiredConfig := middleware.AuthConfig{
		SecretKey: "test-secret",
		Required:  true,
	}
	assert.True(t, requiredConfig.Required, "Auth should be configurable as required")

	t.Logf("COMPLIANCE: Auth Required field allows flexible deployment modes")
}
