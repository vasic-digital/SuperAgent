package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewPrettierFormatter creates a Prettier formatter (JS/TS/HTML/CSS/etc.)
func NewPrettierFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "prettier",
		Type:            formatters.FormatterTypeUnified,
		Architecture:    "node",
		GitHubURL:       "https://github.com/prettier/prettier",
		Version:         "3.4.2",
		Languages:       []string{"javascript", "typescript", "json", "html", "css", "scss", "markdown", "yaml", "graphql"},
		License:         "MIT",
		InstallMethod:   "npm",
		BinaryPath:      "prettier",
		ConfigFormat:    "json",
		Performance:     "medium",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   true,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(
		metadata,
		"prettier",
		[]string{"--stdin-filepath", "temp.js"},
		true,
		logger,
	)
}
