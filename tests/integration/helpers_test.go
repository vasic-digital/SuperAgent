package integration

import (
	"context"
	"net/http"
	"os"
	"testing"

	"dev.helix.agent/internal/models"
)

// =============================================================================
// Shared test constants and helper functions
// =============================================================================

const (
	// HelixAgentBaseURL is the HelixAgent service URL for integration tests
	HelixAgentBaseURL = "http://localhost:7061"
	// TestEmail is the test user email for authentication
	TestEmail = "admin@helixagent.ai"
	// TestPassword is the test user password for authentication
	TestPassword = "HelixAgentPass123"
)

// getTestAPIKey returns the API key for testing, falling back to a default test key
func getTestAPIKey() string {
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey == "" {
		apiKey = "sk-bd15ed2afe4c4f62a7e8b9c10d4e5f6a"
	}
	return apiKey
}

// checkAuthAndHandleFailure checks if the response indicates auth failure.
// Returns true if test should continue normally, false if test should be skipped due to auth issues.
func checkAuthAndHandleFailure(t *testing.T, resp *http.Response, body []byte, endpoint string) bool {
	t.Helper()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		t.Logf("HelixAgent auth not configured for test API key on %s (status %d) - skipping proxy test", endpoint, resp.StatusCode)
		return false
	}
	return true
}

// =============================================================================
// Mock LLM Provider for ensemble and memory integration tests
// =============================================================================

// MockBaseLLMProvider is a basic mock LLM provider for integration testing
type MockBaseLLMProvider struct {
	name            string
	completeFunc    func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)
	streamFunc      func(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error)
	healthCheckFunc func() error
}

func (m *MockBaseLLMProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, req)
	}
	return &models.LLMResponse{
		ID:      "mock-response",
		Content: "Mock response from " + m.name,
	}, nil
}

func (m *MockBaseLLMProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, req)
	}
	ch := make(chan *models.LLMResponse, 1)
	go func() {
		defer close(ch)
		resp, _ := m.Complete(ctx, req)
		ch <- resp
	}()
	return ch, nil
}

func (m *MockBaseLLMProvider) HealthCheck() error {
	if m.healthCheckFunc != nil {
		return m.healthCheckFunc()
	}
	return nil
}

func (m *MockBaseLLMProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportsStreaming: true,
		Metadata:         make(map[string]string),
	}
}

func (m *MockBaseLLMProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, nil
}
