package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// =============================================================================
// Types Tests
// =============================================================================

func TestRenderStyleConstants(t *testing.T) {
	assert.Equal(t, RenderStyle("theater"), RenderStyleTheater)
	assert.Equal(t, RenderStyle("novel"), RenderStyleNovel)
	assert.Equal(t, RenderStyle("screenplay"), RenderStyleScreenplay)
	assert.Equal(t, RenderStyle("minimal"), RenderStyleMinimal)
	assert.Equal(t, RenderStyle("plain"), RenderStylePlain)
}

func TestCLIClientConstants(t *testing.T) {
	assert.Equal(t, CLIClient("opencode"), CLIClientOpenCode)
	assert.Equal(t, CLIClient("crush"), CLIClientCrush)
	assert.Equal(t, CLIClient("helixcode"), CLIClientHelixCode)
	assert.Equal(t, CLIClient("kilocode"), CLIClientKiloCode)
	assert.Equal(t, CLIClient("unknown"), CLIClientUnknown)
}

func TestProgressBarStyleConstants(t *testing.T) {
	assert.Equal(t, ProgressBarStyle("ascii"), ProgressBarStyleASCII)
	assert.Equal(t, ProgressBarStyle("unicode"), ProgressBarStyleUnicode)
	assert.Equal(t, ProgressBarStyle("block"), ProgressBarStyleBlock)
	assert.Equal(t, ProgressBarStyle("dots"), ProgressBarStyleDots)
}

func TestColorSchemeConstants(t *testing.T) {
	assert.Equal(t, ColorScheme("none"), ColorSchemeNone)
	assert.Equal(t, ColorScheme("8"), ColorScheme8)
	assert.Equal(t, ColorScheme("256"), ColorScheme256)
	assert.Equal(t, ColorScheme("true"), ColorSchemeTrueColor)
}

func TestDefaultRenderConfig(t *testing.T) {
	config := DefaultRenderConfig()

	require.NotNil(t, config)
	assert.Equal(t, RenderStyleTheater, config.Style)
	assert.Equal(t, ProgressBarStyleUnicode, config.ProgressStyle)
	assert.Equal(t, ColorScheme256, config.ColorScheme)
	assert.True(t, config.ShowResources)
	assert.True(t, config.ShowLogs)
	assert.Equal(t, 5, config.LogLines)
	assert.Equal(t, 80, config.Width)
	assert.True(t, config.Animate)
	assert.Equal(t, 100*time.Millisecond, config.RefreshRate)
}

func TestProgressBarContent_Fields(t *testing.T) {
	now := time.Now()
	eta := 5 * time.Minute
	content := ProgressBarContent{
		TaskID:      "task-1",
		TaskName:    "Build Project",
		TaskType:    "build",
		Progress:    75.5,
		Message:     "Compiling...",
		Status:      "running",
		StartedAt:   now,
		ETA:         &eta,
		CurrentStep: 3,
		TotalSteps:  4,
	}

	assert.Equal(t, "task-1", content.TaskID)
	assert.Equal(t, "Build Project", content.TaskName)
	assert.Equal(t, "build", content.TaskType)
	assert.Equal(t, 75.5, content.Progress)
	assert.Equal(t, "Compiling...", content.Message)
	assert.Equal(t, "running", content.Status)
	assert.Equal(t, now, content.StartedAt)
	assert.Equal(t, 5*time.Minute, *content.ETA)
	assert.Equal(t, 3, content.CurrentStep)
	assert.Equal(t, 4, content.TotalSteps)
}

func TestStatusTableContent_Fields(t *testing.T) {
	content := StatusTableContent{
		Tasks: []TaskStatusRow{
			{ID: "1", Name: "Task1", Status: "running"},
			{ID: "2", Name: "Task2", Status: "pending"},
		},
		TotalCount: 2,
		Timestamp:  time.Now(),
	}

	assert.Len(t, content.Tasks, 2)
	assert.Equal(t, 2, content.TotalCount)
}

func TestTaskStatusRow_Fields(t *testing.T) {
	row := TaskStatusRow{
		ID:       "task-123",
		Name:     "Test Task",
		Type:     "test",
		Status:   "running",
		Progress: 50.0,
		Duration: 30 * time.Second,
		WorkerID: "worker-1",
		Message:  "Processing...",
	}

	assert.Equal(t, "task-123", row.ID)
	assert.Equal(t, "Test Task", row.Name)
	assert.Equal(t, "test", row.Type)
	assert.Equal(t, "running", row.Status)
	assert.Equal(t, 50.0, row.Progress)
	assert.Equal(t, 30*time.Second, row.Duration)
	assert.Equal(t, "worker-1", row.WorkerID)
	assert.Equal(t, "Processing...", row.Message)
}

