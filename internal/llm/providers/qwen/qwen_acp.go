package qwen

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"dev.helix.agent/internal/models"
)

// QwenACPProvider implements the LLMProvider interface using Qwen Code's ACP
// (Agent Communication Protocol) over stdin/stdout.
//
// This provides a more powerful integration than CLI:
// - Session management (conversation history)
// - Streaming responses
// - Proper authentication handling
// - Tool support
type QwenACPProvider struct {
	model       string
	timeout     time.Duration
	maxTokens   int
	cwd         string
	initialized bool

	// Process management
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	scanner   *bufio.Scanner
	mu        sync.Mutex
	requestID int64
	sessionID string
	isRunning bool
	startOnce sync.Once
	startErr  error

	// Response channels for async handling
	responses map[int64]chan *acpResponse
	respMu    sync.RWMutex
}

// ACP Protocol Version
const acpProtocolVersion = 1

// ACP Message Types
type acpRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type acpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *acpError       `json:"error,omitempty"`
	Method  string          `json:"method,omitempty"` // For notifications
	Params  json.RawMessage `json:"params,omitempty"` // For notifications
}

type acpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ACP Request/Response Types
type initializeRequest struct {
	ClientCapabilities clientCapabilities `json:"clientCapabilities"`
}

type clientCapabilities struct {
	FileSystem bool `json:"fileSystem"`
}

type initializeResponse struct {
	ProtocolVersion   int               `json:"protocolVersion"`
	AgentInfo         agentInfo         `json:"agentInfo"`
	AgentCapabilities agentCapabilities `json:"agentCapabilities"`
	AuthMethods       []authMethod      `json:"authMethods"`
}

type agentInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

type agentCapabilities struct {
	LoadSession        bool               `json:"loadSession"`
	PromptCapabilities promptCapabilities `json:"promptCapabilities"`
}

type promptCapabilities struct {
	Image           bool `json:"image"`
	Audio           bool `json:"audio"`
	EmbeddedContext bool `json:"embeddedContext"`
}

type authMethod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type newSessionRequest struct {
	CWD string `json:"cwd"`
}

type newSessionResponse struct {
	SessionID string `json:"sessionId"`
}

