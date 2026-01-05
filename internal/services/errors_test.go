package services

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMServiceError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *LLMServiceError
		contains []string
	}{
		{
			name: "basic error message",
			err: &LLMServiceError{
				Type:    ErrorTypeProvider,
				Message: "Provider failed",
			},
			contains: []string{"provider_error", "Provider failed"},
		},
		{
			name: "error with provider",
			err: &LLMServiceError{
				Type:     ErrorTypeProvider,
				Message:  "Provider failed",
				Provider: "deepseek",
			},
			contains: []string{"provider_error", "Provider failed", "deepseek"},
		},
		{
			name: "error with cause",
			err: &LLMServiceError{
				Type:    ErrorTypeNetwork,
				Message: "Network error",
				Cause:   errors.New("connection refused"),
			},
			contains: []string{"network_error", "Network error", "connection refused"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, substr := range tt.contains {
				assert.Contains(t, errStr, substr)
			}
		})
	}
}

func TestLLMServiceError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &LLMServiceError{
		Type:    ErrorTypeProvider,
		Message: "wrapper",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())
	assert.True(t, errors.Is(err, cause))
}

func TestLLMServiceError_Is(t *testing.T) {
	err1 := &LLMServiceError{Type: ErrorTypeProvider}
	err2 := &LLMServiceError{Type: ErrorTypeProvider}
	err3 := &LLMServiceError{Type: ErrorTypeTimeout}

	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
	assert.False(t, err1.Is(errors.New("other")))
}

func TestNewProviderError(t *testing.T) {
	cause := errors.New("api call failed")
	err := NewProviderError("claude", cause)

	assert.Equal(t, ErrorTypeProvider, err.Type)
	assert.Equal(t, "claude", err.Provider)
	assert.Equal(t, http.StatusBadGateway, err.HTTPStatus)
	assert.True(t, err.Retryable)
	assert.Contains(t, err.Message, "claude")
}

func TestNewConfigurationError(t *testing.T) {
	cause := errors.New("missing api key")
	err := NewConfigurationError("Provider not configured", cause)

	assert.Equal(t, ErrorTypeConfiguration, err.Type)
	assert.Equal(t, http.StatusServiceUnavailable, err.HTTPStatus)
	assert.False(t, err.Retryable)
	assert.Contains(t, err.Message, "not configured")
}

func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError("openai", 30*time.Second)

	assert.Equal(t, ErrorTypeRateLimit, err.Type)
	assert.Equal(t, http.StatusTooManyRequests, err.HTTPStatus)
	assert.Equal(t, 30*time.Second, err.RetryAfter)
	assert.True(t, err.Retryable)
}

func TestNewTimeoutError(t *testing.T) {
	cause := errors.New("context deadline exceeded")
	err := NewTimeoutError("gemini", cause)

	assert.Equal(t, ErrorTypeTimeout, err.Type)
	assert.Equal(t, http.StatusGatewayTimeout, err.HTTPStatus)
	assert.True(t, err.Retryable)
}

func TestNewNetworkError(t *testing.T) {
	cause := errors.New("dial tcp: connection refused")
	err := NewNetworkError("ollama", cause)

	assert.Equal(t, ErrorTypeNetwork, err.Type)
	assert.Equal(t, http.StatusBadGateway, err.HTTPStatus)
	assert.True(t, err.Retryable)
}

func TestNewValidationError(t *testing.T) {
	details := map[string]interface{}{"field": "messages", "issue": "required"}
	err := NewValidationError("Invalid request", details)

	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.False(t, err.Retryable)
	assert.Equal(t, details, err.Details)
}

func TestNewServiceUnavailableError(t *testing.T) {
	err := NewServiceUnavailableError("Maintenance in progress", 5*time.Minute)

	assert.Equal(t, ErrorTypeServiceUnavailable, err.Type)
	assert.Equal(t, http.StatusServiceUnavailable, err.HTTPStatus)
	assert.Equal(t, 5*time.Minute, err.RetryAfter)
	assert.True(t, err.Retryable)
}

