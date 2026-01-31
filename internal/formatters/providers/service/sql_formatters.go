package service

import (
	"fmt"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// NewSQLFluffFormatter creates a new sqlfluff service formatter
func NewSQLFluffFormatter(baseURL string, logger *logrus.Logger) *ServiceFormatter {
	serviceURL := fmt.Sprintf("%s:9220", baseURL)

	metadata := &formatters.FormatterMetadata{
		Name:            "sqlfluff",
		Type:            formatters.FormatterTypeService,
		Architecture:    "python",
		GitHubURL:       "https://github.com/sqlfluff/sqlfluff",
		Version:         "3.4.1",
		Languages:       []string{"sql"},
		License:         "MIT",
		InstallMethod:   "pip",
		ServiceURL:      serviceURL,
		ConfigFormat:    "toml",
		Performance:     "medium",
		Complexity:      "medium",
		SupportsStdin:   true,
		SupportsInPlace: false,
		SupportsCheck:   true,
		SupportsConfig:  true,
	}

	return NewServiceFormatter(metadata, serviceURL, 30*time.Second, logger)
}
