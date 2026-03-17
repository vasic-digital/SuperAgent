package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/modelsdev"
	"dev.helix.agent/internal/utils"
)

// findGeminiBinary locates the gemini binary by first trying exec.LookPath,
// then checking common npm global install locations when PATH lookup fails.
func findGeminiBinary() (string, error) {
	// First try standard PATH lookup
	path, err := exec.LookPath("gemini")
	if err == nil {
		return path, nil
	}

	// Get home directory for fallback paths
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return "", fmt.Errorf(
			"gemini not found in PATH and cannot determine home directory: %w",
			err,
		)
	}

	// Check common npm global install locations
	candidatePaths := []string{
		filepath.Join(homeDir, ".npm-global", "bin", "gemini"),
		filepath.Join(homeDir, ".local", "bin", "gemini"),
		"/usr/local/bin/gemini",
		filepath.Join(homeDir, "node_modules", ".bin", "gemini"),
	}

	for _, candidate := range candidatePaths {
		if info, statErr := os.Stat(candidate); statErr == nil &&
			!info.IsDir() {
			return candidate, nil
		}
	}

	// Check nvm versions directory (glob for any node version)
	nvmPattern := filepath.Join(
		homeDir, ".nvm", "versions", "node", "*", "bin", "gemini",
	)
	matches, globErr := filepath.Glob(nvmPattern)
	if globErr == nil && len(matches) > 0 {
		// Return the last match (typically the newest version)
		return matches[len(matches)-1], nil
	}

	return "", fmt.Errorf(
		"gemini command not found in PATH or common install locations: %w",
		err,
	)
}

// GeminiCLIProvider implements the LLMProvider interface using Gemini CLI
// from Google. This is used when Gemini CLI is installed and authenticated.
//
// Gemini CLI Features:
// - JSON output format for structured responses
// - Session management with --resume
// - Multiple model support (gemini-2.5-pro, gemini-2.5-flash, etc.)
// - Streaming via --output-format stream-json
// - Non-interactive approval via --approval-mode auto_edit
type GeminiCLIProvider struct {
	model           string
	cliPath         string
	cliAvailable    bool
	cliCheckOnce    sync.Once
	cliCheckErr     error
	timeout         time.Duration
	maxOutputTokens int
	sessionID       string
	apiKey          string

	// Dynamic model discovery
	availableModels     []string
	modelsDiscovered    bool
	modelsDiscoveryOnce sync.Once
}

// geminiJSONResponse represents the JSON output from Gemini CLI
type geminiJSONResponse struct {
	Content string                 `json:"content"`
	Model   string                 `json:"model"`
	Usage   geminiUsageResponse    `json:"usage"`
	Error   string                 `json:"error,omitempty"`
}

// geminiUsageResponse represents the usage field from Gemini CLI JSON output
type geminiUsageResponse struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// geminiStreamEvent represents a single JSONL event from Gemini CLI streaming
type geminiStreamEvent struct {
	Type      string              `json:"type"`
	SessionID string              `json:"sessionId,omitempty"`
	Content   string              `json:"content,omitempty"`
	Usage     *geminiUsageResponse `json:"usage,omitempty"`
	Error     string              `json:"error,omitempty"`
}

// Known Gemini CLI models
var knownGeminiCLIModels = []string{
	"gemini-3.1-pro-preview",
	"gemini-3-pro-preview",
	"gemini-3-flash-preview",
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
	"gemini-2.0-flash",
}

// GeminiCLIConfig holds configuration for the Gemini CLI provider
type GeminiCLIConfig struct {
	Model           string
	Timeout         time.Duration
	MaxOutputTokens int
	APIKey          string
}

// DefaultGeminiCLIConfig returns default configuration
func DefaultGeminiCLIConfig() GeminiCLIConfig {
	apiKey := os.Getenv("GEMINI_API_KEY")
	return GeminiCLIConfig{
		Model:           "",
		Timeout:         180 * time.Second,
		MaxOutputTokens: 8192,
		APIKey:          apiKey,
	}
}

