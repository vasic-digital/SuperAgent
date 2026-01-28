package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"dev.helix.agent/internal/services"
)

// ============================================================================
// Output Format Types
// ============================================================================

// OutputFormat specifies the output format for debate responses
type OutputFormat string

const (
	// OutputFormatANSI uses ANSI escape codes (for terminal output)
	OutputFormatANSI OutputFormat = "ansi"
	// OutputFormatMarkdown uses clean Markdown (for API clients)
	OutputFormatMarkdown OutputFormat = "markdown"
	// OutputFormatPlain uses plain text without formatting
	OutputFormatPlain OutputFormat = "plain"
)

// ============================================================================
// Markdown Formatting Functions (Clean, No ANSI)
// ============================================================================

// FormatDebateTeamIntroductionMarkdown formats the debate team introduction in clean Markdown
func FormatDebateTeamIntroductionMarkdown(topic string, members []*services.DebateTeamMember) string {
	var sb strings.Builder

	// Header
	sb.WriteString("\n")
	sb.WriteString("# HelixAgent AI Debate Ensemble\n\n")
	sb.WriteString("> Five AI minds deliberate to synthesize the optimal response.\n\n")

	// Topic
	topicDisplay := topic
	if len(topicDisplay) > 70 {
		topicDisplay = topicDisplay[:70] + "..."
	}
	sb.WriteString(fmt.Sprintf("**Topic:** %s\n\n", topicDisplay))

	// Debate Team
	sb.WriteString("---\n\n")
	sb.WriteString("## Debate Team\n\n")
	sb.WriteString("| Role | Model | Provider |\n")
	sb.WriteString("|------|-------|----------|\n")

	for _, member := range members {
		if member == nil {
			continue
		}
		roleName := getRoleName(member.Role)
		oauthTag := ""
		if member.IsOAuth {
			oauthTag = " [OAuth]"
		}
		sb.WriteString(fmt.Sprintf("| **%s** | %s | %s%s |\n",
			roleName, member.ModelName, member.ProviderName, oauthTag))

		// Show all fallbacks using Fallbacks slice (preferred) or legacy Fallback chain
		if len(member.Fallbacks) > 0 {
			for i, fb := range member.Fallbacks {
				fbOAuthTag := ""
				if fb.IsOAuth {
					fbOAuthTag = " [OAuth]"
				}
				sb.WriteString(fmt.Sprintf("| └─ Fallback %d | %s | %s%s |\n",
					i+1, fb.ModelName, fb.ProviderName, fbOAuthTag))
			}
		} else if member.Fallback != nil {
			// Legacy single fallback support
			sb.WriteString(fmt.Sprintf("| └─ Fallback | %s | %s |\n",
				member.Fallback.ModelName, member.Fallback.ProviderName))
		}
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("## The Deliberation\n\n")

	return sb.String()
}

// FormatPhaseHeaderMarkdown formats a phase header in clean Markdown
func FormatPhaseHeaderMarkdown(phase services.ValidationPhase, phaseNum int) string {
	icon := "▸"
	phaseName := string(phase)

	switch phase {
	case services.PhaseInitialResponse:
		icon = "🔍"
		phaseName = "INITIAL RESPONSE"
	case services.PhaseValidation:
		icon = "✓"
		phaseName = "VALIDATION"
	case services.PhasePolishImprove:
		icon = "✨"
		phaseName = "POLISH & IMPROVE"
	case services.PhaseFinalConclusion:
		icon = "📜"
		phaseName = "FINAL CONCLUSION"
	}

	return fmt.Sprintf("\n### %s Phase %d: %s\n\n", icon, phaseNum, phaseName)
}

// FormatPhaseContentMarkdown formats debate phase content in Markdown (as a quote block)
func FormatPhaseContentMarkdown(content string) string {
	// Wrap each line in a quote block for visual separation
	lines := strings.Split(content, "\n")
	var sb strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			sb.WriteString("> " + line + "\n")
		} else {
			sb.WriteString(">\n")
		}
	}
	return sb.String()
}

// FormatFinalResponseMarkdown formats the final consensus response in Markdown
func FormatFinalResponseMarkdown(content string) string {
	return fmt.Sprintf("\n---\n\n## Final Answer\n\n%s\n", content)
}

// FormatConsensusHeaderMarkdown formats the consensus section header in Markdown
func FormatConsensusHeaderMarkdown(confidenceScore float64) string {
	return fmt.Sprintf("\n---\n\n## Consensus (Confidence: %.1f%%)\n\n", confidenceScore*100)
}

