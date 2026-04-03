// Package taskweaver provides Taskweaver CLI agent integration.
// Taskweaver: Task planning AI.
package taskweaver

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Taskweaver provides Taskweaver integration
type Taskweaver struct {
	*base.BaseIntegration
}

// New creates a new Taskweaver integration
func New() *Taskweaver {
	info := agents.AgentInfo{
		Type:        agents.TypeTaskweaver,
		Name:        "Taskweaver",
		Description: "Task planning AI",
		Vendor:      "Taskweaver",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Taskweaver{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Taskweaver) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Taskweaver) IsAvailable() bool {
	_, err := exec.LookPath("taskweaver")
	return err == nil
}

var _ agents.AgentIntegration = (*Taskweaver)(nil)
