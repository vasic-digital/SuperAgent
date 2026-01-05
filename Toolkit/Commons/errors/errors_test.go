package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestProviderError_Error(t *testing.T) {
	err := &ProviderError{
		Provider: "test-provider",
		Code:     "test_code",
		Message:  "test message",
	}

	expected := "[test-provider] test_code: test message"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNewProviderError(t *testing.T) {
	err := NewProviderError("test-provider", "test_code", "test message")

	if err.Provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got %s", err.Provider)
	}

	if err.Code != "test_code" {
		t.Errorf("Expected code 'test_code', got %s", err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", err.Message)
	}
}

func TestProviderError_WithStatusCode(t *testing.T) {
	err := NewProviderError("test", "code", "msg").WithStatusCode(404)

	if err.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", err.StatusCode)
	}
}

func TestProviderError_WithDetails(t *testing.T) {
	details := map[string]string{"key": "value"}
	err := NewProviderError("test", "code", "msg").WithDetails(details)

	if err.Details == nil {
		t.Error("Expected details to be set")
	}

	// Check that the details contain the expected values
	detailsMap, ok := err.Details.(map[string]string)
	if !ok {
		t.Errorf("Expected details to be map[string]string, got %T", err.Details)
	}

	if detailsMap["key"] != "value" {
		t.Errorf("Expected details['key'] = 'value', got %s", detailsMap["key"])
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Type:    "invalid_request",
		Message: "bad request",
	}

	expected := "invalid_request: bad request"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestRateLimitError_Error(t *testing.T) {
	err := &RateLimitError{
		RetryAfter: 30,
		Message:    "too many requests",
	}

	expected := "rate limit exceeded: too many requests (retry after 30 seconds)"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestAuthenticationError_Error(t *testing.T) {
	err := &AuthenticationError{
		Message: "invalid key",
	}

	expected := "authentication failed: invalid key"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "api_key",
		Message: "cannot be empty",
	}

	expected := "validation error for field 'api_key': cannot be empty"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNetworkError_Error(t *testing.T) {
	// Test with underlying error
	underlying := &ValidationError{Field: "test", Message: "error"}
	err := &NetworkError{
		Underlying: underlying,
		Message:    "connection failed",
	}

	result := err.Error()
	if !strings.Contains(result, "network error: connection failed") {
		t.Errorf("Expected error message to contain 'network error: connection failed', got %s", result)
	}
	if !strings.Contains(result, "validation error for field 'test'") {
		t.Errorf("Expected error message to contain underlying error, got %s", result)
	}

	// Test without underlying error
	err = &NetworkError{
		Message: "connection failed",
	}

	expected := "network error: connection failed"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNetworkError_Unwrap(t *testing.T) {
	underlying := &ValidationError{Field: "test", Message: "error"}
	err := &NetworkError{
		Underlying: underlying,
		Message:    "connection failed",
	}

	if err.Unwrap() != underlying {
		t.Errorf("Expected Unwrap to return underlying error")
	}
}

func TestTimeoutError_Error(t *testing.T) {
	err := &TimeoutError{
		Operation: "API call",
		Timeout:   "30s",
	}

	expected := "timeout error in API call after 30s"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNewErrorHandler(t *testing.T) {
	handler := NewErrorHandler("test-provider")

	if handler.provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got %s", handler.provider)
	}
}

func TestErrorHandler_HandleHTTPError_Unauthorized(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusUnauthorized}
	body := []byte("unauthorized")

	err := handler.HandleHTTPError(resp, body)

	authErr, ok := err.(*AuthenticationError)
	if !ok {
		t.Errorf("Expected AuthenticationError, got %T", err)
	}

	if authErr.Message != "invalid API key" {
		t.Errorf("Expected message 'invalid API key', got %s", authErr.Message)
	}
}

func TestErrorHandler_HandleHTTPError_Forbidden(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusForbidden}
	body := []byte("forbidden")

	err := handler.HandleHTTPError(resp, body)

	authErr, ok := err.(*AuthenticationError)
	if !ok {
		t.Errorf("Expected AuthenticationError, got %T", err)
	}

	if authErr.Message != "insufficient permissions" {
		t.Errorf("Expected message 'insufficient permissions', got %s", authErr.Message)
	}
}

func TestErrorHandler_HandleHTTPError_RateLimit(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
	}
	resp.Header.Set("Retry-After", "30")
	body := []byte("rate limit exceeded")

	err := handler.HandleHTTPError(resp, body)

	rateErr, ok := err.(*RateLimitError)
	if !ok {
		t.Errorf("Expected RateLimitError, got %T", err)
	}

	if rateErr.Message != "rate limit exceeded" {
		t.Errorf("Expected message 'rate limit exceeded', got %s", rateErr.Message)
	}

	if rateErr.RetryAfter != 30 {
		t.Errorf("Expected retry after 30, got %d", rateErr.RetryAfter)
	}
}

