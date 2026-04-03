// Package forge provides Forge CLI agent integration.
// Forge: AI code forge.
package forge

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Forge provides Forge integration
type Forge struct {
	*base.BaseIntegration
}

// New creates a new Forge integration
func New() *Forge {
	info := agents.AgentInfo{
		Type:        agents.TypeForge,
		Name:        "Forge",
		Description: "AI code forge",
		Vendor:      "Forge",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Forge{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Forge) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Forge) IsAvailable() bool {
	_, err := exec.LookPath("forge")
	return err == nil
}

var _ agents.AgentIntegration = (*Forge)(nil)
