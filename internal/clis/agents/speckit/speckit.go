// Package speckit provides Spec Kit CLI agent integration.
// Spec Kit: AI specification.
package speckit

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// SpecKit provides Spec Kit integration
type SpecKit struct {
	*base.BaseIntegration
}

// New creates a new Spec Kit integration
func New() *SpecKit {
	info := agents.AgentInfo{
		Type:        agents.TypeSpecKit,
		Name:        "Spec Kit",
		Description: "AI specification",
		Vendor:      "Spec",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &SpecKit{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *SpecKit) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *SpecKit) IsAvailable() bool {
	_, err := exec.LookPath("speckit")
	return err == nil
}

var _ agents.AgentIntegration = (*SpecKit)(nil)
