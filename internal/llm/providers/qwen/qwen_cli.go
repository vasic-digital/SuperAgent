package qwen

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

// QwenCLIProvider implements the LLMProvider interface using Qwen Code CLI
// This is used when OAuth credentials are present but the API rejects them
// (OAuth tokens from Qwen Code are product-restricted)
type QwenCLIProvider struct {
	model           string
	cliPath         string // Path to qwen CLI binary
	cliAvailable    bool
	cliCheckOnce    sync.Once
	cliCheckErr     error
	timeout         time.Duration
	maxOutputTokens int
}

// QwenCLIConfig holds configuration for the CLI provider
type QwenCLIConfig struct {
	Model           string
	Timeout         time.Duration
	MaxOutputTokens int
}

// DefaultQwenCLIConfig returns default configuration
func DefaultQwenCLIConfig() QwenCLIConfig {
	return QwenCLIConfig{
		Model:           "qwen-plus",
		Timeout:         120 * time.Second,
		MaxOutputTokens: 4096,
	}
}

// NewQwenCLIProvider creates a new Qwen Code CLI provider
func NewQwenCLIProvider(config QwenCLIConfig) *QwenCLIProvider {
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}
	if config.MaxOutputTokens == 0 {
		config.MaxOutputTokens = 4096
	}

	return &QwenCLIProvider{
		model:           config.Model,
		timeout:         config.Timeout,
		maxOutputTokens: config.MaxOutputTokens,
	}
}

// NewQwenCLIProviderWithModel creates a CLI provider with a specific model
func NewQwenCLIProviderWithModel(model string) *QwenCLIProvider {
	config := DefaultQwenCLIConfig()
	config.Model = model
	return NewQwenCLIProvider(config)
}

// IsCLIAvailable checks if Qwen Code CLI is installed and available
func (p *QwenCLIProvider) IsCLIAvailable() bool {
	p.cliCheckOnce.Do(func() {
		// Check for qwen command in PATH
		path, err := exec.LookPath("qwen")
		if err != nil {
			p.cliCheckErr = fmt.Errorf("qwen command not found in PATH: %w", err)
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
			p.cliCheckErr = fmt.Errorf("qwen command failed: %w (output: %s)", err, string(output))
			p.cliAvailable = false
			return
		}

		// Check if user is logged in by trying to get the current session
		cmd = exec.CommandContext(ctx, path, "auth", "status")
		output, err = cmd.CombinedOutput()
		if err != nil || strings.Contains(string(output), "not logged in") {
			p.cliCheckErr = fmt.Errorf("qwen CLI not authenticated: %w (output: %s)", err, string(output))
			p.cliAvailable = false
			return
		}

		p.cliAvailable = true
	})

	return p.cliAvailable
}

// GetCLIError returns the error from CLI availability check
func (p *QwenCLIProvider) GetCLIError() error {
	p.IsCLIAvailable() // Ensure check is done
	return p.cliCheckErr
}

// Complete implements the LLMProvider interface
func (p *QwenCLIProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if !p.IsCLIAvailable() {
		return nil, fmt.Errorf("Qwen Code CLI not available: %v", p.cliCheckErr)
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

	// Build qwen command arguments
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

	// Execute qwen command
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
			return nil, fmt.Errorf("qwen CLI timed out after %v", p.timeout)
		}
		return nil, fmt.Errorf("qwen CLI failed: %w (stderr: %s)", err, stderr.String())
	}

	output := stdout.String()
	if output == "" {
		return nil, fmt.Errorf("qwen CLI returned empty response")
	}

	// Estimate token count (rough approximation: 4 chars per token)
	promptTokens := len(prompt) / 4
	completionTokens := len(output) / 4

	return &models.LLMResponse{
		Content:      output,
		ProviderName: "qwen-cli",
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

// CompleteStream implements streaming for Qwen CLI
func (p *QwenCLIProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if !p.IsCLIAvailable() {
		return nil, fmt.Errorf("Qwen Code CLI not available: %v", p.cliCheckErr)
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

	// Build qwen command arguments
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
		return nil, fmt.Errorf("failed to start qwen CLI: %w", err)
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
				ProviderName: "qwen-cli",
				FinishReason: "",
			}
		}

		// Wait for command to finish
		_ = cmd.Wait()

		// Send final response
		responseChan <- &models.LLMResponse{
			Content:      fullContent.String(),
			ProviderName: "qwen-cli",
			FinishReason: "stop",
		}
	}()

	return responseChan, nil
}

// HealthCheck checks if Qwen CLI is available and working
func (p *QwenCLIProvider) HealthCheck() error {
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

// GetCapabilities returns the capabilities of Qwen CLI provider
func (p *QwenCLIProvider) GetCapabilities() *models.ProviderCapabilities {
	// Available models through Qwen Code CLI
	supportedModels := []string{
		"qwen-plus",
		"qwen-turbo",
		"qwen-max",
		"qwen2.5-72b-instruct",
		"qwen2.5-32b-instruct",
		"qwen2.5-14b-instruct",
		"qwen2.5-7b-instruct",
		"qwen2.5-coder-32b-instruct",
		"qwen2.5-coder-14b-instruct",
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
func (p *QwenCLIProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	if !p.IsCLIAvailable() {
		return false, []string{fmt.Sprintf("Qwen Code CLI not available: %v", p.cliCheckErr)}
	}
	return true, nil
}

// GetName returns the provider name
func (p *QwenCLIProvider) GetName() string {
	return "qwen-cli"
}

// GetProviderType returns the provider type
func (p *QwenCLIProvider) GetProviderType() string {
	return "qwen"
}

// GetCurrentModel returns the current model
func (p *QwenCLIProvider) GetCurrentModel() string {
	return p.model
}

// SetModel sets the model to use
func (p *QwenCLIProvider) SetModel(model string) {
	p.model = model
}

// IsQwenCodeInstalled is a standalone function to check if Qwen Code is installed
func IsQwenCodeInstalled() bool {
	_, err := exec.LookPath("qwen")
	if err != nil {
		return false
	}
	return true
}

// GetQwenCodePath returns the path to qwen command if installed
func GetQwenCodePath() (string, error) {
	path, err := exec.LookPath("qwen")
	if err != nil {
		return "", fmt.Errorf("qwen command not found in PATH: %w", err)
	}
	return path, nil
}

// IsQwenCodeAuthenticated checks if Qwen Code is logged in
func IsQwenCodeAuthenticated() bool {
	path, err := GetQwenCodePath()
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

// CanUseQwenOAuth returns true if Qwen OAuth can be used via CLI
// This checks: 1) OAuth credentials exist, 2) qwen CLI installed, 3) qwen CLI authenticated
func CanUseQwenOAuth() bool {
	// Check if OAuth is enabled and credentials exist
	if os.Getenv("QWEN_CODE_USE_OAUTH_CREDENTIALS") != "true" {
		return false
	}

	// Check for credential file
	homeDir, _ := os.UserHomeDir()
	credPath := homeDir + "/.qwen/oauth_creds.json"
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		return false
	}

	// Check if CLI is installed and authenticated
	return IsQwenCodeInstalled() && IsQwenCodeAuthenticated()
}
