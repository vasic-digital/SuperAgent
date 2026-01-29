package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewStyluaFormatter creates a StyLua Lua formatter
func NewStyluaFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "stylua",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/JohnnyMorganz/StyLua",
		Version:         "2.0.2",
		Languages:       []string{"lua"},
		License:         "MPL 2.0",
		InstallMethod:   "cargo",
		BinaryPath:      "stylua",
		ConfigFormat:    "toml",
		Performance:     "fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   true,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(metadata, "stylua", []string{"-"}, true, logger)
}
