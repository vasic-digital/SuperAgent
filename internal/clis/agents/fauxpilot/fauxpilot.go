// Package fauxpilot provides FauxPilot CLI agent integration.
// FauxPilot: Open source Copilot alternative.
package fauxpilot

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// FauxPilot provides FauxPilot integration
type FauxPilot struct {
	*base.BaseIntegration
}

// New creates a new FauxPilot integration
func New() *FauxPilot {
	info := agents.AgentInfo{
		Type:        agents.TypeFauxPilot,
		Name:        "FauxPilot",
		Description: "Open source Copilot alternative",
		Vendor:      "FauxPilot",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &FauxPilot{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *FauxPilot) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *FauxPilot) IsAvailable() bool {
	_, err := exec.LookPath("fauxpilot")
	return err == nil
}

var _ agents.AgentIntegration = (*FauxPilot)(nil)
