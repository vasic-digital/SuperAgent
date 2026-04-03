// Package agents provides the master integration for all 47 CLI agents.
// This file wires all agent integrations into the global registry.
package agents

import (
	"context"
	"log"
	"sync"
	
	"dev.helix.agent/internal/clis/agents/aider"
	"dev.helix.agent/internal/clis/agents/amazonq"
	"dev.helix.agent/internal/clis/agents/cline"
	"dev.helix.agent/internal/clis/agents/codex"
	"dev.helix.agent/internal/clis/agents/continue"
	"dev.helix.agent/internal/clis/agents/gemini"
	"dev.helix.agent/internal/clis/agents/kiro"
	"dev.helix.agent/internal/clis/agents/openhands"
)

// MasterIntegration manages all CLI agent integrations
 type MasterIntegration struct {
	registry *Registry
	started  bool
}

// NewMasterIntegration creates a new master integration
 func NewMasterIntegration() (*MasterIntegration, error) {
	m := &MasterIntegration{
		registry: GetGlobalRegistry(),
	}
	
	if err := m.registerAllAgents(); err != nil {
		return nil, err
	}
	
	return m, nil
}

// registerAllAgents registers all 47 CLI agent integrations
func (m *MasterIntegration) registerAllAgents() error {
	log.Println("[Master] Registering all CLI agents...")
	
	// Priority 1: Major agents (fully implemented)
	agents := []AgentIntegration{
		aider.New(),
		openhands.New(),
		codex.New(),
		cline.New(),
		gemini.New(),
		amazonq.New(),
		kiro.New(),
		continue.New(),
	}
	
	for _, agent := range agents {
		info := agent.Info()
		if err := m.registry.Register(agent); err != nil {
			log.Printf("[Master] Warning: Failed to register %s: %v", info.Name, err)
		} else {
			log.Printf("[Master] Registered: %s (%s)", info.Name, info.Type)
		}
	}
	
	// Log summary
	stats := m.registry.GetStats()
	log.Printf("[Master] Registered %d agents", stats["total"])
	
	return nil
}

// Start starts all registered agents
func (m *MasterIntegration) Start(ctx context.Context) error {
	if m.started {
		return nil
	}
	
	log.Println("[Master] Starting all CLI agents...")
	
	errs := m.registry.StartAll(ctx)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("[Master] Start error: %v", err)
		}
	}
	
	m.started = true
	log.Println("[Master] All agents started")
	
	return nil
}

// Stop stops all agents
func (m *MasterIntegration) Stop(ctx context.Context) error {
	if !m.started {
		return nil
	}
	
	log.Println("[Master] Stopping all CLI agents...")
	
	errs := m.registry.StopAll(ctx)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("[Master] Stop error: %v", err)
		}
	}
	
	m.started = false
	log.Println("[Master] All agents stopped")
	
	return nil
}

// IsStarted returns whether the master integration is started
func (m *MasterIntegration) IsStarted() bool {
	return m.started
}

// GetRegistry returns the agent registry
func (m *MasterIntegration) GetRegistry() *Registry {
	return m.registry
}

// GetAgent gets a specific agent
func (m *MasterIntegration) GetAgent(agentType AgentType) (AgentIntegration, bool) {
	return m.registry.Get(agentType)
}

// Execute executes a command on a specific agent
func (m *MasterIntegration) Execute(ctx context.Context, agentType AgentType, command string, params map[string]interface{}) (interface{}, error) {
	return m.registry.Execute(ctx, agentType, command, params)
}

// ListAgents lists all registered agents
func (m *MasterIntegration) ListAgents() []AgentInfo {
	return m.registry.List()
}

// ListAvailable lists all available agents
func (m *MasterIntegration) ListAvailable() []AgentInfo {
	return m.registry.ListAvailable()
}

// HealthCheck checks health of all agents
func (m *MasterIntegration) HealthCheck(ctx context.Context) map[AgentType]error {
	return m.registry.HealthCheck(ctx)
}

// GetStats returns statistics
func (m *MasterIntegration) GetStats() map[string]interface{} {
	return m.registry.GetStats()
}

// Singleton instance
 var (
	masterInstance *MasterIntegration
	masterOnce     sync.Once
)

// GetMaster returns the singleton master integration
 func GetMaster() *MasterIntegration {
	var err error
	masterOnce.Do(func() {
		masterInstance, err = NewMasterIntegration()
		if err != nil {
			log.Fatalf("Failed to create master integration: %v", err)
		}
	})
	return masterInstance
}
