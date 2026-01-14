package services

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDebateCommLogger tests logger creation
func TestNewDebateCommLogger(t *testing.T) {
	logger := logrus.New()
	commLogger := NewDebateCommLogger(logger)

	assert.NotNil(t, commLogger)
	assert.NotNil(t, commLogger.logger)
	assert.True(t, commLogger.enableColors)
	assert.Equal(t, "unknown", commLogger.cliAgent)
}

// TestSetCLIAgent tests setting CLI agent
func TestSetCLIAgent(t *testing.T) {
	logger := logrus.New()
	commLogger := NewDebateCommLogger(logger)

	testCases := []struct {
		agent    string
		expected string
	}{
		{"OpenCode", "opencode"},
		{"ClaudeCode", "claudecode"},
		{"KiloCode", "kilocode"},
		{"CRUSH", "crush"},
		{"aider", "aider"},
	}

	for _, tc := range testCases {
		t.Run(tc.agent, func(t *testing.T) {
			commLogger.SetCLIAgent(tc.agent)
			assert.Equal(t, tc.expected, commLogger.cliAgent)
		})
	}
}

// TestSetColorsEnabled tests color toggle
func TestSetColorsEnabled(t *testing.T) {
	logger := logrus.New()
	commLogger := NewDebateCommLogger(logger)

	assert.True(t, commLogger.enableColors)

	commLogger.SetColorsEnabled(false)
	assert.False(t, commLogger.enableColors)

	commLogger.SetColorsEnabled(true)
	assert.True(t, commLogger.enableColors)
}

// TestRoleAbbreviations tests role abbreviation mapping
func TestRoleAbbreviations(t *testing.T) {
	expectedAbbrevs := map[string]string{
		"analyst":     "A",
		"proposer":    "P",
		"critic":      "C",
		"synthesizer": "S",
		"mediator":    "M",
		"default":     "D",
	}

	for role, abbrev := range expectedAbbrevs {
		t.Run(role, func(t *testing.T) {
			assert.Equal(t, abbrev, RoleAbbreviations[role])
		})
	}
}

// TestRoleColors tests role color mapping
func TestRoleColors(t *testing.T) {
	expectedColors := map[string]string{
		"analyst":     ColorAnalyst,
		"proposer":    ColorProposer,
		"critic":      ColorCritic,
		"synthesizer": ColorSynthesizer,
		"mediator":    ColorMediator,
		"default":     ColorWhite,
	}

	for role, color := range expectedColors {
		t.Run(role, func(t *testing.T) {
			assert.Equal(t, color, RoleColors[role])
		})
	}
}

