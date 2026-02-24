package services

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// newDebateFormatterTestLogger creates a silent logger for formatter tests.
func newDebateFormatterTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	return logger
}

// newTestFormatterExecutor creates a FormatterExecutor for testing.
// It uses an empty registry so format calls will return errors (no formatter found)
// rather than panicking from nil dereference.
func newTestFormatterExecutor() *formatters.FormatterExecutor {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	registry := formatters.NewFormatterRegistry(&formatters.RegistryConfig{}, logger)
	cfg := &formatters.ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
		MaxRetries:     0,
		EnableCache:    false,
		EnableMetrics:  false,
		EnableTracing:  false,
	}
	return formatters.NewFormatterExecutor(registry, cfg, logger)
}

// =============================================================================
// DefaultDebateFormatterConfig Tests
// =============================================================================

func TestDefaultDebateFormatterConfig(t *testing.T) {
	cfg := DefaultDebateFormatterConfig()

	require.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.True(t, cfg.AutoFormat)
	assert.Empty(t, cfg.FormatLanguages)
	assert.Empty(t, cfg.IgnoreLanguages)
	assert.Equal(t, 50000, cfg.MaxCodeBlockSize)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.True(t, cfg.ContinueOnError)
}

func TestDefaultDebateFormatterConfig_ValuesAreReasonable(t *testing.T) {
	cfg := DefaultDebateFormatterConfig()

	assert.True(t, cfg.MaxCodeBlockSize > 0,
		"MaxCodeBlockSize should be positive")
	assert.True(t, cfg.Timeout > 0,
		"Timeout should be positive")
}

// =============================================================================
// NewDebateFormatterIntegration Tests
// =============================================================================

func TestNewDebateFormatterIntegration_WithAllParams(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	executor := newTestFormatterExecutor()
	cfg := DefaultDebateFormatterConfig()

	dfi := NewDebateFormatterIntegration(executor, cfg, logger)

	require.NotNil(t, dfi)
	assert.Equal(t, executor, dfi.executor)
	assert.Equal(t, cfg, dfi.config)
	assert.Equal(t, logger, dfi.logger)
}

func TestNewDebateFormatterIntegration_NilConfig(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	executor := newTestFormatterExecutor()

	dfi := NewDebateFormatterIntegration(executor, nil, logger)

	require.NotNil(t, dfi)
	assert.NotNil(t, dfi.config, "Nil config should be replaced with default")
	assert.True(t, dfi.config.Enabled)
	assert.True(t, dfi.config.AutoFormat)
}

func TestNewDebateFormatterIntegration_NilExecutor(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := DefaultDebateFormatterConfig()

	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	require.NotNil(t, dfi)
	assert.Nil(t, dfi.executor)
}

// =============================================================================
// GetConfig / SetConfig Tests
// =============================================================================

func TestDebateFormatterIntegration_GetConfig(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := DefaultDebateFormatterConfig()
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	retrieved := dfi.GetConfig()
	assert.Equal(t, cfg, retrieved)
}

func TestDebateFormatterIntegration_SetConfig(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	newCfg := &DebateFormatterConfig{
		Enabled:          false,
		AutoFormat:       false,
		FormatLanguages:  []string{"go", "python"},
		IgnoreLanguages:  []string{"markdown"},
		MaxCodeBlockSize: 10000,
		Timeout:          10 * time.Second,
		ContinueOnError:  false,
	}

	dfi.SetConfig(newCfg)

	assert.Equal(t, newCfg, dfi.GetConfig())
	assert.False(t, dfi.config.Enabled)
	assert.False(t, dfi.config.AutoFormat)
	assert.Equal(t, 10000, dfi.config.MaxCodeBlockSize)
}

// =============================================================================
// extractCodeBlocks Tests
// =============================================================================

func TestDebateFormatterIntegration_extractCodeBlocks_NoCodeBlocks(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	blocks := dfi.extractCodeBlocks("This is plain text with no code blocks.")
	assert.Empty(t, blocks)
}

func TestDebateFormatterIntegration_extractCodeBlocks_SingleBlock(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	content := "Here is some code:\n```go\npackage main\n\nfunc main() {}\n```\nEnd of code."

	blocks := dfi.extractCodeBlocks(content)

	require.Len(t, blocks, 1)
	assert.Equal(t, "go", blocks[0].Language)
	assert.Contains(t, blocks[0].Code, "package main")
	assert.Contains(t, blocks[0].Code, "func main() {}")
	assert.Contains(t, blocks[0].Original, "```go")
	assert.Contains(t, blocks[0].Original, "```")
}

func TestDebateFormatterIntegration_extractCodeBlocks_MultipleBlocks(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	content := `First code block:
` + "```python\ndef hello():\n    print('hello')\n```" + `

Second code block:
` + "```javascript\nconst x = 42;\n```" + `

Third code block:
` + "```go\npackage main\n```"

	blocks := dfi.extractCodeBlocks(content)

	require.Len(t, blocks, 3)
	assert.Equal(t, "python", blocks[0].Language)
	assert.Equal(t, "javascript", blocks[1].Language)
	assert.Equal(t, "go", blocks[2].Language)
}

