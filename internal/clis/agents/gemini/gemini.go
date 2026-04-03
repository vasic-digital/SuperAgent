// Package gemini provides Gemini CLI agent integration.
// Gemini: Google's AI coding assistant CLI.
package gemini

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// Gemini provides Gemini CLI integration
 type Gemini struct {
	*base.BaseIntegration
	config *Config
}

// Config holds Gemini configuration
 type Config struct {
	base.BaseConfig
	ProjectID     string
	Location      string
	ModelName     string
}

// New creates a new Gemini integration
 func New() *Gemini {
	info := agents.AgentInfo{
		Type:        agents.TypeGeminiCLI,
		Name:        "Gemini CLI",
		Description: "Google's AI coding assistant CLI",
		Vendor:      "Google",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_generation",
			"code_explanation",
			"chat",
			"vertex_ai_integration",
			"gcp_integration",
		},
		IsEnabled: true,
		Priority:  2,
	}
	
	return &Gemini{
		BaseIntegration: base.NewBaseIntegration(info),
		config: &Config{
			BaseConfig: base.BaseConfig{
				Model:     "gemini-1.5-pro",
				AutoStart: true,
			},
			Location:  "us-central1",
			ModelName: "gemini-1.5-pro",
		},
	}
}

// Initialize initializes Gemini
func (g *Gemini) Initialize(ctx context.Context, config interface{}) error {
	if err := g.BaseIntegration.Initialize(ctx, config); err != nil {
		return err
	}
	
	if cfg, ok := config.(*Config); ok {
		g.config = cfg
	}
	
	// Set up GCP credentials if available
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		g.config.ProjectID = projectID
	}
	
	return nil
}

// Execute executes a Gemini command
func (g *Gemini) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	if !g.IsStarted() {
		if err := g.Start(ctx); err != nil {
			return nil, err
		}
	}
	
	switch command {
	case "chat":
		return g.chat(ctx, params)
	case "generate":
		return g.generate(ctx, params)
	case "explain":
		return g.explain(ctx, params)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// chat sends a chat message
func (g *Gemini) chat(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	message, _ := params["message"].(string)
	if message == "" {
		return nil, fmt.Errorf("message required")
	}
	
	args := g.buildArgs()
	args = append(args, "chat", message)
	
	output, err := g.ExecuteCommand(ctx, "gcloud", args...)
	if err != nil {
		return nil, fmt.Errorf("gemini chat failed: %w\n%s", err, string(output))
	}
	
	return map[string]interface{}{
		"response": string(output),
		"success":  true,
	}, nil
}

// generate generates code
func (g *Gemini) generate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt required")
	}
	
	args := g.buildArgs()
	args = append(args, "generate", prompt)
	
	output, err := g.ExecuteCommand(ctx, "gcloud", args...)
	if err != nil {
		return nil, fmt.Errorf("gemini generate failed: %w", err)
	}
	
	return map[string]interface{}{
		"code":    string(output),
		"success": true,
	}, nil
}

// explain explains code
func (g *Gemini) explain(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	file, _ := params["file"].(string)
	if file == "" {
		return nil, fmt.Errorf("file required")
	}
	
	args := g.buildArgs()
	args = append(args, "explain", file)
	
	output, err := g.ExecuteCommand(ctx, "gcloud", args...)
	if err != nil {
		return nil, fmt.Errorf("gemini explain failed: %w", err)
	}
	
	return map[string]interface{}{
		"explanation": string(output),
		"success":     true,
	}, nil
}

// buildArgs builds command arguments
func (g *Gemini) buildArgs() []string {
	var args []string{"ai", "models"}
	
	if g.config.ProjectID != "" {
		args = append(args, "--project", g.config.ProjectID)
	}
	
	if g.config.Location != "" {
		args = append(args, "--location", g.config.Location)
	}
	
	if g.config.ModelName != "" {
		args = append(args, "--model", g.config.ModelName)
	}
	
	return args
}

// IsAvailable checks if gcloud CLI is available
func (g *Gemini) IsAvailable() bool {
	_, err := exec.LookPath("gcloud")
	return err == nil
}

var _ agents.AgentIntegration = (*Gemini)(nil)
