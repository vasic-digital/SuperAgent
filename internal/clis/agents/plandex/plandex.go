// Package plandex provides Plandex CLI agent integration.
// Plandex: AI task planning.
package plandex

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Plandex provides Plandex integration
type Plandex struct {
	*base.BaseIntegration
}

// New creates a new Plandex integration
func New() *Plandex {
	info := agents.AgentInfo{
		Type:        agents.TypePlandex,
		Name:        "Plandex",
		Description: "AI task planning",
		Vendor:      "Plandex",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Plandex{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Plandex) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Plandex) IsAvailable() bool {
	_, err := exec.LookPath("plandex")
	return err == nil
}

var _ agents.AgentIntegration = (*Plandex)(nil)