func TestResourceGaugeContent_Fields(t *testing.T) {
	content := ResourceGaugeContent{
		TaskID:        "task-1",
		CPUPercent:    75.5,
		MemoryPercent: 60.0,
		MemoryBytes:   1024 * 1024 * 512,
		MemoryMax:     1024 * 1024 * 1024,
		IOReadBytes:   1024 * 1024,
		IOWriteBytes:  512 * 1024,
		NetBytesSent:  2048,
		NetBytesRecv:  4096,
		OpenFDs:       25,
		ThreadCount:   8,
	}

	assert.Equal(t, "task-1", content.TaskID)
	assert.Equal(t, 75.5, content.CPUPercent)
	assert.Equal(t, 60.0, content.MemoryPercent)
	assert.Equal(t, int64(536870912), content.MemoryBytes)
	assert.Equal(t, int64(1073741824), content.MemoryMax)
	assert.Equal(t, 25, content.OpenFDs)
	assert.Equal(t, 8, content.ThreadCount)
}

func TestLogLineContent_Fields(t *testing.T) {
	now := time.Now()
	log := LogLineContent{
		Timestamp: now,
		Level:     "INFO",
		Message:   "Test message",
		Fields: map[string]interface{}{
			"key": "value",
		},
	}

	assert.Equal(t, now, log.Timestamp)
	assert.Equal(t, "INFO", log.Level)
	assert.Equal(t, "Test message", log.Message)
	assert.Equal(t, "value", log.Fields["key"])
}

func TestWorkerStatsContent_Fields(t *testing.T) {
	stats := WorkerStatsContent{
		ActiveWorkers:  5,
		IdleWorkers:    3,
		TotalWorkers:   8,
		TasksCompleted: 100,
		TasksFailed:    5,
	}

	assert.Equal(t, 5, stats.ActiveWorkers)
	assert.Equal(t, 3, stats.IdleWorkers)
	assert.Equal(t, 8, stats.TotalWorkers)
	assert.Equal(t, int64(100), stats.TasksCompleted)
	assert.Equal(t, int64(5), stats.TasksFailed)
}

func TestQueueStatsContent_Fields(t *testing.T) {
	stats := QueueStatsContent{
		PendingCount: 10,
		RunningCount: 5,
		ByPriority:   map[string]int64{"high": 3, "normal": 7},
		ByStatus:     map[string]int64{"pending": 10, "running": 5},
	}

	assert.Equal(t, int64(10), stats.PendingCount)
	assert.Equal(t, int64(5), stats.RunningCount)
	assert.Equal(t, int64(3), stats.ByPriority["high"])
	assert.Equal(t, int64(10), stats.ByStatus["pending"])
}

func TestSystemInfoContent_Fields(t *testing.T) {
	info := SystemInfoContent{
		CPUCores:          8,
		CPUUsedPercent:    45.5,
		MemoryTotalMB:     16384,
		MemoryUsedMB:      8192,
		MemoryUsedPercent: 50.0,
		LoadAvg1:          1.5,
		LoadAvg5:          2.0,
		LoadAvg15:         1.8,
	}

	assert.Equal(t, 8, info.CPUCores)
	assert.Equal(t, 45.5, info.CPUUsedPercent)
	assert.Equal(t, int64(16384), info.MemoryTotalMB)
	assert.Equal(t, 50.0, info.MemoryUsedPercent)
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		status   models.TaskStatus
		expected string
	}{
		{models.TaskStatusPending, IconPending},
		{models.TaskStatusQueued, IconPending},
		{models.TaskStatusRunning, IconRunning},
		{models.TaskStatusCompleted, IconCompleted},
		{models.TaskStatusFailed, IconFailed},
		{models.TaskStatusStuck, IconStuck},
		{models.TaskStatusCancelled, IconCancelled},
		{models.TaskStatusPaused, IconPaused},
		{models.TaskStatus("unknown"), IconPending}, // default
	}

	for _, tc := range tests {
		icon := GetStatusIcon(tc.status)
		assert.Equal(t, tc.expected, icon, "Status: %s", tc.status)
	}
}

