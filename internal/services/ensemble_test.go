package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superagent/superagent/internal/models"
)

// MockLLMProvider implements LLMProvider for testing
type MockLLMProvider struct {
	name       string
	response   *models.LLMResponse
	err        error
	delay      time.Duration
	streamResp []*models.LLMResponse
	streamErr  error
}

func (m *MockLLMProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delay):
		}
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *MockLLMProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if m.streamErr != nil {
		return nil, m.streamErr
	}

	ch := make(chan *models.LLMResponse, len(m.streamResp))
	go func() {
		defer close(ch)
		for _, resp := range m.streamResp {
			select {
			case <-ctx.Done():
				return
			case ch <- resp:
			}
		}
	}()
	return ch, nil
}

func newMockProvider(name string, content string, confidence float64) *MockLLMProvider {
	return &MockLLMProvider{
		name: name,
		response: &models.LLMResponse{
			ID:           name + "-resp",
			Content:      content,
			Confidence:   confidence,
			ProviderID:   name,
			ProviderName: name,
			TokensUsed:   100,
			ResponseTime: 500,
			FinishReason: "stop",
		},
	}
}

func TestNewEnsembleService(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)

	require.NotNil(t, service)
	assert.Equal(t, "confidence_weighted", service.strategy)
	assert.Equal(t, 30*time.Second, service.timeout)
	assert.NotNil(t, service.providers)
}

func TestEnsembleService_RegisterProvider(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)

	provider := newMockProvider("test-provider", "test response", 0.9)
	service.RegisterProvider("test-provider", provider)

	providers := service.GetProviders()
	assert.Len(t, providers, 1)
	assert.Contains(t, providers, "test-provider")
}

func TestEnsembleService_RemoveProvider(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)

	service.RegisterProvider("provider1", newMockProvider("provider1", "response1", 0.9))
	service.RegisterProvider("provider2", newMockProvider("provider2", "response2", 0.8))

	assert.Len(t, service.GetProviders(), 2)

	service.RemoveProvider("provider1")

	assert.Len(t, service.GetProviders(), 1)
	assert.Contains(t, service.GetProviders(), "provider2")
}

func TestEnsembleService_GetProviders(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)

	t.Run("empty providers", func(t *testing.T) {
		providers := service.GetProviders()
		assert.Empty(t, providers)
	})

	t.Run("with providers", func(t *testing.T) {
		service.RegisterProvider("p1", newMockProvider("p1", "r1", 0.9))
		service.RegisterProvider("p2", newMockProvider("p2", "r2", 0.8))

		providers := service.GetProviders()
		assert.Len(t, providers, 2)
	})
}

