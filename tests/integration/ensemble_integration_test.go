// Package integration provides ensemble voting integration tests.
// These tests validate ensemble strategies (confidence-weighted, majority vote,
// quality-weighted) using real strategy implementations and mock providers.
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/services"
)

// =============================================================================
// Test: Ensemble — Majority Vote with Basic Consensus
// =============================================================================

func TestIntegration_Ensemble_MajorityVote_BasicConsensus(t *testing.T) {
	ensemble := services.NewEnsembleService("majority_vote", 30*time.Second)

	// Create 3 providers that return identical content
	consensusContent := "The answer is 42. This is the definitive response."
	for i := 0; i < 3; i++ {
		name := "majority-provider-" + string(rune('A'+i))
		p := &MockBaseLLMProvider{
			name: name,
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:           "resp-" + name,
					Content:      consensusContent,
					Confidence:   0.90,
					ProviderName: name,
					FinishReason: "stop",
					TokensUsed:   20,
					ResponseTime: 200,
				}, nil
			},
		}
		ensemble.RegisterProvider(name, p)
	}

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "majority-test",
		Prompt: "What is the answer?",
		Messages: []models.Message{
			{Role: "user", Content: "What is the answer?"},
		},
	}

	result, err := ensemble.RunEnsemble(ctx, req)
	require.NoError(t, err, "RunEnsemble should not fail with 3 responsive providers")
	require.NotNil(t, result)

	// All 3 providers returned identical content — majority should win
	assert.NotNil(t, result.Selected, "A response must be selected")
	assert.Equal(t, consensusContent, result.Selected.Content,
		"Selected response must match the consensus content")
	assert.Equal(t, "majority_vote", result.VotingMethod)
	assert.Len(t, result.Responses, 3, "All 3 provider responses should be collected")
}

// =============================================================================
// Test: Ensemble — Confidence-Weighted Vote with different confidence scores
// =============================================================================

func TestIntegration_Ensemble_WeightedVote_HighConfidence(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)

	// Provider A: low confidence
	ensemble.RegisterProvider("weighted-low", &MockBaseLLMProvider{
		name: "weighted-low",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return &models.LLMResponse{
				ID:           "resp-low",
				Content:      "Low confidence response with reasonable length for scoring purposes in this test.",
				Confidence:   0.40,
				ProviderName: "weighted-low",
				FinishReason: "stop",
				TokensUsed:   15,
				ResponseTime: 500,
			}, nil
		},
	})

	// Provider B: medium confidence
	ensemble.RegisterProvider("weighted-med", &MockBaseLLMProvider{
		name: "weighted-med",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return &models.LLMResponse{
				ID:           "resp-med",
				Content:      "Medium confidence response with decent quality content for the weighted voting strategy evaluation.",
				Confidence:   0.70,
				ProviderName: "weighted-med",
				FinishReason: "stop",
				TokensUsed:   20,
				ResponseTime: 300,
			}, nil
		},
	})

	// Provider C: high confidence
	ensemble.RegisterProvider("weighted-high", &MockBaseLLMProvider{
		name: "weighted-high",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return &models.LLMResponse{
				ID:           "resp-high",
				Content:      "High confidence response — this provider is most certain about its answer and provides clear content.",
				Confidence:   0.98,
				ProviderName: "weighted-high",
				FinishReason: "stop",
				TokensUsed:   22,
				ResponseTime: 150,
			}, nil
		},
	})

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "weighted-test",
		Prompt: "Explain something",
		Messages: []models.Message{
			{Role: "user", Content: "Explain something"},
		},
	}

	result, err := ensemble.RunEnsemble(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Selected)

	// The highest-confidence provider should win in confidence-weighted voting
	assert.Equal(t, "resp-high", result.Selected.ID,
		"Confidence-weighted strategy should select the highest confidence response")
	assert.Equal(t, "confidence_weighted", result.VotingMethod)

	// Verify scores are populated
	assert.NotEmpty(t, result.Scores, "Scores map must be populated")
	assert.Len(t, result.Responses, 3)
}

// =============================================================================
// Test: Ensemble — Quality-Weighted (Borda-like ranking by quality factors)
// =============================================================================

