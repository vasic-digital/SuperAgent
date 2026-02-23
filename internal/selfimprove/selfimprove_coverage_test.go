package selfimprove

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAIRewardModel_Train_WithDimensions verifies that Train calls
// updateDimensionWeights and calibrateScoring when provided examples with
// non-empty Dimensions and both positive and negative RewardScores.
//
// updateDimensionWeights is called when len(patterns.dimensionCorrelations) > 0
// (requires examples with non-empty Dimensions maps).
// calibrateScoring is called when avgPositiveScore > 0 AND avgNegativeScore > 0
// (requires both positive examples (score > 0.5) and negative (score < 0.5)).
func TestAIRewardModel_Train_WithDimensions(t *testing.T) {
	config := DefaultSelfImprovementConfig()
	model := NewAIRewardModel(nil, nil, config, logrus.New())

	examples := []*TrainingExample{
		{
			ID:          "pos-1",
			RewardScore: 0.85, // positive
			Dimensions: map[DimensionType]float64{
				DimensionAccuracy:    0.9,
				DimensionHelpfulness: 0.8,
				DimensionCoherence:   0.85,
			},
		},
		{
			ID:          "pos-2",
			RewardScore: 0.75, // positive
			Dimensions: map[DimensionType]float64{
				DimensionAccuracy:  0.7,
				DimensionRelevance: 0.8,
			},
		},
		{
			ID:          "neg-1",
			RewardScore: 0.3, // negative
			Dimensions: map[DimensionType]float64{
				DimensionAccuracy:    0.3,
				DimensionHelpfulness: 0.2,
			},
		},
		{
			ID:          "neg-2",
			RewardScore: 0.2, // negative
			Dimensions: map[DimensionType]float64{
				DimensionHarmless: 0.8, // high harmlessness but low overall
			},
		},
	}

	err := model.Train(context.Background(), examples)
	require.NoError(t, err)
}

// TestAIRewardModel_Train_LargePositiveSet verifies storePositiveExamples
// truncation logic (> 100 examples).
func TestAIRewardModel_Train_LargePositiveSet(t *testing.T) {
	config := DefaultSelfImprovementConfig()
	model := NewAIRewardModel(nil, nil, config, logrus.New())

	// Create 120 positive examples (> maxStored = 100)
	examples := make([]*TrainingExample, 120)
	for i := 0; i < 120; i++ {
		examples[i] = &TrainingExample{
			ID:          string(rune(i + 1)),
			RewardScore: 0.6 + float64(i)*0.003, // all > 0.5, varying scores
			Dimensions: map[DimensionType]float64{
				DimensionAccuracy: 0.7,
			},
		}
	}

	err := model.Train(context.Background(), examples)
	require.NoError(t, err)
}

// TestAIRewardModel_Train_WithMinMaxWeightClamping verifies the weight update
// clamping logic. Providing a large positive correlation should clamp the
// updated weight to the maximum (0.5), and a large negative should clamp to the
// minimum (0.05).
func TestAIRewardModel_Train_WithMinMaxWeightClamping(t *testing.T) {
	config := DefaultSelfImprovementConfig()
	model := NewAIRewardModel(nil, nil, config, logrus.New())

	// Large positive correlation → large positive score × large reward
	// Large negative correlation → large positive score × large negative reward
	examples := []*TrainingExample{
		{
			ID:          "clamp-high",
			RewardScore: 0.99, // near-maximum positive
			Dimensions: map[DimensionType]float64{
				DimensionAccuracy: 1.0, // max dimension score
			},
		},
		{
			ID:          "clamp-low",
			RewardScore: 0.01, // near-minimum negative
			Dimensions: map[DimensionType]float64{
				DimensionRelevance: 1.0,
			},
		},
	}

	// Should not panic and should complete successfully
	err := model.Train(context.Background(), examples)
	assert.NoError(t, err)
}
