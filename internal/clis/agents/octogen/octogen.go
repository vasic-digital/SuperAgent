// Package octogen provides Octogen CLI agent integration.
// Octogen: AI coding platform.
package octogen

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Octogen provides Octogen integration
type Octogen struct {
	*base.BaseIntegration
}

// New creates a new Octogen integration
func New() *Octogen {
	info := agents.AgentInfo{
		Type:        agents.TypeOctogen,
		Name:        "Octogen",
		Description: "AI coding platform",
		Vendor:      "Octogen",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Octogen{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Octogen) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Octogen) IsAvailable() bool {
	_, err := exec.LookPath("octogen")
	return err == nil
}

var _ agents.AgentIntegration = (*Octogen)(nil)