func TestIntegration_Ensemble_QualityWeighted_Ranking(t *testing.T) {
	ensemble := services.NewEnsembleService("quality_weighted", 30*time.Second)

	// Provider with perfect quality signals: fast, good finish, moderate length
	ensemble.RegisterProvider("quality-best", &MockBaseLLMProvider{
		name: "quality-best",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return &models.LLMResponse{
				ID:           "resp-quality-best",
				Content:      "A well-crafted response with good length, proper finish reason, and efficient token use. This tests quality scoring.",
				Confidence:   0.85,
				ProviderName: "quality-best",
				FinishReason: "stop",
				TokensUsed:   20,
				ResponseTime: 500, // Fast
			}, nil
		},
	})

	// Provider with poor quality signals: slow, truncated, verbose
	ensemble.RegisterProvider("quality-poor", &MockBaseLLMProvider{
		name: "quality-poor",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return &models.LLMResponse{
				ID:           "resp-quality-poor",
				Content:      "Short",
				Confidence:   0.50,
				ProviderName: "quality-poor",
				FinishReason: "length",
				TokensUsed:   100,
				ResponseTime: 15000, // Very slow
			}, nil
		},
	})

	// Provider with middle-of-the-road quality
	ensemble.RegisterProvider("quality-mid", &MockBaseLLMProvider{
		name: "quality-mid",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return &models.LLMResponse{
				ID:           "resp-quality-mid",
				Content:      "A moderate response that is reasonable but not outstanding in any particular dimension.",
				Confidence:   0.65,
				ProviderName: "quality-mid",
				FinishReason: "stop",
				TokensUsed:   30,
				ResponseTime: 3000,
			}, nil
		},
	})

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "quality-test",
		Prompt: "Rank by quality",
		Messages: []models.Message{
			{Role: "user", Content: "Rank by quality"},
		},
	}

	result, err := ensemble.RunEnsemble(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Selected)

	// The best-quality provider should be selected
	assert.Equal(t, "resp-quality-best", result.Selected.ID,
		"Quality-weighted strategy should select the highest quality response")
	assert.Equal(t, "quality_weighted", result.VotingMethod)

	// Verify scores exist for all responses
	assert.Len(t, result.Scores, 3, "Scores should exist for all 3 responses")

	// Verify the best has the highest score
	bestScore := result.Scores["resp-quality-best"]
	poorScore := result.Scores["resp-quality-poor"]
	assert.Greater(t, bestScore, poorScore,
		"Best quality provider should have higher score than poor quality")
}

// =============================================================================
// Test: Ensemble — No Consensus Fallback
// =============================================================================

func TestIntegration_Ensemble_NoConsensus_Fallback(t *testing.T) {
	ensemble := services.NewEnsembleService("majority_vote", 30*time.Second)

	// 3 providers with completely different content — no majority possible
	contents := []string{
		"The answer involves quantum mechanics and wave functions that collapse upon observation of the system state.",
		"Actually the solution is purely mathematical — it can be derived from first principles of set theory and logic.",
		"In my analysis, the problem is fundamentally philosophical and cannot be resolved through empirical methods alone.",
	}

	for i, content := range contents {
		c := content // capture
		name := "no-consensus-" + string(rune('A'+i))
		ensemble.RegisterProvider(name, &MockBaseLLMProvider{
			name: name,
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					ID:           "resp-nc-" + string(rune('A'+i)),
					Content:      c,
					Confidence:   0.70 + float64(i)*0.05,
					ProviderName: name,
					FinishReason: "stop",
					TokensUsed:   25,
					ResponseTime: 400,
				}, nil
			},
		})
	}

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "no-consensus-test",
		Prompt: "Explain this complex topic",
		Messages: []models.Message{
			{Role: "user", Content: "Explain this complex topic"},
		},
	}

	result, err := ensemble.RunEnsemble(ctx, req)
	require.NoError(t, err, "Ensemble should not error — it should fall back")
	require.NotNil(t, result)
	require.NotNil(t, result.Selected, "A response must still be selected via fallback")

	// When majority vote has no consensus, it falls back to confidence-weighted
	// So the selected response should be the one with best overall score
	assert.NotEmpty(t, result.Selected.Content)
	assert.Len(t, result.Responses, 3)
}

// =============================================================================
// Test: Ensemble — Single Provider Pass-Through
// =============================================================================

func TestIntegration_Ensemble_SingleProvider_PassThrough(t *testing.T) {
	// Test each strategy with a single provider
	strategies := []string{"confidence_weighted", "majority_vote", "quality_weighted"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			ensemble := services.NewEnsembleService(strategy, 30*time.Second)

			singleContent := "This is the only response from the single provider in the ensemble."
			ensemble.RegisterProvider("sole-provider", &MockBaseLLMProvider{
				name: "sole-provider",
				completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
					return &models.LLMResponse{
						ID:           "sole-resp",
						Content:      singleContent,
						Confidence:   0.88,
						ProviderName: "sole-provider",
						FinishReason: "stop",
						TokensUsed:   18,
						ResponseTime: 250,
					}, nil
				},
			})

			ctx := context.Background()
			req := &models.LLMRequest{
				ID:     "single-test-" + strategy,
				Prompt: "Single provider",
				Messages: []models.Message{
					{Role: "user", Content: "Single provider"},
				},
			}

			result, err := ensemble.RunEnsemble(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Selected)

			// With only one provider, the single response must pass through
			assert.Equal(t, singleContent, result.Selected.Content,
				"Single provider response should pass through unchanged")
			assert.Equal(t, "sole-resp", result.Selected.ID)
			assert.Len(t, result.Responses, 1)
		})
	}
}

// =============================================================================
// Test: Ensemble — Provider Failure Handling
// =============================================================================