func TestErrorHandler_HandleHTTPError_BadRequest(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusBadRequest}

	// Test with API error JSON
	apiErr := APIError{
		Type:    "invalid_request",
		Code:    "invalid_param",
		Message: "bad parameter",
	}
	body, _ := json.Marshal(apiErr)

	err := handler.HandleHTTPError(resp, body)

	result, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected APIError, got %T", err)
	}

	if result.Type != "invalid_request" {
		t.Errorf("Expected type 'invalid_request', got %s", result.Type)
	}
}

func TestErrorHandler_HandleHTTPError_ServerError(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusInternalServerError}
	body := []byte("internal server error")

	err := handler.HandleHTTPError(resp, body)

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Errorf("Expected ProviderError, got %T", err)
	}

	if provErr.Code != "server_error" {
		t.Errorf("Expected code 'server_error', got %s", provErr.Code)
	}

	if provErr.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", provErr.StatusCode)
	}
}

func TestErrorHandler_HandleHTTPError_Default(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusNotFound}
	body := []byte("not found")

	err := handler.HandleHTTPError(resp, body)

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Errorf("Expected ProviderError, got %T", err)
	}

	if provErr.Code != "http_error" {
		t.Errorf("Expected code 'http_error', got %s", provErr.Code)
	}

	if provErr.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", provErr.StatusCode)
	}
}

func TestErrorHandler_parseAPIError(t *testing.T) {
	handler := NewErrorHandler("test-provider")

	// Test valid API error
	apiErr := APIError{
		Type:    "invalid_request",
		Code:    "invalid_param",
		Message: "bad parameter",
	}
	body, _ := json.Marshal(apiErr)

	err := handler.parseAPIError(body)

	result, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected APIError, got %T", err)
	}

	if result.Type != "invalid_request" {
		t.Errorf("Expected type 'invalid_request', got %s", result.Type)
	}
}

func TestErrorHandler_parseAPIError_InvalidJSON(t *testing.T) {
	handler := NewErrorHandler("test-provider")

	body := []byte("invalid json")

	err := handler.parseAPIError(body)

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Errorf("Expected ProviderError, got %T", err)
	}

	if provErr.Code != "api_error" {
		t.Errorf("Expected code 'api_error', got %s", provErr.Code)
	}
}

func TestErrorHandler_HandleNetworkError(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	underlying := &ValidationError{Field: "test", Message: "error"}

	err := handler.HandleNetworkError(underlying, "API call")

	netErr, ok := err.(*NetworkError)
	if !ok {
		t.Errorf("Expected NetworkError, got %T", err)
	}

	if netErr.Underlying != underlying {
		t.Errorf("Expected underlying error to be set")
	}

	if !strings.Contains(netErr.Message, "failed during API call") {
		t.Errorf("Expected message to contain 'failed during API call', got %s", netErr.Message)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{&NetworkError{}, true},
		{&RateLimitError{}, true},
		{&TimeoutError{}, true},
		{&ProviderError{StatusCode: 500}, true},
		{&ProviderError{StatusCode: 400}, false},
		{&APIError{Type: "server_error"}, true},
		{&APIError{Type: "rate_limit_error"}, true},
		{&APIError{Type: "invalid_request"}, false},
		{&AuthenticationError{}, false},
		{&ValidationError{}, false},
	}

	for _, test := range tests {
		result := IsRetryable(test.err)
		if result != test.expected {
			t.Errorf("IsRetryable(%T) = %v, expected %v", test.err, result, test.expected)
		}
	}
}

