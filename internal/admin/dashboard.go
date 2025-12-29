package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superagent/superagent/internal/database"
	"github.com/superagent/superagent/internal/services"
)

type AdminDashboardServer struct {
	router   *gin.Engine
	service  *services.ModelMetadataService
	database *database.ModelMetadataRepository
	logger   *log.Logger
}

func NewAdminDashboardServer(service *services.ModelMetadataService, repo *database.ModelMetadataRepository) *AdminDashboardServer {
	logger := log.New(os.Stdout, "[ADMIN] ", log.LstdFlags)

	return &AdminDashboardServer{
		router:   gin.Default(),
		service:  service,
		database: repo,
		logger:   logger,
	}
}

func (ads *AdminDashboardServer) SetupRoutes() {
	// Static files
	ads.router.Static("/admin/static", "./admin/static")

	// Dashboard routes
	ads.router.GET("/admin/dashboard", ads.serveDashboard)
	ads.router.GET("/admin/dashboard/data", ads.getDashboardData)
	ads.router.POST("/admin/dashboard/refresh", ads.triggerRefresh)

	// API routes for dashboard
	ads.router.GET("/admin/api/health", ads.getHealthStatus)
	ads.router.GET("/admin/api/metrics", ads.getMetrics)
	ads.router.GET("/admin/api/providers", ads.getProvidersStatus)
	ads.router.GET("/admin/api/refresh-history", ads.getRefreshHistory)

	ads.logger.Println("Admin dashboard routes configured")
}

func (ads *AdminDashboardServer) serveDashboard(c *gin.Context) {
	dashboardPath := "./admin/models-dashboard.html"
	if _, err := os.Stat(dashboardPath); os.IsNotExist(err) {
		// Fallback to embedded dashboard
		ads.serveEmbeddedDashboard(c)
		return
	}

	c.File(dashboardPath)
}