func TestIntegration_Ensemble_PartialFailure_StillSelects(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)

	// Provider A: succeeds
	ensemble.RegisterProvider("partial-ok", &MockBaseLLMProvider{
		name: "partial-ok",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return &models.LLMResponse{
				ID:           "resp-ok",
				Content:      "Successful response from the surviving provider in the ensemble test.",
				Confidence:   0.92,
				ProviderName: "partial-ok",
				FinishReason: "stop",
				TokensUsed:   15,
				ResponseTime: 200,
			}, nil
		},
	})

	// Provider B: fails
	ensemble.RegisterProvider("partial-fail", &MockBaseLLMProvider{
		name: "partial-fail",
		completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
			return nil, assert.AnError
		},
	})

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "partial-failure-test",
		Prompt: "Test partial failure",
		Messages: []models.Message{
			{Role: "user", Content: "Test partial failure"},
		},
	}

	result, err := ensemble.RunEnsemble(ctx, req)
	require.NoError(t, err, "Ensemble should succeed if at least one provider responds")
	require.NotNil(t, result)
	require.NotNil(t, result.Selected)

	assert.Equal(t, "resp-ok", result.Selected.ID)
	assert.Len(t, result.Responses, 1, "Only 1 successful response should be collected")

	// Metadata should track failures
	if md, ok := result.Metadata["failed_providers"]; ok {
		failedCount, _ := md.(int)
		assert.Equal(t, 1, failedCount)
	}
}

// =============================================================================
// Test: Ensemble — All Providers Fail
// =============================================================================

func TestIntegration_Ensemble_AllProvidersFail(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 5*time.Second)

	for i := 0; i < 3; i++ {
		name := "all-fail-" + string(rune('A'+i))
		ensemble.RegisterProvider(name, &MockBaseLLMProvider{
			name: name,
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, assert.AnError
			},
		})
	}

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "all-fail-test",
		Prompt: "Test all fail",
		Messages: []models.Message{
			{Role: "user", Content: "Test all fail"},
		},
	}

	result, err := ensemble.RunEnsemble(ctx, req)
	assert.Error(t, err, "Ensemble should return error when all providers fail")
	assert.Nil(t, result)
}

// =============================================================================
// Test: Ensemble — ProcessEnsemble with pre-collected responses
// =============================================================================

func TestIntegration_Ensemble_ProcessEnsemble_SelectsBest(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)

	// Pre-build responses (simulating already-collected provider outputs)
	responses := []*models.LLMResponse{
		{
			ID:           "pre-resp-1",
			Content:      "First provider response with moderate quality and reasonable length for testing.",
			Confidence:   0.75,
			ProviderName: "provider-1",
			FinishReason: "stop",
			TokensUsed:   18,
			ResponseTime: 300,
		},
		{
			ID:           "pre-resp-2",
			Content:      "Second provider response — this is the best one with highest quality metrics overall.",
			Confidence:   0.95,
			ProviderName: "provider-2",
			FinishReason: "stop",
			TokensUsed:   20,
			ResponseTime: 100,
		},
		{
			ID:           "pre-resp-3",
			Content:      "Third provider response, decent but not as strong as the second in terms of confidence.",
			Confidence:   0.60,
			ProviderName: "provider-3",
			FinishReason: "stop",
			TokensUsed:   22,
			ResponseTime: 800,
		},
	}

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "process-test",
		Prompt: "Process pre-collected",
	}

	// Without ensemble config, ProcessEnsemble selects best response
	selected, err := ensemble.ProcessEnsemble(ctx, req, responses)
	require.NoError(t, err)
	require.NotNil(t, selected)

	// The best response should be selected (highest confidence as primary factor)
	assert.Equal(t, "pre-resp-2", selected.ID,
		"ProcessEnsemble should select the best response")
}

// =============================================================================
// Test: Ensemble — No Providers registered
// =============================================================================

func TestIntegration_Ensemble_NoProviders(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 10*time.Second)

	ctx := context.Background()
	req := &models.LLMRequest{
		ID:     "no-providers",
		Prompt: "Test",
		Messages: []models.Message{
			{Role: "user", Content: "Test"},
		},
	}

	result, err := ensemble.RunEnsemble(ctx, req)
	assert.Error(t, err, "RunEnsemble with no providers should return error")
	assert.Nil(t, result)
}

// =============================================================================
// Test: Ensemble — Register and Remove Provider dynamically
// =============================================================================

func TestIntegration_Ensemble_RegisterRemove_Dynamic(t *testing.T) {
	ensemble := services.NewEnsembleService("confidence_weighted", 30*time.Second)

	p1 := &MockBaseLLMProvider{name: "dynamic-1"}
	p2 := &MockBaseLLMProvider{name: "dynamic-2"}

	ensemble.RegisterProvider("dynamic-1", p1)
	ensemble.RegisterProvider("dynamic-2", p2)

	providers := ensemble.GetProviders()
	assert.Len(t, providers, 2)

	// Remove one
	ensemble.RemoveProvider("dynamic-1")

	providers = ensemble.GetProviders()
	assert.Len(t, providers, 1)
	assert.Contains(t, providers, "dynamic-2")
	assert.NotContains(t, providers, "dynamic-1")
}
