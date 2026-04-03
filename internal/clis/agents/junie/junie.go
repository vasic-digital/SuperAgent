// Package junie provides Junie CLI agent integration.
// Junie: JetBrains AI.
package junie

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Junie provides Junie integration
type Junie struct {
	*base.BaseIntegration
}

// New creates a new Junie integration
func New() *Junie {
	info := agents.AgentInfo{
		Type:        agents.TypeJunie,
		Name:        "Junie",
		Description: "JetBrains AI",
		Vendor:      "Junie",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Junie{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Junie) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Junie) IsAvailable() bool {
	_, err := exec.LookPath("junie")
	return err == nil
}

var _ agents.AgentIntegration = (*Junie)(nil)
