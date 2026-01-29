package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewRuffFormatter creates a Ruff Python formatter (30x faster than Black)
func NewRuffFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "ruff",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/astral-sh/ruff",
		Version:         "0.9.6",
		Languages:       []string{"python"},
		License:         "MIT",
		InstallMethod:   "pip",
		BinaryPath:      "ruff",
		ConfigFormat:    "toml",
		Performance:     "very_fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   true,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(
		metadata,
		"ruff",
		[]string{"format", "--silent"},
		true,
		logger,
	)
}
