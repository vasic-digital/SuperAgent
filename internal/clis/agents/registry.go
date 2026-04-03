// Package agents provides a unified registry for all CLI agent integrations.
package agents

import (
	"context"
	"fmt"
	"sync"
)

// AgentType represents the type of CLI agent
 type AgentType string

// All 47 CLI agent types
const (
	TypeAider          AgentType = "aider"
	TypeClaudeCode     AgentType = "claude_code"
	TypeCodex          AgentType = "codex"
	TypeOpenHands      AgentType = "openhands"
	TypeCline          AgentType = "cline"
	TypeGeminiCLI      AgentType = "gemini_cli"
	TypeAmazonQ        AgentType = "amazon_q"
	TypeKiro           AgentType = "kiro"
	TypeContinue       AgentType = "continue"
	TypeAgentDeck      AgentType = "agent_deck"
	TypeBridle         AgentType = "bridle"
	TypeClaudePlugins  AgentType = "claude_plugins"
	TypeClaudeSquad    AgentType = "claude_squad"
	TypeCodai          AgentType = "codai"
	TypeCodenameGoose  AgentType = "codename_goose"
	TypeCodexSkills    AgentType = "codex_skills"
	TypeConduit        AgentType = "conduit"
	TypeCopilotCLI     AgentType = "copilot_cli"
	TypeCrush          AgentType = "crush"
	TypeDeepseekCLI    AgentType = "deepseek_cli"
	TypeFauxpilot      AgentType = "fauxpilot"
	TypeForge          AgentType = "forge"
	TypeGetShitDone    AgentType = "get_shit_done"
	TypeGitMCP         AgentType = "git_mcp"
	TypeGptEngineer    AgentType = "gpt_engineer"
	TypeGptme          AgentType = "gptme"
	TypeJunie          AgentType = "junie"
	TypeKiloCode       AgentType = "kilo_code"
	TypeMistralCode    AgentType = "mistral_code"
	TypeMobileAgent    AgentType = "mobile_agent"
	TypeMultiagentCoding AgentType = "multiagent_coding"
	TypeNanocoder      AgentType = "nanocoder"
	TypeNoi            AgentType = "noi"
	TypeOctogen        AgentType = "octogen"
	TypeOllamaCode     AgentType = "ollama_code"
	TypeOpencodeCLI    AgentType = "opencode_cli"
	TypePlandex        AgentType = "plandex"
	TypePostgresMCP    AgentType = "postgres_mcp"
	TypeQwenCode       AgentType = "qwen_code"
	TypeShai           AgentType = "shai"
	TypeSnowCLI        AgentType = "snow_cli"
	TypeSpecKit        AgentType = "spec_kit"
	TypeSuperset       AgentType = "superset"
	TypeTaskweaver     AgentType = "taskweaver"
	TypeUIUXProMax     AgentType = "ui_ux_pro_max"
	TypeVtcode         AgentType = "vtcode"
	TypeWarp           AgentType = "warp"
)

// AllAgentTypes returns all supported agent types
 func AllAgentTypes() []AgentType {
	return []AgentType{
		TypeAider, TypeClaudeCode, TypeCodex, TypeOpenHands, TypeCline,
		TypeGeminiCLI, TypeAmazonQ, TypeKiro, TypeContinue, TypeAgentDeck,
		TypeBridle, TypeClaudePlugins, TypeClaudeSquad, TypeCodai,
		TypeCodenameGoose, TypeCodexSkills, TypeConduit, TypeCopilotCLI,
		TypeCrush, TypeDeepseekCLI, TypeFauxpilot, TypeForge, TypeGetShitDone,
		TypeGitMCP, TypeGptEngineer, TypeGptme, TypeJunie, TypeKiloCode,
		TypeMistralCode, TypeMobileAgent, TypeMultiagentCoding, TypeNanocoder,
		TypeNoi, TypeOctogen, TypeOllamaCode, TypeOpencodeCLI, TypePlandex,
		TypePostgresMCP, TypeQwenCode, TypeShai, TypeSnowCLI, TypeSpecKit,
		TypeSuperset, TypeTaskweaver, TypeUIUXProMax, TypeVtcode, TypeWarp,
	}
}

// AgentInfo holds information about an agent
 type AgentInfo struct {
	Type        AgentType
	Name        string
	Description string
	Vendor      string
	Version     string
	Capabilities []string
	IsEnabled   bool
	Priority    int
}

// AgentIntegration defines the interface for CLI agent integrations
 type AgentIntegration interface {
	Info() AgentInfo
	Initialize(ctx context.Context, config interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error)
	Health(ctx context.Context) error
	IsAvailable() bool
}

// Registry manages all agent integrations
 type Registry struct {
	mu       sync.RWMutex
	agents   map[AgentType]AgentIntegration
	started  map[AgentType]bool
}

// NewRegistry creates a new agent registry
 func NewRegistry() *Registry {
	return &Registry{
		agents:  make(map[AgentType]AgentIntegration),
		started: make(map[AgentType]bool),
	}
}

// Register registers an agent integration
 func (r *Registry) Register(agent AgentIntegration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	info := agent.Info()
	if _, exists := r.agents[info.Type]; exists {
		return fmt.Errorf("agent %s already registered", info.Type)
	}
	
	r.agents[info.Type] = agent
	return nil
}

// Get retrieves an agent integration
 func (r *Registry) Get(agentType AgentType) (AgentIntegration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agent, ok := r.agents[agentType]
	return agent, ok
}

// List returns all registered agents
 func (r *Registry) List() []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var infos []AgentInfo
	for _, agent := range r.agents {
		infos = append(infos, agent.Info())
	}
	
	return infos
}

// Global registry instance
 var (
	globalRegistry *Registry
	once           sync.Once
)

// GetGlobalRegistry returns the global registry instance
 func GetGlobalRegistry() *Registry {
	once.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}
