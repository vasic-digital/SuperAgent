// Package opencodecli provides Opencode CLI CLI agent integration.
// Opencode CLI: Opencode AI.
package opencodecli

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// OpencodeCLI provides Opencode CLI integration
type OpencodeCLI struct {
	*base.BaseIntegration
}

// New creates a new Opencode CLI integration
func New() *OpencodeCLI {
	info := agents.AgentInfo{
		Type:        agents.TypeOpencodeCLI,
		Name:        "Opencode CLI",
		Description: "Opencode AI",
		Vendor:      "Opencode",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &OpencodeCLI{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *OpencodeCLI) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *OpencodeCLI) IsAvailable() bool {
	_, err := exec.LookPath("opencodecli")
	return err == nil
}

var _ agents.AgentIntegration = (*OpencodeCLI)(nil)
