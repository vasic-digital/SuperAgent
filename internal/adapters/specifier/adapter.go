package specifier

import (
	"context"

	helixspec "digital.vasic.helixspecifier/pkg/types"
)

// SpecAdapter wraps the HelixSpecifier engine for use by HelixAgent.
type SpecAdapter struct {
	engine helixspec.SpecEngine
}

// NewSpecAdapter creates a new spec adapter wrapping a SpecEngine.
func NewSpecAdapter(engine helixspec.SpecEngine) *SpecAdapter {
	if engine == nil {
		return nil
	}
	return &SpecAdapter{engine: engine}
}

// ClassifyEffort classifies effort level for a request.
func (a *SpecAdapter) ClassifyEffort(
	ctx context.Context,
	request string,
) (*helixspec.EffortClassification, error) {
	return a.engine.ClassifyEffort(ctx, request)
}

// ExecuteFlow runs the full spec-driven development flow.
func (a *SpecAdapter) ExecuteFlow(
	ctx context.Context,
	request string,
	classification *helixspec.EffortClassification,
) (*helixspec.FlowResult, error) {
	return a.engine.ExecuteFlow(ctx, request, classification)
}

// ResumeFlow resumes a previously cached flow.
func (a *SpecAdapter) ResumeFlow(
	ctx context.Context,
	flowID string,
	request string,
) (*helixspec.FlowResult, error) {
	return a.engine.ResumeFlow(ctx, flowID, request)
}

// GetFlowStatus returns flow status.
func (a *SpecAdapter) GetFlowStatus(
	flowID string,
) (*helixspec.FlowResult, error) {
	return a.engine.GetFlowStatus(flowID)
}

// Health returns engine health.
func (a *SpecAdapter) Health(ctx context.Context) error {
	return a.engine.Health(ctx)
}

// Name returns the engine name.
func (a *SpecAdapter) Name() string {
	return a.engine.Name()
}

// Version returns the engine version.
func (a *SpecAdapter) Version() string {
	return a.engine.Version()
}

// IsReady returns true if the adapter is initialized and healthy.
func (a *SpecAdapter) IsReady() bool {
	return a.engine != nil
}
