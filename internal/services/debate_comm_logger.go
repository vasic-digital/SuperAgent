package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DebateCommLogger provides Retrofit-like logging for debate LLM communication
// Format: [ROLE: Model Name] <--- Request / ---> Response
// Supports colored output for different CLI agents
type DebateCommLogger struct {
	logger       *logrus.Logger
	enableColors bool
	cliAgent     string // Detected CLI agent (opencode, claudecode, etc.)
}

// ANSI color codes for terminal output
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
	ColorDim     = "\033[2m"

	// Role-specific colors
	ColorAnalyst     = "\033[38;5;39m"  // Light blue
	ColorProposer    = "\033[38;5;82m"  // Light green
	ColorCritic      = "\033[38;5;208m" // Orange
	ColorSynthesizer = "\033[38;5;141m" // Light purple
	ColorMediator    = "\033[38;5;220m" // Gold

	// Communication direction colors
	ColorRequest  = "\033[38;5;45m"  // Cyan for outgoing
	ColorResponse = "\033[38;5;46m"  // Bright green for incoming
	ColorFallback = "\033[38;5;214m" // Orange for fallback
	ColorError    = "\033[38;5;196m" // Red for errors
	ColorStream   = "\033[38;5;51m"  // Aqua for streaming
)

// RoleAbbreviations maps debate roles to short codes
var RoleAbbreviations = map[string]string{
	"analyst":     "A",
	"proposer":    "P",
	"critic":      "C",
	"synthesizer": "S",
	"mediator":    "M",
	"default":     "D",
}

// RoleColors maps debate roles to their colors
var RoleColors = map[string]string{
	"analyst":     ColorAnalyst,
	"proposer":    ColorProposer,
	"critic":      ColorCritic,
	"synthesizer": ColorSynthesizer,
	"mediator":    ColorMediator,
	"default":     ColorWhite,
}

// NewDebateCommLogger creates a new debate communication logger
func NewDebateCommLogger(logger *logrus.Logger) *DebateCommLogger {
	return &DebateCommLogger{
		logger:       logger,
		enableColors: true,
		cliAgent:     "unknown",
	}
}

// SetCLIAgent sets the detected CLI agent for proper formatting
func (dcl *DebateCommLogger) SetCLIAgent(agent string) {
	dcl.cliAgent = strings.ToLower(agent)
}

// SetColorsEnabled enables or disables colored output
func (dcl *DebateCommLogger) SetColorsEnabled(enabled bool) {
	dcl.enableColors = enabled
}

// formatRoleTag creates the role tag like [A: Claude Opus 4.5]
func (dcl *DebateCommLogger) formatRoleTag(role, provider, model string) string {
	abbrev := RoleAbbreviations[strings.ToLower(role)]
	if abbrev == "" {
		abbrev = RoleAbbreviations["default"]
	}

	// Clean up model name for display
	modelDisplay := formatModelName(model)

	if dcl.enableColors {
		roleColor := RoleColors[strings.ToLower(role)]
		if roleColor == "" {
			roleColor = RoleColors["default"]
		}
		return fmt.Sprintf("%s%s[%s: %s]%s", roleColor, ColorBold, abbrev, modelDisplay, ColorReset)
	}

	return fmt.Sprintf("[%s: %s]", abbrev, modelDisplay)
}

// formatModelName creates a human-readable model name
func formatModelName(model string) string {
	// Common model name mappings for better readability
	modelMappings := map[string]string{
		"claude-opus-4-5-20251101":     "Claude Opus 4.5",
		"claude-sonnet-4-5-20250929":   "Claude Sonnet 4.5",
		"claude-haiku-4-5-20251001":    "Claude Haiku 4.5",
		"claude-opus-4-20250514":       "Claude Opus 4",
		"claude-sonnet-4-20250514":     "Claude Sonnet 4",
		"deepseek-chat":                "DeepSeek Chat",
		"deepseek-coder":               "DeepSeek Coder",
		"gemini-2.0-flash":             "Gemini 2.0 Flash",
		"gemini-1.5-pro":               "Gemini 1.5 Pro",
		"mistral-large-latest":         "Mistral Large",
		"mistral-medium-latest":        "Mistral Medium",
		"grok-code":                    "Grok Code",
		"qwen-max":                     "Qwen Max",
		"qwen-plus":                    "Qwen Plus",
		"qwen-turbo":                   "Qwen Turbo",
		"gpt-4":                        "GPT-4",
		"gpt-4-turbo":                  "GPT-4 Turbo",
		"llama-3.3-70b-versatile":      "Llama 3.3 70B",
	}

	if mapped, ok := modelMappings[model]; ok {
		return mapped
	}

	// Capitalize first letter of each word for unknown models
	return model
}

