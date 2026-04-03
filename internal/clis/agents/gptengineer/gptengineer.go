// Package gptengineer provides GPT Engineer CLI agent integration.
// GPT Engineer: AI engineer.
package gptengineer

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// GPTEngineer provides GPT Engineer integration
type GPTEngineer struct {
	*base.BaseIntegration
}

// New creates a new GPT Engineer integration
func New() *GPTEngineer {
	info := agents.AgentInfo{
		Type:        agents.TypeGPTEngineer,
		Name:        "GPT Engineer",
		Description: "AI engineer",
		Vendor:      "GPT",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &GPTEngineer{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *GPTEngineer) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *GPTEngineer) IsAvailable() bool {
	_, err := exec.LookPath("gptengineer")
	return err == nil
}

var _ agents.AgentIntegration = (*GPTEngineer)(nil)
