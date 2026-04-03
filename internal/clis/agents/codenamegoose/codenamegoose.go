// Package codenamegoose provides Codename Goose CLI agent integration.
// Codename Goose: AI agent framework.
package codenamegoose

import (
	"context"
	"fmt"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// CodenameGoose provides Codename Goose integration
type CodenameGoose struct {
	*base.BaseIntegration
}

// New creates a new Codename Goose integration
func New() *CodenameGoose {
	info := agents.AgentInfo{
		Type:        agents.TypeCodenameGoose,
		Name:        "Codename Goose",
		Description: "AI agent framework",
		Vendor:      "Codename",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_assistance",
		},
		IsEnabled: true,
		Priority:  5,
	}
	
	return &CodenameGoose{
		BaseIntegration: base.NewBaseIntegration(info),
	}
}

// Execute executes a command
func (a *CodenameGoose) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("command %s not implemented", command)
}

// IsAvailable checks if available
func (a *CodenameGoose) IsAvailable() bool {
	_, err := exec.LookPath("codenamegoose")
	return err == nil
}

var _ agents.AgentIntegration = (*CodenameGoose)(nil)
