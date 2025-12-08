package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates a simple API key passed in the X-API-Key header.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		// Fallback to environment variable if header not provided
		if apiKey == "" {
			apiKey = os.Getenv("SUPERAGENT_API_KEY")
		}
		// Simple in-memory check; in a real setup this would query a store
		if apiKey == "" || apiKey != "dev-key-123" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
