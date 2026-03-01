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

	// Create message for architects
	msg := NewMessage("system", MessageTypeSystem,
		fmt.Sprintf("Create architectural design for: %s", req.Topic))

	// Have architects propose designs
	for _, architect := range architects {
		// TODO: Call actual agent.Process
		response := &AgentResponse{
			AgentID:    architect.ID,
			AgentRole:  architect.Role,
			Provider:   architect.Provider,
			Model:      architect.Model,
			Content:    fmt.Sprintf("Architectural design by %s", architect.Name),
			Confidence: 0.9,
			Score:      architect.Score,
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

	// Get design artifact if available
	var design string
	if artifact, ok := context.Artifacts["design-"+req.ID]; ok {
		design = artifact.Content
	}

	// Create message for generators
	msgContent := fmt.Sprintf("Generate code for: %s", req.Topic)
	if design != "" {
		msgContent += fmt.Sprintf("\n\nDesign:\n%s", design)
	}

	msg := NewMessage("system", MessageTypeSystem, msgContent)

	// Have generators produce code
	for _, generator := range generators {
		// TODO: Call actual agent.Process
		response := &AgentResponse{
			AgentID:    generator.ID,
			AgentRole:  generator.Role,
			Provider:   generator.Provider,
			Model:      generator.Model,
			Content:    fmt.Sprintf("Generated code by %s", generator.Name),
			Confidence: 0.85,
			Score:      generator.Score,
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

	// Get code artifact
	codeArtifact, hasCode := context.Artifacts["code-"+req.ID]
	if !hasCode {
		return phase, fmt.Errorf("no code available for debate")
	}

	// Critic agents critique the code
	critics := o.pool.GetByRole(RoleCritic)
	for _, critic := range critics {
		// TODO: Call actual agent.Process
		response := &AgentResponse{
			AgentID:    critic.ID,
			AgentRole:  critic.Role,
			Provider:   critic.Provider,
			Model:      critic.Model,
			Content:    fmt.Sprintf("Critique by %s: Review code for issues", critic.Name),
			Confidence: 0.8,
			Score:      critic.Score,
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Generator agents defend or revise
	generators := o.pool.GetByRole(RoleGenerator)
	for _, generator := range generators {
		response := &AgentResponse{
			AgentID:    generator.ID,
			AgentRole:  generator.Role,
			Provider:   generator.Provider,
			Model:      generator.Model,
			Content:    fmt.Sprintf("Defense by %s: Addressing critiques", generator.Name),
			Confidence: 0.82,
			Score:      generator.Score,
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
		response := &AgentResponse{
			AgentID:    tester.ID,
			AgentRole:  tester.Role,
			Provider:   tester.Provider,
			Model:      tester.Model,
			Content:    fmt.Sprintf("Tests by %s", tester.Name),
			Confidence: 0.9,
			Score:      tester.Score,
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Validator agents verify correctness
	validators := o.pool.GetByRole(RoleValidator)
	for _, validator := range validators {
		response := &AgentResponse{
			AgentID:    validator.ID,
			AgentRole:  validator.Role,
			Provider:   validator.Provider,
			Model:      validator.Model,
			Content:    fmt.Sprintf("Validation by %s", validator.Name),
			Confidence: 0.95,
			Score:      validator.Score,
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
		response := &AgentResponse{
			AgentID:    refactorer.ID,
			AgentRole:  refactorer.Role,
			Provider:   refactorer.Provider,
			Model:      refactorer.Model,
			Content:    fmt.Sprintf("Refactoring by %s", refactorer.Name),
			Confidence: 0.88,
			Score:      refactorer.Score,
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Performance agents optimize
	performers := o.pool.GetByRole(RolePerformance)
	for _, performer := range performers {
		response := &AgentResponse{
			AgentID:    performer.ID,
			AgentRole:  performer.Role,
			Provider:   performer.Provider,
			Model:      performer.Model,
			Content:    fmt.Sprintf("Optimization by %s", performer.Name),
			Confidence: 0.9,
			Score:      performer.Score,
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
		response := &AgentResponse{
			AgentID:    sec.ID,
			AgentRole:  sec.Role,
			Provider:   sec.Provider,
			Model:      sec.Model,
			Content:    fmt.Sprintf("Security review by %s", sec.Name),
			Confidence: 0.85,
			Score:      sec.Score,
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	// Final validation
	validators := o.pool.GetByRole(RoleValidator)
	for _, validator := range validators {
		response := &AgentResponse{
			AgentID:    validator.ID,
			AgentRole:  validator.Role,
			Provider:   validator.Provider,
			Model:      validator.Model,
			Content:    fmt.Sprintf("Final validation by %s", validator.Name),
			Confidence: 0.95,
			Score:      validator.Score,
		}

		phase.Responses = append(phase.Responses, *response)
		context.AddResponse(response)
	}

	o.logger.WithField("responses", len(phase.Responses)).Info("Integration phase completed")
	return phase, nil
}
