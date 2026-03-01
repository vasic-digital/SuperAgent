package comprehensive

import (
	"context"
	"fmt"

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
