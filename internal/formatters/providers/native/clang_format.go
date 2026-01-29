package native

import (
	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewClangFormatFormatter creates a clang-format C/C++ formatter
func NewClangFormatFormatter(logger *logrus.Logger) *NativeFormatter {
	metadata := &formatters.FormatterMetadata{
		Name:            "clang-format",
		Type:            formatters.FormatterTypeNative,
		Architecture:    "binary",
		GitHubURL:       "https://github.com/llvm/llvm-project",
		Version:         "19.1.8",
		Languages:       []string{"c", "cpp", "java", "javascript", "objectivec", "protobuf"},
		License:         "Apache 2.0",
		InstallMethod:   "apt/brew",
		BinaryPath:      "clang-format",
		ConfigFormat:    "yaml",
		Performance:     "fast",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: true,
		SupportsCheck:   false,
		SupportsConfig:  true,
	}

	return NewNativeFormatter(metadata, "clang-format", []string{"-style=Google"}, true, logger)
}