// LogRequest logs an outgoing request to an LLM provider (Retrofit-like)
// Format: [A: Claude Opus 4.5] <--- Sending request...
func (dcl *DebateCommLogger) LogRequest(role, provider, model string, promptLength int, round int) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s<---%s", ColorRequest, ColorReset)
	} else {
		arrow = "<---"
	}

	message := fmt.Sprintf("%s %s Sending request (round %d, prompt: %d chars)",
		tag, arrow, round, promptLength)

	dcl.logger.Info(message)
}

// LogStreamStart logs the start of a streaming response
// Format: [A: Claude Opus 4.5] ---> [STREAM START]
func (dcl *DebateCommLogger) LogStreamStart(role, provider, model string) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow, streamTag string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorResponse, ColorReset)
		streamTag = fmt.Sprintf("%s[STREAM START]%s", ColorStream, ColorReset)
	} else {
		arrow = "--->"
		streamTag = "[STREAM START]"
	}

	message := fmt.Sprintf("%s %s %s", tag, arrow, streamTag)
	dcl.logger.Info(message)
}

// LogStreamChunk logs a chunk of streaming response
// Format: [A: Claude Opus 4.5] ---> [CHUNK] 128 bytes
func (dcl *DebateCommLogger) LogStreamChunk(role, provider, model string, chunkSize int, totalReceived int) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow, chunkTag string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorStream, ColorReset)
		chunkTag = fmt.Sprintf("%s[CHUNK]%s", ColorDim, ColorReset)
	} else {
		arrow = "--->"
		chunkTag = "[CHUNK]"
	}

	message := fmt.Sprintf("%s %s %s +%d bytes (total: %d)",
		tag, arrow, chunkTag, chunkSize, totalReceived)
	dcl.logger.Debug(message)
}

// LogStreamEnd logs the end of streaming
// Format: [A: Claude Opus 4.5] ---> [STREAM END] 2048 bytes in 1.23s
func (dcl *DebateCommLogger) LogStreamEnd(role, provider, model string, totalBytes int, duration time.Duration) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow, endTag string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorResponse, ColorReset)
		endTag = fmt.Sprintf("%s[STREAM END]%s", ColorStream, ColorReset)
	} else {
		arrow = "--->"
		endTag = "[STREAM END]"
	}

	message := fmt.Sprintf("%s %s %s %d bytes in %.2fs",
		tag, arrow, endTag, totalBytes, duration.Seconds())
	dcl.logger.Info(message)
}

// LogResponse logs a received response from an LLM provider (Retrofit-like)
// Format: [A: Claude Opus 4.5] ---> Received 2048 bytes in 1.23s (quality: 0.85)
func (dcl *DebateCommLogger) LogResponse(role, provider, model string, contentLength int, duration time.Duration, qualityScore float64) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorResponse, ColorReset)
	} else {
		arrow = "--->"
	}

	message := fmt.Sprintf("%s %s Received %d bytes in %.2fs (quality: %.2f)",
		tag, arrow, contentLength, duration.Seconds(), qualityScore)

	dcl.logger.Info(message)
}

