package router

import (
	"github.com/gin-gonic/gin"
	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/pkg/metrics"
)

// SetupRouter creates and configures the main HTTP router.
func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(metrics.Handler()))

	// TODO: Add API routes here
	// Protected routes will be added here

	return r
}
