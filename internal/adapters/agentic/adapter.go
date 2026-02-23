// Package agentic provides an adapter bridging HelixAgent to the Agentic module.
package agentic

import (
	"context"
	"time"

	agenticmod "digital.vasic.agentic/agentic"
	"github.com/sirupsen/logrus"
)

// Adapter bridges HelixAgent to the Agentic module.
type Adapter struct {
	logger *logrus.Logger
}

// New creates a new AgenticAdapter.
func New(logger *logrus.Logger) *Adapter {
	if logger == nil {
		logger = logrus.New()
	}
	return &Adapter{logger: logger}
}

// NewWorkflow creates a new workflow with the given name and description.
func (a *Adapter) NewWorkflow(name, description string, config *agenticmod.WorkflowConfig) *agenticmod.Workflow {
	return agenticmod.NewWorkflow(name, description, config, a.logger)
}

// ExecuteWorkflow creates and runs a minimal workflow with a single no-op node.
// This is a convenience method for simple cases. For complex workflows, use NewWorkflow directly.
func (a *Adapter) ExecuteWorkflow(ctx context.Context, name string, params map[string]any) (*agenticmod.WorkflowState, error) {
	cfg := &agenticmod.WorkflowConfig{
		MaxIterations: 10,
		MaxRetries:    3,
		Timeout:       30 * time.Second,
	}
	wf := a.NewWorkflow(name, "auto-generated workflow", cfg)

	// Add a single pass-through node
	node := &agenticmod.Node{
		ID:   "start",
		Name: "start",
		Type: agenticmod.NodeTypeAgent,
		Handler: func(ctx context.Context, state *agenticmod.WorkflowState, input *agenticmod.NodeInput) (*agenticmod.NodeOutput, error) {
			return &agenticmod.NodeOutput{
				Result:    params,
				ShouldEnd: true,
			}, nil
		},
	}
	if err := wf.AddNode(node); err != nil {
		return nil, err
	}
	if err := wf.SetEntryPoint("start"); err != nil {
		return nil, err
	}
	if err := wf.AddEndNode("start"); err != nil {
		return nil, err
	}

	return wf.Execute(ctx, &agenticmod.NodeInput{Context: params})
}