// NewGeminiCLIProvider creates a new Gemini CLI provider
func NewGeminiCLIProvider(config GeminiCLIConfig) *GeminiCLIProvider {
	if config.Timeout == 0 {
		config.Timeout = 180 * time.Second
	}
	if config.MaxOutputTokens == 0 {
		config.MaxOutputTokens = 8192
	}
	if config.APIKey == "" {
		config.APIKey = os.Getenv("GEMINI_API_KEY")
	}

	p := &GeminiCLIProvider{
		model:           config.Model,
		timeout:         config.Timeout,
		maxOutputTokens: config.MaxOutputTokens,
		apiKey:          config.APIKey,
	}

	if p.model == "" {
		p.model = p.GetBestAvailableModel()
	}

	return p
}

// NewGeminiCLIProviderWithModel creates a CLI provider with a specific model
func NewGeminiCLIProviderWithModel(model string) *GeminiCLIProvider {
	config := DefaultGeminiCLIConfig()
	config.Model = model
	return NewGeminiCLIProvider(config)
}

// IsCLIAvailable checks if Gemini CLI is installed and available
func (p *GeminiCLIProvider) IsCLIAvailable() bool {
	p.cliCheckOnce.Do(func() {
		path, err := findGeminiBinary()
		if err != nil {
			p.cliCheckErr = fmt.Errorf(
				"gemini command not found: %w", err,
			)
			p.cliAvailable = false
			return
		}
		p.cliPath = path

		ctx, cancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cancel()

		cmd := exec.CommandContext(ctx, path, "--version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			p.cliCheckErr = fmt.Errorf(
				"gemini command failed: %w (output: %s)",
				err, string(output),
			)
			p.cliAvailable = false
			return
		}

		if !IsGeminiCLIAuthenticated() && p.apiKey == "" {
			p.cliCheckErr = fmt.Errorf(
				"gemini CLI not authenticated and no GEMINI_API_KEY provided",
			)
			p.cliAvailable = false
			return
		}

		p.cliAvailable = true
	})

	return p.cliAvailable
}

// GetCLIError returns the error from CLI availability check
func (p *GeminiCLIProvider) GetCLIError() error {
	p.IsCLIAvailable()
	return p.cliCheckErr
}

// Complete implements the LLMProvider interface
func (p *GeminiCLIProvider) Complete(
	ctx context.Context,
	req *models.LLMRequest,
) (*models.LLMResponse, error) {
	if !p.IsCLIAvailable() {
		return nil, fmt.Errorf(
			"Gemini CLI not available: %v", p.cliCheckErr,
		)
	}

	prompt := buildPromptFromMessages(req)
	if prompt == "" {
		return nil, fmt.Errorf("no prompt provided")
	}

	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	model := p.model
	if req.ModelParams.Model != "" {
		model = req.ModelParams.Model
	}
	if !utils.ValidateCommandArg(model) {
		return nil, fmt.Errorf("model name contains invalid characters")
	}

	args := []string{
		"-p", prompt,
		"--output-format", "json",
		"--approval-mode", "auto_edit",
	}

	if model != "" {
		args = append(args, "-m", model)
	}

	if p.sessionID != "" {
		if utils.ValidateCommandArg(p.sessionID) {
			args = append(args, "--resume", p.sessionID)
		}
	}

	cmd := exec.CommandContext(cmdCtx, p.cliPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf(
				"gemini CLI timed out after %v", p.timeout,
			)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			stdoutStr := strings.TrimSpace(stdout.String())
			stderrStr := strings.TrimSpace(stderr.String())
			var errorDetail strings.Builder
			errorDetail.WriteString(
				fmt.Sprintf("exit code %d", exitErr.ExitCode()),
			)
			if stderrStr != "" {
				errorDetail.WriteString(
					fmt.Sprintf(", stderr: %s", stderrStr),
				)
			}
			if stdoutStr != "" {
				errorDetail.WriteString(
					fmt.Sprintf(", stdout: %s", stdoutStr),
				)
			}
			return nil, fmt.Errorf(
				"gemini CLI failed: %s", errorDetail.String(),
			)
		}
		return nil, fmt.Errorf(
			"gemini CLI failed: %w (output: %s)", err, stderr.String(),
		)
	}

	rawOutput := stdout.String()
	if rawOutput == "" {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf(
				"gemini CLI returned empty response with stderr: %s",
				stderrStr,
			)
		}
		return nil, fmt.Errorf("gemini CLI returned empty response")
	}

	content, sessionID, promptTokens, completionTokens :=
		p.parseJSONResponse(rawOutput, prompt)

	if sessionID != "" {
		if utils.ValidateCommandArg(sessionID) {
			p.sessionID = sessionID
		}
	}

	return &models.LLMResponse{
		ID:           fmt.Sprintf("gemini-cli-%s", sessionID),
		ProviderID:   "gemini-cli",
		ProviderName: "gemini-cli",
		Content:      content,
		FinishReason: "stop",
		TokensUsed:   promptTokens + completionTokens,
		ResponseTime: duration.Milliseconds(),
		CreatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"model":             model,
			"session_id":        sessionID,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"source":            "gemini-cli",
		},
	}, nil
}

