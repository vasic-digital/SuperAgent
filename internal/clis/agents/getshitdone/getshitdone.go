// Package getshitdone provides Get Shit Done CLI agent integration.
// Get Shit Done: Productivity AI.
package getshitdone

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// GetShitDone provides Get Shit Done integration
type GetShitDone struct {
	*base.BaseIntegration
}

// New creates a new Get Shit Done integration
func New() *GetShitDone {
	info := agents.AgentInfo{
		Type:        agents.TypeGetShitDone,
		Name:        "Get Shit Done",
		Description: "Productivity AI",
		Vendor:      "Get",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &GetShitDone{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *GetShitDone) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *GetShitDone) IsAvailable() bool {
	_, err := exec.LookPath("getshitdone")
	return err == nil
}

var _ agents.AgentIntegration = (*GetShitDone)(nil)
