// Package errors provides standardized error types and utilities for AI providers.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ProviderError represents a provider-specific error.
type ProviderError struct {
	Provider   string
	Code       string
	Message    string
	StatusCode int
	Details    interface{}
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
}

// NewProviderError creates a new provider error.
func NewProviderError(provider, code, message string) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Code:     code,
		Message:  message,
	}
}

// WithStatusCode sets the HTTP status code.
func (e *ProviderError) WithStatusCode(code int) *ProviderError {
	e.StatusCode = code
	return e
}

// WithDetails adds additional details.
func (e *ProviderError) WithDetails(details interface{}) *ProviderError {
	e.Details = details
	return e
}

// APIError represents an error returned by an API.
type APIError struct {
	Type    string      `json:"type"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Param   string      `json:"param,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// RateLimitError represents a rate limit error.
type RateLimitError struct {
	RetryAfter int // seconds
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (retry after %d seconds)", e.Message, e.RetryAfter)
}

// AuthenticationError represents an authentication error.
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Message)
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NetworkError represents a network-related error.
type NetworkError struct {
	Underlying error
	Message    string
}

func (e *NetworkError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("network error: %s (%v)", e.Message, e.Underlying)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Underlying
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	Operation string
	Timeout   string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout error in %s after %s", e.Operation, e.Timeout)
}

// ErrorHandler provides utilities for handling different types of errors.
type ErrorHandler struct {
	provider string
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(provider string) *ErrorHandler {
	return &ErrorHandler{provider: provider}
}

// HandleHTTPError handles HTTP errors and converts them to appropriate error types.
func (h *ErrorHandler) HandleHTTPError(resp *http.Response, body []byte) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{Message: "invalid API key"}
	case http.StatusForbidden:
		return &AuthenticationError{Message: "insufficient permissions"}
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		return &RateLimitError{
			Message:    "rate limit exceeded",
			RetryAfter: parseRetryAfter(retryAfter),
		}
	case http.StatusBadRequest:
		return h.parseAPIError(body)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return &ProviderError{
			Provider:   h.provider,
			Code:       "server_error",
			Message:    fmt.Sprintf("server error: %s", string(body)),
			StatusCode: resp.StatusCode,
		}
	default:
		return &ProviderError{
			Provider:   h.provider,
			Code:       "http_error",
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}
}

// parseAPIError attempts to parse API-specific error from response body.
func (h *ErrorHandler) parseAPIError(body []byte) error {
	// Try to parse as APIError
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Type != "" {
		return &apiErr
	}

	// Fallback to generic provider error
	return &ProviderError{
		Provider: h.provider,
		Code:     "api_error",
		Message:  string(body),
	}
}

// HandleNetworkError handles network-related errors.
func (h *ErrorHandler) HandleNetworkError(err error, operation string) error {
	return &NetworkError{
		Underlying: err,
		Message:    fmt.Sprintf("failed during %s", operation),
	}
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	switch err.(type) {
	case *NetworkError, *RateLimitError, *TimeoutError:
		return true
	case *ProviderError:
		pe := err.(*ProviderError)
		return pe.StatusCode >= 500
	case *APIError:
		ae := err.(*APIError)
		return ae.Type == "server_error" || ae.Type == "rate_limit_error"
	default:
		return false
	}
}

// IsRateLimit checks if an error is a rate limit error.
func IsRateLimit(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// IsAuth checks if an error is an authentication error.
func IsAuth(err error) bool {
	_, ok := err.(*AuthenticationError)
	return ok
}

// GetRetryAfter extracts retry-after duration from rate limit errors.
func GetRetryAfter(err error) int {
	if rle, ok := err.(*RateLimitError); ok {
		return rle.RetryAfter
	}
	return 0
}

// parseRetryAfter parses the Retry-After header.
func parseRetryAfter(retryAfter string) int {
	if retryAfter == "" {
		return 60 // default 1 minute
	}
	// Try to parse as integer (seconds)
	if seconds := atoi(retryAfter); seconds > 0 {
		return seconds
	}
	// Could parse as HTTP date, but for simplicity return default
	return 60
}

// atoi is a simple string to int conversion.
func atoi(s string) int {
	result := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		result = result*10 + int(r-'0')
	}
	return result
}