// parseJSONResponse extracts content from Gemini CLI JSON output
func (p *GeminiCLIProvider) parseJSONResponse(
	rawOutput string,
	prompt string,
) (string, string, int, int) {
	rawOutput = strings.TrimSpace(rawOutput)

	var jsonResp geminiJSONResponse
	if err := json.Unmarshal([]byte(rawOutput), &jsonResp); err == nil {
		if jsonResp.Content != "" {
			promptTokens := jsonResp.Usage.PromptTokenCount
			completionTokens := jsonResp.Usage.CandidatesTokenCount

			if promptTokens == 0 {
				promptTokens = len(prompt) / 4
			}
			if completionTokens == 0 {
				completionTokens = len(jsonResp.Content) / 4
			}

			return jsonResp.Content, "", promptTokens, completionTokens
		}
	}

	// Fallback: treat raw output as plain text
	promptTokens := len(prompt) / 4
	completionTokens := len(rawOutput) / 4
	return rawOutput, "", promptTokens, completionTokens
}

// CompleteStream implements streaming for Gemini CLI
func (p *GeminiCLIProvider) CompleteStream(
	ctx context.Context,
	req *models.LLMRequest,
) (<-chan *models.LLMResponse, error) {
	if !p.IsCLIAvailable() {
		return nil, fmt.Errorf(
			"Gemini CLI not available: %v", p.cliCheckErr,
		)
	}

	prompt := buildPromptFromMessages(req)
	if prompt == "" {
		return nil, fmt.Errorf("no prompt provided")
	}

	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)

	model := p.model
	if req.ModelParams.Model != "" {
		model = req.ModelParams.Model
	}
	if !utils.ValidateCommandArg(model) {
		cancel()
		return nil, fmt.Errorf("model name contains invalid characters")
	}

	args := []string{
		"-p", prompt,
		"--output-format", "stream-json",
		"--approval-mode", "auto_edit",
	}

	if model != "" {
		args = append(args, "-m", model)
	}

	if p.sessionID != "" {
		if utils.ValidateCommandArg(p.sessionID) {
			args = append(args, "--resume", p.sessionID)
		}
	}

	cmd := exec.CommandContext(cmdCtx, p.cliPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start gemini CLI: %w", err)
	}

	responseChan := make(chan *models.LLMResponse)

	go func() {
		defer close(responseChan)
		defer cancel()

		scanner := bufio.NewScanner(stdout)
		var fullContent strings.Builder

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var event geminiStreamEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				// Non-JSON line; emit as raw content
				fullContent.WriteString(line)
				fullContent.WriteString("\n")
				responseChan <- &models.LLMResponse{
					Content:      line,
					ProviderName: "gemini-cli",
					FinishReason: "",
				}
				continue
			}

			switch event.Type {
			case "init":
				if event.SessionID != "" {
					if utils.ValidateCommandArg(event.SessionID) {
						p.sessionID = event.SessionID
					}
				}
			case "message":
				fullContent.WriteString(event.Content)
				responseChan <- &models.LLMResponse{
					Content:      event.Content,
					ProviderName: "gemini-cli",
					FinishReason: "",
				}
			case "tool_use", "tool_result":
				// Pass through tool events as metadata
				responseChan <- &models.LLMResponse{
					Content:      event.Content,
					ProviderName: "gemini-cli",
					FinishReason: "",
					Metadata: map[string]interface{}{
						"event_type": event.Type,
					},
				}
			case "error":
				responseChan <- &models.LLMResponse{
					Content:      event.Error,
					ProviderName: "gemini-cli",
					FinishReason: "error",
					Metadata: map[string]interface{}{
						"error": event.Error,
					},
				}
			case "result":
				finalContent := event.Content
				if finalContent == "" {
					finalContent = fullContent.String()
				}

				var tokensUsed int
				if event.Usage != nil {
					tokensUsed = event.Usage.TotalTokenCount
				}
				if tokensUsed == 0 {
					tokensUsed = len(finalContent) / 4
				}

				responseChan <- &models.LLMResponse{
					Content:      finalContent,
					ProviderName: "gemini-cli",
					FinishReason: "stop",
					TokensUsed:   tokensUsed,
				}
			}
		}

		_ = cmd.Wait()

		// If no result event was received, send final aggregated response
		finalContent := fullContent.String()
		if finalContent != "" {
			responseChan <- &models.LLMResponse{
				Content:      finalContent,
				ProviderName: "gemini-cli",
				FinishReason: "stop",
			}
		}
	}()

	return responseChan, nil
}

