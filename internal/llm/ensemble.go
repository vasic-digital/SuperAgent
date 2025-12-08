package llm

import (
	"github.com/superagent/superagent/internal/models"
)

// RunEnsemble executes a simple ensemble of LLM providers and returns the aggregated responses
func RunEnsemble(req *models.LLMRequest) ([]*models.LLMResponse, *models.LLMResponse, error) {
	providers := []LLMProvider{&DeepSeekProvider{}, &ClaudeProvider{}, &GeminiProvider{}, &QwenProvider{}, &ZaiProvider{}}

	var responses []*models.LLMResponse
	var selected *models.LLMResponse

	for _, p := range providers {
		resp, err := p.Complete(req)
		if err != nil {
			continue
		}
		if resp != nil {
			responses = append(responses, resp)
			if selected == nil {
				selected = resp
			}
		}
	}
	return responses, selected, nil
}
