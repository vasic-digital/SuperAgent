//go:build security
// +build security

// Package security contains security validation tests for input handling.
package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// =============================================================================
// INPUT VALIDATION TESTS
// =============================================================================

// setupSecurityTestRouter creates a router with security validation middleware
func setupSecurityTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Mock chat completion endpoint
	r.POST("/v1/chat/completions", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		// Basic input validation
		messages, ok := req["messages"].([]interface{})
		if !ok || len(messages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "messages required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      "sec-test-id",
			"object":  "chat.completion",
			"choices": []map[string]interface{}{},
		})
	})

	// Mock user endpoint
	r.POST("/v1/users", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		// Validate email format
		email, _ := req["email"].(string)
		if !isValidEmail(email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": "user-1", "email": email})
	})

	// Echo endpoint for testing sanitization
	r.GET("/v1/echo", func(c *gin.Context) {
		msg := c.Query("message")
		// Sanitize output
		msg = sanitizeOutput(msg)
		c.JSON(http.StatusOK, gin.H{"message": msg})
	})

	return r
}

func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 254 {
		return false
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return false
	}
	// Reject dangerous characters (allow + for subaddressing)
	dangerous := []string{"'", "\"", ";", "--", "<", ">", "\\", "/", "|", "&"}
	for _, d := range dangerous {
		if strings.Contains(email, d) {
			return false
		}
	}
	return true
}

