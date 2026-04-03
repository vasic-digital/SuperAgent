// Package snowcli provides Snow CLI CLI agent integration.
// Snow CLI: Snowflake CLI.
package snowcli

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// SnowCLI provides Snow CLI integration
type SnowCLI struct {
	*base.BaseIntegration
}

// New creates a new Snow CLI integration
func New() *SnowCLI {
	info := agents.AgentInfo{
		Type:        agents.TypeSnowCLI,
		Name:        "Snow CLI",
		Description: "Snowflake CLI",
		Vendor:      "Snow",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &SnowCLI{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *SnowCLI) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *SnowCLI) IsAvailable() bool {
	_, err := exec.LookPath("snowcli")
	return err == nil
}

var _ agents.AgentIntegration = (*SnowCLI)(nil)