// TestFormatModelName tests model name formatting
func TestFormatModelName(t *testing.T) {
	testCases := []struct {
		model    string
		expected string
	}{
		{"claude-opus-4-5-20251101", "Claude Opus 4.5"},
		{"claude-sonnet-4-5-20250929", "Claude Sonnet 4.5"},
		{"deepseek-chat", "DeepSeek Chat"},
		{"gemini-2.0-flash", "Gemini 2.0 Flash"},
		{"qwen-max", "Qwen Max"},
		{"unknown-model-xyz", "unknown-model-xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			result := formatModelName(tc.model)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFormatRoleTag tests role tag formatting with colors
func TestFormatRoleTag(t *testing.T) {
	logger := logrus.New()
	commLogger := NewDebateCommLogger(logger)

	t.Run("with_colors_enabled", func(t *testing.T) {
		commLogger.SetColorsEnabled(true)
		tag := commLogger.formatRoleTag("analyst", "claude", "claude-opus-4-5-20251101")

		// Should contain ANSI color codes and proper formatting
		assert.Contains(t, tag, "[A:")
		assert.Contains(t, tag, "Claude Opus 4.5")
		assert.Contains(t, tag, ColorReset)
	})

	t.Run("with_colors_disabled", func(t *testing.T) {
		commLogger.SetColorsEnabled(false)
		tag := commLogger.formatRoleTag("analyst", "claude", "claude-opus-4-5-20251101")

		assert.Equal(t, "[A: Claude Opus 4.5]", tag)
		assert.NotContains(t, tag, "\033[")
	})

	t.Run("unknown_role", func(t *testing.T) {
		commLogger.SetColorsEnabled(false)
		tag := commLogger.formatRoleTag("unknown_role", "deepseek", "deepseek-chat")

		// Should use default abbreviation
		assert.Equal(t, "[D: DeepSeek Chat]", tag)
	})
}

// TestLogRequest tests request logging
func TestLogRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogRequest("analyst", "claude", "claude-opus-4-5-20251101", 1024, 1)

	output := buf.String()
	assert.Contains(t, output, "[A: Claude Opus 4.5]")
	assert.Contains(t, output, "<---")
	assert.Contains(t, output, "Sending request")
	assert.Contains(t, output, "round 1")
	assert.Contains(t, output, "1024 chars")
}

// TestLogResponse tests response logging
func TestLogResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogResponse("mediator", "gemini", "gemini-2.0-flash", 2048, 1500*time.Millisecond, 0.85)

	output := buf.String()
	assert.Contains(t, output, "[M: Gemini 2.0 Flash]")
	assert.Contains(t, output, "--->")
	assert.Contains(t, output, "Received")
	assert.Contains(t, output, "2048 bytes")
	assert.Contains(t, output, "1.50s")
	assert.Contains(t, output, "quality: 0.85")
}

// TestLogResponsePreview tests response preview logging
func TestLogResponsePreview(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	t.Run("short_content", func(t *testing.T) {
		buf.Reset()
		commLogger.LogResponsePreview("analyst", "deepseek", "deepseek-chat", "Short response", 100)
		output := buf.String()
		// Logger may escape quotes, so just check for the content
		assert.Contains(t, output, "Short response")
		assert.Contains(t, output, "14 bytes")
	})

	t.Run("long_content_truncated", func(t *testing.T) {
		buf.Reset()
		longContent := strings.Repeat("This is a very long response content. ", 10)
		commLogger.LogResponsePreview("proposer", "qwen", "qwen-max", longContent, 50)
		output := buf.String()
		assert.Contains(t, output, "...")
	})

	t.Run("multiline_content", func(t *testing.T) {
		buf.Reset()
		multilineContent := "Line 1\nLine 2\nLine 3"
		commLogger.LogResponsePreview("critic", "mistral", "mistral-large-latest", multilineContent, 100)
		output := buf.String()
		// Check that the combined content appears (newlines converted to spaces)
		assert.Contains(t, output, "Line 1")
		assert.Contains(t, output, "Line 2")
		assert.Contains(t, output, "Line 3")
	})
}

// TestLogStreamStart tests stream start logging
func TestLogStreamStart(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogStreamStart("synthesizer", "claude", "claude-sonnet-4-5-20250929")

	output := buf.String()
	assert.Contains(t, output, "[S: Claude Sonnet 4.5]")
	assert.Contains(t, output, "--->")
	assert.Contains(t, output, "[STREAM START]")
}

// TestLogStreamChunk tests stream chunk logging
func TestLogStreamChunk(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogStreamChunk("analyst", "deepseek", "deepseek-chat", 128, 1024)

	output := buf.String()
	assert.Contains(t, output, "[CHUNK]")
	assert.Contains(t, output, "+128 bytes")
	assert.Contains(t, output, "total: 1024")
}

// TestLogStreamEnd tests stream end logging
func TestLogStreamEnd(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogStreamEnd("mediator", "gemini", "gemini-2.0-flash", 4096, 2500*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "[M: Gemini 2.0 Flash]")
	assert.Contains(t, output, "[STREAM END]")
	assert.Contains(t, output, "4096 bytes")
	assert.Contains(t, output, "2.50s")
}

// TestLogFallbackAttempt tests fallback attempt logging
func TestLogFallbackAttempt(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogFallbackAttempt("analyst", "claude", "claude-opus-4-5-20251101", "deepseek", "deepseek-chat", 1)

	output := buf.String()
	assert.Contains(t, output, "[A: Claude Opus 4.5]")
	assert.Contains(t, output, "--->")
	assert.Contains(t, output, "[FALLBACK #1: DeepSeek Chat]")
	assert.Contains(t, output, "Attempting fallback")
}

// TestLogFallbackSuccess tests fallback success logging
func TestLogFallbackSuccess(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogFallbackSuccess("mediator", "claude", "claude-opus-4-5-20251101", "gemini", "gemini-2.0-flash", 1, 1500*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "[M: Claude Opus 4.5]")
	assert.Contains(t, output, "[FALLBACK #1: Gemini 2.0 Flash]")
	assert.Contains(t, output, "Success!")
	assert.Contains(t, output, "1.50s")
}

// TestLogFallbackChain tests fallback chain logging
func TestLogFallbackChain(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	chain := []FallbackChainEntry{
		{Provider: "claude", Model: "claude-opus-4-5-20251101", Success: false, Error: errors.New("timeout"), Duration: 5 * time.Second},
		{Provider: "deepseek", Model: "deepseek-chat", Success: false, Error: errors.New("rate limit"), Duration: 2 * time.Second},
		{Provider: "gemini", Model: "gemini-2.0-flash", Success: true, Duration: 1 * time.Second},
	}

	commLogger.LogFallbackChain("analyst", chain, "Final response content here", 8*time.Second)

	output := buf.String()
	assert.Contains(t, output, "[A: Claude Opus 4.5]")
	assert.Contains(t, output, "[FALLBACK: DeepSeek Chat]")
	assert.Contains(t, output, "[FALLBACK: Gemini 2.0 Flash]")
	assert.Contains(t, output, "Final response")
	assert.Contains(t, output, "8.00s")
}

// TestLogFallbackChainEmpty tests empty fallback chain
func TestLogFallbackChainEmpty(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	// Empty chain should not log anything
	commLogger.LogFallbackChain("analyst", []FallbackChainEntry{}, "content", time.Second)

	assert.Empty(t, buf.String())
}

// TestLogError tests error logging
func TestLogError(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogError("critic", "claude", "claude-opus-4-5-20251101", errors.New("connection timeout"))

	output := buf.String()
	assert.Contains(t, output, "[C: Claude Opus 4.5]")
	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "connection timeout")
}

