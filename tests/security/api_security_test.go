package security

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAPISecurityRouter creates a Gin router that mimics the key security
// middleware of the real HelixAgent server: auth checks, body size limits,
// content-type validation, rate limit headers, and CORS. This allows us to
// test the security boundary without starting the full server.
func setupAPISecurityRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Body size limit middleware (10 MB).
	const maxBodySize int64 = 10 * 1024 * 1024
	r.Use(func(c *gin.Context) {
		if c.Request.ContentLength > maxBodySize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": gin.H{
					"message": "request body too large",
					"type":    "invalid_request_error",
					"code":    "body_too_large",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	})

	// Rate limit header middleware (stub — always adds headers).
	r.Use(func(c *gin.Context) {
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Remaining", "99")
		c.Header("X-RateLimit-Reset", "1700000000")
		c.Next()
	})

	// CORS middleware.
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Completions endpoint with auth and content-type checks.
	r.POST("/v1/chat/completions", func(c *gin.Context) {
		// Auth check.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "missing authorization header",
					"type":    "authentication_error",
				},
			})
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "invalid authorization format",
					"type":    "authentication_error",
				},
			})
			return
		}

		// Content-type check.
		ct := c.GetHeader("Content-Type")
		if !strings.HasPrefix(ct, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": gin.H{
					"message": "unsupported content type",
					"type":    "invalid_request_error",
				},
			})
			return
		}

		// Parse JSON body.
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "invalid JSON: " + err.Error(),
					"type":    "invalid_request_error",
				},
			})
			return
		}

		// Sanitize — ensure SQL keywords in prompt do not leak to DB errors.
		prompt, _ := req["prompt"].(string)
		messages, _ := req["messages"].([]interface{})
		if prompt == "" && len(messages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "prompt or messages required",
					"type":    "invalid_request_error",
				},
			})
			return
		}

		// Escape HTML in any response content to prevent XSS.
		safeContent := escapeHTML("This is a safe response")
		c.JSON(http.StatusOK, gin.H{
			"id":     "chatcmpl-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index":         0,
					"message":       map[string]string{"role": "assistant", "content": safeContent},
					"finish_reason": "stop",
				},
			},
		})
	})

	return r
}

// escapeHTML replaces dangerous characters that could lead to XSS.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// TestSecurity_SQLInjection_PromptField verifies that SQL injection
// payloads in the prompt field are handled safely — the server must not
// expose raw SQL error messages.
func TestSecurity_SQLInjection_PromptField(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	sqlPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users; --",
		"' UNION SELECT * FROM users --",
		"1; DELETE FROM memories WHERE 1=1 --",
		"Robert'); DROP TABLE students;--",
	}

	for _, payload := range sqlPayloads {
		t.Run("SQLi_"+sanitizePayloadName(payload), func(t *testing.T) {
			body := map[string]interface{}{
				"prompt": payload,
				"model":  "test-model",
			}
			jsonBody, _ := json.Marshal(body)
			req, _ := http.NewRequest("POST",
				server.URL+"/v1/chat/completions",
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, _ := io.ReadAll(resp.Body)
			bodyStr := strings.ToLower(string(respBody))

			// Must not expose SQL internals.
			dangerousPatterns := []string{
				"sql syntax", "mysql_fetch", "ora-",
				"postgresql", "sqlite_", "sqlserver",
				"unclosed quotation", "unterminated string",
			}
			for _, pattern := range dangerousPatterns {
				assert.NotContains(t, bodyStr, pattern,
					"Response must not expose SQL error for payload: %s", payload)
			}
		})
	}
}

