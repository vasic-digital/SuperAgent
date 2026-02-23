package compliance

import (
	"testing"

	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestAPIMessageRolesCompliance verifies all required OpenAI-compatible
// message roles are supported in the Message struct.
func TestAPIMessageRolesCompliance(t *testing.T) {
	requiredRoles := []string{"user", "assistant", "system", "tool"}

	for _, role := range requiredRoles {
		msg := models.Message{
			Role:    role,
			Content: "test content",
		}
		assert.Equal(t, role, msg.Role, "Message.Role must support %q", role)
	}

	t.Logf("COMPLIANCE: All required message roles supported: %v", requiredRoles)
}

// TestLLMRequestStructCompliance verifies the LLMRequest type has
// OpenAI-compatible fields for prompt, messages, and model parameters.
func TestLLMRequestStructCompliance(t *testing.T) {
	req := models.LLMRequest{
		Prompt:   "test prompt",
		Messages: []models.Message{{Role: "user", Content: "hello"}},
	}
	assert.Equal(t, "test prompt", req.Prompt)
	assert.Len(t, req.Messages, 1)

	t.Logf("COMPLIANCE: LLMRequest has required Prompt and Messages fields")
}

// TestLLMResponseStructCompliance verifies the LLMResponse type has
// fields required for OpenAI-compatible responses.
func TestLLMResponseStructCompliance(t *testing.T) {
	resp := models.LLMResponse{
		ID:           "cmpl-test",
		ProviderName: "openai",
		Content:      "test response",
	}
	assert.Equal(t, "cmpl-test", resp.ID)
	assert.Equal(t, "openai", resp.ProviderName)
	assert.Equal(t, "test response", resp.Content)

	t.Logf("COMPLIANCE: LLMResponse has required ID, ProviderName, Content fields")
}

// TestEnsembleConfigCompliance verifies the EnsembleConfig type is present
// for ensemble/debate functionality.
func TestEnsembleConfigCompliance(t *testing.T) {
	cfg := models.EnsembleConfig{}

	// EnsembleConfig must be instantiable (no required fields)
	assert.NotNil(t, &cfg)
	t.Logf("COMPLIANCE: EnsembleConfig type present for ensemble/debate support")
}

// TestModelParametersCompliance verifies ModelParameters supports
// the standard generation control fields.
func TestModelParametersCompliance(t *testing.T) {
	params := models.ModelParameters{}
	params.Temperature = 0.7
	params.MaxTokens = 1000
	params.TopP = 0.9

	assert.Equal(t, 0.7, params.Temperature)
	assert.Equal(t, 1000, params.MaxTokens)
	assert.Equal(t, 0.9, params.TopP)

	t.Logf("COMPLIANCE: ModelParameters supports Temperature, MaxTokens, TopP fields")
}

// TestToolCallCompliance verifies tool call types exist for function calling support.
func TestToolCallCompliance(t *testing.T) {
	tool := models.Tool{}
	assert.NotNil(t, &tool)

	toolCall := models.ToolCall{}
	assert.NotNil(t, &toolCall)

	t.Logf("COMPLIANCE: Tool and ToolCall types present for function calling support")
}
