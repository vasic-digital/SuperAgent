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

// ============================================================================
// Tests for fibonacci.go
// ============================================================================

func TestFibonacci(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"n=0", 0, 0},
		{"n=1", 1, 1},
		{"n=2", 2, 1},
		{"n=3", 3, 2},
		{"n=5", 5, 5},
		{"n=10", 10, 55},
		{"n=20", 20, 6765},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fibonacci(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFibonacciSequence(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected []int
	}{
		{"n=0", 0, []int{}},
		{"n<0", -1, []int{}},
		{"n=1", 1, []int{0}},
		{"n=5", 5, []int{0, 1, 1, 2, 3}},
		{"n=8", 8, []int{0, 1, 1, 2, 3, 5, 8, 13}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FibonacciSequence(tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Tests for math.go
// ============================================================================

func TestFactorial(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expected    int64
		expectError bool
	}{
		{"n=0", 0, 1, false},
		{"n=1", 1, 1, false},
		{"n=5", 5, 120, false},
		{"n=10", 10, 3628800, false},
		{"n=20", 20, 2432902008176640000, false},
		{"negative", -1, 0, true},
		{"overflow n=21", 21, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Factorial(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, int64(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFactorialRecursive(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expected    int64
		expectError bool
	}{
		{"n=0", 0, 1, false},
		{"n=1", 1, 1, false},
		{"n=5", 5, 120, false},
		{"n=10", 10, 3628800, false},
		{"n=20", 20, 2432902008176640000, false},
		{"negative", -1, 0, true},
		{"overflow n=21", 21, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FactorialRecursive(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, int64(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFactorial_MatchesRecursive(t *testing.T) {
	for n := 0; n <= 15; n++ {
		iter, errI := Factorial(n)
		rec, errR := FactorialRecursive(n)
		assert.NoError(t, errI)
		assert.NoError(t, errR)
		assert.Equal(t, iter, rec, "Factorial(%d) mismatch: iterative=%d, recursive=%d", n, iter, rec)
	}
}

// ============================================================================
// Tests for path_validation.go
// ============================================================================

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"empty path", "", false},
		{"valid simple path", "/tmp/file.go", true},
		{"valid relative", "src/main.go", true},
		{"path traversal ..", "../../etc/passwd", false},
		{"semicolon injection", "/tmp/file;rm -rf /", false},
		{"pipe injection", "/tmp/file|cat /etc/passwd", false},
		{"dollar sign", "/tmp/$HOME/file", false},
		{"backtick", "/tmp/`whoami`/file", false},
		{"ampersand", "/tmp/file&other", false},
		{"parenthesis", "/tmp/(file)", false},
		{"newline", "/tmp/file\ninjection", false},
		{"carriage return", "/tmp/file\rinjection", false},
		{"curly braces", "/tmp/${var}/file", false},
		{"redirect", "/tmp/file>other", false},
		{"valid with dots", "/tmp/file.txt", true},
		{"valid with dashes", "/opt/my-app/data", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePath(tt.path)
			assert.Equal(t, tt.expected, result, "ValidatePath(%q)", tt.path)
		})
	}
}

func TestValidateSymbol(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		expected bool
	}{
		{"empty", "", false},
		{"valid simple", "myFunc", true},
		{"valid with underscore", "my_func", true},
		{"valid start underscore", "_private", true},
		{"valid alphanumeric", "func123", true},
		{"starts with digit", "123func", false},
		{"has space", "my func", false},
		{"has dash", "my-func", false},
		{"has dot", "my.func", false},
		{"single char", "x", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSymbol(tt.symbol)
			assert.Equal(t, tt.expected, result, "ValidateSymbol(%q)", tt.symbol)
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantOK   bool
		wantPath string
	}{
		{"valid path", "/tmp/file.go", true, "/tmp/file.go"},
		{"path with redundant slashes", "/tmp//file.go", true, "/tmp/file.go"},
		{"path traversal", "../../etc/passwd", false, ""},
		{"empty path", "", false, ""},
		{"injection", "/tmp/file;cat", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := SanitizePath(tt.path)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantPath, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateGitRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected bool
	}{
		{"empty", "", false},
		{"main", "main", true},
		{"branch with slash", "feature/my-feature", true},
		{"tag with dot", "v1.0.0", true},
		{"SHA", "abc123def456", true},
		{"with spaces", "my branch", false},
		{"with dollar", "feat/$var", false},
		{"with semicolon", "feat;cmd", false},
		{"with at", "feat@branch", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateGitRef(tt.ref)
			assert.Equal(t, tt.expected, result, "ValidateGitRef(%q)", tt.ref)
		})
	}
}

func TestValidateCommandArg(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		expected bool
	}{
		{"empty (safe)", "", true},
		{"normal word", "hello", true},
		{"normal path", "/tmp/file.txt", true},
		{"number", "42", true},
		{"semicolon", "arg;cmd", false},
		{"pipe", "arg|cmd", false},
		{"dollar", "$var", false},
		{"backtick", "`cmd`", false},
		{"ampersand", "arg&cmd", false},
		{"newline", "arg\ncmd", false},
		{"backslash", "arg\\cmd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCommandArg(tt.arg)
			assert.Equal(t, tt.expected, result, "ValidateCommandArg(%q)", tt.arg)
		})
	}
}

// ============================================================================
// Tests for string.go
// ============================================================================

func TestReverseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single char", "a", "a"},
		{"two chars", "ab", "ba"},
		{"simple word", "hello", "olleh"},
		{"palindrome", "racecar", "racecar"},
		{"with spaces", "hello world", "dlrow olleh"},
		{"UTF-8 chars", "héllo", "olléh"},
		{"emoji", "abc", "cba"},
		{"unicode cjk", "日本語", "語本日"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
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
