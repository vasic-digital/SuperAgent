package modelsdev

import (
	"fmt"
	"time"
)

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

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrorResponse) Error() string {
	return fmt.Sprintf("validation errors: %d issues found", len(e.Errors))
}

type RateLimitError struct {
	RetryAfter time.Duration `json:"retry_after"`
	Message    string        `json:"message"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (retry after: %v)", e.Message, e.RetryAfter)
}