type promptRequest struct {
	SessionID string         `json:"sessionId"`
	Prompt    []contentBlock `json:"prompt"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type promptResponse struct {
	StopReason string         `json:"stopReason"`
	Result     []contentBlock `json:"result"`
}

// sessionUpdateNotification is sent during streaming
type sessionUpdateNotification struct {
	SessionID string         `json:"sessionId"`
	Updates   []contentBlock `json:"updates"`
}

// QwenACPConfig holds configuration for the ACP provider
type QwenACPConfig struct {
	Model     string
	Timeout   time.Duration
	MaxTokens int
	CWD       string
}

// DefaultQwenACPConfig returns default configuration
func DefaultQwenACPConfig() QwenACPConfig {
	return QwenACPConfig{
		Model:     "qwen-turbo",
		Timeout:   180 * time.Second,
		MaxTokens: 8192,
		CWD:       ".",
	}
}

// NewQwenACPProvider creates a new Qwen ACP provider
func NewQwenACPProvider(config QwenACPConfig) *QwenACPProvider {
	if config.Timeout == 0 {
		config.Timeout = 180 * time.Second
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 8192
	}
	if config.CWD == "" {
		config.CWD = "."
	}

	return &QwenACPProvider{
		model:     config.Model,
		timeout:   config.Timeout,
		maxTokens: config.MaxTokens,
		cwd:       config.CWD,
		responses: make(map[int64]chan *acpResponse),
	}
}

// NewQwenACPProviderWithModel creates an ACP provider with a specific model
func NewQwenACPProviderWithModel(model string) *QwenACPProvider {
	config := DefaultQwenACPConfig()
	config.Model = model
	return NewQwenACPProvider(config)
}

// Start starts the Qwen ACP process
func (p *QwenACPProvider) Start() error {
	p.startOnce.Do(func() {
		p.startErr = p.startProcess()
	})
	return p.startErr
}

func (p *QwenACPProvider) startProcess() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if qwen command exists
	qwenPath, err := exec.LookPath("qwen")
	if err != nil {
		return fmt.Errorf("qwen command not found: %w", err)
	}

	// Start qwen in ACP mode
	p.cmd = exec.Command(qwenPath, "--acp")
	p.cmd.Dir = p.cwd

	// Set up pipes
	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	p.stdout, err = p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start qwen --acp: %w", err)
	}

	p.scanner = bufio.NewScanner(p.stdout)
	p.scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer
	p.isRunning = true

	// Start response reader goroutine
	go p.readResponses()

	// Initialize the ACP connection
	if err := p.initialize(); err != nil {
		p.Stop()
		return fmt.Errorf("failed to initialize ACP: %w", err)
	}

	// Create a new session
	if err := p.createSession(); err != nil {
		p.Stop()
		return fmt.Errorf("failed to create session: %w", err)
	}

	p.initialized = true
	return nil
}

// readResponses reads responses from stdout in a goroutine
func (p *QwenACPProvider) readResponses() {
	for p.scanner.Scan() {
		line := p.scanner.Text()
		if line == "" {
			continue
		}

		var resp acpResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}

		// Handle notifications (no ID)
		if resp.Method != "" {
			p.handleNotification(&resp)
			continue
		}

		// Handle responses (has ID)
		p.respMu.RLock()
		ch, ok := p.responses[resp.ID]
		p.respMu.RUnlock()

		if ok {
			select {
			case ch <- &resp:
			default:
			}
		}
	}

	p.mu.Lock()
	p.isRunning = false
	p.mu.Unlock()
}

// handleNotification handles ACP notifications
func (p *QwenACPProvider) handleNotification(resp *acpResponse) {
	// Handle session/update notifications for streaming
	if resp.Method == "session/update" {
		// Parse notification and handle streaming updates
		// This could be used for real-time streaming in the future
	}
}

// sendRequest sends an ACP request and waits for response
func (p *QwenACPProvider) sendRequest(ctx context.Context, method string, params interface{}) (*acpResponse, error) {
	if !p.isRunning {
		return nil, fmt.Errorf("ACP process not running")
	}

	id := atomic.AddInt64(&p.requestID, 1)

	req := acpRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Create response channel
	respCh := make(chan *acpResponse, 1)
	p.respMu.Lock()
	p.responses[id] = respCh
	p.respMu.Unlock()

	defer func() {
		p.respMu.Lock()
		delete(p.responses, id)
		p.respMu.Unlock()
	}()

	// Send request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	p.mu.Lock()
	_, err = fmt.Fprintf(p.stdin, "%s\n", reqBytes)
	p.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, fmt.Errorf("ACP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// initialize sends the initialize request
func (p *QwenACPProvider) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := initializeRequest{
		ClientCapabilities: clientCapabilities{
			FileSystem: false, // We don't need filesystem access
		},
	}

	resp, err := p.sendRequest(ctx, "initialize", params)
	if err != nil {
		return err
	}

	var initResp initializeResponse
	if err := json.Unmarshal(resp.Result, &initResp); err != nil {
		return fmt.Errorf("failed to parse initialize response: %w", err)
	}

	if initResp.ProtocolVersion != acpProtocolVersion {
		return fmt.Errorf("unsupported protocol version: %d", initResp.ProtocolVersion)
	}

	return nil
}

// createSession creates a new ACP session
func (p *QwenACPProvider) createSession() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := newSessionRequest{
		CWD: p.cwd,
	}

	resp, err := p.sendRequest(ctx, "session/new", params)
	if err != nil {
		return err
	}

	var sessionResp newSessionResponse
	if err := json.Unmarshal(resp.Result, &sessionResp); err != nil {
		return fmt.Errorf("failed to parse session response: %w", err)
	}

	p.sessionID = sessionResp.SessionID
	return nil
}

// Stop stops the ACP process
func (p *QwenACPProvider) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		p.stdin.Close()
		p.cmd.Process.Kill()
		p.cmd.Wait()
	}
	p.isRunning = false
	p.initialized = false
}

// IsAvailable checks if ACP is available
func (p *QwenACPProvider) IsAvailable() bool {
	_, err := exec.LookPath("qwen")
	return err == nil
}

// Complete implements the LLMProvider interface
func (p *QwenACPProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	// Ensure ACP is started
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ACP: %w", err)
	}

	// Build prompt from messages
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

	if prompt == "" {
		return nil, fmt.Errorf("no prompt provided")
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Send prompt request
	params := promptRequest{
		SessionID: p.sessionID,
		Prompt: []contentBlock{
			{Type: "text", Text: prompt},
		},
	}

	startTime := time.Now()
	resp, err := p.sendRequest(timeoutCtx, "session/prompt", params)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("prompt request failed: %w", err)
	}

	var promptResp promptResponse
	if err := json.Unmarshal(resp.Result, &promptResp); err != nil {
		return nil, fmt.Errorf("failed to parse prompt response: %w", err)
	}

	// Extract text from result
	var content strings.Builder
	for _, block := range promptResp.Result {
		if block.Type == "text" {
			content.WriteString(block.Text)
		}
	}

	// Estimate token count
	promptTokens := len(prompt) / 4
	completionTokens := content.Len() / 4

	return &models.LLMResponse{
		ID:           fmt.Sprintf("qwen-acp-%d", time.Now().UnixNano()),
		ProviderID:   "qwen-acp",
		ProviderName: "qwen-acp",
		Content:      content.String(),
		FinishReason: promptResp.StopReason,
		TokensUsed:   promptTokens + completionTokens,
		ResponseTime: duration.Milliseconds(),
		CreatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"source":            "qwen-acp",
			"session_id":        p.sessionID,
			"model":             p.model,
			"stop_reason":       promptResp.StopReason,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"latency":           duration.String(),
		},
	}, nil
}

// CompleteStream implements streaming completion using ACP
func (p *QwenACPProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	// For now, use non-streaming and send as single chunk
	// Full streaming would require handling session/update notifications
	ch := make(chan *models.LLMResponse, 1)

	go func() {
		defer close(ch)

		resp, err := p.Complete(ctx, req)
		if err != nil {
			ch <- &models.LLMResponse{
				ProviderID:   "qwen-acp",
				ProviderName: "qwen-acp",
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			}
			return
		}
		ch <- resp
	}()

	return ch, nil
}

// HealthCheck checks if the ACP process is healthy
func (p *QwenACPProvider) HealthCheck() error {
	if !p.IsAvailable() {
		return fmt.Errorf("qwen command not available")
	}

	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start ACP: %w", err)
	}

	if !p.isRunning {
		return fmt.Errorf("ACP process not running")
	}

	return nil
}

// GetCapabilities returns the capabilities of the ACP provider
func (p *QwenACPProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels: []string{
			"qwen-plus",
			"qwen-turbo",
			"qwen-max",
			"qwen2.5-72b-instruct",
			"qwen2.5-32b-instruct",
			"qwen2.5-coder-32b-instruct",
		},
		SupportsStreaming: true,
		SupportsTools:     false, // ACP tool support could be added
		Limits: models.ModelLimits{
			MaxTokens:             p.maxTokens,
			MaxConcurrentRequests: 1, // Single session for now
		},
	}
}

// ValidateConfig validates the configuration
func (p *QwenACPProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	if !p.IsAvailable() {
		return false, []string{"qwen command not available"}
	}
	return true, nil
}

// GetName returns the provider name
func (p *QwenACPProvider) GetName() string {
	return "qwen-acp"
}

// GetProviderType returns the provider type
func (p *QwenACPProvider) GetProviderType() string {
	return "qwen"
}

// GetCurrentModel returns the current model
func (p *QwenACPProvider) GetCurrentModel() string {
	return p.model
}

// SetModel sets the model to use
func (p *QwenACPProvider) SetModel(model string) {
	p.model = model
}

// IsQwenACPAvailable checks if Qwen ACP is available
func IsQwenACPAvailable() bool {
	_, err := exec.LookPath("qwen")
	return err == nil
}

// CanUseQwenACP returns true if Qwen ACP can be used
// This checks: 1) qwen CLI installed, 2) qwen CLI authenticated
func CanUseQwenACP() bool {
	if !IsQwenCodeInstalled() {
		return false
	}
	return IsQwenCodeAuthenticated()
}
