package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewYamlfmtFormatter creates a yamlfmt YAML formatter
func NewYamlfmtFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "yamlfmt",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/google/yamlfmt",
		Version:         "0.14.0",
		Languages:       []string{"yaml", "yml"},
		License:         "Apache 2.0",
		InstallMethod:   "binary",
		BinaryPath:      "yamlfmt",
		ConfigFormat:    "yaml",
		Performance:     "fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   false,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(metadata, "yamlfmt", []string{"-in"}, true, logger)
}
