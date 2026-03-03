package comprehensive

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// PhaseOrchestrator manages the execution of debate phases
type PhaseOrchestrator struct {
	pool   *AgentPool
	logger *logrus.Logger
}

// NewPhaseOrchestrator creates a new phase orchestrator
func NewPhaseOrchestrator(pool *AgentPool, logger *logrus.Logger) *PhaseOrchestrator {
	if logger == nil {
		logger = logrus.New()
	}

	return &PhaseOrchestrator{
		pool:   pool,
		logger: logger,
	}
}

// PlanningPhase executes the planning phase
func (o *PhaseOrchestrator) PlanningPhase(ctx context.Context, req *DebateRequest, context *Context) (*PhaseResult, error) {
	o.logger.Info("Phase 0: Planning")

	phase := &PhaseResult{
		Phase:     "planning",
		Responses: make([]AgentResponse, 0),
	}

	// Get architect agents
	architects := o.pool.GetByRole(RoleArchitect)
	if len(architects) == 0 {
		return phase, fmt.Errorf("no architect agents available")
	}

	// Have architects propose designs
	for _, architect := range architects {
		specialized := NewArchitectAgent(architect, o.pool)
		msg := &Message{
			Type:    MessageTypeProposal,
			Content: req.Topic,
		}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", architect.Name).Warn("Architect processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Store design artifact
	if len(phase.Responses) > 0 {
		artifact := &Artifact{
			ID:      "design-" + req.ID,
			Type:    ArtifactTypeDesign,
			Name:    "architecture.md",
			Content: phase.Responses[0].Content,
			AgentID: phase.Responses[0].AgentID,
			Version: 1,
		}
		context.AddArtifact(artifact)
	}

	o.logger.WithField("responses", len(phase.Responses)).Info("Planning phase completed")
	return phase, nil
}

// GenerationPhase executes the generation phase
func (o *PhaseOrchestrator) GenerationPhase(ctx context.Context, req *DebateRequest, context *Context) (*PhaseResult, error) {
	o.logger.Info("Phase 1: Generation")

	phase := &PhaseResult{
		Phase:     "generation",
		Responses: make([]AgentResponse, 0),
	}

	// Get generator agents
	generators := o.pool.GetByRole(RoleGenerator)
	if len(generators) == 0 {
		return phase, fmt.Errorf("no generator agents available")
	}

	// Have generators produce code
	for _, generator := range generators {
		specialized := NewGeneratorAgent(generator, o.pool)
		msg := &Message{
			Type:    MessageTypeProposal,
			Content: req.Topic,
		}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", generator.Name).Warn("Generator processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Store code artifact
	if len(phase.Responses) > 0 {
		artifact := &Artifact{
			ID:      "code-" + req.ID,
			Type:    ArtifactTypeCode,
			Name:    "implementation.go",
			Content: phase.Responses[0].Content,
			AgentID: phase.Responses[0].AgentID,
			Version: 1,
		}
		context.AddArtifact(artifact)
	}

	o.logger.WithField("responses", len(phase.Responses)).Info("Generation phase completed")
	return phase, nil
}

// DebatePhase executes the debate phase
func (o *PhaseOrchestrator) DebatePhase(ctx context.Context, req *DebateRequest, context *Context, round int) (*PhaseResult, error) {
	o.logger.WithField("round", round).Info("Phase 2-3: Debate")

	phase := &PhaseResult{
		Phase:     fmt.Sprintf("debate_round_%d", round),
		Round:     round,
		Responses: make([]AgentResponse, 0),
	}

	// Check if code artifact exists
	if _, hasCode := context.Artifacts["code-"+req.ID]; !hasCode {
		return phase, fmt.Errorf("no code available for debate")
	}

	// Critic agents critique the code
	critics := o.pool.GetByRole(RoleCritic)
	for _, critic := range critics {
		specialized := NewCriticAgent(critic, o.pool)
		msg := &Message{
			Type:    MessageTypeCritique,
			Content: req.Topic,
		}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", critic.Name).Warn("Critic processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Generator agents defend or revise
	generators := o.pool.GetByRole(RoleGenerator)
	for _, generator := range generators {
		specialized := NewGeneratorAgent(generator, o.pool)
		msg := &Message{
			Type:    MessageTypeDefense,
			Content: req.Topic,
		}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", generator.Name).Warn("Generator defense failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	o.logger.WithFields(logrus.Fields{
		"round":     round,
		"responses": len(phase.Responses),
	}).Info("Debate round completed")

	return phase, nil
}

// ValidationPhase executes the validation phase
func (o *PhaseOrchestrator) ValidationPhase(ctx context.Context, req *DebateRequest, context *Context) (*PhaseResult, error) {
	o.logger.Info("Phase 4: Validation")

	phase := &PhaseResult{
		Phase:     "validation",
		Responses: make([]AgentResponse, 0),
	}

	// Tester agents create tests
	testers := o.pool.GetByRole(RoleTester)
	for _, tester := range testers {
		specialized := NewTesterAgent(tester, o.pool)
		msg := &Message{Type: MessageTypeProposal, Content: req.Topic}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", tester.Name).Warn("Tester processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Validator agents verify correctness
	validators := o.pool.GetByRole(RoleValidator)
	for _, validator := range validators {
		specialized := NewValidatorAgent(validator, o.pool)
		msg := &Message{Type: MessageTypeProposal, Content: req.Topic}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", validator.Name).Warn("Validator processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Store test artifact
	if len(phase.Responses) > 0 {
		artifact := &Artifact{
			ID:      "tests-" + req.ID,
			Type:    ArtifactTypeTest,
			Name:    "implementation_test.go",
			Content: "package main\n\nimport \"testing\"",
			AgentID: phase.Responses[0].AgentID,
			Version: 1,
		}
		context.AddArtifact(artifact)
	}

	o.logger.WithField("responses", len(phase.Responses)).Info("Validation phase completed")
	return phase, nil
}

// RefactoringPhase executes the refactoring phase
func (o *PhaseOrchestrator) RefactoringPhase(ctx context.Context, req *DebateRequest, context *Context) (*PhaseResult, error) {
	o.logger.Info("Phase 5: Refactoring")

	phase := &PhaseResult{
		Phase:     "refactoring",
		Responses: make([]AgentResponse, 0),
	}

	// Refactoring agents improve code
	refactorers := o.pool.GetByRole(RoleRefactoring)
	for _, refactorer := range refactorers {
		specialized := NewRefactoringAgent(refactorer, o.pool)
		msg := &Message{Type: MessageTypeProposal, Content: req.Topic}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", refactorer.Name).Warn("Refactoring agent processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Performance agents optimize
	performers := o.pool.GetByRole(RolePerformance)
	for _, performer := range performers {
		specialized := NewPerformanceAgent(performer, o.pool)
		msg := &Message{Type: MessageTypeProposal, Content: req.Topic}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", performer.Name).Warn("Performance agent processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	o.logger.WithField("responses", len(phase.Responses)).Info("Refactoring phase completed")
	return phase, nil
}

// IntegrationPhase executes the integration phase
func (o *PhaseOrchestrator) IntegrationPhase(ctx context.Context, req *DebateRequest, context *Context) (*PhaseResult, error) {
	o.logger.Info("Phase 6: Integration")

	phase := &PhaseResult{
		Phase:     "integration",
		Responses: make([]AgentResponse, 0),
	}

	// Security review
	securityAgents := o.pool.GetByRole(RoleSecurity)
	for _, sec := range securityAgents {
		specialized := NewSecurityAgent(sec, o.pool)
		msg := &Message{Type: MessageTypeCritique, Content: req.Topic}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", sec.Name).Warn("Security agent processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Final validation
	validators := o.pool.GetByRole(RoleValidator)
	for _, validator := range validators {
		specialized := NewValidatorAgent(validator, o.pool)
		msg := &Message{Type: MessageTypeProposal, Content: req.Topic}
		response, err := specialized.Process(ctx, msg, context)
		if err != nil {
			o.logger.WithError(err).WithField("agent", validator.Name).Warn("Validator processing failed")
			continue
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	o.logger.WithField("responses", len(phase.Responses)).Info("Integration phase completed")
	return phase, nil
}