// LogResponsePreview logs a response with content preview
// Format: [A: Claude Opus 4.5] ---> "The analysis shows that..." (2048 bytes)
func (dcl *DebateCommLogger) LogResponsePreview(role, provider, model string, content string, maxPreviewLen int) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorResponse, ColorReset)
	} else {
		arrow = "--->"
	}

	preview := content
	if len(preview) > maxPreviewLen {
		preview = preview[:maxPreviewLen] + "..."
	}
	// Clean up newlines for single-line preview
	preview = strings.ReplaceAll(preview, "\n", " ")
	preview = strings.ReplaceAll(preview, "  ", " ")

	var previewText string
	if dcl.enableColors {
		previewText = fmt.Sprintf("%s\"%s\"%s", ColorDim, preview, ColorReset)
	} else {
		previewText = fmt.Sprintf("\"%s\"", preview)
	}

	message := fmt.Sprintf("%s %s %s (%d bytes)",
		tag, arrow, previewText, len(content))

	dcl.logger.Info(message)
}

// LogFallbackAttempt logs a fallback attempt
// Format: [A: Claude Opus 4.5] ---> [FALLBACK #1: DeepSeek Chat] Attempting...
func (dcl *DebateCommLogger) LogFallbackAttempt(originalRole, originalProvider, originalModel, fallbackProvider, fallbackModel string, fallbackIndex int) {
	originalTag := dcl.formatRoleTag(originalRole, originalProvider, originalModel)

	var arrow, fallbackTag string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorFallback, ColorReset)
		fallbackTag = fmt.Sprintf("%s%s[FALLBACK #%d: %s]%s",
			ColorFallback, ColorBold,
			fallbackIndex,
			formatModelName(fallbackModel),
			ColorReset)
	} else {
		arrow = "--->"
		fallbackTag = fmt.Sprintf("[FALLBACK #%d: %s]", fallbackIndex, formatModelName(fallbackModel))
	}

	message := fmt.Sprintf("%s %s %s Attempting fallback...",
		originalTag, arrow, fallbackTag)

	dcl.logger.Warn(message)
}

// LogFallbackSuccess logs a successful fallback
// Format: [A: Claude Opus 4.5] ---> [FALLBACK #1: DeepSeek Chat] ---> Success! (1.5s)
func (dcl *DebateCommLogger) LogFallbackSuccess(originalRole, originalProvider, originalModel, fallbackProvider, fallbackModel string, fallbackIndex int, duration time.Duration) {
	originalTag := dcl.formatRoleTag(originalRole, originalProvider, originalModel)

	var arrow, fallbackTag, successArrow, successTag string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorFallback, ColorReset)
		fallbackTag = fmt.Sprintf("%s%s[FALLBACK #%d: %s]%s",
			ColorFallback, ColorBold,
			fallbackIndex,
			formatModelName(fallbackModel),
			ColorReset)
		successArrow = fmt.Sprintf("%s--->%s", ColorGreen, ColorReset)
		successTag = fmt.Sprintf("%s%sSuccess!%s", ColorGreen, ColorBold, ColorReset)
	} else {
		arrow = "--->"
		fallbackTag = fmt.Sprintf("[FALLBACK #%d: %s]", fallbackIndex, formatModelName(fallbackModel))
		successArrow = "--->"
		successTag = "Success!"
	}

	message := fmt.Sprintf("%s %s %s %s %s (%.2fs)",
		originalTag, arrow, fallbackTag, successArrow, successTag, duration.Seconds())

	dcl.logger.Info(message)
}

// LogFallbackChain logs the complete fallback chain that was used
// Format: [A: Claude Opus 4.5] ---> [FALLBACK: DeepSeek] ---> [FALLBACK: Gemini] ---> Content
func (dcl *DebateCommLogger) LogFallbackChain(role string, chain []FallbackChainEntry, finalContent string, totalDuration time.Duration) {
	if len(chain) == 0 {
		return
	}

	var parts []string

	// First entry is the original provider
	originalTag := dcl.formatRoleTag(role, chain[0].Provider, chain[0].Model)
	parts = append(parts, originalTag)

	// Add each fallback in the chain
	for i := 1; i < len(chain); i++ {
		entry := chain[i]
		var fallbackPart string
		if dcl.enableColors {
			arrow := fmt.Sprintf("%s--->%s", ColorFallback, ColorReset)
			fallbackTag := fmt.Sprintf("%s[FALLBACK: %s]%s", ColorFallback, formatModelName(entry.Model), ColorReset)
			fallbackPart = fmt.Sprintf("%s %s", arrow, fallbackTag)
		} else {
			fallbackPart = fmt.Sprintf("---> [FALLBACK: %s]", formatModelName(entry.Model))
		}
		parts = append(parts, fallbackPart)
	}

	// Final arrow and content preview
	var finalPart string
	preview := finalContent
	if len(preview) > 50 {
		preview = preview[:50] + "..."
	}
	preview = strings.ReplaceAll(preview, "\n", " ")

	if dcl.enableColors {
		arrow := fmt.Sprintf("%s--->%s", ColorResponse, ColorReset)
		finalPart = fmt.Sprintf("%s %s\"%s\"%s (%.2fs)",
			arrow, ColorDim, preview, ColorReset, totalDuration.Seconds())
	} else {
		finalPart = fmt.Sprintf("---> \"%s\" (%.2fs)", preview, totalDuration.Seconds())
	}
	parts = append(parts, finalPart)

	message := strings.Join(parts, " ")
	dcl.logger.Info(message)
}

