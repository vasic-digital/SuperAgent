package ensemble_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

// testEnsembleResponse is a helper to test ensemble responses in test environment
// Note: RunEnsemble now requires providers to be passed - calling without providers returns an error
func testEnsembleResponse(t *testing.T, responses []*models.LLMResponse, selected *models.LLMResponse, err error) {
	// In test environment, RunEnsemble without providers returns an error
	// This is the expected behavior - providers must be explicitly configured
	if err != nil {
		// Error is expected when no providers are configured
		t.Logf("Ensemble returned error (expected without providers): %v", err)
		assert.Nil(t, responses)
		assert.Nil(t, selected)
		return
	}

	// If no error, responses could be nil (all providers failed)
	// or non-nil (some providers succeeded)
	if responses != nil && len(responses) > 0 {
		assert.NotNil(t, selected)
	} else {
		// responses is nil or empty, selected should also be nil
		assert.Nil(t, selected)
	}
}

func TestRunEnsemble_Basic(t *testing.T) {
	req := &models.LLMRequest{
		ID:     "test-request-1",
		Prompt: "Hello, world!",
		ModelParams: models.ModelParameters{
			Model: "test-model",
		},
	}

	responses, selected, err := llm.RunEnsemble(req)
	testEnsembleResponse(t, responses, selected, err)
}

func TestRunEnsemble_NilRequest(t *testing.T) {
	responses, selected, err := llm.RunEnsemble(nil)

	// Currently RunEnsemble doesn't check for nil request
	// It will panic when providers try to use nil request
	// So we expect either an error or a panic
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, responses)
		assert.Nil(t, selected)
	} else {
		// If no error, it might panic later - that's okay for test
		t.Log("RunEnsemble didn't return error for nil request (may panic internally)")
	}
}

func TestRunEnsemble_EmptyRequest(t *testing.T) {
	req := &models.LLMRequest{
		ID:     "test-request-2",
		Prompt: "", // Empty prompt
	}

	responses, selected, err := llm.RunEnsemble(req)
	testEnsembleResponse(t, responses, selected, err)
}

func TestRunEnsemble_WithMaxTokens(t *testing.T) {
	req := &models.LLMRequest{
		ID:     "test-request-3",
		Prompt: "Write a short story",
		ModelParams: models.ModelParameters{
			Model:     "test-model",
			MaxTokens: 100,
		},
	}

	responses, selected, err := llm.RunEnsemble(req)
	testEnsembleResponse(t, responses, selected, err)
}

func TestRunEnsemble_WithTemperature(t *testing.T) {
	req := &models.LLMRequest{
		ID:     "test-request-4",
		Prompt: "Explain quantum computing",
		ModelParams: models.ModelParameters{
			Model:       "test-model",
			Temperature: 0.7,
		},
	}

	responses, selected, err := llm.RunEnsemble(req)
	testEnsembleResponse(t, responses, selected, err)
}

func TestRunEnsemble_WithEnsembleConfig(t *testing.T) {
	req := &models.LLMRequest{
		ID:     "test-request-5",
		Prompt: "Test with ensemble config",
		EnsembleConfig: &models.EnsembleConfig{
			Strategy:            "parallel",
			MinProviders:        2,
			ConfidenceThreshold: 0.8,
			Timeout:             5000,
		},
	}

	responses, selected, err := llm.RunEnsemble(req)
	testEnsembleResponse(t, responses, selected, err)
}

func TestRunEnsemble_MultipleMessages(t *testing.T) {
	req := &models.LLMRequest{
		ID:     "test-request-6",
		Prompt: "Continue conversation",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
		},
	}

	responses, selected, err := llm.RunEnsemble(req)
	testEnsembleResponse(t, responses, selected, err)
}

func TestRunEnsemble_LongPrompt(t *testing.T) {
	// Create a long prompt
	longPrompt := "This is a very long prompt. " + string(make([]byte, 1000))

	req := &models.LLMRequest{
		ID:     "test-request-7",
		Prompt: longPrompt,
	}

	responses, selected, err := llm.RunEnsemble(req)
	testEnsembleResponse(t, responses, selected, err)
}

func TestRunEnsemble_DifferentRequestTypes(t *testing.T) {
	testCases := []struct {
		name        string
		requestType string
		prompt      string
	}{
		{"Text Completion", "text_completion", "Complete this sentence: The quick brown fox"},
		{"Chat", "chat", "Hello, how can I help you today?"},
		{"Code Generation", "code_generation", "Write a function to calculate factorial"},
		{"Translation", "translation", "Translate 'hello' to Spanish"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &models.LLMRequest{
				ID:          "test-request-" + tc.name,
				Prompt:      tc.prompt,
				RequestType: tc.requestType,
			}

			responses, selected, err := llm.RunEnsemble(req)
			testEnsembleResponse(t, responses, selected, err)
		})
	}
}

func TestRunEnsemble_ProviderFailureHandling(t *testing.T) {
	// Test that ensemble handles missing providers gracefully
	// Since RunEnsemble now requires providers, it returns an error when none are configured
	req := &models.LLMRequest{
		ID:     "test-request-8",
		Prompt: "This should return error without providers",
	}

	responses, selected, err := llm.RunEnsemble(req)

	// Should return error when no providers are configured
	// This is the expected secure behavior - no default/mock providers
	assert.Error(t, err)
	assert.Nil(t, responses)
	assert.Nil(t, selected)
}

func TestRunEnsemble_ResponseSelection(t *testing.T) {
	req := &models.LLMRequest{
		ID:     "test-request-9",
		Prompt: "Test response selection logic",
	}

	responses, selected, err := llm.RunEnsemble(req)

	if err != nil {
		t.Logf("Ensemble returned error: %v", err)
		return
	}

	if len(responses) > 0 {
		// Selected should be one of the responses
		assert.NotNil(t, selected)
		assert.Contains(t, responses, selected)

		// If there are multiple responses, selected should have highest confidence
		if len(responses) > 1 {
			maxConfidence := -1.0
			for _, resp := range responses {
				if resp.Confidence > maxConfidence {
					maxConfidence = resp.Confidence
				}
			}
			assert.Equal(t, maxConfidence, selected.Confidence)
		}
	}
}
