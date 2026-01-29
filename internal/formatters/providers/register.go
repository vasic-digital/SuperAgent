package providers

import (
	"os"

	"dev.helix.agent/internal/formatters"
	"dev.helix.agent/internal/formatters/providers/native"
	"dev.helix.agent/internal/formatters/providers/service"
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

	// Service formatters (optional, requires Docker services running)
	serviceBaseURL := os.Getenv("FORMATTER_SERVICE_BASE_URL")
	if serviceBaseURL == "" {
		serviceBaseURL = "http://localhost"
	}

	enableServiceFormatters := os.Getenv("FORMATTER_ENABLE_SERVICES")
	if enableServiceFormatters == "true" || enableServiceFormatters == "1" {
		logger.Info("Registering service formatters...")

		// Python service formatters
		register("autopep8", service.NewAutopep8Formatter(serviceBaseURL, logger),
			service.NewAutopep8Formatter(serviceBaseURL, logger).GetMetadata())

		register("yapf", service.NewYapfFormatter(serviceBaseURL, logger),
			service.NewYapfFormatter(serviceBaseURL, logger).GetMetadata())

		// SQL formatter
		register("sqlfluff", service.NewSQLFluffFormatter(serviceBaseURL, logger),
			service.NewSQLFluffFormatter(serviceBaseURL, logger).GetMetadata())

		// Ruby formatters
		register("rubocop", service.NewRubocopFormatter(serviceBaseURL, logger),
			service.NewRubocopFormatter(serviceBaseURL, logger).GetMetadata())

		register("standardrb", service.NewStandardRBFormatter(serviceBaseURL, logger),
			service.NewStandardRBFormatter(serviceBaseURL, logger).GetMetadata())

		// PHP formatters
		register("php-cs-fixer", service.NewPHPCSFixerFormatter(serviceBaseURL, logger),
			service.NewPHPCSFixerFormatter(serviceBaseURL, logger).GetMetadata())

		register("laravel-pint", service.NewLaravelPintFormatter(serviceBaseURL, logger),
			service.NewLaravelPintFormatter(serviceBaseURL, logger).GetMetadata())

		// Other languages
		register("perltidy", service.NewPerltidyFormatter(serviceBaseURL, logger),
			service.NewPerltidyFormatter(serviceBaseURL, logger).GetMetadata())

		register("cljfmt", service.NewCljfmtFormatter(serviceBaseURL, logger),
			service.NewCljfmtFormatter(serviceBaseURL, logger).GetMetadata())

		register("spotless", service.NewSpotlessFormatter(serviceBaseURL, logger),
			service.NewSpotlessFormatter(serviceBaseURL, logger).GetMetadata())

		register("npm-groovy-lint", service.NewGroovyLintFormatter(serviceBaseURL, logger),
			service.NewGroovyLintFormatter(serviceBaseURL, logger).GetMetadata())

		register("styler", service.NewStylerFormatter(serviceBaseURL, logger),
			service.NewStylerFormatter(serviceBaseURL, logger).GetMetadata())

		register("air", service.NewAirFormatter(serviceBaseURL, logger),
			service.NewAirFormatter(serviceBaseURL, logger).GetMetadata())

		register("psscriptanalyzer", service.NewPSScriptAnalyzerFormatter(serviceBaseURL, logger),
			service.NewPSScriptAnalyzerFormatter(serviceBaseURL, logger).GetMetadata())

		logger.Infof("Service formatters registration complete")
	} else {
		logger.Info("Service formatters disabled (set FORMATTER_ENABLE_SERVICES=true to enable)")
	}

	logger.Infof("Formatter registration complete: %d registered, %d failed", registered, failed)

	return nil
}