func TestGetStatusColor(t *testing.T) {
	tests := []struct {
		status   models.TaskStatus
		expected string
	}{
		{models.TaskStatusPending, ColorYellow},
		{models.TaskStatusQueued, ColorYellow},
		{models.TaskStatusRunning, ColorCyan},
		{models.TaskStatusCompleted, ColorGreen},
		{models.TaskStatusFailed, ColorRed},
		{models.TaskStatusStuck, ColorBrightYellow},
		{models.TaskStatusCancelled, ColorMagenta},
		{models.TaskStatusPaused, ColorBlue},
		{models.TaskStatus("unknown"), ColorWhite}, // default
	}

	for _, tc := range tests {
		color := GetStatusColor(tc.status)
		assert.Equal(t, tc.expected, color, "Status: %s", tc.status)
	}
}

func TestSpinnerFrames(t *testing.T) {
	assert.Len(t, SpinnerFrames, 10)
	assert.Equal(t, IconSpinner1, SpinnerFrames[0])
	assert.Equal(t, IconSpinner10, SpinnerFrames[9])
}

// =============================================================================
// Renderer Tests
// =============================================================================

func TestNewRenderer(t *testing.T) {
	// With nil config
	r := NewRenderer(nil, CLIClientOpenCode)
	require.NotNil(t, r)
	assert.NotNil(t, r.config)
	assert.Equal(t, CLIClientOpenCode, r.client)

	// With custom config
	config := &RenderConfig{
		Style:     RenderStyleMinimal,
		Width:     120,
		ColorScheme: ColorSchemeNone,
	}
	r = NewRenderer(config, CLIClientCrush)
	assert.Equal(t, RenderStyleMinimal, r.config.Style)
	assert.Equal(t, 120, r.config.Width)
	assert.Equal(t, CLIClientCrush, r.client)
}

func TestRenderer_SetWriter(t *testing.T) {
	r := NewRenderer(nil, CLIClientOpenCode)
	buf := &bytes.Buffer{}

	r.SetWriter(buf)
	// The writer is private, so we can't directly test it
	// But this ensures the method doesn't panic
}

func TestRenderer_RenderProgressBar_Plain(t *testing.T) {
	config := &RenderConfig{
		Style:         RenderStylePlain,
		ProgressStyle: ProgressBarStyleASCII,
		ColorScheme:   ColorSchemeNone,
		Width:         80,
	}
	r := NewRenderer(config, CLIClientUnknown)
	r.isTTY = false // Force plain output

	content := &ProgressBarContent{
		Progress: 50.0,
		Message:  "Testing",
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "50.0%")
	assert.Contains(t, output, "Testing")
}

func TestRenderer_RenderProgressBar_Unicode(t *testing.T) {
	config := &RenderConfig{
		Style:         RenderStyleTheater,
		ProgressStyle: ProgressBarStyleUnicode,
		ColorScheme:   ColorScheme256,
		Width:         80,
	}
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true // Force colorized output

	content := &ProgressBarContent{
		Progress: 75.0,
		Message:  "Building",
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "75.0%")
	assert.Contains(t, output, "Building")
	// Should contain ANSI color codes
	assert.Contains(t, output, "\033[")
}

func TestRenderer_RenderProgressBar_Block(t *testing.T) {
	config := &RenderConfig{
		Style:         RenderStyleTheater,
		ProgressStyle: ProgressBarStyleBlock,
		ColorScheme:   ColorScheme256,
		Width:         80,
	}
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &ProgressBarContent{
		Progress: 60.0,
		Message:  "",
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "60.0%")
}

func TestRenderer_RenderProgressBar_Dots(t *testing.T) {
	config := &RenderConfig{
		Style:         RenderStyleTheater,
		ProgressStyle: ProgressBarStyleDots,
		ColorScheme:   ColorScheme256,
		Width:         80,
	}
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &ProgressBarContent{
		Progress: 40.0,
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "40.0%")
}

func TestRenderer_RenderProgressBar_OverflowProtection(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = false

	// Test progress > 100
	content := &ProgressBarContent{
		Progress: 150.0,
		Message:  "Overflow",
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "150.0%")
}

func TestRenderer_RenderProgressBar_NarrowWidth(t *testing.T) {
	config := &RenderConfig{
		Width: 20, // Very narrow
	}
	r := NewRenderer(config, CLIClientUnknown)
	r.isTTY = false

	content := &ProgressBarContent{
		Progress: 50.0,
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "50.0%")
}

func TestRenderer_RenderResourceGauge_Plain(t *testing.T) {
	config := &RenderConfig{
		ColorScheme: ColorSchemeNone,
	}
	r := NewRenderer(config, CLIClientUnknown)
	r.isTTY = false

	content := &ResourceGaugeContent{
		CPUPercent:   50.0,
		MemoryBytes:  1024 * 1024 * 100,
		IOReadBytes:  1024 * 1024,
		IOWriteBytes: 512 * 1024,
		NetBytesSent: 2048,
		NetBytesRecv: 4096,
	}

	output := r.RenderResourceGauge(content)
	assert.Contains(t, output, "CPU: 50.0%")
	assert.Contains(t, output, "Memory:")
	assert.Contains(t, output, "I/O:")
	assert.Contains(t, output, "Net:")
}

