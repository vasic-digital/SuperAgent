package llm

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"

	"dev.helix.agent/internal/models"
)

// DefaultMaxConcurrentProviders is the default maximum number of concurrent
// LLM provider calls during ensemble execution. This prevents overwhelming
// the system when many providers are configured.
const DefaultMaxConcurrentProviders = 10

// ensembleMaxConcurrent controls the semaphore size for concurrent provider
// calls. It can be changed at runtime via SetMaxConcurrentProviders.
var ensembleMaxConcurrent int64 = DefaultMaxConcurrentProviders

// SetMaxConcurrentProviders sets the maximum number of concurrent provider
// calls allowed during ensemble execution. A value of 0 or negative resets
// to the default. This is safe for concurrent use.
func SetMaxConcurrentProviders(n int) {
	if n <= 0 {
		n = DefaultMaxConcurrentProviders
	}
	atomic.StoreInt64(&ensembleMaxConcurrent, int64(n))
}

// GetMaxConcurrentProviders returns the current maximum number of concurrent
// provider calls. This is safe for concurrent use.
func GetMaxConcurrentProviders() int {
	return int(atomic.LoadInt64(&ensembleMaxConcurrent))
}

// EnsembleConfig holds configuration for ensemble execution
type EnsembleConfig struct {
	Providers []LLMProvider
}

// RunEnsemble executes a parallel ensemble of LLM providers and returns the aggregated responses.
// IMPORTANT: Use services.ProviderRegistry.GetEnsembleService() for production code.
// This standalone function requires pre-configured providers to be passed in.
func RunEnsemble(req *models.LLMRequest) ([]*models.LLMResponse, *models.LLMResponse, error) {
	return RunEnsembleWithProviders(context.Background(), req, nil)
}

// RunEnsembleWithProviders executes a parallel ensemble with the given providers.
// If providers is nil or empty, returns an error requiring explicit provider configuration.
// Concurrent provider calls are limited by a semaphore (see SetMaxConcurrentProviders).
// The provided context is propagated to all provider calls, enabling cancellation.
func RunEnsembleWithProviders(ctx context.Context, req *models.LLMRequest, providers []LLMProvider) ([]*models.LLMResponse, *models.LLMResponse, error) {
	if req == nil {
		return nil, nil, fmt.Errorf("request cannot be nil")
	}

	if len(providers) == 0 {
		return nil, nil, fmt.Errorf("no providers configured - use services.ProviderRegistry for proper credential injection")
	}

	maxConcurrent := GetMaxConcurrentProviders()
	sem := semaphore.NewWeighted(int64(maxConcurrent))

	var wg sync.WaitGroup
	respCh := make(chan *models.LLMResponse, len(providers))

	for _, p := range providers {
		pp := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Acquire semaphore slot; respects context cancellation
			// while waiting for a slot.
			if err := sem.Acquire(ctx, 1); err != nil {
				return
			}
			defer sem.Release(1)

			r, err := pp.Complete(ctx, req)
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
