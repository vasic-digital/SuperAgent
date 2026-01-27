package claude

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"dev.helix.agent/internal/models"
)

// ClaudeCLIProvider implements the LLMProvider interface using Claude Code CLI
// This is used when OAuth credentials are present but the API rejects them
// (OAuth tokens from Claude Code are product-restricted)
type ClaudeCLIProvider struct {
	model           string
	cliPath         string // Path to claude CLI binary
	cliAvailable    bool
	cliCheckOnce    sync.Once
	cliCheckErr     error
	timeout         time.Duration
	maxOutputTokens int
}

// ClaudeCLIConfig holds configuration for the CLI provider
type ClaudeCLIConfig struct {
	Model           string
	Timeout         time.Duration
	MaxOutputTokens int
}

// DefaultClaudeCLIConfig returns default configuration
func DefaultClaudeCLIConfig() ClaudeCLIConfig {
	return ClaudeCLIConfig{
		Model:           "claude-sonnet-4-20250514",
		Timeout:         120 * time.Second,
		MaxOutputTokens: 4096,
	}
}

// NewClaudeCLIProvider creates a new Claude Code CLI provider
func NewClaudeCLIProvider(config ClaudeCLIConfig) *ClaudeCLIProvider {
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}
	if config.MaxOutputTokens == 0 {
		config.MaxOutputTokens = 4096
	}

	return &ClaudeCLIProvider{
		model:           config.Model,
		timeout:         config.Timeout,
		maxOutputTokens: config.MaxOutputTokens,
	}
}

// NewClaudeCLIProviderWithModel creates a CLI provider with a specific model
func NewClaudeCLIProviderWithModel(model string) *ClaudeCLIProvider {
	config := DefaultClaudeCLIConfig()
	config.Model = model
	return NewClaudeCLIProvider(config)
}

// IsCLIAvailable checks if Claude Code CLI is installed and available
func (p *ClaudeCLIProvider) IsCLIAvailable() bool {
	p.cliCheckOnce.Do(func() {
		// Check for claude command in PATH
		path, err := exec.LookPath("claude")
		if err != nil {
			p.cliCheckErr = fmt.Errorf("claude command not found in PATH: %w", err)
			p.cliAvailable = false
			return
		}
		p.cliPath = path

		// Verify it works by checking version
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, path, "--version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			p.cliCheckErr = fmt.Errorf("claude command failed: %w (output: %s)", err, string(output))
			p.cliAvailable = false
			return
		}

		// Check if user is logged in by trying to get the current session
		cmd = exec.CommandContext(ctx, path, "auth", "status")
		output, err = cmd.CombinedOutput()
		if err != nil || strings.Contains(string(output), "not logged in") {
			p.cliCheckErr = fmt.Errorf("claude CLI not authenticated: %w (output: %s)", err, string(output))
			p.cliAvailable = false
			return
		}

		p.cliAvailable = true
	})

	return p.cliAvailable
}

// GetCLIError returns the error from CLI availability check
func (p *ClaudeCLIProvider) GetCLIError() error {
	p.IsCLIAvailable() // Ensure check is done
	return p.cliCheckErr
}

