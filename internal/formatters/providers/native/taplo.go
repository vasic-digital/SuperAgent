package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewTaploFormatter creates a Taplo TOML formatter
func NewTaploFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "taplo",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/tamasfe/taplo",
		Version:         "0.9.3",
		Languages:       []string{"toml"},
		License:         "MIT",
		InstallMethod:   "cargo",
		BinaryPath:      "taplo",
		ConfigFormat:    "toml",
		Performance:     "fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   false,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(metadata, "taplo", []string{"fmt", "-"}, true, logger)
}
