// Package comprehensive provides the complete multi-agent debate system
// as specified in the debate documentation.
//
// This implements:
// - 8 specialized agent roles (Architect, Generator, Critic, Refactoring, Tester, Validator, Security, Performance)
// - 6-phase debate workflow (Planning → Generation → Debate → Validation → Refactoring → Integration)
// - Test-driven validation
// - Adversarial Red/Blue team dynamics
// - Reflexion self-correction
// - Quality gates and convergence criteria
//
// Usage:
//
//	config := comprehensive.DefaultConfig()
//	system := comprehensive.NewSystem(config)
//	result, err := system.ConductDebate(ctx, debateReq)
package comprehensive

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Config configures the comprehensive debate system
type Config struct {
	// Agent configuration
	EnableArchitect   bool
	EnableGenerator   bool
	EnableCritic      bool
	EnableRefactoring bool
	EnableTester      bool
	EnableValidator   bool
	EnableSecurity    bool
	EnablePerformance bool

	// Phase configuration
	EnablePlanningPhase    bool
	EnableGenerationPhase  bool
	EnableDebatePhase      bool
	EnableValidationPhase  bool
	EnableRefactoringPhase bool
	EnableIntegrationPhase bool

	// Debate parameters
	MaxRounds        int
	MinConsensus     float64
	QualityThreshold float64
	TestPassRate     float64

	// Convergence criteria
	MaxIterations   int
	EarlyStopRounds int

	// Tool configuration
	EnableTools   bool
	EnableSandbox bool
	ToolTimeout   time.Duration

	Logger *logrus.Logger
}

// DefaultConfig returns sensible defaults for the debate system
func DefaultConfig() *Config {
	return &Config{
		EnableArchitect:   true,
		EnableGenerator:   true,
		EnableCritic:      true,
		EnableRefactoring: true,
		EnableTester:      true,
		EnableValidator:   true,
		EnableSecurity:    true,
		EnablePerformance: true,

		EnablePlanningPhase:    true,
		EnableGenerationPhase:  true,
		EnableDebatePhase:      true,
		EnableValidationPhase:  true,
		EnableRefactoringPhase: true,
		EnableIntegrationPhase: true,

		MaxRounds:        10,
		MinConsensus:     0.8,
		QualityThreshold: 0.95,
		TestPassRate:     0.95,

		MaxIterations:   10,
		EarlyStopRounds: 3,

		EnableTools:   true,
		EnableSandbox: true,
		ToolTimeout:   30 * time.Second,

		Logger: logrus.New(),
	}
}

// System is the main comprehensive debate system
type System struct {
	config       *Config
	logger       *logrus.Logger
	pool         *AgentPool
	orchestrator *PhaseOrchestrator
}

// NewSystem creates a new comprehensive debate system
func NewSystem(config *Config) *System {
	if config == nil {
		config = DefaultConfig()
	}

	pool := NewAgentPool(config.Logger)
	orchestrator := NewPhaseOrchestrator(pool, config.Logger)

	return &System{
		config:       config,
		logger:       config.Logger,
		pool:         pool,
		orchestrator: orchestrator,
	}
}

// DebateRequest represents a request to conduct a debate
type DebateRequest struct {
	ID        string
	Topic     string
	Context   string
	Codebase  string
	Language  string
	MaxRounds int
	Timeout   time.Duration
	Tools     []string
	Metadata  map[string]interface{}
}

// DebateResponse represents the result of a debate
type DebateResponse struct {
	ID              string
	Success         bool
	Consensus       *ConsensusResult
	Phases          []*PhaseResult
	Participants    []string
	RoundsConducted int
	Duration        time.Duration
	QualityScore    float64
	TestPassRate    float64
	LessonsLearned  []string
	CodeChanges     []CodeChange
	Metadata        map[string]interface{}
}

// PhaseResult represents a phase in the debate
type PhaseResult struct {
	Phase          string
	Round          int
	Responses      []AgentResponse
	ConsensusLevel float64
	KeyInsights    []string
	Duration       time.Duration
}

// DebateConfig represents configuration for a debate
type DebateConfig struct {
	Topic        string
	MaxRounds    int
	EnableCognee bool
	Strategy     string
	Participants []ParticipantConfig
}

