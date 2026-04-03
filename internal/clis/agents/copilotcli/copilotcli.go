// Package copilotcli provides Copilot CLI CLI agent integration.
// Copilot CLI: GitHub Copilot CLI.
package copilotcli

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// CopilotCLI provides Copilot CLI integration
type CopilotCLI struct {
	*base.BaseIntegration
}

// New creates a new Copilot CLI integration
func New() *CopilotCLI {
	info := agents.AgentInfo{
		Type:        agents.TypeCopilotCLI,
		Name:        "Copilot CLI",
		Description: "GitHub Copilot CLI",
		Vendor:      "Copilot",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &CopilotCLI{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *CopilotCLI) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *CopilotCLI) IsAvailable() bool {
	_, err := exec.LookPath("copilotcli")
	return err == nil
}

var _ agents.AgentIntegration = (*CopilotCLI)(nil)