// FormatRequestIndicatorMarkdown formats a request indicator in Markdown
func FormatRequestIndicatorMarkdown(role services.DebateRole, provider, model string) string {
	roleName := getRoleName(role)
	return fmt.Sprintf("**[%s]** _Requesting %s (%s)..._\n\n", roleName, model, provider)
}

// FormatResponseIndicatorMarkdown formats a response indicator in Markdown
func FormatResponseIndicatorMarkdown(role services.DebateRole, provider, model string, duration time.Duration) string {
	roleName := getRoleName(role)
	return fmt.Sprintf("**[%s]** Response from %s (%s) in %s\n\n",
		roleName, model, provider, formatDuration(duration))
}

// FormatFallbackIndicatorMarkdown formats a fallback indicator in Markdown
func FormatFallbackIndicatorMarkdown(fromProvider, fromModel, toProvider, toModel, reason string) string {
	return fmt.Sprintf("⚠️ **Fallback:** %s/%s → %s/%s (%s)\n\n",
		fromProvider, fromModel, toProvider, toModel, reason)
}

// FormatFallbackTriggeredMarkdown formats a detailed fallback triggered indicator
// Includes exact error cause and category icon for CLI agent plugins
func FormatFallbackTriggeredMarkdown(role, primaryProvider, primaryModel, fallbackProvider, fallbackModel, errorMsg, errorCategory string, duration time.Duration) string {
	var sb strings.Builder

	// Error category icon
	categoryIcon := getCategoryIcon(errorCategory)

	sb.WriteString(fmt.Sprintf("\n⚡ **[%s] Fallback Triggered**\n", role))
	sb.WriteString(fmt.Sprintf("   Primary: %s/%s (%s)\n", primaryProvider, primaryModel, formatDuration(duration)))
	sb.WriteString(fmt.Sprintf("   %s **Error:** %s\n", categoryIcon, errorMsg))
	sb.WriteString(fmt.Sprintf("   → Trying: %s/%s\n\n", fallbackProvider, fallbackModel))

	return sb.String()
}

// FormatFallbackSuccessMarkdown formats a fallback success indicator
func FormatFallbackSuccessMarkdown(role, fallbackProvider, fallbackModel string, attemptNum int, duration time.Duration) string {
	return fmt.Sprintf("🔄 **[%s] Fallback Succeeded** - %s/%s (attempt %d, %s)\n\n",
		role, fallbackProvider, fallbackModel, attemptNum, formatDuration(duration))
}

// FormatFallbackFailedMarkdown formats a fallback failed indicator
func FormatFallbackFailedMarkdown(role, fallbackProvider, fallbackModel, errorMsg, errorCategory string, attemptNum int, duration time.Duration) string {
	categoryIcon := getCategoryIcon(errorCategory)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("⛔ **[%s] Fallback %d Failed** - %s/%s (%s)\n", role, attemptNum, fallbackProvider, fallbackModel, formatDuration(duration)))
	sb.WriteString(fmt.Sprintf("   %s **Error:** %s\n\n", categoryIcon, errorMsg))

	return sb.String()
}

// FormatFallbackExhaustedMarkdown formats an all-fallbacks-exhausted indicator
func FormatFallbackExhaustedMarkdown(role string, totalAttempts int) string {
	return fmt.Sprintf("💀 **[%s] ALL FALLBACKS EXHAUSTED** - %d attempts failed, no response available\n\n", role, totalAttempts)
}

