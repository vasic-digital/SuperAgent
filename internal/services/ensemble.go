package services

import (
	"context"
	"sort"

	"github.com/superagent/superagent/internal/models"
)

type EnsembleService struct {
	// TODO: Add provider registry
}

func NewEnsembleService() *EnsembleService {
	return &EnsembleService{}
}

func (e *EnsembleService) ProcessEnsemble(ctx context.Context, req *models.LLMRequest, responses []*models.LLMResponse) (*models.LLMResponse, error) {
	if req.EnsembleConfig == nil {
		// No ensemble, return best response
		return e.selectBestResponse(responses), nil
	}

	config := req.EnsembleConfig

	switch config.Strategy {
	case "confidence_weighted":
		return e.confidenceWeightedVoting(responses, config), nil
	case "majority_vote":
		return e.majorityVoting(responses, config), nil
	default:
		return e.selectBestResponse(responses), nil
	}
}

func (e *EnsembleService) confidenceWeightedVoting(responses []*models.LLMResponse, config *models.EnsembleConfig) *models.LLMResponse {
	if len(responses) == 0 {
		return nil
	}

	// Sort by confidence
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].Confidence > responses[j].Confidence
	})

	// Select highest confidence if above threshold
	if responses[0].Confidence >= config.ConfidenceThreshold {
		responses[0].Selected = true
		responses[0].SelectionScore = responses[0].Confidence
		return responses[0]
	}

	// Fallback to best
	if config.FallbackToBest {
		responses[0].Selected = true
		responses[0].SelectionScore = responses[0].Confidence
		return responses[0]
	}

	return nil
}

func (e *EnsembleService) majorityVoting(responses []*models.LLMResponse, config *models.EnsembleConfig) *models.LLMResponse {
	// Simplified majority voting
	if len(responses) == 0 {
		return nil
	}

	// For now, return first
	responses[0].Selected = true
	responses[0].SelectionScore = 1.0
	return responses[0]
}

func (e *EnsembleService) selectBestResponse(responses []*models.LLMResponse) *models.LLMResponse {
	if len(responses) == 0 {
		return nil
	}

	best := responses[0]
	for _, resp := range responses[1:] {
		if resp.Confidence > best.Confidence {
			best = resp
		}
	}

	best.Selected = true
	best.SelectionScore = best.Confidence
	return best
}
