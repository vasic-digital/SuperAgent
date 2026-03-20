// Package helixqa provides an adapter layer between HelixAgent's
// testing infrastructure and the HelixQA module.
//
// This adapter enables comprehensive QA testing of HelixAgent services
// and applications using the HelixQA orchestration framework with
// crash detection, step validation, and multi-platform support.
package helixqa

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"digital.vasic.helixqa/pkg/config"
	"digital.vasic.helixqa/pkg/orchestrator"
)

// Adapter bridges HelixAgent's testing needs with HelixQA.
type Adapter struct {
	orchestrator *orchestrator.Orchestrator
	config       *config.Config
}

// NewAdapter creates a new HelixQA adapter with default configuration.
func NewAdapter() (*Adapter, error) {
	cfg := defaultConfig()
	orch := orchestrator.New(cfg)
	return &Adapter{
		orchestrator: orch,
		config:       cfg,
	}, nil
}

// NewAdapterWithConfig creates a new HelixQA adapter with custom configuration.
func NewAdapterWithConfig(cfg *config.Config) (*Adapter, error) {
	orch := orchestrator.New(cfg)
	return &Adapter{
		orchestrator: orch,
		config:       cfg,
	}, nil
}

// Run executes HelixQA tests with the configured test banks.
func (a *Adapter) Run(ctx context.Context) (*orchestrator.Result, error) {
	return a.orchestrator.Run(ctx)
}

// RunWithBanks executes HelixQA tests with specific test bank paths.
func (a *Adapter) RunWithBanks(ctx context.Context, banks []string) (*orchestrator.Result, error) {
	cfg := *a.config
	cfg.Banks = banks
	orch := orchestrator.New(&cfg)
	return orch.Run(ctx)
}

// defaultConfig returns a default HelixQA configuration for HelixAgent.
func defaultConfig() *config.Config {
	verbose := false
	if os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1" {
		verbose = true
	}
	return &config.Config{
		Banks:          []string{},
		Platforms:      []config.Platform{config.PlatformAll},
		Device:         "",
		PackageName:    "",
		OutputDir:      defaultOutputDir(),
		Speed:          config.SpeedNormal,
		ReportFormat:   config.ReportMarkdown,
		ValidateSteps:  true,
		Record:         false,
		Verbose:        verbose,
		Timeout:        30 * time.Minute,
		StepTimeout:    2 * time.Minute,
		BrowserURL:     "",
		DesktopProcess: "helixagent",
	}
}

// defaultOutputDir returns the default output directory for HelixQA reports.
func defaultOutputDir() string {
	// Try to use configured reports directory, fallback to ./reports/helixqa
	if reportsDir := os.Getenv("REPORTS_DIR"); reportsDir != "" {
		return filepath.Join(reportsDir, "helixqa")
	}
	// Fallback to current directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return filepath.Join(cwd, "reports", "helixqa")
}

// globalAdapter is the singleton HelixQA adapter instance.
var globalAdapter *Adapter

// GetAdapter returns the global HelixQA adapter, initializing it if needed.
func GetAdapter() (*Adapter, error) {
	if globalAdapter == nil {
		adapter, err := NewAdapter()
		if err != nil {
			return nil, err
		}
		globalAdapter = adapter
	}
	return globalAdapter, nil
}

// SetAdapter sets the global HelixQA adapter (for testing).
func SetAdapter(adapter *Adapter) {
	globalAdapter = adapter
}
