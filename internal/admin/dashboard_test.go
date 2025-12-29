package admin

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/superagent/superagent/internal/database"
	"github.com/superagent/superagent/internal/services"
)

func TestAdminDashboardServer_GetDashboardData(t *testing.T) {
	// Setup mock service and repository
	mockService := &services.ModelMetadataService{}
	mockRepo := &database.ModelMetadataRepository{}

	server := NewAdminDashboardServer(mockService, mockRepo)

	// Create a test context
	w := &responseWriter{}
	c, _ := gin.CreateTestContext(w)

	// Test the dashboard data endpoint
	server.getDashboardData(c)

	// Verify response
	assert.Equal(t, 200, w.statusCode)
	assert.Contains(t, string(w.body), "totalModels")
}

func TestAdminDashboardServer_TriggerRefresh(t *testing.T) {
	// This would require mocking the service methods
	// For now, just test that the server initializes correctly
	mockService := &services.ModelMetadataService{}
	mockRepo := &database.ModelMetadataRepository{}

	server := NewAdminDashboardServer(mockService, mockRepo)
	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
}

func TestAdminSecurityMiddleware(t *testing.T) {
	middleware := AdminSecurityMiddleware()
	assert.NotNil(t, middleware)

	// Test middleware execution
	w := &responseWriter{}
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		URL: &url.URL{Path: "/admin/dashboard"},
	}

	middleware(c)

	// Check security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

// Mock response writer for testing
type responseWriter struct {
	statusCode int
	body       []byte
	headers    http.Header
}

func (w *responseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
