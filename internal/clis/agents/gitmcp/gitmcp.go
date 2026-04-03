// Package gitmcp provides Git MCP CLI agent integration.
// Git MCP: Git MCP server.
package gitmcp

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// GitMCP provides Git MCP integration
type GitMCP struct {
	*base.BaseIntegration
}

// New creates a new Git MCP integration
func New() *GitMCP {
	info := agents.AgentInfo{
		Type:        agents.TypeGitMCP,
		Name:        "Git MCP",
		Description: "Git MCP server",
		Vendor:      "Git",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &GitMCP{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *GitMCP) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *GitMCP) IsAvailable() bool {
	_, err := exec.LookPath("gitmcp")
	return err == nil
}

var _ agents.AgentIntegration = (*GitMCP)(nil)
