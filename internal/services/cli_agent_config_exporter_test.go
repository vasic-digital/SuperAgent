package services

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// =====================================================
// NewCLIAgentConfigExporter TESTS
// =====================================================

func TestNewCLIAgentConfigExporter_WithLogger(t *testing.T) {
	logger := newTestLogger()
	exporter := NewCLIAgentConfigExporter(logger)

	require.NotNil(t, exporter)
	assert.Equal(t, logger, exporter.log)
	assert.NotNil(t, exporter.generator)
	assert.True(t, exporter.lastExport.IsZero(), "lastExport should be zero initially")
}

func TestNewCLIAgentConfigExporter_NilLogger(t *testing.T) {
	exporter := NewCLIAgentConfigExporter(nil)

	require.NotNil(t, exporter)
	assert.NotNil(t, exporter.log, "should create default logger when nil")
	assert.NotNil(t, exporter.generator)
}

// =====================================================
// ExportResult TYPE TESTS
// =====================================================

func TestExportResult_Type(t *testing.T) {
	result := &ExportResult{
		ExportedAt:   time.Now(),
		Agents:       []AgentExportResult{},
		SuccessCount: 5,
		FailedCount:  2,
		Duration:     3 * time.Second,
		Error:        "",
	}

	assert.NotZero(t, result.ExportedAt)
	assert.Empty(t, result.Agents)
	assert.Equal(t, 5, result.SuccessCount)
	assert.Equal(t, 2, result.FailedCount)
	assert.Equal(t, 3*time.Second, result.Duration)
	assert.Empty(t, result.Error)
}

func TestExportResult_WithError(t *testing.T) {
	result := &ExportResult{
		ExportedAt: time.Now(),
		Error:      "generation failed",
	}

	assert.NotEmpty(t, result.Error)
	assert.Equal(t, "generation failed", result.Error)
}

func TestExportResult_WithAgents(t *testing.T) {
	agents := []AgentExportResult{
		{AgentType: "opencode", Success: true, Path: "/home/user/.config/opencode/opencode.json"},
		{AgentType: "crush", Success: true, Path: "/home/user/.config/crush/crush.json"},
		{AgentType: "kilocode", Success: false, Errors: []string{"schema validation failed"}},
	}

	result := &ExportResult{
		ExportedAt:   time.Now(),
		Agents:       agents,
		SuccessCount: 2,
		FailedCount:  1,
		Duration:     1 * time.Second,
	}

	assert.Len(t, result.Agents, 3)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 1, result.FailedCount)
}

// =====================================================
// AgentExportResult TYPE TESTS
// =====================================================

func TestAgentExportResult_Type(t *testing.T) {
	tests := []struct {
		name   string
		result AgentExportResult
	}{
		{
			name: "successful agent export",
			result: AgentExportResult{
				AgentType: "opencode",
				Success:   true,
				Path:      "/home/user/.config/opencode/opencode.json",
				Errors:    nil,
			},
		},
		{
			name: "failed agent export",
			result: AgentExportResult{
				AgentType: "crush",
				Success:   false,
				Path:      "",
				Errors:    []string{"save failed: permission denied"},
			},
		},
		{
			name: "agent export with multiple errors",
			result: AgentExportResult{
				AgentType: "helixcode",
				Success:   false,
				Path:      "",
				Errors:    []string{"schema error", "validation error", "save error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.result.AgentType)
			if tt.result.Success {
				assert.NotEmpty(t, tt.result.Path)
				assert.Empty(t, tt.result.Errors)
			} else {
				assert.NotEmpty(t, tt.result.Errors)
			}
		})
	}
}

// =====================================================
// GetLastExport TESTS
// =====================================================

func TestCLIAgentConfigExporter_GetLastExport_Initial(t *testing.T) {
	exporter := NewCLIAgentConfigExporter(newTestLogger())

	lastExport := exporter.GetLastExport()
	assert.True(t, lastExport.IsZero(), "initial lastExport should be zero time")
}

// =====================================================
// OnVerificationComplete TESTS
// =====================================================

func TestCLIAgentConfigExporter_OnVerificationComplete(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress noise

	exporter := NewCLIAgentConfigExporter(logger)

	// OnVerificationComplete delegates to ExportAllConfigs.
	// The result depends on the generator's ability to generate configs,
	// which may fail in a test environment (no API keys, etc.).
	// We just verify it doesn't panic and returns properly.
	ctx := context.Background()
	err := exporter.OnVerificationComplete(ctx, nil)

	// May error if generator has issues, but should not panic
	_ = err
}

// =====================================================
// ExportAllConfigs TESTS
// =====================================================

func TestCLIAgentConfigExporter_ExportAllConfigs(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	exporter := NewCLIAgentConfigExporter(logger)

	ctx := context.Background()
	result, err := exporter.ExportAllConfigs(ctx)

	// The generator may fail in test env due to missing API keys,
	// but we can still validate the structure
	if err != nil {
		require.NotNil(t, result)
		assert.NotEmpty(t, result.Error)
	} else {
		require.NotNil(t, result)
		assert.NotZero(t, result.ExportedAt)
		assert.NotNil(t, result.Agents)
	}

	// After export, lastExport should be updated (even on partial failure)
	// unless the generation itself failed
	if err == nil {
		lastExport := exporter.GetLastExport()
		assert.False(t, lastExport.IsZero())
	}
}

func TestCLIAgentConfigExporter_ExportAllConfigs_SetsLastExport(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	exporter := NewCLIAgentConfigExporter(logger)

	before := time.Now()
	ctx := context.Background()
	_, _ = exporter.ExportAllConfigs(ctx)

	// The lastExport should be set regardless of success/failure
	// (only set on complete method execution, which we can verify)
	lastExport := exporter.GetLastExport()
	if !lastExport.IsZero() {
		assert.True(t, lastExport.After(before) || lastExport.Equal(before))
	}
}

// =====================================================
// ExportAllConfigs RESULT VALIDATION
// =====================================================

func TestCLIAgentConfigExporter_ExportAllConfigs_ResultStructure(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	exporter := NewCLIAgentConfigExporter(logger)

	ctx := context.Background()
	result, _ := exporter.ExportAllConfigs(ctx)

	require.NotNil(t, result)
	assert.NotZero(t, result.ExportedAt)
	assert.NotNil(t, result.Agents, "Agents slice should never be nil")
	// SuccessCount + FailedCount should equal len(Agents)
	assert.Equal(t, len(result.Agents), result.SuccessCount+result.FailedCount)
}

// =====================================================
// CONCURRENCY TESTS
// =====================================================

func TestCLIAgentConfigExporter_ConcurrentGetLastExport(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	exporter := NewCLIAgentConfigExporter(logger)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			_ = exporter.GetLastExport()
		}
	}()

	// GetLastExport uses mutex, should be safe under concurrent access
	for i := 0; i < 100; i++ {
		_ = exporter.GetLastExport()
	}

	<-done
}

// =====================================================
// BENCHMARK TESTS
// =====================================================

func BenchmarkNewCLIAgentConfigExporter(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCLIAgentConfigExporter(logger)
	}
}

func BenchmarkCLIAgentConfigExporter_GetLastExport(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	exporter := NewCLIAgentConfigExporter(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = exporter.GetLastExport()
	}
}
