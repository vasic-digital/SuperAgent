package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewShfmtFormatter creates a shfmt shell script formatter
func NewShfmtFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "shfmt",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/mvdan/sh",
		Version:         "3.10.0",
		Languages:       []string{"bash", "sh", "shell"},
		License:         "BSD-3-Clause",
		InstallMethod:   "binary",
		BinaryPath:      "shfmt",
		ConfigFormat:    "editorconfig",
		Performance:     "fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   false,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(metadata, "shfmt", []string{"-i", "2"}, true, logger)
}
