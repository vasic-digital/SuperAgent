package modelsdev

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Structure(t *testing.T) {
	err := ValidationError{
		Field:   "email",
		Message: "Invalid email format",
	}

	assert.Equal(t, "email", err.Field)
	assert.Equal(t, "Invalid email format", err.Message)
}

func TestValidationErrorResponse_Error(t *testing.T) {
	t.Run("single error", func(t *testing.T) {
		err := &ValidationErrorResponse{
			Errors: []ValidationError{
				{Field: "email", Message: "Invalid email"},
			},
		}

		assert.Contains(t, err.Error(), "validation errors: 1 issues found")
	})

	t.Run("multiple errors", func(t *testing.T) {
		err := &ValidationErrorResponse{
			Errors: []ValidationError{
				{Field: "email", Message: "Invalid email"},
				{Field: "name", Message: "Name is required"},
				{Field: "age", Message: "Must be positive"},
			},
		}

		assert.Contains(t, err.Error(), "validation errors: 3 issues found")
	})

	t.Run("no errors", func(t *testing.T) {
		err := &ValidationErrorResponse{
			Errors: []ValidationError{},
		}

		assert.Contains(t, err.Error(), "validation errors: 0 issues found")
	})
}

func TestRateLimitError_Error(t *testing.T) {
	t.Run("with short retry", func(t *testing.T) {
		err := &RateLimitError{
			RetryAfter: 5 * time.Second,
			Message:    "Too many requests",
		}

		result := err.Error()
		assert.Contains(t, result, "rate limit exceeded")
		assert.Contains(t, result, "Too many requests")
		assert.Contains(t, result, "5s")
	})

	t.Run("with long retry", func(t *testing.T) {
		err := &RateLimitError{
			RetryAfter: 2 * time.Minute,
			Message:    "Request quota exceeded",
		}

		result := err.Error()
		assert.Contains(t, result, "rate limit exceeded")
		assert.Contains(t, result, "Request quota exceeded")
		assert.Contains(t, result, "2m")
	})

	t.Run("zero retry after", func(t *testing.T) {
		err := &RateLimitError{
			RetryAfter: 0,
			Message:    "Rate limited",
		}

		result := err.Error()
		assert.Contains(t, result, "Rate limited")
		assert.Contains(t, result, "0s")
	})
}

func TestAPIError_Variations(t *testing.T) {
	t.Run("authentication error", func(t *testing.T) {
		err := &APIError{
			Type:    "authentication_error",
			Message: "Invalid API key",
			Code:    401,
		}

		result := err.Error()
		assert.Contains(t, result, "authentication_error")
		assert.Contains(t, result, "Invalid API key")
		assert.Contains(t, result, "401")
	})

	t.Run("rate limit error", func(t *testing.T) {
		err := &APIError{
			Type:    "rate_limit_error",
			Message: "Rate limit exceeded",
			Code:    429,
			Details: "Retry after 60 seconds",
		}

		result := err.Error()
		assert.Contains(t, result, "rate_limit_error")
		assert.Contains(t, result, "Rate limit exceeded")
		assert.Contains(t, result, "429")
		assert.Contains(t, result, "Retry after 60 seconds")
	})

	t.Run("permission error", func(t *testing.T) {
		err := &APIError{
			Type:    "permission_denied",
			Message: "Access denied",
			Code:    403,
		}

		result := err.Error()
		assert.Contains(t, result, "permission_denied")
		assert.Contains(t, result, "403")
	})

	t.Run("validation error", func(t *testing.T) {
		err := &APIError{
			Type:    "validation_error",
			Message: "Invalid parameters",
			Code:    400,
			Details: "Field 'model' is required",
		}

		result := err.Error()
		assert.Contains(t, result, "validation_error")
		assert.Contains(t, result, "Invalid parameters")
		assert.Contains(t, result, "Field 'model' is required")
	})
}

func TestAPIError_IsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected bool
	}{
		{"404 error", 404, true},
		{"200 success", 200, false},
		{"500 error", 500, false},
		{"401 error", 401, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{Code: tt.code}
			assert.Equal(t, tt.expected, err.IsNotFound())
		})
	}
}

func TestAPIError_IsRateLimited(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected bool
	}{
		{"429 error", 429, true},
		{"200 success", 200, false},
		{"500 error", 500, false},
		{"404 error", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{Code: tt.code}
			assert.Equal(t, tt.expected, err.IsRateLimited())
		})
	}
}

