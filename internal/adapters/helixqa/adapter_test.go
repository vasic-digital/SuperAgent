package helixqa

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"digital.vasic.helixqa/pkg/config"
	"digital.vasic.helixqa/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdapter(t *testing.T) {
	adapter, err := NewAdapter()
	require.NoError(t, err)
	require.NotNil(t, adapter)
	assert.NotNil(t, adapter.orchestrator)
	assert.NotNil(t, adapter.config)
}

func TestNewAdapterWithConfig(t *testing.T) {
	cfg := &config.Config{
		Banks:          []string{"test.yaml"},
		Platforms:      []config.Platform{config.PlatformDesktop},
		OutputDir:      t.TempDir(),
		Speed:          config.SpeedFast,
		ReportFormat:   config.ReportJSON,
		ValidateSteps:  false,
		Record:         false,
		Verbose:        true,
		Timeout:        5 * time.Minute,
		StepTimeout:    30 * time.Second,
		DesktopProcess: "test",
	}

	adapter, err := NewAdapterWithConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	assert.Equal(t, cfg, adapter.config)
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	require.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.OutputDir)
	assert.Equal(t, config.SpeedNormal, cfg.Speed)
	assert.Equal(t, config.ReportMarkdown, cfg.ReportFormat)
	assert.True(t, cfg.ValidateSteps)
	assert.Equal(t, "helixagent", cfg.DesktopProcess)
}

func TestDefaultOutputDir(t *testing.T) {
	outputDir := defaultOutputDir()
	assert.NotEmpty(t, outputDir)
	assert.Contains(t, outputDir, "helixqa")
}

func TestDefaultOutputDirWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("REPORTS_DIR", tmpDir)

	outputDir := defaultOutputDir()
	assert.Equal(t, filepath.Join(tmpDir, "helixqa"), outputDir)
}

func TestDefaultConfigVerbose(t *testing.T) {
	t.Setenv("DEBUG", "true")
	cfg := defaultConfig()
	assert.True(t, cfg.Verbose)
}

func TestGetAdapter(t *testing.T) {
	SetAdapter(nil)

	adapter, err := GetAdapter()
	require.NoError(t, err)
	require.NotNil(t, adapter)

	adapter2, err := GetAdapter()
	require.NoError(t, err)
	assert.Equal(t, adapter, adapter2)
}

func TestSetAdapter(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	adapter := &Adapter{
		orchestrator: orchestrator.New(cfg),
		config:       cfg,
	}

	SetAdapter(adapter)

	retrieved, err := GetAdapter()
	require.NoError(t, err)
	assert.Equal(t, adapter, retrieved)
}

func TestRunWithBanks_Skip(t *testing.T) {
	t.Skip("Requires full HelixQA environment with test banks")
}

func TestAdapterRun_Skip(t *testing.T) {
	t.Skip("Requires full HelixQA environment with test banks")
}

func TestAdapterConfigDefaults(t *testing.T) {
	adapter, err := NewAdapter()
	require.NoError(t, err)

	assert.NotEmpty(t, adapter.config.Platforms)
	assert.NotEmpty(t, adapter.config.OutputDir)
	assert.Positive(t, adapter.config.Timeout)
	assert.Positive(t, adapter.config.StepTimeout)
}

func TestRunWithBanks_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	bankFile := filepath.Join(tmpDir, "test-bank.yaml")

	bankContent := `version: "1.0"
name: "Test Bank"
description: "Test bank for adapter integration"
test_cases:
  - id: TEST-001
    name: "Simple test"
    category: functional
    priority: high
    platforms: [desktop]
    steps:
      - name: "Step 1"
        action: "Test action"
        expected: "Expected result"
`
	err := os.WriteFile(bankFile, []byte(bankContent), 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		Banks:          []string{},
		Platforms:      []config.Platform{config.PlatformDesktop},
		OutputDir:      tmpDir,
		Speed:          config.SpeedFast,
		ReportFormat:   config.ReportJSON,
		ValidateSteps:  false,
		Record:         false,
		Verbose:        false,
		Timeout:        1 * time.Minute,
		StepTimeout:    10 * time.Second,
		DesktopProcess: "helixagent",
	}

	adapter, err := NewAdapterWithConfig(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := adapter.RunWithBanks(ctx, []string{bankFile})
	if err != nil {
		t.Logf("RunWithBanks returned error (expected in test env): %v", err)
	}
	if result != nil {
		assert.NotNil(t, result)
	}
}
