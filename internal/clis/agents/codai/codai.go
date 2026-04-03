// Package codai provides Codai CLI agent integration.
// Codai: AI coding assistant.
package codai

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Codai provides Codai integration
type Codai struct {
	*base.BaseIntegration
}

// New creates a new Codai integration
func New() *Codai {
	info := agents.AgentInfo{
		Type:        agents.TypeCodai,
		Name:        "Codai",
		Description: "AI coding assistant",
		Vendor:      "Codai",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Codai{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Codai) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Codai) IsAvailable() bool {
	_, err := exec.LookPath("codai")
	return err == nil
}

var _ agents.AgentIntegration = (*Codai)(nil)
