package modelsdev

import (
	"errors"
	"fmt"
	"time"
)

// Standard errors
var (
	// ErrModelNotFound is returned when a model is not found
	ErrModelNotFound = errors.New("model not found")

	// ErrProviderNotFound is returned when a provider is not found
	ErrProviderNotFound = errors.New("provider not found")

	// ErrInvalidModelID is returned when an invalid model ID is provided
	ErrInvalidModelID = errors.New("invalid model ID")

	// ErrInvalidProviderID is returned when an invalid provider ID is provided
	ErrInvalidProviderID = errors.New("invalid provider ID")

	// ErrCacheMiss is returned when an item is not found in cache
	ErrCacheMiss = errors.New("cache miss")

	// ErrCacheExpired is returned when a cached item has expired
	ErrCacheExpired = errors.New("cache entry expired")

	// ErrServiceUnavailable is returned when the service is unavailable
	ErrServiceUnavailable = errors.New("service unavailable")
)

// APIError represents an error returned by the Models.dev API
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("API error [%s]: %s (code: %d, details: %s)", e.Type, e.Message, e.Code, e.Details)
	}
	return fmt.Sprintf("API error [%s]: %s (code: %d)", e.Type, e.Message, e.Code)
}

// IsNotFound returns true if the error is a 404 not found error
func (e *APIError) IsNotFound() bool {
	return e.Code == 404
}

// IsRateLimited returns true if the error is a 429 rate limit error
func (e *APIError) IsRateLimited() bool {
	return e.Code == 429
}

// IsRetryable returns true if the error is retryable (5xx or rate limit)
func (e *APIError) IsRetryable() bool {
	return e.Code >= 500 || e.Code == 429
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a response containing validation errors
type ValidationErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrorResponse) Error() string {
	return fmt.Sprintf("validation errors: %d issues found", len(e.Errors))
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	RetryAfter time.Duration `json:"retry_after"`
	Message    string        `json:"message"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (retry after: %v)", e.Message, e.RetryAfter)
}

// NetworkError represents a network-related error
type NetworkError struct {
	Message    string `json:"message"`
	Underlying error  `json:"-"`
}

func (e *NetworkError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("network error: %s (caused by: %v)", e.Message, e.Underlying)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Underlying
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Message string        `json:"message"`
	Timeout time.Duration `json:"timeout"`
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout error: %s (timeout: %v)", e.Message, e.Timeout)
}

// Helper functions for error checking

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrModelNotFound) || errors.Is(err, ErrProviderNotFound) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsNotFound()
	}
	return false
}

// IsRateLimited checks if the error is a rate limit error
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRateLimited()
	}
	return false
}

// IsRetryable checks if the error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return true
	}
	var networkErr *NetworkError
	if errors.As(err, &networkErr) {
		return true
	}
	return false
}

// IsCacheMiss checks if the error is a cache miss
func IsCacheMiss(err error) bool {
	return err != nil && (errors.Is(err, ErrCacheMiss) || errors.Is(err, ErrCacheExpired))
}