func TestRenderer_RenderResourceGauge_Colorized(t *testing.T) {
	config := &RenderConfig{
		ColorScheme: ColorScheme256,
		Width:       80,
	}
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &ResourceGaugeContent{
		CPUPercent:    95.0, // High - should be red
		MemoryPercent: 80.0, // Medium - should be yellow
		MemoryBytes:   1024 * 1024 * 800,
		MemoryMax:     1024 * 1024 * 1024,
		IOReadBytes:   1024 * 1024,
		IOWriteBytes:  512 * 1024,
	}

	output := r.RenderResourceGauge(content)
	assert.Contains(t, output, "Resources:")
	assert.Contains(t, output, "CPU:")
	assert.Contains(t, output, "Memory:")
}

func TestRenderer_RenderStatusTable_Empty(t *testing.T) {
	r := NewRenderer(nil, CLIClientOpenCode)

	content := &StatusTableContent{
		Tasks: []TaskStatusRow{},
	}

	output := r.RenderStatusTable(content)
	assert.Equal(t, "No tasks in queue.", output)
}

func TestRenderer_RenderStatusTable_Plain(t *testing.T) {
	config := &RenderConfig{
		ColorScheme: ColorSchemeNone,
		Width:       80,
	}
	r := NewRenderer(config, CLIClientUnknown)
	r.isTTY = false

	content := &StatusTableContent{
		Tasks: []TaskStatusRow{
			{ID: "1", Name: "Task1", Type: "build", Status: "running", Progress: 50.0, Duration: 30 * time.Second},
			{ID: "2", Name: "Task2", Type: "test", Status: "pending", Progress: 0.0, Duration: 0},
		},
		TotalCount: 2,
	}

	output := r.RenderStatusTable(content)
	assert.Contains(t, output, "BACKGROUND TASKS")
	assert.Contains(t, output, "Task1")
	assert.Contains(t, output, "Task2")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "pending")
}

func TestRenderer_RenderStatusTable_Colorized(t *testing.T) {
	config := &RenderConfig{
		ColorScheme: ColorScheme256,
		Width:       100,
	}
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &StatusTableContent{
		Tasks: []TaskStatusRow{
			{ID: "1", Name: "Build Application", Type: "build", Status: "running", Progress: 75.0, Duration: time.Minute},
		},
		TotalCount: 1,
	}

	output := r.RenderStatusTable(content)
	assert.Contains(t, output, "BACKGROUND TASKS")
	// Should contain Unicode box characters
	assert.Contains(t, output, BoxDoubleVertical)
}

func TestRenderer_RenderTaskPanel(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	eta := 5 * time.Minute
	content := &TaskPanelContent{
		Task: &models.BackgroundTask{
			ID:       "12345678-1234",
			TaskName: "Build Project",
			Status:   models.TaskStatusRunning,
		},
		Progress: &ProgressBarContent{
			Progress: 60.0,
			Message:  "Compiling...",
		},
		Resources: &ResourceGaugeContent{
			CPUPercent:  50.0,
			MemoryBytes: 1024 * 1024 * 100,
		},
		Logs: []LogLineContent{
			{Timestamp: time.Now(), Level: "INFO", Message: "Starting build"},
			{Timestamp: time.Now(), Level: "DEBUG", Message: "Parsing files"},
		},
		ETA: &eta,
	}

	output := r.RenderTaskPanel(content)
	assert.Contains(t, output, "Build Project")
	assert.Contains(t, output, "12345678")
	assert.Contains(t, output, "60.0%")
	assert.Contains(t, output, "ETA:")
}

func TestRenderer_RenderTaskPanel_NilTask(t *testing.T) {
	r := NewRenderer(nil, CLIClientOpenCode)

	content := &TaskPanelContent{
		Task: nil,
	}

	output := r.RenderTaskPanel(content)
	assert.Empty(t, output)
}

