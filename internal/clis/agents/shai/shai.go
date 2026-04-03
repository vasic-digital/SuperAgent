// Package shai provides Shai CLI agent integration.
// Shai: AI shell assistant.
package shai

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Shai provides Shai integration
type Shai struct {
	*base.BaseIntegration
}

// New creates a new Shai integration
func New() *Shai {
	info := agents.AgentInfo{
		Type:        agents.TypeShai,
		Name:        "Shai",
		Description: "AI shell assistant",
		Vendor:      "Shai",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Shai{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Shai) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Shai) IsAvailable() bool {
	_, err := exec.LookPath("shai")
	return err == nil
}

var _ agents.AgentIntegration = (*Shai)(nil)
