// Package agents provides the agent factory for creating and managing specialized agents.
package agents

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"dev.helix.agent/internal/debate/topology"
)

// AgentFactory creates specialized agents from templates and provider information.
type AgentFactory struct {
	templateRegistry *TemplateRegistry
	discoverer       CapabilityDiscoverer
	mu               sync.RWMutex
}

// NewAgentFactory creates a new agent factory.
func NewAgentFactory() *AgentFactory {
	return &AgentFactory{
		templateRegistry: NewTemplateRegistry(),
	}
}

// NewAgentFactoryWithRegistry creates a factory with a custom registry.
func NewAgentFactoryWithRegistry(registry *TemplateRegistry) *AgentFactory {
	return &AgentFactory{
		templateRegistry: registry,
	}
}

// SetCapabilityDiscoverer sets the capability discoverer for runtime discovery.
func (f *AgentFactory) SetCapabilityDiscoverer(discoverer CapabilityDiscoverer) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.discoverer = discoverer
}

// GetTemplateRegistry returns the template registry.
func (f *AgentFactory) GetTemplateRegistry() *TemplateRegistry {
	return f.templateRegistry
}

// CreateFromTemplate creates a specialized agent from a template.
func (f *AgentFactory) CreateFromTemplate(templateID, provider, model string) (*SpecializedAgent, error) {
	template, ok := f.templateRegistry.Get(templateID)
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template.CreateAgent(provider, model)
}

// CreateForDomain creates a specialized agent for a specific domain.
func (f *AgentFactory) CreateForDomain(domain Domain, provider, model string) (*SpecializedAgent, error) {
	templates := f.templateRegistry.GetByDomain(domain)
	if len(templates) == 0 {
		// Create a basic agent with the domain
		return NewSpecializedAgent(fmt.Sprintf("%s Specialist", domain), provider, model, domain), nil
	}

	// Use the first matching template (highest priority)
	return templates[0].CreateAgent(provider, model)
}

// CreateForRole creates a specialized agent optimized for a specific role.
func (f *AgentFactory) CreateForRole(role topology.AgentRole, provider, model string) (*SpecializedAgent, error) {
	templates := f.templateRegistry.GetByRole(role)
	if len(templates) == 0 {
		// Create a general agent and assign the role
		agent := NewSpecializedAgent(fmt.Sprintf("%s Agent", role), provider, model, DomainGeneral)
		agent.PrimaryRole = role
		return agent, nil
	}

	// Use the first matching template
	agent, err := templates[0].CreateAgent(provider, model)
	if err != nil {
		return nil, err
	}

	// Ensure the role is set correctly
	agent.PrimaryRole = role
	return agent, nil
}

// CreateWithDiscovery creates an agent and performs capability discovery.
func (f *AgentFactory) CreateWithDiscovery(ctx context.Context, templateID, provider, model string) (*SpecializedAgent, error) {
	agent, err := f.CreateFromTemplate(templateID, provider, model)
	if err != nil {
		return nil, err
	}

	f.mu.RLock()
	discoverer := f.discoverer
	f.mu.RUnlock()

	if discoverer != nil {
		if err := agent.DiscoverCapabilities(ctx, discoverer); err != nil {
			// Log but don't fail - use template defaults
			agent.Metadata["discovery_error"] = err.Error()
		}
	}

	return agent, nil
}

// ProviderSpec describes a provider for agent creation.
type ProviderSpec struct {
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
	Score    float64 `json:"score"` // LLMsVerifier score
}

// CreateDebateTeam creates a complete set of agents for all debate roles.
func (f *AgentFactory) CreateDebateTeam(providers []ProviderSpec) ([]*SpecializedAgent, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider required")
	}

	agents := make([]*SpecializedAgent, 0)

	// Map roles to optimal domains
	roleDomainMap := map[topology.AgentRole]Domain{
		topology.RoleProposer:  DomainCode,
		topology.RoleCritic:    DomainSecurity,
		topology.RoleReviewer:  DomainGeneral,
		topology.RoleOptimizer: DomainOptimization,
		topology.RoleModerator: DomainReasoning,
		topology.RoleArchitect: DomainArchitecture,
		topology.RoleRedTeam:   DomainSecurity,
		topology.RoleBlueTeam:  DomainSecurity,
		topology.RoleValidator: DomainGeneral,
		topology.RoleTestAgent: DomainDebug,
		topology.RoleTeacher:   DomainReasoning,
	}

	// Sort providers by score (highest first)
	sortedProviders := make([]ProviderSpec, len(providers))
	copy(sortedProviders, providers)
	sort.Slice(sortedProviders, func(i, j int) bool {
		return sortedProviders[i].Score > sortedProviders[j].Score
	})

	roles := []topology.AgentRole{
		topology.RoleProposer, topology.RoleCritic, topology.RoleReviewer,
		topology.RoleOptimizer, topology.RoleModerator, topology.RoleArchitect,
		topology.RoleRedTeam, topology.RoleBlueTeam, topology.RoleValidator,
	}

	// Create agents, cycling through providers
	for i, role := range roles {
		provider := sortedProviders[i%len(sortedProviders)]
		domain := roleDomainMap[role]

		agent, err := f.CreateForDomain(domain, provider.Provider, provider.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to create agent for %s: %w", role, err)
		}

		agent.PrimaryRole = role
		agent.Score = provider.Score

		agents = append(agents, agent)
	}

	return agents, nil
}

