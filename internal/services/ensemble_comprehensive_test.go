package services

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMProviderForEnsemble is a mock LLM provider for ensemble testing
type MockLLMProviderForEnsemble struct {
	CompleteFunc       func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)
	CompleteStreamFunc func(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error)
	delay              time.Duration
}

func (m *MockLLMProviderForEnsemble) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, req)
	}
	return &models.LLMResponse{
		ID:         "mock-response",
		Content:    "mock content",
		Confidence: 0.8,
	}, nil
}

func (m *MockLLMProviderForEnsemble) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if m.CompleteStreamFunc != nil {
		return m.CompleteStreamFunc(ctx, req)
	}
	ch := make(chan *models.LLMResponse, 1)
	ch <- &models.LLMResponse{
		ID:         "mock-stream",
		Content:    "mock stream content",
		Confidence: 0.8,
	}
	close(ch)
	return ch, nil
}

var _ LLMProvider = (*MockLLMProviderForEnsemble)(nil)

func TestNewEnsembleService(t *testing.T) {
	t.Run("creates service with default values", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		assert.NotNil(t, service)
		assert.Equal(t, "confidence_weighted", service.strategy)
		assert.Equal(t, 30*time.Second, service.timeout)
		assert.NotNil(t, service.providers)
		assert.Empty(t, service.providers)
	})

	t.Run("creates service with different strategy", func(t *testing.T) {
		service := NewEnsembleService("majority_vote", 60*time.Second)

		assert.Equal(t, "majority_vote", service.strategy)
		assert.Equal(t, 60*time.Second, service.timeout)
	})
}

func TestEnsembleService_RegisterProvider(t *testing.T) {
	t.Run("registers provider successfully", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)
		provider := &MockLLMProviderForEnsemble{}

		service.RegisterProvider("test-provider", provider)

		providers := service.GetProviders()
		assert.Contains(t, providers, "test-provider")
	})

	t.Run("registers multiple providers", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		service.RegisterProvider("provider-1", &MockLLMProviderForEnsemble{})
		service.RegisterProvider("provider-2", &MockLLMProviderForEnsemble{})
		service.RegisterProvider("provider-3", &MockLLMProviderForEnsemble{})

		providers := service.GetProviders()
		assert.Len(t, providers, 3)
	})

	t.Run("updates existing provider", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		provider1 := &MockLLMProviderForEnsemble{}
		provider2 := &MockLLMProviderForEnsemble{}

		service.RegisterProvider("test-provider", provider1)
		service.RegisterProvider("test-provider", provider2)

		// Should only have one provider with the updated instance
		providers := service.GetProviders()
		assert.Len(t, providers, 1)
	})
}

func TestEnsembleService_RemoveProvider(t *testing.T) {
	t.Run("removes registered provider", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)
		service.RegisterProvider("test-provider", &MockLLMProviderForEnsemble{})

		service.RemoveProvider("test-provider")

		providers := service.GetProviders()
		assert.NotContains(t, providers, "test-provider")
		assert.Empty(t, providers)
	})

	t.Run("removing non-existent provider is safe", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		// Should not panic
		service.RemoveProvider("non-existent")

		providers := service.GetProviders()
		assert.Empty(t, providers)
	})
}

func TestEnsembleService_GetProviders(t *testing.T) {
	t.Run("returns empty slice when no providers", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		providers := service.GetProviders()
		assert.Empty(t, providers)
	})

	t.Run("returns all registered providers", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)
		service.RegisterProvider("provider-a", &MockLLMProviderForEnsemble{})
		service.RegisterProvider("provider-b", &MockLLMProviderForEnsemble{})

		providers := service.GetProviders()
		assert.Len(t, providers, 2)
	})

	t.Run("returns defensive copy", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)
		service.RegisterProvider("provider-a", &MockLLMProviderForEnsemble{})

		providers1 := service.GetProviders()
		providers2 := service.GetProviders()

		// Modifying one should not affect the other
		providers1 = append(providers1, "extra")
		assert.Len(t, providers2, 1)
	})
}