// TestLogAllFallbacksExhausted tests exhausted fallbacks logging
func TestLogAllFallbacksExhausted(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogAllFallbacksExhausted("proposer", "claude", "claude-opus-4-5-20251101", 3)

	output := buf.String()
	assert.Contains(t, output, "[P: Claude Opus 4.5]")
	assert.Contains(t, output, "[EXHAUSTED]")
	assert.Contains(t, output, "All 3 fallbacks failed")
}

// TestLogDebatePhase tests debate phase logging
func TestLogDebatePhase(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogDebatePhase("Getting participant responses", 2)

	output := buf.String()
	assert.Contains(t, output, "═══ DEBATE PHASE:")
	assert.Contains(t, output, "Getting participant responses")
	assert.Contains(t, output, "Round 2")
}

// TestLogDebateSummary tests debate summary logging
func TestLogDebateSummary(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogDebateSummary(1, 5, 10*time.Second, 0.85, 2)

	output := buf.String()
	assert.Contains(t, output, "═══ ROUND 1 SUMMARY ═══")
	assert.Contains(t, output, "Participants: 5")
	assert.Contains(t, output, "Duration: 10.00s")
	assert.Contains(t, output, "Avg Quality: 0.85")
	assert.Contains(t, output, "Fallbacks Used: 2")
}

// TestLogDebateSummaryNoFallbacks tests summary without fallbacks
func TestLogDebateSummaryNoFallbacks(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	commLogger.LogDebateSummary(3, 5, 8*time.Second, 0.92, 0)

	output := buf.String()
	assert.Contains(t, output, "═══ ROUND 3 SUMMARY ═══")
	assert.NotContains(t, output, "Fallbacks Used")
}

// TestCLIAgentColors tests CLI agent color support detection
func TestCLIAgentColors(t *testing.T) {
	// All 18 supported CLI agents
	supportedAgents := []string{
		"opencode", "claudecode", "kilocode", "crush", "helixcode",
		"kiro", "aider", "cline", "codenamegoose", "deepseekcli",
		"forge", "geminicli", "gptengineer", "mistralcode", "ollamacode",
		"plandex", "qwencode", "amazonq",
	}

	for _, agent := range supportedAgents {
		t.Run(agent, func(t *testing.T) {
			config := CLIAgentColors(agent)
			assert.True(t, config["colors"], "CLI agent %s should support colors", agent)
		})
	}

	t.Run("unknown_agent", func(t *testing.T) {
		config := CLIAgentColors("unknown")
		assert.False(t, config["colors"])
	})

	t.Run("empty_string", func(t *testing.T) {
		config := CLIAgentColors("")
		assert.True(t, config["colors"]) // Default should be true
	})
}

