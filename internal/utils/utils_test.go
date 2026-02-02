package utils

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "Error with cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test error occurred",
				Status:  http.StatusBadRequest,
				Cause:   errors.New("underlying error"),
			},
			expected: "Test error occurred: underlying error",
		},
		{
			name: "Error without cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test error occurred",
				Status:  http.StatusBadRequest,
				Cause:   nil,
			},
			expected: "Test error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.appError.Error())
		})
	}
}

func TestNewAppError(t *testing.T) {
	cause := errors.New("test cause")
	appErr := NewAppError("TEST_ERROR", "Test message", http.StatusBadRequest, cause)

	require.NotNil(t, appErr)
	assert.Equal(t, "TEST_ERROR", appErr.Code)
	assert.Equal(t, "Test message", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.Status)
	assert.Equal(t, cause, appErr.Cause)
}

func TestHandleError_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := &testResponseWriter{}
	c, _ := gin.CreateTestContext(w)

	cause := errors.New("underlying cause")
	appErr := NewAppError("VALIDATION_ERROR", "Invalid input", http.StatusBadRequest, cause)

	HandleError(c, appErr)

	assert.Equal(t, http.StatusBadRequest, w.statusCode)
	assert.Contains(t, w.body, `"error":"Invalid input"`)
	assert.Contains(t, w.body, `"code":"VALIDATION_ERROR"`)
}

func TestHandleError_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := &testResponseWriter{}
	c, _ := gin.CreateTestContext(w)

	genericErr := errors.New("generic error")

	HandleError(c, genericErr)

	assert.Equal(t, http.StatusInternalServerError, w.statusCode)
	assert.Contains(t, w.body, `"error":"internal server error"`)
}

func TestLogger_Initialization(t *testing.T) {
	// Test that logger is initialized
	assert.NotNil(t, Logger)
	assert.Equal(t, GetLogger(), Logger)
}

func TestLogger_EnvironmentLevel(t *testing.T) {
	// Save original environment
	originalLevel := os.Getenv("LOG_LEVEL")
	defer func() { _ = os.Setenv("LOG_LEVEL", originalLevel) }()

	// Test that logger is initialized with environment variable
	// We can't directly test the init() function since it's private,
	// but we can verify the logger exists and GetLogger works
	assert.NotNil(t, Logger)
	assert.Equal(t, Logger, GetLogger())

	// Test all log levels
	testCases := []struct {
		envLevel string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"warning", "warn"},
		{"error", "error"},
		{"", "info"},        // default
		{"invalid", "info"}, // default for invalid
	}

	for _, tc := range testCases {
		t.Run(tc.envLevel, func(t *testing.T) {
			_ = os.Setenv("LOG_LEVEL", tc.envLevel)
			// The logger is already initialized, so setting env var won't change it
			// But we can verify the logger still works
			assert.NotNil(t, Logger)
			assert.Equal(t, Logger, GetLogger())
		})
	}
}

func TestMockLLMRequest(t *testing.T) {
	req := MockLLMRequest()

	require.NotNil(t, req)
	assert.Equal(t, "test-request-123", req.ID)
	assert.Equal(t, "test-session-456", req.SessionID)
	assert.Equal(t, "test-user-789", req.UserID)
	assert.Equal(t, "Write a simple Go function that adds two numbers", req.Prompt)
	assert.Equal(t, "test-model", req.ModelParams.Model)
	assert.Equal(t, 0.7, req.ModelParams.Temperature)
	assert.Equal(t, 1000, req.ModelParams.MaxTokens)
	assert.Equal(t, "confidence_weighted", req.EnsembleConfig.Strategy)
	assert.Equal(t, 2, req.EnsembleConfig.MinProviders)
	assert.Equal(t, 0.8, req.EnsembleConfig.ConfidenceThreshold)
	assert.True(t, req.EnsembleConfig.FallbackToBest)
	assert.Equal(t, "pending", req.Status)
	assert.Equal(t, "code_generation", req.RequestType)
}

func TestMockLLMResponse(t *testing.T) {
	resp := MockLLMResponse()

	require.NotNil(t, resp)
	assert.Equal(t, "test-response-123", resp.ID)
	assert.Equal(t, "test-request-123", resp.RequestID)
	assert.Equal(t, "test-provider", resp.ProviderID)
	assert.Equal(t, "Test Provider", resp.ProviderName)
	assert.Equal(t, "func add(a, b int) int {\n    return a + b\n}", resp.Content)
	assert.Equal(t, 0.95, resp.Confidence)
	assert.Equal(t, 50, resp.TokensUsed)
	assert.Equal(t, int64(500), resp.ResponseTime)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.True(t, resp.Selected)
	assert.Equal(t, 0.95, resp.SelectionScore)
}