func TestDebateFormatterIntegration_extractCodeBlocks_NoLanguage(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	content := "```\nsome code here\n```"

	blocks := dfi.extractCodeBlocks(content)

	require.Len(t, blocks, 1)
	assert.Equal(t, "", blocks[0].Language)
	assert.Contains(t, blocks[0].Code, "some code here")
}

func TestDebateFormatterIntegration_extractCodeBlocks_VariousLanguages(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	languages := []string{"go", "python", "javascript", "rust", "c", "cpp", "java", "ruby", "sql"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			content := fmt.Sprintf("```%s\ncode here\n```", lang)
			blocks := dfi.extractCodeBlocks(content)
			require.Len(t, blocks, 1)
			assert.Equal(t, lang, blocks[0].Language)
		})
	}
}

func TestDebateFormatterIntegration_extractCodeBlocks_Positions(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	content := "prefix\n```go\ncode\n```\nsuffix"

	blocks := dfi.extractCodeBlocks(content)

	require.Len(t, blocks, 1)
	assert.True(t, blocks[0].Start >= 0)
	assert.True(t, blocks[0].End > blocks[0].Start)
	assert.Equal(t, blocks[0].Original, content[blocks[0].Start:blocks[0].End])
}

// =============================================================================
// shouldFormat Tests
// =============================================================================

func TestDebateFormatterIntegration_shouldFormat_EmptyLanguage(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	block := CodeBlock{Language: "", Code: "some code"}
	assert.False(t, dfi.shouldFormat(block))
}

func TestDebateFormatterIntegration_shouldFormat_TooLargeBlock(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 10,
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	block := CodeBlock{Language: "go", Code: "this is more than 10 bytes"}
	assert.False(t, dfi.shouldFormat(block))
}

func TestDebateFormatterIntegration_shouldFormat_ExactlyAtLimit(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 5,
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	block := CodeBlock{Language: "go", Code: "12345"}
	assert.True(t, dfi.shouldFormat(block))
}

func TestDebateFormatterIntegration_shouldFormat_AllLanguagesAllowed(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 50000,
		FormatLanguages:  []string{}, // empty = all
		IgnoreLanguages:  []string{}, // empty = none
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	block := CodeBlock{Language: "go", Code: "package main"}
	assert.True(t, dfi.shouldFormat(block))
}

func TestDebateFormatterIntegration_shouldFormat_WhitelistFilter(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 50000,
		FormatLanguages:  []string{"go", "python"},
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	tests := []struct {
		language string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"Go", true},   // case insensitive
		{"PYTHON", true}, // case insensitive
		{"javascript", false},
		{"rust", false},
		{"java", false},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			block := CodeBlock{Language: tt.language, Code: "code"}
			assert.Equal(t, tt.expected, dfi.shouldFormat(block))
		})
	}
}

func TestDebateFormatterIntegration_shouldFormat_IgnoreList(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 50000,
		FormatLanguages:  []string{}, // all allowed
		IgnoreLanguages:  []string{"markdown", "text", "plaintext"},
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	tests := []struct {
		language string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"markdown", false},
		{"text", false},
		{"plaintext", false},
		{"Markdown", false}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			block := CodeBlock{Language: tt.language, Code: "code"}
			assert.Equal(t, tt.expected, dfi.shouldFormat(block))
		})
	}
}

// =============================================================================
// FormatDebateResponse Tests
// =============================================================================

func TestDebateFormatterIntegration_FormatDebateResponse_Disabled(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		Enabled:    false,
		AutoFormat: true,
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	response := "```go\npackage main\n```"
	result, err := dfi.FormatDebateResponse(context.Background(), response, "agent1", "session1")

	require.NoError(t, err)
	assert.Equal(t, response, result, "Disabled integration should return response unchanged")
}

func TestDebateFormatterIntegration_FormatDebateResponse_AutoFormatDisabled(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		Enabled:    true,
		AutoFormat: false,
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	response := "```go\npackage main\n```"
	result, err := dfi.FormatDebateResponse(context.Background(), response, "agent1", "session1")

	require.NoError(t, err)
	assert.Equal(t, response, result, "AutoFormat disabled should return response unchanged")
}

func TestDebateFormatterIntegration_FormatDebateResponse_NoCodeBlocks(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	response := "This is a plain text response with no code blocks."
	result, err := dfi.FormatDebateResponse(context.Background(), response, "agent1", "session1")

	require.NoError(t, err)
	assert.Equal(t, response, result)
}

func TestDebateFormatterIntegration_FormatDebateResponse_ContinueOnError(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	// Use a real executor with nil registry - it will fail on format
	executor := newTestFormatterExecutor()
	cfg := &DebateFormatterConfig{
		Enabled:          true,
		AutoFormat:       true,
		MaxCodeBlockSize: 50000,
		Timeout:          5 * time.Second,
		ContinueOnError:  true,
	}
	dfi := NewDebateFormatterIntegration(executor, cfg, logger)

	response := "Here is code:\n```go\npackage main\n```\nMore text."
	result, err := dfi.FormatDebateResponse(context.Background(), response, "agent1", "session1")

	require.NoError(t, err, "ContinueOnError should suppress formatting failures")
	assert.Equal(t, response, result, "On error with continue, original response is returned")
}