func TestEnsembleService_RunEnsemble(t *testing.T) {
	t.Run("returns error when no providers", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		req := &models.LLMRequest{Prompt: "test"}
		result, err := service.RunEnsemble(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &LLMServiceError{}, err)
	})

	t.Run("runs ensemble with single provider", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)
		service.RegisterProvider("single-provider", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:         "resp-1",
					Content:    "response content",
					Confidence: 0.9,
					ProviderID: "single-provider",
				}, nil
			},
		})

		req := &models.LLMRequest{Prompt: "test"}
		result, err := service.RunEnsemble(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Responses, 1)
		assert.Equal(t, "single-provider", result.Selected.ProviderName)
	})

	t.Run("runs ensemble with multiple providers", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		service.RegisterProvider("provider-1", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:         "resp-1",
					Content:    "response 1",
					Confidence: 0.7,
					ProviderID: "provider-1",
				}, nil
			},
		})

		service.RegisterProvider("provider-2", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:         "resp-2",
					Content:    "response 2",
					Confidence: 0.9,
					ProviderID: "provider-2",
				}, nil
			},
		})

		req := &models.LLMRequest{Prompt: "test"}
		result, err := service.RunEnsemble(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Responses, 2)
		// Should select higher confidence response
		assert.Equal(t, "resp-2", result.Selected.ID)
	})

	t.Run("handles provider errors gracefully", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		service.RegisterProvider("failing-provider", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, errors.New("provider failed")
			},
		})

		service.RegisterProvider("working-provider", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:         "resp-1",
					Content:    "working response",
					Confidence: 0.8,
					ProviderID: "working-provider",
				}, nil
			},
		})

		req := &models.LLMRequest{Prompt: "test"}
		result, err := service.RunEnsemble(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Responses, 1)
		assert.Equal(t, "working-provider", result.Selected.ProviderName)
	})

	t.Run("returns error when all providers fail", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		service.RegisterProvider("provider-1", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, errors.New("error 1")
			},
		})

		service.RegisterProvider("provider-2", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, errors.New("error 2")
			},
		})

		req := &models.LLMRequest{Prompt: "test"}
		result, err := service.RunEnsemble(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, *AllProvidersFailedError{}, err)
	})

	t.Run("respects timeout", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 50*time.Millisecond)

		service.RegisterProvider("slow-provider", &MockLLMProviderForEnsemble{
			delay: 200 * time.Millisecond,
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:         "slow",
					Content:    "slow response",
					Confidence: 0.9,
				}, nil
			},
		})

		req := &models.LLMRequest{Prompt: "test"}
		start := time.Now()
		result, err := service.RunEnsemble(context.Background(), req)
		duration := time.Since(start)

		// Should timeout and return error
		require.Error(t, err)
		assert.Nil(t, result)
		// Should complete near timeout, not after slow provider finishes
		assert.Less(t, duration, 150*time.Millisecond)
	})

	t.Run("filters providers by preference", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		service.RegisterProvider("provider-a", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{ID: "a", Content: "a", Confidence: 0.5}, nil
			},
		})

		service.RegisterProvider("provider-b", &MockLLMProviderForEnsemble{
			CompleteFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{ID: "b", Content: "b", Confidence: 0.9}, nil
			},
		})

		req := &models.LLMRequest{
			Prompt: "test",
			EnsembleConfig: &models.EnsembleConfig{
				PreferredProviders: []string{"provider-a"},
			},
		}
		result, err := service.RunEnsemble(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		// Should only use preferred provider
		assert.Len(t, result.Responses, 1)
	})
}

func TestEnsembleService_RunEnsembleStream(t *testing.T) {
	t.Run("returns error when no providers for streaming", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		req := &models.LLMRequest{Prompt: "test"}
		stream, err := service.RunEnsembleStream(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, stream)
	})

	t.Run("streams from single provider", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		service.RegisterProvider("stream-provider", &MockLLMProviderForEnsemble{
			CompleteStreamFunc: func(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
				ch := make(chan *models.LLMResponse, 2)
				ch <- &models.LLMResponse{ID: "chunk-1", Content: "Hello", Confidence: 0.8}
				ch <- &models.LLMResponse{ID: "chunk-2", Content: " World", Confidence: 0.8}
				close(ch)
				return ch, nil
			},
		})

		req := &models.LLMRequest{Prompt: "test"}
		stream, err := service.RunEnsembleStream(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, stream)

		var chunks []string
		for resp := range stream {
			chunks = append(chunks, resp.Content)
		}

		assert.Equal(t, []string{"Hello", " World"}, chunks)
	})
}

