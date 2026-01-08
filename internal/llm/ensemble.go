package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/helixagent/helixagent/internal/models"
)

// EnsembleConfig holds configuration for ensemble execution
type EnsembleConfig struct {
	Providers []LLMProvider
}

// RunEnsemble executes a parallel ensemble of LLM providers and returns the aggregated responses.
// IMPORTANT: Use services.ProviderRegistry.GetEnsembleService() for production code.
// This standalone function requires pre-configured providers to be passed in.
func RunEnsemble(req *models.LLMRequest) ([]*models.LLMResponse, *models.LLMResponse, error) {
	return RunEnsembleWithProviders(req, nil)
}

// RunEnsembleWithProviders executes a parallel ensemble with the given providers.
// If providers is nil or empty, returns an error requiring explicit provider configuration.
func RunEnsembleWithProviders(req *models.LLMRequest, providers []LLMProvider) ([]*models.LLMResponse, *models.LLMResponse, error) {
	if req == nil {
		return nil, nil, fmt.Errorf("request cannot be nil")
	}

	if len(providers) == 0 {
		return nil, nil, fmt.Errorf("no providers configured - use services.ProviderRegistry for proper credential injection")
	}

	var wg sync.WaitGroup
	respCh := make(chan *models.LLMResponse, len(providers))

	for _, p := range providers {
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