// TestFormatRetrofitLog tests the main Retrofit format function
func TestFormatRetrofitLog(t *testing.T) {
	logger := logrus.New()
	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	t.Run("request_direction", func(t *testing.T) {
		log := commLogger.FormatRetrofitLog(
			"request",
			"analyst",
			"claude",
			"claude-opus-4-5-20251101",
			"What is your analysis?",
			map[string]interface{}{"round": 1},
		)

		assert.Contains(t, log, "[A: Claude Opus 4.5]")
		assert.Contains(t, log, "<---")
		assert.Contains(t, log, "What is your analysis?")
		assert.Contains(t, log, "round=1")
	})

	t.Run("response_direction", func(t *testing.T) {
		log := commLogger.FormatRetrofitLog(
			"response",
			"mediator",
			"gemini",
			"gemini-2.0-flash",
			"Here is my response content",
			map[string]interface{}{
				"duration": 1500 * time.Millisecond,
				"bytes":    2048,
				"quality":  0.85,
			},
		)

		assert.Contains(t, log, "[M: Gemini 2.0 Flash]")
		assert.Contains(t, log, "--->")
		assert.Contains(t, log, "Here is my response")
		assert.Contains(t, log, "1.50s")
		assert.Contains(t, log, "2048 bytes")
		assert.Contains(t, log, "quality=0.85")
	})

	t.Run("fallback_direction", func(t *testing.T) {
		log := commLogger.FormatRetrofitLog(
			"fallback",
			"proposer",
			"deepseek",
			"deepseek-chat",
			"Fallback response",
			map[string]interface{}{"fallback_index": 2},
		)

		assert.Contains(t, log, "[P: DeepSeek Chat]")
		assert.Contains(t, log, "--->")
		assert.Contains(t, log, "fallback=#2")
	})

	t.Run("error_direction", func(t *testing.T) {
		log := commLogger.FormatRetrofitLog(
			"error",
			"critic",
			"qwen",
			"qwen-max",
			"Error occurred",
			nil,
		)

		assert.Contains(t, log, "[C: Qwen Max]")
		assert.Contains(t, log, "--->")
	})

	t.Run("long_content_truncated", func(t *testing.T) {
		longContent := strings.Repeat("This is very long content. ", 20)
		log := commLogger.FormatRetrofitLog(
			"response",
			"synthesizer",
			"mistral",
			"mistral-large",
			longContent,
			nil,
		)

		assert.Contains(t, log, "...")
		assert.Less(t, len(log), len(longContent)+100)
	})
}

// TestFallbackChainEntry tests the FallbackChainEntry struct
func TestFallbackChainEntry(t *testing.T) {
	entry := FallbackChainEntry{
		Provider: "claude",
		Model:    "claude-opus-4-5-20251101",
		Success:  true,
		Error:    nil,
		Duration: 1500 * time.Millisecond,
	}

	assert.Equal(t, "claude", entry.Provider)
	assert.Equal(t, "claude-opus-4-5-20251101", entry.Model)
	assert.True(t, entry.Success)
	assert.Nil(t, entry.Error)
	assert.Equal(t, 1500*time.Millisecond, entry.Duration)
}

// TestColorConstantsExist tests that all color constants are defined
func TestColorConstantsExist(t *testing.T) {
	// Basic colors
	assert.NotEmpty(t, ColorReset)
	assert.NotEmpty(t, ColorRed)
	assert.NotEmpty(t, ColorGreen)
	assert.NotEmpty(t, ColorYellow)
	assert.NotEmpty(t, ColorBlue)
	assert.NotEmpty(t, ColorMagenta)
	assert.NotEmpty(t, ColorCyan)
	assert.NotEmpty(t, ColorWhite)
	assert.NotEmpty(t, ColorBold)
	assert.NotEmpty(t, ColorDim)

	// Role-specific colors
	assert.NotEmpty(t, ColorAnalyst)
	assert.NotEmpty(t, ColorProposer)
	assert.NotEmpty(t, ColorCritic)
	assert.NotEmpty(t, ColorSynthesizer)
	assert.NotEmpty(t, ColorMediator)

	// Communication colors
	assert.NotEmpty(t, ColorRequest)
	assert.NotEmpty(t, ColorResponse)
	assert.NotEmpty(t, ColorFallback)
	assert.NotEmpty(t, ColorError)
	assert.NotEmpty(t, ColorStream)
}

// TestAllDebateRolesHaveAbbreviationsAndColors tests coverage of all roles
func TestAllDebateRolesHaveAbbreviationsAndColors(t *testing.T) {
	debateRoles := []string{"analyst", "proposer", "critic", "synthesizer", "mediator", "default"}

	for _, role := range debateRoles {
		t.Run(role, func(t *testing.T) {
			abbrev, hasAbbrev := RoleAbbreviations[role]
			assert.True(t, hasAbbrev, "Role %s should have abbreviation", role)
			assert.NotEmpty(t, abbrev)

			color, hasColor := RoleColors[role]
			assert.True(t, hasColor, "Role %s should have color", role)
			assert.NotEmpty(t, color)
		})
	}
}