// FormatFallbackChainMarkdown formats a complete fallback chain in Markdown
func FormatFallbackChainMarkdown(position services.DebateTeamPosition, chain []FallbackAttempt) string {
	if len(chain) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n🔗 **Fallback Chain for Position %d:**\n", position))

	for i, attempt := range chain {
		status := "❌"
		if attempt.Success {
			status = "✅"
		}
		sb.WriteString(fmt.Sprintf("   %d. %s %s/%s (%s)\n",
			i+1, status, attempt.Provider, attempt.Model, formatDuration(attempt.Duration)))
		if attempt.Error != "" {
			categoryIcon := getCategoryIcon(categorizeErrorString(attempt.Error))
			sb.WriteString(fmt.Sprintf("      %s %s\n", categoryIcon, attempt.Error))
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

// getCategoryIcon returns the appropriate icon for an error category
func getCategoryIcon(category string) string {
	switch category {
	case "rate_limit":
		return "🚦"
	case "timeout":
		return "⏱️"
	case "auth":
		return "🔑"
	case "quota":
		return "📊"
	case "connection":
		return "🔌"
	case "unavailable":
		return "🚫"
	case "overloaded":
		return "🔥"
	case "invalid_request":
		return "⚠️"
	case "empty_response":
		return "📭"
	default:
		return "❓"
	}
}

// categorizeErrorString categorizes an error string into a category
func categorizeErrorString(errorMsg string) string {
	if errorMsg == "" {
		return "unknown"
	}

	lowerErr := strings.ToLower(errorMsg)

	switch {
	case strings.Contains(lowerErr, "rate limit") || strings.Contains(lowerErr, "ratelimit"):
		return "rate_limit"
	case strings.Contains(lowerErr, "timeout") || strings.Contains(lowerErr, "timed out"):
		return "timeout"
	case strings.Contains(lowerErr, "auth") || strings.Contains(lowerErr, "unauthorized") ||
		strings.Contains(lowerErr, "invalid api key") || strings.Contains(lowerErr, "401"):
		return "auth"
	case strings.Contains(lowerErr, "quota") || strings.Contains(lowerErr, "exceeded"):
		return "quota"
	case strings.Contains(lowerErr, "connection") || strings.Contains(lowerErr, "network") ||
		strings.Contains(lowerErr, "dial") || strings.Contains(lowerErr, "refused"):
		return "connection"
	case strings.Contains(lowerErr, "unavailable") || strings.Contains(lowerErr, "503"):
		return "unavailable"
	case strings.Contains(lowerErr, "overloaded") || strings.Contains(lowerErr, "capacity"):
		return "overloaded"
	case strings.Contains(lowerErr, "invalid") || strings.Contains(lowerErr, "400"):
		return "invalid_request"
	case strings.Contains(lowerErr, "empty") || strings.Contains(lowerErr, "no content"):
		return "empty_response"
	default:
		return "unknown"
	}
}

// FormatPhaseFooterMarkdown formats a phase footer in Markdown
func FormatPhaseFooterMarkdown(duration time.Duration) string {
	return fmt.Sprintf("\n_Phase completed in %s_\n\n---\n", formatDuration(duration))
}

// ============================================================================
// Plain Text Formatting Functions (No Formatting At All)
// ============================================================================

// FormatDebateTeamIntroductionPlain formats the debate team introduction in plain text
func FormatDebateTeamIntroductionPlain(topic string, members []*services.DebateTeamMember) string {
	var sb strings.Builder

	sb.WriteString("\nHELIXAGENT AI DEBATE ENSEMBLE\n")
	sb.WriteString("Five AI minds deliberate to synthesize the optimal response.\n\n")

	topicDisplay := topic
	if len(topicDisplay) > 70 {
		topicDisplay = topicDisplay[:70] + "..."
	}
	sb.WriteString(fmt.Sprintf("Topic: %s\n\n", topicDisplay))

	sb.WriteString("Debate Team:\n")
	for _, member := range members {
		if member == nil {
			continue
		}
		roleName := getRoleName(member.Role)
		sb.WriteString(fmt.Sprintf("  %s: %s (%s)\n",
			roleName, member.ModelName, member.ProviderName))
		if member.Fallback != nil {
			sb.WriteString(fmt.Sprintf("    Fallback: %s (%s)\n",
				member.Fallback.ModelName, member.Fallback.ProviderName))
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

// FormatPhaseHeaderPlain formats a phase header in plain text
func FormatPhaseHeaderPlain(phase services.ValidationPhase, phaseNum int) string {
	return fmt.Sprintf("\n=== PHASE %d: %s ===\n\n", phaseNum, strings.ToUpper(string(phase)))
}

// FormatFinalResponsePlain formats the final response in plain text
func FormatFinalResponsePlain(content string) string {
	return fmt.Sprintf("\n=== FINAL ANSWER ===\n\n%s\n", content)
}

// ============================================================================
// Universal Formatting Functions (Auto-detect format)
// ============================================================================

// FormatOutput formats content based on the specified output format
func FormatOutput(format OutputFormat, content string) string {
	switch format {
	case OutputFormatANSI:
		return content // Assume already formatted with ANSI
	case OutputFormatMarkdown:
		return StripANSI(content)
	case OutputFormatPlain:
		return StripANSI(StripMarkdown(content))
	default:
		return content
	}
}

// FormatDebateTeamIntroductionForFormat formats the debate team introduction for the specified format
func FormatDebateTeamIntroductionForFormat(format OutputFormat, topic string, members []*services.DebateTeamMember) string {
	switch format {
	case OutputFormatANSI:
		return FormatDebateTeamIntroduction(topic, members)
	case OutputFormatMarkdown:
		return FormatDebateTeamIntroductionMarkdown(topic, members)
	case OutputFormatPlain:
		return FormatDebateTeamIntroductionPlain(topic, members)
	default:
		return FormatDebateTeamIntroductionMarkdown(topic, members)
	}
}

// FormatPhaseHeaderForFormat formats a phase header for the specified format
func FormatPhaseHeaderForFormat(format OutputFormat, phase services.ValidationPhase, phaseNum int) string {
	switch format {
	case OutputFormatANSI:
		return FormatPhaseHeader(phase, phaseNum)
	case OutputFormatMarkdown:
		return FormatPhaseHeaderMarkdown(phase, phaseNum)
	case OutputFormatPlain:
		return FormatPhaseHeaderPlain(phase, phaseNum)
	default:
		return FormatPhaseHeaderMarkdown(phase, phaseNum)
	}
}

// FormatFinalResponseForFormat formats the final response for the specified format
func FormatFinalResponseForFormat(format OutputFormat, content string) string {
	switch format {
	case OutputFormatANSI:
		return FormatFinalResponse(content)
	case OutputFormatMarkdown:
		return FormatFinalResponseMarkdown(content)
	case OutputFormatPlain:
		return FormatFinalResponsePlain(content)
	default:
		return FormatFinalResponseMarkdown(content)
	}
}

// ============================================================================
// Stripping Functions
// ============================================================================

// StripMarkdown removes common Markdown formatting from text
func StripMarkdown(s string) string {
	// Remove headers
	s = regexp.MustCompile(`(?m)^#{1,6}\s+`).ReplaceAllString(s, "")

	// Remove bold/italic
	s = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(s, "$1")
	s = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(s, "$1")
	s = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(s, "$1")
	s = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(s, "$1")

	// Remove code blocks
	s = regexp.MustCompile("```[^`]*```").ReplaceAllString(s, "")
	s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "$1")

	// Remove links
	s = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(s, "$1")

	// Remove horizontal rules
	s = regexp.MustCompile(`(?m)^---+$`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?m)^\*\*\*+$`).ReplaceAllString(s, "")

	// Remove blockquotes
	s = regexp.MustCompile(`(?m)^>\s*`).ReplaceAllString(s, "")

	// Remove table formatting
	s = regexp.MustCompile(`\|`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`(?m)^[-:]+$`).ReplaceAllString(s, "")

	return s
}

// StripAllFormatting removes both ANSI codes and Markdown formatting
func StripAllFormatting(s string) string {
	return StripMarkdown(StripANSI(s))
}

// ============================================================================
// Enhanced ANSI Stripping (with regex for edge cases)
// ============================================================================

// ansiRegex matches all ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSIRegex removes all ANSI escape codes using regex (more thorough than StripANSI)
func StripANSIRegex(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// ContainsANSI checks if a string contains ANSI escape codes
func ContainsANSI(s string) bool {
	return ansiRegex.MatchString(s)
}

// ============================================================================
// Output Format Detection
// ============================================================================

// DetectOutputFormat detects the appropriate output format based on client hints
func DetectOutputFormat(acceptHeader string, userAgent string, formatHint string) OutputFormat {
	// Explicit format hint takes priority
	switch strings.ToLower(formatHint) {
	case "ansi", "terminal", "tty":
		return OutputFormatANSI
	case "markdown", "md":
		return OutputFormatMarkdown
	case "plain", "text":
		return OutputFormatPlain
	}

	// Check Accept header
	if strings.Contains(acceptHeader, "text/plain") {
		return OutputFormatPlain
	}

	// Check User-Agent for known terminal clients
	terminalClients := []string{
		"curl", "wget", "httpie", "terminal",
	}
	userAgentLower := strings.ToLower(userAgent)
	for _, client := range terminalClients {
		if strings.Contains(userAgentLower, client) {
			return OutputFormatANSI
		}
	}

	// Check User-Agent for known API clients that don't support ANSI
	apiClients := []string{
		"opencode", "crush", "claude", "cline", "continue",
		"cursor", "vscode", "visual studio",
		"postman", "insomnia", "httpbin",
		"python", "node", "java", "go-http",
	}
	for _, client := range apiClients {
		if strings.Contains(userAgentLower, client) {
			return OutputFormatMarkdown
		}
	}

	// Default to Markdown for API responses (safest choice)
	return OutputFormatMarkdown
}

// FormatConsensusHeaderSimpleMarkdown formats the consensus section header in Markdown (without confidence score)
func FormatConsensusHeaderSimpleMarkdown() string {
	return "\n---\n\n## Consensus\n\n"
}

// FormatConsensusHeaderForFormat formats the consensus header for the specified format
func FormatConsensusHeaderForFormat(format OutputFormat) string {
	switch format {
	case OutputFormatANSI:
		return FormatConsensusHeader()
	case OutputFormatMarkdown:
		return FormatConsensusHeaderSimpleMarkdown()
	case OutputFormatPlain:
		return "\n=== CONSENSUS ===\n\n"
	default:
		return FormatConsensusHeaderSimpleMarkdown()
	}
}

// FormatRequestIndicatorForFormat formats a request indicator for the specified format
func FormatRequestIndicatorForFormat(format OutputFormat, position services.DebateTeamPosition, role services.DebateRole, provider, model string) string {
	switch format {
	case OutputFormatANSI:
		return FormatRequestIndicator(position, role, provider, model)
	case OutputFormatMarkdown:
		return FormatRequestIndicatorMarkdown(role, provider, model)
	case OutputFormatPlain:
		return fmt.Sprintf("[%s] Request sent to %s (%s)\n", getRoleName(role), provider, model)
	default:
		return FormatRequestIndicatorMarkdown(role, provider, model)
	}
}

// FormatResponseIndicatorSimpleMarkdown formats a response indicator in simple Markdown (role + duration only)
func FormatResponseIndicatorSimpleMarkdown(role services.DebateRole, responseTime time.Duration) string {
	roleName := getRoleName(role)
	return fmt.Sprintf("**[%s]** Response received (%s)\n", roleName, formatDuration(responseTime))
}

// FormatResponseIndicatorForFormat formats a response indicator for the specified format
func FormatResponseIndicatorForFormat(format OutputFormat, position services.DebateTeamPosition, role services.DebateRole, responseTime time.Duration) string {
	switch format {
	case OutputFormatANSI:
		return FormatResponseIndicator(position, role, responseTime)
	case OutputFormatMarkdown:
		return FormatResponseIndicatorSimpleMarkdown(role, responseTime)
	case OutputFormatPlain:
		return fmt.Sprintf("[%s] Response received (%s)\n", getRoleName(role), formatDuration(responseTime))
	default:
		return FormatResponseIndicatorSimpleMarkdown(role, responseTime)
	}
}

// FormatFallbackIndicatorSimpleMarkdown formats a fallback indicator in simple Markdown
func FormatFallbackIndicatorSimpleMarkdown(role services.DebateRole, fallbackProvider, fallbackModel string, responseTime time.Duration) string {
	roleName := getRoleName(role)
	return fmt.Sprintf("**[%s]** _Fallback to %s (%s)_ - %s\n", roleName, fallbackProvider, fallbackModel, formatDuration(responseTime))
}

// FormatPhaseContentForFormat formats debate phase content for the specified format
func FormatPhaseContentForFormat(format OutputFormat, content string) string {
	switch format {
	case OutputFormatANSI:
		return FormatPhaseContent(content)
	case OutputFormatMarkdown:
		// For Markdown, use a quote block for visual separation (no ANSI codes)
		return FormatPhaseContentMarkdown(content)
	case OutputFormatPlain:
		return content // Plain text - no formatting
	default:
		return FormatPhaseContentMarkdown(content)
	}
}

// FormatFallbackIndicatorForFormat formats a fallback indicator for the specified format
func FormatFallbackIndicatorForFormat(format OutputFormat, position services.DebateTeamPosition, role services.DebateRole, fallbackProvider, fallbackModel string, responseTime time.Duration) string {
	switch format {
	case OutputFormatANSI:
		return FormatFallbackIndicator(position, role, fallbackProvider, fallbackModel, responseTime)
	case OutputFormatMarkdown:
		return FormatFallbackIndicatorSimpleMarkdown(role, fallbackProvider, fallbackModel, responseTime)
	case OutputFormatPlain:
		return fmt.Sprintf("[%s] Fallback to %s (%s) - %s\n", getRoleName(role), fallbackProvider, fallbackModel, formatDuration(responseTime))
	default:
		return FormatFallbackIndicatorSimpleMarkdown(role, fallbackProvider, fallbackModel, responseTime)
	}
}

// FormatFallbackWithErrorForFormat formats a detailed fallback indicator with error cause
// This is sent to CLI agent plugins to display the exact failure reason
func FormatFallbackWithErrorForFormat(format OutputFormat, role services.DebateRole, primaryProvider, primaryModel, fallbackProvider, fallbackModel, errorMsg string, attemptNum int, duration time.Duration) string {
	errorCategory := categorizeErrorString(errorMsg)
	categoryIcon := getCategoryIcon(errorCategory)

	switch format {
	case OutputFormatANSI:
		// ANSI format with colors for terminal
		return formatFallbackWithErrorANSI(role, primaryProvider, primaryModel, fallbackProvider, fallbackModel, errorMsg, errorCategory, categoryIcon, attemptNum, duration)
	case OutputFormatMarkdown:
		// Clean Markdown for API clients
		return FormatFallbackTriggeredMarkdown(getRoleName(role), primaryProvider, primaryModel, fallbackProvider, fallbackModel, errorMsg, errorCategory, duration)
	case OutputFormatPlain:
		// Plain text
		return fmt.Sprintf("[%s] Fallback from %s/%s to %s/%s\n  Error: %s (%s)\n",
			getRoleName(role), primaryProvider, primaryModel, fallbackProvider, fallbackModel, errorMsg, errorCategory)
	default:
		return FormatFallbackTriggeredMarkdown(getRoleName(role), primaryProvider, primaryModel, fallbackProvider, fallbackModel, errorMsg, errorCategory, duration)
	}
}

// formatFallbackWithErrorANSI formats a fallback indicator with ANSI colors
func formatFallbackWithErrorANSI(role services.DebateRole, primaryProvider, primaryModel, fallbackProvider, fallbackModel, errorMsg, errorCategory, categoryIcon string, attemptNum int, duration time.Duration) string {
	roleColor := getRoleColor(role)
	roleName := getRoleName(role)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s⚡ [%s]%s Fallback triggered\n", ANSIBrightYellow, roleName, ANSIReset))
	sb.WriteString(fmt.Sprintf("   %sPrimary:%s %s/%s (%s)\n", ANSIDim, ANSIReset, primaryProvider, primaryModel, formatDuration(duration)))
	sb.WriteString(fmt.Sprintf("   %s%s Error:%s %s\n", ANSIBrightRed, categoryIcon, ANSIReset, errorMsg))
	if fallbackProvider != "" {
		sb.WriteString(fmt.Sprintf("   %s→ Trying:%s %s%s/%s%s\n", ANSIDim, ANSIReset, roleColor, fallbackProvider, fallbackModel, ANSIReset))
	}
	sb.WriteString("\n")
	return sb.String()
}

// FormatFallbackChainWithErrorsForFormat formats a complete fallback chain with all error causes
func FormatFallbackChainWithErrorsForFormat(format OutputFormat, position services.DebateTeamPosition, role services.DebateRole, chain []FallbackAttempt, totalTime time.Duration) string {
	if len(chain) == 0 {
		return ""
	}

	switch format {
	case OutputFormatANSI:
		return FormatFallbackChainIndicator(position, role, chain, totalTime)
	case OutputFormatMarkdown:
		return FormatFallbackChainMarkdown(position, chain)
	case OutputFormatPlain:
		return formatFallbackChainPlain(position, chain)
	default:
		return FormatFallbackChainMarkdown(position, chain)
	}
}

// formatFallbackChainPlain formats a fallback chain in plain text
func formatFallbackChainPlain(position services.DebateTeamPosition, chain []FallbackAttempt) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\nFallback Chain for Position %d:\n", position))
	for i, attempt := range chain {
		status := "FAILED"
		if attempt.Success {
			status = "OK"
		}
		sb.WriteString(fmt.Sprintf("  %d. [%s] %s/%s (%s)\n", i+1, status, attempt.Provider, attempt.Model, formatDuration(attempt.Duration)))
		if attempt.Error != "" {
			sb.WriteString(fmt.Sprintf("     Error: %s\n", attempt.Error))
		}
	}
	return sb.String()
}

// IsTerminalClient checks if the user agent suggests a terminal client
func IsTerminalClient(userAgent string) bool {
	userAgentLower := strings.ToLower(userAgent)
	terminalPatterns := []string{
		"curl", "wget", "httpie",
		"terminal", "tty", "console",
		"cli", "command-line",
	}
	for _, pattern := range terminalPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}
	return false
}