func TestDebateFormatterIntegration_FormatDebateResponse_ErrorWithoutContinue(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	executor := newTestFormatterExecutor()
	cfg := &DebateFormatterConfig{
		Enabled:          true,
		AutoFormat:       true,
		MaxCodeBlockSize: 50000,
		Timeout:          5 * time.Second,
		ContinueOnError:  false,
	}
	dfi := NewDebateFormatterIntegration(executor, cfg, logger)

	response := "Code:\n```go\npackage main\n```"
	_, err := dfi.FormatDebateResponse(context.Background(), response, "agent1", "session1")

	assert.Error(t, err, "Without ContinueOnError, formatting failure should return error")
}

func TestDebateFormatterIntegration_FormatDebateResponse_SkipsFilteredBlocks(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		Enabled:          true,
		AutoFormat:       true,
		FormatLanguages:  []string{"python"}, // Only format Python
		MaxCodeBlockSize: 50000,
		Timeout:          5 * time.Second,
		ContinueOnError:  true,
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	// Go block should be skipped, empty language block should be skipped
	response := "```go\npackage main\n```\n\n```\nplain\n```"
	result, err := dfi.FormatDebateResponse(context.Background(), response, "agent1", "session1")

	require.NoError(t, err)
	assert.Equal(t, response, result, "Filtered blocks should not be modified")
}

// =============================================================================
// CodeBlock Tests
// =============================================================================

func TestCodeBlock_Fields(t *testing.T) {
	block := CodeBlock{
		Original: "```go\nfunc main() {}\n```",
		Language: "go",
		Code:     "func main() {}",
		Start:    10,
		End:      40,
	}

	assert.Equal(t, "go", block.Language)
	assert.Equal(t, "func main() {}", block.Code)
	assert.Equal(t, 10, block.Start)
	assert.Equal(t, 40, block.End)
	assert.Contains(t, block.Original, "```go")
}

// =============================================================================
// DebateFormatterConfig Tests
// =============================================================================

func TestDebateFormatterConfig_Fields(t *testing.T) {
	cfg := DebateFormatterConfig{
		Enabled:          true,
		AutoFormat:       true,
		FormatLanguages:  []string{"go", "python", "rust"},
		IgnoreLanguages:  []string{"markdown"},
		MaxCodeBlockSize: 100000,
		Timeout:          60 * time.Second,
		ContinueOnError:  false,
	}

	assert.True(t, cfg.Enabled)
	assert.True(t, cfg.AutoFormat)
	assert.Len(t, cfg.FormatLanguages, 3)
	assert.Len(t, cfg.IgnoreLanguages, 1)
	assert.Equal(t, 100000, cfg.MaxCodeBlockSize)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.False(t, cfg.ContinueOnError)
}

// =============================================================================
// Integration with extractCodeBlocks and shouldFormat
// =============================================================================

func TestDebateFormatterIntegration_ExtractAndFilter_Combined(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 50000,
		FormatLanguages:  []string{"go"},
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	content := "```go\npackage main\n```\n```python\nimport os\n```\n```\nplain\n```"

	blocks := dfi.extractCodeBlocks(content)
	require.Len(t, blocks, 3)

	// Only Go should pass the filter
	assert.True(t, dfi.shouldFormat(blocks[0]))  // go
	assert.False(t, dfi.shouldFormat(blocks[1])) // python
	assert.False(t, dfi.shouldFormat(blocks[2])) // no language
}

func TestDebateFormatterIntegration_ExtractAndFilter_SizeLimit(t *testing.T) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 20,
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	content := "```go\nshort\n```\n```go\nthis is a much longer block of code that exceeds the limit\n```"

	blocks := dfi.extractCodeBlocks(content)
	require.Len(t, blocks, 2)

	assert.True(t, dfi.shouldFormat(blocks[0]))  // short, under limit
	assert.False(t, dfi.shouldFormat(blocks[1])) // exceeds limit
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkDebateFormatterIntegration_extractCodeBlocks(b *testing.B) {
	logger := newDebateFormatterTestLogger()
	dfi := NewDebateFormatterIntegration(nil, nil, logger)

	content := "Text\n```go\npackage main\n```\nMore text\n```python\nimport os\n```\nEnd"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dfi.extractCodeBlocks(content)
	}
}

func BenchmarkDebateFormatterIntegration_shouldFormat(b *testing.B) {
	logger := newDebateFormatterTestLogger()
	cfg := &DebateFormatterConfig{
		MaxCodeBlockSize: 50000,
		FormatLanguages:  []string{"go", "python", "javascript", "rust"},
	}
	dfi := NewDebateFormatterIntegration(nil, cfg, logger)

	block := CodeBlock{Language: "go", Code: "package main"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dfi.shouldFormat(block)
	}
}