// TestLogOutputContainsExpectedElements tests comprehensive output
func TestLogOutputContainsExpectedElements(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	commLogger := NewDebateCommLogger(logger)
	commLogger.SetColorsEnabled(false)

	// Test a complete debate round simulation
	t.Run("full_debate_round_simulation", func(t *testing.T) {
		// Log phase start
		commLogger.LogDebatePhase("Starting deliberation", 1)

		// Log request for analyst
		commLogger.LogRequest("analyst", "claude", "claude-opus-4-5-20251101", 500, 1)

		// Log response for analyst
		commLogger.LogResponse("analyst", "claude", "claude-opus-4-5-20251101", 1024, time.Second, 0.9)

		// Log request for proposer
		commLogger.LogRequest("proposer", "deepseek", "deepseek-chat", 600, 1)

		// Log error for proposer (to trigger fallback)
		commLogger.LogError("proposer", "deepseek", "deepseek-chat", errors.New("rate limited"))

		// Log fallback attempt
		commLogger.LogFallbackAttempt("proposer", "deepseek", "deepseek-chat", "gemini", "gemini-2.0-flash", 1)

		// Log fallback success
		commLogger.LogFallbackSuccess("proposer", "deepseek", "deepseek-chat", "gemini", "gemini-2.0-flash", 1, 800*time.Millisecond)

		// Log summary
		commLogger.LogDebateSummary(1, 2, 3*time.Second, 0.85, 1)

		output := buf.String()

		// Verify all expected elements are present
		require.Contains(t, output, "═══ DEBATE PHASE:")
		require.Contains(t, output, "[A: Claude Opus 4.5]")
		require.Contains(t, output, "[P: DeepSeek Chat]")
		require.Contains(t, output, "<---")
		require.Contains(t, output, "--->")
		require.Contains(t, output, "[ERROR]")
		require.Contains(t, output, "[FALLBACK #1:")
		require.Contains(t, output, "Success!")
		require.Contains(t, output, "═══ ROUND 1 SUMMARY ═══")
	})
}

// TestCLIAgentIntegration tests CLI agent configuration integration
func TestCLIAgentIntegration(t *testing.T) {
	logger := logrus.New()
	commLogger := NewDebateCommLogger(logger)

	// Test all 18 CLI agents
	cliAgents := []string{
		"opencode", "claudecode", "kilocode", "crush", "helixcode",
		"kiro", "aider", "cline", "codenamegoose", "deepseekcli",
		"forge", "geminicli", "gptengineer", "mistralcode", "ollamacode",
		"plandex", "qwencode", "amazonq",
	}

	for _, agent := range cliAgents {
		t.Run(agent, func(t *testing.T) {
			commLogger.SetCLIAgent(agent)
			colorConfig := CLIAgentColors(agent)
			commLogger.SetColorsEnabled(colorConfig["colors"])

			// All supported agents should have colors enabled
			assert.True(t, commLogger.enableColors)
			assert.Equal(t, agent, commLogger.cliAgent)
		})
	}
}

// BenchmarkFormatRoleTag benchmarks role tag formatting
func BenchmarkFormatRoleTag(b *testing.B) {
	logger := logrus.New()
	commLogger := NewDebateCommLogger(logger)

	b.Run("with_colors", func(b *testing.B) {
		commLogger.SetColorsEnabled(true)
		for i := 0; i < b.N; i++ {
			commLogger.formatRoleTag("analyst", "claude", "claude-opus-4-5-20251101")
		}
	})

	b.Run("without_colors", func(b *testing.B) {
		commLogger.SetColorsEnabled(false)
		for i := 0; i < b.N; i++ {
			commLogger.formatRoleTag("analyst", "claude", "claude-opus-4-5-20251101")
		}
	})
}

// BenchmarkFormatModelName benchmarks model name formatting
func BenchmarkFormatModelName(b *testing.B) {
	b.Run("known_model", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatModelName("claude-opus-4-5-20251101")
		}
	})

	b.Run("unknown_model", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatModelName("custom-unknown-model-v1")
		}
	})
}

// BenchmarkLogRequest benchmarks request logging
func BenchmarkLogRequest(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{}) // Discard output
	commLogger := NewDebateCommLogger(logger)

	for i := 0; i < b.N; i++ {
		commLogger.LogRequest("analyst", "claude", "claude-opus-4-5-20251101", 1024, 1)
	}
}