func TestEnsembleService_RunEnsemble_NoProviders(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	ctx := context.Background()

	result, err := service.RunEnsemble(ctx, &models.LLMRequest{
		Prompt: "test",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no providers available")
}

func TestEnsembleService_RunEnsemble_SingleProvider(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	ctx := context.Background()

	service.RegisterProvider("provider1", newMockProvider("provider1", "test response", 0.9))

	result, err := service.RunEnsemble(ctx, &models.LLMRequest{
		Prompt: "test",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Responses, 1)
	assert.NotNil(t, result.Selected)
	assert.Equal(t, "test response", result.Selected.Content)
}

func TestEnsembleService_RunEnsemble_MultipleProviders(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	ctx := context.Background()

	service.RegisterProvider("provider1", newMockProvider("provider1", "response1", 0.7))
	service.RegisterProvider("provider2", newMockProvider("provider2", "response2", 0.9))
	service.RegisterProvider("provider3", newMockProvider("provider3", "response3", 0.8))

	result, err := service.RunEnsemble(ctx, &models.LLMRequest{
		Prompt: "test",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Responses, 3)
	assert.NotNil(t, result.Selected)
	assert.Equal(t, "confidence_weighted", result.VotingMethod)
}

func TestEnsembleService_RunEnsemble_ProviderError(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	ctx := context.Background()

	goodProvider := newMockProvider("good", "good response", 0.9)
	badProvider := &MockLLMProvider{
		name: "bad",
		err:  errors.New("provider failed"),
	}

	service.RegisterProvider("good", goodProvider)
	service.RegisterProvider("bad", badProvider)

	result, err := service.RunEnsemble(ctx, &models.LLMRequest{
		Prompt: "test",
	})

	// Should still succeed if at least one provider works
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Responses, 1)
}

func TestEnsembleService_RunEnsemble_AllProvidersFail(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	ctx := context.Background()

	badProvider1 := &MockLLMProvider{err: errors.New("failed1")}
	badProvider2 := &MockLLMProvider{err: errors.New("failed2")}

	service.RegisterProvider("bad1", badProvider1)
	service.RegisterProvider("bad2", badProvider2)

	result, err := service.RunEnsemble(ctx, &models.LLMRequest{
		Prompt: "test",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestEnsembleService_RunEnsemble_WithPreferredProviders(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	ctx := context.Background()

	service.RegisterProvider("provider1", newMockProvider("provider1", "response1", 0.9))
	service.RegisterProvider("provider2", newMockProvider("provider2", "response2", 0.8))
	service.RegisterProvider("provider3", newMockProvider("provider3", "response3", 0.7))

	result, err := service.RunEnsemble(ctx, &models.LLMRequest{
		Prompt: "test",
		EnsembleConfig: &models.EnsembleConfig{
			PreferredProviders: []string{"provider2"},
			MinProviders:       1,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	// Should only have preferred provider
	assert.GreaterOrEqual(t, len(result.Responses), 1)
}

func TestConfidenceWeightedStrategy_Vote(t *testing.T) {
	strategy := &ConfidenceWeightedStrategy{}

	t.Run("no responses", func(t *testing.T) {
		_, _, err := strategy.Vote([]*models.LLMResponse{}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no responses")
	})

	t.Run("single response", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "response", Confidence: 0.9},
		}
		selected, scores, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Equal(t, "1", selected.ID)
		assert.Contains(t, scores, "1")
	})

	t.Run("multiple responses selects highest", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "response1", Confidence: 0.5, FinishReason: "stop", TokensUsed: 50, ResponseTime: 500},
			{ID: "2", Content: "response2", Confidence: 0.9, FinishReason: "stop", TokensUsed: 50, ResponseTime: 500},
			{ID: "3", Content: "response3", Confidence: 0.7, FinishReason: "stop", TokensUsed: 50, ResponseTime: 500},
		}
		selected, scores, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.Equal(t, "2", selected.ID)
		assert.Len(t, scores, 3)
	})
}

func TestMajorityVoteStrategy_Vote(t *testing.T) {
	strategy := &MajorityVoteStrategy{}

	t.Run("no responses", func(t *testing.T) {
		_, _, err := strategy.Vote([]*models.LLMResponse{}, nil)
		assert.Error(t, err)
	})

	t.Run("single response", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "unique response", Confidence: 0.9},
		}
		selected, _, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.NotNil(t, selected)
	})

	t.Run("similar responses form majority", func(t *testing.T) {
		// Responses with same first 100 chars should group
		responses := []*models.LLMResponse{
			{ID: "1", Content: "The capital of France is Paris.", Confidence: 0.8},
			{ID: "2", Content: "The capital of France is Paris.", Confidence: 0.9},
			{ID: "3", Content: "Different response entirely.", Confidence: 0.7},
		}
		selected, _, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.NotNil(t, selected)
	})
}

func TestQualityWeightedStrategy_Vote(t *testing.T) {
	strategy := &QualityWeightedStrategy{}

	t.Run("no responses", func(t *testing.T) {
		_, _, err := strategy.Vote([]*models.LLMResponse{}, nil)
		assert.Error(t, err)
	})

	t.Run("selects highest quality", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "Short.", Confidence: 0.9, FinishReason: "stop", TokensUsed: 10, ResponseTime: 100},
			{ID: "2", Content: "A medium length response with good content.", Confidence: 0.8, FinishReason: "stop", TokensUsed: 30, ResponseTime: 200},
			{ID: "3", Content: "A really long response that goes on and on...", Confidence: 0.7, FinishReason: "length", TokensUsed: 100, ResponseTime: 5000},
		}
		selected, scores, err := strategy.Vote(responses, &models.LLMRequest{})

		require.NoError(t, err)
		assert.NotNil(t, selected)
		assert.Len(t, scores, 3)
	})
}

func TestEnsembleService_ProcessEnsemble(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	ctx := context.Background()

	t.Run("without ensemble config", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "response1", Confidence: 0.7},
			{ID: "2", Content: "response2", Confidence: 0.9},
		}

		selected, err := service.ProcessEnsemble(ctx, &models.LLMRequest{}, responses)

		require.NoError(t, err)
		assert.Equal(t, "2", selected.ID)
	})

	t.Run("with confidence weighted strategy", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "response1", Confidence: 0.7},
			{ID: "2", Content: "response2", Confidence: 0.9},
		}

		selected, err := service.ProcessEnsemble(ctx, &models.LLMRequest{
			EnsembleConfig: &models.EnsembleConfig{
				Strategy:            "confidence_weighted",
				ConfidenceThreshold: 0.5,
				FallbackToBest:      true,
			},
		}, responses)

		require.NoError(t, err)
		assert.NotNil(t, selected)
	})

	t.Run("with majority vote strategy", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "response1", Confidence: 0.9},
		}

		selected, err := service.ProcessEnsemble(ctx, &models.LLMRequest{
			EnsembleConfig: &models.EnsembleConfig{
				Strategy: "majority_vote",
			},
		}, responses)

		require.NoError(t, err)
		assert.NotNil(t, selected)
	})

	t.Run("empty responses", func(t *testing.T) {
		selected, err := service.ProcessEnsemble(ctx, &models.LLMRequest{}, []*models.LLMResponse{})

		require.NoError(t, err)
		assert.Nil(t, selected)
	})
}

