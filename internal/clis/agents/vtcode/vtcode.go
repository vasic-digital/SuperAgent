// Package vtcode provides VTCode CLI agent integration.
// VTCode: VT coding assistant.
package vtcode

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// VTCode provides VTCode integration
type VTCode struct {
	*base.BaseIntegration
}

// New creates a new VTCode integration
func New() *VTCode {
	info := agents.AgentInfo{
		Type:        agents.TypeVTCode,
		Name:        "VTCode",
		Description: "VT coding assistant",
		Vendor:      "VTCode",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &VTCode{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *VTCode) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *VTCode) IsAvailable() bool {
	_, err := exec.LookPath("vtcode")
	return err == nil
}

var _ agents.AgentIntegration = (*VTCode)(nil)
