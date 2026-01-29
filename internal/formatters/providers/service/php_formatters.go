package service

import (
	"fmt"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewPHPCSFixerFormatter creates a new php-cs-fixer service formatter
func NewPHPCSFixerFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9240", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "php-cs-fixer",
		Type:          formatters.FormatterTypeService,
		Architecture:  "php",
		GitHubURL:     "https://github.com/PHP-CS-Fixer/PHP-CS-Fixer",
		Version:       "3.68.0",
		Languages:     []string{"php"},
		License:       "MIT",
		InstallMethod: "composer",
		ServiceURL:    serviceURL,
		ConfigFormat:  "php",
		Performance:   "medium",
		Complexity:    "medium",
		SupportsStdin:  false,
		SupportsInPlace: true,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewLaravelPintFormatter creates a new laravel-pint service formatter
func NewLaravelPintFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9241", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "laravel-pint",
		Type:          formatters.FormatterTypeService,
		Architecture:  "php",
		GitHubURL:     "https://github.com/laravel/pint",
		Version:       "1.19.0",
		Languages:     []string{"php"},
		License:       "MIT",
		InstallMethod: "composer",
		ServiceURL:    serviceURL,
		ConfigFormat:  "json",
		Performance:   "fast",
		Complexity:    "easy",
		SupportsStdin:  false,
		SupportsInPlace: true,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}
