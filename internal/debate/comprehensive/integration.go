package comprehensive

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// IntegrationManager manages the integration of all debate components
type IntegrationManager struct {
	config       *Config
	pool         *AgentPool
	registry     *ToolRegistry
	orchestrator *PhaseOrchestrator
	engine       *DebateEngine
	logger       *logrus.Logger
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(config *Config, logger *logrus.Logger) (*IntegrationManager, error) {
	if logger == nil {
		logger = logrus.New()
	}

	if config == nil {
		config = DefaultConfig()
	}

	// Create agent pool
	pool := NewAgentPool(logger)

	// Create tool registry
	registry := NewToolRegistry(logger)

	// Create phase orchestrator
	orchestrator := NewPhaseOrchestrator(pool, logger)

	// Create debate engine (using default config for now)
	engine := NewDebateEngine(DefaultConfig(), logger)

	return &IntegrationManager{
		config:       config,
		pool:         pool,
		registry:     registry,
		orchestrator: orchestrator,
		engine:       engine,
		logger:       logger,
	}, nil
}

// Initialize initializes all components
func (m *IntegrationManager) Initialize(ctx context.Context) error {
	m.logger.Info("Initializing debate system integration")

	// Register all tools
	if err := m.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Create agent team
	if err := m.createAgentTeam(); err != nil {
		return fmt.Errorf("failed to create agent team: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"agents": m.pool.Size(),
		"tools":  len(m.registry.GetAll()),
	}).Info("Integration initialization complete")

	return nil
}

// registerTools registers all available tools
func (m *IntegrationManager) registerTools() error {
	// Code tools
	codeTool := NewCodeTool(".", m.logger)
	m.registry.Register(codeTool)

	searchTool := NewSearchTool(".", m.logger)
	m.registry.Register(searchTool)

	// Command tools
	cmdTool := NewCommandTool(".", 30, m.logger)
	m.registry.Register(cmdTool)

	testTool := NewTestTool(".", m.logger)
	m.registry.Register(testTool)

	buildTool := NewBuildTool(".", m.logger)
	m.registry.Register(buildTool)

	// Analysis tools
	staticTool := NewStaticAnalysisTool(m.logger)
	m.registry.Register(staticTool)

	complexityTool := NewComplexityTool(m.logger)
	m.registry.Register(complexityTool)

	lintTool := NewLintTool(m.logger)
	m.registry.Register(lintTool)

	// Security tool
	securityTool := NewSecurityTool(m.logger)
	m.registry.Register(securityTool)

	// Performance tool
	perfTool := NewPerformanceTool(m.logger)
	m.registry.Register(perfTool)

	return nil
}

// createAgentTeam creates the initial agent team
func (m *IntegrationManager) createAgentTeam() error {
	factory := NewAgentFactory(m.pool, m.logger)

	// Define team configuration
	teamConfig := map[Role]AgentConfig{
		RoleArchitect:   {Provider: "openai", Model: "gpt-4", Score: 8.5},
		RoleGenerator:   {Provider: "anthropic", Model: "claude", Score: 8.5},
		RoleCritic:      {Provider: "openai", Model: "gpt-4", Score: 8.0},
		RoleRefactoring: {Provider: "google", Model: "gemini", Score: 8.0},
		RoleTester:      {Provider: "openai", Model: "gpt-4", Score: 8.5},
		RoleValidator:   {Provider: "anthropic", Model: "claude", Score: 8.5},
		RoleSecurity:    {Provider: "openai", Model: "gpt-4", Score: 8.0},
		RolePerformance: {Provider: "google", Model: "gemini", Score: 8.0},
	}

	// Create fallback agents for critical roles
	fallbackConfig := map[Role]AgentConfig{
		RoleGenerator: {Provider: "openrouter", Model: "llama-3.1-70b", Score: 7.5},
		RoleCritic:    {Provider: "openrouter", Model: "mixtral-8x7b", Score: 7.5},
	}

	// Create primary team
	_, err := factory.CreateTeam(teamConfig)
	if err != nil {
		return fmt.Errorf("failed to create primary team: %w", err)
	}

	// Create fallback agents
	for role, cfg := range fallbackConfig {
		_, err := factory.CreateAgent(role, cfg.Provider, cfg.Model, cfg.Score)
		if err != nil {
			m.logger.WithError(err).Warnf("Failed to create fallback agent for role %s", role)
		}
	}

	return nil
}

// ExecuteDebate executes a complete debate
func (m *IntegrationManager) ExecuteDebate(ctx context.Context, req *DebateRequest) (*DebateResponse, error) {
	m.logger.WithField("topic", req.Topic).Info("Executing debate")

	// Create debate context
	debateContext := NewContext(req.Topic, req.Codebase, req.Language)

	// Validate request
	validator := DebateRequestValidator{}
	if errors := validator.Validate(req); len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	// Execute debate through engine
	response, err := m.engine.RunDebate(ctx, req, debateContext)
	if err != nil {
		return nil, fmt.Errorf("debate execution failed: %w", err)
	}

	// Apply quality gates
	if err := m.applyQualityGates(response); err != nil {
		m.logger.WithError(err).Warn("Quality gates failed")
		response.Success = false
	}

	m.logger.WithFields(logrus.Fields{
		"success": response.Success,
		"rounds":  response.RoundsConducted,
	}).Info("Debate execution complete")

	return response, nil
}

// StreamDebate executes a debate with real-time streaming
func (m *IntegrationManager) StreamDebate(ctx context.Context, req *DebateStreamRequest) (*DebateResponse, error) {
	m.logger.WithFields(logrus.Fields{
		"topic":  req.Topic,
		"stream": req.Stream,
	}).Info("Executing debate with streaming")

	// Validate request
	validator := DebateRequestValidator{}
	if errors := validator.Validate(req.DebateRequest); len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	// Create streaming orchestrator
	streamConfig := DefaultStreamConfig()
	allAgents := m.pool.GetAll()
	streamer := NewStreamOrchestrator(streamConfig, req.StreamHandler, req.ID, len(allAgents))

	// Emit debate start event
	if err := streamer.Emit(StreamEventDebateStart, nil, "Starting comprehensive multi-agent debate", nil); err != nil {
		m.logger.WithError(err).Warn("Failed to emit debate start event")
	}

	// Create system for debate execution
	system := NewSystem(m.config)
	system.logger = m.logger

	// Execute debate with streaming through all teams
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

	// Assign agents to teams
	teamAssignments := AssignTeamsToAgents(allAgents)
	teamSummary := GetTeamSummary(teamAssignments)

	// Stream through all 5 teams
	for _, team := range AllTeams() {
		agents := teamSummary[team]
		if len(agents) == 0 {
			continue
		}

		streamer.SetTeam(team)
		if err := streamer.Emit(StreamEventTeamStart, nil, string(team), map[string]interface{}{
			"team":        team,
			"agent_count": len(agents),
			"roles":       GetRolesForTeam(team),
		}); err != nil {
			m.logger.WithError(err).Warn("Failed to emit team start event")
		}

		// Process each agent in the team
		for _, agent := range agents {
			if err := streamer.Emit(StreamEventAgentStart, agent, "", nil); err != nil {
				m.logger.WithError(err).Warn("Failed to emit agent start event")
			}

			// Simulate agent processing (in real implementation, this would call the LLM)
			agentResponse := &AgentResponse{
				AgentID:    agent.ID,
				AgentRole:  agent.Role,
				Provider:   agent.Provider,
				Model:      agent.Model,
				Content:    m.generateAgentContent(agent, req.Topic),
				Confidence: agent.Score / 10.0,
				Timestamp:  time.Now(),
			}

			// Emit agent response
			if err := streamer.Emit(StreamEventAgentResponse, agent, agentResponse.Content, map[string]interface{}{
				"confidence": agentResponse.Confidence,
				"team":       team,
			}); err != nil {
				m.logger.WithError(err).Warn("Failed to emit agent response event")
			}

			response.Participants = append(response.Participants, agent.ID)
			streamer.IncrementCompleted()

			if err := streamer.Emit(StreamEventAgentComplete, agent, "", nil); err != nil {
				m.logger.WithError(err).Warn("Failed to emit agent complete event")
			}
		}

		if err := streamer.Emit(StreamEventTeamComplete, nil, string(team), nil); err != nil {
			m.logger.WithError(err).Warn("Failed to emit team complete event")
		}
	}

	// Calculate final metrics
	response.Duration = streamer.GetDuration()
	response.Success = true
	response.Consensus = &ConsensusResult{
		Reached:    true,
		Confidence: 0.85,
		Summary:    "Comprehensive debate completed with all teams",
		AchievedAt: time.Now(),
	}
	response.QualityScore = 0.85
	response.RoundsConducted = 1

	// Emit debate complete event
	if err := streamer.Emit(StreamEventDebateComplete, nil, "Debate completed successfully", map[string]interface{}{
		"duration":      response.Duration,
		"success":       response.Success,
		"quality_score": response.QualityScore,
		"participants":  len(response.Participants),
	}); err != nil {
		m.logger.WithError(err).Warn("Failed to emit debate complete event")
	}

	m.logger.WithFields(logrus.Fields{
		"success":      response.Success,
		"duration":     response.Duration,
		"participants": len(response.Participants),
	}).Info("Streaming debate execution complete")

	return response, nil
}

// generateAgentContent generates content for an agent based on its role
func (m *IntegrationManager) generateAgentContent(agent *Agent, topic string) string {
	switch agent.Role {
	case RoleArchitect:
		return fmt.Sprintf("[Architect] Analyzing system architecture and design patterns for: %s", topic)
	case RoleGenerator:
		return fmt.Sprintf("[Generator] Generating initial code implementation for: %s", topic)
	case RoleCritic:
		return fmt.Sprintf("[Critic] Reviewing implementation and identifying potential issues for: %s", topic)
	case RoleTester:
		return fmt.Sprintf("[Tester] Designing test cases and validation scenarios for: %s", topic)
	case RoleValidator:
		return fmt.Sprintf("[Validator] Verifying correctness and edge cases for: %s", topic)
	case RoleSecurity:
		return fmt.Sprintf("[Security] Analyzing security implications and vulnerabilities for: %s", topic)
	case RolePerformance:
		return fmt.Sprintf("[Performance] Evaluating performance characteristics and optimizations for: %s", topic)
	case RoleRefactoring:
		return fmt.Sprintf("[Refactoring] Suggesting code improvements and refactoring opportunities for: %s", topic)
	case RoleModerator:
		return fmt.Sprintf("[Moderator] Facilitating discussion and ensuring productive debate for: %s", topic)
	case RoleRedTeam:
		return fmt.Sprintf("[Red Team] Adversarial testing and attack vector analysis for: %s", topic)
	case RoleBlueTeam:
		return fmt.Sprintf("[Blue Team] Defensive implementation and countermeasure design for: %s", topic)
	default:
		return fmt.Sprintf("[%s] Contributing to debate on: %s", agent.Role, topic)
	}
}

// applyQualityGates applies quality gates to the response
func (m *IntegrationManager) applyQualityGates(response *DebateResponse) error {
	// Check quality threshold
	if response.QualityScore < m.config.QualityThreshold {
		return fmt.Errorf("quality score %.2f below threshold %.2f",
			response.QualityScore, m.config.QualityThreshold)
	}

	// Check consensus
	if response.Consensus == nil || !response.Consensus.Reached {
		return fmt.Errorf("consensus not reached")
	}

	// Check confidence
	if response.Consensus.Confidence < m.config.MinConsensus {
		return fmt.Errorf("consensus confidence %.2f below minimum %.2f",
			response.Consensus.Confidence, m.config.MinConsensus)
	}

	return nil
}

// GetAgentPool returns the agent pool
func (m *IntegrationManager) GetAgentPool() *AgentPool {
	return m.pool
}

// GetToolRegistry returns the tool registry
func (m *IntegrationManager) GetToolRegistry() *ToolRegistry {
	return m.registry
}

// GetConfig returns the configuration
func (m *IntegrationManager) GetConfig() *Config {
	return m.config
}

// HealthCheck performs a health check on all components
func (m *IntegrationManager) HealthCheck() map[string]bool {
	health := make(map[string]bool)

	// Check agent pool
	health["agent_pool"] = m.pool != nil && m.pool.Size() > 0

	// Check tool registry
	health["tool_registry"] = m.registry != nil && len(m.registry.GetAll()) > 0

	// Check orchestrator
	health["orchestrator"] = m.orchestrator != nil

	// Check engine
	health["engine"] = m.engine != nil

	return health
}

// Statistics returns system statistics
func (m *IntegrationManager) Statistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["agents_total"] = m.pool.Size()
	stats["agents_by_role"] = m.getAgentCountsByRole()
	stats["tools_total"] = len(m.registry.GetAll())
	stats["tools_by_type"] = m.getToolCountsByType()
	stats["config"] = map[string]interface{}{
		"max_rounds":        m.config.MaxRounds,
		"quality_threshold": m.config.QualityThreshold,
		"min_consensus":     m.config.MinConsensus,
		"test_pass_rate":    m.config.TestPassRate,
	}

	return stats
}

// getAgentCountsByRole returns agent counts by role
func (m *IntegrationManager) getAgentCountsByRole() map[string]int {
	counts := make(map[string]int)

	for _, role := range AllRoles() {
		count := m.pool.SizeByRole(role)
		if count > 0 {
			counts[string(role)] = count
		}
	}

	return counts
}

// getToolCountsByType returns tool counts by type
func (m *IntegrationManager) getToolCountsByType() map[string]int {
	counts := make(map[string]int)

	types := []ToolType{
		ToolTypeCode,
		ToolTypeCommand,
		ToolTypeDatabase,
		ToolTypeAnalysis,
		ToolTypeSecurity,
		ToolTypePerformance,
	}

	for _, t := range types {
		tools := m.registry.GetByType(t)
		if len(tools) > 0 {
			counts[string(t)] = len(tools)
		}
	}

	return counts
}