func TestIsRateLimit(t *testing.T) {
	if !IsRateLimit(&RateLimitError{}) {
		t.Error("Expected IsRateLimit to return true for RateLimitError")
	}

	if IsRateLimit(&NetworkError{}) {
		t.Error("Expected IsRateLimit to return false for NetworkError")
	}
}

func TestIsAuth(t *testing.T) {
	if !IsAuth(&AuthenticationError{}) {
		t.Error("Expected IsAuth to return true for AuthenticationError")
	}

	if IsAuth(&NetworkError{}) {
		t.Error("Expected IsAuth to return false for NetworkError")
	}
}

func TestGetRetryAfter(t *testing.T) {
	err := &RateLimitError{RetryAfter: 30}

	if GetRetryAfter(err) != 30 {
		t.Errorf("Expected 30, got %d", GetRetryAfter(err))
	}

	if GetRetryAfter(&NetworkError{}) != 0 {
		t.Errorf("Expected 0 for non-rate-limit error, got %d", GetRetryAfter(&NetworkError{}))
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"30", 30},
		{"", 60},
		{"invalid", 60},
		{"0", 60},
	}

	for _, test := range tests {
		result := parseRetryAfter(test.input)
		if result != test.expected {
			t.Errorf("parseRetryAfter(%s) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestAtoi(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"", 0},
		{"abc", 0},
		{"123abc", 0},
	}

	for _, test := range tests {
		result := atoi(test.input)
		if result != test.expected {
			t.Errorf("atoi(%s) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestProviderError_Chaining(t *testing.T) {
	err := NewProviderError("test-provider", "test_code", "test message").
		WithStatusCode(500).
		WithDetails(map[string]string{"key": "value"})

	if err.Provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got %s", err.Provider)
	}

	if err.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", err.StatusCode)
	}

	if err.Details == nil {
		t.Error("Expected details to be set")
	}
}

func TestAPIError_Fields(t *testing.T) {
	err := &APIError{
		Type:    "invalid_request",
		Code:    "missing_field",
		Message: "field 'name' is required",
		Param:   "name",
		Details: map[string]interface{}{"field": "name"},
	}

	expected := "invalid_request: field 'name' is required"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	if err.Param != "name" {
		t.Errorf("Expected param 'name', got %s", err.Param)
	}
}

func TestNetworkError_NilUnderlying(t *testing.T) {
	err := &NetworkError{
		Underlying: nil,
		Message:    "connection timeout",
	}

	// Unwrap should return nil
	if err.Unwrap() != nil {
		t.Error("Expected Unwrap to return nil")
	}

	expected := "network error: connection timeout"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorHandler_HandleHTTPError_BadGateway(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusBadGateway}
	body := []byte("bad gateway")

	err := handler.HandleHTTPError(resp, body)

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Errorf("Expected ProviderError, got %T", err)
	}

	if provErr.StatusCode != 502 {
		t.Errorf("Expected status code 502, got %d", provErr.StatusCode)
	}

	if provErr.Code != "server_error" {
		t.Errorf("Expected code 'server_error', got %s", provErr.Code)
	}
}

func TestErrorHandler_HandleHTTPError_ServiceUnavailable(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusServiceUnavailable}
	body := []byte("service unavailable")

	err := handler.HandleHTTPError(resp, body)

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Errorf("Expected ProviderError, got %T", err)
	}

	if provErr.StatusCode != 503 {
		t.Errorf("Expected status code 503, got %d", provErr.StatusCode)
	}
}

func TestErrorHandler_HandleHTTPError_GatewayTimeout(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{StatusCode: http.StatusGatewayTimeout}
	body := []byte("gateway timeout")

	err := handler.HandleHTTPError(resp, body)

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Errorf("Expected ProviderError, got %T", err)
	}

	if provErr.StatusCode != 504 {
		t.Errorf("Expected status code 504, got %d", provErr.StatusCode)
	}
}

