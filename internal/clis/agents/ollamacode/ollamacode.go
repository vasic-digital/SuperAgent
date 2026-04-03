// Package ollamacode provides Ollama Code CLI agent integration.
// Ollama Code: Local LLM coding.
package ollamacode

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// OllamaCode provides Ollama Code integration
type OllamaCode struct {
	*base.BaseIntegration
}

// New creates a new Ollama Code integration
func New() *OllamaCode {
	info := agents.AgentInfo{
		Type:        agents.TypeOllamaCode,
		Name:        "Ollama Code",
		Description: "Local LLM coding",
		Vendor:      "Ollama",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &OllamaCode{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *OllamaCode) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *OllamaCode) IsAvailable() bool {
	_, err := exec.LookPath("ollamacode")
	return err == nil
}

var _ agents.AgentIntegration = (*OllamaCode)(nil)