// Complete implements the LLMProvider interface
func (p *ClaudeCLIProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if !p.IsCLIAvailable() {
		return nil, fmt.Errorf("Claude Code CLI not available: %v", p.cliCheckErr)
	}

	// Build the prompt from messages
	var promptBuilder strings.Builder
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			promptBuilder.WriteString("System: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		case "user":
			promptBuilder.WriteString("Human: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		case "assistant":
			promptBuilder.WriteString("Assistant: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		}
	}

	prompt := promptBuilder.String()
	if prompt == "" && req.Prompt != "" {
		prompt = req.Prompt
	}

	if prompt == "" {
		return nil, fmt.Errorf("no prompt provided")
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Determine model to use
	model := p.model
	if req.ModelParams.Model != "" {
		model = req.ModelParams.Model
	}

	// Build claude command arguments
	args := []string{
		"-p", prompt, // Print mode - just returns the response
		"--model", model,
	}

	// Add max tokens if specified
	maxTokens := p.maxOutputTokens
	if req.ModelParams.MaxTokens > 0 {
		maxTokens = req.ModelParams.MaxTokens
	}
	args = append(args, "--max-tokens", fmt.Sprintf("%d", maxTokens))

	// Execute claude command
	cmd := exec.CommandContext(cmdCtx, p.cliPath, args...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		// Check for context cancellation
		if cmdCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("claude CLI timed out after %v", p.timeout)
		}
		return nil, fmt.Errorf("claude CLI failed: %w (stderr: %s)", err, stderr.String())
	}

	output := stdout.String()
	if output == "" {
		return nil, fmt.Errorf("claude CLI returned empty response")
	}

	// Estimate token count (rough approximation: 4 chars per token)
	promptTokens := len(prompt) / 4
	completionTokens := len(output) / 4

	return &models.LLMResponse{
		Content:      output,
		ProviderName: "claude-cli",
		FinishReason: "stop",
		TokensUsed:   promptTokens + completionTokens,
		ResponseTime: duration.Milliseconds(),
		Metadata: map[string]interface{}{
			"model":             model,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
	}, nil
}

// CompleteStream implements streaming for Claude CLI (reads output line by line)
func (p *ClaudeCLIProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if !p.IsCLIAvailable() {
		return nil, fmt.Errorf("Claude Code CLI not available: %v", p.cliCheckErr)
	}

	// Build the prompt from messages
	var promptBuilder strings.Builder
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			promptBuilder.WriteString("System: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		case "user":
			promptBuilder.WriteString("Human: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		case "assistant":
			promptBuilder.WriteString("Assistant: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		}
	}

	prompt := promptBuilder.String()
	if prompt == "" && req.Prompt != "" {
		prompt = req.Prompt
	}

	if prompt == "" {
		return nil, fmt.Errorf("no prompt provided")
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)

	// Determine model to use
	model := p.model
	if req.ModelParams.Model != "" {
		model = req.ModelParams.Model
	}

	// Build claude command arguments for streaming
	args := []string{
		"-p", prompt,
		"--model", model,
	}

	maxTokens := p.maxOutputTokens
	if req.ModelParams.MaxTokens > 0 {
		maxTokens = req.ModelParams.MaxTokens
	}
	args = append(args, "--max-tokens", fmt.Sprintf("%d", maxTokens))

	cmd := exec.CommandContext(cmdCtx, p.cliPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start claude CLI: %w", err)
	}

	// Create response channel
	responseChan := make(chan *models.LLMResponse)

	// Read output and send chunks in goroutine
	go func() {
		defer close(responseChan)
		defer cancel()

		scanner := bufio.NewScanner(stdout)
		var fullContent strings.Builder

		for scanner.Scan() {
			line := scanner.Text()
			fullContent.WriteString(line)
			fullContent.WriteString("\n")

			responseChan <- &models.LLMResponse{
				Content:      line,
				ProviderName: "claude-cli",
				FinishReason: "",
			}
		}

		// Wait for command to finish
		_ = cmd.Wait()

		// Send final response
		responseChan <- &models.LLMResponse{
			Content:      fullContent.String(),
			ProviderName: "claude-cli",
			FinishReason: "stop",
		}
	}()

	return responseChan, nil
}

// HealthCheck checks if Claude CLI is available and working
func (p *ClaudeCLIProvider) HealthCheck() error {
	if !p.IsCLIAvailable() {
		return p.cliCheckErr
	}

	// Try a minimal prompt to verify it works
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := p.Complete(ctx, &models.LLMRequest{
		Prompt: "Reply with just 'OK'",
		ModelParams: models.ModelParameters{
			MaxTokens: 10,
		},
	})
	if err != nil {
		return err
	}

	if resp.Content == "" {
		return fmt.Errorf("health check returned empty response")
	}

	return nil
}

// GetCapabilities returns the capabilities of Claude CLI provider
func (p *ClaudeCLIProvider) GetCapabilities() *models.ProviderCapabilities {
	// Available models through Claude Code CLI
	supportedModels := []string{
		"claude-opus-4-5-20251101",
		"claude-sonnet-4-5-20250929",
		"claude-haiku-4-5-20251001",
		"claude-opus-4-20250514",
		"claude-sonnet-4-20250514",
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}

	return &models.ProviderCapabilities{
		SupportedModels:   supportedModels,
		SupportsStreaming: true,
		SupportsTools:     false, // CLI doesn't support tools directly
		Limits: models.ModelLimits{
			MaxTokens:             8192,
			MaxConcurrentRequests: 1, // CLI doesn't support concurrent requests
		},
	}
}

// ValidateConfig validates the configuration
func (p *ClaudeCLIProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	if !p.IsCLIAvailable() {
		return false, []string{fmt.Sprintf("Claude Code CLI not available: %v", p.cliCheckErr)}
	}
	return true, nil
}

// GetName returns the provider name
func (p *ClaudeCLIProvider) GetName() string {
	return "claude-cli"
}

// GetProviderType returns the provider type
func (p *ClaudeCLIProvider) GetProviderType() string {
	return "claude"
}

// GetCurrentModel returns the current model
func (p *ClaudeCLIProvider) GetCurrentModel() string {
	return p.model
}

// SetModel sets the model to use
func (p *ClaudeCLIProvider) SetModel(model string) {
	p.model = model
}

// IsClaudeCodeInstalled is a standalone function to check if Claude Code is installed
func IsClaudeCodeInstalled() bool {
	_, err := exec.LookPath("claude")
	if err != nil {
		return false
	}
	return true
}

// GetClaudeCodePath returns the path to claude command if installed
func GetClaudeCodePath() (string, error) {
	path, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude command not found in PATH: %w", err)
	}
	return path, nil
}

// IsClaudeCodeAuthenticated checks if Claude Code is logged in
func IsClaudeCodeAuthenticated() bool {
	path, err := GetClaudeCodePath()
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "auth", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	// Check if output indicates logged in status
	outputStr := strings.ToLower(string(output))
	return !strings.Contains(outputStr, "not logged in") &&
		!strings.Contains(outputStr, "no session") &&
		!strings.Contains(outputStr, "unauthenticated")
}

// CanUseClaudeOAuth returns true if Claude OAuth can be used via CLI
// This checks: 1) OAuth credentials exist, 2) claude CLI installed, 3) claude CLI authenticated
func CanUseClaudeOAuth() bool {
	// Check if OAuth is enabled and credentials exist
	if os.Getenv("CLAUDE_CODE_USE_OAUTH_CREDENTIALS") != "true" {
		return false
	}

	// Check for credential file
	homeDir, _ := os.UserHomeDir()
	credPath := homeDir + "/.claude/.credentials.json"
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		return false
	}

	// Check if CLI is installed and authenticated
	return IsClaudeCodeInstalled() && IsClaudeCodeAuthenticated()
}