// AgentPool manages a pool of specialized agents.
type AgentPool struct {
	agents   map[string]*SpecializedAgent
	byRole   map[topology.AgentRole][]*SpecializedAgent
	byDomain map[Domain][]*SpecializedAgent
	factory  *AgentFactory
	mu       sync.RWMutex
}

// NewAgentPool creates a new agent pool.
func NewAgentPool(factory *AgentFactory) *AgentPool {
	return &AgentPool{
		agents:   make(map[string]*SpecializedAgent),
		byRole:   make(map[topology.AgentRole][]*SpecializedAgent),
		byDomain: make(map[Domain][]*SpecializedAgent),
		factory:  factory,
	}
}

// Add adds an agent to the pool.
func (p *AgentPool) Add(agent *SpecializedAgent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.agents[agent.ID] = agent

	// Index by role
	p.byRole[agent.PrimaryRole] = append(p.byRole[agent.PrimaryRole], agent)

	// Index by domain
	domain := agent.Specialization.PrimaryDomain
	p.byDomain[domain] = append(p.byDomain[domain], agent)
}

// Get retrieves an agent by ID.
func (p *AgentPool) Get(id string) (*SpecializedAgent, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, ok := p.agents[id]
	return agent, ok
}

// GetByRole returns agents with the specified primary role.
func (p *AgentPool) GetByRole(role topology.AgentRole) []*SpecializedAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := p.byRole[role]
	result := make([]*SpecializedAgent, len(agents))
	copy(result, agents)
	return result
}

// GetByDomain returns agents with the specified primary domain.
func (p *AgentPool) GetByDomain(domain Domain) []*SpecializedAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := p.byDomain[domain]
	result := make([]*SpecializedAgent, len(agents))
	copy(result, agents)
	return result
}

// GetAll returns all agents in the pool.
func (p *AgentPool) GetAll() []*SpecializedAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*SpecializedAgent, 0, len(p.agents))
	for _, agent := range p.agents {
		result = append(result, agent)
	}
	return result
}

// Remove removes an agent from the pool.
func (p *AgentPool) Remove(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, ok := p.agents[id]
	if !ok {
		return
	}

	delete(p.agents, id)

	// Remove from role index
	roleAgents := p.byRole[agent.PrimaryRole]
	for i, a := range roleAgents {
		if a.ID == id {
			p.byRole[agent.PrimaryRole] = append(roleAgents[:i], roleAgents[i+1:]...)
			break
		}
	}

	// Remove from domain index
	domain := agent.Specialization.PrimaryDomain
	domainAgents := p.byDomain[domain]
	for i, a := range domainAgents {
		if a.ID == id {
			p.byDomain[domain] = append(domainAgents[:i], domainAgents[i+1:]...)
			break
		}
	}
}

// Size returns the number of agents in the pool.
func (p *AgentPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.agents)
}

// SelectBestForRole selects the best agent for a role based on composite scoring.
func (p *AgentPool) SelectBestForRole(role topology.AgentRole, preferredDomain Domain) *SpecializedAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var best *SpecializedAgent
	var bestScore float64 = -1

	for _, agent := range p.agents {
		score := agent.ScoreAgent(role, preferredDomain)
		if score.CompositeScore > bestScore {
			bestScore = score.CompositeScore
			best = agent
		}
	}

	return best
}

