// Package mistralcode provides Mistral Code CLI agent integration.
// Mistral Code: Mistral AI coding.
package mistralcode

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// MistralCode provides Mistral Code integration
type MistralCode struct {
	*base.BaseIntegration
}

// New creates a new Mistral Code integration
func New() *MistralCode {
	info := agents.AgentInfo{
		Type:        agents.TypeMistralCode,
		Name:        "Mistral Code",
		Description: "Mistral AI coding",
		Vendor:      "Mistral",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &MistralCode{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *MistralCode) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *MistralCode) IsAvailable() bool {
	_, err := exec.LookPath("mistralcode")
	return err == nil
}

var _ agents.AgentIntegration = (*MistralCode)(nil)
