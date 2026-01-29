package service

import (
	"fmt"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewPerltidyFormatter creates a new perltidy service formatter
func NewPerltidyFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9250", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "perltidy",
		Type:          formatters.FormatterTypeService,
		Architecture:  "perl",
		GitHubURL:     "https://github.com/perltidy/perltidy",
		Version:       "20260109.01",
		Languages:     []string{"perl"},
		License:       "GPL-2.0",
		InstallMethod: "cpan",
		ServiceURL:    serviceURL,
		ConfigFormat:  "ini",
		Performance:   "fast",
		Complexity:    "easy",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  false,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewCljfmtFormatter creates a new cljfmt service formatter
func NewCljfmtFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9260", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "cljfmt",
		Type:          formatters.FormatterTypeService,
		Architecture:  "clojure",
		GitHubURL:     "https://github.com/weavejester/cljfmt",
		Version:       "0.12.0",
		Languages:     []string{"clojure"},
		License:       "EPL-1.0",
		InstallMethod: "clojure",
		ServiceURL:    serviceURL,
		ConfigFormat:  "edn",
		Performance:   "medium",
		Complexity:    "easy",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewSpotlessFormatter creates a new spotless service formatter
func NewSpotlessFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9270", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "spotless",
		Type:          formatters.FormatterTypeService,
		Architecture:  "jvm",
		GitHubURL:     "https://github.com/diffplug/spotless",
		Version:       "7.0.0.BETA4",
		Languages:     []string{"java", "kotlin", "scala", "groovy"},
		License:       "Apache-2.0",
		InstallMethod: "gradle",
		ServiceURL:    serviceURL,
		ConfigFormat:  "groovy",
		Performance:   "medium",
		Complexity:    "medium",
		SupportsStdin:  false,
		SupportsInPlace: true,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewGroovyLintFormatter creates a new npm-groovy-lint service formatter
func NewGroovyLintFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9280", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "npm-groovy-lint",
		Type:          formatters.FormatterTypeService,
		Architecture:  "node",
		GitHubURL:     "https://github.com/nvuillam/npm-groovy-lint",
		Version:       "15.0.4",
		Languages:     []string{"groovy"},
		License:       "GPL-3.0",
		InstallMethod: "npm",
		ServiceURL:    serviceURL,
		ConfigFormat:  "json",
		Performance:   "medium",
		Complexity:    "medium",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewStylerFormatter creates a new styler service formatter
func NewStylerFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9290", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "styler",
		Type:          formatters.FormatterTypeService,
		Architecture:  "r",
		GitHubURL:     "https://github.com/r-lib/styler",
		Version:       "1.10.3",
		Languages:     []string{"r"},
		License:       "MIT",
		InstallMethod: "cran",
		ServiceURL:    serviceURL,
		ConfigFormat:  "yaml",
		Performance:   "slow",
		Complexity:    "easy",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  false,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewAirFormatter creates a new air service formatter
func NewAirFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9291", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "air",
		Type:          formatters.FormatterTypeService,
		Architecture:  "r",
		GitHubURL:     "https://github.com/r-lib/air",
		Version:       "0.2.0",
		Languages:     []string{"r"},
		License:       "MIT",
		InstallMethod: "cran",
		ServiceURL:    serviceURL,
		ConfigFormat:  "yaml",
		Performance:   "very_fast",
		Complexity:    "easy",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  false,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewPSScriptAnalyzerFormatter creates a new PSScriptAnalyzer service formatter
func NewPSScriptAnalyzerFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9300", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "psscriptanalyzer",
		Type:          formatters.FormatterTypeService,
		Architecture:  "powershell",
		GitHubURL:     "https://github.com/PowerShell/PSScriptAnalyzer",
		Version:       "1.23.0",
		Languages:     []string{"powershell"},
		License:       "MIT",
		InstallMethod: "powershell",
		ServiceURL:    serviceURL,
		ConfigFormat:  "psd1",
		Performance:   "medium",
		Complexity:    "medium",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}