func TestEnsembleService_selectBestResponse(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)

	t.Run("empty responses", func(t *testing.T) {
		result := service.selectBestResponse([]*models.LLMResponse{})
		assert.Nil(t, result)
	})

	t.Run("selects highest confidence", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Confidence: 0.5},
			{ID: "2", Confidence: 0.9},
			{ID: "3", Confidence: 0.7},
		}
		result := service.selectBestResponse(responses)

		require.NotNil(t, result)
		assert.Equal(t, "2", result.ID)
		assert.True(t, result.Selected)
	})
}

func TestEnsembleService_confidenceWeightedVoting(t *testing.T) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)

	t.Run("above threshold", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Confidence: 0.9},
			{ID: "2", Confidence: 0.7},
		}
		config := &models.EnsembleConfig{
			ConfidenceThreshold: 0.8,
			FallbackToBest:      true,
		}

		result := service.confidenceWeightedVoting(responses, config)

		require.NotNil(t, result)
		assert.Equal(t, "1", result.ID)
	})

	t.Run("below threshold with fallback", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Confidence: 0.7},
			{ID: "2", Confidence: 0.5},
		}
		config := &models.EnsembleConfig{
			ConfidenceThreshold: 0.9,
			FallbackToBest:      true,
		}

		result := service.confidenceWeightedVoting(responses, config)

		require.NotNil(t, result)
		assert.Equal(t, "1", result.ID)
	})

	t.Run("below threshold without fallback", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Confidence: 0.7},
		}
		config := &models.EnsembleConfig{
			ConfidenceThreshold: 0.9,
			FallbackToBest:      false,
		}

		result := service.confidenceWeightedVoting(responses, config)
		assert.Nil(t, result)
	})
}

func TestEnsembleService_majorityVoting(t *testing.T) {
	service := NewEnsembleService("majority_vote", 30*time.Second)

	t.Run("returns first response", func(t *testing.T) {
		responses := []*models.LLMResponse{
			{ID: "1", Content: "response1"},
			{ID: "2", Content: "response2"},
		}

		result := service.majorityVoting(responses, nil)

		require.NotNil(t, result)
		assert.Equal(t, "1", result.ID)
		assert.True(t, result.Selected)
	})

	t.Run("empty responses", func(t *testing.T) {
		result := service.majorityVoting([]*models.LLMResponse{}, nil)
		assert.Nil(t, result)
	})
}

func TestEnsembleResult(t *testing.T) {
	result := &EnsembleResult{
		Responses: []*models.LLMResponse{
			{ID: "1"},
			{ID: "2"},
		},
		Selected:     &models.LLMResponse{ID: "1"},
		VotingMethod: "confidence_weighted",
		Scores:       map[string]float64{"1": 0.9, "2": 0.7},
		Metadata:     map[string]any{"count": 2},
	}

	assert.Len(t, result.Responses, 2)
	assert.Equal(t, "1", result.Selected.ID)
	assert.Equal(t, "confidence_weighted", result.VotingMethod)
	assert.Len(t, result.Scores, 2)
}

func BenchmarkEnsembleService_RunEnsemble(b *testing.B) {
	service := NewEnsembleService("confidence_weighted", 30*time.Second)
	service.RegisterProvider("p1", newMockProvider("p1", "response1", 0.9))
	service.RegisterProvider("p2", newMockProvider("p2", "response2", 0.8))

	ctx := context.Background()
	req := &models.LLMRequest{Prompt: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.RunEnsemble(ctx, req)
	}
}

func BenchmarkConfidenceWeightedStrategy_Vote(b *testing.B) {
	strategy := &ConfidenceWeightedStrategy{}
	responses := []*models.LLMResponse{
		{ID: "1", Content: "response1", Confidence: 0.9, FinishReason: "stop", TokensUsed: 50, ResponseTime: 500},
		{ID: "2", Content: "response2", Confidence: 0.8, FinishReason: "stop", TokensUsed: 60, ResponseTime: 600},
		{ID: "3", Content: "response3", Confidence: 0.7, FinishReason: "stop", TokensUsed: 70, ResponseTime: 700},
	}
	req := &models.LLMRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = strategy.Vote(responses, req)
	}
}
