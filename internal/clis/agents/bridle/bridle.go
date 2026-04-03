// Package bridle provides Bridle CLI agent integration.
// Bridle: AI agent framework.
package bridle

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Bridle provides Bridle integration
type Bridle struct {
	*base.BaseIntegration
}

// New creates a new Bridle integration
func New() *Bridle {
	info := agents.AgentInfo{
		Type:        agents.TypeBridle,
		Name:        "Bridle",
		Description: "AI agent framework",
		Vendor:      "Bridle",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &Bridle{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *Bridle) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *Bridle) IsAvailable() bool {
	_, err := exec.LookPath("bridle")
	return err == nil
}

var _ agents.AgentIntegration = (*Bridle)(nil)
