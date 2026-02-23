// Package planning provides a bridge adapter between HelixAgent and the
// extracted digital.vasic.planning module.
package planning

import (
	"context"

	"github.com/sirupsen/logrus"

	extplanning "digital.vasic.planning/planning"
)

// Adapter bridges HelixAgent to the digital.vasic.planning module.
type Adapter struct {
	logger *logrus.Logger
}

// New creates a new planning Adapter with the provided logger.
func New(logger *logrus.Logger) *Adapter {
	if logger == nil {
		logger = logrus.New()
	}
	return &Adapter{logger: logger}
}

// NewHiPlan creates a new HiPlan instance using the extracted planning module.
func (a *Adapter) NewHiPlan(
	config extplanning.HiPlanConfig,
	generator extplanning.MilestoneGenerator,
	executor extplanning.StepExecutor,
) *extplanning.HiPlan {
	return extplanning.NewHiPlan(config, generator, executor, a.logger)
}

// NewMCTS creates a new MCTS instance using the extracted planning module.
func (a *Adapter) NewMCTS(
	config extplanning.MCTSConfig,
	actionGen extplanning.MCTSActionGenerator,
	rewardFunc extplanning.MCTSRewardFunction,
	rolloutPolicy extplanning.MCTSRolloutPolicy,
) *extplanning.MCTS {
	return extplanning.NewMCTS(config, actionGen, rewardFunc, rolloutPolicy, a.logger)
}

// NewTreeOfThoughts creates a new TreeOfThoughts instance using the extracted planning module.
func (a *Adapter) NewTreeOfThoughts(
	config extplanning.TreeOfThoughtsConfig,
	generator extplanning.ThoughtGenerator,
	evaluator extplanning.ThoughtEvaluator,
) *extplanning.TreeOfThoughts {
	return extplanning.NewTreeOfThoughts(config, generator, evaluator, a.logger)
}

// NewLLMMilestoneGenerator creates an LLMMilestoneGenerator with the provided generate function.
func (a *Adapter) NewLLMMilestoneGenerator(
	generateFunc func(ctx context.Context, prompt string) (string, error),
) *extplanning.LLMMilestoneGenerator {
	return extplanning.NewLLMMilestoneGenerator(generateFunc, a.logger)
}

// NewLLMThoughtGenerator creates an LLMThoughtGenerator with the provided generate function.
func (a *Adapter) NewLLMThoughtGenerator(
	generateFunc func(ctx context.Context, prompt string) (string, error),
	temperature float64,
) *extplanning.LLMThoughtGenerator {
	return extplanning.NewLLMThoughtGenerator(generateFunc, temperature, a.logger)
}

// NewLLMThoughtEvaluator creates an LLMThoughtEvaluator with the provided evaluate function.
func (a *Adapter) NewLLMThoughtEvaluator(
	evaluateFunc func(ctx context.Context, prompt string) (string, error),
) *extplanning.LLMThoughtEvaluator {
	return extplanning.NewLLMThoughtEvaluator(evaluateFunc, a.logger)
}

// DefaultHiPlanConfig returns the default HiPlan configuration.
func (a *Adapter) DefaultHiPlanConfig() extplanning.HiPlanConfig {
	return extplanning.DefaultHiPlanConfig()
}

// DefaultMCTSConfig returns the default MCTS configuration.
func (a *Adapter) DefaultMCTSConfig() extplanning.MCTSConfig {
	return extplanning.DefaultMCTSConfig()
}

// DefaultTreeOfThoughtsConfig returns the default Tree of Thoughts configuration.
func (a *Adapter) DefaultTreeOfThoughtsConfig() extplanning.TreeOfThoughtsConfig {
	return extplanning.DefaultTreeOfThoughtsConfig()
}
