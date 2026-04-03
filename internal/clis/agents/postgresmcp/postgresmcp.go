// Package postgresmcp provides Postgres MCP CLI agent integration.
// Postgres MCP: PostgreSQL MCP.
package postgresmcp

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// PostgresMCP provides Postgres MCP integration
type PostgresMCP struct {
	*base.BaseIntegration
}

// New creates a new Postgres MCP integration
func New() *PostgresMCP {
	info := agents.AgentInfo{
		Type:        agents.TypePostgresMCP,
		Name:        "Postgres MCP",
		Description: "PostgreSQL MCP",
		Vendor:      "Postgres",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &PostgresMCP{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *PostgresMCP) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *PostgresMCP) IsAvailable() bool {
	_, err := exec.LookPath("postgresmcp")
	return err == nil
}

var _ agents.AgentIntegration = (*PostgresMCP)(nil)