// LogError logs an error during LLM communication
// Format: [A: Claude Opus 4.5] ---> [ERROR] Connection timeout
func (dcl *DebateCommLogger) LogError(role, provider, model string, err error) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow, errorTag string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorError, ColorReset)
		errorTag = fmt.Sprintf("%s%s[ERROR]%s", ColorError, ColorBold, ColorReset)
	} else {
		arrow = "--->"
		errorTag = "[ERROR]"
	}

	message := fmt.Sprintf("%s %s %s %v", tag, arrow, errorTag, err)
	dcl.logger.Error(message)
}

// LogAllFallbacksExhausted logs when all fallbacks fail
// Format: [A: Claude Opus 4.5] ---> [EXHAUSTED] All 3 fallbacks failed
func (dcl *DebateCommLogger) LogAllFallbacksExhausted(role, provider, model string, fallbackCount int) {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow, exhaustedTag string
	if dcl.enableColors {
		arrow = fmt.Sprintf("%s--->%s", ColorError, ColorReset)
		exhaustedTag = fmt.Sprintf("%s%s[EXHAUSTED]%s", ColorError, ColorBold, ColorReset)
	} else {
		arrow = "--->"
		exhaustedTag = "[EXHAUSTED]"
	}

	message := fmt.Sprintf("%s %s %s All %d fallbacks failed",
		tag, arrow, exhaustedTag, fallbackCount)

	dcl.logger.Error(message)
}

// LogDebatePhase logs the start of a debate phase
// Format: ═══ DEBATE PHASE: Round 1 - Getting participant responses ═══
func (dcl *DebateCommLogger) LogDebatePhase(phase string, round int) {
	var message string
	if dcl.enableColors {
		message = fmt.Sprintf("%s═══ DEBATE PHASE: %s (Round %d) ═══%s",
			ColorCyan, phase, round, ColorReset)
	} else {
		message = fmt.Sprintf("═══ DEBATE PHASE: %s (Round %d) ═══", phase, round)
	}
	dcl.logger.Info(message)
}

// LogDebateSummary logs a summary of the debate round
func (dcl *DebateCommLogger) LogDebateSummary(round int, participantCount int, totalDuration time.Duration, avgQuality float64, fallbacksUsed int) {
	var header, footer string
	if dcl.enableColors {
		header = fmt.Sprintf("%s%s═══ ROUND %d SUMMARY ═══%s", ColorCyan, ColorBold, round, ColorReset)
		footer = fmt.Sprintf("%s═══════════════════════%s", ColorCyan, ColorReset)
	} else {
		header = fmt.Sprintf("═══ ROUND %d SUMMARY ═══", round)
		footer = "═══════════════════════"
	}

	dcl.logger.Info(header)
	dcl.logger.Infof("  Participants: %d", participantCount)
	dcl.logger.Infof("  Duration: %.2fs", totalDuration.Seconds())
	dcl.logger.Infof("  Avg Quality: %.2f", avgQuality)
	if fallbacksUsed > 0 {
		if dcl.enableColors {
			dcl.logger.Infof("  %sFallbacks Used: %d%s", ColorFallback, fallbacksUsed, ColorReset)
		} else {
			dcl.logger.Infof("  Fallbacks Used: %d", fallbacksUsed)
		}
	}
	dcl.logger.Info(footer)
}

