// Package deepseekcli provides DeepSeek CLI CLI agent integration.
// DeepSeek CLI: DeepSeek AI assistant.
package deepseekcli

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// DeepSeekCLI provides DeepSeek CLI integration
type DeepSeekCLI struct {
	*base.BaseIntegration
}

// New creates a new DeepSeek CLI integration
func New() *DeepSeekCLI {
	info := agents.AgentInfo{
		Type:        agents.TypeDeepSeekCLI,
		Name:        "DeepSeek CLI",
		Description: "DeepSeek AI assistant",
		Vendor:      "DeepSeek",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &DeepSeekCLI{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *DeepSeekCLI) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *DeepSeekCLI) IsAvailable() bool {
	_, err := exec.LookPath("deepseekcli")
	return err == nil
}

var _ agents.AgentIntegration = (*DeepSeekCLI)(nil)
