// Package crush provides a Crush agent implementation.
package crush

import (
	"context"
	"fmt"
	"log"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// Agent implements the Agent interface for Crush.
type Agent struct {
	config toolkit.AgentConfig
	client toolkit.Provider
}

// NewAgent creates a new Crush agent.
func NewAgent(config map[string]interface{}) (toolkit.Agent, error) {
	builder := toolkit.NewDefaultAgentConfigBuilder()
	cfg, err := builder.Build(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	agentConfig, ok := cfg.(*toolkit.AgentConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}

	// Get provider from registry
	registry := toolkit.NewProviderRegistry()
	provider, exists := registry.Get(agentConfig.Provider)
	if !exists {
		return nil, fmt.Errorf("provider %s not found in registry", agentConfig.Provider)
	}

	return &Agent{
		config: *agentConfig,
		client: provider,
	}, nil
}

// Name returns the name of the agent.
func (a *Agent) Name() string {
	return "crush"
}

// Execute executes a task with the given configuration.
func (a *Agent) Execute(ctx context.Context, task string, config interface{}) (string, error) {
	log.Printf("Crush Agent: Executing task: %s", task)

	// Create a chat request for the task
	chatReq := toolkit.ChatRequest{
		Model: a.config.Model,
		Messages: []toolkit.ChatMessage{
			{
				Role:    "system",
				Content: "You are Crush, an advanced AI assistant specialized in complex problem-solving and analysis.",
			},
			{
				Role:    "user",
				Content: task,
			},
		},
		MaxTokens:   a.config.MaxTokens,
		Temperature: a.config.Temperature,
	}

	// Execute the chat completion
	response, err := a.client.Chat(ctx, chatReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute chat completion: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// ValidateConfig validates the agent configuration.
func (a *Agent) ValidateConfig(config interface{}) error {
	agentConfig, ok := config.(*toolkit.AgentConfig)
	if !ok {
		return fmt.Errorf("invalid config type")
	}

	return agentConfig.Validate()
}

// Capabilities returns the capabilities of the Crush agent.
func (a *Agent) Capabilities() []string {
	return []string{
		"problem_solving",
		"analysis",
		"research",
		"decision_making",
		"strategy_planning",
		"complex_reasoning",
		"data_analysis",
	}
}

// Factory function for creating Crush agents.
func Factory(config map[string]interface{}) (toolkit.Agent, error) {
	return NewAgent(config)
}

// Register registers the Crush agent with the registry.
func Register(registry *toolkit.AgentFactoryRegistry) error {
	return registry.Register("crush", Factory)
}
