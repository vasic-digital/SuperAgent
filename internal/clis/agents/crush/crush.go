// Package crush provides Crush CLI agent integration.
// Crush: AI code crusher.
package crush

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Crush provides Crush integration
type Crush struct {
	*base.BaseIntegration
}

// New creates a new Crush integration
func New() *Crush {
	info := agents.AgentInfo{
		Type:        agents.TypeCrush,
		Name:        "Crush",
		Description: "AI code crusher",
		Vendor:      "Crush",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Crush{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Crush) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Crush) IsAvailable() bool {
	_, err := exec.LookPath("crush")
	return err == nil
}

var _ agents.AgentIntegration = (*Crush)(nil)