func TestConfidenceWeightedStrategy(t *testing.T) {
	t.Run("votes on single response", func(t *testing.T) {
		strategy := &ConfidenceWeightedStrategy{}
		responses := []*models.LLMResponse{
			{ID: "resp-1", Content: "content", Confidence: 0.8},
		}

		selected, scores, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Equal(t, "resp-1", selected.ID)
		assert.NotEmpty(t, scores)
	})

	t.Run("selects highest confidence response", func(t *testing.T) {
		strategy := &ConfidenceWeightedStrategy{}
		responses := []*models.LLMResponse{
			{ID: "low", Content: "low confidence", Confidence: 0.5},
			{ID: "high", Content: "high confidence", Confidence: 0.9},
			{ID: "medium", Content: "medium confidence", Confidence: 0.7},
		}

		selected, scores, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.Equal(t, "high", selected.ID)
		assert.Equal(t, 0.9, scores["high"])
	})

	t.Run("returns error for empty responses", func(t *testing.T) {
		strategy := &ConfidenceWeightedStrategy{}

		selected, scores, err := strategy.Vote([]*models.LLMResponse{}, &models.LLMRequest{})

		require.Error(t, err)
		assert.Nil(t, selected)
		assert.Nil(t, scores)
	})

	t.Run("applies preferred provider weights", func(t *testing.T) {
		strategy := &ConfidenceWeightedStrategy{}
		responses := []*models.LLMResponse{
			{ID: "resp-1", Content: "content", Confidence: 0.8, ProviderName: "provider-a"},
			{ID: "resp-2", Content: "content", Confidence: 0.8, ProviderName: "provider-b"},
		}

		req := &models.LLMRequest{
			EnsembleConfig: &models.EnsembleConfig{
				PreferredProviders: []string{"provider-b", "provider-a"},
			},
		}

		selected, _, err := strategy.Vote(responses, req)

		require.NoError(t, err)
		// provider-b is preferred, so it should win despite same confidence
		assert.Equal(t, "resp-2", selected.ID)
	})
}

func TestConfidenceWeightedStrategy_applyQualityWeights(t *testing.T) {
	strategy := &ConfidenceWeightedStrategy{}

	t.Run("applies length factor for optimal length", func(t *testing.T) {
		resp := &models.LLMResponse{
			Content:    string(make([]byte, 500)), // 500 chars
			Confidence: 1.0,
		}

		score := strategy.applyQualityWeights(resp, 1.0)
		// Should boost score for optimal length
		assert.Greater(t, score, 1.0)
	})

	t.Run("reduces score for very long content", func(t *testing.T) {
		resp := &models.LLMResponse{
			Content:    string(make([]byte, 3000)), // 3000 chars
			Confidence: 1.0,
		}

		score := strategy.applyQualityWeights(resp, 1.0)
		// Should reduce score for very long content
		assert.Less(t, score, 1.0)
	})

	t.Run("applies response time factor", func(t *testing.T) {
		fastResp := &models.LLMResponse{
			Content:      "fast",
			Confidence:   1.0,
			ResponseTime: 500, // 500ms
		}
		slowResp := &models.LLMResponse{
			Content:      "slow",
			Confidence:   1.0,
			ResponseTime: 15000, // 15s
		}

		fastScore := strategy.applyQualityWeights(fastResp, 1.0)
		slowScore := strategy.applyQualityWeights(slowResp, 1.0)

		assert.Greater(t, fastScore, slowScore)
	})

	t.Run("applies finish reason factor", func(t *testing.T) {
		stopResp := &models.LLMResponse{
			Content:      "complete",
			Confidence:   1.0,
			FinishReason: "stop",
		}
		filterResp := &models.LLMResponse{
			Content:      "filtered",
			Confidence:   1.0,
			FinishReason: "content_filter",
		}

		stopScore := strategy.applyQualityWeights(stopResp, 1.0)
		filterScore := strategy.applyQualityWeights(filterResp, 1.0)

		assert.Greater(t, stopScore, filterScore)
	})
}

