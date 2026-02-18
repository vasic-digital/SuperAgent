package challenges

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"digital.vasic.challenges/pkg/registry"
)

func TestNewOrchestrator_Defaults(t *testing.T) {
	o := NewOrchestrator(OrchestratorConfig{})
	assert.NotNil(t, o.registry)
	assert.NotNil(t, o.runner)
	assert.NotNil(t, o.collector)
	assert.NotNil(t, o.reporter)
	assert.Equal(t, "challenge-results", o.config.ResultsDir)
	assert.Equal(t, 2, o.config.MaxConcurrency)
	assert.Equal(t, 10*time.Minute, o.config.Timeout)
}

func TestNewOrchestrator_CustomConfig(t *testing.T) {
	o := NewOrchestrator(OrchestratorConfig{
		ResultsDir:     "/tmp/results",
		MaxConcurrency: 4,
		StallThreshold: 30 * time.Second,
		Timeout:        5 * time.Minute,
	})
	assert.Equal(t, "/tmp/results", o.config.ResultsDir)
	assert.Equal(t, 4, o.config.MaxConcurrency)
	assert.Equal(t, 5*time.Minute, o.config.Timeout)
}

func TestOrchestrator_RegisterAll_NoScriptsDir(t *testing.T) {
	o := NewOrchestrator(OrchestratorConfig{})
	err := o.RegisterAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestOrchestrator_RegisterAll_WithScripts(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "challenges", "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))

	scripts := []string{
		"provider_test_challenge.sh",
		"security_scan_challenge.sh",
		"not_a_challenge.txt",
	}
	for _, s := range scripts {
		require.NoError(t, os.WriteFile(
			filepath.Join(scriptsDir, s),
			[]byte("#!/bin/bash\necho ok\n"), 0755,
		))
	}

	o := NewOrchestrator(OrchestratorConfig{
		ProjectRoot: dir,
	})
	err := o.RegisterAll()
	require.NoError(t, err)

	list := o.List()
	assert.Len(t, list, 2)
}

func TestOrchestrator_List_Empty(t *testing.T) {
	o := NewOrchestrator(OrchestratorConfig{})
	list := o.List()
	assert.Empty(t, list)
}

func TestOrchestrator_Run_Empty(t *testing.T) {
	o := NewOrchestrator(OrchestratorConfig{
		ResultsDir: t.TempDir(),
	})
	result, err := o.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
}

func TestOrchestrator_Run_WithScripts(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "challenges", "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))

	require.NoError(t, os.WriteFile(
		filepath.Join(scriptsDir, "pass_challenge.sh"),
		[]byte("#!/bin/bash\necho 'ok'\nexit 0\n"), 0755,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(scriptsDir, "fail_challenge.sh"),
		[]byte("#!/bin/bash\necho 'fail'\nexit 1\n"), 0755,
	))

	o := NewOrchestrator(OrchestratorConfig{
		ProjectRoot: dir,
		ResultsDir:  filepath.Join(dir, "results"),
		Timeout:     10 * time.Second,
	})
	require.NoError(t, o.RegisterAll())

	result, err := o.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 1, result.Failed)
}

func TestOrchestrator_RunSingle(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "challenges", "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))

	require.NoError(t, os.WriteFile(
		filepath.Join(scriptsDir, "single_challenge.sh"),
		[]byte("#!/bin/bash\necho 'single'\nexit 0\n"), 0755,
	))

	o := NewOrchestrator(OrchestratorConfig{
		ProjectRoot: dir,
		ResultsDir:  filepath.Join(dir, "results"),
		Timeout:     10 * time.Second,
	})
	require.NoError(t, o.RegisterAll())

	result, err := o.RunSingle(
		context.Background(), "single-challenge",
	)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Passed)
}