func TestRenderer_RenderTaskPanel_Plain(t *testing.T) {
	config := &RenderConfig{
		ColorScheme:   ColorSchemeNone,
		ShowResources: true,
	}
	r := NewRenderer(config, CLIClientUnknown)
	r.isTTY = false

	eta := 10 * time.Minute
	content := &TaskPanelContent{
		Task: &models.BackgroundTask{
			ID:       "abcdefgh-1234",
			TaskName: "Test Suite",
			Status:   models.TaskStatusRunning,
		},
		Progress: &ProgressBarContent{
			Progress: 80.0,
		},
		Resources: &ResourceGaugeContent{
			CPUPercent:  25.0,
			MemoryBytes: 1024 * 1024 * 50,
		},
		ETA: &eta,
	}

	output := r.RenderTaskPanel(content)
	assert.Contains(t, output, "Test Suite")
	assert.Contains(t, output, "abcdefgh")
	assert.Contains(t, output, "80.0%")
	assert.Contains(t, output, "CPU: 25.0%")
	assert.Contains(t, output, "ETA:")
}

func TestRenderer_RenderDashboard(t *testing.T) {
	config := DefaultRenderConfig()
	config.Width = 100
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &DashboardContent{
		Title:     "MY TASKS",
		Timestamp: time.Now(),
		WorkerStats: WorkerStatsContent{
			ActiveWorkers:  3,
			IdleWorkers:    2,
			TotalWorkers:   5,
			TasksCompleted: 100,
			TasksFailed:    2,
		},
		QueueStats: QueueStatsContent{
			PendingCount: 5,
			RunningCount: 3,
		},
		Tasks: []TaskPanelContent{
			{
				Task: &models.BackgroundTask{
					ID:       "task-1234-5678",
					TaskName: "Build",
					Status:   models.TaskStatusRunning,
				},
				Progress: &ProgressBarContent{Progress: 50.0},
			},
		},
	}

	output := r.RenderDashboard(content)
	assert.Contains(t, output, "MY TASKS")
	assert.Contains(t, output, "3 active")
	assert.Contains(t, output, "2 idle")
	assert.Contains(t, output, "100 completed")
	assert.Contains(t, output, "5 pending")
	assert.Contains(t, output, "Build")
}

func TestRenderer_RenderDashboard_DefaultTitle(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &DashboardContent{
		Title:       "", // Empty title
		WorkerStats: WorkerStatsContent{},
		QueueStats:  QueueStatsContent{},
		Tasks:       []TaskPanelContent{},
	}

	output := r.RenderDashboard(content)
	assert.Contains(t, output, "HELIXAGENT BACKGROUND TASKS")
}

func TestRenderer_RenderDashboard_Plain(t *testing.T) {
	config := &RenderConfig{
		ColorScheme: ColorSchemeNone,
		Width:       80,
	}
	r := NewRenderer(config, CLIClientUnknown)
	r.isTTY = false

	content := &DashboardContent{
		Title: "PLAIN DASHBOARD",
		WorkerStats: WorkerStatsContent{
			ActiveWorkers:  2,
			IdleWorkers:    1,
			TasksCompleted: 50,
			TasksFailed:    1,
		},
		QueueStats: QueueStatsContent{
			PendingCount: 3,
			RunningCount: 2,
		},
		Tasks: []TaskPanelContent{},
	}

	output := r.RenderDashboard(content)
	assert.Contains(t, output, "PLAIN DASHBOARD")
	assert.Contains(t, output, "2 active")
	assert.Contains(t, output, "50 completed")
}

func TestRenderer_GetSpinnerFrame(t *testing.T) {
	r := NewRenderer(nil, CLIClientOpenCode)

	frames := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		frames = append(frames, r.GetSpinnerFrame())
	}

	// Should cycle through all frames
	assert.Equal(t, SpinnerFrames, frames)

	// Should wrap around
	frame11 := r.GetSpinnerFrame()
	assert.Equal(t, SpinnerFrames[0], frame11)
}

// =============================================================================
// Detection Tests
// =============================================================================

func TestDetectCLIClient_OpenCode(t *testing.T) {
	// Save and restore env
	oldVal := os.Getenv("OPENCODE")
	defer os.Setenv("OPENCODE", oldVal)

	os.Setenv("OPENCODE", "1")
	client := DetectCLIClient()
	assert.Equal(t, CLIClientOpenCode, client)

	os.Unsetenv("OPENCODE")
	os.Setenv("OPENCODE_VERSION", "1.0.0")
	client = DetectCLIClient()
	assert.Equal(t, CLIClientOpenCode, client)
	os.Unsetenv("OPENCODE_VERSION")
}

func TestDetectCLIClient_Crush(t *testing.T) {
	oldVal := os.Getenv("CRUSH_CLI")
	defer os.Setenv("CRUSH_CLI", oldVal)

	os.Setenv("CRUSH_CLI", "1")
	client := DetectCLIClient()
	assert.Equal(t, CLIClientCrush, client)
	os.Unsetenv("CRUSH_CLI")
}