func TestMajorityVoteStrategy(t *testing.T) {
	t.Run("finds majority by content similarity", func(t *testing.T) {
		strategy := &MajorityVoteStrategy{}
		responses := []*models.LLMResponse{
			{ID: "resp-1", Content: "agreed answer", Confidence: 0.7},
			{ID: "resp-2", Content: "agreed answer", Confidence: 0.8},
			{ID: "resp-3", Content: "different answer", Confidence: 0.9},
		}

		selected, scores, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.Equal(t, "resp-1", selected.ID) // First in majority group
		assert.NotEmpty(t, scores)
	})

	t.Run("falls back to confidence when no majority", func(t *testing.T) {
		strategy := &MajorityVoteStrategy{}
		responses := []*models.LLMResponse{
			{ID: "resp-1", Content: "answer a", Confidence: 0.9},
			{ID: "resp-2", Content: "answer b", Confidence: 0.5},
			{ID: "resp-3", Content: "answer c", Confidence: 0.7},
		}

		selected, _, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		// Should select highest confidence
		assert.Equal(t, "resp-1", selected.ID)
	})

	t.Run("returns error for empty responses", func(t *testing.T) {
		strategy := &MajorityVoteStrategy{}

		selected, scores, err := strategy.Vote([]*models.LLMResponse{}, &models.LLMRequest{})

		require.Error(t, err)
		assert.Nil(t, selected)
		assert.Nil(t, scores)
	})

	t.Run("handles single response", func(t *testing.T) {
		strategy := &MajorityVoteStrategy{}
		responses := []*models.LLMResponse{
			{ID: "resp-1", Content: "only answer", Confidence: 0.8},
		}

		selected, _, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.Equal(t, "resp-1", selected.ID)
		assert.True(t, selected.Selected)
	})
}

func TestQualityWeightedStrategy(t *testing.T) {
	t.Run("calculates quality scores", func(t *testing.T) {
		strategy := &QualityWeightedStrategy{}
		responses := []*models.LLMResponse{
			{ID: "resp-1", Content: "good response", Confidence: 0.9, ResponseTime: 1000, TokensUsed: 50},
			{ID: "resp-2", Content: "poor response", Confidence: 0.5, ResponseTime: 5000, TokensUsed: 100},
		}

		selected, scores, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Len(t, scores, 2)
		// resp-1 should have higher quality score
		assert.Greater(t, scores["resp-1"], scores["resp-2"])
	})

	t.Run("returns error for empty responses", func(t *testing.T) {
		strategy := &QualityWeightedStrategy{}

		selected, scores, err := strategy.Vote([]*models.LLMResponse{}, &models.LLMRequest{})

		require.Error(t, err)
		assert.Nil(t, selected)
		assert.Nil(t, scores)
	})
}

func TestQualityWeightedStrategy_calculateQualityScore(t *testing.T) {
	strategy := &QualityWeightedStrategy{}

	t.Run("calculates complete quality score", func(t *testing.T) {
		resp := &models.LLMResponse{
			Content:      string(make([]byte, 500)), // Good length
			Confidence:   0.9,
			ResponseTime: 1000, // 1 second
			TokensUsed:   100,
			FinishReason: "stop",
		}

		score := strategy.calculateQualityScore(resp)

		// Score should be between 0 and 1
		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})

	t.Run("handles zero tokens used", func(t *testing.T) {
		resp := &models.LLMResponse{
			Content:      "response",
			Confidence:   0.5,
			ResponseTime: 2000,
			TokensUsed:   0,
			FinishReason: "stop",
		}

		score := strategy.calculateQualityScore(resp)

		// Should not panic or return NaN
		assert.False(t, math.IsNaN(score))
		assert.GreaterOrEqual(t, score, 0.0)
	})
}

func TestEnsembleService_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent provider registration", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := string(rune('a' + idx))
				service.RegisterProvider(name, &MockLLMProviderForEnsemble{})
			}(i)
		}

		wg.Wait()

		providers := service.GetProviders()
		assert.Len(t, providers, 10)
	})

	t.Run("handles concurrent read and write", func(t *testing.T) {
		service := NewEnsembleService("confidence_weighted", 30*time.Second)

		// Pre-populate
		for i := 0; i < 5; i++ {
			service.RegisterProvider(string(rune('a'+i)), &MockLLMProviderForEnsemble{})
		}

		var wg sync.WaitGroup

		// Concurrent reads
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = service.GetProviders()
			}()
		}

		// Concurrent writes
		for i := 5; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := string(rune('a' + idx))
				service.RegisterProvider(name, &MockLLMProviderForEnsemble{})
			}(i)
		}

		wg.Wait()
	})
}
