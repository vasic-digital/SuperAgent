// Package selfimprove provides an adapter bridging HelixAgent to the SelfImprove module.
package selfimprove

import (
	"context"

	selfimprovemod "digital.vasic.selfimprove/selfimprove"
	"github.com/sirupsen/logrus"
)

// Adapter bridges HelixAgent to the SelfImprove module.
type Adapter struct {
	logger *logrus.Logger
}

// New creates a new SelfImprove adapter.
func New(logger *logrus.Logger) *Adapter {
	if logger == nil {
		logger = logrus.New()
	}
	return &Adapter{logger: logger}
}

// NewRewardModel creates a new AI reward model.
func (a *Adapter) NewRewardModel(config *selfimprovemod.SelfImprovementConfig) *selfimprovemod.AIRewardModel {
	if config == nil {
		config = selfimprovemod.DefaultSelfImprovementConfig()
	}
	return selfimprovemod.NewAIRewardModel(nil, nil, config, a.logger)
}

// Train trains the reward model with provided examples.
func (a *Adapter) Train(ctx context.Context, model *selfimprovemod.AIRewardModel, examples []*selfimprovemod.TrainingExample) error {
	return model.Train(ctx, examples)
}