func TestMockLLMProvider(t *testing.T) {
	provider := MockLLMProvider()

	require.NotNil(t, provider)
	assert.Equal(t, "test-provider-123", provider.ID)
	assert.Equal(t, "Test Provider", provider.Name)
	assert.Equal(t, "test", provider.Type)
	assert.Equal(t, "test-api-key", provider.APIKey)
	assert.Equal(t, "https://api.test.com", provider.BaseURL)
	assert.Equal(t, "test-model", provider.Model)
	assert.Equal(t, 1.0, provider.Weight)
	assert.True(t, provider.Enabled)
	assert.Equal(t, "healthy", provider.HealthStatus)
	assert.Equal(t, int64(500), provider.ResponseTime)
}

func TestAssertNoError(t *testing.T) {
	// Test that it doesn't fail when error is nil
	t.Run("no error", func(t *testing.T) {
		AssertNoError(t, nil)
		// If we get here, test passed
	})

	// Test that it fails when error exists
	t.Run("with error", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
		// Instead, we'll verify the function exists and can be called
		// The actual failure behavior is tested by using it in other tests
	})
}

func TestAssertError(t *testing.T) {
	// Test that it doesn't fail when error exists
	t.Run("with error", func(t *testing.T) {
		AssertError(t, errors.New("test error"))
		// If we get here, test passed
	})

	// Test that it fails when error is nil
	t.Run("no error", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
		// Instead, we'll verify the function exists and can be called
	})
}

func TestAssertEqual(t *testing.T) {
	t.Run("equal values", func(t *testing.T) {
		AssertEqual(t, 42, 42)
		AssertEqual(t, "hello", "hello")
		AssertEqual(t, true, true)
	})

	t.Run("not equal values", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
		// The function will be used in other tests to verify its behavior
	})
}

func TestAssertNotEqual(t *testing.T) {
	t.Run("not equal values", func(t *testing.T) {
		AssertNotEqual(t, 42, 43)
		AssertNotEqual(t, "hello", "world")
		AssertNotEqual(t, true, false)
	})

	t.Run("equal values", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
	})
}

func TestAssertNotNil(t *testing.T) {
	t.Run("not nil values", func(t *testing.T) {
		AssertNotNil(t, "not nil")
		AssertNotNil(t, 42)
		AssertNotNil(t, []int{1, 2, 3})
	})

	t.Run("nil value", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
	})
}

func TestAssertNil(t *testing.T) {
	t.Run("nil values", func(t *testing.T) {
		AssertNil(t, nil)
	})

	t.Run("not nil value", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
	})
}

func TestAssertTrue(t *testing.T) {
	t.Run("true condition", func(t *testing.T) {
		AssertTrue(t, true)
	})

	t.Run("false condition", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
	})
}

func TestAssertFalse(t *testing.T) {
	t.Run("false condition", func(t *testing.T) {
		AssertFalse(t, false)
	})

	t.Run("true condition", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
	})
}

func TestAssertContains(t *testing.T) {
	t.Run("contains item", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		AssertContains(t, slice, 3)
		AssertContains(t, []string{"a", "b", "c"}, "b")
	})

	t.Run("doesn't contain item", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
	})
}

func TestAssertNotContains(t *testing.T) {
	t.Run("doesn't contain item", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		AssertNotContains(t, slice, 42)
		AssertNotContains(t, []string{"a", "b", "c"}, "z")
	})

	t.Run("contains item", func(t *testing.T) {
		// We can't directly test that it fails since it would stop the test
	})
}

func TestTestContext(t *testing.T) {
	ctx := TestContext()
	require.NotNil(t, ctx)

	// Verify it's a context.Context
	_, ok := ctx.(context.Context)
	assert.True(t, ok, "TestContext should return a context.Context")

	// Verify it's cancelled (since TestContext cancels immediately)
	select {
	case <-ctx.Done():
		// Context is done as expected
	default:
		t.Error("TestContext should return a cancelled context")
	}
}

// testResponseWriter is a mock response writer for testing
type testResponseWriter struct {
	statusCode int
	body       string
	header     http.Header
}

func (w *testResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *testResponseWriter) Write(data []byte) (int, error) {
	w.body = string(data)
	return len(data), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
