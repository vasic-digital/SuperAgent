package background

import (
	"encoding/json"
	"testing"
	"time"

	internalbackground "dev.helix.agent/internal/background"
	internalmodels "dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
	extractedmodels "digital.vasic.models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// convertToInternalTask Tests
// =============================================================================

func TestConvertToInternalTask_Normal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	extracted := &extractedmodels.BackgroundTask{
		ID:       "task-123",
		TaskType: "code_format",
		TaskName: "Format Go files",
		Status:   extractedmodels.TaskStatus("running"),
		Priority: extractedmodels.TaskPriority("high"),
		Progress: 50.0,
		Payload:  json.RawMessage(`{"file":"main.go"}`),
		Config: extractedmodels.TaskConfig{
			TimeoutSeconds:        1800,
			HeartbeatIntervalSecs: 10,
		},
		MaxRetries:        3,
		RetryCount:        1,
		RetryDelaySeconds: 60,
		RequiredCPUCores:  2,
		RequiredMemoryMB:  512,
		CreatedAt:         now,
		UpdatedAt:         now,
		ScheduledAt:       now,
		ErrorHistory:      json.RawMessage(`[]`),
		Tags:              json.RawMessage(`["go"]`),
		Metadata:          json.RawMessage(`{}`),
	}

	result := convertToInternalTask(extracted)

	require.NotNil(t, result)
	assert.Equal(t, "task-123", result.ID)
	assert.Equal(t, "code_format", result.TaskType)
	assert.Equal(t, "Format Go files", result.TaskName)
	assert.Equal(t, internalmodels.TaskStatus("running"), result.Status)
	assert.Equal(t, internalmodels.TaskPriority("high"), result.Priority)
	assert.Equal(t, 50.0, result.Progress)
	assert.Equal(t, 3, result.MaxRetries)
	assert.Equal(t, 1, result.RetryCount)
	assert.Equal(t, 2, result.RequiredCPUCores)
	assert.Equal(t, 512, result.RequiredMemoryMB)
}

func TestConvertToInternalTask_NilInput(t *testing.T) {
	result := convertToInternalTask(nil)
	assert.Nil(t, result)
}

func TestConvertToInternalTask_EmptyTask(t *testing.T) {
	extracted := &extractedmodels.BackgroundTask{}
	result := convertToInternalTask(extracted)
	require.NotNil(t, result)
	assert.Empty(t, result.ID)
}

// =============================================================================
// convertToExtractedTask Tests
// =============================================================================

func TestConvertToExtractedTask_Normal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	internal := &internalmodels.BackgroundTask{
		ID:               "task-456",
		TaskType:         "lint",
		TaskName:         "Lint project",
		Status:           internalmodels.TaskStatusPending,
		Priority:         internalmodels.TaskPriorityNormal,
		Progress:         0,
		Payload:          json.RawMessage(`{}`),
		Config:           internalmodels.DefaultTaskConfig(),
		MaxRetries:       5,
		RequiredCPUCores: 1,
		RequiredMemoryMB: 256,
		ErrorHistory:     json.RawMessage(`[]`),
		Tags:             json.RawMessage(`[]`),
		Metadata:         json.RawMessage(`{}`),
		CreatedAt:        now,
		UpdatedAt:        now,
		ScheduledAt:      now,
	}

	result := convertToExtractedTask(internal)

	require.NotNil(t, result)
	assert.Equal(t, "task-456", result.ID)
	assert.Equal(t, "lint", result.TaskType)
	assert.Equal(t, extractedmodels.TaskStatus("pending"), result.Status)
	assert.Equal(t, extractedmodels.TaskPriority("normal"), result.Priority)
}

func TestConvertToExtractedTask_NilInput(t *testing.T) {
	result := convertToExtractedTask(nil)
	assert.Nil(t, result)
}

// =============================================================================
// convertToInternalTasks Tests
// =============================================================================