func sanitizeOutput(s string) string {
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// TestInputValidation_EmptyFields tests handling of empty/null input fields
func TestInputValidation_EmptyFields(t *testing.T) {
	router := setupSecurityTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	testCases := []struct {
		name         string
		body         map[string]interface{}
		expectedCode int
	}{
		{
			name:         "EmptyMessages",
			body:         map[string]interface{}{"model": "test", "messages": []interface{}{}},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "NullMessages",
			body:         map[string]interface{}{"model": "test", "messages": nil},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "MissingMessages",
			body:         map[string]interface{}{"model": "test"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ValidRequest",
			body: map[string]interface{}{
				"model":    "test",
				"messages": []interface{}{map[string]string{"role": "user", "content": "test"}},
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest("POST", server.URL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

// TestInputValidation_OverflowValues tests handling of extreme/overflow values
func TestInputValidation_OverflowValues(t *testing.T) {
	router := setupSecurityTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	testCases := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "VeryLargeMaxTokens",
			body: map[string]interface{}{
				"model":      "test",
				"messages":   []interface{}{map[string]string{"role": "user", "content": "test"}},
				"max_tokens": 999999999999,
			},
		},
		{
			name: "NegativeTemperature",
			body: map[string]interface{}{
				"model":       "test",
				"messages":    []interface{}{map[string]string{"role": "user", "content": "test"}},
				"temperature": -100,
			},
		},
		{
			name: "ExcessiveTemperature",
			body: map[string]interface{}{
				"model":       "test",
				"messages":    []interface{}{map[string]string{"role": "user", "content": "test"}},
				"temperature": 9999,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest("POST", server.URL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should not crash - either accept with normalization or reject with 400
			assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
				"Server should handle extreme values gracefully, got %d", resp.StatusCode)
		})
	}
}

// TestInputValidation_EmailFormat tests email validation
func TestInputValidation_EmailFormat(t *testing.T) {
	router := setupSecurityTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	testCases := []struct {
		name        string
		email       string
		expectValid bool
	}{
		{"ValidEmail", "user@example.com", true},
		{"ValidEmailWithSubaddress", "user+tag@example.com", true}, // subaddressing is valid
		{"SQLInjectionEmail", "user'--@example.com", false},
		{"XSSEmail", "<script>@example.com", false},
		{"CommandInjectionEmail", "user;ls@example.com", false},
		{"PathTraversalEmail", "../etc/passwd@example.com", false},
		{"TooShortEmail", "a@b", false},
		{"MissingAt", "userexample.com", false},
		{"MissingDomain", "user@", false},
		{"DoubleQuoteEmail", "user\"@example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{
				"email":    tc.email,
				"username": "testuser",
			}
			jsonBody, _ := json.Marshal(body)
			req, _ := http.NewRequest("POST", server.URL+"/v1/users", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			if tc.expectValid {
				assert.Equal(t, http.StatusCreated, resp.StatusCode, "Valid email should be accepted")
			} else {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Invalid email should be rejected")
			}
		})
	}
}

// TestInputValidation_OutputSanitization tests output sanitization
func TestInputValidation_OutputSanitization(t *testing.T) {
	router := setupSecurityTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	testCases := []struct {
		name        string
		input       string
		notExpected []string
	}{
		{
			name:        "ScriptTag",
			input:       "<script>alert('XSS')</script>",
			notExpected: []string{"<script>", "</script>"},
		},
		{
			name:        "ImageTag",
			input:       "<img src=x onerror=alert('XSS')>",
			notExpected: []string{"<img", "onerror="},
		},
		{
			name:        "QuoteEscape",
			input:       `" onclick="alert('XSS')"`,
			notExpected: []string{`" onclick=`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", server.URL+"/v1/echo?message="+tc.input, nil)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)

			message, _ := response["message"].(string)
			for _, notExp := range tc.notExpected {
				assert.NotContains(t, message, notExp, "Output should be sanitized")
			}
		})
	}
}

// =============================================================================
// PATH TRAVERSAL TESTS
// =============================================================================

// TestPathTraversal tests for path traversal vulnerabilities
func TestPathTraversal(t *testing.T) {
	router := gin.New()
	router.Use(gin.Recovery())

	// Secure file endpoint that validates paths
	router.GET("/v1/files/*path", func(c *gin.Context) {
		path := c.Param("path")

		// Check for path traversal attempts
		if strings.Contains(path, "..") ||
			strings.Contains(path, "\\") ||
			strings.HasPrefix(path, "/etc") ||
			strings.HasPrefix(path, "/root") ||
			strings.HasPrefix(path, "/home") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"path": path})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	pathTraversalPayloads := []string{
		"/../etc/passwd",
		"/..%2F..%2F..%2Fetc%2Fpasswd",
		"/....//....//etc/passwd",
		"\\..\\..\\..\\etc\\passwd",
		"/%2e%2e/%2e%2e/etc/passwd",
	}

	for _, payload := range pathTraversalPayloads {
		t.Run("PathTraversal_"+sanitizeTestName(payload), func(t *testing.T) {
			req, _ := http.NewRequest("GET", server.URL+"/v1/files"+payload, nil)
			resp, err := client.Do(req)
			if err != nil {
				return // Connection error is acceptable
			}
			defer resp.Body.Close()

			// Should not return 200 OK with file contents
			// 400 (bad request) or 404 (not found) are both acceptable rejections
			assert.NotEqual(t, http.StatusOK, resp.StatusCode,
				"Path traversal payload should not succeed: %s", payload)
		})
	}
}

// =============================================================================
// CONTENT TYPE VALIDATION TESTS
// =============================================================================

// TestContentTypeValidation tests content-type header validation
func TestContentTypeValidation(t *testing.T) {
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/v1/data", func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")

		// Only accept JSON
		if !strings.HasPrefix(contentType, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "unsupported content type"})
			return
		}

		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"received": true})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	testCases := []struct {
		name         string
		contentType  string
		body         string
		expectedCode int
	}{
		{"ValidJSON", "application/json", `{"key": "value"}`, http.StatusOK},
		{"TextPlain", "text/plain", `{"key": "value"}`, http.StatusUnsupportedMediaType},
		{"TextHTML", "text/html", `<script>alert('XSS')</script>`, http.StatusUnsupportedMediaType},
		{"ApplicationXML", "application/xml", `<data>value</data>`, http.StatusUnsupportedMediaType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", server.URL+"/v1/data", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", tc.contentType)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func sanitizeTestName(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "%", "_")
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}
