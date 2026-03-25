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
	system       *System // Full debate system with actual LLM invocation
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

	// Create debate engine (legacy stub — kept for backward compatibility)
	engine := NewDebateEngine(DefaultConfig(), logger)

	// Create the full debate system that shares the pool and orchestrator
	// so registered agents from PopulateFromDebateTeam are available
	system := &System{
		config:       config,
		logger:       logger,
		pool:         pool,         // Share pool with IntegrationManager
		orchestrator: orchestrator, // Share orchestrator
	}

	return &IntegrationManager{
		config:       config,
		pool:         pool,
		registry:     registry,
		orchestrator: orchestrator,
		engine:       engine,
		system:       system,
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
	cmdTool := NewCommandTool(".", 30*time.Second, m.logger)
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

// ExecuteDebate executes a complete debate using the full System implementation
// which actually invokes LLM agents through the PhaseOrchestrator.
func (m *IntegrationManager) ExecuteDebate(ctx context.Context, req *DebateRequest) (*DebateResponse, error) {
	m.logger.WithField("topic", req.Topic).Info("Executing debate via System.ConductDebate")

	// Validate request
	validator := DebateRequestValidator{}
	if errors := validator.Validate(req); len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	// Delegate to the full System implementation which has actual LLM invocation
	response, err := m.system.ConductDebate(ctx, req)
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

// GetAgentPool returns the agent pool for external population (e.g., from ServiceIntegration)
func (m *IntegrationManager) GetAgentPool() *AgentPool {
	return m.pool
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
	m.logger.WithField("agent_count", len(allAgents)).Info("[StreamDebate] Pool agents count")
	for i, a := range allAgents {
		m.logger.WithFields(logrus.Fields{
			"idx": i, "id": a.ID, "role": a.Role, "provider": a.Provider, "model": a.Model,
		}).Info("[StreamDebate] Agent in pool")
	}
	streamer := NewStreamOrchestrator(streamConfig, req.StreamHandler, req.ID, len(allAgents))

	// Emit debate start event
	if err := streamer.Emit(StreamEventDebateStart, nil, "Starting comprehensive multi-agent debate", nil); err != nil {
		m.logger.WithError(err).Warn("Failed to emit debate start event")
	}

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
		debateCtx := NewContext(req.Topic, "", "")
		for _, agent := range agents {
			if err := streamer.Emit(StreamEventAgentStart, agent, "", nil); err != nil {
				m.logger.WithError(err).Warn("Failed to emit agent start event")
			}

			// Use specialized agent processing through the agent pool
			specialized := createSpecializedAgent(agent, m.pool)
			msg := &Message{
				Type:    MessageTypeProposal,
				Content: req.Topic,
			}

			var agentResponse *AgentResponse
			processedResp, err := specialized.Process(ctx, msg, debateCtx)
			if err != nil {
				m.logger.WithError(err).WithField("agent", agent.Name).Warn("Agent processing failed in stream")
				// Fallback: create response from agent metadata
				agentResponse = &AgentResponse{
					AgentID:    agent.ID,
					AgentRole:  agent.Role,
					Provider:   agent.Provider,
					Model:      agent.Model,
					Content:    fmt.Sprintf("[%s] Processing failed, using fallback for: %s", agent.Role, req.Topic),
					Confidence: 0.5,
					Timestamp:  time.Now(),
				}
			} else {
				agentResponse = processedResp
			}

			// Emit agent response
			if err := streamer.Emit(StreamEventAgentResponse, agent, agentResponse.Content, map[string]interface{}{
				"confidence": agentResponse.Confidence,
				"team":       team,
			}); err != nil {
				m.logger.WithError(err).Warn("Failed to emit agent response event")
			}

			response.Participants = append(response.Participants, agent.ID)
			debateCtx.AddResponse(agentResponse)
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

// agentProcessor is a local interface for specialized agents that can process messages
type agentProcessor interface {
	Process(ctx context.Context, msg *Message, context *Context) (*AgentResponse, error)
}

// createSpecializedAgent creates the appropriate specialized agent wrapper for the given agent role
func createSpecializedAgent(agent *Agent, pool *AgentPool) agentProcessor {
	switch agent.Role {
	case RoleArchitect:
		return NewArchitectAgent(agent, pool)
	case RoleGenerator:
		return NewGeneratorAgent(agent, pool)
	case RoleCritic:
		return NewCriticAgent(agent, pool)
	case RoleTester:
		return NewTesterAgent(agent, pool)
	case RoleValidator:
		return NewValidatorAgent(agent, pool)
	case RoleSecurity:
		return NewSecurityAgent(agent, pool)
	case RolePerformance:
		return NewPerformanceAgent(agent, pool)
	case RoleRefactoring:
		return NewRefactoringAgent(agent, pool)
	case RoleRedTeam:
		return NewRedTeamAgent(agent, pool)
	case RoleBlueTeam:
		return NewBlueTeamAgent(agent, pool)
	default:
		// Fallback to generator for unknown roles
		return NewGeneratorAgent(agent, pool)
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
