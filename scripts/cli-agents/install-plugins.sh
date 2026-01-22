#!/bin/bash
# ============================================================================
# HelixAgent CLI Agent Plugin Installer
# ============================================================================
# Installs HelixAgent integration plugins for all supported CLI agents
#
# Usage: ./install-plugins.sh [--agent=NAME] [--dry-run] [--uninstall]
# ============================================================================

set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Plugin directories
PLUGIN_SOURCE_DIR="$PROJECT_ROOT/plugins"
PLUGIN_TEMPLATES_DIR="$SCRIPT_DIR/plugins/templates"
PLUGIN_OUTPUT_DIR="$SCRIPT_DIR/plugins/generated"
SYSTEM_PLUGIN_DIR="${HOME}/.helix-plugins"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Parse arguments
DRY_RUN=false
UNINSTALL=false
SPECIFIC_AGENT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --uninstall)
            UNINSTALL=true
            shift
            ;;
        --agent=*)
            SPECIFIC_AGENT="${1#*=}"
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [--agent=NAME] [--dry-run] [--uninstall]"
            echo ""
            echo "Options:"
            echo "  --agent=NAME  Install plugins for specific agent only"
            echo "  --dry-run     Show what would be done without making changes"
            echo "  --uninstall   Remove installed plugins"
            echo ""
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create directories
mkdir -p "$PLUGIN_TEMPLATES_DIR"
mkdir -p "$PLUGIN_OUTPUT_DIR"
mkdir -p "$SYSTEM_PLUGIN_DIR"

# ============================================================================
# Plugin Definitions
# ============================================================================

# Core plugins that all CLI agents need
CORE_PLUGINS=(
    "helix-integration"      # Core HelixAgent API integration
    "event-handler"          # Event subscription and handling
    "verifier-client"        # LLMsVerifier integration
    "debate-ui"              # AI Debate visualization
    "streaming-adapter"      # Streaming response adapter
    "mcp-bridge"             # MCP protocol bridge
)

