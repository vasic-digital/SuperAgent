// Package llmops provides an adapter bridging HelixAgent to the LLMOps module.
package llmops

import (
	"context"

	llmopsmod "digital.vasic.llmops/llmops"
	"github.com/sirupsen/logrus"
)

// Adapter bridges HelixAgent to the LLMOps module.
type Adapter struct {
	logger *logrus.Logger
}

// New creates a new LLMOps adapter.
func New(logger *logrus.Logger) *Adapter {
	if logger == nil {
		logger = logrus.New()
	}
	return &Adapter{logger: logger}
}

// NewEvaluator creates a new in-memory continuous evaluator.
func (a *Adapter) NewEvaluator() *llmopsmod.InMemoryContinuousEvaluator {
	return llmopsmod.NewInMemoryContinuousEvaluator(nil, nil, nil, a.logger)
}

// NewExperimentManager creates a new in-memory experiment manager.
func (a *Adapter) NewExperimentManager() *llmopsmod.InMemoryExperimentManager {
	return llmopsmod.NewInMemoryExperimentManager(a.logger)
}

// CreateDataset creates a new dataset in the evaluator.
func (a *Adapter) CreateDataset(ctx context.Context, evaluator *llmopsmod.InMemoryContinuousEvaluator, name string, datasetType llmopsmod.DatasetType) (*llmopsmod.Dataset, error) {
	ds := &llmopsmod.Dataset{Name: name, Type: datasetType}
	return ds, evaluator.CreateDataset(ctx, ds)
}
