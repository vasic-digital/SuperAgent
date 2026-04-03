// Package nanocoder provides Nanocoder CLI agent integration.
// Nanocoder: Lightweight code AI.
package nanocoder

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Nanocoder provides Nanocoder integration
type Nanocoder struct {
	*base.BaseIntegration
}

// New creates a new Nanocoder integration
func New() *Nanocoder {
	info := agents.AgentInfo{
		Type:        agents.TypeNanocoder,
		Name:        "Nanocoder",
		Description: "Lightweight code AI",
		Vendor:      "Nanocoder",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Nanocoder{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Nanocoder) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Nanocoder) IsAvailable() bool {
	_, err := exec.LookPath("nanocoder")
	return err == nil
}

var _ agents.AgentIntegration = (*Nanocoder)(nil)