# Agent-specific plugin requirements
declare -A AGENT_PLUGINS
AGENT_PLUGINS["claude_code"]="helix-integration,event-handler,debate-ui,mcp-bridge"
AGENT_PLUGINS["aider"]="helix-integration,event-handler,streaming-adapter"
AGENT_PLUGINS["cline"]="helix-integration,event-handler,debate-ui,mcp-bridge"
AGENT_PLUGINS["opencode"]="helix-integration,event-handler,verifier-client,mcp-bridge"
AGENT_PLUGINS["kilo_code"]="helix-integration,event-handler,debate-ui"
AGENT_PLUGINS["gemini_cli"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["qwen_code"]="helix-integration,event-handler"
AGENT_PLUGINS["deepseek_cli"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["forge"]="helix-integration,event-handler,mcp-bridge"
AGENT_PLUGINS["codename_goose"]="helix-integration,verifier-client"
AGENT_PLUGINS["amazon_q"]="helix-integration,event-handler"
AGENT_PLUGINS["kiro"]="helix-integration,event-handler"
AGENT_PLUGINS["gpt_engineer"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["mistral_code"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["ollama_code"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["plandex"]="helix-integration,event-handler"
AGENT_PLUGINS["codex"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["vtcode"]="helix-integration,event-handler"
AGENT_PLUGINS["nanocoder"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["gitmcp"]="helix-integration,mcp-bridge"
AGENT_PLUGINS["taskweaver"]="helix-integration,event-handler"
AGENT_PLUGINS["octogen"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["fauxpilot"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["bridle"]="helix-integration,verifier-client"
AGENT_PLUGINS["agent_deck"]="helix-integration,event-handler,debate-ui"
AGENT_PLUGINS["claude_squad"]="helix-integration,event-handler,debate-ui"
AGENT_PLUGINS["codai"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["emdash"]="helix-integration,event-handler"
AGENT_PLUGINS["get_shit_done"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["github_copilot_cli"]="helix-integration,event-handler"
AGENT_PLUGINS["github_spec_kit"]="helix-integration,mcp-bridge"
AGENT_PLUGINS["gptme"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["mobile_agent"]="helix-integration,event-handler"
AGENT_PLUGINS["multiagent_coding"]="helix-integration,event-handler,debate-ui"
AGENT_PLUGINS["noi"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["openhands"]="helix-integration,event-handler,mcp-bridge"
AGENT_PLUGINS["postgres_mcp"]="helix-integration,mcp-bridge"
AGENT_PLUGINS["shai"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["snowcli"]="helix-integration,event-handler"
AGENT_PLUGINS["superset"]="helix-integration,streaming-adapter"
AGENT_PLUGINS["warp"]="helix-integration,event-handler"
AGENT_PLUGINS["cheshire_cat"]="helix-integration,event-handler,debate-ui"
AGENT_PLUGINS["conduit"]="helix-integration,verifier-client"
AGENT_PLUGINS["crush"]="helix-integration,event-handler,debate-ui"
AGENT_PLUGINS["helixcode"]="helix-integration,event-handler,verifier-client,debate-ui,mcp-bridge"

# ============================================================================
# Plugin Generation Functions
# ============================================================================

generate_helix_integration_plugin() {
    local agent_name="$1"
    local output_dir="$2"

    mkdir -p "$output_dir"

    cat > "$output_dir/helix_integration.go" << 'EOF'
// HelixAgent Integration Plugin
// Provides core API integration for CLI agents

package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

// Plugin metadata
var PluginName = "helix-integration"
var PluginVersion = "1.0.0"

// Config holds plugin configuration
type Config struct {
    HelixAgentURL string `json:"helix_agent_url"`
    APIKey        string `json:"api_key"`
    Timeout       int    `json:"timeout"`
    RetryAttempts int    `json:"retry_attempts"`
    RetryBackoff  int    `json:"retry_backoff_ms"`
}

// HelixClient provides API access to HelixAgent
type HelixClient struct {
    config     Config
    httpClient *http.Client
}

// NewHelixClient creates a new client instance
func NewHelixClient(config Config) *HelixClient {
    return &HelixClient{
        config: config,
        httpClient: &http.Client{
            Timeout: time.Duration(config.Timeout) * time.Millisecond,
        },
    }
}

// ChatCompletion sends a chat completion request
func (c *HelixClient) ChatCompletion(ctx context.Context, messages []map[string]string, model string) (string, error) {
    payload := map[string]interface{}{
        "model":    model,
        "messages": messages,
        "stream":   false,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("marshal request: %w", err)
    }

    url := c.config.HelixAgentURL + "/v1/chat/completions"
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return "", fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    if c.config.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
    }

    var lastErr error
    for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
        if attempt > 0 {
            time.Sleep(time.Duration(c.config.RetryBackoff*attempt) * time.Millisecond)
        }

        resp, err := c.httpClient.Do(req)
        if err != nil {
            lastErr = err
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            respBody, _ := io.ReadAll(resp.Body)
            lastErr = fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
            continue
        }

        var result map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return "", fmt.Errorf("decode response: %w", err)
        }

        choices, ok := result["choices"].([]interface{})
        if !ok || len(choices) == 0 {
            return "", fmt.Errorf("no choices in response")
        }

        choice := choices[0].(map[string]interface{})
        message := choice["message"].(map[string]interface{})
        content := message["content"].(string)

        return content, nil
    }

    return "", fmt.Errorf("all retry attempts failed: %v", lastErr)
}

// Init initializes the plugin
func Init() error {
    fmt.Println("[helix-integration] Plugin initialized")
    return nil
}

// Shutdown gracefully shuts down the plugin
func Shutdown() error {
    fmt.Println("[helix-integration] Plugin shutdown")
    return nil
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Printf("%s v%s\n", PluginName, PluginVersion)
        return
    }
    fmt.Println("HelixAgent Integration Plugin")
}
EOF

    cat > "$output_dir/manifest.json" << EOF
{
    "name": "helix-integration",
    "version": "1.0.0",
    "description": "Core HelixAgent API integration for ${agent_name}",
    "author": "HelixAgent Team",
    "license": "MIT",
    "agent": "${agent_name}",
    "entry_point": "helix_integration.go",
    "dependencies": [],
    "config_schema": {
        "helix_agent_url": {
            "type": "string",
            "default": "http://localhost:8080",
            "description": "HelixAgent server URL"
        },
        "api_key": {
            "type": "string",
            "default": "",
            "description": "API key for authentication"
        },
        "timeout": {
            "type": "integer",
            "default": 120000,
            "description": "Request timeout in milliseconds"
        }
    }
}
EOF

    log_success "Generated helix-integration plugin for $agent_name"
}

generate_event_handler_plugin() {
    local agent_name="$1"
    local output_dir="$2"

    mkdir -p "$output_dir"

    cat > "$output_dir/event_handler.go" << 'EOF'
// Event Handler Plugin
// Manages event subscriptions and handling for HelixAgent events

package main

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
)

var PluginName = "event-handler"
var PluginVersion = "1.0.0"

// Event represents a HelixAgent event
type Event struct {
    Type      string                 `json:"type"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

// EventHandler manages event subscriptions
type EventHandler struct {
    helixURL   string
    subscribed []string
    handlers   map[string][]func(Event)
    ctx        context.Context
    cancel     context.CancelFunc
    mu         sync.RWMutex
}

// NewEventHandler creates a new event handler
func NewEventHandler(helixURL string, subscriptions []string) *EventHandler {
    ctx, cancel := context.WithCancel(context.Background())
    return &EventHandler{
        helixURL:   helixURL,
        subscribed: subscriptions,
        handlers:   make(map[string][]func(Event)),
        ctx:        ctx,
        cancel:     cancel,
    }
}

// Subscribe registers a handler for an event type
func (h *EventHandler) Subscribe(eventType string, handler func(Event)) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.handlers[eventType] = append(h.handlers[eventType], handler)
}

// Start begins listening for events via SSE
func (h *EventHandler) Start() error {
    go h.listen()
    return nil
}

// Stop stops the event listener
func (h *EventHandler) Stop() {
    h.cancel()
}

func (h *EventHandler) listen() {
    url := h.helixURL + "/v1/events/stream?subscribe=" + strings.Join(h.subscribed, ",")

    for {
        select {
        case <-h.ctx.Done():
            return
        default:
            h.connectSSE(url)
            time.Sleep(5 * time.Second) // Reconnect delay
        }
    }
}

func (h *EventHandler) connectSSE(url string) {
    req, err := http.NewRequestWithContext(h.ctx, "GET", url, nil)
    if err != nil {
        return
    }
    req.Header.Set("Accept", "text/event-stream")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    var eventData strings.Builder

    for scanner.Scan() {
        line := scanner.Text()

        if strings.HasPrefix(line, "data:") {
            eventData.WriteString(strings.TrimPrefix(line, "data:"))
        } else if line == "" && eventData.Len() > 0 {
            var event Event
            if err := json.Unmarshal([]byte(eventData.String()), &event); err == nil {
                h.dispatch(event)
            }
            eventData.Reset()
        }
    }
}

func (h *EventHandler) dispatch(event Event) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    // Exact match handlers
    if handlers, ok := h.handlers[event.Type]; ok {
        for _, handler := range handlers {
            go handler(event)
        }
    }

    // Wildcard handlers
    for pattern, handlers := range h.handlers {
        if strings.HasSuffix(pattern, "*") {
            prefix := strings.TrimSuffix(pattern, "*")
            if strings.HasPrefix(event.Type, prefix) {
                for _, handler := range handlers {
                    go handler(event)
                }
            }
        }
    }
}

func Init() error {
    fmt.Println("[event-handler] Plugin initialized")
    return nil
}

func Shutdown() error {
    fmt.Println("[event-handler] Plugin shutdown")
    return nil
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Printf("%s v%s\n", PluginName, PluginVersion)
        return
    }
    fmt.Println("Event Handler Plugin")
}
EOF

    cat > "$output_dir/manifest.json" << EOF
{
    "name": "event-handler",
    "version": "1.0.0",
    "description": "Event subscription and handling for ${agent_name}",
    "author": "HelixAgent Team",
    "license": "MIT",
    "agent": "${agent_name}",
    "entry_point": "event_handler.go",
    "dependencies": ["helix-integration"],
    "config_schema": {
        "subscriptions": {
            "type": "array",
            "default": ["debate.*", "warning.*", "error.*"],
            "description": "Event patterns to subscribe to"
        }
    }
}
EOF

    log_success "Generated event-handler plugin for $agent_name"
}

generate_verifier_client_plugin() {
    local agent_name="$1"
    local output_dir="$2"

    mkdir -p "$output_dir"

    cat > "$output_dir/verifier_client.go" << 'EOF'
// LLMsVerifier Client Plugin
// Integrates with LLMsVerifier for provider verification

package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

var PluginName = "verifier-client"
var PluginVersion = "1.0.0"

// VerificationResult represents provider verification results
type VerificationResult struct {
    Provider      string    `json:"provider"`
    Model         string    `json:"model"`
    Score         float64   `json:"score"`
    Verified      bool      `json:"verified"`
    VerifiedAt    time.Time `json:"verified_at"`
    ResponseTime  int64     `json:"response_time_ms"`
    ErrorRate     float64   `json:"error_rate"`
    Capabilities  []string  `json:"capabilities"`
}

// VerifierClient provides access to LLMsVerifier API
type VerifierClient struct {
    baseURL    string
    httpClient *http.Client
}

// NewVerifierClient creates a new verifier client
func NewVerifierClient(baseURL string) *VerifierClient {
    return &VerifierClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// GetProviderScore retrieves the current score for a provider
func (c *VerifierClient) GetProviderScore(ctx context.Context, provider string) (*VerificationResult, error) {
    url := fmt.Sprintf("%s/v1/providers/%s/score", c.baseURL, provider)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("verifier API error: %d", resp.StatusCode)
    }

    var result VerificationResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// ListVerifiedProviders returns all verified providers
func (c *VerifierClient) ListVerifiedProviders(ctx context.Context) ([]VerificationResult, error) {
    url := c.baseURL + "/v1/providers/verified"

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var results []VerificationResult
    if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
        return nil, err
    }

    return results, nil
}

func Init() error {
    fmt.Println("[verifier-client] Plugin initialized")
    return nil
}

func Shutdown() error {
    fmt.Println("[verifier-client] Plugin shutdown")
    return nil
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Printf("%s v%s\n", PluginName, PluginVersion)
        return
    }
    fmt.Println("LLMsVerifier Client Plugin")
}
EOF

    cat > "$output_dir/manifest.json" << EOF
{
    "name": "verifier-client",
    "version": "1.0.0",
    "description": "LLMsVerifier integration for ${agent_name}",
    "author": "HelixAgent Team",
    "license": "MIT",
    "agent": "${agent_name}",
    "entry_point": "verifier_client.go",
    "dependencies": [],
    "config_schema": {
        "verifier_url": {
            "type": "string",
            "default": "http://localhost:8081",
            "description": "LLMsVerifier server URL"
        },
        "verify_on_startup": {
            "type": "boolean",
            "default": true,
            "description": "Verify providers on startup"
        }
    }
}
EOF

    log_success "Generated verifier-client plugin for $agent_name"
}

generate_debate_ui_plugin() {
    local agent_name="$1"
    local output_dir="$2"

    mkdir -p "$output_dir"

    cat > "$output_dir/debate_ui.go" << 'EOF'
// AI Debate UI Plugin
// Provides visualization for AI Debate responses

package main

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

var PluginName = "debate-ui"
var PluginVersion = "1.0.0"

// DebatePhase represents a phase in the debate
type DebatePhase struct {
    Name       string  `json:"name"`
    Icon       string  `json:"icon"`
    Status     string  `json:"status"`
    Confidence float64 `json:"confidence"`
}

// DebateResponse represents a formatted debate response
type DebateResponse struct {
    Topic           string        `json:"topic"`
    Phases          []DebatePhase `json:"phases"`
    CurrentPhase    int           `json:"current_phase"`
    FinalResponse   string        `json:"final_response"`
    OverallConfidence float64     `json:"overall_confidence"`
    Participants    []string      `json:"participants"`
}

// FormatDebateProgress formats the debate progress for display
func FormatDebateProgress(response *DebateResponse) string {
    var sb strings.Builder

    sb.WriteString("\n")
    sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
    sb.WriteString("║                    AI DEBATE IN PROGRESS                      ║\n")
    sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

    for i, phase := range response.Phases {
        status := "⏳"
        if i < response.CurrentPhase {
            status = "✅"
        } else if i == response.CurrentPhase {
            status = "🔄"
        }

        line := fmt.Sprintf("║ %s %s %-40s %s ║\n",
            status, phase.Icon, phase.Name,
            formatConfidence(phase.Confidence))
        sb.WriteString(line)
    }

    sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
    sb.WriteString(fmt.Sprintf("║ Participants: %-46s ║\n",
        strings.Join(response.Participants[:min(3, len(response.Participants))], ", ")))
    sb.WriteString(fmt.Sprintf("║ Overall Confidence: %-40s ║\n",
        formatConfidence(response.OverallConfidence)))
    sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n")

    return sb.String()
}

func formatConfidence(conf float64) string {
    bars := int(conf * 10)
    return fmt.Sprintf("[%s%s] %.0f%%",
        strings.Repeat("█", bars),
        strings.Repeat("░", 10-bars),
        conf*100)
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// DefaultPhases returns the default debate phases
func DefaultPhases() []DebatePhase {
    return []DebatePhase{
        {Name: "Initial Response", Icon: "🔍", Status: "pending", Confidence: 0},
        {Name: "Validation", Icon: "✓", Status: "pending", Confidence: 0},
        {Name: "Polish & Improve", Icon: "✨", Status: "pending", Confidence: 0},
        {Name: "Final Conclusion", Icon: "📜", Status: "pending", Confidence: 0},
    }
}

func Init() error {
    fmt.Println("[debate-ui] Plugin initialized")
    return nil
}

func Shutdown() error {
    fmt.Println("[debate-ui] Plugin shutdown")
    return nil
}

func main() {
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "--version":
            fmt.Printf("%s v%s\n", PluginName, PluginVersion)
            return
        case "--demo":
            demo := &DebateResponse{
                Topic: "AI Ethics",
                Phases: DefaultPhases(),
                CurrentPhase: 2,
                Participants: []string{"Claude", "Gemini", "DeepSeek"},
                OverallConfidence: 0.87,
            }
            demo.Phases[0].Status = "complete"
            demo.Phases[0].Confidence = 0.85
            demo.Phases[1].Status = "complete"
            demo.Phases[1].Confidence = 0.90
            fmt.Println(FormatDebateProgress(demo))
            return
        }
    }

    // Process stdin JSON
    var response DebateResponse
    if err := json.NewDecoder(os.Stdin).Decode(&response); err != nil {
        fmt.Println("Usage: debate-ui [--version|--demo] or pipe JSON")
        return
    }
    fmt.Println(FormatDebateProgress(&response))
}
EOF

    cat > "$output_dir/manifest.json" << EOF
{
    "name": "debate-ui",
    "version": "1.0.0",
    "description": "AI Debate visualization for ${agent_name}",
    "author": "HelixAgent Team",
    "license": "MIT",
    "agent": "${agent_name}",
    "entry_point": "debate_ui.go",
    "dependencies": ["helix-integration"],
    "config_schema": {
        "show_progress": {
            "type": "boolean",
            "default": true,
            "description": "Show debate progress"
        },
        "show_participants": {
            "type": "boolean",
            "default": true,
            "description": "Show participant list"
        }
    }
}
EOF

    log_success "Generated debate-ui plugin for $agent_name"
}

generate_streaming_adapter_plugin() {
    local agent_name="$1"
    local output_dir="$2"

    mkdir -p "$output_dir"

    cat > "$output_dir/streaming_adapter.go" << 'EOF'
// Streaming Adapter Plugin
// Handles streaming responses from HelixAgent

package main

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
)

var PluginName = "streaming-adapter"
var PluginVersion = "1.0.0"

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    Choices []struct {
        Index int `json:"index"`
        Delta struct {
            Content string `json:"content"`
        } `json:"delta"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
}

// StreamingClient handles streaming API calls
type StreamingClient struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

// NewStreamingClient creates a new streaming client
func NewStreamingClient(baseURL, apiKey string) *StreamingClient {
    return &StreamingClient{
        baseURL:    baseURL,
        apiKey:     apiKey,
        httpClient: &http.Client{},
    }
}

// StreamCompletion streams a chat completion
func (c *StreamingClient) StreamCompletion(ctx context.Context, messages []map[string]string, model string, onChunk func(string)) error {
    payload := map[string]interface{}{
        "model":    model,
        "messages": messages,
        "stream":   true,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    url := c.baseURL + "/v1/chat/completions"
    req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "text/event-stream")
    if c.apiKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.apiKey)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            }
            return err
        }

        line = strings.TrimSpace(line)
        if !strings.HasPrefix(line, "data:") {
            continue
        }

        data := strings.TrimPrefix(line, "data:")
        data = strings.TrimSpace(data)

        if data == "[DONE]" {
            break
        }

        var chunk StreamChunk
        if err := json.Unmarshal([]byte(data), &chunk); err != nil {
            continue
        }

        for _, choice := range chunk.Choices {
            if choice.Delta.Content != "" {
                onChunk(choice.Delta.Content)
            }
        }
    }

    return nil
}

func Init() error {
    fmt.Println("[streaming-adapter] Plugin initialized")
    return nil
}

func Shutdown() error {
    fmt.Println("[streaming-adapter] Plugin shutdown")
    return nil
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Printf("%s v%s\n", PluginName, PluginVersion)
        return
    }
    fmt.Println("Streaming Adapter Plugin")
}
EOF

    cat > "$output_dir/manifest.json" << EOF
{
    "name": "streaming-adapter",
    "version": "1.0.0",
    "description": "Streaming response adapter for ${agent_name}",
    "author": "HelixAgent Team",
    "license": "MIT",
    "agent": "${agent_name}",
    "entry_point": "streaming_adapter.go",
    "dependencies": ["helix-integration"],
    "config_schema": {
        "buffer_size": {
            "type": "integer",
            "default": 4096,
            "description": "Streaming buffer size"
        },
        "chunk_timeout": {
            "type": "integer",
            "default": 30000,
            "description": "Chunk timeout in milliseconds"
        }
    }
}
EOF

    log_success "Generated streaming-adapter plugin for $agent_name"
}

generate_mcp_bridge_plugin() {
    local agent_name="$1"
    local output_dir="$2"

    mkdir -p "$output_dir"

    cat > "$output_dir/mcp_bridge.go" << 'EOF'
// MCP Bridge Plugin
// Bridges MCP protocol between CLI agent and HelixAgent

package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

var PluginName = "mcp-bridge"
var PluginVersion = "1.0.0"

// MCPRequest represents an MCP request
type MCPRequest struct {
    JSONRPC string                 `json:"jsonrpc"`
    ID      interface{}            `json:"id"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
    JSONRPC string                 `json:"jsonrpc"`
    ID      interface{}            `json:"id"`
    Result  map[string]interface{} `json:"result,omitempty"`
    Error   *MCPError              `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// MCPBridge handles MCP protocol bridging
type MCPBridge struct {
    helixURL   string
    httpClient *http.Client
}

// NewMCPBridge creates a new MCP bridge
func NewMCPBridge(helixURL string) *MCPBridge {
    return &MCPBridge{
        helixURL: helixURL,
        httpClient: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

// Forward forwards an MCP request to HelixAgent
func (b *MCPBridge) Forward(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
    body, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    url := b.helixURL + "/v1/mcp"
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url,
        strings.NewReader(string(body)))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := b.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var mcpResp MCPResponse
    if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
        return nil, err
    }

    return &mcpResp, nil
}

// ListTools lists available MCP tools
func (b *MCPBridge) ListTools(ctx context.Context) ([]string, error) {
    req := &MCPRequest{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "tools/list",
    }

    resp, err := b.Forward(ctx, req)
    if err != nil {
        return nil, err
    }

    if resp.Error != nil {
        return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
    }

    tools, ok := resp.Result["tools"].([]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid tools response")
    }

    var names []string
    for _, t := range tools {
        if tool, ok := t.(map[string]interface{}); ok {
            if name, ok := tool["name"].(string); ok {
                names = append(names, name)
            }
        }
    }

    return names, nil
}

func Init() error {
    fmt.Println("[mcp-bridge] Plugin initialized")
    return nil
}

func Shutdown() error {
    fmt.Println("[mcp-bridge] Plugin shutdown")
    return nil
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Printf("%s v%s\n", PluginName, PluginVersion)
        return
    }
    fmt.Println("MCP Bridge Plugin")
}
EOF

    # Fix missing import
    sed -i '11a\    "strings"' "$output_dir/mcp_bridge.go"

    cat > "$output_dir/manifest.json" << EOF
{
    "name": "mcp-bridge",
    "version": "1.0.0",
    "description": "MCP protocol bridge for ${agent_name}",
    "author": "HelixAgent Team",
    "license": "MIT",
    "agent": "${agent_name}",
    "entry_point": "mcp_bridge.go",
    "dependencies": ["helix-integration"],
    "config_schema": {
        "timeout": {
            "type": "integer",
            "default": 60000,
            "description": "MCP request timeout in milliseconds"
        }
    }
}
EOF

    log_success "Generated mcp-bridge plugin for $agent_name"
}

# ============================================================================
# Plugin Installation Functions
# ============================================================================

generate_plugins_for_agent() {
    local agent_name="$1"
    local plugins_str="${AGENT_PLUGINS[$agent_name]}"

    if [[ -z "$plugins_str" ]]; then
        # Use default core plugins
        plugins_str="helix-integration,event-handler"
    fi

    local agent_output_dir="$PLUGIN_OUTPUT_DIR/$agent_name"
    mkdir -p "$agent_output_dir"

    IFS=',' read -ra plugins <<< "$plugins_str"

    for plugin in "${plugins[@]}"; do
        local plugin_dir="$agent_output_dir/$plugin"

        case "$plugin" in
            "helix-integration")
                generate_helix_integration_plugin "$agent_name" "$plugin_dir"
                ;;
            "event-handler")
                generate_event_handler_plugin "$agent_name" "$plugin_dir"
                ;;
            "verifier-client")
                generate_verifier_client_plugin "$agent_name" "$plugin_dir"
                ;;
            "debate-ui")
                generate_debate_ui_plugin "$agent_name" "$plugin_dir"
                ;;
            "streaming-adapter")
                generate_streaming_adapter_plugin "$agent_name" "$plugin_dir"
                ;;
            "mcp-bridge")
                generate_mcp_bridge_plugin "$agent_name" "$plugin_dir"
                ;;
            *)
                log_warning "Unknown plugin: $plugin"
                ;;
        esac
    done

    # Create agent plugin index
    cat > "$agent_output_dir/index.json" << EOF
{
    "agent": "${agent_name}",
    "plugins": [$(echo "$plugins_str" | sed 's/,/", "/g' | sed 's/^/"/' | sed 's/$/"/')],
    "generated_at": "$(date -Iseconds)",
    "version": "1.0.0"
}
EOF

    log_success "Generated all plugins for $agent_name"
}

