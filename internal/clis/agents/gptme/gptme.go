// Package gptme provides GPTMe CLI agent integration.
// GPTMe: AI assistant.
package gptme

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// GPTMe provides GPTMe integration
type GPTMe struct {
	*base.BaseIntegration
}

// New creates a new GPTMe integration
func New() *GPTMe {
	info := agents.AgentInfo{
		Type:        agents.TypeGPTMe,
		Name:        "GPTMe",
		Description: "AI assistant",
		Vendor:      "GPTMe",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &GPTMe{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *GPTMe) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *GPTMe) IsAvailable() bool {
	_, err := exec.LookPath("gptme")
	return err == nil
}

var _ agents.AgentIntegration = (*GPTMe)(nil)
