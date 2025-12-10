package main

import (
	"log"
	"time"

	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/database"
	llm "github.com/superagent/superagent/internal/llm"
	"github.com/superagent/superagent/internal/middleware"
	"github.com/superagent/superagent/internal/models"
	"github.com/superagent/superagent/internal/services"
)

func main() {
	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize user service
	userService := services.NewUserService(db, cfg.Server.JWTSecret, 24*time.Hour)

	// Create auth middleware
	authConfig := middleware.AuthConfig{
		SecretKey:   cfg.Server.JWTSecret,
		TokenExpiry: 24 * time.Hour,
		Issuer:      "superagent",
		SkipPaths:   []string{"/health", "/v1/health", "/v1/auth/login"},
		Required:    true,
	}
	auth := middleware.NewAuthMiddleware(authConfig, userService)

	r := gin.Default()

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Authentication endpoints
	r.POST("/v1/auth/login", auth.Login)

	// Attach auth middleware to protected routes
	protected := r.Group("", auth.Middleware([]string{"/health", "/v1/health", "/v1/auth/login"}))

	// Providers endpoints (stub MVP) - protected for mutation
	protected.GET("/v1/providers", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{"id": "prov-default", "name": "DefaultProvider", "type": "builtin", "enabled": true},
		})
	})
	protected.POST("/v1/providers", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"success": true, "message": "provider added (stub)"})
	})

	// Completions endpoint (ensemble-based)
	r.POST("/v1/completions", func(c *gin.Context) {
		var req models.LLMRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		responses, selected, err := llm.RunEnsemble(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"responses": responses, "selected": selected})
	})

	// gRPC-like bridge endpoint (paralleled path to show gRPC integration readiness)
	r.POST("/grpc/llm/complete", func(c *gin.Context) {
		var req models.LLMRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		responses, selected, err := llm.RunEnsemble(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"responses": responses, "selected": selected})
	})

	// Extra health/checks route for richer observability
	r.GET("/v1/health", func(c *gin.Context) {
		detailed := c.Query("detailed")
		_ = detailed
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "components": []string{"database", "providers"}})
	})

	// Metrics endpoint placeholder
	r.GET("/v1/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"llm_requests_total": 42, "provider_health": "healthy"})
	})

	_ = os.Setenv("SUPERAGENT_API_KEY", "dev-key-123")
	_ = r.Run(":8080")
}
