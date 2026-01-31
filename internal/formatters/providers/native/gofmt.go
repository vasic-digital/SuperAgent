package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewGofmtFormatter creates a gofmt Go formatter
func NewGofmtFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "gofmt",
		Type:            formatters.FormatterTypeBuiltin,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/golang/go",
		Version:         "go1.24.11",
		Languages:       []string{"go"},
		License:         "BSD-3-Clause",
		InstallMethod:   "builtin",
		BinaryPath:      "gofmt",
		ConfigFormat:    "none",
		Performance:     "fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   false,
		SupportsConfig:  false,
	}

	return NewNativeFormatter(
		metadata,
		"gofmt",
		[]string{}, // no args needed
		true,
		logger,
	)
}
