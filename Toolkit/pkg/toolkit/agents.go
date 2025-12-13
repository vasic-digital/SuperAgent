// Package toolkit provides a registry for managing coding agents.
package toolkit

import (
	"fmt"
	"sync"
)

// AgentRegistry manages the registration and retrieval of coding agents.
type AgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]Agent
}

// NewAgentRegistry creates a new AgentRegistry.
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]Agent),
	}
}

// Register registers an agent with the given name.
func (r *AgentRegistry) Register(name string, agent Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[name]; exists {
		return fmt.Errorf("agent %s already registered", name)
	}

	r.agents[name] = agent
	return nil
}

// Unregister removes an agent from the registry.
func (r *AgentRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, name)
}

// Get retrieves an agent by name.
func (r *AgentRegistry) Get(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, ok := r.agents[name]
	return agent, ok
}

// List returns a list of all registered agent names.
func (r *AgentRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

// GetAll returns all registered agents.
func (r *AgentRegistry) GetAll() map[string]Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agents := make(map[string]Agent)
	for name, agent := range r.agents {
		agents[name] = agent
	}
	return agents
}

// AgentFactory defines a function type for creating agents.
type AgentFactory func(config map[string]interface{}) (Agent, error)

// AgentFactoryRegistry manages agent factories.
type AgentFactoryRegistry struct {
	mu        sync.RWMutex
	factories map[string]AgentFactory
}

// NewAgentFactoryRegistry creates a new AgentFactoryRegistry.
func NewAgentFactoryRegistry() *AgentFactoryRegistry {
	return &AgentFactoryRegistry{
		factories: make(map[string]AgentFactory),
	}
}

// Register registers an agent factory.
func (r *AgentFactoryRegistry) Register(name string, factory AgentFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("factory for agent %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Create creates an agent using the registered factory.
func (r *AgentFactoryRegistry) Create(name string, config map[string]interface{}) (Agent, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no factory registered for agent %s", name)
	}

	return factory(config)
}

// List returns a list of all registered factory names.
func (r *AgentFactoryRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}