func TestDetectCLIClient_HelixCode(t *testing.T) {
	oldVal := os.Getenv("HELIXCODE")
	defer os.Setenv("HELIXCODE", oldVal)

	os.Setenv("HELIXCODE", "1")
	client := DetectCLIClient()
	assert.Equal(t, CLIClientHelixCode, client)
	os.Unsetenv("HELIXCODE")
}

func TestDetectCLIClient_KiloCode(t *testing.T) {
	oldVal := os.Getenv("KILOCODE")
	defer os.Setenv("KILOCODE", oldVal)

	os.Setenv("KILOCODE", "1")
	client := DetectCLIClient()
	assert.Equal(t, CLIClientKiloCode, client)
	os.Unsetenv("KILOCODE")
}

func TestDetectCLIClient_ClaudeCode(t *testing.T) {
	oldVal := os.Getenv("CLAUDE_CODE")
	defer os.Setenv("CLAUDE_CODE", oldVal)

	os.Setenv("CLAUDE_CODE", "1")
	client := DetectCLIClient()
	assert.Equal(t, CLIClientOpenCode, client) // Claude Code treated as OpenCode
	os.Unsetenv("CLAUDE_CODE")
}

func TestDetectFromUserAgent(t *testing.T) {
	tests := []struct {
		userAgent string
		expected  CLIClient
	}{
		{"opencode/1.0.0", CLIClientOpenCode},
		{"crush-cli/2.0.0", CLIClientCrush},
		{"helixcode/1.0.0", CLIClientHelixCode},
		{"kilocode/1.0.0", CLIClientKiloCode},
		{"Mozilla/5.0", CLIClientUnknown},
		{"", CLIClientUnknown},
	}

	for _, tc := range tests {
		client := detectFromUserAgent(tc.userAgent)
		assert.Equal(t, tc.expected, client, "User-Agent: %s", tc.userAgent)
	}
}

func TestDetectCLIClientInfo(t *testing.T) {
	info := DetectCLIClientInfo()

	require.NotNil(t, info)
	// Terminal width/height should have defaults
	assert.GreaterOrEqual(t, info.TerminalWidth, 0)
	assert.GreaterOrEqual(t, info.TerminalHeight, 0)
}

func TestDetectColorSupport(t *testing.T) {
	// Save current environment
	noColor := os.Getenv("NO_COLOR")
	forceColor := os.Getenv("FORCE_COLOR")
	term := os.Getenv("TERM")
	defer func() {
		os.Setenv("NO_COLOR", noColor)
		os.Setenv("FORCE_COLOR", forceColor)
		os.Setenv("TERM", term)
	}()

	// Test NO_COLOR
	os.Setenv("NO_COLOR", "1")
	os.Unsetenv("FORCE_COLOR")
	assert.False(t, detectColorSupport())

	// Test FORCE_COLOR
	os.Unsetenv("NO_COLOR")
	os.Setenv("FORCE_COLOR", "1")
	assert.True(t, detectColorSupport())

	// Test dumb terminal
	os.Unsetenv("FORCE_COLOR")
	os.Setenv("TERM", "dumb")
	assert.False(t, detectColorSupport())

	// Test 256color terminal
	os.Setenv("TERM", "xterm-256color")
	assert.True(t, detectColorSupport())
}

func TestDetectUnicodeSupport(t *testing.T) {
	// Save current environment
	lang := os.Getenv("LANG")
	lcAll := os.Getenv("LC_ALL")
	term := os.Getenv("TERM")
	defer func() {
		os.Setenv("LANG", lang)
		os.Setenv("LC_ALL", lcAll)
		os.Setenv("TERM", term)
	}()

	// Test UTF-8 locale
	os.Setenv("LANG", "en_US.UTF-8")
	os.Unsetenv("LC_ALL")
	assert.True(t, detectUnicodeSupport())

	// Test xterm
	os.Unsetenv("LANG")
	os.Setenv("TERM", "xterm")
	assert.True(t, detectUnicodeSupport())
}

func TestGetTerminalSize(t *testing.T) {
	// Save current environment
	cols := os.Getenv("COLUMNS")
	lines := os.Getenv("LINES")
	defer func() {
		os.Setenv("COLUMNS", cols)
		os.Setenv("LINES", lines)
	}()

	// Test with environment variables
	os.Setenv("COLUMNS", "120")
	os.Setenv("LINES", "40")

	w, h := getTerminalSize()
	assert.Equal(t, 120, w)
	assert.Equal(t, 40, h)

	// Test defaults
	os.Unsetenv("COLUMNS")
	os.Unsetenv("LINES")
	w, h = getTerminalSize()
	assert.Equal(t, 80, w)  // default
	assert.Equal(t, 24, h)  // default
}

