package modelsdev

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := &ClientConfig{
		APIKey:    "test-key",
		BaseURL:   "https://test.api.com",
		Timeout:   10 * time.Second,
		UserAgent: "TestAgent/1.0",
	}

	client := NewClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "https://test.api.com", client.baseURL)
	assert.Equal(t, "TestAgent/1.0", client.userAgent)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.rateLimiter)
}

func TestNewClientDefaults(t *testing.T) {
	client := NewClient(nil)

	assert.NotNil(t, client)
	assert.Equal(t, DefaultBaseURL, client.baseURL)
	assert.Equal(t, DefaultUserAgent, client.userAgent)
	assert.Equal(t, "", client.apiKey)
	assert.NotNil(t, client.rateLimiter)
}

func TestRateLimiter_Wait_Success(t *testing.T) {
	limiter := NewRateLimiter(10, time.Second)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}
}

func TestRateLimiter_Wait_Exhausted(t *testing.T) {
	limiter := NewRateLimiter(1, time.Second)
	ctx := context.Background()

	err := limiter.Wait(ctx)
	assert.NoError(t, err)

	done := make(chan bool)
	go func() {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
		done <- true
	}()

	select {
	case <-done:
		t.Error("Expected rate limiter to block")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter(5, 100*time.Millisecond)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}

	time.Sleep(150 * time.Millisecond)

	err := limiter.Wait(ctx)
	assert.NoError(t, err)
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Type:    "invalid_request",
		Message: "Bad request",
		Code:    400,
		Details: "Missing field",
	}

	expected := "API error [invalid_request]: Bad request (code: 400, details: Missing field)"
	assert.Equal(t, expected, err.Error())
}

func TestAPIError_Error_WithoutDetails(t *testing.T) {
	err := &APIError{
		Type:    "not_found",
		Message: "Resource not found",
		Code:    404,
	}

	expected := "API error [not_found]: Resource not found (code: 404)"
	assert.Equal(t, expected, err.Error())
}

func TestModelInfo_Capabilities(t *testing.T) {
	model := ModelInfo{
		ID:       "test-model",
		Name:     "Test Model",
		Provider: "test-provider",
		Capabilities: ModelCapabilities{
			Vision:          true,
			FunctionCalling: true,
			Streaming:       true,
		},
	}

	assert.True(t, model.Capabilities.Vision)
	assert.True(t, model.Capabilities.FunctionCalling)
	assert.True(t, model.Capabilities.Streaming)
	assert.False(t, model.Capabilities.Audio)
}

func TestModelPricing(t *testing.T) {
	pricing := ModelPricing{
		InputPrice:  0.00001,
		OutputPrice: 0.00002,
		Currency:    "USD",
		Unit:        "tokens",
	}

	assert.Equal(t, 0.00001, pricing.InputPrice)
	assert.Equal(t, 0.00002, pricing.OutputPrice)
	assert.Equal(t, "USD", pricing.Currency)
	assert.Equal(t, "tokens", pricing.Unit)
}
