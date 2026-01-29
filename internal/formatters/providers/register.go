package providers

import (
	"dev.helix.agent/internal/formatters"
	"dev.helix.agent/internal/formatters/providers/native"
	"github.com/sirupsen/logrus"
)

// RegisterAllFormatters registers all available formatters with the registry
func RegisterAllFormatters(registry *formatters.FormatterRegistry, logger *logrus.Logger) error {
	// Track registration stats
	registered := 0
	failed := 0

	// Helper function to register and track
	register := func(name string, formatter formatters.Formatter, metadata *formatters.FormatterMetadata) {
		if err := registry.Register(formatter, metadata); err != nil {
			logger.Warnf("Failed to register %s: %v", name, err)
			failed++
		} else {
			registered++
		}
	}

	// Python formatters
	register("black", native.NewBlackFormatter(logger), &formatters.FormatterMetadata{
		Name: "black", Type: formatters.FormatterTypeNative, Version: "26.1a1",
		Languages: []string{"python"}, Performance: "medium",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: true, SupportsConfig: true,
	})

	register("ruff", native.NewRuffFormatter(logger), &formatters.FormatterMetadata{
		Name: "ruff", Type: formatters.FormatterTypeNative, Version: "0.9.6",
		Languages: []string{"python"}, Performance: "very_fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: true, SupportsConfig: true,
	})

	// JavaScript/TypeScript formatters
	register("prettier", native.NewPrettierFormatter(logger), &formatters.FormatterMetadata{
		Name: "prettier", Type: formatters.FormatterTypeUnified, Version: "3.4.2",
		Languages: []string{"javascript", "typescript", "json", "html", "css", "scss", "markdown", "yaml", "graphql"},
		Performance: "medium",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: true, SupportsConfig: true,
	})

	register("biome", native.NewBiomeFormatter(logger), &formatters.FormatterMetadata{
		Name: "biome", Type: formatters.FormatterTypeNative, Version: "1.9.4",
		Languages: []string{"javascript", "typescript", "json", "jsx", "tsx"},
		Performance: "very_fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: true, SupportsConfig: true,
	})

	// Systems languages
	register("rustfmt", native.NewRustfmtFormatter(logger), &formatters.FormatterMetadata{
		Name: "rustfmt", Type: formatters.FormatterTypeNative, Version: "1.8.1",
		Languages: []string{"rust"}, Performance: "fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: true, SupportsConfig: true,
	})

	register("gofmt", native.NewGofmtFormatter(logger), &formatters.FormatterMetadata{
		Name: "gofmt", Type: formatters.FormatterTypeBuiltin, Version: "go1.24.11",
		Languages: []string{"go"}, Performance: "fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: false, SupportsConfig: false,
	})

	register("clang-format", native.NewClangFormatFormatter(logger), &formatters.FormatterMetadata{
		Name: "clang-format", Type: formatters.FormatterTypeNative, Version: "19.1.8",
		Languages: []string{"c", "cpp", "java", "javascript", "objectivec", "protobuf"},
		Performance: "fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: false, SupportsConfig: true,
	})

	// Scripting languages
	register("shfmt", native.NewShfmtFormatter(logger), &formatters.FormatterMetadata{
		Name: "shfmt", Type: formatters.FormatterTypeNative, Version: "3.10.0",
		Languages: []string{"bash", "sh", "shell"}, Performance: "fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: false, SupportsConfig: true,
	})

	register("stylua", native.NewStyluaFormatter(logger), &formatters.FormatterMetadata{
		Name: "stylua", Type: formatters.FormatterTypeNative, Version: "2.0.2",
		Languages: []string{"lua"}, Performance: "fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: true, SupportsConfig: true,
	})

	// Data formats
	register("yamlfmt", native.NewYamlfmtFormatter(logger), &formatters.FormatterMetadata{
		Name: "yamlfmt", Type: formatters.FormatterTypeNative, Version: "0.14.0",
		Languages: []string{"yaml", "yml"}, Performance: "fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: false, SupportsConfig: true,
	})

	register("taplo", native.NewTaploFormatter(logger), &formatters.FormatterMetadata{
		Name: "taplo", Type: formatters.FormatterTypeNative, Version: "0.9.3",
		Languages: []string{"toml"}, Performance: "fast",
		SupportsStdin: true, SupportsInPlace: true, SupportsCheck: false, SupportsConfig: true,
	})

	logger.Infof("Formatter registration complete: %d registered, %d failed", registered, failed)

	return nil
}