func TestNewAllProvidersFailedError(t *testing.T) {
	causes := []error{
		errors.New("provider 1 failed"),
		errors.New("provider 2 failed"),
		nil, // should be skipped
		errors.New("provider 3 failed"),
	}

	err := NewAllProvidersFailedError(3, 4, causes)

	assert.Equal(t, ErrorTypeAllProvidersFailed, err.Type)
	assert.Equal(t, http.StatusBadGateway, err.HTTPStatus)
	assert.True(t, err.Retryable)
	assert.Equal(t, 3, err.FailedCount)
	assert.Equal(t, 4, err.TotalCount)
	assert.Contains(t, err.Message, "3 providers failed")

	// Check details
	require.NotNil(t, err.Details)
	errMsgs, ok := err.Details["errors"].([]string)
	require.True(t, ok)
	assert.Len(t, errMsgs, 3) // nil should be skipped
}

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		provider     string
		expectedType ErrorType
	}{
		{
			name:         "rate limit error from message",
			err:          errors.New("rate limit exceeded, please try again later"),
			provider:     "openai",
			expectedType: ErrorTypeRateLimit,
		},
		{
			name:         "429 status in error",
			err:          errors.New("API returned status 429: too many requests"),
			provider:     "claude",
			expectedType: ErrorTypeRateLimit,
		},
		{
			name:         "timeout error",
			err:          errors.New("context deadline exceeded"),
			provider:     "deepseek",
			expectedType: ErrorTypeTimeout,
		},
		{
			name:         "network connection refused",
			err:          errors.New("dial tcp 127.0.0.1:11434: connection refused"),
			provider:     "ollama",
			expectedType: ErrorTypeNetwork,
		},
		{
			name:         "configuration error - api key",
			err:          errors.New("invalid api key provided"),
			provider:     "gemini",
			expectedType: ErrorTypeConfiguration,
		},
		{
			name:         "configuration error - not available",
			err:          errors.New("provider not available"),
			provider:     "qwen",
			expectedType: ErrorTypeConfiguration,
		},
		{
			name:         "validation error",
			err:          errors.New("invalid request: messages field is required"),
			provider:     "zai",
			expectedType: ErrorTypeValidation,
		},
		{
			name:         "generic provider error",
			err:          errors.New("unknown error from provider"),
			provider:     "test",
			expectedType: ErrorTypeProvider,
		},
		{
			name:         "already categorized error",
			err:          NewTimeoutError("already", nil),
			provider:     "ignored",
			expectedType: ErrorTypeTimeout,
		},
		{
			name:         "nil error",
			err:          nil,
			provider:     "test",
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeError(tt.err, tt.provider)
			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expectedType, result.Type)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "retryable provider error",
			err:       NewProviderError("test", nil),
			retryable: true,
		},
		{
			name:      "non-retryable configuration error",
			err:       NewConfigurationError("bad config", nil),
			retryable: false,
		},
		{
			name:      "retryable rate limit",
			err:       NewRateLimitError("test", time.Second),
			retryable: true,
		},
		{
			name:      "non-retryable validation error",
			err:       NewValidationError("bad input", nil),
			retryable: false,
		},
		{
			name:      "non-LLM error defaults to false",
			err:       errors.New("random error"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.retryable, IsRetryable(tt.err))
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{
			name:   "provider error - 502",
			err:    NewProviderError("test", nil),
			status: http.StatusBadGateway,
		},
		{
			name:   "configuration error - 503",
			err:    NewConfigurationError("bad config", nil),
			status: http.StatusServiceUnavailable,
		},
		{
			name:   "rate limit - 429",
			err:    NewRateLimitError("test", time.Second),
			status: http.StatusTooManyRequests,
		},
		{
			name:   "timeout - 504",
			err:    NewTimeoutError("test", nil),
			status: http.StatusGatewayTimeout,
		},
		{
			name:   "validation - 400",
			err:    NewValidationError("bad input", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "generic error - 500",
			err:    errors.New("unknown"),
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.status, GetHTTPStatus(tt.err))
		})
	}
}

func TestGetRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected time.Duration
	}{
		{
			name:     "rate limit with retry after",
			err:      NewRateLimitError("test", 30*time.Second),
			expected: 30 * time.Second,
		},
		{
			name:     "service unavailable with retry after",
			err:      NewServiceUnavailableError("maintenance", 5*time.Minute),
			expected: 5 * time.Minute,
		},
		{
			name:     "provider error - no retry after",
			err:      NewProviderError("test", nil),
			expected: 0,
		},
		{
			name:     "generic error - no retry after",
			err:      errors.New("unknown"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetRetryAfter(tt.err))
		})
	}
}

func TestToOpenAIError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := NewProviderError("test", nil)
		result := err.ToOpenAIError()

		errObj, ok := result["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "PROVIDER_ERROR", errObj["code"])
		assert.Equal(t, "provider_error", errObj["type"])
	})

	t.Run("all providers failed converts type", func(t *testing.T) {
		err := NewAllProvidersFailedError(3, 3, nil)
		result := err.ToOpenAIError()

		errObj, ok := result["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "server_error", errObj["type"])
	})

	t.Run("includes retry after", func(t *testing.T) {
		err := NewRateLimitError("test", 30*time.Second)
		result := err.ToOpenAIError()

		assert.Equal(t, float64(30), result["retry_after"])
	})
}

// Integration-style tests to verify error flow
func TestErrorCategorization_RealWorldScenarios(t *testing.T) {
	scenarios := []struct {
		name         string
		rawError     string
		expectedType ErrorType
		expectedCode int
	}{
		{
			name:         "OpenAI rate limit response",
			rawError:     "Error code: 429 - {'error': {'message': 'Rate limit reached'}}",
			expectedType: ErrorTypeRateLimit,
			expectedCode: http.StatusTooManyRequests,
		},
		{
			name:         "Claude API authentication failure",
			rawError:     "authentication failed: invalid api key",
			expectedType: ErrorTypeConfiguration,
			expectedCode: http.StatusServiceUnavailable,
		},
		{
			name:         "Ollama connection refused",
			rawError:     "Post \"http://localhost:11434/api/generate\": dial tcp [::1]:11434: connection refused",
			expectedType: ErrorTypeNetwork,
			expectedCode: http.StatusBadGateway,
		},
		{
			name:         "Request timeout",
			rawError:     "context deadline exceeded while waiting for response",
			expectedType: ErrorTypeTimeout,
			expectedCode: http.StatusGatewayTimeout,
		},
		{
			name:         "DeepSeek API error",
			rawError:     "deepseek api returned 500: internal server error",
			expectedType: ErrorTypeProvider,
			expectedCode: http.StatusBadGateway,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			err := CategorizeError(errors.New(sc.rawError), "test-provider")

			assert.Equal(t, sc.expectedType, err.Type, "Error type mismatch")
			assert.Equal(t, sc.expectedCode, err.HTTPStatus, "HTTP status mismatch")
		})
	}
}
