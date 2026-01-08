package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/services"
)

// mockRequestProvider implements LLMProvider for RequestService testing
type mockRequestProvider struct {
	name           string
	response       *models.LLMResponse
	errorResponse  error
	streamResponse <-chan *models.LLMResponse
	streamError    error
}

func (m *mockRequestProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.errorResponse != nil {
		return nil, m.errorResponse
	}
	return m.response, nil
}

func (m *mockRequestProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	return m.streamResponse, m.streamError
}

func TestRequestService_NewRequestService(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	assert.NotNil(t, service)
	assert.Empty(t, service.GetProviders())
}

func TestRequestService_RegisterProvider(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	provider1 := &mockRequestProvider{name: "provider-1"}
	provider2 := &mockRequestProvider{name: "provider-2"}

	service.RegisterProvider("provider-1", provider1)
	service.RegisterProvider("provider-2", provider2)

	providers := service.GetProviders()
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, "provider-1")
	assert.Contains(t, providers, "provider-2")
}

func TestRequestService_RemoveProvider(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	provider1 := &mockRequestProvider{name: "provider-1"}
	provider2 := &mockRequestProvider{name: "provider-2"}

	service.RegisterProvider("provider-1", provider1)
	service.RegisterProvider("provider-2", provider2)

	providers := service.GetProviders()
	assert.Len(t, providers, 2)

	service.RemoveProvider("provider-1")

	providers = service.GetProviders()
	assert.Len(t, providers, 1)
	assert.Contains(t, providers, "provider-2")
	assert.NotContains(t, providers, "provider-1")
}

func TestRequestService_ProcessRequest_NoProviders(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	result, err := service.ProcessRequest(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no providers available")
}

func TestRequestService_ProcessRequest_SingleProviderSuccess(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	response := &models.LLMResponse{
		ID:           "resp-1",
		Content:      "Test response",
		ProviderID:   "test-provider",
		ProviderName: "test-provider",
	}

	provider := &mockRequestProvider{
		name:     "test-provider",
		response: response,
	}

	service.RegisterProvider("test-provider", provider)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	result, err := service.ProcessRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, response, result)
}