// FallbackChainEntry represents an entry in the fallback chain for logging
type FallbackChainEntry struct {
	Provider string
	Model    string
	Success  bool
	Error    error
	Duration time.Duration
}

// CLIAgentColors returns CLI-agent-specific color configuration
// Different CLI agents may have different terminal capabilities
func CLIAgentColors(agent string) map[string]bool {
	// All 18 supported CLI agents and their color support
	colorSupport := map[string]bool{
		"opencode":      true,  // Full ANSI color support
		"claudecode":    true,  // Full ANSI color support
		"kilocode":      true,  // Full ANSI color support
		"crush":         true,  // Full ANSI color support
		"helixcode":     true,  // Full ANSI color support
		"kiro":          true,  // Full ANSI color support
		"aider":         true,  // Full ANSI color support
		"cline":         true,  // VS Code terminal support
		"codenamegoose": true,  // Full ANSI color support
		"deepseekcli":   true,  // Full ANSI color support
		"forge":         true,  // Full ANSI color support
		"geminicli":     true,  // Full ANSI color support
		"gptengineer":   true,  // Full ANSI color support
		"mistralcode":   true,  // Full ANSI color support
		"ollamacode":    true,  // Full ANSI color support
		"plandex":       true,  // Full ANSI color support
		"qwencode":      true,  // Full ANSI color support
		"amazonq":       true,  // Full ANSI color support
		"unknown":       false, // Disable colors for unknown agents
	}

	if supported, ok := colorSupport[strings.ToLower(agent)]; ok {
		return map[string]bool{"colors": supported}
	}
	return map[string]bool{"colors": true} // Default to colors enabled
}

// FormatRetrofitLog creates a complete Retrofit-style log entry
// This is the main function for generating formatted communication logs
func (dcl *DebateCommLogger) FormatRetrofitLog(
	direction string, // "request" or "response"
	role string,
	provider string,
	model string,
	content string,
	metadata map[string]interface{},
) string {
	tag := dcl.formatRoleTag(role, provider, model)

	var arrow string
	switch direction {
	case "request":
		if dcl.enableColors {
			arrow = fmt.Sprintf("%s<---%s", ColorRequest, ColorReset)
		} else {
			arrow = "<---"
		}
	case "response":
		if dcl.enableColors {
			arrow = fmt.Sprintf("%s--->%s", ColorResponse, ColorReset)
		} else {
			arrow = "--->"
		}
	case "fallback":
		if dcl.enableColors {
			arrow = fmt.Sprintf("%s--->%s", ColorFallback, ColorReset)
		} else {
			arrow = "--->"
		}
	case "error":
		if dcl.enableColors {
			arrow = fmt.Sprintf("%s--->%s", ColorError, ColorReset)
		} else {
			arrow = "--->"
		}
	}

	// Build metadata string
	var metaParts []string
	if round, ok := metadata["round"].(int); ok {
		metaParts = append(metaParts, fmt.Sprintf("round=%d", round))
	}
	if duration, ok := metadata["duration"].(time.Duration); ok {
		metaParts = append(metaParts, fmt.Sprintf("%.2fs", duration.Seconds()))
	}
	if bytes, ok := metadata["bytes"].(int); ok {
		metaParts = append(metaParts, fmt.Sprintf("%d bytes", bytes))
	}
	if quality, ok := metadata["quality"].(float64); ok {
		metaParts = append(metaParts, fmt.Sprintf("quality=%.2f", quality))
	}
	if fallbackIdx, ok := metadata["fallback_index"].(int); ok {
		metaParts = append(metaParts, fmt.Sprintf("fallback=#%d", fallbackIdx))
	}

	metaStr := ""
	if len(metaParts) > 0 {
		metaStr = fmt.Sprintf(" (%s)", strings.Join(metaParts, ", "))
	}

	// Create content preview
	preview := content
	if len(preview) > 80 {
		preview = preview[:80] + "..."
	}
	preview = strings.ReplaceAll(preview, "\n", " ")

	return fmt.Sprintf("%s %s %s%s", tag, arrow, preview, metaStr)
}
