package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// DebateFormatterIntegration integrates code formatters into AI Debate responses
type DebateFormatterIntegration struct {
	executor *formatters.FormatterExecutor
	config   *DebateFormatterConfig
	logger   *logrus.Logger
}

// DebateFormatterConfig configures formatter integration with debates
type DebateFormatterConfig struct {
	Enabled          bool
	AutoFormat       bool
	FormatLanguages  []string // Empty = all languages
	IgnoreLanguages  []string // Languages to skip
	MaxCodeBlockSize int      // Max bytes to format
	Timeout          time.Duration
	ContinueOnError  bool // Continue if formatting fails
}

// CodeBlock represents an extracted code block
type CodeBlock struct {
	Original string // Original markdown block
	Language string // Language identifier
	Code     string // Code content only
	Start    int    // Start position in original text
	End      int    // End position in original text
}

// NewDebateFormatterIntegration creates a new debate formatter integration
func NewDebateFormatterIntegration(
	executor *formatters.FormatterExecutor,
	config *DebateFormatterConfig,
	logger *logrus.Logger,
) *DebateFormatterIntegration {
	if config == nil {
		config = DefaultDebateFormatterConfig()
	}

	return &DebateFormatterIntegration{
		executor: executor,
		config:   config,
		logger:   logger,
	}
}

// DefaultDebateFormatterConfig returns default configuration
func DefaultDebateFormatterConfig() *DebateFormatterConfig {
	return &DebateFormatterConfig{
		Enabled:          true,
		AutoFormat:       true,
		FormatLanguages:  []string{}, // All languages
		IgnoreLanguages:  []string{}, // None ignored
		MaxCodeBlockSize: 50000,      // 50KB max
		Timeout:          30 * time.Second,
		ContinueOnError:  true,
	}
}

// FormatDebateResponse formats all code blocks in a debate response
func (d *DebateFormatterIntegration) FormatDebateResponse(
	ctx context.Context,
	response string,
	agentName string,
	sessionID string,
) (string, error) {
	if !d.config.Enabled || !d.config.AutoFormat {
		return response, nil
	}

	// Extract code blocks
	codeBlocks := d.extractCodeBlocks(response)
	if len(codeBlocks) == 0 {
		d.logger.Debug("No code blocks found in debate response")
		return response, nil
	}

	d.logger.Infof("Found %d code blocks to format", len(codeBlocks))

	// Format each code block
	formatted := response
	for i, block := range codeBlocks {
		if !d.shouldFormat(block) {
			d.logger.Debugf("Skipping code block %d (%s): filtered", i+1, block.Language)
			continue
		}

		formattedCode, err := d.formatCodeBlock(ctx, block, agentName, sessionID)
		if err != nil {
			if d.config.ContinueOnError {
				d.logger.Warnf("Failed to format code block %d: %v (continuing)", i+1, err)
				continue
			}
			return "", fmt.Errorf("failed to format code block %d: %w", i+1, err)
		}

		// Replace original block with formatted block
		formatted = strings.Replace(formatted, block.Original, formattedCode, 1)
		d.logger.Debugf("Formatted code block %d (%s)", i+1, block.Language)
	}

	return formatted, nil
}

// extractCodeBlocks extracts all code blocks from markdown
func (d *DebateFormatterIntegration) extractCodeBlocks(content string) []CodeBlock {
	// Regex to match markdown code blocks: ```language\ncode\n```
	re := regexp.MustCompile("(?s)```([a-z0-9+#-]*)\n(.*?)\n```")
	matches := re.FindAllStringSubmatchIndex(content, -1)

	blocks := make([]CodeBlock, 0, len(matches))

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		language := content[match[2]:match[3]]
		code := content[match[4]:match[5]]
		original := content[match[0]:match[1]]

		blocks = append(blocks, CodeBlock{
			Original: original,
			Language: language,
			Code:     code,
			Start:    match[0],
			End:      match[1],
		})
	}

	return blocks
}

// formatCodeBlock formats a single code block
func (d *DebateFormatterIntegration) formatCodeBlock(
	ctx context.Context,
	block CodeBlock,
	agentName string,
	sessionID string,
) (string, error) {
	req := &formatters.FormatRequest{
		Content:   block.Code,
		Language:  block.Language,
		Timeout:   d.config.Timeout,
		AgentName: agentName,
		SessionID: sessionID,
		CheckOnly: false,
	}

	result, err := d.executor.Execute(ctx, req)
	if err != nil {
		return "", fmt.Errorf("executor failed: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("formatting failed: %v", result.Error)
	}

	// Reconstruct markdown code block with formatted code
	return fmt.Sprintf("```%s\n%s\n```", block.Language, result.Content), nil
}

// shouldFormat determines if a code block should be formatted
func (d *DebateFormatterIntegration) shouldFormat(block CodeBlock) bool {
	// Check size limit
	if len(block.Code) > d.config.MaxCodeBlockSize {
		d.logger.Debugf("Code block too large: %d bytes (max %d)", len(block.Code), d.config.MaxCodeBlockSize)
		return false
	}

	// Check if language is empty
	if block.Language == "" {
		return false
	}

	// Normalize language
	language := strings.ToLower(block.Language)

	// Check explicit language filters
	if len(d.config.FormatLanguages) > 0 {
		for _, lang := range d.config.FormatLanguages {
			if strings.ToLower(lang) == language {
				return true
			}
		}
		return false
	}

	// Check ignore list
	if len(d.config.IgnoreLanguages) > 0 {
		for _, lang := range d.config.IgnoreLanguages {
			if strings.ToLower(lang) == language {
				return false
			}
		}
	}

	return true
}

// GetConfig returns the current configuration
func (d *DebateFormatterIntegration) GetConfig() *DebateFormatterConfig {
	return d.config
}

// SetConfig updates the configuration
func (d *DebateFormatterIntegration) SetConfig(config *DebateFormatterConfig) {
	d.config = config
}