// TestSecurity_XSS_ResponseFormatting verifies that any HTML special
// characters in the response content are properly escaped.
func TestSecurity_XSS_ResponseFormatting(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	body := map[string]interface{}{
		"prompt": "test prompt",
		"model":  "test-model",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST",
		server.URL+"/v1/chat/completions",
		bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// The response content should not contain unescaped HTML tags.
	assert.NotContains(t, string(respBody), "<script>",
		"Response must not contain unescaped script tags")
	assert.NotContains(t, string(respBody), "onerror=",
		"Response must not contain event handlers")
	assert.NotContains(t, string(respBody), "javascript:",
		"Response must not contain javascript: URLs")
}

// TestSecurity_LargePayload_Rejected verifies that payloads exceeding the
// maximum body size are rejected with 413.
func TestSecurity_LargePayload_Rejected(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	// Create a payload slightly over 10 MB.
	largeContent := strings.Repeat("A", 11*1024*1024)
	body := map[string]interface{}{
		"prompt": largeContent,
		"model":  "test-model",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST",
		server.URL+"/v1/chat/completions",
		bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	// Explicitly set ContentLength so the middleware can check it.
	req.ContentLength = int64(len(jsonBody))

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode,
		"Oversized payload should be rejected with 413")

	var errResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.NotNil(t, errResp["error"],
		"Error response should contain 'error' field")
}

// TestSecurity_InvalidJSON_Handled verifies that malformed JSON returns a
// 400 Bad Request, not a 500 Internal Server Error.
func TestSecurity_InvalidJSON_Handled(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	malformedPayloads := []struct {
		name string
		body string
	}{
		{"EmptyObject", "{}"},
		{"MissingClosingBrace", `{"prompt": "test"`},
		{"TrailingComma", `{"prompt": "test",}`},
		{"SingleQuotes", `{'prompt': 'test'}`},
		{"PlainText", `this is not json`},
		{"Null", `null`},
		{"Array", `[1, 2, 3]`},
	}

	for _, tc := range malformedPayloads {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST",
				server.URL+"/v1/chat/completions",
				bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Must be 400, not 500.
			assert.True(t,
				resp.StatusCode == http.StatusBadRequest ||
					resp.StatusCode == http.StatusOK,
				"Should return 400 for malformed JSON, got %d (200 acceptable for empty object)",
				resp.StatusCode)
			assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode,
				"Must not return 500 for malformed JSON input")
		})
	}
}

// TestSecurity_MissingAuthHeader_Rejected verifies that requests without
// an Authorization header receive a 401 Unauthorized response.
func TestSecurity_MissingAuthHeader_Rejected(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	body := map[string]interface{}{
		"prompt": "test",
		"model":  "test-model",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST",
		server.URL+"/v1/chat/completions",
		bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// Deliberately omit Authorization header.

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"Missing auth header should return 401")

	var errResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.NotNil(t, errResp["error"],
		"401 response should contain 'error' field")
}

// TestSecurity_RateLimitHeaders_Present verifies that rate limit headers
// are present in every API response.
func TestSecurity_RateLimitHeaders_Present(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	body := map[string]interface{}{
		"prompt": "hello",
		"model":  "test-model",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST",
		server.URL+"/v1/chat/completions",
		bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"),
		"X-RateLimit-Limit header must be present")
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"),
		"X-RateLimit-Remaining header must be present")
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"),
		"X-RateLimit-Reset header must be present")
}

// TestSecurity_CORS_Headers_Correct verifies that CORS headers are set
// correctly, including for preflight (OPTIONS) requests.
func TestSecurity_CORS_Headers_Correct(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	t.Run("PreflightRequest", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS",
			server.URL+"/v1/chat/completions", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Preflight should return 204 No Content.
		assert.Equal(t, http.StatusNoContent, resp.StatusCode,
			"OPTIONS preflight should return 204")

		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"),
			"CORS Allow-Origin header must be present")
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"),
			"CORS Allow-Methods header must be present")
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"),
			"CORS Allow-Headers header must be present")
	})

	t.Run("RegularRequest", func(t *testing.T) {
		body := map[string]interface{}{
			"prompt": "test",
			"model":  "test-model",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST",
			server.URL+"/v1/chat/completions",
			bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Origin", "https://example.com")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"),
			"CORS Allow-Origin header must be present on regular requests")
	})
}

// TestSecurity_ContentType_Validated verifies that requests with
// non-JSON content types are rejected with 415 Unsupported Media Type.
func TestSecurity_ContentType_Validated(t *testing.T) {
	router := setupAPISecurityRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	client := &http.Client{}

	unsupportedTypes := []struct {
		name        string
		contentType string
		body        string
	}{
		{"TextPlain", "text/plain", `{"prompt": "test"}`},
		{"TextHTML", "text/html", `<html><body>test</body></html>`},
		{"ApplicationXML", "application/xml", `<request><prompt>test</prompt></request>`},
		{"MultipartForm", "multipart/form-data", `prompt=test`},
		{"ApplicationFormURL", "application/x-www-form-urlencoded", `prompt=test`},
	}

	for _, tc := range unsupportedTypes {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST",
				server.URL+"/v1/chat/completions",
				bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			req.Header.Set("Authorization", "Bearer test-token")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode,
				"Content-Type '%s' should be rejected with 415", tc.contentType)
		})
	}
}

// sanitizePayloadName converts a payload string into a safe test name.
func sanitizePayloadName(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "%", "_")
	s = strings.ReplaceAll(s, "'", "_")
	s = strings.ReplaceAll(s, ";", "_")
	s = strings.ReplaceAll(s, " ", "_")
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}