func TestAPIError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected bool
	}{
		{"429 rate limit", 429, true},
		{"500 server error", 500, true},
		{"502 bad gateway", 502, true},
		{"503 unavailable", 503, true},
		{"200 success", 200, false},
		{"400 bad request", 400, false},
		{"401 unauthorized", 401, false},
		{"404 not found", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{Code: tt.code}
			assert.Equal(t, tt.expected, err.IsRetryable())
		})
	}
}

func TestNetworkError(t *testing.T) {
	t.Run("with underlying error", func(t *testing.T) {
		underlying := &APIError{Code: 500, Message: "internal"}
		err := &NetworkError{
			Message:    "connection failed",
			Underlying: underlying,
		}

		result := err.Error()
		assert.Contains(t, result, "network error")
		assert.Contains(t, result, "connection failed")
		assert.Contains(t, result, "internal")

		// Test Unwrap
		assert.Equal(t, underlying, err.Unwrap())
	})

	t.Run("without underlying error", func(t *testing.T) {
		err := &NetworkError{
			Message: "connection refused",
		}

		result := err.Error()
		assert.Contains(t, result, "network error")
		assert.Contains(t, result, "connection refused")
		assert.NotContains(t, result, "caused by")

		// Test Unwrap
		assert.Nil(t, err.Unwrap())
	})
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Message: "request timed out",
		Timeout: 30 * time.Second,
	}

	result := err.Error()
	assert.Contains(t, result, "timeout error")
	assert.Contains(t, result, "request timed out")
	assert.Contains(t, result, "30s")
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ErrModelNotFound",
			err:      ErrModelNotFound,
			expected: true,
		},
		{
			name:     "ErrProviderNotFound",
			err:      ErrProviderNotFound,
			expected: true,
		},
		{
			name:     "APIError 404",
			err:      &APIError{Code: 404},
			expected: true,
		},
		{
			name:     "APIError 500",
			err:      &APIError{Code: 500},
			expected: false,
		},
		{
			name:     "other error",
			err:      ErrInvalidModelID,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "RateLimitError",
			err:      &RateLimitError{Message: "rate limited"},
			expected: true,
		},
		{
			name:     "APIError 429",
			err:      &APIError{Code: 429},
			expected: true,
		},
		{
			name:     "APIError 500",
			err:      &APIError{Code: 500},
			expected: false,
		},
		{
			name:     "other error",
			err:      ErrModelNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRateLimited(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "APIError 429",
			err:      &APIError{Code: 429},
			expected: true,
		},
		{
			name:     "APIError 500",
			err:      &APIError{Code: 500},
			expected: true,
		},
		{
			name:     "APIError 400",
			err:      &APIError{Code: 400},
			expected: false,
		},
		{
			name:     "RateLimitError",
			err:      &RateLimitError{},
			expected: true,
		},
		{
			name:     "NetworkError",
			err:      &NetworkError{Message: "connection failed"},
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrModelNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCacheMiss(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ErrCacheMiss",
			err:      ErrCacheMiss,
			expected: true,
		},
		{
			name:     "ErrCacheExpired",
			err:      ErrCacheExpired,
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrModelNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCacheMiss(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStandardErrors(t *testing.T) {
	// Verify standard errors are defined
	assert.NotNil(t, ErrModelNotFound)
	assert.NotNil(t, ErrProviderNotFound)
	assert.NotNil(t, ErrInvalidModelID)
	assert.NotNil(t, ErrInvalidProviderID)
	assert.NotNil(t, ErrCacheMiss)
	assert.NotNil(t, ErrCacheExpired)
	assert.NotNil(t, ErrServiceUnavailable)

	// Verify error messages
	assert.Equal(t, "model not found", ErrModelNotFound.Error())
	assert.Equal(t, "provider not found", ErrProviderNotFound.Error())
	assert.Equal(t, "invalid model ID", ErrInvalidModelID.Error())
	assert.Equal(t, "invalid provider ID", ErrInvalidProviderID.Error())
	assert.Equal(t, "cache miss", ErrCacheMiss.Error())
	assert.Equal(t, "cache entry expired", ErrCacheExpired.Error())
	assert.Equal(t, "service unavailable", ErrServiceUnavailable.Error())
}