// ParticipantConfig represents configuration for a participant
type ParticipantConfig struct {
	Role     Role
	Provider string
	Model    string
}

// CodeChange represents a code change
type CodeChange struct {
	FilePath   string
	ChangeType string // "create", "update", "delete"
	OldContent string
	NewContent string
	Rationale  string
	AgentID    string
	Validated  bool
}

// ConductDebate conducts a comprehensive multi-agent debate
func (s *System) ConductDebate(ctx context.Context, req *DebateRequest) (*DebateResponse, error) {
	startTime := time.Now()
	s.logger.WithFields(logrus.Fields{
		"debate_id": req.ID,
		"topic":     req.Topic,
	}).Info("Starting comprehensive multi-agent debate")

	// Validate request
	if req.Topic == "" {
		return nil, fmt.Errorf("debate topic cannot be empty")
	}

	// Set defaults
	maxRounds := req.MaxRounds
	if maxRounds == 0 {
		maxRounds = s.config.MaxRounds
	}

	response := &DebateResponse{
		ID:             req.ID,
		Success:        false,
		Consensus:      NewConsensusResult(),
		Phases:         make([]*PhaseResult, 0),
		Participants:   make([]string, 0),
		LessonsLearned: make([]string, 0),
		CodeChanges:    make([]CodeChange, 0),
		Metadata:       make(map[string]interface{}),
	}

	// Phase 0: Planning
	if s.config.EnablePlanningPhase {
		if err := s.runPlanningPhase(ctx, req, response); err != nil {
			s.logger.WithError(err).Warn("Planning phase failed")
		}
	}

	// Phase 1: Generation
	if s.config.EnableGenerationPhase {
		if err := s.runGenerationPhase(ctx, req, response); err != nil {
			s.logger.WithError(err).Warn("Generation phase failed")
		}
	}

	// Phase 2-3: Debate rounds
	if s.config.EnableDebatePhase {
		for round := 1; round <= maxRounds; round++ {
			s.logger.WithField("round", round).Info("Debate round")

			// Run debate round
			phaseResult, err := s.runDebateRound(ctx, req, response, round)
			if err != nil {
				s.logger.WithError(err).WithField("round", round).Warn("Debate round failed")
				continue
			}

			response.Phases = append(response.Phases, phaseResult)
			response.RoundsConducted = round

			// Check convergence
			if s.checkConvergence(response) {
				s.logger.WithField("round", round).Info("Debate converged")
				break
			}
		}
	}

	// Phase 4: Validation
	if s.config.EnableValidationPhase {
		if err := s.runValidationPhase(ctx, req, response); err != nil {
			s.logger.WithError(err).Warn("Validation phase failed")
		}
	}

	// Phase 5: Refactoring
	if s.config.EnableRefactoringPhase {
		if err := s.runRefactoringPhase(ctx, req, response); err != nil {
			s.logger.WithError(err).Warn("Refactoring phase failed")
		}
	}

	// Phase 6: Integration
	if s.config.EnableIntegrationPhase {
		if err := s.runIntegrationPhase(ctx, req, response); err != nil {
			s.logger.WithError(err).Warn("Integration phase failed")
		}
	}

	// Calculate final metrics
	response.Duration = time.Since(startTime)
	response.Success = response.Consensus != nil && response.Consensus.Reached

	s.logger.WithFields(logrus.Fields{
		"debate_id":     req.ID,
		"success":       response.Success,
		"rounds":        response.RoundsConducted,
		"duration":      response.Duration,
		"quality_score": response.QualityScore,
	}).Info("Comprehensive debate completed")

	return response, nil
}

// runPlanningPhase executes the planning phase using the PhaseOrchestrator
func (s *System) runPlanningPhase(ctx context.Context, req *DebateRequest, resp *DebateResponse) error {
	debateCtx := &Context{
		Topic:     req.Topic,
		Language:  req.Language,
		Responses: make([]*AgentResponse, 0),
		Artifacts: make(map[string]*Artifact),
	}
	phase, err := s.orchestrator.PlanningPhase(ctx, req, debateCtx)
	if err != nil {
		return err
	}
	resp.Phases = append(resp.Phases, phase)
	return nil
}