func (ads *AdminDashboardServer) serveEmbeddedDashboard(c *gin.Context) {
	dashboardHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SuperAgent - Models.dev Admin Dashboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <header class="bg-white shadow-sm border-b">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                <h1 class="text-2xl font-bold text-gray-900">SuperAgent Models.dev Admin</h1>
            </div>
        </header>

        <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
            <div class="bg-white shadow rounded-lg p-6">
                <h2 class="text-lg font-medium text-gray-900 mb-4">Dashboard</h2>
                <div id="dashboard-content">
                    <p class="text-gray-600">Loading dashboard data...</p>
                </div>
            </div>
        </main>
    </div>

    <script>
        async function loadDashboard() {
            try {
                const response = await fetch('/admin/dashboard/data');
                const data = await response.json();

                document.getElementById('dashboard-content').innerHTML = \`
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                        <div class="bg-blue-50 p-4 rounded-lg">
                            <h3 class="font-medium text-blue-900">Total Models</h3>
                            <p class="text-2xl font-bold text-blue-600">\${data.totalModels || 0}</p>
                        </div>
                        <div class="bg-green-50 p-4 rounded-lg">
                            <h3 class="font-medium text-green-900">Healthy</h3>
                            <p class="text-2xl font-bold text-green-600">\${data.healthy ? 'Yes' : 'No'}</p>
                        </div>
                        <div class="bg-yellow-50 p-4 rounded-lg">
                            <h3 class="font-medium text-yellow-900">Last Refresh</h3>
                            <p class="text-sm text-yellow-600">\${data.lastRefresh || 'Never'}</p>
                        </div>
                    </div>

                    <div class="mt-6">
                        <button onclick="triggerRefresh()" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
                            Refresh Models
                        </button>
                    </div>
                \`;
            } catch (error) {
                document.getElementById('dashboard-content').innerHTML = '<p class="text-red-600">Failed to load dashboard data</p>';
            }
        }

        async function triggerRefresh() {
            try {
                const response = await fetch('/admin/dashboard/refresh', { method: 'POST' });
                const result = await response.json();
                alert('Refresh ' + (result.success ? 'completed' : 'failed'));
                loadDashboard();
            } catch (error) {
                alert('Refresh failed');
            }
        }

        loadDashboard();
        setInterval(loadDashboard, 30000); // Refresh every 30 seconds
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(200, dashboardHTML)
}

func (ads *AdminDashboardServer) getDashboardData(c *gin.Context) {
	health := ads.service.GetHealthStatus(c.Request.Context())

	// Get basic stats
	totalModels := health.TotalModels
	lastRefresh := ""
	if !health.LastRefreshTime.IsZero() {
		lastRefresh = health.LastRefreshTime.Format(time.RFC3339)
	}

	data := gin.H{
		"totalModels": totalModels,
		"healthy":     health.APIHealthy && health.CacheHealthy && health.CircuitBreakerHealthy,
		"lastRefresh": lastRefresh,
		"errorMessage": health.ErrorMessage,
	}

	c.JSON(200, data)
}

func (ads *AdminDashboardServer) triggerRefresh(c *gin.Context) {
	incremental := c.Query("incremental") == "true"

	var err error
	if incremental {
		err = ads.service.RefreshModelsIncremental(c.Request.Context())
	} else {
		err = ads.service.RefreshModels(c.Request.Context())
	}

	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "type": map[bool]string{true: "incremental", false: "full"}[incremental]})
}

func (ads *AdminDashboardServer) getHealthStatus(c *gin.Context) {
	health := ads.service.GetHealthStatus(c.Request.Context())
	c.JSON(200, health)
}

func (ads *AdminDashboardServer) getMetrics(c *gin.Context) {
	metrics := ads.service.GetMetrics()
	if metrics == nil {
		c.JSON(200, gin.H{"message": "Metrics collection not available"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Metrics are available via /metrics endpoint",
		"circuit_breaker_healthy": true, // This would be determined by actual metrics
	})
}

func (ads *AdminDashboardServer) getProvidersStatus(c *gin.Context) {
	// In a real implementation, this would query provider health
	// For now, return mock data
	providers := []gin.H{
		{
			"name":       "OpenAI",
			"healthy":    true,
			"modelCount": 150,
			"lastSync":   time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
		},
		{
			"name":       "Anthropic",
			"healthy":    true,
			"modelCount": 25,
			"lastSync":   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
		{
			"name":       "Google",
			"healthy":    false,
			"modelCount": 89,
			"lastSync":   time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			"error":      "Sync in progress",
		},
	}

	c.JSON(200, gin.H{"providers": providers})
}

func (ads *AdminDashboardServer) getRefreshHistory(c *gin.Context) {
	histories, err := ads.service.GetRefreshHistory(c.Request.Context(), 20)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Convert to JSON-friendly format
	var result []gin.H
	for _, h := range histories {
		result = append(result, gin.H{
			"refresh_type":      h.RefreshType,
			"status":           h.Status,
			"models_refreshed": h.ModelsRefreshed,
			"models_failed":    h.ModelsFailed,
			"duration_seconds": h.DurationSeconds,
			"started_at":       h.StartedAt.Format(time.RFC3339),
			"completed_at":     h.CompletedAt,
		})
	}

	c.JSON(200, gin.H{"histories": result})
}

func (ads *AdminDashboardServer) Start(addr string) error {
	ads.logger.Printf("Starting admin dashboard server on %s", addr)
	return ads.router.Run(addr)
}

// Middleware to add security headers for admin routes
func AdminSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Check for admin authentication (simplified)
		if strings.HasPrefix(c.Request.URL.Path, "/admin") {
			// In a real implementation, you'd check for proper authentication
			// For now, we'll allow all requests
		}

		c.Next()
	}
}

// Example usage in main application
func SetupAdminDashboard(r *gin.Engine, service *services.ModelMetadataService, repo *database.ModelMetadataRepository) {
	dashboard := NewAdminDashboardServer(service, repo)

	// Add security middleware
	r.Use(AdminSecurityMiddleware())

	// Setup dashboard routes
	dashboard.SetupRoutes()

	// Mount dashboard routes on main router
	r.Any("/admin/*path", func(c *gin.Context) {
		dashboard.router.HandleContext(c)
	})
}