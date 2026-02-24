package e2e

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtClaimsE2E defines the JWT token claims structure for E2E tests
type jwtClaimsE2E struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// generateE2EJWT generates a valid JWT token for E2E tests
func generateE2EJWT() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "helixagent-test-secret-key-for-challenges-1767638342"
	}

	claims := &jwtClaimsE2E{
		UserID:   "1",
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "helixagent",
			Subject:   "1",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return tokenString
}

// getE2EAPIKey returns a valid API key for E2E tests
// It generates a JWT if the HELIXAGENT_API_KEY is not already a JWT
func getE2EAPIKey() string {
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if len(apiKey) > 3 && apiKey[:3] == "eyJ" {
		// Already a JWT token
		return apiKey
	}
	// Generate a JWT token
	return generateE2EJWT()
}
