package service

import (
	"fmt"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewAutopep8Formatter creates a new autopep8 service formatter
func NewAutopep8Formatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9211", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:            "autopep8",
		Type:            formatters.FormatterTypeService,
		Architecture:    "python",
		GitHubURL:       "https://github.com/hhatto/autopep8",
		Version:         "2.0.4",
		Languages:       []string{"python"},
		License:         "MIT",
		InstallMethod:   "pip",
		ServiceURL:      serviceURL,
		ConfigFormat:    "ini",
		Performance:     "medium",
		Complexity:      "easy",
		SupportsStdin:   true,
		SupportsInPlace: false,
		SupportsCheck:   false,
		SupportsConfig:  true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}

// NewYapfFormatter creates a new yapf service formatter
func NewYapfFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9210", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:            "yapf",
		Type:            formatters.FormatterTypeService,
		Architecture:    "python",
		GitHubURL:       "https://github.com/google/yapf",
		Version:         "0.40.2",
		Languages:       []string{"python"},
		License:         "Apache-2.0",
		InstallMethod:   "pip",
		ServiceURL:      serviceURL,
		ConfigFormat:    "ini",
		Performance:     "slow",
		Complexity:      "medium",
		SupportsStdin:   true,
		SupportsInPlace: false,
		SupportsCheck:   false,
		SupportsConfig:  true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}
