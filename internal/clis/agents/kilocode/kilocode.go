// Package kilocode provides Kilo Code CLI agent integration.
// Kilo Code: Lightweight code AI.
package kilocode

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// KiloCode provides Kilo Code integration
type KiloCode struct {
	*base.BaseIntegration
}

// New creates a new Kilo Code integration
func New() *KiloCode {
	info := agents.AgentInfo{
		Type:        agents.TypeKiloCode,
		Name:        "Kilo Code",
		Description: "Lightweight code AI",
		Vendor:      "Kilo",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &KiloCode{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *KiloCode) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *KiloCode) IsAvailable() bool {
	_, err := exec.LookPath("kilocode")
	return err == nil
}

var _ agents.AgentIntegration = (*KiloCode)(nil)