// buildPromptFromMessages constructs a prompt string from request messages
func buildPromptFromMessages(req *models.LLMRequest) string {
	var promptBuilder strings.Builder
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			promptBuilder.WriteString("System: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		case "user":
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n")
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

	return prompt
}

// HealthCheck checks if Gemini CLI is available and working
func (p *GeminiCLIProvider) HealthCheck() error {
	if !p.IsCLIAvailable() {
		return p.cliCheckErr
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 30*time.Second,
	)
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

// GetCapabilities returns the capabilities of Gemini CLI provider
func (p *GeminiCLIProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:   p.GetAvailableModels(),
		SupportsStreaming: true,
		SupportsTools:     false,
		Limits: models.ModelLimits{
			MaxTokens:             8192,
			MaxConcurrentRequests: 1,
		},
	}
}

// ValidateConfig validates the configuration
func (p *GeminiCLIProvider) ValidateConfig(
	config map[string]interface{},
) (bool, []string) {
	if !p.IsCLIAvailable() {
		return false, []string{
			fmt.Sprintf("Gemini CLI not available: %v", p.cliCheckErr),
		}
	}
	return true, nil
}

// GetName returns the provider name
func (p *GeminiCLIProvider) GetName() string {
	return "gemini-cli"
}

// GetProviderType returns the provider type
func (p *GeminiCLIProvider) GetProviderType() string {
	return "gemini"
}

// GetCurrentModel returns the current model
func (p *GeminiCLIProvider) GetCurrentModel() string {
	return p.model
}

// SetModel sets the model to use
func (p *GeminiCLIProvider) SetModel(model string) {
	p.model = model
}

// IsGeminiCLIInstalled checks if Gemini CLI is installed
func IsGeminiCLIInstalled() bool {
	_, err := findGeminiBinary()
	return err == nil
}

// IsGeminiCLIAuthenticated checks if Gemini CLI is authenticated
// by inspecting the ~/.gemini/ directory for credential files
func IsGeminiCLIAuthenticated() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	geminiDir := filepath.Join(homeDir, ".gemini")
	info, err := os.Stat(geminiDir)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check for known credential files in ~/.gemini/
	credFiles := []string{
		"credentials.json",
		"settings.json",
		"auth.json",
	}
	for _, credFile := range credFiles {
		credPath := filepath.Join(geminiDir, credFile)
		fInfo, fErr := os.Stat(credPath)
		if fErr == nil && !fInfo.IsDir() && fInfo.Size() > 0 {
			return true
		}
	}

	return false
}

// GetGeminiCLIPath returns the path to gemini command if installed
func GetGeminiCLIPath() (string, error) {
	path, err := findGeminiBinary()
	if err != nil {
		return "", fmt.Errorf(
			"gemini command not found: %w", err,
		)
	}
	return path, nil
}

// CanUseGeminiCLI returns true if Gemini CLI can be used
func CanUseGeminiCLI() bool {
	if os.Getenv("GEMINI_API_KEY") != "" {
		return IsGeminiCLIInstalled()
	}
	return IsGeminiCLIInstalled() && IsGeminiCLIAuthenticated()
}

