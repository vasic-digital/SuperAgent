// Package noi provides Noi CLI agent integration.
// Noi: AI assistant.
package noi

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Noi provides Noi integration
type Noi struct {
	*base.BaseIntegration
}

// New creates a new Noi integration
func New() *Noi {
	info := agents.AgentInfo{
		Type:        agents.TypeNoi,
		Name:        "Noi",
		Description: "AI assistant",
		Vendor:      "Noi",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Noi{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Noi) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Noi) IsAvailable() bool {
	_, err := exec.LookPath("noi")
	return err == nil
}

var _ agents.AgentIntegration = (*Noi)(nil)