func TestErrorHandler_HandleHTTPError_RateLimitEmptyHeader(t *testing.T) {
	handler := NewErrorHandler("test-provider")
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
	}
	// No Retry-After header
	body := []byte("rate limit exceeded")

	err := handler.HandleHTTPError(resp, body)

	rateErr, ok := err.(*RateLimitError)
	if !ok {
		t.Errorf("Expected RateLimitError, got %T", err)
	}

	// Default should be 60 seconds
	if rateErr.RetryAfter != 60 {
		t.Errorf("Expected default retry after 60, got %d", rateErr.RetryAfter)
	}
}

func TestErrorHandler_parseAPIError_EmptyType(t *testing.T) {
	handler := NewErrorHandler("test-provider")

	// JSON with empty type field
	body := []byte(`{"code": "error_code", "message": "error message"}`)

	err := handler.parseAPIError(body)

	// Should fall back to ProviderError since Type is empty
	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Errorf("Expected ProviderError for empty type, got %T", err)
	}

	if provErr.Code != "api_error" {
		t.Errorf("Expected code 'api_error', got %s", provErr.Code)
	}
}

func TestIsRetryable_UnknownError(t *testing.T) {
	err := errors.New("unknown error type")

	result := IsRetryable(err)

	if result {
		t.Error("Expected unknown error to not be retryable")
	}
}

func TestIsRetryable_ProviderError4xx(t *testing.T) {
	err := &ProviderError{StatusCode: 404}

	result := IsRetryable(err)

	if result {
		t.Error("Expected 4xx error to not be retryable")
	}
}

func TestIsRetryable_ProviderError5xx(t *testing.T) {
	statusCodes := []int{500, 502, 503, 504, 520}

	for _, code := range statusCodes {
		err := &ProviderError{StatusCode: code}
		result := IsRetryable(err)

		if !result {
			t.Errorf("Expected %d error to be retryable", code)
		}
	}
}

func TestIsRetryable_APIErrorTypes(t *testing.T) {
	retryableTypes := []string{"server_error", "rate_limit_error"}
	nonRetryableTypes := []string{"invalid_request", "authentication_error", "validation_error"}

	for _, errType := range retryableTypes {
		err := &APIError{Type: errType}
		if !IsRetryable(err) {
			t.Errorf("Expected APIError type '%s' to be retryable", errType)
		}
	}

	for _, errType := range nonRetryableTypes {
		err := &APIError{Type: errType}
		if IsRetryable(err) {
			t.Errorf("Expected APIError type '%s' to not be retryable", errType)
		}
	}
}

