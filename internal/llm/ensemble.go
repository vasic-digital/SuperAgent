package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/superagent/superagent/internal/llm/providers/claude"
	"github.com/superagent/superagent/internal/llm/providers/deepseek"
	"github.com/superagent/superagent/internal/llm/providers/gemini"
	"github.com/superagent/superagent/internal/llm/providers/ollama"
	"github.com/superagent/superagent/internal/llm/providers/qwen"
	"github.com/superagent/superagent/internal/llm/providers/zai"
	"github.com/superagent/superagent/internal/models"
)

// RunEnsemble executes a parallel ensemble of LLM providers and returns the aggregated responses
func RunEnsemble(req *models.LLMRequest) ([]*models.LLMResponse, *models.LLMResponse, error) {
	if req == nil {
		return nil, nil, fmt.Errorf("request cannot be nil")
	}

	// Initialize providers with default configurations
	provs := []LLMProvider{
		ollama.NewOllamaProvider("", ""),
		deepseek.NewDeepSeekProvider("", "", ""),
		claude.NewClaudeProvider("", "", ""),
		gemini.NewGeminiProvider("", "", ""),
		qwen.NewQwenProvider("", "", ""),
		zai.NewZAIProvider("", "", ""),
	}

	var wg sync.WaitGroup
	respCh := make(chan *models.LLMResponse, len(provs))

	for _, p := range provs {
		pp := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := pp.Complete(context.Background(), req)
			if err == nil && r != nil {
				respCh <- r
			}
		}()
	}

	go func() {
		wg.Wait()
		close(respCh)
	}()

	var responses []*models.LLMResponse
	for r := range respCh {
		responses = append(responses, r)
	}

	// Choose the best by highest confidence if available
	var selected *models.LLMResponse
	max := -1.0
	for _, r := range responses {
		if r != nil && r.Confidence > max {
			max = r.Confidence
			selected = r
		}
	}
	if len(responses) == 0 {
		return nil, nil, nil
	}
	return responses, selected, nil
}