// SelectTopNForRole selects the top N agents for a role.
func (p *AgentPool) SelectTopNForRole(role topology.AgentRole, preferredDomain Domain, n int) []*SpecializedAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	type agentScore struct {
		agent *SpecializedAgent
		score float64
	}

	scores := make([]agentScore, 0, len(p.agents))
	for _, agent := range p.agents {
		s := agent.ScoreAgent(role, preferredDomain)
		scores = append(scores, agentScore{agent: agent, score: s.CompositeScore})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]*SpecializedAgent, 0, n)
	for i := 0; i < n && i < len(scores); i++ {
		result = append(result, scores[i].agent)
	}

	return result
}

// ToTopologyAgents converts all agents to topology.Agent for use in debates.
func (p *AgentPool) ToTopologyAgents() []*topology.Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*topology.Agent, 0, len(p.agents))
	for _, agent := range p.agents {
		result = append(result, agent.ToTopologyAgent())
	}
	return result
}

// TeamBuilder builds optimized debate teams from an agent pool.
type TeamBuilder struct {
	pool *AgentPool
}

// NewTeamBuilder creates a new team builder.
func NewTeamBuilder(pool *AgentPool) *TeamBuilder {
	return &TeamBuilder{pool: pool}
}

// TeamConfig configures team building.
type TeamConfig struct {
	RequiredRoles    []topology.AgentRole          // Roles that must be filled
	PreferredDomains map[topology.AgentRole]Domain // Preferred domain per role
	MinAgentsPerRole int                           // Minimum agents per role (for fallback)
	AllowRoleSharing bool                          // Allow same agent for multiple roles
}

// DefaultTeamConfig returns a sensible default configuration.
func DefaultTeamConfig() *TeamConfig {
	return &TeamConfig{
		RequiredRoles: []topology.AgentRole{
			topology.RoleProposer, topology.RoleCritic, topology.RoleReviewer,
			topology.RoleOptimizer, topology.RoleModerator,
		},
		PreferredDomains: map[topology.AgentRole]Domain{
			topology.RoleProposer:  DomainCode,
			topology.RoleCritic:    DomainSecurity,
			topology.RoleReviewer:  DomainGeneral,
			topology.RoleOptimizer: DomainOptimization,
			topology.RoleModerator: DomainReasoning,
			topology.RoleArchitect: DomainArchitecture,
			topology.RoleRedTeam:   DomainSecurity,
			topology.RoleBlueTeam:  DomainSecurity,
			topology.RoleValidator: DomainGeneral,
		},
		MinAgentsPerRole: 1,
		AllowRoleSharing: false,
	}
}

// TeamAssignment represents an agent assigned to a role.
type TeamAssignment struct {
	Agent     *SpecializedAgent
	Role      topology.AgentRole
	Score     *AgentScore
	IsPrimary bool
}

// BuildTeam builds an optimized team based on configuration.
func (tb *TeamBuilder) BuildTeam(config *TeamConfig) ([]*TeamAssignment, error) {
	if config == nil {
		config = DefaultTeamConfig()
	}

	assignments := make([]*TeamAssignment, 0)
	usedAgents := make(map[string]bool)

	for _, role := range config.RequiredRoles {
		// Get preferred domain for this role
		preferredDomain := DomainGeneral
		if d, ok := config.PreferredDomains[role]; ok {
			preferredDomain = d
		}

		// Find best agents for this role
		candidates := tb.pool.SelectTopNForRole(role, preferredDomain, config.MinAgentsPerRole+2)

		assigned := 0
		for _, candidate := range candidates {
			if !config.AllowRoleSharing && usedAgents[candidate.ID] {
				continue
			}

			score := candidate.ScoreAgent(role, preferredDomain)

			// Clone the agent's primary role for this assignment
			assignment := &TeamAssignment{
				Agent:     candidate,
				Role:      role,
				Score:     score,
				IsPrimary: assigned == 0,
			}
			assignments = append(assignments, assignment)

			usedAgents[candidate.ID] = true
			assigned++

			if assigned >= config.MinAgentsPerRole {
				break
			}
		}

		if assigned == 0 {
			return nil, fmt.Errorf("no suitable agent found for role: %s", role)
		}
	}

	return assignments, nil
}

// BuildTeamTopologyAgents builds a team and returns topology.Agent instances.
func (tb *TeamBuilder) BuildTeamTopologyAgents(config *TeamConfig) ([]*topology.Agent, error) {
	assignments, err := tb.BuildTeam(config)
	if err != nil {
		return nil, err
	}

	result := make([]*topology.Agent, 0, len(assignments))
	for _, assignment := range assignments {
		agent := assignment.Agent.ToTopologyAgent()
		agent.Role = assignment.Role // Override with assigned role
		result = append(result, agent)
	}

	return result, nil
}
