// Package warp provides Warp CLI agent integration.
// Warp: AI terminal.
package warp

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Warp provides Warp integration
type Warp struct {
	*base.BaseIntegration
}

// New creates a new Warp integration
func New() *Warp {
	info := agents.AgentInfo{
		Type:        agents.TypeWarp,
		Name:        "Warp",
		Description: "AI terminal",
		Vendor:      "Warp",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Warp{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Warp) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Warp) IsAvailable() bool {
	_, err := exec.LookPath("warp")
	return err == nil
}

var _ agents.AgentIntegration = (*Warp)(nil)