func TestOrchestrator_RunSingle_NotFound(t *testing.T) {
	o := NewOrchestrator(OrchestratorConfig{
		ResultsDir: t.TempDir(),
	})
	_, err := o.RunSingle(
		context.Background(), "nonexistent",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOrchestrator_Filter(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "challenges", "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))

	for _, name := range []string{
		"alpha_challenge.sh",
		"beta_challenge.sh",
		"gamma_challenge.sh",
	} {
		require.NoError(t, os.WriteFile(
			filepath.Join(scriptsDir, name),
			[]byte("#!/bin/bash\necho ok\nexit 0\n"), 0755,
		))
	}

	o := NewOrchestrator(OrchestratorConfig{
		ProjectRoot: dir,
		ResultsDir:  filepath.Join(dir, "results"),
		Filter:      []string{"alpha-challenge"},
		Timeout:     10 * time.Second,
	})
	require.NoError(t, o.RegisterAll())

	result, err := o.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}

func TestOrchestrator_CategoryFilter(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "challenges", "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))

	for _, name := range []string{
		"provider_test_challenge.sh",
		"security_scan_challenge.sh",
	} {
		require.NoError(t, os.WriteFile(
			filepath.Join(scriptsDir, name),
			[]byte("#!/bin/bash\necho ok\nexit 0\n"), 0755,
		))
	}

	o := NewOrchestrator(OrchestratorConfig{
		ProjectRoot: dir,
		ResultsDir:  filepath.Join(dir, "results"),
		Category:    "provider",
		Timeout:     10 * time.Second,
	})
	require.NoError(t, o.RegisterAll())

	result, err := o.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}

func TestDetectCategory(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"provider_comprehensive_challenge.sh", "provider"},
		{"security_scanning_challenge.sh", "security"},
		{"debate_team_challenge.sh", "debate"},
		{"cli_agent_config_challenge.sh", "cli"},
		{"mcp_adapter_challenge.sh", "mcp"},
		{"bigdata_comprehensive_challenge.sh", "bigdata"},
		{"memory_system_challenge.sh", "memory"},
		{"unknown_challenge.sh", "shell"},
		{"release_build_challenge.sh", "release"},
		{"speckit_auto_activation_challenge.sh", "speckit"},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			assert.Equal(t, tc.expected,
				detectCategory(tc.filename))
		})
	}
}

func TestRegisterShellChallengesEnhanced_Basic(t *testing.T) {
	dir := t.TempDir()

	scripts := []string{
		"provider_test_challenge.sh",
		"security_test_challenge.sh",
		"not_a_challenge.txt",
	}
	for _, s := range scripts {
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, s),
			[]byte("#!/bin/bash\necho ok\n"), 0755,
		))
	}

	reg := registry.NewRegistry()
	err := RegisterShellChallengesEnhanced(reg, dir, "")
	require.NoError(t, err)
	assert.Equal(t, 2, reg.Count())
}

func TestRegisterShellChallengesEnhanced_NonexistentDir(
	t *testing.T,
) {
	reg := registry.NewRegistry()
	err := RegisterShellChallengesEnhanced(
		reg, "/nonexistent", "",
	)
	assert.Error(t, err)
}

func TestOrchestrator_EnvLoading(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".env"),
		[]byte("TEST_KEY=test_value\n"), 0644,
	))

	o := NewOrchestrator(OrchestratorConfig{
		ProjectRoot: dir,
	})
	assert.Equal(t, "test_value", o.envVars["TEST_KEY"])
}

func TestOrchestrator_Run_Parallel(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "challenges", "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0755))

	for _, name := range []string{
		"alpha_challenge.sh",
		"beta_challenge.sh",
	} {
		require.NoError(t, os.WriteFile(
			filepath.Join(scriptsDir, name),
			[]byte("#!/bin/bash\necho ok\nexit 0\n"), 0755,
		))
	}

	o := NewOrchestrator(OrchestratorConfig{
		ProjectRoot:    dir,
		ResultsDir:     filepath.Join(dir, "results"),
		Parallel:       true,
		MaxConcurrency: 2,
		Timeout:        10 * time.Second,
	})
	require.NoError(t, o.RegisterAll())

	result, err := o.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Passed)
}

func TestChallengeInfo_Fields(t *testing.T) {
	info := ChallengeInfo{
		ID:          "test-id",
		Name:        "Test Name",
		Description: "Test Description",
		Category:    "test",
	}
	assert.Equal(t, "test-id", info.ID)
	assert.Equal(t, "Test Name", info.Name)
	assert.Equal(t, "Test Description", info.Description)
	assert.Equal(t, "test", info.Category)
}