// DiscoverModels attempts to discover available models using 3-tier
// discovery: CLI help, models.dev, hardcoded fallback
func (p *GeminiCLIProvider) DiscoverModels() []string {
	p.modelsDiscoveryOnce.Do(func() {
		ctx, cancel := context.WithTimeout(
			context.Background(), 30*time.Second,
		)
		defer cancel()

		// Tier 1: CLI help output
		if p.IsCLIAvailable() {
			cliModels := p.discoverModelsFromCLI(ctx)
			if len(cliModels) > 0 {
				p.availableModels = cliModels
				p.modelsDiscovered = true
				return
			}
		}

		// Tier 2: models.dev API
		modelsDevModels := p.discoverModelsFromModelsDev(ctx)
		if len(modelsDevModels) > 0 {
			p.availableModels = modelsDevModels
			p.modelsDiscovered = true
			return
		}

		// Tier 3: hardcoded fallback
		p.availableModels = knownGeminiCLIModels
	})

	return p.availableModels
}

// discoverModelsFromCLI tries to get models from Gemini CLI help output
func (p *GeminiCLIProvider) discoverModelsFromCLI(
	ctx context.Context,
) []string {
	commands := [][]string{
		{"--help"},
	}

	for _, args := range commands {
		cmd := exec.CommandContext(ctx, p.cliPath, args...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			discovered := parseGeminiCLIModelsOutput(string(output))
			if len(discovered) > 0 {
				return discovered
			}
		}
	}

	return nil
}

// discoverModelsFromModelsDev fetches models from models.dev API
func (p *GeminiCLIProvider) discoverModelsFromModelsDev(
	ctx context.Context,
) []string {
	client := modelsdev.NewClient(nil)

	opts := &modelsdev.ListModelsOptions{
		Limit: 100,
	}

	resp, err := client.ListProviderModels(ctx, "google", opts)
	if err != nil || resp == nil {
		return nil
	}

	var allModels []string
	for _, m := range resp.Models {
		if m.ID != "" {
			allModels = append(allModels, m.ID)
		}
	}

	return allModels
}

// parseGeminiCLIModelsOutput parses CLI output to extract model names
func parseGeminiCLIModelsOutput(output string) []string {
	var discovered []string
	lines := strings.Split(output, "\n")

	modelPatterns := []string{
		"gemini-3", "gemini-2.5", "gemini-2.0", "gemini-1.5",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range modelPatterns {
			if strings.Contains(strings.ToLower(line), pattern) {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					modelName := parts[0]
					modelName = strings.Trim(modelName, ".,:-*")
					if len(modelName) > 2 {
						discovered = append(discovered, modelName)
					}
				}
				break
			}
		}
	}

	return discovered
}

// GetAvailableModels returns the list of available models
func (p *GeminiCLIProvider) GetAvailableModels() []string {
	return p.DiscoverModels()
}

// IsModelAvailable checks if a specific model is available
func (p *GeminiCLIProvider) IsModelAvailable(model string) bool {
	available := p.GetAvailableModels()
	for _, m := range available {
		if m == model {
			return true
		}
	}
	return false
}

// GetBestAvailableModel returns the best available model based on priority
func (p *GeminiCLIProvider) GetBestAvailableModel() string {
	available := p.GetAvailableModels()

	modelPriority := []string{
		"gemini-2.5-pro",
		"gemini-2.5-flash",
		"gemini-3-pro-preview",
		"gemini-3.1-pro-preview",
		"gemini-3-flash-preview",
		"gemini-2.5-flash-lite",
		"gemini-2.0-flash",
	}

	for _, preferred := range modelPriority {
		for _, model := range available {
			if strings.Contains(
				strings.ToLower(model),
				strings.ToLower(preferred),
			) {
				return model
			}
		}
	}

	if len(available) > 0 {
		return available[0]
	}
	return "gemini-2.5-pro"
}

// GetKnownGeminiCLIModels returns the list of known Gemini CLI models
func GetKnownGeminiCLIModels() []string {
	return knownGeminiCLIModels
}

// DiscoverGeminiCLIModels is a standalone function to discover models
func DiscoverGeminiCLIModels() ([]string, error) {
	if !IsGeminiCLIInstalled() {
		return knownGeminiCLIModels, fmt.Errorf(
			"gemini CLI not installed, returning known models",
		)
	}

	path, err := GetGeminiCLIPath()
	if err != nil {
		return knownGeminiCLIModels, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 15*time.Second,
	)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--help")
	output, err := cmd.CombinedOutput()
	if err == nil {
		discovered := parseGeminiCLIModelsOutput(string(output))
		if len(discovered) > 0 {
			return discovered, nil
		}
	}

	return knownGeminiCLIModels, fmt.Errorf(
		"could not discover models from CLI, returning known models",
	)
}
