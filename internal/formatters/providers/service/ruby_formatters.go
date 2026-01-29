package service

import (
	"fmt"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewRubocopFormatter creates a new rubocop service formatter
func NewRubocopFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9230", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "rubocop",
		Type:          formatters.FormatterTypeService,
		Architecture:  "ruby",
		GitHubURL:     "https://github.com/rubocop/rubocop",
		Version:       "1.72.0",
		Languages:     []string{"ruby"},
		License:       "MIT",
		InstallMethod: "gem",
		ServiceURL:    serviceURL,
		ConfigFormat:  "yaml",
		Performance:   "medium",
		Complexity:    "medium",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewStandardRBFormatter creates a new standardrb service formatter
func NewStandardRBFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9231", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:          "standardrb",
		Type:          formatters.FormatterTypeService,
		Architecture:  "ruby",
		GitHubURL:     "https://github.com/standardrb/standard",
		Version:       "1.42.1",
		Languages:     []string{"ruby"},
		License:       "MIT",
		InstallMethod: "gem",
		ServiceURL:    serviceURL,
		ConfigFormat:  "yaml",
		Performance:   "fast",
		Complexity:    "easy",
		SupportsStdin:  true,
		SupportsInPlace: false,
		SupportsCheck:  true,
		SupportsConfig: true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}
