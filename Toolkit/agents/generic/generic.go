// Package generic provides a generic agent implementation that can work with any config format.
package generic

import (
	"context"
	"fmt"
	"log"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// Agent implements the Agent interface for a generic agent.
type Agent struct {
	config toolkit.AgentConfig
	client toolkit.Provider
}

// NewAgent creates a new generic agent.
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
	return "generic"
}

// Execute executes a task with the given configuration.
func (a *Agent) Execute(ctx context.Context, task string, config interface{}) (string, error) {
	log.Printf("Generic Agent: Executing task: %s", task)

	// Create a chat request for the task
	systemPrompt := "You are a helpful AI assistant. Execute the following task to the best of your ability."

	chatReq := toolkit.ChatRequest{
		Model: a.config.Model,
		Messages: []toolkit.ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
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

// Capabilities returns the capabilities of the generic agent.
func (a *Agent) Capabilities() []string {
	return []string{
		"general_assistance",
		"conversation",
		"information_retrieval",
		"task_execution",
		"analysis",
		"creative_writing",
		"problem_solving",
	}
}

// Factory function for creating generic agents.
func Factory(config map[string]interface{}) (toolkit.Agent, error) {
	return NewAgent(config)
}

// Register registers the generic agent with the registry.
func Register(registry *toolkit.AgentFactoryRegistry) error {
	return registry.Register("generic", Factory)
}