// runGenerationPhase executes the generation phase using the PhaseOrchestrator
func (s *System) runGenerationPhase(ctx context.Context, req *DebateRequest, resp *DebateResponse) error {
	debateCtx := &Context{
		Topic:     req.Topic,
		Language:  req.Language,
		Responses: make([]*AgentResponse, 0),
		Artifacts: make(map[string]*Artifact),
	}
	phase, err := s.orchestrator.GenerationPhase(ctx, req, debateCtx)
	if err != nil {
		return err
	}
	resp.Phases = append(resp.Phases, phase)
	return nil
}

// runDebateRound executes a debate round using the PhaseOrchestrator
func (s *System) runDebateRound(ctx context.Context, req *DebateRequest, resp *DebateResponse, round int) (*PhaseResult, error) {
	debateCtx := &Context{
		Topic:     req.Topic,
		Language:  req.Language,
		Responses: make([]*AgentResponse, 0),
		Artifacts: make(map[string]*Artifact),
	}
	// Ensure a code artifact exists for the debate phase
	debateCtx.Artifacts["code-"+req.ID] = &Artifact{
		ID:   "code-" + req.ID,
		Type: ArtifactTypeCode,
		Name: "implementation.go",
	}
	return s.orchestrator.DebatePhase(ctx, req, debateCtx, round)
}

// runValidationPhase executes the validation phase using the PhaseOrchestrator
func (s *System) runValidationPhase(ctx context.Context, req *DebateRequest, resp *DebateResponse) error {
	debateCtx := &Context{
		Topic:     req.Topic,
		Language:  req.Language,
		Responses: make([]*AgentResponse, 0),
		Artifacts: make(map[string]*Artifact),
	}
	phase, err := s.orchestrator.ValidationPhase(ctx, req, debateCtx)
	if err != nil {
		return err
	}
	resp.Phases = append(resp.Phases, phase)
	return nil
}

// runRefactoringPhase executes the refactoring phase using the PhaseOrchestrator
func (s *System) runRefactoringPhase(ctx context.Context, req *DebateRequest, resp *DebateResponse) error {
	debateCtx := &Context{
		Topic:     req.Topic,
		Language:  req.Language,
		Responses: make([]*AgentResponse, 0),
		Artifacts: make(map[string]*Artifact),
	}
	phase, err := s.orchestrator.RefactoringPhase(ctx, req, debateCtx)
	if err != nil {
		return err
	}
	resp.Phases = append(resp.Phases, phase)
	return nil
}

// runIntegrationPhase executes the integration phase using the PhaseOrchestrator
func (s *System) runIntegrationPhase(ctx context.Context, req *DebateRequest, resp *DebateResponse) error {
	debateCtx := &Context{
		Topic:     req.Topic,
		Language:  req.Language,
		Responses: make([]*AgentResponse, 0),
		Artifacts: make(map[string]*Artifact),
	}
	phase, err := s.orchestrator.IntegrationPhase(ctx, req, debateCtx)
	if err != nil {
		return err
	}
	resp.Phases = append(resp.Phases, phase)
	return nil
}

// checkConvergence checks if the debate has converged based on quality and consensus
func (s *System) checkConvergence(resp *DebateResponse) bool {
	if len(resp.Phases) == 0 {
		return false
	}

	// Check if quality threshold is met
	if resp.QualityScore >= s.config.QualityThreshold {
		return true
	}

	// Check if consensus level in recent phases exceeds minimum
	lastPhase := resp.Phases[len(resp.Phases)-1]
	if lastPhase.ConsensusLevel >= s.config.MinConsensus {
		return true
	}

	// Check early stopping: if last N rounds show no improvement
	if resp.RoundsConducted >= s.config.EarlyStopRounds && len(resp.Phases) >= s.config.EarlyStopRounds {
		noImprovement := true
		for i := len(resp.Phases) - s.config.EarlyStopRounds; i < len(resp.Phases)-1; i++ {
			if resp.Phases[i+1].ConsensusLevel > resp.Phases[i].ConsensusLevel {
				noImprovement = false
				break
			}
		}
		if noImprovement {
			s.logger.Info("Early stopping: no improvement in recent rounds")
			return true
		}
	}

	return false
}