func TestConvertToInternalTasks_Normal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	extracted := []*extractedmodels.BackgroundTask{
		{ID: "task-1", TaskType: "test", CreatedAt: now, UpdatedAt: now, ScheduledAt: now,
			ErrorHistory: json.RawMessage(`[]`), Tags: json.RawMessage(`[]`), Metadata: json.RawMessage(`{}`)},
		{ID: "task-2", TaskType: "test", CreatedAt: now, UpdatedAt: now, ScheduledAt: now,
			ErrorHistory: json.RawMessage(`[]`), Tags: json.RawMessage(`[]`), Metadata: json.RawMessage(`{}`)},
	}

	result := convertToInternalTasks(extracted)

	require.Len(t, result, 2)
	assert.Equal(t, "task-1", result[0].ID)
	assert.Equal(t, "task-2", result[1].ID)
}

func TestConvertToInternalTasks_NilInput(t *testing.T) {
	result := convertToInternalTasks(nil)
	assert.Nil(t, result)
}

func TestConvertToInternalTasks_EmptySlice(t *testing.T) {
	result := convertToInternalTasks([]*extractedmodels.BackgroundTask{})
	require.NotNil(t, result)
	assert.Len(t, result, 0)
}

// =============================================================================
// convertToExtractedTasks Tests
// =============================================================================

func TestConvertToExtractedTasks_Normal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	internal := []*internalmodels.BackgroundTask{
		{ID: "task-a", TaskType: "test", CreatedAt: now, UpdatedAt: now, ScheduledAt: now,
			ErrorHistory: json.RawMessage(`[]`), Tags: json.RawMessage(`[]`), Metadata: json.RawMessage(`{}`)},
		{ID: "task-b", TaskType: "test", CreatedAt: now, UpdatedAt: now, ScheduledAt: now,
			ErrorHistory: json.RawMessage(`[]`), Tags: json.RawMessage(`[]`), Metadata: json.RawMessage(`{}`)},
		{ID: "task-c", TaskType: "test", CreatedAt: now, UpdatedAt: now, ScheduledAt: now,
			ErrorHistory: json.RawMessage(`[]`), Tags: json.RawMessage(`[]`), Metadata: json.RawMessage(`{}`)},
	}

	result := convertToExtractedTasks(internal)

	require.Len(t, result, 3)
	assert.Equal(t, "task-a", result[0].ID)
	assert.Equal(t, "task-c", result[2].ID)
}

func TestConvertToExtractedTasks_NilInput(t *testing.T) {
	result := convertToExtractedTasks(nil)
	assert.Nil(t, result)
}

// =============================================================================
// convertTaskStatus Tests
// =============================================================================

func TestConvertTaskStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    internalmodels.TaskStatus
		expected extractedmodels.TaskStatus
	}{
		{"pending", internalmodels.TaskStatusPending, extractedmodels.TaskStatus("pending")},
		{"queued", internalmodels.TaskStatusQueued, extractedmodels.TaskStatus("queued")},
		{"running", internalmodels.TaskStatusRunning, extractedmodels.TaskStatus("running")},
		{"paused", internalmodels.TaskStatusPaused, extractedmodels.TaskStatus("paused")},
		{"completed", internalmodels.TaskStatusCompleted, extractedmodels.TaskStatus("completed")},
		{"failed", internalmodels.TaskStatusFailed, extractedmodels.TaskStatus("failed")},
		{"stuck", internalmodels.TaskStatusStuck, extractedmodels.TaskStatus("stuck")},
		{"cancelled", internalmodels.TaskStatusCancelled, extractedmodels.TaskStatus("cancelled")},
		{"dead_letter", internalmodels.TaskStatusDeadLetter, extractedmodels.TaskStatus("dead_letter")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTaskStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// convertTaskPriority Tests
// =============================================================================

func TestConvertTaskPriority(t *testing.T) {
	tests := []struct {
		name     string
		input    internalmodels.TaskPriority
		expected extractedmodels.TaskPriority
	}{
		{"critical", internalmodels.TaskPriorityCritical, extractedmodels.TaskPriority("critical")},
		{"high", internalmodels.TaskPriorityHigh, extractedmodels.TaskPriority("high")},
		{"normal", internalmodels.TaskPriorityNormal, extractedmodels.TaskPriority("normal")},
		{"low", internalmodels.TaskPriorityLow, extractedmodels.TaskPriority("low")},
		{"background", internalmodels.TaskPriorityBackground, extractedmodels.TaskPriority("background")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTaskPriority(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// convertToInternalResourceSnapshot Tests
// =============================================================================

func TestConvertToInternalResourceSnapshot_Normal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	extracted := &extractedmodels.ResourceSnapshot{
		ID:            "snap-1",
		TaskID:        "task-1",
		CPUPercent:    45.5,
		MemoryPercent: 60.0,
		ThreadCount:   8,
		SampledAt:     now,
	}

	result := convertToInternalResourceSnapshot(extracted)

	require.NotNil(t, result)
	assert.Equal(t, "snap-1", result.ID)
	assert.Equal(t, "task-1", result.TaskID)
	assert.Equal(t, 45.5, result.CPUPercent)
	assert.Equal(t, 60.0, result.MemoryPercent)
	assert.Equal(t, 8, result.ThreadCount)
}

func TestConvertToInternalResourceSnapshot_NilInput(t *testing.T) {
	result := convertToInternalResourceSnapshot(nil)
	assert.Nil(t, result)
}

// =============================================================================
// convertToExtractedResourceSnapshot Tests
// =============================================================================

func TestConvertToExtractedResourceSnapshot_Normal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	internal := &internalmodels.ResourceSnapshot{
		ID:            "snap-2",
		TaskID:        "task-2",
		CPUPercent:    30.0,
		MemoryPercent: 40.0,
		SampledAt:     now,
	}

	result := convertToExtractedResourceSnapshot(internal)

	require.NotNil(t, result)
	assert.Equal(t, "snap-2", result.ID)
	assert.Equal(t, "task-2", result.TaskID)
	assert.Equal(t, 30.0, result.CPUPercent)
}

func TestConvertToExtractedResourceSnapshot_NilInput(t *testing.T) {
	result := convertToExtractedResourceSnapshot(nil)
	assert.Nil(t, result)
}

// =============================================================================
// convertToInternalSystemResources Tests
// =============================================================================

func TestConvertToInternalSystemResources_Normal(t *testing.T) {
	extracted := &extractedbackground.SystemResources{
		TotalCPUCores:     8,
		AvailableCPUCores: 4.0,
		TotalMemoryMB:     16384,
		AvailableMemoryMB: 8192,
		CPULoadPercent:    50.0,
		MemoryUsedPercent: 50.0,
		LoadAvg1:          2.0,
		LoadAvg5:          1.5,
		LoadAvg15:         1.0,
	}

	result := convertToInternalSystemResources(extracted)

	require.NotNil(t, result)
	assert.Equal(t, 8, result.TotalCPUCores)
	assert.Equal(t, 4.0, result.AvailableCPUCores)
	assert.Equal(t, int64(16384), result.TotalMemoryMB)
	assert.Equal(t, int64(8192), result.AvailableMemoryMB)
	assert.Equal(t, 50.0, result.CPULoadPercent)
}

func TestConvertToInternalSystemResources_NilInput(t *testing.T) {
	result := convertToInternalSystemResources(nil)
	assert.Nil(t, result)
}

// =============================================================================
// convertToExtractedSystemResources Tests
// =============================================================================

func TestConvertToExtractedSystemResources_Normal(t *testing.T) {
	internal := &internalbackground.SystemResources{
		TotalCPUCores:     16,
		AvailableCPUCores: 8.0,
		TotalMemoryMB:     32768,
		AvailableMemoryMB: 16384,
	}

	result := convertToExtractedSystemResources(internal)

	require.NotNil(t, result)
	assert.Equal(t, 16, result.TotalCPUCores)
	assert.Equal(t, 8.0, result.AvailableCPUCores)
}

func TestConvertToExtractedSystemResources_NilInput(t *testing.T) {
	result := convertToExtractedSystemResources(nil)
	assert.Nil(t, result)
}

// =============================================================================
// convertToInternalStuckAnalysis Tests
// =============================================================================

func TestConvertToInternalStuckAnalysis_Normal(t *testing.T) {
	extracted := &extractedbackground.StuckAnalysis{
		IsStuck:         true,
		Reason:          "no heartbeat for 10m",
		Recommendations: []string{"restart task", "check logs"},
	}

	result := convertToInternalStuckAnalysis(extracted)

	require.NotNil(t, result)
	assert.True(t, result.IsStuck)
	assert.Equal(t, "no heartbeat for 10m", result.Reason)
	assert.Len(t, result.Recommendations, 2)
}

func TestConvertToInternalStuckAnalysis_NilInput(t *testing.T) {
	result := convertToInternalStuckAnalysis(nil)
	assert.Nil(t, result)
}

// =============================================================================
// convertToExtractedStuckAnalysis Tests
// =============================================================================

func TestConvertToExtractedStuckAnalysis_Normal(t *testing.T) {
	internal := &internalbackground.StuckAnalysis{
		IsStuck: false,
		Reason:  "",
	}

	result := convertToExtractedStuckAnalysis(internal)

	require.NotNil(t, result)
	assert.False(t, result.IsStuck)
}

func TestConvertToExtractedStuckAnalysis_NilInput(t *testing.T) {
	result := convertToExtractedStuckAnalysis(nil)
	assert.Nil(t, result)
}

// =============================================================================
// Roundtrip conversion tests (internal -> extracted -> internal)
// =============================================================================

func TestRoundtrip_Task(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &internalmodels.BackgroundTask{
		ID:                "roundtrip-task",
		TaskType:          "test",
		TaskName:          "Roundtrip test",
		Status:            internalmodels.TaskStatusRunning,
		Priority:          internalmodels.TaskPriorityHigh,
		Progress:          75.0,
		MaxRetries:        3,
		RetryCount:        1,
		RetryDelaySeconds: 30,
		RequiredCPUCores:  4,
		RequiredMemoryMB:  2048,
		Payload:           json.RawMessage(`{"key":"value"}`),
		ErrorHistory:      json.RawMessage(`[]`),
		Tags:              json.RawMessage(`["test"]`),
		Metadata:          json.RawMessage(`{"env":"test"}`),
		CreatedAt:         now,
		UpdatedAt:         now,
		ScheduledAt:       now,
	}

	extracted := convertToExtractedTask(original)
	require.NotNil(t, extracted)

	recovered := convertToInternalTask(extracted)
	require.NotNil(t, recovered)

	assert.Equal(t, original.ID, recovered.ID)
	assert.Equal(t, original.TaskType, recovered.TaskType)
	assert.Equal(t, original.TaskName, recovered.TaskName)
	assert.Equal(t, original.Status, recovered.Status)
	assert.Equal(t, original.Priority, recovered.Priority)
	assert.Equal(t, original.Progress, recovered.Progress)
	assert.Equal(t, original.MaxRetries, recovered.MaxRetries)
	assert.Equal(t, original.RequiredCPUCores, recovered.RequiredCPUCores)
}

func TestRoundtrip_ResourceSnapshot(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &internalmodels.ResourceSnapshot{
		ID:            "snap-roundtrip",
		TaskID:        "task-roundtrip",
		CPUPercent:    55.5,
		MemoryPercent: 72.3,
		ThreadCount:   12,
		SampledAt:     now,
	}

	extracted := convertToExtractedResourceSnapshot(original)
	require.NotNil(t, extracted)

	recovered := convertToInternalResourceSnapshot(extracted)
	require.NotNil(t, recovered)

	assert.Equal(t, original.ID, recovered.ID)
	assert.Equal(t, original.TaskID, recovered.TaskID)
	assert.Equal(t, original.CPUPercent, recovered.CPUPercent)
	assert.Equal(t, original.MemoryPercent, recovered.MemoryPercent)
	assert.Equal(t, original.ThreadCount, recovered.ThreadCount)
}