func TestGetRenderConfigForClient_OpenCode(t *testing.T) {
	config := GetRenderConfigForClient(CLIClientOpenCode)

	assert.Equal(t, RenderStyleTheater, config.Style)
	assert.Equal(t, ProgressBarStyleUnicode, config.ProgressStyle)
	assert.True(t, config.ShowResources)
	assert.True(t, config.ShowLogs)
}

func TestGetRenderConfigForClient_Crush(t *testing.T) {
	config := GetRenderConfigForClient(CLIClientCrush)

	assert.Equal(t, RenderStyleMinimal, config.Style)
	assert.False(t, config.ShowResources)
	assert.Equal(t, 3, config.LogLines)
}

func TestGetRenderConfigForClient_HelixCode(t *testing.T) {
	config := GetRenderConfigForClient(CLIClientHelixCode)

	assert.Equal(t, RenderStyleTheater, config.Style)
	assert.Equal(t, ProgressBarStyleUnicode, config.ProgressStyle)
	assert.True(t, config.ShowResources)
	assert.True(t, config.ShowLogs)
}

func TestGetRenderConfigForClient_KiloCode(t *testing.T) {
	config := GetRenderConfigForClient(CLIClientKiloCode)

	assert.Equal(t, RenderStyleScreenplay, config.Style)
	assert.Equal(t, ProgressBarStyleBlock, config.ProgressStyle)
}

func TestFormatForClient_NoColor(t *testing.T) {
	// Save current environment
	noColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", noColor)

	os.Setenv("NO_COLOR", "1")

	content := "\033[31mRed Text\033[0m"
	result := FormatForClient(CLIClientOpenCode, content)

	// ANSI codes should be stripped
	assert.NotContains(t, result, "\033[31m")
	assert.Contains(t, result, "Red Text")
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"\033[31mRed\033[0m", "Red"},
		{"\033[1m\033[32mBold Green\033[0m", "Bold Green"},
		{"No colors", "No colors"},
		{"\033[0m\033[0m", ""},
		{"", ""},
	}

	for _, tc := range tests {
		result := stripANSI(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %q", tc.input)
	}
}

func TestConvertToASCII(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{BoxHorizontal, "-"},
		{BoxVertical, "|"},
		{BoxTopLeft, "+"},
		{BoxDoubleHorizontal, "="},
		{ProgressFilled, "#"},
		{ProgressEmpty, "."},
		{IconPending, "o"},
		{IconRunning, "*"},
		{IconCompleted, "+"},
		{IconFailed, "x"},
		{"Normal text", "Normal text"},
	}

	for _, tc := range tests {
		result := convertToASCII(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %q", tc.input)
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten!", 12, "exactly ten!"},
		{"this is a very long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tc := range tests {
		result := truncateString(tc.input, tc.maxLen)
		assert.Equal(t, tc.expected, result, "Input: %q, MaxLen: %d", tc.input, tc.maxLen)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0B"},
		{500, "500B"},
		{1023, "1023B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1073741824, "1.0GB"},
		{1099511627776, "1.0TB"},
	}

	for _, tc := range tests {
		result := formatBytes(tc.bytes)
		assert.Equal(t, tc.expected, result, "Bytes: %d", tc.bytes)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{1 * time.Second, "1.0s"},
		{30 * time.Second, "30.0s"},
		{90 * time.Second, "1m30s"},
		{5 * time.Minute, "5m0s"},
		{65 * time.Minute, "1h5m"},
		{2 * time.Hour, "2h0m"},
	}

	for _, tc := range tests {
		result := formatDuration(tc.duration)
		assert.Equal(t, tc.expected, result, "Duration: %v", tc.duration)
	}
}

func TestGetLevelColor(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"ERROR", ColorRed},
		{"error", ColorRed},
		{"FATAL", ColorRed},
		{"PANIC", ColorRed},
		{"WARN", ColorYellow},
		{"WARNING", ColorYellow},
		{"INFO", ColorCyan},
		{"DEBUG", ColorDim},
		{"TRACE", ColorDim},
		{"OTHER", ColorWhite},
		{"", ColorWhite},
	}

	for _, tc := range tests {
		result := getLevelColor(tc.level)
		assert.Equal(t, tc.expected, result, "Level: %s", tc.level)
	}
}

