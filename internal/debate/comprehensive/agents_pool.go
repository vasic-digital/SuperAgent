package comprehensive

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// AgentPool manages a pool of agents
type AgentPool struct {
	agents map[string]*Agent
	byRole map[Role][]*Agent
	byID   map[string]*Agent
	mu     sync.RWMutex
	logger *logrus.Logger
}

// NewAgentPool creates a new agent pool
func NewAgentPool(logger *logrus.Logger) *AgentPool {
	if logger == nil {
		logger = logrus.New()
	}

	return &AgentPool{
		agents: make(map[string]*Agent),
		byRole: make(map[Role][]*Agent),
		byID:   make(map[string]*Agent),
		logger: logger,
	}
}

// Add adds an agent to the pool
func (p *AgentPool) Add(agent *Agent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.agents[agent.ID] = agent
	p.byID[agent.ID] = agent
	p.byRole[agent.Role] = append(p.byRole[agent.Role], agent)

	p.logger.WithFields(logrus.Fields{
		"agent_id":   agent.ID,
		"agent_role": agent.Role,
		"provider":   agent.Provider,
	}).Debug("Agent added to pool")
}

// Get retrieves an agent by ID
func (p *AgentPool) Get(id string) (*Agent, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, ok := p.byID[id]
	return agent, ok
}

// GetByRole retrieves all agents with a specific role
func (p *AgentPool) GetByRole(role Role) []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := p.byRole[role]
	result := make([]*Agent, len(agents))
	copy(result, agents)
	return result
}

// GetAll retrieves all agents
func (p *AgentPool) GetAll() []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Agent, 0, len(p.agents))
	for _, agent := range p.agents {
		result = append(result, agent)
	}
	return result
}

// GetActive returns only active agents
func (p *AgentPool) GetActive() []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*Agent
	for _, agent := range p.agents {
		if agent.IsActive {
			result = append(result, agent)
		}
	}
	return result
}

// GetActiveByRole returns active agents with a specific role
func (p *AgentPool) GetActiveByRole(role Role) []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*Agent
	for _, agent := range p.byRole[role] {
		if agent.IsActive {
			result = append(result, agent)
		}
	}
	return result
}

// Size returns the total number of agents
func (p *AgentPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.agents)
}

// SizeByRole returns the number of agents with a specific role
func (p *AgentPool) SizeByRole(role Role) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.byRole[role])
}

// Remove removes an agent from the pool
func (p *AgentPool) Remove(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, ok := p.byID[id]
	if !ok {
		return false
	}

	delete(p.agents, id)
	delete(p.byID, id)

	// Remove from role index
	agents := p.byRole[agent.Role]
	for i, a := range agents {
		if a.ID == id {
			p.byRole[agent.Role] = append(agents[:i], agents[i+1:]...)
			break
		}
	}

	p.logger.WithField("agent_id", id).Debug("Agent removed from pool")
	return true
}

// SelectBestForRole selects the best agent for a role based on score
func (p *AgentPool) SelectBestForRole(role Role) *Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := p.byRole[role]
	if len(agents) == 0 {
		return nil
	}

	var best *Agent
	bestScore := -1.0

	for _, agent := range agents {
		if agent.IsActive && agent.Score > bestScore {
			best = agent
			bestScore = agent.Score
		}
	}

	return best
}

// SelectTopNForRole selects the top N agents for a role
func (p *AgentPool) SelectTopNForRole(role Role, n int) []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := p.byRole[role]
	if len(agents) == 0 {
		return nil
	}

	// Filter active agents
	var active []*Agent
	for _, agent := range agents {
		if agent.IsActive {
			active = append(active, agent)
		}
	}

	// Sort by score (simple bubble sort for small N)
	for i := 0; i < len(active); i++ {
		for j := i + 1; j < len(active); j++ {
			if active[j].Score > active[i].Score {
				active[i], active[j] = active[j], active[i]
			}
		}
	}

	// Return top N
	if n > len(active) {
		n = len(active)
	}
	return active[:n]
}

// HasRole checks if any agent has a specific role
func (p *AgentPool) HasRole(role Role) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.byRole[role]) > 0
}

// Clear removes all agents from the pool
func (p *AgentPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.agents = make(map[string]*Agent)
	p.byRole = make(map[Role][]*Agent)
	p.byID = make(map[string]*Agent)

	p.logger.Debug("Agent pool cleared")
}

// AgentFactory creates agents
type AgentFactory struct {
	pool   *AgentPool
	logger *logrus.Logger
}

// NewAgentFactory creates a new agent factory
func NewAgentFactory(pool *AgentPool, logger *logrus.Logger) *AgentFactory {
	if logger == nil {
		logger = logrus.New()
	}

	return &AgentFactory{
		pool:   pool,
		logger: logger,
	}
}

// CreateAgent creates a new agent with the given parameters
func (f *AgentFactory) CreateAgent(role Role, provider, model string, score float64) (*Agent, error) {
	if !role.IsValid() {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	if provider == "" {
		return nil, fmt.Errorf("provider cannot be empty")
	}

	if model == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}

	agent := NewAgent(role, provider, model, score)

	if f.pool != nil {
		f.pool.Add(agent)
	}

	f.logger.WithFields(logrus.Fields{
		"agent_id":   agent.ID,
		"agent_role": agent.Role,
		"provider":   agent.Provider,
		"model":      agent.Model,
	}).Info("Agent created")

	return agent, nil
}

// CreateTeam creates a full team of agents
func (f *AgentFactory) CreateTeam(config map[Role]AgentConfig) ([]*Agent, error) {
	var agents []*Agent

	for role, cfg := range config {
		agent, err := f.CreateAgent(role, cfg.Provider, cfg.Model, cfg.Score)
		if err != nil {
			return nil, fmt.Errorf("failed to create agent for role %s: %w", role, err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// AgentConfig represents configuration for creating an agent
type AgentConfig struct {
	Provider string
	Model    string
	Score    float64
}

// BaseAgent provides a base implementation of AgentInterface
type BaseAgent struct {
	agent  *Agent
	pool   *AgentPool
	logger *logrus.Logger
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(agent *Agent, pool *AgentPool, logger *logrus.Logger) *BaseAgent {
	if logger == nil {
		logger = logrus.New()
	}

	return &BaseAgent{
		agent:  agent,
		pool:   pool,
		logger: logger,
	}
}

// GetID returns the agent's ID
func (b *BaseAgent) GetID() string {
	return b.agent.ID
}

// GetRole returns the agent's role
func (b *BaseAgent) GetRole() Role {
	return b.agent.Role
}

// GetCapabilities returns the agent's capabilities
func (b *BaseAgent) GetCapabilities() []Capability {
	return b.agent.Capabilities
}

// CanHandle checks if the agent can handle a task
func (b *BaseAgent) CanHandle(taskType string) bool {
	// Base implementation - override in specific agents
	return true
}

// UpdateScore updates the agent's score
func (b *BaseAgent) UpdateScore(score float64) {
	b.agent.Score = score
}

// Process processes a message - must be overridden
func (b *BaseAgent) Process(ctx context.Context, msg *Message, context *Context) (*AgentResponse, error) {
	return nil, fmt.Errorf("Process method must be implemented by specific agent")
}