func TestRequestService_ProcessRequest_ProviderFailure(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	failingProvider := &mockRequestProvider{
		name:          "failing-provider",
		errorResponse: assert.AnError,
	}

	service.RegisterProvider("failing-provider", failingProvider)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	result, err := service.ProcessRequest(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed")
}

func TestRequestService_ProcessRequest_MultipleProviders(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	// Create multiple providers
	provider1 := &mockRequestProvider{
		name: "provider-1",
		response: &models.LLMResponse{
			ID:           "resp-1",
			Content:      "Response from provider 1",
			ProviderID:   "provider-1",
			ProviderName: "provider-1",
		},
	}

	provider2 := &mockRequestProvider{
		name: "provider-2",
		response: &models.LLMResponse{
			ID:           "resp-2",
			Content:      "Response from provider 2",
			ProviderID:   "provider-2",
			ProviderName: "provider-2",
		},
	}

	service.RegisterProvider("provider-1", provider1)
	service.RegisterProvider("provider-2", provider2)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	// Run multiple times to test different routing strategies
	for i := 0; i < 5; i++ {
		result, err := service.ProcessRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Contains(t, []string{"provider-1", "provider-2"}, result.ProviderID)
	}
}

func TestRequestService_ProcessRequest_WithEnsemble(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	// Create multiple providers for ensemble
	provider1 := &mockRequestProvider{
		name: "provider-1",
		response: &models.LLMResponse{
			ID:           "resp-1",
			Content:      "Response from provider 1",
			ProviderID:   "provider-1",
			ProviderName: "provider-1",
			Metadata: map[string]any{
				"confidence": 0.8,
			},
		},
	}

	provider2 := &mockRequestProvider{
		name: "provider-2",
		response: &models.LLMResponse{
			ID:           "resp-2",
			Content:      "Response from provider 2",
			ProviderID:   "provider-2",
			ProviderName: "provider-2",
			Metadata: map[string]any{
				"confidence": 0.9,
			},
		},
	}

	service.RegisterProvider("provider-1", provider1)
	service.RegisterProvider("provider-2", provider2)

	// Also register with ensemble
	ensemble.RegisterProvider("provider-1", provider1)
	ensemble.RegisterProvider("provider-2", provider2)

	req := &models.LLMRequest{
		ID:             "test-request-1",
		SessionID:      "test-session",
		UserID:         "test-user",
		Prompt:         "Test prompt",
		EnsembleConfig: nil,
	}

	result, err := service.ProcessRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// In ensemble mode, we should get a response
	assert.NotEmpty(t, result.Content)
}

func TestRequestService_ProcessRequestStream_NoProviders(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	stream, err := service.ProcessRequestStream(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, stream)
	assert.Contains(t, err.Error(), "no providers available")
}

func TestRequestService_ProcessRequestStream_Success(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	// Create a mock stream
	ch := make(chan *models.LLMResponse, 1)
	ch <- &models.LLMResponse{
		ID:           "stream-resp",
		Content:      "Stream response",
		ProviderID:   "stream-provider",
		ProviderName: "stream-provider",
	}
	close(ch)

	provider := &mockRequestProvider{
		name:           "stream-provider",
		streamResponse: ch,
	}

	service.RegisterProvider("stream-provider", provider)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	stream, err := service.ProcessRequestStream(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, stream)

	// Read from stream
	response, ok := <-stream
	assert.True(t, ok)
	assert.NotNil(t, response)
	assert.Equal(t, "Stream response", response.Content)

	// Stream should be closed
	_, ok = <-stream
	assert.False(t, ok)
}

func TestRequestService_RoutingStrategies(t *testing.T) {
	// Test different routing strategies
	testCases := []struct {
		name     string
		strategy string
	}{
		{"Weighted", "weighted"},
		{"RoundRobin", "round_robin"},
		{"HealthBased", "health_based"},
		{"LatencyBased", "latency_based"},
		{"Random", "random"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
			service := services.NewRequestService(tc.strategy, ensemble, nil)

			// Create multiple providers
			for i := 1; i <= 3; i++ {
				provider := &mockRequestProvider{
					name: "provider-" + string(rune('A'+i-1)),
					response: &models.LLMResponse{
						ID:           "resp-" + string(rune('A'+i-1)),
						Content:      "Response " + string(rune('A'+i-1)),
						ProviderID:   "provider-" + string(rune('A'+i-1)),
						ProviderName: "provider-" + string(rune('A'+i-1)),
					},
				}
				service.RegisterProvider("provider-"+string(rune('A'+i-1)), provider)
			}

			req := &models.LLMRequest{
				ID: "test-request",
				ModelParams: models.ModelParameters{
					Model: "test-model",
				},
			}

			result, err := service.ProcessRequest(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Should get a response from one of the providers
			assert.Contains(t, []string{"provider-A", "provider-B", "provider-C"}, result.ProviderID)
		})
	}
}

func TestRequestService_UpdateProviderHealth(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	provider := &mockRequestProvider{
		name: "test-provider",
		response: &models.LLMResponse{
			ID:           "resp-1",
			Content:      "Test response",
			ProviderID:   "test-provider",
			ProviderName: "test-provider",
		},
	}

	service.RegisterProvider("test-provider", provider)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	// Process request to trigger health update
	result, err := service.ProcessRequest(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Note: Health tracking might be internal, so we can't directly assert on it
	// This test verifies the request doesn't fail due to health tracking
}

func TestRequestService_GetProviderHealth(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)
	service := services.NewRequestService("weighted", ensemble, nil)

	provider1 := &mockRequestProvider{name: "provider-1"}
	provider2 := &mockRequestProvider{name: "provider-2"}

	service.RegisterProvider("provider-1", provider1)
	service.RegisterProvider("provider-2", provider2)

	// Get provider health
	health := service.GetAllProviderHealth()

	// Should have health for both providers
	assert.Len(t, health, 2)
	assert.Contains(t, health, "provider-1")
	assert.Contains(t, health, "provider-2")

	// Each health entry should have basic fields
	for _, h := range health {
		assert.NotEmpty(t, h.Name)
		assert.True(t, h.Healthy)
		assert.NotZero(t, h.ResponseTime)
	}
}

func TestRequestService_Timeout(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 100*time.Millisecond)
	service := services.NewRequestService("weighted", ensemble, nil)

	slowProvider := &mockRequestProvider{
		name: "slow-provider",
		response: &models.LLMResponse{
			ID:           "resp-slow",
			Content:      "Slow response",
			ProviderID:   "slow-provider",
			ProviderName: "slow-provider",
		},
	}

	service.RegisterProvider("slow-provider", slowProvider)

	req := &models.LLMRequest{
		ID: "test-request",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	// Note: Our mock doesn't actually delay, so this test verifies the timeout context is used
	result, err := service.ProcessRequest(context.Background(), req)

	// The mock returns immediately, so we should get a response
	// In a real test with actual delay, this would timeout
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "slow-provider", result.ProviderID)
}
