// Package amazonq provides Amazon Q CLI agent integration.
// Amazon Q: AWS AI coding assistant and developer tool.
package amazonq

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"dev.helix.agent/internal/clis/agents"
	"dev.helix.agent/internal/clis/agents/base"
)

// AmazonQ provides Amazon Q CLI integration
 type AmazonQ struct {
	*base.BaseIntegration
	config *Config
}

// Config holds Amazon Q configuration
 type Config struct {
	base.BaseConfig
	Profile       string
	Region        string
	Context       []string
}

// New creates a new Amazon Q integration
 func New() *AmazonQ {
	info := agents.AgentInfo{
		Type:        agents.TypeAmazonQ,
		Name:        "Amazon Q",
		Description: "AWS AI coding assistant and developer tool",
		Vendor:      "Amazon Web Services",
		Version:     "1.0.0",
		Capabilities: []string{
			"code_generation",
			"code_explanation",
			"aws_integration",
			"cloudformation_support",
			"lambda_development",
			"security_scanning",
			"chat",
		},
		IsEnabled: true,
		Priority:  2,
	}
	
	return &AmazonQ{
		BaseIntegration: base.NewBaseIntegration(info),
		config: &Config{
			BaseConfig: base.BaseConfig{
				AutoStart: true,
			},
			Profile: "default",
			Region:  "us-east-1",
		},
	}
}

// Initialize initializes Amazon Q
func (a *AmazonQ) Initialize(ctx context.Context, config interface{}) error {
	if err := a.BaseIntegration.Initialize(ctx, config); err != nil {
		return err
	}
	
	if cfg, ok := config.(*Config); ok {
		a.config = cfg
	}
	
	// Set AWS credentials if available
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		a.config.Profile = profile
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		a.config.Region = region
	}
	
	return nil
}

// Execute executes an Amazon Q command
func (a *AmazonQ) Execute(ctx context.Context, command string, params map[string]interface{}) (interface{}, error) {
	if !a.IsStarted() {
		if err := a.Start(ctx); err != nil {
			return nil, err
		}
	}
	
	switch command {
	case "chat":
		return a.chat(ctx, params)
	case "transform":
		return a.transform(ctx, params)
	case "review":
		return a.review(ctx, params)
	case "test":
		return a.test(ctx, params)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// chat sends a chat message
func (a *AmazonQ) chat(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	message, _ := params["message"].(string)
	if message == "" {
		return nil, fmt.Errorf("message required")
	}
	
	args := a.buildArgs()
	args = append(args, "chat", message)
	
	output, err := a.ExecuteCommand(ctx, "q", args...)
	if err != nil {
		return nil, fmt.Errorf("amazon q chat failed: %w\n%s", err, string(output))
	}
	
	return map[string]interface{}{
		"response": string(output),
		"success":  true,
	}, nil
}

// transform transforms code
func (a *AmazonQ) transform(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt required")
	}
	
	args := a.buildArgs()
	args = append(args, "transform", prompt)
	
	output, err := a.ExecuteCommand(ctx, "q", args...)
	if err != nil {
		return nil, fmt.Errorf("amazon q transform failed: %w", err)
	}
	
	return map[string]interface{}{
		"result":  string(output),
		"success": true,
	}, nil
}

// review reviews code
func (a *AmazonQ) review(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	file, _ := params["file"].(string)
	if file == "" {
		return nil, fmt.Errorf("file required")
	}
	
	args := a.buildArgs()
	args = append(args, "review", file)
	
	output, err := a.ExecuteCommand(ctx, "q", args...)
	if err != nil {
		return nil, fmt.Errorf("amazon q review failed: %w", err)
	}
	
	return map[string]interface{}{
		"review":  string(output),
		"success": true,
	}, nil
}

// test generates tests
func (a *AmazonQ) test(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	file, _ := params["file"].(string)
	if file == "" {
		return nil, fmt.Errorf("file required")
	}
	
	args := a.buildArgs()
	args = append(args, "test", file)
	
	output, err := a.ExecuteCommand(ctx, "q", args...)
	if err != nil {
		return nil, fmt.Errorf("amazon q test failed: %w", err)
	}
	
	return map[string]interface{}{
		"tests":   string(output),
		"success": true,
	}, nil
}

// buildArgs builds command arguments
func (a *AmazonQ) buildArgs() []string {
	var args []string
	
	if a.config.Profile != "" {
		args = append(args, "--profile", a.config.Profile)
	}
	
	if a.config.Region != "" {
		args = append(args, "--region", a.config.Region)
	}
	
	return args
}

// IsAvailable checks if Amazon Q CLI is available
func (a *AmazonQ) IsAvailable() bool {
	_, err := exec.LookPath("q")
	return err == nil
}

var _ agents.AgentIntegration = (*AmazonQ)(nil)
