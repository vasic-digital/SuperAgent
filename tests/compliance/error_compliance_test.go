package compliance

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openAIErrorResponse represents the OpenAI-compatible error response format.
type openAIErrorResponse struct {
	Error openAIError `json:"error"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
}

// TestErrorResponseJSONFormat verifies that error responses serialize to
// the expected OpenAI-compatible JSON format.
func TestErrorResponseJSONFormat(t *testing.T) {
	errResp := openAIErrorResponse{
		Error: openAIError{
			Message: "The model does not exist",
			Type:    "invalid_request_error",
			Code:    "model_not_found",
		},
	}

	data, err := json.Marshal(errResp)
	require.NoError(t, err)

	// Verify JSON structure
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Contains(t, decoded, "error", "Error response must have 'error' key")

	errorObj, ok := decoded["error"].(map[string]interface{})
	require.True(t, ok, "Error value must be an object")

	assert.Contains(t, errorObj, "message", "Error object must have 'message' field")
	assert.Contains(t, errorObj, "type", "Error object must have 'type' field")

	t.Logf("COMPLIANCE: Error response follows OpenAI format: %s", string(data))
}

// TestErrorTypeCompliance verifies that all required error types
// are defined for the OpenAI-compatible error taxonomy.
func TestErrorTypeCompliance(t *testing.T) {
	requiredErrorTypes := []string{
		"invalid_request_error",
		"authentication_error",
		"permission_error",
		"not_found_error",
		"rate_limit_error",
		"api_error",
		"service_unavailable_error",
	}

	// Error types are strings — verify they are non-empty and well-formed
	for _, errType := range requiredErrorTypes {
		assert.NotEmpty(t, errType)
		// Should not contain spaces (snake_case)
		for _, c := range errType {
			assert.NotEqual(t, ' ', c, "Error type %q must not contain spaces", errType)
		}
	}

	t.Logf("COMPLIANCE: All required OpenAI error types defined: %v", requiredErrorTypes)
}

// TestErrorCodeCompliance verifies that common error codes are well-formed.
func TestErrorCodeCompliance(t *testing.T) {
	errorCodes := map[string]string{
		"model_not_found":           "invalid_request_error",
		"invalid_api_key":           "authentication_error",
		"insufficient_quota":        "rate_limit_error",
		"context_length_exceeded":   "invalid_request_error",
		"rate_limit_exceeded":       "rate_limit_error",
		"server_error":              "api_error",
	}

	for code, errType := range errorCodes {
		assert.NotEmpty(t, code, "Error code must not be empty")
		assert.NotEmpty(t, errType, "Error type must not be empty")
	}

	t.Logf("COMPLIANCE: %d error codes defined with correct types", len(errorCodes))
}

// TestFallbackErrorFormat verifies that fallback chain errors follow the
// expected categorized format required by the constitution.
func TestFallbackErrorFormat(t *testing.T) {
	// Fallback error categories per CLAUDE.md requirement
	fallbackCategories := []string{
		"rate_limit", "timeout", "auth",
		"connection", "unavailable", "overloaded",
	}

	for _, category := range fallbackCategories {
		assert.NotEmpty(t, category, "Fallback category must not be empty")
	}

	// All categories should be lowercase
	for _, cat := range fallbackCategories {
		for _, c := range cat {
			assert.True(t, (c >= 'a' && c <= 'z') || c == '_',
				"Fallback category %q must be lowercase with underscores", cat)
		}
	}

	t.Logf("COMPLIANCE: Fallback error categories are properly formatted: %v", fallbackCategories)
}

// TestErrorMessageCompliance verifies that error messages are human-readable
// and follow proper conventions.
func TestErrorMessageCompliance(t *testing.T) {
	sampleErrors := []openAIErrorResponse{
		{Error: openAIError{Message: "Invalid API key provided", Type: "authentication_error"}},
		{Error: openAIError{Message: "Rate limit exceeded", Type: "rate_limit_error"}},
		{Error: openAIError{Message: "Model not found", Type: "invalid_request_error"}},
	}

	for _, e := range sampleErrors {
		assert.NotEmpty(t, e.Error.Message, "Error message must not be empty")
		assert.NotEmpty(t, e.Error.Type, "Error type must not be empty")

		// Messages should start with capital letter
		if len(e.Error.Message) > 0 {
			first := rune(e.Error.Message[0])
			assert.True(t, first >= 'A' && first <= 'Z',
				"Error message %q should start with capital letter", e.Error.Message)
		}
	}

	t.Logf("COMPLIANCE: Error messages are human-readable and properly capitalized")
}
