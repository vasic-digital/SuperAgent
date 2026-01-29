package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewBiomeFormatter creates a Biome formatter (35x faster than Prettier)
func NewBiomeFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "biome",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/biomejs/biome",
		Version:         "1.9.4",
		Languages:       []string{"javascript", "typescript", "json", "jsx", "tsx"},
		License:         "MIT",
		InstallMethod:   "npm",
		BinaryPath:      "biome",
		ConfigFormat:    "json",
		Performance:     "very_fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   true,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(metadata, "biome", []string{"format", "--stdin-file-path=temp.js"}, true, logger)
}
