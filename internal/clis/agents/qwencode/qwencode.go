// Package qwencode provides Qwen Code CLI agent integration.
// Qwen Code: Alibaba Qwen coding.
package qwencode

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// QwenCode provides Qwen Code integration
type QwenCode struct {
	*base.BaseIntegration
}

// New creates a new Qwen Code integration
func New() *QwenCode {
	info := agents.AgentInfo{
		Type:        agents.TypeQwenCode,
		Name:        "Qwen Code",
		Description: "Alibaba Qwen coding",
		Vendor:      "Qwen",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &QwenCode{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *QwenCode) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *QwenCode) IsAvailable() bool {
	_, err := exec.LookPath("qwencode")
	return err == nil
}

var _ agents.AgentIntegration = (*QwenCode)(nil)