func TestRateLimitError_ZeroRetryAfter(t *testing.T) {
	err := &RateLimitError{
		RetryAfter: 0,
		Message:    "rate limit hit",
	}

	expected := "rate limit exceeded: rate limit hit (retry after 0 seconds)"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestGetRetryAfter_NonRateLimitErrors(t *testing.T) {
	errors := []error{
		&AuthenticationError{Message: "test"},
		&ValidationError{Field: "test", Message: "test"},
		&TimeoutError{Operation: "test", Timeout: "30s"},
		&ProviderError{Code: "test"},
	}

	for _, err := range errors {
		result := GetRetryAfter(err)
		if result != 0 {
			t.Errorf("Expected GetRetryAfter to return 0 for %T, got %d", err, result)
		}
	}
}

func TestParseRetryAfter_LargeNumber(t *testing.T) {
	result := parseRetryAfter("3600")
	if result != 3600 {
		t.Errorf("Expected 3600, got %d", result)
	}
}

func TestAtoi_LargeNumber(t *testing.T) {
	result := atoi("999999999")
	if result != 999999999 {
		t.Errorf("Expected 999999999, got %d", result)
	}
}

func TestAtoi_LeadingZeros(t *testing.T) {
	result := atoi("007")
	if result != 7 {
		t.Errorf("Expected 7, got %d", result)
	}
}

func TestErrorTypesImplementError(t *testing.T) {
	// Verify all error types implement the error interface
	errors := []error{
		&ProviderError{Provider: "test", Code: "code", Message: "msg"},
		&APIError{Type: "type", Message: "msg"},
		&RateLimitError{Message: "msg", RetryAfter: 60},
		&AuthenticationError{Message: "msg"},
		&ValidationError{Field: "field", Message: "msg"},
		&NetworkError{Message: "msg"},
		&TimeoutError{Operation: "op", Timeout: "30s"},
	}

	for _, err := range errors {
		if err.Error() == "" {
			t.Errorf("Error() returned empty string for %T", err)
		}
	}
}

func TestProviderError_ProviderField(t *testing.T) {
	providers := []string{"openai", "anthropic", "cohere", "huggingface"}

	for _, provider := range providers {
		err := NewProviderError(provider, "test_code", "test message")
		if err.Provider != provider {
			t.Errorf("Expected provider '%s', got '%s'", provider, err.Provider)
		}

		if !strings.Contains(err.Error(), provider) {
			t.Errorf("Expected error message to contain provider name '%s'", provider)
		}
	}
}

func TestValidationError_Fields(t *testing.T) {
	tests := []struct {
		field   string
		message string
	}{
		{"api_key", "is required"},
		{"temperature", "must be between 0 and 2"},
		{"max_tokens", "must be a positive integer"},
	}

	for _, tt := range tests {
		err := &ValidationError{Field: tt.field, Message: tt.message}

		if !strings.Contains(err.Error(), tt.field) {
			t.Errorf("Expected error to contain field name '%s'", tt.field)
		}

		if !strings.Contains(err.Error(), tt.message) {
			t.Errorf("Expected error to contain message '%s'", tt.message)
		}
	}
}

func TestNetworkError_WrappedError(t *testing.T) {
	underlying := &ValidationError{Field: "test", Message: "underlying error"}
	err := &NetworkError{
		Underlying: underlying,
		Message:    "network failure",
	}

	// Test that errors.Is can find the underlying error
	if err.Unwrap() != underlying {
		t.Error("Expected Unwrap to return the underlying error")
	}
}

func TestTimeoutError_Operations(t *testing.T) {
	operations := []string{"API call", "database query", "file upload", "token refresh"}
	timeouts := []string{"30s", "1m", "5m", "30m"}

	for i, op := range operations {
		timeout := timeouts[i]
		err := &TimeoutError{Operation: op, Timeout: timeout}

		if !strings.Contains(err.Error(), op) {
			t.Errorf("Expected error to contain operation '%s'", op)
		}

		if !strings.Contains(err.Error(), timeout) {
			t.Errorf("Expected error to contain timeout '%s'", timeout)
		}
	}
}

func TestErrorHandler_MultipleProviders(t *testing.T) {
	providers := []string{"openai", "anthropic", "cohere"}

	for _, provider := range providers {
		handler := NewErrorHandler(provider)

		if handler.provider != provider {
			t.Errorf("Expected provider '%s', got '%s'", provider, handler.provider)
		}

		// Test that errors from this handler include the provider name
		resp := &http.Response{StatusCode: http.StatusNotFound}
		err := handler.HandleHTTPError(resp, []byte("not found"))

		provErr, ok := err.(*ProviderError)
		if !ok {
			t.Errorf("Expected ProviderError, got %T", err)
			continue
		}

		if provErr.Provider != provider {
			t.Errorf("Expected provider '%s' in error, got '%s'", provider, provErr.Provider)
		}
	}
}

func TestIsRateLimit_AllTypes(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{&RateLimitError{}, true},
		{&AuthenticationError{}, false},
		{&ValidationError{}, false},
		{&ProviderError{}, false},
		{&NetworkError{}, false},
		{&TimeoutError{}, false},
		{&APIError{}, false},
		{errors.New("random error"), false},
	}

	for _, tt := range tests {
		result := IsRateLimit(tt.err)
		if result != tt.expected {
			t.Errorf("IsRateLimit(%T) = %v, expected %v", tt.err, result, tt.expected)
		}
	}
}

func TestIsAuth_AllTypes(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{&AuthenticationError{}, true},
		{&RateLimitError{}, false},
		{&ValidationError{}, false},
		{&ProviderError{}, false},
		{&NetworkError{}, false},
		{&TimeoutError{}, false},
		{&APIError{}, false},
		{errors.New("random error"), false},
	}

	for _, tt := range tests {
		result := IsAuth(tt.err)
		if result != tt.expected {
			t.Errorf("IsAuth(%T) = %v, expected %v", tt.err, result, tt.expected)
		}
	}
}