// =============================================================================
// Color/ANSI Code Constants Tests
// =============================================================================

func TestANSIColorConstants(t *testing.T) {
	assert.Equal(t, "\033[0m", ColorReset)
	assert.Equal(t, "\033[1m", ColorBold)
	assert.Equal(t, "\033[2m", ColorDim)
	assert.Equal(t, "\033[31m", ColorRed)
	assert.Equal(t, "\033[32m", ColorGreen)
	assert.Equal(t, "\033[33m", ColorYellow)
	assert.Equal(t, "\033[34m", ColorBlue)
	assert.Equal(t, "\033[36m", ColorCyan)
}

func TestBoxDrawingConstants(t *testing.T) {
	assert.Equal(t, "─", BoxHorizontal)
	assert.Equal(t, "│", BoxVertical)
	assert.Equal(t, "┌", BoxTopLeft)
	assert.Equal(t, "┐", BoxTopRight)
	assert.Equal(t, "└", BoxBottomLeft)
	assert.Equal(t, "┘", BoxBottomRight)
	assert.Equal(t, "═", BoxDoubleHorizontal)
	assert.Equal(t, "║", BoxDoubleVertical)
}

func TestProgressCharConstants(t *testing.T) {
	assert.Equal(t, "█", ProgressFilled)
	assert.Equal(t, "░", ProgressEmpty)
	assert.Equal(t, "=", ProgressASCIIFilled)
	assert.Equal(t, " ", ProgressASCIIEmpty)
	assert.Equal(t, ">", ProgressASCIITip)
}

func TestStatusIconConstants(t *testing.T) {
	assert.Equal(t, "○", IconPending)
	assert.Equal(t, "◉", IconRunning)
	assert.Equal(t, "✓", IconCompleted)
	assert.Equal(t, "✗", IconFailed)
	assert.Equal(t, "⚠", IconStuck)
	assert.Equal(t, "⊘", IconCancelled)
	assert.Equal(t, "⏸", IconPaused)
}

// =============================================================================
// Edge Cases and Concurrent Access Tests
// =============================================================================

func TestRenderer_RenderProgressBar_ZeroProgress(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &ProgressBarContent{
		Progress: 0.0,
		Message:  "Starting",
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "0.0%")
}

func TestRenderer_RenderProgressBar_FullProgress(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &ProgressBarContent{
		Progress: 100.0,
		Message:  "Complete",
	}

	output := r.RenderProgressBar(content)
	assert.Contains(t, output, "100.0%")
}

func TestRenderer_RenderResourceGauge_HighCPU(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &ResourceGaugeContent{
		CPUPercent:    95.0, // Should trigger red color
		MemoryPercent: 50.0,
		MemoryBytes:   500 * 1024 * 1024,
		MemoryMax:     1024 * 1024 * 1024,
	}

	output := r.RenderResourceGauge(content)
	assert.Contains(t, output, "CPU:")
	// Should contain red color code
	assert.Contains(t, output, ColorRed)
}

func TestRenderer_RenderResourceGauge_HighMemory(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = true

	content := &ResourceGaugeContent{
		CPUPercent:    50.0,
		MemoryPercent: 92.0, // Should trigger red color
		MemoryBytes:   920 * 1024 * 1024,
		MemoryMax:     1024 * 1024 * 1024,
	}

	output := r.RenderResourceGauge(content)
	assert.Contains(t, output, "Memory:")
}

func TestRenderer_ConcurrentSpinnerFrame(t *testing.T) {
	r := NewRenderer(nil, CLIClientOpenCode)

	done := make(chan struct{})
	const numGoroutines = 10
	const framesPerGoroutine = 100

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < framesPerGoroutine; j++ {
				frame := r.GetSpinnerFrame()
				// Should be a valid spinner frame
				found := false
				for _, f := range SpinnerFrames {
					if f == frame {
						found = true
						break
					}
				}
				assert.True(t, found, "Invalid spinner frame: %s", frame)
			}
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestRenderer_RenderStatusTable_LongNames(t *testing.T) {
	config := DefaultRenderConfig()
	r := NewRenderer(config, CLIClientOpenCode)
	r.isTTY = false

	content := &StatusTableContent{
		Tasks: []TaskStatusRow{
			{
				ID:     "1",
				Name:   "This is a very long task name that should be truncated",
				Type:   "very-long-type-name",
				Status: "running",
			},
		},
	}

	output := r.RenderStatusTable(content)
	// Names should be truncated, not cause overflow
	assert.True(t, len(strings.Split(output, "\n")[3]) < 150)
}
