package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Providers endpoints (stub MVP)
	r.GET("/v1/providers", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{"id": "prov-default", "name": "DefaultProvider", "type": "builtin", "enabled": true},
		})
	})
	r.POST("/v1/providers", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"success": true, "message": "provider added (stub)"})
	})

	_ = r.Run(":8080")
}
