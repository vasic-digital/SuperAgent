package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewBlackFormatter creates a Black Python formatter
func NewBlackFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "black",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "python",
		GitHubURL:       "https://github.com/psf/black",
		Version:         "26.1a1",
		Languages:       []string{"python"},
		License:         "MIT",
		InstallMethod:   "pip",
		BinaryPath:      "black",
		ConfigFormat:    "toml",
		Performance:     "medium",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   true,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(
		metadata,
		"black",                    // binary name
		[]string{"--quiet"},        // default args
		true,                       // supports stdin
		logger,
	)
}
