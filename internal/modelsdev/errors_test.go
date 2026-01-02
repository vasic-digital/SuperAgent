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