install_plugins_for_agent() {
    local agent_name="$1"
    local source_dir="$PLUGIN_OUTPUT_DIR/$agent_name"
    local target_dir="$SYSTEM_PLUGIN_DIR/$agent_name"

    if [[ ! -d "$source_dir" ]]; then
        log_error "Plugins not generated for $agent_name"
        return 1
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would install plugins: $source_dir -> $target_dir"
        return 0
    fi

    mkdir -p "$target_dir"
    cp -r "$source_dir"/* "$target_dir/"

    log_success "Installed plugins for $agent_name: $target_dir"
}

uninstall_plugins_for_agent() {
    local agent_name="$1"
    local target_dir="$SYSTEM_PLUGIN_DIR/$agent_name"

    if [[ ! -d "$target_dir" ]]; then
        log_warning "No plugins installed for $agent_name"
        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would remove: $target_dir"
        return 0
    fi

    rm -rf "$target_dir"
    log_success "Uninstalled plugins for $agent_name"
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    log_info "=============================================="
    log_info "HelixAgent CLI Agent Plugin Installer"
    log_info "=============================================="

    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "Running in DRY-RUN mode"
    fi

    # Get list of agents to process
    local agents_to_process=()

    if [[ -n "$SPECIFIC_AGENT" ]]; then
        agents_to_process=("$SPECIFIC_AGENT")
    else
        agents_to_process=("${!AGENT_PLUGINS[@]}")
    fi

    local total=${#agents_to_process[@]}
    local success=0
    local failed=0

    if [[ "$UNINSTALL" == "true" ]]; then
        log_info "Uninstalling plugins..."
        for agent in "${agents_to_process[@]}"; do
            if uninstall_plugins_for_agent "$agent"; then
                success=$((success + 1))
            else
                failed=$((failed + 1))
            fi
        done
    else
        log_info "Generating plugins..."
        for agent in "${agents_to_process[@]}"; do
            if generate_plugins_for_agent "$agent"; then
                success=$((success + 1))
            else
                failed=$((failed + 1))
            fi
        done

        log_info "Installing plugins..."
        for agent in "${agents_to_process[@]}"; do
            install_plugins_for_agent "$agent"
        done
    fi

    echo ""
    log_info "=============================================="
    log_info "Summary"
    log_info "=============================================="
    log_info "Total agents: $total"
    log_success "Processed: $success"
    if [[ $failed -gt 0 ]]; then
        log_error "Failed: $failed"
    fi

    log_info "Plugin directory: $SYSTEM_PLUGIN_DIR"
    log_success "Plugin installation complete!"
}

main "$@"
